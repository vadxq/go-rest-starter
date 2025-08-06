package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Message 队列消息
type Message struct {
	ID        string          `json:"id"`
	Topic     string          `json:"topic"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
	Retries   int             `json:"retries"`
	MaxRetries int            `json:"max_retries"`
}

// Handler 消息处理器
type Handler func(ctx context.Context, msg *Message) error

// Queue 队列接口
type Queue interface {
	// Publish 发布消息
	Publish(ctx context.Context, topic string, payload interface{}) error
	// Subscribe 订阅主题
	Subscribe(ctx context.Context, topic string, handler Handler) error
	// PublishDelayed 发布延迟消息
	PublishDelayed(ctx context.Context, topic string, payload interface{}, delay time.Duration) error
	// Close 关闭队列
	Close() error
}

// RedisQueue Redis队列实现
type RedisQueue struct {
	client      *redis.Client
	handlers    map[string][]Handler
	mu          sync.RWMutex
	workerPool  chan struct{}
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewRedisQueue 创建Redis队列
func NewRedisQueue(client *redis.Client, maxWorkers int) Queue {
	ctx, cancel := context.WithCancel(context.Background())
	
	rq := &RedisQueue{
		client:     client,
		handlers:   make(map[string][]Handler),
		workerPool: make(chan struct{}, maxWorkers),
		ctx:        ctx,
		cancel:     cancel,
	}
	
	// 初始化工作池
	for i := 0; i < maxWorkers; i++ {
		rq.workerPool <- struct{}{}
	}
	
	// 启动延迟消息处理器
	rq.wg.Add(1)
	go rq.processDelayedMessages()
	
	return rq
}

// Publish 发布消息
func (rq *RedisQueue) Publish(ctx context.Context, topic string, payload interface{}) error {
	// 序列化payload
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	// 创建消息
	msg := &Message{
		ID:        generateMessageID(),
		Topic:     topic,
		Payload:   data,
		Timestamp: time.Now(),
		Retries:   0,
		MaxRetries: 3,
	}
	
	// 序列化消息
	msgData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	// 发布到Redis
	key := fmt.Sprintf("queue:%s", topic)
	if err := rq.client.LPush(ctx, key, msgData).Err(); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	
	return nil
}

// Subscribe 订阅主题
func (rq *RedisQueue) Subscribe(ctx context.Context, topic string, handler Handler) error {
	// 注册处理器
	rq.mu.Lock()
	rq.handlers[topic] = append(rq.handlers[topic], handler)
	rq.mu.Unlock()
	
	// 启动消费者
	rq.wg.Add(1)
	go rq.consume(topic)
	
	return nil
}

// consume 消费消息
func (rq *RedisQueue) consume(topic string) {
	defer rq.wg.Done()
	
	key := fmt.Sprintf("queue:%s", topic)
	
	for {
		select {
		case <-rq.ctx.Done():
			return
		default:
			// 从队列中获取消息（阻塞1秒）
			result, err := rq.client.BRPop(rq.ctx, time.Second, key).Result()
			if err != nil {
				if err == redis.Nil {
					continue // 超时，继续等待
				}
				// 记录错误并继续
				continue
			}
			
			if len(result) < 2 {
				continue
			}
			
			// 获取工作令牌
			<-rq.workerPool
			
			// 异步处理消息
			go func(data string) {
				defer func() {
					rq.workerPool <- struct{}{} // 归还工作令牌
				}()
				
				// 反序列化消息
				var msg Message
				if err := json.Unmarshal([]byte(data), &msg); err != nil {
					return
				}
				
				// 处理消息
				rq.processMessage(&msg)
			}(result[1])
		}
	}
}

// processMessage 处理消息
func (rq *RedisQueue) processMessage(msg *Message) {
	rq.mu.RLock()
	handlers := rq.handlers[msg.Topic]
	rq.mu.RUnlock()
	
	for _, handler := range handlers {
		ctx, cancel := context.WithTimeout(rq.ctx, 30*time.Second)
		err := handler(ctx, msg)
		cancel()
		
		if err != nil {
			// 处理失败，重试
			if msg.Retries < msg.MaxRetries {
				msg.Retries++
				rq.retryMessage(msg)
			} else {
				// 超过最大重试次数，发送到死信队列
				rq.sendToDeadLetter(msg, err)
			}
		}
	}
}

// retryMessage 重试消息
func (rq *RedisQueue) retryMessage(msg *Message) {
	// 计算重试延迟（指数退避）
	delay := time.Duration(msg.Retries) * time.Second * 2
	
	// 发布延迟消息
	rq.PublishDelayed(rq.ctx, msg.Topic, msg, delay)
}

// sendToDeadLetter 发送到死信队列
func (rq *RedisQueue) sendToDeadLetter(msg *Message, err error) {
	deadLetterKey := fmt.Sprintf("dead_letter:%s", msg.Topic)
	
	// 添加错误信息
	type DeadLetterMessage struct {
		*Message
		Error     string    `json:"error"`
		FailedAt  time.Time `json:"failed_at"`
	}
	
	dlMsg := &DeadLetterMessage{
		Message:  msg,
		Error:    err.Error(),
		FailedAt: time.Now(),
	}
	
	data, _ := json.Marshal(dlMsg)
	rq.client.LPush(rq.ctx, deadLetterKey, data)
}

// PublishDelayed 发布延迟消息
func (rq *RedisQueue) PublishDelayed(ctx context.Context, topic string, payload interface{}, delay time.Duration) error {
	// 序列化payload
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	// 创建消息
	msg := &Message{
		ID:        generateMessageID(),
		Topic:     topic,
		Payload:   data,
		Timestamp: time.Now(),
	}
	
	// 序列化消息
	msgData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	// 添加到延迟队列（使用有序集合）
	score := float64(time.Now().Add(delay).Unix())
	key := "delayed_queue"
	if err := rq.client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: msgData,
	}).Err(); err != nil {
		return fmt.Errorf("failed to publish delayed message: %w", err)
	}
	
	return nil
}

// processDelayedMessages 处理延迟消息
func (rq *RedisQueue) processDelayedMessages() {
	defer rq.wg.Done()
	
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-rq.ctx.Done():
			return
		case <-ticker.C:
			// 获取到期的消息
			now := float64(time.Now().Unix())
			key := "delayed_queue"
			
			// 获取所有到期的消息
			messages, err := rq.client.ZRangeByScore(rq.ctx, key, &redis.ZRangeBy{
				Min: "0",
				Max: fmt.Sprintf("%f", now),
			}).Result()
			
			if err != nil {
				continue
			}
			
			for _, msgData := range messages {
				// 反序列化消息
				var msg Message
				if err := json.Unmarshal([]byte(msgData), &msg); err != nil {
					continue
				}
				
				// 发布到正常队列
				if err := rq.Publish(rq.ctx, msg.Topic, msg.Payload); err != nil {
					continue
				}
				
				// 从延迟队列中删除
				rq.client.ZRem(rq.ctx, key, msgData)
			}
		}
	}
}

// Close 关闭队列
func (rq *RedisQueue) Close() error {
	rq.cancel()
	rq.wg.Wait()
	return nil
}

// generateMessageID 生成消息ID
func generateMessageID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Nanosecond())
}


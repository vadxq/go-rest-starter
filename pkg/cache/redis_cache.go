package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis缓存实现
type redisCache struct {
	client            *redis.Client
	defaultExpiration time.Duration
}

// 创建Redis缓存
func newRedisCache(opts Options) (Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     opts.RedisAddress,
		Password: opts.RedisPassword,
		DB:       opts.RedisDB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("无法连接到Redis: %w", err)
	}

	return &redisCache{
		client:            client,
		defaultExpiration: opts.DefaultExpiration,
	}, nil
}

// 获取缓存
func (c *redisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return val, nil
}

// 设置缓存
func (c *redisCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	if expiration == 0 {
		expiration = c.defaultExpiration
	}
	
	return c.client.Set(ctx, key, value, expiration).Err()
}

// 删除缓存
func (c *redisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// 清空缓存
func (c *redisCache) Clear(ctx context.Context) error {
	return c.client.FlushAll(ctx).Err()
}

// 获取对象
func (c *redisCache) GetObject(ctx context.Context, key string, value interface{}) error {
	data, err := c.Get(ctx, key)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(data, value)
}

// 设置对象
func (c *redisCache) SetObject(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	
	return c.Set(ctx, key, data, expiration)
} 
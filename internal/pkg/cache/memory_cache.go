package cache

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

var (
	// ErrNotFound 表示键在缓存中不存在
	ErrNotFound = errors.New("键不存在")
	
	// ErrExpired 表示缓存项已过期
	ErrExpired = errors.New("缓存项已过期")
)

// 内存缓存项
type item struct {
	value      []byte
	expiration int64
}

// 检查是否过期
func (i *item) isExpired() bool {
	if i.expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > i.expiration
}

// 内存缓存实现
type memoryCache struct {
	items             sync.Map
	defaultExpiration time.Duration
	janitor           *janitor
}

// 创建内存缓存
func newMemoryCache(opts Options) (Cache, error) {
	cache := &memoryCache{
		defaultExpiration: opts.DefaultExpiration,
	}
	
	// 如果设置了清理间隔，启动清理协程
	if opts.CleanupInterval > 0 {
		cache.janitor = newJanitor(opts.CleanupInterval)
		cache.janitor.run(cache)
	}
	
	return cache, nil
}

// 获取缓存
func (c *memoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	value, ok := c.items.Load(key)
	if !ok {
		return nil, ErrNotFound
	}
	
	item, ok := value.(*item)
	if !ok {
		return nil, errors.New("无效的缓存项类型")
	}
	
	if item.isExpired() {
		c.items.Delete(key)
		return nil, ErrExpired
	}
	
	return item.value, nil
}

// 设置缓存
func (c *memoryCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	var exp int64
	
	if expiration == 0 {
		expiration = c.defaultExpiration
	}
	
	if expiration > 0 {
		exp = time.Now().Add(expiration).UnixNano()
	}
	
	c.items.Store(key, &item{
		value:      value,
		expiration: exp,
	})
	
	return nil
}

// 删除缓存
func (c *memoryCache) Delete(ctx context.Context, key string) error {
	c.items.Delete(key)
	return nil
}

// 清空缓存
func (c *memoryCache) Clear(ctx context.Context) error {
	c.items = sync.Map{}
	return nil
}

// 获取对象
func (c *memoryCache) GetObject(ctx context.Context, key string, value interface{}) error {
	data, err := c.Get(ctx, key)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(data, value)
}

// 设置对象
func (c *memoryCache) SetObject(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	
	return c.Set(ctx, key, data, expiration)
}

// 清理过期项的协程
type janitor struct {
	interval time.Duration
	stopCh   chan bool
}

// 创建清理协程
func newJanitor(interval time.Duration) *janitor {
	return &janitor{
		interval: interval,
		stopCh:   make(chan bool),
	}
}

// 运行清理协程
func (j *janitor) run(cache *memoryCache) {
	ticker := time.NewTicker(j.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				cache.deleteExpired()
			case <-j.stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

// 停止清理协程
func (j *janitor) stop() {
	j.stopCh <- true
}

// 删除过期项
func (c *memoryCache) deleteExpired() {
	c.items.Range(func(key, value interface{}) bool {
		item, ok := value.(*item)
		if !ok {
			c.items.Delete(key)
			return true
		}
		
		if item.isExpired() {
			c.items.Delete(key)
		}
		
		return true
	})
} 
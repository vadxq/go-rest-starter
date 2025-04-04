package cache

import (
	"context"
	"time"
)

// Cache 定义缓存接口
type Cache interface {
	// Get 从缓存中获取值
	Get(ctx context.Context, key string) ([]byte, error)
	
	// Set 设置缓存值
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
	
	// Delete 从缓存中删除特定键
	Delete(ctx context.Context, key string) error
	
	// Clear 清空缓存
	Clear(ctx context.Context) error
	
	// GetObject 获取并解析为指定类型的对象
	GetObject(ctx context.Context, key string, value interface{}) error
	
	// SetObject 将对象序列化后存入缓存
	SetObject(ctx context.Context, key string, value interface{}, expiration time.Duration) error
}

// Options 缓存选项
type Options struct {
	// Redis地址
	RedisAddress string
	
	// Redis密码
	RedisPassword string
	
	// Redis数据库
	RedisDB int
	
	// 默认过期时间
	DefaultExpiration time.Duration
	
	// 清理间隔
	CleanupInterval time.Duration
}

// NewCache 创建缓存实例
func NewCache(opts Options) (Cache, error) {
	if opts.RedisAddress != "" {
		return newRedisCache(opts)
	}
	return newMemoryCache(opts)
} 
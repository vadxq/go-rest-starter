package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// CacheAside Cache-Aside模式（最常用的缓存模式）
type CacheAside struct {
	cache  Cache
	loader DataLoader
	ttl    time.Duration
}

// DataLoader 数据加载器
type DataLoader func(ctx context.Context, key string) (interface{}, error)

// NewCacheAside 创建Cache-Aside模式缓存
func NewCacheAside(cache Cache, loader DataLoader, ttl time.Duration) *CacheAside {
	return &CacheAside{
		cache:  cache,
		loader: loader,
		ttl:    ttl,
	}
}

// Get 获取数据（先查缓存，缓存没有则从数据源加载）
func (ca *CacheAside) Get(ctx context.Context, key string, dest interface{}) error {
	// 先从缓存获取
	err := ca.cache.GetObject(ctx, key, dest)
	if err == nil {
		return nil
	}
	
	// 缓存未命中，从数据源加载
	data, err := ca.loader(ctx, key)
	if err != nil {
		return err
	}
	
	// 写入缓存（异步，避免阻塞）
	go ca.cache.SetObject(context.Background(), key, data, ca.ttl)
	
	// 将数据复制到目标
	return copyValue(data, dest)
}

// Invalidate 失效缓存
func (ca *CacheAside) Invalidate(ctx context.Context, key string) error {
	return ca.cache.Delete(ctx, key)
}

// SingleFlight 防止缓存击穿（同一时间只允许一个请求去加载数据）
type SingleFlight struct {
	cache   Cache
	loader  DataLoader
	ttl     time.Duration
	flights map[string]*flightGroup
	mu      sync.Mutex
}

type flightGroup struct {
	wg   sync.WaitGroup
	val  interface{}
	err  error
}

// NewSingleFlight 创建SingleFlight缓存
func NewSingleFlight(cache Cache, loader DataLoader, ttl time.Duration) *SingleFlight {
	return &SingleFlight{
		cache:   cache,
		loader:  loader,
		ttl:     ttl,
		flights: make(map[string]*flightGroup),
	}
}

// Get 获取数据（防止缓存击穿）
func (sf *SingleFlight) Get(ctx context.Context, key string, dest interface{}) error {
	// 先从缓存获取
	err := sf.cache.GetObject(ctx, key, dest)
	if err == nil {
		return nil
	}
	
	// 检查是否有正在进行的加载
	sf.mu.Lock()
	if fg, ok := sf.flights[key]; ok {
		sf.mu.Unlock()
		// 等待加载完成
		fg.wg.Wait()
		if fg.err != nil {
			return fg.err
		}
		return copyValue(fg.val, dest)
	}
	
	// 创建新的flight group
	fg := &flightGroup{}
	fg.wg.Add(1)
	sf.flights[key] = fg
	sf.mu.Unlock()
	
	// 加载数据
	fg.val, fg.err = sf.loader(ctx, key)
	if fg.err == nil {
		// 写入缓存
		sf.cache.SetObject(ctx, key, fg.val, sf.ttl)
		copyValue(fg.val, dest)
	}
	
	// 标记完成
	fg.wg.Done()
	
	// 清理flight group
	sf.mu.Lock()
	delete(sf.flights, key)
	sf.mu.Unlock()
	
	return fg.err
}

// copyValue 复制值（简单的JSON序列化/反序列化）
func copyValue(src, dest interface{}) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}
# Cache Package

## Design Decision: Redis-Only Cache

This package exclusively uses Redis for caching, without any in-memory cache implementation. This design decision was made for the following reasons:

### Why No Memory Cache?

1. **Memory Leak Prevention**
   - In-memory caches can lead to memory leaks if not properly managed
   - Growing cache size can cause out-of-memory issues in production
   - Difficult to control memory usage across different deployment environments

2. **Consistency in Distributed Systems**
   - Memory caches are local to each instance
   - In a multi-instance deployment, each instance would have different cached data
   - Redis provides a single source of truth for all instances

3. **Better Monitoring and Control**
   - Redis provides built-in monitoring tools
   - Easy to set memory limits and eviction policies
   - Clear visibility into cache usage and performance

4. **Data Persistence Options**
   - Redis can persist data to disk if needed
   - Graceful recovery after restarts
   - No cache warming required after deployment

5. **Production Best Practices**
   - Most production systems use Redis or similar external cache
   - Easier to scale horizontally
   - Better resource isolation

## Usage

```go
// Initialize cache (requires Redis)
cacheOpts := cache.Options{
    RedisAddress:      "localhost:6379",
    RedisPassword:     "password",
    RedisDB:           0,
    DefaultExpiration: 10 * time.Minute,
}

cache, err := cache.NewCache(cacheOpts)
if err != nil {
    // Handle error - cache is not available
    log.Warn("Cache not available, continuing without cache")
}

// Use cache
if cache != nil {
    // Set value
    err = cache.Set(ctx, "key", []byte("value"), 5*time.Minute)
    
    // Get value
    value, err := cache.Get(ctx, "key")
    
    // Set object
    user := &User{ID: 1, Name: "John"}
    err = cache.SetObject(ctx, "user:1", user, 10*time.Minute)
    
    // Get object
    var cachedUser User
    err = cache.GetObject(ctx, "user:1", &cachedUser)
}
```

## Cache Strategies

The package provides various caching strategies that all use Redis as the backend:

- **Cache-Aside Pattern**: Read from cache, if miss, read from database and update cache
- **Write-Through Pattern**: Write to both cache and database
- **Single Flight**: Prevent cache stampede by ensuring only one request loads data
- **Bloom Filter Cache**: Use bloom filter to prevent cache penetration

## Configuration

Redis connection is configured through environment variables or config file:

```yaml
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
```

## Fallback Behavior

If Redis is not available:
- The application continues to work without cache
- All cache operations return immediately without error
- Data is fetched directly from the database

This ensures the application remains functional even if the cache layer fails.
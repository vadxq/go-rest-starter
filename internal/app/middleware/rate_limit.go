package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	RequestsPerSecond int           // 每秒允许请求数
	Burst             int           // 突发请求数
	CleanupInterval   time.Duration // 清理过期记录的间隔
}

// DefaultRateLimitConfig 默认速率限制配置
var DefaultRateLimitConfig = RateLimitConfig{
	RequestsPerSecond: 10,
	Burst:             20,
	CleanupInterval:   10 * time.Minute,
}

// rateLimiter 速率限制器
type rateLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimitMiddleware 基于 IP 的速率限制中间件
type RateLimitMiddleware struct {
	config   RateLimitConfig
	limiters map[string]*rateLimiter
	mu       sync.RWMutex
}

// NewRateLimitMiddleware 创建新的速率限制中间件
func NewRateLimitMiddleware(config RateLimitConfig) *RateLimitMiddleware {
	rlm := &RateLimitMiddleware{
		config:   config,
		limiters: make(map[string]*rateLimiter),
	}

	// 启动清理 goroutine
	go rlm.cleanup()

	return rlm
}

// Handler 速率限制中间件处理函数
func (rlm *RateLimitMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 获取客户端 IP
		ip := getClientIP(r)

		// 获取或创建限制器
		limiter := rlm.getLimiter(ip)

		// 检查是否允许请求
		if !limiter.Allow() {
			writeRateLimitResponse(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getLimiter 获取或创建 IP 对应的限制器
func (rlm *RateLimitMiddleware) getLimiter(ip string) *rate.Limiter {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	limiterInfo, exists := rlm.limiters[ip]
	if !exists {
		limiterInfo = &rateLimiter{
			limiter: rate.NewLimiter(
				rate.Limit(rlm.config.RequestsPerSecond),
				rlm.config.Burst,
			),
			lastSeen: time.Now(),
		}
		rlm.limiters[ip] = limiterInfo
	} else {
		limiterInfo.lastSeen = time.Now()
	}

	return limiterInfo.limiter
}

// cleanup 定期清理过期的限制器
func (rlm *RateLimitMiddleware) cleanup() {
	ticker := time.NewTicker(rlm.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rlm.mu.Lock()
		cutoff := time.Now().Add(-rlm.config.CleanupInterval * 2)

		for ip, limiterInfo := range rlm.limiters {
			if limiterInfo.lastSeen.Before(cutoff) {
				delete(rlm.limiters, ip)
			}
		}
		rlm.mu.Unlock()
	}
}

// writeRateLimitResponse 写入速率限制响应
func writeRateLimitResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-RateLimit-Limit", "10")
	w.Header().Set("X-RateLimit-Remaining", "0")
	w.Header().Set("Retry-After", "60")
	w.WriteHeader(http.StatusTooManyRequests)
	
	response := `{
		"error": {
			"type": "RATE_LIMIT_EXCEEDED",
			"message": "请求频率过高，请稍后再试",
			"details": "Rate limit exceeded. Please try again later."
		}
	}`
	
	w.Write([]byte(response))
}

// getClientIP 获取客户端真实IP地址
func getClientIP(r *http.Request) string {
	// 检查 X-Forwarded-For 头
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// 取第一个IP地址
		if idx := len(xff); idx > 0 {
			return xff[:idx]
		}
	}

	// 检查 X-Real-IP 头
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 检查 X-Forwarded-For 头的第一个地址
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// 默认使用 RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
package middleware

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// Metrics 基础性能指标
type Metrics struct {
	TotalRequests   atomic.Uint64
	ActiveRequests  atomic.Int64
	TotalErrors     atomic.Uint64
	StartTime       time.Time
}

// GlobalMetrics 全局指标实例
var GlobalMetrics = &Metrics{
	StartTime: time.Now(),
}

// MonitoringMiddleware 监控中间件（简化版）
func MonitoringMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// 增加计数
		GlobalMetrics.TotalRequests.Add(1)
		GlobalMetrics.ActiveRequests.Add(1)
		defer GlobalMetrics.ActiveRequests.Add(-1)
		
		// 包装响应写入器
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		
		// 执行请求
		next.ServeHTTP(ww, r)
		
		// 记录错误
		if ww.Status() >= 400 {
			GlobalMetrics.TotalErrors.Add(1)
		}
		
		// 添加响应时间头
		duration := time.Since(start)
		w.Header().Set("X-Response-Time", strconv.FormatInt(duration.Milliseconds(), 10)+"ms")
	})
}

// GetMetricsSnapshot 获取指标快照
func GetMetricsSnapshot() MetricsSnapshot {
	uptime := time.Since(GlobalMetrics.StartTime)
	total := GlobalMetrics.TotalRequests.Load()
	errors := GlobalMetrics.TotalErrors.Load()
	
	var errorRate float64
	if total > 0 {
		errorRate = float64(errors) / float64(total) * 100
	}
	
	return MetricsSnapshot{
		TotalRequests:  total,
		ActiveRequests: GlobalMetrics.ActiveRequests.Load(),
		TotalErrors:    errors,
		ErrorRate:      errorRate,
		Uptime:         uptime,
		QPS:            float64(total) / uptime.Seconds(),
	}
}

// MetricsSnapshot 指标快照
type MetricsSnapshot struct {
	TotalRequests  uint64        `json:"total_requests"`
	ActiveRequests int64         `json:"active_requests"`
	TotalErrors    uint64        `json:"total_errors"`
	ErrorRate      float64       `json:"error_rate"`
	Uptime         time.Duration `json:"uptime_seconds"`
	QPS            float64       `json:"qps"`
}

// MetricsHandler 指标端点处理器
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := GetMetricsSnapshot()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
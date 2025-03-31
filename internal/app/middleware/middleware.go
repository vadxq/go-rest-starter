package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// TraceIDKeyType 定义 TraceID 的上下文 Key 类型
type TraceIDKeyType struct{}

// TraceIDKey 是用于在上下文中存储 TraceID 的 Key
var TraceIDKey TraceIDKeyType

// GetTraceID 从上下文中获取 TraceID
func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// TraceID 是一个中间件，用于生成和传递请求的唯一 ID
func TraceID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Request-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), TraceIDKey, traceID)
		w.Header().Set("X-Request-ID", traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware 是一个自定义的日志中间件
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		latency := time.Since(start)

		log.Info().
			Str("proto", r.Proto).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", ww.Status()).
			Str("latency", latency.String()).
			Str("ip", r.RemoteAddr).
			Str("user-agent", r.UserAgent()).
			Str("trace_id", GetTraceID(r.Context())).
			Msg("请求处理完成")
	})
}

// CORSMiddleware 处理跨域请求
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RecoveryMiddleware 恢复中间件，处理 panic
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Error().
					Str("trace_id", GetTraceID(r.Context())).
					Interface("error", err).
					Msg("服务器发生 panic")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"code":500,"message":"内部服务器错误"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

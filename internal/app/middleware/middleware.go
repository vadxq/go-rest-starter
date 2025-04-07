package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// 上下文键类型
type contextKey string

const (
	// 请求上下文键
	traceIDKey   contextKey = "trace_id"    // 请求跟踪ID
	userIDKey    contextKey = "user_id"     // 用户ID
	userRoleKey  contextKey = "user_role"   // 用户角色
	requestIDKey contextKey = "request_id"  // 请求ID
	clientIPKey  contextKey = "client_ip"   // 客户端IP
	startTimeKey contextKey = "start_time"  // 请求开始时间
)

// GetTraceID 从上下文中获取 TraceID
func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// GetUserIDFromContext 从上下文中获取用户ID（使用contextKey方式获取）
func GetUserIDFromContext(ctx context.Context) uint {
	if ctx == nil {
		return 0
	}
	if userID, ok := ctx.Value(userIDKey).(uint); ok {
		return userID
	}
	return 0
}

// GetUserRole 从上下文中获取用户角色
func GetUserRole(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if role, ok := ctx.Value(userRoleKey).(string); ok {
		return role
	}
	return ""
}

// SetUserInfo 在上下文中设置用户信息
func SetUserInfo(ctx context.Context, userID uint, role string) context.Context {
	ctx = context.WithValue(ctx, userIDKey, userID)
	return context.WithValue(ctx, userRoleKey, role)
}

// GetClientIP 从上下文中获取客户端IP
func GetClientIP(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if ip, ok := ctx.Value(clientIPKey).(string); ok {
		return ip
	}
	return ""
}

// GetRequestLatency 获取请求延迟时间
func GetRequestLatency(ctx context.Context) time.Duration {
	if ctx == nil {
		return 0
	}
	if startTime, ok := ctx.Value(startTimeKey).(time.Time); ok {
		return time.Since(startTime)
	}
	return 0
}

// RequestContext 请求上下文中间件，设置请求相关信息
func RequestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置请求开始时间
		ctx := context.WithValue(r.Context(), startTimeKey, time.Now())
		
		// 设置客户端IP
		clientIP := r.Header.Get("X-Forwarded-For")
		if clientIP == "" {
			clientIP = r.RemoteAddr
		}
		ctx = context.WithValue(ctx, clientIPKey, clientIP)
		
		// 生成请求ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		ctx = context.WithValue(ctx, requestIDKey, requestID)
		
		// 设置跟踪ID（兼容现有代码）
		traceID := requestID
		ctx = context.WithValue(ctx, traceIDKey, traceID)
		
		// 设置响应头
		w.Header().Set("X-Request-ID", requestID)
		
		// 继续处理请求
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware 是一个自定义的日志中间件
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 包装响应写入器以获取状态码
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		
		// 处理请求
		next.ServeHTTP(ww, r)
		
		// 计算请求处理延迟
		latency := GetRequestLatency(r.Context())
		
		// 构建日志事件
		logEvent := log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("query", r.URL.RawQuery).
			Int("status", ww.Status()).
			Str("latency", latency.String()).
			Int("size", ww.BytesWritten()).
			Str("ip", GetClientIP(r.Context())).
			Str("user_agent", r.UserAgent()).
			Str("trace_id", GetTraceID(r.Context()))
		
		// 添加用户信息（如果有）
		userID, ok := GetUserID(r.Context())
		if ok && userID != 0 {
			logEvent = logEvent.Uint("user_id", userID)
		}
		
		// 记录日志
		logEvent.Msg(fmt.Sprintf("%s %s - %d", r.Method, r.URL.Path, ww.Status()))
	})
}

// CORSMiddleware 处理跨域请求
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
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
				// 记录错误日志
				log.Error().
					Str("trace_id", GetTraceID(r.Context())).
					Str("path", r.URL.Path).
					Str("method", r.Method).
					Interface("error", err).
					Msg("服务器发生 panic")
				
				// 返回错误响应
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				traceID := GetTraceID(r.Context())
				errResp := fmt.Sprintf(`{"code":500,"message":"内部服务器错误","error_code":"internal_server_error","trace_id":"%s","timestamp":%d}`, 
					traceID, time.Now().Unix())
				w.Write([]byte(errResp))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// TraceID 是一个中间件，用于生成和传递请求的唯一 ID（向后兼容）
func TraceID(next http.Handler) http.Handler {
	return RequestContext(next)
}

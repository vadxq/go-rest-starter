package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	apperrors "github.com/vadxq/go-rest-starter/pkg/errors"
)

// 上下文键类型
type contextKey string

const (
	// 请求上下文键
	reqContextKey contextKey = "request_context" // 请求上下文对象
)

// ReqContext 请求上下文结构体
type ReqContext struct {
	TraceID    string    // 请求跟踪ID
	RequestID  string    // 请求ID
	UserID     uint      // 用户ID (如果已认证)
	UserRole   string    // 用户角色 (如果已认证)
	ClientIP   string    // 客户端IP
	StartTime  time.Time // 请求开始时间
	RequestURI string    // 请求URI
	Method     string    // 请求方法
}

// GetUserIDFromContext 从请求上下文中获取用户ID
func GetUserIDFromContext(ctx context.Context) (uint, bool) {
	reqCtx := GetRequestContext(ctx)
	if reqCtx == nil || reqCtx.UserID == 0 {
		return 0, false
	}
	return reqCtx.UserID, true
}

// GetRequestContext 从context.Context获取请求上下文
func GetRequestContext(ctx context.Context) *ReqContext {
	if ctx == nil {
		return nil
	}
	if rc, ok := ctx.Value(reqContextKey).(*ReqContext); ok {
		return rc
	}
	return nil
}

// RequestContext 请求上下文中间件，设置请求相关信息
func RequestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 创建请求上下文
		reqCtx := &ReqContext{
			RequestID:  r.Header.Get("X-Request-ID"),
			ClientIP:   r.Header.Get("X-Forwarded-For"),
			StartTime:  time.Now(),
			RequestURI: r.RequestURI,
			Method:     r.Method,
		}

		// 如果没有请求ID，则生成一个
		if reqCtx.RequestID == "" {
			reqCtx.RequestID = middleware.GetReqID(r.Context())
		}

		// 设置跟踪ID与请求ID相同
		reqCtx.TraceID = reqCtx.RequestID

		// 如果没有客户端IP，则使用RemoteAddr
		if reqCtx.ClientIP == "" {
			reqCtx.ClientIP = r.RemoteAddr
		}

		// 设置响应头
		w.Header().Set("X-Request-ID", reqCtx.RequestID)

		// 将请求上下文添加到请求上下文
		ctx := context.WithValue(r.Context(), reqContextKey, reqCtx)

		// 继续处理请求
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware 日志中间件，记录请求日志
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 获取请求上下文
		reqCtx := GetRequestContext(r.Context())
		if reqCtx == nil {
			// 如果没有请求上下文，则创建一个
			reqCtx = &ReqContext{
				StartTime:  time.Now(),
				RequestURI: r.RequestURI,
				Method:     r.Method,
			}
		}

		// 获取请求主体大小
		var requestSize int64
		if r.ContentLength > 0 {
			requestSize = r.ContentLength
		}

		// 包装响应写入器以获取状态码
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		// 处理请求
		next.ServeHTTP(ww, r)

		// 计算请求处理延迟
		latency := time.Since(reqCtx.StartTime)

		// 构建日志事件参数
		args := []interface{}{
			"method", reqCtx.Method,
			"path", reqCtx.RequestURI,
			"query", r.URL.RawQuery,
			"status", ww.Status(),
			"latency", latency.String(),
			"size", ww.BytesWritten(),
			"req_size", requestSize,
			"ip", reqCtx.ClientIP,
			"user_agent", r.UserAgent(),
			"trace_id", reqCtx.TraceID,
		}

		// 添加用户信息（如果有）
		if reqCtx.UserID != 0 {
			args = append(args, "user_id", reqCtx.UserID)
		}

		// 记录日志
		slog.Info(fmt.Sprintf("%s %s - %d", reqCtx.Method, reqCtx.RequestURI, ww.Status()), args...)
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
		defer apperrors.RecoverPanicWithCallback("HTTP请求处理", func(err interface{}, stack []byte) {
			// 获取请求上下文
			reqCtx := GetRequestContext(r.Context())

			// 返回错误响应
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)

			// 构建错误响应
			response := struct {
				Success bool `json:"success"`
				Error   struct {
					Type    string `json:"type"`
					Message string `json:"message"`
				} `json:"error"`
			}{
				Success: false,
				Error: struct {
					Type    string `json:"type"`
					Message string `json:"message"`
				}{
					Type:    "INTERNAL_ERROR",
					Message: "服务器内部错误",
				},
			}

			// 如果有跟踪ID，添加到响应中
			if reqCtx != nil && reqCtx.TraceID != "" {
				response.Error.Message = "服务器内部错误，请稍后重试"
			}

			json.NewEncoder(w).Encode(response)
		})

		next.ServeHTTP(w, r)
	})
}

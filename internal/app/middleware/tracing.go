package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/vadxq/go-rest-starter/pkg/logger"
)

// TracingMiddleware 请求追踪中间件
func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 获取或生成请求ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = middleware.GetReqID(r.Context())
			if requestID == "" {
				requestID = fmt.Sprintf("%d", middleware.NextRequestID())
			}
		}

		// 获取或生成链路追踪ID
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			traceID = r.Header.Get("X-B3-TraceId") // 支持Zipkin B3格式
			if traceID == "" {
				traceID = requestID // 如果没有trace ID，使用request ID
			}
		}

		// 获取span ID（如果存在）
		spanID := r.Header.Get("X-Span-ID")
		if spanID == "" {
			spanID = r.Header.Get("X-B3-SpanId") // 支持Zipkin B3格式
		}

		// 获取parent span ID（如果存在）
		parentSpanID := r.Header.Get("X-Parent-Span-ID")
		if parentSpanID == "" {
			parentSpanID = r.Header.Get("X-B3-ParentSpanId") // 支持Zipkin B3格式
		}

		// 设置响应头
		w.Header().Set("X-Request-ID", requestID)
		w.Header().Set("X-Trace-ID", traceID)
		if spanID != "" {
			w.Header().Set("X-Span-ID", spanID)
		}

		// 创建带有追踪信息的上下文
		ctx := r.Context()
		ctx = logger.WithRequestID(ctx, requestID)
		ctx = logger.WithTraceID(ctx, traceID)
		ctx = context.WithValue(ctx, "span_id", spanID)
		ctx = context.WithValue(ctx, "parent_span_id", parentSpanID)

		// 将追踪信息添加到Chi上下文
		ctx = context.WithValue(ctx, middleware.RequestIDKey, requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetTraceInfo 从上下文获取追踪信息
func GetTraceInfo(ctx context.Context) TraceInfo {
	return TraceInfo{
		RequestID:    logger.GetRequestID(ctx),
		TraceID:      logger.GetTraceID(ctx),
		SpanID:       getStringFromContext(ctx, "span_id"),
		ParentSpanID: getStringFromContext(ctx, "parent_span_id"),
	}
}

// TraceInfo 追踪信息
type TraceInfo struct {
	RequestID    string `json:"request_id"`
	TraceID      string `json:"trace_id"`
	SpanID       string `json:"span_id,omitempty"`
	ParentSpanID string `json:"parent_span_id,omitempty"`
}

// getStringFromContext 从上下文获取字符串值
func getStringFromContext(ctx context.Context, key string) string {
	if value := ctx.Value(key); value != nil {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// RequestContextMiddleware 请求上下文中间件
func RequestContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 创建请求上下文
		ctx := r.Context()
		
		// 添加请求方法和路径
		ctx = context.WithValue(ctx, "http_method", r.Method)
		ctx = context.WithValue(ctx, "http_path", r.URL.Path)
		ctx = context.WithValue(ctx, "http_query", r.URL.RawQuery)
		ctx = context.WithValue(ctx, "user_agent", r.UserAgent())
		ctx = context.WithValue(ctx, "remote_addr", r.RemoteAddr)
		
		// 添加常用请求头
		ctx = context.WithValue(ctx, "content_type", r.Header.Get("Content-Type"))
		ctx = context.WithValue(ctx, "accept", r.Header.Get("Accept"))
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetHTTPRequestContext 获取HTTP请求上下文信息
func GetHTTPRequestContext(ctx context.Context) HTTPRequestContext {
	return HTTPRequestContext{
		Method:      getStringFromContext(ctx, "http_method"),
		Path:        getStringFromContext(ctx, "http_path"),
		Query:       getStringFromContext(ctx, "http_query"),
		UserAgent:   getStringFromContext(ctx, "user_agent"),
		RemoteAddr:  getStringFromContext(ctx, "remote_addr"),
		ContentType: getStringFromContext(ctx, "content_type"),
		Accept:      getStringFromContext(ctx, "accept"),
	}
}

// HTTPRequestContext HTTP请求上下文信息
type HTTPRequestContext struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Query       string `json:"query,omitempty"`
	UserAgent   string `json:"user_agent"`
	RemoteAddr  string `json:"remote_addr"`
	ContentType string `json:"content_type,omitempty"`
	Accept      string `json:"accept,omitempty"`
}
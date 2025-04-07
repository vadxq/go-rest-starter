package handlers

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/rs/zerolog"

	"github.com/vadxq/go-rest-starter/api/v1/dto"
	"github.com/vadxq/go-rest-starter/internal/app/middleware"
	"github.com/vadxq/go-rest-starter/internal/pkg/errors"
)

// ResponseWriter 响应写入器，用于统一处理HTTP响应
type ResponseWriter struct {
	Logger zerolog.Logger
}

// NewResponseWriter 创建一个新的响应写入器
func NewResponseWriter(logger zerolog.Logger) *ResponseWriter {
	return &ResponseWriter{Logger: logger}
}

// Error 写入错误响应
func (rw *ResponseWriter) Error(w http.ResponseWriter, r *http.Request, err error) {
	// 获取错误信息
	var statusCode int = http.StatusInternalServerError
	var message string = "服务器内部错误"
	var errorCode string = "internal_server_error"
	
	// 获取堆栈信息
	_, file, line, _ := runtime.Caller(1)
	
	// 处理自定义错误
	if appErr, ok := err.(errors.AppError); ok {
		statusCode = appErr.Code()
		message = appErr.Error()
		errorCode = appErr.ErrorCode()
	}
	
	// 获取请求信息
	traceID := middleware.GetTraceID(r.Context())
	requestMethod := r.Method
	requestPath := r.URL.Path
	clientIP := r.RemoteAddr
	
	// 记录错误日志
	logEvent := rw.Logger.Error().
		Str("trace_id", traceID).
		Str("method", requestMethod).
		Str("path", requestPath).
		Str("client_ip", clientIP).
		Str("error_code", errorCode).
		Int("status", statusCode).
		Str("error_file", file).
		Int("error_line", line).
		Err(err)
	
	// 添加用户信息（如果存在）
	userID, ok := middleware.GetUserID(r.Context())
	if ok && userID != 0 {
		logEvent = logEvent.Uint("user_id", userID)
	}
	
	// 输出日志
	logEvent.Msg("请求处理错误")
	
	// 构建错误响应
	errorResponse := dto.ErrorResponse{
		Code:      statusCode,
		Message:   message,
		ErrorCode: errorCode,
		TraceID:   traceID,
		Timestamp: time.Now().Unix(),
	}
	
	// 发送响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}

// Success 写入成功响应
func (rw *ResponseWriter) Success(w http.ResponseWriter, r *http.Request, data interface{}, statusCode int) {
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	
	// 获取请求信息
	traceID := middleware.GetTraceID(r.Context())
	
	// 记录成功日志
	logEvent := rw.Logger.Info().
		Str("trace_id", traceID).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Int("status", statusCode)
	
	// 添加用户信息（如果存在）
	userID, ok := middleware.GetUserID(r.Context())
	if ok && userID != 0 {
		logEvent = logEvent.Uint("user_id", userID)
	}
	
	// 输出日志
	logEvent.Msg("请求处理成功")
	
	// 构建成功响应
	successResponse := dto.Response{
		Code:      statusCode,
		Data:      data,
		TraceID:   traceID,
		Timestamp: time.Now().Unix(),
	}
	
	// 发送响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(successResponse)
}

// NoContent 返回无内容响应
func (rw *ResponseWriter) NoContent(w http.ResponseWriter, r *http.Request) {
	// 记录日志
	rw.Logger.Info().
		Str("trace_id", middleware.GetTraceID(r.Context())).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Int("status", http.StatusNoContent).
		Msg("请求处理成功(无内容)")
	
	w.WriteHeader(http.StatusNoContent)
}

// 向后兼容的处理方法
func handleError(logger zerolog.Logger, w http.ResponseWriter, r *http.Request, err error) {
	writer := ResponseWriter{Logger: logger}
	writer.Error(w, r, err)
} 
package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// Logger 日志记录器接口
type Logger interface {
	Debug(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
	With(keysAndValues ...any) Logger
	WithContext(ctx context.Context) Logger
}

// StructuredLogger 结构化日志记录器
type StructuredLogger struct {
	logger *slog.Logger
	ctx    context.Context
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `yaml:"level" json:"level"`             // 日志级别: debug, info, warn, error
	File       string `yaml:"file" json:"file"`               // 日志文件路径，空则输出到控制台
	Console    bool   `yaml:"console" json:"console"`         // 是否输出到控制台
	MaxSize    int    `yaml:"max_size" json:"max_size"`       // 文件最大大小(MB)
	MaxBackups int    `yaml:"max_backups" json:"max_backups"` // 保留的备份文件数
	MaxAge     int    `yaml:"max_age" json:"max_age"`         // 保留的天数
	Compress   bool   `yaml:"compress" json:"compress"`       // 是否压缩
}

// ContextKey 上下文键类型
type ContextKey string

const (
	// TraceIDKey 链路追踪ID键
	TraceIDKey ContextKey = "trace_id"
	// RequestIDKey 请求ID键
	RequestIDKey ContextKey = "request_id"
	// UserIDKey 用户ID键
	UserIDKey ContextKey = "user_id"
)

// NewLogger 创建新的日志记录器
func NewLogger(config *LogConfig) (*StructuredLogger, error) {
	level := parseLevel(config.Level)

	var writers []io.Writer

	// 控制台输出
	if config.Console || config.File == "" {
		writers = append(writers, os.Stdout)
	}

	// 文件输出
	if config.File != "" {
		file, err := createLogFile(config.File)
		if err != nil {
			return nil, err
		}
		writers = append(writers, file)
	}

	// 如果没有配置任何输出，默认输出到控制台
	if len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}

	var writer io.Writer
	if len(writers) == 1 {
		writer = writers[0]
	} else {
		writer = io.MultiWriter(writers...)
	}

	// 创建handler
	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// 格式化时间
			if a.Key == slog.TimeKey {
				return slog.String("timestamp", a.Value.Time().Format(time.RFC3339))
			}

			// 格式化源码位置
			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				// 只显示文件名和行号
				return slog.String("source", filepath.Base(source.File)+":"+
					func() string {
						return slog.IntValue(source.Line).String()
					}())
			}

			return a
		},
	})

	logger := slog.New(handler)

	return &StructuredLogger{
		logger: logger,
		ctx:    context.Background(),
	}, nil
}

// Default 创建默认日志记录器
func Default() *StructuredLogger {
	logger, _ := NewLogger(&LogConfig{
		Level:   "info",
		Console: true,
	})
	return logger
}

// parseLevel 解析日志级别
func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// createLogFile 创建日志文件
func createLogFile(filename string) (*os.File, error) {
	// 添加日期到文件名
	dir := filepath.Dir(filename)
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	nameWithoutExt := base[:len(base)-len(ext)]
	
	// 创建带日期的文件名
	dateStr := time.Now().Format("2006-01-02")
	newFilename := filepath.Join(dir, nameWithoutExt+"-"+dateStr+ext)
	
	// 确保目录存在
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	return os.OpenFile(newFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}

// Debug 输出调试级别日志
func (l *StructuredLogger) Debug(msg string, keysAndValues ...any) {
	l.log(slog.LevelDebug, msg, keysAndValues...)
}

// Info 输出信息级别日志
func (l *StructuredLogger) Info(msg string, keysAndValues ...any) {
	l.log(slog.LevelInfo, msg, keysAndValues...)
}

// Warn 输出警告级别日志
func (l *StructuredLogger) Warn(msg string, keysAndValues ...any) {
	l.log(slog.LevelWarn, msg, keysAndValues...)
}

// Error 输出错误级别日志
func (l *StructuredLogger) Error(msg string, keysAndValues ...any) {
	l.log(slog.LevelError, msg, keysAndValues...)
}

// log 内部日志记录方法
func (l *StructuredLogger) log(level slog.Level, msg string, keysAndValues ...any) {
	// 添加调用者信息
	pc, file, line, ok := runtime.Caller(2)
	if ok {
		funcName := runtime.FuncForPC(pc).Name()
		keysAndValues = append(keysAndValues,
			"caller", filepath.Base(file)+":"+slog.IntValue(line).String(),
			"function", filepath.Base(funcName),
		)
	}

	// 从上下文中提取信息
	if l.ctx != nil {
		if traceID := l.ctx.Value(TraceIDKey); traceID != nil {
			keysAndValues = append(keysAndValues, "trace_id", traceID)
		}
		if requestID := l.ctx.Value(RequestIDKey); requestID != nil {
			keysAndValues = append(keysAndValues, "request_id", requestID)
		}
		if userID := l.ctx.Value(UserIDKey); userID != nil {
			keysAndValues = append(keysAndValues, "user_id", userID)
		}
	}

	l.logger.Log(l.ctx, level, msg, keysAndValues...)
}

// With 添加字段到日志记录器
func (l *StructuredLogger) With(keysAndValues ...any) Logger {
	return &StructuredLogger{
		logger: l.logger.With(keysAndValues...),
		ctx:    l.ctx,
	}
}

// WithContext 为日志记录器设置上下文
func (l *StructuredLogger) WithContext(ctx context.Context) Logger {
	return &StructuredLogger{
		logger: l.logger,
		ctx:    ctx,
	}
}

// GetTraceID 从上下文中获取链路追踪ID
func GetTraceID(ctx context.Context) string {
	if traceID := ctx.Value(TraceIDKey); traceID != nil {
		if str, ok := traceID.(string); ok {
			return str
		}
	}
	// 尝试从Chi中间件获取
	if reqID := middleware.GetReqID(ctx); reqID != "" {
		return reqID
	}
	return ""
}

// GetRequestID 从上下文中获取请求ID
func GetRequestID(ctx context.Context) string {
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		if str, ok := requestID.(string); ok {
			return str
		}
	}
	return ""
}

// GetUserID 从上下文中获取用户ID
func GetUserID(ctx context.Context) string {
	if userID := ctx.Value(UserIDKey); userID != nil {
		if str, ok := userID.(string); ok {
			return str
		}
	}
	return ""
}

// WithTraceID 在上下文中设置链路追踪ID
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// WithRequestID 在上下文中设置请求ID
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// WithUserID 在上下文中设置用户ID
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware(logger Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 获取或生成请求ID
			requestID := middleware.GetReqID(r.Context())
			if requestID == "" {
				requestID = fmt.Sprintf("%d", middleware.NextRequestID())
			}

			// 获取或生成链路追踪ID
			traceID := r.Header.Get("X-Trace-ID")
			if traceID == "" {
				traceID = requestID
			}

			// 设置响应头
			w.Header().Set("X-Request-ID", requestID)
			w.Header().Set("X-Trace-ID", traceID)

			// 创建带有追踪信息的上下文
			ctx := WithRequestID(r.Context(), requestID)
			ctx = WithTraceID(ctx, traceID)

			// 创建带有上下文的日志记录器
			ctxLogger := logger.WithContext(ctx)

			// 记录请求开始
			ctxLogger.Info("request started",
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"user_agent", r.UserAgent(),
				"remote_addr", r.RemoteAddr,
			)

			// 创建响应记录器
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// 处理请求
			next.ServeHTTP(ww, r.WithContext(ctx))

			// 记录请求完成
			duration := time.Since(start)
			ctxLogger.Info("request completed",
				"status", ww.Status(),
				"duration_ms", duration.Milliseconds(),
				"bytes", ww.BytesWritten(),
			)

			// 记录慢请求
			if duration > 5*time.Second {
				ctxLogger.Warn("slow request detected",
					"duration_ms", duration.Milliseconds(),
					"threshold_ms", 5000,
				)
			}
		})
	}
}

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware(logger Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// 获取堆栈信息
					buf := make([]byte, 4096)
					n := runtime.Stack(buf, false)
					stackTrace := string(buf[:n])

					// 记录panic
					logger.WithContext(r.Context()).Error("panic recovered",
						"error", err,
						"stack_trace", stackTrace,
						"method", r.Method,
						"path", r.URL.Path,
					)

					// 返回500错误
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
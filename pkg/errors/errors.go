package errors

import (
	"fmt"
	"log/slog"
	"net/http"
)

// ErrorType 是错误类型的枚举
type ErrorType string

// 预定义错误类型
const (
	// ErrorTypeValidation 验证错误
	ErrorTypeValidation ErrorType = "VALIDATION_ERROR"
	// ErrorTypeNotFound 资源不存在
	ErrorTypeNotFound ErrorType = "NOT_FOUND"
	// ErrorTypeUnauthorized 未授权
	ErrorTypeUnauthorized ErrorType = "UNAUTHORIZED"
	// ErrorTypeForbidden 禁止访问
	ErrorTypeForbidden ErrorType = "FORBIDDEN"
	// ErrorTypeInternal 内部服务器错误
	ErrorTypeInternal ErrorType = "INTERNAL_ERROR"
	// ErrorTypeBadRequest 错误的请求
	ErrorTypeBadRequest ErrorType = "BAD_REQUEST"
	// ErrorTypeConflict 资源冲突
	ErrorTypeConflict ErrorType = "CONFLICT"
)

// Error 结构化错误
type Error struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Err     error     `json:"-"`
}

// Error 实现标准error接口
func (e *Error) Error() string {
	// 生产环境不输出内部错误详情
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap 实现errors.Unwrap接口
func (e *Error) Unwrap() error {
	return e.Err
}

// StatusCode 返回对应的HTTP状态码
func (e *Error) StatusCode() int {
	switch e.Type {
	case ErrorTypeValidation:
		return http.StatusBadRequest
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeUnauthorized:
		return http.StatusUnauthorized
	case ErrorTypeForbidden:
		return http.StatusForbidden
	case ErrorTypeInternal:
		return http.StatusInternalServerError
	case ErrorTypeBadRequest:
		return http.StatusBadRequest
	case ErrorTypeConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// New 创建新的错误
func New(errType ErrorType, message string, err error) *Error {
	return &Error{
		Type:    errType,
		Message: message,
		Err:     err,
	}
}

// ValidationError 创建验证错误
func ValidationError(message string, err error) *Error {
	return New(ErrorTypeValidation, message, err)
}

// NotFoundError 创建未找到错误
func NotFoundError(entity string, err error) *Error {
	return New(ErrorTypeNotFound, fmt.Sprintf("%s not found", entity), err)
}

// UnauthorizedError 创建未授权错误
func UnauthorizedError(message string, err error) *Error {
	return New(ErrorTypeUnauthorized, message, err)
}

// ForbiddenError 创建禁止访问错误
func ForbiddenError(message string, err error) *Error {
	return New(ErrorTypeForbidden, message, err)
}

// InternalError 创建内部服务器错误
func InternalError(message string, err error) *Error {
	return New(ErrorTypeInternal, message, err)
}

// BadRequestError 创建错误请求错误
func BadRequestError(message string, err error) *Error {
	return New(ErrorTypeBadRequest, message, err)
}

// ConflictError 创建资源冲突错误
func ConflictError(message string, err error) *Error {
	return New(ErrorTypeConflict, message, err)
}

// AsError 尝试将标准error转换为自定义Error类型
func AsError(err error) *Error {
	if err == nil {
		return nil
	}

	// 如果err已经是*Error类型，则直接返回
	if e, ok := err.(*Error); ok {
		return e
	}

	// 否则包装为内部错误
	return InternalError("unexpected error", err)
}

// RecoverPanic 用于从panic中恢复并记录错误
func RecoverPanic(source string) {
	if r := recover(); r != nil {
		// 生产环境只记录必要信息
		slog.Error("panic recovered", 
			"source", source,
			"error", fmt.Sprintf("%v", r))
	}
}

// RecoverPanicWithCallback 从panic中恢复，并执行回调函数
func RecoverPanicWithCallback(source string, callback func(err interface{})) {
	if r := recover(); r != nil {
		slog.Error("panic recovered", 
			"source", source,
			"error", fmt.Sprintf("%v", r))
		
		if callback != nil {
			callback(r)
		}
	}
}

package errors

import (
	"fmt"
)

// AppError 应用错误接口
type AppError interface {
	error
	Code() int         // HTTP状态码
	ErrorCode() string // 业务错误代码
	WithContext(key string, value interface{}) AppError // 添加上下文信息
	Context() map[string]interface{} // 获取上下文信息
}

// 基础错误类型
type baseError struct {
	message   string                 // 错误消息
	code      int                    // HTTP状态码
	errorCode string                 // 业务错误代码
	context   map[string]interface{} // 上下文信息
}

// Error 实现error接口
func (e *baseError) Error() string {
	return e.message
}

// Code 返回HTTP状态码
func (e *baseError) Code() int {
	return e.code
}

// ErrorCode 返回业务错误代码
func (e *baseError) ErrorCode() string {
	return e.errorCode
}

// WithContext 添加上下文信息
func (e *baseError) WithContext(key string, value interface{}) AppError {
	if e.context == nil {
		e.context = make(map[string]interface{})
	}
	e.context[key] = value
	return e
}

// Context 获取上下文信息
func (e *baseError) Context() map[string]interface{} {
	return e.context
}

// NotFoundError 资源未找到错误
type NotFoundError struct {
	baseError
}

// ValidationError 验证错误
type ValidationError struct {
	baseError
}

// UnauthorizedError 未授权错误
type UnauthorizedError struct {
	baseError
}

// ForbiddenError 禁止访问错误
type ForbiddenError struct {
	baseError
}

// InternalError 内部错误
type InternalError struct {
	baseError
	err error // 原始错误
}

// Error 重写Error方法以包含原始错误
func (e *InternalError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s: %v", e.message, e.err)
	}
	return e.message
}

// Unwrap 返回原始错误
func (e *InternalError) Unwrap() error {
	return e.err
}

// NewNotFoundError 创建资源未找到错误
func NewNotFoundError(message string) *NotFoundError {
	return &NotFoundError{
		baseError: baseError{
			message:   message,
			code:      404,
			errorCode: "resource_not_found",
			context:   make(map[string]interface{}),
		},
	}
}

// NewValidationError 创建验证错误
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		baseError: baseError{
			message:   message,
			code:      400,
			errorCode: "validation_failed",
			context:   make(map[string]interface{}),
		},
	}
}

// NewUnauthorizedError 创建未授权错误
func NewUnauthorizedError(message string) *UnauthorizedError {
	return &UnauthorizedError{
		baseError: baseError{
			message:   message,
			code:      401,
			errorCode: "unauthorized",
			context:   make(map[string]interface{}),
		},
	}
}

// NewForbiddenError 创建禁止访问错误
func NewForbiddenError(message string) *ForbiddenError {
	return &ForbiddenError{
		baseError: baseError{
			message:   message,
			code:      403,
			errorCode: "forbidden",
			context:   make(map[string]interface{}),
		},
	}
}

// NewInternalError 创建内部错误
func NewInternalError(err error) *InternalError {
	message := "内部服务器错误"
	if err != nil {
		message = fmt.Sprintf("内部服务器错误: %v", err)
	}
	
	return &InternalError{
		baseError: baseError{
			message:   message,
			code:      500,
			errorCode: "internal_server_error",
			context:   make(map[string]interface{}),
		},
		err: err,
	}
}

// BadRequestError 创建请求参数错误
func NewBadRequestError(message string) AppError {
	return &baseError{
		message:   message,
		code:      400,
		errorCode: "bad_request",
		context:   make(map[string]interface{}),
	}
}

// ConflictError 创建资源冲突错误
func NewConflictError(message string) AppError {
	return &baseError{
		message:   message,
		code:      409,
		errorCode: "resource_conflict",
		context:   make(map[string]interface{}),
	}
}

// ServiceUnavailableError 创建服务不可用错误
func NewServiceUnavailableError(message string) AppError {
	return &baseError{
		message:   message,
		code:      503,
		errorCode: "service_unavailable",
		context:   make(map[string]interface{}),
	}
}

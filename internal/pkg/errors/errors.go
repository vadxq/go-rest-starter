package errors

import (
	"fmt"
)

// AppError 应用错误接口
type AppError interface {
	error
	Code() int
}

// NotFoundError 资源未找到错误
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID %s not found", e.Resource, e.ID)
}

func (e *NotFoundError) Code() int {
	return 404
}

// ValidationError 验证错误
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func (e *ValidationError) Code() int {
	return 400
}

// UnauthorizedError 未授权错误
type UnauthorizedError struct {
	Message string
}

func (e *UnauthorizedError) Error() string {
	return e.Message
}

func (e *UnauthorizedError) Code() int {
	return 401
}

// ForbiddenError 禁止访问错误
type ForbiddenError struct {
	Message string
}

func (e *ForbiddenError) Error() string {
	return e.Message
}

func (e *ForbiddenError) Code() int {
	return 403
}

// InternalError 内部错误
type InternalError struct {
	Err error
}

func (e *InternalError) Error() string {
	return fmt.Sprintf("内部服务器错误: %v", e.Err)
}

func (e *InternalError) Code() int {
	return 500
}

// NewNotFoundError 创建资源未找到错误
func NewNotFoundError(resource, id string) *NotFoundError {
	return &NotFoundError{
		Resource: resource,
		ID:       id,
	}
}

// NewValidationError 创建验证错误
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		Message: message,
	}
}

// NewUnauthorizedError 创建未授权错误
func NewUnauthorizedError(message string) *UnauthorizedError {
	return &UnauthorizedError{
		Message: message,
	}
}

// NewForbiddenError 创建禁止访问错误
func NewForbiddenError(message string) *ForbiddenError {
	return &ForbiddenError{
		Message: message,
	}
}

// NewInternalError 创建内部错误
func NewInternalError(err error) *InternalError {
	return &InternalError{
		Err: err,
	}
}

package dto

import "time"

// CreateUserInput 创建用户请求
type CreateUserInput struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// UpdateUserInput 更新用户请求
type UpdateUserInput struct {
	Name     string `json:"name" validate:"omitempty,min=2,max=100"`
	Email    string `json:"email" validate:"omitempty,email"`
	Password string `json:"password" validate:"omitempty,min=6"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"trace_id,omitempty"`
}

// ListResponse 列表分页响应
type ListResponse struct {
	Data  interface{} `json:"data"`            // 列表数据
	Page  int         `json:"page"`            // 当前页码
	Size  int         `json:"size"`            // 每页大小
	Total int64       `json:"total"`           // 总记录数
}

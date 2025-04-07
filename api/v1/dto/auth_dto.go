package dto

// 基础响应结构
type Response struct {
	Code      int         `json:"code"`      // HTTP状态码
	Data      interface{} `json:"data"`      // 响应数据
	TraceID   string      `json:"trace_id"`  // 请求跟踪ID
	Timestamp int64       `json:"timestamp"` // 响应时间戳
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code      int    `json:"code"`       // HTTP状态码
	Message   string `json:"message"`    // 错误消息
	ErrorCode string `json:"error_code"` // 错误代码
	TraceID   string `json:"trace_id"`   // 请求跟踪ID
	Timestamp int64  `json:"timestamp"`  // 响应时间戳
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	TokenType    string       `json:"token_type"`
	User         UserResponse `json:"user"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// TokenResponse 令牌响应
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
} 
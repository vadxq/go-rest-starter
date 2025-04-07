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
	ErrorCode string `json:"error_code"` // 错误代码，可选
	TraceID   string `json:"trace_id,omitempty"`   // 请求跟踪ID
	Timestamp int64  `json:"timestamp,omitempty"`  // 响应时间戳
}

// ListResponse 列表分页响应
type ListResponse struct {
	Data  interface{} `json:"data"`   // 列表数据
	Page  int         `json:"page"`   // 当前页码
	Size  int         `json:"size"`   // 每页大小
	Total int64       `json:"total"`  // 总记录数
} 
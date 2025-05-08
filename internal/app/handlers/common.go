package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	apperrors "github.com/vadxq/go-rest-starter/pkg/errors"
)

// Response 标准API响应结构
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo 错误信息结构
type ErrorInfo struct {
	Type    string   `json:"type"`
	Message string   `json:"message"`
	Fields  []string `json:"fields,omitempty"`
}

// RespondJSON 发送JSON响应
func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	response := Response{
		Success: status >= 200 && status < 300,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("响应JSON序列化失败", "error", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}

// RespondError 发送错误响应
func RespondError(w http.ResponseWriter, err error) {
	var appErr *apperrors.Error

	// 尝试将err转换为应用错误类型
	if !errors.As(err, &appErr) {
		// 如果不是应用错误，则将其包装为内部错误
		appErr = apperrors.InternalError("内部服务器错误", err)
	}

	// 构建错误响应
	response := Response{
		Success: false,
		Error: &ErrorInfo{
			Type:    string(appErr.Type),
			Message: appErr.Message,
			Fields:  appErr.Fields,
		},
	}

	// 获取HTTP状态码
	status := appErr.StatusCode()

	// 记录错误
	if status >= 500 {
		slog.Error(appErr.Message, "error", appErr, "type", string(appErr.Type))
	} else {
		slog.Debug(appErr.Message, "error", appErr, "type", string(appErr.Type))
	}

	// 发送响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("错误响应JSON序列化失败", "error", err)
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}

// DecodeJSON 从请求体解析JSON数据
func DecodeJSON(r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return apperrors.BadRequestError("无效的JSON数据", err)
	}
	return nil
}

// BindJSON 从请求体解析JSON并验证
func BindJSON(r *http.Request, v interface{}, validate func(interface{}) error) error {
	// 解析JSON
	if err := DecodeJSON(r, v); err != nil {
		return err
	}

	// 验证数据
	if validate != nil {
		if err := validate(v); err != nil {
			return apperrors.ValidationError("数据验证失败", err)
		}
	}

	return nil
}

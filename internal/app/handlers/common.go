package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/vadxq/go-rest-starter/api/v1/dto"
	"github.com/vadxq/go-rest-starter/internal/app/middleware"
	apperrors "github.com/vadxq/go-rest-starter/internal/pkg/errors"
)

// handleError 统一处理服务错误
func handleError(logger zerolog.Logger, w http.ResponseWriter, r *http.Request, err error) {
	statusCode := http.StatusInternalServerError
	message := "服务器错误"

	var appErr apperrors.AppError
	if errors.As(err, &appErr) {
		statusCode = appErr.Code()
		message = appErr.Error()
	}

	// 记录错误日志
	logger.Error().
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Int("status_code", statusCode).
		Str("trace_id", middleware.GetTraceID(r.Context())).
		Err(err).
		Msg("请求处理失败")

	// 构建响应
	errorResponse := dto.ErrorResponse{
		Code:    statusCode,
		Message: message,
		TraceID: middleware.GetTraceID(r.Context()),
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
} 
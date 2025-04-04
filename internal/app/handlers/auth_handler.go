package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"

	"github.com/vadxq/go-rest-starter/api/v1/dto"
	"github.com/vadxq/go-rest-starter/internal/app/services"
)

// AuthHandler 处理认证相关的HTTP请求
type AuthHandler struct {
	authService services.AuthService
	logger      zerolog.Logger
	validator   *validator.Validate
}

// NewAuthHandler 创建一个新的AuthHandler实例
func NewAuthHandler(as services.AuthService, logger zerolog.Logger, v *validator.Validate) *AuthHandler {
	return &AuthHandler{
		authService: as,
		logger:      logger,
		validator:   v,
	}
}

// Login 处理用户登录请求
// @Summary 用户登录
// @Description 通过邮箱和密码进行登录，并获取访问令牌
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.LoginRequest true "登录请求体"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, r, "无效的请求体", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.renderError(w, r, "请求参数验证失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	response, err := h.authService.Login(r.Context(), req)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	h.renderJSON(w, r, response, http.StatusOK)
}

// RefreshToken 处理令牌刷新请求
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.RefreshTokenRequest true "刷新令牌请求体"
// @Success 200 {object} dto.TokenResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.renderError(w, r, "无效的请求体", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.renderError(w, r, "请求参数验证失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	response, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	h.renderJSON(w, r, response, http.StatusOK)
}

// 渲染错误响应
func (h *AuthHandler) renderError(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
	h.logger.Error().
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Int("status_code", statusCode).
		Msg(message)

	response := dto.ErrorResponse{
		Code:    statusCode,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// 渲染JSON响应
func (h *AuthHandler) renderJSON(w http.ResponseWriter, r *http.Request, data interface{}, statusCode int) {
	h.logger.Info().
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Int("status_code", statusCode).
		Msg("响应成功")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// 处理服务错误
func (h *AuthHandler) handleServiceError(w http.ResponseWriter, r *http.Request, err error) {
	handleError(h.logger, w, r, err)
} 
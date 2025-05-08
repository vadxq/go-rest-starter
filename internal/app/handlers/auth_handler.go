package handlers

import (
	"net/http"
	"strings"

	"log/slog"

	"github.com/go-playground/validator/v10"

	"github.com/vadxq/go-rest-starter/internal/app/dto"
	"github.com/vadxq/go-rest-starter/internal/app/services"
	apperrors "github.com/vadxq/go-rest-starter/pkg/errors"
)

// AuthHandler 处理认证相关的HTTP请求
type AuthHandler struct {
	authService services.AuthService
	logger      *slog.Logger
	validator   *validator.Validate
}

// NewAuthHandler 创建一个新的AuthHandler实例
func NewAuthHandler(as services.AuthService, logger *slog.Logger, v *validator.Validate) *AuthHandler {
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
// @Success 200 {object} Response{data=dto.LoginResponse}
// @Failure 400,401,500 {object} Response{error=ErrorInfo}
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest

	if err := BindJSON(r, &req, func(v interface{}) error {
		return h.validator.Struct(v)
	}); err != nil {
		RespondError(w, err)
		return
	}

	response, err := h.authService.Login(r.Context(), req)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, response)
}

// RefreshToken 处理令牌刷新请求
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.RefreshTokenRequest true "刷新令牌请求体"
// @Success 200 {object} Response{data=dto.TokenResponse}
// @Failure 400,401,500 {object} Response{error=ErrorInfo}
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshTokenRequest

	if err := BindJSON(r, &req, func(v interface{}) error {
		return h.validator.Struct(v)
	}); err != nil {
		RespondError(w, err)
		return
	}

	response, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusOK, response)
}

// Logout 处理用户登出请求
// @Summary 用户登出
// @Description 使当前用户的访问令牌失效
// @Tags auth
// @Accept json
// @Produce json
// @Success 204 {object} nil
// @Failure 401,500 {object} Response{error=ErrorInfo}
// @Router /api/v1/auth/logout [post]
// @Security BearerAuth
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// 从Authorization头部获取访问令牌
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		RespondError(w, apperrors.UnauthorizedError("未提供授权令牌", nil))
		return
	}

	// 分离Bearer前缀和令牌
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		RespondError(w, apperrors.UnauthorizedError("授权格式无效", nil))
		return
	}

	accessToken := parts[1]

	// 调用服务执行登出
	err := h.authService.Logout(r.Context(), accessToken)
	if err != nil {
		RespondError(w, err)
		return
	}

	// 成功登出返回204状态码
	RespondJSON(w, http.StatusNoContent, nil)
}

package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"

	"github.com/vadxq/go-rest-starter/api/v1/dto"
	"github.com/vadxq/go-rest-starter/internal/app/services"
	apperrors "github.com/vadxq/go-rest-starter/internal/pkg/errors"
)

// UserHandler 处理用户相关的 HTTP 请求
type UserHandler struct {
	userService services.UserService
	logger      zerolog.Logger
	validator   *validator.Validate
}

// NewUserHandler 创建一个新的 UserHandler 实例
func NewUserHandler(us services.UserService, logger zerolog.Logger, v *validator.Validate) *UserHandler {
	return &UserHandler{
		userService: us,
		logger:      logger,
		validator:   v,
	}
}

// GetUser 获取用户详情
// @Summary 获取用户详情
// @Description 根据用户ID获取用户详细信息
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		h.renderError(w, r, "ID参数缺失", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	// 转换为 DTO
	response := dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	h.renderJSON(w, r, response, http.StatusOK)
}

// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建一个新的用户
// @Tags users
// @Accept json
// @Produce json
// @Param body body dto.CreateUserInput true "创建用户请求体"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input dto.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.renderError(w, r, "无效的请求体", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(input); err != nil {
		h.renderError(w, r, "请求参数验证失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.userService.CreateUser(r.Context(), input)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	// 转换为 DTO
	response := dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	h.renderJSON(w, r, response, http.StatusCreated)
}

// UpdateUser 更新用户
// @Summary 更新用户
// @Description 根据用户ID更新用户信息
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Param body body dto.UpdateUserInput true "更新用户请求体"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		h.renderError(w, r, "ID参数缺失", http.StatusBadRequest)
		return
	}

	var input dto.UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.renderError(w, r, "无效的请求体", http.StatusBadRequest)
		return
	}

	user, err := h.userService.UpdateUser(r.Context(), userID, input)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	// 转换为 DTO
	response := dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	h.renderJSON(w, r, response, http.StatusOK)
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 根据用户ID删除用户
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Success 204 {object} dto.ErrorResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		h.renderError(w, r, "ID参数缺失", http.StatusBadRequest)
		return
	}

	err := h.userService.DeleteUser(r.Context(), userID)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}

	h.renderJSON(w, r, nil, http.StatusNoContent)
}

// renderError 渲染错误响应
func (h *UserHandler) renderError(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
	h.logger.Error().
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Int("status_code", statusCode).
		Msg(message)

	response := dto.ErrorResponse{
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// renderJSON 渲染 JSON 响应
func (h *UserHandler) renderJSON(w http.ResponseWriter, r *http.Request, data interface{}, statusCode int) {
	h.logger.Info().
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Int("status_code", statusCode).
		Msg("响应成功")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// handleServiceError 处理服务错误
func (h *UserHandler) handleServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var appErr apperrors.AppError
	if errors.As(err, &appErr) {
		h.renderError(w, r, appErr.Error(), appErr.Code())
	} else {
		h.renderError(w, r, "服务器错误", http.StatusInternalServerError)
	}
}

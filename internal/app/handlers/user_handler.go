package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"

	"github.com/vadxq/go-rest-starter/api/v1/dto"
	"github.com/vadxq/go-rest-starter/internal/app/services"
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
// @Router /api/v1/users/{id} [get]
// @Security BearerAuth
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
// @Router /api/v1/users [post]
// @Security BearerAuth
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
// @Router /api/v1/users/{id} [put]
// @Security BearerAuth
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
// @Success 204 {object} nil
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/users/{id} [delete]
// @Security BearerAuth
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

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取用户列表
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "页码，默认为1" default(1)
// @Param page_size query int false "每页大小，默认为10" default(10)
// @Success 200 {object} dto.ListResponse{data=[]dto.UserResponse}
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/users [get]
// @Security BearerAuth
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// 获取分页参数
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	
	pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}
	
	// 限制每页最大数量
	if pageSize > 100 {
		pageSize = 100
	}
	
	// 调用Service层获取用户列表
	users, total, err := h.userService.ListUsers(r.Context(), page, pageSize)
	if err != nil {
		h.handleServiceError(w, r, err)
		return
	}
	
	// 转换为DTO
	usersResponse := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		usersResponse = append(usersResponse, dto.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}
	
	// 构建分页响应
	response := dto.ListResponse{
		Data:  usersResponse,
		Page:  page,
		Size:  pageSize,
		Total: total,
	}
	
	h.renderJSON(w, r, response, http.StatusOK)
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
	handleError(h.logger, w, r, err)
}

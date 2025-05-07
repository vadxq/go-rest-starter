package handlers

import (
	"net/http"
	"strconv"

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
// @Success 200 {object} Response{data=dto.UserResponse}
// @Failure 400,404,500 {object} Response{error=ErrorInfo}
// @Router /api/v1/users/{id} [get]
// @Security BearerAuth
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		RespondError(w, apperrors.BadRequestError("ID参数缺失", nil))
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		RespondError(w, err)
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

	RespondJSON(w, http.StatusOK, response)
}

// CreateUser 创建用户
// @Summary 创建用户
// @Description 创建一个新的用户
// @Tags users
// @Accept json
// @Produce json
// @Param body body dto.CreateUserInput true "创建用户请求体"
// @Success 201 {object} Response{data=dto.UserResponse}
// @Failure 400,500 {object} Response{error=ErrorInfo}
// @Router /api/v1/users [post]
// @Security BearerAuth
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input dto.CreateUserInput

	if err := BindJSON(r, &input, func(v interface{}) error {
		return h.validator.Struct(v)
	}); err != nil {
		RespondError(w, err)
		return
	}

	user, err := h.userService.CreateUser(r.Context(), input)
	if err != nil {
		RespondError(w, err)
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

	RespondJSON(w, http.StatusCreated, response)
}

// UpdateUser 更新用户
// @Summary 更新用户
// @Description 根据用户ID更新用户信息
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Param body body dto.UpdateUserInput true "更新用户请求体"
// @Success 200 {object} Response{data=dto.UserResponse}
// @Failure 400,404,500 {object} Response{error=ErrorInfo}
// @Router /api/v1/users/{id} [put]
// @Security BearerAuth
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		RespondError(w, apperrors.BadRequestError("ID参数缺失", nil))
		return
	}

	var input dto.UpdateUserInput
	if err := BindJSON(r, &input, nil); err != nil {
		RespondError(w, err)
		return
	}

	user, err := h.userService.UpdateUser(r.Context(), userID, input)
	if err != nil {
		RespondError(w, err)
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

	RespondJSON(w, http.StatusOK, response)
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 根据用户ID删除用户
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Success 204 {object} nil
// @Failure 400,404,500 {object} Response{error=ErrorInfo}
// @Router /api/v1/users/{id} [delete]
// @Security BearerAuth
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		RespondError(w, apperrors.BadRequestError("ID参数缺失", nil))
		return
	}

	err := h.userService.DeleteUser(r.Context(), userID)
	if err != nil {
		RespondError(w, err)
		return
	}

	RespondJSON(w, http.StatusNoContent, nil)
}

// ListUsers 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取用户列表
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "页码，默认为1" default(1)
// @Param page_size query int false "每页大小，默认为10" default(10)
// @Success 200 {object} Response{data=dto.ListResponse{data=[]dto.UserResponse}}
// @Failure 500 {object} Response{error=ErrorInfo}
// @Router /api/v1/users [get]
// @Security BearerAuth
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// 解析分页参数
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := 1
	pageSize := 10

	if pageStr != "" {
		pageVal, err := strconv.Atoi(pageStr)
		if err == nil && pageVal > 0 {
			page = pageVal
		}
	}

	if pageSizeStr != "" {
		pageSizeVal, err := strconv.Atoi(pageSizeStr)
		if err == nil && pageSizeVal > 0 {
			pageSize = pageSizeVal
		}
	}

	users, total, err := h.userService.ListUsers(r.Context(), page, pageSize)
	if err != nil {
		RespondError(w, err)
		return
	}

	// 转换为 DTO
	userResponses := make([]dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = dto.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
	}

	response := dto.ListResponse{
		Data:  userResponses,
		Total: total,
		Page:  page,
		Size:  pageSize,
	}

	RespondJSON(w, http.StatusOK, response)
}

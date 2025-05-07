package services

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/vadxq/go-rest-starter/api/v1/dto"
	"github.com/vadxq/go-rest-starter/internal/app/models"
	"github.com/vadxq/go-rest-starter/internal/app/repository"
	"github.com/vadxq/go-rest-starter/internal/pkg/cache"
	apperrors "github.com/vadxq/go-rest-starter/internal/pkg/errors"
)

const (
	// 用户缓存键前缀
	userCachePrefix = "user:"

	// 用户列表缓存键
	userListCacheKey = "user:list"

	// 用户缓存过期时间
	userCacheTTL = 30 * time.Minute
)

// UserService 用户服务接口
type UserService interface {
	CreateUser(ctx context.Context, input dto.CreateUserInput) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
	UpdateUser(ctx context.Context, id string, input dto.UpdateUserInput) (*models.User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, page, pageSize int) ([]*models.User, int64, error)
}

// userService 用户服务实现
type userService struct {
	userRepo  repository.UserRepository
	validator *validator.Validate
	db        *gorm.DB
	cache     cache.Cache
}

// NewUserService 创建用户服务
func NewUserService(ur repository.UserRepository, v *validator.Validate, db *gorm.DB, c cache.Cache) UserService {
	return &userService{
		userRepo:  ur,
		validator: v,
		db:        db,
		cache:     c,
	}
}

// 获取用户缓存键
func getUserCacheKey(id string) string {
	return fmt.Sprintf("%s%s", userCachePrefix, id)
}

// CreateUser 创建用户
func (s *userService) CreateUser(ctx context.Context, input dto.CreateUserInput) (*models.User, error) {
	// 验证输入
	if err := s.validator.Struct(input); err != nil {
		return nil, apperrors.ValidationError("输入数据验证失败", err)
	}

	// 检查邮箱是否已存在
	exists, err := s.userRepo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, err // 错误已经在仓库层包装
	}

	if exists {
		return nil, apperrors.ConflictError("邮箱已被注册", nil)
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.InternalError("密码加密失败", err)
	}

	user := &models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPassword),
		Role:     "user", // 默认角色
	}

	// 开启事务
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.Create(ctx, tx, user); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err // 错误已经在仓库层包装
	}

	// 清除用户列表缓存
	_ = s.cache.Delete(ctx, userListCacheKey)

	return user, nil
}

// GetByID 根据ID获取用户
func (s *userService) GetByID(ctx context.Context, id string) (*models.User, error) {
	// 尝试从缓存获取
	cacheKey := getUserCacheKey(id)
	var user models.User

	err := s.cache.GetObject(ctx, cacheKey, &user)
	if err == nil {
		return &user, nil
	}

	// 缓存未命中，从数据库获取
	user2, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err // 错误已经在仓库层包装
	}

	// 存入缓存
	_ = s.cache.SetObject(ctx, cacheKey, user2, userCacheTTL)

	return user2, nil
}

// UpdateUser 更新用户
func (s *userService) UpdateUser(ctx context.Context, id string, input dto.UpdateUserInput) (*models.User, error) {
	// 验证输入
	if err := s.validator.Struct(input); err != nil {
		return nil, apperrors.ValidationError("输入数据验证失败", err)
	}

	// 获取用户
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err // 错误已经在仓库层包装
	}

	// 更新用户字段
	if input.Name != "" {
		user.Name = input.Name
	}

	if input.Email != "" && input.Email != user.Email {
		// 检查新邮箱是否存在
		exists, err := s.userRepo.ExistsByEmail(ctx, input.Email)
		if err != nil {
			return nil, err // 错误已经在仓库层包装
		}

		if exists {
			return nil, apperrors.ConflictError("邮箱已被注册", nil)
		}

		user.Email = input.Email
	}

	if input.Password != "" {
		// 加密密码
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, apperrors.InternalError("密码加密失败", err)
		}

		user.Password = string(hashedPassword)
	}

	// 开启事务
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.Update(ctx, tx, user); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err // 错误已经在仓库层包装
	}

	// 更新缓存
	cacheKey := getUserCacheKey(id)
	_ = s.cache.SetObject(ctx, cacheKey, user, userCacheTTL)

	// 清除用户列表缓存
	_ = s.cache.Delete(ctx, userListCacheKey)

	return user, nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(ctx context.Context, id string) error {
	// 获取用户
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err // 错误已经在仓库层包装
	}

	// 开启事务
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.Delete(ctx, tx, user.ID); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err // 错误已经在仓库层包装
	}

	// 删除缓存
	cacheKey := getUserCacheKey(id)
	_ = s.cache.Delete(ctx, cacheKey)

	// 清除用户列表缓存
	_ = s.cache.Delete(ctx, userListCacheKey)

	return nil
}

// ListUsers 获取用户列表
func (s *userService) ListUsers(ctx context.Context, page, pageSize int) ([]*models.User, int64, error) {
	// 生成缓存键，包含分页信息
	cacheKey := fmt.Sprintf("%s:%d:%d", userListCacheKey, page, pageSize)

	// 尝试从缓存获取
	var cachedResult struct {
		Users []*models.User `json:"users"`
		Total int64          `json:"total"`
	}

	err := s.cache.GetObject(ctx, cacheKey, &cachedResult)
	if err == nil {
		return cachedResult.Users, cachedResult.Total, nil
	}

	// 缓存未命中，从数据库获取
	users, total, err := s.userRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err // 错误已经在仓库层包装
	}

	// 存入缓存
	cachedResult = struct {
		Users []*models.User `json:"users"`
		Total int64          `json:"total"`
	}{
		Users: users,
		Total: total,
	}

	_ = s.cache.SetObject(ctx, cacheKey, cachedResult, userCacheTTL)

	return users, total, nil
}

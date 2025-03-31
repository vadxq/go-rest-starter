package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/vadxq/go-rest-starter/api/v1/dto"
	"github.com/vadxq/go-rest-starter/internal/app/models"
	"github.com/vadxq/go-rest-starter/internal/app/repository"
	apperrors "github.com/vadxq/go-rest-starter/internal/pkg/errors"
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
}

// NewUserService 创建用户服务
func NewUserService(ur repository.UserRepository, v *validator.Validate, db *gorm.DB) UserService {
	return &userService{
		userRepo:  ur,
		validator: v,
		db:        db,
	}
}

// CreateUser 创建用户
func (s *userService) CreateUser(ctx context.Context, input dto.CreateUserInput) (*models.User, error) {
	if err := s.validator.Struct(input); err != nil {
		return nil, apperrors.NewValidationError(err.Error())
	}

	// 检查邮箱是否已存在
	existingUser, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err == nil && existingUser != nil {
		return nil, apperrors.NewValidationError("邮箱已被使用")
	} else if err != nil && !errors.As(err, &apperrors.NotFoundError{}) {
		return nil, err
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.NewInternalError(fmt.Errorf("密码加密失败: %w", err))
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
		return nil, err
	}

	return user, nil
}

// GetByID 根据ID获取用户
func (s *userService) GetByID(ctx context.Context, id string) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// UpdateUser 更新用户
func (s *userService) UpdateUser(ctx context.Context, id string, input dto.UpdateUserInput) (*models.User, error) {
	if err := s.validator.Struct(input); err != nil {
		return nil, apperrors.NewValidationError(err.Error())
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if input.Name != "" {
		user.Name = input.Name
	}

	if input.Email != "" && input.Email != user.Email {
		// 检查新邮箱是否已存在
		existingUser, err := s.userRepo.GetByEmail(ctx, input.Email)
		if err == nil && existingUser != nil && existingUser.ID != user.ID {
			return nil, apperrors.NewValidationError("邮箱已被使用")
		} else if err != nil && !errors.As(err, &apperrors.NotFoundError{}) {
			return nil, err
		}
		user.Email = input.Email
	}

	if input.Password != "" {
		// 加密新密码
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, apperrors.NewInternalError(fmt.Errorf("密码加密失败: %w", err))
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
		return nil, err
	}

	return user, nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(ctx context.Context, id string) error {
	// 开启事务
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.userRepo.Delete(ctx, tx, id); err != nil {
			return err
		}
		return nil
	})

	return err
}

// ListUsers 获取用户列表
func (s *userService) ListUsers(ctx context.Context, page, pageSize int) ([]*models.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	users, err := s.userRepo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

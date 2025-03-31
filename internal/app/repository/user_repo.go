package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"gorm.io/gorm"

	"github.com/vadxq/go-rest-starter/internal/app/models"
	apperrors "github.com/vadxq/go-rest-starter/internal/pkg/errors"
)

// UserRepository 用户仓库接口
type UserRepository interface {
	Create(ctx context.Context, tx *gorm.DB, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, tx *gorm.DB, user *models.User) error
	Delete(ctx context.Context, tx *gorm.DB, id string) error
	List(ctx context.Context, offset, limit int) ([]*models.User, error)
	Count(ctx context.Context) (int64, error)
}

// userRepository 用户仓库实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create 创建用户
func (r *userRepository) Create(ctx context.Context, tx *gorm.DB, user *models.User) error {
	result := tx.WithContext(ctx).Create(user)
	if result.Error != nil {
		return fmt.Errorf("创建用户失败: %w", result.Error)
	}
	return nil
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("User", id)
		}
		return nil, fmt.Errorf("获取用户失败: %w", result.Error)
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("User", email)
		}
		return nil, fmt.Errorf("获取用户失败: %w", result.Error)
	}
	return &user, nil
}

// Update 更新用户
func (r *userRepository) Update(ctx context.Context, tx *gorm.DB, user *models.User) error {
	result := tx.WithContext(ctx).Save(user)
	if result.Error != nil {
		return fmt.Errorf("更新用户失败: %w", result.Error)
	}
	return nil
}

// Delete 删除用户
func (r *userRepository) Delete(ctx context.Context, tx *gorm.DB, id string) error {
	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return apperrors.NewValidationError("无效的用户ID")
	}

	result := tx.WithContext(ctx).Delete(&models.User{}, userID)
	if result.Error != nil {
		return fmt.Errorf("删除用户失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.NewNotFoundError("User", id)
	}
	return nil
}

// List 获取用户列表
func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*models.User, error) {
	var users []*models.User
	result := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("获取用户列表失败: %w", result.Error)
	}
	return users, nil
}

// Count 获取用户总数
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&models.User{}).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("获取用户总数失败: %w", result.Error)
	}
	return count, nil
}

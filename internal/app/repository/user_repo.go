package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/vadxq/go-rest-starter/internal/app/models"
	apperrors "github.com/vadxq/go-rest-starter/internal/pkg/errors"
)

// UserRepository 定义了用户仓库接口
type UserRepository interface {
	Create(ctx context.Context, tx *gorm.DB, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	Update(ctx context.Context, tx *gorm.DB, user *models.User) error
	Delete(ctx context.Context, tx *gorm.DB, id uint) error
	List(ctx context.Context, page, pageSize int) ([]*models.User, int64, error)
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建一个新的 UserRepository 实例
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

// GetByID 根据 ID 获取用户
func (r *userRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).First(&user, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, apperrors.NewNotFoundError("用户不存在")
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
		if result.Error == gorm.ErrRecordNotFound {
			return nil, apperrors.NewNotFoundError("用户不存在")
		}
		return nil, fmt.Errorf("获取用户失败: %w", result.Error)
	}
	return &user, nil
}

// ExistsByEmail 检查邮箱是否存在
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&models.User{}).Where("email = ?", email).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("检查邮箱是否存在失败: %w", result.Error)
	}
	return count > 0, nil
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
func (r *userRepository) Delete(ctx context.Context, tx *gorm.DB, id uint) error {
	result := tx.WithContext(ctx).Delete(&models.User{}, id)
	if result.Error != nil {
		return fmt.Errorf("删除用户失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.NewNotFoundError(fmt.Sprintf("用户(ID:%d)不存在", id))
	}
	return nil
}

// List 获取用户列表
func (r *userRepository) List(ctx context.Context, page, pageSize int) ([]*models.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	
	offset := (page - 1) * pageSize
	
	var users []*models.User
	result := r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Find(&users)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("获取用户列表失败: %w", result.Error)
	}
	
	var total int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取用户总数失败: %w", err)
	}
	
	return users, total, nil
}

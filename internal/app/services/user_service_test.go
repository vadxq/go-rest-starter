package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/vadxq/go-rest-starter/internal/app/dto"
	"github.com/vadxq/go-rest-starter/internal/app/models"
	apperrors "github.com/vadxq/go-rest-starter/pkg/errors"
)

// MockUserRepository 是 UserRepository 的模拟实现
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, tx *gorm.DB, user *models.User) error {
	args := m.Called(ctx, tx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, tx *gorm.DB, user *models.User) error {
	args := m.Called(ctx, tx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, tx *gorm.DB, id uint) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, page, pageSize int) ([]*models.User, int64, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.User), args.Get(1).(int64), args.Error(2)
}

// MockDB 是 gorm.DB 的模拟实现
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Transaction(fc func(tx *gorm.DB) error) error {
	args := m.Called(fc)
	if args.Error(0) != nil {
		return args.Error(0)
	}
	// 执行回调函数
	return fc(nil)
}

// MockCache 是缓存的模拟实现
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCache) SetObject(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCache) GetObject(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockCache) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockCache) Clear(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestUserService_CreateUser(t *testing.T) {
	// 设置测试数据
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCache)
	validator := validator.New()

	service := NewUserService(mockRepo, validator, nil, mockCache)

	ctx := context.Background()
	input := dto.CreateUserInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	// 成功创建用户的测试
	t.Run("Success", func(t *testing.T) {
		// 设置期望
		mockRepo.On("ExistsByEmail", ctx, input.Email).Return(false, nil)
		mockCache.On("Delete", ctx, userListCacheKey).Return(nil)

		// 执行测试
		user, err := service.CreateUser(ctx, input)

		// 断言
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, input.Name, user.Name)
		assert.Equal(t, input.Email, user.Email)
		assert.Equal(t, "user", user.Role)

		// 验证密码是否被正确加密
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
		assert.NoError(t, err)

		// 验证模拟调用
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	// 邮箱已存在的测试
	t.Run("EmailExists", func(t *testing.T) {
		mockRepo2 := new(MockUserRepository)
		service2 := NewUserService(mockRepo2, validator, nil, mockCache)

		// 设置期望
		mockRepo2.On("ExistsByEmail", ctx, input.Email).Return(true, nil)

		// 执行测试
		user, err := service2.CreateUser(ctx, input)

		// 断言
		assert.Error(t, err)
		assert.Nil(t, user)

		appErr, ok := err.(*apperrors.Error)
		assert.True(t, ok)
		assert.Equal(t, apperrors.ErrorTypeConflict, appErr.Type)

		// 验证模拟调用
		mockRepo2.AssertExpectations(t)
	})

	// 验证失败的测试
	t.Run("ValidationError", func(t *testing.T) {
		mockRepo3 := new(MockUserRepository)
		service3 := NewUserService(mockRepo3, validator, nil, mockCache)

		invalidInput := dto.CreateUserInput{
			Name:     "", // 空名称应该失败
			Email:    "invalid-email",
			Password: "123", // 密码太短
		}

		// 执行测试
		user, err := service3.CreateUser(ctx, invalidInput)

		// 断言
		assert.Error(t, err)
		assert.Nil(t, user)

		appErr, ok := err.(*apperrors.Error)
		assert.True(t, ok)
		assert.Equal(t, apperrors.ErrorTypeValidation, appErr.Type)
	})
}

func TestUserService_GetByID(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockCache := new(MockCache)
	validator := validator.New()
	service := NewUserService(mockRepo, validator, nil, mockCache)

	ctx := context.Background()
	userID := "1"
	expectedUser := &models.User{
		Name:  "Test User",
		Email: "test@example.com",
		Role:  "user",
	}
	expectedUser.ID = 1

	// 缓存命中的测试
	t.Run("CacheHit", func(t *testing.T) {
		cacheKey := getUserCacheKey(userID)
		mockCache.On("GetObject", ctx, cacheKey, mock.AnythingOfType("*models.User")).Return(nil).Run(func(args mock.Arguments) {
			user := args[2].(*models.User)
			*user = *expectedUser
		})

		// 执行测试
		user, err := service.GetByID(ctx, userID)

		// 断言
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, expectedUser.Name, user.Name)
		assert.Equal(t, expectedUser.Email, user.Email)

		// 验证模拟调用
		mockCache.AssertExpectations(t)
	})

	// 缓存未命中，从数据库获取的测试
	t.Run("CacheMissDBSuccess", func(t *testing.T) {
		mockRepo2 := new(MockUserRepository)
		mockCache2 := new(MockCache)
		service2 := NewUserService(mockRepo2, validator, nil, mockCache2)

		cacheKey := getUserCacheKey(userID)
		
		// 设置期望
		mockCache2.On("GetObject", ctx, cacheKey, mock.AnythingOfType("*models.User")).Return(errors.New("cache miss"))
		mockRepo2.On("GetByID", ctx, userID).Return(expectedUser, nil)
		mockCache2.On("SetObject", ctx, cacheKey, expectedUser, userCacheTTL).Return(nil)

		// 执行测试
		user, err := service2.GetByID(ctx, userID)

		// 断言
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, expectedUser.Name, user.Name)
		assert.Equal(t, expectedUser.Email, user.Email)

		// 验证模拟调用
		mockRepo2.AssertExpectations(t)
		mockCache2.AssertExpectations(t)
	})

	// 用户不存在的测试
	t.Run("UserNotFound", func(t *testing.T) {
		mockRepo3 := new(MockUserRepository)
		mockCache3 := new(MockCache)
		service3 := NewUserService(mockRepo3, validator, nil, mockCache3)

		cacheKey := getUserCacheKey(userID)

		// 设置期望
		mockCache3.On("GetObject", ctx, cacheKey, mock.AnythingOfType("*models.User")).Return(errors.New("cache miss"))
		mockRepo3.On("GetByID", ctx, userID).Return(nil, apperrors.NotFoundError("用户", nil))

		// 执行测试
		user, err := service3.GetByID(ctx, userID)

		// 断言
		assert.Error(t, err)
		assert.Nil(t, user)

		appErr, ok := err.(*apperrors.Error)
		assert.True(t, ok)
		assert.Equal(t, apperrors.ErrorTypeNotFound, appErr.Type)

		// 验证模拟调用
		mockRepo3.AssertExpectations(t)
		mockCache3.AssertExpectations(t)
	})
}
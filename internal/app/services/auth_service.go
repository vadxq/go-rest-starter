package services

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/vadxq/go-rest-starter/internal/app/dto"
	"github.com/vadxq/go-rest-starter/internal/app/repository"
	"github.com/vadxq/go-rest-starter/pkg/cache"
	apperrors "github.com/vadxq/go-rest-starter/pkg/errors"
	"github.com/vadxq/go-rest-starter/pkg/jwt"
)

const (
	// 令牌缓存键前缀
	tokenCachePrefix = "token:"

	// 令牌黑名单缓存键前缀
	tokenBlacklistPrefix = "blacklist:"
)

// AuthService 认证服务接口
type AuthService interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponse, error)
	Logout(ctx context.Context, accessToken string) error
}

// authService 认证服务实现
type authService struct {
	userRepo  repository.UserRepository
	validator *validator.Validate
	db        *gorm.DB
	jwtConfig *jwt.Config
	cache     cache.Cache
}

// NewAuthService 创建认证服务
func NewAuthService(ur repository.UserRepository, v *validator.Validate, db *gorm.DB, jwtConfig *jwt.Config, c cache.Cache) AuthService {
	return &authService{
		userRepo:  ur,
		validator: v,
		db:        db,
		jwtConfig: jwtConfig,
		cache:     c,
	}
}

// Login 用户登录
func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	// 验证请求数据
	if err := s.validator.Struct(req); err != nil {
		return nil, apperrors.ValidationError("输入数据验证失败", err)
	}

	// 获取用户
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// 不管是没找到还是数据库错误，都返回相同的错误信息，避免枚举攻击
		return nil, apperrors.UnauthorizedError("邮箱或密码错误", nil)
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, apperrors.UnauthorizedError("邮箱或密码错误", nil)
	}

	// 生成访问令牌
	accessToken, err := jwt.GenerateAccessToken(user.ID, user.Role, s.jwtConfig)
	if err != nil {
		return nil, apperrors.InternalError("生成访问令牌失败", err)
	}

	// 生成刷新令牌
	refreshToken, err := jwt.GenerateRefreshToken(user.ID, s.jwtConfig)
	if err != nil {
		return nil, apperrors.InternalError("生成刷新令牌失败", err)
	}

	// 缓存令牌 - 可以用于快速验证或令牌追踪
	tokenKey := fmt.Sprintf("%s%d", tokenCachePrefix, user.ID)
	if s.cache != nil {
		_ = s.cache.SetObject(ctx, tokenKey, map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		}, s.jwtConfig.AccessTokenExp)
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtConfig.AccessTokenExp.Seconds()),
		TokenType:    "Bearer",
		User: dto.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

// RefreshToken 刷新令牌
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponse, error) {
	// 检查令牌是否在黑名单中
	blacklistKey := fmt.Sprintf("%s%s", tokenBlacklistPrefix, refreshToken)
	var blacklisted bool
	if s.cache != nil {
		err := s.cache.GetObject(ctx, blacklistKey, &blacklisted)
		if err == nil && blacklisted {
			return nil, apperrors.UnauthorizedError("刷新令牌已被撤销", nil)
		}
	}

	// 解析刷新令牌
	userId, err := jwt.ParseRefreshToken(refreshToken, s.jwtConfig.Secret)
	if err != nil {
		return nil, apperrors.UnauthorizedError("无效的刷新令牌", nil)
	}

	// 用户ID转为字符串
	userIdStr := fmt.Sprintf("%d", userId)

	// 获取用户
	user, err := s.userRepo.GetByID(ctx, userIdStr)
	if err != nil {
		return nil, apperrors.UnauthorizedError("用户不存在", nil)
	}

	// 生成新的访问令牌
	accessToken, err := jwt.GenerateAccessToken(user.ID, user.Role, s.jwtConfig)
	if err != nil {
		return nil, apperrors.InternalError("生成访问令牌失败", err)
	}

	// 更新缓存中的令牌
	tokenKey := fmt.Sprintf("%s%d", tokenCachePrefix, user.ID)
	if s.cache != nil {
		_ = s.cache.SetObject(ctx, tokenKey, map[string]string{
			"access_token": accessToken,
		}, s.jwtConfig.AccessTokenExp)
	}

	return &dto.TokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   int64(s.jwtConfig.AccessTokenExp.Seconds()),
		TokenType:   "Bearer",
	}, nil
}

// Logout 用户登出
func (s *authService) Logout(ctx context.Context, accessToken string) error {
	// 解析令牌以获取用户ID
	claims, err := jwt.ParseToken(accessToken, s.jwtConfig.Secret)
	if err != nil {
		return apperrors.UnauthorizedError("无效的访问令牌", nil)
	}

	// 将令牌加入黑名单
	if s.cache != nil {
		blacklistKey := fmt.Sprintf("%s%s", tokenBlacklistPrefix, accessToken)
		_ = s.cache.SetObject(ctx, blacklistKey, true, s.jwtConfig.AccessTokenExp)

		// 清除用户令牌缓存
		tokenKey := fmt.Sprintf("%s%d", tokenCachePrefix, claims.UserID)
		_ = s.cache.Delete(ctx, tokenKey)
	}

	return nil
}

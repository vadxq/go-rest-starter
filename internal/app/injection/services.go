package injection

import (
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/vadxq/go-rest-starter/internal/app/config"
	"github.com/vadxq/go-rest-starter/internal/app/services"
	"github.com/vadxq/go-rest-starter/internal/pkg/cache"
	"github.com/vadxq/go-rest-starter/internal/pkg/jwt"
)

// Services 所有服务的集合
// 包含所有业务逻辑层对象，处理核心业务规则
type Services struct {
	// 用户相关业务逻辑
	UserService services.UserService
	
	// 认证相关业务逻辑
	AuthService services.AuthService
	
	// 可以在此添加更多服务...
	// ProductService services.ProductService
	// OrderService services.OrderService
}

// InitServices 初始化所有服务
// 这是依赖注入的第二层，依赖于仓库层
func InitServices(
	repos *Repositories,
	validate *validator.Validate,
	db *gorm.DB,
	config *config.AppConfig,
	cacheInstance cache.Cache,
) *Services {
	// 参数验证
	if repos == nil {
		log.Fatal().Msg("仓库依赖不能为空")
	}
	if validate == nil {
		log.Fatal().Msg("验证器不能为空")
	}
	if db == nil {
		log.Fatal().Msg("数据库连接不能为空")
	}
	if config == nil {
		log.Fatal().Msg("配置不能为空")
	}
	
	// 创建JWT配置
	jwtConfig := createJWTConfig(config)
	
	// 创建所有服务实例
	userService := services.NewUserService(repos.UserRepo, validate, db, cacheInstance)
	authService := services.NewAuthService(repos.UserRepo, validate, db, jwtConfig, cacheInstance)

	// 返回服务集合
	return &Services{
		UserService: userService,
		AuthService: authService,
	}
}

// createJWTConfig 从应用配置创建JWT配置
// 这是一个辅助函数，用于创建JWT服务所需的配置
func createJWTConfig(config *config.AppConfig) *jwt.Config {
	if config.JWT.Secret == "" {
		log.Warn().Msg("JWT密钥为空，这可能导致安全问题")
	}
	
	return &jwt.Config{
		Secret:          config.JWT.Secret,
		AccessTokenExp:  config.JWT.AccessTokenExp,
		RefreshTokenExp: config.JWT.RefreshTokenExp,
		Issuer:          config.JWT.Issuer,
	}
} 
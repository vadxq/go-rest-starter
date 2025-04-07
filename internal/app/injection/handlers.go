package injection

import (
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/vadxq/go-rest-starter/internal/app/handlers"
)

// Handlers 所有处理器的集合
// 包含所有表现层对象，负责处理HTTP请求和响应
type Handlers struct {
	// 用户相关HTTP处理器
	UserHandler *handlers.UserHandler
	
	// 认证相关HTTP处理器
	AuthHandler *handlers.AuthHandler
	
	// 可以在此添加更多处理器...
	// ProductHandler *handlers.ProductHandler
	// OrderHandler *handlers.OrderHandler
}

// InitHandlers 初始化所有处理器
// 这是依赖注入的第三层，依赖于服务层
func InitHandlers(
	services *Services,
	logger zerolog.Logger,
	validate *validator.Validate,
) *Handlers {
	// 参数验证
	if services == nil {
		log.Fatal().Msg("服务依赖不能为空")
	}
	if validate == nil {
		log.Fatal().Msg("验证器不能为空")
	}
	
	// 创建所有处理器实例
	userHandler := handlers.NewUserHandler(services.UserService, logger, validate)
	authHandler := handlers.NewAuthHandler(services.AuthService, logger, validate)
	
	// 返回处理器集合
	return &Handlers{
		UserHandler: userHandler,
		AuthHandler: authHandler,
	}
} 
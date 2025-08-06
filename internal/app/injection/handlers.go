package injection

import (
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/vadxq/go-rest-starter/internal/app/handlers"
)

// Handlers 包含所有HTTP处理器
type Handlers struct {
	UserHandler   *handlers.UserHandler
	AuthHandler   *handlers.AuthHandler
	HealthHandler *handlers.HealthHandler
}

// InitHandlers 初始化所有HTTP处理器
func InitHandlers(
	services *Services,
	logger *slog.Logger,
	validator *validator.Validate,
	db *gorm.DB,
	redis *redis.Client,
) *Handlers {
	// 初始化用户处理器
	userHandler := handlers.NewUserHandler(
		services.UserService,
		logger,
		validator,
	)

	// 初始化认证处理器
	authHandler := handlers.NewAuthHandler(
		services.AuthService,
		logger,
		validator,
	)

	// 初始化健康检查处理器
	healthHandler := handlers.NewHealthHandler(
		db,
		redis,
		logger,
	)

	return &Handlers{
		UserHandler:   userHandler,
		AuthHandler:   authHandler,
		HealthHandler: healthHandler,
	}
}

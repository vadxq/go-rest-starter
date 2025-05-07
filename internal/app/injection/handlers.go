package injection

import (
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"

	"github.com/vadxq/go-rest-starter/internal/app/handlers"
)

// Handlers 包含所有HTTP处理器
type Handlers struct {
	UserHandler *handlers.UserHandler
	AuthHandler *handlers.AuthHandler
}

// InitHandlers 初始化所有HTTP处理器
func InitHandlers(
	services *Services,
	logger zerolog.Logger,
	validator *validator.Validate,
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

	return &Handlers{
		UserHandler: userHandler,
		AuthHandler: authHandler,
	}
}

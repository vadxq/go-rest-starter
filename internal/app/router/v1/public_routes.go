package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/vadxq/go-rest-starter/internal/app/handlers"
)

// RouterConfig 路由配置
type RouterConfig struct {
	UserHandler *handlers.UserHandler
	AuthHandler *handlers.AuthHandler
	JWTSecret   string
}

// SetupPublicRoutes 设置公共路由（不需要认证）
func SetupPublicRoutes(r chi.Router, config RouterConfig) {
	// 认证相关路由
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", config.AuthHandler.Login)          // 登录
		r.Post("/refresh", config.AuthHandler.RefreshToken) // 刷新令牌
		// 可以添加注册、忘记密码等路由
	})
}

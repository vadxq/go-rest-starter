package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/vadxq/go-rest-starter/internal/app/handlers"
	custommiddleware "github.com/vadxq/go-rest-starter/internal/app/middleware"
)

// SetupProtectedRoutes 设置受保护路由（需要认证）
func SetupProtectedRoutes(r chi.Router, config RouterConfig, jwtConfig *custommiddleware.JWTConfig) {
	// 创建需要JWT认证的路由组
	r.Group(func(r chi.Router) {
		r.Use(custommiddleware.JWTAuth(jwtConfig))

		// 用户登出（需要认证的认证相关路由）
		r.Route("/account", func(r chi.Router) {
			r.Post("/logout", config.AuthHandler.Logout)
		})

		// 用户资源路由
		SetupUserRoutes(r, config.UserHandler)
	})
}

// SetupUserRoutes 设置用户相关路由
func SetupUserRoutes(r chi.Router, userHandler *handlers.UserHandler) {
	r.Route("/users", func(r chi.Router) {
		// 用户集合操作
		r.Get("/", userHandler.ListUsers)                                               // 获取用户列表
		r.With(custommiddleware.RequireRole("admin")).Post("/", userHandler.CreateUser) // 创建用户 (仅管理员)

		// 用户实例操作
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", userHandler.GetUser)       // 获取用户详情
			r.Put("/", userHandler.UpdateUser)    // 更新用户
			r.Delete("/", userHandler.DeleteUser) // 删除用户
		})
	})
}

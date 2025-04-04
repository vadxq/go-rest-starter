package v1

import (
	"github.com/go-chi/chi/v5"

	"github.com/vadxq/go-rest-starter/internal/app/handlers"
	custommiddleware "github.com/vadxq/go-rest-starter/internal/app/middleware"
)

// SetupUserRoutes 设置用户相关路由
func SetupUserRoutes(r chi.Router, userHandler *handlers.UserHandler) {
	r.Route("/users", func(r chi.Router) {
		// 列表路由
		r.Get("/", userHandler.ListUsers)

		// 创建用户 - 只有管理员可以创建用户
		r.With(custommiddleware.RequireRole("admin")).Post("/", userHandler.CreateUser)

		// 用户ID相关路由
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", userHandler.GetUser)
			r.Put("/", userHandler.UpdateUser)
			r.Delete("/", userHandler.DeleteUser)
		})
	})
} 
package v1

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/vadxq/go-rest-starter/internal/app/handlers"
	custommiddleware "github.com/vadxq/go-rest-starter/internal/app/middleware"
)

// SetupRoutes 设置API路由
func SetupRoutes(r chi.Router, userHandler *handlers.UserHandler) {
	// 全局中间件
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// 自定义中间件
	r.Use(custommiddleware.TraceID)
	r.Use(custommiddleware.LoggingMiddleware)

	// 健康检查
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API v1 路由
	r.Route("/api/v1", func(r chi.Router) {
		// 用户相关路由
		r.Route("/users", func(r chi.Router) {
			r.Post("/", userHandler.CreateUser)

			// 用户ID相关路由
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", userHandler.GetUser)
				r.Put("/", userHandler.UpdateUser)
				r.Delete("/", userHandler.DeleteUser)
			})
		})
	})
}

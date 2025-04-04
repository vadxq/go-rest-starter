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
func SetupRoutes(r chi.Router, userHandler *handlers.UserHandler, authHandler *handlers.AuthHandler) {
	// 全局中间件
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// 自定义中间件
	r.Use(custommiddleware.TraceID)
	r.Use(custommiddleware.LoggingMiddleware)
	r.Use(custommiddleware.CORSMiddleware) // 添加CORS中间件，确保API可以被跨域访问

	// 设置Swagger路由
	SetupSwaggerRoutes(r)

	// 健康检查
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API v1 路由
	r.Route("/api/v1", func(r chi.Router) {
		// 认证路由 - 不需要认证
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.RefreshToken)
		})

		// 用户相关路由
		SetupUserRoutes(r, userHandler)
	})
}

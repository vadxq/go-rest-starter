package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	v1 "github.com/vadxq/go-rest-starter/api/v1"
	"github.com/vadxq/go-rest-starter/internal/app/handlers"
	custommiddleware "github.com/vadxq/go-rest-starter/internal/app/middleware"
)

// Handler 接口用于定义处理程序集
type Handler interface {
	SetupRoutes(r chi.Router)
}

// RouterConfig 路由配置
type RouterConfig struct {
	UserHandler *handlers.UserHandler
	AuthHandler *handlers.AuthHandler
	JWTSecret   string
}

// Setup 设置所有API路由
func Setup(r chi.Router, config RouterConfig) {
	// 全局中间件
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// 自定义中间件
	r.Use(custommiddleware.TraceID)
	r.Use(custommiddleware.LoggingMiddleware)
	r.Use(custommiddleware.CORSMiddleware)

	// 设置Swagger路由
	v1.SetupSwaggerRoutes(r)

	// 健康检查
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// 认证路由 - 不需要JWT认证
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", config.AuthHandler.Login)
			r.Post("/refresh", config.AuthHandler.RefreshToken)
		})

		// 需要JWT认证的路由
		r.Group(func(r chi.Router) {
			// JWT认证中间件
			jwtConfig := &custommiddleware.JWTConfig{
				Secret: config.JWTSecret,
				ExcludePaths: []string{
					"/api/v1/auth",
					"/swagger",
					"/health",
				},
			}
			r.Use(custommiddleware.JWTAuth(jwtConfig))

			// 用户相关路由
			v1.SetupUserRoutes(r, config.UserHandler)
		})
	})
} 

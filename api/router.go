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
	// 基础中间件
	r.Use(middleware.RequestID)                            // 请求ID
	r.Use(middleware.RealIP)                              // 真实IP
	r.Use(custommiddleware.RequestContext)                // 请求上下文
	r.Use(custommiddleware.LoggingMiddleware)             // 日志
	r.Use(custommiddleware.RecoveryMiddleware)            // 恢复
	r.Use(middleware.Timeout(60 * time.Second))           // 超时
	r.Use(middleware.CleanPath)                           // 清理路径
	r.Use(middleware.StripSlashes)                        // 去除尾部斜杠
	
	// 安全中间件
	r.Use(custommiddleware.CORSMiddleware)                // 跨域
	r.Use(securityHeaders)                                // 安全头

	// API文档路由
	v1.SetupSwaggerRoutes(r)

	// 健康检查和状态监控
	setupUtilityRoutes(r)

	// API v1
	setupV1Routes(r, config)
}

// setupUtilityRoutes 设置实用路由（健康检查、状态监控等）
func setupUtilityRoutes(r chi.Router) {
	// 健康检查
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	// 版本信息
	r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"version":"1.0.0"}`))
	})
	
	// 状态监控（可扩展）
	r.Route("/status", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"running"}`))
		})
	})
}

// setupV1Routes 设置 API v1 路由
func setupV1Routes(r chi.Router, config RouterConfig) {
	r.Route("/api/v1", func(r chi.Router) {
		// 不需要认证的路由组
		r.Group(func(r chi.Router) {
			// 认证相关路由
			r.Route("/auth", func(r chi.Router) {
				r.Post("/login", config.AuthHandler.Login)
				r.Post("/refresh", config.AuthHandler.RefreshToken)
				// 可以添加注册、忘记密码等路由
			})
		})

		// 需要JWT认证的路由组
		r.Group(func(r chi.Router) {
			// JWT认证中间件
			jwtConfig := &custommiddleware.JWTConfig{
				Secret: config.JWTSecret,
				ExcludePaths: []string{
					"/api/v1/auth",
					"/swagger",
					"/health",
					"/version",
					"/status",
				},
			}
			r.Use(custommiddleware.JWTAuth(jwtConfig))

			// 用户相关路由
			v1.SetupUserRoutes(r, config.UserHandler)
			
			// 其他需要认证的资源路由可以在这里添加
		})
	})
}

// securityHeaders 添加安全相关的HTTP头
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 防止MIME类型嗅探
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// 开启XSS过滤
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		// 防止点击劫持
		w.Header().Set("X-Frame-Options", "DENY")
		// HTTP严格传输安全
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		// 引用策略
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		
		next.ServeHTTP(w, r)
	})
} 

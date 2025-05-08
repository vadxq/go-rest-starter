package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/vadxq/go-rest-starter/api/app" // 导入生成的文档
)

// SetupSwaggerRoutes 设置Swagger文档路由
func SetupSwaggerRoutes(r chi.Router) {
	// 设置Swagger UI路由
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // Swagger JSON API的URL
	))

	// 重定向根路径到Swagger UI
	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusMovedPermanently)
	})
}

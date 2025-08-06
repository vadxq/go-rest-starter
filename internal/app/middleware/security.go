package middleware

import (
	"net/http"
	"strings"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	// CSP配置
	ContentSecurityPolicy string
	// HSTS配置
	StrictTransportSecurity string
	// 允许的来源
	AllowedOrigins []string
	// 是否启用各项安全特性
	EnableCSP        bool
	EnableHSTS       bool
	EnableXSS        bool
	EnableNoSniff    bool
	EnableFrameDeny  bool
	EnableReferrer   bool
}

// DefaultSecurityConfig 默认安全配置
var DefaultSecurityConfig = SecurityConfig{
	ContentSecurityPolicy: "default-src 'self'; " +
		"script-src 'self' 'unsafe-inline'; " +
		"style-src 'self' 'unsafe-inline'; " +
		"img-src 'self' data: https:; " +
		"font-src 'self' data:; " +
		"connect-src 'self'; " +
		"media-src 'self'; " +
		"object-src 'none'; " +
		"frame-ancestors 'none'; " +
		"base-uri 'self'; " +
		"form-action 'self'; " +
		"upgrade-insecure-requests;",
	StrictTransportSecurity: "max-age=31536000; includeSubDomains; preload",
	AllowedOrigins:         []string{"https://example.com"},
	EnableCSP:              true,
	EnableHSTS:             true,
	EnableXSS:              true,
	EnableNoSniff:          true,
	EnableFrameDeny:        true,
	EnableReferrer:         true,
}

// SecurityMiddleware 安全中间件
func SecurityMiddleware(config *SecurityConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = &DefaultSecurityConfig
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 防止MIME类型嗅探
			if config.EnableNoSniff {
				w.Header().Set("X-Content-Type-Options", "nosniff")
			}

			// XSS保护
			if config.EnableXSS {
				w.Header().Set("X-XSS-Protection", "1; mode=block")
			}

			// 防止点击劫持
			if config.EnableFrameDeny {
				w.Header().Set("X-Frame-Options", "DENY")
			}

			// HTTP严格传输安全（仅在HTTPS下有效）
			if config.EnableHSTS && (r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https") {
				w.Header().Set("Strict-Transport-Security", config.StrictTransportSecurity)
			}

			// 引用策略
			if config.EnableReferrer {
				w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			}

			// 内容安全策略
			if config.EnableCSP {
				w.Header().Set("Content-Security-Policy", config.ContentSecurityPolicy)
			}

			// 权限策略（Feature Policy / Permissions Policy）
			w.Header().Set("Permissions-Policy", 
				"accelerometer=(), " +
				"camera=(), " +
				"geolocation=(), " +
				"gyroscope=(), " +
				"magnetometer=(), " +
				"microphone=(), " +
				"payment=(), " +
				"usb=()")

			// 防止浏览器缓存敏感信息
			if strings.Contains(r.URL.Path, "/api/") {
				w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// BasicSecurityHeaders 基础安全头中间件（简化版）
func BasicSecurityHeaders(next http.Handler) http.Handler {
	return SecurityMiddleware(&DefaultSecurityConfig)(next)
}

// NoCacheMiddleware 禁用缓存中间件
func NoCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}

// SecureRedirectMiddleware HTTPS重定向中间件
func SecureRedirectMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查是否已经是HTTPS
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			next.ServeHTTP(w, r)
			return
		}

		// 构建HTTPS URL
		target := "https://" + r.Host + r.URL.Path
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}

		// 执行301永久重定向
		http.Redirect(w, r, target, http.StatusMovedPermanently)
	})
}
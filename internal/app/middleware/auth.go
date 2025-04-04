package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
	
	"github.com/vadxq/go-rest-starter/api/v1/dto"
	jwtpkg "github.com/vadxq/go-rest-starter/internal/pkg/jwt"
)

// UserIDKey 用户ID键
type UserIDKey struct{}

// RoleKey 角色键
type RoleKey struct{}

// JWTConfig JWT中间件配置
type JWTConfig struct {
	Secret     string   // JWT密钥
	ExcludePaths []string // 排除的路径（不需要认证）
}

// JWTAuth JWT认证中间件
func JWTAuth(config *JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 跳过OPTIONS请求
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// 检查路径是否在排除列表中
			path := r.URL.Path
			for _, excludePath := range config.ExcludePaths {
				if strings.HasPrefix(path, excludePath) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// 从请求头中获取令牌
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				renderError(w, "缺少认证令牌", http.StatusUnauthorized)
				return
			}

			// 提取令牌
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				renderError(w, "认证令牌格式无效", http.StatusUnauthorized)
				return
			}
			tokenString := tokenParts[1]

			// 解析令牌
			claims, err := jwtpkg.ParseToken(tokenString, config.Secret)
			if err != nil {
				log.Error().Err(err).Str("token", tokenString).Msg("解析令牌失败")
				renderError(w, "无效的认证令牌", http.StatusUnauthorized)
				return
			}

			// 将用户ID和角色添加到上下文
			ctx := context.WithValue(r.Context(), UserIDKey{}, claims.UserID)
			ctx = context.WithValue(ctx, RoleKey{}, claims.Role)

			// 继续处理请求
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID 从上下文中获取用户ID
func GetUserID(ctx context.Context) (uint, bool) {
	userID, ok := ctx.Value(UserIDKey{}).(uint)
	return userID, ok
}

// GetRole 从上下文中获取角色
func GetRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(RoleKey{}).(string)
	return role, ok
}

// RequireRole 要求特定角色的中间件
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := GetRole(r.Context())
			if !ok || userRole != role {
				renderError(w, "没有权限访问", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// 渲染错误响应
func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := dto.ErrorResponse{
		Code:    statusCode,
		Message: message,
	}
	
	json.NewEncoder(w).Encode(response)
} 

package jwt

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Config JWT配置
type Config struct {
	Secret          string        // JWT密钥
	AccessTokenExp  time.Duration // 访问令牌过期时间
	RefreshTokenExp time.Duration // 刷新令牌过期时间
	Issuer          string        // 签发者
}

// Claims 自定义JWT声明
type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateAccessToken 生成访问令牌
func GenerateAccessToken(userID uint, role string, config *Config) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.AccessTokenExp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Secret))
}

// GenerateRefreshToken 生成刷新令牌
func GenerateRefreshToken(userID uint, config *Config) (string, error) {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.RefreshTokenExp)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    config.Issuer,
		Subject:   fmt.Sprintf("%d", userID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Secret))
}

// ParseToken 解析并验证访问令牌
func ParseToken(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非预期的签名方法: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("无效的令牌")
}

// ParseRefreshToken 解析并验证刷新令牌
func ParseRefreshToken(tokenString string, secret string) (uint, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非预期的签名方法: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		// 从Subject中获取用户ID
		userID, err := strconv.ParseUint(claims.Subject, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("无效的用户ID: %w", err)
		}
		return uint(userID), nil
	}

	return 0, fmt.Errorf("无效的令牌")
}

// ValidateToken 验证令牌是否有效
func ValidateToken(tokenString string, secret string) bool {
	_, err := ParseToken(tokenString, secret)
	return err == nil
} 
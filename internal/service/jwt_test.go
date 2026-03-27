package service

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// TestJWTService_ExpiredToken 测试过期 token 应该被拒绝
func TestJWTService_ExpiredToken(t *testing.T) {
	jwtService := NewJWTService("test-secret")

	// 创建一个已过期的 token
	username := "testuser"
	claims := JWTClaims{
		UserID:   123,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // 1小时前过期
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)), // 2小时前签发
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	assert.NoError(t, err)

	// 验证过期 token 应该返回错误
	_, err = jwtService.ValidateToken(tokenString)
	assert.Error(t, err, "Expired token should return error")
}

// TestJWTService_ValidToken 测试有效 token 应该通过验证
func TestJWTService_ValidToken(t *testing.T) {
	jwtService := NewJWTService("test-secret")

	username := "testuser"
	tokenString, err := jwtService.GenerateToken(123, &username)
	assert.NoError(t, err)

	// 验证有效 token 应该成功
	claims, err := jwtService.ValidateToken(tokenString)
	assert.NoError(t, err, "Valid token should pass validation")
	assert.Equal(t, uint(123), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
}

// TestJWTService_ConfigSecret 测试 JWT service 应该从配置中读取 secret
func TestJWTService_ConfigSecret(t *testing.T) {
	// 这个测试验证 JWT secret 应该来自配置
	// 而不是硬编码在代码中
	secretFromConfig := "config-secret"

	jwtService := NewJWTService(secretFromConfig)
	username := "testuser"
	tokenString, err := jwtService.GenerateToken(123, &username)
	assert.NoError(t, err)

	// 用相同的 secret 可以验证
	claims, err := jwtService.ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.Equal(t, uint(123), claims.UserID)

	// 用不同的 secret 无法验证
	jwtService2 := NewJWTService("different-secret")
	_, err = jwtService2.ValidateToken(tokenString)
	assert.Error(t, err, "Token signed with different secret should fail")
}

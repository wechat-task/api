package service

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// TestJWTService_ExpiredToken tests that expired token should be rejected
func TestJWTService_ExpiredToken(t *testing.T) {
	jwtService := NewJWTService("test-secret")

	// Create an expired token
	username := "testuser"
	claims := JWTClaims{
		UserID:   123,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)), // issued 2 hours ago
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	assert.NoError(t, err)

	// Verify that expired token returns error
	_, err = jwtService.ValidateToken(tokenString)
	assert.Error(t, err, "Expired token should return error")
}

// TestJWTService_ValidToken tests that valid token passes verification
func TestJWTService_ValidToken(t *testing.T) {
	jwtService := NewJWTService("test-secret")

	username := "testuser"
	tokenString, err := jwtService.GenerateToken(123, &username)
	assert.NoError(t, err)

	// Verify valid token succeeds
	claims, err := jwtService.ValidateToken(tokenString)
	assert.NoError(t, err, "Valid token should pass validation")
	assert.Equal(t, uint(123), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
}

// TestJWTService_ConfigSecret tests that JWT service reads secret from config
func TestJWTService_ConfigSecret(t *testing.T) {
	// This test verifies JWT secret should come from config, not hardcoded
	secretFromConfig := "config-secret"

	jwtService := NewJWTService(secretFromConfig)
	username := "testuser"
	tokenString, err := jwtService.GenerateToken(123, &username)
	assert.NoError(t, err)

	// Can verify with the same secret
	claims, err := jwtService.ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.Equal(t, uint(123), claims.UserID)

	// Cannot verify with different secret
	jwtService2 := NewJWTService("different-secret")
	_, err = jwtService2.ValidateToken(tokenString)
	assert.Error(t, err, "Token signed with different secret should fail")
}

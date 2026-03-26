package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/wechat-task/api/internal/service"
	"net/http"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) BeginAuth(c *gin.Context) {
	options, sessionID, err := h.authService.BeginAuth()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("session_id", sessionID, 300, "/", "", false, true)

	c.JSON(http.StatusOK, options)
}

func (h *AuthHandler) FinishAuth(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session not found"})
		return
	}

	user, isNewUser, err := h.authService.FinishAuth(sessionID, c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	statusCode := http.StatusOK
	if isNewUser {
		statusCode = http.StatusCreated
	}

	c.JSON(statusCode, gin.H{
		"user_id":     user.ID,
		"username":    user.Username,
		"is_new_user": isNewUser,
	})
}

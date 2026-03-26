package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/wechat-task/api/internal/model"
	"github.com/wechat-task/api/internal/service"
	"net/http"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type BeginRegistrationRequest struct {
	DisplayName string `json:"display_name" binding:"required"`
	Icon        string `json:"icon"`
}

func (h *AuthHandler) BeginRegistration(c *gin.Context) {
	var req BeginRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	options, sessionID, err := h.authService.BeginRegistration(req.DisplayName, req.Icon)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("session_id", sessionID, 300, "/", "", false, true)

	c.JSON(http.StatusOK, options)
}

func (h *AuthHandler) FinishRegistration(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session not found"})
		return
	}

	_, sessionData, err := h.authService.GetSessionService().GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session"})
		return
	}

	user := &model.User{
		DisplayName: string(sessionData.UserID),
	}
	user.SetWebAuthnID(sessionData.UserID)

	credential, err := h.authService.GetWebAuthn().FinishRegistration(user, *sessionData, c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newUser := &model.User{
		DisplayName: string(sessionData.UserID),
	}
	newUser.SetWebAuthnID(sessionData.UserID)

	if err := h.authService.GetUserRepository().Create(newUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dbCredential := &model.Credential{}
	dbCredential.FromWebAuthnCredential(*credential, newUser.ID)

	if err := h.authService.GetCredentialRepository().Create(dbCredential); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.authService.GetSessionService().DeleteSession(sessionID)

	c.JSON(http.StatusCreated, gin.H{
		"user_id":     newUser.ID,
		"webauthn_id": string(newUser.WebAuthnID()),
	})
}

func (h *AuthHandler) BeginLogin(c *gin.Context) {
	options, sessionID, err := h.authService.BeginLogin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("session_id", sessionID, 300, "/", "", false, true)

	c.JSON(http.StatusOK, options)
}

func (h *AuthHandler) FinishLogin(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session not found"})
		return
	}

	_, sessionData, err := h.authService.GetSessionService().GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session"})
		return
	}

	credentialID := extractCredentialID(c)

	dbCredential, err := h.authService.GetCredentialRepository().GetByCredentialID(credentialID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credential"})
		return
	}

	user, err := h.authService.GetUserRepository().GetByWebAuthnID(dbCredential.CredentialID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	_, err = h.authService.GetWebAuthn().FinishLogin(user, *sessionData, c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	h.authService.GetSessionService().DeleteSession(sessionID)

	c.JSON(http.StatusOK, gin.H{
		"user_id":  user.ID,
		"username": user.Username,
	})
}

func extractCredentialID(c *gin.Context) []byte {
	var response map[string]interface{}
	if err := c.ShouldBindJSON(&response); err != nil {
		return nil
	}

	if credentialID, ok := response["id"].(string); ok {
		return []byte(credentialID)
	}

	return nil
}

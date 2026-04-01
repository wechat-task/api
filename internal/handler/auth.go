package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wechat-task/api/internal/logger"
	"github.com/wechat-task/api/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
	jwtService  *service.JWTService
}

func NewAuthHandler(authService *service.AuthService, jwtService *service.JWTService) *AuthHandler {
	return &AuthHandler{authService: authService, jwtService: jwtService}
}

type RegisterOptionsRequest struct {
	Username string `json:"username" binding:"required" example:"john_doe"`
}

// RegisterOptions godoc
// @Summary      Begin passkey registration
// @Description  Generate WebAuthn registration options. User must provide a username.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body  RegisterOptionsRequest  true  "Username (required)"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /auth/passkey/register/options [post]
func (h *AuthHandler) RegisterOptions(c *gin.Context) {
	var req RegisterOptionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	options, sessionID, err := h.authService.BeginRegistration(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"options":    options,
		"session_id": sessionID,
	})
}

// RegisterVerify godoc
// @Summary      Verify passkey registration
// @Description  Complete WebAuthn registration by verifying the credential response
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body  object  true  "session_id and WebAuthn credential response"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /auth/passkey/register/verify [post]
func (h *AuthHandler) RegisterVerify(c *gin.Context) {
	sessionID, err := extractSessionID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.FinishRegistration(sessionID, c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	token, err := h.jwtService.GenerateToken(user.ID, user.Username)
	if err != nil {
		logger.Errorf("failed to generate JWT token for user %d: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token":    token,
		"user_id":  user.ID,
		"username": user.Username,
	})
}

// LoginOptions godoc
// @Summary      Begin passkey login
// @Description  Generate WebAuthn login options
// @Tags         auth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]string
// @Router       /auth/passkey/login/options [post]
func (h *AuthHandler) LoginOptions(c *gin.Context) {
	options, sessionID, err := h.authService.BeginLogin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"options": gin.H{
			"publicKey": gin.H{
				"challenge":        options.Response.Challenge,
				"timeout":          options.Response.Timeout,
				"rpId":             options.Response.RelyingPartyID,
				"allowCredentials": []interface{}{},
			},
		},
		"session_id": sessionID,
	})
}

// LoginVerify godoc
// @Summary      Verify passkey login
// @Description  Complete WebAuthn login by verifying the assertion response
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body  object  true  "session_id and WebAuthn assertion response"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /auth/passkey/login/verify [post]
func (h *AuthHandler) LoginVerify(c *gin.Context) {
	sessionID, err := extractSessionID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.FinishLogin(sessionID, c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	token, err := h.jwtService.GenerateToken(user.ID, user.Username)
	if err != nil {
		logger.Errorf("failed to generate JWT token for user %d: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":    token,
		"user_id":  user.ID,
		"username": user.Username,
	})
}

// extractSessionID reads the request body, extracts session_id, and restores the body for WebAuthn parsing.
func extractSessionID(c *gin.Context) (string, error) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return "", errors.New("failed to read request body")
	}

	var payload struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(bodyBytes, &payload); err != nil || payload.SessionID == "" {
		return "", errors.New("session_id is required")
	}

	// Restore body so WebAuthn library can parse it
	c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	return payload.SessionID, nil
}

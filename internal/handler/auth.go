package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/wechat-task/api/internal/service"
	"net/http"
)

type AuthHandler struct {
	authService *service.AuthService
	userService *service.UserService
}

func NewAuthHandler(authService *service.AuthService, userService *service.UserService) *AuthHandler {
	return &AuthHandler{authService: authService, userService: userService}
}

type BeginAuthRequest struct {
	Username string `json:"username" example:"john_doe"`
}

// BeginAuth godoc
// @Summary      Begin authentication
// @Description  Initiate Passkeys authentication flow with optional username
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      BeginAuthRequest  false  "Username (optional for registration)"
// @Success      200  {object}  protocol.CredentialCreation
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /auth/start [post]
func (h *AuthHandler) BeginAuth(c *gin.Context) {
	var req BeginAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty request body for backward compatibility
		req = BeginAuthRequest{Username: ""}
	}

	options, sessionID, err := h.authService.BeginAuth(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("session_id", sessionID, 300, "/", "", false, true)

	c.JSON(http.StatusOK, options)
}

// FinishAuth godoc
// @Summary      Finish authentication
// @Description  Complete Passkeys authentication (registration or login)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      protocol.CredentialCreationResponse  true  "WebAuthn credential response"
// @Success      200  {object}  map[string]interface{}  "Existing user"
// @Success      201  {object}  map[string]interface{}  "New user"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      401  {object}  map[string]string  "Authentication failed"
// @Router       /auth/finish [post]
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

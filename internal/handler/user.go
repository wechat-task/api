package handler

import (
	"github.com/gin-gonic/gin"
	_ "github.com/wechat-task/api/internal/model"
	"github.com/wechat-task/api/internal/service"
	"net/http"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type SetUsernameRequest struct {
	Username string `json:"username" binding:"required" example:"john_doe"`
}

// GetCurrentUser godoc
// @Summary      Get current user
// @Description  Get authenticated user's profile information
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  model.User  "User profile"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      404  {object}  map[string]string  "User not found"
// @Router       /user/me [get]
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// SetUsername godoc
// @Summary      Set username
// @Description  Set a display username for the authenticated user
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      SetUsernameRequest  true  "Username"
// @Success      200  {object}  model.User  "Updated user profile"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Router       /user/username [put]
func (h *UserHandler) SetUsername(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req SetUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.SetUsername(userID.(uint), req.Username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

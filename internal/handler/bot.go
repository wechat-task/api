package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wechat-task/api/internal/service"
)

// BotHandler handles bot management HTTP requests.
type BotHandler struct {
	botService *service.BotService
}

// NewBotHandler creates a new BotHandler.
func NewBotHandler(botService *service.BotService) *BotHandler {
	return &BotHandler{botService: botService}
}

type updateBotRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// CreateBot godoc
// @Summary      Create bot
// @Description  Create a new iLink bot binding. Returns a QR code for the user to scan. The bot status will be automatically updated to active once scanned.
// @Tags         bot
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      201  {object}  map[string]interface{}  "Created bot with QR code"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /bots [post]
func (h *BotHandler) CreateBot(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	result, err := h.botService.CreateBot(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"bot":          result.Bot,
		"qrcode_image": result.QRCodeImage,
	})
}

// ListBots godoc
// @Summary      List bots
// @Description  List all bots belonging to the authenticated user
// @Tags         bot
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.Bot
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Router       /bots [get]
func (h *BotHandler) ListBots(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	bots, err := h.botService.ListBots(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bots)
}

// GetBot godoc
// @Summary      Get bot
// @Description  Get a bot by ID
// @Tags         bot
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Bot ID"
// @Success      200  {object}  model.Bot
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      404  {object}  map[string]string  "Not found"
// @Router       /bots/{id} [get]
func (h *BotHandler) GetBot(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	botID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bot id"})
		return
	}

	bot, err := h.botService.GetBot(userID.(uint), uint(botID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	c.JSON(http.StatusOK, bot)
}

// UpdateBot godoc
// @Summary      Update bot
// @Description  Update a bot's name and description (both optional)
// @Tags         bot
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int                true  "Bot ID"
// @Param        request  body      updateBotRequest   true  "Bot updates"
// @Success      200  {object}  model.Bot
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      404  {object}  map[string]string  "Not found"
// @Router       /bots/{id} [put]
func (h *BotHandler) UpdateBot(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	botID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bot id"})
		return
	}

	var req updateBotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bot, err := h.botService.UpdateBot(userID.(uint), uint(botID), &service.UpdateBotRequest{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	c.JSON(http.StatusOK, bot)
}

// DeleteBot godoc
// @Summary      Delete bot
// @Description  Delete a bot by ID
// @Tags         bot
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Bot ID"
// @Success      200  {object}  map[string]string  "Deleted"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      404  {object}  map[string]string  "Not found"
// @Router       /bots/{id} [delete]
func (h *BotHandler) DeleteBot(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	botID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bot id"})
		return
	}

	if err := h.botService.DeleteBot(userID.(uint), uint(botID)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bot not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

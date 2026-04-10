package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/wechat-task/api/internal/service"
)

// ChannelHandler handles channel management HTTP requests.
type ChannelHandler struct {
	channelService *service.ChannelService
}

// NewChannelHandler creates a new ChannelHandler.
func NewChannelHandler(channelService *service.ChannelService) *ChannelHandler {
	return &ChannelHandler{channelService: channelService}
}

// CreateWechatClawbotChannel godoc
// @Summary      Create WeChat clawbot channel
// @Description  Create a WeChat clawbot channel for a bot. Returns a QR code for the user to scan.
// @Tags         channel
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      int  true  "Bot ID"
// @Success      201  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /bots/{id}/channels/wechat-clawbot [post]
func (h *ChannelHandler) CreateWechatClawbotChannel(c *gin.Context) {
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

	result, err := h.channelService.CreateWechatClawbotChannel(userID.(uint), uint(botID))
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "bot not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"channel":      result.Channel,
		"qrcode_image": result.QRCodeImage,
	})
}

// ListChannels godoc
// @Summary      List channels
// @Description  List all channels for a bot
// @Tags         channel
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      int  true  "Bot ID"
// @Success      200  {array}   model.Channel
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /bots/{id}/channels [get]
func (h *ChannelHandler) ListChannels(c *gin.Context) {
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

	channels, err := h.channelService.ListChannels(userID.(uint), uint(botID))
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "bot not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, channels)
}

// DeleteChannel godoc
// @Summary      Delete channel
// @Description  Delete a channel by ID
// @Tags         channel
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id         path      int  true  "Bot ID"
// @Param        channelId  path      int  true  "Channel ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /bots/{id}/channels/{channelId} [delete]
func (h *ChannelHandler) DeleteChannel(c *gin.Context) {
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

	channelID, err := strconv.ParseUint(c.Param("channelId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel id"})
		return
	}

	if err := h.channelService.DeleteChannel(userID.(uint), uint(botID), uint(channelID)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

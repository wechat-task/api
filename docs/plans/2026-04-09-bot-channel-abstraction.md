# Bot Channel Abstraction Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Refactor bot to support multiple messaging channels by introducing a `channels` table and decoupling iLink-specific logic from the bot model.

**Architecture:** Simplify `Bot` to pure identity (name + owner). Move all platform-specific connection data into a `Channel` model with JSONB config. Channel service dispatches by type — wechat_clawbot delegates to existing iLink client. Separate API endpoints per channel type.

**Tech Stack:** Go, Gin, GORM, PostgreSQL, existing iLink client (unchanged)

---

### Task 1: Add migration for channels table and bot column cleanup

**Files:**
- Create: `internal/database/migrations/000006_create_channels_table.up.sql`

**Step 1: Create migration SQL**

```sql
-- Create channels table
CREATE TABLE IF NOT EXISTS channels (
    id SERIAL PRIMARY KEY,
    bot_id INTEGER NOT NULL,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    config JSONB DEFAULT '{}',
    last_cursor TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_channels_bot_id FOREIGN KEY (bot_id) REFERENCES bots(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_channels_bot_id ON channels(bot_id);
CREATE INDEX IF NOT EXISTS idx_channels_type ON channels(type);

-- Migrate existing iLink data from bots into channels
INSERT INTO channels (bot_id, type, status, config, last_cursor, created_at, updated_at)
SELECT
    id,
    'wechat_clawbot',
    CASE
        WHEN status = 'pending' AND qrcode_id IS NOT NULL THEN 'pending'
        WHEN status = 'active' THEN 'active'
        ELSE status
    END,
    jsonb_build_object(
        'ilink_bot_id', COALESCE(ilink_bot_id, ''),
        'ilink_user_id', COALESCE(ilink_user_id, ''),
        'bot_token', COALESCE(bot_token, ''),
        'base_url', COALESCE(base_url, ''),
        'qrcode_id', COALESCE(qrcode_id, ''),
        'qrcode_image', COALESCE(qrcode_image, '')
    ),
    last_cursor,
    created_at,
    updated_at
FROM bots;

-- Set default name for bots without one
UPDATE bots SET name = 'My Bot' WHERE name IS NULL;

-- Make name NOT NULL
ALTER TABLE bots ALTER COLUMN name SET NOT NULL;

-- Drop iLink-specific columns from bots
ALTER TABLE bots DROP COLUMN IF EXISTS bot_token;
ALTER TABLE bots DROP COLUMN IF EXISTS base_url;
ALTER TABLE bots DROP COLUMN IF EXISTS ilink_bot_id;
ALTER TABLE bots DROP COLUMN IF EXISTS ilink_user_id;
ALTER TABLE bots DROP COLUMN IF EXISTS last_cursor;
ALTER TABLE bots DROP COLUMN IF EXISTS qrcode_id;
ALTER TABLE bots DROP COLUMN IF EXISTS qrcode_image;
```

**Step 2: Verify migration compiles and runs**

Run: `go build -o server .`
Expected: clean build

**Step 3: Commit**

```bash
git add internal/database/migrations/000006_create_channels_table.up.sql
git commit -m "feat: add channels table migration with bot data migration"
```

---

### Task 2: Create Channel model

**Files:**
- Modify: `internal/model/bot.go` — remove iLink fields, make Name required
- Create: `internal/model/channel.go`

**Step 1: Update Bot model**

Replace entire `internal/model/bot.go` with:

```go
package model

import "time"

// Bot represents a user's bot that can be bound to multiple messaging channels.
type Bot struct {
	ID          uint      `json:"id" gorm:"primaryKey" example:"1"`
	UserID      uint      `json:"user_id" gorm:"not null;index" example:"1"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null" example:"My Bot"`
	Description *string   `json:"description" gorm:"type:text" example:"Work assistant bot"`
	Status      string    `json:"status" gorm:"not null;default:pending" example:"pending"` // pending, active, disconnected, expired
	Channels    []Channel `json:"channels,omitempty" gorm:"foreignKey:BotID"`
	CreatedAt   time.Time `json:"created_at" example:"2026-03-30T10:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2026-03-30T10:00:00Z"`
}
```

**Step 2: Create Channel model**

Create `internal/model/channel.go`:

```go
package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// ChannelType enumerates supported channel types.
type ChannelType string

const (
	ChannelTypeWechatClawbot ChannelType = "wechat_clawbot"
	ChannelTypeLark          ChannelType = "lark"
)

// Channel represents a messaging channel bound to a bot.
type Channel struct {
	ID         uint        `json:"id" gorm:"primaryKey" example:"1"`
	BotID      uint        `json:"bot_id" gorm:"not null;index" example:"1"`
	Type       ChannelType `json:"type" gorm:"type:varchar(50);not null" example:"wechat_clawbot"`
	Status     string      `json:"status" gorm:"not null;default:pending" example:"pending"` // pending, active, disconnected, expired
	Config     ChannelConfig `json:"config" gorm:"type:jsonb;default:'{}'"`
	LastCursor *string     `json:"last_cursor,omitempty" gorm:"column:last_cursor"`
	CreatedAt  time.Time   `json:"created_at" example:"2026-03-30T10:00:00Z"`
	UpdatedAt  time.Time   `json:"updated_at" example:"2026-03-30T10:00:00Z"`
}

// ChannelConfig stores type-specific configuration as JSONB.
type ChannelConfig map[string]interface{}

// Scan implements sql.Scanner for ChannelConfig.
func (c *ChannelConfig) Scan(value interface{}) error {
	if value == nil {
		*c = ChannelConfig{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, c)
}

// Value implements driver.Valuer for ChannelConfig.
func (c ChannelConfig) Value() (driver.Value, error) {
	if c == nil {
		return "{}", nil
	}
	bytes, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

// GetString retrieves a string value from config.
func (c ChannelConfig) GetString(key string) string {
	if v, ok := c[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// Set stores a key-value pair in config.
func (c ChannelConfig) Set(key string, value interface{}) {
	c[key] = value
}
```

**Step 3: Verify compilation**

Run: `go fmt ./... && go build -o server .`
Expected: clean build

**Step 4: Commit**

```bash
git add internal/model/bot.go internal/model/channel.go
git commit -m "feat: update Bot model and add Channel model with JSONB config"
```

---

### Task 3: Create ChannelRepository

**Files:**
- Create: `internal/repository/channel.go`
- Modify: `internal/repository/bot.go` — remove iLink-specific queries, preload channels

**Step 1: Update BotRepository**

Replace entire `internal/repository/bot.go` with:

```go
package repository

import (
	"github.com/wechat-task/api/internal/model"
	"gorm.io/gorm"
)

// BotRepository handles database operations for bots.
type BotRepository struct {
	db *gorm.DB
}

// NewBotRepository creates a new BotRepository.
func NewBotRepository(db *gorm.DB) *BotRepository {
	return &BotRepository{db: db}
}

// Create inserts a new bot record.
func (r *BotRepository) Create(bot *model.Bot) error {
	return r.db.Create(bot).Error
}

// GetByID returns a bot by its ID, with channels preloaded.
func (r *BotRepository) GetByID(id uint) (*model.Bot, error) {
	var bot model.Bot
	err := r.db.Preload("Channels").First(&bot, id).Error
	return &bot, err
}

// GetByUserID returns all bots belonging to a user, with channels preloaded.
func (r *BotRepository) GetByUserID(userID uint) ([]model.Bot, error) {
	var bots []model.Bot
	err := r.db.Preload("Channels").Where("user_id = ?", userID).Find(&bots).Error
	return bots, err
}

// Update saves bot changes.
func (r *BotRepository) Update(bot *model.Bot) error {
	return r.db.Save(bot).Error
}

// Delete removes a bot by ID (cascades to channels via FK).
func (r *BotRepository) Delete(id uint) error {
	return r.db.Delete(&model.Bot{}, id).Error
}
```

**Step 2: Create ChannelRepository**

Create `internal/repository/channel.go`:

```go
package repository

import (
	"github.com/wechat-task/api/internal/model"
	"gorm.io/gorm"
)

// ChannelRepository handles database operations for channels.
type ChannelRepository struct {
	db *gorm.DB
}

// NewChannelRepository creates a new ChannelRepository.
func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

// Create inserts a new channel record.
func (r *ChannelRepository) Create(channel *model.Channel) error {
	return r.db.Create(channel).Error
}

// GetByID returns a channel by its ID.
func (r *ChannelRepository) GetByID(id uint) (*model.Channel, error) {
	var channel model.Channel
	err := r.db.First(&channel, id).Error
	return &channel, err
}

// GetByBotID returns all channels for a bot.
func (r *ChannelRepository) GetByBotID(botID uint) ([]model.Channel, error) {
	var channels []model.Channel
	err := r.db.Where("bot_id = ?", botID).Find(&channels).Error
	return channels, err
}

// Update saves channel changes.
func (r *ChannelRepository) Update(channel *model.Channel) error {
	return r.db.Save(channel).Error
}

// Delete removes a channel by ID.
func (r *ChannelRepository) Delete(id uint) error {
	return r.db.Delete(&model.Channel{}, id).Error
}

// GetByStatus returns all channels with the given status.
func (r *ChannelRepository) GetByStatus(status string) ([]model.Channel, error) {
	var channels []model.Channel
	err := r.db.Where("status = ?", status).Find(&channels).Error
	return channels, err
}

// GetByType returns all channels of a given type and status.
func (r *ChannelRepository) GetByType(channelType model.ChannelType, status string) ([]model.Channel, error) {
	var channels []model.Channel
	err := r.db.Where("type = ? AND status = ?", channelType, status).Find(&channels).Error
	return channels, err
}
```

**Step 3: Verify compilation**

Run: `go fmt ./... && go build -o server .`
Expected: clean build

**Step 4: Commit**

```bash
git add internal/repository/bot.go internal/repository/channel.go
git commit -m "feat: add ChannelRepository and update BotRepository to preload channels"
```

---

### Task 4: Refactor BotService to pure CRUD

**Files:**
- Modify: `internal/service/bot.go` — remove iLink/QR code logic, make it pure CRUD

**Step 1: Replace BotService**

Replace entire `internal/service/bot.go` with:

```go
package service

import (
	"errors"
	"fmt"

	"github.com/wechat-task/api/internal/model"
	"github.com/wechat-task/api/internal/repository"
)

// BotService handles bot management business logic.
type BotService struct {
	repo *repository.BotRepository
}

// NewBotService creates a new BotService.
func NewBotService(repo *repository.BotRepository) *BotService {
	return &BotService{repo: repo}
}

// CreateBotRequest holds fields for creating a bot.
type CreateBotRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

// CreateBot creates a new bot.
func (s *BotService) CreateBot(userID uint, req *CreateBotRequest) (*model.Bot, error) {
	bot := &model.Bot{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Status:      "pending",
	}

	if err := s.repo.Create(bot); err != nil {
		return nil, fmt.Errorf("create bot: %w", err)
	}

	return bot, nil
}

// GetBot returns a bot by ID, verifying ownership.
func (s *BotService) GetBot(userID, botID uint) (*model.Bot, error) {
	bot, err := s.repo.GetByID(botID)
	if err != nil {
		return nil, err
	}
	if bot.UserID != userID {
		return nil, errors.New("bot not found")
	}
	return bot, nil
}

// ListBots returns all bots for a user.
func (s *BotService) ListBots(userID uint) ([]model.Bot, error) {
	return s.repo.GetByUserID(userID)
}

// UpdateBotRequest holds optional fields for updating a bot.
type UpdateBotRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// UpdateBot updates a bot's name and description.
func (s *BotService) UpdateBot(userID, botID uint, req *UpdateBotRequest) (*model.Bot, error) {
	bot, err := s.GetBot(userID, botID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		bot.Name = *req.Name
	}
	if req.Description != nil {
		bot.Description = req.Description
	}

	if err := s.repo.Update(bot); err != nil {
		return nil, fmt.Errorf("update bot: %w", err)
	}

	return bot, nil
}

// DeleteBot removes a bot, verifying ownership.
func (s *BotService) DeleteBot(userID, botID uint) error {
	bot, err := s.GetBot(userID, botID)
	if err != nil {
		return err
	}
	return s.repo.Delete(bot.ID)
}
```

**Step 2: Verify compilation**

Run: `go fmt ./... && go build -o server .`
Expected: clean build (note: channel service not yet created, will be in next task)

**Step 3: Commit**

```bash
git add internal/service/bot.go
git commit -m "refactor: simplify BotService to pure CRUD without iLink logic"
```

---

### Task 5: Create ChannelService with wechat_clawbot support

**Files:**
- Create: `internal/service/channel.go`

**Step 1: Create ChannelService**

Create `internal/service/channel.go`:

```go
package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/wechat-task/api/internal/ilink"
	"github.com/wechat-task/api/internal/logger"
	"github.com/wechat-task/api/internal/model"
	"github.com/wechat-task/api/internal/repository"
)

// ChannelService handles channel management business logic.
type ChannelService struct {
	channelRepo *repository.ChannelRepository
	botRepo     *repository.BotRepository
	ilinkCli    *ilink.Client
}

// NewChannelService creates a new ChannelService.
func NewChannelService(channelRepo *repository.ChannelRepository, botRepo *repository.BotRepository) *ChannelService {
	return &ChannelService{
		channelRepo: channelRepo,
		botRepo:     botRepo,
		ilinkCli:    ilink.NewClient(""),
	}
}

// CreateWechatClawbotResult is returned when creating a wechat_clawbot channel.
type CreateWechatClawbotResult struct {
	Channel     *model.Channel `json:"channel"`
	QRCodeImage string         `json:"qrcode_image"`
}

// CreateWechatClawbotChannel creates a new wechat_clawbot channel with QR code.
func (s *ChannelService) CreateWechatClawbotChannel(userID, botID uint) (*CreateWechatClawbotResult, error) {
	if err := s.verifyBotOwnership(userID, botID); err != nil {
		return nil, err
	}

	qrResp, err := s.ilinkCli.GetQRCode(3)
	if err != nil {
		return nil, fmt.Errorf("get qrcode: %w", err)
	}

	config := model.ChannelConfig{
		"qrcode_id":    qrResp.QRCode,
		"qrcode_image": qrResp.QRCodeImgContent,
	}

	channel := &model.Channel{
		BotID:  botID,
		Type:   model.ChannelTypeWechatClawbot,
		Status: "pending",
		Config: config,
	}

	if err := s.channelRepo.Create(channel); err != nil {
		return nil, fmt.Errorf("create channel: %w", err)
	}

	go s.pollWechatQRCodeStatus(channel.ID, qrResp.QRCode)

	logger.Infof("WechatClawbot channel created (id=%d) for bot (id=%d), QR polling started", channel.ID, botID)
	return &CreateWechatClawbotResult{
		Channel:     channel,
		QRCodeImage: qrResp.QRCodeImgContent,
	}, nil
}

// pollWechatQRCodeStatus runs in a goroutine to check QR code scan status.
func (s *ChannelService) pollWechatQRCodeStatus(channelID uint, qrcodeID string) {
	confirmed, err := s.ilinkCli.WaitForConfirmation(qrcodeID, 2*time.Second, 5*time.Minute)
	if err != nil {
		logger.Errorf("Channel (id=%d) QR code polling failed: %v", channelID, err)
		ch, findErr := s.channelRepo.GetByID(channelID)
		if findErr != nil {
			return
		}
		if ch.Status == "pending" {
			ch.Status = "expired"
			_ = s.channelRepo.Update(ch)
		}
		return
	}

	ch, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		logger.Errorf("Channel (id=%d) not found after confirmation: %v", channelID, err)
		return
	}

	ch.Status = "active"
	ch.Config = model.ChannelConfig{
		"ilink_bot_id": confirmed.ILinkBotID,
		"ilink_user_id": confirmed.ILinkUserID,
		"bot_token":    confirmed.BotToken,
		"base_url":     confirmed.BaseURL,
	}

	if err := s.channelRepo.Update(ch); err != nil {
		logger.Errorf("Channel (id=%d) failed to update after confirmation: %v", channelID, err)
		return
	}

	// Update bot status to active
	bot, err := s.botRepo.GetByID(ch.BotID)
	if err == nil && bot.Status == "pending" {
		bot.Status = "active"
		_ = s.botRepo.Update(bot)
	}

	logger.Infof("Channel (id=%d) activated: ilink_bot_id=%s", channelID, confirmed.ILinkBotID)
}

// ListChannels returns all channels for a bot.
func (s *ChannelService) ListChannels(userID, botID uint) ([]model.Channel, error) {
	if err := s.verifyBotOwnership(userID, botID); err != nil {
		return nil, err
	}
	return s.channelRepo.GetByBotID(botID)
}

// DeleteChannel removes a channel, verifying ownership via bot.
func (s *ChannelService) DeleteChannel(userID, botID, channelID uint) error {
	if err := s.verifyBotOwnership(userID, botID); err != nil {
		return err
	}

	channel, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return errors.New("channel not found")
	}
	if channel.BotID != botID {
		return errors.New("channel not found")
	}

	return s.channelRepo.Delete(channelID)
}

// RecoverPendingChannels finds all pending channels and restarts their connection flows.
func (s *ChannelService) RecoverPendingChannels() {
	channels, err := s.channelRepo.GetByStatus("pending")
	if err != nil {
		logger.Errorf("Failed to recover pending channels: %v", err)
		return
	}

	if len(channels) == 0 {
		return
	}

	logger.Infof("Recovering %d pending channel(s)", len(channels))
	for _, ch := range channels {
		switch ch.Type {
		case model.ChannelTypeWechatClawbot:
			qrcodeID := ch.Config.GetString("qrcode_id")
			if qrcodeID == "" {
				ch.Status = "expired"
				_ = s.channelRepo.Update(&ch)
				continue
			}
			go s.pollWechatQRCodeStatus(ch.ID, qrcodeID)
		default:
			logger.Warnf("Unknown channel type during recovery: %s", ch.Type)
		}
	}
}

// verifyBotOwnership checks that the bot belongs to the user.
func (s *ChannelService) verifyBotOwnership(userID, botID uint) error {
	bot, err := s.botRepo.GetByID(botID)
	if err != nil {
		return errors.New("bot not found")
	}
	if bot.UserID != userID {
		return errors.New("bot not found")
	}
	return nil
}
```

**Step 2: Verify compilation**

Run: `go fmt ./... && go build -o server .`
Expected: clean build

**Step 3: Commit**

```bash
git add internal/service/channel.go
git commit -m "feat: add ChannelService with wechat_clawbot QR code support"
```

---

### Task 6: Refactor BotHandler and create ChannelHandler

**Files:**
- Modify: `internal/handler/bot.go` — require name in create, use new service signatures
- Create: `internal/handler/channel.go`

**Step 1: Refactor BotHandler**

Replace entire `internal/handler/bot.go` with:

```go
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

type createBotRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

type updateBotRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// CreateBot godoc
// @Summary      Create bot
// @Description  Create a new bot with a name. Channels can be added separately.
// @Tags         bot
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      createBotRequest  true  "Bot creation data"
// @Success      201  {object}  model.Bot
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /bots [post]
func (h *BotHandler) CreateBot(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req createBotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	bot, err := h.botService.CreateBot(userID.(uint), &service.CreateBotRequest{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, bot)
}

// ListBots godoc
// @Summary      List bots
// @Description  List all bots belonging to the authenticated user, including their channels
// @Tags         bot
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   model.Bot
// @Failure      401  {object}  map[string]string
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
// @Description  Get a bot by ID with its channels
// @Tags         bot
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Bot ID"
// @Success      200  {object}  model.Bot
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
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
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
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
// @Description  Delete a bot by ID (cascades to channels)
// @Tags         bot
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Bot ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
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
```

**Step 2: Create ChannelHandler**

Create `internal/handler/channel.go`:

```go
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
// @Param        botId  path      int  true  "Bot ID"
// @Success      201  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /bots/{botId}/channels/wechat-clawbot [post]
func (h *ChannelHandler) CreateWechatClawbotChannel(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	botID, err := strconv.ParseUint(c.Param("botId"), 10, 32)
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
// @Param        botId  path      int  true  "Bot ID"
// @Success      200  {array}   model.Channel
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /bots/{botId}/channels [get]
func (h *ChannelHandler) ListChannels(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	botID, err := strconv.ParseUint(c.Param("botId"), 10, 32)
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
// @Param        botId      path      int  true  "Bot ID"
// @Param        channelId  path      int  true  "Channel ID"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /bots/{botId}/channels/{channelId} [delete]
func (h *ChannelHandler) DeleteChannel(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	botID, err := strconv.ParseUint(c.Param("botId"), 10, 32)
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
```

**Step 3: Verify compilation**

Run: `go fmt ./... && go build -o server .`
Expected: clean build

**Step 4: Commit**

```bash
git add internal/handler/bot.go internal/handler/channel.go
git commit -m "feat: refactor BotHandler and add ChannelHandler with type-specific endpoints"
```

---

### Task 7: Wire dependencies and routes in main.go

**Files:**
- Modify: `main.go` — add ChannelRepository, ChannelService, ChannelHandler, new routes

**Step 1: Update main.go**

Replace lines 80-88 and 114-122 in `main.go`:

Old (lines 80-88):
```go
	botRepo := repository.NewBotRepository(db)
	botService := service.NewBotService(botRepo)
	botService.RecoverPendingBots()

	jwtService := service.NewJWTService(cfg.JWT.Secret)

	authHandler := handler.NewAuthHandler(authService, jwtService)
	userHandler := handler.NewUserHandler(userService)
	botHandler := handler.NewBotHandler(botService)
```

New:
```go
	botRepo := repository.NewBotRepository(db)
	channelRepo := repository.NewChannelRepository(db)

	botService := service.NewBotService(botRepo)
	channelService := service.NewChannelService(channelRepo, botRepo)
	channelService.RecoverPendingChannels()

	jwtService := service.NewJWTService(cfg.JWT.Secret)

	authHandler := handler.NewAuthHandler(authService, jwtService)
	userHandler := handler.NewUserHandler(userService)
	botHandler := handler.NewBotHandler(botService)
	channelHandler := handler.NewChannelHandler(channelService)
```

Old (lines 114-122):
```go
	bots := r.Group("/api/v1/bots")
	bots.Use(middleware.Auth(jwtService))
	{
		bots.POST("", botHandler.CreateBot)
		bots.GET("", botHandler.ListBots)
		bots.GET("/:id", botHandler.GetBot)
		bots.PUT("/:id", botHandler.UpdateBot)
		bots.DELETE("/:id", botHandler.DeleteBot)
	}
```

New:
```go
	bots := r.Group("/api/v1/bots")
	bots.Use(middleware.Auth(jwtService))
	{
		bots.POST("", botHandler.CreateBot)
		bots.GET("", botHandler.ListBots)
		bots.GET("/:id", botHandler.GetBot)
		bots.PUT("/:id", botHandler.UpdateBot)
		bots.DELETE("/:id", botHandler.DeleteBot)

		channels := bots.Group("/:botId/channels")
		{
			channels.POST("/wechat-clawbot", channelHandler.CreateWechatClawbotChannel)
			channels.GET("", channelHandler.ListChannels)
			channels.DELETE("/:channelId", channelHandler.DeleteChannel)
		}
	}
```

**Step 2: Verify compilation**

Run: `go fmt ./... && go mod tidy && go build -o server .`
Expected: clean build

**Step 3: Commit**

```bash
git add main.go
git commit -m "feat: wire ChannelService and register channel routes in main.go"
```

---

### Task 8: Verify everything works end-to-end

**Step 1: Run all tests**

Run: `go test ./...`
Expected: all pass

**Step 2: Run full verification**

Run: `go fmt ./... && go mod tidy && go build -o server .`
Expected: clean

**Step 3: Final commit (if any formatting fixes needed)**

```bash
git add -A
git commit -m "chore: final cleanup after bot-channel refactoring"
```

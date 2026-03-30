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

// BotService handles bot management business logic.
type BotService struct {
	repo     *repository.BotRepository
	ilinkCli *ilink.Client
}

// NewBotService creates a new BotService.
func NewBotService(repo *repository.BotRepository) *BotService {
	return &BotService{
		repo:     repo,
		ilinkCli: ilink.NewClient(""),
	}
}

// CreateBotResult is returned by CreateBot.
type CreateBotResult struct {
	Bot         *model.Bot
	QRCodeImage string
}

// CreateBot creates a new bot in pending status with a QR code for binding.
func (s *BotService) CreateBot(userID uint) (*CreateBotResult, error) {
	qrResp, err := s.ilinkCli.GetQRCode(3)
	if err != nil {
		return nil, fmt.Errorf("get qrcode: %w", err)
	}

	qrcodeID := qrResp.QRCode
	bot := &model.Bot{
		UserID:   userID,
		Status:   "pending",
		QRCodeID: &qrcodeID,
	}

	if err := s.repo.Create(bot); err != nil {
		return nil, fmt.Errorf("create bot: %w", err)
	}

	// Start background polling for QR code confirmation
	go s.pollQRCodeStatus(bot.ID, qrcodeID)

	logger.Infof("Bot created (id=%d), QR code polling started", bot.ID)
	return &CreateBotResult{Bot: bot, QRCodeImage: qrResp.QRCodeImgContent}, nil
}

// pollQRCodeStatus runs in a goroutine to check QR code scan status.
func (s *BotService) pollQRCodeStatus(botID uint, qrcodeID string) {
	confirmed, err := s.ilinkCli.WaitForConfirmation(qrcodeID, 2*time.Second, 5*time.Minute)
	if err != nil {
		logger.Errorf("Bot (id=%d) QR code polling failed: %v", botID, err)
		// Mark as expired
		bot, findErr := s.repo.GetByID(botID)
		if findErr != nil {
			return
		}
		if bot.Status == "pending" {
			bot.Status = "expired"
			_ = s.repo.Update(bot)
		}
		return
	}

	bot, err := s.repo.GetByID(botID)
	if err != nil {
		logger.Errorf("Bot (id=%d) not found after confirmation: %v", botID, err)
		return
	}

	bot.Status = "active"
	bot.BotToken = &confirmed.BotToken
	bot.BaseURL = &confirmed.BaseURL
	bot.ILinkBotID = &confirmed.ILinkBotID
	bot.ILinkUserID = &confirmed.ILinkUserID
	bot.QRCodeID = nil

	if err := s.repo.Update(bot); err != nil {
		logger.Errorf("Bot (id=%d) failed to update after confirmation: %v", botID, err)
		return
	}

	logger.Infof("Bot (id=%d) activated: ilink_bot_id=%s", botID, confirmed.ILinkBotID)
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
		bot.Name = req.Name
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

// RecoverPendingBots finds all pending bots and restarts QR code polling.
// Should be called on server startup to handle restarts gracefully.
func (s *BotService) RecoverPendingBots() {
	bots, err := s.repo.GetByStatus("pending")
	if err != nil {
		logger.Errorf("Failed to recover pending bots: %v", err)
		return
	}

	if len(bots) == 0 {
		return
	}

	logger.Infof("Recovering %d pending bot(s)", len(bots))
	for i := range bots {
		if bots[i].QRCodeID == nil {
			// No QR code to poll, mark as expired
			bots[i].Status = "expired"
			_ = s.repo.Update(&bots[i])
			continue
		}
		go s.pollQRCodeStatus(bots[i].ID, *bots[i].QRCodeID)
	}
}

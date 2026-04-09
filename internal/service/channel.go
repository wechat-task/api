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
		"ilink_bot_id":  confirmed.ILinkBotID,
		"ilink_user_id": confirmed.ILinkUserID,
		"bot_token":     confirmed.BotToken,
		"base_url":      confirmed.BaseURL,
	}

	if err := s.channelRepo.Update(ch); err != nil {
		logger.Errorf("Channel (id=%d) failed to update after confirmation: %v", channelID, err)
		return
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

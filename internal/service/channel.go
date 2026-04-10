package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/wechat-task/api/internal/channels"
	"github.com/wechat-task/api/internal/logger"
	"github.com/wechat-task/api/internal/model"
	"github.com/wechat-task/api/internal/repository"
)

// ChannelService handles channel management business logic.
type ChannelService struct {
	channelRepo *repository.ChannelRepository
	botRepo     *repository.BotRepository
	providers   channels.Registry
}

// NewChannelService creates a new ChannelService.
func NewChannelService(channelRepo *repository.ChannelRepository, botRepo *repository.BotRepository, providers channels.Registry) *ChannelService {
	return &ChannelService{
		channelRepo: channelRepo,
		botRepo:     botRepo,
		providers:   providers,
	}
}

// CreateChannelResult is the generic result for channel creation.
type CreateChannelResult struct {
	Channel *model.Channel `json:"channel"`
	Display map[string]any `json:"display,omitempty"`
}

// CreateChannel creates a new channel of the specified type.
func (s *ChannelService) CreateChannel(userID, botID uint, channelType model.ChannelType, params map[string]any) (*CreateChannelResult, error) {
	if err := s.verifyBotOwnership(userID, botID); err != nil {
		return nil, err
	}

	provider := s.providers.Get(channelType)
	if provider == nil {
		return nil, fmt.Errorf("unsupported channel type: %s", channelType)
	}

	result, err := provider.Connect(context.Background(), params)
	if err != nil {
		return nil, fmt.Errorf("connect channel: %w", err)
	}

	channel := &model.Channel{
		BotID:  botID,
		Type:   channelType,
		Status: result.Status,
		Config: result.Config,
	}

	if err := s.channelRepo.Create(channel); err != nil {
		return nil, fmt.Errorf("create channel: %w", err)
	}

	if result.Status == "pending" {
		go s.waitForConfirmation(channel, provider)
	}

	logger.Infof("Channel created (id=%d, type=%s, status=%s) for bot (id=%d)",
		channel.ID, channelType, result.Status, botID)

	return &CreateChannelResult{
		Channel: channel,
		Display: result.Display,
	}, nil
}

// waitForConfirmation polls for channel activation.
func (s *ChannelService) waitForConfirmation(channel *model.Channel, provider channels.Provider) {
	result, err := provider.Recover(context.Background(), channel)
	if err != nil {
		logger.Errorf("Channel (id=%d) confirmation failed: %v", channel.ID, err)
		ch, findErr := s.channelRepo.GetByID(channel.ID)
		if findErr != nil {
			return
		}
		if ch.Status == "pending" {
			ch.Status = "expired"
			_ = s.channelRepo.Update(ch)
		}
		return
	}

	if result == nil {
		return
	}

	ch, err := s.channelRepo.GetByID(channel.ID)
	if err != nil {
		logger.Errorf("Channel (id=%d) not found after confirmation: %v", channel.ID, err)
		return
	}

	ch.Status = result.Status
	for k, v := range result.Config {
		ch.Config[k] = v
	}

	if err := s.channelRepo.Update(ch); err != nil {
		logger.Errorf("Channel (id=%d) failed to update after confirmation: %v", channel.ID, err)
		return
	}

	logger.Infof("Channel (id=%d) activated: status=%s", channel.ID, result.Status)
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
	pendingChannels, err := s.channelRepo.GetByStatus("pending")
	if err != nil {
		logger.Errorf("Failed to recover pending channels: %v", err)
		return
	}

	if len(pendingChannels) == 0 {
		return
	}

	logger.Infof("Recovering %d pending channel(s)", len(pendingChannels))
	for _, ch := range pendingChannels {
		provider := s.providers.Get(ch.Type)
		if provider == nil {
			logger.Warnf("Unknown channel type during recovery: %s", ch.Type)
			continue
		}
		go s.waitForConfirmation(&ch, provider)
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

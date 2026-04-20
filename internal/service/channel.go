package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/wechat-task/api/internal/channels"
	ilink "github.com/wechat-task/api/internal/channels/ilink"
	"github.com/wechat-task/api/internal/logger"
	"github.com/wechat-task/api/internal/model"
	"github.com/wechat-task/api/internal/repository"
)

// ChannelService handles channel management business logic.
type ChannelService struct {
	channelRepo *repository.ChannelRepository
	botRepo     *repository.BotRepository
	ctxRepo     *repository.ChannelContextRepository
	providers   channels.Registry
}

// NewChannelService creates a new ChannelService.
func NewChannelService(
	channelRepo *repository.ChannelRepository,
	botRepo *repository.BotRepository,
	ctxRepo *repository.ChannelContextRepository,
	providers channels.Registry,
) *ChannelService {
	return &ChannelService{
		channelRepo: channelRepo,
		botRepo:     botRepo,
		ctxRepo:     ctxRepo,
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

	// Start polling for messages after activation
	go s.startPolling(ch)
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

// StartActiveChannelPollers starts message pollers for all active wechat_clawbot channels.
func (s *ChannelService) StartActiveChannelPollers() {
	activeChannels, err := s.channelRepo.GetByType(model.ChannelTypeWechatClawbot, "active")
	if err != nil {
		logger.Errorf("Failed to get active wechat channels: %v", err)
		return
	}

	if len(activeChannels) == 0 {
		return
	}

	logger.Infof("Starting pollers for %d active wechat_clawbot channel(s)", len(activeChannels))
	for _, ch := range activeChannels {
		go s.startPolling(&ch)
	}
}

// startPolling starts a long-poll loop for a wechat_clawbot channel.
func (s *ChannelService) startPolling(channel *model.Channel) {
	botToken := channel.Config.GetString("bot_token")
	baseURL := channel.Config.GetString("base_url")
	if botToken == "" {
		logger.Errorf("Channel (id=%d) missing bot_token, cannot poll", channel.ID)
		return
	}

	client := ilink.NewAuthenticatedClient(botToken, baseURL)

	// Use last_cursor if available for resuming
	cursor := ""
	if channel.LastCursor != nil {
		cursor = *channel.LastCursor
	}

	logger.Infof("Polling started for channel (id=%d)", channel.ID)

	err := client.PollLoop(context.Background(), cursor, func(msg ilink.WeixinMessage) error {
		// Extract text from message
		var text string
		for _, item := range msg.ItemList {
			if item.TextItem != nil {
				text = item.TextItem.Text
				break
			}
		}

		logger.Infof("Channel (id=%d) received message from %s: %s",
			channel.ID, msg.FromUserID, text)

		// Store contextToken for this user
		ctx := &model.ChannelContext{
			ChannelID:    channel.ID,
			UserID:       msg.FromUserID,
			ContextToken: msg.ContextToken,
		}
		if text != "" {
			ctx.LastMessage = &text
		}
		if err := s.ctxRepo.Upsert(ctx); err != nil {
			logger.Errorf("Failed to store context for channel=%d user=%s: %v",
				channel.ID, msg.FromUserID, err)
		}

		// Update channel's last_cursor for resume after restart
		return nil
	})

	if err != nil {
		logger.Errorf("Channel (id=%d) polling stopped with error: %v", channel.ID, err)
	}
}

// SendMessage sends a text message to a channel.
func (s *ChannelService) SendMessage(userID, botID, channelID uint, text string) error {
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
	if channel.Status != "active" {
		return errors.New("channel is not active")
	}

	// WeChat clawbot does not support proactive messaging
	if channel.Type == model.ChannelTypeWechatClawbot {
		return errors.New("WeChat channel does not support proactive messaging, messages can only be sent as replies to users")
	}

	provider := s.providers.Get(channel.Type)
	if provider == nil {
		return fmt.Errorf("unsupported channel type: %s", channel.Type)
	}

	if err := provider.SendText(context.Background(), channel.Config, "", text); err != nil {
		logger.Errorf("Failed to send message via channel (id=%d, type=%s): %v", channelID, channel.Type, err)
		return err
	}

	logger.Infof("Message sent via channel (id=%d, type=%s)", channelID, channel.Type)
	return nil
}

// DeliverToChannel sends a message to a channel without ownership checks (for system-initiated delivery).
func (s *ChannelService) DeliverToChannel(ctx context.Context, botID, channelID uint, text string) error {
	channel, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return fmt.Errorf("channel not found: %w", err)
	}
	if channel.BotID != botID {
		return fmt.Errorf("channel does not belong to bot")
	}
	if channel.Status != "active" {
		return fmt.Errorf("channel is not active")
	}
	provider := s.providers.Get(channel.Type)
	if provider == nil {
		return fmt.Errorf("unsupported channel type: %s", channel.Type)
	}
	return provider.SendText(ctx, channel.Config, "", text)
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

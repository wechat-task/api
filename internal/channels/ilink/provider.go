package ilink

import (
	"context"
	"fmt"
	"time"

	"github.com/wechat-task/api/internal/channels"
	"github.com/wechat-task/api/internal/model"
)

const defaultBotType = 3
const defaultPollInterval = 2 * time.Second
const defaultMaxWait = 5 * time.Minute

// Provider implements channels.Provider for the iLink protocol.
type Provider struct{}

// NewProvider creates a new iLink channel provider.
func NewProvider() *Provider {
	return &Provider{}
}

// Type returns the channel type for iLink.
func (p *Provider) Type() model.ChannelType {
	return model.ChannelTypeWechatClawbot
}

// Connect initiates a WeChat clawbot connection by generating a QR code.
func (p *Provider) Connect(ctx context.Context, params map[string]any) (*channels.ConnectResult, error) {
	client := NewClient("")

	botType := defaultBotType
	if bt, ok := params["bot_type"]; ok {
		if btInt, ok := bt.(int); ok {
			botType = btInt
		}
	}

	qrResp, err := client.GetQRCode(botType)
	if err != nil {
		return nil, fmt.Errorf("get qrcode: %w", err)
	}

	return &channels.ConnectResult{
		Status: "pending",
		Config: model.ChannelConfig{
			"qrcode_id":    qrResp.QRCode,
			"qrcode_image": qrResp.QRCodeImgContent,
		},
		Display: map[string]any{
			"qrcode_image": qrResp.QRCodeImgContent,
		},
	}, nil
}

// Recover resumes a pending iLink connection by polling QR code status.
func (p *Provider) Recover(ctx context.Context, channel *model.Channel) (*channels.ConnectResult, error) {
	if channel.Status != "pending" {
		return nil, nil
	}

	qrcodeID := channel.Config.GetString("qrcode_id")
	if qrcodeID == "" {
		return nil, fmt.Errorf("missing qrcode_id in config")
	}

	client := NewClient("")
	confirmed, err := client.WaitForConfirmation(qrcodeID, defaultPollInterval, defaultMaxWait)
	if err != nil {
		return nil, err
	}

	return &channels.ConnectResult{
		Status: "active",
		Config: model.ChannelConfig{
			"ilink_bot_id":  confirmed.ILinkBotID,
			"ilink_user_id": confirmed.ILinkUserID,
			"bot_token":     confirmed.BotToken,
			"base_url":      confirmed.BaseURL,
		},
	}, nil
}

// SendText sends a text message via the iLink protocol.
func (p *Provider) SendText(ctx context.Context, config model.ChannelConfig, recipient, text string) error {
	botToken := config.GetString("bot_token")
	baseURL := config.GetString("base_url")
	client := NewAuthenticatedClient(botToken, baseURL)
	return client.SendTextMessage(recipient, "", text)
}

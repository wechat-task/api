package lark

import (
	"context"
	"fmt"

	"github.com/wechat-task/api/internal/channels"
	"github.com/wechat-task/api/internal/model"
)

// Provider implements channels.Provider for Lark webhook.
type Provider struct{}

// NewProvider creates a new Lark webhook channel provider.
func NewProvider() *Provider {
	return &Provider{}
}

// Type returns the channel type for Lark.
func (p *Provider) Type() model.ChannelType {
	return model.ChannelTypeLark
}

// Connect validates and stores the Lark webhook configuration.
// Lark webhooks are push-only, so they are immediately active.
func (p *Provider) Connect(ctx context.Context, params map[string]any) (*channels.ConnectResult, error) {
	webhookURL, _ := params["webhook_url"].(string)
	secret, _ := params["secret"].(string)

	if webhookURL == "" {
		return nil, fmt.Errorf("webhook_url is required")
	}

	return &channels.ConnectResult{
		Status: "active",
		Config: model.ChannelConfig{
			"webhook_url": webhookURL,
			"secret":      secret,
		},
	}, nil
}

// Recover is a no-op for Lark since webhooks are stateless.
func (p *Provider) Recover(ctx context.Context, channel *model.Channel) (*channels.ConnectResult, error) {
	return nil, nil
}

// SendText sends a text message via the Lark webhook.
func (p *Provider) SendText(ctx context.Context, config model.ChannelConfig, recipient, text string) error {
	webhookURL := config.GetString("webhook_url")
	secret := config.GetString("secret")
	client := NewClient(webhookURL, secret)
	return client.SendText(text)
}

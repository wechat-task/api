package channels

import (
	"context"

	"github.com/wechat-task/api/internal/model"
)

// ConnectResult holds the result of initiating a channel connection.
type ConnectResult struct {
	Status  string
	Config  model.ChannelConfig
	Display map[string]any
}

// Connector handles the setup flow for a specific channel type.
type Connector interface {
	// Type returns the ChannelType this connector handles.
	Type() model.ChannelType

	// Connect initiates a new channel connection.
	Connect(ctx context.Context, params map[string]any) (*ConnectResult, error)

	// Recover resumes a pending connection flow after server restart.
	Recover(ctx context.Context, channel *model.Channel) (*ConnectResult, error)
}

// Sender handles outbound messages for a specific channel type.
type Sender interface {
	// SendText sends a plain text message.
	SendText(ctx context.Context, config model.ChannelConfig, recipient, text string) error
}

// Provider combines connector and sender for a complete channel protocol.
type Provider interface {
	Connector
	Sender
}

// Registry holds all channel providers keyed by type.
type Registry map[model.ChannelType]Provider

// NewRegistry creates a registry with the given providers.
func NewRegistry(providers ...Provider) Registry {
	r := make(Registry, len(providers))
	for _, p := range providers {
		r[p.Type()] = p
	}
	return r
}

// Get returns the provider for a channel type, or nil if not found.
func (r Registry) Get(t model.ChannelType) Provider {
	return r[t]
}

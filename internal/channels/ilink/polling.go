package ilink

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wechat-task/api/internal/logger"
)

// MessageHandler is called for each inbound message during polling.
type MessageHandler func(msg WeixinMessage) error

// GetUpdates performs a single long-poll request for new messages.
func (c *Client) GetUpdates(cursor string) (*GetUpdatesResponse, error) {
	reqBody := GetUpdatesRequest{
		GetUpdatesBuf: cursor,
		BaseInfo:      BaseInfo{ChannelVersion: ChannelVersion},
	}

	data, err := c.doPOST("/ilink/bot/getupdates", reqBody)
	if err != nil {
		return nil, fmt.Errorf("get updates: %w", err)
	}

	var resp GetUpdatesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal updates response: %w", err)
	}

	if resp.Ret != 0 {
		return nil, fmt.Errorf("get updates failed: ret=%d", resp.Ret)
	}

	return &resp, nil
}

// PollLoop starts a blocking loop that calls GetUpdates and dispatches messages.
// It manages the cursor internally. Returns on error or when ctx is cancelled.
func (c *Client) PollLoop(ctx context.Context, cursor string, handler MessageHandler) error {
	for {
		select {
		case <-ctx.Done():
			logger.Info("iLink polling stopped by context")
			return ctx.Err()
		default:
		}

		resp, err := c.GetUpdates(cursor)
		if err != nil {
			logger.Errorf("iLink get updates error: %v", err)
			return err
		}

		for _, msg := range resp.Msgs {
			if err := handler(msg); err != nil {
				logger.Errorf("iLink message handler error: %v", err)
			}
		}

		if resp.GetUpdatesBuf != "" {
			cursor = resp.GetUpdatesBuf
		}
	}
}

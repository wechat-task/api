package ilink

import (
	"encoding/json"
	"fmt"

	"github.com/wechat-task/api/internal/logger"
)

// SendMessage sends a message with arbitrary items.
func (c *Client) SendMessage(toUserID, contextToken string, items []Item) error {
	reqBody := SendMessageRequest{
		Msg: OutboundMessage{
			ToUserID:     toUserID,
			MessageType:  2, // bot outbound
			MessageState: 2, // FINISH
			ContextToken: contextToken,
			ItemList:     items,
		},
	}

	data, err := c.doPOST("/ilink/bot/sendmessage", reqBody)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	var resp APIResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("unmarshal send response: %w", err)
	}

	if resp.Ret != 0 {
		return fmt.Errorf("send message failed: ret=%d", resp.Ret)
	}

	logger.Debugf("iLink message sent to %s", toUserID)
	return nil
}

// SendTextMessage is a convenience wrapper for sending a text reply.
func (c *Client) SendTextMessage(toUserID, contextToken, text string) error {
	return c.SendMessage(toUserID, contextToken, []Item{
		{Type: 1, TextItem: &TextItem{Text: text}},
	})
}

// SendTyping sends a "typing" indicator to the user.
func (c *Client) SendTyping(toUserID, typingTicket string) error {
	reqBody := map[string]string{
		"to_user_id":    toUserID,
		"typing_ticket": typingTicket,
	}

	data, err := c.doPOST("/ilink/bot/sendtyping", reqBody)
	if err != nil {
		return fmt.Errorf("send typing: %w", err)
	}

	var resp APIResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("unmarshal typing response: %w", err)
	}

	if resp.Ret != 0 {
		return fmt.Errorf("send typing failed: ret=%d", resp.Ret)
	}

	return nil
}

// GetConfig retrieves configuration including typing_ticket.
func (c *Client) GetConfig() (map[string]any, error) {
	data, err := c.doPOST("/ilink/bot/getconfig", nil)
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}

	var resp map[string]any
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal config response: %w", err)
	}

	return resp, nil
}

package ilink

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/wechat-task/api/internal/logger"
)

// GetQRCode fetches a QR code for bot login.
func (c *Client) GetQRCode(botType int) (*QRCodeResponse, error) {
	if botType == 0 {
		botType = 3
	}

	data, err := c.doGET("/ilink/bot/get_bot_qrcode", map[string]string{
		"bot_type": fmt.Sprintf("%d", botType),
	})
	if err != nil {
		return nil, fmt.Errorf("get bot qrcode: %w", err)
	}

	var resp QRCodeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal qrcode response: %w", err)
	}

	if resp.Ret != 0 {
		return nil, fmt.Errorf("get bot qrcode failed: ret=%d", resp.Ret)
	}

	logger.Infof("iLink QR code generated: %s", resp.QRCodeImgContent)
	return &resp, nil
}

// GetQRCodeStatus checks the current status of a QR code.
func (c *Client) GetQRCodeStatus(qrcode string) (*QRCodeStatusResponse, error) {
	data, err := c.doGET("/ilink/bot/get_qrcode_status", map[string]string{
		"qrcode": qrcode,
	})
	if err != nil {
		return nil, fmt.Errorf("get qrcode status: %w", err)
	}

	var resp QRCodeStatusResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal status response: %w", err)
	}

	return &resp, nil
}

// WaitForConfirmation polls QR code status until confirmed or expired.
// Polls every interval until maxWait is reached.
func (c *Client) WaitForConfirmation(qrcode string, interval, maxWait time.Duration) (*QRCodeStatusResponse, error) {
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		resp, err := c.GetQRCodeStatus(qrcode)
		if err != nil {
			return nil, err
		}

		switch resp.Status {
		case "confirmed":
			c.botToken = resp.BotToken
			if resp.BaseURL != "" {
				c.baseURL = resp.BaseURL
			}
			logger.Infof("iLink login confirmed: bot_id=%s, user_id=%s", resp.ILinkBotID, resp.ILinkUserID)
			return resp, nil
		case "expired":
			return nil, fmt.Errorf("QR code expired")
		case "wait":
			time.Sleep(interval)
		default:
			return nil, fmt.Errorf("unknown QR code status: %s", resp.Status)
		}
	}

	return nil, fmt.Errorf("wait for confirmation timed out after %v", maxWait)
}

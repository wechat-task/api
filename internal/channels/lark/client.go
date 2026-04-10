package lark

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/wechat-task/api/internal/logger"
)

// Client is the Lark webhook client.
type Client struct {
	httpClient *http.Client
	webhookURL string
	secret     string
}

// NewClient creates a Lark webhook client with the full webhook URL and optional secret.
func NewClient(webhookURL, secret string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		webhookURL: webhookURL,
		secret:     secret,
	}
}

// sign generates the HMAC-SHA256 signature for request verification.
func (c *Client) sign(timestamp int64) string {
	if c.secret == "" {
		return ""
	}
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, c.secret)
	h := hmac.New(sha256.New, []byte(c.secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Send sends a raw webhook request.
func (c *Client) Send(req *WebhookRequest) error {
	if c.secret != "" {
		timestamp := time.Now().Unix()
		req.TS = strconv.FormatInt(timestamp, 10)
		req.Sign = c.sign(timestamp)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, c.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	logger.Debugf("Lark webhook POST %s", c.webhookURL)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// SendText sends a plain text message via the webhook.
func (c *Client) SendText(text string) error {
	return c.Send(&WebhookRequest{
		MsgType: "text",
		Content: TextContent{Text: text},
	})
}

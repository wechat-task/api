package ilink

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	DefaultBaseURL  = "https://ilinkai.weixin.qq.com"
	ChannelVersion  = "1.0.2"
	LongPollTimeout = 40 * time.Second
)

// Client is the iLink API client.
type Client struct {
	httpClient *http.Client
	baseURL    string
	botToken   string // empty until login completes
}

// NewClient creates an unauthenticated client (for login flow).
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		httpClient: &http.Client{Timeout: LongPollTimeout},
		baseURL:    baseURL,
	}
}

// NewAuthenticatedClient creates a client with an existing token.
func NewAuthenticatedClient(botToken, baseURL string) *Client {
	c := NewClient(baseURL)
	c.botToken = botToken
	return c
}

// SetToken configures the client with a bot token after login.
func (c *Client) SetToken(token string) {
	c.botToken = token
}

// newUIN generates the anti-replay X-WECHAT-UIN header value.
func newUIN() string {
	var buf [4]byte
	_, _ = rand.Read(buf[:])
	return base64.StdEncoding.EncodeToString(
		fmt.Appendf(nil, "%d", binary.LittleEndian.Uint32(buf[:])),
	)
}

// authHeaders returns required headers for authenticated requests.
func (c *Client) authHeaders() http.Header {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	h.Set("AuthorizationType", "ilink_bot_token")
	h.Set("X-WECHAT-UIN", newUIN())
	if c.botToken != "" {
		h.Set("Authorization", "Bearer "+c.botToken)
	}
	return h
}

// doGET performs a GET request with query parameters.
func (c *Client) doGET(path string, params map[string]string) ([]byte, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}
	if len(params) > 0 {
		q := u.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header = c.authHeaders()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// doPOST performs a POST request with JSON body.
func (c *Client) doPOST(path string, payload any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	reqURL := c.baseURL + path
	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header = c.authHeaders()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

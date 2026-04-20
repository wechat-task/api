package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIProvider implements LLMProvider using the OpenAI-compatible Chat Completions API.
type OpenAIProvider struct {
	APIKey  string
	BaseURL string
}

func (p *OpenAIProvider) Complete(ctx context.Context, req LLMRequest) (*LLMResponse, error) {
	baseURL := strings.TrimRight(p.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}

	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	type request struct {
		Model       string    `json:"model"`
		Messages    []message `json:"messages"`
		Temperature float64   `json:"temperature,omitempty"`
		MaxTokens   int       `json:"max_tokens,omitempty"`
	}

	type usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	}

	type response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage usage `json:"usage"`
	}

	msgs := []message{}
	if req.SystemPrompt != "" {
		msgs = append(msgs, message{Role: "system", Content: req.SystemPrompt})
	}
	msgs = append(msgs, message{Role: "user", Content: req.Prompt})

	maxTokens := 0
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}

	body := request{
		Model:       req.Model,
		Messages:    msgs,
		MaxTokens:   maxTokens,
		Temperature: req.Temperature,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)

	httpClient := &http.Client{Timeout: 60 * time.Second}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai api error: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai api returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result response
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	content := ""
	if len(result.Choices) > 0 {
		content = result.Choices[0].Message.Content
	}

	return &LLMResponse{
		Content: content,
		TokenUsage: TokenUsage{
			InputTokens:  result.Usage.PromptTokens,
			OutputTokens: result.Usage.CompletionTokens,
			Total:        result.Usage.TotalTokens,
		},
	}, nil
}

package llm

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// AnthropicProvider implements LLMProvider using the Anthropic Messages API.
type AnthropicProvider struct {
	APIKey  string
	BaseURL string
}

func (p *AnthropicProvider) Complete(ctx context.Context, req LLMRequest) (*LLMResponse, error) {
	opts := []option.RequestOption{
		option.WithAPIKey(p.APIKey),
	}
	if p.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(p.BaseURL))
	}

	client := anthropic.NewClient(opts...)

	maxTokens := int64(4096)
	if req.MaxTokens > 0 {
		maxTokens = int64(req.MaxTokens)
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(req.Model),
		MaxTokens: maxTokens,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(req.Prompt)),
		},
	}

	if req.SystemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: req.SystemPrompt},
		}
	}

	if req.Temperature > 0 {
		params.Temperature = anthropic.Float(req.Temperature)
	}

	resp, err := client.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("anthropic api error: %w", err)
	}

	content := ""
	for _, block := range resp.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	return &LLMResponse{
		Content: content,
		TokenUsage: TokenUsage{
			InputTokens:  int(resp.Usage.InputTokens),
			OutputTokens: int(resp.Usage.OutputTokens),
			Total:        int(resp.Usage.InputTokens + resp.Usage.OutputTokens),
		},
	}, nil
}

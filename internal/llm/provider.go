package llm

import "context"

// LLMRequest is the provider-agnostic request for LLM completion.
type LLMRequest struct {
	Model        string
	Prompt       string
	SystemPrompt string
	Temperature  float64
	MaxTokens    int
}

// LLMResponse is the provider-agnostic response from LLM completion.
type LLMResponse struct {
	Content    string
	TokenUsage TokenUsage
}

// TokenUsage tracks token consumption.
type TokenUsage struct {
	InputTokens  int
	OutputTokens int
	Total        int
}

// LLMProvider is the interface every LLM provider must implement.
type LLMProvider interface {
	Complete(ctx context.Context, req LLMRequest) (*LLMResponse, error)
}

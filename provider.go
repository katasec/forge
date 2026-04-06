package forge

import "context"

// ProviderRequest is the input to a single LLM call.
type ProviderRequest struct {
	Messages     []Message        `json:"messages"`
	Tools        []ToolDefinition `json:"tools,omitempty"`
	SystemPrompt string           `json:"system_prompt,omitempty"`
}

// ProviderResponse is the output of a single LLM call.
type ProviderResponse struct {
	Message      Message      `json:"message"`
	FinishReason FinishReason `json:"finish_reason"`
	Usage        TokenUsage   `json:"usage"`
}

// Provider makes a single LLM call. It does not loop.
type Provider interface {
	Generate(ctx context.Context, req ProviderRequest) (*ProviderResponse, error)
}

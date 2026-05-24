package provider

import (
	"context"

	"github.com/katasec/forge/message"
	"github.com/katasec/forge/tool"
)

// FinishReason indicates why the agent loop terminated.
type FinishReason string

const (
	FinishReasonStop      FinishReason = "stop"
	FinishReasonToolUse   FinishReason = "tool_use"
	FinishReasonIterLimit FinishReason = "iter_limit"
	FinishReasonError     FinishReason = "error"
)

// TokenUsage tracks token consumption across provider calls.
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Request is the input to a single LLM call.
type Request struct {
	Messages     []message.Message `json:"messages"`
	Tools        []tool.Definition `json:"tools,omitempty"`
	SystemPrompt string            `json:"system_prompt,omitempty"`
}

// Response is the output of a single LLM call.
type Response struct {
	Message      message.Message `json:"message"`
	FinishReason FinishReason    `json:"finish_reason"`
	Usage        TokenUsage      `json:"usage"`
}

// Provider makes a single LLM call. It does not loop.
type Provider interface {
	Generate(ctx context.Context, req Request) (*Response, error)
}

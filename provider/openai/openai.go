// Package openai implements forge.Provider using the OpenAI-compatible chat
// completions API. Works with OpenAI, xAI (Grok), Together, Groq, and any
// other provider that speaks the OpenAI format.
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/katasec/forge"
)

// Provider implements forge.Provider using the OpenAI-compatible API.
type Provider struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

// New creates an OpenAI-compatible provider for the given base URL, API key, and model.
func New(baseURL, apiKey, model string) *Provider {
	return &Provider{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		client:  &http.Client{},
	}
}

// --- OpenAI API request/response types ---

type request struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type response struct {
	Choices []choice `json:"choices"`
	Usage   usage    `json:"usage"`
}

type choice struct {
	Message      message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

// Generate sends a request to the OpenAI-compatible chat completions endpoint.
func (p *Provider) Generate(ctx context.Context, req forge.ProviderRequest) (*forge.ProviderResponse, error) {
	// Convert forge messages to OpenAI format.
	var msgs []message
	if req.SystemPrompt != "" {
		msgs = append(msgs, message{Role: "system", Content: req.SystemPrompt})
	}
	for _, m := range req.Messages {
		if m.Role == forge.RoleSystem {
			continue
		}
		msgs = append(msgs, message{
			Role:    string(m.Role),
			Content: m.Content,
		})
	}

	body := request{
		Model:    p.model,
		Messages: msgs,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", p.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (%d): %s", httpResp.StatusCode, string(respBody))
	}

	var apiResp response
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	ch := apiResp.Choices[0]

	finishReason := forge.FinishReasonStop
	if ch.FinishReason == "tool_calls" {
		finishReason = forge.FinishReasonToolUse
	}

	return &forge.ProviderResponse{
		Message: forge.Message{
			Role:    forge.RoleAssistant,
			Content: ch.Message.Content,
		},
		FinishReason: finishReason,
		Usage: forge.TokenUsage{
			InputTokens:  apiResp.Usage.PromptTokens,
			OutputTokens: apiResp.Usage.CompletionTokens,
		},
	}, nil
}

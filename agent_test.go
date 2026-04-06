package forge

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// mockProvider is a test double that returns pre-configured responses.
type mockProvider struct {
	responses []*ProviderResponse
	errors    []error
	calls     int
}

func (m *mockProvider) Generate(_ context.Context, _ ProviderRequest) (*ProviderResponse, error) {
	i := m.calls
	m.calls++
	if i < len(m.errors) && m.errors[i] != nil {
		return nil, m.errors[i]
	}
	if i < len(m.responses) {
		return m.responses[i], nil
	}
	// Default: stop with empty message.
	return &ProviderResponse{
		Message:      Message{Role: RoleAssistant, Content: "default"},
		FinishReason: FinishReasonStop,
	}, nil
}

func TestNewAgentNilProvider(t *testing.T) {
	_, err := NewAgent(Config{})
	if err == nil {
		t.Fatal("expected error for nil provider")
	}
}

func TestNewAgentDefaultErrorPolicy(t *testing.T) {
	agent, err := NewAgent(Config{
		Provider: &mockProvider{},
	})
	if err != nil {
		t.Fatalf("NewAgent error: %v", err)
	}
	if agent.errorPolicy != ErrorPolicyStop {
		t.Errorf("errorPolicy = %q, want %q", agent.errorPolicy, ErrorPolicyStop)
	}
}

func TestAgentRunStop(t *testing.T) {
	provider := &mockProvider{
		responses: []*ProviderResponse{
			{
				Message:      Message{Role: RoleAssistant, Content: "hello back"},
				FinishReason: FinishReasonStop,
				Usage:        TokenUsage{InputTokens: 10, OutputTokens: 5},
			},
		},
	}

	agent, _ := NewAgent(Config{Provider: provider})
	resp, err := agent.Run(context.Background(), AgentRequest{
		Messages: []Message{{Role: RoleUser, Content: "hello"}},
	})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}

	if resp.FinishReason != FinishReasonStop {
		t.Errorf("FinishReason = %q, want %q", resp.FinishReason, FinishReasonStop)
	}
	if resp.Usage.InputTokens != 10 || resp.Usage.OutputTokens != 5 {
		t.Errorf("Usage = %+v", resp.Usage)
	}
	if len(resp.Messages) != 2 { // user + assistant
		t.Errorf("got %d messages, want 2", len(resp.Messages))
	}
	if resp.ConversationID == "" {
		t.Error("expected ConversationID to be generated")
	}
}

func TestAgentRunIterLimit(t *testing.T) {
	provider := &mockProvider{
		responses: []*ProviderResponse{
			{
				Message: Message{
					Role:      RoleAssistant,
					ToolCalls: []ToolCall{{ID: "c1", Name: "echo", Arguments: json.RawMessage(`{"text":"hi"}`)}},
				},
				FinishReason: FinishReasonToolUse,
			},
			// Would loop forever, but iter limit stops it.
			{
				Message: Message{
					Role:      RoleAssistant,
					ToolCalls: []ToolCall{{ID: "c2", Name: "echo", Arguments: json.RawMessage(`{"text":"hi"}`)}},
				},
				FinishReason: FinishReasonToolUse,
			},
		},
	}

	type echoInput struct {
		Text string `json:"text"`
	}

	agent, _ := NewAgent(Config{
		Provider:      provider,
		MaxIterations: 2,
		Tools: []Tool{
			Func[echoInput]("echo", "echoes", func(_ context.Context, in echoInput) (string, error) {
				return in.Text, nil
			}),
		},
	})

	resp, err := agent.Run(context.Background(), AgentRequest{
		Messages: []Message{{Role: RoleUser, Content: "go"}},
	})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.FinishReason != FinishReasonIterLimit {
		t.Errorf("FinishReason = %q, want %q", resp.FinishReason, FinishReasonIterLimit)
	}
}

func TestAgentRunProviderError(t *testing.T) {
	provider := &mockProvider{
		errors: []error{errors.New("provider down")},
	}

	agent, _ := NewAgent(Config{Provider: provider})
	_, err := agent.Run(context.Background(), AgentRequest{
		Messages: []Message{{Role: RoleUser, Content: "hello"}},
	})
	if err == nil {
		t.Fatal("expected error from provider")
	}
	if err.Error() != "provider down" {
		t.Errorf("error = %q, want %q", err.Error(), "provider down")
	}
}

func TestAgentRunToolErrorStop(t *testing.T) {
	provider := &mockProvider{
		responses: []*ProviderResponse{
			{
				Message: Message{
					Role:      RoleAssistant,
					ToolCalls: []ToolCall{{ID: "c1", Name: "broken", Arguments: json.RawMessage(`{}`)}},
				},
				FinishReason: FinishReasonToolUse,
			},
		},
	}

	type emptyInput struct{}

	agent, _ := NewAgent(Config{
		Provider:    provider,
		ErrorPolicy: ErrorPolicyStop,
		Tools: []Tool{
			Func[emptyInput]("broken", "always fails", func(_ context.Context, _ emptyInput) (string, error) {
				return "", errors.New("tool broke")
			}),
		},
	})

	resp, err := agent.Run(context.Background(), AgentRequest{
		Messages: []Message{{Role: RoleUser, Content: "go"}},
	})
	if err != nil {
		t.Fatalf("Run error: %v (tool errors should not be fatal)", err)
	}
	if resp.FinishReason != FinishReasonError {
		t.Errorf("FinishReason = %q, want %q", resp.FinishReason, FinishReasonError)
	}
	if len(resp.Errors) != 1 {
		t.Fatalf("got %d errors, want 1", len(resp.Errors))
	}
	if resp.Errors[0].Message != "tool broke" {
		t.Errorf("error message = %q", resp.Errors[0].Message)
	}
	// Tool results should still be in the message history.
	lastMsg := resp.Messages[len(resp.Messages)-1]
	if lastMsg.Role != RoleTool {
		t.Errorf("last message role = %q, want %q", lastMsg.Role, RoleTool)
	}
}

func TestAgentRunToolErrorContinue(t *testing.T) {
	type emptyInput struct{}

	provider := &mockProvider{
		responses: []*ProviderResponse{
			{
				Message: Message{
					Role:      RoleAssistant,
					ToolCalls: []ToolCall{{ID: "c1", Name: "broken", Arguments: json.RawMessage(`{}`)}},
				},
				FinishReason: FinishReasonToolUse,
			},
			// After seeing the error, LLM stops.
			{
				Message:      Message{Role: RoleAssistant, Content: "I see the tool failed"},
				FinishReason: FinishReasonStop,
			},
		},
	}

	agent, _ := NewAgent(Config{
		Provider:    provider,
		ErrorPolicy: ErrorPolicyContinue,
		Tools: []Tool{
			Func[emptyInput]("broken", "always fails", func(_ context.Context, _ emptyInput) (string, error) {
				return "", errors.New("tool broke")
			}),
		},
	})

	resp, err := agent.Run(context.Background(), AgentRequest{
		Messages: []Message{{Role: RoleUser, Content: "go"}},
	})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.FinishReason != FinishReasonStop {
		t.Errorf("FinishReason = %q, want %q (should continue past error)", resp.FinishReason, FinishReasonStop)
	}
	if len(resp.Errors) != 1 {
		t.Fatalf("got %d errors, want 1 (errors should still be collected)", len(resp.Errors))
	}
	if provider.calls != 2 {
		t.Errorf("provider called %d times, want 2", provider.calls)
	}
}

func TestAgentRunWithMemory(t *testing.T) {
	store := NewInMemoryStore()
	ctx := context.Background()

	// Pre-populate memory.
	store.Save(ctx, "conv-1", []Message{
		{ID: "prev-1", Role: RoleUser, Content: "earlier message"},
	})

	provider := &mockProvider{
		responses: []*ProviderResponse{
			{
				Message:      Message{Role: RoleAssistant, Content: "I remember"},
				FinishReason: FinishReasonStop,
				Usage:        TokenUsage{InputTokens: 20, OutputTokens: 10},
			},
		},
	}

	agent, _ := NewAgent(Config{
		Provider: provider,
		Memory:   store,
	})

	resp, err := agent.Run(ctx, AgentRequest{
		ConversationID: "conv-1",
		Messages:       []Message{{Role: RoleUser, Content: "new message"}},
	})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}

	// Should have: earlier + new + assistant = 3 messages.
	if len(resp.Messages) != 3 {
		t.Fatalf("got %d messages, want 3", len(resp.Messages))
	}
	if resp.Messages[0].Content != "earlier message" {
		t.Errorf("Messages[0] = %q, want %q", resp.Messages[0].Content, "earlier message")
	}

	// Memory should be updated with all 3 messages.
	saved, _ := store.Load(ctx, "conv-1")
	if len(saved) != 3 {
		t.Fatalf("saved %d messages, want 3", len(saved))
	}
}

func TestAgentRunUsageAccumulation(t *testing.T) {
	type emptyInput struct{}

	provider := &mockProvider{
		responses: []*ProviderResponse{
			{
				Message: Message{
					Role:      RoleAssistant,
					ToolCalls: []ToolCall{{ID: "c1", Name: "noop", Arguments: json.RawMessage(`{}`)}},
				},
				FinishReason: FinishReasonToolUse,
				Usage:        TokenUsage{InputTokens: 10, OutputTokens: 5},
			},
			{
				Message:      Message{Role: RoleAssistant, Content: "done"},
				FinishReason: FinishReasonStop,
				Usage:        TokenUsage{InputTokens: 20, OutputTokens: 8},
			},
		},
	}

	agent, _ := NewAgent(Config{
		Provider: provider,
		Tools: []Tool{
			Func[emptyInput]("noop", "does nothing", func(_ context.Context, _ emptyInput) (string, error) {
				return "ok", nil
			}),
		},
	})

	resp, err := agent.Run(context.Background(), AgentRequest{
		Messages: []Message{{Role: RoleUser, Content: "go"}},
	})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.Usage.InputTokens != 30 {
		t.Errorf("InputTokens = %d, want 30", resp.Usage.InputTokens)
	}
	if resp.Usage.OutputTokens != 13 {
		t.Errorf("OutputTokens = %d, want 13", resp.Usage.OutputTokens)
	}
}

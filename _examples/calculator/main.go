// Calculator example — demonstrates forge with math tools and a mock provider.
//
// This uses a mock provider that simulates an LLM deciding to call tools.
// Replace MockProvider with a real provider (Anthropic, OpenAI, etc.) to
// connect to an actual LLM.
//
// Run: go run .
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/katasec/forge"
)

// --- Tool input types ---

type AddInput struct {
	A float64 `json:"a" jsonschema:"description=First number"`
	B float64 `json:"b" jsonschema:"description=Second number"`
}

type MultiplyInput struct {
	A float64 `json:"a" jsonschema:"description=First number"`
	B float64 `json:"b" jsonschema:"description=Second number"`
}

// --- Mock provider ---

// MockProvider simulates an LLM that uses calculator tools.
// On the first call it requests tool use; on the second it returns the final answer.
type MockProvider struct {
	calls int
}

func (p *MockProvider) Generate(_ context.Context, req forge.ProviderRequest) (*forge.ProviderResponse, error) {
	p.calls++

	// First call: "LLM" decides to use the add tool.
	if p.calls == 1 {
		return &forge.ProviderResponse{
			Message: forge.Message{
				Role: forge.RoleAssistant,
				ToolCalls: []forge.ToolCall{
					{
						ID:        "call-1",
						Name:      "add",
						Arguments: json.RawMessage(`{"a": 12, "b": 30}`),
					},
				},
			},
			FinishReason: forge.FinishReasonToolUse,
			Usage:        forge.TokenUsage{InputTokens: 25, OutputTokens: 15},
		}, nil
	}

	// Second call: "LLM" sees the tool result and formulates the answer.
	// Look at the last message to find the tool result.
	var toolResult string
	for _, msg := range req.Messages {
		if msg.Role == forge.RoleTool && len(msg.ToolResults) > 0 {
			toolResult = msg.ToolResults[0].Content
		}
	}

	return &forge.ProviderResponse{
		Message: forge.Message{
			Role:    forge.RoleAssistant,
			Content: fmt.Sprintf("The answer is %s!", toolResult),
		},
		FinishReason: forge.FinishReasonStop,
		Usage:        forge.TokenUsage{InputTokens: 40, OutputTokens: 10},
	}, nil
}

func main() {
	// Create tools.
	addTool := forge.Func[AddInput]("add", "Add two numbers", func(_ context.Context, in AddInput) (string, error) {
		result := in.A + in.B
		b, _ := json.Marshal(result)
		return string(b), nil
	})

	mulTool := forge.Func[MultiplyInput]("multiply", "Multiply two numbers", func(_ context.Context, in MultiplyInput) (string, error) {
		result := in.A * in.B
		b, _ := json.Marshal(result)
		return string(b), nil
	})

	// Create a logging middleware.
	logging := forge.Middleware(func(next forge.RunFunc) forge.RunFunc {
		return func(ctx context.Context, req forge.ProviderRequest) (*forge.ProviderResponse, error) {
			fmt.Printf("[middleware] Calling provider with %d messages\n", len(req.Messages))
			resp, err := next(ctx, req)
			if err == nil {
				fmt.Printf("[middleware] Provider returned: finish_reason=%s\n", resp.FinishReason)
			}
			return resp, err
		}
	})

	// Build the agent.
	agent, err := forge.NewAgent(forge.Config{
		Provider:      &MockProvider{},
		Tools:         []forge.Tool{addTool, mulTool},
		Middleware:     []forge.Middleware{logging},
		SystemPrompt:  "You are a helpful calculator assistant.",
		MaxIterations: 5,
		ErrorPolicy:   forge.ErrorPolicyContinue,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Run the agent.
	fmt.Println("User: What is 12 + 30?")
	fmt.Println(strings.Repeat("-", 40))

	resp, err := agent.Run(context.Background(), forge.AgentRequest{
		Messages: []forge.Message{
			{Role: forge.RoleUser, Content: "What is 12 + 30?"},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("Assistant: %s\n", resp.Messages[len(resp.Messages)-1].Content)
	fmt.Printf("Finish reason: %s\n", resp.FinishReason)
	fmt.Printf("Tokens: %d in, %d out\n", resp.Usage.InputTokens, resp.Usage.OutputTokens)
	fmt.Printf("Conversation: %d messages\n", len(resp.Messages))
}

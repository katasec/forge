package forge

import (
	"context"
	"fmt"
)

// ToolExecutor executes a batch of tool calls and returns the results.
type ToolExecutor interface {
	Execute(ctx context.Context, calls []ToolCall) []ToolResult
}

// SequentialExecutor invokes tools one at a time via a ToolRegistry.
type SequentialExecutor struct {
	Registry *ToolRegistry
}

// Execute processes each tool call in order. Missing tools and invocation
// errors are returned as ToolResults with IsError set to true.
func (e *SequentialExecutor) Execute(ctx context.Context, calls []ToolCall) []ToolResult {
	results := make([]ToolResult, 0, len(calls))
	for _, call := range calls {
		tool, ok := e.Registry.Get(call.Name)
		if !ok {
			results = append(results, ToolResult{
				CallID:  call.ID,
				Content: fmt.Sprintf("tool not found: %s", call.Name),
				IsError: true,
			})
			continue
		}

		content, err := tool.Invoke(ctx, call.Arguments)
		if err != nil {
			results = append(results, ToolResult{
				CallID:  call.ID,
				Content: err.Error(),
				IsError: true,
			})
			continue
		}

		results = append(results, ToolResult{
			CallID:  call.ID,
			Content: content,
			IsError: false,
		})
	}
	return results
}

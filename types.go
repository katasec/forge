package forge

import "encoding/json"

// Role identifies the sender of a message in a conversation.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
	RoleSystem    Role = "system"
)

// Message represents a single message in a conversation.
type Message struct {
	ID          string       `json:"id"`
	Role        Role         `json:"role"`
	Content     string       `json:"content"`
	ToolCalls   []ToolCall   `json:"tool_calls,omitempty"`
	ToolResults []ToolResult `json:"tool_results,omitempty"`
}

// ToolCall represents a request from the LLM to invoke a tool.
type ToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ToolResult represents the outcome of a tool invocation.
type ToolResult struct {
	CallID  string `json:"call_id"`
	Content string `json:"content"`
	IsError bool   `json:"is_error"`
}

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

// ErrorPolicy controls agent behavior when a tool invocation fails.
type ErrorPolicy string

const (
	ErrorPolicyStop     ErrorPolicy = "stop"
	ErrorPolicyContinue ErrorPolicy = "continue"
)

// ToolError wraps a tool invocation failure.
type ToolError struct {
	CallID  string `json:"call_id"`
	Message string `json:"message"`
}

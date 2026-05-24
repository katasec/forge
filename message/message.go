package message

import "github.com/katasec/forge/tool"

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
	ID          string        `json:"id"`
	Role        Role          `json:"role"`
	Content     string        `json:"content"`
	ToolCalls   []tool.Call   `json:"tool_calls,omitempty"`
	ToolResults []tool.Result `json:"tool_results,omitempty"`
}

// UserMessage creates a user-role message with the given content.
func UserMessage(content string) Message {
	return Message{Role: RoleUser, Content: content}
}

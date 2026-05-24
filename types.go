package forge

import (
	"github.com/katasec/forge/message"
	"github.com/katasec/forge/provider"
	"github.com/katasec/forge/tool"
)

type Role = message.Role

const (
	RoleUser      = message.RoleUser
	RoleAssistant = message.RoleAssistant
	RoleTool      = message.RoleTool
	RoleSystem    = message.RoleSystem
)

type Message = message.Message

func UserMessage(content string) Message {
	return message.UserMessage(content)
}

type ToolCall = tool.Call
type ToolResult = tool.Result
type ToolError = tool.Error

type FinishReason = provider.FinishReason

const (
	FinishReasonStop      = provider.FinishReasonStop
	FinishReasonToolUse   = provider.FinishReasonToolUse
	FinishReasonIterLimit = provider.FinishReasonIterLimit
	FinishReasonError     = provider.FinishReasonError
)

type TokenUsage = provider.TokenUsage

type ErrorPolicy string

const (
	ErrorPolicyStop     ErrorPolicy = "stop"
	ErrorPolicyContinue ErrorPolicy = "continue"
)

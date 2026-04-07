# Forge — Type Specification & Design

> This is the authoritative specification for the forge Go agent framework.
> All implementation should conform to this document.

## Package

```
module github.com/katasec/forge
go 1.23
```

Root package: `package forge` — all interfaces, core types, and default implementations.

Sub-packages hold swappable backend implementations (following the `database/sql` pattern):

```
forge/                          root — interfaces, types, Agent, Config, defaults
forge/provider/anthropic/       Anthropic Messages API provider
forge/provider/openai/          OpenAI-compatible provider (OpenAI, xAI, Together, Groq)
```

Future sub-packages (created when implementations exist, not preemptively):

```
forge/memory/sqlite/            SQLite-backed MemoryStore
forge/memory/redis/             Redis-backed MemoryStore
forge/executor/concurrent/      Parallel tool executor
```

---

## 1. Core Types (`types.go`)

### Role

```go
type Role string

const (
    RoleUser      Role = "user"
    RoleAssistant Role = "assistant"
    RoleTool      Role = "tool"
    RoleSystem    Role = "system"
)
```

### Message

```go
type Message struct {
    ID        string       `json:"id"`
    Role      Role         `json:"role"`
    Content   string       `json:"content"`
    ToolCalls []ToolCall   `json:"tool_calls,omitempty"`
    ToolResults []ToolResult `json:"tool_results,omitempty"`
}
```

- `ID` is assigned by the caller or generated (UUID) when empty.
- `ToolCalls` is populated only on assistant messages when the LLM requests tool use.
- `ToolResults` is populated only on tool-role messages returning results.

### ToolCall & ToolResult

```go
type ToolCall struct {
    ID        string          `json:"id"`
    Name      string          `json:"name"`
    Arguments json.RawMessage `json:"arguments"`
}

type ToolResult struct {
    CallID  string `json:"call_id"`
    Content string `json:"content"`
    IsError bool   `json:"is_error"`
}
```

- `ToolCall.ID` is provider-assigned (e.g. Anthropic's `toolu_*` IDs).
- `ToolResult.CallID` must match the corresponding `ToolCall.ID`.
- `ToolResult.IsError` signals that the content is an error message, not a successful result.

### FinishReason

```go
type FinishReason string

const (
    FinishReasonStop      FinishReason = "stop"       // LLM chose to stop (natural completion)
    FinishReasonToolUse   FinishReason = "tool_use"    // LLM requested tool calls
    FinishReasonIterLimit FinishReason = "iter_limit"  // Agent hit MaxIterations
    FinishReasonError     FinishReason = "error"       // Unrecoverable error terminated the loop
)
```

**Semantics:**
- `stop` — the provider returned a response with no tool calls. The agent loop terminates normally.
- `tool_use` — the provider requested tool calls. The loop continues to execute them. This value appears in `ProviderResponse` but never as a final `AgentResponse.FinishReason` (the loop always processes tool calls before returning).
- `iter_limit` — the loop hit `Config.MaxIterations`. The agent returns whatever content the last assistant message contained.
- `error` — a provider error occurred, or a tool error occurred with `ErrorPolicyStop`.

### TokenUsage

```go
type TokenUsage struct {
    InputTokens  int `json:"input_tokens"`
    OutputTokens int `json:"output_tokens"`
}
```

Accumulated across all provider calls in a single `Agent.Run` invocation.

### ErrorPolicy

```go
type ErrorPolicy string

const (
    ErrorPolicyStop     ErrorPolicy = "stop"
    ErrorPolicyContinue ErrorPolicy = "continue"
)
```

Controls behavior when a tool invocation returns an error (`ToolResult.IsError == true`):
- `stop` — the agent loop terminates immediately with `FinishReasonError`.
- `continue` — the error is fed back to the LLM as a tool result so it can recover or try a different approach.

### ToolError

```go
type ToolError struct {
    CallID  string `json:"call_id"`
    Message string `json:"message"`
}
```

Wraps a tool invocation failure. Used internally; surfaced to the caller in `AgentResponse.Errors`.

---

## 2. Metadata (`metadata.go`)

Allows callers to attach arbitrary key-value metadata to a context, accessible by tools and middleware.

```go
type metadataKey struct{}

type Metadata struct {
    Values map[string]string
}

func WithMetadata(ctx context.Context, m Metadata) context.Context
func MetadataFromContext(ctx context.Context) (Metadata, bool)
```

- `WithMetadata` stores a `Metadata` in the context using the unexported `metadataKey`.
- `MetadataFromContext` retrieves it; returns `false` if not present.
- `Metadata.Values` is never nil after construction — `WithMetadata` should initialize the map if nil.

---

## 3. Tool System (`tool.go`, `registry.go`)

### Tool Interface

```go
type Tool interface {
    Name() string
    Description() string
    Schema() ToolSchema
    Invoke(ctx context.Context, args json.RawMessage) (string, error)
}
```

- `Schema()` returns the JSON Schema for the tool's parameters.
- `Invoke()` returns a string result or an error.
- An error from `Invoke` is converted to a `ToolResult{IsError: true}`.

### ToolSchema & ToolDefinition

```go
type ToolSchema struct {
    Parameters json.RawMessage `json:"parameters"`
}

type ToolDefinition struct {
    Name        string     `json:"name"`
    Description string     `json:"description"`
    Schema      ToolSchema `json:"schema"`
}
```

`ToolDefinition` is what gets sent to the provider so it knows which tools are available.

### Func Helper

```go
func Func[T any](name, description string, fn func(ctx context.Context, input T) (string, error)) Tool
```

Generic convenience that:
1. Derives `ToolSchema.Parameters` from `T` using `invopop/jsonschema`.
2. On `Invoke`, unmarshals `args` into `T`, then calls `fn`.
3. Returns a `Tool` implementation (unexported struct is fine).

`T` must be a struct. The JSON schema is generated once at construction time, not per-call.

### ToolRegistry

```go
type ToolRegistry struct { /* unexported fields */ }

func NewToolRegistry() *ToolRegistry
func (r *ToolRegistry) Register(tools ...Tool)
func (r *ToolRegistry) Get(name string) (Tool, bool)
func (r *ToolRegistry) Definitions() []ToolDefinition
```

- `Register` adds tools. Duplicate names overwrite silently (last-write-wins).
- `Get` returns a tool by name.
- `Definitions` returns all registered tools as `ToolDefinition` for passing to a provider.

---

## 4. Executor (`executor.go`)

```go
type ToolExecutor interface {
    Execute(ctx context.Context, calls []ToolCall) []ToolResult
}

type SequentialExecutor struct {
    Registry *ToolRegistry
}

func (e *SequentialExecutor) Execute(ctx context.Context, calls []ToolCall) []ToolResult
```

`SequentialExecutor.Execute`:
1. Iterates `calls` in order.
2. For each call, looks up the tool via `Registry.Get(call.Name)`.
3. If not found: returns `ToolResult{CallID: call.ID, Content: "tool not found: <name>", IsError: true}`.
4. If found: calls `tool.Invoke(ctx, call.Arguments)`.
   - On success: `ToolResult{CallID: call.ID, Content: result, IsError: false}`.
   - On error: `ToolResult{CallID: call.ID, Content: err.Error(), IsError: true}`.
5. Returns all results in the same order as the input calls.

---

## 5. Memory (`memory.go`)

```go
type MemoryStore interface {
    Load(ctx context.Context, conversationID string) ([]Message, error)
    Save(ctx context.Context, conversationID string, messages []Message) error
    Clear(ctx context.Context, conversationID string) error
}

type InMemoryStore struct { /* sync.RWMutex + map[string][]Message */ }

func NewInMemoryStore() *InMemoryStore
```

- `Load` returns a copy of the stored messages (not a reference to the internal slice).
- `Save` replaces the entire message history for that conversation ID.
- `Clear` deletes the conversation.
- `InMemoryStore` is safe for concurrent use.

---

## 6. Provider & Middleware (`provider.go`, `middleware.go`)

### Provider

```go
type ProviderRequest struct {
    Messages    []Message        `json:"messages"`
    Tools       []ToolDefinition `json:"tools,omitempty"`
    SystemPrompt string          `json:"system_prompt,omitempty"`
}

type ProviderResponse struct {
    Message    Message      `json:"message"`
    FinishReason FinishReason `json:"finish_reason"`
    Usage      TokenUsage   `json:"usage"`
}

type Provider interface {
    Generate(ctx context.Context, req ProviderRequest) (*ProviderResponse, error)
}
```

- `Generate` makes a single LLM call. It does **not** loop.
- The provider is responsible for translating `ProviderRequest` into the LLM's native API format and back.
- A provider error (non-nil error return) always terminates the agent loop.

### Middleware

```go
type RunFunc func(ctx context.Context, req ProviderRequest) (*ProviderResponse, error)

type Middleware func(next RunFunc) RunFunc
```

Middleware wraps `RunFunc` in the standard decorator pattern.

**Composition order:** middlewares are applied innermost-last. Given `[A, B, C]`:

```
request → A → B → C → provider.Generate → C → B → A → response
```

Applied as:

```go
run := provider.Generate  // the innermost RunFunc
for i := len(middlewares) - 1; i >= 0; i-- {
    run = middlewares[i](run)
}
```

Use cases: logging, token counting, rate limiting, retry, prompt injection.

---

## 7. Config & Agent (`config.go`, `agent.go`)

### Config

```go
type Config struct {
    Provider      Provider
    Tools         []Tool
    Middleware    []Middleware
    Memory        MemoryStore    // optional, nil means no persistence
    SystemPrompt  string         // optional
    MaxIterations int            // 0 means no limit
    ErrorPolicy   ErrorPolicy    // defaults to ErrorPolicyStop
}
```

### Agent

```go
type Agent struct { /* unexported fields */ }

func NewAgent(cfg Config) (*Agent, error)
```

`NewAgent` validates:
- `Provider` must not be nil.
- Builds a `ToolRegistry` from `cfg.Tools`.
- Creates a `SequentialExecutor` with the registry.
- Applies `cfg.Middleware` to build the composed `RunFunc`.
- Defaults `ErrorPolicy` to `ErrorPolicyStop` if empty.

### AgentRequest & AgentResponse

```go
type AgentRequest struct {
    ConversationID string    `json:"conversation_id"`
    Messages       []Message `json:"messages"`
}

type AgentResponse struct {
    ConversationID string       `json:"conversation_id"`
    Messages       []Message    `json:"messages"`     // full conversation history
    FinishReason   FinishReason `json:"finish_reason"`
    Usage          TokenUsage   `json:"usage"`
    Errors         []ToolError  `json:"errors,omitempty"`
}
```

### Agent Loop — `Agent.Run(ctx, req)`

```
func (a *Agent) Run(ctx context.Context, req AgentRequest) (*AgentResponse, error)
```

Pseudocode:

```
1.  conversationID = req.ConversationID or generate UUID
2.  messages = []
3.  if memory != nil:
4.      messages = memory.Load(ctx, conversationID)
5.  append req.Messages to messages
6.  usage = TokenUsage{}
7.  errors = []ToolError{}
8.  iteration = 0
9.
10. LOOP:
11.     if maxIterations > 0 && iteration >= maxIterations:
12.         finishReason = FinishReasonIterLimit
13.         break LOOP
14.
15.     providerReq = ProviderRequest{
16.         Messages:     messages,
17.         Tools:        registry.Definitions(),
18.         SystemPrompt: config.SystemPrompt,
19.     }
20.
21.     providerResp, err = composedRunFunc(ctx, providerReq)
22.     if err != nil:
23.         return nil, err   // provider errors are always fatal
24.
25.     usage.InputTokens  += providerResp.Usage.InputTokens
26.     usage.OutputTokens += providerResp.Usage.OutputTokens
27.     append providerResp.Message to messages
28.     iteration++
29.
30.     if providerResp.FinishReason == FinishReasonStop:
31.         finishReason = FinishReasonStop
32.         break LOOP
33.
34.     // FinishReason is tool_use — execute the tool calls
35.     toolResults = executor.Execute(ctx, providerResp.Message.ToolCalls)
36.
37.     // Check for tool errors
38.     for each result in toolResults where result.IsError:
39.         append ToolError{CallID: result.CallID, Message: result.Content} to errors
40.         if errorPolicy == ErrorPolicyStop:
41.             finishReason = FinishReasonError
42.             // still append the tool message so the conversation is coherent
43.             toolMsg = Message{Role: RoleTool, ToolResults: toolResults}
44.             append toolMsg to messages
45.             break LOOP
46.
47.     // Feed results back to the LLM
48.     toolMsg = Message{Role: RoleTool, ToolResults: toolResults}
49.     append toolMsg to messages
50.
51. END LOOP
52.
53. if memory != nil:
54.     memory.Save(ctx, conversationID, messages)
55.
56. return AgentResponse{
57.     ConversationID: conversationID,
58.     Messages:       messages,
59.     FinishReason:   finishReason,
60.     Usage:          usage,
61.     Errors:         errors,
62. }
```

**Key behaviors:**
- Provider errors (line 23) are always fatal — they return an error, not an `AgentResponse`.
- Tool errors with `ErrorPolicyContinue` are collected in `errors` but the loop continues, letting the LLM see the error and adapt.
- Tool errors with `ErrorPolicyStop` break the loop immediately but still include the tool results in the message history.
- `FinishReasonToolUse` never appears in the final `AgentResponse` — the loop always processes tool calls.
- Memory is saved once at the end, with the complete conversation.
- Context cancellation is respected: `composedRunFunc` and `tool.Invoke` should check `ctx`.

---

## 8. Future Refinements

Tracked separately in `docs/design/future-refinements.md`. Not in scope for the initial implementation:

- Parallel tool execution (a `ConcurrentExecutor`)
- Streaming provider responses
- Conversation branching / forking
- Persistent memory stores (SQLite, Redis)
- Structured output / response format constraints
- Token budget management (auto-truncate history)
- Agent-to-agent delegation

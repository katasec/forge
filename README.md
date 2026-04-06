# forge

A provider-agnostic Go library for building AI agent loops with pluggable tools, memory, and middleware.

Forge handles the **LLM call → tool execution → response** cycle. You supply a provider (Anthropic, OpenAI, etc.), register tools, and forge runs the loop — including error handling, iteration limits, and conversation memory.

## Install

```bash
go get github.com/katasec/forge
```

## Quick Start

```bash
export ANTHROPIC_API_KEY=sk-ant-...
```

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/katasec/forge"
)

func main() {
    // Create a provider — see _examples/hello-world for full implementation.
    provider := NewAnthropicProvider(os.Getenv("ANTHROPIC_API_KEY"), "claude-sonnet-4-20250514")

    // Build the agent.
    agent, err := forge.NewAgent(forge.Config{
        Provider:     provider,
        SystemPrompt: "You are a helpful assistant. Keep responses brief.",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Run it.
    resp, err := agent.Run(context.Background(), forge.AgentRequest{
        Messages: []forge.Message{
            {Role: forge.RoleUser, Content: "Hello! What are you?"},
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(resp.Messages[len(resp.Messages)-1].Content)
}
```

Swap to xAI Grok by changing one line:

```go
// provider := NewAnthropicProvider(os.Getenv("ANTHROPIC_API_KEY"), "claude-sonnet-4-20250514")
provider := NewOpenAIProvider("https://api.x.ai/v1", os.Getenv("XAI_API_KEY"), "grok-3-mini")
```

The `OpenAIProvider` works with any OpenAI-compatible API (xAI, OpenAI, Together, Groq, etc.). See [`_examples/hello-world`](./_examples/hello-world) for the full runnable code.

## Core Concepts

### Provider

The `Provider` interface makes a single LLM call. Forge doesn't ship providers — you implement one for your LLM of choice:

```go
type Provider interface {
    Generate(ctx context.Context, req ProviderRequest) (*ProviderResponse, error)
}
```

### Tools

Define tools with `Func[T]` — the JSON schema for parameters is derived from the Go struct at construction time using [invopop/jsonschema](https://github.com/invopop/jsonschema):

```go
type SearchInput struct {
    Query string `json:"query" jsonschema:"description=Search query"`
    Limit int    `json:"limit" jsonschema:"description=Max results"`
}

searchTool := forge.Func[SearchInput]("search", "Search the database", func(ctx context.Context, in SearchInput) (string, error) {
    // ... your search logic
    return results, nil
})
```

Or implement the `Tool` interface directly for full control:

```go
type Tool interface {
    Name() string
    Description() string
    Schema() ToolSchema
    Invoke(ctx context.Context, args json.RawMessage) (string, error)
}
```

### Agent Loop

`Agent.Run` executes this loop:

1. Load conversation history from memory (if configured)
2. Call the provider with messages + tool definitions
3. If the provider says **stop** → return the response
4. If the provider requests **tool use** → execute tools, feed results back, go to 2
5. If **iteration limit** hit → return with `FinishReasonIterLimit`
6. Save conversation to memory

### Error Policy

Controls what happens when a tool returns an error:

- `ErrorPolicyStop` (default) — terminate the loop immediately
- `ErrorPolicyContinue` — feed the error back to the LLM so it can adapt

### Middleware

Intercept provider calls for logging, retries, rate limiting, etc:

```go
logging := forge.Middleware(func(next forge.RunFunc) forge.RunFunc {
    return func(ctx context.Context, req forge.ProviderRequest) (*forge.ProviderResponse, error) {
        log.Printf("calling provider with %d messages", len(req.Messages))
        resp, err := next(ctx, req)
        if err == nil {
            log.Printf("provider returned: %s", resp.FinishReason)
        }
        return resp, err
    }
})

agent, _ := forge.NewAgent(forge.Config{
    Provider:   myProvider,
    Middleware: []forge.Middleware{logging},
})
```

Middleware composes as decorators: given `[A, B, C]`, request flows `A → B → C → provider → C → B → A`.

### Memory

Persist conversations across `Agent.Run` calls:

```go
store := forge.NewInMemoryStore()

agent, _ := forge.NewAgent(forge.Config{
    Provider: myProvider,
    Memory:   store,
})

// First call — starts a conversation.
resp, _ := agent.Run(ctx, forge.AgentRequest{
    ConversationID: "conv-1",
    Messages:       []forge.Message{{Role: forge.RoleUser, Content: "Hi"}},
})

// Second call — continues the same conversation.
resp, _ = agent.Run(ctx, forge.AgentRequest{
    ConversationID: "conv-1",
    Messages:       []forge.Message{{Role: forge.RoleUser, Content: "What did I just say?"}},
})
```

Implement `MemoryStore` for persistent storage (SQLite, Redis, etc.):

```go
type MemoryStore interface {
    Load(ctx context.Context, conversationID string) ([]Message, error)
    Save(ctx context.Context, conversationID string, messages []Message) error
    Clear(ctx context.Context, conversationID string) error
}
```

### Metadata

Attach arbitrary key-value data to the context, accessible by tools and middleware:

```go
ctx := forge.WithMetadata(context.Background(), forge.Metadata{
    Values: map[string]string{"user_id": "123", "tenant": "acme"},
})

// Inside a tool:
if meta, ok := forge.MetadataFromContext(ctx); ok {
    userID := meta.Values["user_id"]
}
```

## Examples

See the [`_examples`](./_examples) directory for runnable demos:

- **[hello-world](./_examples/hello-world)** — Simplest possible example: call Claude or xAI with one flag swap
- **[calculator](./_examples/calculator)** — Agent with math tools, middleware, and a mock provider

## License

MIT

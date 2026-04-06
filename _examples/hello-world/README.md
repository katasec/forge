# Hello World

The simplest possible forge example — call an LLM and get a response.

## Run with Claude (Anthropic)

```bash
export ANTHROPIC_API_KEY=sk-ant-...
go run .
```

## Run with Grok (xAI)

```bash
export XAI_API_KEY=xai-...
go run . -provider xai
```

## What's in here

| File | What it does |
|------|-------------|
| `main.go` | Picks a provider, builds an agent, sends "Hello!", prints the response |
| `anthropic.go` | `AnthropicProvider` — implements `forge.Provider` using the Anthropic Messages API |
| `openai_compat.go` | `OpenAIProvider` — implements `forge.Provider` for any OpenAI-compatible API (xAI, OpenAI, Together, Groq, etc.) |

## Swapping providers

The only thing that changes is one line:

```go
// Claude
provider := NewAnthropicProvider(os.Getenv("ANTHROPIC_API_KEY"), "claude-sonnet-4-20250514")

// xAI Grok
provider := NewOpenAIProvider("https://api.x.ai/v1", os.Getenv("XAI_API_KEY"), "grok-3-mini")

// OpenAI
provider := NewOpenAIProvider("https://api.openai.com/v1", os.Getenv("OPENAI_API_KEY"), "gpt-4o")
```

Everything else — agent config, tools, middleware, memory — stays the same regardless of provider.

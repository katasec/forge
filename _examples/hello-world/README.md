# Hello World

The smallest forge example: pick a provider, build an agent with `forge.Config`, ask one question, and print the latest assistant text.

## Run with Claude

```bash
export ANTHROPIC_API_KEY=sk-ant-...
go run .
```

## Run with Grok

```bash
export XAI_API_KEY=xai-...
go run . -provider xai
```

## Run with xAI Search

```bash
export XAI_API_KEY=xai-...
go run . -provider xai-search
```

## What this shows

- `forge.Config` as the agent setup point
- provider swapping without changing app code
- `agent.Ask(ctx, prompt)` for the common path
- `resp.LastText()` for the latest assistant answer
- provider-specific access to xAI citations when search is enabled

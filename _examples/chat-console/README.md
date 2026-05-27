# Chat Console

An interactive console app that shows the intended forge developer experience:

- configure an agent once with `forge.Config`
- plug in an explicit memory implementation with `memory/inmem`
- talk to it with `Ask`
- read the latest answer with `LastText`
- keep conversation context in memory across turns

## Run

```bash
export ANTHROPIC_API_KEY=sk-ant-...
go run .
```

Use OpenAI's Responses API:

```bash
export OPENAI_API_KEY=sk-...
go run . -provider openai
```

Use xAI's Responses API:

```bash
export XAI_API_KEY=xai-...
go run . -provider xai
```

Use xAI Responses API with web search:

```bash
export XAI_API_KEY=xai-...
go run . -provider xai-search
```

Type `exit` or `quit` to stop.

# Future Refinements

> Ideas and extensions tracked for after the initial implementation ships. None of these are in scope for v1.

| Refinement | Notes |
|------------|-------|
| Parallel tool execution | `ConcurrentExecutor` using goroutines with configurable concurrency |
| Streaming provider responses | `Provider.GenerateStream` returning a channel or iterator |
| Conversation branching | Fork a conversation to explore multiple paths |
| Persistent memory stores | SQLite, Redis, or file-backed `MemoryStore` implementations |
| Structured output | Response format constraints passed to providers |
| Token budget management | Auto-truncate or summarize history when approaching limits |
| Agent-to-agent delegation | One agent invoking another as a tool |

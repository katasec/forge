# Calculator

Demonstrates tool use with a mock provider — no API key needed.

## Run

```bash
go run .
```

## Output

```
User: What is 12 + 30?
----------------------------------------
[middleware] Calling provider with 1 messages
[middleware] Provider returned: finish_reason=tool_use
[middleware] Calling provider with 3 messages
[middleware] Provider returned: finish_reason=stop
----------------------------------------
Assistant: The answer is 42!
Finish reason: stop
Tokens: 65 in, 25 out
Conversation: 4 messages
```

## What's in here

Everything is in `main.go`:

- **Tools**: `add` and `multiply` using `forge.Func[T]` with typed inputs
- **Mock provider**: Simulates an LLM that decides to call the `add` tool, then formulates a response from the result
- **Logging middleware**: Prints each provider call and its finish reason
- **Agent loop**: Shows the full cycle — provider call → tool execution → provider call → stop

This example is useful for understanding how the agent loop works without needing API credentials.

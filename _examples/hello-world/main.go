// Hello World — the simplest possible forge example.
//
// Shows how to call Claude with your Anthropic API key, and how to
// swap to xAI's Grok by changing one line.
//
// Usage:
//
//	export ANTHROPIC_API_KEY=sk-ant-...
//	go run .
//
//	# Or use xAI (OpenAI-compatible) instead:
//	export XAI_API_KEY=xai-...
//	go run . -provider xai
//
//	# Or use xAI Responses API with web search:
//	export XAI_API_KEY=xai-...
//	go run . -provider xai-search
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/katasec/forge"
	"github.com/katasec/forge/provider/anthropic"
	"github.com/katasec/forge/provider/openai"
	"github.com/katasec/forge/provider/xai"
)

func main() {
	providerFlag := flag.String("provider", "anthropic", "Provider to use: anthropic, xai, or xai-search")
	flag.Parse()

	// Pick your provider — this is the only thing that changes.
	var provider forge.Provider
	var xaiProvider *xai.Provider // for citation access
	switch *providerFlag {
	case "anthropic":
		key := os.Getenv("ANTHROPIC_API_KEY")
		if key == "" {
			log.Fatal("Set ANTHROPIC_API_KEY environment variable")
		}
		provider = anthropic.New(key, "claude-sonnet-4-20250514")
	case "xai":
		key := os.Getenv("XAI_API_KEY")
		if key == "" {
			log.Fatal("Set XAI_API_KEY environment variable")
		}
		provider = openai.New("https://api.x.ai/v1", key, "grok-3-mini")
	case "xai-search":
		key := os.Getenv("XAI_API_KEY")
		if key == "" {
			log.Fatal("Set XAI_API_KEY environment variable")
		}
		xaiProvider = xai.New(key, "grok-4-1-fast-non-reasoning", xai.WithWebSearch())
		provider = xaiProvider
	default:
		log.Fatalf("Unknown provider: %s (use 'anthropic', 'xai', or 'xai-search')", *providerFlag)
	}

	// Build the agent — same code regardless of provider.
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
	fmt.Printf("\n[%s | tokens: %d in, %d out]\n", *providerFlag, resp.Usage.InputTokens, resp.Usage.OutputTokens)

	// Show citations if using xai-search.
	if xaiProvider != nil {
		if citations := xaiProvider.LastCitations(); len(citations) > 0 {
			fmt.Println("\nSources:")
			for i, c := range citations {
				fmt.Printf("  [%d] %s — %s\n", i+1, c.Title, c.URL)
			}
		}
	}
}

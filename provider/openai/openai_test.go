package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/katasec/forge"
)

// Compile-time check that *Provider satisfies forge.Provider.
var _ forge.Provider = (*Provider)(nil)

func TestNew(t *testing.T) {
	p := New("https://api.openai.com/v1", "test-key", "gpt-4")
	if p == nil {
		t.Fatal("New returned nil")
	}
	if p.baseURL != "https://api.openai.com/v1" {
		t.Errorf("baseURL = %q, want %q", p.baseURL, "https://api.openai.com/v1")
	}
	if p.apiKey != "test-key" {
		t.Errorf("apiKey = %q, want %q", p.apiKey, "test-key")
	}
	if p.model != "gpt-4" {
		t.Errorf("model = %q, want %q", p.model, "gpt-4")
	}
}

func TestGenerate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer test-key")
		}

		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Model != "gpt-4" {
			t.Errorf("model = %q, want %q", req.Model, "gpt-4")
		}
		// System prompt should be the first message.
		if len(req.Messages) == 0 || req.Messages[0].Role != "system" {
			t.Error("expected system message as first message")
		}

		resp := response{
			Choices: []choice{{
				Message:      message{Role: "assistant", Content: "Hello!"},
				FinishReason: "stop",
			}},
			Usage: usage{PromptTokens: 8, CompletionTokens: 3},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p := New(srv.URL, "test-key", "gpt-4")

	resp, err := p.Generate(context.Background(), forge.ProviderRequest{
		SystemPrompt: "You are helpful.",
		Messages: []forge.Message{
			{Role: forge.RoleUser, Content: "Hi"},
		},
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if resp.Message.Content != "Hello!" {
		t.Errorf("content = %q, want %q", resp.Message.Content, "Hello!")
	}
	if resp.Message.Role != forge.RoleAssistant {
		t.Errorf("role = %q, want %q", resp.Message.Role, forge.RoleAssistant)
	}
	if resp.FinishReason != forge.FinishReasonStop {
		t.Errorf("finishReason = %q, want %q", resp.FinishReason, forge.FinishReasonStop)
	}
	if resp.Usage.InputTokens != 8 || resp.Usage.OutputTokens != 3 {
		t.Errorf("usage = %+v, want {8, 3}", resp.Usage)
	}
}

func TestGenerateNoChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := response{Choices: []choice{}, Usage: usage{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p := New(srv.URL, "test-key", "gpt-4")

	_, err := p.Generate(context.Background(), forge.ProviderRequest{
		Messages: []forge.Message{{Role: forge.RoleUser, Content: "Hi"}},
	})
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
}

func TestGenerateAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":"rate limited"}`))
	}))
	defer srv.Close()

	p := New(srv.URL, "test-key", "gpt-4")

	_, err := p.Generate(context.Background(), forge.ProviderRequest{
		Messages: []forge.Message{{Role: forge.RoleUser, Content: "Hi"}},
	})
	if err == nil {
		t.Fatal("expected error for 429 response")
	}
}

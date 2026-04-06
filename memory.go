package forge

import (
	"context"
	"sync"
)

// MemoryStore persists conversation message history.
type MemoryStore interface {
	Load(ctx context.Context, conversationID string) ([]Message, error)
	Save(ctx context.Context, conversationID string, messages []Message) error
	Clear(ctx context.Context, conversationID string) error
}

// InMemoryStore is a thread-safe in-memory implementation of MemoryStore.
type InMemoryStore struct {
	mu   sync.RWMutex
	data map[string][]Message
}

// NewInMemoryStore creates an empty InMemoryStore.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string][]Message),
	}
}

// Load returns a copy of the stored messages for the given conversation.
func (s *InMemoryStore) Load(_ context.Context, conversationID string) ([]Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msgs, ok := s.data[conversationID]
	if !ok {
		return nil, nil
	}

	cp := make([]Message, len(msgs))
	copy(cp, msgs)
	return cp, nil
}

// Save replaces the entire message history for the given conversation.
func (s *InMemoryStore) Save(_ context.Context, conversationID string, messages []Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cp := make([]Message, len(messages))
	copy(cp, messages)
	s.data[conversationID] = cp
	return nil
}

// Clear deletes the conversation history.
func (s *InMemoryStore) Clear(_ context.Context, conversationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, conversationID)
	return nil
}

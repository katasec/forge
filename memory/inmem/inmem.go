package inmem

import (
	"context"
	"sync"

	"github.com/katasec/forge/message"
)

// Store is a thread-safe in-memory memory store.
type Store struct {
	mu   sync.RWMutex
	data map[string][]message.Message
}

// New creates an empty in-memory store.
func New() *Store {
	return &Store{
		data: make(map[string][]message.Message),
	}
}

// Load returns a copy of the stored messages for the given conversation.
func (s *Store) Load(_ context.Context, conversationID string) ([]message.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	msgs, ok := s.data[conversationID]
	if !ok {
		return nil, nil
	}

	cp := make([]message.Message, len(msgs))
	copy(cp, msgs)
	return cp, nil
}

// Save replaces the entire message history for the given conversation.
func (s *Store) Save(_ context.Context, conversationID string, messages []message.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cp := make([]message.Message, len(messages))
	copy(cp, messages)
	s.data[conversationID] = cp
	return nil
}

// Clear deletes the conversation history.
func (s *Store) Clear(_ context.Context, conversationID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, conversationID)
	return nil
}

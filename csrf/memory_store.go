package csrf

import (
	"sync"
)

// MemoryCSRFStore is a simple in-memory implementation of CSRFStore
type MemoryCSRFStore struct {
	tokens sync.Map // thread-safe map to store tokens
}

// NewMemoryCSRFStore creates a new MemoryCSRFStore
func NewMemoryCSRFStore() *MemoryCSRFStore {
	return &MemoryCSRFStore{
		tokens: sync.Map{},
	}
}

// Store saves a token with the given key
func (s *MemoryCSRFStore) Store(key, token string) error {
	if key == "" || token == "" {
		return ErrKeyOrTokenEmpty
	}
	s.tokens.Store(key, token)
	return nil
}

// Validate checks if the provided token matches the one stored under the given key
func (s *MemoryCSRFStore) Validate(key, token string) error {
	if key == "" || token == "" {
		return ErrKeyOrTokenEmpty
	}
	value, ok := s.tokens.Load(key)
	if !ok {
		return ErrTokenNotFound
	}
	storedToken, ok := value.(string)
	if !ok {
		return ErrTokenNotFound
	}
	if storedToken != token {
		return ErrTokenMismatch
	}
	return nil
}

package csrf

import (
	"crypto/subtle"
	"sync"
	"time"
)

// TokenEntry stores a token with its expiration time
type TokenEntry struct {
	Token      string
	Expiration time.Time
}

// DefaultMemoryCSRFStore includes token expiration and cleanup
type DefaultMemoryCSRFStore struct {
	tokens   sync.Map // map[string]TokenEntry
	lifetime time.Duration
	cleanup  *time.Ticker
}

// NewDefaultMemoryCSRFStore creates a store with periodic cleanup
func NewDefaultMemoryCSRFStore(durations ...time.Duration) *DefaultMemoryCSRFStore {
	var (
		lt = DefaultExpirationTime
		ct = DefaultCleanupIntervalTime
	)
	if len(durations) > 0 {
		lt = durations[0]
	}
	if len(durations) > 1 {
		ct = durations[1]
	}

	store := &DefaultMemoryCSRFStore{
		tokens:   sync.Map{},
		lifetime: lt,
		cleanup:  time.NewTicker(ct),
	}

	// Start cleanup routine
	go store.cleanupRoutine()

	return store
}

// Store saves a token with expiration
func (s *DefaultMemoryCSRFStore) Store(key, token string) error {
	if key == "" || token == "" {
		return ErrKeyOrTokenEmpty
	}

	s.tokens.Store(key, TokenEntry{
		Token:      token,
		Expiration: time.Now().Add(s.lifetime),
	})
	return nil
}

func (s *DefaultMemoryCSRFStore) Validate(key, token string) error {
	_, err := s.validEntry(key, token)
	return err
}

// Consume validates and removes a token in one atomic operation.
func (s *DefaultMemoryCSRFStore) Consume(key, token string) error {
	entry, err := s.validEntry(key, token)
	if err != nil {
		return err
	}
	if !s.tokens.CompareAndDelete(key, entry) {
		return ErrTokenNotFound
	}
	return nil
}

func (s *DefaultMemoryCSRFStore) validEntry(key, token string) (TokenEntry, error) {
	if key == "" || token == "" {
		return TokenEntry{}, ErrKeyOrTokenEmpty
	}

	value, ok := s.tokens.Load(key)
	if !ok {
		return TokenEntry{}, ErrTokenNotFound
	}

	entry, ok := value.(TokenEntry)
	if !ok {
		return TokenEntry{}, ErrTokenNotFound
	}

	// Check if the token is expired
	if time.Now().After(entry.Expiration) {
		s.tokens.CompareAndDelete(key, entry)
		return TokenEntry{}, ErrTokenExpired
	}

	if subtle.ConstantTimeCompare([]byte(entry.Token), []byte(token)) != 1 {
		return TokenEntry{}, ErrTokenMismatch
	}

	return entry, nil
}

// cleanupRoutine periodically removes expired tokens
func (s *DefaultMemoryCSRFStore) cleanupRoutine() {
	for range s.cleanup.C {
		now := time.Now()
		s.tokens.Range(func(key, value interface{}) bool {
			entry, ok := value.(TokenEntry)
			if !ok || now.After(entry.Expiration) {
				s.tokens.CompareAndDelete(key, value)
			}
			return true
		})
	}
}

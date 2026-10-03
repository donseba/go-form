package csrf

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestConsume(t *testing.T) {
	stores := map[string]func() Store{
		"memory":  func() Store { return NewMemoryCSRFStore() },
		"default": func() Store { return NewDefaultMemoryCSRFStore() },
	}
	for name, newStore := range stores {
		t.Run(name, func(t *testing.T) {
			store := newStore()
			consumer := store.(TokenConsumer)
			if err := store.Store("session-token", "token"); err != nil {
				t.Fatal(err)
			}
			if err := consumer.Consume("session-token", "forged"); !errors.Is(err, ErrTokenMismatch) {
				t.Fatalf("forged token: %v", err)
			}
			var accepted atomic.Int32
			var workers sync.WaitGroup
			for range 32 {
				workers.Add(1)
				go func() {
					defer workers.Done()
					if err := consumer.Consume("session-token", "token"); err == nil {
						accepted.Add(1)
					} else if !errors.Is(err, ErrTokenNotFound) {
						t.Errorf("unexpected consumption error: %v", err)
					}
				}()
			}
			workers.Wait()
			if accepted.Load() != 1 {
				t.Fatalf("token consumed %d times", accepted.Load())
			}
			if err := store.Validate("session-token", "token"); !errors.Is(err, ErrTokenNotFound) {
				t.Fatalf("consumed token still stored: %v", err)
			}
		})
	}
}

func TestConsumeExpiredToken(t *testing.T) {
	store := NewDefaultMemoryCSRFStore(-time.Second, time.Hour)
	if err := store.Store("key", "token"); err != nil {
		t.Fatal(err)
	}
	if err := store.Consume("key", "token"); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("expired token: %v", err)
	}
}

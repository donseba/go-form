package csrf

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"
)

// Context key for storing CSRF token and session ID
type contextKey string

const (
	CSRFTokenContextKey contextKey = "csrf_token"
	SessionIDContextKey contextKey = "session_id"
)

var (
	ErrTokenMismatch           = errors.New("csrf token mismatch")
	ErrTokenNotFound           = errors.New("csrf token not found")
	ErrTokenExpired            = errors.New("csrf token expired")
	ErrKeyOrTokenEmpty         = errors.New("key or token must not be empty")
	DefaultSessionID           = "session_id"
	DefaultExpirationTime      = 10 * time.Minute
	DefaultCleanupIntervalTime = 10 * time.Minute
)

// Store defines an interface for storing and retrieving CSRF tokens.
// Implementations can use any session storage mechanism.
type Store interface {
	// Store saves a token with the given key
	Store(key, token string) error

	// Validate checks if the token is valid for the key
	Validate(key, token string) error
}

// GenerateCSRFToken creates a secure random token for CSRF protection
func GenerateCSRFToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

package form

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/donseba/go-form/csrf"
)

// CSRF errors
var (
	DefaultCSRFField = "_csrf"
)

// CSRFOptions configures how the CSRF middleware behaves
type CSRFOptions struct {
	// ErrorHandler lets you customize error handling instead of returning HTTP errors
	ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)
}

// DefaultCSRFOptions returns the default options for CSRF protection
func DefaultCSRFOptions() CSRFOptions {
	return CSRFOptions{
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			switch {
			case errors.Is(err, csrf.ErrTokenMismatch):
				http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			case errors.Is(err, csrf.ErrTokenExpired):
				http.Error(w, "CSRF token expired", http.StatusForbidden)
			case errors.Is(err, csrf.ErrKeyOrTokenEmpty):
				http.Error(w, "CSRF token or session ID is empty", http.StatusBadRequest)
			case errors.Is(err, csrf.ErrTokenNotFound):
				http.Error(w, "CSRF token not found", http.StatusBadRequest)
			default:
				http.Error(w, "CSRF validation error: "+err.Error(), http.StatusBadRequest)
			}
		},
	}
}

// CSRFMiddleware creates middleware for CSRF protection with default options
func (f *Form) CSRFMiddleware() func(next http.Handler) http.Handler {
	return f.CSRFMiddlewareWithOptions(DefaultCSRFOptions())
}

// CSRFMiddlewareWithOptions creates middleware for CSRF protection with custom options
func (f *Form) CSRFMiddlewareWithOptions(options CSRFOptions) func(next http.Handler) http.Handler {
	if !f.HasCSRFStore() {
		f.SetCSRFStore(csrf.NewMemoryCSRFStore())
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get session key (from cookie or create one)
			sessionID, err := getOrCreateSessionID(w, r)
			if err != nil {
				if options.ErrorHandler != nil {
					options.ErrorHandler(w, r, err)
					return
				}
				http.Error(w, "Session error", http.StatusInternalServerError)
				return
			}

			// Store the session ID in the context
			ctx := context.WithValue(r.Context(), csrf.SessionIDContextKey, sessionID)
			r = r.WithContext(ctx)

			// For safe methods (GET, HEAD), generate and store a token
			if r.Method == http.MethodGet || r.Method == http.MethodHead {
				token, err := csrf.GenerateCSRFToken()
				if err != nil {
					if options.ErrorHandler != nil {
						options.ErrorHandler(w, r, err)
						return
					}
					http.Error(w, "Failed to generate CSRF token", http.StatusInternalServerError)
					return
				}

				// Store the token
				err = f.GetCSRFStore().Store(sessionID, token)
				if err != nil {
					if options.ErrorHandler != nil {
						options.ErrorHandler(w, r, err)
						return
					}
					http.Error(w, "Failed to store CSRF token", http.StatusInternalServerError)
					return
				}

				// Add token to context
				ctx = context.WithValue(r.Context(), csrf.CSRFTokenContextKey, token)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// For unsafe methods, validate the token
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete || r.Method == http.MethodPatch {
				submittedToken := r.FormValue(DefaultCSRFField)
				if submittedToken == "" {
					if options.ErrorHandler != nil {
						options.ErrorHandler(w, r, errors.New("token not found"))
						return
					}
					http.Error(w, "Missing CSRF token", http.StatusBadRequest)
					return
				}

				// Validate the token
				err := f.GetCSRFStore().Validate(sessionID, submittedToken)
				if err != nil {
					if options.ErrorHandler != nil {
						options.ErrorHandler(w, r, err)
						return
					}
					if errors.Is(err, csrf.ErrTokenMismatch) {
						http.Error(w, "Invalid CSRF token", http.StatusForbidden)
					} else {
						http.Error(w, "CSRF validation error", http.StatusInternalServerError)
					}
					return
				}

				// Generate a fresh token for the next request
				token, err := csrf.GenerateCSRFToken()
				if err != nil {
					if options.ErrorHandler != nil {
						options.ErrorHandler(w, r, err)
						return
					}
					http.Error(w, "Failed to generate CSRF token", http.StatusInternalServerError)
					return
				}

				// Store the token
				err = f.GetCSRFStore().Store(sessionID, token)
				if err != nil {
					if options.ErrorHandler != nil {
						options.ErrorHandler(w, r, err)
						return
					}
					http.Error(w, "Failed to store CSRF token", http.StatusInternalServerError)
					return
				}

				// Add token to context
				ctx = context.WithValue(r.Context(), csrf.CSRFTokenContextKey, token)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// For other methods, just pass through
			next.ServeHTTP(w, r)
		})
	}
}

// Helper function to get or create a session ID
func getOrCreateSessionID(w http.ResponseWriter, r *http.Request) (string, error) {
	// Check for existing session cookie
	cookie, err := r.Cookie(csrf.DefaultSessionID)
	if err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	// Create a new session ID
	sessionID, err := csrf.GenerateCSRFToken()
	if err != nil {
		return "", err
	}

	// Set the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     csrf.DefaultSessionID,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil, // Set secure if using HTTPS
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(csrf.DefaultExpirationTime),
	})

	return sessionID, nil
}

// getSessionID retrieves the session ID from the request context or cookie
func getSessionID(r *http.Request) (string, error) {
	// First check if sessionID is in context
	if sessionID, ok := r.Context().Value(csrf.SessionIDContextKey).(string); ok && sessionID != "" {
		return sessionID, nil
	}

	// Then check for cookie
	cookie, err := r.Cookie(csrf.DefaultSessionID)
	if err != nil || cookie.Value == "" {
		return "", errors.New("session ID not found")
	}

	return cookie.Value, nil
}

// GetCSRFToken retrieves the CSRF token from the request context
func GetCSRFToken(r *http.Request) (string, bool) {
	token, ok := r.Context().Value(csrf.CSRFTokenContextKey).(string)
	return token, ok
}

// InjectCSRFToken adds the CSRF token to the form Info struct
func InjectCSRFToken(r *http.Request, info *Info) {
	if token, ok := GetCSRFToken(r); ok && token != "" {
		info.CsrfValue = token
		if info.CsrfField == "" {
			info.CsrfField = DefaultCSRFField
		}
	}
}

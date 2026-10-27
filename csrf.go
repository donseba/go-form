package form

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/donseba/go-form/v2/csrf"
)

// CSRF errors
var (
	DefaultCSRFField = "_csrf"
)

// CSRFOptions configures how the CSRF middleware behaves
type CSRFOptions struct {
	// ErrorHandler lets you customize error handling instead of returning HTTP errors
	ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

	// FieldName is the form field the middleware reads the submitted token
	// from; empty means DefaultCSRFField. Forms that WithRequestInfo or
	// WithContextInfo give the middleware's token render it in this field,
	// unless their Info.CsrfField names another one.
	FieldName string
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
				http.Error(w, "CSRF token not found", http.StatusForbidden)
			default:
				http.Error(w, "CSRF validation error: "+err.Error(), http.StatusBadRequest)
			}
		},
	}
}

// fieldName is the form field the token is submitted in.
func (o CSRFOptions) fieldName() string {
	if o.FieldName != "" {
		return o.FieldName
	}

	return DefaultCSRFField
}

// CSRFMiddleware creates middleware for CSRF protection with default options
func (f *Form) CSRFMiddleware() func(next http.Handler) http.Handler {
	return f.CSRFMiddlewareWithOptions(DefaultCSRFOptions())
}

// CSRFMiddlewareWithOptions creates middleware for CSRF protection with
// custom options. Safe requests get a session, and GET and HEAD requests a
// token for the forms they render. A submission (POST, PUT, PATCH or
// DELETE) must come with the session and a token issued to it, which it
// uses up, and gets a fresh token. Every request it lets through sets or
// renews the session cookie, for as long as the store keeps tokens (see
// csrf.ExpiringStore).
func (f *Form) CSRFMiddlewareWithOptions(options CSRFOptions) func(next http.Handler) http.Handler {
	if !f.HasCSRFStore() {
		f.SetCSRFStore(csrf.NewDefaultMemoryCSRFStore())
	}

	return func(next http.Handler) http.Handler {
		return csrfHandler{form: f, options: options, next: next}
	}
}

// csrfHandler is the CSRF middleware of form around next.
type csrfHandler struct {
	form    *Form
	options CSRFOptions
	next    http.Handler
}

// csrfFieldContextKey is the context key of the form field the middleware
// reads the token from (CSRFOptions.FieldName), for InjectCSRFTokenContext.
type csrfFieldContextKey struct{}

// ServeHTTP runs the protected handler for r when r passes the checks of
// the middleware.
func (h csrfHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if unsafeMethod(r.Method) {
		h.serveSubmission(w, r)

		return
	}

	// Safe requests may create a session.
	sessionID, err := sessionIDOrNew(r)
	if err != nil {
		h.fail(w, r, err, "Session error", http.StatusInternalServerError)

		return
	}

	r = h.withSession(r, sessionID)

	// GET and HEAD render forms, which need a token.
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		h.serveWithToken(w, r, sessionID)

		return
	}

	// For other methods, just pass through
	h.setSessionCookie(w, r, sessionID)
	h.next.ServeHTTP(w, r)
}

// serveSubmission runs the handler for r, a submission, when it carries a
// token issued to its session. The token is used up; the handler gets a
// fresh one, to render the form again.
func (h csrfHandler) serveSubmission(w http.ResponseWriter, r *http.Request) {
	// Submissions must use an existing session.
	sessionID, err := getSessionID(r)
	if err != nil {
		h.fail(w, r, err, "Session error", http.StatusInternalServerError)

		return
	}

	r = h.withSession(r, sessionID)

	submittedToken := r.FormValue(h.options.fieldName())
	if submittedToken == "" {
		h.fail(w, r, csrf.ErrKeyOrTokenEmpty, "Missing CSRF token", http.StatusBadRequest)

		return
	}

	// Consume only the submitted token; other forms stay valid.
	err = consumeCSRFToken(h.form.GetCSRFStore(), csrfStoreKey(sessionID, submittedToken), submittedToken)
	if err != nil {
		message, status := "CSRF validation error", http.StatusInternalServerError
		if invalidToken(err) {
			message, status = "Invalid CSRF token", http.StatusForbidden
		}

		h.fail(w, r, err, message, status)

		return
	}

	h.serveWithToken(w, r, sessionID)
}

// invalidToken reports whether err, an error of the store, means that the
// submitted token is not valid for the session: unknown, another or expired.
func invalidToken(err error) bool {
	return errors.Is(err, csrf.ErrTokenMismatch) || errors.Is(err, csrf.ErrTokenNotFound) || errors.Is(err, csrf.ErrTokenExpired)
}

// serveWithToken stores a fresh token for the session, renews the session
// cookie and runs the handler with the token in the request's context.
func (h csrfHandler) serveWithToken(w http.ResponseWriter, r *http.Request, sessionID string) {
	token, err := csrf.GenerateCSRFToken()
	if err != nil {
		h.fail(w, r, err, "Failed to generate CSRF token", http.StatusInternalServerError)

		return
	}

	err = h.form.GetCSRFStore().Store(csrfStoreKey(sessionID, token), token)
	if err != nil {
		h.fail(w, r, err, "Failed to store CSRF token", http.StatusInternalServerError)

		return
	}

	h.setSessionCookie(w, r, sessionID)

	ctx := context.WithValue(r.Context(), csrf.CSRFTokenContextKey, token)
	h.next.ServeHTTP(w, r.WithContext(ctx))
}

// withSession returns r with the session ID and the token's form field in
// its context.
func (h csrfHandler) withSession(r *http.Request, sessionID string) *http.Request {
	ctx := context.WithValue(r.Context(), csrf.SessionIDContextKey, sessionID)
	ctx = context.WithValue(ctx, csrfFieldContextKey{}, h.options.fieldName())

	return r.WithContext(ctx)
}

// setSessionCookie sets the session cookie, or renews it, so that it lasts
// as long as a token issued now (see Form.sessionLifetime).
func (h csrfHandler) setSessionCookie(w http.ResponseWriter, r *http.Request, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     csrf.DefaultSessionID,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil, // Set secure if using HTTPS
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(h.form.sessionLifetime()),
	})
}

// fail answers a request the middleware refuses, with the ErrorHandler of
// the options when there is one, and with message and status otherwise.
func (h csrfHandler) fail(w http.ResponseWriter, r *http.Request, err error, message string, status int) {
	if h.options.ErrorHandler != nil {
		h.options.ErrorHandler(w, r, err)

		return
	}

	http.Error(w, message, status)
}

// sessionLifetime is how long the session cookie lasts after a token is
// issued: the lifetime of the store's tokens when the store reports one
// (csrf.ExpiringStore), and csrf.DefaultExpirationTime otherwise.
func (f *Form) sessionLifetime() time.Duration {
	if store, ok := f.GetCSRFStore().(csrf.ExpiringStore); ok && store.Lifetime() > 0 {
		return store.Lifetime()
	}

	return csrf.DefaultExpirationTime
}

// unsafeMethod reports whether a request with method submits something, so
// that it must carry a token: POST, PUT, PATCH or DELETE.
func unsafeMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	}

	return false
}

// csrfStoreKey keeps each issued token independent and bound to its session.
// Hashing keeps session IDs and tokens out of storage keys and diagnostics.
func csrfStoreKey(sessionID, token string) string {
	sum := sha256.Sum256([]byte(strconv.Itoa(len(sessionID)) + ":" + sessionID + token))
	return hex.EncodeToString(sum[:])
}

func consumeCSRFToken(store csrf.Store, key, token string) error {
	if consumer, ok := store.(csrf.TokenConsumer); ok {
		return consumer.Consume(key, token)
	}
	if err := store.Validate(key, token); err != nil {
		return err
	}
	// Preserve one-use behavior for existing Store implementations. An atomic
	// TokenConsumer is needed to also reject simultaneous reuse across instances.
	replacement, err := csrf.GenerateCSRFToken()
	if err != nil {
		return err
	}
	return store.Store(key, replacement)
}

// sessionIDOrNew returns the session ID of the session cookie, or a new
// one when the request has none.
func sessionIDOrNew(r *http.Request) (string, error) {
	cookie, err := r.Cookie(csrf.DefaultSessionID)
	if err == nil && cookie.Value != "" {
		return cookie.Value, nil
	}

	return csrf.GenerateCSRFToken()
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
	if r == nil {
		return
	}
	InjectCSRFTokenContext(r.Context(), info)
}

// InjectCSRFTokenContext adds the CSRF token from ctx to the form Info struct.
// It is useful when rendering happens below the HTTP controller layer. An
// empty Info.CsrfField becomes the field the middleware reads the token
// from (CSRFOptions.FieldName, or DefaultCSRFField).
func InjectCSRFTokenContext(ctx context.Context, info *Info) {
	if ctx == nil || info == nil {
		return
	}

	token, ok := ctx.Value(csrf.CSRFTokenContextKey).(string)
	if !ok || token == "" {
		return
	}

	info.CsrfValue = token

	if info.CsrfField == "" {
		info.CsrfField = csrfFieldName(ctx)
	}
}

// csrfFieldName is the form field the CSRF middleware that handles the
// request of ctx reads the token from: CSRFOptions.FieldName, or
// DefaultCSRFField.
func csrfFieldName(ctx context.Context) string {
	if name, ok := ctx.Value(csrfFieldContextKey{}).(string); ok && name != "" {
		return name
	}

	return DefaultCSRFField
}

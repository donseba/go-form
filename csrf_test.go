package form

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/donseba/go-form/csrf"
	"github.com/donseba/go-form/templates"
)

// TestGenerateCSRFToken verifies that token generation creates valid tokens
func TestGenerateCSRFToken(t *testing.T) {
	token, err := csrf.GenerateCSRFToken()
	if err != nil {
		t.Fatalf("GenerateCSRFToken() error = %v", err)
	}
	if token == "" {
		t.Error("GenerateCSRFToken() returned empty token")
	}

	// Generate a few tokens and ensure they're different
	tokens := make(map[string]bool)
	for i := 0; i < 10; i++ {
		token, err := csrf.GenerateCSRFToken()
		if err != nil {
			t.Fatalf("GenerateCSRFToken() error = %v", err)
		}
		if tokens[token] {
			t.Errorf("GenerateCSRFToken() returned duplicate token: %s", token)
		}
		tokens[token] = true
	}
}

// TestInjectCSRFToken tests that tokens are correctly injected into form Info structs
func TestInjectCSRFToken(t *testing.T) {
	// Create a request with CSRF token in context
	testToken := "test-csrf-token"
	r, _ := http.NewRequest("GET", "/", nil)
	ctx := context.WithValue(r.Context(), csrf.CSRFTokenContextKey, testToken)
	r = r.WithContext(ctx)

	// Test with default CsrfField
	info := &Info{}
	InjectCSRFToken(r, info)
	if info.CsrfValue != testToken {
		t.Errorf("InjectCSRFToken() CsrfValue = %v, want %v", info.CsrfValue, testToken)
	}
	if info.CsrfField != DefaultCSRFField {
		t.Errorf("InjectCSRFToken() CsrfField = %v, want %v", info.CsrfField, DefaultCSRFField)
	}

	// Test with custom CsrfField
	customField := "custom_csrf"
	info = &Info{CsrfField: customField}
	InjectCSRFToken(r, info)
	if info.CsrfValue != testToken {
		t.Errorf("InjectCSRFToken() CsrfValue = %v, want %v", info.CsrfValue, testToken)
	}
	if info.CsrfField != customField {
		t.Errorf("InjectCSRFToken() CsrfField = %v, want %v", info.CsrfField, customField)
	}
}

// TestCSRFMiddleware_GET tests the middleware handling for GET requests
func TestCSRFMiddleware_GET(t *testing.T) {
	f := NewForm(templates.BootstrapV5)

	// Create a test handler that verifies a token is present in the context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := GetCSRFToken(r)
		if !ok {
			t.Error("CSRFMiddleware didn't add token to context")
		}
		if token == "" {
			t.Error("CSRFMiddleware added empty token to context")
		}
		w.WriteHeader(http.StatusOK)
	})

	// Apply middleware
	handler := f.CSRFMiddleware()(testHandler)

	// Create test request and response
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Execute the handler with middleware
	handler.ServeHTTP(w, r)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("CSRFMiddleware() status = %v, want %v", w.Code, http.StatusOK)
	}

	// Check if cookie was set
	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == csrf.DefaultSessionID {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil {
		t.Error("CSRFMiddleware() didn't set session cookie")
	}
}

// TestCSRFMiddleware_POST_Valid tests the middleware handling for POST requests with valid tokens
func TestCSRFMiddleware_POST_Valid(t *testing.T) {
	f := NewForm(templates.BootstrapV5)
	store := f.GetCSRFStore()

	// Generate a session ID and token
	sessionID := "test-session-id"
	testToken := "test-csrf-token"

	// Store the token
	_ = store.Store(sessionID, testToken)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if a new token was generated
		newToken, ok := GetCSRFToken(r)
		if !ok {
			t.Error("CSRFMiddleware didn't add new token to context after validation")
		}
		if newToken == testToken {
			t.Error("CSRFMiddleware didn't refresh token after validation")
		}
		w.WriteHeader(http.StatusOK)
	})

	// Apply middleware
	handler := f.CSRFMiddleware()(testHandler)

	// Create test POST request with token
	formData := url.Values{}
	formData.Set(DefaultCSRFField, testToken)
	r := httptest.NewRequest("POST", "/", strings.NewReader(formData.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Add session cookie
	r.AddCookie(&http.Cookie{
		Name:  csrf.DefaultSessionID,
		Value: sessionID,
	})

	w := httptest.NewRecorder()

	// Execute the handler with middleware
	handler.ServeHTTP(w, r)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("CSRFMiddleware() status = %v, want %v", w.Code, http.StatusOK)
	}
}

// TestCSRFMiddleware_POST_Invalid tests the middleware handling for POST requests with invalid tokens
func TestCSRFMiddleware_POST_Invalid(t *testing.T) {
	f := NewForm(templates.BootstrapV5)
	store := f.GetCSRFStore()

	// Generate a session ID and token
	sessionID := "test-session-id"
	testToken := "test-csrf-token"

	// Store the token
	_ = store.Store(sessionID, testToken)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This should not be called
		t.Error("Handler was called despite invalid CSRF token")
		w.WriteHeader(http.StatusOK)
	})

	// Apply middleware
	handler := f.CSRFMiddleware()(testHandler)

	// Create test POST request with invalid token
	formData := url.Values{}
	formData.Set(DefaultCSRFField, "wrong-token")
	r := httptest.NewRequest("POST", "/", strings.NewReader(formData.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Add session cookie
	r.AddCookie(&http.Cookie{
		Name:  csrf.DefaultSessionID,
		Value: sessionID,
	})

	w := httptest.NewRecorder()

	// Execute the handler with middleware
	handler.ServeHTTP(w, r)

	// Check response - should be forbidden
	if w.Code != http.StatusForbidden {
		t.Errorf("CSRFMiddleware() status = %v, want %v", w.Code, http.StatusForbidden)
	}
}

// TestCSRFMiddleware_POST_MissingToken tests the middleware handling for POST requests without tokens
func TestCSRFMiddleware_POST_MissingToken(t *testing.T) {
	f := NewForm(templates.BootstrapV5)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This should not be called
		t.Error("Handler was called despite missing CSRF token")
		w.WriteHeader(http.StatusOK)
	})

	// Apply middleware
	handler := f.CSRFMiddleware()(testHandler)

	// Create test POST request without token
	r := httptest.NewRequest("POST", "/", nil)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()

	// Execute the handler with middleware
	handler.ServeHTTP(w, r)

	// Check response - should be bad request
	if w.Code != http.StatusBadRequest {
		t.Errorf("CSRFMiddleware() status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

// TestCSRFEndToEnd tests a complete request flow including form rendering and submission
func TestCSRFEndToEnd(t *testing.T) {
	// Create form renderer
	f := NewForm(templates.BootstrapV5)

	// Create a simple test form
	type TestForm struct {
		Info `target:"/test" method:"post"`
		Name string `form:"input,text" required:"true"`
	}

	// First request: GET the form
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create and render form
		form := TestForm{
			Info: Info{
				Target:     "/test",
				Method:     "post",
				SubmitText: "Submit",
				CsrfField:  DefaultCSRFField,
			},
		}

		// Inject CSRF token
		InjectCSRFToken(r, &form.Info)

		// Write token to response for testing
		w.Header().Set("X-CSRF-Token", form.CsrfValue)
		w.WriteHeader(http.StatusOK)
	})

	// Apply middleware
	middlewareHandler := f.CSRFMiddleware()(handler)

	// Make GET request
	rGet := httptest.NewRequest("GET", "/test", nil)
	wGet := httptest.NewRecorder()
	middlewareHandler.ServeHTTP(wGet, rGet)

	// Check response
	if wGet.Code != http.StatusOK {
		t.Fatalf("GET request status = %v, want %v", wGet.Code, http.StatusOK)
	}

	// Extract token and session cookie for next request
	token := wGet.Header().Get("X-CSRF-Token")
	if token == "" {
		t.Fatal("No CSRF token in response")
	}

	cookies := wGet.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == csrf.DefaultSessionID {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("No session cookie in response")
	}

	// Second request: POST the form with the token
	formData := url.Values{}
	formData.Set("name", "Test User")
	formData.Set(DefaultCSRFField, token)

	rPost := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	rPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rPost.AddCookie(sessionCookie)

	wPost := httptest.NewRecorder()

	// Different handler for POST to verify CSRF token was validated
	postHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This should be called if CSRF validation succeeds
		w.WriteHeader(http.StatusOK)
	})

	postMiddlewareHandler := f.CSRFMiddleware()(postHandler)
	postMiddlewareHandler.ServeHTTP(wPost, rPost)

	// Check response
	if wPost.Code != http.StatusOK {
		t.Errorf("POST request status = %v, want %v", wPost.Code, http.StatusOK)
	}
}

// TestMultipleFormSubmissions tests that CSRF tokens are properly refreshed on multiple submissions
func TestMultipleFormSubmissions(t *testing.T) {
	// Create form renderer
	f := NewForm(templates.BootstrapV5)

	// Create a sequence of tokens to verify refreshing
	var tokens []string

	// Handler that logs the current token and returns a 200 status
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := GetCSRFToken(r)
		if !ok {
			t.Error("No CSRF token in request context")
		}
		tokens = append(tokens, token)

		w.WriteHeader(http.StatusOK)
	})

	// Apply middleware
	middlewareHandler := f.CSRFMiddleware()(handler)

	// Initial GET request to set up session
	rGet := httptest.NewRequest("GET", "/", nil)
	wGet := httptest.NewRecorder()
	middlewareHandler.ServeHTTP(wGet, rGet)

	// Get the session cookie
	var sessionCookie *http.Cookie
	for _, cookie := range wGet.Result().Cookies() {
		if cookie.Name == csrf.DefaultSessionID {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("No session cookie in response")
	}

	// Make three POST submissions to verify token changes
	for i := 0; i < 3; i++ {
		// Get the current token
		currentToken := tokens[len(tokens)-1]

		// Create form data with current token
		formData := url.Values{}
		formData.Set(DefaultCSRFField, currentToken)

		rPost := httptest.NewRequest("POST", "/", strings.NewReader(formData.Encode()))
		rPost.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rPost.AddCookie(sessionCookie)

		wPost := httptest.NewRecorder()
		middlewareHandler.ServeHTTP(wPost, rPost)

		// Check response
		if wPost.Code != http.StatusOK {
			t.Errorf("POST request %d status = %v, want %v", i+1, wPost.Code, http.StatusOK)
		}
	}

	// Verify all tokens are different
	if len(tokens) != 4 { // 1 GET + 3 POST requests
		t.Errorf("Expected 4 tokens, got %d", len(tokens))
	}

	tokenSet := make(map[string]bool)
	for i, token := range tokens {
		if tokenSet[token] {
			t.Errorf("Token %d is a duplicate: %s", i, token)
		}
		tokenSet[token] = true
	}
}

// MockCSRFStore is a mock implementation of CSRFStore for testing
type MockCSRFStore struct {
	tokens map[string]string
	calls  map[string]int
}

func NewMockCSRFStore() *MockCSRFStore {
	return &MockCSRFStore{
		tokens: make(map[string]string),
		calls:  make(map[string]int),
	}
}

func (m *MockCSRFStore) Store(key, token string) error {
	m.calls["Store"]++
	m.tokens[key] = token
	return nil
}

func (m *MockCSRFStore) Get(key string) (string, error) {
	m.calls["Get"]++
	token, ok := m.tokens[key]
	if !ok {
		return "", errors.New("token not found")
	}
	return token, nil
}

func (m *MockCSRFStore) Delete(key string) error {
	m.calls["Delete"]++
	delete(m.tokens, key)
	return nil
}

func (m *MockCSRFStore) Validate(key, token string) error {
	m.calls["Validate"]++
	storedToken, ok := m.tokens[key]
	if !ok {
		return errors.New("token not found")
	}
	if storedToken != token {
		return csrf.ErrTokenMismatch
	}
	return nil
}

func (m *MockCSRFStore) ValidateCSRFToken(r *http.Request, csrfField, token string) error {
	m.calls["ValidateCSRFToken"]++
	sessionID, err := getSessionID(r)
	if err != nil {
		return err
	}
	return m.Validate(sessionID, token)
}

func (m *MockCSRFStore) AddCSRFToken(r *http.Request, csrfField string) error {
	m.calls["AddCSRFToken"]++
	sessionID, err := getSessionID(r)
	if err != nil {
		return err
	}
	token, err := csrf.GenerateCSRFToken()
	if err != nil {
		return err
	}
	m.tokens[sessionID] = token
	*r = *r.WithContext(context.WithValue(r.Context(), csrf.CSRFTokenContextKey, token))
	return nil
}

func (m *MockCSRFStore) DeleteToken(r *http.Request, csrfField string) error {
	m.calls["DeleteToken"]++
	sessionID, err := getSessionID(r)
	if err != nil {
		return err
	}
	delete(m.tokens, sessionID)
	return nil
}

// TestCustomCSRFStore tests that a custom CSRF store can be used
func TestCustomCSRFStore(t *testing.T) {
	mockStore := NewMockCSRFStore()
	f := NewForm(templates.BootstrapV5)
	f.SetCSRFStore(mockStore)

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Apply middleware
	middlewareHandler := f.CSRFMiddleware()(handler)

	// Make GET request
	rGet := httptest.NewRequest("GET", "/", nil)
	wGet := httptest.NewRecorder()
	middlewareHandler.ServeHTTP(wGet, rGet)

	// Check that the mock store was used
	if mockStore.calls["Store"] != 1 {
		t.Errorf("Store() should be called 1 time, got %d", mockStore.calls["Store"])
	}
}

// TestCSRFFailureCases tests various failure scenarios for CSRF protection
func TestCSRFFailureCases(t *testing.T) {
	// Create form renderer
	f := NewForm(templates.BootstrapV5)

	// Success tracker
	handlerCalled := false

	// Create a simple handler that just sets the flag
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Apply CSRF middleware
	middlewareHandler := f.CSRFMiddleware()(handler)

	// Setup: First make a GET request to establish a session and token
	rGet := httptest.NewRequest(http.MethodGet, "/", nil)
	wGet := httptest.NewRecorder()
	middlewareHandler.ServeHTTP(wGet, rGet)

	// Extract session cookie
	var sessionCookie *http.Cookie
	for _, cookie := range wGet.Result().Cookies() {
		if cookie.Name == csrf.DefaultSessionID {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("No session cookie in response")
	}

	// Test cases
	tests := []struct {
		name                  string
		setupRequest          func() *http.Request
		expectedStatus        int
		handlerShouldBeCalled bool
	}{
		{
			name: "Missing CSRF token",
			setupRequest: func() *http.Request {
				// Empty form data without token
				formData := url.Values{}
				r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formData.Encode()))
				r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				r.AddCookie(sessionCookie)
				return r
			},
			expectedStatus:        http.StatusBadRequest,
			handlerShouldBeCalled: false,
		},
		{
			name: "Invalid CSRF token",
			setupRequest: func() *http.Request {
				// Form with made-up token
				formData := url.Values{}
				formData.Set(DefaultCSRFField, "invalid-token-that-doesnt-exist")
				r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formData.Encode()))
				r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				r.AddCookie(sessionCookie)
				return r
			},
			expectedStatus:        http.StatusForbidden,
			handlerShouldBeCalled: false,
		},
		{
			name: "Missing session cookie",
			setupRequest: func() *http.Request {
				// Form with some token but no session cookie
				formData := url.Values{}
				formData.Set(DefaultCSRFField, "some-token")
				r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formData.Encode()))
				r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				// Deliberately not adding cookie
				return r
			},
			expectedStatus:        http.StatusBadRequest, // Session error
			handlerShouldBeCalled: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset flag
			handlerCalled = false

			// Setup request
			r := tc.setupRequest()
			w := httptest.NewRecorder()

			// Execute
			middlewareHandler.ServeHTTP(w, r)

			// Verify status code
			if w.Code != tc.expectedStatus {
				t.Errorf("Status = %v, want %v", w.Code, tc.expectedStatus)
			}

			// Verify if handler was called
			if handlerCalled != tc.handlerShouldBeCalled {
				t.Errorf("Handler called = %v, want %v", handlerCalled, tc.handlerShouldBeCalled)
			}
		})
	}
}

// TestCSRFTokenReuseAttempt tests that a token cannot be reused after it's been consumed
func TestCSRFTokenReuseAttempt(t *testing.T) {
	// Create form renderer
	f := NewForm(templates.BootstrapV5)

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the token from context and put it in a header for testing
		if token, ok := GetCSRFToken(r); ok {
			w.Header().Set("X-CSRF-Token", token)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Apply CSRF middleware
	middlewareHandler := f.CSRFMiddleware()(handler)

	// Initial GET request to get token and session
	rGet := httptest.NewRequest(http.MethodGet, "/", nil)
	wGet := httptest.NewRecorder()
	middlewareHandler.ServeHTTP(wGet, rGet)

	// Get token from response header
	token := wGet.Header().Get("X-CSRF-Token")
	if token == "" {
		t.Fatal("No CSRF token in response")
	}

	// Get session cookie
	var sessionCookie *http.Cookie
	for _, cookie := range wGet.Result().Cookies() {
		if cookie.Name == csrf.DefaultSessionID {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("No session cookie in response")
	}

	// First POST with the token - should succeed
	formData := url.Values{}
	formData.Set(DefaultCSRFField, token)
	rPost1 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formData.Encode()))
	rPost1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rPost1.AddCookie(sessionCookie)

	wPost1 := httptest.NewRecorder()
	middlewareHandler.ServeHTTP(wPost1, rPost1)

	// Verify success
	if wPost1.Code != http.StatusOK {
		t.Errorf("First POST status = %v, want %v", wPost1.Code, http.StatusOK)
	}

	// Second POST with the same token - should fail
	rPost2 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(formData.Encode()))
	rPost2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rPost2.AddCookie(sessionCookie)

	wPost2 := httptest.NewRecorder()
	middlewareHandler.ServeHTTP(wPost2, rPost2)

	// Verify failure
	if wPost2.Code != http.StatusForbidden {
		t.Errorf("Second POST with same token status = %v, want %v", wPost2.Code, http.StatusForbidden)
	}
}

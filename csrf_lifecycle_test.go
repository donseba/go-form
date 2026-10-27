package form

import (
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/donseba/go-form/v2/csrf"
)

func csrfRequest(handler http.Handler, method, path, token string, cookie *http.Cookie) *httptest.ResponseRecorder {
	values := url.Values{DefaultCSRFField: {token}}
	req := httptest.NewRequest(method, path, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func csrfTestHandler(f *Form) http.Handler {
	return f.CSRFMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _ := GetCSRFToken(r)
		w.Header().Set("X-CSRF-Token", token)
	}))
}

func TestCSRFFormsSurviveOtherRequests(t *testing.T) {
	stores := map[string]func() csrf.Store{
		"default": func() csrf.Store { return csrf.NewDefaultMemoryCSRFStore() },
		"memory":  func() csrf.Store { return csrf.NewMemoryCSRFStore() },
		"legacy":  func() csrf.Store { return NewMockCSRFStore() },
	}
	for name, newStore := range stores {
		t.Run(name, func(t *testing.T) {
			f := NewForm()
			f.SetCSRFStore(newStore())
			handler := csrfTestHandler(f)
			first := csrfRequest(handler, http.MethodGet, "/edit", "", nil)
			cookie := first.Result().Cookies()[0]
			firstToken := first.Header().Get("X-CSRF-Token")
			second := csrfRequest(handler, http.MethodGet, "/other-tab", "", cookie)
			secondToken := second.Header().Get("X-CSRF-Token")
			csrfRequest(handler, http.MethodGet, "/fragment", "", cookie)
			csrfRequest(handler, http.MethodHead, "/fragment", "", cookie)

			post := csrfRequest(handler, http.MethodPost, "/edit", firstToken, cookie)
			if post.Code != http.StatusOK {
				t.Fatalf("earlier form rejected after GET/HEAD: %d %s", post.Code, post.Body.String())
			}
			refreshed := post.Header().Get("X-CSRF-Token")
			if refreshed == "" || refreshed == firstToken {
				t.Fatal("submission did not issue a fresh token")
			}
			if post := csrfRequest(handler, http.MethodPost, "/other-tab", secondToken, cookie); post.Code != http.StatusOK {
				t.Fatalf("other form rejected after a submission: %d %s", post.Code, post.Body.String())
			}
			if post := csrfRequest(handler, http.MethodPost, "/edit", refreshed, cookie); post.Code != http.StatusOK {
				t.Fatalf("refreshed token rejected: %d %s", post.Code, post.Body.String())
			}
			if post := csrfRequest(handler, http.MethodPost, "/edit", firstToken, cookie); post.Code != http.StatusForbidden {
				t.Fatalf("consumed token accepted: %d", post.Code)
			}
		})
	}
}

func TestCSRFConcurrentForms(t *testing.T) {
	f := NewForm()
	handler := csrfTestHandler(f)
	first := csrfRequest(handler, http.MethodGet, "/edit", "", nil)
	cookie := first.Result().Cookies()[0]
	var requests sync.WaitGroup
	for range 32 {
		requests.Add(1)
		go func() {
			defer requests.Done()
			get := csrfRequest(handler, http.MethodGet, "/fragment", "", cookie)
			token := get.Header().Get("X-CSRF-Token")
			post := csrfRequest(handler, http.MethodPost, "/edit", token, cookie)
			if post.Code != http.StatusOK {
				t.Errorf("concurrent form rejected: %d %s", post.Code, post.Body.String())
			}
		}()
	}
	requests.Wait()
}

func TestCSRFTokenRemainsBoundToSession(t *testing.T) {
	f := NewForm()
	handler := csrfTestHandler(f)
	first := csrfRequest(handler, http.MethodGet, "/edit", "", nil)
	second := csrfRequest(handler, http.MethodGet, "/edit", "", nil)
	token := first.Header().Get("X-CSRF-Token")
	if post := csrfRequest(handler, http.MethodPost, "/edit", token, second.Result().Cookies()[0]); post.Code != http.StatusForbidden {
		t.Fatalf("cross-session token accepted: %d", post.Code)
	}
	if post := csrfRequest(handler, http.MethodPost, "/edit", token, first.Result().Cookies()[0]); post.Code != http.StatusOK {
		t.Fatalf("cross-session rejection consumed the real token: %d", post.Code)
	}
}

func TestCSRFExpiredTokenRejected(t *testing.T) {
	f := NewForm()
	f.SetCSRFStore(csrf.NewDefaultMemoryCSRFStore(-time.Second, time.Hour))
	handler := csrfTestHandler(f)
	get := csrfRequest(handler, http.MethodGet, "/edit", "", nil)
	post := csrfRequest(handler, http.MethodPost, "/edit", get.Header().Get("X-CSRF-Token"), get.Result().Cookies()[0])
	if post.Code != http.StatusForbidden {
		t.Fatalf("expired token status = %d, want 403", post.Code)
	}
}

func TestCSRFSubmissionWithoutSessionDoesNotCreateCookie(t *testing.T) {
	post := csrfRequest(csrfTestHandler(NewForm()), http.MethodPost, "/edit", "token", nil)
	if post.Code != http.StatusBadRequest || len(post.Result().Cookies()) != 0 {
		t.Fatalf("missing session status = %d, cookies = %v", post.Code, post.Result().Cookies())
	}
}

// csrfFormHandler renders a form's hidden fields with the request's token,
// as a page below the middleware does.
func csrfFormHandler(f *Form, info Info) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hidden, err := f.Hidden(nil, WithRequestInfo(r, struct{ Name string }{}, info))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		_, _ = w.Write([]byte(hidden))
	})
}

// csrfHiddenInput is the name and value of the hidden token input in body.
func csrfHiddenInput(t *testing.T, body string) (name, value string) {
	t.Helper()

	match := regexp.MustCompile(`<input type="hidden" name="([^"]+)" value="([^"]+)"`).FindStringSubmatch(body)
	if match == nil {
		t.Fatalf("no hidden token input in %s", body)
	}

	return match[1], html.UnescapeString(match[2])
}

func TestCSRFCustomFieldName(t *testing.T) {
	tests := map[string]Info{
		"from the middleware": {Target: "/", Method: http.MethodPost},
		"set on the form":     {Target: "/", Method: http.MethodPost, CsrfField: "token"},
	}

	for name, info := range tests {
		t.Run(name, func(t *testing.T) {
			f := NewForm()
			handler := f.CSRFMiddlewareWithOptions(CSRFOptions{FieldName: "token"})(csrfFormHandler(f, info))

			get := httptest.NewRecorder()
			handler.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/", nil))

			field, token := csrfHiddenInput(t, get.Body.String())
			if field != "token" {
				t.Fatalf("token rendered in %q, want token", field)
			}

			values := url.Values{field: {token}}
			post := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
			post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			post.AddCookie(get.Result().Cookies()[0])

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, post)

			if rec.Code != http.StatusOK {
				t.Fatalf("submission with the token in %q: %d %s", field, rec.Code, rec.Body.String())
			}

			if field, _ := csrfHiddenInput(t, rec.Body.String()); field != "token" {
				t.Errorf("fresh token rendered in %q, want token", field)
			}
		})
	}
}

func TestCSRFDefaultFieldNameIgnoresOtherFields(t *testing.T) {
	f := NewForm()
	handler := csrfTestHandler(f)
	get := csrfRequest(handler, http.MethodGet, "/", "", nil)

	values := url.Values{"token": {get.Header().Get("X-CSRF-Token")}}
	post := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	post.AddCookie(get.Result().Cookies()[0])

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, post)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("token outside %q: status %d, want 400", DefaultCSRFField, rec.Code)
	}
}

// sessionCookie is the session cookie that rec sets, or nil.
func sessionCookie(rec *httptest.ResponseRecorder) *http.Cookie {
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == csrf.DefaultSessionID {
			return cookie
		}
	}

	return nil
}

// expiresIn reports whether cookie expires about lifetime from now.
func expiresIn(cookie *http.Cookie, lifetime time.Duration) bool {
	left := time.Until(cookie.Expires)

	return left > lifetime-5*time.Second && left <= lifetime+time.Second
}

func TestCSRFRenewsSessionCookie(t *testing.T) {
	stores := map[string]struct {
		store    csrf.Store
		lifetime time.Duration
	}{
		"default store":  {csrf.NewDefaultMemoryCSRFStore(), csrf.DefaultExpirationTime},
		"longer tokens":  {csrf.NewDefaultMemoryCSRFStore(30 * time.Minute), 30 * time.Minute},
		"without expiry": {csrf.NewMemoryCSRFStore(), csrf.DefaultExpirationTime},
	}

	for name, test := range stores {
		t.Run(name, func(t *testing.T) {
			f := NewForm()
			f.SetCSRFStore(test.store)
			handler := csrfTestHandler(f)

			first := csrfRequest(handler, http.MethodGet, "/", "", nil)

			created := sessionCookie(first)
			if created == nil || !expiresIn(created, test.lifetime) {
				t.Fatalf("new session cookie %v, want one that expires in %s", created, test.lifetime)
			}

			// A later page renews the cookie, so that its form outlives
			// the cookie set by the first page.
			later := csrfRequest(handler, http.MethodGet, "/", "", created)

			renewed := sessionCookie(later)
			if renewed == nil || renewed.Value != created.Value || !expiresIn(renewed, test.lifetime) {
				t.Fatalf("later page: cookie %v, want %s renewed for %s", renewed, created.Value, test.lifetime)
			}

			// So does an accepted submission, which gets a fresh token.
			post := csrfRequest(handler, http.MethodPost, "/", later.Header().Get("X-CSRF-Token"), created)

			submitted := sessionCookie(post)
			if post.Code != http.StatusOK || submitted == nil || submitted.Value != created.Value {
				t.Fatalf("submission: status %d, cookie %v, want %s renewed", post.Code, submitted, created.Value)
			}

			// A refused submission does not.
			refused := csrfRequest(handler, http.MethodPost, "/", "wrong", created)
			if refused.Code != http.StatusForbidden || sessionCookie(refused) != nil {
				t.Fatalf("refused submission: status %d, cookies %v", refused.Code, refused.Result().Cookies())
			}
		})
	}
}

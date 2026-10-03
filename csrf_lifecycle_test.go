package form

import (
	"net/http"
	"net/http/httptest"
	"net/url"
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

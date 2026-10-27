package main

import (
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// TestThemesConcurrently renders every theme at once: each response shows
// the theme it asked for, and with -race no request writes what another
// reads, as SetTheme on a shared renderer did.
func TestThemesConcurrently(t *testing.T) {
	marks := map[string]string{
		"bootstrap":  `class="form-control"`,
		"tailwind":   "border border-gray-200 block w-full",
		"tailwindv4": "focus-visible:ring-offset-white",
		"plain":      "padding: 0.375rem 0.75rem",
	}

	handler := newHandler()

	var requests sync.WaitGroup

	for round := range 8 {
		for theme, mark := range marks {
			requests.Add(1)

			go func() {
				defer requests.Done()

				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, httptest.NewRequest("GET", "/?theme="+theme, nil))

				body := rec.Body.String()
				if rec.Code != 200 || !strings.Contains(body, mark) {
					t.Errorf("round %d, theme %s: status %d, mark %q missing", round, theme, rec.Code, mark)
				}

				for other, otherMark := range marks {
					if other != theme && strings.Contains(body, otherMark) {
						t.Errorf("round %d, theme %s: rendered with %s", round, theme, other)
					}
				}
			}()
		}
	}

	requests.Wait()
}

// TestUnknownThemeShowsDefault shows the first theme for a theme the
// example does not have.
func TestUnknownThemeShowsDefault(t *testing.T) {
	rec := httptest.NewRecorder()
	newHandler().ServeHTTP(rec, httptest.NewRequest("GET", "/?theme=unknown", nil))

	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `class="form-control"`) {
		t.Fatalf("status %d, body without the bootstrap theme: %s", rec.Code, rec.Body.String())
	}
}

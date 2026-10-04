package form

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestMapFormClearsSubmittedEmptyValues(t *testing.T) {
	type Post struct {
		Title      string
		Excerpt    string
		Untouched  string
		Image      SortedSelect[string]
		Categories SortedMultiSelect[string]
	}

	// An edit form is filled with the stored values first.
	post := Post{
		Title:      "Hello",
		Excerpt:    "Old excerpt",
		Untouched:  "kept",
		Image:      NewSortedSelect(map[string]string{"": "—", "a": "a.png"}),
		Categories: NewSortedMultiSelect(map[string]string{"x": "X", "y": "Y"}),
	}
	_ = post.Image.Set("a")
	_ = post.Categories.Set([]string{"x", "y"})

	// The user clears the excerpt, picks no image and unticks every category.
	body := url.Values{"Title": {"Hello"}, "Excerpt": {""}, "Image": {""}}
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err := MapForm(r, &post); err != nil {
		t.Fatal(err)
	}

	if post.Excerpt != "" {
		t.Errorf("excerpt not cleared: %q", post.Excerpt)
	}

	if post.Untouched != "kept" {
		t.Errorf("field missing from the form changed: %q", post.Untouched)
	}

	if post.Image.Get() != "" {
		t.Errorf("image not cleared: %q", post.Image.Get())
	}

	if len(post.Categories.Get()) != 0 {
		t.Errorf("categories not cleared: %v", post.Categories.Get())
	}
}

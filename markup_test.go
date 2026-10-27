package form

import (
	"regexp"
	"strings"
	"testing"

	"github.com/donseba/go-form/v2/templates"
)

type markupInput struct {
	Subject string `form:"input,hidden" name:"subject" label:"Subject" placeholder:"Not shown"`
	Name    string `form:"input,text"   name:"name"    label:"Name"                            description:"Your full name" required:"true"`
	Remarks string `form:"textarea"     name:"remarks" label:"Remarks"`
	Agree   bool   `form:"checkbox"     name:"agree"   label:"I agree"`
}

// doubleSpace finds two spaces inside a tag on one line, such as
// `<div  class=`.
var doubleSpace = regexp.MustCompile(`<\w+  |[^\s]  +(class|style|id|role|for|aria-[a-z]+)=`)

func TestMarkupHasNoDoubleSpaces(t *testing.T) {
	for _, theme := range []string{"bootstrap", "tailwind", "tailwindv4", "plain"} {
		f := NewForm()
		f.SetTheme(theme)

		model := WithInfo(markupInput{}, Info{Target: "/send", Method: "post", CancelTarget: "/back"})
		errs := FieldErrors{FieldValidationError{Field: "name", Err: "Too short"}}

		html, err := f.Render(nil, model, errs)
		if err != nil {
			t.Fatalf("%s: %v", theme, err)
		}

		if found := doubleSpace.FindString(string(html)); found != "" {
			t.Errorf("%s: double space at %q in\n%s", theme, found, html)
		}
	}
}

func TestHiddenInputCarriesNameAndValueOnly(t *testing.T) {
	f := NewForm()
	f.SetTheme("bootstrap")

	html, err := f.Field(nil, markupInput{Subject: "Hello"}, "subject", nil)
	if err != nil {
		t.Fatal(err)
	}

	want := `<input type="hidden" id="subject" name="subject" value="Hello">`
	if strings.TrimSpace(string(html)) != want {
		t.Errorf("hidden input\n got %s\nwant %s", html, want)
	}
}

// TestRequiredMarkerIsVisuallyHidden checks that the "(required)" for screen
// readers gets the theme's class for visually hidden text, or an inline
// style in the plain theme, so it does not show on screen; a theme of the
// application's own without VisuallyHidden keeps the class sr-only.
func TestRequiredMarkerIsVisuallyHidden(t *testing.T) {
	if _, err := templates.RegisterClassTheme("visually-hidden-test", templates.ThemeClasses{
		Label: templates.StyleOption{Class: "label"},
	}); err != nil {
		t.Fatal(err)
	}

	for theme, want := range map[string]string{
		"bootstrap":            `<span class="visually-hidden">(required)</span>`,
		"tailwind":             `<span class="sr-only">(required)</span>`,
		"tailwindv4":           `<span class="sr-only">(required)</span>`,
		"plain":                `<span style="` + templates.VisuallyHiddenStyle + `" class="sr-only">(required)</span>`,
		"visually-hidden-test": `<span class="sr-only">(required)</span>`,
	} {
		f := NewForm()
		f.SetTheme(theme)

		html, err := f.Field(nil, markupInput{}, "name", nil)
		if err != nil {
			t.Fatalf("%s: %v", theme, err)
		}

		if !strings.Contains(string(html), want) {
			t.Errorf("%s: want %s in\n%s", theme, want, html)
		}
	}
}

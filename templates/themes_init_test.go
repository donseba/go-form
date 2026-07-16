package templates_test

import (
	"html/template"
	"io"
	"testing"

	"github.com/donseba/go-form/v2/templates"
	"github.com/donseba/go-form/v2/types"
)

type dummyLoc struct{}

func (dummyLoc) GetLocale() string { return "en" }

func TestInitThemes_LoadsEmbeddedTemplates(t *testing.T) {
	if err := templates.InitThemes(); err != nil {
		t.Fatalf("InitThemes: %v", err)
	}

	for _, name := range []string{"bootstrap", "tailwind", "tailwindv4", "plain"} {
		theme, ok := templates.GetTheme(name)
		if !ok {
			t.Fatalf("theme %q not registered", name)
		}
		if theme.Templates == nil {
			t.Fatalf("theme %q templates not loaded", name)
		}

		// Smoke-check that key templates exist.
		for _, tmplName := range []string{"form", "input", "button", "wrapper", "label"} {
			if theme.Templates.Lookup(tmplName) == nil {
				t.Fatalf("theme %q missing template %q", name, tmplName)
			}
		}

		// Also ensure that templates can be cloned and executed multiple times (common pattern for injecting funcs).
		cl, err := theme.Templates.Clone()
		if err != nil {
			t.Fatalf("theme %q: clone: %v", name, err)
		}
		cl = cl.Funcs(map[string]any{
			"fields": func() template.HTML { return "" },
		})
		for i := 0; i < 2; i++ {
			if err := cl.ExecuteTemplate(io.Discard, "form", map[string]any{
				"Field": types.FormField{Target: "/", Method: "POST", Label: "Save"},
				"Loc":   dummyLoc{},
			}); err != nil {
				t.Fatalf("theme %q: execute (iteration %d): %v", name, i, err)
			}
		}
	}
}

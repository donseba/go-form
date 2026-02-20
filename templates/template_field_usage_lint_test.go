package templates

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/donseba/go-form/types"
)

// Single “dummy” data struct used only for template execution.
// Field is a types.FormField so invalid `.Field.*` lookups error.
// Loc is any struct with GetLocale.

type dummyLoc struct{}

func (dummyLoc) GetLocale() string { return "en" }

// This test is a "template linter" that ensures every gohtml theme template only references
// known `.Field.*` properties.
func TestGoHTMLTemplates_FieldAccessIsValid(t *testing.T) {
	// Allowlist of valid `.Field.<Name>` properties.
	// This is intentionally strict and matches `types.FormField`.
	allowedFieldProps := map[string]struct{}{
		"Type":         {},
		"InputType":    {},
		"Name":         {},
		"Id":           {},
		"Label":        {},
		"Value":        {},
		"Placeholder":  {},
		"Required":     {},
		"Hidden":       {},
		"Disabled":     {},
		"Min":          {},
		"Max":          {},
		"Step":         {},
		"MaxLength":    {},
		"Description":  {},
		"Rows":         {},
		"Cols":         {},
		"Legend":       {},
		"Values":       {},
		"Fields":       {},
		"Target":       {},
		"Method":       {},
		"CancelTarget": {},
		"CancelText":   {},
		"Attributes":   {},
		"GroupBefore":  {},
		"GroupAfter":   {},
		"Class":        {},
		"Data":         {},
		"ValueMap":     {}, // allow ValueMap for multicheckbox
	}

	allowedRoot := map[string]struct{}{
		"Field": {},
		"Loc":   {},
		"Type":  {}, // used by input.gohtml: {{.Type}}

		// input-group template context
		"GroupBefore": {},
		"GroupAfter":  {},
		"Input":       {},
	}

	// Helpers provided by theme loader + renderer.
	allowedFuncs := template.FuncMap{
		"themeClass":           func(string) string { return "" },
		"themeStyle":           func(string) template.HTML { return "" },
		"themeAttr":            func(string) string { return "" },
		"default":              func(v any, fb any) any { return v },
		"form_print":           func(...any) string { return "" },
		"form_data_attributes": func(...any) string { return "" },
		"form_attributes":      func(...any) string { return "" },
		// callable blocks in some templates
		"errors": func() []string { return nil },
		"field":  func() template.HTML { return "" },
		"fields": func() template.HTML { return "" },
		"label":  func() template.HTML { return "" },
	}

	dummy := struct {
		Field       types.FormField
		Loc         dummyLoc
		Type        string
		GroupBefore template.HTML
		GroupAfter  template.HTML
		Input       template.HTML
	}{
		Field:       types.FormField{},
		Loc:         dummyLoc{},
		Type:        "text",
		GroupBefore: "",
		GroupAfter:  "",
		Input:       "",
	}

	// Walk embedded template FS and lint each .gohtml.
	var files []string
	err := fs.WalkDir(TemplateFS, "gohtml", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".gohtml" {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		t.Fatalf("walk embedded templates: %v", err)
	}

	if len(files) == 0 {
		t.Fatalf("no gohtml templates found in embedded FS")
	}

	for _, path := range files {
		path := path
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		t.Run(name, func(t *testing.T) {
			b, err := fs.ReadFile(TemplateFS, path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			tplStr := string(b)

			tpl, err := template.New(name).Funcs(allowedFuncs).Option("missingkey=error").Parse(tplStr)
			if err != nil {
				t.Fatalf("failed to parse template %s: %v", path, err)
			}

			var sb strings.Builder
			err = tpl.Execute(&sb, dummy)
			if err == nil {
				// Execution succeeded: we still want to verify `.Field.<prop>` names
				// against our allowlist, because execution may not hit some branches.
				referenced := fieldPropsReferenced(tplStr)
				for prop := range referenced {
					if _, ok := allowedFieldProps[prop]; !ok {
						t.Fatalf("unknown .Field.%s in template %s", prop, path)
					}
				}
				return
			}

			refd := fieldPropsReferenced(tplStr)
			for prop := range refd {
				if _, ok := allowedFieldProps[prop]; !ok {
					t.Fatalf("unknown .Field.%s in template %s (execution error: %v)", prop, path, err)
				}
			}

			for _, root := range rootsReferenced(tplStr) {
				if _, ok := allowedRoot[root]; !ok {
					_ = ok
				}
			}

			t.Fatalf("template execution failed (likely invalid field access): %v", err)
		})
	}
}

// fieldPropsReferenced extracts `.Field.<Prop>` occurrences from a template string.
// It’s intentionally simple and regex-free to keep it dependency-free.
func fieldPropsReferenced(tpl string) map[string]struct{} {
	out := map[string]struct{}{}

	needle := ".Field."
	for i := 0; i < len(tpl); {
		j := strings.Index(tpl[i:], needle)
		if j < 0 {
			break
		}
		j += i
		start := j + len(needle)
		end := start
		for end < len(tpl) {
			c := tpl[end]
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
				end++
				continue
			}
			break
		}
		if end > start {
			out[tpl[start:end]] = struct{}{}
		}
		i = end
	}

	return out
}

// rootsReferenced is a best-effort scan of `.Foo` occurrences. It’s not used as a strict gate.
func rootsReferenced(tpl string) []string {
	var out []string
	needle := "."
	for i := 0; i < len(tpl); {
		j := strings.Index(tpl[i:], needle)
		if j < 0 {
			break
		}
		j += i
		start := j + 1
		if start >= len(tpl) {
			break
		}
		if tpl[start] == 'F' && strings.HasPrefix(tpl[start:], "Field") {
			out = append(out, "Field")
		}
		if tpl[start] == 'L' && strings.HasPrefix(tpl[start:], "Loc") {
			out = append(out, "Loc")
		}
		if tpl[start] == 'T' && strings.HasPrefix(tpl[start:], "Type") {
			out = append(out, "Type")
		}
		i = start
	}
	return out
}

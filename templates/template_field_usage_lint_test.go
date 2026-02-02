package templates

import (
	"fmt"
	"strings"
	"testing"
	"text/template"

	"github.com/donseba/go-form/types"
)

// Single “dummy” data struct used only for template execution.
// Field is a types.FormField so invalid `.Field.*` lookups error.
// Loc is any struct with GetLocale.
type dummyLoc struct{}

func (dummyLoc) GetLocale() string { return "en" }

// This test is a "template linter" that ensures every template string only references
// known `.Field.*` properties.
//
// Why?
// - `html/template` will happily compile templates that reference unknown fields.
// - At runtime you'll get "can't evaluate field X" errors only when executing that specific template.
// - When adding a new template set (TailwindV4), it's easy to ship a typo like `.Field.Placehoder`.
func TestTemplateMaps_FieldAccessIsValid(t *testing.T) {
	templateSets := []struct {
		name string
		tm   types.TemplateMap
	}{
		{"Plain", Plain},
		{"BootstrapV5", BootstrapV5},
		{"TailwindV3", TailwindV3},
		{"TailwindV4", TailwindV4},
	}

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
	}

	allowedRoot := map[string]struct{}{
		"Field": {},
		"Loc":   {},
		"Type":  {}, // used by baseInput: {{.Type}}

		// input-group wrapper template context is not the same as other templates
		"GroupBefore": {},
		"GroupAfter":  {},
		"Input":       {},
	}

	// Helpers provided by Form.init(...) that templates may call.
	allowedFuncs := template.FuncMap{
		"baseInput":            func(...any) (string, error) { return "", nil },
		"form_print":           func(...any) string { return "" },
		"form_data_attributes": func(...any) string { return "" },
		"form_attributes":      func(...any) string { return "" },
		"errors":               func() []string { return nil },
		"field":                func() string { return "" },
		"fields":               func() string { return "" },
		"label":                func() string { return "" },
	}

	dummy := struct {
		Field       types.FormField
		Loc         dummyLoc
		Type        string
		GroupBefore string
		GroupAfter  string
		Input       string
	}{
		Field:       types.FormField{},
		Loc:         dummyLoc{},
		Type:        "text",
		GroupBefore: "",
		GroupAfter:  "",
		Input:       "",
	}

	for _, ts := range templateSets {
		ts := ts
		t.Run(ts.name, func(t *testing.T) {
			for ft, byInput := range ts.tm {
				for it, tplStr := range byInput {
					name := fmt.Sprintf("%s/%s/%s", ts.name, ft, it)

					// Quick static scan for `.Field.` occurrences so we can produce clearer errors
					// and also catch cases like `.Field` (no dot) if they exist.
					if strings.Contains(tplStr, ".Field") {
						// Check for `.Field` not followed by a dot, which is almost always a mistake.
						// (Allow `.Field}}` etc.)
						for _, idx := range stringsIndexes(tplStr, ".Field") {
							if idx+len(".Field") < len(tplStr) {
								n := tplStr[idx+len(".Field")]
								if n != '.' && n != ' ' && n != '}' && n != ')' && n != '\n' && n != '\t' {
									// .FieldX is suspicious, but template parsing will fail anyway; keep going.
								}
							}
						}
					}

					tpl, err := template.New(name).Funcs(allowedFuncs).Option("missingkey=error").Parse(tplStr)
					if err != nil {
						t.Fatalf("failed to parse template: %v", err)
					}

					// Validate root lookups and `.Field.<prop>` lookups by executing.
					// Use placeholder funcs so execution doesn't depend on the Form runtime.
					var sb strings.Builder
					err = tpl.Execute(&sb, dummy)
					if err == nil {
						// Execution succeeded: we still want to verify `.Field.<prop>` names
						// against our allowlist, because execution may not hit some branches.
						referenced := fieldPropsReferenced(tplStr)
						for prop := range referenced {
							if _, ok := allowedFieldProps[prop]; !ok {
								t.Fatalf("unknown .Field.%s in template", prop)
							}
						}
						continue
					}

					// When execution fails, try to produce a clearer message.
					refd := fieldPropsReferenced(tplStr)
					for prop := range refd {
						if _, ok := allowedFieldProps[prop]; !ok {
							t.Fatalf("unknown .Field.%s in template (execution error: %v)", prop, err)
						}
					}

					// Also check root identifiers used.
					for _, root := range rootsReferenced(tplStr) {
						if _, ok := allowedRoot[root]; !ok {
							// Not fatal: templates may reference $ vars etc. We keep the check lightweight.
							_ = ok
						}
					}

					t.Fatalf("template execution failed (likely invalid field access): %v", err)
				}
			}
		})
	}
}

func stringsIndexes(s, sub string) []int {
	if sub == "" {
		return nil
	}
	var idxs []int
	for i := 0; ; {
		j := strings.Index(s[i:], sub)
		if j < 0 {
			break
		}
		j += i
		idxs = append(idxs, j)
		i = j + len(sub)
		if i >= len(s) {
			break
		}
	}
	return idxs
}

// fieldPropsReferenced extracts `.Field.<Prop>` occurrences from a template string.
// It’s intentionally simple and regex-free to keep it dependency-free.
func fieldPropsReferenced(tpl string) map[string]struct{} {
	out := map[string]struct{}{}

	// Look for `.Field.` and read an identifier following it.
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
		i = start
	}
	return out
}

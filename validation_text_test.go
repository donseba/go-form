package form

import (
	"html/template"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/donseba/go-form/v2/templates"
	"github.com/donseba/go-form/v2/types"
)

func TestValidateRefusesInvalidUTF8(t *testing.T) {
	type input struct {
		Name     string `form:"input,text"`
		Password string `form:"input,password"`
		Note     string `form:"textarea"       invalidMsg:"Use plain text"`
	}

	errs := NewForm().ValidateForm(&input{Name: "a\xffb", Password: "\xc3", Note: "\xff"})

	want := FieldErrors{
		FieldValidationError{Field: "Name", Rule: "invalid", Err: "Enter a valid value."},
		FieldValidationError{Field: "Password", Rule: "invalid", Err: "Enter a valid value."},
		FieldValidationError{Field: "Note", Rule: "invalid", Err: "Use plain text"},
	}
	if !reflect.DeepEqual(errs, want) {
		t.Errorf("got %#v, want %#v", errs, want)
	}

	if errs := NewForm().ValidateForm(&input{Name: "Zoë", Note: "line\nline"}); len(errs) > 0 {
		t.Errorf("valid text refused: %v", errs)
	}
}

func TestValidateSingleLine(t *testing.T) {
	type input struct {
		Subject string `form:"input,text" line:"true"`
		Body    string `form:"textarea"`
	}

	for _, value := range []string{"a\nb", "a\rb", "a\tb", "a\x00b", "a b", "a\u0085b"} {
		errs := NewForm().ValidateForm(&input{Subject: value, Body: value})

		want := FieldErrors{
			FieldValidationError{Field: "Subject", Rule: "line", Err: "Use a single line, without tabs or line breaks."},
		}
		if !reflect.DeepEqual(errs, want) {
			t.Errorf("%q: got %#v", value, errs)
		}
	}

	if errs := NewForm().ValidateForm(&input{Subject: "Hello, Zoë!"}); len(errs) > 0 {
		t.Errorf("one line refused: %v", errs)
	}

	if SingleLine("a\xff") || !SingleLine("") || !SingleLine("plain") {
		t.Error("SingleLine")
	}
}

func TestSchemaSingleLine(t *testing.T) {
	schema := Schema{Fields: []Spec{
		{Name: "subject", SingleLine: true, Messages: map[string]string{"line": "One line, please"}},
		{Name: "text", Type: types.FieldTypeTextArea},
	}}

	errs := NewForm().ValidateSchemaLocalized(schema, url.Values{"subject": {"a\nb"}, "text": {"a\nb"}}, &DefaultLocalizer{})

	want := FieldErrors{FieldValidationError{Field: "subject", Rule: "line", Err: "One line, please"}}
	if !reflect.DeepEqual(errs, want) {
		t.Errorf("got %#v", errs)
	}

	errs = NewForm().ValidateSchemaLocalized(schema, url.Values{"text": {"\xff"}}, &DefaultLocalizer{})
	if len(errs) != 1 || errs[0].(FieldValidationError).Rule != "invalid" {
		t.Errorf("invalid UTF-8 in a schema: %v", errs)
	}
}

func TestValidURLExported(t *testing.T) {
	cases := []struct {
		rule, value string
		valid       bool
	}{
		{"", "https://example.com", true},
		{"", "/path", true},
		{"", "javascript:alert(1)", false},
		{"http,https", "ftp://example.com", false},
		{"link", "#top", true},
		{"link", "//evil.example", false},
	}

	for _, c := range cases {
		if got := ValidURL(c.rule, c.value); got != c.valid {
			t.Errorf("ValidURL(%q, %q) = %v", c.rule, c.value, got)
		}
	}
}

func TestFieldErrorRule(t *testing.T) {
	type address struct {
		City string `name:"city" required:"true"`
	}

	type input struct {
		Name    string  `required:"true"`
		Email   string  `                form:"input,email"`
		Slug    string  `                                   pattern:"^[a-z]+$"`
		Address address `                                                      name:"address"`
	}

	errs := NewForm().ValidateForm(&input{Email: "nope", Slug: "A"})

	rules := map[string]string{}

	for _, err := range errs {
		validationErr := err.(FieldValidationError)
		rules[validationErr.Field] = validationErr.Rule
	}

	want := map[string]string{"Name": "required", "Email": "email", "Slug": "pattern", "address.city": "required"}
	if !reflect.DeepEqual(rules, want) {
		t.Errorf("got %v, want %v", rules, want)
	}
}

func TestLocalizedValidation(t *testing.T) {
	renderer := NewTranslatedForm(func(loc types.Localizer, key string, args ...any) string {
		return loc.GetLocale() + ":" + key
	})

	slug := func(v any, field reflect.StructField, loc Localizer) FieldErrors {
		if s, _ := v.(string); strings.ToLower(s) != s {
			return FieldErrors{FieldValidationError{Field: field.Name, Err: "Use lowercase letters"}}
		}

		return nil
	}

	renderer.RegisterLocalizedValidation("slug", slug)

	type input struct {
		Slug  string `name:"slug" validate:"slug"`
		Other string `            validate:"slug"  slugMsg:"Lowercase only"`
		Plain string `            validate:"plain"`
	}

	renderer.RegisterValidationMethod("plain", func(v any, field reflect.StructField) FieldErrors {
		return FieldErrors{FieldValidationError{Field: field.Name, Err: "as written"}}
	})

	errs := renderer.ValidateFormLocalized(&input{Slug: "A", Other: "B"}, &DefaultLocalizer{})

	want := FieldErrors{
		FieldValidationError{Field: "slug", Rule: "slug", Err: "en:Use lowercase letters"},
		FieldValidationError{Field: "Other", Rule: "slug", Err: "en:Lowercase only"},
		FieldValidationError{Field: "Plain", Rule: "plain", Err: "as written"},
	}
	if !reflect.DeepEqual(errs, want) {
		t.Errorf("got %#v, want %#v", errs, want)
	}
}

// TestHoneypotThemeClass checks that a theme with a Honeypot class hides
// the spam trap with that class alone, without a style attribute that a
// Content-Security-Policy would ignore, and that a theme without one gets
// HoneypotClass and the inline style, so it is hidden without CSS.
func TestHoneypotThemeClass(t *testing.T) {
	cases := []struct {
		name    string
		classes templates.ThemeClasses
		want    string
	}{
		{
			name:    "honeypot-class-test",
			classes: templates.ThemeClasses{Honeypot: templates.StyleOption{Class: "trap"}},
			want:    `aria-hidden="true" class="trap">`,
		},
		{
			name:    "honeypot-none-test",
			classes: templates.ThemeClasses{Input: templates.StyleOption{Class: "control"}},
			want:    `aria-hidden="true" class="form-honeypot" style="` + templates.HoneypotStyle + `">`,
		},
		{
			name:    "honeypot-style-test",
			classes: templates.ThemeClasses{Honeypot: templates.StyleOption{Style: "position:fixed;top:-100px"}},
			want:    `aria-hidden="true" class="form-honeypot" style="position:fixed;top:-100px">`,
		},
	}

	for _, tc := range cases {
		if _, err := templates.RegisterClassTheme(tc.name, tc.classes); err != nil {
			t.Fatal(err)
		}

		renderer := NewForm()
		renderer.SetTheme(tc.name)

		html, err := renderer.Honeypot(nil, "website")
		if err != nil {
			t.Fatal(err)
		}

		if !strings.HasPrefix(string(html), `<input type="text" name="website" value="" tabindex="-1" autocomplete="off" `) ||
			!strings.HasSuffix(string(html), tc.want) {
			t.Errorf("%s: %s, want it to end with %s", tc.name, html, tc.want)
		}
	}

	bootstrap, err := NewForm().Honeypot(nil, "website")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(bootstrap), `class="visually-hidden"`) || strings.Contains(string(bootstrap), "style=") {
		t.Errorf("bootstrap honeypot: %s", bootstrap)
	}
}

// TestHoneypotPlainTheme checks the spam trap of the plain theme, which
// has no Honeypot class: HoneypotClass with the inline style
// HoneypotStyle, so it is hidden without any CSS, and an escaped name.
func TestHoneypotPlainTheme(t *testing.T) {
	renderer := NewForm()
	renderer.SetTheme("plain")
	honeypot := renderer.FuncMap()["form_honeypot"].(func(string) (template.HTML, error))

	html, err := honeypot(`web"site`)
	if err != nil {
		t.Fatal(err)
	}

	for _, part := range []string{`type="text"`, `name="web&#34;site"`, `tabindex="-1"`, `aria-hidden="true"`, `autocomplete="off"`, `class="form-honeypot"`, `style="position:absolute;left:-10000px;`} {
		if !strings.Contains(string(html), part) {
			t.Errorf("%s misses %s", html, part)
		}
	}
}

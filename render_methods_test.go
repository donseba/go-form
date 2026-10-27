package form

import (
	"net/url"
	"strings"
	"testing"
)

type methodsInput struct {
	Name string `form:"input,text" name:"name" label:"Name"`
}

func TestRenderMethodsMatchTemplateFunctions(t *testing.T) {
	f := NewTranslatedForm(optionTranslate)
	f.SetTheme("plain")

	model := WithInfo(methodsInput{}, Info{Target: "/send", Method: "post", Honeypot: "website"})
	loc := testLocalizer{Locale: "it"}

	render, err := f.Render(loc, model, nil)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	want, _ := f.formRenderLocalized(loc, model, nil)
	if render != want {
		t.Errorf("Render differs from form_render_localized:\n%s\n%s", render, want)
	}

	field, err := f.Field(loc, model, "name", FieldErrors{FieldValidationError{Field: "name", Err: "Too long"}})
	if err != nil {
		t.Fatalf("Field: %v", err)
	}

	if !strings.Contains(string(field), "T(Name)") || !strings.Contains(string(field), "Too long") {
		t.Errorf("Field is not translated or misses its error: %s", field)
	}

	open, err := f.Open(loc, model)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if !strings.Contains(string(open), `action="/send"`) || !strings.Contains(string(open), `name="website"`) {
		t.Errorf("Open misses the action or the honeypot: %s", open)
	}

	buttons, err := f.Buttons(loc, model)
	if err != nil {
		t.Fatalf("Buttons: %v", err)
	}

	if !strings.Contains(string(buttons), "T(Submit)") {
		t.Errorf("Buttons is not translated: %s", buttons)
	}

	closing, err := f.Close()
	if err != nil || !strings.Contains(string(closing), "</form>") {
		t.Errorf("Close = %q, %v", closing, err)
	}

	trap, err := f.Honeypot(loc, "website")
	if err != nil || !strings.Contains(string(trap), `name="website"`) {
		t.Errorf("Honeypot = %q, %v", trap, err)
	}
}

func TestRenderMethodsWithoutLocalizer(t *testing.T) {
	f := NewForm()
	f.SetTheme("plain")

	html, err := f.Field(nil, methodsInput{}, "name", nil)
	if err != nil || !strings.Contains(string(html), "Name") {
		t.Errorf("Field without a localizer = %q, %v", html, err)
	}

	if _, err := f.Open(nil, methodsInput{}); err == nil {
		t.Error("Open of a model without form metadata did not fail")
	}
}

func TestInfoHiddenRendersAfterHoneypot(t *testing.T) {
	f := NewForm()
	f.SetTheme("plain")

	info := Info{
		Target:    "/send",
		Method:    "post",
		Honeypot:  "website",
		CsrfValue: "token",
		Hidden:    url.Values{"back": {`/news?a=1&b="2"`}, "anchor": {"one", "two"}},
	}

	for name, render := range map[string]func() (string, error){
		"open": func() (string, error) {
			html, err := f.Open(nil, WithInfo(methodsInput{}, info))
			return string(html), err
		},
		"render": func() (string, error) {
			html, err := f.Render(nil, WithInfo(methodsInput{}, info), nil)
			return string(html), err
		},
	} {
		html, err := render()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}

		csrf := strings.Index(html, `value="token"`)
		trap := strings.Index(html, `name="website"`)
		anchor := strings.Index(
			html,
			`<input type="hidden" name="anchor" value="one"><input type="hidden" name="anchor" value="two">`,
		)
		back := strings.Index(html, `<input type="hidden" name="back" value="/news?a=1&amp;b=&#34;2&#34;">`)

		if csrf < 0 || trap < csrf || anchor < trap || back < anchor {
			t.Errorf("%s: want CSRF, honeypot, then hidden inputs in name order, escaped:\n%s", name, html)
		}
	}
}

// A field the model does not have is named with the fields it has, so a
// template's typing mistake shows the right name.
func TestFieldNamesTheFieldsOfTheModel(t *testing.T) {
	type address struct {
		City string `form:"input,text" name:"city" label:"City"`
	}

	type person struct {
		Name    string  `form:"input,text"  name:"name"    label:"Name"`
		Email   string  `form:"input,email" name:"email"   label:"Email"`
		Address address `form:"group"       name:"address" label:"Address"`
	}

	f := NewForm()
	f.SetTheme("plain")

	_, err := f.Field(nil, person{}, "nmae", nil)
	if err == nil {
		t.Fatal("Field rendered a field the model does not have")
	}

	want := `form: no field "nmae" (fields: name, email, address, address.city)`
	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err, want)
	}
}

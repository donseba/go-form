package form

import (
	"html/template"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type contactModel struct {
	Name    string                `form:"input,text"  label:"Name"   required:"true"`
	Email   string                `form:"input,email" label:"Email"`
	Address addressModel          `                                                  legend:"Address"`
	Upload  *multipart.FileHeader `form:"input,file"  label:"Upload"                                  accept:".pdf"`
}

type addressModel struct {
	City string `name:"city" form:"input,text" label:"City"`
}

// execute renders the page template text with the form's functions.
func execute(t *testing.T, f *Form, text string, data any) string {
	t.Helper()

	page := template.Must(template.New("page").Funcs(f.FuncMap()).Parse(text))

	var out strings.Builder
	if err := page.Execute(&out, data); err != nil {
		t.Fatalf("execute: %v", err)
	}

	return out.String()
}

func TestFormFieldByField(t *testing.T) {
	f := NewForm()
	model := WithInfo(
		contactModel{Name: "Ada"},
		Info{Target: "/contact", Method: "post", CsrfValue: "token", Honeypot: "website", SubmitText: "Send"},
	)
	errs := FieldErrors{FieldValidationError{Field: "Email", Err: "Enter a valid email address."}}

	html := execute(
		t,
		f,
		`{{ form_open .Model }}<div class="row">{{ form_field .Model "Name" .Errors }}{{ form_field .Model "Email" .Errors }}</div>{{ form_field .Model "Address.city" .Errors }}{{ range form_errors .Errors "Email" }}<p>{{ . }}</p>{{ end }}{{ form_buttons .Model }}{{ form_close }}`,
		map[string]any{"Model": model, "Errors": errs},
	)

	for _, want := range []string{
		`action="/contact"`,
		`enctype="multipart/form-data"`,
		`name="_csrf" value="token"`,
		`name="website"`,
		`<div class="row">`,
		`value="Ada"`,
		`id="Address.city"`,
		`aria-invalid="true"`,
		`<p>Enter a valid email address.</p>`,
		`type="submit"`,
		`</form>`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %s in %s", want, html)
		}
	}

	if strings.Contains(html, `id="Upload"`) {
		t.Errorf("a field that was not asked for rendered: %s", html)
	}

	if strings.Index(html, `<form`) > strings.Index(html, `name="website"`) {
		t.Errorf("hidden fields not inside the form: %s", html)
	}
}

func TestFormFieldUnknown(t *testing.T) {
	f := NewForm()
	if _, err := f.formField(contactModel{}, "Missing", nil); err == nil {
		t.Error("an unknown field rendered without an error")
	}

	if _, err := f.formOpen(contactModel{}); err == nil {
		t.Error("form_open without form metadata rendered without an error")
	}
}

func TestFormRenderHoneypotAndEnctype(t *testing.T) {
	html := renderString(t, NewForm(), WithInfo(contactModel{}, Info{Target: "/", Method: "post", Honeypot: "website"}), nil)

	if !strings.Contains(html, `<input type="text" name="website" value="" tabindex="-1" autocomplete="off" aria-hidden="true"`) {
		t.Errorf("honeypot missing: %s", html)
	}

	if !strings.Contains(html, `enctype="multipart/form-data"`) {
		t.Errorf("form with a file input without multipart encoding: %s", html)
	}

	plain := renderString(t, NewForm(), WithInfo(readOnlyModel{}, Info{Target: "/", Method: "post"}), nil)
	if strings.Contains(plain, "enctype") || strings.Contains(plain, "tabindex") {
		t.Errorf("form without files or honeypot changed: %s", plain)
	}

	explicit := renderString(
		t,
		NewForm(),
		WithInfo(contactModel{}, Info{Target: "/", Method: "post", Attributes: map[string]string{"enctype": "text/plain"}}),
		nil,
	)
	if !strings.Contains(explicit, `enctype="text/plain"`) || strings.Contains(explicit, "multipart") {
		t.Errorf("explicit enctype replaced: %s", explicit)
	}
}

func TestIsSpam(t *testing.T) {
	post := func(values url.Values) *http.Request {
		r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		return r
	}

	if IsSpam(post(url.Values{"website": {""}, "name": {"Ada"}}), "website") {
		t.Error("empty honeypot counted as spam")
	}

	if !IsSpam(post(url.Values{"website": {"http://spam.example"}}), "website") {
		t.Error("filled honeypot not counted as spam")
	}
}

func TestFieldErrorsMessages(t *testing.T) {
	errs := FieldErrors{
		FieldValidationError{Field: "Email", Err: "a"},
		FieldValidationError{Field: "Name", Err: "b"},
		FieldValidationError{Field: "Email", Err: "c"},
	}

	if got := strings.Join(errs.Messages("Email"), ","); got != "a,c" {
		t.Errorf("messages of Email: %q", got)
	}

	if errs.Messages("City") != nil {
		t.Error("messages for a field without errors")
	}
}

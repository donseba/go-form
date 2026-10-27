package form

import (
	"net/url"
	"strings"
	"testing"

	"github.com/donseba/go-form/v2/types"
)

var contactSchema = Schema{Fields: []Spec{
	{Name: "name", Label: "Name", Required: true, MaxLength: 5, AutoComplete: "name"},
	{Name: "email", Label: "Email", InputType: types.InputFieldTypeEmail, Required: true},
	{
		Name:        "topic",
		Label:       "Topic",
		Type:        types.FieldTypeDropdown,
		Placeholder: "—",
		Options:     []types.FieldValue{{Value: "sales", Name: "Sales"}, {Value: "help", Name: "Help"}},
	},
	{Name: "guests", Label: "Guests", InputType: types.InputFieldTypeNumber, Min: "1", Max: "4"},
	{Name: "day", Label: "Day", InputType: types.InputFieldTypeDate},
	{Name: "code", Label: "Code", Pattern: "^[0-9]+$", Messages: map[string]string{"pattern": "Use digits"}},
	{Name: "message", Label: "Message", Type: types.FieldTypeTextArea, Rows: 4},
	{Name: "consent", Label: "I agree", Type: types.FieldTypeCheckbox, Required: true},
	{
		Name:    "days",
		Label:   "Days",
		Type:    types.FieldTypeMultiCheckbox,
		Options: []types.FieldValue{{Value: "mon", Name: "Monday"}, {Value: "tue", Name: "Tuesday"}},
	},
}}

func TestMapValues(t *testing.T) {
	r := postForm(url.Values{"name": {"  Ada "}, "consent": {"on"}, "days": {"mon", "tue"}, "other": {"x"}})
	values := MapValues(r, contactSchema)

	if values.Get("name") != "Ada" || values.Get("consent") != "on" || len(values["days"]) != 2 {
		t.Errorf("values: %v", values)
	}

	if _, ok := values["other"]; ok {
		t.Error("a value outside the schema was mapped")
	}
}

func TestValidateSchemaLocalized(t *testing.T) {
	f := NewForm()

	valid := url.Values{
		"name":    {"Ada"},
		"email":   {"ada@example.com"},
		"topic":   {"help"},
		"guests":  {"2"},
		"day":     {"2026-10-06"},
		"code":    {"42"},
		"consent": {"on"},
		"days":    {"tue"},
	}
	if errs := f.ValidateSchemaLocalized(contactSchema, valid, &DefaultLocalizer{}); len(errs) != 0 {
		t.Errorf("valid values refused: %v", errs)
	}

	invalid := url.Values{
		"name":   {"Adelaide"},
		"email":  {"ada"},
		"topic":  {"spam"},
		"guests": {"9"},
		"day":    {"tomorrow"},
		"code":   {"x"},
		"days":   {"sun"},
	}
	errs := f.ValidateSchemaLocalized(contactSchema, invalid, &DefaultLocalizer{})

	want := map[string]string{
		"name":    "Use at most 5 characters.",
		"email":   "Enter a valid email address.",
		"topic":   "The value 'spam' is not allowed.",
		"guests":  "Enter a value of at most 4.",
		"day":     "The value 'tomorrow' is not allowed.",
		"code":    "Use digits",
		"consent": "This field is required.",
		"days":    "The value 'sun' is not allowed.",
	}

	for name, message := range want {
		if got := strings.Join(errs.Messages(name), ";"); got != message {
			t.Errorf("%s: got %q, want %q", name, got, message)
		}
	}

	spaces := url.Values{"name": {"   "}}
	if errs := f.ValidateSchemaLocalized(contactSchema, spaces, &DefaultLocalizer{}); len(errs.Messages("name")) != 1 {
		t.Errorf("spaces only accepted as a required name: %v", errs)
	}
}

func TestRenderSchema(t *testing.T) {
	values := url.Values{"name": {"Ada"}, "topic": {"help"}, "consent": {"on"}, "days": {"tue"}, "message": {"Hi"}}
	errs := FieldErrors{FieldValidationError{Field: "email", Err: "Enter a valid email address."}}
	model := WithInfo(contactSchema.Bind(values), Info{Target: "/contact", Method: "post", SubmitText: "Send"})

	html := renderString(t, NewForm(), model, errs)

	for _, want := range []string{
		`action="/contact"`,
		`value="Ada"`,
		`maxlength="5"`,
		`autocomplete="name"`,
		`type="email"`,
		`<option value="help" selected`,
		`<option value="" >—</option>`,
		`type="number"`,
		`pattern="^[0-9]&#43;$"`,
		`>Hi</textarea>`,
		`Enter a valid email address.`,
		`aria-invalid="true"`,
		`Send`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %s in %s", want, html)
		}
	}

	if !strings.Contains(control(t, html, "consent"), "checked") || !strings.Contains(control(t, html, "days_1"), "checked") ||
		strings.Contains(control(t, html, "days_0"), "checked") {
		t.Errorf("checkboxes: %s", html)
	}

	field := execute(t, NewForm(), `{{ form_field .Model "email" .Errors }}`, map[string]any{"Model": model, "Errors": errs})
	if !strings.Contains(field, `id="email"`) || strings.Contains(field, `id="name"`) {
		t.Errorf("form_field of a schema: %s", field)
	}

	if !HasFields(contactSchema) || HasFields(Schema{}) {
		t.Error("HasFields of a schema")
	}
}

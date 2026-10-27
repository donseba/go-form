package form

import (
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

// mailRule is a reusable rule as an application registers it: an address
// whose domain has a dot, with a message that is a translation key.
func mailRule(value any, field reflect.StructField, _ Localizer) FieldErrors {
	address, _ := value.(string)

	_, domain, _ := strings.Cut(address, "@")
	if address == "" || strings.Contains(domain, ".") {
		return nil
	}

	return FieldErrors{FieldValidationError{Field: field.Name, Err: "Enter an address with a domain, such as name@example.com"}}
}

func TestLocalizedRuleIsReusable(t *testing.T) {
	catalogues := map[string]map[string]string{
		"nl": {
			"Enter an address with a domain, such as name@example.com": "Vul een adres met een domein in, zoals naam@voorbeeld.nl",
		},
	}

	f := NewTranslatedForm(func(loc Localizer, key string, args ...any) string {
		if text, ok := catalogues[loc.GetLocale()][key]; ok {
			return text
		}

		return key
	})
	f.RegisterLocalizedValidation("mail", mailRule)

	type subscribe struct {
		Email string `form:"input,email" name:"email" validate:"mail"`
	}

	type comment struct {
		Address string `form:"input,email" name:"address" validate:"mail" mailMsg:"Use your own address"`
	}

	var in subscribe

	errs, err := f.Bind(
		postForm(url.Values{"email": {" jane@localhost "}}),
		&in,
		MapOptions{TrimSpace: true},
		testLocalizer{Locale: "nl"},
	)
	if err != nil {
		t.Fatal(err)
	}

	want := FieldErrors{
		FieldValidationError{Field: "email", Rule: "mail", Err: "Vul een adres met een domein in, zoals naam@voorbeeld.nl"},
	}
	if !reflect.DeepEqual(errs, want) {
		t.Errorf("Dutch: %#v", errs)
	}

	errs = f.ValidateFormLocalized(&subscribe{Email: "jane@localhost"}, testLocalizer{Locale: "en"})
	if errs.Messages("email")[0] != "Enter an address with a domain, such as name@example.com" {
		t.Errorf("English: %v", errs)
	}

	errs = f.ValidateFormLocalized(&comment{Address: "jane@localhost"}, testLocalizer{Locale: "nl"})
	if len(errs) != 1 || errs.Messages("address")[0] != "Use your own address" {
		t.Errorf("message per field: %v", errs)
	}

	if errs := f.ValidateForm(&subscribe{Email: "jane@example.com"}); len(errs) != 0 {
		t.Errorf("valid address: %v", errs)
	}
}

func TestOwnMessagesInTheLocalizersLanguage(t *testing.T) {
	type input struct {
		Name string `form:"input,text" name:"name" required:"true"`
	}

	cases := []struct {
		name   string
		form   *Form
		locale string
		want   string
	}{
		{name: "no translation function", form: NewForm(), locale: "nl", want: "Dit veld is verplicht."},
		{name: "region", form: NewForm(), locale: "nl-BE", want: "Dit veld is verplicht."},
		{name: "no catalogue", form: NewForm(), locale: "it", want: "This field is required."},
		{
			name: "missing from the application's catalogue",
			form: NewTranslatedForm(func(_ Localizer, key string, _ ...any) string {
				return key
			}),
			locale: "nl_NL",
			want:   "Dit veld is verplicht.",
		},
		{
			name: "the application's translation wins",
			form: NewTranslatedForm(func(_ Localizer, key string, _ ...any) string {
				return "Verplicht!"
			}),
			locale: "nl",
			want:   "Verplicht!",
		},
	}

	for _, c := range cases {
		errs := c.form.ValidateFormLocalized(&input{}, testLocalizer{Locale: c.locale})
		if got := errs.Messages("name"); len(got) != 1 || got[0] != c.want {
			t.Errorf("%s: %v, want %q", c.name, got, c.want)
		}
	}

	// Arguments are filled in, and labels go through translate as well.
	limited := NewForm().validationErrorTranslated(testLocalizer{Locale: "nl"}, TranslationKeyMaxLength, 5)
	if !strings.Contains(limited, "5") || strings.Contains(limited, "characters") {
		t.Errorf("max length in Dutch: %q", limited)
	}

	if got := NewForm().translate(testLocalizer{Locale: "nl"}, "Cancel", nil); got != "Annuleren" {
		t.Errorf("Cancel in Dutch: %q", got)
	}
}

func TestWallClockField(t *testing.T) {
	type event struct {
		Start time.Time `form:"input,datetime-local" name:"start" wallclock:"true"`
		Sent  time.Time `form:"input,datetime-local" name:"sent"`
	}

	amsterdam, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		t.Skip("no time zone data:", err)
	}

	var in event

	values := url.Values{"start": {"2026-10-04T19:30"}, "sent": {"2026-10-04T19:30"}}
	if err := MapFormWith(postForm(values), &in, MapOptions{Location: amsterdam}); err != nil {
		t.Fatal(err)
	}

	if !in.Start.Equal(time.Date(2026, 10, 4, 19, 30, 0, 0, time.UTC)) {
		t.Errorf("wall clock read in a zone: %v", in.Start)
	}

	if got := in.Sent.UTC().Format("15:04"); got != "17:30" {
		t.Errorf("other field not read in the zone: %s", got)
	}

	html, err := NewForm().Render(zonedLocalizer{amsterdam}, in, nil)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Count(string(html), `value="2026-10-04T19:30"`) != 2 {
		t.Errorf("not shown as entered: %s", html)
	}

	// The wall-clock time becomes an instant in the zone of the event.
	if got := WallTime(in.Start, amsterdam).UTC().Format("15:04"); got != "17:30" {
		t.Errorf("WallTime = %s", got)
	}
}

// TestSchemaValidateRunsRegisteredRules checks that Spec.Validate checks a
// field of a Schema with registered rules, as the validate tag does for a
// struct: with the rule's translated message and Rule, its Messages
// override, and after go-form's own rules.
func TestSchemaValidateRunsRegisteredRules(t *testing.T) {
	catalogues := map[string]map[string]string{
		"nl": {
			"Enter an address with a domain, such as name@example.com": "Vul een adres met een domein in, zoals naam@voorbeeld.nl",
		},
	}

	f := NewTranslatedForm(func(loc Localizer, key string, args ...any) string {
		if text, ok := catalogues[loc.GetLocale()][key]; ok {
			return text
		}

		return key
	})
	f.RegisterLocalizedValidation("mail", mailRule)

	schema := Schema{Fields: []Spec{
		{Name: "email", InputType: "email", Validate: "mail"},
		{Name: "reply", InputType: "email", Validate: "mail", Messages: map[string]string{"mail": "Use your own address"}},
	}}

	errs := f.ValidateSchemaLocalized(
		schema,
		url.Values{"email": {"jane@localhost"}, "reply": {"jane@localhost"}},
		testLocalizer{Locale: "nl"},
	)

	want := FieldErrors{
		FieldValidationError{Field: "email", Rule: "mail", Err: "Vul een adres met een domein in, zoals naam@voorbeeld.nl"},
		FieldValidationError{Field: "reply", Rule: "mail", Err: "Use your own address"},
	}
	if !reflect.DeepEqual(errs, want) {
		t.Errorf("errors = %#v, want %#v", errs, want)
	}

	if errs := f.ValidateSchemaLocalized(
		schema,
		url.Values{"email": {"jane@example.com"}},
		testLocalizer{Locale: "nl"},
	); len(
		errs,
	) != 0 {
		t.Errorf("valid address refused: %v", errs)
	}
}

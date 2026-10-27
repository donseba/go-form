package form

import (
	"math"
	"net/url"
	"strings"
	"testing"

	"github.com/donseba/go-form/v2/types"
)

func TestValidEmail(t *testing.T) {
	valid := []string{"name@example.com", "first.last+tag@sub.example.org", "a@b"}
	invalid := []string{
		"name",
		"name@",
		"@example.com",
		"Name <name@example.com>",
		"<name@example.com>",
		"a b@example.com",
		strings.Repeat("a", 250) + "@example.com",
	}

	for _, address := range valid {
		if !ValidEmail(address) {
			t.Errorf("%q refused", address)
		}
	}

	for _, address := range invalid {
		if ValidEmail(address) {
			t.Errorf("%q accepted", address)
		}
	}

	type Input struct {
		Email  string `form:"input,email" label:"Email"`
		Backup string `form:"input,text"  label:"Backup" email:"true"`
		Copies string `form:"input,email" label:"Copies"              multiple:"true"`
	}

	errs := NewForm().ValidateForm(&Input{Email: "Name <name@example.com>", Backup: "nope", Copies: "a@example.com, b@example.com"})
	if len(errs.Messages("Email")) != 1 || len(errs.Messages("Backup")) != 1 || len(errs.Messages("Copies")) != 0 {
		t.Errorf("email errors: %v", errs)
	}
}

func TestValidateURLRules(t *testing.T) {
	cases := []struct {
		rule, value string
		valid       bool
	}{
		{"true", "https://example.com/x", true},
		{"true", "/path", true},
		{"true", "javascript:alert(1)", false},
		{"http,https", "https://example.com", true},
		{"http,https", "HTTP://example.com", true},
		{"http,https", "ftp://example.com", false},
		{"http,https", "https:example.com", false},
		{"http,https", "/path", false},
		{"http,https,mailto", "mailto:a@example.com", true},
		{"link", "/en/news", true},
		{"link", "#contact", true},
		{"link", "https://example.com", true},
		{"link", "mailto:a@example.com", true},
		{"link", "tel:+31201234567", true},
		{"link", "//evil.example", false},
		{"link", "javascript:alert(1)", false},
		{"link", "data:text/html,x", false},
		{"link", "ftp://example.com", false},
		{"link", "https://exa mple.com", false},
		{"link", "mailto:", false},
	}

	for _, c := range cases {
		if got := ValidURL(c.rule, c.value); got != c.valid {
			t.Errorf("url:%q %q: got %v, want %v", c.rule, c.value, got, c.valid)
		}
	}
}

func TestValidateEqual(t *testing.T) {
	type Input struct {
		Password string `form:"input,password" label:"Password"`
		Repeat   string `form:"input,password" label:"Repeat"   equal:"Password"`
		Code     string `form:"input,text"                                       name:"code"`
		Again    string `form:"input,text"                      equal:"code"                 equalMsg:"Type the same code twice"`
	}

	errs := NewForm().ValidateForm(&Input{Password: "secret", Repeat: "secret", Code: "a", Again: "a"})
	if len(errs) != 0 {
		t.Errorf("equal values refused: %v", errs)
	}

	errs = NewForm().ValidateForm(&Input{Password: "secret", Repeat: "other", Code: "a", Again: "b"})
	if got := strings.Join(errs.Messages("Repeat"), ";"); got != "The value must match Password." {
		t.Errorf("repeat: %q", got)
	}

	if got := strings.Join(errs.Messages("Again"), ";"); got != "Type the same code twice" {
		t.Errorf("again: %q", got)
	}
}

func TestPerRuleMessages(t *testing.T) {
	type Input struct {
		Slug string `form:"input,text" required:"true" requiredMsg:"Choose an address" pattern:"^[a-z-]+$" patternMsg:"Use lowercase letters and dashes" minLength:"3" minLengthMsg:"Use %d characters or more"`
		Old  string `form:"input,text" required:"true"                                 pattern:"^[a-z]+$"                                                                                                       errorMsg:"Wrong"`
	}

	errs := NewForm().ValidateForm(&Input{})
	if got := strings.Join(
		errs.Messages("Slug"),
		";",
	); got != "Choose an address;Use 3 characters or more;Use lowercase letters and dashes" {
		t.Errorf("required: %q", got)
	}

	errs = NewForm().ValidateForm(&Input{Slug: "AB", Old: "A"})
	if got := strings.Join(errs.Messages("Slug"), ";"); got != "Use 3 characters or more;Use lowercase letters and dashes" {
		t.Errorf("rules: %q", got)
	}

	// errorMsg still replaces the message of every rule.
	if got := strings.Join(errs.Messages("Old"), ";"); got != "Wrong" {
		t.Errorf("errorMsg: %q", got)
	}

	// Per-rule messages are text of the application, so they are
	// translated.
	translated := NewTranslatedForm(optionTranslate).ValidateForm(&Input{Slug: "AB"})
	if got := translated.Messages("Slug")[1]; got != "T(Use lowercase letters and dashes)" {
		t.Errorf("rule message not translated: %q", got)
	}
}

func TestMessagesWithoutContext(t *testing.T) {
	type Input struct {
		Name  string  `form:"input,text"   required:"true"`
		Code  string  `form:"input,text"                   pattern:"^[0-9]+$"`
		Price float64 `form:"input,number"                                    step:"0.5"`
	}

	errs := NewForm().ValidateForm(&Input{Code: "x", Price: 0.3})
	for _, err := range errs {
		_, message := err.FieldError()
		if strings.Contains(message, "form||") {
			t.Errorf("message shows its context: %q", message)
		}

		if strings.Contains(message, "[0-9]") {
			t.Errorf("message shows the pattern: %q", message)
		}
	}

	if got := strings.Join(errs.Messages("Price"), ";"); got != "Enter a multiple of 0.5." {
		t.Errorf("step: %q", got)
	}

	// A translator returning a message untranslated shows its English text.
	untranslated := NewTranslatedForm(func(_ Localizer, key string, _ ...any) string { return key })
	if got := untranslated.ValidateForm(&Input{}).Messages("Name"); len(got) != 1 || got[0] != "This field is required." {
		t.Errorf("untranslated message: %q", got)
	}
}

// stepMessages are the messages of the step rule of form's field X.
func stepMessages(form any) []string {
	var messages []string

	for _, err := range NewForm().ValidateForm(form) {
		if failed, ok := err.(FieldValidationError); ok && failed.Rule == "step" {
			messages = append(messages, failed.Err)
		}
	}

	return messages
}

func TestStepCountsExactly(t *testing.T) {
	type tenths struct {
		X float64 `step:"0.1"`
	}

	type cents struct {
		X float64 `step:"0.01"`
	}

	type fromQuarter struct {
		X float64 `step:"0.5" min:"0.25"`
	}

	type oddFromOne struct {
		X int `step:"2" min:"1"`
	}

	type tenthsFloat32 struct {
		X float32 `step:"0.1"`
	}

	type pairs struct {
		X int64 `step:"2"`
	}

	type pairsUnsigned struct {
		X uint64 `step:"2"`
	}

	type fivesUnsigned struct {
		X uint64 `step:"5"`
	}

	cases := []struct {
		name  string
		form  any
		valid bool
	}{
		{"0.3 on 0.1", &tenths{X: 0.3}, true},
		{"0.7 on 0.1", &tenths{X: 0.7}, true},
		{"-0.3 on 0.1", &tenths{X: -0.3}, true},
		{"0.35 on 0.1", &tenths{X: 0.35}, false},
		{"19.99 on 0.01", &cents{X: 19.99}, true},
		{"19.995 on 0.01", &cents{X: 19.995}, false},
		{"0.75 from 0.25", &fromQuarter{X: 0.75}, true},
		{"1.25 from 0.25", &fromQuarter{X: 1.25}, true},
		{"0.5 from 0.25", &fromQuarter{X: 0.5}, false},
		{"1 from 0.25", &fromQuarter{X: 1}, false},
		{"3 from 1", &oddFromOne{X: 3}, true},
		{"4 from 1", &oddFromOne{X: 4}, false},
		{"float32 0.3", &tenthsFloat32{X: 0.3}, true},
		{"float32 0.7", &tenthsFloat32{X: 0.7}, true},
		{"float32 0.35", &tenthsFloat32{X: 0.35}, false},
		{"largest int64", &pairs{X: math.MaxInt64}, false},
		{"largest int64 less one", &pairs{X: math.MaxInt64 - 1}, true},
		{"smallest int64", &pairs{X: math.MinInt64}, true},
		{"largest uint64 on 2", &pairsUnsigned{X: math.MaxUint64}, false},
		{"largest uint64 less one on 2", &pairsUnsigned{X: math.MaxUint64 - 1}, true},
		{"largest uint64 on 5", &fivesUnsigned{X: math.MaxUint64}, true},
		{"largest uint64 less one on 5", &fivesUnsigned{X: math.MaxUint64 - 1}, false},
	}

	for _, c := range cases {
		messages := stepMessages(c.form)
		if c.valid && len(messages) > 0 {
			t.Errorf("%s refused: %v", c.name, messages)
		}

		if !c.valid && len(messages) != 1 {
			t.Errorf("%s: %v, want one step error", c.name, messages)
		}
	}
}

func TestStepWithoutSteps(t *testing.T) {
	type anyStep struct {
		X float64 `step:"any"`
	}

	type zeroStep struct {
		X float64 `step:"0"`
	}

	type negativeStep struct {
		X float64 `step:"-1"`
	}

	type wordStep struct {
		X float64 `step:"abc"`
	}

	for _, form := range []any{&anyStep{X: 0.123}, &zeroStep{X: 0.123}, &negativeStep{X: 0.123}, &wordStep{X: 0.123}} {
		if messages := stepMessages(form); len(messages) > 0 {
			t.Errorf("%T checked steps: %v", form, messages)
		}
	}
}

func TestStepMessageCountsFromMin(t *testing.T) {
	type fromZero struct {
		X float64 `step:"0.5"`
	}

	type fromStep struct {
		X float64 `step:"0.5" min:"1"`
	}

	type fromQuarter struct {
		X float64 `step:"0.5" min:"0.25"`
	}

	type ownMessage struct {
		X float64 `step:"0.5" min:"0.25" stepMsg:"Use steps of %g"`
	}

	cases := []struct {
		form any
		want string
	}{
		{&fromZero{X: 0.3}, "Enter a multiple of 0.5."},
		{&fromStep{X: 1.3}, "Enter a multiple of 0.5."},
		{&fromQuarter{X: 0.5}, "Enter a value in steps of 0.5 from 0.25."},
		{&ownMessage{X: 0.5}, "Use steps of 0.5"},
	}

	for _, c := range cases {
		if got := strings.Join(stepMessages(c.form), ";"); got != c.want {
			t.Errorf("%T: %q, want %q", c.form, got, c.want)
		}
	}

	errs := NewForm().ValidateFormLocalized(&fromQuarter{X: 0.5}, testLocalizer{Locale: "nl"})
	if got := strings.Join(errs.Messages("X"), ";"); got != "Vul een waarde in stappen van 0.5 vanaf 0.25 in." {
		t.Errorf("Dutch: %q", got)
	}
}

func TestSchemaStepCountsExactly(t *testing.T) {
	schema := Schema{Fields: []Spec{
		{Name: "tenths", InputType: types.InputFieldTypeNumber, Step: "0.1"},
		{Name: "quarters", InputType: types.InputFieldTypeNumber, Min: "0.25", Max: "10", Step: "0.5"},
	}}

	f := NewForm()
	onStep := url.Values{"tenths": {"0.3"}, "quarters": {"0.75"}}

	if errs := f.ValidateSchemaLocalized(schema, onStep, &DefaultLocalizer{}); len(errs) != 0 {
		t.Errorf("values on a step refused: %v", errs)
	}

	errs := f.ValidateSchemaLocalized(schema, url.Values{"tenths": {"0.35"}, "quarters": {"1"}}, &DefaultLocalizer{})
	if len(errs.Messages("tenths")) != 1 ||
		strings.Join(errs.Messages("quarters"), ";") != "Enter a value in steps of 0.5 from 0.25." {
		t.Errorf("values between steps: %v", errs)
	}
}

func TestOnStep(t *testing.T) {
	cases := []struct {
		value, min, step string
		want             bool
	}{
		{"0.3", "", "0.1", true},
		{"-0.3", "", "0.1", true},
		{"0.35", "", "0.1", false},
		{"0.75", "0.25", "0.5", true},
		{"1", "0.25", "0.5", false},
		{"1", "0.5", "0.5", true},
		{"5", "abc", "1", true},
		{"0.123", "", "any", true},
		{"0.123", "", "0", true},
		{"0.123", "", "abc", true},
		{"abc", "", "1", false},
		{"NaN", "", "1", false},
		{"Inf", "", "any", false},
	}

	for _, c := range cases {
		if got := OnStep(c.value, c.min, c.step); got != c.want {
			t.Errorf("OnStep(%q, %q, %q) = %v, want %v", c.value, c.min, c.step, got, c.want)
		}
	}
}

// keysSelect is a multi-select of an application whose keys need not be
// among its options.
type keysSelect struct {
	keys []string
}

func (s keysSelect) String() string {
	return strings.Join(s.keys, ",")
}

func (s keysSelect) SortedMapper() []SortedMap {
	return []SortedMap{ssEntry{"a", "A"}}
}

func (s keysSelect) GetKeysAsStrings() []string {
	return s.keys
}

// failingSelect is a select of an application whose pointer holds an error
// for the form to report.
type failingSelect struct {
	key string
}

func (s failingSelect) String() string {
	return s.key
}

func (s failingSelect) SortedMapper() []SortedMap {
	return []SortedMap{ssEntry{"a", "A"}}
}

func (s *failingSelect) Error() string {
	return "failing select"
}

func (s *failingSelect) Unwrap() error {
	return SortedSelectError{Key: TranslationKeySortedSelectNotFound, Args: []any{"z"}}
}

func TestValidateSortedMapperKeys(t *testing.T) {
	type input struct {
		Tags   keysSelect
		Choice failingSelect
		Other  failingSelect
	}

	model := input{
		Tags:   keysSelect{keys: []string{"a", "", "x", "y"}},
		Choice: failingSelect{key: "a"},
		Other:  failingSelect{key: "q"},
	}

	var got []string

	for _, err := range NewForm().ValidateForm(&model) {
		validationErr := err.(FieldValidationError)
		got = append(got, validationErr.Field+":"+validationErr.Rule+":"+validationErr.Err)
	}

	want := []string{
		"Tags:values:The choice 'x' is not allowed.",
		"Tags:values:The choice 'y' is not allowed.",
		"Choice:values:The choice 'z' is not available.",
		"Other:values:The choice 'q' is not allowed.",
		"Other:values:The choice 'z' is not available.",
	}

	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}

	// A struct passed by value has no address, so its selects hold no
	// error to report.
	if errs := NewForm().ValidateForm(input{Choice: failingSelect{key: "a"}}); len(errs) != 0 {
		t.Errorf("by value: %v", errs)
	}
}

func TestItemsOfMultiSelect(t *testing.T) {
	type input struct {
		Tags SortedMultiSelect[string] `form:"multicheckbox" minItems:"2" maxItems:"3"`
	}

	source := map[string]string{"a": "A", "b": "B", "c": "C", "d": "D"}
	tests := map[string]struct {
		selected []string
		rule     string
	}{
		"too few":  {[]string{"a"}, "minItems"},
		"enough":   {[]string{"a", "b"}, ""},
		"too many": {[]string{"a", "b", "c", "d"}, "maxItems"},
	}

	f := NewForm()

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			errs := f.ValidateForm(&input{Tags: NewSortedMultiSelect(source).Selected(test.selected...)})

			var rules []string

			for _, err := range errs {
				rules = append(rules, err.(FieldValidationError).Rule)
			}

			if strings.Join(rules, ",") != test.rule {
				t.Errorf("rules %v, want %q", rules, test.rule)
			}
		})
	}
}

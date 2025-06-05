package form

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type TestForm struct {
	Username    string `form:"input,text" label:"Username" required:"true"`
	Age         int    `form:"input,number" label:"Age" min:"18" max:"99"`
	Email       string `form:"input,email" label:"Email" required:"true"`
	Description string `form:"textarea" label:"Description" maxLength:"10" minLength:"2"`
}

func TestValidateForm(t *testing.T) {
	cases := []struct {
		name   string
		form   TestForm
		errors int
	}{
		{
			name:   "all valid",
			form:   TestForm{Username: "user", Age: 25, Email: "user@example.com", Description: "valid"},
			errors: 0,
		},
		{
			name:   "missing required",
			form:   TestForm{Username: "", Age: 25, Email: "", Description: "valid"},
			errors: 2,
		},
		{
			name:   "age too low",
			form:   TestForm{Username: "user", Age: 10, Email: "user@example.com", Description: "valid"},
			errors: 1,
		},
		{
			name:   "age too high",
			form:   TestForm{Username: "user", Age: 120, Email: "user@example.com", Description: "valid"},
			errors: 1,
		},
		{
			name:   "invalid email",
			form:   TestForm{Username: "user", Age: 25, Email: "notanemail", Description: "valid"},
			errors: 1,
		},
		{
			name:   "description too long",
			form:   TestForm{Username: "user", Age: 25, Email: "user@example.com", Description: "This description is way too long"},
			errors: 1,
		},
		{
			name:   "description to short",
			form:   TestForm{Username: "user", Age: 25, Email: "user@example.com", Description: "æ"},
			errors: 1,
		},
		{
			name:   "description exactly min length",
			form:   TestForm{Username: "user", Age: 25, Email: "user@example.com", Description: "ab"},
			errors: 0,
		},
	}

	f := NewForm(nil)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errList := f.internalFormValidation(&c.form, &DefaultLocalizer{})
			if len(errList) != c.errors {
				t.Errorf("expected %d errors, got %d: %+v", c.errors, len(errList), errList)
			}
		})
	}
}

func TestValidateForm_Values(t *testing.T) {
	type ValuesForm struct {
		Color string `form:"dropdown" label:"Color" values:"red:Red;green:Green;blue:Blue"`
	}

	cases := []struct {
		name   string
		form   ValuesForm
		errors int
	}{
		{
			name:   "valid value",
			form:   ValuesForm{Color: "red"},
			errors: 0,
		},
		{
			name:   "invalid value",
			form:   ValuesForm{Color: "yellow"},
			errors: 1,
		},
		{
			name:   "empty value",
			form:   ValuesForm{Color: ""},
			errors: 0, // not required, so empty is allowed
		},
	}

	f := NewForm(nil)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errList := f.internalFormValidation(&c.form, &DefaultLocalizer{})
			if len(errList) != c.errors {
				t.Errorf("expected %d errors, got %d: %+v", c.errors, len(errList), errList)
			}
		})
	}
}

func isHexColor(val any, field reflect.StructField) (out FieldErrors) {
	s, ok := val.(string)
	if !ok || s == "" {
		return nil
	}
	if len(s) != 7 || s[0] != '#' {
		out = append(out, FieldValidationError{
			Field: field.Name,
			Err:   "must be a valid hex color (e.g. #aabbcc)",
		})

		return out
	}
	for _, c := range s[1:] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			out = append(out, FieldValidationError{
				Field: field.Name,
				Err:   "must be a valid hex color (e.g. #aabbcc)",
			})
			return out
		}
	}

	return
}

func TestValidateForm_CustomValidation(t *testing.T) {
	f := NewForm(nil)
	f.RegisterValidationMethod("isHexColor", isHexColor)

	type CustomForm struct {
		Color string `form:"input,text" label:"Color" validate:"isHexColor"`
	}

	cases := []struct {
		name   string
		form   CustomForm
		errors int
	}{
		{
			name:   "valid hex color",
			form:   CustomForm{Color: "#aabbcc"},
			errors: 0,
		},
		{
			name:   "invalid hex color - missing #",
			form:   CustomForm{Color: "aabbcc"},
			errors: 1,
		},
		{
			name:   "invalid hex color - wrong length",
			form:   CustomForm{Color: "#abc"},
			errors: 1,
		},
		{
			name:   "invalid hex color - bad char",
			form:   CustomForm{Color: "#aabbcg"},
			errors: 1,
		},
		{
			name:   "empty value",
			form:   CustomForm{Color: ""},
			errors: 0,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errList := f.ValidateForm(&c.form, &DefaultLocalizer{})
			if len(errList) != c.errors {
				t.Errorf("expected %d errors, got %d: %+v", c.errors, len(errList), errList)
			}
		})
	}
}

// --- Enumerator test ---
type ColorEnum string

func (c ColorEnum) Enum() []any {
	return []any{"red", "green", "blue"}
}

func TestValidateForm_Enumerator(t *testing.T) {
	type EnumForm struct {
		Color ColorEnum `form:"dropdown" label:"Color"`
	}
	cases := []struct {
		name   string
		form   EnumForm
		errors int
	}{
		{"valid enum", EnumForm{Color: "red"}, 0},
		{"invalid enum", EnumForm{Color: "yellow"}, 1},
		{"empty enum", EnumForm{Color: ""}, 0},
	}

	f := NewForm(nil)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errList := f.internalFormValidation(&c.form, &DefaultLocalizer{})
			if len(errList) != c.errors {
				t.Errorf("expected %d errors, got %d: %+v", c.errors, len(errList), errList)
			}
		})
	}
}

// --- Mapper test ---
type ColorMap string

func (c ColorMap) Mapper() map[string]string {
	return map[string]string{"r": "Red", "g": "Green", "b": "Blue"}
}

func (c ColorMap) String() string { return string(c) }

func TestValidateForm_Mapper(t *testing.T) {
	type MapForm struct {
		Color ColorMap `form:"dropdown" label:"Color"`
	}
	cases := []struct {
		name   string
		form   MapForm
		errors int
	}{
		{"valid map", MapForm{Color: "r"}, 0},
		{"invalid map", MapForm{Color: "x"}, 1},
		{"empty map", MapForm{Color: ""}, 0},
	}

	f := NewForm(nil)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errList := f.internalFormValidation(&c.form, &DefaultLocalizer{})
			if len(errList) != c.errors {
				t.Errorf("expected %d errors, got %d: %+v", c.errors, len(errList), errList)
			}
		})
	}
}

// --- SortedMapper test ---
type colorPair struct {
	k, v string
}

func (c colorPair) Key() string   { return c.k }
func (c colorPair) Value() string { return c.v }

type ColorSortedMap string

func (c ColorSortedMap) String() string { return string(c) }

func (c ColorSortedMap) SortedMapper() []SortedMap {
	return []SortedMap{
		colorPair{"r", "Red"},
		colorPair{"g", "Green"},
		colorPair{"b", "Blue"},
	}
}

func TestValidateForm_SortedMapper(t *testing.T) {
	type SortedMapForm struct {
		Color ColorSortedMap `form:"dropdown" label:"Color"`
	}
	cases := []struct {
		name   string
		form   SortedMapForm
		errors int
	}{
		{"valid sortedmap", SortedMapForm{Color: "g"}, 0},
		{"invalid sortedmap", SortedMapForm{Color: "x"}, 1},
		{"empty sortedmap", SortedMapForm{Color: ""}, 0},
	}

	f := NewForm(nil)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errList := f.internalFormValidation(&c.form, &DefaultLocalizer{})
			if len(errList) != c.errors {
				t.Errorf("expected %d errors, got %d: %+v", c.errors, len(errList), errList)
			}
		})
	}
}

// --- Translation tests ---
type testLocalizer struct{ Locale string }

func (l testLocalizer) GetLocale() string { return l.Locale }

var testTranslations = map[string]map[string]string{
	"en": {
		"form.validation.required": "is required",
		"form.validation.min":      "must be >= %v",
	},
	"it": {
		"form.validation.required": "è obbligatorio",
		"form.validation.min":      "deve essere almeno %v",
	},
}

func testTranslate(loc Localizer, key string, args ...any) string {
	locale := "en"
	if l, ok := loc.(testLocalizer); ok {
		locale = l.Locale
	}
	msg := key
	if m, ok := testTranslations[locale]; ok {
		if t, ok := m[key]; ok {
			msg = t
		}
	}
	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}

func TestValidateForm_Translation(t *testing.T) {
	type TranslationForm struct {
		Name string `form:"input,text" label:"Name" required:"true"`
		Age  int    `form:"input,number" label:"Age" required:"true" min:"18"`
	}
	f := NewTranslatedForm(nil, testTranslate)

	cases := []struct {
		name   string
		form   TranslationForm
		locale string
		expect string
	}{
		{"required error in EN", TranslationForm{Name: "", Age: 0}, "en", "is required"},
		{"required error in IT", TranslationForm{Name: "", Age: 0}, "it", "è obbligatorio"},
		{"min error in EN", TranslationForm{Name: "John", Age: 10}, "en", ">= 18"},
		{"min error in IT", TranslationForm{Name: "John", Age: 10}, "it", "almeno 18"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errList := f.ValidateForm(&c.form, testLocalizer{Locale: c.locale})
			found := false
			for _, err := range errList {
				_, fieldErr := err.FieldError()
				if strings.Contains(fieldErr, c.expect) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected error containing '%s', got: %+v", c.expect, errList)
			}
		})
	}
}

func TestExtendedValidations(t *testing.T) {
	type ExtendedForm struct {
		PatternField  string  `pattern:"^foo[0-9]+$"`
		URLField      string  `url:"true"`
		BoolField     bool    `bool:"true"`
		ZeroInt       int     `zero:"true"`
		ZeroString    string  `zero:"true"`
		SliceField    []int   `minItems:"2" maxItems:"3"`
		PrefixField   string  `prefix:"abc"`
		SuffixField   string  `suffix:"xyz"`
		ContainsField string  `contains:"mid"`
		StepInt       int     `step:"5"`
		StepFloat     float64 `step:"0.5"`
		CustomMsg     string  `pattern:"^bar$" errorMsg:"custom error!"`
	}

	valid := ExtendedForm{
		PatternField:  "foo123",
		URLField:      "https://example.com",
		BoolField:     true,
		ZeroInt:       0,
		ZeroString:    "",
		SliceField:    []int{1, 2},
		PrefixField:   "abcde",
		SuffixField:   "wxyz",
		ContainsField: "amidb",
		StepInt:       10,
		StepFloat:     2.0,
		CustomMsg:     "bar",
	}

	cases := []struct {
		name   string
		form   ExtendedForm
		errors int
	}{
		{"all valid", valid, 0},
		{"pattern fail", func() ExtendedForm { f := valid; f.PatternField = "bar"; return f }(), 1},
		{"url fail", func() ExtendedForm { f := valid; f.URLField = "notaurl"; return f }(), 1},
		{"bool fail", func() ExtendedForm { f := valid; f.BoolField = false; return f }(), 1},
		{"zero int fail", func() ExtendedForm { f := valid; f.ZeroInt = 1; return f }(), 1},
		{"zero string fail", func() ExtendedForm { f := valid; f.ZeroString = "notzero"; return f }(), 1},
		{"minItems fail", func() ExtendedForm { f := valid; f.SliceField = []int{1}; return f }(), 1},
		{"maxItems fail", func() ExtendedForm { f := valid; f.SliceField = []int{1, 2, 3, 4}; return f }(), 1},
		{"prefix fail", func() ExtendedForm { f := valid; f.PrefixField = "def"; return f }(), 1},
		{"suffix fail", func() ExtendedForm { f := valid; f.SuffixField = "abc"; return f }(), 1},
		{"contains fail", func() ExtendedForm { f := valid; f.ContainsField = "nope"; return f }(), 1},
		{"step int fail", func() ExtendedForm { f := valid; f.StepInt = 7; return f }(), 1},
		{"step float fail", func() ExtendedForm { f := valid; f.StepFloat = 2.1; return f }(), 1},
		{"custom error msg", func() ExtendedForm { f := valid; f.CustomMsg = "baz"; return f }(), 1},
	}

	f := NewForm(nil)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errList := f.internalFormValidation(&c.form, &DefaultLocalizer{})
			if len(errList) != c.errors {
				t.Errorf("expected %d errors, got %d: %+v", c.errors, len(errList), errList)
			}
			if c.name == "custom error msg" {
				if len(errList) == 0 {
					t.Errorf("expected at least one error for custom error msg case")
				} else {
					_, errMsg := errList[0].FieldError()
					if !strings.Contains(errMsg, "custom error!") {
						t.Errorf("expected custom error message, got: %s", errMsg)
					}
				}
			}
		})
	}
}

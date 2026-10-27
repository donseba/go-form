package form

import (
	"errors"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/donseba/go-form/v2/types"
)

type strictInput struct {
	X       float64    `form:"input,number"         min:"0" max:"100"`
	Small   float32    `form:"input,number"`
	N       int        `form:"input,number"         min:"1"           name:"n"`
	Tiny    int8       `form:"input,number"`
	At      *time.Time `form:"input,datetime-local"                   name:"at"`
	Agree   bool       `form:"checkbox"`
	ID      [16]byte   `form:"input,hidden"`
	Price   float64    `form:"input,number"                                          invalidMsg:"Enter an amount such as 12.50"`
	Address struct {
		Zip int `form:"input,number" name:"zip"`
	} `                                              name:"address"`
}

// strictNames are the field names of errs.
func strictNames(errs FieldErrors) []string {
	var names []string

	for _, err := range errs {
		name, _ := err.FieldError()
		names = append(names, name)
	}

	return names
}

func TestMapFormSkipsUnreadableValues(t *testing.T) {
	r := postForm(url.Values{
		"X": {"NaN"}, "Small": {"1e40"}, "n": {"abc"}, "Tiny": {"300"},
		"at": {"garbage"}, "Agree": {"maybe"}, "ID": {"not-a-uuid"},
	})

	input := strictInput{X: 5, N: 2}
	if err := MapForm(r, &input); err != nil {
		t.Fatalf("MapForm without Strict returned %v", err)
	}

	if input.X != 5 || input.N != 2 || input.Tiny != 0 || input.Small != 0 || input.At != nil {
		t.Errorf("unreadable values were set: %+v", input)
	}
}

func TestMapFormStrict(t *testing.T) {
	r := postForm(url.Values{
		"X": {"NaN"}, "Small": {"1e40"}, "n": {"abc"}, "Tiny": {"300"},
		"at": {"2026-13-45T99:99"}, "Agree": {"maybe"}, "ID": {"not-a-uuid"},
		"Price": {"Inf"}, "address.zip": {"12a"},
	})

	var input strictInput
	err := MapFormWith(r, &input, MapOptions{Strict: true})

	var invalid *InvalidValuesError
	if !errors.As(err, &invalid) {
		t.Fatalf("got %v, want an *InvalidValuesError", err)
	}

	want := []string{"X", "Small", "n", "Tiny", "at", "Agree", "ID", "Price", "address.zip"}
	if got := strictNames(invalid.Errors); !reflect.DeepEqual(got, want) {
		t.Errorf("fields %v, want %v", got, want)
	}

	for _, fieldErr := range invalid.Errors {
		validationErr := fieldErr.(FieldValidationError)
		if validationErr.Rule != "invalid" {
			t.Errorf("%s: rule %q", validationErr.Field, validationErr.Rule)
		}

		want := "Enter a valid value."
		if validationErr.Field == "Price" {
			want = "Enter an amount such as 12.50"
		}

		if validationErr.Err != want {
			t.Errorf("%s: message %q, want %q", validationErr.Field, validationErr.Err, want)
		}
	}
}

func TestMapFormStrictStartsNoGoroutines(t *testing.T) {
	before := runtime.NumGoroutine()

	for range 50 {
		var input strictInput

		_ = MapFormWith(postForm(url.Values{"n": {"abc"}}), &input, MapOptions{Strict: true})
	}

	// Each call made a Form with a CSRF store whose goroutine never stops.
	if after := runtime.NumGoroutine(); after > before+5 {
		t.Fatalf("goroutines went from %d to %d", before, after)
	}
}

func TestMapFormStrictMapsTheRest(t *testing.T) {
	r := postForm(url.Values{"X": {"12.5"}, "n": {"x"}, "Agree": {"on"}, "ID": {"6ba7b810-9dad-11d1-80b4-00c04fd430c8"}})

	var input strictInput

	err := MapFormWith(r, &input, MapOptions{Strict: true})
	if err == nil {
		t.Fatal("no error for n")
	}

	if input.X != 12.5 || !input.Agree || input.ID[0] != 0x6b || input.ID[15] != 0xc8 {
		t.Errorf("valid values were not mapped: %+v", input)
	}

	if err := MapFormWith(postForm(url.Values{"X": {"1"}, "n": {"3"}}), &input, MapOptions{Strict: true}); err != nil {
		t.Errorf("valid submission: %v", err)
	}
}

func TestMapFormLocalized(t *testing.T) {
	renderer := NewTranslatedForm(func(loc types.Localizer, key string, args ...any) string {
		return loc.GetLocale() + ":" + key
	})

	errs, err := renderer.MapFormLocalized(
		postForm(url.Values{"n": {"abc"}, "Price": {"x"}}),
		&strictInput{},
		MapOptions{},
		&DefaultLocalizer{},
	)
	if err != nil {
		t.Fatal(err)
	}

	want := FieldErrors{
		FieldValidationError{Field: "n", Rule: "invalid", Err: "en:form||Value is not valid"},
		FieldValidationError{Field: "Price", Rule: "invalid", Err: "en:Enter an amount such as 12.50"},
	}
	if !reflect.DeepEqual(errs, want) {
		t.Errorf("got %#v, want %#v", errs, want)
	}

	if _, err := renderer.MapFormLocalized(
		postForm(nil),
		strictInput{},
		MapOptions{},
		&DefaultLocalizer{},
	); !errors.Is(
		err,
		ErrMapFormNotPointer,
	) {
		t.Errorf("not a pointer: %v", err)
	}
}

func TestValidateRefusesNonFiniteNumbers(t *testing.T) {
	type input struct {
		X float64 `min:"0" max:"100" step:"0.5"`
		Y float32
		Z float64
	}

	errs := NewForm().ValidateForm(&input{X: math.NaN(), Y: float32(math.Inf(1)), Z: 3})

	want := FieldErrors{
		FieldValidationError{Field: "X", Rule: "invalid", Err: "Enter a valid value."},
		FieldValidationError{Field: "Y", Rule: "invalid", Err: "Enter a valid value."},
	}
	if !reflect.DeepEqual(errs, want) {
		t.Errorf("got %#v, want %#v", errs, want)
	}
}

func TestValidateMinMaxOtherNumberKinds(t *testing.T) {
	type input struct {
		A int64   `min:"1"`
		B uint8   `        max:"10"`
		C float32 `        max:"1"`
	}

	errs := NewForm().ValidateForm(&input{A: 0, B: 11, C: 2})
	if got := strictNames(errs); !reflect.DeepEqual(got, []string{"A", "B", "C"}) {
		t.Errorf("got %v", errs)
	}
}

func TestSchemaRefusesNonFiniteNumbers(t *testing.T) {
	schema := Schema{Fields: []Spec{{Name: "n", InputType: types.InputFieldTypeNumber, Max: "10"}}}

	for _, value := range []string{"NaN", "Inf", "-Inf", "+infinity"} {
		errs := NewForm().ValidateSchemaLocalized(schema, url.Values{"n": {value}}, &DefaultLocalizer{})
		if len(errs) != 1 || errs[0].(FieldValidationError).Rule != "values" {
			t.Errorf("%s: %v", value, errs)
		}
	}
}

func TestMapFormCheckboxValues(t *testing.T) {
	type input struct {
		Off   bool
		Zero  bool
		False bool
		Empty bool
		Upper bool
		Short bool
	}

	r := postForm(url.Values{
		"Off": {"off"}, "Zero": {"0"}, "False": {"false"}, "Empty": {""},
		"Upper": {"TRUE"}, "Short": {"t"},
	})

	got := input{Off: true, Zero: true, False: true, Empty: true}

	err := MapFormWith(r, &got, MapOptions{Strict: true})
	if err != nil {
		t.Fatal(err)
	}

	want := input{Upper: true, Short: true}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestMapFormNotStruct(t *testing.T) {
	number := 0

	if err := MapForm(postForm(nil), &number); !errors.Is(err, ErrMapFormNotStruct) {
		t.Errorf("got %v, want ErrMapFormNotStruct", err)
	}
}

// multiInput has a field after its multi-select, which mapping must reach.
type multiInput struct {
	Tags SortedMultiSelect[string]
	Name string `name:"name"`
}

func TestMapFormMultiSelectSkipsUnknownKeys(t *testing.T) {
	values := url.Values{"Tags": {"a", "x"}, "name": {"Ada"}}

	got := multiInput{Tags: NewSortedMultiSelect(map[string]string{"a": "A"}).Selected("a")}
	if err := MapForm(postForm(values), &got); err != nil {
		t.Fatalf("MapForm returned %v", err)
	}

	// As for a SortedSelect, an unknown key leaves nothing selected.
	if keys := got.Tags.GetKeysAsStrings(); len(keys) != 0 || got.Name != "Ada" {
		t.Errorf("got tags %v and name %q, want no tags and Ada", keys, got.Name)
	}

	got = multiInput{Tags: NewSortedMultiSelect(map[string]string{"a": "A"})}
	err := MapFormWith(postForm(values), &got, MapOptions{Strict: true})

	var invalid *InvalidValuesError
	if !errors.As(err, &invalid) || !reflect.DeepEqual(strictNames(invalid.Errors), []string{"Tags"}) {
		t.Fatalf("Strict: got %v, want an invalid value for Tags", err)
	}

	if got.Name != "Ada" {
		t.Errorf("Strict: name %q, want Ada", got.Name)
	}

	got = multiInput{Tags: NewSortedMultiSelect(map[string]string{"a": "A"})}

	errs, err := (&Form{}).Bind(postForm(values), &got, MapOptions{}, nil)
	if err != nil || !reflect.DeepEqual(strictNames(errs), []string{"Tags"}) {
		t.Errorf("Bind: errors %v and %v, want an invalid value for Tags", errs, err)
	}
}

func TestMapFormMultiSelectWithoutSource(t *testing.T) {
	// Nothing ticked is an empty selection, also without options.
	var empty multiInput

	err := MapFormWith(postForm(url.Values{"name": {"Ada"}}), &empty, MapOptions{Strict: true})
	if err != nil || empty.Name != "Ada" {
		t.Errorf("nothing ticked: got %v and name %q, want no error and Ada", err, empty.Name)
	}

	// A key needs an option to select.
	var ticked multiInput

	err = MapFormWith(postForm(url.Values{"Tags": {"a"}, "name": {"Ada"}}), &ticked, MapOptions{Strict: true})

	var invalid *InvalidValuesError
	if !errors.As(err, &invalid) || !reflect.DeepEqual(strictNames(invalid.Errors), []string{"Tags"}) || ticked.Name != "Ada" {
		t.Errorf("a key: got %v and name %q, want an invalid value for Tags and Ada", err, ticked.Name)
	}
}

func TestMapFormMultiSelectOnGet(t *testing.T) {
	type input struct {
		Tags SortedMultiSelect[string]
	}

	source := map[string]string{"a": "A", "b": "B"}
	got := input{Tags: NewSortedMultiSelect(source).Selected("b")}

	// Without values a GET leaves the selection as it is.
	if err := MapForm(&http.Request{Method: http.MethodGet, Form: url.Values{}}, &got); err != nil {
		t.Fatal(err)
	}

	if keys := got.Tags.GetKeysAsStrings(); !reflect.DeepEqual(keys, []string{"b"}) {
		t.Errorf("without values: %v, want [b]", keys)
	}

	if err := MapForm(&http.Request{Method: http.MethodGet, Form: url.Values{"Tags": {"a"}}}, &got); err != nil {
		t.Fatal(err)
	}

	if keys := got.Tags.GetKeysAsStrings(); !reflect.DeepEqual(keys, []string{"a"}) {
		t.Errorf("with a value: %v, want [a]", keys)
	}
}

func TestMapFormUnsupportedArray(t *testing.T) {
	type input struct {
		Codes [4]int
		Name  string
	}

	var reported []string

	MapFormErrorHandler = func(field string, _ error) {
		reported = append(reported, field)
	}

	t.Cleanup(func() {
		MapFormErrorHandler = nil
	})

	got := input{Codes: [4]int{1, 2, 3, 4}}

	err := MapFormWith(postForm(url.Values{"Codes": {"9"}, "Name": {"x"}}), &got, MapOptions{Strict: true})
	if err != nil {
		t.Fatal(err)
	}

	if got.Codes != [4]int{1, 2, 3, 4} || got.Name != "x" {
		t.Errorf("got %+v", got)
	}

	if !reflect.DeepEqual(reported, []string{"Codes"}) {
		t.Errorf("reported %v, want [Codes]", reported)
	}
}

package form

import (
	"fmt"
	"strings"
	"testing"
)

func TestValueSorted_SortedMapper(t *testing.T) {
	type testKey int
	vs := ValueSorted[testKey]{}
	vs.SetSource(map[testKey]string{
		2: "two",
		1: "one",
		3: "three",
	})
	entries := vs.SortedMapper()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	// Should be sorted by key string: "1", "2", "3"
	expected := []struct{ k, v string }{
		{"1", "one"},
		{"2", "two"},
		{"3", "three"},
	}
	for i, e := range entries {
		if e.Key() != expected[i].k || e.Value() != expected[i].v {
			t.Errorf("entry %d: expected (%s,%s), got (%s,%s)", i, expected[i].k, expected[i].v, e.Key(), e.Value())
		}
	}
}

func TestValueSorted_ScanAndValue(t *testing.T) {
	vs := &ValueSorted[int]{}
	vs.SetSource(map[int]string{1: "one"})
	err := vs.Scan(1)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if v, _ := vs.Value(); v != 1 {
		t.Errorf("expected value 1, got %v", v)
	}
}

func TestValueSorted_ScanInvalidType(t *testing.T) {
	vs := &ValueSorted[int]{}
	vs.SetSource(map[int]string{1: "one"})
	err := vs.Scan("not an int")
	if err == nil {
		t.Error("expected error for invalid type, got nil")
	}
}

func TestValueSorted_String(t *testing.T) {
	vs := ValueSorted[string]{value: "foo"}
	if vs.String() != "foo" {
		t.Errorf("expected 'foo', got '%s'", vs.String())
	}
}

func TestValueSorted_SetNotFoundError(t *testing.T) {
	vs := &ValueSorted[int]{source: map[int]string{1: "one"}}
	err := vs.Set(2)
	expected := "form||ValueSorted: value '2' not found in source"
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Errorf("expected error containing '%s', got '%v'", expected, err)
	}
}

func TestValueSorted_SetFromKeyNotFoundError(t *testing.T) {
	vs := &ValueSorted[int]{source: map[int]string{1: "one"}}
	err := vs.SetFromKey("2")
	if err == nil || !strings.Contains(err.Error(), TranslationKeyValueSortedKeyNotFound) {
		t.Errorf("expected error containing '%s', got '%v'", TranslationKeyValueSortedKeyNotFound, err)
	}
}

func TestValueSorted_ScanTypeError(t *testing.T) {
	vs := &ValueSorted[int]{source: map[int]string{1: "one"}}
	err := vs.Scan("not an int")
	if err == nil || !strings.Contains(err.Error(), TranslationKeyValueSortedTypeError) {
		t.Errorf("expected error containing '%s', got '%v'", TranslationKeyValueSortedTypeError, err)
	}
}

func TestValueSorted_ScanNotFoundError(t *testing.T) {
	vs := &ValueSorted[int]{source: map[int]string{1: "one"}}
	err := vs.Scan(2)
	// The error should contain the interpolated value, not the format string
	expected := "form||ValueSorted: value '2' not found in source"
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Errorf("expected error containing '%s', got '%v'", expected, err)
	}
}

func TestValueSorted_UnmarshalJSONNotFoundError(t *testing.T) {
	vs := &ValueSorted[int]{}
	data := []byte(`{"value":2,"source":{"1":"one"}}`)
	err := vs.UnmarshalJSON(data)
	// The error should contain the interpolated value, which will be '0' (zero value)
	expected := "form||ValueSorted: value '0' not found in source"
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Errorf("expected error containing '%s', got '%v'", expected, err)
	}
}

var testTranslationSorted = map[string]map[string]string{
	"en": {
		TranslationKeyValueSortedNotFound: "value '%v' not found (EN)",
	},
	"it": {
		TranslationKeyValueSortedNotFound: "valore '%v' non trovato (IT)",
	},
}

func testTranslateSorted(loc Localizer, key string, args ...any) string {
	locale := "en"
	if l, ok := loc.(testLocalizer); ok {
		locale = l.Locale
	}
	msg := key
	if m, ok := testTranslationSorted[locale]; ok {
		if t, ok := m[key]; ok {
			msg = t
		}
	}
	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}

func TestValueSorted_TranslationWithVariable(t *testing.T) {
	loc := testLocalizer{Locale: "it"}
	f := NewTranslatedForm(nil, testTranslateSorted)
	vs := &ValueSorted[int]{source: map[int]string{1: "one"}}
	err := vs.Set(42)
	var key string
	var args []any
	if vserr, ok := err.(ValueSortedError); ok {
		key = vserr.Key
		args = vserr.Args
	} else {
		t.Fatalf("expected ValueSortedError, got %T", err)
	}
	translated := f.validationErrorTranslated(loc, key, args...)
	if translated != "valore '42' non trovato (IT)" {
		t.Errorf("expected translated error 'valore '42' non trovato (IT)', got '%s'", translated)
	}
}

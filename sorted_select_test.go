package form

import (
	"fmt"
	"strings"
	"testing"
)

func TestSortedSelect_SortedMapper(t *testing.T) {
	type testKey int
	ss := SortedSelect[testKey]{}
	ss.SetSource(map[testKey]string{
		2: "two",
		1: "one",
		3: "three",
	})
	entries := ss.SortedMapper()
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

func TestSortedSelect_ScanAndValue(t *testing.T) {
	ss := &SortedSelect[int]{}
	ss.SetSource(map[int]string{1: "one"})
	err := ss.Scan(1)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if v, _ := ss.Value(); v != 1 {
		t.Errorf("expected value 1, got %v", v)
	}
}

func TestSortedSelect_ScanInvalidType(t *testing.T) {
	ss := &SortedSelect[int]{}
	ss.SetSource(map[int]string{1: "one"})
	err := ss.Scan("not an int")
	if err == nil {
		t.Error("expected error for invalid type, got nil")
	}
}

func TestSortedSelect_String(t *testing.T) {
	ss := SortedSelect[string]{value: "foo"}
	if ss.String() != "foo" {
		t.Errorf("expected 'foo', got '%s'", ss.String())
	}
}

func TestSortedSelect_SetNotFoundError(t *testing.T) {
	ss := &SortedSelect[int]{source: map[int]string{1: "one"}}
	err := ss.Set(2)
	expected := "form||SortedSelect: value '2' not found in source"
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Errorf("expected error containing '%s', got '%v'", expected, err)
	}
}

func TestSortedSelect_SetFromKeyNotFoundError(t *testing.T) {
	ss := &SortedSelect[int]{source: map[int]string{1: "one"}}
	err := ss.SetFromKey("2")
	if err == nil || !strings.Contains(err.Error(), TranslationKeySortedSelectKeyNotFound) {
		t.Errorf("expected error containing '%s', got '%v'", TranslationKeySortedSelectKeyNotFound, err)
	}
}

func TestSortedSelect_ScanTypeError(t *testing.T) {
	ss := &SortedSelect[int]{source: map[int]string{1: "one"}}
	err := ss.Scan("not an int")
	if err == nil || !strings.Contains(err.Error(), TranslationKeySortedSelectTypeError) {
		t.Errorf("expected error containing '%s', got '%v'", TranslationKeySortedSelectTypeError, err)
	}
}

func TestSortedSelect_ScanNotFoundError(t *testing.T) {
	ss := &SortedSelect[int]{source: map[int]string{1: "one"}}
	err := ss.Scan(2)
	// The error should contain the interpolated value, not the format string
	expected := "form||SortedSelect: value '2' not found in source"
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Errorf("expected error containing '%s', got '%v'", expected, err)
	}
}

func TestSortedSelect_UnmarshalJSONNotFoundError(t *testing.T) {
	ss := &SortedSelect[int]{}
	data := []byte(`{"value":2,"source":{"1":"one"}}`)
	err := ss.UnmarshalJSON(data)
	// The error should contain the interpolated value, which will be '0' (zero value)
	expected := "form||SortedSelect: value '0' not found in source"
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Errorf("expected error containing '%s', got '%v'", expected, err)
	}
}

var testTranslationSorted = map[string]map[string]string{
	"en": {
		TranslationKeySortedSelectNotFound: "value '%v' not found (EN)",
	},
	"it": {
		TranslationKeySortedSelectNotFound: "valore '%v' non trovato (IT)",
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

func TestSortedSelect_TranslationWithVariable(t *testing.T) {
	loc := testLocalizer{Locale: "it"}
	f := NewTranslatedForm(testTranslateSorted)
	ss := &SortedSelect[int]{source: map[int]string{1: "one"}}
	err := ss.Set(42)
	var key string
	var args []any
	if sserr, ok := err.(SortedSelectError); ok {
		key = sserr.Key
		args = sserr.Args
	} else {
		t.Fatalf("expected SortedSelectError, got %T", err)
	}
	translated := f.validationErrorTranslated(loc, key, args...)
	if translated != "valore '42' non trovato (IT)" {
		t.Errorf("expected translated error 'valore '42' non trovato (IT)', got '%s'", translated)
	}
}

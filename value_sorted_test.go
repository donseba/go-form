package form

import (
	"testing"
)

func TestValueSorted_SortedMapper(t *testing.T) {
	type testKey int
	vs := ValueSorted[testKey]{
		Source: map[testKey]string{
			2: "two",
			1: "one",
			3: "three",
		},
	}
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
	vs := &ValueSorted[int]{Source: map[int]string{1: "one"}}
	err := vs.Scan(1)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if v, _ := vs.Value(); v != 1 {
		t.Errorf("expected value 1, got %v", v)
	}
}

func TestValueSorted_ScanInvalidType(t *testing.T) {
	vs := &ValueSorted[int]{Source: map[int]string{1: "one"}}
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

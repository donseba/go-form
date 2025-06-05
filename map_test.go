package form

import (
	"net/http"
	"net/url"
	"testing"
)

type testStruct struct {
	Name   string
	Age    int
	Active bool
	Score  float64
	Note   string
}

func TestMapForm(t *testing.T) {
	form := url.Values{}
	form.Set("Name", "Alice")
	form.Set("Age", "30")
	form.Set("Active", "true")
	form.Set("Score", "99.5")
	form.Set("Note", "Hello")

	r := &http.Request{Form: form}

	var s testStruct
	err := MapForm(r, &s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Name != "Alice" {
		t.Errorf("expected Name 'Alice', got '%s'", s.Name)
	}
	if s.Age != 30 {
		t.Errorf("expected Age 30, got %d", s.Age)
	}
	if !s.Active {
		t.Errorf("expected Active true, got false")
	}
	if s.Score != 99.5 {
		t.Errorf("expected Score 99.5, got %f", s.Score)
	}
	if s.Note != "Hello" {
		t.Errorf("expected Note 'Hello', got '%s'", s.Note)
	}
}

func TestMapFormErrors(t *testing.T) {
	var s testStruct
	err := MapForm(nil, s) // not a pointer
	if err == nil {
		t.Error("expected error for non-pointer dst")
	}

	err = MapForm(nil, &struct{}{}) // nil request, but should not panic
	if err != nil {
		t.Errorf("unexpected error for nil request: %v", err)
	}
}

type nestedStruct struct {
	City string
	Zip  int
}

type parentStruct struct {
	Name   string
	Nested nestedStruct
}

func TestMapFormNestedStruct(t *testing.T) {
	form := url.Values{}
	form.Set("Name", "Bob")
	form.Set("Nested.City", "Amsterdam")
	form.Set("Nested.Zip", "12345")

	r := &http.Request{Form: form}

	var s parentStruct
	err := MapForm(r, &s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Name != "Bob" {
		t.Errorf("expected Name 'Bob', got '%s'", s.Name)
	}

	if s.Nested.City != "Amsterdam" {
		t.Errorf("expected Nested.City 'Amsterdam', got '%s'", s.Nested.City)
	}

	if s.Nested.Zip != 12345 {
		t.Errorf("expected Nested.Zip 12345, got %d", s.Nested.Zip)
	}
}

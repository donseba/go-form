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

// UUID representation as [16]byte
type UUIDStruct struct {
	UUID [16]byte
	Name string
}

func TestMapFormUUID(t *testing.T) {
	form := url.Values{}
	form.Set("Name", "Document")
	form.Set("UUID", "123e4567-e89b-12d3-a456-426614174000")

	r := &http.Request{Form: form}

	var s UUIDStruct
	err := MapForm(r, &s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected byte representation of UUID "123e4567-e89b-12d3-a456-426614174000"
	expected := [16]byte{
		0x12, 0x3e, 0x45, 0x67,
		0xe8, 0x9b,
		0x12, 0xd3,
		0xa4, 0x56,
		0x42, 0x66, 0x14, 0x17, 0x40, 0x00,
	}

	if s.UUID != expected {
		t.Errorf("UUID not correctly parsed")
	}

	if s.Name != "Document" {
		t.Errorf("expected Name 'Document', got '%s'", s.Name)
	}
}

type CheckboxStruct struct {
	Name          string
	Subscribed    bool // will be set with "on" value
	Notifications bool // will be set with "true" value
	Marketing     bool // will be missing from form (should be false)
}

func TestMapFormCheckboxOn(t *testing.T) {
	form := url.Values{}
	form.Set("Name", "User")
	form.Set("Subscribed", "on") // HTML standard checkbox value
	form.Set("Notifications", "true")
	// Marketing field is intentionally not set

	r := &http.Request{Form: form}

	var s CheckboxStruct
	err := MapForm(r, &s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s.Name != "User" {
		t.Errorf("expected Name 'User', got '%s'", s.Name)
	}

	if !s.Subscribed {
		t.Errorf("expected Subscribed to be true with 'on' value, got false")
	}

	if !s.Notifications {
		t.Errorf("expected Notifications to be true with 'true' value, got false")
	}

	if s.Marketing {
		t.Errorf("expected Marketing to be false when not present in form, got true")
	}
}

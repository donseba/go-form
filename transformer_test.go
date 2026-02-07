package form

import (
	"testing"

	"github.com/donseba/go-form/types"
)

type ModelA struct {
	TextField  string `required:"true" name:"SomeRandomTextFieldName"`
	IntField   int64
	FloatField float64
	SubGroup   struct {
		SubTextField string `required:"true"`
		SubIntField  int64
	} `legend:"legendSubGroup"`
}

// MockSortedMapEntry implements SortedMap interface for testing
type MockSortedMapEntry struct {
	Key_   string
	Value_ string
}

func (m *MockSortedMapEntry) Key() string {
	return m.Key_
}

func (m *MockSortedMapEntry) Value() string {
	return m.Value_
}

// MockSortedMapper implements SortedMapper interface for testing
type MockSortedMapper struct {
	Entries []SortedMap
}

func (m *MockSortedMapper) String() string {
	return "MockSortedMapper"
}

func (m *MockSortedMapper) SortedMapper() []SortedMap {
	return m.Entries
}

// ModelWithSortedMapper contains a SortedMapper field for testing
type ModelWithSortedMapper struct {
	Name           string
	SortedDropdown *MockSortedMapper `form:"dropdown"`
}

// --- Enum tests ---
type StatusEnum string

const (
	StatusActive   StatusEnum = "active"
	StatusInactive StatusEnum = "inactive"
	StatusPending  StatusEnum = "pending"
)

func (s StatusEnum) String() string {
	return string(s)
}

func (s StatusEnum) Enum() []any {
	return []any{StatusActive, StatusInactive, StatusPending}
}

// ModelWithEnumNoTranslate has an enum field WITHOUT translation
type ModelWithEnumNoTranslate struct {
	Status StatusEnum `form:"dropdown" label:"Status"`
}

// ModelWithEnumTranslate has an enum field WITH translation enabled
type ModelWithEnumTranslate struct {
	Status StatusEnum `form:"dropdown" label:"Status" translate:"true"`
}

func TestNewTransformer(t *testing.T) {
	out, err := NewTransformer(&ModelA{})
	if err != nil {
		t.Error(err)
	}

	expected := `{"fields":[{"type":"","name":"SomeRandomTextFieldName","id":"SomeRandomTextFieldName","label":"SomeRandomTextFieldName","value":"","required":true},{"type":"input","inputType":"number","name":"IntField","id":"IntField","label":"IntField","value":0,"step":"1"},{"type":"input","inputType":"number","name":"FloatField","id":"FloatField","label":"FloatField","value":0,"step":"any"},{"type":"group","name":"SubGroup","id":"SubGroup","label":"SubGroup","value":{"SubTextField":"","SubIntField":0},"legend":"legendSubGroup","fields":[{"type":"","name":"SubGroup.SubTextField","id":"SubGroup.SubTextField","label":"SubTextField","value":"","required":true},{"type":"input","inputType":"number","name":"SubGroup.SubIntField","id":"SubGroup.SubIntField","label":"SubIntField","value":0,"step":"1"}]}]}`
	if string(out.JSON()) != expected {
		t.Error("transformer render changed")
	}
}

func TestSortedMapperTransformer(t *testing.T) {
	// Create a model with a SortedMapper field
	model := &ModelWithSortedMapper{
		Name: "TestModel",
		SortedDropdown: &MockSortedMapper{
			Entries: []SortedMap{
				&MockSortedMapEntry{Key_: "opt1", Value_: "Option 1"},
				&MockSortedMapEntry{Key_: "opt2", Value_: "Option 2"},
				&MockSortedMapEntry{Key_: "opt3", Value_: "Option 3"},
			},
		},
	}

	transformer, err := NewTransformer(model)
	if err != nil {
		t.Fatalf("NewTransformer failed: %v", err)
	}

	if transformer == nil {
		t.Fatal("transformer is nil")
	}

	// Find the SortedDropdown field
	var sortedDropdownField *types.FormField
	for i := range transformer.Fields {
		if transformer.Fields[i].Name == "SortedDropdown" {
			sortedDropdownField = &transformer.Fields[i]
			break
		}
	}

	if sortedDropdownField == nil {
		t.Fatal("SortedDropdown field not found")
	}

	// Verify the field type is DropdownMapped
	if sortedDropdownField.Type != types.FieldTypeDropdownMapped {
		t.Errorf("Expected field type %s, got %s", types.FieldTypeDropdownMapped, sortedDropdownField.Type)
	}

	// Verify the values are correctly mapped from SortedMapper
	if len(sortedDropdownField.Values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(sortedDropdownField.Values))
	}

	expectedValues := []struct {
		value string
		name  string
	}{
		{value: "opt1", name: "Option 1"},
		{value: "opt2", name: "Option 2"},
		{value: "opt3", name: "Option 3"},
	}

	for i, expected := range expectedValues {
		if i >= len(sortedDropdownField.Values) {
			break
		}
		actual := sortedDropdownField.Values[i]
		if actual.Value != expected.value {
			t.Errorf("Value[%d]: expected %s, got %s", i, expected.value, actual.Value)
		}
		if actual.Name != expected.name {
			t.Errorf("Name[%d]: expected %s, got %s", i, expected.name, actual.Name)
		}
		if actual.Disabled {
			t.Errorf("Value[%d] should not be disabled", i)
		}
	}
}

func TestEnumWithoutTranslateTag(t *testing.T) {
	// Test enum without translate tag - should use plain enum values
	model := &ModelWithEnumNoTranslate{
		Status: StatusActive,
	}

	transformer, err := NewTransformer(model)
	if err != nil {
		t.Fatalf("NewTransformer failed: %v", err)
	}

	if transformer == nil {
		t.Fatal("transformer is nil")
	}

	// Find the Status field
	var statusField *types.FormField
	for i := range transformer.Fields {
		if transformer.Fields[i].Name == "Status" {
			statusField = &transformer.Fields[i]
			break
		}
	}

	if statusField == nil {
		t.Fatal("Status field not found")
	}

	// Verify the field type is Dropdown
	if statusField.Type != types.FieldTypeDropdown {
		t.Errorf("Expected field type dropdown, got %s", statusField.Type)
	}

	// Verify the values use plain enum values (not translation keys)
	if len(statusField.Values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(statusField.Values))
	}

	expectedValues := []struct {
		value string
		name  string
	}{
		{value: "active", name: "active"},
		{value: "inactive", name: "inactive"},
		{value: "pending", name: "pending"},
	}

	for i, expected := range expectedValues {
		if i >= len(statusField.Values) {
			break
		}
		actual := statusField.Values[i]
		if actual.Value != expected.value {
			t.Errorf("Value[%d]: expected %s, got %s", i, expected.value, actual.Value)
		}
		if actual.Name != expected.name {
			t.Errorf("Name[%d]: expected %s, got %s (should be plain enum value, not translation key)", i, expected.name, actual.Name)
		}
	}
}

func TestEnumWithTranslateTag(t *testing.T) {
	// Test enum with translate="true" tag - should use translation keys
	model := &ModelWithEnumTranslate{
		Status: StatusActive,
	}

	transformer, err := NewTransformer(model)
	if err != nil {
		t.Fatalf("NewTransformer failed: %v", err)
	}

	if transformer == nil {
		t.Fatal("transformer is nil")
	}

	// Find the Status field
	var statusField *types.FormField
	for i := range transformer.Fields {
		if transformer.Fields[i].Name == "Status" {
			statusField = &transformer.Fields[i]
			break
		}
	}

	if statusField == nil {
		t.Fatal("Status field not found")
	}

	// Verify the field type is Dropdown
	if statusField.Type != types.FieldTypeDropdown {
		t.Errorf("Expected field type dropdown, got %s", statusField.Type)
	}

	// Verify the values use translation keys (enum||TypeName.value format)
	if len(statusField.Values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(statusField.Values))
	}

	expectedValues := []struct {
		value string
		name  string
	}{
		{value: "active", name: "enum||StatusEnum.active"},
		{value: "inactive", name: "enum||StatusEnum.inactive"},
		{value: "pending", name: "enum||StatusEnum.pending"},
	}

	for i, expected := range expectedValues {
		if i >= len(statusField.Values) {
			break
		}
		actual := statusField.Values[i]
		if actual.Value != expected.value {
			t.Errorf("Value[%d]: expected %s, got %s", i, expected.value, actual.Value)
		}
		if actual.Name != expected.name {
			t.Errorf("Name[%d]: expected %s, got %s (should be translation key)", i, expected.name, actual.Name)
		}
	}
}

func TestEnumWithDefaultEnumTranslationGlobal(t *testing.T) {
	// Save original value
	originalValue := DefaultEnumTranslation
	defer func() { DefaultEnumTranslation = originalValue }()

	// Test with global DefaultEnumTranslation = true
	DefaultEnumTranslation = true

	model := &ModelWithEnumNoTranslate{
		Status: StatusActive,
	}

	transformer, err := NewTransformer(model)
	if err != nil {
		t.Fatalf("NewTransformer failed: %v", err)
	}

	if transformer == nil {
		t.Fatal("transformer is nil")
	}

	// Find the Status field
	var statusField *types.FormField
	for i := range transformer.Fields {
		if transformer.Fields[i].Name == "Status" {
			statusField = &transformer.Fields[i]
			break
		}
	}

	if statusField == nil {
		t.Fatal("Status field not found")
	}

	// Should use translation keys because DefaultEnumTranslation is true
	expectedValues := []struct {
		value string
		name  string
	}{
		{value: "active", name: "enum||StatusEnum.active"},
		{value: "inactive", name: "enum||StatusEnum.inactive"},
		{value: "pending", name: "enum||StatusEnum.pending"},
	}

	for i, expected := range expectedValues {
		if i >= len(statusField.Values) {
			break
		}
		actual := statusField.Values[i]
		if actual.Name != expected.name {
			t.Errorf("Name[%d]: expected %s, got %s (should use global default)", i, expected.name, actual.Name)
		}
	}
}

func TestEnumTranslateTagOverridesGlobal(t *testing.T) {
	// Save original value
	originalValue := DefaultEnumTranslation
	defer func() { DefaultEnumTranslation = originalValue }()

	// Set global to true
	DefaultEnumTranslation = true

	// But use translate="false" to opt-out
	type ModelWithEnumNoTranslateOverride struct {
		Status StatusEnum `form:"dropdown" label:"Status" translate:"false"`
	}

	model := &ModelWithEnumNoTranslateOverride{
		Status: StatusActive,
	}

	transformer, err := NewTransformer(model)
	if err != nil {
		t.Fatalf("NewTransformer failed: %v", err)
	}

	// Find the Status field
	var statusField *types.FormField
	for i := range transformer.Fields {
		if transformer.Fields[i].Name == "Status" {
			statusField = &transformer.Fields[i]
			break
		}
	}

	if statusField == nil {
		t.Fatal("Status field not found")
	}

	// Should NOT use translation keys because translate="false" explicitly overrides global
	if statusField.Values[0].Name != "active" {
		t.Errorf("Expected plain value 'active', got %s (struct tag should override global default)", statusField.Values[0].Name)
	}
}

package templates

import (
	"fmt"
	"testing"

	"github.com/donseba/go-form/types"
)

// This test acts as a lightweight "linter" for template maps.
// It ensures each template set provides all required FieldType/InputType entries.
//
// Rationale: missing templates usually surface only at runtime when rendering a form,
// which is easy to miss when adding a new template set (e.g. TailwindV4).
func TestTemplateMaps_HaveRequiredEntries(t *testing.T) {
	templateSets := []struct {
		name string
		tm   types.TemplateMap
	}{
		{"Plain", Plain},
		{"BootstrapV5", BootstrapV5},
		{"TailwindV3", TailwindV3},
		{"TailwindV4", TailwindV4},
	}

	// Required templates used by the renderer.
	// Note: FieldTypeInput must include InputFieldTypeNone, because it's used as the base template.
	required := map[types.FieldType][]types.InputFieldType{
		types.FieldTypeBase:           {types.InputFieldTypeNone},
		types.FieldTypeInput:          {types.InputFieldTypeNone},
		types.FieldTypeLabel:          {types.InputFieldTypeNone},
		types.FieldTypeWrapper:        {types.InputFieldTypeNone},
		types.FieldTypeError:          {types.InputFieldTypeNone},
		types.FieldTypeForm:           {types.InputFieldTypeNone},
		types.FieldTypeGroup:          {types.InputFieldTypeNone},
		types.FieldTypeInputGroup:     {types.InputFieldTypeNone},
		types.FieldTypeCheckbox:       {types.InputFieldTypeNone},
		types.FieldTypeRadios:         {types.InputFieldTypeNone, types.InputFieldTypeRadioStruct},
		types.FieldTypeDropdown:       {types.InputFieldTypeNone},
		types.FieldTypeDropdownMapped: {types.InputFieldTypeNone},
		types.FieldTypeTextArea:       {types.InputFieldTypeNone},
	}

	for _, ts := range templateSets {
		ts := ts
		t.Run(ts.name, func(t *testing.T) {
			for ft, inputTypes := range required {
				m, ok := ts.tm[ft]
				if !ok {
					t.Fatalf("missing field type %q", ft)
				}

				for _, it := range inputTypes {
					if _, ok := m[it]; !ok {
						t.Fatalf("missing template for field type %q input type %q", ft, it)
					}
				}
			}
		})
	}
}

func TestTemplateMaps_NoEmptyTemplates(t *testing.T) {
	templateSets := []struct {
		name string
		tm   types.TemplateMap
	}{
		{"Plain", Plain},
		{"BootstrapV5", BootstrapV5},
		{"TailwindV3", TailwindV3},
		{"TailwindV4", TailwindV4},
	}

	for _, ts := range templateSets {
		ts := ts
		t.Run(ts.name, func(t *testing.T) {
			for ft, m := range ts.tm {
				for it, tpl := range m {
					if tpl == "" {
						t.Fatalf("empty template string for field type %q input type %q", ft, it)
					}
				}
			}
		})
	}
}

func Example_requiredTemplateCoverage() {
	// This example exists so the required template coverage list is easy to discover.
	// (No output.)
	_ = fmt.Sprintf
}

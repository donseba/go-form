package form

import (
	"github.com/donseba/go-form/v2/types"
)

// collapseStructRadioGroups inspects group fields and, when a parent group declared as
// `form:"radios,radio_group"` contains child fields that represent struct-based radio options
// (Type=radios, InputType=radio_struct) sharing the same Name, it:
//   - converts the parent group into a FieldTypeRadios group (InputType=radio_group)
//   - populates parent.Values from the child fields
//   - sets parent.Value to the selected option (first checked)
//   - clears parent.Fields (so we don't render nested wrappers/labels)
func collapseStructRadioGroups(fields []types.FormField) []types.FormField {
	for i := range fields {
		// Recurse first
		if len(fields[i].Fields) > 0 {
			fields[i].Fields = collapseStructRadioGroups(fields[i].Fields)
		}

		if fields[i].Type != types.FieldTypeGroup {
			continue
		}
		// Explicit opt-in: parent must be declared as radios + radio_group.
		if !(fields[i].Type == types.FieldTypeGroup && fields[i].InputType == types.InputFieldTypeRadioGroup) {
			continue
		}
		if len(fields[i].Fields) == 0 {
			continue
		}

		// Detect: all children are struct-radio options and share the same Name.
		var radioName string
		allRadio := true
		for _, ch := range fields[i].Fields {
			if ch.Type != types.FieldTypeRadios || ch.InputType != types.InputFieldTypeRadioStruct {
				allRadio = false
				break
			}
			if radioName == "" {
				radioName = ch.Name
			} else if ch.Name != radioName {
				allRadio = false
				break
			}
		}
		if !allRadio || radioName == "" {
			continue
		}

		// Collapse into a single radio group.
		fields[i].Type = types.FieldTypeRadios
		fields[i].InputType = types.InputFieldTypeRadioGroup
		fields[i].Name = radioName
		fields[i].Id = radioName

		fields[i].Values = nil
		fields[i].Value = ""
		for _, ch := range fields[i].Fields {
			fields[i].Values = append(fields[i].Values, types.FieldValue{
				Value:    ch.Id,
				Name:     ch.Label,
				Disabled: ch.Disabled,
			})
			if s, ok := ch.Value.(string); ok && s != "" {
				fields[i].Value = ch.Id
			}
		}

		fields[i].Fields = nil
	}
	return fields
}

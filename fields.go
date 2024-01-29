package form

type FieldType string

const (
	FieldTypeGroup          FieldType = "group"
	FieldTypeCheckbox       FieldType = "checkbox"
	FieldTypeChecklist      FieldType = "checklist"
	FieldTypeInput          FieldType = "input"
	FieldTypeLabel          FieldType = "label"
	FieldTypeRadios         FieldType = "radios"
	FieldTypeDropdown       FieldType = "dropdown"
	FieldTypeDropdownMapped FieldType = "dropdownmapped"
	FieldTypeSubmit         FieldType = "submit"
	FieldTypeTextArea       FieldType = "textarea"
)

type InputFieldType string

const (
	InputFieldTypeText     InputFieldType = "text"
	InputFieldTypePassword InputFieldType = "password"
	InputFieldTypeEmail    InputFieldType = "email"
	InputFieldTypeTel      InputFieldType = "tel"
	InputFieldTypeNumber   InputFieldType = "number"
	InputFieldTypeNone     InputFieldType = ""
)

func (i InputFieldType) String() string {
	return string(i)
}

func (i InputFieldType) Enum() []any {
	return []interface{}{
		InputFieldTypeText,
		InputFieldTypePassword,
		InputFieldTypeEmail,
		InputFieldTypeTel,
		InputFieldTypeNumber,
	}
}

type FieldValue struct {
	Name     string `json:"name,omitempty"`
	Value    string `json:"value,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
}

type FormField struct {
	Id          string         `json:"id,omitempty"`
	Placeholder string         `json:"placeholder,omitempty"`
	Name        string         `json:"name,omitempty"`
	Value       interface{}    `json:"value,omitempty"`
	Type        FieldType      `json:"type,omitempty"`
	InputType   InputFieldType `json:"inputType,omitempty"`
	Label       string         `json:"label,omitempty"`
	Step        string         `json:"step,omitempty"`
	Rows        string         `json:"rows,omitempty"`
	Cols        string         `json:"cols,omitempty"`
	Values      []FieldValue   `json:"values,omitempty"`
	Required    bool           `json:"required"`
	Fields      []FormField    `json:"fields,omitempty"`
	Legend      string         `json:"legend,omitempty"`
}

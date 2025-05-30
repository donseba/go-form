package types

// FieldType represents the type of form field
type FieldType string

// InputFieldType represents the HTML input type
type InputFieldType string

func (i InputFieldType) String() string {
	return string(i)
}

// FieldValue represents a value in a dropdown or radio group
type FieldValue struct {
	Value    string `json:"value"`
	Name     string `json:"name"`
	Disabled bool   `json:"disabled"`
}

// FormField represents a form field
type FormField struct {
	Type        FieldType      `json:"type"`
	InputType   InputFieldType `json:"inputType,omitempty"`
	Name        string         `json:"name"`
	Id          string         `json:"id,omitempty"`
	Label       string         `json:"label,omitempty"`
	Value       interface{}    `json:"value,omitempty"`
	Placeholder string         `json:"placeholder,omitempty"`
	Required    bool           `json:"required,omitempty"`
	Hidden      bool           `json:"hidden,omitempty"`
	Step        string         `json:"step,omitempty"`
	Rows        string         `json:"rows,omitempty"`
	Cols        string         `json:"cols,omitempty"`
	Legend      string         `json:"legend,omitempty"`
	Values      []FieldValue   `json:"values,omitempty"`
	Fields      []FormField    `json:"fields,omitempty"`
	Target      string         `json:"target,omitempty"`
	Method      string         `json:"method,omitempty"`
}

// TemplateMap represents a map of field types to their input type templates
type TemplateMap map[FieldType]map[InputFieldType]string

// Constants for field types
const (
	FieldTypeInput          FieldType = "input"
	FieldTypeCheckbox       FieldType = "checkbox"
	FieldTypeRadios         FieldType = "radios"
	FieldTypeDropdown       FieldType = "dropdown"
	FieldTypeDropdownMapped FieldType = "dropdownmapped"
	FieldTypeTextArea       FieldType = "textarea"
	FieldTypeGroup          FieldType = "group"
	FieldTypeError          FieldType = "error"
	FieldTypeLabel          FieldType = "label"
	FieldTypeWrapper        FieldType = "wrapper"
	FieldTypeForm           FieldType = "form"
)

// Constants for input types
const (
	InputFieldTypeText        InputFieldType = "text"
	InputFieldTypePassword    InputFieldType = "password"
	InputFieldTypeEmail       InputFieldType = "email"
	InputFieldTypeTel         InputFieldType = "tel"
	InputFieldTypeNumber      InputFieldType = "number"
	InputFieldTypeDate        InputFieldType = "date"
	InputFieldTypeNone        InputFieldType = ""
	InputFieldTypeRadio       InputFieldType = "radio"
	InputFieldTypeHidden      InputFieldType = "hidden"
	InputFieldTypeRadioStruct InputFieldType = "radio_struct"
)

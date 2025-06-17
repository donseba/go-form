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
	Type        FieldType         `json:"type"`
	InputType   InputFieldType    `json:"inputType,omitempty"`
	Name        string            `json:"name"`
	Id          string            `json:"id,omitempty"`
	Label       string            `json:"label,omitempty"`
	Value       interface{}       `json:"value,omitempty"`
	Placeholder string            `json:"placeholder,omitempty"`
	Description string            `json:"description,omitempty"`
	Required    bool              `json:"required,omitempty"`
	Hidden      bool              `json:"hidden,omitempty"`
	Min         string            `json:"min,omitempty"`
	Max         string            `json:"max,omitempty"`
	MaxLength   string            `json:"maxLength,omitempty"`
	Step        string            `json:"step,omitempty"`
	Rows        string            `json:"rows,omitempty"`
	Cols        string            `json:"cols,omitempty"`
	Legend      string            `json:"legend,omitempty"`
	Values      []FieldValue      `json:"values,omitempty"`
	Fields      []FormField       `json:"fields,omitempty"`
	Target      string            `json:"target,omitempty"`
	Method      string            `json:"method,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	GroupBefore string            `json:"groupBefore,omitempty"`
	GroupAfter  string            `json:"groupAfter,omitempty"`
	Class       string            `json:"class,omitempty"`
	Data        map[string]string `json:"data,omitempty"` // Data attributes
}

// TemplateMap represents a map of field types to their input type templates
type TemplateMap map[FieldType]map[InputFieldType]string

// Constants for field types
const (
	FieldTypeBase           FieldType = "base"
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
	FieldTypeInputGroup     FieldType = "inputgroup"
)

// Constants for input types
const (
	InputFieldTypeColor         InputFieldType = "color"
	InputFieldTypeDate          InputFieldType = "date"
	InputFieldTypeDateTimeLocal InputFieldType = "datetime-local"
	InputFieldTypeEmail         InputFieldType = "email"
	InputFieldTypeHidden        InputFieldType = "hidden"
	InputFieldTypeImage         InputFieldType = "image"
	InputFieldTypeMonth         InputFieldType = "month"
	InputFieldTypeNumber        InputFieldType = "number"
	InputFieldTypePassword      InputFieldType = "password"
	InputFieldTypeRadioStruct   InputFieldType = "radio_struct"
	InputFieldTypeRange         InputFieldType = "range"
	InputFieldTypeSearch        InputFieldType = "search"
	InputFieldTypeSubmit        InputFieldType = "submit"
	InputFieldTypeTel           InputFieldType = "tel"
	InputFieldTypeText          InputFieldType = "text"
	InputFieldTypeTime          InputFieldType = "time"
	InputFieldTypeUrl           InputFieldType = "url"
	InputFieldTypeWeek          InputFieldType = "week"
	InputFieldTypeNone          InputFieldType = ""
)

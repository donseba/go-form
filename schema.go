package form

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/donseba/go-form/v2/types"
)

type (
	// Schema is a form defined at runtime rather than by a struct, such as
	// a form an editor puts together or the settings a theme declares. It
	// renders through the same themes and checks the same rules as a struct
	// with tags:
	//
	//	schema := form.Schema{Fields: []form.Spec{
	//		{Name: "name", Label: "Name", Required: true, MaxLength: 100},
	//		{Name: "email", Label: "Email", InputType: types.InputFieldTypeEmail, Required: true},
	//		{Name: "topic", Label: "Topic", Type: types.FieldTypeDropdown, Placeholder: "—", Options: options},
	//	}}
	//
	//	values := form.MapValues(r, schema)
	//	errs := renderer.ValidateSchemaLocalized(schema, values, loc)
	//	model := form.WithInfo(schema.Bind(values), form.Info{Target: "/contact", Method: "post"})
	Schema struct {
		Fields []Spec
	}

	// Spec is one field of a Schema: what a struct field says with tags.
	Spec struct {
		// Name is the form key; Label defaults to it.
		Name  string
		Label string
		// Type is the kind of control: input (the default), textarea,
		// dropdown, radios, checkbox or multicheckbox.
		Type types.FieldType
		// InputType is the type of an input, such as email, number or
		// date; text by default.
		InputType   types.InputFieldType
		Placeholder string
		Description string
		// Options are the choices of a dropdown, radios or multicheckbox.
		// A submitted value must be one of them.
		Options []types.FieldValue

		Required  bool
		MinLength int
		MaxLength int
		// Min, Max and Step limit a number input.
		Min, Max, Step string
		// Pattern is a regular expression the value must match, as the
		// pattern tag.
		Pattern string
		// URL is a rule of the url tag: "true", "http,https" or "link".
		URL string
		// SingleLine allows one line without control characters, such as
		// tabs or line breaks, like the line tag.
		SingleLine bool
		// Validate names rules registered with RegisterValidationMethod
		// or RegisterLocalizedValidation, separated by commas, like the
		// validate tag, such as "mail". They check the first submitted
		// value as a string, after go-form's own rules, also when it is
		// empty, as they check a struct's field.
		Validate string
		// Messages replace go-form's messages per rule, such as
		// {"pattern": "Use lowercase letters and dashes"}, like the
		// patternMsg tag.
		Messages map[string]string

		AutoComplete string
		ReadOnly     bool
		Disabled     bool
		Rows         int
		Class        string
	}

	// BoundSchema is a schema with the values to show, made by Schema.Bind.
	// Render it like a struct, with WithInfo and form_render.
	BoundSchema struct {
		Schema Schema
		Values url.Values
	}
)

// Bind returns the schema with values to render, such as the submitted
// values after a failed validation or the stored ones.
func (s Schema) Bind(values url.Values) BoundSchema {
	return BoundSchema{Schema: s, Values: values}
}

// MapValues returns the submitted values of the schema's fields, without
// white space around them (except in password inputs). A checkbox is "on"
// when ticked and absent otherwise; a multicheckbox has a value per ticked
// box.
func MapValues(r *http.Request, s Schema) url.Values {
	if r.Form == nil {
		_ = r.ParseMultipartForm(defaultMaxMemory)
	}

	values := url.Values{}
	for _, spec := range s.Fields {
		spec.addSubmitted(values, r.Form[spec.Name])
	}

	return values
}

// addSubmitted adds the values submitted for the field to values, as
// MapValues returns them.
func (spec Spec) addSubmitted(values url.Values, submitted []string) {
	switch spec.fieldType() {
	case types.FieldTypeMultiCheckbox:
		for _, value := range submitted {
			if value = strings.TrimSpace(value); value != "" {
				values.Add(spec.Name, value)
			}
		}
	case types.FieldTypeCheckbox:
		if truthy(first(submitted)) {
			values.Set(spec.Name, "on")
		}
	default:
		if len(submitted) == 0 {
			return
		}

		value := submitted[0]
		if spec.inputType() != types.InputFieldTypePassword {
			value = strings.TrimSpace(value)
		}

		values.Set(spec.Name, value)
	}
}

// ValidateSchemaLocalized checks values against the rules of the schema,
// with messages for loc. Values of fields outside the schema are ignored.
func (f *Form) ValidateSchemaLocalized(s Schema, values url.Values, loc Localizer) FieldErrors {
	var errs FieldErrors
	for _, spec := range s.Fields {
		errs = append(errs, f.validateSpec(spec, values[spec.Name], loc)...)
	}

	return errs
}

// validateSpec checks the submitted values of one field, with the rules of
// the struct tags the spec stands for, then with the registered rules of
// Spec.Validate.
func (f *Form) validateSpec(spec Spec, submitted []string, loc Localizer) FieldErrors {
	errs := f.validateSpecRules(spec, submitted, loc)
	if spec.Validate == "" {
		return errs
	}

	return append(errs, f.validateCustom(spec.structField(), reflect.ValueOf(first(submitted)), loc)...)
}

// validateSpecRules checks the submitted values of one field with go-form's
// own rules.
func (f *Form) validateSpecRules(spec Spec, submitted []string, loc Localizer) FieldErrors {
	field := spec.structField()
	getErr := f.messageFunc(field, loc)
	invalid := func(value string) FieldErrors {
		return FieldErrors{
			FieldValidationError{Field: spec.Name, Rule: "values", Err: getErr("values", TranslationKeyInvalidValue, value)},
		}
	}

	switch spec.fieldType() {
	case types.FieldTypeCheckbox:
		checked := truthy(first(submitted))

		return validateRequired(field, reflect.ValueOf(checked), getErr)
	case types.FieldTypeMultiCheckbox:
		chosen := slices.DeleteFunc(slices.Clone(submitted), func(value string) bool {
			return value == ""
		})

		if value, ok := spec.unoffered(chosen); ok {
			return invalid(value)
		}

		return validateRequired(field, reflect.ValueOf(chosen), getErr)
	}

	value := first(submitted)

	if errs := validateRequired(field, reflect.ValueOf(value), getErr); len(errs) > 0 || value == "" {
		return errs
	}

	if len(spec.Options) > 0 && !spec.offers(value) {
		return invalid(value)
	}

	switch spec.inputType() {
	case types.InputFieldTypeNumber, types.InputFieldTypeRange:
		number, err := strconv.ParseFloat(value, 64)
		if err != nil || !finite(number) {
			return invalid(value)
		}

		var errs FieldErrors
		errs = append(errs, validateMinMax(f, field, reflect.ValueOf(number), loc, getErr)...)
		errs = append(errs, validateStep(f, field, reflect.ValueOf(number), loc, getErr)...)

		return errs
	case types.InputFieldTypeDate,
		types.InputFieldTypeTime,
		types.InputFieldTypeMonth,
		types.InputFieldTypeWeek,
		types.InputFieldTypeDateTimeLocal:
		var target time.Time
		if err := parseTimeToFieldValue(reflect.ValueOf(&target).Elem(), field, value, nil); err != nil {
			return invalid(value)
		}

		return nil
	}

	return f.validateField(field, reflect.ValueOf(value), loc)
}

// structField is the struct field with tags the spec stands for, so that
// the rules of struct tags check it.
func (spec Spec) structField() reflect.StructField {
	form := string(spec.fieldType())
	if spec.inputType() != "" {
		form += "," + string(spec.inputType())
	}

	tags := []string{
		tagPair(tagName, spec.Name),
		tagPair(tagLabel, spec.label()),
		tagPair(tagForm, form),
	}

	if spec.Required {
		tags = append(tags, tagPair(tagRequired, "true"))
	}

	if spec.MinLength > 0 {
		tags = append(tags, tagPair(tagMinLength, strconv.Itoa(spec.MinLength)))
	}

	if spec.MaxLength > 0 {
		tags = append(tags, tagPair(tagMaxLength, strconv.Itoa(spec.MaxLength)))
	}

	if spec.SingleLine {
		tags = append(tags, tagPair("line", "true"))
	}

	for tag, value := range map[string]string{
		tagMin: spec.Min, tagMax: spec.Max, tagStep: spec.Step, tagPattern: spec.Pattern, "url": spec.URL, "validate": spec.Validate,
	} {
		if value != "" {
			tags = append(tags, tagPair(tag, value))
		}
	}

	for rule, message := range spec.Messages {
		tags = append(tags, tagPair(rule+"Msg", message))
	}

	return reflect.StructField{
		Name: "Field",
		Type: reflect.TypeOf(""),
		Tag:  reflect.StructTag(strings.Join(tags, " ")),
	}
}

// tagPair is one key:"value" pair of a struct tag.
func tagPair(key, value string) string {
	return key + ":" + strconv.Quote(value)
}

// formField is the field the spec renders as, showing values.
func (spec Spec) formField(values url.Values) types.FormField {
	field := types.FormField{
		Type:         spec.fieldType(),
		InputType:    spec.inputType(),
		Name:         spec.Name,
		Id:           spec.Name,
		Label:        spec.label(),
		Placeholder:  spec.Placeholder,
		Description:  spec.Description,
		Required:     spec.Required,
		Disabled:     spec.Disabled,
		ReadOnly:     spec.ReadOnly,
		AutoComplete: spec.AutoComplete,
		Class:        spec.Class,
		Values:       spec.Options,
		Min:          spec.Min,
		Max:          spec.Max,
		Step:         spec.Step,
		Pattern:      clientPattern(spec.Pattern),
		Value:        values.Get(spec.Name),
	}

	if spec.MinLength > 0 {
		field.MinLength = strconv.Itoa(spec.MinLength)
	}

	if spec.MaxLength > 0 {
		field.MaxLength = strconv.Itoa(spec.MaxLength)
	}

	if spec.Rows > 0 {
		field.Rows = strconv.Itoa(spec.Rows)
	}

	switch field.Type {
	case types.FieldTypeRadios:
		// The options render as one group of radio buttons.
		field.InputType = types.InputFieldTypeRadioGroup
	case types.FieldTypeCheckbox:
		field.Value = truthy(values.Get(spec.Name))
	case types.FieldTypeMultiCheckbox:
		field.ValueMap = map[string]bool{}
		for _, value := range values[spec.Name] {
			field.ValueMap[value] = true
		}
	}

	return field
}

// fieldType is the type of control, input by default.
func (spec Spec) fieldType() types.FieldType {
	if spec.Type == "" {
		return types.FieldTypeInput
	}

	return spec.Type
}

// inputType is the type of an input, text by default.
func (spec Spec) inputType() types.InputFieldType {
	if spec.InputType == "" && spec.fieldType() == types.FieldTypeInput {
		return types.InputFieldTypeText
	}

	return spec.InputType
}

// label is the label of the field, its name by default.
func (spec Spec) label() string {
	if spec.Label == "" {
		return spec.Name
	}

	return spec.Label
}

// unoffered returns the first of values that is not one of the options.
func (spec Spec) unoffered(values []string) (string, bool) {
	for _, value := range values {
		if !spec.offers(value) {
			return value, true
		}
	}

	return "", false
}

// offers reports whether value is one of the options that are not
// disabled.
func (spec Spec) offers(value string) bool {
	return slices.ContainsFunc(spec.Options, func(option types.FieldValue) bool {
		return option.Value == value && !option.Disabled
	})
}

// truthy reports whether a submitted checkbox value means ticked.
func truthy(value string) bool {
	checked, err := strconv.ParseBool(value)

	return value == "on" || (err == nil && checked)
}

// boundSchemaOf returns model as a BoundSchema when it is a Schema or a
// BoundSchema, or a pointer to one; a Schema has no values. ok is false for
// other models, and isNil is set for a nil pointer to one, which has no
// fields.
func boundSchemaOf(model any) (bound BoundSchema, isNil, ok bool) {
	switch schema := model.(type) {
	case Schema:
		return schema.Bind(nil), false, true
	case *Schema:
		if schema == nil {
			return BoundSchema{}, true, true
		}

		return schema.Bind(nil), false, true
	case BoundSchema:
		return schema, false, true
	case *BoundSchema:
		if schema == nil {
			return BoundSchema{}, true, true
		}

		return *schema, false, true
	}

	return BoundSchema{}, false, false
}

// schemaFields returns the fields of a Schema or BoundSchema model.
func schemaFields(model any) ([]types.FormField, bool, error) {
	bound, isNil, ok := boundSchemaOf(model)
	if !ok {
		return nil, false, nil
	}

	if isNil {
		return nil, true, fmt.Errorf("form model must not be nil")
	}

	fields := make([]types.FormField, 0, len(bound.Schema.Fields))
	for _, spec := range bound.Schema.Fields {
		fields = append(fields, spec.formField(bound.Values))
	}

	return fields, true, nil
}

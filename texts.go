package form

import (
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/donseba/go-form/v2/types"
)

// Texts returns the texts of forms, each once, in order of appearance, for
// an application to extract into its catalogues: field labels (the name
// when there is no label), legends, descriptions, placeholders, option
// labels that are translated (a values tag, a struct radio group, an
// Enumerator's keys, a select with translate:"true" or RenderModel.Options
// with Translate set), the submit and cancel texts of form metadata, and
// the messages per rule (requiredMsg and any other tag ending in Msg,
// errorMsg included). The form's translation function translates them
// when they render or validate, except errorMsg, which shows as written.
//
// models are what the form renders: structs or pointers to them (a nil
// pointer stands for its zero value), RenderModels made with WithInfo, and
// Schemas or BoundSchemas. The options of a select are only known from a
// value that holds them. go-form's own messages are not included; they are
// in its catalogues (locales).
//
//	texts := form.Texts(SignupInput{}, form.WithInfo(ContactInput{}, info), schema)
func Texts(models ...any) []string {
	collector := &textCollector{seen: map[string]bool{}}

	for _, model := range models {
		collector.model(model)
	}

	return collector.texts
}

// textCollector gathers texts once each, in order.
type textCollector struct {
	texts []string
	seen  map[string]bool
}

// add adds text unless it is empty or known.
func (c *textCollector) add(text string) {
	if text == "" || c.seen[text] {
		return
	}

	c.seen[text] = true
	c.texts = append(c.texts, text)
}

// model adds the texts of one model.
func (c *textCollector) model(model any) {
	inner := model
	wrapped := false

	var render RenderModel

	switch m := model.(type) {
	case RenderModel:
		render, inner, wrapped = m, m.Model, true
	case *RenderModel:
		if m == nil {
			return
		}

		render, inner, wrapped = *m, m.Model, true
	}

	if inner == nil {
		return
	}

	if bound, _, ok := boundSchemaOf(inner); ok {
		c.rendered(model)

		for _, spec := range bound.Schema.Fields {
			for _, rule := range slices.Sorted(maps.Keys(spec.Messages)) {
				c.add(spec.Messages[rule])
			}
		}

		return
	}

	value := reflect.ValueOf(inner)

	// A nil pointer stands for the zero value of its struct.
	if value.Kind() == reflect.Pointer && value.IsNil() {
		if value.Type().Elem().Kind() != reflect.Struct {
			return
		}

		value = reflect.New(value.Type().Elem())
		inner = value.Interface()
	}

	if wrapped {
		render.Model = inner
		c.rendered(render)
	} else {
		c.rendered(inner)
	}

	c.messages(value, overridesOf(value).merge(render.Overrides))
}

// rendered adds the texts the model renders with.
func (c *textCollector) rendered(model any) {
	transformer, err := NewTransformer(model)
	if err != nil {
		return
	}

	c.fields(transformer.Fields)
}

// fields adds the texts of fields and their nested fields.
func (c *textCollector) fields(fields []types.FormField) {
	for _, field := range fields {
		if field.Type == types.FieldTypeForm {
			c.add(field.Label)

			if field.CancelTarget != "" {
				c.add(field.CancelText)
			}

			continue
		}

		switch {
		case field.Type == types.FieldTypeGroup:
			// A group shows its legend, which defaults to its label.
			c.add(field.Legend)
		case field.InputType != types.InputFieldTypeHidden:
			c.add(field.Label)
			c.add(field.Placeholder)
			c.add(field.Description)
		}

		for _, option := range field.Values {
			if option.Translate {
				c.add(option.Name)
			}
		}

		c.fields(field.Fields)
	}
}

// messages adds the messages per rule of the fields of value, a struct or
// a pointer to one, and of its nested structs.
func (c *textCollector) messages(value reflect.Value, overrides Overrides) {
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			value = reflect.New(value.Type().Elem())
		}

		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < value.NumField(); i++ {
		structField := value.Type().Field(i)
		if !structField.IsExported() || structField.Type == reflect.TypeOf(Info{}) {
			continue
		}

		name := validationFieldName(structField)
		field := overrides.apply(name, structField)

		c.ruleMessages(field.Tag)

		if nested, ok := nestedFormValue(field, value.Field(i)); ok {
			c.messages(nested, overridesOf(nested).merge(below(overrides, name)))
		}
	}
}

// ruleMessages adds the messages per rule of tag: the values of the tags
// whose name ends in Msg.
func (c *textCollector) ruleMessages(tag reflect.StructTag) {
	for _, entry := range parseTag(tag) {
		if strings.HasSuffix(entry.key, "Msg") {
			c.add(entry.value)
		}
	}
}

// nestedFormValue returns the struct that value, the value of field, holds
// when field is a nested form: the struct itself, the one a pointer points
// to, or the zero value of its struct for a nil pointer. ok is false for
// other fields (see nestedForm), such as a time.
func nestedFormValue(field reflect.StructField, value reflect.Value) (reflect.Value, bool) {
	if !nestedForm(field) || isTime(field.Type) {
		return value, false
	}

	if value.Kind() == reflect.Pointer && value.IsNil() {
		value = reflect.New(value.Type().Elem())
	}

	if value.Kind() == reflect.Pointer {
		value = value.Elem()
	}

	return value, value.Kind() == reflect.Struct
}

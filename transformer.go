package form

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type (
	Enumerator interface{ Enum() []any }
	Mapper     interface {
		Mapper() map[string]string
		String() string
	}

	// SortedMapper is new addition to provide sorted key value pairs
	SortedMapper interface {
		String() string
		SortedMapper() []SortedMap
	}

	SortedMap interface {
		Key() string
		Value() string
	}
)

var (
	enumType   = reflect.TypeOf((*Enumerator)(nil)).Elem()
	mapperType = reflect.TypeOf((*Mapper)(nil)).Elem()

	//new addition to provide sorted key value pairs
	sortedMapperType = reflect.TypeOf((*SortedMapper)(nil)).Elem()
)

type Transformer struct {
	Fields []FormField `json:"fields"`
}

func NewTransformer(model interface{}) (*Transformer, error) {
	modelValue := reflect.ValueOf(model)
	modelType := modelValue.Type()

	// check if we received a pointer
	if modelValue.Kind() == reflect.Ptr || modelValue.Kind() == reflect.Interface {
		modelValue = modelValue.Elem()
		modelType = modelType.Elem()
	}

	tr := &Transformer{}
	fields, err := tr.scanModel(modelValue, modelType)
	if err != nil {
		return nil, err
	}

	tr.Fields = fields

	return tr, nil
}

func (t *Transformer) JSON() json.RawMessage {
	out, _ := json.Marshal(t)

	return out
}

// scanModel the incoming interface and ensure we can work with it.
func (t *Transformer) scanModel(rValue reflect.Value, rType reflect.Type, names ...string) ([]FormField, error) {
	var fields []FormField

	for i := 0; i < rType.NumField(); i++ {
		tags := rType.Field(i).Tag

		name := tags.Get("name")
		if name == "" {
			name = rType.Field(i).Name
		}

		nname := append(names, name)

		field := FormField{
			Label:       tags.Get("label"),
			Placeholder: tags.Get("placeholder"),
			Name:        strings.Join(nname, "."),
			Value:       rValue.Field(i).Interface(),
		}

		if field.Label == "" {
			field.Label = name
		}

		if tags.Get("required") == "true" {
			field.Required = true
		}

		if rType.Field(i).Type.Implements(enumType) {
			enums := reflect.New(rType.Field(i).Type).Interface().(Enumerator).Enum()
			var fieldValue []FieldValue
			for _, v := range enums {
				fieldValue = append(fieldValue, FieldValue{
					Value:    fmt.Sprint(v),
					Name:     fmt.Sprint(v),
					Disabled: false,
				})
			}

			field.Type = FieldTypeDropdown
			field.Values = fieldValue

			fields = append(fields, field)

			continue
		}

		if rType.Field(i).Type.Implements(mapperType) {
			maps := rValue.Field(i).Interface().(Mapper).Mapper()
			var fieldValue []FieldValue
			for k, v := range maps {
				fieldValue = append(fieldValue, FieldValue{
					Value:    k,
					Name:     v,
					Disabled: false,
				})
			}

			field.Type = FieldTypeDropdownMapped
			field.Values = fieldValue

			fields = append(fields, field)
			continue
		}

		//new addition to provide sorted key value pairs
		if rType.Field(i).Type.Implements(sortedMapperType) {
			maps := rValue.Field(i).Interface().(SortedMapper).SortedMapper()
			var fieldValue []FieldValue
			for _, v := range maps {
				fieldValue = append(fieldValue, FieldValue{
					Value:    v.Key(),
					Name:     v.Value(),
					Disabled: false,
				})
			}

			field.Type = FieldTypeDropdownMapped
			field.Values = fieldValue

			fields = append(fields, field)
			continue
		}

		inputType := InputFieldType(tags.Get("inputType"))

		fType := rType.Field(i).Type
		fValue := rValue.Field(i)

		if fValue.Kind() == reflect.Ptr && fValue.IsNil() {
			fValue = reflect.New(fValue.Type().Elem()).Elem()
			fType = fValue.Type()
		}

		if fValue.Kind() == reflect.Ptr || fValue.Kind() == reflect.Interface {
			fValue = fValue.Elem()
			fType = fType.Elem()
		}

		switch fType.Kind() {
		case reflect.String:

			if inputType == "" {
				inputType = InputFieldTypeText
			}

			field.Type = FieldTypeInput
			field.InputType = inputType

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:

			if inputType == "" {
				inputType = InputFieldTypeNumber
			}

			field.Type = FieldTypeInput
			field.InputType = inputType
			field.Step = "1"

		case reflect.Float32, reflect.Float64:
			if inputType == "" {
				inputType = InputFieldTypeNumber
			}

			field.Type = FieldTypeInput
			field.InputType = inputType
			field.Step = "any"

		case reflect.Bool:
			fieldType := FieldTypeCheckbox
			if len(names) > 0 && names[len(names)-1] == name {
				// radio-options use the same 'name' as their parent for grouping
				fieldType = FieldTypeRadios
			}

			field.Type = fieldType
		case reflect.Slice, reflect.Array:
		case reflect.Map:
		case reflect.Struct:
			field.Type = FieldTypeGroup
			field.Legend = tags.Get("legend")

			var err error
			field.Fields, err = t.scanModel(fValue, fType, nname...)
			if err != nil {
				return nil, err
			}
		}

		fields = append(fields, field)
	}

	return fields, nil
}

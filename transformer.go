package form

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/donseba/go-form/types"
)

const (
	tagName        = "name"
	tagLabel       = "label"
	tagPlaceholder = "placeholder"
	tagRequired    = "required"
	tagInputType   = "inputType"
	tagLegend      = "legend"
	tagType        = "type"
	tagStep        = "step"
	tagRows        = "rows"
	tagCols        = "cols"
)

type (
	Enumerator interface{ Enum() []any }
	Mapper     interface {
		Mapper() map[string]string
		String() string
	}

	// SortedMapper is a new addition to provide sorted key value pairs
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

	// Check for form metadata
	if rType.Field(0).Anonymous && rType.Field(0).Type == reflect.TypeOf(Info{}) {
		info := rValue.Field(0).Interface().(Info)
		formField := FormField{
			Type:   types.FieldTypeForm,
			Target: info.Target,
			Method: info.Method,
		}
		fields = append(fields, formField)
	}

	for i := 0; i < rType.NumField(); i++ {
		// Skip the Info field as we've already processed it
		if i == 0 && rType.Field(i).Anonymous && rType.Field(i).Type == reflect.TypeOf(Info{}) {
			continue
		}

		tags := rType.Field(i).Tag

		name := tags.Get(tagName)
		fieldName := rType.Field(i).Name
		if name == "" {
			name = fieldName
		}

		nname := append(names, name)
		field := FormField{
			Label:       tags.Get(tagLabel),
			Placeholder: tags.Get(tagPlaceholder),
			Name:        strings.Join(nname, "."),
			Value:       rValue.Field(i).Interface(),
		}

		if field.Label == "" {
			field.Label = name
		}

		if tags.Get(tagRequired) == "true" {
			field.Required = true
		}

		if tags.Get("primary") == "true" {
			field.Hidden = true
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

			field.Type = types.FieldTypeDropdown
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

			field.Type = types.FieldTypeDropdownMapped
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

			field.Type = types.FieldTypeDropdownMapped
			field.Values = fieldValue

			fields = append(fields, field)
			continue
		}

		inputType := types.InputFieldType(tags.Get(tagInputType))

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
				inputType = types.InputFieldTypeText
			}

			typ := types.FieldType(tags.Get(tagType))
			if typ == "" {
				typ = types.FieldTypeInput
			}

			field.Type = typ
			field.InputType = inputType
			if tags.Get(tagRows) != "" {
				field.Rows = tags.Get(tagRows)
			}
			if tags.Get(tagCols) != "" {
				field.Cols = tags.Get(tagCols)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:

			if inputType == "" {
				inputType = types.InputFieldTypeNumber
			}

			field.Type = types.FieldTypeInput
			field.InputType = inputType

			if tags.Get(tagStep) != "" {
				field.Step = tags.Get(tagStep)
			} else {
				field.Step = "1"
			}
		case reflect.Float32, reflect.Float64:
			if inputType == "" {
				inputType = types.InputFieldTypeNumber
			}

			field.Type = types.FieldTypeInput
			field.InputType = inputType

			if tags.Get(tagStep) != "" {
				field.Step = tags.Get(tagStep)
			} else {
				field.Step = "any"
			}
		case reflect.Bool:
			fieldType := types.FieldTypeCheckbox
			if len(names) > 0 && names[len(names)-1] == name {
				// radio-options use the same 'name' as their parent for grouping
				fieldType = types.FieldTypeRadios
			}

			field.Type = fieldType

			// replace the last slice element with the field name
			nname[len(nname)-1] = fieldName
			field.Id = strings.Join(nname, ".")
		case reflect.Slice, reflect.Array:
		case reflect.Map:
		case reflect.Struct:
			field.Type = types.FieldTypeGroup
			field.Legend = tags.Get(tagLegend)
			if field.Legend == "" {
				field.Legend = field.Label
			}

			var err error
			field.Fields, err = t.scanModel(fValue, fType, nname...)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unsupported type: %s", fType.Kind())
		}

		fields = append(fields, field)
	}

	return fields, nil
}

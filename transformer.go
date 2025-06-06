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
	tagLegend      = "legend"
	tagValues      = "values"
	tagForm        = "form"
	tagGroup       = "group"
	tagStep        = "step"
	tagRows        = "rows"
	tagCols        = "cols"
	tagMin         = "min"
	tagMax         = "max"
	tagMaxLength   = "maxLength"
	tagDescription = "description"
)

var DefaultSubmitText = "Submit"

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
	Fields []types.FormField `json:"fields"`
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
func (t *Transformer) scanModel(rValue reflect.Value, rType reflect.Type, names ...string) ([]types.FormField, error) {
	var fields []types.FormField

	// Check for form metadata
	if rType.Field(0).Anonymous && rType.Field(0).Type == reflect.TypeOf(Info{}) {
		info := rValue.Field(0).Interface().(Info)

		label := info.SubmitText
		if label == "" {
			label = DefaultSubmitText
		}

		formField := types.FormField{
			Type:   types.FieldTypeForm,
			Target: info.Target,
			Method: info.Method,
			Label:  label,
		}

		for k, v := range info.Attributes {
			if formField.Attributes == nil {
				formField.Attributes = make(map[string]string)
			}
			formField.Attributes[k] = v
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
		field := types.FormField{
			Id:          strings.Join(nname, "."),
			Label:       tags.Get(tagLabel),
			Placeholder: tags.Get(tagPlaceholder),
			Description: tags.Get(tagDescription),
			Name:        strings.Join(nname, "."),
			Value:       rValue.Field(i).Interface(),
		}

		groupTag := tags.Get(tagGroup)
		if groupTag != "" {
			if strings.Contains(groupTag, ",") {
				parts := strings.SplitN(groupTag, ",", 2)
				field.GroupBefore = strings.TrimSpace(parts[0])
				if len(parts) > 1 {
					field.GroupAfter = strings.TrimSpace(parts[1])
				}
			} else {
				field.GroupBefore = groupTag
			}
		}

		formTag := tags.Get(tagForm)
		if strings.Contains(formTag, ",") {
			// If the tag contains a comma, it is a form field
			field.Type = types.FieldType(strings.Split(formTag, ",")[0])
			field.InputType = types.InputFieldType(strings.Split(formTag, ",")[1])
		} else {
			// Otherwise, it is a regular input field
			field.Type = types.FieldType(formTag)
		}

		if field.Label == "" {
			field.Label = name
		}

		if tags.Get(tagRequired) == "true" {
			field.Required = true
		}

		if rType.Field(i).Type.Implements(enumType) {
			enums := reflect.New(rType.Field(i).Type).Interface().(Enumerator).Enum()
			var fieldValue []types.FieldValue
			for _, v := range enums {
				fieldValue = append(fieldValue, types.FieldValue{
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

		if rType.Field(i).Tag.Get(tagValues) != "" {
			// If the tag contains values, it is a dropdown or radio field
			values := strings.Split(rType.Field(i).Tag.Get(tagValues), ";")
			var fieldValue []types.FieldValue
			for _, v := range values {
				if strings.Contains(v, ":") {
					parts := strings.SplitN(v, ":", 2)
					if len(parts) != 2 {
						return nil, fmt.Errorf("invalid value format in tag %s for field %s", tagValues, fieldName)
					}
					fieldValue = append(fieldValue, types.FieldValue{
						Value:    strings.TrimSpace(parts[0]),
						Name:     strings.TrimSpace(parts[1]),
						Disabled: false,
					})
					continue
				}

				fieldValue = append(fieldValue, types.FieldValue{
					Value:    v,
					Name:     v,
					Disabled: false,
				})
			}

			// Set the field type based on the form tag
			if field.Type == types.FieldTypeRadios {
				field.Values = fieldValue
			} else if field.Type == types.FieldTypeDropdown {
				field.Values = fieldValue
			} else {
				field.Type = types.FieldTypeDropdown
				field.Values = fieldValue
			}

			fields = append(fields, field)
			continue
		}

		if rType.Field(i).Type.Implements(mapperType) {
			maps := rValue.Field(i).Interface().(Mapper).Mapper()
			var fieldValue []types.FieldValue
			for k, v := range maps {
				fieldValue = append(fieldValue, types.FieldValue{
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
			var fieldValue []types.FieldValue
			for _, v := range maps {
				fieldValue = append(fieldValue, types.FieldValue{
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
			if tags.Get(tagRows) != "" {
				field.Rows = tags.Get(tagRows)
			}
			if tags.Get(tagCols) != "" {
				field.Cols = tags.Get(tagCols)
			}
			if tags.Get(tagMaxLength) != "" {
				field.MaxLength = tags.Get(tagMaxLength)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:

			if field.InputType == "" {
				field.InputType = types.InputFieldTypeNumber
			}

			field.Type = types.FieldTypeInput

			if tags.Get(tagStep) != "" {
				field.Step = tags.Get(tagStep)
			} else {
				field.Step = "1"
			}
			if tags.Get(tagMin) != "" {
				field.Min = tags.Get(tagMin)
			}
			if tags.Get(tagMax) != "" {
				field.Max = tags.Get(tagMax)
			}
		case reflect.Float32, reflect.Float64:
			if field.InputType == "" {
				field.InputType = types.InputFieldTypeNumber
			}

			field.Type = types.FieldTypeInput

			if tags.Get(tagStep) != "" {
				field.Step = tags.Get(tagStep)
			} else {
				field.Step = "any"
			}
			if tags.Get(tagMin) != "" {
				field.Min = tags.Get(tagMin)
			}
			if tags.Get(tagMax) != "" {
				field.Max = tags.Get(tagMax)
			}
		case reflect.Bool:
			fieldType := types.FieldTypeCheckbox
			if len(names) > 0 && names[len(names)-1] == name {
				// radio-options use the same 'name' as their parent for grouping
				fieldType = types.FieldTypeRadios
				field.InputType = types.InputFieldTypeRadioStruct // Set the new input type for struct-based radio buttons
			}

			field.Type = fieldType

			// replace the last slice element with the field name
			nname[len(nname)-1] = fieldName
			field.Id = strings.Join(nname, ".")

			// For radio buttons, set the value to the field name if checked
			if fieldType == types.FieldTypeRadios {
				if rValue.Field(i).Bool() {
					field.Value = fieldName
				} else {
					field.Value = "" // Set empty value when false
				}
			}
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

package form

import (
	"encoding/json"
	"fmt"
	"math"
	"mime/multipart"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/donseba/go-form/v2/types"
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
	tagClass       = "class"
	tagMaxLength   = "maxLength"
	tagDescription = "description"
	tagData        = "data"
	tagNot         = "not"
	// Allow disabling fields via struct tags.
	tagDisabled = "disabled"
	// Enable translation support for enum values
	tagTranslate = "translate"

	tagReadOnly     = "readonly"
	tagAutoComplete = "autocomplete"
	tagAccept       = "accept"
	tagMultiple     = "multiple"
	tagMinLength    = "minLength"
	tagPattern      = "pattern"
	// tagWallClock marks a datetime-local field whose time is a time of
	// day on the clock, read and shown without a time zone.
	tagWallClock = "wallclock"
)

var (
	DefaultSubmitText      = "Submit"
	DefaultEnumTranslation = false // Set to true to enable translation for all enums by default
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
	Fields []types.FormField `json:"fields"`

	// location is the time zone date-times are shown in; nil shows them in
	// their own zone.
	location *time.Location
	// overrides replace struct tags of fields, by the name they render
	// with (Overrides).
	overrides Overrides
	// options are the choices of fields given at render time, by the name
	// they render with (FieldOptions).
	options FieldOptions
}

func NewTransformer(model interface{}) (*Transformer, error) {
	return newTransformer(model, nil)
}

// newTransformer is NewTransformer showing date-times in location (when not
// nil).
func newTransformer(model interface{}, location *time.Location) (*Transformer, error) {
	var (
		renderInfo *Info
		overrides  Overrides
		options    FieldOptions
	)

	if wrapped, ok := model.(*RenderModel); ok && wrapped == nil {
		return nil, fmt.Errorf("form model must not be nil")
	}

	if wrapped, ok := renderModelOf(model); ok {
		model = wrapped.Model
		renderInfo = &wrapped.Info
		overrides = wrapped.Overrides
		options = wrapped.Options
	}

	// A form defined at runtime.
	if fields, ok, err := schemaFields(model); ok {
		if err != nil {
			return nil, err
		}

		if renderInfo != nil {
			fields = append([]types.FormField{formFieldFromInfo(*renderInfo)}, fields...)
		}

		return &Transformer{Fields: fields, location: location}, nil
	}

	modelValue := reflect.ValueOf(model)
	if !modelValue.IsValid() {
		return nil, fmt.Errorf("form model must not be nil")
	}
	modelType := modelValue.Type()

	// check if we received a pointer
	for modelValue.Kind() == reflect.Ptr || modelValue.Kind() == reflect.Interface {
		if modelValue.IsNil() {
			return nil, fmt.Errorf("form model must not be nil")
		}
		modelValue = modelValue.Elem()
		modelType = modelValue.Type()
	}
	if modelValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("form model must be a struct, got %s", modelValue.Kind())
	}

	tr := &Transformer{
		location:  location,
		overrides: overridesOf(modelValue).merge(overrides),
		options:   options,
	}
	fields, err := tr.scanModel(modelValue, modelType)
	if err != nil {
		return nil, err
	}
	if renderInfo != nil {
		fields = removeFormFields(fields)
		fields = append([]types.FormField{formFieldFromInfo(*renderInfo)}, fields...)
	}

	tr.Fields = collapseStructRadioGroups(fields)

	return tr, nil
}

func removeFormFields(fields []types.FormField) []types.FormField {
	out := fields[:0]
	for _, field := range fields {
		if field.Type != types.FieldTypeForm {
			out = append(out, field)
		}
	}
	return out
}

func formFieldFromInfo(info Info) types.FormField {
	label := info.SubmitText
	if label == "" {
		label = DefaultSubmitText
	}
	field := types.FormField{
		Type:         types.FieldTypeForm,
		Target:       info.Target,
		Method:       info.Method,
		Label:        label,
		CancelTarget: info.CancelTarget,
		CancelText:   info.CancelText,
		ReadOnly:     info.ReadOnly,
		Honeypot:     info.Honeypot,
	}
	for key, value := range info.Attributes {
		if field.Attributes == nil {
			field.Attributes = make(map[string]string)
		}
		field.Attributes[key] = value
	}
	return field
}

func (t *Transformer) JSON() json.RawMessage {
	out, _ := json.Marshal(t)

	return out
}

// scanModel the incoming interface and ensure we can work with it.
func (t *Transformer) scanModel(rValue reflect.Value, rType reflect.Type, names ...string) ([]types.FormField, error) {
	var fields []types.FormField

	// Check for form metadata
	if rType.NumField() > 0 && rType.Field(0).Anonymous && rType.Field(0).Type == reflect.TypeOf(Info{}) {
		info := rValue.Field(0).Interface().(Info)
		fields = append(fields, formFieldFromInfo(info))
	}

	for i := 0; i < rType.NumField(); i++ {
		// Skip the Info field as we've already processed it
		if i == 0 && rType.Field(i).Anonymous && rType.Field(i).Type == reflect.TypeOf(Info{}) {
			continue
		}

		// Unexported fields are the model's own, such as a limit an
		// Overrider uses; they never render.
		if !rType.Field(i).IsExported() {
			continue
		}

		name := rType.Field(i).Tag.Get(tagName)
		fieldName := rType.Field(i).Name
		if name == "" {
			name = fieldName
		}

		nname := append(names, name)

		// The tags with the overrides of this render.
		tags := t.overrides.apply(strings.Join(nname, "."), rType.Field(i)).Tag
		field := types.FormField{
			Id:          strings.Join(nname, "."),
			Label:       tags.Get(tagLabel),
			Placeholder: tags.Get(tagPlaceholder),
			Description: tags.Get(tagDescription),
			Name:        strings.Join(nname, "."),
			Class:       tags.Get(tagClass),
			Value:       rValue.Field(i).Interface(),
		}

		// Process data attributes if present
		if dataTag := tags.Get(tagData); dataTag != "" {
			field.Data = make(map[string]string)

			// Improved parser for data attributes that handles complex values
			var i, start int
			var inQuote bool
			var inBrace int

			for i < len(dataTag) {
				switch dataTag[i] {
				case '"':
					// Toggle quote state if not escaped
					if i == 0 || dataTag[i-1] != '\\' {
						inQuote = !inQuote
					}
				case '{':
					if !inQuote {
						inBrace++
					}
				case '}':
					if !inQuote {
						inBrace--
					}
				case ',':
					// Only split on commas that are not inside quotes or braces
					if !inQuote && inBrace == 0 {
						// Process the attribute
						attr := dataTag[start:i]
						if strings.Contains(attr, "=") {
							parts := strings.SplitN(attr, "=", 2)
							if len(parts) == 2 {
								key := strings.TrimSpace(parts[0])
								value := strings.TrimSpace(parts[1])
								field.Data[key] = value
							}
						}
						start = i + 1 // Move past the comma
					}
				}
				i++
			}

			// Process the last attribute
			if start < len(dataTag) {
				attr := dataTag[start:]
				if strings.Contains(attr, "=") {
					parts := strings.SplitN(attr, "=", 2)
					if len(parts) == 2 {
						key := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(parts[1])
						field.Data[key] = value
					}
				}
			}
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
			parts := strings.Split(formTag, ",")
			field.Type = types.FieldType(parts[0])
			field.InputType = types.InputFieldType(parts[1])
		} else {
			field.Type = types.FieldType(formTag)
		}

		// Special case: explicit declaration of a struct-based radio group.
		// `form:"radios,radio_group"` means: the Go field is a struct group, but it should render as a single radio group.
		if field.Type == types.FieldTypeRadios && field.InputType == types.InputFieldTypeRadioGroup {
			field.Type = types.FieldTypeGroup
		}

		if field.Label == "" {
			field.Label = name
		}

		if tags.Get(tagRequired) == "true" {
			field.Required = true
		}
		if tags.Get(tagDisabled) == "true" {
			field.Disabled = true
		}

		field.ReadOnly = tags.Get(tagReadOnly) == "true"
		field.AutoComplete = tags.Get(tagAutoComplete)
		field.Accept = tags.Get(tagAccept)
		field.Multiple = tags.Get(tagMultiple) == "true"

		// Check if translation is enabled: struct tag takes precedence over global default
		tagValue := tags.Get(tagTranslate)
		shouldTranslate := tagValue == "true" || (tagValue != "false" && DefaultEnumTranslation)

		// Options given at render time replace those of the tags.
		if choices, ok := t.options[field.Name]; ok {
			withChoices(&field, rValue.Field(i), choices)
			fields = append(fields, field)

			continue
		}

		if rType.Field(i).Type.Implements(enumType) {
			enums := reflect.New(rType.Field(i).Type).Interface().(Enumerator).Enum()
			var fieldValue []types.FieldValue

			var typeName string
			if shouldTranslate {
				typeName = rType.Field(i).Type.Name()
				if typeName == "" {
					// Fallback for unnamed types (aliases, pointers, etc.)
					typeName = strings.ReplaceAll(rType.Field(i).Type.String(), ".", "_")
					typeName = strings.ReplaceAll(typeName, "*", "")
				}
			}

			for _, v := range enums {
				val := fmt.Sprint(v)
				enumName := val
				if shouldTranslate {
					enumName = fmt.Sprintf("enum||%s.%s", typeName, val)
				}
				fieldValue = append(fieldValue, types.FieldValue{
					Value:     val,
					Name:      enumName,
					Disabled:  false,
					Translate: shouldTranslate,
				})
			}

			field.Type = types.FieldTypeDropdown
			field.Values = fieldValue

			fields = append(fields, field)

			continue
		}

		if values := tags.Get(tagValues); values != "" {
			// Labels written in a values tag are text of the application,
			// like field labels, so they are translated unless
			// translate:"false".
			withChoices(&field, rValue.Field(i), parseValuesTag(values, tagValue != "false"))
			fields = append(fields, field)

			continue
		}

		if rType.Field(i).Type.Implements(mapperType) {
			maps := rValue.Field(i).Interface().(Mapper).Mapper()
			var fieldValue []types.FieldValue

			for k, v := range maps {
				fieldValue = append(fieldValue, types.FieldValue{
					Value:     k,
					Name:      v,
					Disabled:  false,
					Translate: shouldTranslate,
				})
			}

			field.Type = types.FieldTypeDropdownMapped
			field.Values = fieldValue
			// Set Value as string for template eq compatibility
			field.Value = fmt.Sprint(rValue.Field(i).Interface())

			fields = append(fields, field)
			continue
		}

		//new addition to provide sorted key value pairs
		if rType.Field(i).Type.Implements(sortedMapperType) {
			maps := rValue.Field(i).Interface().(SortedMapper).SortedMapper()
			var fieldValue []types.FieldValue

			for _, v := range maps {
				fieldValue = append(fieldValue, types.FieldValue{
					Value:     v.Key(),
					Name:      v.Value(),
					Disabled:  false,
					Translate: shouldTranslate,
				})
			}

			// Use multicheckbox if form tag is multicheckbox, else fallback to dropdownmapped
			if field.Type == types.FieldTypeMultiCheckbox || tags.Get(tagForm) == "multicheckbox" {
				field.Type = types.FieldTypeMultiCheckbox
				// For checkboxes, Value should be a map[string]bool for checked state
				valueMap := map[string]bool{}
				if rValue.Field(i).IsValid() && rValue.Field(i).CanInterface() {
					if sms, ok := rValue.Field(i).Interface().(MultiSelectGetter); ok {
						for _, v := range sms.GetKeysAsStrings() {
							valueMap[v] = true
						}
					}
				}
				field.ValueMap = valueMap
			} else {
				field.Type = types.FieldTypeDropdownMapped
				field.Value = fmt.Sprint(rValue.Field(i).Interface())
			}

			field.Values = fieldValue

			field.SortByLabel = sortsByLabel(rValue.Field(i), shouldTranslate)

			fields = append(fields, field)

			continue
		}

		// Uploaded files: *multipart.FileHeader or []*multipart.FileHeader.
		if single, ok := fileField(rType.Field(i).Type); ok {
			fields = append(fields, fileInput(field, single))

			continue
		}

		// check if time.Time or time.Time pointer
		if rType.Field(i).Type == reflect.TypeOf(time.Time{}) || rType.Field(i).Type == reflect.TypeOf(&time.Time{}) {
			var elem time.Time
			if rType.Field(i).Type == reflect.TypeOf(&time.Time{}) {
				if rValue.Field(i).IsNil() {
					elem = time.Time{}
				} else {
					elem = rValue.Field(i).Elem().Interface().(time.Time)
				}
			} else {
				elem = rValue.Field(i).Interface().(time.Time)
			}

			switch {
			case elem.IsZero():
				// An unset time shows an empty input instead of 0001-01-01.
				field.Value = ""
			case field.InputType == types.InputFieldTypeDate:
				field.Value = elem.Format(time.DateOnly)
			case field.InputType == types.InputFieldTypeTime:
				field.Value = elem.Format(time.TimeOnly)
			case field.InputType == types.InputFieldTypeDateTimeLocal:
				field.Value = t.dateTimeLocalValue(elem, tags)
			case field.InputType == types.InputFieldTypeMonth:
				field.Value = elem.Format(monthLayout)
			case field.InputType == types.InputFieldTypeWeek:
				// For week, we use the ISO week date format
				year, week := elem.ISOWeek()
				field.Value = fmt.Sprintf("%d-W%02d", year, week)
			default:
				field.Value = elem.Format(time.DateOnly)
			}

			if tags.Get(tagStep) != "" {
				field.Step = tags.Get(tagStep)
			} else {
				field.Step = "60"
			}

			fields = append(fields, field)

			continue
		}

		fType := rType.Field(i).Type
		fValue := rValue.Field(i)

		field.Value = shownValue(fValue, field.Value)

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
			field.MinLength = tags.Get(tagMinLength)
			field.Pattern = clientPattern(tags.Get(tagPattern))

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

			// A nested model's own overrides yield to the outer ones.
			t.overrides = overridesOf(fValue).prefixed(strings.Join(nname, ".")).merge(t.overrides)

			var err error
			field.Fields, err = t.scanModel(fValue, fType, nname...)
			if err != nil {
				return nil, err
			}

			// If the struct field itself is declared as radios, treat contained bool fields
			// as radio options and build the Values slice for the radio-group template.
			if tags.Get(tagForm) == string(types.FieldTypeRadios) {
				field.Type = types.FieldTypeRadios
				field.InputType = types.InputFieldTypeRadioStruct

				// The value is the chosen option, none until one is
				// ticked; MapForm reads it back (see mapRadioGroup).
				field.Value = ""

				for _, sub := range field.Fields {
					v := fmt.Sprint(sub.Value)
					// Option labels are the labels of the bool fields, which
					// are translated like any other label.
					field.Values = append(field.Values, types.FieldValue{
						Value:     sub.Id,
						Name:      sub.Label,
						Disabled:  sub.Disabled,
						Translate: true,
					})
					if v != "" {
						field.Value = sub.Id
					}
				}
			}
		default:
			return nil, fmt.Errorf("unsupported type: %s", fType.Kind())
		}
		if field.Type == "" {
			field.Type = types.FieldTypeInput
			field.InputType = types.InputFieldTypeText
		}

		fields = append(fields, field)
	}

	return fields, nil
}

// NOTE: collapseStructRadioGroups is implemented in transformer_radio_struct.go

// sortsByLabel reports whether the options of a select, value, are sorted
// by their labels when the form renders, in the language of the render:
// when the select sorts by label (SortByLabel) and its labels are
// translated.
func sortsByLabel(value reflect.Value, translated bool) bool {
	sorter, ok := value.Interface().(labelSorter)

	return ok && translated && sorter.SortedByLabel()
}

// fileInput is field as a file input for uploaded files: for one file when
// single is set, or for several. A file input cannot be filled in again, so
// it shows no value.
func fileInput(field types.FormField, single bool) types.FormField {
	if field.Type == "" || field.Type == types.FieldTypeInput {
		field.Type = types.FieldTypeInput
		field.InputType = types.InputFieldTypeFile
	}

	field.Value = ""
	if !single {
		field.Multiple = true
	}

	return field
}

// dateTimeLocalValue is elem as the value of a datetime-local input: in the
// zone of the localizer (see TimeLocalizer), in the format browsers accept,
// with seconds only when the step tag asks for them. A wall-clock field
// (wallclock:"true") shows the time as it is.
func (t *Transformer) dateTimeLocalValue(elem time.Time, tags reflect.StructTag) string {
	if t.location != nil && tags.Get(tagWallClock) != "true" {
		elem = elem.In(t.location)
	}

	return elem.Format(dateTimeLocalFormat(tags.Get(tagStep)))
}

// shownValue is what a field whose value is value shows: for a pointer to a
// string, bool or number the value it points to, or nothing for a nil one,
// and shown, the value as it is, for anything else.
func shownValue(value reflect.Value, shown any) any {
	pointed, isNil := pointee(value)
	if isNil {
		return ""
	}

	if pointed.Kind() != reflect.Pointer {
		return pointed.Interface()
	}

	return shown
}

// Layouts of the values of month and datetime-local inputs.
const (
	monthLayout                = "2006-01"
	dateTimeLocalLayout        = "2006-01-02T15:04"
	dateTimeLocalSecondsLayout = "2006-01-02T15:04:05"
)

// dateTimeLocalFormat is the layout of a datetime-local value: with seconds
// only when the step tag allows them, as browsers refuse seconds otherwise.
func dateTimeLocalFormat(step string) string {
	seconds, err := strconv.ParseFloat(step, 64)
	if err != nil || seconds <= 0 || math.Mod(seconds, 60) == 0 {
		return dateTimeLocalLayout
	}

	return dateTimeLocalSecondsLayout
}

var (
	fileHeaderType  = reflect.TypeOf(&multipart.FileHeader{})
	fileHeadersType = reflect.TypeOf([]*multipart.FileHeader{})
)

// fileField reports whether t holds uploaded files, and whether it holds one
// (*multipart.FileHeader) rather than several ([]*multipart.FileHeader).
func fileField(t reflect.Type) (single bool, ok bool) {
	switch t {
	case fileHeaderType:
		return true, true
	case fileHeadersType:
		return false, true
	}

	return false, false
}

// clientPattern is the pattern tag as the pattern attribute, when browsers
// check it the way the server does. Browsers match the whole value, while
// the server looks for a match anywhere, so only a pattern anchored with ^
// and $ means the same to both. Go-only syntax, such as flags or \A, would
// make browsers ignore the attribute, so such patterns stay on the server.
func clientPattern(pattern string) string {
	if !strings.HasPrefix(pattern, "^") || !strings.HasSuffix(pattern, "$") || strings.HasSuffix(pattern, `\$`) {
		return ""
	}

	for _, goOnly := range []string{"(?i", "(?m", "(?s", "(?U", "(?P<", `\A`, `\z`, "[[:", `\pL`, `\PL`} {
		if strings.Contains(pattern, goOnly) {
			return ""
		}
	}

	// ^a|b$ is "starts with a or ends with b" on the server, but "a or b"
	// in a browser.
	if topLevelAlternation(pattern) {
		return ""
	}

	return pattern
}

// topLevelAlternation reports whether pattern has a | outside groups and
// character classes.
func topLevelAlternation(pattern string) bool {
	depth := 0
	inClass := false

	for i := 0; i < len(pattern); i++ {
		switch c := pattern[i]; {
		case c == '\\':
			i++
		case inClass:
			inClass = c != ']'
		case c == '[':
			inClass = true
		case c == '(':
			depth++
		case c == ')':
			depth--
		case c == '|' && depth == 0:
			return true
		}
	}

	return false
}

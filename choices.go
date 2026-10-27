package form

import (
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/donseba/go-form/v2/types"
)

// FieldOptions are the choices of fields that are only known when the form
// renders, such as the options of a poll, by the field's name as it renders
// (its name tag or Go name, such as "option", or "address.country" in a
// nested struct). They replace the field's values tag:
//
//	type VoteInput struct {
//		Poll   string `form:"input,hidden" name:"poll"`
//		Option string `form:"radios" name:"option" label:"Your choice" required:"true"`
//	}
//
//	model := form.WithInfo(VoteInput{Poll: poll.ID}, info)
//	model.Options = form.FieldOptions{"option": choices} // []types.FieldValue
//
// A radios field renders them as a group of radio buttons, a multicheckbox
// field as a checkbox each, and any other field as a dropdown. Labels are
// shown as written unless an option sets Translate, so they may hold any
// text, ";" and ":" included. The field may be a string, an integer kind
// (int, uint8, …), a pointer to one, or for a multicheckbox a slice of
// them. Validation with the same RenderModel (ValidateForm,
// ValidateFormLocalized and Bind) checks that a submitted value is one of
// the options and not disabled, with the rule "values".
type FieldOptions map[string][]types.FieldValue

// offered is the set of the values of choices a submission may hold: those
// that are not disabled.
func offered(choices []types.FieldValue) map[string]bool {
	allowed := make(map[string]bool, len(choices))

	for _, choice := range choices {
		if !choice.Disabled {
			allowed[choice.Value] = true
		}
	}

	return allowed
}

// parseValuesTag reads a values tag, such as "a:Option A;b:Option B" or
// "a;b", into options. translate is whether their labels are translated.
func parseValuesTag(tag string, translate bool) []types.FieldValue {
	var choices []types.FieldValue

	for _, entry := range strings.Split(tag, ";") {
		value, label, labelled := strings.Cut(entry, ":")
		value = strings.TrimSpace(value)

		if labelled {
			label = strings.TrimSpace(label)
		} else {
			label = value
		}

		choices = append(choices, types.FieldValue{
			Value:     value,
			Name:      label,
			Translate: translate,
		})
	}

	return choices
}

// withChoices makes field a control with the options choices, showing the
// value of the struct field value: a group of radio buttons for a radios
// field, a checkbox per option for a multicheckbox field and a dropdown
// for any other.
func withChoices(field *types.FormField, value reflect.Value, choices []types.FieldValue) {
	field.Values = slices.Clone(choices)

	switch field.Type {
	case types.FieldTypeRadios:
		field.InputType = types.InputFieldTypeRadioGroup
		field.Value = choiceValue(value)
	case types.FieldTypeMultiCheckbox:
		field.InputType = ""
		field.ValueMap = chosenKeys(value)
	default:
		field.Type = types.FieldTypeDropdown
		field.InputType = ""
		field.Value = choiceValue(value)
	}
}

// choiceValue is value as the option values it is compared with: the text
// of a string or integer (also behind a pointer), "" for a nil pointer, or
// the value as it is.
func choiceValue(value reflect.Value) any {
	value, isNil := pointee(value)
	if isNil {
		return ""
	}

	if key, ok := scalarKey(value); ok {
		return key
	}

	if !value.IsValid() || !value.CanInterface() {
		return ""
	}

	return value.Interface()
}

// chosenKeys are the options value holds, a slice of strings or integers
// (or one of them), as the set a multicheckbox shows ticked.
func chosenKeys(value reflect.Value) map[string]bool {
	chosen := map[string]bool{}

	value, isNil := pointee(value)
	if isNil || !value.IsValid() {
		return chosen
	}

	if key, ok := scalarKey(value); ok {
		if key != "" {
			chosen[key] = true
		}

		return chosen
	}

	if value.Kind() != reflect.Slice && value.Kind() != reflect.Array {
		return chosen
	}

	for i := 0; i < value.Len(); i++ {
		if key, ok := scalarKey(value.Index(i)); ok {
			chosen[key] = true
		}
	}

	return chosen
}

// scalarKey is the text of a string or integer value, as a form submits
// it. ok is false for other kinds.
func scalarKey(value reflect.Value) (key string, ok bool) {
	switch value.Kind() {
	case reflect.String:
		return value.String(), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(value.Int(), 10), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(value.Uint(), 10), true
	}

	return "", false
}

// pointee is the value a pointer to a string, bool or number points to,
// and whether that pointer is nil. Other values are returned as they are.
func pointee(value reflect.Value) (reflect.Value, bool) {
	if value.Kind() != reflect.Pointer || !scalarKind(value.Type().Elem().Kind()) {
		return value, false
	}

	if value.IsNil() {
		return value, true
	}

	return value.Elem(), false
}

// scalarKind reports whether kind is a string, bool or number kind.
func scalarKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}

	return false
}

// unoffered returns a value of value, a string, an integer or a slice of
// them, that allowed does not hold. An empty string means nothing is
// chosen, and so does 0 for an integer field unless 0 is an option or
// zeroChosen is set, for the value of a pointer, whose nil means nothing
// is chosen: required:"true" reports those.
func unoffered(value reflect.Value, allowed map[string]bool, zeroChosen bool) (string, bool) {
	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		if value.Type().Elem().Kind() == reflect.Uint8 {
			return "", false
		}

		for i := 0; i < value.Len(); i++ {
			key, ok := scalarKey(value.Index(i))
			if !ok || key == "" {
				continue
			}

			if !allowed[key] {
				return key, true
			}
		}

		return "", false
	}

	key, ok := scalarKey(value)
	if !ok || key == "" || allowed[key] {
		return "", false
	}

	if key == "0" && value.Kind() != reflect.String && !zeroChosen {
		return "", false
	}

	return key, true
}

// choiceErrors is the error of a value of value, the value of field (that
// of a pointer field points to), that allowed does not hold, with the
// rule "values".
func choiceErrors(field reflect.StructField, value reflect.Value, allowed map[string]bool, getErr messageFunc) FieldErrors {
	key, invalid := unoffered(value, allowed, field.Type != nil && field.Type.Kind() == reflect.Pointer)
	if !invalid {
		return nil
	}

	return FieldErrors{ruleError(field, getErr, "values", TranslationKeyInvalidValue, key)}
}

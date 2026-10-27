package form

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strings"

	"github.com/donseba/go-form/v2/types"
)

// Carry is a submission that failed validation, kept to show the form again
// on another request, such as the page a redirect after the post leads to:
// the submitted values of the form's fields and the errors per field.
//
//	carry, err := form.NewCarry(SignupInput{}, r.PostForm, errs)
//	data, err := carry.Encode(2048) // into a cookie, encrypted
//
// and on the page the visitor returns to:
//
//	carry, err := form.DecodeCarry(data)
//	model := form.WithInfo(SignupInput{}, info)
//	model.Carry = &carry
//
// form_render and form_field then show the carried values in place of the
// model's, and the carried errors at their fields when the template passes
// no errors of its own. A Carry never holds the value of a password input,
// a file input or the form's honeypot (Info.Honeypot), nor values that are
// not a field of the form, such as a CSRF token. Values can be personal
// data, such as an e-mail address: keep the encoding out of reach of
// others, for example encrypted.
type Carry struct {
	// Values are the submitted values, by field name as the field renders
	// (such as "address.city").
	Values url.Values `json:"v,omitempty"`
	// Errors are the messages per field name.
	Errors map[string][]string `json:"e,omitempty"`
}

// ErrCarryTooLarge is returned by Carry.Encode when the errors alone do not
// fit in the limit.
var ErrCarryTooLarge = errors.New("form: carry exceeds its limit")

// NewCarry keeps a failed submission of model, a struct with go-form's tags,
// a RenderModel or a Schema: the values of its fields in values, such as
// r.PostForm, and the errors errs. Values that are not a field of the
// model, the values of password and file inputs and the honeypot of a
// RenderModel are left out.
func NewCarry(model any, values url.Values, errs FieldErrors) (Carry, error) {
	tr, err := NewTransformer(model)
	if err != nil {
		return Carry{}, err
	}

	honeypot := ""
	if info, ok := modelInfo(model); ok {
		honeypot = info.Honeypot
	}

	carry := Carry{}

	for _, name := range carriedNames(tr.Fields) {
		submitted, ok := values[name]
		if !ok || name == honeypot {
			continue
		}

		if carry.Values == nil {
			carry.Values = url.Values{}
		}

		carry.Values[name] = slices.Clone(submitted)
	}

	for _, fieldErr := range errs {
		field, message := fieldErr.FieldError()

		if carry.Errors == nil {
			carry.Errors = map[string][]string{}
		}

		carry.Errors[field] = append(carry.Errors[field], message)
	}

	return carry, nil
}

// carriedNames are the names of the fields whose values a Carry keeps:
// every field with a value, also nested ones, but no password or file
// inputs.
func carriedNames(fields []types.FormField) []string {
	var names []string

	for _, field := range fields {
		switch {
		case field.Type == types.FieldTypeForm:
			continue
		case field.Type == types.FieldTypeGroup:
			names = append(names, carriedNames(field.Fields)...)

			continue
		case field.InputType == types.InputFieldTypePassword, field.InputType == types.InputFieldTypeFile:
			continue
		}

		if field.Name != "" {
			names = append(names, field.Name)
		}
	}

	return names
}

// Encode returns the carry as compact JSON of at most limit bytes (no limit
// when limit is zero or less). When it does not fit, the values that take
// the most room are left out, the longest first, until it does: such a
// field shows the model's value again, with its errors. When the errors
// alone do not fit, Encode returns ErrCarryTooLarge.
func (c Carry) Encode(limit int) ([]byte, error) {
	kept := Carry{Values: url.Values{}, Errors: c.Errors}
	maps.Copy(kept.Values, c.Values)

	for {
		data, err := marshalCarry(kept)
		if err != nil {
			return nil, err
		}

		if limit <= 0 || len(data) <= limit {
			return data, nil
		}

		longest := longestValue(kept.Values)
		if longest == "" {
			return nil, fmt.Errorf("%w: %d bytes of errors, limit %d", ErrCarryTooLarge, len(data), limit)
		}

		delete(kept.Values, longest)
	}
}

// marshalCarry is the JSON of c, with <, > and & as they are.
func marshalCarry(c Carry) ([]byte, error) {
	if len(c.Values) == 0 {
		c.Values = nil
	}

	var buffer bytes.Buffer

	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(c); err != nil {
		return nil, err
	}

	return bytes.TrimSuffix(buffer.Bytes(), []byte("\n")), nil
}

// longestValue is the name whose values are longest together, the first by
// name among equals, or "" when there are none.
func longestValue(values url.Values) string {
	longest, size := "", -1

	for _, name := range slices.Sorted(maps.Keys(values)) {
		length := 0
		for _, value := range values[name] {
			length += len(value)
		}

		if length > size {
			longest, size = name, length
		}
	}

	return longest
}

// DecodeCarry reads a carry that Encode returned.
func DecodeCarry(data []byte) (Carry, error) {
	var carry Carry

	if err := json.Unmarshal(data, &carry); err != nil {
		return Carry{}, fmt.Errorf("form: decode carry: %w", err)
	}

	return carry, nil
}

// FieldErrors are the carried errors, by field name in order.
func (c Carry) FieldErrors() FieldErrors {
	var errs FieldErrors

	for _, field := range slices.Sorted(maps.Keys(c.Errors)) {
		for _, message := range c.Errors[field] {
			errs = append(errs, FieldValidationError{Field: field, Err: message})
		}
	}

	return errs
}

// matches reports whether the carry belongs to the form with fields: its
// hidden inputs hold the values the carry was submitted with, so that of
// several forms of one kind on a page, such as a comment form per post,
// only the one that was sent shows the carry.
func (c Carry) matches(fields []types.FormField) bool {
	for _, field := range fields {
		if !c.matches(field.Fields) {
			return false
		}

		if field.InputType != types.InputFieldTypeHidden {
			continue
		}

		submitted, ok := c.Values[field.Name]
		if !ok {
			continue
		}

		if first(submitted) != fmt.Sprint(valueOrEmpty(field.Value)) {
			return false
		}
	}

	return true
}

// restore shows the carried values in fields, in place of the model's.
// Hidden inputs keep the model's values; password and file inputs show
// none, as the carry has no values for them.
func (c Carry) restore(fields []types.FormField) {
	for i := range fields {
		field := &fields[i]

		if field.Type == types.FieldTypeGroup {
			c.restore(field.Fields)

			continue
		}

		if field.InputType == types.InputFieldTypeHidden {
			continue
		}

		submitted, ok := c.Values[field.Name]

		switch field.Type {
		case types.FieldTypeCheckbox:
			// An unticked box sends nothing.
			field.Value = ok && truthy(first(submitted))
		case types.FieldTypeMultiCheckbox:
			field.ValueMap = map[string]bool{}
			for _, value := range submitted {
				field.ValueMap[strings.TrimSpace(value)] = true
			}
		default:
			if ok {
				field.Value = first(submitted)
			}
		}
	}
}

// first is the first of values, or "".
func first(values []string) string {
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

// valueOrEmpty is value, or "" for nil.
func valueOrEmpty(value any) any {
	if value == nil {
		return ""
	}

	return value
}

// modelCarry is the carry of a RenderModel, or nil.
func modelCarry(v any) *Carry {
	if wrapped, ok := renderModelOf(v); ok {
		return wrapped.Carry
	}

	return nil
}

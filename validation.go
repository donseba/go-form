package form

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

type FieldValidationError struct {
	Field string
	Err   string
}

func (e FieldValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Err)
}

func (e FieldValidationError) FieldError() (field, err string) {
	return e.Field, e.Err
}

// ValidateForm validates struct fields based on struct tags.
// Returns a slice of FieldValidationError.
func ValidateForm(form any) FieldErrors {
	var errList FieldErrors
	v := reflect.ValueOf(form)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		name := field.Name

		// Required
		req := field.Tag.Get("required")
		if req == "true" {
			if isEmptyValue(value) {
				errList = append(errList, FieldValidationError{Field: name, Err: "is required"})
			}
		}

		// Min/Max/Step for numbers
		if value.Kind() == reflect.Int || value.Kind() == reflect.Float64 {
			var val float64
			if value.Kind() == reflect.Int {
				val = float64(value.Int())
			} else {
				val = value.Float()
			}
			if minTag := field.Tag.Get("min"); minTag != "" {
				min, _ := strconv.ParseFloat(minTag, 64)
				if val < min {
					errList = append(errList, FieldValidationError{Field: name, Err: fmt.Sprintf("must be >= %v", min)})
				}
			}
			if maxTag := field.Tag.Get("max"); maxTag != "" {
				max, _ := strconv.ParseFloat(maxTag, 64)
				if val > max {
					errList = append(errList, FieldValidationError{Field: name, Err: fmt.Sprintf("must be <= %v", max)})
				}
			}
		}

		// maxLength for strings and textareas
		if (value.Kind() == reflect.String) && field.Tag.Get("maxLength") != "" {
			maxLen, err := strconv.Atoi(field.Tag.Get("maxLength"))
			if err == nil && utf8.RuneCountInString(value.String()) > maxLen {
				errList = append(errList, FieldValidationError{Field: name, Err: fmt.Sprintf("must be at most %d characters", maxLen)})
			}
		}

		// minLength for strings and textareas
		if (value.Kind() == reflect.String) && field.Tag.Get("minLength") != "" {
			minLen, err := strconv.Atoi(field.Tag.Get("minLength"))
			if err == nil && utf8.RuneCountInString(value.String()) < minLen {
				errList = append(errList, FieldValidationError{Field: name, Err: fmt.Sprintf("must be at least %d characters", minLen)})
			}
		}

		// values check for radios/dropdowns
		if vals := field.Tag.Get("values"); vals != "" && value.Kind() == reflect.String {
			allowed := map[string]struct{}{}
			for _, v := range strings.Split(vals, ";") {
				if strings.Contains(v, ":") {
					parts := strings.SplitN(v, ":", 2)
					allowed[strings.TrimSpace(parts[0])] = struct{}{}
				} else {
					allowed[strings.TrimSpace(v)] = struct{}{}
				}
			}
			if value.String() != "" {
				if _, ok := allowed[value.String()]; !ok {
					errList = append(errList, FieldValidationError{Field: name, Err: "has an invalid value"})
				}
			}
		}

		// Email format (simple check)
		if field.Tag.Get("form") == "input,email" && value.Kind() == reflect.String {
			if val := value.String(); val != "" && !strings.Contains(val, "@") {
				errList = append(errList, FieldValidationError{Field: name, Err: "must be a valid email address"})
			}
		}

		// Enumerator validation
		if value.IsValid() && value.Type().Implements(reflect.TypeOf((*Enumerator)(nil)).Elem()) {
			enumVals := value.Interface().(Enumerator).Enum()
			valStr := fmt.Sprint(value.Interface())
			found := false
			for _, v := range enumVals {
				if fmt.Sprint(v) == valStr {
					found = true
					break
				}
			}
			if !found && valStr != "" {
				errList = append(errList, FieldValidationError{Field: name, Err: "has an invalid value (not in Enum)"})
			}
		}

		// Mapper validation
		if value.IsValid() && value.Type().Implements(reflect.TypeOf((*Mapper)(nil)).Elem()) {
			maps := value.Interface().(Mapper).Mapper()
			valStr := fmt.Sprint(value.Interface())
			if _, ok := maps[valStr]; !ok && valStr != "" {
				errList = append(errList, FieldValidationError{Field: name, Err: "has an invalid value (not in Mapper)"})
			}
		}

		// SortedMapper validation
		if value.IsValid() && value.Type().Implements(reflect.TypeOf((*SortedMapper)(nil)).Elem()) {
			smaps := value.Interface().(SortedMapper).SortedMapper()
			valStr := fmt.Sprint(value.Interface())
			found := false
			for _, sm := range smaps {
				if sm.Key() == valStr {
					found = true
					break
				}
			}
			if !found && valStr != "" {
				errList = append(errList, FieldValidationError{Field: name, Err: "has an invalid value (not in SortedMapper)"})
			}
		}
	}
	return errList
}

func (f *Form) ValidateForm(form any) FieldErrors {
	errList := ValidateForm(form) // built-in validations

	v := reflect.ValueOf(form)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		validateTag := field.Tag.Get("validate")
		if validateTag == "" {
			continue
		}
		for _, validatorName := range strings.Split(validateTag, ",") {
			validatorName = strings.TrimSpace(validatorName)
			if validatorName == "" {
				continue
			}
			if fn, ok := f.validators[validatorName]; ok {
				errList = append(errList, fn(value.Interface(), field)...)
			}
		}
	}
	return errList
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	}
	return false
}

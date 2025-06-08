package form

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	TranslationKeyRequired            = "form.validation.required"
	TranslationKeyMin                 = "form.validation.min"
	TranslationKeyMax                 = "form.validation.max"
	TranslationKeyMaxLength           = "form.validation.maxLength"
	TranslationKeyMinLength           = "form.validation.minLength"
	TranslationKeyInvalidValue        = "form.validation.invalidValue"
	TranslationKeyInvalidEmail        = "form.validation.invalidEmail"
	TranslationKeyInvalidEnum         = "form.validation.invalidEnum"
	TranslationKeyInvalidMapper       = "form.validation.invalidMapper"
	TranslationKeyInvalidSortedMapper = "form.validation.invalidSortedMapper"
	TranslationKeyPattern             = "form.validation.pattern"
	TranslationKeyURL                 = "form.validation.url"
	TranslationKeyBool                = "form.validation.bool"
	TranslationKeyZero                = "form.validation.zero"
	TranslationKeyMinItems            = "form.validation.minItems"
	TranslationKeyMaxItems            = "form.validation.maxItems"
	TranslationKeyPrefix              = "form.validation.prefix"
	TranslationKeySuffix              = "form.validation.suffix"
	TranslationKeyContains            = "form.validation.contains"
	TranslationKeyStep                = "form.validation.step"
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

// Helper functions for each validation type
func validateRequired(f *Form, field reflect.StructField, value reflect.Value, loc Localizer, getErr func(string, any) string) (errs FieldErrors) {
	req := field.Tag.Get("required")
	if req == "true" {
		if isEmptyValue(value) {
			errs = append(errs, FieldValidationError{Field: field.Name, Err: getErr(TranslationKeyRequired, nil)})
		}
	}
	return
}

func validateMinMax(f *Form, field reflect.StructField, value reflect.Value, loc Localizer, getErr func(string, any) string) (errs FieldErrors) {
	if value.Kind() == reflect.Int || value.Kind() == reflect.Float64 {
		var val float64
		if value.Kind() == reflect.Int {
			val = float64(value.Int())
		} else {
			val = value.Float()
		}
		if minTag := field.Tag.Get("min"); minTag != "" {
			minVal, _ := strconv.ParseFloat(minTag, 64)
			if val < minVal {
				errs = append(errs, FieldValidationError{Field: field.Name, Err: getErr(TranslationKeyMin, minVal)})
			}
		}
		if maxTag := field.Tag.Get("max"); maxTag != "" {
			maxVal, _ := strconv.ParseFloat(maxTag, 64)
			if val > maxVal {
				errs = append(errs, FieldValidationError{Field: field.Name, Err: getErr(TranslationKeyMax, maxVal)})
			}
		}
	}
	return
}

func validateStep(f *Form, field reflect.StructField, value reflect.Value, loc Localizer, getErr func(string, any) string) (errs FieldErrors) {
	if stepTag := field.Tag.Get("step"); stepTag != "" && (value.Kind() == reflect.Int || value.Kind() == reflect.Float64) {
		step, err := strconv.ParseFloat(stepTag, 64)
		var val float64
		if value.Kind() == reflect.Int {
			val = float64(value.Int())
		} else {
			val = value.Float()
		}
		if err == nil && step > 0 {
			mod := val / step
			if mod != float64(int64(mod)) {
				errs = append(errs, FieldValidationError{Field: field.Name, Err: getErr(TranslationKeyStep, step)})
			}
		}
	}
	return
}

func validateLength(f *Form, field reflect.StructField, value reflect.Value, loc Localizer, getErr func(string, any) string) (errs FieldErrors) {
	if value.Kind() == reflect.String {
		if maxLength := field.Tag.Get("maxLength"); maxLength != "" {
			maxLen, err := strconv.Atoi(maxLength)
			if err == nil && utf8.RuneCountInString(value.String()) > maxLen {
				errs = append(errs, FieldValidationError{Field: field.Name, Err: getErr(TranslationKeyMaxLength, maxLen)})
			}
		}
		if minLength := field.Tag.Get("minLength"); minLength != "" {
			minLen, err := strconv.Atoi(minLength)
			if err == nil && utf8.RuneCountInString(value.String()) < minLen {
				errs = append(errs, FieldValidationError{Field: field.Name, Err: getErr(TranslationKeyMinLength, minLen)})
			}
		}
	}
	return
}

func validatePattern(f *Form, field reflect.StructField, value reflect.Value, loc Localizer, getErr func(string, any) string) (errs FieldErrors) {
	if pattern := field.Tag.Get("pattern"); pattern != "" && value.Kind() == reflect.String {
		matched, err := regexp.MatchString(pattern, value.String())
		if err == nil && !matched {
			errs = append(errs, FieldValidationError{Field: field.Name, Err: getErr(TranslationKeyPattern, pattern)})
		}
	}
	return
}

func validateURL(f *Form, field reflect.StructField, value reflect.Value, loc Localizer, getErr func(string, any) string) (errs FieldErrors) {
	if field.Tag.Get("url") == "true" && value.Kind() == reflect.String {
		str := value.String()
		if str != "" {
			_, err := url.ParseRequestURI(str)
			if err != nil {
				errs = append(errs, FieldValidationError{Field: field.Name, Err: getErr(TranslationKeyURL, nil)})
			}
		}
	}
	return
}

func validateBool(f *Form, field reflect.StructField, value reflect.Value, loc Localizer, getErr func(string, any) string) (errs FieldErrors) {
	if field.Tag.Get("bool") == "true" && value.Kind() == reflect.Bool {
		if !value.Bool() {
			errs = append(errs, FieldValidationError{Field: field.Name, Err: getErr(TranslationKeyBool, nil)})
		}
	}
	return
}

func validateZero(f *Form, field reflect.StructField, value reflect.Value, loc Localizer, getErr func(string, any) string) (errs FieldErrors) {
	if field.Tag.Get("zero") == "true" {
		if !isEmptyValue(value) {
			errs = append(errs, FieldValidationError{Field: field.Name, Err: getErr(TranslationKeyZero, nil)})
		}
	}
	return
}

func validateSliceArrayLength(f *Form, field reflect.StructField, value reflect.Value, loc Localizer, getErr func(string, any) string) (errs FieldErrors) {
	if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
		if minItems := field.Tag.Get("minItems"); minItems != "" {
			min, err := strconv.Atoi(minItems)
			if err == nil && value.Len() < min {
				errs = append(errs, FieldValidationError{Field: field.Name, Err: getErr(TranslationKeyMinItems, min)})
			}
		}
		if maxItems := field.Tag.Get("maxItems"); maxItems != "" {
			max, err := strconv.Atoi(maxItems)
			if err == nil && value.Len() > max {
				errs = append(errs, FieldValidationError{Field: field.Name, Err: getErr(TranslationKeyMaxItems, max)})
			}
		}
	}
	return
}

func validatePrefixSuffixContains(f *Form, field reflect.StructField, value reflect.Value, loc Localizer, getErr func(string, any) string) (errs FieldErrors) {
	if value.Kind() == reflect.String {
		str := value.String()
		if prefix := field.Tag.Get("prefix"); prefix != "" {
			if !strings.HasPrefix(str, prefix) {
				errs = append(errs, FieldValidationError{Field: field.Name, Err: getErr(TranslationKeyPrefix, prefix)})
			}
		}
		if suffix := field.Tag.Get("suffix"); suffix != "" {
			if !strings.HasSuffix(str, suffix) {
				errs = append(errs, FieldValidationError{Field: field.Name, Err: getErr(TranslationKeySuffix, suffix)})
			}
		}
		if contains := field.Tag.Get("contains"); contains != "" {
			if !strings.Contains(str, contains) {
				errs = append(errs, FieldValidationError{Field: field.Name, Err: getErr(TranslationKeyContains, contains)})
			}
		}
	}
	return
}

func validateValues(f *Form, field reflect.StructField, value reflect.Value, loc Localizer, getErr func(string, any) string) (errs FieldErrors) {
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
				errs = append(errs, FieldValidationError{Field: field.Name, Err: getErr(TranslationKeyInvalidValue, value.String())})
			}
		}
	}
	return
}

func validateEmail(f *Form, field reflect.StructField, value reflect.Value, loc Localizer, getErr func(string, any) string) (errs FieldErrors) {
	if field.Tag.Get("form") == "input,email" && value.Kind() == reflect.String {
		if val := value.String(); val != "" && !strings.Contains(val, "@") {
			errs = append(errs, FieldValidationError{Field: field.Name, Err: getErr(TranslationKeyInvalidEmail, nil)})
		}
	}
	return
}

func validateEnum(f *Form, field reflect.StructField, value reflect.Value, loc Localizer, getErr func(string, any) string) (errs FieldErrors) {
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
			errs = append(errs, FieldValidationError{Field: field.Name, Err: getErr(TranslationKeyInvalidEnum, valStr)})
		}
	}
	return
}

func validateMapper(f *Form, field reflect.StructField, value reflect.Value, loc Localizer, getErr func(string, any) string) (errs FieldErrors) {
	if value.IsValid() && value.Type().Implements(reflect.TypeOf((*Mapper)(nil)).Elem()) {
		maps := value.Interface().(Mapper).Mapper()
		valStr := fmt.Sprint(value.Interface())
		if _, ok := maps[valStr]; !ok && valStr != "" {
			errs = append(errs, FieldValidationError{Field: field.Name, Err: getErr(TranslationKeyInvalidMapper, valStr)})
		}
	}
	return
}

func validateSortedMapper(f *Form, field reflect.StructField, value reflect.Value, loc Localizer, getErr func(string, any) string) (errs FieldErrors) {
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
			errs = append(errs, FieldValidationError{Field: field.Name, Err: getErr(TranslationKeyInvalidSortedMapper, valStr)})
		}
	}
	return
}

// internalFormValidation validates struct fields based on struct tags.
// Returns FieldErrors a slice or FieldError.
func (f *Form) internalFormValidation(form any, loc Localizer) FieldErrors {
	var errList FieldErrors
	v := reflect.ValueOf(form)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Custom error message
		errorMsg := field.Tag.Get("errorMsg")
		getErr := func(defaultKey string, arg any) string {
			if errorMsg != "" {
				return errorMsg
			}
			return f.validationErrorTranslated(loc, defaultKey, arg)
		}

		errList = append(errList,
			validateRequired(f, field, value, loc, getErr)...)
		errList = append(errList,
			validateMinMax(f, field, value, loc, getErr)...)
		errList = append(errList,
			validateStep(f, field, value, loc, getErr)...)
		errList = append(errList,
			validateLength(f, field, value, loc, getErr)...)
		errList = append(errList,
			validatePattern(f, field, value, loc, getErr)...)
		errList = append(errList,
			validateURL(f, field, value, loc, getErr)...)
		errList = append(errList,
			validateBool(f, field, value, loc, getErr)...)
		errList = append(errList,
			validateZero(f, field, value, loc, getErr)...)
		errList = append(errList,
			validateSliceArrayLength(f, field, value, loc, getErr)...)
		errList = append(errList,
			validatePrefixSuffixContains(f, field, value, loc, getErr)...)
		errList = append(errList,
			validateValues(f, field, value, loc, getErr)...)
		errList = append(errList,
			validateEmail(f, field, value, loc, getErr)...)
		errList = append(errList,
			validateEnum(f, field, value, loc, getErr)...)
		errList = append(errList,
			validateMapper(f, field, value, loc, getErr)...)
		errList = append(errList,
			validateSortedMapper(f, field, value, loc, getErr)...)
	}
	return errList
}

func (f *Form) validationErrorTranslated(loc Localizer, key string, args ...any) string {
	if f.translationEnabled && f.translationFunc != nil {
		if len(args) == 1 && args[0] == nil {
			return f.translationFunc(loc, key)
		}

		return f.translationFunc(loc, key, args...)
	}
	return fmt.Sprintf(key, args...)
}

func (f *Form) ValidateForm(form any) FieldErrors {
	return f.ValidateFormLocalized(form, &DefaultLocalizer{})
}

func (f *Form) ValidateFormLocalized(form any, loc Localizer) FieldErrors {
	errList := f.internalFormValidation(form, loc) // built-in validations

	v := reflect.ValueOf(form)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		// Handle nested structs (excluding time.Time)
		if value.Kind() == reflect.Struct && field.Type.PkgPath() != "time" {
			nestedErrs := f.ValidateFormLocalized(value.Addr().Interface(), loc)
			for _, err := range nestedErrs {
				f, e := err.FieldError()
				errList = append(errList, FieldValidationError{
					Field: field.Name + "." + f,
					Err:   e,
				})
			}
			continue
		}
		if value.Kind() == reflect.Ptr && !value.IsNil() && value.Elem().Kind() == reflect.Struct && field.Type.Elem().PkgPath() != "time" {
			nestedErrs := f.ValidateFormLocalized(value.Interface(), loc)
			for _, err := range nestedErrs {
				f, e := err.FieldError()

				errList = append(errList, FieldValidationError{
					Field: field.Name + "." + f,
					Err:   e,
				})
			}
			continue
		}
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
	default:
		// For unhandled kinds, treat as empty only if zero value
		z := reflect.Zero(v.Type())
		return reflect.DeepEqual(v.Interface(), z.Interface())
	}
}

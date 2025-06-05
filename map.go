package form

import (
	"net/http"
	"reflect"
	"strconv"
)

var (
	ErrMapFormNotPointer = &MapFormError{"dst must be a pointer to struct"}
	ErrMapFormNotStruct  = &MapFormError{"dst must be a pointer to struct"}
)

// MapForm maps form values from an http.Request to a struct based on the `name` tag.
// Only exported fields are set. Supports string, int, float64, and bool fields.
func MapForm(r *http.Request, dst any, prefixes ...string) error {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return ErrMapFormNotPointer
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return ErrMapFormNotStruct
	}
	prefix := ""
	if len(prefixes) > 0 {
		prefix = prefixes[0]
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fv := v.Field(i)
		if !fv.CanSet() {
			continue
		}

		// Recursively map nested structs (skip special types)
		if fv.Kind() == reflect.Struct {
			if fv.CanAddr() {
				name := field.Tag.Get("name")
				if name == "" {
					name = field.Name
				}

				_ = MapForm(r, fv.Addr().Interface(), prefix+name+".")
			}
			continue
		}
		formKey := field.Tag.Get("name")
		if formKey == "" {
			formKey = field.Name
		}
		formValue := r.FormValue(prefix + formKey)
		if formValue == "" {
			continue
		}

		switch fv.Kind() {
		case reflect.String:
			fv.SetString(formValue)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if iv, err := strconv.ParseInt(formValue, 10, 64); err == nil {
				fv.SetInt(iv)
			}
		case reflect.Float32, reflect.Float64:
			if fv64, err := strconv.ParseFloat(formValue, 64); err == nil {
				fv.SetFloat(fv64)
			}
		case reflect.Bool:
			if bv, err := strconv.ParseBool(formValue); err == nil {
				fv.SetBool(bv)
			}
		}
	}
	return nil
}

type MapFormError struct {
	msg string
}

func (e *MapFormError) Error() string {
	return e.msg
}

package form

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/donseba/go-form/types"
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

		// Recursively map nested structs (skip time.Time and Info)
		if fv.Kind() == reflect.Struct &&
			!(field.Type.PkgPath() == "time" && field.Type.Name() == "Time") &&
			field.Type != reflect.TypeOf(Info{}) &&
			fv.CanAddr() {
			name := field.Tag.Get("name")
			if name == "" {
				name = field.Name
			}
			_ = MapForm(r, fv.Addr().Interface(), prefix+name+".")
			continue
		}
		formKey := field.Tag.Get("name")
		if formKey == "" {
			formKey = field.Name
		}

		// Handle boolean fields specially
		if fv.Kind() == reflect.Bool {
			// Check if value exists in form - useful for checkboxes
			if r != nil && r.Form != nil {
				formValue := r.FormValue(prefix + formKey)

				// Check if form value exists
				if _, exists := r.Form[prefix+formKey]; exists {
					// For checkboxes, "on" means true
					if formValue == "on" || formValue == "true" || formValue == "1" {
						fv.SetBool(true)
					} else if formValue == "off" || formValue == "false" || formValue == "0" || formValue == "" {
						fv.SetBool(false)
					} else if bv, err := strconv.ParseBool(formValue); err == nil {
						fv.SetBool(bv)
					}
				} else {
					// If checkbox isn't in the form at all, it's unchecked
					fv.SetBool(false)
				}
			}
			continue
		}

		// For non-boolean fields, proceed as before
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
		case reflect.Array:
			// Handle UUID arrays ([16]byte typically)
			if fv.Type().Elem().Kind() == reflect.Uint8 && fv.Len() == 16 {
				// Try to parse UUID string (format: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx")
				if len(formValue) == 36 {
					hexString := strings.ReplaceAll(formValue, "-", "")
					if len(hexString) == 32 {
						for i := 0; i < fv.Len() && i < len(hexString)/2; i++ {
							// Convert each pair of hex chars to byte
							if b, err := strconv.ParseUint(hexString[i*2:i*2+2], 16, 8); err == nil {
								fv.Index(i).SetUint(b)
							}
						}
					}
				}
			} else {
				fmt.Printf("unsupported array field type %s for field %s\n", fv.Type().Elem().Kind(), field.Name)
			}
		case reflect.Struct:
			if field.Type.PkgPath() == "time" && field.Type.Name() == "Time" {
				err := parseTimeToFieldValue(fv, field, formValue)
				if err != nil {
					fmt.Printf("error parsing time field %s: %s", field.Name, err)
					continue
				}
			}
		case reflect.Ptr:
			if field.Type.Elem().PkgPath() == "time" && field.Type.Elem().Name() == "Time" {
				err := parseTimeToFieldValue(fv, field, formValue)
				if err != nil {
					fmt.Printf("error parsing time field %s: %s", field.Name, err)
					continue
				}
			}
		default:
			fmt.Printf("unsupported field type %s for field %s\n", fv.Kind(), field.Name)
		}
	}
	return nil
}

func parseTimeToFieldValue(fv reflect.Value, field reflect.StructField, formValue string) error {
	fieldTag := field.Tag.Get("form")
	timetype := "datetime-local"
	if fieldTag != "" {
		parts := strings.Split(fieldTag, ",")
		if len(parts) == 2 && parts[0] == "input" {
			// if the tag is set, we use it to determine the type
			timetype = parts[1]
		}

		var (
			layout  = "2006-01-02T15:04"
			layout2 = ""
		)
		switch timetype {
		case types.InputFieldTypeDateTimeLocal.String():
			layout = time.DateTime
			layout2 = "2006-01-02T15:04"
		case types.InputFieldTypeDate.String():
			layout = time.DateOnly
		case types.InputFieldTypeTime.String():
			layout = time.TimeOnly
		case types.InputFieldTypeMonth.String():
			layout = "01"
		case types.InputFieldTypeWeek.String():
			layout = "2006-W01"
			tt, err := WeekStringToTime(formValue)
			if err != nil {
				fmt.Println("Error parsing week string:", err)
				return err // or optionally return err
			}

			layout = "2006-01-02"
			formValue = tt.Format(layout)
		}

		parsed, err := time.Parse(layout, formValue)
		if err != nil {
			// If the layout is datetime-local, we try to parse it with a different layout
			if layout2 != "" {
				parsed, err = time.Parse(layout2, formValue)
				if err != nil {
					return err // or optionally return err
				}
			} else {
				return err
			}
		}

		fv.Set(reflect.ValueOf(parsed))
	}
	return nil
}

func WeekStringToTime(weekStr string) (time.Time, error) {
	// Example input: "2025-W23"
	parts := strings.Split(weekStr, "-W")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid week format")
	}
	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, err
	}
	week, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, err
	}
	// Start from Jan 4th, which is always in week 1
	jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, time.UTC)

	// Find the Monday of week 1
	for jan4.Weekday() != time.Monday {
		jan4 = jan4.AddDate(0, 0, -1)
	}
	// Add weeks
	t := jan4.AddDate(0, 0, (week-1)*7)
	return t, nil
}

type MapFormError struct {
	msg string
}

func (e *MapFormError) Error() string {
	return e.msg
}

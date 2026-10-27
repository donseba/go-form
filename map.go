package form

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/donseba/go-form/v2/types"
)

var (
	ErrMapFormNotPointer = &MapFormError{"dst must be a pointer to struct"}
	ErrMapFormNotStruct  = &MapFormError{"dst must be a pointer to struct"}
)

// MapFormErrorHandler, when set, receives the submitted values MapForm could
// not set, by form field name: a key a SortedSelect or SortedMultiSelect
// does not offer, an option a struct radio group does not have, a date in
// the wrong format, a number that does not parse, does not fit the field or
// is not finite (NaN, Inf), or a field type MapForm does not support.
// MapForm skips those values and never writes output itself; the field keeps
// its zero or previous value. MapOptions.Strict returns them as field
// errors instead of only skipping them. Set it,
// for example, to log such values while debugging:
//
//	form.MapFormErrorHandler = func(field string, err error) {
//		slog.Debug("form value ignored", "field", field, "error", err)
//	}
var MapFormErrorHandler func(field string, err error)

// mapFormError hands a value MapForm skipped to MapFormErrorHandler.
func mapFormError(field string, err error) {
	if MapFormErrorHandler != nil {
		MapFormErrorHandler(field, err)
	}
}

// MapOptions changes how MapFormWith maps a submission. The zero value maps
// like MapForm.
type MapOptions struct {
	// Prefix goes before every form key, like the prefix argument of
	// MapForm.
	Prefix string
	// Location is the time zone a datetime-local value is read in: the
	// zone the visitor entered it in, usually the site's. Nil reads it as
	// UTC. A time the clocks skip or show twice is read as
	// ParseDateTimeLocal reads it. Render the form with a TimeLocalizer of
	// the same zone. Date, time, month and week inputs hold no instant and
	// are always read as UTC.
	Location *time.Location
	// TrimSpace removes white space around submitted strings, except in
	// password inputs, where spaces count, and fields with trim:"false".
	// The tag trim:"true" trims one field without this option.
	TrimSpace bool
	// Strict reports the submitted values that cannot be read as their
	// field's type: a number such as "abc", NaN or one too large for the
	// field, a date or time in the wrong format, a key a select or
	// multi-select does not offer, an option a struct radio group does not
	// have, a checkbox value other than on/off, true/false or 1/0, or a
	// malformed UUID. MapFormWith still maps every other value and returns
	// them as an *InvalidValuesError; Form.MapFormLocalized returns them as
	// translated field errors. Without Strict such values are skipped, so
	// the field keeps its zero or previous value.
	Strict bool
	// NormalizeNewlines reads the line breaks of submitted strings as
	// "\n": browsers send a textarea's line breaks as "\r\n", but count
	// them as one character for maxlength, as maxLength does after
	// normalizing. Password inputs are taken as typed.
	NormalizeNewlines bool
}

// InvalidValuesError is the error MapFormWith returns with
// MapOptions.Strict when submitted values cannot be read as their field's
// type. Every other value is mapped.
type InvalidValuesError struct {
	// Errors has an error per field, named as the field renders (such as
	// "address.city"), with the rule "invalid" and go-form's English
	// message (TranslationKeyInvalid), or the field's invalidMsg or
	// errorMsg tag. Form.MapFormLocalized translates them.
	Errors FieldErrors

	values []invalidValue
}

// invalidValue is a submitted value that could not be mapped.
type invalidValue struct {
	name  string
	field reflect.StructField
}

// Error lists the fields with invalid values.
func (e *InvalidValuesError) Error() string {
	names := make([]string, 0, len(e.values))
	for _, value := range e.values {
		names = append(names, value.name)
	}

	return "form: invalid values for " + strings.Join(names, ", ")
}

// fieldErrors are the errors with messages for loc.
func (e *InvalidValuesError) fieldErrors(f *Form, loc Localizer) FieldErrors {
	errs := make(FieldErrors, 0, len(e.values))
	for _, value := range e.values {
		message := f.messageFunc(value.field, loc)("invalid", TranslationKeyInvalid, nil)
		errs = append(errs, FieldValidationError{Field: value.name, Rule: "invalid", Err: message})
	}

	return errs
}

// MapFormLocalized maps form values from r to dst like MapFormWith with
// options.Strict set, and returns the values it could not read as field
// errors with messages for loc, ready to add to those of validation:
//
//	errs, err := renderer.MapFormLocalized(r, &input, form.MapOptions{TrimSpace: true}, loc)
//	if err != nil {
//		http.Error(w, "Bad request", http.StatusBadRequest)
//		return
//	}
//
//	errs = append(errs, renderer.ValidateFormLocalized(&input, loc)...)
//
// err is only set when dst is not a pointer to a struct.
func (f *Form) MapFormLocalized(r *http.Request, dst any, options MapOptions, loc Localizer) (FieldErrors, error) {
	options.Strict = true

	err := MapFormWith(r, dst, options)

	var invalid *InvalidValuesError
	if errors.As(err, &invalid) {
		return invalid.fieldErrors(f, loc), nil
	}

	return nil, err
}

// Bind maps a submission from r onto dst, a pointer to a struct with
// go-form's tags, like MapFormLocalized (strictly), and validates it like
// ValidateFormLocalized, with messages for loc (nil for DefaultLocalizer).
// It returns the field errors of both; a field whose submitted value could
// not be read has that error only, rather than what validation says about
// the value it was left with:
//
//	var input SignupInput
//
//	errs, err := renderer.Bind(r, &input, form.MapOptions{TrimSpace: true, NormalizeNewlines: true}, loc)
//	if err != nil {
//		http.Error(w, "Bad request", http.StatusBadRequest)
//		return
//	}
//
//	if len(errs) > 0 {
//		// Show the form again with errs.
//	}
//
// The overrides of an Overrider dst apply to validation. dst may also be a
// RenderModel whose Model is a pointer to a struct, such as the model the
// form rendered with: the submission is mapped onto its Model, and its
// Overrides and Options apply to validation. err is only set when dst is
// not a pointer to a struct; the errors are then nil.
func (f *Form) Bind(r *http.Request, dst any, options MapOptions, loc Localizer) (FieldErrors, error) {
	loc = orDefaultLocalizer(loc)

	target := dst
	if wrapped, ok := renderModelOf(dst); ok {
		target = wrapped.Model
	}

	errs, err := f.MapFormLocalized(r, target, options, loc)
	if err != nil {
		return nil, err
	}

	unreadable := make(map[string]bool, len(errs))

	for _, fieldErr := range errs {
		field, _ := fieldErr.FieldError()
		unreadable[field] = true
	}

	for _, fieldErr := range f.ValidateFormLocalized(dst, loc) {
		if field, _ := fieldErr.FieldError(); !unreadable[field] {
			errs = append(errs, fieldErr)
		}
	}

	return errs, nil
}

// defaultMaxMemory is the memory a multipart form may use before its files
// go to disk, as for http.Request.FormValue. Call r.ParseMultipartForm before
// mapping to use another limit, and limit the request body with
// http.MaxBytesReader.
const defaultMaxMemory = 32 << 20

// MapForm maps form values from an http.Request to a struct based on the `name` tag.
// Only exported fields are set. Supports string, int, uint, float and bool
// fields, pointers to strings and numbers (nil when the value is empty),
// slices of strings and integers (a value per ticked box of a
// multicheckbox), time.Time, selects and uploaded files
// (*multipart.FileHeader and []*multipart.FileHeader). MapFormWith maps
// with options such as a time zone.
func MapForm(r *http.Request, dst any, prefixes ...string) error {
	options := MapOptions{}
	if len(prefixes) > 0 {
		options.Prefix = prefixes[0]
	}

	return MapFormWith(r, dst, options)
}

// MapFormWith maps form values from r to dst like MapForm, with options:
//
//	err := form.MapFormWith(r, &input, form.MapOptions{Location: site.Location, TrimSpace: true})
func MapFormWith(r *http.Request, dst any, options MapOptions) error {
	m := &mapper{r: r, options: options}
	if err := m.mapStruct(dst, options.Prefix); err != nil {
		return err
	}

	if !options.Strict || len(m.invalid) == 0 {
		return nil
	}

	// A Form without translations gives the default messages. NewForm would
	// start a CSRF store, with a goroutine that never stops, per call.
	invalid := &InvalidValuesError{values: m.invalid}
	invalid.Errors = invalid.fieldErrors(&Form{}, &DefaultLocalizer{})

	return invalid
}

// mapper maps one submission, collecting the values it cannot read.
type mapper struct {
	r       *http.Request
	options MapOptions
	invalid []invalidValue
}

// keySetter is a select, such as a SortedSelect, whose option a submitted
// key selects.
type keySetter interface {
	SetFromKey(key string) error
}

// keysSetter is a multi-select, such as a SortedMultiSelect, whose options
// the submitted keys select.
type keysSetter interface {
	SetFromKeys(keys []string) error
}

// skip records a submitted value of field, named name, that could not be
// read, and hands it to MapFormErrorHandler.
func (m *mapper) skip(name string, field reflect.StructField, err error) {
	mapFormError(name, err)

	m.invalid = append(m.invalid, invalidValue{name: name, field: field})
}

// mapStruct maps the values below prefix onto dst, a pointer to a struct.
func (m *mapper) mapStruct(dst any, prefix string) error {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return ErrMapFormNotPointer
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return ErrMapFormNotStruct
	}

	// Parse the submission up front, so that checkboxes and files are found
	// whatever the order of the fields. Without a multipart body this
	// parses the form as usual.
	if m.r != nil && m.r.Form == nil {
		_ = m.r.ParseMultipartForm(defaultMaxMemory)
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		m.mapField(t.Field(i), v.Field(i), prefix)
	}

	return nil
}

// mapField maps the value submitted for field, a field below prefix, onto
// fv, its value: a struct radio group, a nested struct field by field, a
// checkbox, uploaded files, a select, the ticked boxes of a multicheckbox,
// or one value.
func (m *mapper) mapField(field reflect.StructField, fv reflect.Value, prefix string) {
	if !fv.CanSet() {
		return
	}

	name := prefix + validationFieldName(field)

	if mapsNested(field, fv) {
		if key, ok := radioGroupKey(field, fv, name); ok {
			m.mapRadioGroup(field, fv, name, key)

			return
		}

		// An error of a nested form does not stop the others
		// from being mapped; its invalid values are recorded.
		_ = m.mapStruct(fv.Addr().Interface(), name+".")

		return
	}

	if fv.Kind() == reflect.Bool {
		m.mapBool(field, fv, name)

		return
	}

	if single, ok := fileField(field.Type); ok {
		m.mapFiles(fv, name, single)

		return
	}

	formValue := m.cleaned(field, m.r.FormValue(name))
	_, submitted := m.r.Form[name]

	// A select, such as a SortedSelect, reads the submitted key. A
	// submitted empty value selects the empty choice, so a pre-filled
	// value can be cleared.
	if setter, ok := addressOf[keySetter](fv); ok {
		if submitted {
			m.selectKey(setter, field, name, formValue)
		}

		return
	}

	if setter, ok := addressOf[keysSetter](fv); ok {
		m.selectKeys(setter, field, name)

		return
	}

	// The ticked boxes of a multicheckbox, or the choices of a select
	// with multiple, into a slice of strings or integers.
	if listField(field.Type) {
		m.mapList(field, fv, name)

		return
	}

	if formValue == "" {
		// A submitted empty value clears a pre-filled string, time or
		// pointer.
		if submitted && clearable(field.Type) {
			fv.Set(reflect.Zero(fv.Type()))
		}

		return
	}

	m.setValue(field, fv, name, formValue)
}

// mapsNested reports whether fv, the value of field, is a nested form
// whose fields are mapped one by one: a struct other than time.Time, Info
// or a select, such as a SortedSelect, which takes one value.
func mapsNested(field reflect.StructField, fv reflect.Value) bool {
	if fv.Kind() != reflect.Struct || isTime(field.Type) || field.Type == reflect.TypeOf(Info{}) || !fv.CanAddr() {
		return false
	}

	_, single := addressOf[keySetter](fv)
	_, multiple := addressOf[keysSetter](fv)

	return !single && !multiple
}

// addressOf returns the address of fv as a T, such as the setter of a
// select, and whether it is one.
func addressOf[T any](fv reflect.Value) (T, bool) {
	if !fv.CanAddr() {
		var none T

		return none, false
	}

	setter, ok := fv.Addr().Interface().(T)

	return setter, ok
}

// mapBool sets fv, a bool, from the value submitted as name: "on", "true"
// and "1" tick it, and "off", "false", "0", "" and a checkbox that is not in
// the form, as browsers leave out an unticked one, clear it. Other values
// count as strconv.ParseBool reads them, or are recorded as invalid.
// Without a parsed form the field keeps its value.
func (m *mapper) mapBool(field reflect.StructField, fv reflect.Value, name string) {
	if m.r == nil || m.r.Form == nil {
		return
	}

	if _, exists := m.r.Form[name]; !exists {
		fv.SetBool(false)

		return
	}

	formValue := m.r.FormValue(name)

	switch formValue {
	case "on", "true", "1":
		fv.SetBool(true)
	case "off", "false", "0", "":
		fv.SetBool(false)
	default:
		checked, err := strconv.ParseBool(formValue)
		if err != nil {
			m.skip(name, field, err)

			return
		}

		fv.SetBool(checked)
	}
}

// mapFiles sets fv, a *multipart.FileHeader when single is set and a
// []*multipart.FileHeader otherwise, from the files uploaded as name.
// Without a file the field keeps its value.
func (m *mapper) mapFiles(fv reflect.Value, name string, single bool) {
	// r.ParseForm leaves a multipart body unread.
	if m.r.MultipartForm == nil {
		_ = m.r.ParseMultipartForm(defaultMaxMemory)
	}

	var files []*multipart.FileHeader
	if m.r.MultipartForm != nil {
		files = m.r.MultipartForm.File[name]
	}

	switch {
	case len(files) == 0:
	case single:
		fv.Set(reflect.ValueOf(files[0]))
	default:
		fv.Set(reflect.ValueOf(files))
	}
}

// cleaned is value, the value submitted for field, with its line breaks as
// "\n" and without white space around it, as the options and the tags of
// field ask, when field is a string or a pointer to one.
func (m *mapper) cleaned(field reflect.StructField, value string) string {
	kind := field.Type.Kind()
	if kind != reflect.String && (kind != reflect.Pointer || field.Type.Elem().Kind() != reflect.String) {
		return value
	}

	if m.options.NormalizeNewlines && !isPassword(field) {
		value = normalizeNewlines(value)
	}

	if trims(field, m.options) {
		value = strings.TrimSpace(value)
	}

	return value
}

// selectKey selects the option key of a select, the value of field. An
// unknown key leaves the selection empty and is recorded as invalid;
// mapping does not abort the whole form.
func (m *mapper) selectKey(setter keySetter, field reflect.StructField, name, key string) {
	if err := setter.SetFromKey(key); err != nil && key != "" {
		m.skip(name, field, err)
	}
}

// selectKeys selects the options of a multi-select, the value of field,
// with the keys submitted as name (see listValues), without empty ones. As
// for a select, a key it does not offer leaves nothing selected. Such a key,
// and keys for a multi-select without a source, are recorded as invalid;
// mapping goes on with the other fields.
func (m *mapper) selectKeys(setter keysSetter, field reflect.StructField, name string) {
	submitted, ok := m.listValues(name)
	if !ok {
		return
	}

	keys := make([]string, 0, len(submitted))

	for _, key := range submitted {
		if key != "" {
			keys = append(keys, key)
		}
	}

	if err := setter.SetFromKeys(keys); err != nil && len(keys) > 0 {
		m.skip(name, field, err)
	}
}

// radioGroupKey reports whether fv, the value of field, named name, is a
// struct radio group: a struct tagged form:"radios" or
// form:"radios,radio_group" whose exported fields are bools with the name
// of the group, its options. key is the name its radio buttons are
// submitted as: the name of the group for form:"radios", and the name of
// its options below it for form:"radios,radio_group".
func radioGroupKey(field reflect.StructField, fv reflect.Value, name string) (key string, ok bool) {
	group := validationFieldName(field)

	switch field.Tag.Get(tagForm) {
	case string(types.FieldTypeRadios):
		key = name
	case string(types.FieldTypeRadios) + "," + string(types.InputFieldTypeRadioGroup):
		key = name + "." + group
	default:
		return "", false
	}

	t := fv.Type()
	for i := 0; i < t.NumField(); i++ {
		option := t.Field(i)
		if option.IsExported() && (option.Type.Kind() != reflect.Bool || validationFieldName(option) != group) {
			return "", false
		}
	}

	return key, true
}

// mapRadioGroup ticks the option of fv, a struct radio group named name
// (see radioGroupKey), that was submitted as key, and clears the others.
// An option's value is the name of the group, a dot and the Go name of its
// field, as the group renders it, such as "Plan.Monthly". Nothing
// submitted clears every option, as for an unticked checkbox; a value that
// is no option too, and it is recorded as invalid. Without a parsed form
// the options keep their values.
func (m *mapper) mapRadioGroup(field reflect.StructField, fv reflect.Value, name, key string) {
	if m.r == nil || m.r.Form == nil {
		return
	}

	chosen := m.r.FormValue(key)
	found := false

	t := fv.Type()
	for i := 0; i < t.NumField(); i++ {
		if !t.Field(i).IsExported() {
			continue
		}

		selected := chosen == name+"."+t.Field(i).Name
		if selected {
			found = true
		}

		fv.Field(i).SetBool(selected)
	}

	if chosen != "" && !found {
		m.skip(key, field, fmt.Errorf("option %q not found in radio group", chosen))
	}
}

// listValues are the values submitted as name for a field with a value per
// ticked box or chosen option. Browsers send nothing when no box is
// ticked, so on a POST no values mean an empty selection, like an
// unchecked checkbox; otherwise ok is false and the field keeps its value.
func (m *mapper) listValues(name string) (values []string, ok bool) {
	if m.r.Method == http.MethodPost && m.r.PostForm != nil {
		values = m.r.PostForm[name]
	} else if m.r.Form != nil {
		values = m.r.Form[name]
	}

	return values, len(values) > 0 || m.r.Method == http.MethodPost
}

// clearable reports whether a submitted empty value clears a field of type
// t: a string, a time or a pointer to a string or number.
func clearable(t reflect.Type) bool {
	return t.Kind() == reflect.String || isTime(t) || scalarPointer(t)
}

// scalarPointer reports whether t is a pointer to a string or number, which
// a submitted value sets to a new value.
func scalarPointer(t reflect.Type) bool {
	if t.Kind() != reflect.Pointer {
		return false
	}

	kind := t.Elem().Kind()

	return kind != reflect.Bool && scalarKind(kind)
}

// setValue sets fv, the value of field, from value, a submitted value that
// is not empty: a string or number, a pointer to one, a UUID ([16]byte) or
// a time. A value that cannot be read is recorded as invalid; a field type
// MapForm does not support is only handed to MapFormErrorHandler.
func (m *mapper) setValue(field reflect.StructField, fv reflect.Value, name, value string) {
	// A pointer to a string or number gets a new value.
	if scalarPointer(field.Type) {
		target := reflect.New(field.Type.Elem())
		if err := setText(target.Elem(), value); err != nil {
			m.skip(name, field, err)

			return
		}

		fv.Set(target)

		return
	}

	switch fv.Kind() {
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		if err := setText(fv, value); err != nil {
			m.skip(name, field, err)
		}
	case reflect.Array:
		m.setArray(field, fv, name, value)
	case reflect.Struct, reflect.Pointer:
		// Other structs, such as Info, and pointers are not read.
		if !isTime(field.Type) {
			return
		}

		if err := parseTimeToFieldValue(fv, field, value, m.options.Location); err != nil {
			m.skip(name, field, err)
		}
	default:
		mapFormError(name, fmt.Errorf("unsupported field type %s", fv.Kind()))
	}
}

// setArray sets fv, the value of field, from a UUID such as
// "6ba7b810-9dad-11d1-80b4-00c04fd430c8" when it is a [16]byte; a value
// that is no UUID is recorded as invalid. Other arrays are not supported.
func (m *mapper) setArray(field reflect.StructField, fv reflect.Value, name, value string) {
	if fv.Type().Elem().Kind() != reflect.Uint8 || fv.Len() != 16 {
		mapFormError(name, fmt.Errorf("unsupported array field type %s", fv.Type().Elem().Kind()))

		return
	}

	if !setUUID(fv, value) {
		m.skip(name, field, fmt.Errorf("invalid UUID %q", value))
	}
}

// listField reports whether t is a slice of strings or integers, which
// holds a value per ticked box of a multicheckbox. A []byte is not one.
func listField(t reflect.Type) bool {
	if t.Kind() != reflect.Slice {
		return false
	}

	switch t.Elem().Kind() {
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	}

	return false
}

// mapList sets fv, a slice of strings or integers, from the values
// submitted as name (see listValues), without empty ones; no values on a
// POST make it empty (nil). When a value cannot be read, the field keeps
// its value and the value is recorded as invalid.
func (m *mapper) mapList(field reflect.StructField, fv reflect.Value, name string) {
	submitted, ok := m.listValues(name)
	if !ok {
		return
	}

	list := reflect.Zero(fv.Type())

	for _, value := range submitted {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		item := reflect.New(fv.Type().Elem()).Elem()
		if err := setText(item, value); err != nil {
			m.skip(name, field, err)

			return
		}

		list = reflect.Append(list, item)
	}

	fv.Set(list)
}

// setText sets fv, a string or number, from a submitted value. A number
// that does not parse, does not fit fv or is not finite is an error, and
// leaves fv as it was.
func setText(fv reflect.Value, value string) error {
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		number, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}

		if fv.OverflowInt(number) {
			return fmt.Errorf("value %q does not fit %s", value, fv.Type())
		}

		fv.SetInt(number)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		number, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}

		if fv.OverflowUint(number) {
			return fmt.Errorf("value %q does not fit %s", value, fv.Type())
		}

		fv.SetUint(number)
	case reflect.Float32, reflect.Float64:
		number, err := strconv.ParseFloat(value, fv.Type().Bits())
		if err != nil {
			return err
		}

		// NaN and Inf parse, but compare false with every limit.
		if !finite(number) {
			return fmt.Errorf("value %q is not a finite number", value)
		}

		fv.SetFloat(number)
	default:
		return fmt.Errorf("unsupported field type %s", fv.Kind())
	}

	return nil
}

// setUUID sets fv, a [16]byte, from a UUID such as
// "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx", and reports whether s is one.
// fv is left as it is otherwise.
func setUUID(fv reflect.Value, s string) bool {
	hexString := strings.ReplaceAll(s, "-", "")
	if len(s) != 36 || len(hexString) != 32 {
		return false
	}

	var parsed [16]byte
	for i := range parsed {
		b, err := strconv.ParseUint(hexString[i*2:i*2+2], 16, 8)
		if err != nil {
			return false
		}

		parsed[i] = byte(b)
	}

	for i := range parsed {
		fv.Index(i).SetUint(uint64(parsed[i]))
	}

	return true
}

// newlines turns "\r\n" and a lone "\r" into "\n".
var newlines = strings.NewReplacer("\r\n", "\n", "\r", "\n")

// normalizeNewlines returns s with its line breaks as "\n".
func normalizeNewlines(s string) string {
	if !strings.Contains(s, "\r") {
		return s
	}

	return newlines.Replace(s)
}

// trims reports whether MapFormWith trims the value of field: with
// trim:"true", or with the TrimSpace option unless the field is a password
// input or has trim:"false".
func trims(field reflect.StructField, options MapOptions) bool {
	switch field.Tag.Get("trim") {
	case "true":
		return true
	case "false":
		return false
	}

	return options.TrimSpace && !isPassword(field)
}

// timeLayouts are the layouts a value of each input type may come in.
// Browsers leave out the seconds of a time or date-time unless the step asks
// for them; fractions of a second are always accepted.
var timeLayouts = map[string][]string{
	types.InputFieldTypeDateTimeLocal.String(): {dateTimeLocalLayout, dateTimeLocalSecondsLayout, time.DateTime},
	types.InputFieldTypeDate.String():          {time.DateOnly},
	types.InputFieldTypeTime.String():          {"15:04", time.TimeOnly},
	types.InputFieldTypeMonth.String():         {monthLayout},
}

// parseTimeToFieldValue sets fv, a time.Time or *time.Time, from a
// submitted value. A datetime-local value is read in location when it is
// not nil (see ParseDateTimeLocal); other values hold no instant and are
// read as UTC.
func parseTimeToFieldValue(fv reflect.Value, field reflect.StructField, formValue string, location *time.Location) error {
	fieldTag := field.Tag.Get("form")
	if fieldTag == "" {
		return nil
	}

	timetype := types.InputFieldTypeDateTimeLocal.String()

	parts := strings.Split(fieldTag, ",")
	if len(parts) == 2 && parts[0] == "input" {
		// if the tag is set, we use it to determine the type
		timetype = parts[1]
	}

	var (
		parsed time.Time
		err    error
	)

	switch timetype {
	case types.InputFieldTypeWeek.String():
		parsed, err = WeekStringToTime(formValue)
	case types.InputFieldTypeDateTimeLocal.String():
		// A wall-clock field holds the time as entered, as UTC.
		if field.Tag.Get(tagWallClock) == "true" {
			location = nil
		}

		parsed, err = ParseDateTimeLocal(formValue, location)
	default:
		layouts, ok := timeLayouts[timetype]
		if !ok {
			layouts = timeLayouts[types.InputFieldTypeDateTimeLocal.String()]
		}

		for _, layout := range layouts {
			parsed, err = time.Parse(layout, formValue)
			if err == nil {
				break
			}
		}
	}

	if err != nil {
		return err
	}

	if fv.Kind() == reflect.Pointer {
		fv.Set(reflect.ValueOf(&parsed))
	} else {
		fv.Set(reflect.ValueOf(parsed))
	}

	return nil
}

// ParseDateTimeLocal reads value, the value of a datetime-local input such
// as "2026-03-29T14:30", as a wall-clock time in location, as MapFormWith
// reads it with MapOptions.Location; a nil location reads it as UTC.
// Seconds and fractions of a second are accepted.
//
// Where location changes its offset, a wall-clock time can be missing or
// occur twice, and time.ParseInLocation leaves open which instant it
// returns. ParseDateTimeLocal moves a time the clocks skip (02:30 when they
// go from 02:00 to 03:00) forward by the length of the gap, to 03:30, and
// reads a time that occurs twice (02:30 when they go back from 03:00 to
// 02:00) as the first of the two, before the clocks go back.
func ParseDateTimeLocal(value string, location *time.Location) (time.Time, error) {
	var (
		parsed time.Time
		err    error
	)

	for _, layout := range timeLayouts[types.InputFieldTypeDateTimeLocal.String()] {
		parsed, err = time.Parse(layout, value)
		if err == nil {
			return WallTime(parsed, location), nil
		}
	}

	return time.Time{}, err
}

// WallTime is the instant at which the clocks in location show the date and
// time of day of wall; the location of wall itself is ignored. A time the
// clocks skip or show twice is read as ParseDateTimeLocal reads it. A nil
// location is UTC.
func WallTime(wall time.Time, location *time.Location) time.Time {
	if location == nil {
		location = time.UTC
	}

	year, month, day := wall.Date()
	wall = time.Date(year, month, day, wall.Hour(), wall.Minute(), wall.Second(), wall.Nanosecond(), time.UTC)

	// The offsets a day before and after the wall-clock time, read as UTC,
	// are those around the one change of offset it can be near.
	_, before := wall.Add(-24 * time.Hour).In(location).Zone()
	_, after := wall.Add(24 * time.Hour).In(location).Zone()

	// Before a change, and the first of two times when the clocks go back.
	early := wall.Add(-time.Duration(before) * time.Second).In(location)
	if _, offset := early.Zone(); offset == before {
		return early
	}

	// After a change.
	late := wall.Add(-time.Duration(after) * time.Second).In(location)
	if _, offset := late.Zone(); offset == after {
		return late
	}

	// In a gap the clocks skip, the offset before the gap moves the time
	// forward by its length.
	return early
}

// isTime reports whether t is time.Time or *time.Time.
func isTime(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t.PkgPath() == "time" && t.Name() == "Time"
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

package form

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"mime"
	"mime/multipart"
	"net/mail"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// translationContext starts the translation keys of go-form's own texts.
const translationContext = "form||"

var (
	TranslationKeyRequired                      = "form||Validation required"
	TranslationKeyMin                           = "form||Value should be greater than or equal to %v"
	TranslationKeyMax                           = "form||Value should be less than or equal to %v"
	TranslationKeyMaxLength                     = "form||Value should not exceed %d characters"
	TranslationKeyMinLength                     = "form||Value should be at least %d characters"
	TranslationKeyInvalidValue                  = "form||Invalid value '%s' provided"
	TranslationKeyInvalidEmail                  = "form||Invalid email format"
	TranslationKeyInvalidEnum                   = "form||Invalid enum value '%s' provided"
	TranslationKeyInvalidMapper                 = "form||Invalid mapper value '%s' provided"
	TranslationKeyInvalidSortedMapper           = "form||Invalid sorted mapper value '%s' provided"
	TranslationKeyPattern                       = "form||Value is not in the expected format"
	TranslationKeyURL                           = "form||Invalid URL format"
	TranslationKeyBool                          = "form||Value should be either true or false"
	TranslationKeyZero                          = "form||Value should be zero"
	TranslationKeyMinItems                      = "form||Value should have at least %d items"
	TranslationKeyMaxItems                      = "form||Value should have at most %d items"
	TranslationKeyPrefix                        = "form||Value should start with '%s'"
	TranslationKeySuffix                        = "form||Value should end with '%s'"
	TranslationKeyContains                      = "form||Value should contain '%s'"
	TranslationKeyStep                          = "form||Value should be a multiple of %g"
	TranslationKeyCSRFTokenMissing              = "form||CSRF token is missing"
	TranslationKeyCSRFTokenInvalid              = "form||Invalid CSRF token"
	TranslationKeyCSRFTokenError                = "form||Error processing CSRF token"
	TranslationKeySortedSelectNotFound          = "form||SortedSelect: value '%v' not found in source"
	TranslationKeySortedSelectKeyNotFound       = "form||SortedSelect: key '%s' not found in source"
	TranslationKeySortedSelectSourceNil         = "form||SortedSelect: source map is nil"
	TranslationKeySortedSelectTypeError         = "form||Cannot assign or convert value of type %T to %s"
	TranslationKeySortedSelectUnmarshalNotFound = "form||SortedSelect: value '%v' not found in source during unmarshal"
	TranslationKeyEqual                         = "form||Value should match %s"
	TranslationKeyAccept                        = "form||File should be of type %s"
	TranslationKeyMaxSize                       = "form||File should not exceed %s"
	// TranslationKeyInvalid is the message of a value that cannot be read
	// as the field's type, such as "abc" for a number or a date in the
	// wrong format (MapOptions.Strict), a number that is not finite (NaN,
	// Inf) or text that is not valid UTF-8. Its rule is "invalid", so the
	// tag invalidMsg replaces it.
	TranslationKeyInvalid = "form||Value is not valid"
	// TranslationKeyLine is the message of line:"true": the value must be
	// one line without control characters such as tabs.
	TranslationKeyLine = "form||Value should be a single line without tabs or line breaks"
	// TranslationKeyStepFromMin is the message of a value between two steps
	// that are counted from a min that is not itself a multiple of the step,
	// such as 0.75 for step:"0.5" min:"0.25"; its verbs are the step and
	// the min.
	TranslationKeyStepFromMin = "form||Value should be in steps of %g from %g"
)

// TranslationKeyRequiredLabel is the text after the label of a required field
// that only screen readers announce; the asterisk is hidden from them.
var TranslationKeyRequiredLabel = "form||(required)"

// FieldValidationError represents a validation error for a specific field.
type FieldValidationError struct {
	Field string
	Err   string
	// Rule is the rule that failed, such as "required", "maxLength",
	// "pattern", "values" or "invalid", or the name of a custom validator
	// (the validate tag) when it leaves Rule empty. The message per rule
	// tag is the rule followed by Msg (requiredMsg, …).
	Rule string
}

// Error implements the error interface for FieldValidationError.
func (e FieldValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Err)
}

// FieldError returns the field and error message.
func (e FieldValidationError) FieldError() (field, err string) {
	return e.Field, e.Err
}

// messageFunc returns the message of a failed rule, such as "required" or
// "pattern": the field's message for the rule (requiredMsg, patternMsg, …),
// its errorMsg, or the translation of key with arg.
// A message of go-form gets all args; a field's own message for the rule
// only the first, the rule's argument, such as the step.
type messageFunc func(rule, key string, args ...any) string

// Helper functions for each validation type
func validateRequired(field reflect.StructField, value reflect.Value, getErr messageFunc) (errs FieldErrors) {
	req := field.Tag.Get("required")
	if req == "true" {
		// A nil pointer is nothing filled in; any other, such as one to
		// 0, is filled in.
		missing := isEmptyValue(value)
		text, _ := pointee(value)
		// Only spaces count as nothing filled in: browsers accept them for
		// required fields, so the server must not. A password is taken as
		// typed: spaces may be part of it.
		if text.Kind() == reflect.String && !isPassword(field) && strings.TrimSpace(text.String()) == "" {
			missing = true
		}

		if selected, ok := selectionMade(value); ok {
			missing = !selected
		}
		if missing {
			errs = append(errs, ruleError(field, getErr, "required", TranslationKeyRequired, nil))
		}
	}
	return
}

// selectionMade reports whether a select field (a SortedSelect,
// SortedMultiSelect or another Mapper or SortedMapper) has a choice selected.
// A single select counts as empty when its key is "" or not one of its
// options, so a placeholder such as "" → "—" does not satisfy required:"true".
// A multi-select needs at least one key that is not "". ok is false for
// fields that are not selects.
func selectionMade(value reflect.Value) (selected bool, ok bool) {
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return false, false
		}
		value = value.Elem()
	}
	if !value.IsValid() || !value.CanInterface() {
		return false, false
	}

	switch v := value.Interface().(type) {
	case MultiSelectGetter:
		for _, key := range v.GetKeysAsStrings() {
			if key != "" {
				return true, true
			}
		}
		return false, true
	case SortedMapper:
		key := v.String()
		if key == "" {
			return false, true
		}
		for _, option := range v.SortedMapper() {
			if option.Key() == key {
				return true, true
			}
		}
		return false, true
	case Mapper:
		key := v.String()
		if key == "" {
			return false, true
		}
		_, found := v.Mapper()[key]
		return found, true
	}

	return false, false
}

// ruleError is the error of field for the rule that failed, with the
// message getErr gives the rule: the field's own message for it, or key,
// go-form's message, with args.
func ruleError(field reflect.StructField, getErr messageFunc, rule, key string, args ...any) FieldValidationError {
	return FieldValidationError{Field: validationFieldName(field), Rule: rule, Err: getErr(rule, key, args...)}
}

// numberValue returns the value of an int or float field as a float64. ok is
// false for other kinds.
func numberValue(value reflect.Value) (number float64, ok bool) {
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(value.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(value.Uint()), true
	case reflect.Float32, reflect.Float64:
		return value.Float(), true
	}

	return 0, false
}

// finite reports whether number is neither NaN nor infinite.
func finite(number float64) bool {
	return !math.IsNaN(number) && !math.IsInf(number, 0)
}

// validateFinite refuses NaN and infinite numbers, which compare false with
// every limit, so that min and max would let them through.
func validateFinite(
	f *Form,
	field reflect.StructField,
	value reflect.Value,
	loc Localizer,
	getErr messageFunc,
) (errs FieldErrors) {
	if number, ok := numberValue(value); ok && !finite(number) {
		errs = append(errs, ruleError(field, getErr, "invalid", TranslationKeyInvalid, nil))
	}

	return
}

func validateMinMax(
	f *Form,
	field reflect.StructField,
	value reflect.Value,
	loc Localizer,
	getErr messageFunc,
) (errs FieldErrors) {
	val, ok := numberValue(value)
	if !ok || !finite(val) {
		return nil
	}

	if minTag := field.Tag.Get("min"); minTag != "" {
		minVal, _ := strconv.ParseFloat(minTag, 64)
		if val < minVal {
			errs = append(errs, ruleError(field, getErr, "min", TranslationKeyMin, minVal))
		}
	}

	if maxTag := field.Tag.Get("max"); maxTag != "" {
		maxVal, _ := strconv.ParseFloat(maxTag, 64)
		if val > maxVal {
			errs = append(errs, ruleError(field, getErr, "max", TranslationKeyMax, maxVal))
		}
	}

	return errs
}

// validateStep checks the step tag: a number must be a whole number of steps
// from min, or from zero without a min, as a browser counts them. It counts
// exactly, so 0.3 is on step:"0.1". A step of "any", zero or less, or no
// number at all lets every number through.
func validateStep(f *Form, field reflect.StructField, value reflect.Value, loc Localizer, getErr messageFunc) (errs FieldErrors) {
	stepTag := field.Tag.Get("step")
	if stepTag == "" {
		return nil
	}

	number, ok := exactValue(value)
	if !ok {
		return nil
	}

	step, ok := exactStep(stepTag)
	if !ok {
		return nil
	}

	base := new(big.Rat)
	if minimum, ok := exactString(field.Tag.Get("min")); ok {
		base = minimum
	}

	if stepsFrom(number, base, step) {
		return nil
	}

	stepNumber, _ := step.Float64()
	message := getErr("step", TranslationKeyStep, stepNumber)

	// A min that is a whole number of steps from zero has the steps of zero,
	// which the shorter message names.
	if !stepsFrom(base, new(big.Rat), step) {
		minimum, _ := base.Float64()
		message = getErr("step", TranslationKeyStepFromMin, stepNumber, minimum)
	}

	return FieldErrors{FieldValidationError{Field: validationFieldName(field), Rule: "step", Err: message}}
}

// OnStep reports whether value is a whole number of steps of step from min,
// or from zero when min is empty, as the step and min tags check it: exactly,
// so "0.3" is on a step of "0.1", and "0.75" is on a step of "0.5" from
// "0.25". A value that is no number is on no step; a step of "any", zero or
// less, or no number at all has every number on it. A min that is no number
// counts as zero.
func OnStep(value, min, step string) bool {
	number, ok := exactString(value)
	if !ok {
		return false
	}

	steps, ok := exactStep(step)
	if !ok {
		return true
	}

	base, ok := exactString(min)
	if !ok {
		base = new(big.Rat)
	}

	return stepsFrom(number, base, steps)
}

// stepsFrom reports whether number is a whole number of steps from base.
func stepsFrom(number, base, step *big.Rat) bool {
	steps := new(big.Rat).Sub(number, base)
	steps.Quo(steps, step)

	return steps.IsInt()
}

// exactStep reads a step tag as an exact number above zero. ok is false for
// "any", zero or less, or no number, which check no steps.
func exactStep(tag string) (step *big.Rat, ok bool) {
	step, ok = exactString(tag)
	if !ok || step.Sign() <= 0 {
		return nil, false
	}

	return step, true
}

// exactString reads text as a finite number, exactly as it reads as a
// float64 (a browser's number): "0.1" is one tenth. ok is false for text
// that is no number, NaN or infinite.
func exactString(text string) (number *big.Rat, ok bool) {
	parsed, err := strconv.ParseFloat(text, 64)
	if err != nil || !finite(parsed) {
		return nil, false
	}

	return exactFloat(parsed, 64), true
}

// exactFloat is number as the shortest decimal that reads back as it in
// bitSize bits, exactly: 0.1 is one tenth, not the binary number nearest
// to it, and a float32 0.3 is three tenths.
func exactFloat(number float64, bitSize int) *big.Rat {
	exact, ok := new(big.Rat).SetString(strconv.FormatFloat(number, 'g', -1, bitSize))
	if !ok {
		return new(big.Rat).SetFloat64(number)
	}

	return exact
}

// exactValue returns the value of an int, uint or float field as an exact
// number. ok is false for other kinds, NaN and infinity.
func exactValue(value reflect.Value) (number *big.Rat, ok bool) {
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return new(big.Rat).SetInt64(value.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return new(big.Rat).SetInt(new(big.Int).SetUint64(value.Uint())), true
	case reflect.Float32, reflect.Float64:
		if !finite(value.Float()) {
			return nil, false
		}

		bitSize := 64
		if value.Kind() == reflect.Float32 {
			bitSize = 32
		}

		return exactFloat(value.Float(), bitSize), true
	}

	return nil, false
}

func validateLength(
	f *Form,
	field reflect.StructField,
	value reflect.Value,
	loc Localizer,
	getErr messageFunc,
) (errs FieldErrors) {
	if value.Kind() == reflect.String {
		if maxLength := field.Tag.Get("maxLength"); maxLength != "" {
			maxLen, err := strconv.Atoi(maxLength)
			if err == nil && utf8.RuneCountInString(value.String()) > maxLen {
				errs = append(errs, ruleError(field, getErr, "maxLength", TranslationKeyMaxLength, maxLen))
			}
		}
		if minLength := field.Tag.Get("minLength"); minLength != "" {
			minLen, err := strconv.Atoi(minLength)
			if err == nil && utf8.RuneCountInString(minLengthText(field, value.String())) < minLen {
				errs = append(errs, ruleError(field, getErr, "minLength", TranslationKeyMinLength, minLen))
			}
		}
	}
	return
}

// minLengthText is the part of text, the value of field, that counts
// towards its minimum length: leading and trailing spaces do not count,
// except in a password, which is taken as typed.
func minLengthText(field reflect.StructField, text string) string {
	if isPassword(field) {
		return text
	}

	return strings.TrimSpace(text)
}

func validatePattern(
	f *Form,
	field reflect.StructField,
	value reflect.Value,
	loc Localizer,
	getErr messageFunc,
) (errs FieldErrors) {
	if pattern := field.Tag.Get("pattern"); pattern != "" && value.Kind() == reflect.String {
		matched, err := regexp.MatchString(pattern, value.String())
		if err == nil && !matched {
			// The pattern itself means nothing to people; it is only passed
			// to a message that asks for it.
			var arg any
			if strings.Contains(TranslationKeyPattern, "%") {
				arg = pattern
			}

			errs = append(errs, ruleError(field, getErr, "pattern", TranslationKeyPattern, arg))
		}
	}
	return
}

// validateText refuses text that is not valid UTF-8, which no browser sends,
// and checks line:"true": one line, without control characters such as
// tabs, line breaks or NUL, and without the Unicode line and paragraph
// separators, for values such as a name or a subject that end up in e-mail
// headers and lists.
func validateText(f *Form, field reflect.StructField, value reflect.Value, loc Localizer, getErr messageFunc) (errs FieldErrors) {
	if value.Kind() != reflect.String {
		return
	}

	text := value.String()
	if !utf8.ValidString(text) {
		errs = append(errs, ruleError(field, getErr, "invalid", TranslationKeyInvalid, nil))

		return
	}

	if field.Tag.Get("line") == "true" && !SingleLine(text) {
		errs = append(errs, ruleError(field, getErr, "line", TranslationKeyLine, nil))
	}

	return
}

// SingleLine reports whether s is one line of valid UTF-8 without control
// characters (tabs, line breaks, NUL, …) and without the Unicode line and
// paragraph separators: the rule of line:"true".
func SingleLine(s string) bool {
	if !utf8.ValidString(s) {
		return false
	}

	return strings.IndexFunc(s, func(r rune) bool {
		return unicode.IsControl(r) || r == '\u2028' || r == '\u2029'
	}) < 0
}

// validateURL checks the url tag: url:"true" for a URL with a scheme or an
// absolute path, url:"http,https" for an absolute URL with one of the listed
// schemes, url:"link" for a link on a page: an http(s) URL, a path, an
// #anchor, or a mailto: or tel: address. Scripts (javascript:, vbscript:,
// data:) are never valid.
func validateURL(f *Form, field reflect.StructField, value reflect.Value, loc Localizer, getErr messageFunc) (errs FieldErrors) {
	rule := field.Tag.Get("url")
	if rule == "" || rule == "false" || value.Kind() != reflect.String {
		return
	}

	if str := value.String(); str != "" && !ValidURL(rule, str) {
		errs = append(errs, ruleError(field, getErr, "url", TranslationKeyURL, nil))
	}

	return
}

// ValidURL reports whether s is a URL the url tag rule allows, the check
// of url:"…" for code that validates URLs outside a form:
//
//   - "true" (or ""): a URL with a scheme, or an absolute path;
//   - "link": a link on a page: an http(s) URL, a path (/news), an anchor
//     (#contact), or a mailto: or tel: address;
//   - a list of schemes such as "http,https": an absolute URL with one of
//     them, and a host for http and https.
//
// URLs with white space or control characters and javascript:, vbscript:
// and data: URLs are never valid.
func ValidURL(rule, s string) bool {
	if rule == "" {
		rule = "true"
	}

	if strings.IndexFunc(s, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) >= 0 {
		return false
	}

	switch rule {
	case "true":
		u, err := url.ParseRequestURI(s)
		return err == nil && !scriptScheme(u.Scheme)
	case "link":
		// //host is another site without a scheme; a path or anchor stays
		// on this one.
		if strings.HasPrefix(s, "/") || strings.HasPrefix(s, "#") {
			_, err := url.Parse(s)
			return err == nil && !strings.HasPrefix(s, "//")
		}

		return schemeURL(s, []string{"http", "https", "mailto", "tel"})
	default:
		return schemeURL(s, strings.Split(rule, ","))
	}
}

// schemeURL reports whether s is an absolute URL with one of schemes: with a
// host, or with an opaque part for schemes such as mailto: and tel:.
func schemeURL(s string, schemes []string) bool {
	u, err := url.Parse(s)
	if err != nil || scriptScheme(u.Scheme) {
		return false
	}

	allowed := false

	for _, scheme := range schemes {
		if strings.EqualFold(strings.TrimSpace(scheme), u.Scheme) {
			allowed = true
			break
		}
	}

	if !allowed {
		return false
	}

	if u.Host != "" {
		return true
	}

	scheme := strings.ToLower(u.Scheme)

	return u.Opaque != "" && scheme != "http" && scheme != "https"
}

// scriptScheme reports whether a URL scheme runs code or embeds content
// when followed.
func scriptScheme(scheme string) bool {
	switch strings.ToLower(scheme) {
	case "javascript", "vbscript", "data":
		return true
	}

	return false
}

func validateBool(f *Form, field reflect.StructField, value reflect.Value, loc Localizer, getErr messageFunc) (errs FieldErrors) {
	if field.Tag.Get("bool") == "true" && value.Kind() == reflect.Bool {
		if !value.Bool() {
			errs = append(errs, ruleError(field, getErr, "bool", TranslationKeyBool, nil))
		}
	}
	return
}

func validateZero(f *Form, field reflect.StructField, value reflect.Value, loc Localizer, getErr messageFunc) (errs FieldErrors) {
	if field.Tag.Get("zero") == "true" {
		if !isEmptyValue(value) {
			errs = append(errs, ruleError(field, getErr, "zero", TranslationKeyZero, nil))
		}
	}
	return
}

// validateSliceArrayLength checks the minItems and maxItems tags against the
// number of items of value (see itemCount).
func validateSliceArrayLength(
	f *Form,
	field reflect.StructField,
	value reflect.Value,
	loc Localizer,
	getErr messageFunc,
) (errs FieldErrors) {
	count, ok := itemCount(value)
	if !ok {
		return nil
	}

	if minItems := field.Tag.Get("minItems"); minItems != "" {
		minCount, err := strconv.Atoi(minItems)
		if err == nil && count < minCount {
			errs = append(errs, ruleError(field, getErr, "minItems", TranslationKeyMinItems, minCount))
		}
	}

	if maxItems := field.Tag.Get("maxItems"); maxItems != "" {
		maxCount, err := strconv.Atoi(maxItems)
		if err == nil && count > maxCount {
			errs = append(errs, ruleError(field, getErr, "maxItems", TranslationKeyMaxItems, maxCount))
		}
	}

	return errs
}

// itemCount is the number of items of value that minItems and maxItems
// count, and whether they count them: the length of a slice or array, or
// the number of selected options of a multi-select, such as a
// SortedMultiSelect (MultiSelectGetter).
func itemCount(value reflect.Value) (int, bool) {
	if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
		return value.Len(), true
	}

	if !value.IsValid() || !value.CanInterface() {
		return 0, false
	}

	if multi, ok := value.Interface().(MultiSelectGetter); ok {
		return len(multi.GetKeysAsStrings()), true
	}

	return 0, false
}

func validatePrefixSuffixContains(
	f *Form,
	field reflect.StructField,
	value reflect.Value,
	loc Localizer,
	getErr messageFunc,
) (errs FieldErrors) {
	if value.Kind() == reflect.String {
		str := value.String()
		if prefix := field.Tag.Get("prefix"); prefix != "" {
			if !strings.HasPrefix(str, prefix) {
				errs = append(errs, ruleError(field, getErr, "prefix", TranslationKeyPrefix, prefix))
			}
		}
		if suffix := field.Tag.Get("suffix"); suffix != "" {
			if !strings.HasSuffix(str, suffix) {
				errs = append(errs, ruleError(field, getErr, "suffix", TranslationKeySuffix, suffix))
			}
		}
		if contains := field.Tag.Get("contains"); contains != "" {
			if !strings.Contains(str, contains) {
				errs = append(errs, ruleError(field, getErr, "contains", TranslationKeyContains, contains))
			}
		}
	}
	return
}

// validateValues checks the values tag: a string or integer field, or a
// slice of them, must hold options of the tag. An empty string, and 0 when
// it is not an option, mean nothing is chosen.
func validateValues(
	f *Form,
	field reflect.StructField,
	value reflect.Value,
	loc Localizer,
	getErr messageFunc,
) (errs FieldErrors) {
	tag := field.Tag.Get(tagValues)
	if tag == "" {
		return nil
	}

	return choiceErrors(field, value, offered(parseValuesTag(tag, false)), getErr)
}

// validateEmail checks email inputs (form:"input,email") and fields with
// email:"true": the value must be a plain address such as
// name@example.com, without a display name, of at most 254 characters. With
// multiple:"true" it may be a comma-separated list of addresses.
func validateEmail(
	f *Form,
	field reflect.StructField,
	value reflect.Value,
	loc Localizer,
	getErr messageFunc,
) (errs FieldErrors) {
	isEmail := strings.TrimSpace(field.Tag.Get("form")) == "input,email" || field.Tag.Get("email") == "true"
	if !isEmail || value.Kind() != reflect.String || value.String() == "" {
		return
	}

	addresses := []string{value.String()}
	if field.Tag.Get(tagMultiple) == "true" {
		addresses = strings.Split(value.String(), ",")
	}

	for _, address := range addresses {
		if !ValidEmail(strings.TrimSpace(address)) {
			errs = append(errs, ruleError(field, getErr, "email", TranslationKeyInvalidEmail, nil))
			break
		}
	}

	return
}

// ValidEmail reports whether s is a plain email address, such as
// name@example.com: what net/mail parses, without a display name or angle
// brackets, of at most 254 characters.
func ValidEmail(s string) bool {
	if s == "" || len(s) > 254 {
		return false
	}

	address, err := mail.ParseAddress(s)

	return err == nil && address.Name == "" && address.Address == s
}

// validateEqual checks equal:"Other": the value of field must equal the
// field Other of parent, the same struct, named by its Go name or name tag,
// such as a repeated password. overrides are those of parent's fields.
func (f *Form) validateEqual(
	parent reflect.Value,
	field reflect.StructField,
	value reflect.Value,
	loc Localizer,
	overrides Overrides,
) FieldErrors {
	other := field.Tag.Get("equal")
	if other == "" {
		return nil
	}

	otherField, ok := siblingField(parent.Type(), other)
	if !ok {
		return nil
	}

	otherField = overrides.apply(validationFieldName(otherField), otherField)

	otherValue := parent.FieldByIndex(otherField.Index)
	if !value.CanInterface() || !otherValue.CanInterface() || reflect.DeepEqual(value.Interface(), otherValue.Interface()) {
		return nil
	}

	label := otherField.Tag.Get(tagLabel)
	if label == "" {
		label = validationFieldName(otherField)
	}

	message := f.messageFunc(field, loc)("equal", TranslationKeyEqual, f.translate(loc, label, nil))

	return FieldErrors{FieldValidationError{Field: validationFieldName(field), Rule: "equal", Err: message}}
}

// siblingField finds the field of t with the Go name or name tag name.
func siblingField(t reflect.Type, name string) (reflect.StructField, bool) {
	if field, ok := t.FieldByName(name); ok {
		return field, true
	}

	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Tag.Get(tagName) == name {
			return t.Field(i), true
		}
	}

	return reflect.StructField{}, false
}

// validateFiles checks uploaded files: accept:"image/*,.pdf" for their type
// and maxSize:"10MB" for the size of each file.
func validateFiles(
	f *Form,
	field reflect.StructField,
	value reflect.Value,
	loc Localizer,
	getErr messageFunc,
) (errs FieldErrors) {
	files := uploadedFiles(value)
	if len(files) == 0 {
		return
	}

	if accept := field.Tag.Get(tagAccept); accept != "" {
		for _, file := range files {
			if !AcceptsFile(accept, file) {
				errs = append(errs, ruleError(field, getErr, "accept", TranslationKeyAccept, accept))
				break
			}
		}
	}

	if tag := field.Tag.Get("maxSize"); tag != "" {
		limit, err := ParseSize(tag)
		if err != nil {
			return
		}

		for _, file := range files {
			if file.Size > limit {
				errs = append(errs, ruleError(field, getErr, "maxSize", TranslationKeyMaxSize, tag))
				break
			}
		}
	}

	return
}

// uploadedFiles returns the files of a *multipart.FileHeader or
// []*multipart.FileHeader value.
func uploadedFiles(value reflect.Value) []*multipart.FileHeader {
	if !value.IsValid() || !value.CanInterface() {
		return nil
	}

	switch files := value.Interface().(type) {
	case *multipart.FileHeader:
		if files != nil {
			return []*multipart.FileHeader{files}
		}
	case []*multipart.FileHeader:
		return files
	}

	return nil
}

// AcceptsFile reports whether file matches accept, a list like the accept
// attribute: extensions such as ".pdf" and media types such as "image/png" or
// "image/*". Both the name and the media type are given by the browser, so
// this guides people; check the content itself where it matters.
func AcceptsFile(accept string, file *multipart.FileHeader) bool {
	mediaType, _, _ := mime.ParseMediaType(file.Header.Get("Content-Type"))
	extension := strings.ToLower(filepath.Ext(file.Filename))

	for _, token := range strings.Split(accept, ",") {
		token = strings.ToLower(strings.TrimSpace(token))

		switch {
		case token == "":
		case strings.HasPrefix(token, "."):
			if extension == token {
				return true
			}
		case strings.HasSuffix(token, "/*"):
			if strings.HasPrefix(mediaType, strings.TrimSuffix(token, "*")) {
				return true
			}
		case mediaType == token:
			return true
		}
	}

	return false
}

// ParseSize reads a file size such as "512KB", "10MB", "1.5GB" or "100" (bytes).
// Units are binary: 1KB is 1024 bytes; KiB, MiB and GiB are accepted too.
func ParseSize(size string) (int64, error) {
	text := strings.ToUpper(strings.TrimSpace(size))
	multiplier := 1.0

	for _, unit := range []struct {
		suffix string
		bytes  float64
	}{
		{"KIB", 1 << 10},
		{"MIB", 1 << 20},
		{"GIB", 1 << 30},
		{"KB", 1 << 10},
		{"MB", 1 << 20},
		{"GB", 1 << 30},
		{"K", 1 << 10},
		{"M", 1 << 20},
		{"G", 1 << 30},
		{"B", 1},
	} {
		if strings.HasSuffix(text, unit.suffix) {
			text = strings.TrimSpace(strings.TrimSuffix(text, unit.suffix))
			multiplier = unit.bytes
			break
		}
	}

	number, err := strconv.ParseFloat(text, 64)
	if err != nil || number < 0 {
		return 0, fmt.Errorf("form: invalid size %q", size)
	}

	return int64(number * multiplier), nil
}

func validateEnum(f *Form, field reflect.StructField, value reflect.Value, loc Localizer, getErr messageFunc) (errs FieldErrors) {
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
			errs = append(errs, ruleError(field, getErr, "values", TranslationKeyInvalidEnum, valStr))
		}
	}
	return
}

func validateMapper(
	f *Form,
	field reflect.StructField,
	value reflect.Value,
	loc Localizer,
	getErr messageFunc,
) (errs FieldErrors) {
	if value.IsValid() && value.Type().Implements(reflect.TypeOf((*Mapper)(nil)).Elem()) {
		maps := value.Interface().(Mapper).Mapper()
		valStr := fmt.Sprint(value.Interface())
		if _, ok := maps[valStr]; !ok && valStr != "" {
			errs = append(errs, ruleError(field, getErr, "values", TranslationKeyInvalidMapper, valStr))
		}
	}
	return
}

// validateSortedMapper checks a SortedMapper field, such as a SortedSelect
// or SortedMultiSelect: the keys it holds must be keys of its options, and
// an error the select holds is reported with its translated message.
func validateSortedMapper(
	f *Form,
	field reflect.StructField,
	value reflect.Value,
	loc Localizer,
	getErr messageFunc,
) (errs FieldErrors) {
	if !value.IsValid() || !value.Type().Implements(sortedMapperType) {
		return nil
	}

	options := value.Interface().(SortedMapper).SortedMapper()

	if multi, ok := value.Interface().(MultiSelectGetter); ok {
		return unknownKeys(field, multi.GetKeysAsStrings(), options, getErr)
	}

	// The zero value, such as 0 for SortedSelect[int], means nothing is
	// selected; required:"true" reports that, not this check.
	key := fmt.Sprint(value.Interface())
	if !hasKey(options, key) && key != "" && key != fmt.Sprint(reflect.Zero(value.Type()).Interface()) {
		errs = append(errs, ruleError(field, getErr, "values", TranslationKeyInvalidSortedMapper, key))
	}

	return append(errs, selectError(f, field, value, loc)...)
}

// unknownKeys are the errors of the keys of a multi-select that are not
// keys of its options, one per key. An empty key is no choice.
func unknownKeys(field reflect.StructField, keys []string, options []SortedMap, getErr messageFunc) (errs FieldErrors) {
	for _, key := range keys {
		if key != "" && !hasKey(options, key) {
			errs = append(errs, ruleError(field, getErr, "values", TranslationKeyInvalidSortedMapper, key))
		}
	}

	return errs
}

// hasKey reports whether key is the key of one of options.
func hasKey(options []SortedMap, key string) bool {
	return slices.ContainsFunc(options, func(option SortedMap) bool {
		return option.Key() == key
	})
}

// selectError is the error a select, the value of field, holds as a
// SortedSelectError, with its message translated for loc.
func selectError(f *Form, field reflect.StructField, value reflect.Value, loc Localizer) FieldErrors {
	if !value.CanAddr() {
		return nil
	}

	var selectErr SortedSelectError

	err, ok := value.Addr().Interface().(error)
	if !ok || err == nil || !errors.As(err, &selectErr) {
		return nil
	}

	return FieldErrors{FieldValidationError{
		Field: validationFieldName(field),
		Rule:  "values",
		Err:   f.validationErrorTranslated(loc, selectErr.Key, selectErr.Args...),
	}}
}

// internalFormValidation validates the fields of form, a struct or a
// pointer to one, based on their struct tags, without nested structs and
// custom validators.
func (f *Form) internalFormValidation(form any, loc Localizer) FieldErrors {
	v := reflect.ValueOf(form)

	overrides := overridesOf(v)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	return f.validateTags(v, loc, overrides, nil)
}

// validateTags validates the fields of v, a struct, based on their struct
// tags with overrides applied, and the options of a render (options) in
// place of their values tags.
func (f *Form) validateTags(v reflect.Value, loc Localizer, overrides Overrides, options FieldOptions) FieldErrors {
	var errList FieldErrors
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		name := validationFieldName(t.Field(i))
		field := overrides.apply(name, t.Field(i))
		value := v.Field(i)

		choices, given := options[name]
		if given {
			field.Tag = overrideTag(field.Tag, Tags{tagValues: ""})
		}

		errList = append(errList, f.validateField(field, value, loc)...)
		errList = append(errList, f.validateEqual(v, field, value, loc, overrides)...)

		if given {
			chosen, _ := pointee(value)
			errList = append(errList, choiceErrors(field, chosen, offered(choices), f.messageFunc(field, loc))...)
		}
	}

	return errList
}

// fieldValidators are the rules of struct tags that check one field.
// validateRequired runs before them, with the field's own value; they get
// the value a pointer to a string, bool or number points to.
var fieldValidators = []func(*Form, reflect.StructField, reflect.Value, Localizer, messageFunc) FieldErrors{
	validateFinite,
	validateText,
	validateMinMax,
	validateStep,
	validateLength,
	validatePattern,
	validateURL,
	validateBool,
	validateZero,
	validateSliceArrayLength,
	validatePrefixSuffixContains,
	validateValues,
	validateEmail,
	validateEnum,
	validateMapper,
	validateSortedMapper,
	validateFiles,
}

// validateField checks value against the tags of field.
func (f *Form) validateField(field reflect.StructField, value reflect.Value, loc Localizer) (errs FieldErrors) {
	getErr := f.messageFunc(field, loc)
	errs = append(errs, validateRequired(field, value, getErr)...)

	// A nil pointer has nothing to check beyond required.
	checked, isNil := pointee(value)
	if isNil {
		return errs
	}

	for _, validate := range fieldValidators {
		errs = append(errs, validate(f, field, checked, loc, getErr)...)
	}

	return errs
}

// messageFunc returns the messages of field: per rule (requiredMsg,
// patternMsg, …), errorMsg for every rule of the field, or go-form's own.
func (f *Form) messageFunc(field reflect.StructField, loc Localizer) messageFunc {
	errorMsg := field.Tag.Get("errorMsg")

	return func(rule, defaultKey string, args ...any) string {
		if message := field.Tag.Get(rule + "Msg"); message != "" {
			var arg any
			if len(args) > 0 {
				arg = args[0]
			}

			return f.ruleMessage(loc, message, arg)
		}

		if errorMsg != "" {
			return errorMsg
		}

		return f.validationErrorTranslated(loc, defaultKey, args...)
	}
}

// ruleMessage translates a message set for one rule of a field, such as
// patternMsg:"Use lowercase letters and dashes". Like a label it is text of
// the application. The rule's argument, such as the minimum length, is only
// passed when the message has a verb for it.
func (f *Form) ruleMessage(loc Localizer, message string, arg any) string {
	if arg == nil || !strings.Contains(message, "%") {
		return f.translate(loc, message, nil)
	}

	return f.translate(loc, message, []any{arg})
}

// validationErrorTranslated translates one of go-form's messages. Without a
// translation, the text of go-form's catalogue for the language of loc
// shows, or its English text (ownText), such as "This field is required."
// for TranslationKeyRequired.
func (f *Form) validationErrorTranslated(loc Localizer, key string, args ...any) string {
	if len(args) == 1 && args[0] == nil {
		args = nil
	}

	return f.translateKey(loc, key, args)
}

// withoutContext removes the "form||" context from a message that was not
// translated.
func withoutContext(message string) string {
	return strings.TrimPrefix(message, translationContext)
}

// ValidateForm checks form, a pointer to a struct with go-form's tags,
// against the rules of its tags, with go-form's English messages; see
// ValidateFormLocalized.
func (f *Form) ValidateForm(form any) FieldErrors {
	return f.ValidateFormLocalized(form, &DefaultLocalizer{})
}

// ValidateFormLocalized checks form, a pointer to a struct with go-form's
// tags, against the rules of its tags, with messages for loc. The tags
// include the overrides of an Overrider model; form may also be a
// RenderModel, whose Overrides then apply on top, and whose Options
// replace values tags: a submitted value must be one of them.
func (f *Form) ValidateFormLocalized(form any, loc Localizer) FieldErrors {
	wrapped, ok := renderModelOf(form)
	if !ok {
		return f.validateStruct(form, loc, nil, nil)
	}

	return f.validateStruct(wrapped.Model, loc, wrapped.Overrides, wrapped.Options)
}

// validateStruct checks form, a struct or a pointer to one, with the
// overrides of outer on top of its own and the options of a render.
func (f *Form) validateStruct(form any, loc Localizer, outer Overrides, options FieldOptions) FieldErrors {
	v := reflect.ValueOf(form)

	overrides := overridesOf(v).merge(outer)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	errList := f.validateTags(v, loc, overrides, options) // built-in validations

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		name := validationFieldName(t.Field(i))
		field := overrides.apply(name, t.Field(i))
		value := v.Field(i)

		// A nested struct is validated as a form of its own.
		if nested, ok := nestedModel(field, value); ok {
			errs := f.validateStruct(nested, loc, below(overrides, name), below(options, name))
			errList = append(errList, nestedErrors(field, errs)...)

			continue
		}

		errList = append(errList, f.validateCustom(field, value, loc)...)
	}

	return errList
}

// nestedModel returns a pointer to the struct value holds, the value of
// field, when it is a nested form: a struct, or a pointer to one that is
// not nil, other than a time, an uploaded file or the inside of a select
// (see nestedForm).
func nestedModel(field reflect.StructField, value reflect.Value) (any, bool) {
	if !nestedForm(field) {
		return nil, false
	}

	switch {
	case value.Kind() == reflect.Struct && field.Type.PkgPath() != "time":
		return value.Addr().Interface(), true
	case value.Kind() == reflect.Pointer && !value.IsNil() &&
		value.Elem().Kind() == reflect.Struct && field.Type.Elem().PkgPath() != "time":
		return value.Interface(), true
	}

	return nil, false
}

// validateCustom checks value with the registered rules the validate tag
// of field names (RegisterValidationMethod, RegisterLocalizedValidation).
func (f *Form) validateCustom(field reflect.StructField, value reflect.Value, loc Localizer) FieldErrors {
	validateTag := field.Tag.Get("validate")
	if validateTag == "" {
		return nil
	}

	var errList FieldErrors

	for _, validatorName := range strings.Split(validateTag, ",") {
		validatorName = strings.TrimSpace(validatorName)
		if validatorName == "" {
			continue
		}

		if fn, ok := f.validators[validatorName]; ok {
			for _, validationErr := range fn(value.Interface(), field) {
				errList = append(errList, customError(field, validatorName, validationErr, nil))
			}
		}

		if fn, ok := f.localizedValidators[validatorName]; ok {
			getErr := f.messageFunc(field, loc)
			translate := func(message string) string {
				return getErr(validatorName, message, nil)
			}

			for _, validationErr := range fn(value.Interface(), field, loc) {
				errList = append(errList, customError(field, validatorName, validationErr, translate))
			}
		}
	}

	return errList
}

// nestedErrors are the errors of a nested form, named below field, such as
// "address.city".
func nestedErrors(field reflect.StructField, errs FieldErrors) FieldErrors {
	nested := make(FieldErrors, 0, len(errs))
	for _, err := range errs {
		name, message := err.FieldError()

		rule := ""
		if validationErr, ok := err.(FieldValidationError); ok {
			rule = validationErr.Rule
		}

		nested = append(nested, FieldValidationError{
			Field: validationFieldName(field) + "." + name,
			Err:   message,
			Rule:  rule,
		})
	}

	return nested
}

// customError is an error a custom validator returned for field: named as
// the field renders when it names the Go field, with the validator's name as
// its rule when it has none. translate, when not nil, turns the message into
// the field's message for the rule.
func customError(field reflect.StructField, validator string, err FieldError, translate func(string) string) FieldError {
	name, message := err.FieldError()

	validationErr, ok := err.(FieldValidationError)
	if !ok && name != field.Name && translate == nil {
		return err
	}

	if !ok {
		validationErr = FieldValidationError{Field: name, Err: message}
	}

	if validationErr.Field == field.Name {
		validationErr.Field = validationFieldName(field)
	}

	if validationErr.Rule == "" {
		validationErr.Rule = validator
	}

	if translate != nil {
		validationErr.Err = translate(validationErr.Err)
	}

	return validationErr
}

// validationFieldName matches the name used by form mapping and rendering.
func validationFieldName(field reflect.StructField) string {
	if name := field.Tag.Get(tagName); name != "" {
		return name
	}
	return field.Name
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

		de := reflect.DeepEqual(v.Interface(), z.Interface())

		// Use fmt.Sprint for a more forgiving comparison (e.g., for UUIDs, time.Time, etc.)
		if de == false && fmt.Sprint(v.Interface()) == fmt.Sprint(z.Interface()) {
			return true
		}

		return de
	}
}

// isPassword reports whether field is a password input, whose value counts
// exactly as typed, spaces included.
func isPassword(field reflect.StructField) bool {
	return strings.TrimSpace(field.Tag.Get("form")) == "input,password"
}

// nestedForm reports whether field is a struct, or a pointer to one, whose
// fields are validated as a nested form: not a select, such as a
// SortedSelect, an uploaded file or an unexported field.
func nestedForm(field reflect.StructField) bool {
	if !field.IsExported() {
		return false
	}

	if _, ok := fileField(field.Type); ok {
		return false
	}

	if field.Type.Implements(sortedMapperType) || field.Type.Implements(mapperType) {
		return false
	}

	return true
}

package form

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"reflect"
	"sync"

	"github.com/donseba/go-form/v2/csrf"
	"github.com/donseba/go-form/v2/templates"
	"github.com/donseba/go-form/v2/types"
)

type (
	DefaultLocalizer struct{}

	TranslationFunc func(loc types.Localizer, key string, args ...any) string
	ValidationFunc  func(fieldValue any, fieldStruct reflect.StructField) FieldErrors

	// LocalizedValidationFunc is a custom validator registered with
	// RegisterLocalizedValidation. It gets the localizer of the
	// validation, and the messages of the errors it returns are
	// translation keys: the form translates them for loc, like go-form's
	// own messages.
	LocalizedValidationFunc func(fieldValue any, fieldStruct reflect.StructField, loc Localizer) FieldErrors

	FieldErrors []FieldError

	FieldError interface {
		FieldError() (field, err string)
	}

	Info struct {
		Target     string `json:"target,omitempty"`
		Method     string `json:"method,omitempty"`
		SubmitText string `json:"submit,omitempty"`

		// CancelTarget enables an optional cancel action rendered by templates (as a link).
		// When empty, no cancel control is rendered.
		CancelTarget string `json:"cancel_target,omitempty"`
		// CancelText is the label for the cancel action. If empty and CancelTarget is set,
		// templates should fall back to a sensible default (e.g. "Cancel").
		CancelText string `json:"cancel_text,omitempty"`

		Attributes map[string]string `json:"attributes,omitempty"`
		CsrfValue  string            `json:"csrf_value,omitempty"` // CSRF token value
		CsrfField  string            `json:"csrf_field,omitempty"` // Name of the CSRF field (defaults to "_csrf")

		// ReadOnly shows the form without letting it be changed: every
		// field is disabled and there is no submit button. A cancel link
		// (CancelTarget) still shows, for example as a way back.
		ReadOnly bool `json:"read_only,omitempty"`
		// Honeypot is the name of a spam trap: a text field hidden from
		// people, which many bots fill in. Check it with IsSpam.
		Honeypot string `json:"honeypot,omitempty"`
		// Hidden are hidden inputs the form sends along, such as the page
		// to return to: an input per value, in name order, rendered by
		// form_open and form_render after the CSRF token and the honeypot.
		Hidden url.Values `json:"hidden,omitempty"`
		// IDPrefix goes before the id of every field of the form, followed
		// by a dash, and so before the ids its labels, descriptions and
		// errors refer to (for, aria-labelledby, aria-describedby): with
		// "newsletter" the email field is newsletter-email. Give forms
		// that share a page different prefixes, so their ids do not
		// collide. Empty keeps the ids as the fields' names.
		IDPrefix string `json:"id_prefix,omitempty"`
	}

	// RenderModel associates form metadata with a model without requiring the
	// model's concrete type to embed Info. Construct it with WithInfo,
	// WithContextInfo, or WithRequestInfo.
	RenderModel struct {
		Info  Info
		Model any
		// Overrides replace struct tags of the model's fields for this
		// render, such as a maximum that depends on the places left; they
		// win over those of an Overrider model. ValidateForm and
		// ValidateFormLocalized apply them too when they get the
		// RenderModel.
		Overrides Overrides
		// Options are the choices of fields that are only known at
		// render time, such as the options of a poll, by field name; they
		// replace a values tag. ValidateForm, ValidateFormLocalized and
		// Bind check a submitted value against them when they get the
		// RenderModel. See FieldOptions.
		Options FieldOptions
		// Carry shows a failed submission of the form again: its values
		// in place of the model's (hidden inputs keep the model's), and
		// its errors at their fields when the template passes none. It
		// only applies when the model's hidden inputs hold the values
		// the carry was submitted with; nil shows the model as it is.
		// See NewCarry.
		Carry *Carry
	}

	Form struct {
		validators          map[string]ValidationFunc
		localizedValidators map[string]LocalizedValidationFunc
		translationEnabled  bool
		translationFunc     TranslationFunc
		csrfStore           csrf.Store // CSRF token storage

		// themeName selects which embedded gohtml theme to use ("bootstrap", "tailwind", "tailwindv4", "plain").
		themeName string
	}
)

// WithInfo associates form metadata with a model for rendering. The model's
// fields keep their original names and nesting; RenderModel is transparent to
// the transformer.
func WithInfo(model any, info Info) RenderModel {
	return RenderModel{Info: info, Model: model}
}

// WithContextInfo associates form metadata with a model and injects any CSRF
// token available in ctx.
func WithContextInfo(ctx context.Context, model any, info Info) RenderModel {
	InjectCSRFTokenContext(ctx, &info)
	return WithInfo(model, info)
}

// WithRequestInfo associates form metadata with a model and injects any CSRF
// token available on r.
func WithRequestInfo(r *http.Request, model any, info Info) RenderModel {
	if r != nil {
		return WithContextInfo(r.Context(), model, info)
	}
	return WithInfo(model, info)
}

func (d *DefaultLocalizer) GetLocale() string {
	return "en" // Default locale
}

// SetCSRFStore sets the CSRF store for the Form. Configure a shared store before
// creating middleware when requests may reach different application instances.
func (f *Form) SetCSRFStore(store csrf.Store) {
	f.csrfStore = store
}

// HasCSRFStore checks if the form has a CSRF store
func (f *Form) HasCSRFStore() bool {
	return f.csrfStore != nil
}

// GetCSRFStore returns the CSRF store
func (f *Form) GetCSRFStore() csrf.Store {
	return f.csrfStore
}

// NewTranslatedForm creates a new form with translation support.
func NewTranslatedForm(translationFunc TranslationFunc) *Form {
	f := &Form{
		validators:          make(map[string]ValidationFunc),
		localizedValidators: make(map[string]LocalizedValidationFunc),
		csrfStore:           csrf.NewDefaultMemoryCSRFStore(),
		// Default theme. Users can override via SetTheme(...).
		themeName: "bootstrap",
	}

	f.enableTranslation(translationFunc)
	return f
}

// NewForm creates a new form without translation support.
func NewForm() *Form {
	return &Form{
		validators:          make(map[string]ValidationFunc),
		localizedValidators: make(map[string]LocalizedValidationFunc),
		csrfStore:           csrf.NewDefaultMemoryCSRFStore(),
		// Default theme. Users can override via SetTheme(...).
		themeName: "bootstrap",
	}
}

func (f *Form) enableTranslation(fn TranslationFunc) {
	f.translationEnabled = true
	f.translationFunc = fn
}

// FuncMap returns the template functions that render forms:
//
//   - form_render model errors renders the whole form; form_render_localized
//     localizer model errors renders it translated.
//   - form_field model "Email" errors renders one field with its label,
//     description and errors (form_field_localized localizer model "Email"
//     errors), so a template can lay out the fields itself.
//   - form_errors errors "Email" returns the messages of one field.
//   - form_open model renders the opening form tag with the hidden fields
//     (CSRF token, honeypot, Info.Hidden), form_buttons model the submit and cancel
//     buttons and form_close the closing tag; form_open_localized and
//     form_buttons_localized take a localizer first. form_open takes
//     attributes of the template's own after the model: name/value pairs
//     and template.HTMLAttr values, such as form_open model "class"
//     "booking" (htmxAttrs …); see Open.
//   - form_hidden model renders the hidden fields alone (form_hidden_localized
//     localizer model), and form_honeypot name the spam trap input
//     (Info.Honeypot), for a form whose markup a template writes itself.
//   - form_has_fields model reports whether model has a field to render.
//
// Render, Field, Open, Hidden, Buttons, Close and Honeypot render
// the same from Go.
func (f *Form) FuncMap() template.FuncMap {
	funcMap := template.FuncMap{
		"form_render":            f.formRender,
		"form_render_localized":  f.formRenderLocalized,
		"form_has_fields":        HasFields,
		"form_field":             f.formField,
		"form_field_localized":   f.formFieldLocalized,
		"form_errors":            FieldErrors.Messages,
		"form_open":              f.formOpen,
		"form_open_localized":    f.formOpenLocalized,
		"form_buttons":           f.formButtons,
		"form_buttons_localized": f.formButtonsLocalized,
		"form_close":             f.formClose,
		"form_honeypot":          f.honeypotField,
		"form_hidden":            f.formHidden,
		"form_hidden_localized":  f.formHiddenLocalized,
	}

	return funcMap
}

// HasFields reports whether model contains at least one renderable field. Form
// metadata does not count as a field.
func HasFields(model any) bool {
	transformer, err := NewTransformer(model)
	if err != nil {
		return false
	}
	for _, field := range transformer.Fields {
		if field.Type != types.FieldTypeForm {
			return true
		}
	}
	return false
}

func (f *Form) formRender(v any, errs FieldErrors, kv ...any) (template.HTML, error) {
	return f.formRenderFunc(&DefaultLocalizer{}, v, errs, kv...)
}

func (f *Form) formRenderLocalized(loc types.Localizer, v any, errs FieldErrors, kv ...any) (template.HTML, error) {
	return f.formRenderFunc(loc, v, errs, kv...)
}

var initThemesOnce sync.Once
var initThemesErr error

func ensureThemesLoaded() error {
	initThemesOnce.Do(func() {
		initThemesErr = templates.InitThemes()
	})
	return initThemesErr
}

// themePrint mirrors the legacy form_print behavior.
func (f *Form) themePrint(loc types.Localizer, key string, args ...any) string {
	return f.translate(loc, key, args)
}

// translate translates key, a label or other text, with args. Without a
// translation function the text is formatted as is, go-form's own texts
// show in the language of loc when go-form has a catalogue for it and in
// English otherwise (ownText), and a "form||" context never shows.
func (f *Form) translate(loc types.Localizer, key string, args []any) string {
	// Empty gettext keys represent the translation catalog header. Form
	// metadata frequently leaves optional labels and placeholders empty, so
	// never send an empty value to the application's translator.
	if key == "" {
		return ""
	}

	return f.translateKey(loc, key, args)
}

// translateKey translates key with args as translate does, but sends an
// empty key to the translation function too.
func (f *Form) translateKey(loc types.Localizer, key string, args []any) string {
	translated := key

	switch {
	case f.translationEnabled && f.translationFunc != nil:
		translated = f.translationFunc(loc, key, args...)
	case len(args) > 0:
		translated = fmt.Sprintf(key, args...)
	}

	return withoutContext(ownText(loc, key, args, translated))
}

// requiredLabel is the text screen readers announce after the label of a
// required field. Without a translation it is the English text of
// TranslationKeyRequiredLabel, "(required)".
func (f *Form) requiredLabel(loc types.Localizer) string {
	return f.translate(loc, TranslationKeyRequiredLabel, nil)
}

// RegisterValidationMethod registers fn as the custom validator name, which
// the validate tag of a field names, such as validate:"slug". The messages
// of the errors it returns show as they are; RegisterLocalizedValidation
// registers a validator whose messages are translated. Register validators
// before validating.
func (f *Form) RegisterValidationMethod(name string, fn ValidationFunc) {
	f.validators[name] = fn
}

// RegisterLocalizedValidation registers a custom validator for the
// validate tag whose messages are translated, like go-form's own:
//
//	f.RegisterLocalizedValidation("slug", func(v any, field reflect.StructField, loc form.Localizer) form.FieldErrors {
//		if s, _ := v.(string); s != "" && !slugPattern.MatchString(s) {
//			return form.FieldErrors{form.FieldValidationError{Field: field.Name, Err: "Use lowercase letters, digits and dashes"}}
//		}
//		return nil
//	})
//
// The form translates the message of each error with its translation
// function for loc. A <name>Msg tag on the field, such as slugMsg, replaces
// the message, as errorMsg does; the error's Rule is the validator's name
// unless it sets one. An error that names the Go field is named as the
// field renders. Register validators before validating; a name registered
// with RegisterValidationMethod too runs both.
func (f *Form) RegisterLocalizedValidation(name string, fn LocalizedValidationFunc) {
	if f.localizedValidators == nil {
		f.localizedValidators = make(map[string]LocalizedValidationFunc)
	}

	f.localizedValidators[name] = fn
}

func (f *Form) GetValidationMethod(name string) (ValidationFunc, bool) {
	fn, ok := f.validators[name]
	return fn, ok
}

func scanError(errs FieldErrors) map[string][]string {
	ret := make(map[string][]string)
	for _, err := range errs {
		field, fieldErr := err.FieldError()
		ret[field] = append(ret[field], fieldErr)
	}
	return ret
}

// SetTheme selects which gohtml theme is used for rendering.
func (f *Form) SetTheme(name string) {
	f.themeName = name
}

func (f *Form) themeFieldHTML(theme *templates.Theme, loc types.Localizer, field types.FormField, errorMap map[string][]string) (template.HTML, error) {
	// Render label. A single checkbox renders its own label next to the box;
	// a group of radio buttons or checkboxes is a fieldset, its label the
	// legend.
	var (
		labelH template.HTML
		err    error
	)
	if field.Type != types.FieldTypeCheckbox {
		labelH, err = f.themeExec(
			theme,
			themeTemplate(theme, field, "label", "legend"),
			map[string]any{"Field": field, "Loc": loc},
			nil,
		)
		if err != nil {
			return "", err
		}
	}

	// Render control
	var control template.HTML
	switch field.Type {
	case types.FieldTypeInput:
		control, err = f.themeExec(theme, "input", map[string]any{"Type": field.InputType.String(), "Field": field, "Loc": loc}, nil)
	case types.FieldTypeDropdown, types.FieldTypeDropdownMapped:
		control, err = f.themeExec(theme, "select", map[string]any{"Field": field, "Loc": loc}, nil)
	case types.FieldTypeTextArea:
		control, err = f.themeExec(theme, "textarea", map[string]any{"Field": field, "Loc": loc}, nil)
	case types.FieldTypeCheckbox:
		control, err = f.themeExec(theme, "checkbox", map[string]any{"Field": field, "Loc": loc}, nil)
	case types.FieldTypeRadios:
		name := "radio"
		if radioGroup(field) {
			name = "radio-group"
		}
		control, err = f.themeExec(theme, name, map[string]any{"Field": field, "Loc": loc}, nil)
	case types.FieldTypeMultiCheckbox:
		control, err = f.themeExec(theme, "multicheckbox", map[string]any{"Field": field, "Loc": loc}, nil)
	default:
		return "", fmt.Errorf("unsupported field type %q for theme renderer", field.Type)
	}
	if err != nil {
		return "", err
	}

	control, err = f.themeWrapIfGrouped(theme, field, control)
	if err != nil {
		return "", err
	}

	return f.themeWrapField(theme, loc, field, labelH, control, errorMap)
}

package form

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
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
	}

	// RenderModel associates form metadata with a model without requiring the
	// model's concrete type to embed Info. Construct it with WithInfo,
	// WithContextInfo, or WithRequestInfo.
	RenderModel struct {
		Info  Info
		Model any
	}

	Form struct {
		validators         map[string]ValidationFunc
		translationEnabled bool
		translationFunc    TranslationFunc
		csrfStore          csrf.Store // CSRF token storage

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

// SetCSRFStore sets the CSRF store for the Form
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
		validators: make(map[string]ValidationFunc),
		csrfStore:  csrf.NewDefaultMemoryCSRFStore(),
		// Default theme. Users can override via SetTheme(...).
		themeName: "bootstrap",
	}

	f.enableTranslation(translationFunc)
	return f
}

// NewForm creates a new form without translation support.
func NewForm() *Form {
	return &Form{
		validators: make(map[string]ValidationFunc),
		csrfStore:  csrf.NewDefaultMemoryCSRFStore(),
		// Default theme. Users can override via SetTheme(...).
		themeName: "bootstrap",
	}
}

func (f *Form) enableTranslation(fn TranslationFunc) {
	f.translationEnabled = true
	f.translationFunc = fn
}

func (f *Form) FuncMap() template.FuncMap {
	funcMap := template.FuncMap{
		"form_render":           f.formRender,
		"form_render_localized": f.formRenderLocalized,
		"form_has_fields":       HasFields,
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
	if f.translationEnabled && f.translationFunc != nil {
		return f.translationFunc(loc, key, args...)
	}
	if len(args) > 0 {
		return fmt.Sprintf(key, args...)
	}
	return key
}

func (f *Form) formRenderFunc(loc types.Localizer, v any, errs FieldErrors, _ ...any) (template.HTML, error) {
	theme, err := f.getTheme()
	if err != nil {
		return "", err
	}

	tr, err := NewTransformer(v)
	if err != nil {
		return "", err
	}
	fieldErrors := scanError(errs)

	// Build formField and inner HTML.
	var inner template.HTML
	var formField *types.FormField

	for i, field := range tr.Fields {
		if field.Type == types.FieldTypeForm {
			formField = &tr.Fields[i]
			if formField.Attributes == nil {
				formField.Attributes = make(map[string]string)
			}
			// Submit text is already set by the Transformer into FormField.Label, so don't overwrite it here.
			continue
		}

		if field.Type == types.FieldTypeGroup {
			funcs := template.FuncMap{
				"fields": func() template.HTML {
					return f.renderGroupFields(theme, loc, field.Fields, fieldErrors)
				},
				"errors": func() []string {
					if es, ok := fieldErrors[field.Name]; ok {
						return es
					}
					return nil
				},
			}

			h, err := f.themeExec(theme, "group", map[string]any{"Field": field, "Loc": loc}, funcs)
			if err != nil {
				return "", err
			}
			inner += h
			continue
		}

		fh, err := f.themeFieldHTML(theme, loc, field, fieldErrors)
		if err != nil {
			return "", err
		}
		inner += fh
	}

	if formField == nil {
		return inner, nil
	}

	csrfHTML := csrfHiddenInputHTML(v)

	funcs := template.FuncMap{
		"fields": func() template.HTML { return csrfHTML + inner },
	}

	formData := struct {
		Field types.FormField
		Loc   types.Localizer
	}{
		Field: *formField,
		Loc:   loc,
	}

	h, err := f.themeExec(theme, "form", formData, funcs)
	if err != nil {
		return "", err
	}
	return h, nil
}

func (f *Form) RegisterValidationMethod(name string, fn ValidationFunc) {
	f.validators[name] = fn
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
	// Render label
	labelH, err := f.themeExec(theme, "label", map[string]any{"Field": field, "Loc": loc}, nil)
	if err != nil {
		return "", err
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
		if field.InputType == types.InputFieldTypeRadioStruct || field.InputType == types.InputFieldTypeRadioGroup {
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

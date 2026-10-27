package form

import (
	"html/template"
	"strings"

	"github.com/donseba/go-form/v2/types"
)

// The methods below render a form from Go, as the template functions of
// FuncMap do, for applications that wrap them in template functions of
// their own, such as functions that take the localizer of the page being
// rendered from a context. A nil localizer renders the texts as written
// (DefaultLocalizer).

// Render renders the whole form of model, a struct with form metadata
// (WithInfo) or a Schema, with errs at their fields, translated for loc:
// form_render_localized.
func (f *Form) Render(loc Localizer, model any, errs FieldErrors) (template.HTML, error) {
	return f.formRenderFunc(orDefaultLocalizer(loc), model, errs)
}

// Field renders the field named name of model (as rendered, such as
// "address.city") with its label, description and errors, translated for
// loc: form_field_localized.
func (f *Form) Field(loc Localizer, model any, name string, errs FieldErrors) (template.HTML, error) {
	return f.formFieldLocalized(orDefaultLocalizer(loc), model, name, errs)
}

// Open renders the opening form tag of model, made with WithInfo, with
// its hidden fields: the CSRF token, the honeypot and Info.Hidden. attrs
// are more attributes of the tag, if any: name/value pairs and
// template.HTMLAttr values, such as those another library renders for the
// form:
//
//	open, err := forms.Open(loc, model, "class", "booking", htmxAttrs)
//
// A class adds to the theme's, other attributes replace those of
// Info.Attributes, and action, method and enctype cannot be given: they
// come from Info. form_open_localized with attributes after the model.
func (f *Form) Open(loc Localizer, model any, attrs ...any) (template.HTML, error) {
	return f.formOpenLocalized(orDefaultLocalizer(loc), model, attrs...)
}

// Hidden renders the hidden inputs Open adds after the opening tag of
// model, made with WithInfo: the CSRF token, the honeypot and Info.Hidden,
// for a template that writes the form tag itself: form_hidden_localized.
func (f *Form) Hidden(loc Localizer, model any) (template.HTML, error) {
	return f.formHiddenLocalized(orDefaultLocalizer(loc), model)
}

// Buttons renders the submit button and cancel link of model, made with
// WithInfo, translated for loc: form_buttons_localized.
func (f *Form) Buttons(loc Localizer, model any) (template.HTML, error) {
	return f.formButtonsLocalized(orDefaultLocalizer(loc), model)
}

// Close renders the closing form tag: form_close.
func (f *Form) Close() (template.HTML, error) {
	return f.formClose()
}

// Honeypot renders the spam trap input named name (Info.Honeypot) with
// the form's theme, for a form whose markup a template writes itself:
// form_honeypot. Check it with IsSpam.
func (f *Form) Honeypot(loc Localizer, name string) (template.HTML, error) {
	theme, err := f.getTheme()
	if err != nil {
		return "", err
	}

	data := map[string]any{"Field": types.FormField{Name: name}, "Loc": orDefaultLocalizer(loc)}

	trap, err := f.themeExec(theme, "honeypot", data, nil)
	if err != nil {
		return "", err
	}

	//nolint:gosec // The theme's template escaped the markup.
	return template.HTML(strings.TrimSpace(string(trap))), nil
}

// orDefaultLocalizer is loc, or DefaultLocalizer when loc is nil.
func orDefaultLocalizer(loc Localizer) Localizer {
	if loc == nil {
		return &DefaultLocalizer{}
	}

	return loc
}

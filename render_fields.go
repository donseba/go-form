package form

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/donseba/go-form/v2/templates"
	"github.com/donseba/go-form/v2/types"
)

// renderPlan is a model ready to render: its fields with their errors, and
// its form metadata.
type renderPlan struct {
	theme *templates.Theme
	loc   types.Localizer
	// form is the form metadata (Info); nil renders the fields alone.
	form   *types.FormField
	fields []types.FormField
	errors map[string][]string
	// hidden are the CSRF token and honeypot fields.
	hidden template.HTML
}

// plan prepares v for rendering with loc and errs: date-times in the zone
// of a TimeLocalizer, errors on their fields, every field disabled in a
// read-only form, and multipart encoding for a form with a file input.
func (f *Form) plan(loc types.Localizer, v any, errs FieldErrors) (*renderPlan, error) {
	theme, err := f.getTheme()
	if err != nil {
		return nil, err
	}

	tr, err := newTransformer(v, locationOf(loc))
	if err != nil {
		return nil, err
	}

	p := &renderPlan{theme: theme, loc: loc}

	for i := range tr.Fields {
		if tr.Fields[i].Type == types.FieldTypeForm {
			form := tr.Fields[i]
			p.form = &form

			continue
		}

		p.fields = append(p.fields, tr.Fields[i])
	}

	// A failed submission of this form shows again (RenderModel.Carry).
	if carry := modelCarry(v); carry != nil && carry.matches(p.fields) {
		carry.restore(p.fields)

		if len(errs) == 0 {
			errs = carry.FieldErrors()
		}
	}

	p.errors = scanError(errs)

	readOnly := p.form != nil && p.form.ReadOnly
	prepareFields(p.fields, p.errors, readOnly)
	f.sortOptions(loc, p.fields)

	if p.form == nil {
		return p, nil
	}

	if p.form.Attributes == nil {
		p.form.Attributes = make(map[string]string)
	}

	// Files are only sent with multipart encoding.
	if _, set := p.form.Attributes["enctype"]; !set && hasFileInput(p.fields) {
		p.form.Attributes["enctype"] = "multipart/form-data"
	}

	p.hidden = csrfHiddenInputHTML(v)
	if p.form.Honeypot != "" {
		data := map[string]any{"Field": types.FormField{Name: p.form.Honeypot}, "Loc": loc}

		trap, err := f.themeExec(theme, "honeypot", data, nil)
		if err != nil {
			return nil, err
		}

		p.hidden += trap
	}

	if info, ok := modelInfo(v); ok {
		p.hidden += hiddenInputsHTML(info)
		prefixIDs(p.fields, info.IDPrefix)
	}

	return p, nil
}

// renderModelOf returns v as a RenderModel when it is one, or a pointer to
// one that is not nil.
func renderModelOf(v any) (RenderModel, bool) {
	switch wrapped := v.(type) {
	case RenderModel:
		return wrapped, true
	case *RenderModel:
		if wrapped != nil {
			return *wrapped, true
		}
	}

	return RenderModel{}, false
}

// prefixIDs puts prefix and a dash before the ids of fields and their
// nested fields (Info.IDPrefix). An empty prefix changes nothing.
func prefixIDs(fields []types.FormField, prefix string) {
	if prefix == "" {
		return
	}

	for i := range fields {
		if fields[i].Id != "" {
			fields[i].Id = prefix + "-" + fields[i].Id
		}

		prefixIDs(fields[i].Fields, prefix)
	}
}

// sortOptions orders the options of fields with SortByLabel by their
// labels as they render for loc: translated when they are translated.
func (f *Form) sortOptions(loc types.Localizer, fields []types.FormField) {
	for i := range fields {
		f.sortOptions(loc, fields[i].Fields)

		if !fields[i].SortByLabel || len(fields[i].Values) < 2 {
			continue
		}

		values := fields[i].Values
		labels := make([]string, len(values))

		for j, value := range values {
			labels[j] = value.Name
			if value.Translate {
				labels[j] = f.translate(loc, value.Name, nil)
			}
		}

		order := make([]int, len(values))
		for j := range order {
			order[j] = j
		}

		sort.SliceStable(order, func(a, b int) bool {
			return labelLess(labels[order[a]], labels[order[b]])
		})

		sorted := make([]types.FieldValue, 0, len(values))
		for _, j := range order {
			sorted = append(sorted, values[j])
		}

		fields[i].Values = sorted
	}
}

// locationOf is the time zone of a TimeLocalizer, or nil.
func locationOf(loc types.Localizer) *time.Location {
	if zoned, ok := loc.(TimeLocalizer); ok {
		return zoned.Location()
	}

	return nil
}

// prepareFields gives fields and their nested fields their error messages,
// and disables them in a read-only form.
func prepareFields(fields []types.FormField, errors map[string][]string, readOnly bool) {
	for i := range fields {
		fields[i].Errors = errors[fields[i].Name]
		if readOnly {
			fields[i].Disabled = true
		}

		prepareFields(fields[i].Fields, errors, readOnly)
	}
}

// hasFileInput reports whether fields or their nested fields have a file
// input.
func hasFileInput(fields []types.FormField) bool {
	for _, field := range fields {
		if field.Type == types.FieldTypeInput && field.InputType == types.InputFieldTypeFile {
			return true
		}

		if hasFileInput(field.Fields) {
			return true
		}
	}

	return false
}

// findField finds the field named name (as rendered, such as "address.city"),
// also among nested fields.
func findField(fields []types.FormField, name string) (types.FormField, bool) {
	for _, field := range fields {
		if field.Name == name {
			return field, true
		}

		if found, ok := findField(field.Fields, name); ok {
			return found, true
		}
	}

	return types.FormField{}, false
}

// fieldNames are the names findField finds in fields, groups and their
// fields included, in the order of the form.
func fieldNames(fields []types.FormField) []string {
	var names []string

	for _, field := range fields {
		names = append(names, field.Name)
		names = append(names, fieldNames(field.Fields)...)
	}

	return names
}

func (f *Form) formRenderFunc(loc types.Localizer, v any, errs FieldErrors, _ ...any) (template.HTML, error) {
	p, err := f.plan(loc, v, errs)
	if err != nil {
		return "", err
	}

	var inner template.HTML

	for _, field := range p.fields {
		h, err := f.renderField(p, field)
		if err != nil {
			return "", err
		}

		inner += h
	}

	if p.form == nil {
		return inner, nil
	}

	data, funcs := p.formTemplate(nil)
	funcs["fields"] = func() template.HTML { return p.hidden + inner }

	return f.themeExec(p.theme, "form", data, funcs)
}

// formTemplateData is the data of the form templates (form, form-open,
// form-buttons).
type formTemplateData struct {
	// Field is the form metadata. Its Attributes are those of
	// Info.Attributes that render as they are; class and style merge into
	// the theme's (form_open_class, form_open_style).
	Field types.FormField
	Loc   types.Localizer
}

// formTemplate is the data and the functions of the form templates, with
// the attributes a template gave form_open (attrs): form_open_class and
// form_open_style, the classes and inline style of the form besides the
// theme's, and form_open_attributes, the template's other attributes.
func (p *renderPlan) formTemplate(attrs []formAttribute) (formTemplateData, template.FuncMap) {
	tag := buildOpenTag(p.theme.Classes.Form.Class, p.form.Attributes, attrs)

	field := *p.form
	field.Attributes = tag.attributes

	funcs := template.FuncMap{
		"form_open_class":      func() string { return tag.class },
		"form_open_style":      func() template.CSS { return tag.style },
		"form_open_attributes": func() template.HTMLAttr { return tag.extra },
	}

	return formTemplateData{Field: field, Loc: p.loc}, funcs
}

// renderField renders one field, or a group with its fields.
func (f *Form) renderField(p *renderPlan, field types.FormField) (template.HTML, error) {
	if field.Type != types.FieldTypeGroup {
		return f.themeFieldHTML(p.theme, p.loc, field, p.errors)
	}

	funcs := template.FuncMap{
		"fields": func() template.HTML {
			return f.renderGroupFields(p.theme, p.loc, field.Fields, p.errors)
		},
		"errors": func() []string {
			return p.errors[field.Name]
		},
	}

	return f.themeExec(p.theme, "group", map[string]any{"Field": field, "Loc": p.loc}, funcs)
}

func (f *Form) formField(v any, name string, errs FieldErrors) (template.HTML, error) {
	return f.formFieldLocalized(&DefaultLocalizer{}, v, name, errs)
}

// formFieldLocalized renders the field named name of v (as rendered, such as
// "address.city") with its label, description and errors.
func (f *Form) formFieldLocalized(loc types.Localizer, v any, name string, errs FieldErrors) (template.HTML, error) {
	p, err := f.plan(loc, v, errs)
	if err != nil {
		return "", err
	}

	field, ok := findField(p.fields, name)
	if !ok {
		return "", fmt.Errorf("form: no field %q (fields: %s)", name, strings.Join(fieldNames(p.fields), ", "))
	}

	return f.renderField(p, field)
}

// formOpen is form_open: the opening form tag of v with its hidden
// fields, and the attributes the template gives it (attrs).
func (f *Form) formOpen(v any, attrs ...any) (template.HTML, error) {
	return f.formOpenLocalized(&DefaultLocalizer{}, v, attrs...)
}

// formOpenLocalized renders the opening form tag of v with its hidden
// fields, and the attributes the template gives it (attrs, see
// openAttributes).
func (f *Form) formOpenLocalized(loc types.Localizer, v any, attrs ...any) (template.HTML, error) {
	given, err := openAttributes(attrs)
	if err != nil {
		return "", err
	}

	p, err := f.formPlan(loc, v)
	if err != nil {
		return "", err
	}

	data, funcs := p.formTemplate(given)

	open, err := f.themeExec(p.theme, "form-open", data, funcs)
	if err != nil {
		return "", err
	}

	return open + p.hidden, nil
}

// formHidden is form_hidden: the hidden inputs of v alone.
func (f *Form) formHidden(v any) (template.HTML, error) {
	return f.formHiddenLocalized(&DefaultLocalizer{}, v)
}

// formHiddenLocalized renders the hidden inputs form_open adds after the
// opening tag: the CSRF token, the honeypot and Info.Hidden.
func (f *Form) formHiddenLocalized(loc types.Localizer, v any) (template.HTML, error) {
	p, err := f.formPlan(loc, v)
	if err != nil {
		return "", err
	}

	return p.hidden, nil
}

func (f *Form) formButtons(v any) (template.HTML, error) {
	return f.formButtonsLocalized(&DefaultLocalizer{}, v)
}

// formButtonsLocalized renders the submit button and cancel link of v.
func (f *Form) formButtonsLocalized(loc types.Localizer, v any) (template.HTML, error) {
	p, err := f.formPlan(loc, v)
	if err != nil {
		return "", err
	}

	data, funcs := p.formTemplate(nil)

	return f.themeExec(p.theme, "form-buttons", data, funcs)
}

// formClose renders the closing form tag.
func (f *Form) formClose() (template.HTML, error) {
	theme, err := f.getTheme()
	if err != nil {
		return "", err
	}

	return f.themeExec(theme, "form-close", nil, nil)
}

// formPlan is plan for a model with form metadata.
func (f *Form) formPlan(loc types.Localizer, v any) (*renderPlan, error) {
	p, err := f.plan(loc, v, nil)
	if err != nil {
		return nil, err
	}

	if p.form == nil {
		return nil, fmt.Errorf("form: the model has no form metadata; wrap it with WithInfo")
	}

	return p, nil
}

// Messages returns the messages for the field named name (as rendered, such
// as "address.city"), for example to show them next to a field laid out by
// hand. Templates call it as form_errors.
func (e FieldErrors) Messages(name string) []string {
	var messages []string

	for _, err := range e {
		if field, message := err.FieldError(); field == name {
			messages = append(messages, message)
		}
	}

	return messages
}

// honeypotField renders the spam trap input named name with the form's
// theme: form_honeypot.
func (f *Form) honeypotField(name string) (template.HTML, error) {
	return f.Honeypot(nil, name)
}

// IsSpam reports whether the honeypot field name (Info.Honeypot) was filled
// in. People never see the field; answer such requests as if they succeeded,
// but store and send nothing.
func IsSpam(r *http.Request, name string) bool {
	return r.FormValue(name) != ""
}

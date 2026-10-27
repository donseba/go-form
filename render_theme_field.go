package form

import (
	"html/template"

	"github.com/donseba/go-form/v2/templates"
	"github.com/donseba/go-form/v2/types"
)

// themeWrapIfGrouped wraps control HTML with the theme's input-group template when
// GroupBefore/GroupAfter is set.
func (f *Form) themeWrapIfGrouped(theme *templates.Theme, field types.FormField, control template.HTML) (template.HTML, error) {
	if field.GroupBefore == "" && field.GroupAfter == "" {
		return control, nil
	}
	return f.themeExec(theme, "input-group", map[string]any{
		"GroupBefore": template.HTML(field.GroupBefore),
		"GroupAfter":  template.HTML(field.GroupAfter),
		"Input":       control,
	}, nil)
}

// themeWrapField renders the wrapper template using the provided label and control HTML.
func (f *Form) themeWrapField(theme *templates.Theme, loc types.Localizer, field types.FormField, label template.HTML, control template.HTML, errorMap map[string][]string) (template.HTML, error) {
	if field.InputType == types.InputFieldTypeHidden {
		return control, nil
	}

	funcs := template.FuncMap{
		"label": func() template.HTML { return label },
		"field": func() template.HTML { return control },
		"errors": func() []string {
			if es, ok := errorMap[field.Name]; ok {
				return es
			}
			return nil
		},
	}
	return f.themeExec(
		theme,
		themeTemplate(theme, field, "wrapper", "fieldset"),
		map[string]any{"Field": field, "Loc": loc},
		funcs,
	)
}

// radioGroup reports whether field renders as a group of radio buttons,
// one per option.
func radioGroup(field types.FormField) bool {
	return field.Type == types.FieldTypeRadios &&
		(field.InputType == types.InputFieldTypeRadioStruct || field.InputType == types.InputFieldTypeRadioGroup)
}

// optionGroup reports whether field renders as a group of options, which
// is a fieldset with a legend: a group of radio buttons, or a
// multicheckbox, a checkbox per option.
func optionGroup(field types.FormField) bool {
	return radioGroup(field) || field.Type == types.FieldTypeMultiCheckbox
}

// themeTemplate is the template of a part of field: group for a group of
// options (see optionGroup), which is a fieldset with a legend, when the
// theme has that template, and name otherwise, such as for a theme whose
// templates all come from a folder of its own (Theme.LoadTemplates).
func themeTemplate(theme *templates.Theme, field types.FormField, name, group string) string {
	if optionGroup(field) && theme.Templates.Lookup(group) != nil {
		return group
	}

	return name
}

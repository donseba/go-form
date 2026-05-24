package form

import (
	"html/template"

	"github.com/donseba/go-form/templates"
	"github.com/donseba/go-form/types"
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
	return f.themeExec(theme, "wrapper", map[string]any{"Field": field, "Loc": loc}, funcs)
}

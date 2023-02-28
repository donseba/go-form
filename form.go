package form

import (
	"html/template"
	"strings"
)

type Form struct {
	template      template.Template
	groupTemplate template.Template
}

func NewForm(template, groupTemplate template.Template) Form {
	return Form{
		template:      template,
		groupTemplate: groupTemplate,
	}
}

func (f *Form) FuncMap() template.FuncMap {
	return template.FuncMap{
		"form_render": f.formRender,
	}
}

func (f *Form) formRender(v interface{}, errs []FieldError) (template.HTML, error) {
	tr, err := NewTransformer(v)
	if err != nil {
		return "", err
	}

	fieldErrors := scanError(errs)

	var html template.HTML
	for _, field := range tr.Fields {
		if field.Type == FieldTypeGroup {
			gtpl, err := f.groupTemplate.Clone()
			if err != nil {
				return "", err
			}

			var sb strings.Builder
			gtpl = gtpl.Funcs(template.FuncMap{
				"fields": func() template.HTML {
					var subhtml template.HTML

					for _, subField := range field.Fields {
						fieldhtml, err := f.formFieldHTML(subField, fieldErrors)
						if err != nil {
							continue
						}

						subhtml = subhtml + fieldhtml
					}

					return subhtml
				},
			})

			err = gtpl.Execute(&sb, field)
			if err != nil {
				return "", err
			}
			html = html + template.HTML(sb.String())

			continue
		}

		fieldHTML, err := f.formFieldHTML(field, fieldErrors)
		if err != nil {
			return "", err
		}

		html = html + fieldHTML
	}
	return html, nil
}

func (f *Form) formFieldHTML(field FormField, errors map[string][]string) (template.HTML, error) {
	tpl, err := f.template.Clone()
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	tpl = tpl.Funcs(template.FuncMap{
		"errors": func() []string {
			if errs, ok := errors[field.Name]; ok {
				return errs
			}
			return nil
		},
	})
	err = tpl.Execute(&sb, field)
	if err != nil {
		return "", err
	}

	return template.HTML(sb.String()), nil
}

type FieldError interface {
	FieldError() (field, err string)
}

func scanError(errs []FieldError) map[string][]string {
	ret := make(map[string][]string)
	for _, err := range errs {
		field, fieldErr := err.FieldError()
		ret[field] = append(ret[field], fieldErr)
	}
	return ret
}

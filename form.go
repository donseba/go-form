package form

import (
	"errors"
	"html/template"
	"strings"
)

type Form struct {
	//template      template.Template
	//groupTemplate template.Template

	templateMap map[FieldType]map[InputFieldType]template.Template
}

func NewForm(templateMap map[FieldType]map[InputFieldType]string) Form {
	f := Form{
		templateMap: make(map[FieldType]map[InputFieldType]template.Template),
	}

	for fieldType, inputTemplates := range templateMap {
		f.templateMap[fieldType] = make(map[InputFieldType]template.Template)
		for inputType, tpl := range inputTemplates {
			t, err := template.New("field").Funcs(map[string]any{
				"errors": func() []string { return nil },     // Placeholder for error handling
				"field":  func() template.HTML { return "" }, // Placeholder for field rendering
				"fields": func() template.HTML { return "" }, // Placeholder for group fields
				"label":  func() template.HTML { return "" }, // Placeholder for label rendering
			}).Parse(tpl)
			if err != nil {
				panic(err) // Handle error appropriately in production code
			}
			f.templateMap[fieldType][inputType] = *t
		}
	}

	return f
}

func (f *Form) FuncMap() template.FuncMap {
	return template.FuncMap{
		"form_render": f.formRender,
	}
}

func (f *Form) formRender(v interface{}, errs []FieldError, kv ...any) (template.HTML, error) {
	tr, err := NewTransformer(v)
	if err != nil {
		return "", err
	}

	fieldErrors := scanError(errs)

	var data = make(map[string]any)
	if len(kv) > 0 {
		if len(kv)%2 != 0 {
			return "", errors.New("invalid dict call")
		}

		for i := 0; i < len(kv); i += 2 {
			key, ok := kv[i].(string)
			if !ok {
				return "", errors.New("dict keys must be strings")
			}
			data[key] = kv[i+1]
		}
	}

	var html template.HTML
	for _, field := range tr.Fields {
		if field.Type == FieldTypeGroup {
			tmpl, ok := f.templateMap[FieldTypeGroup][InputFieldTypeNone]
			if !ok {
				return "", errors.New("group template not found for field type: " + string(field.InputType))
			}

			gtpl, err := tmpl.Clone()
			if err != nil {
				return "", err
			}

			var sb strings.Builder
			gtpl = gtpl.Funcs(template.FuncMap{
				"fields": func() template.HTML {
					var subhtml template.HTML

					for _, subField := range field.Fields {
						fMap := copyMap(data)
						fMap["Field"] = field

						fieldhtml, err := f.formFieldHTML(subField, fieldErrors, fMap)
						if err != nil {
							continue
						}

						subhtml = subhtml + fieldhtml
					}

					return subhtml
				},
				"errors": func() []string {
					if errs, ok := fieldErrors[field.Name]; ok {
						return errs
					}
					return nil
				},
			})

			fMap := copyMap(data)
			fMap["Field"] = field

			err = gtpl.Execute(&sb, fMap)
			if err != nil {
				return "", err
			}
			html = html + template.HTML(sb.String())

			continue
		}

		fieldHTML, err := f.formFieldHTML(field, fieldErrors, data)
		if err != nil {
			return "", err
		}

		html = html + fieldHTML
	}
	return html, nil
}

func (f *Form) formFieldHTML(field FormField, errorMap map[string][]string, data map[string]any) (template.HTML, error) {
	tmp, ok := f.templateMap[field.Type][field.InputType]
	if !ok {
		return "", errors.New("template not found for field type: " + string(field.Type) + " and input type: " + string(field.InputType))
	}

	tpl, err := tmp.Clone()
	if err != nil {
		return "", err
	}

	fMap := copyMap(data)
	fMap["Field"] = field

	// generate label for the field
	labelTmp, ok := f.templateMap[FieldTypeLabel][InputFieldTypeNone]
	if !ok {
		return "", errors.New("label template not found for field type: " + string(FieldTypeLabel))
	}

	labelTpl, err := labelTmp.Clone()
	if err != nil {
		return "", err
	}

	var labelSb strings.Builder
	err = labelTpl.Execute(&labelSb, fMap)
	if err != nil {
		return "", err
	}

	var fieldSb strings.Builder
	tpl = tpl.Funcs(template.FuncMap{
		"label": func() template.HTML {
			return template.HTML(labelSb.String())
		},
		"errors": func() []string {
			if errs, ok := errorMap[field.Name]; ok {
				return errs
			}
			return nil
		},
	})

	err = tpl.Execute(&fieldSb, fMap)
	if err != nil {
		return "", err
	}

	// Skip wrapper for hidden fields
	if field.InputType == InputFieldTypeHidden {
		return template.HTML(fieldSb.String()), nil
	}

	// Check if we have a wrapper template
	if wrapperTmp, ok := f.templateMap[FieldTypeWrapper][InputFieldTypeNone]; ok {
		wrapperTpl, err := wrapperTmp.Clone()
		if err != nil {
			return "", err
		}

		var wrapperSb strings.Builder
		wrapperTpl = wrapperTpl.Funcs(template.FuncMap{
			"field": func() template.HTML {
				return template.HTML(fieldSb.String())
			},
			"label": func() template.HTML {
				return template.HTML(labelSb.String())
			},
			"errors": func() []string {
				if errs, ok := errorMap[field.Name]; ok {
					return errs
				}
				return nil
			},
		})

		err = wrapperTpl.Execute(&wrapperSb, fMap)
		if err != nil {
			return "", err
		}
		return template.HTML(wrapperSb.String()), nil
	}

	return template.HTML(fieldSb.String()), nil
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

func copyMap[K, V comparable](m map[K]V) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		result[k] = v
	}
	return result
}

package form

import (
	"errors"
	"fmt"
	"html/template"
	"reflect"
	"strings"

	"github.com/donseba/go-form/types"
)

type (
	Localizer interface {
		// GetLocale returns the locale of the localizer, ie. "en_US"
		GetLocale() string
	}

	DefaultLocalizer struct{}

	TranslationFunc func(loc Localizer, key string, args ...any) string
	ValidationFunc  func(fieldValue any, fieldStruct reflect.StructField) FieldErrors

	FieldErrors []FieldError

	FieldError interface {
		FieldError() (field, err string)
	}

	Info struct {
		Target     string            `json:"target,omitempty"`
		Method     string            `json:"method,omitempty"`
		SubmitText string            `json:"submit,omitempty"`
		Attributes map[string]string `json:"attributes,omitempty"`
	}

	Form struct {
		templateMap        map[types.FieldType]map[types.InputFieldType]template.Template
		validators         map[string]ValidationFunc
		translationEnabled bool
		translationFunc    TranslationFunc
	}
)

func (d *DefaultLocalizer) GetLocale() string {
	return "en" // Default locale
}

func NewTranslatedForm(templateMap map[types.FieldType]map[types.InputFieldType]string, translationFunc TranslationFunc) *Form {
	f := &Form{
		templateMap: make(map[types.FieldType]map[types.InputFieldType]template.Template),
		validators:  make(map[string]ValidationFunc),
	}

	f.enableTranslation(translationFunc)

	return f.init(templateMap)
}

func NewForm(templateMap map[types.FieldType]map[types.InputFieldType]string) *Form {
	f := &Form{
		templateMap: make(map[types.FieldType]map[types.InputFieldType]template.Template),
		validators:  make(map[string]ValidationFunc),
	}

	return f.init(templateMap)
}

func (f *Form) init(templateMap map[types.FieldType]map[types.InputFieldType]string) *Form {
	// First, create the base input template
	baseInputTpl, err := template.New("baseInput").Funcs(map[string]any{
		"form_print": func(loc Localizer, key string, args ...any) string { return "" },
	}).Parse(templateMap[types.FieldTypeBase][types.InputFieldTypeNone])
	if err != nil {
		panic(fmt.Errorf("error parsing base input template: %w", err))
	}

	for fieldType, inputTemplates := range templateMap {
		f.templateMap[fieldType] = make(map[types.InputFieldType]template.Template)
		for inputType, tpl := range inputTemplates {
			// Skip the base template as it's already created
			if fieldType == types.FieldTypeInput && inputType == types.InputFieldTypeNone {
				f.templateMap[fieldType][inputType] = *baseInputTpl
				continue
			}

			// Create a new template with the base template defined
			t := template.New(inputType.String())
			t, tErr := t.Funcs(map[string]any{
				"errors": func() []string { return nil },     // Placeholder for error handling
				"field":  func() template.HTML { return "" }, // Placeholder for field rendering
				"fields": func() template.HTML { return "" }, // Placeholder for group fields
				"label":  func() template.HTML { return "" }, // Placeholder for label rendering
				"form_print": func(loc Localizer, key string, args ...any) string {
					if f.translationEnabled && f.translationFunc != nil {
						return f.translationFunc(loc, key, args...)
					}

					if len(args) > 0 {
						return fmt.Sprintf(key, args...)
					}

					return key
				},
				"form_attributes": func(attributes map[string]string) template.HTMLAttr {
					var sb strings.Builder
					for k, v := range attributes {
						if v != "" {
							sb.WriteString(fmt.Sprintf(` %s="%s"`, k, template.HTMLEscapeString(v)))
						}
					}
					return template.HTMLAttr(sb.String())
				},
				"baseInput": func(kv ...any) template.HTML {
					if baseInputTpl == nil {
						return template.HTML("base input template not defined")
					}

					if len(kv)%2 != 0 {
						return template.HTML("need an even number of arguments for dict")
					}

					data := make(map[string]any)
					for i := 0; i < len(kv); i += 2 {
						key, ok := kv[i].(string)
						if !ok {
							return template.HTML("dict keys must be strings")
						}
						data[key] = kv[i+1]
					}

					var sb strings.Builder
					eErr := baseInputTpl.Execute(&sb, data)
					if eErr != nil {
						return template.HTML(fmt.Sprintf("error executing base input template: %+v", eErr))
					}
					return template.HTML(sb.String())
				},
			}).Parse(tpl)
			if tErr != nil {
				panic(fmt.Errorf("error parsing template for field type %s and input type %s: %w", fieldType, inputType, tErr))
			}
			f.templateMap[fieldType][inputType] = *t
		}
	}

	return f
}

// EnableTranslation enables translation support for the form package.
func (f *Form) enableTranslation(fn TranslationFunc) {
	f.translationEnabled = true
	f.translationFunc = fn
}

func (f *Form) FuncMap() template.FuncMap {
	funcMap := template.FuncMap{
		"form_render":           f.formRender,
		"form_render_localized": f.formRenderLocalized,
	}

	return funcMap
}

func (f *Form) formRender(v any, errs FieldErrors, kv ...any) (template.HTML, error) {
	return f.formRenderFunc(&DefaultLocalizer{}, v, errs, kv...)
}

func (f *Form) formRenderLocalized(loc Localizer, v any, errs FieldErrors, kv ...any) (template.HTML, error) {
	return f.formRenderFunc(loc, v, errs, kv...)
}

func (f *Form) formRenderFunc(loc Localizer, v any, errs FieldErrors, kv ...any) (template.HTML, error) {
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

	var formField *types.FormField
	for i, field := range tr.Fields {
		if field.Type == types.FieldTypeForm {
			formField = &tr.Fields[i]

			continue // Skip the form field itself, as it's handled separately
		}

		if field.Type == types.FieldTypeGroup {
			tmpl, ok := f.templateMap[types.FieldTypeGroup][types.InputFieldTypeNone]
			if !ok {
				return "", errors.New("group template not found for field type: " + string(field.InputType))
			}

			gtpl, err := tmpl.Clone()
			if err != nil {
				return "", err
			}

			var sb strings.Builder

			funcMap := template.FuncMap{
				"fields": func() template.HTML {
					var subHTML template.HTML

					for _, subField := range field.Fields {
						fMap := copyMap(data)
						fMap["Field"] = field

						fieldHTML, err := f.formFieldHTML(loc, subField, fieldErrors, fMap)
						if err != nil {
							continue
						}

						subHTML = subHTML + fieldHTML
					}

					return subHTML
				},
				"errors": func() []string {
					if errs, ok := fieldErrors[field.Name]; ok {
						return errs
					}
					return nil
				},
			}

			gtpl = gtpl.Funcs(funcMap)

			fMap := copyMap(data)
			fMap["Field"] = field
			fMap["Loc"] = loc

			err = gtpl.Execute(&sb, fMap)
			if err != nil {
				return "", fmt.Errorf("error executing group template: %w", err)
			}
			html = html + template.HTML(sb.String())

			continue
		}

		fieldHTML, err := f.formFieldHTML(loc, field, fieldErrors, data)
		if err != nil {
			return "", fmt.Errorf("error generating field HTML for field type %s with inputType %s, : %w", field.Type, field.InputType, err)
		}
		html = html + fieldHTML
	}

	if formField != nil {
		tmpl, ok := f.templateMap[types.FieldTypeForm][types.InputFieldTypeNone]
		if !ok {
			return "", errors.New("form template not found for field type: " + string(types.FieldTypeForm))
		}

		formTmpl, err := tmpl.Clone()
		if err != nil {
			return "", err
		}

		var sb strings.Builder
		formTmpl = formTmpl.Funcs(template.FuncMap{
			"fields": func() template.HTML {
				return html
			},
		})

		formData := struct {
			Field types.FormField
			Loc   Localizer
		}{
			Field: *formField,
			Loc:   loc,
		}

		err = formTmpl.Execute(&sb, formData)
		if err != nil {
			return "", err
		}

		return template.HTML(sb.String()), nil
	}

	return html, nil
}

func (f *Form) formFieldHTML(loc Localizer, field types.FormField, errorMap map[string][]string, data map[string]any) (template.HTML, error) {
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
	fMap["Loc"] = loc

	// generate label for the field
	labelTmp, ok := f.templateMap[types.FieldTypeLabel][types.InputFieldTypeNone]
	if !ok {
		return "", errors.New("label template not found for field type: " + string(types.FieldTypeLabel))
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
	if field.InputType == types.InputFieldTypeHidden {
		return template.HTML(fieldSb.String()), nil
	}

	// Check if we have a wrapper template
	if wrapperTmp, ok := f.templateMap[types.FieldTypeWrapper][types.InputFieldTypeNone]; ok {
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

func copyMap[K, V comparable](m map[K]V) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		result[k] = v
	}
	return result
}

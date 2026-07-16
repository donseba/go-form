package form

import (
	"fmt"
	"html/template"
	"reflect"

	"github.com/donseba/go-form/v2/templates"
	"github.com/donseba/go-form/v2/types"
)

// getTheme loads embedded themes (once) and returns the selected theme.
func (f *Form) getTheme() (*templates.Theme, error) {
	if err := ensureThemesLoaded(); err != nil {
		return nil, err
	}
	if f.themeName == "" {
		return nil, fmt.Errorf("no theme selected")
	}
	theme, ok := templates.GetTheme(f.themeName)
	if !ok || theme == nil || theme.Templates == nil {
		return nil, fmt.Errorf("unknown or unloaded theme %q", f.themeName)
	}
	return theme, nil
}

// csrfHiddenInputHTML extracts CSRF settings from the model and returns a hidden input.
func csrfHiddenInputHTML(v any) template.HTML {
	switch wrapped := v.(type) {
	case RenderModel:
		return csrfHiddenInputHTMLFromInfo(wrapped.Info)
	case *RenderModel:
		if wrapped != nil {
			return csrfHiddenInputHTMLFromInfo(wrapped.Info)
		}
	}

	// Prefer explicit GetFormInfo.
	if fm, ok := v.(interface{ GetFormInfo() Info }); ok {
		info := fm.GetFormInfo()
		return csrfHiddenInputHTMLFromInfo(info)
	}

	rval := reflect.ValueOf(v)
	if rval.Kind() == reflect.Ptr || rval.Kind() == reflect.Interface {
		if !rval.IsNil() {
			rval = rval.Elem()
		}
	}
	if rval.IsValid() && rval.Kind() == reflect.Struct && rval.NumField() > 0 {
		firstField := rval.Field(0)
		if firstField.IsValid() && firstField.Type() == reflect.TypeOf(Info{}) {
			return csrfHiddenInputHTMLFromInfo(firstField.Interface().(Info))
		}
	}

	return ""
}

func csrfHiddenInputHTMLFromInfo(info Info) template.HTML {
	if info.CsrfValue == "" {
		return ""
	}

	csrfFieldName := info.CsrfField
	if csrfFieldName == "" {
		csrfFieldName = DefaultCSRFField
	}

	return template.HTML(fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`,
		template.HTMLEscapeString(csrfFieldName),
		template.HTMLEscapeString(info.CsrfValue)))
}

// renderGroupFields renders all fields inside a group by delegating to themeFieldHTML.
// It ignores per-field render errors (legacy behavior) to avoid nuking the whole form.
func (f *Form) renderGroupFields(theme *templates.Theme, loc types.Localizer, fields []types.FormField, fieldErrors map[string][]string) template.HTML {
	var sub template.HTML
	for _, subField := range fields {
		fh, err := f.themeFieldHTML(theme, loc, subField, fieldErrors)
		if err != nil {
			continue
		}
		sub += fh
	}
	return sub
}

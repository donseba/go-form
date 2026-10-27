package form

import (
	"fmt"
	"html/template"
	"maps"
	"reflect"
	"slices"
	"strings"

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

// modelInfo returns the form metadata of a model: the Info of a
// RenderModel, of GetFormInfo, or of an Info embedded as the first field.
func modelInfo(v any) (Info, bool) {
	if wrapped, ok := renderModelOf(v); ok {
		return wrapped.Info, true
	}

	// Prefer explicit GetFormInfo.
	if fm, ok := v.(interface{ GetFormInfo() Info }); ok {
		return fm.GetFormInfo(), true
	}

	rval := reflect.ValueOf(v)
	if rval.Kind() == reflect.Ptr || rval.Kind() == reflect.Interface {
		if !rval.IsNil() {
			rval = rval.Elem()
		}
	}
	if rval.IsValid() && rval.Kind() == reflect.Struct && rval.NumField() > 0 {
		firstField := rval.Field(0)
		if firstField.IsValid() && firstField.Type() == reflect.TypeOf(Info{}) && firstField.CanInterface() {
			return firstField.Interface().(Info), true
		}
	}

	return Info{}, false
}

// csrfHiddenInputHTML extracts CSRF settings from the model and returns a hidden input.
func csrfHiddenInputHTML(v any) template.HTML {
	info, ok := modelInfo(v)
	if !ok {
		return ""
	}

	return csrfHiddenInputHTMLFromInfo(info)
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

// hiddenInputsHTML renders the hidden inputs of Info.Hidden: one per value,
// in name order.
func hiddenInputsHTML(info Info) template.HTML {
	if len(info.Hidden) == 0 {
		return ""
	}

	var out strings.Builder

	for _, name := range slices.Sorted(maps.Keys(info.Hidden)) {
		for _, value := range info.Hidden[name] {
			out.WriteString(`<input type="hidden" name="`)
			out.WriteString(template.HTMLEscapeString(name))
			out.WriteString(`" value="`)
			out.WriteString(template.HTMLEscapeString(value))
			out.WriteString(`">`)
		}
	}

	//nolint:gosec // Fixed markup around escaped names and values.
	return template.HTML(out.String())
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

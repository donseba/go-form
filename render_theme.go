package form

import (
	"html/template"
	"strings"

	"github.com/donseba/go-form/templates"
	"github.com/donseba/go-form/types"
)

// themeExec clones a theme template set, injects standard form functions, and executes one template.
//
// Contract:
// - It always injects form_print wired to this Form's translation function.
// - It can inject optional overrides (fields/errors/label/field etc) via funcs.
// - It returns template.HTML and preserves template execution errors.
func (f *Form) themeExec(theme *templates.Theme, tmplName string, data any, funcs template.FuncMap) (template.HTML, error) {
	cl, err := theme.Templates.Clone()
	if err != nil {
		return "", err
	}

	// Always inject the translated form_print.
	cl = cl.Funcs(template.FuncMap{
		"form_print": func(loc types.Localizer, key string, args ...any) string { return f.themePrint(loc, key, args...) },
	})

	// Optional overrides (fields/errors/etc.).
	if len(funcs) > 0 {
		cl = cl.Funcs(funcs)
	}

	var out strings.Builder
	if err := cl.ExecuteTemplate(&out, tmplName, data); err != nil {
		return "", err
	}
	return template.HTML(out.String()), nil
}

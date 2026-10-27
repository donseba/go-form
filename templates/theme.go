package templates

import (
	"fmt"
	"html/template"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/donseba/go-form/v2/types"
)

// StyleOption represents either a CSS class or an inline style
type StyleOption struct {
	Class string // CSS class name for class-based themes (Bootstrap, Tailwind)
	Style string // Inline CSS style for style-based themes (Plain)
}

// ThemeClasses represents all CSS classes used in a theme
type ThemeClasses struct {
	// Common UI elements
	Wrapper StyleOption
	Label   StyleOption
	// Fieldset and Legend are added to Wrapper and Label for a group of
	// radio buttons or a multicheckbox, which renders as a fieldset with
	// its label as the legend (see fieldset.gohtml and legend.gohtml), such
	// as a reset of the browser's fieldset border and padding.
	Fieldset    StyleOption
	Legend      StyleOption
	Required    StyleOption
	Error       StyleOption
	Description StyleOption

	// Input elements
	Input           StyleOption
	Select          StyleOption
	Textarea        StyleOption
	TextareaWrapper StyleOption
	Checkbox        StyleOption
	CheckboxWrapper StyleOption
	CheckboxLabel   StyleOption
	Radio           StyleOption
	RadioWrapper    StyleOption
	RadioLabel      StyleOption
	RadioGroup      StyleOption
	Range           StyleOption
	RangeWrapper    StyleOption
	RangeValue      StyleOption
	Color           StyleOption
	Button          StyleOption
	File            StyleOption
	Multicheckbox   StyleOption

	// Form containers
	Form        StyleOption
	FormGroup   StyleOption
	FormHeader  StyleOption
	FormLegend  StyleOption
	FormBody    StyleOption
	FormButtons StyleOption
	Cancel      StyleOption

	// Input groups
	InputGroup     StyleOption
	InputGroupText StyleOption

	// Honeypot hides the spam trap input (Info.Honeypot) from people, such
	// as a class that positions it off-screen. A theme with a class gets
	// that class alone, so it renders no style attribute, which a
	// Content-Security-Policy without 'unsafe-inline' would ignore. Without
	// a class, the input gets HoneypotClass and the style of Honeypot, or
	// HoneypotStyle, so it is hidden without any CSS. Keep it a normal text
	// input, not display:none or the hidden attribute, which bots skip.
	Honeypot StyleOption

	// VisuallyHidden hides text from sight but not from screen readers, such
	// as the "(required)" after the label of a required field. A theme with
	// a class gets that class alone, such as Bootstrap's visually-hidden or
	// Tailwind's sr-only, so it renders no style attribute. Without a class,
	// the text gets VisuallyHiddenClass, which a site's CSS hides, and the
	// style of VisuallyHidden, such as the plain theme's VisuallyHiddenStyle.
	VisuallyHidden StyleOption
}

// VisuallyHiddenClass is the class of text for screen readers only in a
// theme whose VisuallyHidden has no class of its own: a site's CSS hides
// it, such as
//
//	.sr-only { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0, 0, 0, 0); white-space: nowrap; }
const VisuallyHiddenClass = "sr-only"

// VisuallyHiddenStyle is the inline style that hides text for screen
// readers only without any CSS, as the plain theme does.
const VisuallyHiddenStyle = "position:absolute;width:1px;height:1px;padding:0;margin:-1px;overflow:hidden;clip:rect(0,0,0,0);white-space:nowrap;border:0"

// visuallyHidden is the StyleOption of text for screen readers only: the
// theme's class alone, or VisuallyHiddenClass with the theme's style.
func (c ThemeClasses) visuallyHidden() StyleOption {
	if c.VisuallyHidden.Class != "" {
		return StyleOption{Class: c.VisuallyHidden.Class}
	}

	return StyleOption{Class: VisuallyHiddenClass, Style: c.VisuallyHidden.Style}
}

// HoneypotClass is the class of the spam trap input of a theme whose
// Honeypot has no class of its own: a site's CSS can hide it, such as
//
//	.form-honeypot { position: absolute; left: -10000px; width: 1px; height: 1px; overflow: hidden; }
const HoneypotClass = "form-honeypot"

// HoneypotStyle is the inline style that hides the spam trap input of a
// theme whose Honeypot has neither a class nor a style.
const HoneypotStyle = "position:absolute;left:-10000px;width:1px;height:1px;overflow:hidden"

// honeypot is the StyleOption of the spam trap input: the theme's class
// alone, or HoneypotClass with the theme's style or HoneypotStyle.
func (c ThemeClasses) honeypot() StyleOption {
	if c.Honeypot.Class != "" {
		return StyleOption{Class: c.Honeypot.Class}
	}

	option := StyleOption{Class: HoneypotClass, Style: c.Honeypot.Style}
	if option.Style == "" {
		option.Style = HoneypotStyle
	}

	return option
}

// Theme represents a form theme with CSS classes and attributes
type Theme struct {
	Name      string
	Classes   ThemeClasses
	AttrMap   map[string]string
	Templates *template.Template
}

// themeCache stores precompiled templates for themes
var themeCache = struct {
	sync.RWMutex
	themes map[string]*Theme
}{
	themes: make(map[string]*Theme),
}

// RegisterTheme registers a new theme with the given name and classes
func RegisterTheme(name string, classes ThemeClasses, attrMap map[string]string) *Theme {
	themeCache.Lock()
	defer themeCache.Unlock()

	theme := &Theme{
		Name:    name,
		Classes: classes,
		AttrMap: attrMap,
	}

	themeCache.themes[name] = theme
	return theme
}

// RegisterClassTheme registers a class-based theme with go-form's own
// templates, ready to render, such as a theme with the CSS classes of a site:
//
//	theme, err := templates.RegisterClassTheme("site", templates.ThemeClasses{
//		Input: templates.StyleOption{Class: "control"},
//	})
//
// Registering the same name with the same classes again returns the theme as
// it is, so callers need no sync.Once. Other classes replace the theme. Register themes before serving requests,
// and use a name of your own: the built-in names are registered again when
// the first form renders.
func RegisterClassTheme(name string, classes ThemeClasses) (*Theme, error) {
	themeCache.Lock()
	defer themeCache.Unlock()

	if theme, ok := themeCache.themes[name]; ok && theme.Classes == classes && theme.Templates != nil {
		return theme, nil
	}

	theme := &Theme{Name: name, Classes: classes}
	if err := theme.LoadTemplatesFS(TemplateFS, "gohtml"); err != nil {
		return nil, err
	}

	themeCache.themes[name] = theme

	return theme, nil
}

// GetTheme returns a theme by name
func GetTheme(name string) (*Theme, bool) {
	themeCache.RLock()
	defer themeCache.RUnlock()

	theme, found := themeCache.themes[name]
	return theme, found
}

// LoadTemplates loads all .gohtml templates from the given directory and associates them with the theme
func (t *Theme) LoadTemplates(templateDir string) error {
	tmpl := template.New("").Funcs(t.funcs())

	// Walk the template directory
	err := filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process .gohtml files
		if filepath.Ext(path) != ".gohtml" {
			return nil
		}

		// Read the template file
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Get template name (without extension)
		name := filepath.Base(path)
		name = name[:len(name)-len(filepath.Ext(name))]

		// Parse the template
		_, err = tmpl.New(name).Parse(string(content))
		return err
	})

	if err != nil {
		return err
	}

	t.Templates = tmpl
	return nil
}

// LoadTemplatesFS loads all .gohtml templates from the given embedded filesystem
func (t *Theme) LoadTemplatesFS(fsys fs.FS, rootDir string) error {
	// Always start from a fresh template set. This lets InitThemes be called
	// multiple times in tests and ensures template content is not stale.
	tmpl := template.New("").Funcs(t.funcs())

	err := fs.WalkDir(fsys, rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || filepath.Ext(path) != ".gohtml" {
			return nil
		}

		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		_, err = tmpl.New(name).Parse(string(content))

		return err
	})
	if err != nil {
		return err
	}

	t.Templates = tmpl

	return nil
}

// funcs are the helper functions of the theme's templates.
func (t *Theme) funcs() template.FuncMap {
	return template.FuncMap{
		// default returns the fallback value when the piped value is the zero value.
		// Usage: {{ .Something | default "fallback" }}
		"default": funcDefault,
		"themeClass": func(key string) string {
			// For class-based themes, return the class; for inline style themes, return empty
			option := t.getStyleOptionForKey(key)
			return option.Class
		},
		"themeStyle": func(key string) template.CSS {
			option := t.getStyleOptionForKey(key)
			return template.CSS(option.Style)
		},
		"themeAttr": func(key string) string {
			if attr, ok := t.AttrMap[key]; ok {
				return attr
			}
			return ""
		},
		// Placeholders for templates that expect callable blocks/functions.
		// Callers can override these via template clone + Funcs() before executing.
		"fields": func() template.HTML { return "" },
		"field":  func() template.HTML { return "" },
		"label":  func() template.HTML { return "" },
		"errors": func() []string { return nil },
		// The classes, inline style and other attributes the form gets
		// besides the theme's, from Info.Attributes and form_open.
		"form_open_class":      func() string { return "" },
		"form_open_style":      func() template.CSS { return "" },
		"form_open_attributes": func() template.HTMLAttr { return "" },
		"form_print":           funcPrint,
		"form_required_label":  funcRequiredLabel,
		"form_attributes":      funcAttributes,
		"form_data_attributes": funcDataAttributes,
		"form_described_by":    DescribedBy,
		"form_class":           ClassList,
		"form_style":           StyleList,
	}
}

// ClassList joins the non-empty classes with a space, such as a theme's
// class and a field's class tag; empty when there are none, so templates
// leave the class attribute out:
//
//	{{with form_class (themeClass "input") .Field.Class}}class="{{.}}"{{end}}
//
// Templates call it as form_class.
func ClassList(classes ...string) string {
	var parts []string

	for _, class := range classes {
		if class = strings.TrimSpace(class); class != "" {
			parts = append(parts, class)
		}
	}

	return strings.Join(parts, " ")
}

// StyleList joins the non-empty inline styles with "; ", such as a theme's
// style and the style a template gives form_open; empty when there are
// none, so templates leave the style attribute out. A single style keeps
// its closing semicolon:
//
//	{{with form_style (themeStyle "form") .Style}}style="{{.}}"{{end}}
//
// Templates call it as form_style.
func StyleList(styles ...template.CSS) template.CSS {
	var parts []string

	for _, style := range styles {
		if trimmed := strings.TrimSpace(string(style)); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}

	if len(parts) == 1 {
		return template.CSS(parts[0])
	}

	for i := range parts {
		parts[i] = strings.TrimRight(parts[i], "; ")
	}

	//nolint:gosec // The styles are the theme's and those the application gave.
	return template.CSS(strings.Join(parts, "; "))
}

// DescribedBy is the aria-describedby value of a field's control: the IDs of
// its description and of its error messages, as the wrapper template renders
// them ("<id>_description", "<id>_error_0", …). Templates call it as
// form_described_by.
func DescribedBy(field types.FormField) string {
	var ids []string
	if field.Description != "" {
		ids = append(ids, field.Id+"_description")
	}

	for i := range field.Errors {
		ids = append(ids, field.Id+"_error_"+strconv.Itoa(i))
	}

	return strings.Join(ids, " ")
}

// funcDefault returns fallback when v is the zero value of its type.
func funcDefault(v any, fallback any) any {
	if v == nil {
		return fallback
	}

	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return fallback
	}

	// Handle pointers & interfaces
	for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return fallback
		}

		rv = rv.Elem()
	}

	if !rv.IsValid() {
		return fallback
	}

	// Zero checks for common kinds
	switch rv.Kind() {
	case reflect.String:
		if rv.Len() == 0 {
			return fallback
		}
	case reflect.Bool:
		if !rv.Bool() {
			return fallback
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if rv.Int() == 0 {
			return fallback
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if rv.Uint() == 0 {
			return fallback
		}
	case reflect.Float32, reflect.Float64:
		if rv.Float() == 0 {
			return fallback
		}
	case reflect.Slice, reflect.Map, reflect.Array:
		if rv.Len() == 0 {
			return fallback
		}
	default:
		// For other kinds, keep the original value.
	}

	return v
}

// getStyleOptionForKey returns the StyleOption for a given key
func (t *Theme) getStyleOptionForKey(key string) StyleOption {
	// Convert dashed keys to camel case (e.g., "checkbox-wrapper" to "CheckboxWrapper")
	parts := strings.Split(key, "-")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	fieldName := parts[0]
	for i := 1; i < len(parts); i++ {
		fieldName += parts[i]
	}

	// Match the field name to the corresponding StyleOption
	switch fieldName {
	case "wrapper":
		return t.Classes.Wrapper
	case "label":
		return t.Classes.Label
	case "fieldset":
		return t.Classes.Fieldset
	case "legend":
		return t.Classes.Legend
	case "required":
		return t.Classes.Required
	case "error":
		return t.Classes.Error
	case "description":
		return t.Classes.Description
	case "input":
		return t.Classes.Input
	case "select":
		return t.Classes.Select
	case "textarea":
		return t.Classes.Textarea
	case "textareaWrapper":
		return t.Classes.TextareaWrapper
	case "checkbox":
		return t.Classes.Checkbox
	case "checkboxWrapper":
		return t.Classes.CheckboxWrapper
	case "checkboxLabel":
		return t.Classes.CheckboxLabel
	case "radio":
		return t.Classes.Radio
	case "radioWrapper":
		return t.Classes.RadioWrapper
	case "radioLabel":
		return t.Classes.RadioLabel
	case "radioGroup":
		return t.Classes.RadioGroup
	case "range":
		return t.Classes.Range
	case "rangeWrapper":
		return t.Classes.RangeWrapper
	case "rangeValue":
		return t.Classes.RangeValue
	case "color":
		return t.Classes.Color
	case "button":
		return t.Classes.Button
	case "file":
		return t.Classes.File
	case "form":
		return t.Classes.Form
	case "formGroup":
		return t.Classes.FormGroup
	case "formHeader":
		return t.Classes.FormHeader
	case "formLegend":
		return t.Classes.FormLegend
	case "formBody":
		return t.Classes.FormBody
	case "formButtons":
		return t.Classes.FormButtons
	case "cancel":
		return t.Classes.Cancel
	case "inputGroup":
		return t.Classes.InputGroup
	case "inputGroupText":
		return t.Classes.InputGroupText
	case "multicheckbox":
		return t.Classes.Multicheckbox
	case "honeypot":
		return t.Classes.honeypot()
	case "visuallyHidden":
		return t.Classes.visuallyHidden()
	default:
		return StyleOption{}
	}
}

// RenderTemplate renders a template with the given name and data
func (t *Theme) RenderTemplate(name string, data interface{}) (template.HTML, error) {
	var buf strings.Builder
	err := t.Templates.ExecuteTemplate(&buf, name, data)
	if err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

// Define existing function references
var (
	// funcRequiredLabel is replaced by the renderer with the form's
	// translation of TranslationKeyRequiredLabel.
	funcRequiredLabel = func(loc types.Localizer) string {
		_ = loc
		return "(required)"
	}
	funcPrint = func(loc types.Localizer, key string, args ...any) string {
		_ = loc
		if len(args) > 0 {
			return fmt.Sprintf(key, args...)
		}
		return key
	}
	funcAttributes = func(attributes map[string]string) template.HTMLAttr {
		var sb strings.Builder
		// Sorted, so the same form renders the same markup.
		for _, k := range slices.Sorted(maps.Keys(attributes)) {
			v := attributes[k]
			if v != "" {
				sb.WriteString(" ")
				sb.WriteString(k)
				sb.WriteString("=\"")
				sb.WriteString(template.HTMLEscapeString(v))
				sb.WriteString("\"")
			}
		}
		return template.HTMLAttr(sb.String())
	}
	funcDataAttributes = func(data map[string]string) template.HTMLAttr {
		var sb strings.Builder

		for _, k := range slices.Sorted(maps.Keys(data)) {
			v := data[k]
			if v != "" {
				sb.WriteString(" data-")
				sb.WriteString(k)
				sb.WriteString("=\"")
				sb.WriteString(template.HTMLEscapeString(v))
				sb.WriteString("\"")
			}
		}
		return template.HTMLAttr(sb.String())
	}
)

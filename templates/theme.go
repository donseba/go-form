package templates

import (
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/donseba/go-form/types"
)

// StyleOption represents either a CSS class or an inline style
type StyleOption struct {
	Class string // CSS class name for class-based themes (Bootstrap, Tailwind)
	Style string // Inline CSS style for style-based themes (Plain)
}

// ThemeClasses represents all CSS classes used in a theme
type ThemeClasses struct {
	// Common UI elements
	Wrapper     StyleOption
	Label       StyleOption
	Error       StyleOption
	Description StyleOption

	// Input elements
	Input           StyleOption
	Select          StyleOption
	Textarea        StyleOption
	Checkbox        StyleOption
	CheckboxWrapper StyleOption
	CheckboxLabel   StyleOption
	Radio           StyleOption
	RadioWrapper    StyleOption
	RadioLabel      StyleOption
	Range           StyleOption
	Color           StyleOption
	Button          StyleOption
	File            StyleOption

	// Form containers
	Form        StyleOption
	FormGroup   StyleOption
	FormHeader  StyleOption
	FormBody    StyleOption
	FormButtons StyleOption

	// Input groups
	InputGroup     StyleOption
	InputGroupText StyleOption
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

// GetTheme returns a theme by name
func GetTheme(name string) (*Theme, bool) {
	themeCache.RLock()
	defer themeCache.RUnlock()

	theme, found := themeCache.themes[name]
	return theme, found
}

// LoadTemplates loads all .gohtml templates from the given directory and associates them with the theme
func (t *Theme) LoadTemplates(templateDir string) error {
	tmpl := template.New("")

	// Register helper functions
	tmpl.Funcs(template.FuncMap{
		"themeClass": func(key string) string {
			option := t.getStyleOptionForKey(key)
			return option.Class
		},
		"themeStyle": func(key string) string {
			option := t.getStyleOptionForKey(key)
			return option.Style
		},
		"themeAttr": func(key string) string {
			if attr, ok := t.AttrMap[key]; ok {
				return attr
			}
			return ""
		},
	})

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
	tmpl := template.New("")

	// Register helper functions
	tmpl.Funcs(template.FuncMap{
		"themeClass": func(key string) string {
			// For class-based themes, return the class; for inline style themes, return empty
			option := t.getStyleOptionForKey(key)
			return option.Class
		},
		"themeStyle": func(key string) string {
			// For inline style themes, return the style; for class-based themes, return empty
			option := t.getStyleOptionForKey(key)
			return option.Style
		},
		"themeAttr": func(key string) string {
			if attr, ok := t.AttrMap[key]; ok {
				return attr
			}
			return ""
		},
		"form_print":           funcPrint,
		"form_attributes":      funcAttributes,
		"form_data_attributes": funcDataAttributes,
	})

	// Walk the embedded filesystem
	err := fs.WalkDir(fsys, rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process .gohtml files
		if filepath.Ext(path) != ".gohtml" {
			return nil
		}

		// Read the template file
		content, err := fs.ReadFile(fsys, path)
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
	case "range":
		return t.Classes.Range
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
	case "formBody":
		return t.Classes.FormBody
	case "formButtons":
		return t.Classes.FormButtons
	case "inputGroup":
		return t.Classes.InputGroup
	case "inputGroupText":
		return t.Classes.InputGroupText
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
	funcPrint          = func(loc types.Localizer, key string) string { return "" }     // Placeholder
	funcAttributes     = func(attributes map[string]string) template.HTML { return "" } // Placeholder
	funcDataAttributes = func(data map[string]string) template.HTML { return "" }       // Placeholder
)

// SetFuncPrint sets the print function used in templates
func SetFuncPrint(fn func(loc types.Localizer, key string) string) {
	funcPrint = fn
}

// SetFuncAttributes sets the attributes function used in templates
func SetFuncAttributes(fn func(attributes map[string]string) template.HTML) {
	funcAttributes = fn
}

// SetFuncDataAttributes sets the data-attributes function used in templates
func SetFuncDataAttributes(fn func(data map[string]string) template.HTML) {
	funcDataAttributes = fn
}

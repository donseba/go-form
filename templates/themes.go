package templates

import (
	"embed"
)

//go:embed gohtml/*.gohtml
var TemplateFS embed.FS

// BootstrapTheme defines Bootstrap v5 classes for form elements
var BootstrapTheme = ThemeClasses{
	// Common elements
	Wrapper:     StyleOption{Class: "mb-2"},
	Label:       StyleOption{Class: "form-label small mb-1"},
	Error:       StyleOption{Class: "invalid-feedback d-block small"},
	Description: StyleOption{Class: "form-text small"},

	// Input types
	Input:           StyleOption{Class: "form-control"},
	Select:          StyleOption{Class: "form-select form-select-sm"},
	Textarea:        StyleOption{Class: "form-control form-control-sm"},
	Radio:           StyleOption{Class: "form-check-input"},
	RadioWrapper:    StyleOption{Class: "form-check form-check-inline"},
	RadioLabel:      StyleOption{Class: "form-check-label"},
	Checkbox:        StyleOption{Class: "form-check-input"},
	CheckboxWrapper: StyleOption{Class: "form-check"},
	CheckboxLabel:   StyleOption{Class: "form-check-label"},
	Range:           StyleOption{Class: "form-range"},
	Color:           StyleOption{Class: "form-control form-control-color"},
	Button:          StyleOption{Class: "btn btn-primary btn-sm"},
	File:            StyleOption{Class: "form-control form-control-sm"},

	// Form container
	Form:        StyleOption{Class: "mx-auto border rounded shadow-sm p-3"},
	FormGroup:   StyleOption{Class: "card card-sm mb-2"},
	FormHeader:  StyleOption{Class: "card-header py-1"},
	FormBody:    StyleOption{Class: "card-body py-2"},
	FormButtons: StyleOption{Class: "d-grid gap-2 mt-3"},

	// Input groups
	InputGroup:     StyleOption{Class: "input-group"},
	InputGroupText: StyleOption{Class: "input-group-text"},
}

// TailwindTheme defines Tailwind CSS v3 classes for form elements
var TailwindTheme = ThemeClasses{
	// Common elements
	Wrapper:     StyleOption{Class: "mb-2"},
	Label:       StyleOption{Class: "block text-sm font-medium leading-6 text-gray-900"},
	Error:       StyleOption{Class: "mt-1 text-sm text-red-600"},
	Description: StyleOption{Class: "mt-1 text-sm text-gray-500"},

	// Input types
	Input:           StyleOption{Class: "border border-gray-200 block w-full rounded-md px-3 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"},
	Select:          StyleOption{Class: "border border-gray-200 block w-full rounded-md px-3 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"},
	Textarea:        StyleOption{Class: "border border-gray-200 block w-full rounded-md px-3 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"},
	Radio:           StyleOption{Class: "h-4 w-4 border border-gray-200 text-indigo-600 focus:ring-indigo-600"},
	RadioWrapper:    StyleOption{Class: "inline-block mr-4"},
	RadioLabel:      StyleOption{Class: "ml-2 text-sm text-gray-900"},
	Checkbox:        StyleOption{Class: "h-4 w-4 rounded border border-gray-200 text-indigo-600 focus:ring-indigo-600"},
	CheckboxWrapper: StyleOption{Class: "inline-block"},
	CheckboxLabel:   StyleOption{Class: "ml-2 text-sm text-gray-900"},
	Range:           StyleOption{Class: "w-full h-2 rounded-lg appearance-none cursor-pointer bg-gray-200 border border-gray-200"},
	Color:           StyleOption{Class: "h-8 w-8 rounded border border-gray-200 p-0"},
	Button:          StyleOption{Class: "rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed"},
	File:            StyleOption{Class: "border border-gray-200 block w-full rounded-md px-3 py-1.5 text-gray-900 shadow-sm sm:text-sm"},

	// Form container
	Form:        StyleOption{Class: "mx-auto max-w-md rounded-lg border border-gray-200 bg-white p-4 shadow-sm"},
	FormGroup:   StyleOption{Class: "mb-2 rounded-lg border border-gray-200 bg-white"},
	FormHeader:  StyleOption{Class: "border-b border-gray-200 bg-gray-50 px-4 py-2"},
	FormBody:    StyleOption{Class: "p-4"},
	FormButtons: StyleOption{Class: "mt-4 flex justify-end"},

	// Input groups
	InputGroup:     StyleOption{Class: "flex rounded-md shadow-sm"},
	InputGroupText: StyleOption{Class: "inline-flex items-center rounded-l-md border border-r-0 border-gray-300 bg-gray-50 px-3 text-gray-500 text-sm"},
}

// PlainTheme defines simple, unstyled HTML with inline styles
var PlainTheme = ThemeClasses{
	// Common elements
	Wrapper:     StyleOption{Style: "margin-bottom: 0.5rem;"},
	Label:       StyleOption{Style: "display: block; margin-bottom: 0.25rem; font-size: 0.875rem; font-weight: 500; color: #212529;"},
	Error:       StyleOption{Style: "display: block; width: 100%; margin-top: 0.25rem; font-size: 0.75rem; color: #dc3545;"},
	Description: StyleOption{Style: "margin-top: 0.25rem; font-size: 0.75rem; color: #6c757d;"},

	// Input types
	Input:           StyleOption{Style: "width: 100%; padding: 0.375rem 0.75rem; font-size: 0.875rem; line-height: 1.5; color: #212529; background-color: #fff; border: 1px solid #ced4da; border-radius: 0.25rem; transition: border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;"},
	Select:          StyleOption{Style: "width: 100%; padding: 0.375rem 0.75rem; font-size: 0.875rem; line-height: 1.5; color: #212529; background-color: #fff; border: 1px solid #ced4da; border-radius: 0.25rem; transition: border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;"},
	Textarea:        StyleOption{Style: "width: 100%; padding: 0.375rem 0.75rem; font-size: 0.875rem; line-height: 1.5; color: #212529; background-color: #fff; border: 1px solid #ced4da; border-radius: 0.25rem; transition: border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;"},
	Radio:           StyleOption{Style: "width: 1rem; height: 1rem; margin-top: 0.25rem; vertical-align: top; background-color: #fff; border: 1px solid #ced4da; border-radius: 50%; cursor: pointer;"},
	RadioWrapper:    StyleOption{Style: "display: inline-block; margin-right: 1rem;"},
	RadioLabel:      StyleOption{Style: "margin-left: 0.25rem; font-size: 0.875rem; color: #212529;"},
	Checkbox:        StyleOption{Style: "width: 1rem; height: 1rem; margin-top: 0.25rem; vertical-align: top; background-color: #fff; border: 1px solid #ced4da; border-radius: 0.25rem; cursor: pointer;"},
	CheckboxWrapper: StyleOption{Style: "display: inline-block;"},
	CheckboxLabel:   StyleOption{Style: "margin-left: 0.25rem; font-size: 0.875rem; color: #212529;"},
	Range:           StyleOption{Style: "width: 100%; height: 0.5rem; border: 1px solid #ced4da; border-radius: 0.25rem;"},
	Color:           StyleOption{Style: "width: 2rem; height: 2rem; padding: 0; border: 1px solid #ced4da; border-radius: 0.25rem;"},
	Button:          StyleOption{Style: "display: inline-block; font-weight: 400; text-align: center; white-space: nowrap; vertical-align: middle; user-select: none; border: 1px solid transparent; padding: 0.375rem 0.75rem; font-size: 0.875rem; line-height: 1.5; border-radius: 0.25rem; color: #fff; background-color: #0d6efd; border-color: #0d6efd; cursor: pointer; transition: color 0.15s ease-in-out, background-color 0.15s ease-in-out, border-color 0.15s ease-in-out, box-shadow 0.15s ease-in-out;"},
	File:            StyleOption{Style: "width: 100%; padding: 0.375rem 0.75rem; font-size: 0.875rem; line-height: 1.5; color: #212529; background-color: #fff; border: 1px solid #ced4da; border-radius: 0.25rem;"},

	// Form container
	Form:        StyleOption{Style: "max-width: 32rem; margin: 0 auto; padding: 1rem; border: 1px solid #dee2e6; border-radius: 0.25rem; background-color: #fff; box-shadow: 0 0.125rem 0.25rem rgba(0, 0, 0, 0.075);"},
	FormGroup:   StyleOption{Style: "margin-bottom: 0.5rem; border: 1px solid #dee2e6; border-radius: 0.25rem; background-color: #fff;"},
	FormHeader:  StyleOption{Style: "padding: 0.5rem 1rem; border-bottom: 1px solid #dee2e6; background-color: #f8f9fa;"},
	FormBody:    StyleOption{Style: "padding: 0.5rem 1rem;"},
	FormButtons: StyleOption{Style: "margin-top: 1rem; text-align: right;"},

	// Input groups
	InputGroup:     StyleOption{Style: "display: flex; align-items: stretch; width: 100%;"},
	InputGroupText: StyleOption{Style: "display: inline-flex; align-items: center; padding: 0 0.75rem; background: #f8f9fa; border: 1px solid #ced4da; border-right: 0; border-radius: 0.25rem 0 0 0.25rem; color: #6c757d; font-size: 0.875rem;"},
}

// Initialize themes
func InitThemes() {
	// Register themes with appropriate inline style setting
	RegisterTheme("bootstrap", BootstrapTheme, nil) // false = uses CSS classes
	RegisterTheme("tailwind", TailwindTheme, nil)   // false = uses CSS classes
	RegisterTheme("plain", PlainTheme, nil)         // true = uses inline styles

	// Load templates using the embedded filesystem
	if theme, ok := GetTheme("bootstrap"); ok {
		_ = theme.LoadTemplatesFS(TemplateFS, "gohtml")
	}
	if theme, ok := GetTheme("tailwind"); ok {
		_ = theme.LoadTemplatesFS(TemplateFS, "gohtml")
	}
	if theme, ok := GetTheme("plain"); ok {
		_ = theme.LoadTemplatesFS(TemplateFS, "gohtml")
	}
}

# go-form

A flexible Go package for generating and rendering HTML forms from struct definitions, supporting multiple template engines and easy customization.

## Features

- **Automatic form generation** from Go structs and tags
- **Multiple template sets**: Plain HTML, Bootstrap 5, Tailwind CSS, or your own
- **Supports all common field types**: input, checkbox, radio, dropdown, textarea
- **Nested groups** and fieldsets
- **Customizable error handling and validation**
- **Easy integration** with `html/template`

## Installation

```bash
go get github.com/donseba/go-form
```

## Quick Start

```go
import (
    "github.com/donseba/go-form"
    "html/template"
)

type MyForm struct {
    Username string `form:"input,text" label:"Username" required:"true"`
    Password string `form:"input,password" label:"Password" required:"true"`
    Email    string `form:"input,email" label:"Email"`
}

f := form.NewForm(form.DefaultTemplates) // or form.DefaultBootstrapTemplates, etc.
tmpl := template.Must(template.New("page").Funcs(f.FuncMap()).Parse(`{{ form_render .Form .Errors }}`))
```

## Templates

Choose from built-in template sets or define your own. Example sets:
- `form.DefaultTemplates` (plain HTML)
- `form.DefaultBootstrapTemplates` (Bootstrap 5)
- `form.DefaultTailwindTemplates` (Tailwind CSS)

You can also register custom templates for each field type.

## Supported Field Types

```go
Username string `form:"input,text" label:"Username" required:"true"`
Password string `form:"input,password" label:"Password" required:"true"`
Email    string `form:"input,email" label:"Email"`
Age      int    `form:"input,number" label:"Age" step:"1"`
Country  string `form:"dropdown" label:"Country" values:"us:United States;ca:Canada"`
Message  string `form:"textarea" label:"Message" rows:"5"`
```

### Checkboxes and Radios

```go
AcceptTerms bool   `form:"checkbox" label:"Accept Terms" required
```

## Error Handling

Implement the `FieldError` interface to handle form validation errors:

```go
type fieldError struct {
    Field string
    Issue string
}

func (fe fieldError) Error() string {
    return fmt.Sprintf("%s: %s", fe.Field, fe.Issue)
}

func (fe fieldError) FieldError() (field, err string) {
    return fe.Field, fe.Issue
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
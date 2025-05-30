# go-form

A Go package for rendering forms based on struct layout and tags. This package provides a simple way to generate HTML forms with various styling options.

## Features

- Multiple template options:
  - Plain HTML with custom styling
  - Bootstrap 5
  - Tailwind CSS
  - bring your own template
- Support for most common form field types:
  - Input fields (text, password, email, tel, number, date)
  - Checkboxes
  - Radio buttons (both tag-based and struct-based)
  - Dropdowns
  - Text areas
- Nested form groups
- Error handling
- Required field validation
- Customizable styling

## Installation

```bash
go get github.com/donseba/go-form
```

## Usage

See the [example directory](example/) for a complete implementation. The example demonstrates:
- Form struct definition with various field types
- Template selection and rendering
- Error handling
- Form submission handling

### Basic Structure

```go
// Create form instance with desired template
formInstance := form.NewForm(templates.TailwindV3) // or templates.Plain, templates.BootstrapV5

// Create template with form render function
tmpl := template.Must(template.New("example").Funcs(formInstance.FuncMap()).Parse(`
    {{ form_render .Form .Errors }}
`))
```

## Form Field Types

### Input Fields
```go
Username string `form:"input,text" label:"Username" placeholder:"Enter username" required:"true"`
Password string `form:"input,password" label:"Password" placeholder:"Enter password" required:"true"`
Email    string `form:"input,email" label:"Email" placeholder:"Enter email" required:"true"`
Phone    string `form:"input,tel" label:"Phone" placeholder:"Enter phone"`
Age      int    `form:"input,number" label:"Age" placeholder:"Enter age" step:"1"`
Birthday string `form:"input,date" label:"Birthday" placeholder:"Enter birthday"`
```

### Checkboxes and Radio Buttons
```go
// Tag-based radio buttons
Gender string `form:"radios" label:"Gender" values:"male:Male;female:Female;other:Other"`

// Struct-based radio buttons
type GenderOptions struct {
    Male   bool `form:"radios" label:"Male"`
    Female bool `form:"radios" label:"Female"`
    Other  bool `form:"radios" label:"Other"`
}
```

### Dropdowns
```go
Country string `form:"dropdown" label:"Country" values:"us:United States;ca:Canada;uk:United Kingdom"`
```

### Text Areas
```go
Message string `form:"textarea" label:"Message" placeholder:"Enter message" rows:"5" cols:"50"`
```

### Nested Groups
```go
Address struct {
    Street1 string `form:"input,text" label:"Street Address" required:"true"`
    City    string `form:"input,text" label:"City" required:"true"`
    State   string `form:"input,text" label:"State" required:"true"`
    Zip     string `form:"input,text" label:"ZIP Code" required:"true"`
} `legend:"Address Information"`
```

## Radio Button Implementation

The package supports two approaches for implementing radio buttons:

### 1. Tag-based Radio Buttons
Use the `values` tag to define radio options:
```go
Gender string `form:"radios" label:"Gender" values:"male:Male;female:Female;other:Other"`
```
This approach is useful when you want to define the options directly in the struct tag.

### 2. Struct-based Radio Buttons
Use boolean fields in a struct to represent radio options:
```go
type GenderOptions struct {
    Male   bool `form:"radios" label:"Male"`
    Female bool `form:"radios" label:"Female"`
    Other  bool `form:"radios" label:"Other"`
}
```
This approach is useful when you want to represent radio options as boolean fields in your struct.

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
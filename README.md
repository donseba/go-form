# go-form

A Go library for rendering HTML forms from Go structs using struct tags and Go templates. Supports multiple template styles (Plain, Bootstrap 5, Tailwind CSS) and a wide range of HTML input types.

---

## Features

- Define forms as Go structs with struct tags for field type, label, placeholder, and more
- Supports many HTML input types: text, password, email, tel, number, date, color, range, datetime-local, time, week, month, hidden
- Checkbox, radio, dropdown, and textarea fields
- Grouping and nested struct support for form sections
- Built-in template sets: Plain, Bootstrap 5, Tailwind CSS
- Integrates with `html/template` via a FuncMap

---

## Installation

```
go get github.com/donseba/go-form
```

---

## Quick Start

```go
import (
    "github.com/donseba/go-form"
    "github.com/donseba/go-form/templates"
    "html/template"
)

type ExampleForm struct {
    Username string `form:"input,text" label:"Username" placeholder:"Enter your username" required:"true"`
    Password string `form:"input,password" label:"Password" placeholder:"Enter your password" required:"true"`
    Email    string `form:"input,email" label:"Email" placeholder:"Enter your email" required:"true"`
    Age      int    `form:"input,number" label:"Age" placeholder:"Enter your age" step:"1"`
}

f := form.NewForm(templates.Plain) // or templates.BootstrapV5, templates.TailwindV3
funcMap := f.FuncMap()
tmpl := template.Must(template.New("form").Funcs(funcMap).Parse(`{{ form_render .Form nil }}`))
```

---

## Supported Templates

| Template Name            | Description                |
|-------------------------|----------------------------|
| `templates.Plain`       | Plain HTML, minimal styles |
| `templates.BootstrapV5` | Bootstrap 5 form styles    |
| `templates.TailwindV3`  | Tailwind CSS v3 styles     |

---

## Supported Input Fields & Options

| Field Type / Tag Example         | Description         | Options (Struct Tags)                                  |
|----------------------------------|---------------------|-------------------------------------------------------|
| `form:"input,text"`             | Text input          | `label`, `placeholder`, `required`, `maxlength`        |
| `form:"input,password"`         | Password input      | `label`, `placeholder`, `required`                     |
| `form:"input,email"`            | Email input         | `label`, `placeholder`, `required`                     |
| `form:"input,number"`           | Number input        | `label`, `placeholder`, `required`, `min`, `max`, `step`|
| `form:"input,date"`             | Date input          | `label`, `placeholder`, `required`                     |
| `form:"input,datetime-local"`   | DateTime input      | `label`, `placeholder`, `required`                     |
| `form:"input,time"`             | Time input          | `label`, `placeholder`, `required`                     |
| `form:"input,week"`             | Week input          | `label`, `placeholder`, `required`                     |
| `form:"input,month"`            | Month input         | `label`, `placeholder`, `required`                     |
| `form:"input,color"`            | Color input         | `label`, `placeholder`, `required`                     |
| `form:"input,range"`            | Range input         | `label`, `min`, `max`, `step`                          |
| `form:"input,hidden"`           | Hidden input        | `value`                                               |
| `form:"input,search"`           | Search input        | `label`, `placeholder`                                 |
| `form:"input,url"`              | URL input           | `label`, `placeholder`                                 |
| `form:"input,tel"`              | Telephone input     | `label`, `placeholder`                                 |
| `form:"input,image"`            | Image input         | `label`, `src`, `alt`                                  |
| `form:"checkbox"`               | Checkbox            | `label`, `required`                                    |
| `form:"radios"`                 | Radio group         | `label`, `values` (e.g. `a:A;b:B`), `required`         |
| `form:"dropdown"`               | Dropdown/select     | `label`, `values` (e.g. `a:A;b:B`), `required`         |
| `form:"textarea"`               | Multi-line text     | `label`, `placeholder`, `rows`, `cols`, `maxlength`    |

Other supported tags:
- `legend` — For grouping/nested structs (section title)
- `description` — Field description/help text

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.


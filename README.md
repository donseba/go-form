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
- CSRF Protection 

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
- `maxLength` — Maximum length for textarea or string input
- `class` — Custom CSS class for the field
- `data` — Custom data attributes (e.g., `data="custom:value,foo:bar,baz:qux"`)

---

## Validation

### Built-in Validation
- **required**: Ensures the field is not empty.
- **min, max, step**: For numeric fields, enforces minimum, maximum, and step values.
- **minLength, maxLength**: For string/textarea fields, enforces minimum and maximum character count (Unicode-aware).
- **values**: For radios/dropdowns, ensures the value is one of the allowed options.
- **Email format**: Checks for a valid email address format (basic @ check).
- **Enumerator, Mapper, SortedMapper**: If a field implements one of these interfaces, the value must be present in the allowed set returned by Enum(), Mapper(), or SortedMapper().

### Custom Validation
You can add your own validation logic using the `validate` struct tag and by registering a custom validation function:

```go
// 1. Define your validation function (must return form.FieldErrors)
func isHexColor(val any, field reflect.StructField) form.FieldErrors { /* ... */ }

// 2. Register it with your Form instance
f.RegisterValidationMethod("isHexColor", isHexColor)

// 3. Use it in your struct
type MyForm struct {
    Color string `form:"input,text" label:"Color" validate:"isHexColor"`
}

// 4. Call f.ValidateForm(&myForm) to run both built-in and custom validations
```

Custom validators can be chained with commas in the `validate` tag. All errors are collected and can be rendered in your template.

---

## Translation / Internationalization

go-form supports translation of form labels, error messages, and other UI text. You can provide your own translation function and a Localizer implementation to render forms in different languages or customize the wording for your application.

### How to Use

1. **Create a translation function**: This function receives a Localizer, a key, and optional arguments, and returns the translated string.
2. **Implement a Localizer**: This determines the current locale (e.g., from the user session or request).
3. **Create the form with translation support**: Use `form.NewTranslatedForm(template, translateFunc)`.
4. **Pass your Localizer when rendering or validating**: The form will use your translation function and Localizer to fetch translations.

```go
// Example translation function and Localizer
var translations = map[string]map[string]string{
    "en": {"Name": "Name", "form.validation.required": "is required"},
    "it": {"Name": "Nome", "form.validation.required": "è obbligatorio"},
}

type MyLocalizer struct { Locale string }
func (l MyLocalizer) GetLocale() string { return l.Locale }

func myTranslate(loc form.Localizer, key string, args ...any) string {
    locale := "en"
    if l, ok := loc.(MyLocalizer); ok {
        locale = l.Locale
    }
    msg := key
    if m, ok := translations[locale]; ok {
        if t, ok := m[key]; ok {
            msg = t
        }
    }
    if len(args) > 0 {
        return fmt.Sprintf(msg, args...)
    }
    return msg
}

f := form.NewTranslatedForm(templates.Plain, myTranslate)
// When rendering or validating, pass your Localizer:
loc := MyLocalizer{Locale: "it"}
// ...
```

See the example in `example/translation/main.go` for a complete usage demonstration.

---

### Custom Form Attributes
You can set custom HTML attributes on forms (e.g., `hx-post`, `data-*`, etc.) using the `Attributes` field in your form struct:

```go
form := &FormField{
    // ...other fields...
    Attributes: map[string]string{
        "hx-post": "/some-url",
        "data-custom": "value",
    },
}
```

### Input Groups (Prepend/Append)
You can prepend or append content to input fields using the `group` tag. This is supported in all template sets (Plain, Bootstrap 5, Tailwind CSS):

```go
type ExampleForm struct {
    Username string `form:"input,text" label:"Username" group:"@,.com"`
}
```
This will render an input with `@` before and `.com` after the field, styled according to the selected template.

---

### CSRF Protection

go-form includes built-in CSRF (Cross-Site Request Forgery) protection for your forms. This prevents attackers from tricking users into submitting unauthorized requests.

#### Basic Usage

1. Create a form renderer which adds a default CSRF protection by default:
   ```go
   formRenderer := form.NewForm(templates.BootstrapV5)
   ```

2. Apply the CSRF middleware to your handlers:
   ```go
   // With standard http.ServeMux:
   protectedHandler := formRenderer.CSRFMiddleware()(yourHandler) // <-- wrap your handler
   mux.Handle("/", protectedHandler)
   
   // With Chi router:
   import "github.com/go-chi/chi/v5"
   
   router := chi.NewRouter()
   router.Use(formRenderer.CSRFMiddleware()) // <-- load the middleware
   ```

3. Inject the CSRF token into your form object Info before rendering:
```go
   
  loginForm := LoginForm{
    Info: form.Info{
      Target:     "/login",
      Method:     "post",
      SubmitText: "Log In",
    },
  }
  
  form.InjectCSRFToken(r, &loginForm.Info)

```

The middleware automatically:
- Generates a secure random token for each form
- Validates the token on submission
- Refreshes tokens after each submission
- Rejects requests with missing or invalid tokens

#### Custom Error Handling

By default, CSRF validation failures return HTTP error responses. For a better user experience, you can provide custom error handling:

```go
options := form.CSRFOptions{
  ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
    switch {
      case errors.Is(err, csrf.ErrTokenMismatch):
      http.Error(w, "Invalid CSRF token", http.StatusForbidden)
      case errors.Is(err, csrf.ErrTokenExpired):
      http.Error(w, "CSRF token expired", http.StatusForbidden)
      case errors.Is(err, csrf.ErrKeyOrTokenEmpty):
      http.Error(w, "CSRF token or session ID is empty", http.StatusBadRequest)
      case errors.Is(err, csrf.ErrTokenNotFound):
      http.Error(w, "CSRF token not found", http.StatusBadRequest)
      default:
      http.Error(w, "CSRF validation error: "+err.Error(), http.StatusBadRequest)
    }
  },
}

// Use the custom options
protectedHandler := formRenderer.CSRFMiddlewareWithOptions(options)(yourHandler)
```

#### Alternative CSRF Stores

The default in-memory CSRF store is suitable for single-server applications. For production or distributed environments, you can implement a custom `CSRFStore` that uses Redis, a database, or another shared storage mechanism:

```go
// Example Redis CSRF Store implementation
type RedisCSRFStore struct {
    client *redis.Client
    prefix string
    ttl    time.Duration
}

func (s *RedisCSRFStore) Store(key, token string) error {
    return s.client.Set(ctx, s.prefix+key, token, s.ttl).Err()
}

func (s *RedisCSRFStore) Get(key string) (string, error) {
    val, err := s.client.Get(ctx, s.prefix+key).Result()
    if err == redis.Nil {
        return "", csrf.ErrTokenNotFound
    }
    return val, err
}

// ... implement other required methods ...

// Then use it with your form:
store := &RedisCSRFStore{
    client: redisClient,
    prefix: "csrf:",
    ttl:    30 * time.Minute,
}
formRenderer.SetCSRFStore(store)
```

See the example in `example/csrf/main.go` for a complete usage demonstration.

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

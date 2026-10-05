<p align="center">
    <a href="https://docs.gowebthings.com/go-form">
        <img src="./assets/go-form-logo.png" alt="go-form" height="70">
    </a>
</p>

# go-form

[Documentation](https://docs.gowebthings.com/go-form) · Part of [go-webthings](https://gowebthings.com/components).

A Go library for rendering HTML forms from Go structs using struct tags and Go templates. Supports multiple template styles (Plain, Bootstrap 5, Tailwind CSS) and a wide range of HTML input types.

---

## Features

- Define forms as Go structs with struct tags for field type, label, placeholder, and more
- Supports many HTML input types: text, password, email, tel, number, date, color, range, datetime-local, time, week, month, hidden
- Checkbox, radio, dropdown, and textarea fields
- Grouping and nested struct support for form sections
- Built-in themes: Plain, Bootstrap 5, Tailwind CSS v3 and v4
- Integrates with `html/template` via a FuncMap
- CSRF Protection
- **SortedSelect** and **SortedMultiSelect** for type-safe, mapped dropdowns and multi-selects (see examples)

---

## Installation

```sh
go get github.com/donseba/go-form/v2
```

---

## Quick Start

```go
package main

import (
	"html/template"
	"log"
	"os"

	"github.com/donseba/go-form/v2"
)

type Contact struct {
	Name  string `form:"input,text" label:"Name" required:"true"`
	Email string `form:"input,email" label:"Email" required:"true"`
}

func main() {
	renderer := form.NewForm()
	renderer.SetTheme("plain")

	page := template.Must(template.New("contact").Funcs(renderer.FuncMap()).Parse(
		`{{ form_render .Form .Errors }}`,
	))
	model := form.WithInfo(Contact{}, form.Info{
		Target: "/contacts", Method: "post", SubmitText: "Save contact",
	})
	if err := page.Execute(os.Stdout, map[string]any{
		"Form": model, "Errors": form.FieldErrors(nil),
	}); err != nil {
		log.Fatal(err)
	}
}
```

---

The example prints form HTML. In an HTTP handler, use `WithRequestInfo` and wrap the handler with the same renderer's `CSRFMiddleware()` to supply and validate the CSRF token. Use `form.MapForm(r, &input)` to map a submission and `renderer.ValidateForm(&input)` to collect field errors. Form metadata stays outside your application struct.

## Supported Templates

Select a theme with `renderer.SetTheme(name)` before serving requests.

| Theme name | Description |
|------------|-------------|
| `plain` | Plain HTML with minimal inline styles |
| `bootstrap` | Bootstrap 5 classes (default) |
| `tailwind` | Tailwind CSS v3 classes |
| `tailwindv4` | Tailwind CSS v4 classes |

Class-based themes require your application to load the corresponding stylesheet.

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
| `form:"multicheckbox"`          | Multi-checkbox group| `label`, `values` (e.g. `a:A;b:B`), `required`         |

Other supported tags:
- `legend` — For grouping/nested structs (section title)
- `description` — Field description/help text
- `maxLength` — Maximum length for textarea or string input
- `class` — Custom CSS class for the field
- `data` — Custom data attributes (e.g., `data="custom:value,foo:bar,baz:qux"`)
- `translate` — Translate option labels: `translate:"true"` for Enumerator, Mapper and SortedMapper fields (such as `SortedSelect`), `translate:"false"` to keep labels from a `values` tag as written (see [Translating Option Labels](#translating-option-labels))

---

## Validation

### Built-in Validation
- **required**: Ensures the field is not empty. For selects (`SortedSelect`, `SortedMultiSelect`, `Mapper` and `SortedMapper` fields) a choice must be selected: an empty key (such as a `"" → "—"` placeholder) or a key that is not one of the options counts as missing, and a multi-select needs at least one key. The error uses `TranslationKeyRequired`, like other required fields.
- **min, max, step**: For numeric fields, enforces minimum, maximum, and step values.
- **minLength, maxLength**: For string/textarea fields, enforces minimum and maximum character count (Unicode-aware).
- **values**: For radios/dropdowns, ensures the value is one of the allowed options.
- **Email format**: Checks for a valid email address format (basic @ check).
- **Enumerator, Mapper, SortedMapper**: If a field implements one of these interfaces, the value must be present in the allowed set returned by Enum(), Mapper(), or SortedMapper().

#### Using Enumerator, Mapper, and SortedMapper Interfaces

For enum values, implement `Enumerator`:

```go
type Status string
func (s Status) Enum() []any { return []any{"active", "inactive"} }

type MyForm struct {
    Status Status `form:"dropdown" label:"Status"`
}
```

For key-value pairs, use `Mapper` (unordered) or `SortedMapper` (ordered):

```go
type ColorMap string
func (c ColorMap) Mapper() map[string]string {
    return map[string]string{"red": "Red", "blue": "Blue"}
}

// For ordered pairs, implement SortedMapper with []SortedMap
```

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

Validation errors use the same `name` tags as form mapping and rendering,
including nested paths such as `shipping.city_name`. Custom validators still
receive the original Go struct field; errors they return for that field are
mapped to its rendered name.

---

## Translation / Internationalization

go-form supports translation of form labels, error messages, and other UI text. You can provide your own translation function and a Localizer implementation to render forms in different languages or customize the wording for your application.

### How to Use

1. **Create a translation function**: This function receives a Localizer, a key, and optional arguments, and returns the translated string.
2. **Implement a Localizer**: This determines the current locale (e.g., from the user session or request).
3. **Create the form with translation support**: Use `form.NewTranslatedForm(translateFunc)`.
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

f := form.NewTranslatedForm(myTranslate)
f.SetTheme("plain")
// When rendering or validating, pass your Localizer:
loc := MyLocalizer{Locale: "it"}
// ...
```

See the example in `example/translation/main.go` for a complete usage demonstration.

#### Translating Enum Values

By default, enum values display as-is. To enable translation, add `translate:"true"`:

```go
type Status string
func (s Status) Enum() []any { return []any{"active", "inactive"} }

type MyForm struct {
    Status Status `form:"dropdown" label:"Status" translate:"true"`
}
```

To enable translation for **all enums by default**, set the global variable:

```go
import "github.com/donseba/go-form/v2"

func init() {
    form.DefaultEnumTranslation = true  // All enums will be translatable by default
}
```

Individual fields can still opt-out using `translate:"false"`. Translation keys follow the format `enum||{TypeName}.{value}`:

```go
var translations = map[string]map[string]string{
    "en": {"enum||Status.active": "Active", "enum||Status.inactive": "Inactive"},
    "it": {"enum||Status.active": "Attivo", "enum||Status.inactive": "Inattivo"},
}
```

#### Translating Option Labels

Dropdowns, radio groups and multi-checkboxes follow one rule: an option label is passed to the translation function when its option asks for it (`types.FieldValue.Translate`).

- Labels from a `values` tag and the labels of a struct radio group are text of your application, like field labels, so they are translated. Add `translate:"false"` to show them as written.
- Labels from an `Enumerator`, `Mapper` or `SortedMapper` (including `SortedSelect` and `SortedMultiSelect`) are often data, such as names from a database, so they are shown as written. Add `translate:"true"` (or set `DefaultEnumTranslation`) to use them as translation keys:

```go
type MyForm struct {
    // "Small" and "Large" are translated.
    Size string `form:"dropdown" label:"Size" values:"s:Small;l:Large"`
    // The labels of the source map are translation keys.
    Role form.SortedSelect[string] `form:"dropdown" label:"Role" translate:"true"`
    // Page titles from the database are shown as written.
    Page form.SortedSelect[int64] `form:"dropdown" label:"Page"`
}
```

> In v2.4.0 and earlier, multi-checkboxes and radio groups translated every option label while dropdowns translated none. Add `translate:"true"` to `SortedMultiSelect` fields whose labels are translation keys.

#### Translating the Required Marker

Required fields show an asterisk and, for screen readers only, the text `(required)`. That text is the translation key `form||(required)` (`form.TranslationKeyRequiredLabel`); without a translation function it shows in English.

---

### Custom Form Attributes
You can set custom HTML attributes on forms (e.g., `hx-post`, `data-*`, etc.) using `form.Info.Attributes` with `WithInfo` or `WithRequestInfo`:

```go
model := form.WithRequestInfo(r, input, form.Info{
    Target: "/some-url",
    Method: "post",
    Attributes: map[string]string{
        "hx-post": "/some-url",
        "data-custom": "value",
    },
})
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
   formRenderer := form.NewForm()
   formRenderer.SetTheme("bootstrap")
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

3. Associate the form metadata with the model before rendering. This also
   injects the request's CSRF token:
```go

  loginForm := LoginForm{Email: "name@example.com"}
  renderModel := form.WithRequestInfo(r, loginForm, form.Info{
    Target:     "/login",
    Method:     "post",
    SubmitText: "Log In",
  })

```

`WithInfo` works when CSRF is not needed. `WithContextInfo` is useful in a
rendering service that receives a `context.Context` rather than an HTTP
request. These wrappers let application- or provider-owned structs use full
form metadata without embedding `form.Info` or changing their field names.

Models that already embed `form.Info` remain supported. For those models,
`InjectCSRFToken` can still populate the embedded metadata directly.

The middleware automatically:
- Generates a secure random token for each form
- Validates the token on submission
- Refreshes tokens after each submission
- Rejects requests with missing or invalid tokens

Each issued token is stored separately and bound to its session. Loading another
page, an HTMX fragment, or a second form does not invalidate earlier forms.
Submitting a form consumes only its token and supplies a fresh token to the
handler, so forms in other tabs remain usable until their tokens expire.

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

The default in-memory CSRF store belongs to one `Form` instance in one
process. For multiple application instances, configure every instance with a
store backed by the same Redis or database service before creating the
middleware. A request that renders a form and the later submission may reach
different instances. Use the same token expiration on every instance.

Implement `csrf.Store` with `Store` and `Validate` methods. Treat the key as an
opaque identifier for a session-bound token, rather than the raw session ID.
Existing forms must be reloaded when deploying the change from session-only
storage keys.

Stores may also implement `csrf.TokenConsumer` with `Consume(key, token string)
to validate and remove a token atomically. Both built-in memory stores do this.
Shared stores should implement it with a transaction or equivalent atomic
operation to reject simultaneous reuse across application instances. Stores
that only implement `Store` retain sequential one-use behavior through
validation followed by token invalidation.

For example, with
`github.com/redis/go-redis/v9`:

```go
type RedisCSRFStore struct {
    client *redis.Client
}

func (s *RedisCSRFStore) Store(key, token string) error {
    return s.client.Set(context.Background(), "csrf:"+key, token, csrf.DefaultExpirationTime).Err()
}

func (s *RedisCSRFStore) Validate(key, token string) error {
    if key == "" || token == "" {
        return csrf.ErrKeyOrTokenEmpty
    }
    stored, err := s.client.Get(context.Background(), "csrf:"+key).Result()
    if err == redis.Nil {
        return csrf.ErrTokenNotFound
    }
    if err != nil {
        return err
    }
    if subtle.ConstantTimeCompare([]byte(stored), []byte(token)) != 1 {
        return csrf.ErrTokenMismatch
    }
    return nil
}

formRenderer := form.NewForm()
formRenderer.SetCSRFStore(&RedisCSRFStore{client: redisClient})
protectedHandler := formRenderer.CSRFMiddleware()(yourHandler)
```

The snippet uses `context`, `crypto/subtle`, `redis`, `form`, and `csrf` imports.
Configure the shared store on every server. The default memory store is useful
for local development and single-process deployments.

See the example in `example/csrf/main.go` for a complete usage demonstration.

---

## SortedSelect and SortedMultiSelect

`SortedSelect` and `SortedMultiSelect` are generic types for type-safe, mapped dropdowns and multi-selects with custom key types. They support form mapping, validation, database integration, and JSON serialization.

- **SortedSelect** is for single-value dropdowns (e.g., `DepartmentID form.SortedSelect[int64]`).
- **SortedMultiSelect** is for multi-value selections (e.g., `DepartmentsMulti form.SortedMultiSelect[int64]`).
- Both support any comparable Go type as the key: `int`, `int64`, `string`, `float64`, `uuid.UUID`, `time.Time`, etc.
- They work seamlessly with form rendering, validation, and database/sql or JSON marshalling.

**Minimal usage example:**

```go
import "github.com/donseba/go-form/v2"

// Single select
DepartmentID form.SortedSelect[int64] `form:"dropdown" label:"Department"`

// Multi select
DepartmentsMulti form.SortedMultiSelect[int64] `form:"multicheckbox" label:"Departments"`

// Initialize with a source map
form.NewSortedSelect(map[int64]string{1: "HR", 2: "IT"})
form.NewSortedMultiSelect(map[int64]string{1: "HR", 2: "IT"})
```

`MapForm` sets a select from the posted key. A key that is not one of the options leaves the selection empty, so `required:"true"` reports it; `MapForm` does not abort or write output for it. To see such skipped values (also malformed dates), set a handler:

```go
form.MapFormErrorHandler = func(field string, err error) {
    slog.Debug("form value ignored", "field", field, "error", err)
}
```

For advanced usage, see [`example/sortedselect/main.go`](example/sortedselect/main.go) — covers single and multi-select fields, supported key types, pre-filled and user-submitted values, validation, error handling, JSON and DB integration.

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

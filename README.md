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
- Integrates with `html/template` via a FuncMap, rendering a whole form or [field by field](#laying-out-fields-yourself)
- [Forms defined at runtime](#forms-defined-at-runtime) (`Schema`) render and validate like structs
- [File uploads](#file-uploads), [time zones](#time-zones), [trimming](#trimming-white-space), [strict mapping](#strict-mapping) and [read-only forms](#read-only-forms)
- [Binding](#binding-a-submission) a submission in one step, [limits](#limits-that-change-at-runtime) and [options](#options-known-at-render-time) that change at runtime for rendering and validation alike, and [showing a failed submission again](#showing-a-failed-submission-again) after a redirect
- [Text extraction](#extracting-texts) for translators, and [rendering from Go](#rendering-from-go)
- CSRF Protection, and a [honeypot](#spam-honeypot) against spam bots
- Messages in English and Dutch, in the localizer's language ([message catalogues](#message-catalogues))
- **SortedSelect** and **SortedMultiSelect** for type-safe, mapped dropdowns and multi-selects, sorted by key or label or in [your own order](#ordered-selects-and-placeholders)

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

The example prints form HTML. In an HTTP handler, use `WithRequestInfo` and wrap the handler with the same renderer's `CSRFMiddleware()` to supply and validate the CSRF token. Use `form.MapForm(r, &input)` to map a submission and `renderer.ValidateForm(&input)` to collect field errors, or `renderer.Bind` for both in one step (see [Binding a Submission](#binding-a-submission)). Form metadata stays outside your application struct.

## Supported Templates

Select a theme with `renderer.SetTheme(name)` before serving requests.

| Theme name | Description |
|------------|-------------|
| `plain` | Plain HTML with minimal inline styles |
| `bootstrap` | Bootstrap 5 classes (default) |
| `tailwind` | Tailwind CSS v3 classes |
| `tailwindv4` | Tailwind CSS v4 classes |

Class-based themes require your application to load the corresponding stylesheet.

### Your Own Theme

`templates.RegisterClassTheme` registers a theme with your CSS classes and go-form's markup. Registering the same name and classes again returns the same theme, so it can run where the renderer is created:

```go
_, err := templates.RegisterClassTheme("site", templates.ThemeClasses{
    Wrapper: templates.StyleOption{Class: "field"},
    Label:   templates.StyleOption{Class: "field-label"},
    Input:   templates.StyleOption{Class: "control"},
    Error:   templates.StyleOption{Class: "field-error"},
})
if err != nil {
    return err
}

renderer.SetTheme("site")
```

A theme of your own hides text for screen readers only, such as the `(required)` of a required field, with the class `sr-only` unless `VisuallyHidden` names another class, so its CSS needs a rule for that class.

A group of radio buttons, and a multicheckbox, renders as a `<fieldset>` with the field's label as its `<legend>` (`fieldset.gohtml` and `legend.gohtml`), so the question names the group for screen readers: the fieldset gets the classes of `Wrapper` and `Fieldset`, the legend those of `Label` and `Legend`. Use `Fieldset` and `Legend` for what only the group needs, such as a reset of the browser's fieldset border and padding (the plain theme has one inline).

To change the markup, load a complete template set of your own into a theme with `LoadTemplatesFS` (or `LoadTemplates` from disk), such as a copy of go-form's templates in `templates/gohtml` (`input.gohtml`, `select.gohtml`, `wrapper.gohtml`, `fieldset.gohtml`, `form-buttons.gohtml`, …) with your changes. Your templates can use the same functions as go-form's, such as `themeClass`, `form_print`, `form_described_by` and `form_class`, which joins the classes that are not empty, so an attribute without classes can be left out: `{{with form_class (themeClass "input") .Field.Class}}class="{{.}}"{{end}}`. Register themes and load their templates before serving requests, under a name of your own.

The opening form tag comes from `form-open.gohtml`, which `form_open` renders and `form_render` too, through `form.gohtml`. A `form-open.gohtml` of your own, or a `form.gohtml` that writes the tag itself, renders the form's attributes besides `action` and `method` with four functions, as go-form's `form-open.gohtml` does:

- `form_open_class` and `form_open_style`: the classes and inline style of `Info.Attributes` and of `form_open`, besides the theme's (`themeClass "form"`, `themeStyle "form"`); join them with `form_class` and `form_style`;
- `form_attributes .Field.Attributes`: the other attributes of `Info.Attributes`, such as `enctype`;
- `form_open_attributes`: the other attributes the template gave `form_open` (see [Attributes in the Template](#attributes-in-the-template)).

A template that leaves one out drops those attributes.

---

## Supported Input Fields & Options

| Field Type / Tag Example         | Description         | Options (Struct Tags)                                  |
|----------------------------------|---------------------|-------------------------------------------------------|
| `form:"input,text"`             | Text input          | `label`, `placeholder`, `required`, `minLength`, `maxLength`, `pattern` |
| `form:"input,password"`         | Password input      | `label`, `placeholder`, `required`, `autocomplete`     |
| `form:"input,email"`            | Email input         | `label`, `placeholder`, `required`, `multiple`         |
| `form:"input,number"`           | Number input        | `label`, `placeholder`, `required`, `min`, `max`, `step`|
| `form:"input,date"`             | Date input          | `label`, `placeholder`, `required`                     |
| `form:"input,datetime-local"`   | DateTime input      | `label`, `placeholder`, `required`                     |
| `form:"input,time"`             | Time input          | `label`, `placeholder`, `required`                     |
| `form:"input,week"`             | Week input          | `label`, `placeholder`, `required`                     |
| `form:"input,month"`            | Month input         | `label`, `placeholder`, `required`                     |
| `form:"input,color"`            | Color input         | `label`, `placeholder`, `required`                     |
| `form:"input,range"`            | Range input         | `label`, `min`, `max`, `step`                          |
| `form:"input,hidden"`           | Hidden input        | none; it sends the field's value                       |
| `form:"input,file"`             | File input          | `label`, `required`, `accept`, `multiple`, `maxSize` (see [File uploads](#file-uploads)) |
| `form:"input,search"`           | Search input        | `label`, `placeholder`                                 |
| `form:"input,url"`              | URL input           | `label`, `placeholder`                                 |
| `form:"input,tel"`              | Telephone input     | `label`, `placeholder`                                 |
| `form:"input,image"`            | Image input         | `label` (there are no `src` and `alt` tags)            |
| `form:"textarea"`               | Textarea            | `label`, `placeholder`, `required`, `rows`, `cols`, `minLength`, `maxLength` |
| `form:"checkbox"`               | Checkbox            | `label`, `required`                                    |
| `form:"radios"`                 | Radio group         | `label`, `values` (e.g. `a:A;b:B`), `required`         |
| `form:"dropdown"`               | Dropdown/select     | `label`, `values` (e.g. `a:A;b:B`), `required`, `placeholder` (an empty first option) |
| `form:"multicheckbox"`          | Multi-checkbox group| `label`, `values` (e.g. `a:A;b:B`), `required`, `minItems`, `maxItems` (slice fields and `SortedMultiSelect`) |

A field with a `values` tag may be a string, any integer kind (`values:"60:1 minute;300:5 minutes"` on an `int`), a pointer to one, or for a multi-checkbox a slice of them (`[]string`, `[]int`). Options only known when the form renders, such as the answers of a poll, are given with [`RenderModel.Options`](#options-known-at-render-time).

A struct radio group is a struct field with `form:"radios"` whose bool fields carry the struct's `name`: it renders a radio button per bool field, with the field's `label`, and `MapForm` ticks the chosen field and clears the others. `form:"radios,radio_group"`, with a `legend`, declares the same group explicitly.

```go
type PlanChoice struct {
    Monthly bool `name:"plan" label:"Monthly"`
    Yearly  bool `name:"plan" label:"Yearly"`
}

type SignupInput struct {
    Plan PlanChoice `form:"radios" name:"plan" label:"Plan"`
}
```

Number fields may be pointers, such as `*float64`, to tell "not filled in" (`nil`) from `0`. An optional coordinate:

```go
Latitude *float64 `form:"input,number" label:"Latitude" step:"any" min:"-90" max:"90"`
```

Other supported tags:
- `legend` — For grouping/nested structs (section title)
- `description` — Field description/help text
- `maxLength` — Maximum length for textarea or string input
- `class` — Custom CSS class for the field
- `data` — Custom data attributes: `data:"custom=value,foo=bar"` renders `data-custom="value" data-foo="bar"`
- `translate` — Translate option labels: `translate:"true"` for Enumerator, Mapper and SortedMapper fields (such as `SortedSelect`), `translate:"false"` to keep labels from a `values` tag as written (see [Translating Option Labels](#translating-option-labels))
- `disabled:"true"` — Renders the control disabled: inputs (also hidden and range inputs), textareas, selects, checkboxes, radios and multicheckboxes. Browsers do not submit disabled controls, so `MapForm` leaves such a field as it was, except that a disabled checkbox becomes `false` and, on a POST, a disabled multicheckbox empty. Use `readonly:"true"` for a text field whose value must be sent
- `readonly:"true"` — Renders a text-like input or textarea read-only
- `autocomplete` — The `autocomplete` attribute, such as `autocomplete:"email"`, `autocomplete:"current-password"` or `autocomplete:"new-password"`
- `accept`, `multiple` — For file inputs (`accept:"image/*,.pdf"`); `multiple` also for email inputs
- `minLength`, `maxLength` — Checked on the server and rendered as `minlength` and `maxlength`
- `pattern` — Checked on the server; rendered as the `pattern` attribute only when browsers check it the same way: a pattern anchored with `^` and `$`, without Go-only syntax such as `(?i)` and without a top-level `|`. Other patterns are only checked on the server.
- `placeholder` on a dropdown — Renders an empty first option with that label, such as `placeholder:"—"`
- `line:"true"` — The value must be one line without control characters (see [Built-in Validation](#built-in-validation))
- `wallclock:"true"` — A `datetime-local` field read and shown without a time zone (see [Time Zones](#time-zones))

A field with errors renders its control with `aria-invalid="true"` and an `aria-describedby` that names its description and its error messages (`<id>_description`, `<id>_error_0`, …), so screen readers read them with the control.

---

## Laying Out Fields Yourself

`form_render` renders the whole form. To lay out the fields yourself, such as two fields side by side or a field between your own markup, render them one at a time:

```gohtml
{{ form_open .Form }}
  <div class="row">
    <div class="col">{{ form_field .Form "FirstName" .Errors }}</div>
    <div class="col">{{ form_field .Form "LastName" .Errors }}</div>
  </div>
  {{ form_field .Form "address.city" .Errors }}
  <input name="custom" value="…">
  {{ range form_errors .Errors "custom" }}<p class="error">{{ . }}</p>{{ end }}
  {{ form_buttons .Form }}
{{ form_close }}
```

| Function | Renders |
|----------|---------|
| `form_field model name errors` | One field with its label, description and errors, by the name it renders with (the `name` tag, nested as `address.city`). A group renders with all its fields. |
| `form_errors errors name` | The messages for one field, as `[]string` (also `errors.Messages(name)` in Go). |
| `form_open model [attributes…]` | The opening `<form>` tag with its attributes, and the hidden fields: the CSRF token, the honeypot and `Info.Hidden`. Attributes of the template's own follow the model (see [Attributes in the Template](#attributes-in-the-template)). |
| `form_hidden model` | The hidden fields of `form_open` alone, for a `<form>` tag you write yourself. |
| `form_buttons model` | The cancel link and submit button (no submit button in a read-only form). |
| `form_close` | The closing `</form>` tag. |
| `form_honeypot name` | The spam trap input for a form whose markup you write yourself (see [Spam Honeypot](#spam-honeypot)). |
| `form_has_fields model` | Whether the model has a field to render (`form.HasFields` in Go). |

`form_field_localized`, `form_open_localized`, `form_hidden_localized` and `form_buttons_localized` take a localizer first, like `form_render_localized`. Fields render with the selected theme, so they look the same as in `form_render`. The submit button shows `Info.SubmitText`, or `form.DefaultSubmitText` (`Submit`) when it is empty.

A name the model does not have fails the template, and the error lists the fields the model has, each by the name `form_field` takes, with a group's fields after the group. A typing mistake then shows the right name. For a model with `name`, `email` and an `address` group with `city`:

```text
form: no field "nmae" (fields: name, email, address, address.city)
```

### Attributes in the Template

`form_open` takes attributes of the template's own after the model: name/value pairs and `template.HTMLAttr` values, such as the attributes another library renders for a form that submits with HTMX. go-form does not need to know that library:

```gohtml
{{ form_open .Form "class" "booking" (submitAttrs runtime .Action "target" "#booking" "swap" "outerHTML") }}
  {{ form_field .Form "name" .Errors }}
  <button type="submit">Book</button>
{{ form_close }}
```

```html
<form action="/book" method="post" class="site-form booking" hx-post="/book" hx-target="#booking" hx-swap="outerHTML">
<input type="text" name="website" … aria-hidden="true">
<input type="hidden" name="back" value="/workshop#booking">
```

- `class` adds to the theme's class and to `Info.Attributes["class"]`, in one `class` attribute; a class that is already there is not repeated. `style` adds to the theme's and `Info.Attributes["style"]` the same way.
- `action`, `method` and `enctype` come from the form metadata (`Info.Target`, `Info.Method`, `Info.Attributes["enctype"]` or a file input). Giving them in the template is an error, rather than a second attribute the browser ignores.
- Any other attribute replaces the one of the same name in `Info.Attributes`; given twice, the last one counts.
- The values of pairs are escaped. A pair cannot set an event handler (`on…`); that takes a trusted `template.HTMLAttr`. A `nil` value renders the name alone, such as `"novalidate" nil`; a `nil` attribute is skipped, so an optional one can be left out.

From Go, `Open(loc, model, attributes...)` renders the same. A template that has to write the `<form>` tag itself renders the hidden fields with `{{ form_hidden .Form }}` (`Hidden(loc, model)` in Go).

### Hidden Fields

`Info.Hidden` adds hidden inputs to the form, such as the page to return to. `form_open` and `form_render` render them after the CSRF token and the honeypot, one input per value, in name order:

```go
model := form.WithInfo(input, form.Info{
    Target: "/comments", Method: "post", Honeypot: "website",
    Hidden: url.Values{"back": {r.URL.Path + "#comments"}},
})
```

### Several Forms on One Page

Fields render with their name as their id (`id="email"`), which labels, descriptions and errors refer to. Two forms on one page with a field of the same name, such as a newsletter form and a comment form, so collide. `Info.IDPrefix` puts a prefix and a dash before every id of a form, and before the ids that `for`, `aria-labelledby` and `aria-describedby` refer to; names stay as they are:

```go
newsletter := form.WithInfo(SubscribeInput{}, form.Info{Target: "/subscribe", Method: "post", IDPrefix: "newsletter"})
// <label for="newsletter-email" id="newsletter-email_label"> <input id="newsletter-email" name="email" …>
```

Without a prefix the ids are the names, as before.

### Rendering from Go

The template functions are methods of `*form.Form` too, for applications that wrap them in template functions of their own, such as functions that take the localizer of the page from a context. A nil localizer renders as the functions without a localizer do (`form.DefaultLocalizer`, English).

| Method | Template function |
|--------|-------------------|
| `Render(loc, model, errs)` | `form_render_localized` |
| `Field(loc, model, name, errs)` | `form_field_localized` |
| `Open(loc, model, attributes...)` | `form_open_localized`, with or without attributes |
| `Hidden(loc, model)` | `form_hidden_localized` |
| `Buttons(loc, model)` | `form_buttons_localized` |
| `Close()` | `form_close` |
| `Honeypot(loc, name)` | `form_honeypot` |

```go
funcs := template.FuncMap{
    "form_field": func(ctx context.Context, model any, name string, errs form.FieldErrors) (template.HTML, error) {
        return renderer.Field(localizerFrom(ctx), model, name, errs)
    },
}
```

## Limits That Change at Runtime

Some limits are only known when the form renders, such as the places left at an event. `form.Overrides` replace struct tags of fields by the name they render with, for rendering and validation alike, so the browser and the server check the same limit. A model that implements `Overrider` gets them everywhere: when it renders, in `ValidateForm` and in `Bind`:

```go
type BookingInput struct {
    Adults int `form:"input,number" name:"adults" label:"Adults" min:"1" max:"50"`

    left int // the places left; unexported fields never render
}

func (in BookingInput) FormOverrides() form.Overrides {
    return form.Overrides{"adults": {"max": strconv.Itoa(in.left)}}
}

input := BookingInput{left: event.Left}
errs, err := renderer.Bind(r, &input, form.MapOptions{}, loc) // max is event.Left
```

For one render, set `RenderModel.Overrides`; they win over the model's own, and `ValidateForm` applies them when it gets the `RenderModel`:

```go
model := form.WithInfo(input, info)
model.Overrides = form.Overrides{
    "notes":        {"placeholder": "Dietary needs?"},
    "address.city": {"maxLength": "40"},
}
```

Any tag a field reads can be replaced, such as `min`, `max`, `step`, `minLength`, `maxLength`, `pattern`, `placeholder`, `label`, `required`, `accept`, `maxSize` or a message such as `maxMsg`. An empty value removes the tag. The `name` tag cannot be replaced. A nested struct may implement `Overrider` too; the overrides of the outer model win.

## Options Known at Render Time

The options of some fields are only known when the form renders, such as the answers of a poll or the rooms that are free. `RenderModel.Options` gives them per field, by the name the field renders with (`form.FieldOptions`, a `map[string][]types.FieldValue`); they replace a `values` tag:

```go
type VoteInput struct {
    Poll   string `form:"input,hidden" name:"poll"`
    Option string `form:"radios"       name:"option" label:"Your choice" required:"true"`
}

choices := make([]types.FieldValue, 0, len(poll.Options))
for _, option := range poll.Options {
    choices = append(choices, types.FieldValue{Value: option.ID, Name: option.Label})
}

model := form.WithInfo(VoteInput{Poll: poll.ID}, form.Info{Target: "/vote", Method: "post"})
model.Options = form.FieldOptions{"option": choices}
```

```gohtml
{{ form_open .Form }}
    {{ form_field .Form "poll" nil }}
    {{ form_field .Form "option" .Errors }}
    {{ form_buttons .Form }}
{{ form_close }}
```

- A `radios` field renders the options as a group of radio buttons and a `multicheckbox` field as a checkbox each, both in a `<fieldset>` whose `<legend>` is the field's label, and any other field as a dropdown. The field may be a string, an integer kind, a pointer to one, or for a multi-checkbox a slice of them.
- Labels are shown as written, so they may hold any text, `;` and `:` included, and data such as poll answers is not sent to the translation function. Set `Translate` on an option whose label is a translation key; `form.Texts` then extracts it.
- A `Disabled` option renders disabled and is not a valid choice.

Validate with the same `RenderModel`: `ValidateForm`, `ValidateFormLocalized` and `Bind` check that a submitted value is one of the options, with the rule `values` (message `form.TranslationKeyInvalidValue`, or the field's `valuesMsg`). `Bind` takes the `RenderModel` in place of a pointer when its `Model` is a pointer, and maps onto that:

```go
var input VoteInput

model := form.WithInfo(&input, form.Info{Target: "/vote", Method: "post"})
model.Options = form.FieldOptions{"option": choices}

errs, err := renderer.Bind(r, model, form.MapOptions{TrimSpace: true}, loc)
```

A [kept submission](#showing-a-failed-submission-again) shows the chosen option again.

## Read-Only Forms

`Info.ReadOnly` shows a form without letting it be changed, for example to people without the right to edit: every field renders disabled and there is no submit button. A cancel link (`CancelTarget`) still shows, as a way back.

```go
model := form.WithInfo(page, form.Info{Target: "/pages/1", Method: "post", ReadOnly: !canEdit})
```

## File Uploads

Fields of type `*multipart.FileHeader` (one file) and `[]*multipart.FileHeader` (several files, rendered with `multiple`) render as file inputs, and `MapForm` maps the uploaded files to them. A form with a file input gets `enctype="multipart/form-data"` automatically, unless `Info.Attributes` sets another `enctype`.

```go
type Upload struct {
    Title    string                `form:"input,text" label:"Title" required:"true"`
    Document *multipart.FileHeader `form:"input,file" label:"Document" required:"true" accept:".pdf,image/*" maxSize:"10MB"`
}

func upload(w http.ResponseWriter, r *http.Request) {
    r.Body = http.MaxBytesReader(w, r.Body, 11<<20)

    var input Upload
    _ = form.MapForm(r, &input)
    if errs := renderer.ValidateForm(&input); len(errs) > 0 {
        // render the form again with errs
    }
    file, err := input.Document.Open()
    // …
}
```

- `accept` lists extensions (`.pdf`) and media types (`image/png`, `image/*`), like the `accept` attribute. Both the file name and the media type come from the browser, so this guides people; check the content itself where it matters. `form.AcceptsFile(accept, file)` is the same check in Go.
- `maxSize` limits the size of each file: `512KB`, `10MB`, `1.5GB` or bytes. Units are binary (1KB is 1024 bytes). `form.ParseSize` reads such a size. Limit the request body as well, with `http.MaxBytesReader`.
- Without an uploaded file the field keeps its value. `MapForm` parses multipart forms with 32 MB in memory, as `r.FormValue` does; call `r.ParseMultipartForm` first for another limit.

## Time Zones

`MapForm` reads date-times as UTC. When people enter date-times in the site's time zone, map with `MapFormWith` and render with a localizer that has a `Location() *time.Location` method (`form.TimeLocalizer`):

```go
err := form.MapFormWith(r, &event, form.MapOptions{Location: site.Location})

type localizer struct{ locale string; zone *time.Location }
func (l localizer) GetLocale() string { return l.locale }
func (l localizer) Location() *time.Location { return l.zone }
```

```gohtml
{{ form_render_localized .Localizer .Form .Errors }}
```

The zone applies to `datetime-local` inputs, so `time.Time` fields hold the right instant. Date, time, month and week inputs hold no instant and are read and shown as they are. Date-times are accepted with or without seconds; they render with seconds only when the `step` tag is not a whole number of minutes. Month inputs use the format `2006-01`.

Where the zone changes its offset, such as at the start and end of summer time, a time on the clock can be missing or occur twice. go-form reads such times by one rule:

- A time the clocks skip moves forward by the length of the gap: in Amsterdam, where the clocks go from 02:00 to 03:00 on the last Sunday of March, 02:30 is read as 03:30.
- A time that occurs twice is read as the first of the two, before the clocks go back: 02:30 on the last Sunday of October is 02:30 summer time (00:30 UTC).

`form.ParseDateTimeLocal(value, location)` reads a `datetime-local` value such as `2026-03-29T02:30` with that rule, as `MapFormWith` does; a nil location reads it as UTC. `form.WallTime(t, location)` is the instant at which the clocks in `location` show the date and time of day of `t` (the location of `t` itself is ignored), with the same rule; use it for a time of day that belongs to another zone than the site's, such as that of an event:

```go
start, err := form.ParseDateTimeLocal(r.FormValue("start"), nil) // the time on the clock, as UTC
instant := form.WallTime(start, eventZone)
```

A field with `wallclock:"true"` holds such a time on the clock: `MapFormWith` reads it as UTC whatever `MapOptions.Location`, and it renders as stored, without the localizer's zone. Turn it into an instant with `WallTime`:

```go
type EventInput struct {
    Start    time.Time `form:"input,datetime-local" label:"Start" wallclock:"true"`
    TimeZone string    `form:"dropdown" label:"Time zone" values:"Europe/Amsterdam;Europe/London"`
}

zone, _ := time.LoadLocation(input.TimeZone)
startsAt := form.WallTime(input.Start, zone)
```

## Trimming White Space

`MapFormWith(r, &input, form.MapOptions{TrimSpace: true})` removes white space around every submitted string, except in password inputs, where spaces may be part of the password, and fields with `trim:"false"`. Without the option, `trim:"true"` trims one field. A value of only spaces becomes empty, so `required:"true"` reports it. Whatever the option, `required` and `minLength` do not count surrounding spaces (except in password inputs).

## Binding a Submission

`Bind` maps a submission and validates it in one step: it maps strictly (see [Strict Mapping](#strict-mapping)), validates with the overrides of an `Overrider` model, and returns the field errors of both. A field whose value could not be read has that error only, rather than what validation says about the value it was left with:

```go
var input SignupInput

errs, err := renderer.Bind(r, &input, form.MapOptions{TrimSpace: true, NormalizeNewlines: true}, loc)
if err != nil {
    http.Error(w, "Bad request", http.StatusBadRequest) // not a pointer to a struct
    return
}

if len(errs) > 0 {
    // render the form again with errs
}
```

`MapOptions.NormalizeNewlines` reads the line breaks of submitted strings as `\n`. Browsers send a textarea's line breaks as `\r\n` but count them as one character for `maxlength`; normalized, `maxLength` counts them the same way. Password inputs are taken as typed.

`Bind` also takes a `RenderModel` whose `Model` is a pointer, such as the model the form rendered with: it maps onto that pointer and validates with the model's `Overrides` and [`Options`](#options-known-at-render-time).

Besides strings, numbers and booleans, `MapForm` maps pointers to strings and numbers (`*float64`, `*int`, …), which an empty value sets to `nil`, and slices of strings and integers (`[]string`, `[]int`), with a value per ticked box of a multi-checkbox. On a POST, no ticked box means an empty slice.

## Showing a Failed Submission Again

After a POST, many applications redirect, so that reloading the page does not post again. To show the form with the visitor's values and the errors on the page the redirect leads to, keep the failed submission as a `form.Carry`:

```go
errs, err := renderer.Bind(r, &input, form.MapOptions{TrimSpace: true}, loc)
// handle err
if len(errs) > 0 {
    carry, err := form.NewCarry(form.WithInfo(input, info), r.PostForm, errs)
    // handle err
    data, err := carry.Encode(4096) // at most 4096 bytes
    // store data, for example in an encrypted cookie or a database row with a cookie that names it
    http.Redirect(w, r, backURL, http.StatusSeeOther)
    return
}
```

and on the page the visitor returns to:

```go
carry, err := form.DecodeCarry(data)
model := form.WithInfo(SignupInput{}, info)
model.Carry = &carry
```

- `NewCarry(model, values, errs)` keeps the submitted values of the model's fields (a struct, a `RenderModel` or a `Schema`) and the messages of `errs` per field. It never keeps the value of a password input, a file input or the form's honeypot (`Info.Honeypot`), nor values that are not a field of the form, such as a CSRF token.
- `Carry.Encode(limit)` returns compact JSON of at most `limit` bytes (no limit when `limit` is 0). When it does not fit, the longest values are left out first; such a field shows the model's value again, with its errors. When the errors alone do not fit, it returns `form.ErrCarryTooLarge`. `DecodeCarry` reads what `Encode` returned.
- With `RenderModel.Carry` set, `form_render` and `form_field` show the carried values in place of the model's. Hidden inputs keep the model's values.
- A carry only applies when the model's hidden inputs hold the values it was submitted with. Of several forms of one kind on a page, such as a comment form per post with a hidden post ID, only the one that was sent shows it.
- The carried errors only show when the template passes no errors of its own (`nil`); errors it passes win. `Carry.FieldErrors()` returns them (with their messages only, not their `Rule`), for example for `form_errors`.

Values can be personal data, such as an e-mail address: keep the encoding out of reach of others, for example encrypted or on the server.

## Strict Mapping

Without options, `MapForm` skips a submitted value it cannot read as its field's type, such as `abc` or `NaN` for a number or a date in the wrong format: the field keeps its zero or previous value, and only `required:"true"` may notice. With `MapOptions{Strict: true}` such values come back as field errors, so people see which field to correct:

```go
errs, err := renderer.MapFormLocalized(r, &input, form.MapOptions{TrimSpace: true}, loc)
if err != nil {
    http.Error(w, "Bad request", http.StatusBadRequest) // not a pointer to a struct
    return
}

errs = append(errs, renderer.ValidateFormLocalized(&input, loc)...)
```

`MapFormLocalized` maps with `Strict` and returns the unreadable values as `FieldErrors` translated for `loc`: rule `invalid`, message `form.TranslationKeyInvalid` or the field's `invalidMsg` (such as `invalidMsg:"Enter an amount such as 12.50"`). `MapFormWith(r, &input, form.MapOptions{Strict: true})` returns them as an `*form.InvalidValuesError` whose `Errors` have English messages; every other value is still mapped.

Unreadable are: a number that does not parse, does not fit the field (`300` for an `int8`) or is not finite (`NaN`, `Inf`); a date or time in the wrong format; a key a select or multi-select (`SortedSelect`, `SortedMultiSelect`) does not offer, or an option a struct radio group does not have; a checkbox value other than `on`/`off`, `true`/`false` or `1`/`0`; and a malformed UUID. An empty value is never unreadable; `required:"true"` reports it.

## Spam Honeypot

`Info.Honeypot` adds a spam trap to the form: a text field hidden from people (and from screen readers), which many bots fill in. Check it when the form comes in, and answer spam as if it succeeded without storing or sending anything:

```go
model := form.WithRequestInfo(r, input, form.Info{Target: "/contact", Method: "post", Honeypot: "website"})

if form.IsSpam(r, "website") {
    http.Redirect(w, r, "/contact?sent=1", http.StatusSeeOther)
    return
}
```

The field renders with the theme's `honeypot` template, also with `form_open`. A form whose markup a template writes itself renders it with `{{ form_honeypot "website" }}`, with the form's theme.

The field stays a normal text input, never `type="hidden"`, the `hidden` attribute or `display:none`, which bots skip. `ThemeClasses.Honeypot` hides it from people. A theme with a `Honeypot` class, such as Bootstrap's `visually-hidden` or Tailwind's `sr-only` in the built-in themes, renders that class alone, without a `style` attribute, so the field stays hidden under a Content-Security-Policy without `'unsafe-inline'` for styles, which ignores inline styles. A theme without a `Honeypot` class renders the class `form-honeypot` (`templates.HoneypotClass`) with the inline style `templates.HoneypotStyle` (or `Honeypot.Style`), so it is hidden without any CSS but shows under such a policy. A class theme of your own for a site with a strict policy sets the class and hides it in its CSS:

```go
templates.RegisterClassTheme("site", templates.ThemeClasses{
    // …
    Honeypot: templates.StyleOption{Class: "form-honeypot"},
})
```

```css
.form-honeypot { position: absolute; left: -10000px; width: 1px; height: 1px; overflow: hidden; }
```

---

## Validation

### Built-in Validation
- **required**: Ensures the field is not empty. For selects (`SortedSelect`, `SortedMultiSelect`, `Mapper` and `SortedMapper` fields) a choice must be selected: an empty key (such as a `"" → "—"` placeholder) or a key that is not one of the options counts as missing, and a multi-select needs at least one key. The error uses `TranslationKeyRequired`, like other required fields.
- **min, max, step**: For numeric fields (every int, uint and float kind), enforces minimum, maximum, and step values. A value must be a whole number of steps from `min`, or from 0 without a `min`, counted exactly: `0.3` is on `step:"0.1"`, and `0.75` is on `step:"0.5" min:"0.25"`, while `1` is not. A value between two steps gets "Enter a multiple of %g." (`form.TranslationKeyStep`), or "Enter a value in steps of %g from %g." (`form.TranslationKeyStepFromMin`, with the step and the min) when `min` is not itself a whole number of steps from 0. `form.OnStep(value, min, step)` is the same check in Go: an empty `min` counts from 0, and a `step` of `any`, 0 or less allows every number. A [schema](#forms-defined-at-runtime)'s `Min` and `Step` count the same way.
- **Finite numbers**: A float field holding `NaN` or an infinity fails with the rule `invalid` (`form.TranslationKeyInvalid`), whatever its other tags: such values compare false with every limit, so `min` and `max` alone would let them through.
- **Valid text**: Every string field must be valid UTF-8, which browsers always send; other text fails with the rule `invalid`.
- **line**: `line:"true"` (`Spec.SingleLine` in a schema) allows one line without control characters (tabs, line breaks, NUL, …) and without the Unicode line and paragraph separators, for values such as a name or a subject that end up in e-mail headers and lists. `form.SingleLine(s)` is the same check in Go.
- **minLength, maxLength**: For string/textarea fields, enforces minimum and maximum character count (Unicode-aware).
- **minItems, maxItems**: For slice fields (a multi-checkbox into a `[]string`) and for the selected options of a `SortedMultiSelect`, the minimum and maximum number of chosen values.
- **values**: For radios, dropdowns and multi-checkboxes with a `values` tag or [`RenderModel.Options`](#options-known-at-render-time), ensures the value is one of the options; for a multi-checkbox every ticked one. It checks string and integer fields, pointers to them and slices of them. An empty string means nothing is chosen, and so does `0` in an integer field unless `0` is an option (a pointer is `nil` when nothing is chosen); `required:"true"` reports those.
- **Email format**: Email inputs (`form:"input,email"`) and fields with `email:"true"` must hold a plain address such as `name@example.com`: what `net/mail` parses, without a display name or angle brackets, of at most 254 characters. With `multiple:"true"` a comma-separated list. `form.ValidEmail(s)` is the same check in Go.
- **url**: `url:"true"` accepts a URL with a scheme or an absolute path; `url:"http,https"` an absolute URL with one of the listed schemes and a host; `url:"link"` a link on a page: an `http(s)` URL, a path (`/news`), an anchor (`#contact`), or a `mailto:` or `tel:` address. `javascript:`, `vbscript:` and `data:` URLs are never accepted. `form.ValidURL(rule, s)` is the same check in Go, for URLs that do not come from a form.
- **equal**: `equal:"Password"` requires the same value as another field of the struct (by Go name or `name` tag), such as a repeated password.
- **accept, maxSize**: For file fields, see [File Uploads](#file-uploads).
- **pattern**: The value must match the regular expression. The message does not show the expression; set `patternMsg` to say what is expected.
- **prefix, suffix, contains**: The string must start with, end with or contain the text of the tag, such as `prefix:"https://"`.
- **bool**: `bool:"true"` requires a ticked checkbox, such as an agreement to the terms.
- **zero**: `zero:"true"` requires the zero value, such as an empty string.
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
func (c ColorMap) String() string { return string(c) }

// For ordered pairs, implement SortedMapper: SortedMapper() []SortedMap and String() string
```

### Messages Per Rule

Each rule has its own message tag: the rule name followed by `Msg`, such as `requiredMsg`, `minLengthMsg`, `maxLengthMsg`, `patternMsg`, `emailMsg`, `urlMsg`, `minMsg`, `maxMsg`, `stepMsg`, `valuesMsg` (also for selects and enums), `equalMsg`, `acceptMsg` and `maxSizeMsg`. Like labels, these messages are translated, and get the rule's value when they have a verb for it:

```go
Slug string `form:"input,text" label:"Address" required:"true" pattern:"^[a-z0-9-]+$" patternMsg:"Use lowercase letters, digits and dashes" minLength:"3" minLengthMsg:"Use at least %d characters"`
```

`errorMsg` replaces the message of every rule of the field without its own message, and is shown as written. `invalidMsg` replaces the message of a value that cannot be read (see [Strict Mapping](#strict-mapping)) or is not valid UTF-8, and `lineMsg` that of `line:"true"`.

Every error is a `form.FieldValidationError` whose `Rule` names the rule that failed, such as `required`, `maxLength`, `pattern`, `values` (also for selects and enums), `invalid` or `line`, or the name of a custom validator. The errors of a nested struct, named such as `address.city`, keep their `Rule`. The messages are for people, `Rule` is for code that tells them apart:

```go
for _, err := range errs {
    if e, ok := err.(form.FieldValidationError); ok && e.Rule == "required" {
        // …
    }
}
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
mapped to its rendered name, and get the validator's name as their `Rule`.

The messages of `RegisterValidationMethod` validators are shown as written. A validator registered with `RegisterLocalizedValidation` gets the localizer, and the form translates the messages it returns, like go-form's own; a `<name>Msg` tag on the field (such as `slugMsg`) replaces them:

```go
f.RegisterLocalizedValidation("slug", func(v any, field reflect.StructField, loc form.Localizer) form.FieldErrors {
    if s, _ := v.(string); s != "" && !slugPattern.MatchString(s) {
        return form.FieldErrors{form.FieldValidationError{Field: field.Name, Err: "Use lowercase letters, digits and dashes"}}
    }
    return nil
})

type Page struct {
    Slug string `form:"input,text" label:"Address" validate:"slug"`
}
```

Register a rule once and attach it to any field with the `validate` tag, such as a stricter check of e-mail addresses than the email input's own. Its message is a translation key, so each language gets its own text, and the error's `Rule` is the rule's name:

```go
renderer.RegisterLocalizedValidation("mail", func(v any, field reflect.StructField, loc form.Localizer) form.FieldErrors {
    address, _ := v.(string)
    if address == "" || hasMailDomain(address) { // your own check, such as a dot in the domain
        return nil
    }

    return form.FieldErrors{form.FieldValidationError{Field: field.Name, Err: "Enter an address such as name@example.com"}}
})

type SubscribeInput struct {
    Email string `form:"input,email" label:"Email" required:"true" validate:"mail"`
}

type CommentInput struct {
    Email string `form:"input,email" label:"Email" validate:"mail" mailMsg:"Use the address you subscribed with"`
}
```

`Bind` and `ValidateFormLocalized` run it after go-form's own rules, so the visitor sees its message at the field, translated for `loc`; `mailMsg` replaces it for one field. Add the message to your catalogues; `form.Texts` only finds messages in tags.

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
    "en": {"Name": "Name", form.TranslationKeyRequired: "is required"},
    "it": {"Name": "Nome", form.TranslationKeyRequired: "è obbligatorio"},
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

- Labels from a `values` tag and the labels of a struct radio group are text of your application, like field labels, so they are translated. Add `translate:"false"` to a field with a `values` tag to show its labels as written.
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

#### Message Catalogues

go-form's own messages are translation keys with the context `form||`, such as `form||Validation required` (the `TranslationKey…` variables). When nothing translates them, they show in the language of the localizer (`GetLocale()`, such as `nl`, `nl_NL` or `nl-BE`) when go-form has a catalogue for it, and in English otherwise: `form||Validation required` shows as `Dit veld is verplicht.` for a Dutch localizer and `This field is required.` for any language go-form has no catalogue for. Nothing translates them when there is no translation function, or when the function returns a message as it got it (with or without the context), as a translator does for a message missing from its catalogue. The application's own translation wins. [go-translator](https://github.com/donseba/go-translator) shows a missing message as `*key*` by default; configure it with `translator.WithUntranslated(translator.ShowOriginal)` (or `SetUntranslated`) to return it as written.

For another language, translate go-form's messages in your own catalogues: [`locales/translations.pot`](locales/translations.pot) lists them.

#### Extracting Texts

`form.Texts(models...)` returns the texts the translation function gets for forms, each once, in order of appearance, to add to your catalogues: labels (the field name when there is no `label` tag), legends, placeholders, descriptions, option labels that are translated, the submit and cancel texts of form metadata, and the messages per rule (every tag ending in `Msg`, such as `requiredMsg` and `errorMsg`). It takes structs, pointers (a nil pointer stands for the zero value), models made with `WithInfo` and schemas:

```go
texts := form.Texts(SignupInput{}, form.WithInfo(ContactInput{}, info), contactSchema)
```

The options of a select are only known from a value that holds them. go-form's own messages are in its catalogues (see [Message Catalogues](#message-catalogues)).

#### Translating the Required Marker

Required fields show an asterisk and, for screen readers only, the text `(required)`. The theme's `ThemeClasses.VisuallyHidden` hides that text from sight: the built-in themes use Bootstrap's `visually-hidden`, Tailwind's `sr-only` and, in the plain theme, the inline style `templates.VisuallyHiddenStyle`. A theme without a `VisuallyHidden` class renders the class `sr-only` (`templates.VisuallyHiddenClass`), which its CSS hides, and the style of `VisuallyHidden`, if any. That text is the translation key `form||(required)` (`form.TranslationKeyRequiredLabel`). When nothing translates it, it shows like go-form's other messages (see [Message Catalogues](#message-catalogues)): in the localizer's language when go-form has a catalogue for it, such as `(verplicht)` for `nl`, and in English otherwise. The default text of the cancel link (an empty `Info.CancelText`) is the key `Cancel`, which shows as `Annuleren` for `nl` the same way.

---

### Custom Form Attributes
You can set custom HTML attributes on forms (e.g., `hx-post`, `data-*`, etc.) using `form.Info.Attributes` with `WithInfo` or `WithRequestInfo`, or in the template (see [Attributes in the Template](#attributes-in-the-template)):

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

#### Stateless Protection with `http.CrossOriginProtection`

Since Go 1.25 the standard library protects against CSRF without tokens: `http.CrossOriginProtection` refuses cross-origin form posts using the `Sec-Fetch-Site` and `Origin` headers that browsers send. It keeps no state, so it works across application instances without a shared store. Wrap your handlers with it and leave out the CSRF middleware; `WithRequestInfo` then adds no token, and forms render without one:

```go
protection := http.NewCrossOriginProtection()
mux.Handle("/", protection.Handler(yourHandler))
```

Use go-form's token middleware below for browsers that send neither header, or when you need tokens for another reason.

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

Models that embed `form.Info` as their first field remain supported. For those models,
`InjectCSRFToken` can still populate the embedded metadata directly.

The middleware automatically:
- Generates a secure random token for each form
- Validates the token on submission
- Refreshes tokens after each submission
- Renews the session cookie with every token it issues
- Rejects requests with missing or invalid tokens

Each issued token is stored separately and bound to its session. Loading another
page, an HTMX fragment, or a second form does not invalidate earlier forms.
Submitting a form consumes only its token and supplies a fresh token to the
handler, so forms in other tabs remain usable until their tokens expire.

The middleware keeps its session in the cookie `session_id` (`csrf.DefaultSessionID`)
and reads the token from the field `_csrf` (`form.DefaultCSRFField`). Change these
variables at startup when your application uses the same names. `CSRFOptions.FieldName`
sets the field for one middleware:

```go
options := form.DefaultCSRFOptions()
options.FieldName = "token"
protectedHandler := formRenderer.CSRFMiddlewareWithOptions(options)(yourHandler)
```

Forms that get their token from `WithRequestInfo` or `WithContextInfo` render it in
that field. A form that sets `Info.CsrfField` must name the same field.

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
      http.Error(w, "CSRF token not found", http.StatusForbidden)
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

The default store keeps a token for `csrf.DefaultExpirationTime` (10 minutes);
`formRenderer.SetCSRFStore(csrf.NewDefaultMemoryCSRFStore(30*time.Minute))` keeps
tokens longer. The session cookie lasts as long as a token: every page that gets a
token and every accepted submission renew it, for the lifetime of the store's
tokens. A store reports that lifetime with `Lifetime() time.Duration`
(`csrf.ExpiringStore`); for other stores the cookie lasts `csrf.DefaultExpirationTime`.

Implement `csrf.Store` with `Store` and `Validate` methods. Treat the key as an
opaque identifier for a session-bound token, rather than the raw session ID.
Existing forms must be reloaded when deploying the change from session-only
storage keys.

Stores may also implement `csrf.TokenConsumer` with `Consume(key, token string)`
to validate and remove a token atomically. Both built-in memory stores do this.
Shared stores should implement it with a transaction or equivalent atomic
operation to reject simultaneous reuse across application instances. Stores
that only implement `Store` retain sequential one-use behavior through
validation followed by token invalidation. A store whose tokens last another time
than `csrf.DefaultExpirationTime` implements `csrf.ExpiringStore`, so that the
session cookie lasts as long.

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

`MapForm` sets a select from the posted key and a multi-select from the posted keys. A key that is not one of the options leaves the selection empty, so `required:"true"` reports it, and with `MapOptions{Strict: true}` it is a field error (see [Strict Mapping](#strict-mapping)). `MapForm` does not abort or write output for it, and maps the other fields. A select without a source has no options to read a key with: give it its source before mapping. To see such skipped values (also malformed dates), set a handler:

```go
form.MapFormErrorHandler = func(field string, err error) {
    slog.Debug("form value ignored", "field", field, "error", err)
}
```

### Ordered Selects and Placeholders

`SortedSelect` lists its options by key. To list them otherwise:

```go
// In your own order, such as pages in tree order.
parent := form.NewOrderedSelect(
    form.Option[int64]{Value: 1, Label: "Home"},
    form.Option[int64]{Value: 4, Label: "— About"},
).Selected(page.ParentID)

// Alphabetically by label (ignoring case), such as files by name.
file := form.NewSortedSelect(files).SortByLabel().Selected(current)
```

With `translate:"true"`, `SortByLabel` sorts the options again when the form renders, by their translated labels, so they are in alphabetical order in every language:

```go
// The labels are translation keys, listed in the order of their translations.
Module form.SortedSelect[string] `form:"dropdown" label:"Module" translate:"true"`

input.Module = form.NewSortedSelect(titles).SortByLabel().Selected(current)
```

`Selected(v)` selects a value and returns the select, for chaining; a value that is not an option leaves nothing selected, as `Set` does. `NewOrderedMultiSelect`, `SortByLabel` and `Selected(vs...)` work the same for `SortedMultiSelect`. `SetSource` with a map sorts the options by key again.

A `placeholder` tag on a dropdown renders an empty first option, such as `placeholder:"—"`, so no options need a `""` key. Submitting it leaves the select empty, so `required:"true"` reports it.

For advanced usage, see [`example/sortedselect/main.go`](example/sortedselect/main.go) — covers single and multi-select fields, supported key types, pre-filled and user-submitted values, validation, error handling and JSON output.

---

## Forms Defined at Runtime

A form put together at runtime, such as a contact form an editor builds or the settings a theme declares, is a `form.Schema`. It renders through the same themes and checks the same rules as a struct with tags:

```go
schema := form.Schema{Fields: []form.Spec{
    {Name: "name", Label: "Name", Required: true, MaxLength: 100, AutoComplete: "name"},
    {Name: "email", Label: "Email", InputType: types.InputFieldTypeEmail, Required: true},
    {Name: "topic", Label: "Topic", Type: types.FieldTypeDropdown, Placeholder: "—", Options: []types.FieldValue{
        {Value: "sales", Name: "Sales"}, {Value: "help", Name: "Help"},
    }},
    {Name: "code", Label: "Code", Pattern: "^[0-9]{4}$", Messages: map[string]string{"pattern": "Enter four digits"}},
    {Name: "consent", Label: "I agree", Type: types.FieldTypeCheckbox, Required: true},
}}

values := form.MapValues(r, schema)                          // url.Values of the schema's fields, trimmed
errs := renderer.ValidateSchemaLocalized(schema, values, loc) // form.FieldErrors by field name
model := form.WithRequestInfo(r, schema.Bind(values), form.Info{Target: "/contact", Method: "post"})
```

`model` renders with `form_render` or field by field with `form_field`. A `Spec` has the rules of the tags: `Required`, `MinLength`, `MaxLength`, `Min`, `Max`, `Step`, `Pattern`, `URL`, `SingleLine` and, for email inputs, the email check. Numbers must be finite (`NaN` and `Inf` are refused) and text valid UTF-8. Values of dropdowns, radios and multicheckboxes must be one of the `Options`; numbers and dates must parse. `Validate` names [registered rules](#validation), separated by commas like the `validate` tag, such as `Validate: "mail"`; they check the first submitted value as a string, after go-form's own rules. `Messages` replace go-form's messages per rule, like the `…Msg` tags, also those of registered rules.

---

## Upgrading

### From v2.5.0

The module path stays `github.com/donseba/go-form/v2`. Only struct literals that list their fields without names stop compiling; the other changes are in behaviour, grouped by area below. Check the first group before you deploy: those changes can lose submitted values without an error.

**Changes that break compilation**

- Struct literals that list their fields without names no longer compile for the structs with new fields: `FieldValidationError` (`Rule`, the rule that failed), `Info`, `RenderModel`, `templates.ThemeClasses` (`Fieldset` and `Legend` come right after `Label`), `types.FormField` and `Transformer`. Name the fields, such as `form.FieldValidationError{Field: "email", Err: message}`; custom validators that return `FieldValidationError{"email", message}` are the most likely to break.

**Values that can be lost without an error**

- On a POST, `MapForm` sets a slice of strings or integers (`[]string`, `[]int`, …) to `nil` when the request has no value for it, as browsers send nothing for a multicheckbox without ticked boxes; v2.5.0 did not map such slices and left them as they were. `MapForm` maps every exported field of the struct, so a slice field that the form does not render is cleared on every POST. Map into a struct with only the form's fields, or set such fields after mapping.
- A `SortedMultiSelect` that gets a key it does not offer is left with nothing selected, and `MapForm` maps the other fields, as for a `SortedSelect`. v2.5.0's `MapForm` returned an error and left the fields after it unmapped, also on every POST to a multi-select without a source. A handler that answered the error with 400 now gets an empty selection, which `required:"true"` and `minItems` report; map with `MapOptions{Strict: true}` (or `Bind`) to get the unknown key as a field error with the rule `invalid`.
- `disabled:"true"` now renders the `disabled` attribute on every control (inputs, hidden inputs, textareas, selects, checkboxes, radios, multicheckboxes and range inputs), where v2.5.0 only rendered it on buttons, and browsers do not submit disabled controls. After `MapForm` a disabled field keeps the value the struct had before mapping, a ticked disabled checkbox becomes `false`, and on a POST a disabled multicheckbox becomes empty. Fill the struct from the stored record before mapping, or use `readonly:"true"` for a text field whose value must be sent.

**Messages and translations**

- go-form's own messages show their English text when nothing translates them, such as `This field is required.` and `Enter a valid email address.`; v2.5.0 showed the key, such as `form||Validation required`. Nothing translates a message when there is no translation function, or when the function returns it as it got it, with or without `form||`. Tests that compare messages need the new texts.
- These texts follow the localizer's language (`GetLocale()`) when go-form has a catalogue for it: Dutch for `nl`, `nl_NL`, `nl-BE`, …, also for the default cancel text (`Annuleren`) and the screen-reader text of required fields (`(verplicht)`). Your own translations still win. Tests that render with a Dutch localizer and compare English texts need the Dutch ones.
- A translation that equals the key without `form||`, such as `Validation required` for `form||Validation required`, counts as untranslated and shows go-form's text (`This field is required.`). Give such messages a translation of their own, or remove them from your catalogues.
- The keys of `pattern` and `step` changed: `form.TranslationKeyPattern` is `form||Value is not in the expected format`, without the expression, and `form.TranslationKeyStep` is `form||Value should be a multiple of %g` (`%g` instead of `%f`). Existing forms can also show the new `form.TranslationKeyStepFromMin` and `form.TranslationKeyInvalid` (see Validation below). Update and add their translations; go-form has English and Dutch texts for all four.

**Validation**

- `required:"true"` treats a string of only spaces as empty, and also a `*string` that points to `""` or spaces; v2.5.0 only refused `""` and a nil pointer. `minLength` no longer counts spaces at the start and end. Both take password inputs as typed.
- The rules check the value a pointer to a string, bool or number points to (`min`, `max`, `step`, `maxLength`, `pattern`, …); v2.5.0 skipped pointer fields. `min`, `max` and `step` apply to every int, uint and float kind, not only `int` and `float64`.
- `step` counts exactly, from `min` when the field has one and from zero otherwise. v2.5.0 divided floating-point numbers and always counted from zero, so `0.3` failed `step:"0.1"`, and with `min:"0.25" step:"0.5"` it accepted `1` and refused `0.75`; now `0.3` and `0.75` pass and `1` fails. When `min` is not itself a whole number of steps from zero, the message is the new `form.TranslationKeyStepFromMin` (`form||Value should be in steps of %g from %g`) instead of `form.TranslationKeyStep`.
- The new rule `invalid` refuses a float field holding `NaN` or an infinity, and a string field that is not valid UTF-8, whatever their other tags (`form.TranslationKeyInvalid`); v2.5.0 let them through.
- Email inputs (`form:"input,email"`) are checked with `net/mail`: a plain address of at most 254 characters, without a display name; v2.5.0 only looked for an `@`. `url:"true"` refuses `javascript:`, `vbscript:` and `data:` URLs, and URLs with spaces.
- Every `url` tag other than `""` and `"false"` now checks the value; v2.5.0 only checked `url:"true"`. `url:"link"` and lists of schemes such as `url:"http,https"` have rules of their own (see [Built-in Validation](#built-in-validation)), and any other value is read as a list of schemes.
- A `values` tag works on integer fields, where a dropdown failed to render (`incompatible types for comparison`). Validation checks integer fields and slices of strings or integers (every ticked value) against the options, not only string fields; `0` counts as nothing chosen unless it is an option.
- Validation no longer treats selects (`SortedSelect`, `SortedMultiSelect`, `Mapper`, …), uploaded files and unexported struct fields as nested forms. The `validate` tag of a select or file field now runs, and an unexported struct field no longer panics.
- `minItems` and `maxItems` also count the selected options of a `SortedMultiSelect`; v2.5.0 only counted slices and arrays, and ignored these tags on a multi-select.
- `ValidateForm` and `ValidateFormLocalized` validate the model of a `RenderModel`, with its `Overrides` and `Options`; v2.5.0 found no errors in a `*RenderModel` and panicked on a `RenderModel`.

**Mapping (`MapForm`)**

- `MapForm` parses the request before mapping, so a bool field before the first other field is mapped too: `true` when ticked, `false` when not sent. v2.5.0 left such a field as it was unless the handler had parsed the request.
- `MapForm` maps pointers to strings and numbers (an empty value sets `nil`), unsigned integers, slices of strings and integers, and uploaded files (`*multipart.FileHeader`, `[]*multipart.FileHeader`). v2.5.0 skipped these fields, so they now change when the form sends them.
- A number that does not fit the field (`300` for an `int8`, a float too large for a `float32`), `NaN` and infinities leave the field as it was; v2.5.0 wrapped integers around and set the others. A malformed UUID leaves the field as it was instead of setting part of it.
- More skipped values reach `MapFormErrorHandler`: numbers that do not parse, do not fit or are not finite, checkbox values that are not a boolean (such as `yes`), and malformed UUIDs; v2.5.0 skipped them without calling it. With the new `MapOptions{Strict: true}` they come back as field errors (see [Strict Mapping](#strict-mapping)).
- `MapForm` maps struct radio groups (see [Supported Input Fields & Options](#supported-input-fields--options)): the bool field of the chosen option becomes `true` and the others `false`. v2.5.0 set every option to `false`. A `form:"radios"` group with no option chosen renders, where v2.5.0 failed with `incompatible types for comparison`.
- Time inputs accept `15:04`, as browsers send it, and `datetime-local` inputs accept seconds (`2006-01-02T15:04:05`); v2.5.0 refused both. Month inputs map and render as `2006-01` instead of `01`, which lost the year. A `datetime-local` value renders with seconds when the field's `step` is not a whole number of minutes.

**Markup and themes**

- Elements without classes or inline styles render without `class` and `style` attributes, instead of `class=" "` or `style=""`.
- `maxLength` and `minLength` render as `maxlength` and `minlength`, and anchored `pattern` tags as `pattern`, so browsers check them before submitting.
- A field with errors renders its control with `aria-invalid="true"`, and its `aria-describedby` names the error messages (`<id>_error_0`, …) besides the description; the messages get those ids.
- Hidden inputs render with their `id`, `name` and `value` only (and `disabled` and `data-*` attributes), without `placeholder=""` and an `aria-labelledby` that pointed to a label that does not exist.
- A `placeholder` tag on a dropdown renders an empty first option; it used to be ignored.
- `form:"radios"` with a `values` tag renders a group of radio buttons, one per option, instead of a single radio button, and `form:"multicheckbox"` with a `values` tag a checkbox per option instead of a dropdown. Values without a label (`values:"a; b"`) lose surrounding spaces, as validation already read them.
- Groups of radio buttons (`form:"radios"` with options, and struct radio groups) and multicheckboxes render as a `<fieldset>` with the label as its `<legend>` (`fieldset.gohtml`, `legend.gohtml`), instead of a `<div>` with a `<label for>` that pointed to no element; the inner `role="radiogroup"` is gone. The fieldset gets the `Wrapper` classes plus the new `ThemeClasses.Fieldset` and the legend the `Label` classes plus `ThemeClasses.Legend`, so CSS that styles a field's `label` needs `legend` too, and the fieldset needs its border, padding and margin reset where the theme's CSS has no such reset (Bootstrap's reboot and Tailwind's preflight have one). A theme whose templates all come from its own folder (`LoadTemplates`, `LoadTemplatesFS`) without `fieldset.gohtml` keeps the old markup.
- `Info.Attributes["class"]` and `["style"]` merge into the theme's `class` and `style` of the form tag, instead of rendering a second `class` or `style` attribute. A theme with its own `form.gohtml` no longer finds them in `.Field.Attributes`, so it loses them unless it renders `form_open_class` and `form_open_style`, as go-form's own `form-open.gohtml` does.
- A form with a file input gets `enctype="multipart/form-data"`, unless `Info.Attributes` sets another `enctype`.
- The `(required)` for screen readers is hidden by the theme's new `ThemeClasses.VisuallyHidden`: Bootstrap's `visually-hidden` and, in the plain theme, an inline style. Both themes rendered the class `sr-only`, which they do not define, so the text showed on screen. The Tailwind themes and themes without `VisuallyHidden` keep `sr-only`.
- Form attributes and `data-*` attributes render sorted by name, instead of in random order.
- A nil pointer to a string, bool or number renders an empty input instead of `<nil>`.
- Unexported fields of a model no longer render; rendering a model with one used to panic.

**CSRF protection**

- The middleware reads the token from the field `CSRFOptions.FieldName` (by default `form.DefaultCSRFField`), and forms that get their token from `WithRequestInfo` or `WithContextInfo` render it in that field. In v2.5.0 a form with another `Info.CsrfField` failed every submission; set the same name in `CSRFOptions.FieldName`.
- The middleware renews its session cookie (`session_id`) with every token it issues, on pages and accepted submissions, and keeps it for the lifetime of the store's tokens (`csrf.ExpiringStore`; 10 minutes for the default store, or what `NewDefaultMemoryCSRFStore` was given). v2.5.0 set the cookie once, for `csrf.DefaultExpirationTime`, so a form loaded late in the cookie's life failed when it was submitted, and a longer token lifetime did not lengthen the cookie. Every response the middleware lets through now sets the cookie.

**Other**

- `FuncMap()` has 11 new functions: `form_field`, `form_field_localized`, `form_errors`, `form_open`, `form_open_localized`, `form_buttons`, `form_buttons_localized`, `form_close`, `form_hidden`, `form_hidden_localized` and `form_honeypot`. Where you merge it with template functions of your own, functions with the same name replace each other; rename yours.
- A localizer with a `Location() *time.Location` method is now a `form.TimeLocalizer`: rendering with it shows `datetime-local` values in that zone, while `MapForm` still reads them as UTC. Map such forms with `form.MapFormWith(r, &input, form.MapOptions{Location: …})` in the same zone (see [Time Zones](#time-zones)), or tag fields that hold a time on the clock with `wallclock:"true"`.

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

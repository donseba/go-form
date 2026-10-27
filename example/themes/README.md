# Theme-based example (gohtml)

This example showcases the new theme approach:

- themes are defined as token sets (`templates.ThemeClasses`)
- markup is shared via embedded `.gohtml` templates (`templates/gohtml/*.gohtml`)
- the cancel action is rendered when `CancelTarget` is set
- file inputs are supported (`form:"input,file"`) and rendered without a `value` attribute

## Run

```bash
go run ./example/themes
```

Open:

- http://localhost:8084/?theme=bootstrap
- http://localhost:8084/?theme=tailwind
- http://localhost:8084/?theme=tailwindv4
- http://localhost:8084/?theme=plain

## Notes

- The example makes a renderer per theme at start-up, each with its own page template, and picks one
  per request. An application selects its theme once, before it serves requests: `SetTheme` on a
  renderer that is rendering for other requests races with them.
- A form with a file input gets `enctype="multipart/form-data"` automatically; `MapForm` maps uploads to `*multipart.FileHeader` fields.
- The `form` package puts the theme's `label`, control, `wrapper` and `form` templates together, so
  this example renders the whole struct with one call: `{{ form_render .Form .Errors }}`. To lay out
  the fields yourself, render them one at a time with `form_open`, `form_field`, `form_buttons` and
  `form_close` (see [Laying Out Fields Yourself](../../README.md#laying-out-fields-yourself)).

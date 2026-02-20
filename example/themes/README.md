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

- For file uploads the form must use `enctype="multipart/form-data"`.
- This example keeps rendering explicit (it renders `label`, `input`, `wrapper`, and then `form`).
  The next step would be to move this wiring into the `form` package (e.g. a `ThemeRenderer`) so
  users can render a full struct with one call, like they do with the TemplateMap approach.

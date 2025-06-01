package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/donseba/go-form"
	"github.com/donseba/go-form/templates"
)

// Dummy Localizer implementation
// In a real app, use your go-translator Localizer

type MyLocalizer struct {
	Locale string
}

func (l MyLocalizer) GetLocale() string { return l.Locale }

// Dummy translation data
var translations = map[string]map[string]string{
	"en": {
		"Name":                     "Name",
		"Age":                      "Age",
		"Submit":                   "Submit",
		"form.validation.required": "is required",
		"form.validation.min":      "must be >= %v",
		"form.validation.max":      "must be <= %v",
	},
	"it": {
		"Name":                     "Nome",
		"Age":                      "Età",
		"Submit":                   "Invia",
		"form.validation.required": "è obbligatorio",
		"form.validation.min":      "deve essere piu di %v",
		"form.validation.max":      "deve essere meno di %v",
	},
}

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

type TranslationForm struct {
	form.Info
	Name string `form:"input,text" label:"Name" required:"true"`
	Age  int    `form:"input,number" label:"Age" required:"true" min:"10" max:"120"`
}

func main() {
	f := form.NewTranslatedForm(templates.Plain, myTranslate)

	// Add navigation links to other examples
	htmlLinks := `<div style="margin-bottom:1em">
		<a href=\"http://localhost:8000/\">Templates Example</a> |
		<a href=\"http://localhost:8081/\">Validation Example</a>
	</div>`

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		locale := r.URL.Query().Get("locale")
		if locale == "" {
			locale = "en"
		}

		loc := MyLocalizer{Locale: locale}
		fmt.Println("Locale set to:", loc.GetLocale())

		data := TranslationForm{
			Info: form.Info{
				Target:     r.URL.String(),
				Method:     http.MethodPost,
				SubmitText: "Submit",
			},
		}
		errList := form.FieldErrors{}

		if r.Method == http.MethodPost {
			err := r.ParseForm()
			if err != nil {
				http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
				return
			}
			data.Name = r.FormValue("Name")
			data.Age, _ = strconv.Atoi(r.FormValue("Age"))

			errList = f.ValidateForm(&data, loc)
		}

		tmpl := template.Must(template.New("page").Funcs(f.FuncMap()).Parse(`
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
				<script src="https://cdn.tailwindcss.com"></script>
				<meta charset="UTF-8"><title>Form with translations</title>
			</head>
			<body>
			` + htmlLinks + `
			<h2>Translation Example</h2>
			<div>
				<a href="?locale=en">English</a> | <a href="?locale=it">Italian</a>
			</div>

			{{ form_render_localized .Loc .Form .Errors }}
			</body>
			</html>
		`))
		err := tmpl.Execute(w, map[string]any{"Form": data, "Errors": errList, "Loc": loc})
		if err != nil {
			http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Change port to 8082
	fmt.Println("Server running at http://localhost:8082/?locale=en or ?locale=it")
	log.Fatal(http.ListenAndServe(":8082", nil))
}

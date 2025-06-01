package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"reflect"

	"github.com/donseba/go-form"
	"github.com/donseba/go-form/templates"
)

// Custom validation: checks if a string is a valid hex color (e.g. #aabbcc)
func isHexColor(val any, field reflect.StructField) (out form.FieldErrors) {
	s, ok := val.(string)
	if !ok || s == "" {
		return nil
	}
	if len(s) != 7 || s[0] != '#' {
		out = append(out, form.FieldValidationError{
			Field: field.Name,
			Err:   "must be a valid hex color (e.g. #aabbcc)",
		})

		return out
	}
	for _, c := range s[1:] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			out = append(out, form.FieldValidationError{
				Field: field.Name,
				Err:   "must be a valid hex color (e.g. #aabbcc)",
			})
			return out
		}
	}

	return
}

type CustomForm struct {
	form.Info
	Name  string `form:"input,text" label:"Name" required:"true" minLength:"2" maxLength:"20"`
	Color string `form:"input,text" label:"Favorite Color (hex)" validate:"isHexColor"`
}

func main() {
	f := form.NewForm(templates.BootstrapV5)
	f.RegisterValidationMethod("isHexColor", isHexColor)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := CustomForm{
			Info: form.Info{
				Target: r.URL.Path,
				Method: http.MethodPost,
			},
		}

		var errs form.FieldErrors
		if r.Method == http.MethodPost {
			err := r.ParseForm()
			if err != nil {
				http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
				return
			}
			data.Name = r.FormValue("Name")
			data.Color = r.FormValue("Color")
			errs = f.ValidateForm(&data, nil)
		}

		tmpl := template.Must(template.New("page").Funcs(f.FuncMap()).Parse(`
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8"><title>Custom Validation Example</title>
    			<link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
			</head>
			<body>

			<h2>Custom Validation Example</h2>
			<div style="margin-bottom:1em">
				<a href="http://localhost:8080/">Templates Example</a> |
				<a href="http://localhost:8082/">Translation Example</a>
			</div>
			{{ form_render .Form .Errors }}

			</body>
			</html>
		`))

		err := tmpl.Execute(w, map[string]any{
			"Form":   data,
			"Errors": errs,
		})
		if err != nil {
			http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Change port to 8081
	fmt.Println("Server running at http://localhost:8081/")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

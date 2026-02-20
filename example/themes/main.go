package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/donseba/go-form"
)

// ThemeExampleForm showcases:
// - theme-based gohtml templates (ThemeClasses + embedded templates)
// - cancel button rendering
// - file input rendering
//
// NOTE: file uploads require multipart/form-data.

type ThemeExampleForm struct {
	form.Info

	Name      string                         `form:"input,text" label:"Name" placeholder:"Jane" required:"true"`
	Email     string                         `form:"input,email" label:"Email" placeholder:"jane@example.com" required:"true"`
	Upload    string                         `form:"input,file" label:"Profile picture" description:"Choose an image to upload"`
	Interests form.SortedMultiSelect[string] `form:"multicheckbox" label:"Interests" description:"Select your interests"`
	Languages form.SortedMultiSelect[string] `form:"multicheckbox" label:"Languages" description:"Select languages you know"`
}

func main() {
	f := form.NewForm()

	page := template.Must(template.New("page").Funcs(f.FuncMap()).Parse(`
<!doctype html>
<html lang="en">
<head>
 <meta charset="utf-8" />
 <meta name="viewport" content="width=device-width, initial-scale=1" />
 <title>go-form theme example</title>

 <!-- Bootstrap 5 (for the bootstrap theme) -->
 <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">

 <!-- Tailwind (if you switch to tailwind/tailwindv4 theme, this gives you utilities quickly) -->
 <script src="https://cdn.tailwindcss.com"></script>
</head>
<body>
 <div class="container py-4">
   <h1 class="mb-4">Theme-based rendering (gohtml)</h1>

   <div class="mb-3">
     <a href="/?theme=bootstrap">bootstrap</a> |
     <a href="/?theme=tailwind">tailwind</a> |
     <a href="/?theme=tailwindv4">tailwindv4</a> |
     <a href="/?theme=plain">plain</a>
   </div>

   {{ form_render .Form .Errors }}

   {{ if .Posted }}
     <hr />
     <h2>Posted values</h2>
     <pre>{{ .Debug }}</pre>
   {{ end }}
 </div>
</body>
</html>
`))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		themeName := r.URL.Query().Get("theme")
		if themeName == "" {
			themeName = "bootstrap"
		}
		f.SetTheme(themeName)

		data := ThemeExampleForm{
			Info: form.Info{
				Target:       "/?theme=" + themeName,
				Method:       http.MethodPost,
				SubmitText:   "Save",
				CancelTarget: "/?theme=" + themeName,
				CancelText:   "Cancel",
				Attributes: map[string]string{
					"enctype": "multipart/form-data",
				},
			},
		}
		data.Languages = form.NewSortedMultiSelect(map[string]string{
			"Go":         "Go",
			"Python":     "Python",
			"JavaScript": "JavaScript",
			"Rust":       "Rust",
		})
		data.Interests = form.NewSortedMultiSelect(map[string]string{
			"Reading":   "Reading",
			"Traveling": "Traveling",
			"Coding":    "Coding",
			"Music":     "Music",
		})

		posted := r.Method == http.MethodPost
		debug := ""
		var errs form.FieldErrors

		if posted {
			if err := r.ParseMultipartForm(10 << 20); err != nil {
				http.Error(w, "ParseMultipartForm: "+err.Error(), http.StatusBadRequest)
				return
			}
			_ = form.MapForm(r, &data)
			errList := f.ValidateForm(&data)
			errs = errList

			debug = fmt.Sprintf("Name=%q\nEmail=%q\nInterests=%q\nLanguages=%q\n", data.Name, data.Email, data.Interests, data.Languages)
			if fh, _, err := r.FormFile("Upload"); err == nil {
				_ = fh.Close()
				debug += "Upload=<received file>\n"
			} else {
				debug += "Upload=<no file>\n"
			}
		}

		if err := page.Execute(w, map[string]any{
			"Form":   data,
			"Errors": errs,
			"Posted": posted,
			"Debug":  strings.TrimSpace(debug),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	fmt.Println("Theme example running at http://localhost:8084/?theme=bootstrap")
	log.Fatal(http.ListenAndServe(":8084", nil))
}

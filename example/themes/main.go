package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/donseba/go-form/v2"
)

// ThemeExampleForm showcases:
// - theme-based gohtml templates (ThemeClasses + embedded templates)
// - cancel button rendering
// - file input rendering
//
// NOTE: file uploads require multipart/form-data.

type ThemeExampleForm struct {
	form.Info

	Prefix    string                         `form:"dropdown" label:"Prefix" description:"Select a prefix for your name" values:"mr:MR;ms:MS" required:"true"`
	Name      string                         `form:"input,text" label:"Name" placeholder:"Jane" required:"true"`
	Email     string                         `form:"input,email" label:"Email" placeholder:"jane@example.com" required:"true"`
	Upload    string                         `form:"input,file" label:"Profile picture" description:"Choose an image to upload"`
	Interests form.SortedMultiSelect[string] `form:"multicheckbox" label:"Interests" description:"Select your interests"`
	Languages form.SortedMultiSelect[string] `form:"multicheckbox" label:"Languages" description:"Select languages you know"`
}

// themes are the themes the example switches between with ?theme=; the
// first is the default.
var themes = []string{"bootstrap", "tailwind", "tailwindv4", "plain"}

const pageTemplate = `
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
`

// themePage is the page of one theme: a renderer of its own, whose theme
// is set before it serves requests, and the page template with that
// renderer's functions. Calling SetTheme on a shared renderer per request
// would race with the requests that render.
type themePage struct {
	name     string
	renderer *form.Form
	page     *template.Template
}

func main() {
	fmt.Println("Theme example running at http://localhost:8084/?theme=bootstrap")
	log.Fatal(http.ListenAndServe(":8084", newHandler()))
}

// newHandler returns the handler of the example, with a page per theme,
// made at start-up.
func newHandler() http.Handler {
	pages := make(map[string]themePage, len(themes))

	for _, name := range themes {
		renderer := form.NewForm()
		renderer.SetTheme(name)

		pages[name] = themePage{
			name:     name,
			renderer: renderer,
			page:     template.Must(template.New("page").Funcs(renderer.FuncMap()).Parse(pageTemplate)),
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, ok := pages[r.URL.Query().Get("theme")]
		if !ok {
			page = pages[themes[0]]
		}

		page.ServeHTTP(w, r)
	})
}

// ServeHTTP shows the form in the page's theme and, after a submission,
// the posted values.
func (p themePage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data := ThemeExampleForm{
		Info: form.Info{
			Target:       "/?theme=" + p.name,
			Method:       http.MethodPost,
			SubmitText:   "Save",
			CancelTarget: "/?theme=" + p.name,
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
		errs = p.renderer.ValidateForm(&data)

		debug = fmt.Sprintf(
			"Prefix=%q\nName=%q\nEmail=%q\nInterests=%q\nLanguages=%q\n",
			data.Prefix, data.Name, data.Email, data.Interests, data.Languages,
		)

		if fh, _, err := r.FormFile("Upload"); err == nil {
			_ = fh.Close()
			debug += "Upload=<received file>\n"
		} else {
			debug += "Upload=<no file>\n"
		}
	}

	if err := p.page.Execute(w, map[string]any{
		"Form":   data,
		"Errors": errs,
		"Posted": posted,
		"Debug":  strings.TrimSpace(debug),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/donseba/go-form"
)

type DemoForm struct {
	form.Info
	DepartmentID     form.SortedSelect[int64]       `form:"dropdown" label:"Department (single)"`
	DepartmentsMulti form.SortedMultiSelect[int64]  `form:"multicheckbox" label:"Departments (multi)"`
	ColorID          form.SortedSelect[string]      `form:"dropdown" label:"Color (single)"`
	ColorsMulti      form.SortedMultiSelect[string] `form:"multicheckbox" label:"Colors (multi, pre-filled)"`
	PriceID          form.SortedSelect[float64]     `form:"dropdown" label:"Price (single)"`
	UUIDField        form.SortedSelect[uuid.UUID]   `form:"dropdown" label:"UUID (single)" required:"true"`
	TimeField        form.SortedSelect[time.Time]   `form:"dropdown" label:"Time (single)"`
}

func main() {
	f := form.NewForm()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		u1 := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
		u2 := uuid.MustParse("123e4567-e89b-12d3-a456-426614174001")
		t1, _ := time.Parse(time.RFC3339, "2024-01-01T12:00:00Z")
		t2 := t1.Add(time.Hour)

		data := DemoForm{
			Info: form.Info{
				Target:     r.URL.Path,
				Method:     http.MethodPost,
				SubmitText: "Submit",
				Attributes: map[string]string{"custom-attr": "value"},
			},
			DepartmentID: form.NewSortedSelect(map[int64]string{
				1: "Department 1",
				2: "Department 2",
				3: "Department 3",
			}),
			DepartmentsMulti: form.NewSortedMultiSelect(map[int64]string{
				1: "Department 1",
				2: "Department 2",
				3: "Department 3",
			}),
			ColorID: form.NewSortedSelect(map[string]string{
				"r": "Red",
				"g": "Green",
				"b": "Blue",
			}),
			ColorsMulti: form.NewSortedMultiSelect(map[string]string{
				"r": "Red",
				"g": "Green",
				"b": "Blue",
			}),
			PriceID: form.NewSortedSelect(map[float64]string{
				1.99: "Cheap",
				5.49: "Medium",
				9.99: "Expensive",
			}),
			UUIDField: form.NewSortedSelect(map[uuid.UUID]string{
				uuid.Nil: "Select...",
				u1:       "First",
				u2:       "Second",
			}),
			TimeField: form.NewSortedSelect(map[time.Time]string{
				t1: "Now",
				t2: "Later",
			}),
		}

		// Only pre-fill on GET, not POST
		if r.Method != http.MethodPost {
			_ = data.DepartmentID.Set(3)
			_ = data.ColorID.Set("g")
			_ = data.PriceID.Set(5.49)
			_ = data.UUIDField.Set(uuid.Nil)
			_ = data.TimeField.Set(t2)
			_ = data.ColorsMulti.Set([]string{"r", "b"}) // Pre-fill multi-select
			// DepartmentsMulti left empty for demo
		}

		var errs form.FieldErrors
		var submitted DemoForm
		var submittedJSON string
		if r.Method == http.MethodPost {
			submitted = data // copy for display
			err := form.MapForm(r, &data)
			if err != nil {
				http.Error(w, "Error mapping form: "+err.Error(), http.StatusBadRequest)
				return
			}
			errs = f.ValidateForm(&data)
			submitted = data // show submitted values
			if b, err := json.MarshalIndent(submitted, "", "  "); err == nil {
				submittedJSON = string(b)
			} else {
				submittedJSON = err.Error()
			}
		}

		tmpl := template.Must(template.New("page").Funcs(f.FuncMap()).Parse(`
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8"><title>SortedSelect & MultiSelect Demo</title>
    			<link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
			</head>
			<body>
			<div class="container mt-4">
				<h2>SortedSelect & MultiSelect Demo</h2>
				<p>This demo shows both single and multi-select fields using SortedSelect and SortedMultiSelect.</p>
				{{ form_render .Form .Errors }}
				{{ if .Submitted }}
				<hr>
				<h4>Submitted Values</h4>
				<pre><code>{{ .SubmittedJSON }}</code></pre>
				{{ end }}
			</div>
			</body>
			</html>
		`))

		err := tmpl.Execute(w, map[string]any{
			"Form":          data,
			"Errors":        errs,
			"Submitted":     submitted,
			"SubmittedJSON": submittedJSON,
		})
		if err != nil {
			http.Error(w, "Error rendering template: "+err.Error(), http.StatusInternalServerError)
			return
		}
	})

	fmt.Println("Server running at http://localhost:8090/")
	log.Fatal(http.ListenAndServe(":8090", nil))
}

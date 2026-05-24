package form

import (
	"html/template"
	"io"
	"testing"
)

func TestForm_Render_UsesGoHTMLTheme(t *testing.T) {
	type Simple struct {
		Info
		Name string `form:"input,text" label:"Name" required:"true"`
	}

	f := NewForm()
	f.SetTheme("bootstrap")
	data := Simple{Info: Info{Target: "/", Method: "POST", SubmitText: "Save"}}

	tmpl := template.Must(template.New("t").Funcs(f.FuncMap()).Parse(`{{ form_render . nil }}`))
	if err := tmpl.Execute(io.Discard, data); err != nil {
		t.Fatalf("execute: %v", err)
	}
}

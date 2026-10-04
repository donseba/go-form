package form

import (
	"html/template"
	"io"
	"strings"
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

func TestForm_Render_SingleCheckboxHasOneLabel(t *testing.T) {
	type Terms struct {
		Info
		Accept bool `form:"checkbox" label:"Accept Terms"`
	}

	for _, theme := range []string{"bootstrap", "tailwind", "plain"} {
		f := NewForm()
		f.SetTheme(theme)
		data := Terms{Info: Info{Target: "/", Method: "POST", SubmitText: "Save"}}

		var out strings.Builder
		tmpl := template.Must(template.New("t").Funcs(f.FuncMap()).Parse(`{{ form_render . nil }}`))
		if err := tmpl.Execute(&out, data); err != nil {
			t.Fatalf("%s: execute: %v", theme, err)
		}

		if got := strings.Count(out.String(), "Accept Terms"); got != 1 {
			t.Fatalf("%s: want the checkbox label once, got %d times:\n%s", theme, got, out.String())
		}
	}
}

package form

import (
	"strings"
	"testing"
)

type attributesModel struct {
	Email    string `form:"input,email"    label:"Email"    autocomplete:"email"            maxLength:"254" minLength:"3" description:"Your address"`
	Code     string `form:"input,text"     label:"Code"                                                                                              pattern:"^[a-z]*$"     readonly:"true"`
	Loose    string `form:"input,text"     label:"Loose"                                                                                             pattern:"[a-z]"`
	Flags    string `form:"input,text"     label:"Flags"                                                                                             pattern:"(?i)^[a-z]+$"`
	Either   string `form:"input,text"     label:"Either"                                                                                            pattern:"^a|b$"`
	Note     string `form:"textarea"       label:"Note"                                     maxLength:"500"                                                                                 disabled:"true"`
	Size     string `form:"dropdown"       label:"Size"                                                                                                                                     disabled:"true" values:"s:Small;l:Large" placeholder:"—"`
	Agree    bool   `form:"checkbox"       label:"Agree"                                                                                                                                    disabled:"true"`
	Password string `form:"input,password" label:"Password" autocomplete:"current-password"`
}

// renderString renders model with errs and returns the markup.
func renderString(t *testing.T, f *Form, model any, errs FieldErrors) string {
	t.Helper()

	html, err := f.formRender(model, errs)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	return string(html)
}

// control returns the markup of the element with id, up to its closing >.
func control(t *testing.T, html, id string) string {
	t.Helper()

	start := strings.Index(html, `id="`+id+`"`)
	if start < 0 {
		t.Fatalf("no element %q in %s", id, html)
	}

	open := strings.LastIndex(html[:start], "<")
	end := strings.Index(html[start:], ">")

	return html[open : start+end+1]
}

func TestRenderAttributes(t *testing.T) {
	html := renderString(t, NewForm(), attributesModel{}, nil)

	cases := []struct {
		id   string
		want []string
		not  []string
	}{
		{
			"Email",
			[]string{`autocomplete="email"`, `maxlength="254"`, `minlength="3"`, `aria-describedby="Email_description"`},
			[]string{"aria-invalid"},
		},
		{"Code", []string{`pattern="^[a-z]*$"`, "readonly"}, nil},
		// Patterns that mean something else in a browser stay on the server.
		{"Loose", nil, []string{"pattern="}},
		{"Flags", nil, []string{"pattern="}},
		{"Either", nil, []string{"pattern="}},
		{"Note", []string{`maxlength="500"`, "disabled"}, nil},
		{"Size", []string{"disabled"}, nil},
		{"Agree", []string{"disabled"}, nil},
		{"Password", []string{`autocomplete="current-password"`}, nil},
	}

	for _, c := range cases {
		element := control(t, html, c.id)
		for _, want := range c.want {
			if !strings.Contains(element, want) {
				t.Errorf("%s: missing %s in %s", c.id, want, element)
			}
		}

		for _, not := range c.not {
			if strings.Contains(element, not) {
				t.Errorf("%s: unexpected %s in %s", c.id, not, element)
			}
		}
	}

	if !strings.Contains(html, `<option value="" selected>—</option>`) {
		t.Errorf("dropdown placeholder not rendered as an empty option: %s", html)
	}
}

func TestRenderErrorsDescribeControl(t *testing.T) {
	errs := FieldErrors{
		FieldValidationError{Field: "Email", Err: "first"},
		FieldValidationError{Field: "Email", Err: "second"},
	}

	html := renderString(t, NewForm(), attributesModel{}, errs)
	element := control(t, html, "Email")

	if !strings.Contains(element, `aria-invalid="true"`) {
		t.Errorf("invalid control not marked: %s", element)
	}

	if !strings.Contains(element, `aria-describedby="Email_description Email_error_0 Email_error_1"`) {
		t.Errorf("control not described by its errors: %s", element)
	}

	for _, id := range []string{"Email_error_0", "Email_error_1"} {
		if !strings.Contains(html, `id="`+id+`"`) {
			t.Errorf("no error element %s: %s", id, html)
		}
	}
}

type readOnlyModel struct {
	Name  string `form:"input,text" label:"Name"`
	Agree bool   `form:"checkbox"   label:"Agree"`
}

func TestReadOnlyForm(t *testing.T) {
	model := WithInfo(
		readOnlyModel{},
		Info{Target: "/save", Method: "post", ReadOnly: true, CancelTarget: "/back", CancelText: "Back"},
	)
	html := renderString(t, NewForm(), model, nil)

	for _, id := range []string{"Name", "Agree"} {
		if !strings.Contains(control(t, html, id), "disabled") {
			t.Errorf("%s not disabled in a read-only form: %s", id, html)
		}
	}

	if strings.Contains(html, `type="submit"`) {
		t.Errorf("read-only form has a submit button: %s", html)
	}

	if !strings.Contains(html, `href="/back"`) {
		t.Errorf("read-only form lost its cancel link: %s", html)
	}

	editable := renderString(t, NewForm(), WithInfo(readOnlyModel{}, Info{Target: "/save", Method: "post"}), nil)
	if strings.Contains(control(t, editable, "Name"), "disabled") || !strings.Contains(editable, `type="submit"`) {
		t.Errorf("editable form rendered read-only: %s", editable)
	}
}

func TestFormAttributesSorted(t *testing.T) {
	model := WithInfo(
		readOnlyModel{},
		Info{Target: "/", Method: "post", Attributes: map[string]string{"hx-post": "/a", "data-b": "b", "aria-label": "c"}},
	)

	first := renderString(t, NewForm(), model, nil)
	for range 5 {
		if renderString(t, NewForm(), model, nil) != first {
			t.Fatal("form attributes render in a different order each time")
		}
	}

	if !strings.Contains(first, `aria-label="c" data-b="b" hx-post="/a"`) {
		t.Errorf("attributes not sorted: %s", first)
	}
}

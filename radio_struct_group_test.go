package form

import (
	"strings"
	"testing"
)

func TestRender_StructRadioGroup_RendersOptions(t *testing.T) {
	type RadioGroupBlock struct {
		Option1 bool `name:"RadioGroup" label:"first option"`
		Option2 bool `name:"RadioGroup" label:"second option"`
	}

	type M struct {
		Info
		RadioGroup RadioGroupBlock `form:"radios" label:"Radio Group"`
	}

	m := M{}
	m.Info.Target = "/"
	m.Info.Method = "POST"
	m.Info.SubmitText = "Submit"
	m.RadioGroup.Option1 = true

	f := NewForm()
	f.SetTheme("plain")

	out, err := f.formRender(m, nil)
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	html := string(out)
	if !strings.Contains(html, "type=\"radio\"") {
		t.Fatalf("expected radio inputs to render, got: %s", html)
	}
	if !strings.Contains(html, "first option") || !strings.Contains(html, "second option") {
		t.Fatalf("expected radio option labels to render, got: %s", html)
	}
	// Selected should be checked
	if !strings.Contains(html, "checked") {
		t.Fatalf("expected checked radio to render, got: %s", html)
	}
}

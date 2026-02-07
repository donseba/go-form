package form

import (
	"strings"
	"testing"
)

func TestRender_StructRadioGroup_ExplicitRadioGroupInputType(t *testing.T) {
	type RadioGroupBlock struct {
		Option1 bool `name:"RadioGroup" label:"first option"`
		Option2 bool `name:"RadioGroup" label:"second option"`
	}

	type M struct {
		Info
		RadioGroup RadioGroupBlock `form:"radios,radio_group" legend:"Radio Group"`
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
	if !strings.Contains(html, "role=\"radiogroup\"") {
		t.Fatalf("expected radiogroup container, got: %s", html)
	}
	if strings.Count(html, "type=\"radio\"") != 2 {
		t.Fatalf("expected 2 radio inputs, got: %s", html)
	}
	if !strings.Contains(html, "checked") {
		t.Fatalf("expected one option checked, got: %s", html)
	}
}

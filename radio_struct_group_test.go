package form

import (
	"errors"
	"net/url"
	"reflect"
	"regexp"
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

// planBlock is a struct radio group: its bool fields carry the group's name.
type planBlock struct {
	Monthly bool `name:"plan" label:"Monthly"`
	Yearly  bool `name:"plan" label:"Yearly"`
}

// radioInputs are the name and value of every radio button in page.
func radioInputs(t *testing.T, page string) [][2]string {
	t.Helper()

	var inputs [][2]string

	for _, input := range regexp.MustCompile(`<input type="radio"[^>]*>`).FindAllString(page, -1) {
		name := regexp.MustCompile(`name="([^"]*)"`).FindStringSubmatch(input)
		value := regexp.MustCompile(`value="([^"]*)"`).FindStringSubmatch(input)

		if name == nil || value == nil {
			t.Fatalf("radio button without a name or value: %s", input)
		}

		inputs = append(inputs, [2]string{name[1], value[1]})
	}

	return inputs
}

func TestStructRadioGroupRoundTrip(t *testing.T) {
	type implicit struct {
		Plan planBlock `form:"radios" name:"plan" label:"Plan"`
	}

	type explicit struct {
		Plan planBlock `form:"radios,radio_group" name:"plan" legend:"Plan"`
	}

	type nested struct {
		Billing struct {
			Plan planBlock `form:"radios" name:"plan" label:"Plan"`
		} `name:"billing"`
	}

	tests := map[string]struct {
		model any
		plan  func(model any) planBlock
	}{
		"radios": {&implicit{Plan: planBlock{Monthly: true}}, func(model any) planBlock {
			return model.(*implicit).Plan
		}},
		"radios,radio_group": {&explicit{Plan: planBlock{Monthly: true}}, func(model any) planBlock {
			return model.(*explicit).Plan
		}},
		"nested": {&nested{}, func(model any) planBlock {
			return model.(*nested).Billing.Plan
		}},
	}

	f := NewForm()
	f.SetTheme("plain")

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			page, err := f.Render(nil, reflect.ValueOf(test.model).Elem().Interface(), nil)
			if err != nil {
				t.Fatalf("render: %v", err)
			}

			inputs := radioInputs(t, string(page))
			if len(inputs) != 2 {
				t.Fatalf("got radio buttons %v, want two", inputs)
			}

			// Submit the second option, as a browser does.
			yearly := inputs[1]
			submission := postForm(url.Values{yearly[0]: {yearly[1]}})

			if err := MapFormWith(submission, test.model, MapOptions{Strict: true}); err != nil {
				t.Fatalf("MapForm: %v", err)
			}

			if got := test.plan(test.model); got != (planBlock{Yearly: true}) {
				t.Errorf("after submitting %s=%s: %+v, want only Yearly", yearly[0], yearly[1], got)
			}

			// Nothing chosen clears every option.
			if err := MapForm(postForm(url.Values{}), test.model); err != nil {
				t.Fatalf("MapForm: %v", err)
			}

			if got := test.plan(test.model); got != (planBlock{}) {
				t.Errorf("nothing chosen: %+v, want no option", got)
			}
		})
	}
}

func TestStructRadioGroupUnknownOption(t *testing.T) {
	type input struct {
		Plan planBlock `form:"radios" name:"plan" label:"Plan"`
		Name string
	}

	got := input{Plan: planBlock{Monthly: true}}
	err := MapFormWith(postForm(url.Values{"plan": {"plan.Weekly"}, "Name": {"Ada"}}), &got, MapOptions{Strict: true})

	var invalid *InvalidValuesError
	if !errors.As(err, &invalid) || !reflect.DeepEqual(strictNames(invalid.Errors), []string{"plan"}) {
		t.Fatalf("got %v, want an invalid value for plan", err)
	}

	if got.Plan != (planBlock{}) || got.Name != "Ada" {
		t.Errorf("got %+v, want no option and name Ada", got)
	}
}

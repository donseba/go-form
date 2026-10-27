package form

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type bookingInput struct {
	Adults int    `form:"input,number" name:"adults" label:"Adults" min:"0" max:"50"`
	Notes  string `form:"textarea"     name:"notes"                                  maxLength:"2000" placeholder:"Anything else?"`
	Guest  struct {
		Email string `form:"input,email" name:"email" maxLength:"254"`
	} `                    name:"guest"`

	left int
}

// FormOverrides limits the adults to the places left.
func (in bookingInput) FormOverrides() Overrides {
	if in.left == 0 {
		return nil
	}

	return Overrides{"adults": {"min": "1", "max": strconv.Itoa(in.left)}}
}

func TestOverriderAppliesToRenderingAndValidation(t *testing.T) {
	f := NewForm()
	f.SetTheme("plain")

	input := bookingInput{Adults: 4, left: 3}

	html, err := f.Field(nil, &input, "adults", nil)
	if err != nil {
		t.Fatalf("Field: %v", err)
	}

	if !strings.Contains(string(html), `min="1"`) || !strings.Contains(string(html), `max="3"`) {
		t.Errorf("the overrides do not render: %s", html)
	}

	errs := f.ValidateForm(&input)
	if len(errs) != 1 || errs[0].(FieldValidationError).Rule != "max" {
		t.Errorf("ValidateForm = %v, want adults over the overridden max", errs)
	}

	input.left = 0
	if errs := f.ValidateForm(&input); len(errs) != 0 {
		t.Errorf("without overrides the tags apply, got %v", errs)
	}
}

func TestOverriderAppliesToBind(t *testing.T) {
	r := postForm(url.Values{"adults": {"0"}})

	input := bookingInput{left: 2}

	errs, err := NewForm().Bind(r, &input, MapOptions{}, nil)
	if err != nil {
		t.Fatalf("Bind: %v", err)
	}

	if len(errs) != 1 || errs[0].(FieldValidationError).Rule != "min" {
		t.Errorf("Bind = %v, want adults under the overridden min", errs)
	}
}

func TestRenderModelOverridesWinAndReachNestedFields(t *testing.T) {
	f := NewForm()
	f.SetTheme("plain")

	model := WithInfo(bookingInput{left: 3}, Info{Target: "/book", Method: "post"})
	model.Overrides = Overrides{
		"adults":      {"max": "2"},
		"notes":       {"maxLength": "10", "placeholder": ""},
		"guest.email": {"maxLength": "5", "name": "ignored"},
	}

	html, err := f.Render(nil, model, nil)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	out := string(html)

	for _, want := range []string{`max="2"`, `min="1"`, `maxlength="10"`, `maxlength="5"`, `name="guest.email"`} {
		if !strings.Contains(out, want) {
			t.Errorf("render misses %s:\n%s", want, out)
		}
	}

	if strings.Contains(out, "Anything else?") || strings.Contains(out, "ignored") {
		t.Errorf("an empty override did not remove the placeholder, or the name changed:\n%s", out)
	}

	input := bookingInput{Adults: 3, Notes: strings.Repeat("x", 11), left: 3}
	input.Guest.Email = "a@b.example"
	model.Model = &input

	var rules []string

	for _, err := range f.ValidateForm(model) {
		validationErr := err.(FieldValidationError)
		rules = append(rules, validationErr.Field+":"+validationErr.Rule)
	}

	want := []string{"adults:max", "notes:maxLength", "guest.email:maxLength"}
	if !reflect.DeepEqual(rules, want) {
		t.Errorf("ValidateForm of the RenderModel = %v, want %v", rules, want)
	}
}

func TestOverrideTag(t *testing.T) {
	tag := reflect.StructTag(`form:"input,number" name:"n" max:"5" label:"Say \"hi\""`)

	got := overrideTag(tag, Tags{"max": "9", "min": "1", "label": "", "name": "x"})

	if got.Get("max") != "9" || got.Get("min") != "1" || got.Get("name") != "n" || got.Get("form") != "input,number" {
		t.Errorf("overrideTag = %s", got)
	}

	if _, ok := got.Lookup("label"); ok {
		t.Errorf("an empty value did not remove the tag: %s", got)
	}

	if entries := parseTag(tag); len(entries) != 4 || entries[3].value != `Say "hi"` {
		t.Errorf("parseTag = %#v", entries)
	}
}

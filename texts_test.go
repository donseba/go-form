package form

import (
	"reflect"
	"testing"

	"github.com/donseba/go-form/v2/types"
)

type textsAddress struct {
	City string `form:"input,text" name:"city" label:"City" requiredMsg:"Which city?"`
}

type textsInput struct {
	Subject string               `form:"input,hidden" name:"subject" label:"Hidden label"`
	Name    string               `form:"input,text"   name:"name"    label:"Name"         placeholder:"Your name" description:"As on your card" required:"true" requiredMsg:"Tell us your name"`
	Size    string               `form:"dropdown"     name:"size"    label:"Size"                                                                                                               values:"s:Small;l:Large"`
	Code    string               `form:"dropdown"     name:"code"    label:"Code"                                                                                                               values:"x:Kept"          translate:"false"`
	Colour  SortedSelect[string] `form:"dropdown"     name:"colour"  label:"Colour"                                                                                                                                      translate:"true"`
	Page    SortedSelect[string] `form:"dropdown"     name:"page"    label:"Page"`
	Body    string               `form:"textarea"     name:"body"                                                                                                                                                                          errorMsg:"Something is wrong" customRuleMsg:"Custom rule"`
	Address textsAddress         `                    name:"address"                                                                                                                                                                                                                                 legend:"Address"`
	Plain   string
}

func TestTextsOfStruct(t *testing.T) {
	model := WithInfo(textsInput{
		Colour: NewSortedSelect(map[string]string{"red": "Red"}),
		Page:   NewSortedSelect(map[string]string{"home": "Home page"}),
	}, Info{SubmitText: "Send", CancelTarget: "/", CancelText: "Back"})

	want := []string{
		"Send", "Back",
		"Name", "Your name", "As on your card",
		"Size", "Small", "Large",
		"Code",
		"Colour", "Red",
		"Page",
		"body",
		"Address", "City",
		"Plain",
		"Tell us your name", "Something is wrong", "Custom rule", "Which city?",
	}

	if got := Texts(model); !reflect.DeepEqual(got, want) {
		t.Errorf("Texts =\n%q\nwant\n%q", got, want)
	}
}

func TestTextsOfTypesSchemasAndDuplicates(t *testing.T) {
	schema := Schema{Fields: []Spec{
		{Name: "topic", Label: "Topic", Type: types.FieldTypeDropdown, Placeholder: "Pick one", Options: []types.FieldValue{
			{Value: "a", Name: "Sales", Translate: true},
			{Value: "b", Name: "Raw"},
		}, Messages: map[string]string{"required": "Pick a topic"}},
		{Name: "city", Label: "City"},
	}}

	got := Texts((*textsAddress)(nil), schema, nil, textsAddress{})
	want := []string{"City", "Which city?", "Topic", "Pick one", "Sales", "Pick a topic"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Texts = %q, want %q", got, want)
	}
}

func TestTextsIncludeOverrides(t *testing.T) {
	model := WithInfo(textsAddress{}, Info{})
	model.Overrides = Overrides{"city": {"placeholder": "Amsterdam", "maxLengthMsg": "Too long a city"}}

	got := Texts(model)
	want := []string{DefaultSubmitText, "City", "Amsterdam", "Which city?", "Too long a city"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Texts = %q, want %q", got, want)
	}
}

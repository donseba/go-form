package form

import (
	"io/fs"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/donseba/go-form/v2/templates"
	"github.com/donseba/go-form/v2/types"
)

type voteInput struct {
	Poll   string `form:"input,hidden" name:"poll"`
	Option string `form:"radios"       name:"option" label:"Your choice" required:"true"`
}

// pollChoices have labels with the separators of a values tag.
var pollChoices = []types.FieldValue{
	{Value: "o1", Name: "Yes; absolutely"},
	{Value: "o2", Name: "No: never"},
	{Value: "o3", Name: "Closed", Disabled: true},
}

// render renders model with go-form's plain theme.
func render(t *testing.T, model any) string {
	t.Helper()

	f := NewForm()
	f.SetTheme("plain")

	html, err := f.Render(nil, model, nil)
	if err != nil {
		t.Fatal(err)
	}

	return string(html)
}

// rules are the rules of errs by field name.
func rules(errs FieldErrors) map[string]string {
	out := map[string]string{}

	for _, err := range errs {
		if validationErr, ok := err.(FieldValidationError); ok {
			out[validationErr.Field] = validationErr.Rule
		}
	}

	return out
}

// checkedValues are the values of the checked inputs and selected options
// of html.
var checkedValues = regexp.MustCompile(`value="([^"]*)"\s+(?:checked|selected)`)

func checked(html string) []string {
	var values []string
	for _, match := range checkedValues.FindAllStringSubmatch(html, -1) {
		values = append(values, match[1])
	}

	return values
}

func TestOptionsRenderAsRadioGroup(t *testing.T) {
	model := WithInfo(voteInput{Poll: "p1", Option: "o2"}, Info{Target: "/vote", Method: "post"})
	model.Options = FieldOptions{"option": pollChoices}

	html := render(t, model)

	if !strings.Contains(html, "<fieldset") {
		t.Errorf("no radio group: %s", html)
	}

	for _, label := range []string{"Yes; absolutely", "No: never"} {
		if !strings.Contains(html, label) {
			t.Errorf("label %q missing: %s", label, html)
		}
	}

	if got := checked(html); !reflect.DeepEqual(got, []string{"o2"}) {
		t.Errorf("checked = %v, want [o2]", got)
	}

	if strings.Count(html, `type="radio"`) != 3 {
		t.Errorf("want 3 radio buttons: %s", html)
	}
}

func TestOptionsRenderAsSelectAndMultiCheckbox(t *testing.T) {
	type input struct {
		Country string   `name:"country"`
		Level   int      `name:"level"   form:"dropdown"      values:"1:One;2:Two"`
		Tags    []string `name:"tags"    form:"multicheckbox"`
		Address struct {
			Region *uint16 `form:"dropdown" name:"region"`
		} `name:"address"`
	}

	region := uint16(7)

	in := input{Country: "nl", Level: 3, Tags: []string{"a", "c"}}
	in.Address.Region = &region

	model := WithInfo(in, Info{})
	model.Options = FieldOptions{
		"country":        {{Value: "be", Name: "Belgium"}, {Value: "nl", Name: "Netherlands"}},
		"level":          {{Value: "3", Name: "Three"}},
		"tags":           {{Value: "a", Name: "A"}, {Value: "b", Name: "B"}, {Value: "c", Name: "C"}},
		"address.region": {{Value: "7", Name: "Seven"}},
	}

	html := render(t, model)

	if strings.Count(html, "<select") != 3 {
		t.Errorf("want three selects: %s", html)
	}

	if strings.Count(html, `type="checkbox"`) != 3 {
		t.Errorf("want three checkboxes: %s", html)
	}

	if got := checked(html); !reflect.DeepEqual(got, []string{"nl", "3", "a", "c", "7"}) {
		t.Errorf("chosen = %v", got)
	}

	if strings.Contains(html, "One") {
		t.Errorf("the options of the values tag render: %s", html)
	}
}

func TestOptionsValidate(t *testing.T) {
	f := NewForm()

	cases := []struct {
		option string
		want   map[string]string
	}{
		{option: "o1", want: map[string]string{}},
		{option: "", want: map[string]string{"option": "required"}},
		{option: "o9", want: map[string]string{"option": "values"}},
		{option: "o3", want: map[string]string{"option": "values"}},
	}

	for _, c := range cases {
		model := WithInfo(&voteInput{Option: c.option}, Info{})
		model.Options = FieldOptions{"option": pollChoices}

		if got := rules(f.ValidateForm(model)); !reflect.DeepEqual(got, c.want) {
			t.Errorf("option %q: rules %v, want %v", c.option, got, c.want)
		}
	}

	// Without the options nothing limits the value.
	if errs := f.ValidateForm(&voteInput{Option: "o9"}); len(errs) != 0 {
		t.Errorf("without options: %v", errs)
	}
}

func TestOptionsReplaceValuesTagInValidation(t *testing.T) {
	type input struct {
		Plan  string `form:"dropdown" name:"plan"  values:"basic:Basic"`
		Inner struct {
			Size int `form:"dropdown" name:"size"`
		} `                name:"inner"`
	}

	in := &input{Plan: "gold"}
	in.Inner.Size = 4

	model := WithInfo(in, Info{})
	model.Options = FieldOptions{
		"plan":       {{Value: "gold", Name: "Gold"}},
		"inner.size": {{Value: "2", Name: "Two"}},
	}

	got := rules(NewForm().ValidateForm(model))
	if want := map[string]string{"inner.size": "values"}; !reflect.DeepEqual(got, want) {
		t.Errorf("rules %v, want %v", got, want)
	}
}

func TestBindWithOptions(t *testing.T) {
	f := NewForm()

	var in voteInput

	model := WithInfo(&in, Info{})
	model.Options = FieldOptions{"option": pollChoices}

	errs, err := f.Bind(postForm(url.Values{"poll": {"p1"}, "option": {"o9"}}), model, MapOptions{}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if in.Poll != "p1" || in.Option != "o9" {
		t.Errorf("mapped %+v", in)
	}

	if got := rules(errs); !reflect.DeepEqual(got, map[string]string{"option": "values"}) {
		t.Errorf("rules %v", got)
	}

	errs, err = f.Bind(postForm(url.Values{"option": {"o1"}}), &model, MapOptions{}, nil)
	if err != nil || len(errs) != 0 || in.Option != "o1" {
		t.Errorf("valid vote: %v, %v, %+v", errs, err, in)
	}

	// The model must be a pointer to be mapped onto.
	if _, err := f.Bind(postForm(url.Values{}), WithInfo(voteInput{}, Info{}), MapOptions{}, nil); err == nil {
		t.Error("Bind mapped onto a RenderModel of a struct value")
	}
}

func TestCarryWithOptions(t *testing.T) {
	type input struct {
		Poll   string   `form:"input,hidden"  name:"poll"`
		Option string   `form:"radios"        name:"option"`
		Extras []string `form:"multicheckbox" name:"extras"`
	}

	model := WithInfo(input{Poll: "p1"}, Info{})
	model.Options = FieldOptions{
		"option": pollChoices,
		"extras": {{Value: "x", Name: "X"}, {Value: "y", Name: "Y"}},
	}

	errs := FieldErrors{FieldValidationError{Field: "option", Err: "Choose one"}}

	carry, err := NewCarry(model, url.Values{"poll": {"p1"}, "option": {"o1"}, "extras": {"y"}}, errs)
	if err != nil {
		t.Fatal(err)
	}

	data, err := carry.Encode(0)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := DecodeCarry(data)
	if err != nil {
		t.Fatal(err)
	}

	model.Carry = &decoded
	html := render(t, model)

	if got := checked(html); !reflect.DeepEqual(got, []string{"o1", "y"}) {
		t.Errorf("chosen = %v: %s", got, html)
	}

	if !strings.Contains(html, "Choose one") {
		t.Errorf("carried error missing: %s", html)
	}
}

func TestOptionTextsAreExtracted(t *testing.T) {
	model := WithInfo(voteInput{}, Info{})
	model.Options = FieldOptions{"option": {
		{Value: "a", Name: "Translated label", Translate: true},
		{Value: "b", Name: "Data label"},
	}}

	texts := Texts(model)

	if !contains(texts, "Translated label") || contains(texts, "Data label") {
		t.Errorf("texts = %v", texts)
	}
}

func contains(texts []string, text string) bool {
	for _, candidate := range texts {
		if candidate == text {
			return true
		}
	}

	return false
}

func TestValuesTagOnIntegers(t *testing.T) {
	type input struct {
		MaxAge  int    `form:"dropdown"      name:"max_age" values:"0:No time;60:1 minute;300:5 minutes"`
		Size    int    `form:"dropdown"      name:"size"    values:"1:Small;2:Large"                     required:"true"`
		Pointer *int64 `form:"radios"        name:"pointer" values:"5:Five;6:Six"`
		Days    []int  `form:"multicheckbox" name:"days"    values:"1:Monday;2:Tuesday"`
	}

	six := int64(6)
	html := render(t, input{MaxAge: 300, Pointer: &six, Days: []int{2}})

	if got := checked(html); !reflect.DeepEqual(got, []string{"300", "6", "2"}) {
		t.Errorf("chosen = %v: %s", got, html)
	}

	if !strings.Contains(html, "<fieldset") {
		t.Errorf("radios with a values tag render one radio: %s", html)
	}

	f := NewForm()

	cases := []struct {
		values url.Values
		want   map[string]string
	}{
		{values: url.Values{"max_age": {"0"}, "size": {"2"}, "pointer": {"5"}, "days": {"1", "2"}}, want: map[string]string{}},
		{values: url.Values{"max_age": {"30"}, "size": {"2"}}, want: map[string]string{"max_age": "values"}},
		{values: url.Values{"size": {"3"}}, want: map[string]string{"size": "values"}},
		{values: url.Values{"size": {""}}, want: map[string]string{"size": "required"}},
		{values: url.Values{"size": {"1"}, "pointer": {"0"}}, want: map[string]string{"pointer": "values"}},
		{values: url.Values{"size": {"1"}, "days": {"1", "7"}}, want: map[string]string{"days": "values"}},
	}

	for _, c := range cases {
		var in input

		errs, err := f.Bind(postForm(c.values), &in, MapOptions{}, nil)
		if err != nil {
			t.Fatal(err)
		}

		if got := rules(errs); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%v: rules %v, want %v", c.values, got, c.want)
		}
	}
}

func TestValuesTagMessage(t *testing.T) {
	type input struct {
		Size int `form:"dropdown" name:"size" values:"1:Small;2:Large"`
	}

	errs := NewForm().ValidateForm(&input{Size: 3})
	if len(errs) != 1 || errs.Messages("size")[0] != "The value '3' is not allowed." {
		t.Errorf("errs = %v", errs)
	}
}

func TestMapFormLists(t *testing.T) {
	type input struct {
		Tags  []string `name:"tags"`
		Days  []uint8  `name:"days"`
		Other []int    `name:"other"`
	}

	in := input{Other: []int{9}}

	err := MapForm(postForm(url.Values{"tags": {" a ", "", "b"}, "days": {"1", "300"}}), &in)
	if err != nil {
		t.Fatal(err)
	}

	// []uint8 is bytes, not a list of choices.
	if !reflect.DeepEqual(in.Tags, []string{"a", "b"}) || in.Days != nil || in.Other != nil {
		t.Errorf("mapped %+v", in)
	}

	// A list that cannot be read keeps its value.
	type numbers struct {
		Days []int `name:"days"`
	}

	kept := numbers{Days: []int{1}}

	err = MapFormWith(postForm(url.Values{"days": {"2", "x"}}), &kept, MapOptions{Strict: true})
	if err == nil || !reflect.DeepEqual(kept.Days, []int{1}) {
		t.Errorf("unreadable list: %v, %v", err, kept.Days)
	}

	// Without a POST, no values keep the list.
	get, _ := http.NewRequest(http.MethodGet, "/?x=1", nil)
	if err := MapForm(get, &kept); err != nil || !reflect.DeepEqual(kept.Days, []int{1}) {
		t.Errorf("GET without values: %v, %v", err, kept.Days)
	}
}

// TestRadioGroupIsAFieldset checks that a group of radio buttons renders
// as a fieldset whose legend is the field's label, with the required
// marker, the description and the errors inside, and no label that points
// to an element that does not exist; with the classes of Wrapper and
// Fieldset, and of Label and Legend.
func TestRadioGroupIsAFieldset(t *testing.T) {
	type poll struct {
		Option string `form:"radios" name:"option" label:"Your choice" description:"Pick one" required:"true" values:"a:Apples;b:Pears"`
	}

	if _, err := templates.RegisterClassTheme("fieldset-test", templates.ThemeClasses{
		Wrapper:  templates.StyleOption{Class: "field"},
		Fieldset: templates.StyleOption{Class: "field-set"},
		Label:    templates.StyleOption{Class: "label"},
		Legend:   templates.StyleOption{Class: "legend"},
	}); err != nil {
		t.Fatal(err)
	}

	f := NewForm()
	f.SetTheme("fieldset-test")

	errs := FieldErrors{FieldValidationError{Field: "option", Err: "Choose one."}}

	out, err := f.formRender(WithInfo(poll{}, Info{Target: "/vote", Method: "post"}), errs)
	if err != nil {
		t.Fatal(err)
	}

	html := strings.Join(strings.Fields(string(out)), " ")

	for _, want := range []string{
		`<fieldset class="field field-set"> <legend class="label legend" id="option_label"> Your choice <span aria-hidden="true">*</span>`,
		`<div id="option_description">Pick one</div>`,
		`<div id="option_error_0" role="alert">Choose one.</div> </fieldset>`,
		`aria-describedby="option_description option_error_0"`,
		`<label for="option_0" id="option_0_label">Apples</label>`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %s in\n%s", want, html)
		}
	}

	for _, unwanted := range []string{`for="option"`, `role="radiogroup"`, `<label class="label"`} {
		if strings.Contains(html, unwanted) {
			t.Errorf("unexpected %s in\n%s", unwanted, html)
		}
	}
}

// TestMultiCheckboxIsAFieldset checks that a multicheckbox renders as a
// fieldset whose legend is the field's label, as a group of radio buttons
// does, with no label that points to an element that does not exist.
func TestMultiCheckboxIsAFieldset(t *testing.T) {
	type topics struct {
		Topics []string `form:"multicheckbox" name:"topics" label:"Topics" description:"Pick any" values:"a:Apples;b:Pears"`
	}

	if _, err := templates.RegisterClassTheme("fieldset-checkboxes-test", templates.ThemeClasses{
		Wrapper:  templates.StyleOption{Class: "field"},
		Fieldset: templates.StyleOption{Class: "field-set"},
		Label:    templates.StyleOption{Class: "label"},
		Legend:   templates.StyleOption{Class: "legend"},
	}); err != nil {
		t.Fatal(err)
	}

	f := NewForm()
	f.SetTheme("fieldset-checkboxes-test")

	errs := FieldErrors{FieldValidationError{Field: "topics", Err: "Choose one."}}

	out, err := f.formRender(WithInfo(topics{Topics: []string{"b"}}, Info{Target: "/topics", Method: "post"}), errs)
	if err != nil {
		t.Fatal(err)
	}

	html := strings.Join(strings.Fields(string(out)), " ")

	for _, want := range []string{
		`<fieldset class="field field-set"> <legend class="label legend" id="topics_label"> Topics </legend>`,
		`<div id="topics_description">Pick any</div>`,
		`<div id="topics_error_0" role="alert">Choose one.</div> </fieldset>`,
		`<label for="topics_0" id="topics_0_label">Apples</label>`,
		`<label for="topics_1" id="topics_1_label">Pears</label>`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %s in\n%s", want, html)
		}
	}

	for _, unwanted := range []string{`for="topics"`, `<label class="label"`} {
		if strings.Contains(html, unwanted) {
			t.Errorf("unexpected %s in\n%s", unwanted, html)
		}
	}
}

// TestOptionGroupsWithoutFieldsetTemplate checks that a theme without
// fieldset.gohtml and legend.gohtml, such as one whose templates all come
// from a folder of its own, renders groups of radio buttons and
// checkboxes with its wrapper and label templates.
func TestOptionGroupsWithoutFieldsetTemplate(t *testing.T) {
	type choices struct {
		Option string   `form:"radios"        name:"option" label:"Your choice" values:"a:Apples;b:Pears"`
		Topics []string `form:"multicheckbox" name:"topics" label:"Topics"      values:"a:Apples;b:Pears"`
	}

	files := fstest.MapFS{}

	err := fs.WalkDir(templates.TemplateFS, "gohtml", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path.Ext(name) != ".gohtml" {
			return err
		}

		if base := path.Base(name); base == "fieldset.gohtml" || base == "legend.gohtml" {
			return nil
		}

		data, err := fs.ReadFile(templates.TemplateFS, name)
		if err != nil {
			return err
		}

		files[name] = &fstest.MapFile{Data: data}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	theme := templates.RegisterTheme("no-fieldset-test", templates.ThemeClasses{}, nil)
	if err := theme.LoadTemplatesFS(files, "gohtml"); err != nil {
		t.Fatal(err)
	}

	f := NewForm()
	f.SetTheme("no-fieldset-test")

	out, err := f.formRender(WithInfo(choices{}, Info{Target: "/choose", Method: "post"}), nil)
	if err != nil {
		t.Fatal(err)
	}

	html := strings.Join(strings.Fields(string(out)), " ")

	if strings.Contains(html, "<fieldset") || strings.Contains(html, "<legend") {
		t.Errorf("fieldset without its template in\n%s", html)
	}

	for _, want := range []string{`<label for="option" id="option_label">`, `<label for="topics" id="topics_label">`} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %s in\n%s", want, html)
		}
	}
}

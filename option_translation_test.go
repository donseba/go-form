package form

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// optionStatus is an enum whose labels are translation keys with translate:"true".
type optionStatus string

func (optionStatus) Enum() []any {
	return []any{"active", "inactive"}
}

type optionModel struct {
	Info `target:"/submit" method:"post"`
	// Labels from a values tag are translated, like field labels.
	Size string `form:"dropdown" label:"Size" values:"s:Small;l:Large"`
	// translate:"false" keeps them as written.
	Code string `form:"dropdown" label:"Code" values:"s:Small;l:Large" translate:"false"`
	// SortedSelect labels are data: translated only with translate:"true".
	Colour SortedSelect[string]      `form:"dropdown" label:"Colour" translate:"true"`
	Page   SortedSelect[string]      `form:"dropdown" label:"Page"`
	Tags   SortedMultiSelect[string] `form:"multicheckbox" label:"Tags" translate:"true"`
	Pages  SortedMultiSelect[string] `form:"multicheckbox" label:"Pages"`
	Status optionStatus              `form:"dropdown" label:"Status" translate:"true"`
	// Struct radio groups take their option labels from the bool fields.
	Delivery deliveryBlock `form:"radios,radio_group" legend:"Delivery"`
}

type deliveryBlock struct {
	Post   bool `name:"Delivery" label:"Post"`
	Pickup bool `name:"Delivery" label:"Pickup"`
}

func optionTranslate(loc Localizer, key string, args ...any) string {
	_ = loc
	_ = args

	return "T(" + key + ")"
}

func renderOptions(t *testing.T, f *Form) string {
	t.Helper()

	model := optionModel{
		Colour: NewSortedSelect(map[string]string{"red": "Red"}),
		Page:   NewSortedSelect(map[string]string{"home": "Home"}),
		Tags:   NewSortedMultiSelect(map[string]string{"news": "News"}),
		Pages:  NewSortedMultiSelect(map[string]string{"about": "About"}),
	}

	html, err := f.formRenderLocalized(testLocalizer{Locale: "it"}, &model, nil)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	return string(html)
}

// TestOptionLabelsTranslated checks that select, multicheckbox and radio
// group options are translated by the same rule: when the option asks for it.
func TestOptionLabelsTranslated(t *testing.T) {
	f := NewTranslatedForm(optionTranslate)
	f.SetTheme("plain")

	html := renderOptions(t, f)

	for _, want := range []string{
		">T(Small)</option>",
		">T(Large)</option>",
		">Small</option>",
		">T(Red)</option>",
		">Home</option>",
		">T(News)</label>",
		">About</label>",
		">T(enum||optionStatus.active)</option>",
		">T(Post)</label>",
		">T(Pickup)</label>",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("expected %q in HTML:\n%s", want, html)
		}
	}

	for _, unwanted := range []string{"T(Home)", "T(About)", "enum||optionStatus.active</option>"} {
		if strings.Contains(html, unwanted) {
			t.Errorf("unexpected %q in HTML:\n%s", unwanted, html)
		}
	}
}

// TestOptionLabelsWithoutTranslation checks that a form without a translation
// function prints the labels as written.
func TestOptionLabelsWithoutTranslation(t *testing.T) {
	f := NewForm()
	f.SetTheme("plain")

	html := renderOptions(t, f)

	for _, want := range []string{">Small</option>", ">Red</option>", ">News</label>", ">Post</label>"} {
		if !strings.Contains(html, want) {
			t.Errorf("expected %q in HTML:\n%s", want, html)
		}
	}
}

type requiredLabelModel struct {
	Info `target:"/submit" method:"post"`
	Name string `form:"input,text" label:"Name" required:"true"`
}

// TestRequiredLabelTranslated checks the screen reader text of required
// fields: English without translation, translated through its key otherwise.
func TestRequiredLabelTranslated(t *testing.T) {
	plain := NewForm()
	plain.SetTheme("plain")

	html, err := plain.formRender(&requiredLabelModel{}, nil)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	if !strings.Contains(string(html), `<span class="sr-only">(required)</span>`) {
		t.Fatalf("expected English required text, got: %s", html)
	}

	translated := NewTranslatedForm(optionTranslate)
	translated.SetTheme("plain")

	html, err = translated.formRenderLocalized(testLocalizer{Locale: "it"}, &requiredLabelModel{}, nil)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	if !strings.Contains(string(html), `<span class="sr-only">T(form||(required))</span>`) {
		t.Fatalf("expected translated required text, got: %s", html)
	}
}

type requiredSelectModel struct {
	Single   SortedSelect[string]      `form:"dropdown" required:"true"`
	Number   SortedSelect[int]         `form:"dropdown" required:"true"`
	Pointer  *SortedSelect[string]     `form:"dropdown" required:"true"`
	Multiple SortedMultiSelect[string] `form:"multicheckbox" required:"true"`
	Optional SortedSelect[string]      `form:"dropdown"`
}

func newRequiredSelectModel() *requiredSelectModel {
	choices := map[string]string{"": "—", "a": "A"}
	pointer := NewSortedSelect(choices)

	return &requiredSelectModel{
		Single:   NewSortedSelect(choices),
		Number:   NewSortedSelect(map[int]string{1: "One", 2: "Two"}),
		Pointer:  &pointer,
		Multiple: NewSortedMultiSelect(map[string]string{"a": "A", "b": "B"}),
		Optional: NewSortedSelect(choices),
	}
}

func requiredFields(errs FieldErrors) map[string]string {
	fields := map[string]string{}

	for _, err := range errs {
		field, message := err.FieldError()
		fields[field] = message
	}

	return fields
}

// TestRequiredSelect checks that required:"true" on selects needs a choice:
// the empty key and keys that are not options count as missing, and the
// error uses the same key as other required fields.
func TestRequiredSelect(t *testing.T) {
	f := NewTranslatedForm(optionTranslate)

	empty := newRequiredSelectModel()

	missing := requiredFields(f.ValidateFormLocalized(empty, testLocalizer{Locale: "it"}))
	for _, field := range []string{"Single", "Number", "Pointer", "Multiple"} {
		if missing[field] != "T("+TranslationKeyRequired+")" {
			t.Errorf("%s: expected the required error, got %q (all: %v)", field, missing[field], missing)
		}
	}

	if _, ok := missing["Optional"]; ok {
		t.Errorf("optional select reported missing: %v", missing)
	}

	chosen := newRequiredSelectModel()
	_ = chosen.Single.Set("a")
	_ = chosen.Number.Set(2)
	_ = chosen.Pointer.Set("a")
	_ = chosen.Multiple.Set([]string{"b"})

	if errs := f.ValidateForm(chosen); len(errs) != 0 {
		t.Fatalf("selected choices reported missing: %v", errs)
	}

	// Zero is a choice when the select offers it.
	zero := struct {
		Number SortedSelect[int] `form:"dropdown" required:"true"`
	}{Number: NewSortedSelect(map[int]string{0: "None", 1: "One"})}

	if errs := f.ValidateForm(&zero); len(errs) != 0 {
		t.Fatalf("zero option reported missing: %v", errs)
	}

	// An optional select without a choice is neither missing nor invalid.
	optional := struct {
		Number SortedSelect[int] `form:"dropdown"`
	}{Number: NewSortedSelect(map[int]string{1: "One"})}

	if errs := f.ValidateForm(&optional); len(errs) != 0 {
		t.Fatalf("optional select without a choice: %v", errs)
	}
}

// TestRequiredSelectAfterMapForm checks the posted form: choosing the
// placeholder or nothing reports the required error.
func TestRequiredSelectAfterMapForm(t *testing.T) {
	model := newRequiredSelectModel()
	_ = model.Single.Set("a")

	_ = model.Pointer.Set("a")

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url.Values{
		"Single": {""},
		"Number": {"1"},
	}.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err := request.ParseForm(); err != nil {
		t.Fatal(err)
	}

	if err := MapForm(request, model); err != nil {
		t.Fatal(err)
	}

	missing := requiredFields(NewForm().ValidateForm(model))
	if len(missing) != 2 || missing["Single"] == "" || missing["Multiple"] == "" {
		t.Fatalf("expected Single and Multiple missing, got %v", missing)
	}
}

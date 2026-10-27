package form

import (
	"strings"
	"testing"
)

// keys lists the option keys of a select in order.
func keys(mapper SortedMapper) string {
	var out []string
	for _, entry := range mapper.SortedMapper() {
		out = append(out, entry.Key())
	}

	return strings.Join(out, ",")
}

func TestOrderedSelect(t *testing.T) {
	status := NewOrderedSelect(
		Option[string]{Value: "open", Label: "Open"},
		Option[string]{Value: "paid", Label: "Paid"},
		Option[string]{Value: "cancelled", Label: "Cancelled"},
	).Selected("paid")

	if got := keys(status); got != "open,paid,cancelled" {
		t.Errorf("order: %s", got)
	}

	if status.Get() != "paid" {
		t.Errorf("selected: %q", status.Get())
	}

	if err := status.SetFromKey("cancelled"); err != nil || status.Get() != "cancelled" {
		t.Errorf("SetFromKey: %v %q", err, status.Get())
	}

	if NewOrderedSelect(Option[int]{1, "a"}).Selected(9).Get() != 0 {
		t.Error("a value that is not an option was selected")
	}

	status.SetSource(map[string]string{"b": "B", "a": "A"})

	if got := keys(status); got != "a,b" {
		t.Errorf("a new source map is not sorted by key: %s", got)
	}
}

func TestSortByLabel(t *testing.T) {
	files := NewSortedSelect(map[string]string{"c1": "beach.jpg", "a9": "Zebra.png", "b2": "apple.gif"}).SortByLabel()

	if got := keys(files); got != "b2,c1,a9" {
		t.Errorf("by label: %s", got)
	}

	tags := NewSortedMultiSelect(map[int]string{1: "Zeta", 2: "alpha"}).SortByLabel().Selected(1, 2)
	if got := keys(tags); got != "2,1" {
		t.Errorf("multi by label: %s", got)
	}

	if len(tags.Get()) != 2 {
		t.Errorf("multi selected: %v", tags.Get())
	}

	ordered := NewOrderedMultiSelect(Option[string]{"x", "X"}, Option[string]{"a", "A"})
	if got := keys(ordered); got != "x,a" {
		t.Errorf("ordered multi: %s", got)
	}
}

func TestOrderedSelectRenders(t *testing.T) {
	type Input struct {
		Parent SortedSelect[int] `form:"dropdown" label:"Parent" placeholder:"—"`
	}

	input := Input{Parent: NewOrderedSelect(Option[int]{3, "Home"}, Option[int]{1, "About"}).Selected(1)}
	html := renderString(t, NewForm(), input, nil)

	home := strings.Index(html, ">Home<")
	about := strings.Index(html, ">About<")
	placeholder := strings.Index(html, `<option value="" >—</option>`)

	if placeholder < 0 || placeholder > home || home > about {
		t.Errorf("options out of order: %s", html)
	}

	if !strings.Contains(html, `<option value="1" selected`) {
		t.Errorf("selected option: %s", html)
	}
}

type translatedSortInput struct {
	Module SortedSelect[string] `form:"dropdown" name:"module" label:"Module" translate:"true"`
	Plain  SortedSelect[string] `form:"dropdown" name:"plain"  label:"Plain"`
}

// reverseTranslate translates the labels so their order flips: "Alpha"
// becomes "Zulu" and "Zulu" becomes "Alpha".
func reverseTranslate(_ Localizer, key string, _ ...any) string {
	switch key {
	case "Alpha":
		return "Zulu"
	case "Zulu":
		return "Alpha"
	}

	return key
}

func TestSortByLabelSortsTranslatedLabels(t *testing.T) {
	f := NewTranslatedForm(reverseTranslate)
	f.SetTheme("plain")

	options := map[string]string{"a": "Alpha", "m": "Mike", "z": "Zulu"}
	input := translatedSortInput{
		Module: NewSortedSelect(options).SortByLabel(),
		Plain:  NewSortedSelect(options).SortByLabel(),
	}

	html, err := f.Field(testLocalizer{Locale: "nl"}, input, "module", nil)
	if err != nil {
		t.Fatalf("Field: %v", err)
	}

	out := string(html)
	if strings.Index(out, ">Alpha<") >= strings.Index(out, ">Mike<") ||
		strings.Index(out, ">Mike<") >= strings.Index(out, ">Zulu<") {
		t.Errorf("the translated labels are not in order:\n%s", out)
	}

	if strings.Index(out, `value="z"`) >= strings.Index(out, `value="a"`) {
		t.Errorf("z (translated Alpha) does not come before a (translated Zulu):\n%s", out)
	}

	// Labels that are not translated keep the order of SortByLabel.
	html, err = f.Field(testLocalizer{Locale: "nl"}, input, "plain", nil)
	if err != nil {
		t.Fatalf("Field: %v", err)
	}

	out = string(html)
	if strings.Index(out, `value="a"`) >= strings.Index(out, `value="z"`) {
		t.Errorf("untranslated labels changed order:\n%s", out)
	}
}

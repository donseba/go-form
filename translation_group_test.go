package form

import (
	"strings"
	"testing"

	"github.com/donseba/go-form/templates"
)

type translatedGroupModel struct {
	Info    `target:"/submit" method:"post"`
	Address struct {
		Street string
	} `legend:"Address Information"`
}

func TestTranslatedForm_GroupLegendIsTranslated(t *testing.T) {
	translate := func(loc Localizer, key string, args ...any) string {
		_ = loc
		_ = args
		return "T(" + key + ")"
	}

	f := NewTranslatedForm(templates.Plain, translate)

	html, err := f.formRenderLocalized(testLocalizer{Locale: "it"}, &translatedGroupModel{}, nil)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	if !strings.Contains(string(html), "T(Address Information)") {
		t.Fatalf("expected translated group legend in HTML, got: %s", html)
	}
}

package templates_test

import (
	"testing"

	"github.com/donseba/go-form/v2/templates"
	"github.com/donseba/go-form/v2/types"
)

func TestRegisterClassTheme(t *testing.T) {
	classes := templates.ThemeClasses{Input: templates.StyleOption{Class: "control"}}

	theme, err := templates.RegisterClassTheme("class-theme-test", classes)
	if err != nil {
		t.Fatal(err)
	}

	if theme.Templates == nil || theme.Templates.Lookup("input") == nil {
		t.Fatal("templates not loaded")
	}

	again, err := templates.RegisterClassTheme("class-theme-test", classes)
	if err != nil || again != theme {
		t.Errorf("registering the same theme again replaced it: %v", err)
	}

	other, err := templates.RegisterClassTheme(
		"class-theme-test",
		templates.ThemeClasses{Input: templates.StyleOption{Class: "field"}},
	)
	if err != nil || other == theme || other.Classes.Input.Class != "field" {
		t.Errorf("other classes did not replace the theme: %v", err)
	}

	if found, _ := templates.GetTheme("class-theme-test"); found != other {
		t.Error("the replacement is not registered")
	}
}

func TestDescribedBy(t *testing.T) {
	field := types.FormField{Id: "email", Description: "Your address", Errors: []string{"a", "b"}}
	if got := templates.DescribedBy(field); got != "email_description email_error_0 email_error_1" {
		t.Errorf("described by: %q", got)
	}

	if templates.DescribedBy(types.FormField{Id: "x"}) != "" {
		t.Error("a field without description or errors is described")
	}
}

func TestClassList(t *testing.T) {
	for _, c := range []struct {
		in   []string
		want string
	}{
		{nil, ""},
		{[]string{"", " "}, ""},
		{[]string{"control", ""}, "control"},
		{[]string{" control ", "big"}, "control big"},
	} {
		if got := templates.ClassList(c.in...); got != c.want {
			t.Errorf("ClassList(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

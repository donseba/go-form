package templates_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/donseba/go-form/templates"
	"github.com/donseba/go-form/types"
)

type dummyLoc struct{}

func (dummyLoc) GetLocale() string { return "en" }

func TestGoHTML_InputTemplate_RendersFileInputWithoutValueAndWithFileClass(t *testing.T) {
	if err := templates.InitThemes(); err != nil {
		t.Fatalf("InitThemes: %v", err)
	}

	theme, ok := templates.GetTheme("bootstrap")
	if !ok {
		t.Fatal("bootstrap theme not registered")
	}
	if theme.Templates == nil {
		t.Fatal("bootstrap theme templates not loaded")
	}

	data := map[string]any{
		"Type": "file",
		"Field": types.FormField{
			Id:          "upload",
			Name:        "upload",
			Value:       "SHOULD_NOT_RENDER",
			Placeholder: "",
			Required:    false,
			Min:         "",
			Max:         "",
			Step:        "",
			Description: "",
			Data:        map[string]string{},
		},
		"Loc": dummyLoc{},
	}

	var buf bytes.Buffer
	if err := theme.Templates.ExecuteTemplate(&buf, "input", data); err != nil {
		t.Fatalf("execute template: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "type=\"file\"") {
		t.Fatalf("expected file input type, got: %s", out)
	}
	if strings.Contains(out, "value=") {
		t.Fatalf("file inputs must not render a value attribute, got: %s", out)
	}

	// Ensure the file slot is used (bootstrap defines it as `form-control form-control-sm`).
	if !strings.Contains(out, "form-control") {
		t.Fatalf("expected bootstrap file class to be present, got: %s", out)
	}
	if !strings.Contains(out, "form-control-sm") {
		t.Fatalf("expected bootstrap file class (form-control-sm) to be present, got: %s", out)
	}
}

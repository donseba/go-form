package form

import (
	"strings"
	"testing"
)

func TestCancelButton_RenderedOnlyWhenConfigured(t *testing.T) {
	type WithCancel struct {
		Info `target:"/submit" method:"post"`
		Text string `name:"text" form:"input,text" label:"Text"`
	}

	// no cancel by default
	{
		f := NewForm()
		f.SetTheme("plain")
		html, err := f.formRenderFunc(&DefaultLocalizer{}, WithCancel{Text: "x"}, nil)
		if err != nil {
			t.Fatalf("render failed: %v", err)
		}
		out := string(html)
		if strings.Contains(out, "href=\"/back\"") {
			t.Fatalf("cancel link should not be rendered when CancelTarget is empty")
		}
	}

	// cancel enabled
	{
		m := WithCancel{Text: "x"}
		m.Info.CancelTarget = "/back"
		m.Info.CancelText = "Go back"

		themeSets := []struct {
			name  string
			theme string
		}{
			{"Plain", "plain"},
			{"Bootstrap", "bootstrap"},
			{"TailwindV3", "tailwind"},
			{"TailwindV4", "tailwindv4"},
		}

		for _, ts := range themeSets {
			ts := ts
			t.Run(ts.name, func(t *testing.T) {
				f := NewForm()
				f.SetTheme(ts.theme)

				html, err := f.formRenderFunc(&DefaultLocalizer{}, m, nil)
				if err != nil {
					t.Fatalf("render failed: %v", err)
				}
				out := string(html)
				if !strings.Contains(out, "href=\"/back\"") {
					t.Fatalf("expected cancel link to be rendered")
				}
				if !strings.Contains(out, "Go back") {
					t.Fatalf("expected cancel text to be rendered")
				}
			})
		}
	}
}

func TestCancelRendersForThemes(t *testing.T) {
	themeSets := []struct {
		name  string
		theme string
	}{
		{"Plain", "plain"},
		{"Bootstrap", "bootstrap"},
		{"TailwindV3", "tailwind"},
		{"TailwindV4", "tailwindv4"},
	}

	for _, ts := range themeSets {
		ts := ts
		t.Run(ts.name, func(t *testing.T) {
			type F struct {
				Info
				Name string `form:"input,text" label:"Name"`
			}

			m := F{}
			m.Info.CancelTarget = "/cancel"
			m.Info.CancelText = "Cancel"

			f := NewForm()
			f.SetTheme(ts.theme)

			html, err := f.formRender(m, nil)
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			if !strings.Contains(string(html), "Cancel") {
				t.Fatalf("expected cancel text in output")
			}
		})
	}
}

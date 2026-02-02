package form

import (
	"strings"
	"testing"

	"github.com/donseba/go-form/templates"
	"github.com/donseba/go-form/types"
)

func TestCancelButton_RenderedOnlyWhenConfigured(t *testing.T) {
	type WithCancel struct {
		Info `target:"/submit" method:"post"`
		Text string `name:"text" form:"input,text" label:"Text"`
	}

	// no cancel by default
	{
		f := NewForm(templates.Plain)
		html, err := f.formRenderFunc(&DefaultLocalizer{}, WithCancel{Text: "x"}, nil)
		if err != nil {
			t.Fatalf("render failed: %v", err)
		}
		if strings.Contains(string(html), "cancel") || strings.Contains(string(html), "Cancel") {
			// There may be other places, so check for href which we only add for cancel.
			if strings.Contains(string(html), "href=\"/back\"") {
				t.Fatalf("cancel link should not be rendered when CancelTarget is empty")
			}
		}
	}

	// cancel enabled
	{
		m := WithCancel{Text: "x"}
		m.Info.CancelTarget = "/back"
		m.Info.CancelText = "Go back"

		templateSets := []struct {
			name string
			tm   types.TemplateMap
		}{
			{"Plain", templates.Plain},
			{"BootstrapV5", templates.BootstrapV5},
			{"TailwindV3", templates.TailwindV3},
			{"TailwindV4", templates.TailwindV4},
		}

		for _, ts := range templateSets {
			ts := ts
			t.Run(ts.name, func(t *testing.T) {
				f := NewForm(ts.tm)
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

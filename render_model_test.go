package form

import (
	"context"
	"strings"
	"testing"

	"github.com/donseba/go-form/v2/csrf"
	"github.com/donseba/go-form/v2/types"
)

type externalSettings struct {
	APIKey string `name:"api_key" label:"API key"`
}

func TestWithInfoKeepsExternalModelFieldsFlat(t *testing.T) {
	wrapped := WithInfo(&externalSettings{APIKey: "secret"}, Info{
		Target:     "/settings",
		Method:     "post",
		SubmitText: "Save",
	})

	transformer, err := NewTransformer(wrapped)
	if err != nil {
		t.Fatal(err)
	}
	if len(transformer.Fields) != 2 {
		t.Fatalf("got %d fields, want 2", len(transformer.Fields))
	}
	if transformer.Fields[0].Type != types.FieldTypeForm || transformer.Fields[0].Target != "/settings" {
		t.Fatalf("unexpected form field: %#v", transformer.Fields[0])
	}
	if transformer.Fields[1].Name != "api_key" || transformer.Fields[1].Value != "secret" {
		t.Fatalf("unexpected model field: %#v", transformer.Fields[1])
	}
}

func TestWithContextInfoInjectsCSRF(t *testing.T) {
	ctx := context.WithValue(context.Background(), csrf.CSRFTokenContextKey, "token-value")
	wrapped := WithContextInfo(ctx, externalSettings{}, Info{Target: "/settings"})

	if wrapped.Info.CsrfField != DefaultCSRFField || wrapped.Info.CsrfValue != "token-value" {
		t.Fatalf("unexpected CSRF info: %#v", wrapped.Info)
	}

	f := NewForm()
	f.SetTheme("plain")
	html, err := f.formRender(wrapped, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(html), `name="_csrf" value="token-value"`) {
		t.Fatalf("rendered form does not contain CSRF field: %s", html)
	}
}

func TestWithInfoRejectsNilModelDuringTransform(t *testing.T) {
	_, err := NewTransformer(WithInfo(nil, Info{}))
	if err == nil {
		t.Fatal("expected nil model error")
	}
}

func TestWithInfoOverridesEmbeddedInfo(t *testing.T) {
	type settings struct {
		Info
		Name string
	}
	transformer, err := NewTransformer(WithInfo(settings{
		Info: Info{Target: "/old"},
		Name: "Dashboard",
	}, Info{Target: "/new"}))
	if err != nil {
		t.Fatal(err)
	}
	if len(transformer.Fields) != 2 || transformer.Fields[0].Target != "/new" {
		t.Fatalf("unexpected fields: %#v", transformer.Fields)
	}
}

func TestTransformerAcceptsEmptyStruct(t *testing.T) {
	transformer, err := NewTransformer(struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	if len(transformer.Fields) != 0 {
		t.Fatalf("got %d fields, want none", len(transformer.Fields))
	}
}

func TestHasFieldsUnderstandsRenderModels(t *testing.T) {
	if HasFields(WithInfo(struct{}{}, Info{})) {
		t.Fatal("metadata-only model must not report fields")
	}
	if !HasFields(WithInfo(externalSettings{}, Info{})) {
		t.Fatal("external model field was not detected")
	}
}

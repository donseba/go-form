package form

import (
	"reflect"
	"strings"
	"testing"
)

func TestTaggedValidationErrorsRender(t *testing.T) {
	type profile struct {
		Name string `form:"input,text" name:"display_name" label:"Name" required:"true"`
	}
	for _, theme := range []string{"plain", "bootstrap", "tailwind", "tailwindv4"} {
		t.Run(theme, func(t *testing.T) {
			f := NewTranslatedForm(func(_ Localizer, key string, _ ...any) string {
				if key == TranslationKeyRequired {
					return "is verplicht"
				}
				return key
			})
			f.SetTheme(theme)
			model := profile{}
			errs := f.ValidateFormLocalized(&model, &DefaultLocalizer{})
			if len(errs) != 1 {
				t.Fatalf("expected required error, got %v", errs)
			}
			if name, _ := errs[0].FieldError(); name != "display_name" {
				t.Fatalf("error field = %q, want display_name", name)
			}
			out, err := f.formRenderLocalized(&DefaultLocalizer{}, WithInfo(&model, Info{Target: "/profile", Method: "post"}), errs)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(out), `name="display_name"`) || !strings.Contains(string(out), "is verplicht") {
				t.Fatalf("field error missing from rendered form: %s", out)
			}
		})
	}
}

func TestTaggedNestedValidationFieldNames(t *testing.T) {
	type address struct {
		City string `form:"input,text" name:"city_name" required:"true"`
	}
	type profile struct {
		Address address  `name:"shipping"`
		Billing *address `name:"billing"`
	}
	model := profile{Billing: &address{}}
	f := NewForm()
	errs := f.ValidateForm(&model)
	if len(errs) != 2 {
		t.Fatalf("expected two required errors, got %v", errs)
	}
	for i, want := range []string{"shipping.city_name", "billing.city_name"} {
		if name, _ := errs[i].FieldError(); name != want {
			t.Errorf("field %d = %q, want %q", i, name, want)
		}
	}
}

func TestCustomValidationUsesTaggedFieldName(t *testing.T) {
	type profile struct {
		Name string `form:"input,text" name:"display_name" validate:"reject"`
	}
	f := NewForm()
	f.RegisterValidationMethod("reject", func(_ any, field reflect.StructField) FieldErrors {
		if field.Name != "Name" {
			t.Errorf("custom validator received renamed struct field %q", field.Name)
		}
		return FieldErrors{FieldValidationError{Field: field.Name, Err: "custom rejection"}}
	})
	errs := f.ValidateForm(&profile{})
	if len(errs) != 1 {
		t.Fatalf("expected one custom error, got %v", errs)
	}
	if name, _ := errs[0].FieldError(); name != "display_name" {
		t.Fatalf("custom error field = %q, want display_name", name)
	}
}

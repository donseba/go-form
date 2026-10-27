package form

import (
	"fmt"
	"strings"
	"testing"
)

// TestOwnMessagesShowInEnglish checks that each of go-form's messages shows
// its English text, not its key, when nothing translates it: without a
// translation function, and with one whose catalogue lacks the message.
func TestOwnMessagesShowInEnglish(t *testing.T) {
	english := catalogueEntries(t, "en.po")
	delete(english, "")

	translators := map[string]*Form{
		"no translation function": NewForm(),
		"key as is": NewTranslatedForm(func(_ Localizer, key string, args ...any) string {
			return key
		}),
		"without context": NewTranslatedForm(func(_ Localizer, key string, args ...any) string {
			return strings.TrimPrefix(key, "form||")
		}),
		"formatted without context": NewTranslatedForm(func(_ Localizer, key string, args ...any) string {
			if len(args) == 0 {
				return strings.TrimPrefix(key, "form||")
			}

			return fmt.Sprintf(strings.TrimPrefix(key, "form||"), args...)
		}),
	}

	for name, f := range translators {
		for key, text := range english {
			var args []any
			if strings.Contains(key, "%") {
				args = []any{"x"}
				if strings.Contains(key, "%T") {
					args = []any{1, "x"}
				}
			}

			want := text
			if len(args) > 0 {
				want = fmt.Sprintf(text, args...)
			}

			if got := f.validationErrorTranslated(&DefaultLocalizer{}, key, args...); got != want {
				t.Errorf("%s: message %q: got %q, want %q", name, key, got, want)
			}

			if got := f.translate(&DefaultLocalizer{}, key, args); got != want {
				t.Errorf("%s: text %q: got %q, want %q", name, key, got, want)
			}
		}
	}
}

func TestTranslatedMessagesStayTranslated(t *testing.T) {
	f := NewTranslatedForm(func(_ Localizer, key string, _ ...any) string {
		if key == TranslationKeyRequired {
			return "Verplicht"
		}

		return key
	})

	if got := f.validationErrorTranslated(&DefaultLocalizer{}, TranslationKeyRequired); got != "Verplicht" {
		t.Errorf("translated message: %q", got)
	}

	// Texts of the application are not go-form's: they show as written.
	if got := f.translate(&DefaultLocalizer{}, "Validation required", nil); got != "Validation required" {
		t.Errorf("text of the application: %q", got)
	}
}

func TestRequiredMessageInEnglish(t *testing.T) {
	type input struct {
		Name string `form:"input,text" name:"name" required:"true"`
	}

	errs := NewForm().ValidateForm(&input{})
	if got := errs.Messages("name"); len(got) != 1 || got[0] != "This field is required." {
		t.Errorf("required: %q", got)
	}
}

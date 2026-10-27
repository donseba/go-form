package form

import (
	"bufio"
	"bytes"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/donseba/go-form/v2/types"
)

// englishTexts are the English texts of go-form's own messages, keyed by
// their translation keys (the TranslationKey… values and "Cancel"), from
// go-form's English catalogue: "form||Validation required" is "This field
// is required.".
var englishTexts = sync.OnceValue(func() map[string]string {
	return catalogueTexts("en")
})

// ownTexts are the texts of go-form's catalogues read so far, by file name.
var ownTexts sync.Map

// catalogueTexts are the texts of go-form's catalogue that best matches
// language (see catalogueFile), keyed by their translation keys, without
// those that are not translated (an empty msgstr). It is nil when go-form
// has no catalogue for language.
func catalogueTexts(language string) map[string]string {
	name, ok := catalogueFile(language)
	if !ok {
		return nil
	}

	if texts, ok := ownTexts.Load(name); ok {
		return texts.(map[string]string)
	}

	data, err := locales.ReadFile(name)
	if err != nil {
		panic(err) // The catalogue is embedded.
	}

	texts, _ := ownTexts.LoadOrStore(name, parseCatalogue(name, data))

	return texts.(map[string]string)
}

// parseCatalogue reads one of go-form's catalogues, named name: the msgid
// and msgstr of each message, on one line each.
func parseCatalogue(name string, data []byte) map[string]string {
	var err error

	texts := map[string]string{}
	id := ""

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "msgid "):
			id, err = strconv.Unquote(strings.TrimPrefix(line, "msgid "))
		case strings.HasPrefix(line, "msgstr ") && id != "":
			var text string

			text, err = strconv.Unquote(strings.TrimPrefix(line, "msgstr "))
			if text != "" {
				texts[id] = text
			}
		}

		if err != nil {
			panic(fmt.Sprintf("form: %s: %v", name, err)) // The catalogue is go-form's own.
		}
	}

	return texts
}

// ownText returns translated, the translation of key with args, unless
// key is one of go-form's own messages that was not translated: when there
// is no translation function, or the function returned the key as it is
// (with or without its "form||" context and args filled in), as a
// translator does for a message missing from its catalogue. Then it is the
// text of go-form's catalogue for the language of loc (GetLocale, such as
// "nl" or "nl_NL"; see catalogueFile), or the English text when go-form
// has no catalogue for that language, with args filled in: "Dit veld is
// verplicht." or "This field is required." for "form||Validation
// required". A nil loc shows the English text.
func ownText(loc types.Localizer, key string, args []any, translated string) string {
	english, ok := englishTexts()[key]
	if !ok || !untranslated(key, args, translated) {
		return translated
	}

	text := english

	if loc != nil {
		if own, ok := catalogueTexts(loc.GetLocale())[key]; ok {
			text = own
		}
	}

	if len(args) > 0 {
		return fmt.Sprintf(text, args...)
	}

	return text
}

// untranslated reports whether translated is key as it is: with or without
// the "form||" context, with or without args filled in.
func untranslated(key string, args []any, translated string) bool {
	written := []string{key, withoutContext(key)}
	if len(args) > 0 {
		written = append(written, fmt.Sprintf(key, args...), fmt.Sprintf(withoutContext(key), args...))
	}

	return slices.Contains(written, translated)
}

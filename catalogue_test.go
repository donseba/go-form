package form

import (
	"bufio"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// catalogueEntries reads the msgid → msgstr pairs of a PO file of
// go-form's catalogues (locales).
func catalogueEntries(t *testing.T, name string) map[string]string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("locales", name))
	if err != nil {
		t.Fatal(err)
	}

	entries := map[string]string{}
	id := ""

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "msgid "):
			id, err = strconv.Unquote(strings.TrimPrefix(line, "msgid "))
		case strings.HasPrefix(line, "msgstr "):
			entries[id], err = strconv.Unquote(strings.TrimPrefix(line, "msgstr "))
		}

		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}

	return entries
}

// TestCatalogues checks that go-form's catalogues hold every message of
// go-form, translated in each language.
func TestCatalogues(t *testing.T) {
	keys := []string{
		TranslationKeyRequired, TranslationKeyMin, TranslationKeyMax,
		TranslationKeyMaxLength, TranslationKeyMinLength, TranslationKeyInvalidValue,
		TranslationKeyInvalidEmail, TranslationKeyInvalidEnum, TranslationKeyInvalidMapper,
		TranslationKeyInvalidSortedMapper, TranslationKeyPattern, TranslationKeyURL,
		TranslationKeyBool, TranslationKeyZero, TranslationKeyMinItems,
		TranslationKeyMaxItems, TranslationKeyPrefix, TranslationKeySuffix,
		TranslationKeyContains, TranslationKeyStep, TranslationKeyStepFromMin, TranslationKeyCSRFTokenMissing,
		TranslationKeyCSRFTokenInvalid, TranslationKeyCSRFTokenError,
		TranslationKeySortedSelectNotFound, TranslationKeySortedSelectKeyNotFound,
		TranslationKeySortedSelectSourceNil, TranslationKeySortedSelectTypeError,
		TranslationKeySortedSelectUnmarshalNotFound, TranslationKeyRequiredLabel,
		TranslationKeyEqual, TranslationKeyAccept, TranslationKeyMaxSize,
		TranslationKeyInvalid, TranslationKeyLine,
		"Cancel",
	}

	template := catalogueEntries(t, "translations.pot")

	for _, name := range []string{"en_US.po", "nl_NL.po", "en.po", "nl.po"} {
		entries := catalogueEntries(t, name)

		if !strings.Contains(entries[""], "Language: "+strings.TrimSuffix(name, ".po")+"\n") {
			t.Errorf("%s: no language header: %q", name, entries[""])
		}

		for _, key := range keys {
			if entries[key] == "" {
				t.Errorf("%s misses %q", name, key)
			}

			if strings.Count(entries[key], "%") != strings.Count(key, "%") {
				t.Errorf("%s: %q does not keep the verbs of %q", name, entries[key], key)
			}
		}

		if len(entries) != len(template) {
			t.Errorf("%s has %d entries, the template %d", name, len(entries), len(template))
		}
	}

	for _, key := range keys {
		if _, ok := template[key]; !ok {
			t.Errorf("translations.pot misses %q", key)
		}
	}
}

// TestLanguageCatalogues checks that the catalogues per language only hold
// the same translations as those per region.
func TestLanguageCatalogues(t *testing.T) {
	for language, region := range map[string]string{"en.po": "en_US.po", "nl.po": "nl_NL.po"} {
		entries := catalogueEntries(t, language)
		regional := catalogueEntries(t, region)

		delete(entries, "")
		delete(regional, "")

		if !reflect.DeepEqual(entries, regional) {
			t.Errorf("%s and %s differ", language, region)
		}
	}
}

// TestCatalogueFile checks which of go-form's catalogues its messages
// show from for a language name.
func TestCatalogueFile(t *testing.T) {
	cases := []struct {
		language string
		want     string
	}{
		{"nl", "locales/nl.po"},
		{"nl_NL", "locales/nl_NL.po"},
		{"nl-BE", "locales/nl.po"},
		{"NL_be", "locales/nl.po"},
		{"en_GB", "locales/en.po"},
		{"en_US", "locales/en_US.po"},
	}

	for _, c := range cases {
		if name, ok := catalogueFile(c.language); !ok || name != c.want {
			t.Errorf("%s: %q %v, want %q", c.language, name, ok, c.want)
		}
	}

	for _, language := range []string{"", "de", "fr_FR", "../nl", "nl/x"} {
		if _, ok := catalogueFile(language); ok {
			t.Errorf("%q: a catalogue", language)
		}
	}
}

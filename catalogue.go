package form

import (
	"embed"
	"io/fs"
	"strings"
)

// locales are go-form's catalogues of its own messages, per language and
// region (en_US.po, nl_NL.po) and per language only (en.po, nl.po): the
// texts that show when nothing translates a message (see ownText).
//
//go:embed locales/*.po
var locales embed.FS

// catalogueFile is the path of go-form's catalogue that best matches
// language, such as "nl", "nl_NL", "nl-BE" or "en_GB": the catalogue of
// that language and region, else the one of the language alone, such as
// "locales/nl_NL.po" for nl-NL or "locales/nl.po" for nl_BE. It reports
// false when go-form has no catalogue for the language.
func catalogueFile(language string) (string, bool) {
	normalized := strings.ReplaceAll(strings.TrimSpace(language), "-", "_")
	if normalized == "" || strings.ContainsAny(normalized, `/\.`) {
		return "", false
	}

	base, region, _ := strings.Cut(normalized, "_")
	base = strings.ToLower(base)

	candidates := []string{base}
	if region != "" {
		candidates = []string{base + "_" + strings.ToUpper(region), base}
	}

	for _, candidate := range candidates {
		name := "locales/" + candidate + ".po"
		if _, err := fs.Stat(locales, name); err == nil {
			return name, true
		}
	}

	return "", false
}

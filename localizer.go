package form

import "github.com/donseba/go-form/types"

// Localizer is the public interface used by go-form for translations.
//
// It is an alias to types.Localizer to keep the API ergonomic and backwards-compatible
// with code that expects form.Localizer.
type Localizer = types.Localizer

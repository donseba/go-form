package form

import (
	"time"

	"github.com/donseba/go-form/v2/types"
)

// Localizer is the public interface used by go-form for translations.
//
// It is an alias to types.Localizer to keep the API ergonomic and backwards-compatible
// with code that expects form.Localizer.
type Localizer = types.Localizer

// TimeLocalizer is a Localizer that also names the time zone people enter
// date-times in, usually the site's. Rendering with it shows the value of
// a datetime-local input in that zone; map the submission with
// MapFormWith(r, dst, MapOptions{Location: …}) in the same zone. Values of
// date, time, month and week inputs hold no instant and are shown as stored.
type TimeLocalizer interface {
	Localizer
	Location() *time.Location
}

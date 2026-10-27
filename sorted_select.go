package form

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// SortedSelect is a generic helper that holds a typed key value and a map of key->label
// It implements SortedMapper (value receiver) and a SetFromKey method (pointer receiver)
// so MapForm can set the selected value from the posted key string.
// T must be comparable to be used as map key.
type SortedSelect[T comparable] struct {
	value  T
	source map[T]string
	order  selectOrder[T]
}

// entry implements SortedMap for SortedSelect
type ssEntry struct {
	k string
	v string
}

func NewSortedSelect[T comparable](source map[T]string) SortedSelect[T] {
	return SortedSelect[T]{source: source}
}

// NewOrderedSelect returns a select that lists its options in the order
// given, such as pages in tree order or statuses in their natural order:
//
//	status := form.NewOrderedSelect(
//		form.Option[string]{Value: "open", Label: "Open"},
//		form.Option[string]{Value: "paid", Label: "Paid"},
//	).Selected(current)
func NewOrderedSelect[T comparable](options ...Option[T]) SortedSelect[T] {
	source, keys := optionsFrom(options)

	return SortedSelect[T]{source: source, order: selectOrder[T]{keys: keys}}
}

// SortByLabel returns the select listing its options alphabetically by
// label (ignoring case) instead of by key, such as files by name rather than
// by ID.
func (ss SortedSelect[T]) SortByLabel() SortedSelect[T] {
	ss.order.byLabel = true

	return ss
}

// SortedByLabel reports whether the select lists its options by label
// (SortByLabel). With translate:"true" the form sorts the options again by
// their translated labels when it renders.
func (ss SortedSelect[T]) SortedByLabel() bool {
	return ss.order.byLabel
}

// Selected returns the select with val selected, for chaining after
// NewSortedSelect or NewOrderedSelect. A value that is not one of the options
// leaves nothing selected, as Set does.
func (ss SortedSelect[T]) Selected(val T) SortedSelect[T] {
	_ = ss.Set(val)

	return ss
}

func (e ssEntry) Key() string   { return e.k }
func (e ssEntry) Value() string { return e.v }

// Get returns the current value of the SortedSelect field (typed as T)
func (ss SortedSelect[T]) Get() T {
	return ss.value
}

// Set sets the value of the SortedSelect field, ensuring it exists in Source if Source is not nil
func (ss *SortedSelect[T]) Set(val T) error {
	if ss.source != nil {
		found := false
		for k := range ss.source {
			if fmt.Sprint(k) == fmt.Sprint(val) {
				found = true
				break
			}
		}
		if !found {
			var zero T
			ss.value = zero
			return SortedSelectError{Key: TranslationKeySortedSelectNotFound, Args: []any{val}}
		}
	}

	ss.value = val
	return nil
}

// SortedMapper returns the options sorted by key string representation
// (fmt.Sprint), in the given order for NewOrderedSelect, or by label after
// SortByLabel.
func (ss SortedSelect[T]) SortedMapper() []SortedMap {
	return ss.order.entries(ss.source)
}

func (ss *SortedSelect[T]) Scan(val any) error {
	if ss == nil {
		return nil
	}
	v := reflect.ValueOf(val)
	if !v.IsValid() {
		var zero T
		ss.value = zero
		return nil
	}
	var zero T
	typ := reflect.TypeOf(zero)
	var assigned T
	if v.Type().AssignableTo(typ) {
		assigned = v.Interface().(T)
	} else if v.Type().ConvertibleTo(typ) {
		assigned = v.Convert(typ).Interface().(T)
	} else {
		return fmt.Errorf("%s", TranslationKeySortedSelectTypeError)
	}
	found := false
	for k := range ss.source {
		if fmt.Sprint(k) == fmt.Sprint(assigned) {
			found = true
			break
		}
	}
	if !found {
		ss.value = zero
		return SortedSelectError{Key: TranslationKeySortedSelectNotFound, Args: []any{assigned}}
	}
	ss.value = assigned
	return nil
}

func (ss *SortedSelect[T]) Value() (any, error) {
	return ss.value, nil
}

// String returns the current value as a string using fmt.Sprint (no converters required)
func (ss SortedSelect[T]) String() string {
	return fmt.Sprint(ss.value)
}

// SetFromKey sets the value from a string key (as submitted by forms)
func (ss *SortedSelect[T]) SetFromKey(key string) error {
	if ss == nil || ss.source == nil {
		return fmt.Errorf("%s", TranslationKeySortedSelectSourceNil)
	}
	var found bool
	var typedKey T
	for k := range ss.source {
		if fmt.Sprint(k) == key {
			typedKey = k
			found = true
			break
		}
	}
	if !found {
		var zero T
		ss.value = zero
		return fmt.Errorf("%s", TranslationKeySortedSelectKeyNotFound)
	}
	ss.value = typedKey
	return nil
}

// SetSource sets or updates the source map for SortedSelect. A map has no
// order, so the options of an ordered select are sorted by key again (or by
// label after SortByLabel).
func (ss *SortedSelect[T]) SetSource(source map[T]string) {
	ss.source = source
	ss.order.keys = nil
}

// Source returns a copy of the internal source map for read-only access
func (ss SortedSelect[T]) Source() map[T]string {
	return ss.source
}

// MarshalJSON outputs {"value":..., "source":...} with string keys for source
func (ss SortedSelect[T]) MarshalJSON() ([]byte, error) {
	source := make(map[string]string, len(ss.source))
	for k, v := range ss.source {
		source[fmt.Sprint(k)] = v
	}
	return json.Marshal(struct {
		Value  T                 `json:"value"`
		Source map[string]string `json:"source"`
	}{
		Value:  ss.value,
		Source: source,
	})
}

// UnmarshalJSON restores value and Source, validates value
func (ss *SortedSelect[T]) UnmarshalJSON(data []byte) error {
	var temp struct {
		Value  T            `json:"value"`
		Source map[T]string `json:"source"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	ss.source = temp.Source
	ss.value = temp.Value
	if ss.source != nil {
		found := false
		for k := range ss.source {
			if fmt.Sprint(k) == fmt.Sprint(ss.value) {
				found = true
				break
			}
		}
		if !found {
			var zero T
			ss.value = zero
			return SortedSelectError{Key: TranslationKeySortedSelectNotFound, Args: []any{ss.value}}
		}
	}
	return nil
}

// SortedSelectError is a structured error for translation
// Holds the translation key and arguments for formatting
// Implements error interface
// Usage: return SortedSelectError{Key: key, Args: args}
type SortedSelectError struct {
	Key  string
	Args []any
}

func (e SortedSelectError) Error() string {
	if len(e.Args) > 0 {
		return fmt.Sprintf(e.Key, e.Args...)
	}
	return e.Key
}

// Ensure SortedSelect implements SortedMapper
var _ SortedMapper = SortedSelect[string]{}

package form

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

// SortedSelect is a generic helper that holds a typed key value and a map of key->label
// It implements SortedMapper (value receiver) and a SetFromKey method (pointer receiver)
// so MapForm can set the selected value from the posted key string.
// T must be comparable to be used as map key.
type SortedSelect[T comparable] struct {
	value  T
	source map[T]string
}

// entry implements SortedMap for SortedSelect
type ssEntry struct {
	k string
	v string
}

func NewSortedSelect[T comparable](source map[T]string) SortedSelect[T] {
	return SortedSelect[T]{source: source}
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

// SortedMapper returns entries sorted by key string representation
// Uses fmt.Sprint to produce string representations of keys (no converters required).
func (ss SortedSelect[T]) SortedMapper() []SortedMap {
	if ss.source == nil {
		return nil
	}
	// collect keys as strings using fmt.Sprint
	keys := make([]string, 0, len(ss.source))
	keyMap := make(map[string]T)
	for k := range ss.source {
		ks := fmt.Sprint(k)
		// only record the formatted key once to avoid duplicate entries when Format collides
		if _, exists := keyMap[ks]; !exists {
			keys = append(keys, ks)
			keyMap[ks] = k
		}
	}
	sort.Strings(keys)
	out := make([]SortedMap, 0, len(keys))
	for _, ks := range keys {
		k := keyMap[ks]
		out = append(out, ssEntry{ks, ss.source[k]})
	}
	return out
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

// SetSource sets or updates the source map for SortedSelect
func (ss *SortedSelect[T]) SetSource(source map[T]string) {
	ss.source = source
}

// Source returns a copy of the internal source map for read-only access
func (ss SortedSelect[T]) Source() map[T]string {
	return ss.source
}

// MarshalJSON outputs {"value":..., "source":...}
func (ss SortedSelect[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Value  T            `json:"value"`
		Source map[T]string `json:"source"`
	}{
		Value:  ss.value,
		Source: ss.source,
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

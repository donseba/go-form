package form

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

// ValueSorted is a generic helper that holds a typed key value and a map of key->label
// It implements SortedMapper (value receiver) and a SetFromKey method (pointer receiver)
// so MapForm can set the selected value from the posted key string.
// T must be comparable to be used as map key.
type ValueSorted[T comparable] struct {
	value  T
	source map[T]string
}

// entry implements SortedMap for ValueSorted
type vsEntry struct {
	k string
	v string
}

func NewValueSorted[T comparable](source map[T]string) ValueSorted[T] {
	return ValueSorted[T]{source: source}
}

func (e vsEntry) Key() string   { return e.k }
func (e vsEntry) Value() string { return e.v }

// Get returns the current value of the ValueSorted field (typed as T)
func (vs ValueSorted[T]) Get() T {
	return vs.value
}

// Set sets the value of the ValueSorted field, ensuring it exists in Source if Source is not nil
func (vs *ValueSorted[T]) Set(val T) error {
	if vs.source != nil {
		found := false
		for k := range vs.source {
			if fmt.Sprint(k) == fmt.Sprint(val) {
				found = true
				break
			}
		}
		if !found {
			var zero T
			vs.value = zero
			return ValueSortedError{Key: TranslationKeyValueSortedNotFound, Args: []any{val}}
		}
	}

	vs.value = val
	return nil
}

// SortedMapper returns entries sorted by key string representation
// Uses fmt.Sprint to produce string representations of keys (no converters required).
func (vs ValueSorted[T]) SortedMapper() []SortedMap {
	if vs.source == nil {
		return nil
	}
	// collect keys as strings using fmt.Sprint
	keys := make([]string, 0, len(vs.source))
	keyMap := make(map[string]T)
	for k := range vs.source {
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
		out = append(out, vsEntry{ks, vs.source[k]})
	}
	return out
}

func (vs *ValueSorted[T]) Scan(val any) error {
	if vs == nil {
		return nil
	}
	v := reflect.ValueOf(val)
	if !v.IsValid() {
		var zero T
		vs.value = zero
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
		return fmt.Errorf("%s", TranslationKeyValueSortedTypeError)
	}
	found := false
	for k := range vs.source {
		if fmt.Sprint(k) == fmt.Sprint(assigned) {
			found = true
			break
		}
	}
	if !found {
		vs.value = zero
		return ValueSortedError{Key: TranslationKeyValueSortedNotFound, Args: []any{assigned}}
	}
	vs.value = assigned
	return nil
}

func (vs *ValueSorted[T]) Value() (any, error) {
	return vs.value, nil
}

// String returns the current value as a string using fmt.Sprint (no converters required)
func (vs ValueSorted[T]) String() string {
	return fmt.Sprint(vs.value)
}

// SetFromKey sets the value from a string key (as submitted by forms)
func (vs *ValueSorted[T]) SetFromKey(key string) error {
	if vs == nil || vs.source == nil {
		return fmt.Errorf("%s", TranslationKeyValueSortedSourceNil)
	}
	var found bool
	var typedKey T
	for k := range vs.source {
		if fmt.Sprint(k) == key {
			typedKey = k
			found = true
			break
		}
	}
	if !found {
		var zero T
		vs.value = zero
		return fmt.Errorf("%s", TranslationKeyValueSortedKeyNotFound)
	}
	vs.value = typedKey
	return nil
}

// SetSource sets or updates the source map for ValueSorted
func (vs *ValueSorted[T]) SetSource(source map[T]string) {
	vs.source = source
}

// Source returns a copy of the internal source map for read-only access
func (vs ValueSorted[T]) Source() map[T]string {
	return vs.source
}

// MarshalJSON outputs {"value":..., "source":...}
func (vs ValueSorted[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Value  T            `json:"value"`
		Source map[T]string `json:"source"`
	}{
		Value:  vs.value,
		Source: vs.source,
	})
}

// UnmarshalJSON restores value and Source, validates value
func (vs *ValueSorted[T]) UnmarshalJSON(data []byte) error {
	var temp struct {
		Value  T            `json:"value"`
		Source map[T]string `json:"source"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	vs.source = temp.Source
	vs.value = temp.Value
	if vs.source != nil {
		found := false
		for k := range vs.source {
			if fmt.Sprint(k) == fmt.Sprint(vs.value) {
				found = true
				break
			}
		}
		if !found {
			var zero T
			vs.value = zero
			return ValueSortedError{Key: TranslationKeyValueSortedNotFound, Args: []any{vs.value}}
		}
	}
	return nil
}

// ValueSortedError is a structured error for translation
// Holds the translation key and arguments for formatting
// Implements error interface
// Usage: return ValueSortedError{Key: key, Args: args}
type ValueSortedError struct {
	Key  string
	Args []any
}

func (e ValueSortedError) Error() string {
	if len(e.Args) > 0 {
		return fmt.Sprintf(e.Key, e.Args...)
	}
	return e.Key
}

// Ensure ValueSorted implements SortedMapper
var _ SortedMapper = ValueSorted[string]{}

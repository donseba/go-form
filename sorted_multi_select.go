package form

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
)

// SortedMultiSelect is a generic helper for multi-select dropdowns with custom key types.
// Holds a slice of selected values and a map of key->label.
type SortedMultiSelect[T comparable] struct {
	values []T
	source map[T]string
}

type smsEntry struct {
	k string
	v string
}

func (e smsEntry) Key() string   { return e.k }
func (e smsEntry) Value() string { return e.v }

func NewSortedMultiSelect[T comparable](source map[T]string) SortedMultiSelect[T] {
	return SortedMultiSelect[T]{source: source}
}

// Get returns the current selected values.
func (sms SortedMultiSelect[T]) Get() []T {
	if sms.values == nil {
		return nil
	}
	out := make([]T, len(sms.values))
	copy(out, sms.values)
	return out
}

// Set sets the selected values, ensuring all exist in Source.
func (sms *SortedMultiSelect[T]) Set(vals []T) error {
	if sms.source != nil {
		for _, val := range vals {
			found := false
			for k := range sms.source {
				if fmt.Sprint(k) == fmt.Sprint(val) {
					found = true
					break
				}
			}
			if !found {
				sms.values = nil
				return SortedSelectError{Key: TranslationKeySortedSelectNotFound, Args: []any{val}}
			}
		}
	}
	sms.values = vals
	return nil
}

// SetFromKeys sets the selected values from a slice of string keys (as submitted by forms).
func (sms *SortedMultiSelect[T]) SetFromKeys(keys []string) error {
	if sms == nil || sms.source == nil {
		return fmt.Errorf("%s", TranslationKeySortedSelectSourceNil)
	}
	var vals []T

	for _, key := range keys {
		found := false
		var typedKey T
		for k := range sms.source {
			if fmt.Sprint(k) == key {
				typedKey = k
				found = true
				break
			}
		}
		if !found {
			sms.values = nil
			return fmt.Errorf("%s: value '%s' not found in source", TranslationKeySortedSelectKeyNotFound, key)
		}
		vals = append(vals, typedKey)
	}

	sms.values = vals
	return nil
}

// SortedMapper returns entries sorted by key string representation.
func (sms SortedMultiSelect[T]) SortedMapper() []SortedMap {
	if sms.source == nil {
		return nil
	}
	keys := make([]string, 0, len(sms.source))
	keyMap := make(map[string]T)
	for k := range sms.source {
		ks := fmt.Sprint(k)
		if _, exists := keyMap[ks]; !exists {
			keys = append(keys, ks)
			keyMap[ks] = k
		}
	}
	sort.Strings(keys)
	out := make([]SortedMap, 0, len(keys))
	for _, ks := range keys {
		k := keyMap[ks]
		out = append(out, smsEntry{ks, sms.source[k]})
	}
	return out
}

// Scan implements database/sql.Scanner for []T.
func (sms *SortedMultiSelect[T]) Scan(val any) error {
	if sms == nil {
		return nil
	}
	v := reflect.ValueOf(val)
	if !v.IsValid() {
		sms.values = nil
		return nil
	}
	var vals []T
	switch v.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i).Interface()
			var zero T
			typ := reflect.TypeOf(zero)
			var assigned T
			iv := reflect.ValueOf(item)
			if iv.Type().AssignableTo(typ) {
				assigned = iv.Interface().(T)
			} else if iv.Type().ConvertibleTo(typ) {
				assigned = iv.Convert(typ).Interface().(T)
			} else {
				return fmt.Errorf("%s", TranslationKeySortedSelectTypeError)
			}
			vals = append(vals, assigned)
		}
	default:
		return fmt.Errorf("%s", TranslationKeySortedSelectTypeError)
	}
	return sms.Set(vals)
}

// Value returns the selected values for DB storage.
func (sms *SortedMultiSelect[T]) Value() (any, error) {
	return sms.values, nil
}

// String returns the selected values as a comma-separated string.
func (sms SortedMultiSelect[T]) String() string {
	out := make([]string, len(sms.values))
	for i, v := range sms.values {
		out[i] = fmt.Sprint(v)
	}
	return fmt.Sprintf("[%s]", joinStrings(out, ", "))
}

func joinStrings(ss []string, sep string) string {
	if len(ss) == 0 {
		return ""
	}
	out := ss[0]
	for _, s := range ss[1:] {
		out += sep + s
	}
	return out
}

// SetSource sets or updates the source map for SortedMultiSelect.
func (sms *SortedMultiSelect[T]) SetSource(source map[T]string) {
	sms.source = source
}

// Source returns a copy of the internal source map for read-only access.
func (sms SortedMultiSelect[T]) Source() map[T]string {
	return sms.source
}

// MarshalJSON outputs {"values":..., "source":...} with string keys for source
func (sms SortedMultiSelect[T]) MarshalJSON() ([]byte, error) {
	source := make(map[string]string, len(sms.source))
	for k, v := range sms.source {
		source[fmt.Sprint(k)] = v
	}
	return json.Marshal(struct {
		Values []T               `json:"values"`
		Source map[string]string `json:"source"`
	}{
		Values: sms.values,
		Source: source,
	})
}

// UnmarshalJSON restores values and Source, validates values
func (sms *SortedMultiSelect[T]) UnmarshalJSON(data []byte) error {
	var temp struct {
		Values []T          `json:"values"`
		Source map[T]string `json:"source"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	sms.source = temp.Source
	return sms.Set(temp.Values)
}

// MultiSelectGetter provides a unified way to get selected keys as strings for all multi-select types.
type MultiSelectGetter interface {
	GetKeysAsStrings() []string
}

// GetKeysAsStrings returns the selected keys as strings.
func (sms SortedMultiSelect[T]) GetKeysAsStrings() []string {
	keys := make([]string, len(sms.values))
	for i, v := range sms.values {
		keys[i] = fmt.Sprint(v)
	}
	return keys
}

// Ensure SortedMultiSelect implements SortedMapper
var _ SortedMapper = SortedMultiSelect[string]{}

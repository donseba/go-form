package form

import (
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
	Source map[T]string
}

// entry implements SortedMap for ValueSorted
type vsEntry struct {
	k string
	v string
}

func (e vsEntry) Key() string   { return e.k }
func (e vsEntry) Value() string { return e.v }

// SortedMapper returns entries sorted by key string representation
// Uses fmt.Sprint to produce string representations of keys (no converters required).
func (vs ValueSorted[T]) SortedMapper() []SortedMap {
	if vs.Source == nil {
		return nil
	}
	// collect keys as strings using fmt.Sprint
	keys := make([]string, 0, len(vs.Source))
	keyMap := make(map[string]T)
	for k := range vs.Source {
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
		out = append(out, vsEntry{ks, vs.Source[k]})
	}
	return out
}

func (vs ValueSorted[T]) Val() T {
	return vs.value
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
	// use a stable type for T (derived from a zero value) rather than reflecting the current value
	var zero T
	typ := reflect.TypeOf(zero)
	if v.Type().AssignableTo(typ) {
		vs.value = v.Interface().(T)
		return nil
	}
	if v.Type().ConvertibleTo(typ) {
		vs.value = v.Convert(typ).Interface().(T)
		return nil
	}
	return fmt.Errorf("cannot assign or convert value of type %T to %s", val, typ.String())
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
	if vs == nil || vs.Source == nil {
		return fmt.Errorf("ValueSorted: Source map is nil")
	}
	var found bool
	var typedKey T
	for k := range vs.Source {
		if fmt.Sprint(k) == key {
			typedKey = k
			found = true
			break
		}
	}
	if !found {
		var zero T
		vs.value = zero
		return fmt.Errorf("ValueSorted: key '%s' not found in Source", key)
	}
	vs.value = typedKey
	return nil
}

// Ensure ValueSorted implements SortedMapper
var _ SortedMapper = ValueSorted[string]{}

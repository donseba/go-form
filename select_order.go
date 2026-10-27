package form

import (
	"fmt"
	"sort"
	"strings"
)

// Option is one choice of an ordered select: the value it stands for and the
// label people see. NewOrderedSelect and NewOrderedMultiSelect keep options in
// the order they are given.
type Option[T comparable] struct {
	Value T
	Label string
}

// selectOrder is how a SortedSelect or SortedMultiSelect lists its options.
type selectOrder[T comparable] struct {
	// keys is the order of an ordered select; nil sorts the options by key.
	keys []T
	// byLabel sorts the options by label instead.
	byLabel bool
}

// optionsFrom splits options into the source map and their order. A value
// given twice keeps its first place and its last label.
func optionsFrom[T comparable](options []Option[T]) (map[T]string, []T) {
	source := make(map[T]string, len(options))
	keys := make([]T, 0, len(options))

	for _, option := range options {
		if _, exists := source[option.Value]; !exists {
			keys = append(keys, option.Value)
		}

		source[option.Value] = option.Label
	}

	return source, keys
}

// entries lists the options of source: in the order of an ordered select,
// otherwise by key, and by label when byLabel is set. Keys are compared as
// fmt.Sprint prints them, and a key printed twice is listed once.
func (o selectOrder[T]) entries(source map[T]string) []SortedMap {
	if source == nil {
		return nil
	}

	var keys []T
	seen := make(map[string]bool, len(source))

	if o.keys != nil {
		for _, key := range o.keys {
			if _, ok := source[key]; !ok || seen[fmt.Sprint(key)] {
				continue
			}

			seen[fmt.Sprint(key)] = true
			keys = append(keys, key)
		}
	}

	// Keys outside the order (or all keys of an unordered select) follow,
	// sorted by key.
	var rest []T

	for key := range source {
		if !seen[fmt.Sprint(key)] {
			seen[fmt.Sprint(key)] = true
			rest = append(rest, key)
		}
	}

	sort.Slice(rest, func(i, j int) bool {
		return fmt.Sprint(rest[i]) < fmt.Sprint(rest[j])
	})

	keys = append(keys, rest...)

	if o.byLabel {
		sort.SliceStable(keys, func(i, j int) bool {
			return labelLess(source[keys[i]], source[keys[j]])
		})
	}

	out := make([]SortedMap, 0, len(keys))
	for _, key := range keys {
		out = append(out, ssEntry{fmt.Sprint(key), source[key]})
	}

	return out
}

// labelSorter is a select that lists its options by label (SortByLabel).
type labelSorter interface {
	SortedByLabel() bool
}

// labelLess orders labels alphabetically, ignoring case; labels that differ
// only in case keep their order.
func labelLess(a, b string) bool {
	return strings.ToLower(a) < strings.ToLower(b)
}

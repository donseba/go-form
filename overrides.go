package form

import (
	"maps"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

type (
	// Tags are values of a field's struct tags by tag name, such as
	// {"max": "8", "placeholder": "Up to 8 people"}.
	Tags map[string]string

	// Overrides replace the struct tags of fields for one render or
	// validation, by the field's name as it renders (its name tag or Go
	// name, such as "adults", or "address.city" in a nested struct). They
	// express limits that change at runtime, such as the places left at
	// an event, without copying the struct:
	//
	//	model := form.WithInfo(input, info)
	//	model.Overrides = form.Overrides{
	//		"adults": {"max": strconv.Itoa(left)},
	//	}
	//
	// Any tag a field reads can be replaced, such as min, max, step,
	// maxLength, minLength, pattern, placeholder, label, description,
	// required, accept, maxSize or a message such as maxMsg; an empty
	// value removes the tag. The name tag cannot be replaced: it is how
	// a field is found. Overrides apply to rendering and, through
	// Overrider, to validation and Bind.
	Overrides map[string]Tags

	// Overrider is a model whose limits change at runtime: rendering,
	// validation (ValidateForm, ValidateFormLocalized) and Bind apply the
	// overrides it returns, so the browser and the server check the same
	// limits:
	//
	//	type BookingInput struct {
	//		Adults int `form:"input,number" name:"adults" min:"1" max:"50"`
	//
	//		left int // the places left, set before rendering or binding
	//	}
	//
	//	func (in BookingInput) FormOverrides() form.Overrides {
	//		return form.Overrides{"adults": {"max": strconv.Itoa(in.left)}}
	//	}
	//
	// A nested struct may implement it too; the overrides of the outer
	// model win. RenderModel.Overrides win over both.
	Overrider interface {
		FormOverrides() Overrides
	}
)

// overriderType is the type of Overrider.
var overriderType = reflect.TypeOf((*Overrider)(nil)).Elem()

// overridesOf returns the overrides of v when it, or its address, is an
// Overrider.
func overridesOf(v reflect.Value) Overrides {
	if !v.IsValid() {
		return nil
	}

	if v.Kind() == reflect.Pointer && v.IsNil() {
		return nil
	}

	if v.CanInterface() && v.Type().Implements(overriderType) {
		return v.Interface().(Overrider).FormOverrides()
	}

	if v.CanAddr() && v.Addr().CanInterface() && v.Addr().Type().Implements(overriderType) {
		return v.Addr().Interface().(Overrider).FormOverrides()
	}

	return nil
}

// merge returns the overrides of o with those of other on top: a tag both
// set has the value of other.
func (o Overrides) merge(other Overrides) Overrides {
	if len(other) == 0 {
		return o
	}

	if len(o) == 0 {
		return other
	}

	merged := make(Overrides, len(o)+len(other))
	maps.Copy(merged, o)

	for name, tags := range other {
		combined := make(Tags, len(merged[name])+len(tags))
		maps.Copy(combined, merged[name])
		maps.Copy(combined, tags)

		merged[name] = combined
	}

	return merged
}

// prefixed returns the overrides with prefix and a dot before every name,
// for the fields of a struct nested as prefix.
func (o Overrides) prefixed(prefix string) Overrides {
	if len(o) == 0 {
		return nil
	}

	out := make(Overrides, len(o))
	for name, tags := range o {
		out[prefix+"."+name] = tags
	}

	return out
}

// below returns the entries of byName, such as Overrides or FieldOptions,
// of the fields of the struct nested as prefix, named within that struct:
// "address.city" is "city" below "address". It is nil when there are none.
func below[M ~map[string]V, V any](byName M, prefix string) M {
	var out M

	for name, value := range byName {
		rest, ok := strings.CutPrefix(name, prefix+".")
		if !ok {
			continue
		}

		if out == nil {
			out = M{}
		}

		out[rest] = value
	}

	return out
}

// apply returns field with the overrides of the field named name in its
// tags.
func (o Overrides) apply(name string, field reflect.StructField) reflect.StructField {
	tags, ok := o[name]
	if !ok || len(tags) == 0 {
		return field
	}

	field.Tag = overrideTag(field.Tag, tags)

	return field
}

// tagEntry is one key:"value" pair of a struct tag.
type tagEntry struct {
	key   string
	value string
}

// overrideTag returns tag with the values of tags: an existing pair gets
// its new value in its place, other pairs follow in name order, and an
// empty value removes the pair. The name tag is kept as it is.
func overrideTag(tag reflect.StructTag, tags Tags) reflect.StructTag {
	entries := parseTag(tag)
	done := make(map[string]bool, len(tags))

	var pairs []string

	for _, entry := range entries {
		value := entry.value

		if replacement, ok := tags[entry.key]; ok && entry.key != tagName {
			value = replacement
			done[entry.key] = true
		}

		if value == "" {
			continue
		}

		pairs = append(pairs, tagPair(entry.key, value))
	}

	for _, key := range slices.Sorted(maps.Keys(tags)) {
		if done[key] || key == tagName || tags[key] == "" || !validTagKey(key) {
			continue
		}

		pairs = append(pairs, tagPair(key, tags[key]))
	}

	return reflect.StructTag(strings.Join(pairs, " "))
}

// validTagKey reports whether key can be the key of a struct tag pair:
// no space, quote, colon or control character.
func validTagKey(key string) bool {
	return key != "" && !strings.ContainsFunc(key, func(r rune) bool {
		return r <= ' ' || r == ':' || r == '"' || r == 0x7f
	})
}

// parseTag splits a struct tag into its key:"value" pairs, in order, as
// reflect.StructTag.Lookup reads them. Parsing stops at the first pair that
// is not well formed.
func parseTag(tag reflect.StructTag) []tagEntry {
	var entries []tagEntry

	rest := strings.TrimLeft(string(tag), " ")
	for rest != "" {
		entry, after, ok := nextTagEntry(rest)
		if !ok {
			break
		}

		entries = append(entries, entry)
		rest = strings.TrimLeft(after, " ")
	}

	return entries
}

// nextTagEntry reads the key:"value" pair at the start of rest, and returns
// it with the text after it. ok is false when rest does not start with a
// well-formed pair.
func nextTagEntry(rest string) (entry tagEntry, after string, ok bool) {
	// The key runs up to the colon.
	i := 0
	for i < len(rest) && rest[i] > ' ' && rest[i] != ':' && rest[i] != '"' && rest[i] != 0x7f {
		i++
	}

	if i == 0 || i+1 >= len(rest) || rest[i] != ':' || rest[i+1] != '"' {
		return tagEntry{}, "", false
	}

	key := rest[:i]
	rest = rest[i+1:]

	// The value is a quoted string.
	i = 1
	for i < len(rest) && rest[i] != '"' {
		if rest[i] == '\\' {
			i++
		}

		i++
	}

	if i >= len(rest) {
		return tagEntry{}, "", false
	}

	value, err := strconv.Unquote(rest[:i+1])
	if err != nil {
		return tagEntry{}, "", false
	}

	return tagEntry{key: key, value: value}, rest[i+1:], true
}

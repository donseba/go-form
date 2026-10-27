package form

import (
	"fmt"
	"html"
	"html/template"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/donseba/go-form/v2/templates"
)

// A template can give the opening form tag attributes of its own, after
// the model: form_open .Form "class" "booking" (someAttrs …). They are
// name/value pairs and template.HTMLAttr values, such as the attributes
// another library renders for a form that submits with HTMX:
//
//	{{form_open .Form "class" "signup" (submitAttrs runtime .Action "target" "#signup")}}
//
// Precedence:
//
//   - action, method and enctype come from the form metadata (Info.Target,
//     Info.Method, Info.Attributes["enctype"] or a file input); giving them
//     here is an error rather than a second, ignored attribute.
//   - class adds to the theme's class and that of Info.Attributes; a class
//     already there is not repeated. style adds to the theme's style and
//     that of Info.Attributes.
//   - Any other attribute replaces the one of the same name in
//     Info.Attributes; given twice, the last one counts.
//
// The values of pairs are escaped; a pair cannot set an event handler
// (on…), which needs a trusted template.HTMLAttr. A nil value renders the
// name alone, as for novalidate.

// formAttribute is one attribute a template gives the opening form tag. A
// bare attribute, such as novalidate, has a name only.
type formAttribute struct {
	name  string
	value string
	bare  bool
}

// metadataAttributes are the attributes the form metadata sets, with where
// to set them.
var metadataAttributes = map[string]string{
	"action":  "Info.Target",
	"method":  "Info.Method",
	"enctype": `Info.Attributes["enctype"]`,
}

// attributeNamePattern is the attribute name a name/value pair may give.
var attributeNamePattern = regexp.MustCompile(`^[A-Za-z_:@][A-Za-z0-9_:.@-]*$`)

// openAttributes reads the attributes a template gives form_open: name and
// value pairs, and template.HTMLAttr values. nil arguments are skipped, so
// an optional attribute can be left out.
func openAttributes(args []any) ([]formAttribute, error) {
	var attrs []formAttribute

	for i := 0; i < len(args); i++ {
		switch arg := args[i].(type) {
		case nil:
			continue
		case template.HTMLAttr:
			parsed, err := parseAttributes(string(arg))
			if err != nil {
				return nil, err
			}

			attrs = append(attrs, parsed...)
		case string:
			if i+1 >= len(args) {
				return nil, fmt.Errorf("form_open: attribute %q has no value", arg)
			}

			attr, err := pairAttribute(arg, args[i+1])
			if err != nil {
				return nil, err
			}

			attrs = append(attrs, attr)
			i++
		default:
			return nil, fmt.Errorf("form_open: want name/value pairs or template.HTMLAttr, got %T", arg)
		}
	}

	for _, attr := range attrs {
		if source, ok := metadataAttributes[strings.ToLower(attr.name)]; ok {
			return nil, fmt.Errorf("form_open: the %s attribute comes from the form metadata (%s)", attr.name, source)
		}
	}

	return attrs, nil
}

// pairAttribute is the attribute of a name/value pair.
func pairAttribute(name string, value any) (formAttribute, error) {
	if !attributeNamePattern.MatchString(name) {
		return formAttribute{}, fmt.Errorf("form_open: %q is not an attribute name", name)
	}

	if strings.HasPrefix(strings.ToLower(name), "on") {
		return formAttribute{}, fmt.Errorf("form_open: set the event handler %s with a template.HTMLAttr", name)
	}

	switch v := value.(type) {
	case nil:
		return formAttribute{name: name, bare: true}, nil
	case string:
		return formAttribute{name: name, value: v}, nil
	default:
		return formAttribute{name: name, value: fmt.Sprint(v)}, nil
	}
}

// parseAttributes reads the attributes of trusted markup such as
// ` hx-post="/x" hx-target='#y' novalidate`, with their values unescaped.
func parseAttributes(markup string) ([]formAttribute, error) {
	var attrs []formAttribute

	i := skipSpace(markup, 0)
	for i < len(markup) {
		start := i
		for i < len(markup) && !isSpace(markup[i]) && !strings.ContainsRune(`="'<>/`, rune(markup[i])) {
			i++
		}

		name := markup[start:i]
		if name == "" {
			return nil, fmt.Errorf("form_open: cannot read the attributes %q", markup)
		}

		i = skipSpace(markup, i)
		if i >= len(markup) || markup[i] != '=' {
			attrs = append(attrs, formAttribute{name: name, bare: true})

			continue
		}

		value, next, err := attributeValue(markup, skipSpace(markup, i+1))
		if err != nil {
			return nil, err
		}

		attrs = append(attrs, formAttribute{name: name, value: html.UnescapeString(value)})
		i = skipSpace(markup, next)
	}

	return attrs, nil
}

// attributeValue reads the value of an attribute at i of markup, quoted or
// not, and returns it with the position after it.
func attributeValue(markup string, i int) (string, int, error) {
	if i < len(markup) && (markup[i] == '"' || markup[i] == '\'') {
		quote := markup[i]

		end := strings.IndexByte(markup[i+1:], quote)
		if end < 0 {
			return "", 0, fmt.Errorf("form_open: unclosed quote in the attributes %q", markup)
		}

		return markup[i+1 : i+1+end], i + end + 2, nil
	}

	start := i
	for i < len(markup) && !isSpace(markup[i]) {
		i++
	}

	return markup[start:i], i, nil
}

// skipSpace returns the position of the first character from i of s that
// is not white space.
func skipSpace(s string, i int) int {
	for i < len(s) && isSpace(s[i]) {
		i++
	}

	return i
}

// isSpace reports whether c is HTML white space.
func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f'
}

// openTag is the opening form tag's attributes besides action and method,
// as the form-open template renders them.
type openTag struct {
	// attributes are those of Info.Attributes that render as they are.
	attributes map[string]string
	// class are the classes besides the theme's.
	class string
	// style is the inline style besides the theme's.
	style template.CSS
	// extra are the other attributes the template gave, rendered.
	extra template.HTMLAttr
}

// buildOpenTag merges the attributes of the form metadata (info) and those
// a template gave (attrs) as the precedence above says. themeClass is the
// theme's class of the form, which the classes do not repeat.
func buildOpenTag(themeClass string, info map[string]string, attrs []formAttribute) openTag {
	given := make(map[string]bool, len(attrs))
	for _, attr := range attrs {
		given[strings.ToLower(attr.name)] = true
	}

	tag := openTag{attributes: make(map[string]string, len(info))}

	var classes, styles []string

	// The class and style of Info.Attributes merge into the theme's.
	for _, name := range slices.Sorted(maps.Keys(info)) {
		switch lower := strings.ToLower(name); {
		case lower == "class":
			classes = append(classes, info[name])
		case lower == "style":
			styles = append(styles, info[name])
		case !given[lower]:
			tag.attributes[name] = info[name]
		}
	}

	var (
		order []string
		last  = map[string]formAttribute{}
	)

	for _, attr := range attrs {
		lower := strings.ToLower(attr.name)

		switch lower {
		case "class":
			classes = append(classes, attr.value)
		case "style":
			styles = append(styles, attr.value)
		default:
			if _, seen := last[lower]; !seen {
				order = append(order, lower)
			}

			last[lower] = attr
		}
	}

	tag.class = classTokens(themeClass, classes)
	tag.style = joinStyles(styles)

	var extra strings.Builder

	for _, lower := range order {
		attr := last[lower]

		extra.WriteString(" ")
		extra.WriteString(attr.name)

		if !attr.bare {
			extra.WriteString(`="`)
			extra.WriteString(template.HTMLEscapeString(attr.value))
			extra.WriteString(`"`)
		}
	}

	//nolint:gosec // Checked names around escaped values.
	tag.extra = template.HTMLAttr(extra.String())

	return tag
}

// classTokens are the classes of lists, once each, without those of
// themeClass, separated by a space.
func classTokens(themeClass string, lists []string) string {
	seen := map[string]bool{}
	for _, class := range strings.Fields(themeClass) {
		seen[class] = true
	}

	var tokens []string

	for _, list := range lists {
		for _, class := range strings.Fields(list) {
			if seen[class] {
				continue
			}

			seen[class] = true
			tokens = append(tokens, class)
		}
	}

	return strings.Join(tokens, " ")
}

// joinStyles joins inline styles with "; ".
func joinStyles(styles []string) template.CSS {
	css := make([]template.CSS, 0, len(styles))
	for _, style := range styles {
		//nolint:gosec // Styles of the application, escaped for the attribute.
		css = append(css, template.CSS(style))
	}

	return templates.StyleList(css...)
}

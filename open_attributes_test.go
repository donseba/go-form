package form

import (
	"bytes"
	"html/template"
	"net/url"
	"strings"
	"testing"
)

type openInput struct {
	Name string `form:"input,text" name:"name" label:"Name"`
}

// openTagOf is the opening form tag in markup, with its white space
// collapsed.
func openTagOf(t *testing.T, markup template.HTML) string {
	t.Helper()

	tag, _, found := strings.Cut(string(markup), ">")
	if !found {
		t.Fatalf("no tag in %s", markup)
	}

	return strings.Join(strings.Fields(tag), " ") + ">"
}

func TestFormOpenAttributes(t *testing.T) {
	f := NewForm()
	f.SetTheme("bootstrap")

	info := Info{
		Target:   "/book",
		Method:   "post",
		Honeypot: "website",
		Hidden:   url.Values{"back": {"/workshop#booking"}},
		Attributes: map[string]string{
			"class":     "card mx-auto",
			"id":        "from-info",
			"data-kind": "booking",
		},
	}
	model := WithInfo(openInput{}, info)

	htmx := template.HTMLAttr(` hx-post="/book" hx-target='#booking' hx-vals="{&#34;a&#34;:1}" novalidate`)

	open, err := f.formOpen(model, "class", "booking shadow-sm", "id", "booking-form", htmx, "aria-label", nil, nil)
	if err != nil {
		t.Fatalf("form_open: %v", err)
	}

	tag := openTagOf(t, open)
	want := `<form action="/book" method="post" class="mx-auto border rounded shadow-sm p-3 card booking" data-kind="booking" id="booking-form" hx-post="/book" hx-target="#booking" hx-vals="{&#34;a&#34;:1}" novalidate aria-label>`

	if tag != want {
		t.Errorf("tag\n got %s\nwant %s", tag, want)
	}

	for _, hidden := range []string{`name="website"`, `name="back" value="/workshop#booking"`} {
		if !strings.Contains(string(open), hidden) {
			t.Errorf("form_open misses %s: %s", hidden, open)
		}
	}

	if strings.Count(tag, "class=") != 1 {
		t.Errorf("more than one class attribute: %s", tag)
	}
}

func TestFormOpenWithoutAttributesUnchanged(t *testing.T) {
	f := NewForm()
	f.SetTheme("bootstrap")

	model := WithInfo(openInput{}, Info{Target: "/send", Method: "post", Attributes: map[string]string{"data-x": "1"}})

	open, err := f.formOpen(model)
	if err != nil {
		t.Fatalf("form_open: %v", err)
	}

	want := `<form action="/send" method="post" class="mx-auto border rounded shadow-sm p-3" data-x="1">`
	if tag := openTagOf(t, open); tag != want {
		t.Errorf("tag\n got %s\nwant %s", tag, want)
	}
}

func TestFormOpenStyleMerges(t *testing.T) {
	f := NewForm()
	f.SetTheme("plain")

	model := WithInfo(openInput{}, Info{Target: "/send", Method: "post", Attributes: map[string]string{"style": "color: red;"}})

	open, err := f.Open(nil, model, "style", "margin: 0")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	tag := openTagOf(t, open)
	if strings.Count(tag, "style=") != 1 {
		t.Fatalf("not one style attribute: %s", tag)
	}

	if !strings.Contains(tag, `rgba(0, 0, 0, 0.075); color: red; margin: 0"`) {
		t.Errorf("styles not merged after the theme's: %s", tag)
	}
}

func TestFormOpenRefusesMetadataAttributes(t *testing.T) {
	f := NewForm()
	f.SetTheme("plain")

	model := WithInfo(openInput{}, Info{Target: "/send", Method: "post"})

	cases := map[string][]any{
		"method pair":     {"method", "get"},
		"action pair":     {"ACTION", "/elsewhere"},
		"enctype pair":    {"enctype", "multipart/form-data"},
		"method in attrs": {template.HTMLAttr(`hx-post="/x" method="get"`)},
		"event handler":   {"onsubmit", "steal()"},
		"bad name":        {`x" onclick="y`, "1"},
		"no value":        {"class"},
		"other type":      {42},
		"unclosed quote":  {template.HTMLAttr(`hx-post="/x`)},
	}

	for name, attrs := range cases {
		if _, err := f.formOpen(model, attrs...); err == nil {
			t.Errorf("%s: form_open did not fail", name)
		}
	}

	// A trusted template.HTMLAttr may set an event handler.
	if _, err := f.formOpen(model, template.HTMLAttr(`onsubmit="return check()"`)); err != nil {
		t.Errorf("event handler in a template.HTMLAttr: %v", err)
	}
}

func TestFormOpenAttributesInTemplate(t *testing.T) {
	f := NewForm()
	f.SetTheme("plain")

	funcs := f.FuncMap()
	funcs["htmx"] = func(target string) template.HTMLAttr {
		return template.HTMLAttr(` hx-post="/vote" hx-target="` + template.HTMLEscapeString(target) + `"`)
	}

	tmpl := template.Must(template.New("page").Funcs(funcs).Parse(
		`{{form_open . "class" "poll" (htmx "#poll-1")}}{{form_close}}`,
	))

	model := WithInfo(openInput{}, Info{Target: "/vote", Method: "post", Honeypot: "website"})

	var out bytes.Buffer
	if err := tmpl.Execute(&out, model); err != nil {
		t.Fatalf("execute: %v", err)
	}

	html := out.String()
	for _, want := range []string{`class="poll"`, `hx-post="/vote" hx-target="#poll-1"`, `name="website"`, `</form>`} {
		if !strings.Contains(html, want) {
			t.Errorf("template misses %s: %s", want, html)
		}
	}
}

func TestFormHidden(t *testing.T) {
	f := NewForm()
	f.SetTheme("plain")

	model := WithInfo(openInput{}, Info{
		Target:    "/send",
		Method:    "post",
		CsrfValue: "token",
		Honeypot:  "website",
		Hidden:    url.Values{"back": {"/page#form"}},
	})

	hidden, err := f.formHidden(model)
	if err != nil {
		t.Fatalf("form_hidden: %v", err)
	}

	for _, want := range []string{`name="_csrf" value="token"`, `name="website"`, `name="back" value="/page#form"`} {
		if !strings.Contains(string(hidden), want) {
			t.Errorf("form_hidden misses %s: %s", want, hidden)
		}
	}

	if strings.Contains(string(hidden), "<form") {
		t.Errorf("form_hidden renders the form tag: %s", hidden)
	}

	open, _ := f.Open(nil, model)
	if !strings.HasSuffix(string(open), string(hidden)) {
		t.Errorf("Open does not end with the hidden inputs of Hidden:\n%s\n%s", open, hidden)
	}

	if _, err := f.Hidden(nil, openInput{}); err == nil {
		t.Error("Hidden of a model without form metadata did not fail")
	}
}

func TestParseAttributes(t *testing.T) {
	attrs, err := parseAttributes(` a="1"  b='two words' c=bare d e = "&lt;x&gt;" `)
	if err != nil {
		t.Fatal(err)
	}

	want := []formAttribute{
		{name: "a", value: "1"},
		{name: "b", value: "two words"},
		{name: "c", value: "bare"},
		{name: "d", bare: true},
		{name: "e", value: "<x>"},
	}

	if len(attrs) != len(want) {
		t.Fatalf("got %+v, want %+v", attrs, want)
	}

	for i := range want {
		if attrs[i] != want[i] {
			t.Errorf("attribute %d: got %+v, want %+v", i, attrs[i], want[i])
		}
	}
}

package form

import (
	"errors"
	"mime/multipart"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/donseba/go-form/v2/types"
)

type carryInput struct {
	Subject  string                `form:"input,hidden"   name:"subject"`
	Name     string                `form:"input,text"     name:"name"     label:"Name"     required:"true"`
	Email    string                `form:"input,email"    name:"email"    label:"Email"`
	Password string                `form:"input,password" name:"password" label:"Password"`
	Places   int                   `form:"input,number"   name:"places"   label:"Places"`
	Plan     string                `form:"dropdown"       name:"plan"     label:"Plan"                     values:"basic:Basic;premium:Premium"`
	Agree    bool                  `form:"checkbox"       name:"agree"    label:"Agree"`
	Photo    *multipart.FileHeader `form:"input,file"     name:"photo"    label:"Photo"`
}

// posted is a failed post of carryInput, with fields that must never be
// carried.
func posted() url.Values {
	return url.Values{
		"subject":  {"post-1"},
		"name":     {""},
		"email":    {"jane@example.com"},
		"password": {"secret"},
		"places":   {"abc"},
		"plan":     {"premium"},
		"agree":    {"on"},
		"photo":    {"photo.png"},
		"website":  {"http://spam.example"},
		"_csrf":    {"token"},
		"back":     {"/en/news#remarks"},
	}
}

func TestCarryRoundTrip(t *testing.T) {
	model := WithInfo(carryInput{Subject: "post-1"}, Info{Honeypot: "website"})
	errs := FieldErrors{
		FieldValidationError{Field: "name", Err: "Fill in your name"},
		FieldValidationError{Field: "places", Err: "Not a number"},
	}

	carry, err := NewCarry(model, posted(), errs)
	if err != nil {
		t.Fatal(err)
	}

	data, err := carry.Encode(0)
	if err != nil {
		t.Fatal(err)
	}

	for _, secret := range []string{"secret", "spam.example", "token", "photo.png", "/en/news"} {
		if strings.Contains(string(data), secret) {
			t.Errorf("the encoding carries %q: %s", secret, data)
		}
	}

	decoded, err := DecodeCarry(data)
	if err != nil {
		t.Fatal(err)
	}

	want := url.Values{
		"subject": {"post-1"},
		"name":    {""},
		"email":   {"jane@example.com"},
		"places":  {"abc"},
		"plan":    {"premium"},
		"agree":   {"on"},
	}
	if !reflect.DeepEqual(decoded.Values, want) {
		t.Errorf("values = %v, want %v", decoded.Values, want)
	}

	if got := decoded.FieldErrors(); len(got) != 2 || got[0].(FieldValidationError).Field != "name" {
		t.Errorf("errors = %v", got)
	}

	if _, err := DecodeCarry([]byte("not json")); err == nil {
		t.Error("DecodeCarry accepted garbage")
	}
}

// TestCarryLeavesOutHoneypotField checks that a field named as the
// honeypot is not carried either.
func TestCarryLeavesOutHoneypotField(t *testing.T) {
	type trapped struct {
		Name    string `form:"input,text" name:"name"`
		Website string `form:"input,text" name:"website"`
	}

	carry, err := NewCarry(WithInfo(trapped{}, Info{Honeypot: "website"}), url.Values{"name": {"Jane"}, "website": {"x"}}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := carry.Values["website"]; ok || carry.Values.Get("name") != "Jane" {
		t.Errorf("values = %v", carry.Values)
	}
}

func TestCarryRenders(t *testing.T) {
	f := NewForm()
	f.SetTheme("plain")

	carry, err := NewCarry(carryInput{}, posted(), FieldErrors{FieldValidationError{Field: "name", Err: "Fill in your name"}})
	if err != nil {
		t.Fatal(err)
	}

	model := WithInfo(
		carryInput{Subject: "post-1", Name: "Stored", Places: 2, Agree: false},
		Info{Target: "/send", Method: "post"},
	)
	model.Carry = &carry

	html, err := f.Render(nil, model, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, part := range []string{
		`name="subject" value="post-1"`,
		`value="jane@example.com"`,
		`value="abc"`,
		`value="premium" selected`,
		"checked",
		"Fill in your name",
	} {
		if !strings.Contains(string(html), part) {
			t.Errorf("render misses %q:\n%s", part, html)
		}
	}

	if strings.Contains(string(html), "Stored") || strings.Contains(string(html), "secret") {
		t.Errorf("render shows the model's name or the password:\n%s", html)
	}

	// Errors the template passes win over the carried ones.
	field, err := f.Field(nil, model, "name", FieldErrors{FieldValidationError{Field: "name", Err: "Own error"}})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(field), "Own error") || strings.Contains(string(field), "Fill in your name") {
		t.Errorf("field = %s", field)
	}

	// The comment form of another post on the page keeps its own values.
	other := WithInfo(carryInput{Subject: "post-2", Name: "Stored"}, Info{Target: "/send", Method: "post"})
	other.Carry = &carry

	html, err = f.Render(nil, other, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(html), `value="Stored"`) || strings.Contains(string(html), "Fill in your name") {
		t.Errorf("the carry applied to another form:\n%s", html)
	}
}

// TestCarryUnticks checks that a box that was not ticked shows unticked,
// although the model ticks it.
func TestCarryUnticks(t *testing.T) {
	f := NewForm()
	f.SetTheme("plain")

	carry, err := NewCarry(carryInput{}, url.Values{"name": {"Jane"}}, nil)
	if err != nil {
		t.Fatal(err)
	}

	model := WithInfo(carryInput{Agree: true}, Info{Target: "/send", Method: "post"})
	model.Carry = &carry

	field, err := f.Field(nil, model, "agree", nil)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(field), "checked") {
		t.Errorf("box is ticked: %s", field)
	}
}

func TestCarrySchema(t *testing.T) {
	f := NewForm()
	f.SetTheme("plain")

	schema := Schema{Fields: []Spec{
		{Name: "email", InputType: types.InputFieldTypeEmail, Required: true},
		{Name: "pin", InputType: types.InputFieldTypePassword},
		{
			Name:    "topics",
			Type:    types.FieldTypeMultiCheckbox,
			Options: []types.FieldValue{{Value: "a", Name: "A"}, {Value: "b", Name: "B"}},
		},
	}}

	carry, err := NewCarry(
		schema,
		url.Values{"email": {"jane@"}, "pin": {"1234"}, "topics": {"b"}},
		FieldErrors{FieldValidationError{Field: "email", Err: "Not an address"}},
	)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := carry.Values["pin"]; ok {
		t.Fatalf("the password input is carried: %v", carry.Values)
	}

	model := WithInfo(schema.Bind(nil), Info{Target: "/contact", Method: "post"})
	model.Carry = &carry

	html, err := f.Render(nil, model, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, part := range []string{`value="jane@"`, "Not an address", `value="b"`} {
		if !strings.Contains(string(html), part) {
			t.Errorf("render misses %q:\n%s", part, html)
		}
	}
}

func TestCarryEncodeLimit(t *testing.T) {
	carry := Carry{
		Values: url.Values{"name": {"Jane"}, "body": {strings.Repeat("x", 500)}},
		Errors: map[string][]string{"name": {"Too short"}},
	}

	data, err := carry.Encode(100)
	if err != nil {
		t.Fatal(err)
	}

	if len(data) > 100 {
		t.Fatalf("%d bytes over the limit", len(data))
	}

	decoded, err := DecodeCarry(data)
	if err != nil {
		t.Fatal(err)
	}

	if decoded.Values.Get("name") != "Jane" || decoded.Values.Has("body") || len(decoded.Errors["name"]) != 1 {
		t.Errorf("decoded = %+v", decoded)
	}

	// The errors stay; when they alone are too large there is no carry.
	carry.Errors["body"] = []string{strings.Repeat("y", 200)}
	if _, err := carry.Encode(100); !errors.Is(err, ErrCarryTooLarge) {
		t.Errorf("err = %v, want ErrCarryTooLarge", err)
	}

	// Encode leaves the carry itself as it was.
	if !carry.Values.Has("body") {
		t.Error("Encode changed the carry")
	}
}

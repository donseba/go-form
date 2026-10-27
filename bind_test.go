package form

import (
	"net/url"
	"reflect"
	"testing"
)

type bindInput struct {
	Name     string `form:"input,text"     name:"name"     required:"true" maxLength:"5"`
	Age      int    `form:"input,number"   name:"age"                                    min:"18"`
	Body     string `form:"textarea"       name:"body"                     maxLength:"4"`
	Password string `form:"input,password" name:"password"                                        minLength:"3"`
	Address  struct {
		Note string `form:"textarea" name:"note"`
	} `                      name:"address"`
}

func TestBindMapsAndValidates(t *testing.T) {
	r := postForm(url.Values{"name": {" toolong "}, "age": {"12"}, "password": {"secret"}})

	var input bindInput

	errs, err := NewForm().Bind(r, &input, MapOptions{TrimSpace: true}, nil)
	if err != nil {
		t.Fatalf("Bind returned %v", err)
	}

	if input.Name != "toolong" || input.Age != 12 {
		t.Errorf("Bind mapped %+v", input)
	}

	if got := strictNames(errs); !reflect.DeepEqual(got, []string{"name", "age"}) {
		t.Errorf("Bind errors = %v, want name (maxLength) and age (min)", got)
	}
}

func TestBindUnreadableValueReplacesValidation(t *testing.T) {
	r := postForm(url.Values{"name": {"Ann"}, "age": {"abc"}, "password": {"secret"}})

	var input bindInput

	errs, err := NewForm().Bind(r, &input, MapOptions{}, &DefaultLocalizer{})
	if err != nil {
		t.Fatalf("Bind returned %v", err)
	}

	if len(errs) != 1 {
		t.Fatalf("Bind errors = %v, want only the unreadable age", errs)
	}

	if validationErr, ok := errs[0].(FieldValidationError); !ok || validationErr.Field != "age" ||
		validationErr.Rule != "invalid" {
		t.Errorf("Bind error = %#v, want age with rule invalid instead of min", errs[0])
	}
}

func TestBindRejectsNonPointer(t *testing.T) {
	errs, err := NewForm().Bind(postForm(url.Values{}), bindInput{}, MapOptions{}, nil)
	if err == nil || errs != nil {
		t.Errorf("Bind of a struct value = %v, %v; want an error", errs, err)
	}
}

func TestMapOptionsNormalizeNewlines(t *testing.T) {
	r := postForm(url.Values{
		"name":         {"Ann"},
		"age":          {"20"},
		"body":         {"a\r\nb\rc"},
		"password":     {"x\r\ny"},
		"address.note": {"one\r\ntwo"},
	})

	var input bindInput

	errs, err := NewForm().Bind(r, &input, MapOptions{NormalizeNewlines: true}, nil)
	if err != nil {
		t.Fatalf("Bind returned %v", err)
	}

	if input.Body != "a\nb\nc" {
		t.Errorf("Body = %q, want line breaks as \\n", input.Body)
	}

	if input.Address.Note != "one\ntwo" {
		t.Errorf("nested Note = %q, want line breaks as \\n", input.Address.Note)
	}

	if input.Password != "x\r\ny" {
		t.Errorf("Password = %q, want it as typed", input.Password)
	}

	// "a\nb\nc" is 5 characters: over maxLength 4 either way, so the error
	// shows the count is of the normalized text.
	if got := strictNames(errs); !reflect.DeepEqual(got, []string{"body"}) {
		t.Errorf("errors = %v, want body", got)
	}
}

func TestMapOptionsNormalizeNewlinesCountsOneCharacter(t *testing.T) {
	r := postForm(url.Values{"name": {"Ann"}, "age": {"20"}, "password": {"secret"}, "body": {"ab\r\nc"}})

	var input bindInput

	errs, err := NewForm().Bind(r, &input, MapOptions{NormalizeNewlines: true}, nil)
	if err != nil {
		t.Fatalf("Bind returned %v", err)
	}

	if len(errs) != 0 {
		t.Errorf("errors = %v; \"ab\\nc\" is 4 characters, within maxLength", errs)
	}
}

func TestBindKeepsPasswordAsTyped(t *testing.T) {
	r := postForm(url.Values{"name": {"Ann"}, "age": {"20"}, "password": {"  a "}})

	var input bindInput

	errs, err := NewForm().Bind(r, &input, MapOptions{TrimSpace: true, NormalizeNewlines: true}, nil)
	if err != nil {
		t.Fatalf("Bind returned %v", err)
	}

	if input.Password != "  a " || len(errs) != 0 {
		t.Errorf("Password = %q, errors %v; want it as typed and long enough", input.Password, errs)
	}
}

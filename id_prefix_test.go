package form

import (
	"mime/multipart"
	"regexp"
	"strings"
	"testing"

	"github.com/donseba/go-form/v2/templates"
)

type idPrefixInput struct {
	Email   string                    `form:"input,email"   name:"email"   label:"Email"  description:"We never share it"`
	Topics  SortedMultiSelect[string] `form:"multicheckbox" name:"topics"  label:"Topics"`
	Address struct {
		City string `form:"input,text" name:"city" label:"City"`
	} `                     name:"address"                                                legend:"Address"`
}

func TestIDPrefixKeepsFormsApart(t *testing.T) {
	f := NewForm()
	f.SetTheme("plain")

	input := idPrefixInput{Topics: NewSortedMultiSelect(map[string]string{"news": "News"})}
	errs := FieldErrors{FieldValidationError{Field: "email", Err: "Invalid"}}

	html, err := f.Render(nil, WithInfo(input, Info{Target: "/subscribe", Method: "post", IDPrefix: "newsletter"}), errs)
	if err != nil {
		t.Fatal(err)
	}

	out := string(html)

	for _, want := range []string{
		`id="newsletter-email"`,
		`for="newsletter-email"`,
		`id="newsletter-email_label"`,
		`aria-labelledby="newsletter-email_label"`,
		`id="newsletter-email_description"`,
		`id="newsletter-email_error_0"`,
		`aria-describedby="newsletter-email_description newsletter-email_error_0"`,
		`id="newsletter-topics_0"`,
		`id="newsletter-address_legend"`,
		`id="newsletter-address.city"`,
		`name="email"`,
		`name="address.city"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("render misses %s:\n%s", want, out)
		}
	}

	if strings.Contains(out, `id="email"`) || strings.Contains(out, `name="newsletter-`) {
		t.Errorf("an id lost its prefix, or a name got one:\n%s", out)
	}

	field, err := f.Field(nil, WithInfo(input, Info{IDPrefix: "comment"}), "email", nil)
	if err != nil || !strings.Contains(string(field), `id="comment-email"`) {
		t.Errorf("Field with a prefix = %s, %v", field, err)
	}
}

func TestNoIDPrefixKeepsIDs(t *testing.T) {
	f := NewForm()
	f.SetTheme("plain")

	html, err := f.Field(nil, WithInfo(idPrefixInput{}, Info{}), "email", nil)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(html), `id="email"`) || !strings.Contains(string(html), `for="email"`) {
		t.Errorf("ids changed without a prefix:\n%s", html)
	}
}

type classInput struct {
	Upload *multipart.FileHeader `form:"input,file" name:"upload" label:"Upload"`
	Styled *multipart.FileHeader `form:"input,file" name:"styled" label:"Styled" class:"big"`
	Name   string                `form:"input,text" name:"name"   label:"Name"`
}

func TestEmptyClassesLeaveTheAttributeOut(t *testing.T) {
	// A class theme without a file class, as a site registers one.
	if _, err := templates.RegisterClassTheme("test-classes", templates.ThemeClasses{
		Input:   templates.StyleOption{Class: "control"},
		Wrapper: templates.StyleOption{Class: "field"},
	}); err != nil {
		t.Fatal(err)
	}

	f := NewForm()
	f.SetTheme("test-classes")

	html, err := f.Render(nil, WithInfo(classInput{}, Info{Target: "/x", Method: "post"}), nil)
	if err != nil {
		t.Fatal(err)
	}

	out := string(html)

	if regexp.MustCompile(`class="\s*"`).MatchString(out) || strings.Contains(out, `style=""`) {
		t.Errorf("empty class or style attributes render:\n%s", out)
	}

	upload := regexp.MustCompile(`<input type="file"[^>]*name="upload"[^>]*>`).FindString(out)
	if upload == "" || strings.Contains(upload, "class=") {
		t.Errorf("the file input without classes has a class attribute: %s", upload)
	}

	styled := regexp.MustCompile(`<input type="file"[^>]*name="styled"[^>]*>`).FindString(out)
	if !strings.Contains(styled, `class="big"`) {
		t.Errorf("the class tag is lost: %s", styled)
	}

	if !strings.Contains(out, `class="control"`) || !strings.Contains(out, `class="field"`) {
		t.Errorf("theme classes are lost:\n%s", out)
	}
}

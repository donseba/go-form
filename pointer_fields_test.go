package form

import (
	"net/url"
	"reflect"
	"strings"
	"testing"
)

type placeInput struct {
	Name      *string  `form:"input,text"   name:"name"      trim:"true"`
	Latitude  *float64 `form:"input,number" name:"latitude"              label:"Latitude"  step:"any" min:"-90"  max:"90"`
	Longitude *float64 `form:"input,number" name:"longitude"             label:"Longitude" step:"any" min:"-180" max:"180" required:"true"`
	Floors    *int     `form:"input,number" name:"floors"                                             min:"1"`
	Visitors  uint32   `form:"input,number" name:"visitors"`
}

func TestCoordinatesRender(t *testing.T) {
	latitude := 52.370216

	html := render(t, placeInput{Latitude: &latitude})

	for _, want := range []string{
		`name="latitude" value="52.370216"`,
		`name="longitude" value=""`,
		`min="-90"`, `max="90"`, `step="any"`, `min="-180"`, `max="180"`,
	} {
		if !strings.Contains(collapseSpace(html), want) {
			t.Errorf("missing %s: %s", want, html)
		}
	}

	if strings.Contains(html, "nil") || strings.Contains(html, "0x") {
		t.Errorf("a pointer renders as such: %s", html)
	}
}

// collapseSpace puts one space between the attributes of html.
func collapseSpace(html string) string {
	return strings.Join(strings.Fields(html), " ")
}

func TestCoordinatesBind(t *testing.T) {
	f := NewForm()

	cases := []struct {
		values url.Values
		want   map[string]string
		check  func(placeInput) bool
	}{
		{
			values: url.Values{"latitude": {"52.370216"}, "longitude": {"4.895168"}},
			want:   map[string]string{},
			check: func(in placeInput) bool {
				return *in.Latitude == 52.370216 && *in.Longitude == 4.895168
			},
		},
		{
			// 0 is a place on the equator, not nothing filled in.
			values: url.Values{"latitude": {"0"}, "longitude": {"0"}},
			want:   map[string]string{},
			check: func(in placeInput) bool {
				return in.Latitude != nil && *in.Latitude == 0 && in.Longitude != nil
			},
		},
		{
			values: url.Values{"latitude": {""}, "longitude": {""}},
			want:   map[string]string{"longitude": "required"},
			check: func(in placeInput) bool {
				return in.Latitude == nil && in.Longitude == nil
			},
		},
		{
			values: url.Values{"latitude": {"90.5"}, "longitude": {"-181"}},
			want:   map[string]string{"latitude": "max", "longitude": "min"},
		},
		{
			values: url.Values{"latitude": {"north"}, "longitude": {"NaN"}},
			want:   map[string]string{"latitude": "invalid", "longitude": "invalid"},
			check: func(in placeInput) bool {
				return in.Latitude == nil && in.Longitude == nil
			},
		},
		{
			values: url.Values{"longitude": {"1"}, "floors": {"0"}, "name": {"  Dam  "}, "visitors": {"4000000000"}},
			want:   map[string]string{"floors": "min"},
			check: func(in placeInput) bool {
				return *in.Floors == 0 && *in.Name == "Dam" && in.Visitors == 4000000000
			},
		},
		{
			values: url.Values{"longitude": {"1"}, "visitors": {"-1"}},
			want:   map[string]string{"visitors": "invalid"},
		},
	}

	for _, c := range cases {
		var in placeInput

		errs, err := f.Bind(postForm(c.values), &in, MapOptions{}, nil)
		if err != nil {
			t.Fatal(err)
		}

		if got := rules(errs); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%v: rules %v, want %v", c.values, got, c.want)
		}

		if c.check != nil && !c.check(in) {
			t.Errorf("%v: mapped %+v", c.values, in)
		}
	}
}

func TestPointerClearedByEmptyValue(t *testing.T) {
	latitude := 10.0
	in := placeInput{Latitude: &latitude}

	if err := MapForm(postForm(url.Values{"latitude": {""}}), &in); err != nil {
		t.Fatal(err)
	}

	if in.Latitude != nil {
		t.Errorf("latitude = %v, want nil", *in.Latitude)
	}

	// A value that is not submitted keeps the field.
	in.Latitude = &latitude

	if err := MapForm(postForm(url.Values{}), &in); err != nil || in.Latitude == nil {
		t.Errorf("latitude cleared without a value: %v", err)
	}
}

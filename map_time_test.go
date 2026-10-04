package form

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestMapFormOptionalTimes(t *testing.T) {
	type Event struct {
		Start time.Time  `form:"input,datetime-local"`
		End   *time.Time `form:"input,datetime-local"`
		Until *time.Time `form:"input,date"`
	}

	until := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	event := Event{Until: &until}

	body := url.Values{"Start": {"2026-10-04T19:30"}, "End": {"2026-10-04T22:00"}, "Until": {""}}
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(body.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err := MapForm(r, &event); err != nil {
		t.Fatal(err)
	}

	if event.Start.Format("2006-01-02 15:04") != "2026-10-04 19:30" {
		t.Errorf("start: %v", event.Start)
	}

	if event.End == nil || event.End.Format("15:04") != "22:00" {
		t.Errorf("end: %v", event.End)
	}

	if event.Until != nil {
		t.Errorf("cleared date kept: %v", event.Until)
	}
}

func TestTransformTimeValues(t *testing.T) {
	type Event struct {
		Start time.Time  `form:"input,datetime-local"`
		End   *time.Time `form:"input,datetime-local"`
		Day   time.Time  `form:"input,date"`
	}

	start := time.Date(2026, 10, 4, 19, 30, 0, 0, time.UTC)
	transformer, err := NewTransformer(Event{Start: start})
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]string{"Start": "2026-10-04T19:30", "End": "", "Day": ""}
	for _, field := range transformer.Fields {
		if field.Value != want[field.Name] {
			t.Errorf("%s: want %q, got %q", field.Name, want[field.Name], field.Value)
		}
	}
}

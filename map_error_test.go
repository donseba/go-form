package form

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

type mapErrorModel struct {
	Colour SortedSelect[string] `form:"dropdown"`
	Day    time.Time            `form:"input,date"`
}

// postMapErrorModel maps a form with an unknown select key and a malformed
// date, and returns what MapForm wrote to stdout.
func postMapErrorModel(t *testing.T, model *mapErrorModel) string {
	t.Helper()

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(url.Values{
		"Colour": {"purple"},
		"Day":    {"not a date"},
	}.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err := request.ParseForm(); err != nil {
		t.Fatal(err)
	}

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	stdout := os.Stdout
	os.Stdout = writer

	mapErr := MapForm(request, model)

	os.Stdout = stdout
	_ = writer.Close()

	written, _ := io.ReadAll(reader)
	_ = reader.Close()

	if mapErr != nil {
		t.Fatalf("MapForm: %v", mapErr)
	}

	return string(written)
}

// TestMapFormUnknownKeyIsQuiet checks that values MapForm cannot set are
// skipped without writing to stdout, and reach MapFormErrorHandler when set.
func TestMapFormUnknownKeyIsQuiet(t *testing.T) {
	model := &mapErrorModel{Colour: NewSortedSelect(map[string]string{"red": "Red"})}
	_ = model.Colour.Set("red")

	if written := postMapErrorModel(t, model); written != "" {
		t.Fatalf("MapForm wrote to stdout: %q", written)
	}

	if model.Colour.Get() != "" || !model.Day.IsZero() {
		t.Fatalf("unknown values were set: %q %v", model.Colour.Get(), model.Day)
	}

	reported := map[string]error{}

	MapFormErrorHandler = func(field string, err error) {
		reported[field] = err
	}
	t.Cleanup(func() { MapFormErrorHandler = nil })

	if written := postMapErrorModel(t, &mapErrorModel{Colour: NewSortedSelect(map[string]string{"red": "Red"})}); written != "" {
		t.Fatalf("MapForm wrote to stdout: %q", written)
	}

	if len(reported) != 2 || reported["Colour"] == nil || reported["Day"] == nil {
		t.Fatalf("handler got %v", reported)
	}
}

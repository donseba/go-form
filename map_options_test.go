package form

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"testing"
	"time"
)

// postForm is a urlencoded POST of values.
func postForm(values url.Values) *http.Request {
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return r
}

type zonedLocalizer struct{ location *time.Location }

func (zonedLocalizer) GetLocale() string { return "nl" }

func (l zonedLocalizer) Location() *time.Location { return l.location }

func TestMapFormWithLocation(t *testing.T) {
	type Event struct {
		Start time.Time `form:"input,datetime-local"`
		Day   time.Time `form:"input,date"`
	}

	amsterdam, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		t.Skip("no time zone data:", err)
	}

	var event Event

	r := postForm(url.Values{"Start": {"2026-10-04T19:30"}, "Day": {"2026-10-04"}})
	if err := MapFormWith(r, &event, MapOptions{Location: amsterdam}); err != nil {
		t.Fatal(err)
	}

	// 19:30 in Amsterdam (summer time) is 17:30 UTC.
	if got := event.Start.UTC().Format("15:04"); got != "17:30" {
		t.Errorf("date-time not read in the site's zone: %s", got)
	}

	// A date holds no instant: it stays midnight UTC.
	if !event.Day.Equal(time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("date read in a zone: %v", event.Day)
	}

	var plain Event
	if err := MapForm(postForm(url.Values{"Start": {"2026-10-04T19:30"}}), &plain); err != nil {
		t.Fatal(err)
	}

	if plain.Start.Location() != time.UTC || plain.Start.Hour() != 19 {
		t.Errorf("MapForm no longer reads date-times as UTC: %v", plain.Start)
	}

	// Rendered with a TimeLocalizer of the same zone, the value is what
	// was entered.
	html, err := NewForm().formRenderLocalized(zonedLocalizer{amsterdam}, event, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(html), `value="2026-10-04T19:30"`) {
		t.Errorf("date-time not shown in the localizer's zone: %s", html)
	}

	if !strings.Contains(string(html), `value="2026-10-04"`) {
		t.Errorf("date shown in a zone: %s", html)
	}
}

func TestMapFormTimeLayouts(t *testing.T) {
	type Times struct {
		Seconds time.Time `form:"input,datetime-local"`
		Short   time.Time `form:"input,time"`
		Month   time.Time `form:"input,month"`
	}

	var times Times

	r := postForm(url.Values{"Seconds": {"2026-10-04T19:30:15"}, "Short": {"08:45"}, "Month": {"2026-10"}})
	if err := MapForm(r, &times); err != nil {
		t.Fatal(err)
	}

	if times.Seconds.Format(time.DateTime) != "2026-10-04 19:30:15" {
		t.Errorf("date-time with seconds: %v", times.Seconds)
	}

	if times.Short.Format("15:04") != "08:45" {
		t.Errorf("time without seconds: %v", times.Short)
	}

	if times.Month.Format("2006-01") != "2026-10" {
		t.Errorf("month: %v", times.Month)
	}

	html := renderString(t, NewForm(), times, nil)
	if !strings.Contains(html, `value="2026-10"`) {
		t.Errorf("month not rendered as 2006-01: %s", html)
	}
}

func TestMapFormTrim(t *testing.T) {
	type Input struct {
		Name     string `form:"input,text"`
		Tagged   string `form:"input,text"     trim:"true"`
		Kept     string `form:"input,text"     trim:"false"`
		Password string `form:"input,password"`
	}

	values := url.Values{"Name": {"  Ada  "}, "Tagged": {" x "}, "Kept": {" y "}, "Password": {" secret "}}

	var plain Input
	if err := MapForm(postForm(values), &plain); err != nil {
		t.Fatal(err)
	}

	if plain.Name != "  Ada  " || plain.Tagged != "x" {
		t.Errorf("MapForm trimmed %+v", plain)
	}

	var trimmed Input
	if err := MapFormWith(postForm(values), &trimmed, MapOptions{TrimSpace: true}); err != nil {
		t.Fatal(err)
	}

	if trimmed.Name != "Ada" || trimmed.Tagged != "x" || trimmed.Kept != " y " || trimmed.Password != " secret " {
		t.Errorf("TrimSpace: %+v", trimmed)
	}

	var blank Input

	blank.Name = "previous"
	if err := MapFormWith(postForm(url.Values{"Name": {"   "}}), &blank, MapOptions{TrimSpace: true}); err != nil {
		t.Fatal(err)
	}

	if blank.Name != "" {
		t.Errorf("spaces only did not clear the value: %q", blank.Name)
	}
}

// multipartRequest posts fields and files ("name" → file names with their
// content).
func multipartRequest(t *testing.T, fields map[string]string, files map[string][]testFile) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for name, value := range fields {
		_ = writer.WriteField(name, value)
	}

	for name, list := range files {
		for _, file := range list {
			header := textproto.MIMEHeader{}
			header.Set("Content-Disposition", `form-data; name="`+name+`"; filename="`+file.name+`"`)
			header.Set("Content-Type", file.contentType)

			part, err := writer.CreatePart(header)
			if err != nil {
				t.Fatal(err)
			}

			_, _ = part.Write([]byte(file.content))
		}
	}

	_ = writer.Close()

	r, _ := http.NewRequest(http.MethodPost, "/", &body)
	r.Header.Set("Content-Type", writer.FormDataContentType())

	return r
}

type testFile struct {
	name, contentType, content string
}

type uploadModel struct {
	Title       string                  `form:"input,text"`
	Document    *multipart.FileHeader   `form:"input,file" label:"Document"    accept:".pdf,image/*" maxSize:"10B" required:"true"`
	Attachments []*multipart.FileHeader `form:"input,file" label:"Attachments"`
	Agree       bool                    `form:"checkbox"`
}

func TestMapFormFiles(t *testing.T) {
	r := multipartRequest(t, map[string]string{"Title": "Report", "Agree": "on"}, map[string][]testFile{
		"Document":    {{"report.pdf", "application/pdf", "%PDF"}},
		"Attachments": {{"a.txt", "text/plain", "a"}, {"b.txt", "text/plain", "b"}},
	})

	var input uploadModel
	if err := MapForm(r, &input); err != nil {
		t.Fatal(err)
	}

	if input.Title != "Report" || !input.Agree {
		t.Errorf("fields of a multipart form: %+v", input)
	}

	if input.Document == nil || input.Document.Filename != "report.pdf" {
		t.Fatalf("file not mapped: %+v", input.Document)
	}

	if len(input.Attachments) != 2 || input.Attachments[1].Filename != "b.txt" {
		t.Errorf("files not mapped: %+v", input.Attachments)
	}

	if errs := NewForm().ValidateForm(&input); len(errs) != 0 {
		t.Errorf("valid upload refused: %v", errs)
	}

	html := renderString(t, NewForm(), input, nil)
	if !strings.Contains(html, `accept=".pdf,image/*"`) || !strings.Contains(control(t, html, "Attachments"), "multiple") {
		t.Errorf("file inputs: %s", html)
	}
}

func TestValidateFiles(t *testing.T) {
	upload := func(file testFile) *multipart.FileHeader {
		r := multipartRequest(t, nil, map[string][]testFile{"Document": {file}})

		var input uploadModel
		if err := MapForm(r, &input); err != nil {
			t.Fatal(err)
		}

		return input.Document
	}

	cases := []struct {
		name string
		file *multipart.FileHeader
		want string
	}{
		{"missing", nil, "This field is required."},
		{"type", upload(testFile{"run.exe", "application/octet-stream", "MZ"}), "Choose a file of type .pdf,image/*."},
		{"image", upload(testFile{"photo.JPG", "image/jpeg", "jpeg"}), ""},
		{"size", upload(testFile{"big.pdf", "application/pdf", "0123456789AB"}), "The file must not be larger than 10B."},
	}

	for _, c := range cases {
		errs := NewForm().ValidateForm(&uploadModel{Document: c.file})

		got := strings.Join(errs.Messages("Document"), ";")
		if got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestParseSize(t *testing.T) {
	for text, want := range map[string]int64{"100": 100, "10B": 10, "1KB": 1024, "1.5 MB": 1572864, "2MiB": 2 << 20, "1g": 1 << 30} {
		got, err := ParseSize(text)
		if err != nil || got != want {
			t.Errorf("ParseSize(%q) = %d, %v; want %d", text, got, err, want)
		}
	}

	if _, err := ParseSize("big"); err == nil {
		t.Error("ParseSize accepted a size without a number")
	}
}

func TestMapFormBoolBeforeParse(t *testing.T) {
	type Input struct {
		Agree bool   `form:"checkbox"`
		Name  string `form:"input,text"`
	}

	input := Input{}
	if err := MapForm(postForm(url.Values{"Agree": {"on"}, "Name": {"x"}}), &input); err != nil {
		t.Fatal(err)
	}

	if !input.Agree {
		t.Error("a checkbox before the first other field was not mapped")
	}
}

func TestMapFormFilesAfterParseForm(t *testing.T) {
	r := multipartRequest(t, nil, map[string][]testFile{"Document": {{"report.pdf", "application/pdf", "%PDF"}}})
	_ = r.ParseForm()

	var input uploadModel
	if err := MapForm(r, &input); err != nil {
		t.Fatal(err)
	}

	if input.Document == nil {
		t.Error("file not mapped after r.ParseForm")
	}
}

// TestParseDateTimeLocalTransitions checks the times the clocks skip or show
// twice, which time.ParseInLocation reads differently per zone.
func TestParseDateTimeLocalTransitions(t *testing.T) {
	cases := []struct {
		zone  string
		value string
		want  string
	}{
		// Clocks skip 02:00–03:00: moved forward by the gap.
		{"Europe/Amsterdam", "2026-03-29T02:30", "2026-03-29T03:30:00+02:00"},
		{"America/New_York", "2026-03-08T02:30", "2026-03-08T03:30:00-04:00"},
		{"Australia/Sydney", "2026-10-04T02:30", "2026-10-04T03:30:00+11:00"},
		// Lord Howe skips half an hour.
		{"Australia/Lord_Howe", "2026-10-04T02:15", "2026-10-04T02:45:00+11:00"},
		// Clocks go back: the first of the two times.
		{"Europe/Amsterdam", "2026-10-25T02:30", "2026-10-25T02:30:00+02:00"},
		{"America/New_York", "2026-11-01T01:30", "2026-11-01T01:30:00-04:00"},
		{"Australia/Sydney", "2026-04-05T02:30", "2026-04-05T02:30:00+11:00"},
		{"Australia/Lord_Howe", "2026-04-05T01:45", "2026-04-05T01:45:00+11:00"},
		// Next to the changes, times are what they say.
		{"Europe/Amsterdam", "2026-03-29T01:59", "2026-03-29T01:59:00+01:00"},
		{"Europe/Amsterdam", "2026-03-29T03:00", "2026-03-29T03:00:00+02:00"},
		{"Europe/Amsterdam", "2026-10-25T01:59", "2026-10-25T01:59:00+02:00"},
		{"Europe/Amsterdam", "2026-10-25T03:00", "2026-10-25T03:00:00+01:00"},
		{"Europe/Amsterdam", "2026-10-25T10:00", "2026-10-25T10:00:00+01:00"},
		{"America/New_York", "2026-07-01T12:00:30", "2026-07-01T12:00:30-04:00"},
		// Santiago skips midnight.
		{"America/Santiago", "2026-09-06T00:00", "2026-09-06T01:00:00-03:00"},
	}

	for _, c := range cases {
		location, err := time.LoadLocation(c.zone)
		if err != nil {
			t.Skip("no time zone data:", err)
		}

		got, err := ParseDateTimeLocal(c.value, location)
		if err != nil {
			t.Fatal(err)
		}

		if got.Format(time.RFC3339) != c.want || got.Location() != location {
			t.Errorf("%s %s: %s, want %s", c.zone, c.value, got.Format(time.RFC3339), c.want)
		}

		// MapFormWith reads them the same way.
		var event struct {
			Start time.Time `form:"input,datetime-local"`
		}

		if err := MapFormWith(postForm(url.Values{"Start": {c.value}}), &event, MapOptions{Location: location}); err != nil {
			t.Fatal(err)
		}

		if !event.Start.Equal(got) {
			t.Errorf("%s %s: MapFormWith read %s", c.zone, c.value, event.Start.Format(time.RFC3339))
		}
	}

	if got, err := ParseDateTimeLocal(
		"2026-03-29T02:30",
		nil,
	); err != nil ||
		!got.Equal(time.Date(2026, 3, 29, 2, 30, 0, 0, time.UTC)) {
		t.Errorf("no location: %v %v", got, err)
	}

	if _, err := ParseDateTimeLocal("2026-03-29", nil); err == nil {
		t.Error("a date read as a date-time")
	}

	// WallTime reads the date and time of day of a time in any zone.
	santiago, err := time.LoadLocation("America/Santiago")
	if err != nil {
		t.Skip("no time zone data:", err)
	}

	midnight := time.Date(2026, 9, 6, 0, 0, 0, 0, time.FixedZone("elsewhere", 5*60*60))
	if got := WallTime(midnight, santiago).Format(time.RFC3339); got != "2026-09-06T01:00:00-03:00" {
		t.Errorf("WallTime: %s", got)
	}
}

package form

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

type testForm struct {
	DepartmentID SortedSelect[int64] `form:"dropdown"`
}

func TestSortedSelect_MapForm_E2E(t *testing.T) {
	f := &testForm{}
	f.DepartmentID.SetSource(map[int64]string{
		1: "Department 1",
		2: "Department 2",
		3: "Department 3",
	})

	// Simulate form submission with DepartmentID=2
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.PostForm = map[string][]string{
		"DepartmentID": {"2"},
	}
	req.Form = req.PostForm

	err := MapForm(req, f)
	if err != nil {
		t.Fatalf("MapForm failed: %v", err)
	}

	if f.DepartmentID.Get() != 2 {
		t.Errorf("Expected DepartmentID.Get()=2, got %v", f.DepartmentID.Get())
	}

	// Optionally, validate form
	// v := NewForm(nil)
	// errs := v.ValidateForm(f)
	// if len(errs) > 0 {
	//	t.Errorf("Validation errors: %v", errs)
	// }
}

func TestSortedSelect_ScanAndValue_DB(t *testing.T) {
	// int64 key
	vsInt := &SortedSelect[int64]{}
	vsInt.SetSource(map[int64]string{1: "One", 2: "Two"})
	if err := vsInt.Scan(int64(2)); err != nil {
		t.Errorf("Scan failed for int64: %v", err)
	}
	if v, _ := vsInt.Value(); v != int64(2) {
		t.Errorf("Value failed for int64: got %v", v)
	}

	// string key
	vsStr := &SortedSelect[string]{}
	vsStr.SetSource(map[string]string{"a": "Alpha", "b": "Beta"})
	if err := vsStr.Scan("b"); err != nil {
		t.Errorf("Scan failed for string: %v", err)
	}
	if v, _ := vsStr.Value(); v != "b" {
		t.Errorf("Value failed for string: got %v", v)
	}

	// float64 key
	vsFloat := &SortedSelect[float64]{}
	vsFloat.SetSource(map[float64]string{1.1: "One", 2.2: "Two"})
	if err := vsFloat.Scan(2.2); err != nil {
		t.Errorf("Scan failed for float64: %v", err)
	}
	if v, _ := vsFloat.Value(); v != 2.2 {
		t.Errorf("Value failed for float64: got %v", v)
	}

	// uuid.UUID key
	u1 := uuid.New()
	u2 := uuid.New()
	vsUUID := &SortedSelect[uuid.UUID]{}
	vsUUID.SetSource(map[uuid.UUID]string{u1: "First", u2: "Second"})
	if err := vsUUID.Scan(u2); err != nil {
		t.Errorf("Scan failed for uuid.UUID: %v", err)
	}
	if v, _ := vsUUID.Value(); v != u2 {
		t.Errorf("Value failed for uuid.UUID: got %v", v)
	}

	// time.Time key
	t1 := time.Now().Truncate(time.Second)
	t2 := t1.Add(time.Hour)
	vsTime := &SortedSelect[time.Time]{}
	vsTime.SetSource(map[time.Time]string{t1: "Now", t2: "Later"})

	if err := vsTime.Scan(t2); err != nil {
		t.Errorf("Scan failed for time.Time: %v", err)
	}
	if v, _ := vsTime.Value(); !v.(time.Time).Equal(t2) {
		t.Errorf("Value failed for time.Time: got %v", v)
	}

	// Edge case: invalid scan type
	if err := vsInt.Scan("not-an-int"); err == nil {
		t.Errorf("Expected error for invalid scan type, got nil")
	}

	// Edge case: missing key
	if err := vsStr.Scan("z"); err == nil {
		t.Errorf("Expected error for missing key, got nil")
	}
}

func TestSortedSelect_JSON(t *testing.T) {
	vs := &SortedSelect[int64]{}
	vs.SetSource(map[int64]string{1: "One", 2: "Two"})
	_ = vs.Set(2)
	data, err := json.Marshal(vs)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	var vs2 SortedSelect[int64]
	if err := json.Unmarshal(data, &vs2); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if vs2.Get() != 2 {
		t.Errorf("Expected Get()=2 after unmarshal, got %v", vs2.Get())
	}
	if len(vs2.Source()) != 2 {
		t.Errorf("Expected Source length=2 after unmarshal, got %d", len(vs2.Source()))
	}
	// Edge case: value not in Source
	bad := &SortedSelect[int64]{}
	bad.SetSource(map[int64]string{1: "One"})
	badData := []byte(`{"value":2,"source":{"1":"One"}}`)
	if err := json.Unmarshal(badData, bad); err == nil {
		t.Errorf("Expected error for value not in Source, got nil")
	}
}

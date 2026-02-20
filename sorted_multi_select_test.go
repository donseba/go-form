package form

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

type multiDept int

func TestSortedMultiSelect_SetAndGet(t *testing.T) {
	sms := &SortedMultiSelect[multiDept]{}
	sms.SetSource(map[multiDept]string{1: "HR", 2: "IT", 3: "Sales"})
	if err := sms.Set([]multiDept{1, 3}); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	got := sms.Get()
	if !reflect.DeepEqual(got, []multiDept{1, 3}) {
		t.Errorf("expected [1 3], got %v", got)
	}
}

func TestSortedMultiSelect_SetFromKeys(t *testing.T) {
	sms := &SortedMultiSelect[int]{}
	sms.SetSource(map[int]string{1: "HR", 2: "IT"})
	if err := sms.SetFromKeys([]string{"1", "2"}); err != nil {
		t.Fatalf("SetFromKeys failed: %v", err)
	}
	got := sms.Get()
	if !reflect.DeepEqual(got, []int{1, 2}) {
		t.Errorf("expected [1 2], got %v", got)
	}
}

func TestSortedMultiSelect_SetNotFoundError(t *testing.T) {
	sms := &SortedMultiSelect[int]{}
	sms.SetSource(map[int]string{1: "HR"})
	err := sms.Set([]int{2})
	if err == nil {
		t.Error("expected error for not found value, got nil")
	}
}

func TestSortedMultiSelect_ScanAndValue(t *testing.T) {
	sms := &SortedMultiSelect[int]{}
	sms.SetSource(map[int]string{1: "HR", 2: "IT"})
	if err := sms.Scan([]int{1, 2}); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	v, _ := sms.Value()
	if !reflect.DeepEqual(v, []int{1, 2}) {
		t.Errorf("expected [1 2], got %v", v)
	}
}

func TestSortedMultiSelect_JSON(t *testing.T) {
	sms := &SortedMultiSelect[int]{}
	sms.SetSource(map[int]string{1: "HR", 2: "IT"})
	_ = sms.Set([]int{1, 2})
	data, err := json.Marshal(sms)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	var sms2 SortedMultiSelect[int]
	if err := json.Unmarshal(data, &sms2); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if !reflect.DeepEqual(sms2.Get(), []int{1, 2}) {
		t.Errorf("expected [1 2] after unmarshal, got %v", sms2.Get())
	}
}

type testMultiForm struct {
	Colors SortedMultiSelect[string] `form:"multicheckbox"`
}

func TestSortedMultiSelect_MapForm_E2E(t *testing.T) {
	f := &testMultiForm{}
	f.Colors.SetSource(map[string]string{"r": "Red", "g": "Green", "b": "Blue"})

	// Simulate form submission with Colors=[r, b]
	form := make(url.Values)
	form["Colors"] = []string{"r", "b"}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm() // ParseForm populates req.PostForm
	t.Logf("PostForm: %#v, Form: %#v", req.PostForm, req.Form)

	err := MapForm(req, f)
	if err != nil {
		t.Fatalf("MapForm failed: %v", err)
	}

	got := f.Colors.Get()
	if len(got) != 2 || got[0] != "r" || got[1] != "b" {
		t.Errorf("Expected Colors.Get() = [r b], got %v", got)
	}
}

func TestSortedMultiSelect_MapForm_InvalidValue(t *testing.T) {
	f := &testMultiForm{}
	f.Colors.SetSource(map[string]string{"r": "Red", "g": "Green", "b": "Blue"})

	form := make(url.Values)
	form["Colors"] = []string{"r", "x"} // 'x' is not a valid key
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()

	err := MapForm(req, f)
	if err == nil {
		t.Errorf("Expected error for invalid value, got nil")
	}
}

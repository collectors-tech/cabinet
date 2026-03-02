package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestInventoryGradingEnumsConfigurationContract(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileCreate := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Grading Admin"}`), map[string]string{"Content-Type": "application/json"})
	if profileCreate.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", profileCreate.Code, profileCreate.Body.String())
	}
	var createdProfile map[string]any
	if err := json.Unmarshal(profileCreate.Body.Bytes(), &createdProfile); err != nil {
		t.Fatalf("decode profile create: %v", err)
	}
	profileID, _ := createdProfile["id"].(string)
	if profileID == "" {
		t.Fatalf("missing profile id in payload: %v", createdProfile)
	}
	activate := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profileID+`"}`), map[string]string{"Content-Type": "application/json"})
	if activate.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", activate.Code, activate.Body.String())
	}

	saveEnums := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/inventory/grading/enums",
		strings.NewReader(`{"condition_grades":["M","NM","EX"],"packaging_grades":["sealed_mint","sealed_good","loose"]}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if saveEnums.Code != http.StatusOK {
		t.Fatalf("save grading enums status=%d body=%s", saveEnums.Code, saveEnums.Body.String())
	}

	readEnums := doRequest(t, a, http.MethodGet, "/api/inventory/grading/enums", nil, nil)
	if readEnums.Code != http.StatusOK {
		t.Fatalf("read grading enums status=%d body=%s", readEnums.Code, readEnums.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(readEnums.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode grading enums payload: %v", err)
	}
	condition, ok := payload["condition_grades"].([]any)
	if !ok || len(condition) != 3 {
		t.Fatalf("expected condition_grades[3], got %v", payload["condition_grades"])
	}
	packaging, ok := payload["packaging_grades"].([]any)
	if !ok || len(packaging) != 3 {
		t.Fatalf("expected packaging_grades[3], got %v", payload["packaging_grades"])
	}
}

func TestInventoryGradingFieldsPersistOnItemUpdate(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"GRD-001","title":"Grading Item"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create item status=%d body=%s", create.Code, create.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create payload: %v", err)
	}
	itemID, _ := created["id"].(string)
	if itemID == "" {
		t.Fatalf("missing item id in payload: %v", created)
	}

	update := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/items/"+itemID,
		strings.NewReader(`{"grading_status":"graded","grader":"VSS","grade_numeric":9.5,"slabbed":true,"collector_classification":"collector_series","car_grade_type":"NM","packaging_grade_type":"sealed_mint"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if update.Code != http.StatusOK {
		t.Fatalf("update grading status=%d body=%s", update.Code, update.Body.String())
	}
	var updated map[string]any
	if err := json.Unmarshal(update.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode update payload: %v", err)
	}
	if updated["grading_status"] != "graded" {
		t.Fatalf("expected grading_status=graded, got %v", updated["grading_status"])
	}
	if updated["grader"] != "VSS" {
		t.Fatalf("expected grader=VSS, got %v", updated["grader"])
	}
	if updated["collector_classification"] != "collector_series" {
		t.Fatalf("expected collector_classification=collector_series, got %v", updated["collector_classification"])
	}
}

func TestInventoryGradingDefaultsApplyOnCreate(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileCreate := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Grading Defaults"}`), map[string]string{"Content-Type": "application/json"})
	if profileCreate.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", profileCreate.Code, profileCreate.Body.String())
	}
	var createdProfile map[string]any
	if err := json.Unmarshal(profileCreate.Body.Bytes(), &createdProfile); err != nil {
		t.Fatalf("decode profile create: %v", err)
	}
	profileID, _ := createdProfile["id"].(string)
	if profileID == "" {
		t.Fatalf("missing profile id in payload: %v", createdProfile)
	}
	activate := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profileID+`"}`), map[string]string{"Content-Type": "application/json"})
	if activate.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", activate.Code, activate.Body.String())
	}

	setDefaults := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/inventory/grading/defaults",
		strings.NewReader(`{"grading_status":"ungraded","priority":"medium"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if setDefaults.Code != http.StatusOK {
		t.Fatalf("set defaults status=%d body=%s", setDefaults.Code, setDefaults.Body.String())
	}

	create := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"GRD-DEFAULT-001","title":"Default Grading Item"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create item status=%d body=%s", create.Code, create.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(create.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode create payload: %v", err)
	}
	if payload["grading_status"] != "ungraded" {
		t.Fatalf("expected default grading_status=ungraded, got %v", payload["grading_status"])
	}
	if payload["priority"] != "medium" {
		t.Fatalf("expected default priority=medium, got %v", payload["priority"])
	}
}

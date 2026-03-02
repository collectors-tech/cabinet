package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestCollectionDomain004GradingEnumsAreConfigurablePerProfile(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	createProfile := func(name string) string {
		t.Helper()
		resp := doRequest(
			t,
			a,
			http.MethodPost,
			"/api/profiles",
			strings.NewReader(`{"name":"`+name+`"}`),
			map[string]string{"Content-Type": "application/json"},
		)
		if resp.Code != http.StatusCreated {
			t.Fatalf("create profile %s status=%d body=%s", name, resp.Code, resp.Body.String())
		}
		var payload map[string]any
		if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode profile %s payload: %v", name, err)
		}
		id, _ := payload["id"].(string)
		if strings.TrimSpace(id) == "" {
			t.Fatalf("missing profile id for %s", name)
		}
		return id
	}

	activateProfile := func(profileID string) {
		t.Helper()
		resp := doRequest(
			t,
			a,
			http.MethodPut,
			"/api/profiles/active",
			strings.NewReader(`{"profile_id":"`+profileID+`"}`),
			map[string]string{"Content-Type": "application/json"},
		)
		if resp.Code != http.StatusOK {
			t.Fatalf("activate profile %s status=%d body=%s", profileID, resp.Code, resp.Body.String())
		}
	}

	profileA := createProfile("Enums A")
	profileB := createProfile("Enums B")

	activateProfile(profileA)
	saveA := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/inventory/grading/enums",
		strings.NewReader(`{"condition_grades":["M","NM","A-ONLY"],"packaging_grades":["sealed_mint","A-PACK"]}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if saveA.Code != http.StatusOK {
		t.Fatalf("save enums A status=%d body=%s", saveA.Code, saveA.Body.String())
	}

	readA := doRequest(t, a, http.MethodGet, "/api/inventory/grading/enums", nil, nil)
	if readA.Code != http.StatusOK {
		t.Fatalf("read enums A status=%d body=%s", readA.Code, readA.Body.String())
	}
	var enumsA map[string]any
	if err := json.Unmarshal(readA.Body.Bytes(), &enumsA); err != nil {
		t.Fatalf("decode enums A: %v", err)
	}
	conditionA, _ := enumsA["condition_grades"].([]any)
	packagingA, _ := enumsA["packaging_grades"].([]any)
	if len(conditionA) < 3 || conditionA[2] != "a-only" {
		t.Fatalf("profile A expected a-only condition grade, got %v", conditionA)
	}
	if len(packagingA) < 2 || packagingA[1] != "a-pack" {
		t.Fatalf("profile A expected a-pack packaging grade, got %v", packagingA)
	}

	activateProfile(profileB)
	readB := doRequest(t, a, http.MethodGet, "/api/inventory/grading/enums", nil, nil)
	if readB.Code != http.StatusOK {
		t.Fatalf("read enums B status=%d body=%s", readB.Code, readB.Body.String())
	}
	var enumsB map[string]any
	if err := json.Unmarshal(readB.Body.Bytes(), &enumsB); err != nil {
		t.Fatalf("decode enums B: %v", err)
	}
	conditionB, _ := enumsB["condition_grades"].([]any)
	if len(conditionB) > 0 {
		for _, v := range conditionB {
			if v == "a-only" {
				t.Fatalf("profile B unexpectedly contains a-only condition grade: %v", conditionB)
			}
		}
	}
}

package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestPokemonGradedOverrideSaveAndFetch(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Pokemon Graded"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	setActive := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if setActive.Code != http.StatusOK {
		t.Fatalf("set active profile status=%d body=%s", setActive.Code, setActive.Body.String())
	}

	if _, err := a.db.Exec(`INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, grading_status, grader, grade_numeric, slabbed) VALUES ('poke-graded-1', ?, 'Pokemon','Cards','PK-001','Card 1','graded','VSS',9.5,1)`, profile.ID); err != nil {
		t.Fatalf("seed canonical item: %v", err)
	}

	saveResp := doRequest(t, a, http.MethodPost, "/api/integrations/pokemon/graded-overrides", strings.NewReader(`{
		"item_id":"poke-graded-1",
		"grader":"VSS",
		"grade_numeric":9.5,
		"cert_number":"CERT-001",
		"slab_state":"sealed",
		"valuation_override_amount":399.95,
		"currency":"AUD",
		"source_note":"manual_review"
	}`), map[string]string{"Content-Type": "application/json"})
	if saveResp.Code != http.StatusOK {
		t.Fatalf("save graded override status=%d body=%s", saveResp.Code, saveResp.Body.String())
	}

	getResp := doRequest(t, a, http.MethodGet, "/api/integrations/pokemon/graded-overrides?item_id=poke-graded-1", nil, nil)
	if getResp.Code != http.StatusOK {
		t.Fatalf("get graded override status=%d body=%s", getResp.Code, getResp.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(getResp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode graded override payload: %v", err)
	}
	for _, field := range []string{"item_id", "grader", "grade_numeric", "cert_number", "slab_state", "valuation_override_amount", "currency", "source_note", "overridden_at"} {
		if _, ok := payload[field]; !ok {
			t.Fatalf("missing field %q in payload %+v", field, payload)
		}
	}
}

func TestPokemonGradedOverrideRejectsMissingItemID(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodPost, "/api/integrations/pokemon/graded-overrides", strings.NewReader(`{"grader":"VSS"}`), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 status, got=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode missing item payload: %v", err)
	}
	if payload["error"] != "missing_item_id" {
		t.Fatalf("expected missing_item_id error, got %+v", payload)
	}
}

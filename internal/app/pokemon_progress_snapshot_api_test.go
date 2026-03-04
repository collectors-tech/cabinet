package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestPokemonProgressSnapshotReturnsDeterministicPayload(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Pokemon Share"}`), map[string]string{"Content-Type": "application/json"})
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

	item1 := `{"part_number":"SHARE-001","title":"Share 1","brand":"Pokemon","category":"Cards","tags":["set:base-set","language:en"],"profile_id":"` + profile.ID + `"}`
	item2 := `{"part_number":"SHARE-002","title":"Share 2","brand":"Pokemon","category":"Cards","tags":["set:base-set","language:jp"],"profile_id":"` + profile.ID + `"}`
	resp1 := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(item1), map[string]string{"Content-Type": "application/json"})
	if resp1.Code != http.StatusCreated {
		t.Fatalf("create item1 status=%d body=%s", resp1.Code, resp1.Body.String())
	}
	resp2 := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(item2), map[string]string{"Content-Type": "application/json"})
	if resp2.Code != http.StatusCreated {
		t.Fatalf("create item2 status=%d body=%s", resp2.Code, resp2.Body.String())
	}

	resp := doRequest(t, a, http.MethodGet, "/api/integrations/pokemon/progress-snapshot?set_id=base-set&total_count=10", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("snapshot status=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode snapshot payload: %v", err)
	}
	if payload["set_id"] != "base-set" {
		t.Fatalf("expected set_id base-set, got %+v", payload)
	}
	if payload["completion_percent"] != 20.0 {
		t.Fatalf("expected completion_percent 20.0, got %+v", payload["completion_percent"])
	}
	sharePayload, ok := payload["share_payload"].(map[string]any)
	if !ok {
		t.Fatalf("expected share_payload object, got %+v", payload["share_payload"])
	}
	if sharePayload["visibility"] != "private" {
		t.Fatalf("expected default visibility private, got %+v", sharePayload)
	}
	for _, field := range []string{"headline", "summary", "share_link"} {
		if _, ok := sharePayload[field]; !ok {
			t.Fatalf("missing share payload field %q: %+v", field, sharePayload)
		}
	}
	if _, ok := payload["generated_at"]; !ok {
		t.Fatalf("missing generated_at field: %+v", payload)
	}
}

func TestPokemonProgressSnapshotRejectsMissingSetID(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodGet, "/api/integrations/pokemon/progress-snapshot", nil, nil)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload["error"] != "missing_set_id" {
		t.Fatalf("expected missing_set_id, got %+v", payload)
	}
}

package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestPokemonSetProgressRejectsMissingSetID(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodGet, "/api/integrations/pokemon/set-progress", nil, nil)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for missing set_id, got %d body=%s", resp.Code, resp.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["error"] != "missing_set_id" {
		t.Fatalf("expected error missing_set_id, got %#v", payload["error"])
	}
}

func TestPokemonSetProgressComputesVariantLanguageAndGradedBreakdowns(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	for _, body := range []string{
		`{"part_number":"PKM-001","title":"Charizard","tags":["set:base-set","variant:holo","language:en"],"grading_status":"graded"}`,
		`{"part_number":"PKM-002","title":"Blastoise","tags":["set:base-set","variant:regular","language:en"],"grading_status":"ungraded"}`,
		`{"part_number":"PKM-003","title":"Venusaur","tags":["set:base-set","variant:holo","language:jp"],"grading_status":"graded"}`,
		`{"part_number":"PKM-004","title":"Pikachu","tags":["set:jungle","variant:regular","language:en"],"grading_status":"graded"}`,
	} {
		create := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
		if create.Code != http.StatusCreated {
			t.Fatalf("create item status=%d body=%s", create.Code, create.Body.String())
		}
	}

	resp := doRequest(t, a, http.MethodGet, "/api/integrations/pokemon/set-progress?set_id=base-set&total_count=10", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload["set_id"] != "base-set" {
		t.Fatalf("expected set_id=base-set, got %#v", payload["set_id"])
	}
	if payload["owned_count"] != float64(3) {
		t.Fatalf("expected owned_count=3, got %#v", payload["owned_count"])
	}
	if payload["total_count"] != float64(10) {
		t.Fatalf("expected total_count=10, got %#v", payload["total_count"])
	}
	if payload["completion_percent"] != float64(30) {
		t.Fatalf("expected completion_percent=30, got %#v", payload["completion_percent"])
	}

	breakdown, ok := payload["breakdown"].(map[string]any)
	if !ok {
		t.Fatalf("expected breakdown map, got %#v", payload["breakdown"])
	}
	variant, ok := breakdown["variant"].(map[string]any)
	if !ok || variant["holo"] != float64(2) || variant["regular"] != float64(1) {
		t.Fatalf("unexpected variant breakdown: %#v", breakdown["variant"])
	}
	language, ok := breakdown["language"].(map[string]any)
	if !ok || language["en"] != float64(2) || language["jp"] != float64(1) {
		t.Fatalf("unexpected language breakdown: %#v", breakdown["language"])
	}
	graded, ok := breakdown["graded"].(map[string]any)
	if !ok || graded["graded"] != float64(2) || graded["ungraded"] != float64(1) {
		t.Fatalf("unexpected graded breakdown: %#v", breakdown["graded"])
	}
}

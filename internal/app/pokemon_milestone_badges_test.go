package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestPokemonMilestoneEvaluateReturnsDeterministicEvents(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Pokemon Milestone"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct{ ID string `json:"id"` }
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	setActive := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if setActive.Code != http.StatusOK {
		t.Fatalf("set active profile status=%d body=%s", setActive.Code, setActive.Body.String())
	}

	for i := 1; i <= 5; i++ {
		payload := fmt.Sprintf(`{"part_number":"MILE-%03d","title":"Milestone","brand":"Pokemon","category":"Cards","tags":["set:base-set"],"profile_id":"%s"}`, i, profile.ID)
		resp := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(payload), map[string]string{"Content-Type": "application/json"})
		if resp.Code != http.StatusCreated {
			t.Fatalf("seed item status=%d body=%s", resp.Code, resp.Body.String())
		}
	}

	resp := doRequest(t, a, http.MethodPost, "/api/integrations/pokemon/milestone-evaluate", strings.NewReader(`{"set_id":"base-set","total_count":10}`), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusOK {
		t.Fatalf("milestone evaluate status=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["set_id"] != "base-set" {
		t.Fatalf("expected set_id base-set got %+v", payload)
	}
	events, ok := payload["events"].([]any)
	if !ok || len(events) != 2 {
		t.Fatalf("expected 2 milestone events at 50%% completion, got %+v", payload["events"])
	}
}

func TestPokemonMilestoneEvaluateRejectsMissingSetID(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodPost, "/api/integrations/pokemon/milestone-evaluate", strings.NewReader(`{"total_count":10}`), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got=%d body=%s", resp.Code, resp.Body.String())
	}
}

package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFrontlineDiscoveryParsesAssetAndCachesLastKnownGood(t *testing.T) {
	t.Parallel()

	assetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/assets/pd-search.js" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/javascript")
		_, _ = w.Write([]byte(`
window.Glgoliasearch('APP123','KEY456');
const indexName = 'frontline_products_live';
`))
	}))
	defer assetServer.Close()

	a := newTestApp(t)
	reqBody := `{"asset_url":"` + assetServer.URL + `/assets/pd-search.js"}`
	resp := doRequest(t, a, http.MethodPost, "/api/providers/frontline/discovery", strings.NewReader(reqBody), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusOK {
		t.Fatalf("frontline discovery status=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload struct {
		Config struct {
			ApplicationID string   `json:"application_id"`
			SearchKey     string   `json:"search_key"`
			IndexNames    []string `json:"index_names"`
			Source        string   `json:"source"`
		} `json:"config"`
		FallbackUsed bool   `json:"fallback_used"`
		Warning      string `json:"warning"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload.Config.ApplicationID != "APP123" {
		t.Fatalf("application_id mismatch got=%q", payload.Config.ApplicationID)
	}
	if payload.Config.SearchKey != "KEY456" {
		t.Fatalf("search_key mismatch got=%q", payload.Config.SearchKey)
	}
	if len(payload.Config.IndexNames) == 0 || payload.Config.IndexNames[0] != "frontline_products_live" {
		t.Fatalf("index_names mismatch got=%v", payload.Config.IndexNames)
	}
	if payload.Config.Source == "" {
		t.Fatalf("expected source to be set")
	}
	if payload.FallbackUsed {
		t.Fatal("expected fallback_used=false for successful discovery")
	}
	if payload.Warning != "" {
		t.Fatalf("expected empty warning on successful discovery, got=%q", payload.Warning)
	}
}

func TestFrontlineDiscoveryFallsBackToCachedConfigOnDriftFailure(t *testing.T) {
	t.Parallel()

	assetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/assets/pd-search.js":
			w.Header().Set("Content-Type", "application/javascript")
			_, _ = w.Write([]byte(`
window.Glgoliasearch('APP123','KEY456');
const indexName = 'frontline_products_live';
`))
		case "/assets/pd-search-drift.js":
			w.Header().Set("Content-Type", "application/javascript")
			_, _ = w.Write([]byte(`window.invalid = true;`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer assetServer.Close()

	a := newTestApp(t)
	primeBody := `{"asset_url":"` + assetServer.URL + `/assets/pd-search.js"}`
	prime := doRequest(t, a, http.MethodPost, "/api/providers/frontline/discovery", strings.NewReader(primeBody), map[string]string{"Content-Type": "application/json"})
	if prime.Code != http.StatusOK {
		t.Fatalf("prime discovery status=%d body=%s", prime.Code, prime.Body.String())
	}

	driftBody := `{"asset_url":"` + assetServer.URL + `/assets/pd-search-drift.js"}`
	drift := doRequest(t, a, http.MethodPost, "/api/providers/frontline/discovery", strings.NewReader(driftBody), map[string]string{"Content-Type": "application/json"})
	if drift.Code != http.StatusOK {
		t.Fatalf("drift fallback status=%d body=%s", drift.Code, drift.Body.String())
	}
	var payload struct {
		Config struct {
			ApplicationID string `json:"application_id"`
			SearchKey     string `json:"search_key"`
		} `json:"config"`
		FallbackUsed bool   `json:"fallback_used"`
		Warning      string `json:"warning"`
	}
	if err := json.NewDecoder(drift.Body).Decode(&payload); err != nil {
		t.Fatalf("decode drift payload: %v", err)
	}
	if !payload.FallbackUsed {
		t.Fatal("expected fallback_used=true when discovery parsing fails")
	}
	if payload.Warning == "" {
		t.Fatal("expected warning when fallback path is used")
	}
	if payload.Config.ApplicationID != "APP123" || payload.Config.SearchKey != "KEY456" {
		t.Fatalf("expected cached config values, got application_id=%q search_key=%q", payload.Config.ApplicationID, payload.Config.SearchKey)
	}
}

func TestFrontlineRunPersistsAlgoliaCandidatesAndHydratesSnapshot(t *testing.T) {
	t.Parallel()

	var algoliaAppID string
	var algoliaKey string
	searchServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/1/indexes/frontline_products_live/query" {
			http.NotFound(w, r)
			return
		}
		algoliaAppID = strings.TrimSpace(r.Header.Get("X-Algolia-Application-Id"))
		algoliaKey = strings.TrimSpace(r.Header.Get("X-Algolia-API-Key"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "hits":[
    {"objectID":"fl-1","name":"AFX Falcon Slot Car","url":"/products/fl-1","price":44.95,"image":"https://cdn.frontline/fl-1.jpg"},
    {"objectID":"fl-2","title":"AFX Mustang Slot Car","url":"https://frontlinehobbies.com.au/products/fl-2","price_value":"49.95"}
  ],
  "nbHits":2
}`))
	}))
	defer searchServer.Close()

	assetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/assets/pd-search.js" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/javascript")
		_, _ = w.Write([]byte(`
window.Glgoliasearch('APP123','KEY456');
const indexName = 'frontline_products_live';
`))
	}))
	defer assetServer.Close()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"FrontlineProfile"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	activate := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if activate.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", activate.Code, activate.Body.String())
	}

	settingsBody := fmt.Sprintf(`{"settings":{"integration.frontlinehobbies.base_url":"%s","integration.frontlinehobbies.items_per_page":"24"}}`, assetServer.URL)
	saveSettings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(settingsBody), map[string]string{"Content-Type": "application/json"})
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}

	createQuery := doRequest(t, a, http.MethodPost, "/api/scanner/query-sets", strings.NewReader(`{"name":"AFX","keywords":["AFX"],"provider_scope":["frontlinehobbies"],"enabled":true}`), map[string]string{"Content-Type": "application/json"})
	if createQuery.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", createQuery.Code, createQuery.Body.String())
	}
	var qs struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createQuery.Body).Decode(&qs); err != nil {
		t.Fatalf("decode query set: %v", err)
	}

	runBody := fmt.Sprintf(`{"query_set_id":"%s","discovery_asset_url":"%s/assets/pd-search.js","search_url":"%s/1/indexes/frontline_products_live/query"}`, qs.ID, assetServer.URL, searchServer.URL)
	run := doRequest(t, a, http.MethodPost, "/api/providers/frontline/run", strings.NewReader(runBody), map[string]string{"Content-Type": "application/json"})
	if run.Code != http.StatusOK {
		t.Fatalf("frontline run status=%d body=%s", run.Code, run.Body.String())
	}
	var payload struct {
		Total      int              `json:"total"`
		Candidates []map[string]any `json:"candidates"`
		Run        map[string]any   `json:"run"`
		RunSummary map[string]any   `json:"run_summary"`
	}
	if err := json.NewDecoder(run.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if algoliaAppID != "APP123" || algoliaKey != "KEY456" {
		t.Fatalf("expected Algolia headers from discovered config, got app=%q key=%q", algoliaAppID, algoliaKey)
	}
	if payload.Total != 2 || len(payload.Candidates) != 2 {
		t.Fatalf("expected two persisted candidates total=2, got total=%d candidates=%+v", payload.Total, payload.Candidates)
	}
	if saved, ok := payload.Run["saved"].(float64); !ok || int(saved) != 2 {
		t.Fatalf("expected persisted run saved=2, got %+v", payload.Run)
	}
	if total, ok := payload.RunSummary["candidates_total"].(float64); !ok || int(total) != 2 {
		t.Fatalf("expected run summary candidates_total=2, got %+v", payload.RunSummary)
	}
	for _, candidate := range payload.Candidates {
		source, _ := candidate["source"].(string)
		if source != "frontlinehobbies" {
			t.Fatalf("expected persisted frontline source, got %+v", candidate)
		}
	}

	reloaded := doRequest(t, a, http.MethodGet, "/api/scanner/query-sets", nil, nil)
	if reloaded.Code != http.StatusOK {
		t.Fatalf("reload query sets status=%d body=%s", reloaded.Code, reloaded.Body.String())
	}
	var querySetPayload struct {
		QuerySets []map[string]any `json:"query_sets"`
	}
	if err := json.NewDecoder(reloaded.Body).Decode(&querySetPayload); err != nil {
		t.Fatalf("decode reloaded query sets: %v", err)
	}
	if len(querySetPayload.QuerySets) != 1 {
		t.Fatalf("expected one reloaded query set, got %+v", querySetPayload.QuerySets)
	}
	latest := querySetPayload.QuerySets[0]
	if got, _ := latest["last_run_status"].(string); got != "succeeded" {
		t.Fatalf("expected frontline run to persist succeeded snapshot, got %+v", latest)
	}
	if got, ok := latest["last_candidate_count"].(float64); !ok || int(got) != 2 {
		t.Fatalf("expected frontline run to persist candidate count=2, got %+v", latest)
	}
}

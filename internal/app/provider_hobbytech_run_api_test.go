package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHobbytechRunParsesMybcappsProductsWithPagination(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/assets/hobby-search.js":
			w.Header().Set("Content-Type", "application/javascript")
			_, _ = w.Write([]byte(`
const config = { shop: 'hobbytechtoys.com.au', sid: 'SID-1', template: 'search', widget: 'main' };
`))
		case "/services/mybcapps/search":
			page := r.URL.Query().Get("page")
			w.Header().Set("Content-Type", "application/json")
			if page == "2" {
				_, _ = w.Write([]byte(`{"products":[{"id":"hb-2","title":"GFX Viper","url":"https://hobbytech/item/2","price":"29.95"}],"pagination":{"total_pages":2}}`))
				return
			}
			_, _ = w.Write([]byte(`{"products":[{"id":"hb-1","title":"GFX Camaro","url":"https://hobbytech/item/1","price":"24.95"}],"pagination":{"total_pages":2}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"HobbyProfile"}`), map[string]string{"Content-Type": "application/json"})
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

	settingsBody := fmt.Sprintf(`{"settings":{"integration.hobbytechtoys.base_url":"%s","integration.hobbytechtoys.items_per_page":"24"}}`, server.URL)
	saveSettings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(settingsBody), map[string]string{"Content-Type": "application/json"})
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}

	createQuery := doRequest(t, a, http.MethodPost, "/api/scanner/query-sets", strings.NewReader(`{"name":"GFX","keywords":["GFX"],"provider_scope":["hobbytechtoys"],"enabled":true}`), map[string]string{"Content-Type": "application/json"})
	if createQuery.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", createQuery.Code, createQuery.Body.String())
	}
	var qs struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createQuery.Body).Decode(&qs); err != nil {
		t.Fatalf("decode query set: %v", err)
	}

	runBody := fmt.Sprintf(`{"query_set_id":"%s","discovery_asset_url":"%s/assets/hobby-search.js","search_url":"%s/services/mybcapps/search"}`, qs.ID, server.URL, server.URL)
	run := doRequest(t, a, http.MethodPost, "/api/providers/hobbytech/run", strings.NewReader(runBody), map[string]string{"Content-Type": "application/json"})
	if run.Code != http.StatusOK {
		t.Fatalf("hobbytech run status=%d body=%s", run.Code, run.Body.String())
	}
	var payload struct {
		PageCount      int              `json:"page_count"`
		Candidates     []map[string]any `json:"candidates"`
		DriftRecovered bool             `json:"drift_recovered"`
		Run            map[string]any   `json:"run"`
		RunSummary     map[string]any   `json:"run_summary"`
	}
	if err := json.NewDecoder(run.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload.PageCount != 2 {
		t.Fatalf("expected page_count=2 got=%d", payload.PageCount)
	}
	if len(payload.Candidates) != 2 {
		t.Fatalf("expected 2 candidates got=%d", len(payload.Candidates))
	}
	if payload.DriftRecovered {
		t.Fatal("expected drift_recovered=false for healthy run")
	}
	if saved, ok := payload.Run["saved"].(float64); !ok || int(saved) != 2 {
		t.Fatalf("expected persisted run saved=2, got %+v", payload.Run)
	}
	if total, ok := payload.RunSummary["candidates_total"].(float64); !ok || int(total) != 2 {
		t.Fatalf("expected run summary candidates_total=2, got %+v", payload.RunSummary)
	}
	for _, candidate := range payload.Candidates {
		source, _ := candidate["source"].(string)
		if source != "hobbytechtoys" {
			t.Fatalf("expected persisted hobbytech source, got %+v", candidate)
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
		t.Fatalf("expected hobbytech run to persist succeeded snapshot, got %+v", latest)
	}
	if got, ok := latest["last_candidate_count"].(float64); !ok || int(got) != 2 {
		t.Fatalf("expected hobbytech run to persist candidate count=2, got %+v", latest)
	}
}

func TestHobbytechRunRecoversFromSessionDriftWithFallbackDiscovery(t *testing.T) {
	t.Parallel()

	firstSearchAttempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/assets/hobby-search-old.js":
			w.Header().Set("Content-Type", "application/javascript")
			_, _ = w.Write([]byte(`const cfg={shop:'hobbytechtoys.com.au',sid:'OLD-SID',template:'search'};`))
		case "/assets/hobby-search-new.js":
			w.Header().Set("Content-Type", "application/javascript")
			_, _ = w.Write([]byte(`const cfg={shop:'hobbytechtoys.com.au',sid:'NEW-SID',template:'search'};`))
		case "/services/mybcapps/search":
			firstSearchAttempt++
			if firstSearchAttempt == 1 {
				http.Error(w, "session expired", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"products":[{"id":"hb-9","title":"GFX Corvette","url":"https://hobbytech/item/9","price":"34.95"}],"pagination":{"total_pages":1}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"HobbyDriftProfile"}`), map[string]string{"Content-Type": "application/json"})
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

	settingsBody := fmt.Sprintf(`{"settings":{"integration.hobbytechtoys.base_url":"%s","integration.hobbytechtoys.items_per_page":"24"}}`, server.URL)
	saveSettings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(settingsBody), map[string]string{"Content-Type": "application/json"})
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}

	createQuery := doRequest(t, a, http.MethodPost, "/api/scanner/query-sets", strings.NewReader(`{"name":"GFX","keywords":["GFX"],"provider_scope":["hobbytechtoys"],"enabled":true}`), map[string]string{"Content-Type": "application/json"})
	if createQuery.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", createQuery.Code, createQuery.Body.String())
	}
	var qs struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createQuery.Body).Decode(&qs); err != nil {
		t.Fatalf("decode query set: %v", err)
	}

	runBody := fmt.Sprintf(`{"query_set_id":"%s","discovery_asset_url":"%s/assets/hobby-search-old.js","fallback_discovery_asset_urls":["%s/assets/hobby-search-new.js"],"search_url":"%s/services/mybcapps/search"}`, qs.ID, server.URL, server.URL, server.URL)
	run := doRequest(t, a, http.MethodPost, "/api/providers/hobbytech/run", strings.NewReader(runBody), map[string]string{"Content-Type": "application/json"})
	if run.Code != http.StatusOK {
		t.Fatalf("hobbytech run status=%d body=%s", run.Code, run.Body.String())
	}
	var payload struct {
		Candidates     []map[string]any `json:"candidates"`
		DriftRecovered bool             `json:"drift_recovered"`
		Warning        string           `json:"warning"`
	}
	if err := json.NewDecoder(run.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if !payload.DriftRecovered {
		t.Fatal("expected drift_recovered=true after fallback retry")
	}
	if payload.Warning == "" {
		t.Fatal("expected warning message for drift recovery")
	}
	if len(payload.Candidates) != 1 {
		t.Fatalf("expected one candidate after recovery got=%d", len(payload.Candidates))
	}
}

func TestHobbytechRunFallsBackToPublicShopifySuggestWhenBoostDiscoveryIsUnavailable(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/assets/hobby-search.js":
			http.NotFound(w, r)
		case "/services.mybcapps.com/bc-sf-filter/search":
			http.Error(w, "boost direct public request forbidden", http.StatusForbidden)
		case "/search/suggest.json":
			if got := r.URL.Query().Get("q"); got != "AFX" {
				t.Fatalf("expected suggest query AFX, got %q", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"resources": {
					"results": {
						"products": [
							{
								"id": 70615,
								"title": "AFX Low Bridge Supports",
								"url": "/products/afx-low-bridge-supports?_pos=2&_psq=AFX&_psid=session-looking-value&_ss=e#results",
								"price": "12.95",
								"available": true,
								"image": "https://hobbytechtoys.com.au/cdn/afx-low-bridge.jpg?v=70615"
							}
						]
					}
				}
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"HobbySuggestProfile"}`), map[string]string{"Content-Type": "application/json"})
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

	settingsBody := fmt.Sprintf(`{"settings":{"integration.hobbytechtoys.base_url":"%s","integration.hobbytechtoys.items_per_page":"12"}}`, server.URL)
	saveSettings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(settingsBody), map[string]string{"Content-Type": "application/json"})
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}

	createQuery := doRequest(t, a, http.MethodPost, "/api/scanner/query-sets", strings.NewReader(`{"name":"AFX","keywords":["AFX"],"provider_scope":["hobbytechtoys"],"enabled":true}`), map[string]string{"Content-Type": "application/json"})
	if createQuery.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", createQuery.Code, createQuery.Body.String())
	}
	var qs struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createQuery.Body).Decode(&qs); err != nil {
		t.Fatalf("decode query set: %v", err)
	}

	runBody := fmt.Sprintf(`{"query_set_id":"%s"}`, qs.ID)
	run := doRequest(t, a, http.MethodPost, "/api/providers/hobbytech/run", strings.NewReader(runBody), map[string]string{"Content-Type": "application/json"})
	if run.Code != http.StatusOK {
		t.Fatalf("hobbytech run should fall back to public Shopify suggest, status=%d body=%s", run.Code, run.Body.String())
	}
	var payload struct {
		Candidates []map[string]any `json:"candidates"`
		RunSummary map[string]any   `json:"run_summary"`
	}
	if err := json.NewDecoder(run.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if len(payload.Candidates) != 1 {
		t.Fatalf("expected one Shopify suggest fallback candidate, got %+v", payload.Candidates)
	}
	candidate := payload.Candidates[0]
	if source, _ := candidate["source"].(string); source != "hobbytechtoys" {
		t.Fatalf("expected Hobbytech source from fallback, got %+v", candidate)
	}
	if productURL, _ := candidate["url"].(string); productURL != server.URL+"/products/afx-low-bridge-supports" {
		t.Fatalf("expected canonical Hobbytech product URL without search tracking, got %+v", candidate)
	}
	if imageURL, _ := candidate["image"].(string); imageURL != "https://hobbytechtoys.com.au/cdn/afx-low-bridge.jpg?v=70615" {
		t.Fatalf("expected Hobbytech image version query to remain intact, got %+v", candidate)
	}
	if method, _ := payload.RunSummary["data_depth_source"].(string); method != "shopify_search_suggest_json" {
		t.Fatalf("expected Shopify suggest fallback run summary, got %+v", payload.RunSummary)
	}
}

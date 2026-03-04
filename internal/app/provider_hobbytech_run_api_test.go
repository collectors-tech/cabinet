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
		PageCount     int              `json:"page_count"`
		Candidates    []map[string]any `json:"candidates"`
		DriftRecovered bool            `json:"drift_recovered"`
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


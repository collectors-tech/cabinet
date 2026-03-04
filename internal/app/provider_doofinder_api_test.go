package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDoofinderDiscoveryExtractsStoreZoneHashIDAndCaches(t *testing.T) {
	t.Parallel()

	assetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/assets/df.js":
			w.Header().Set("Content-Type", "application/javascript")
			_, _ = w.Write([]byte(`
window.doofinder_config = {
  store: "5f1f7b26-6e08-4aac-a336-9f008fcb5315",
  zone: "eu1",
  search_engines: {
    products: { hashid: "df-abc-123" }
  }
};
`))
		case "/assets/df-drift.js":
			w.Header().Set("Content-Type", "application/javascript")
			_, _ = w.Write([]byte(`window.no_doofinder_here = true;`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer assetServer.Close()

	a := newTestApp(t)
	discovery := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/providers/doofinder/discovery",
		strings.NewReader(fmt.Sprintf(`{"asset_url":"%s/assets/df.js"}`, assetServer.URL)),
		map[string]string{"Content-Type": "application/json"},
	)
	if discovery.Code != http.StatusOK {
		t.Fatalf("doofinder discovery status=%d body=%s", discovery.Code, discovery.Body.String())
	}
	var payload struct {
		Config struct {
			Store   string `json:"store"`
			Zone    string `json:"zone"`
			HashID  string `json:"hashid"`
			Source  string `json:"source"`
			APIBase string `json:"api_base"`
		} `json:"config"`
		FallbackUsed bool   `json:"fallback_used"`
		Warning      string `json:"warning"`
	}
	if err := json.NewDecoder(discovery.Body).Decode(&payload); err != nil {
		t.Fatalf("decode discovery payload: %v", err)
	}
	if payload.Config.Store != "5f1f7b26-6e08-4aac-a336-9f008fcb5315" {
		t.Fatalf("unexpected store=%q", payload.Config.Store)
	}
	if payload.Config.Zone != "eu1" {
		t.Fatalf("unexpected zone=%q", payload.Config.Zone)
	}
	if payload.Config.HashID != "df-abc-123" {
		t.Fatalf("unexpected hashid=%q", payload.Config.HashID)
	}
	if !strings.Contains(payload.Config.APIBase, "doofinder.com") {
		t.Fatalf("expected api_base to include doofinder domain, got=%q", payload.Config.APIBase)
	}
	if payload.FallbackUsed {
		t.Fatal("expected fallback_used=false on successful parse")
	}
	if payload.Warning != "" {
		t.Fatalf("expected empty warning, got=%q", payload.Warning)
	}

	fallback := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/providers/doofinder/discovery",
		strings.NewReader(fmt.Sprintf(`{"asset_url":"%s/assets/df-drift.js"}`, assetServer.URL)),
		map[string]string{"Content-Type": "application/json"},
	)
	if fallback.Code != http.StatusOK {
		t.Fatalf("doofinder fallback discovery status=%d body=%s", fallback.Code, fallback.Body.String())
	}
	var fallbackPayload struct {
		Config struct {
			HashID string `json:"hashid"`
		} `json:"config"`
		FallbackUsed bool   `json:"fallback_used"`
		Warning      string `json:"warning"`
	}
	if err := json.NewDecoder(fallback.Body).Decode(&fallbackPayload); err != nil {
		t.Fatalf("decode fallback payload: %v", err)
	}
	if !fallbackPayload.FallbackUsed {
		t.Fatal("expected fallback_used=true when parse drifts")
	}
	if fallbackPayload.Warning == "" {
		t.Fatal("expected warning for cached fallback")
	}
	if fallbackPayload.Config.HashID != "df-abc-123" {
		t.Fatalf("expected cached hashid, got=%q", fallbackPayload.Config.HashID)
	}
}

func TestDoofinderRunUsesOriginAwareHeadersAndNormalizesCandidates(t *testing.T) {
	t.Parallel()

	lastOrigin := ""
	lastReferer := ""
	dfServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/5/search" {
			http.NotFound(w, r)
			return
		}
		lastOrigin = strings.TrimSpace(r.Header.Get("Origin"))
		lastReferer = strings.TrimSpace(r.Header.Get("Referer"))
		if lastOrigin == "" || lastReferer == "" {
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "results":[
    {"id":"df-1","title":"AFX Mega G+ Camaro","link":"https://www.mrtoys.com.au/p/df-1","price":"49.95","image_link":"https://cdn.mrtoys.com.au/img/df-1.jpg"},
    {"id":"df-2","title":"AFX Mega G+ Corvette","link":"https://www.mrtoys.com.au/p/df-2","price":"52.95","image_link":"https://cdn.mrtoys.com.au/img/df-2.jpg"}
  ],
  "meta":{"total":2}
}`))
	}))
	defer dfServer.Close()

	assetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/assets/df.js" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/javascript")
		_, _ = w.Write([]byte(`
window.doofinder_config = {
  store: "5f1f7b26-6e08-4aac-a336-9f008fcb5315",
  zone: "eu1",
  search_engines: {
    products: { hashid: "df-abc-123" }
  }
};
`))
	}))
	defer assetServer.Close()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"DoofinderProfile"}`), map[string]string{"Content-Type": "application/json"})
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

	settingsBody := fmt.Sprintf(`{"settings":{"integration.mrtoys-com-au.base_url":"https://www.mrtoys.com.au","integration.mrtoys-com-au.provider_scope":"doofinder"}}`)
	saveSettings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(settingsBody), map[string]string{"Content-Type": "application/json"})
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}

	createQuery := doRequest(t, a, http.MethodPost, "/api/scanner/query-sets", strings.NewReader(`{"name":"AFX","keywords":["AFX"],"provider_scope":["mrtoys.com.au"],"enabled":true}`), map[string]string{"Content-Type": "application/json"})
	if createQuery.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", createQuery.Code, createQuery.Body.String())
	}
	var qs struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createQuery.Body).Decode(&qs); err != nil {
		t.Fatalf("decode query set: %v", err)
	}

	run := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/providers/doofinder/run",
		strings.NewReader(fmt.Sprintf(`{
  "query_set_id":"%s",
  "asset_url":"%s/assets/df.js",
  "search_url":"%s/5/search",
  "provider_domain":"mrtoys.com.au",
  "page_size":24
}`, qs.ID, assetServer.URL, dfServer.URL)),
		map[string]string{"Content-Type": "application/json"},
	)
	if run.Code != http.StatusOK {
		t.Fatalf("doofinder run status=%d body=%s", run.Code, run.Body.String())
	}
	var payload struct {
		CandidateCount int              `json:"candidate_count"`
		Total          int              `json:"total"`
		Candidates     []map[string]any `json:"candidates"`
		Discovery      map[string]any   `json:"discovery"`
	}
	if err := json.NewDecoder(run.Body).Decode(&payload); err != nil {
		t.Fatalf("decode run payload: %v", err)
	}
	if payload.CandidateCount != 2 || payload.Total != 2 {
		t.Fatalf("expected two normalized candidates total=2, got candidate_count=%d total=%d", payload.CandidateCount, payload.Total)
	}
	if len(payload.Candidates) != 2 {
		t.Fatalf("expected two candidates, got=%d", len(payload.Candidates))
	}
	for _, field := range []string{"listing_id", "title", "url", "price", "image", "seller", "source"} {
		if _, ok := payload.Candidates[0][field]; !ok {
			t.Fatalf("candidate missing field %q: %+v", field, payload.Candidates[0])
		}
	}
	if payload.Discovery["hashid"] != "df-abc-123" {
		t.Fatalf("expected discovery hashid, got=%v", payload.Discovery["hashid"])
	}
	if payload.Discovery["source"] == "" {
		t.Fatalf("expected discovery source in telemetry, got=%v", payload.Discovery)
	}
	if lastOrigin != "https://www.mrtoys.com.au" {
		t.Fatalf("expected origin header to match provider domain, got=%q", lastOrigin)
	}
	if !strings.HasPrefix(lastReferer, "https://www.mrtoys.com.au") {
		t.Fatalf("expected referer header to match provider domain, got=%q", lastReferer)
	}
}


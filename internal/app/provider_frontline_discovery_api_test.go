package app

import (
	"encoding/json"
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


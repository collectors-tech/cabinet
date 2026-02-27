package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWave4ProvidersRegistryContract(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodGet, "/api/providers/registry", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("providers registry expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var payload struct {
		Providers []map[string]any `json:"providers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode providers payload: %v", err)
	}
	if len(payload.Providers) == 0 {
		t.Fatal("expected provider records in registry payload")
	}

	var amazon map[string]any
	domains := map[string]bool{}
	for _, p := range payload.Providers {
		if p["provider_id"] == "amazon" {
			amazon = p
		}
		if d, ok := p["base_domain"].(string); ok {
			domains[d] = true
		}
	}
	if amazon == nil {
		t.Fatalf("amazon provider not found in registry payload: %+v", payload.Providers)
	}
	for _, field := range []string{"integration_mode", "eligibility_required", "policy_scope_note", "capabilities", "state"} {
		if _, ok := amazon[field]; !ok {
			t.Fatalf("amazon provider missing %q: %+v", field, amazon)
		}
	}
	for _, d := range []string{
		"bonzaslotcars.com.au",
		"frontlinehobbies.com.au",
		"hobbytechtoys.com.au",
		"andrewshobbies.com.au",
		"voglers.com.au",
		"acercmodels.com",
		"mrtoys.com.au",
	} {
		if !domains[d] {
			t.Fatalf("registry missing AU webshop domain %q", d)
		}
	}
}

func TestWave4AmazonRunModeAndNormalizationContract(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wave4Amazon"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p.ID+`"}`), map[string]string{"Content-Type": "application/json"})

	createQuery := doRequest(t, a, http.MethodPost, "/api/scanner/query-sets", strings.NewReader(`{"name":"Q1","keywords":["afx"],"region":"AU","enabled":true}`), map[string]string{"Content-Type": "application/json"})
	if createQuery.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", createQuery.Code, createQuery.Body.String())
	}
	var qs map[string]any
	if err := json.NewDecoder(createQuery.Body).Decode(&qs); err != nil {
		t.Fatalf("decode query set: %v", err)
	}
	querySetID, _ := qs["id"].(string)
	if querySetID == "" {
		t.Fatal("expected query set id")
	}

	if _, err := a.db.Exec(`INSERT INTO app_state(key, value, updated_at) VALUES('provider.amazon.mode','program_api',CURRENT_TIMESTAMP) ON CONFLICT(key) DO UPDATE SET value='program_api', updated_at=CURRENT_TIMESTAMP`); err != nil {
		t.Fatalf("set amazon mode: %v", err)
	}

	okRun := doRequest(t, a, http.MethodPost, "/api/providers/amazon/run", strings.NewReader(`{"query_set_id":"`+querySetID+`"}`), map[string]string{"Content-Type": "application/json"})
	if okRun.Code != http.StatusOK {
		t.Fatalf("amazon run expected 200, got %d body=%s", okRun.Code, okRun.Body.String())
	}
	var okPayload struct {
		Candidates []map[string]any `json:"candidates"`
	}
	if err := json.NewDecoder(okRun.Body).Decode(&okPayload); err != nil {
		t.Fatalf("decode amazon run payload: %v", err)
	}
	if len(okPayload.Candidates) == 0 {
		t.Fatalf("expected normalized amazon candidates, got %+v", okPayload)
	}
	first := okPayload.Candidates[0]
	for _, field := range []string{"listing_id", "title", "price", "url", "seller", "source"} {
		if _, ok := first[field]; !ok {
			t.Fatalf("normalized candidate missing %q: %+v", field, first)
		}
	}

	if _, err := a.db.Exec(`INSERT INTO app_state(key, value, updated_at) VALUES('provider.amazon.mode','disabled',CURRENT_TIMESTAMP) ON CONFLICT(key) DO UPDATE SET value='disabled', updated_at=CURRENT_TIMESTAMP`); err != nil {
		t.Fatalf("disable amazon mode: %v", err)
	}
	disabledRun := doRequest(t, a, http.MethodPost, "/api/providers/amazon/run", strings.NewReader(`{"query_set_id":"`+querySetID+`"}`), map[string]string{"Content-Type": "application/json"})
	if disabledRun.Code != http.StatusConflict {
		t.Fatalf("amazon disabled run expected 409, got %d body=%s", disabledRun.Code, disabledRun.Body.String())
	}
	if !strings.Contains(disabledRun.Body.String(), `"error_code":"PROVIDER_DISABLED"`) {
		t.Fatalf("expected PROVIDER_DISABLED envelope, got %s", disabledRun.Body.String())
	}
}

func TestWave4AUWebshopStockExtractionContract(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/providers/au-webshops/parse-stock",
		strings.NewReader(`{"domain":"bonzaslotcars.com.au","html":"<span>Only 2 left in stock</span>"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if resp.Code != http.StatusOK {
		t.Fatalf("au parse stock expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode au stock payload: %v", err)
	}
	ss, ok := payload["stock_signal"].(map[string]any)
	if !ok {
		t.Fatalf("expected stock_signal object, got %+v", payload)
	}
	for _, field := range []string{"raw", "normalized_state", "source_domain", "observed_at"} {
		if _, ok := ss[field]; !ok {
			t.Fatalf("stock_signal missing %q: %+v", field, ss)
		}
	}
}

func TestWave4ScannerScheduledSummaryAndCandidateDedup(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|100|0","title":"AFX Mega G+","price":{"value":"45.00"},"itemWebUrl":"https://ebay/item/100","seller":{"username":"seller1"}}]}`))
	}))
	defer server.Close()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wave4Scanner"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	_ = doRequest(t, a, http.MethodPut, "/api/profiles/"+p.ID+"/settings", strings.NewReader(`{"settings":{"ebay_base_url":"`+server.URL+`","ebay_bearer_token":"token","ebay_marketplace":"EBAY_AU"}}`), map[string]string{"Content-Type": "application/json"})

	createQuery := doRequest(t, a, http.MethodPost, "/api/scanner/query-sets", strings.NewReader(`{"name":"Q1","keywords":["afx"],"schedule_cron":"*/15 * * * *","enabled":true,"rate_limit_rps":2}`), map[string]string{"Content-Type": "application/json"})
	if createQuery.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", createQuery.Code, createQuery.Body.String())
	}
	var qs map[string]any
	if err := json.NewDecoder(createQuery.Body).Decode(&qs); err != nil {
		t.Fatalf("decode query set: %v", err)
	}
	querySetID, _ := qs["id"].(string)

	scheduled := doRequest(t, a, http.MethodPost, "/api/scanner/run/scheduled", strings.NewReader(`{}`), map[string]string{"Content-Type": "application/json"})
	if scheduled.Code != http.StatusOK {
		t.Fatalf("scheduled run expected 200, got %d body=%s", scheduled.Code, scheduled.Body.String())
	}
	var summary map[string]any
	if err := json.NewDecoder(scheduled.Body).Decode(&summary); err != nil {
		t.Fatalf("decode scheduled summary: %v", err)
	}
	for _, field := range []string{"run_id", "query_sets_executed", "candidates_collected", "failures"} {
		if _, ok := summary[field]; !ok {
			t.Fatalf("scheduled summary missing %q: %+v", field, summary)
		}
	}

	run1 := doRequest(t, a, http.MethodPost, "/api/scanner/run", strings.NewReader(`{"query_set_id":"`+querySetID+`"}`), map[string]string{"Content-Type": "application/json"})
	if run1.Code != http.StatusOK {
		t.Fatalf("run1 status=%d body=%s", run1.Code, run1.Body.String())
	}
	run2 := doRequest(t, a, http.MethodPost, "/api/scanner/run", strings.NewReader(`{"query_set_id":"`+querySetID+`"}`), map[string]string{"Content-Type": "application/json"})
	if run2.Code != http.StatusOK {
		t.Fatalf("run2 status=%d body=%s", run2.Code, run2.Body.String())
	}

	candidates := doRequest(t, a, http.MethodGet, "/api/scanner/candidates?query_set_id="+querySetID, nil, nil)
	if candidates.Code != http.StatusOK {
		t.Fatalf("list candidates status=%d body=%s", candidates.Code, candidates.Body.String())
	}
	var cPayload struct {
		Candidates []map[string]any `json:"candidates"`
	}
	if err := json.NewDecoder(candidates.Body).Decode(&cPayload); err != nil {
		t.Fatalf("decode candidates payload: %v", err)
	}
	if len(cPayload.Candidates) != 1 {
		t.Fatalf("expected deduplicated candidate count=1, got %d payload=%+v", len(cPayload.Candidates), cPayload.Candidates)
	}
	first := cPayload.Candidates[0]
	for _, field := range []string{"listing_id", "title", "url", "seller", "first_seen", "last_seen", "stock_state"} {
		if _, ok := first[field]; !ok {
			t.Fatalf("candidate missing %q: %+v", field, first)
		}
	}
}

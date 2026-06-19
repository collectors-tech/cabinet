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
		for _, field := range []string{"integration_mode", "api_available", "auth_requirement"} {
			if _, ok := p[field]; !ok {
				t.Fatalf("provider %v missing %q: %+v", p["provider_id"], field, p)
			}
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
	for _, field := range []string{"has_token", "setup_instructions", "health", "last_run"} {
		if _, ok := amazon[field]; !ok {
			t.Fatalf("amazon provider missing %q: %+v", field, amazon)
		}
	}
	health, ok := amazon["health"].(map[string]any)
	if !ok {
		t.Fatalf("amazon provider health must be object: %+v", amazon)
	}
	for _, field := range []string{"status", "last_checked_at", "message"} {
		if _, ok := health[field]; !ok {
			t.Fatalf("amazon provider health missing %q: %+v", field, health)
		}
	}
	lastRun, ok := amazon["last_run"].(map[string]any)
	if !ok {
		t.Fatalf("amazon provider last_run must be object: %+v", amazon)
	}
	for _, field := range []string{"status", "finished_at"} {
		if _, ok := lastRun[field]; !ok {
			t.Fatalf("amazon provider last_run missing %q: %+v", field, lastRun)
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
		"hobbyco.com.au",
		"metrohobbies.com.au",
	} {
		if !domains[d] {
			t.Fatalf("registry missing AU webshop domain %q", d)
		}
	}
}

func TestWave4AUWebshopDomainsConfigSourceContract(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wave4AUConfig"}`), map[string]string{"Content-Type": "application/json"})
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

	settingsBody := `{"settings":{"integration.au_webshops.domains":"bonzaslotcars.com.au, metrohobbies.com.au, customcollector.example"}}`
	updateSettings := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/profiles/"+p.ID+"/settings",
		strings.NewReader(settingsBody),
		map[string]string{"Content-Type": "application/json"},
	)
	if updateSettings.Code != http.StatusOK {
		t.Fatalf("update settings status=%d body=%s", updateSettings.Code, updateSettings.Body.String())
	}

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

	auDomains := map[string]bool{}
	for _, provider := range payload.Providers {
		providerID, _ := provider["provider_id"].(string)
		if !strings.HasPrefix(providerID, "au-webshop-") {
			continue
		}
		domain, _ := provider["base_domain"].(string)
		if domain != "" {
			auDomains[domain] = true
		}
	}
	if len(auDomains) != 3 {
		t.Fatalf("expected exactly 3 config-driven AU domains, got %d domains=%v", len(auDomains), auDomains)
	}
	for _, expected := range []string{
		"bonzaslotcars.com.au",
		"metrohobbies.com.au",
		"customcollector.example",
	} {
		if !auDomains[expected] {
			t.Fatalf("registry missing config-driven AU domain %q from %v", expected, auDomains)
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
	source, ok := first["source"].(map[string]any)
	if !ok {
		t.Fatalf("expected source provider object in provider run payload, got %+v", first["source"])
	}
	if got, _ := source["provider_id"].(string); got != "amazon" {
		t.Fatalf("expected amazon source in provider run payload, got %+v", first["source"])
	}

	candidates := doRequest(t, a, http.MethodGet, "/api/scanner/candidates?query_set_id="+querySetID, nil, nil)
	if candidates.Code != http.StatusOK {
		t.Fatalf("list amazon candidates expected 200, got %d body=%s", candidates.Code, candidates.Body.String())
	}
	var candidatePayload struct {
		Candidates []map[string]any `json:"candidates"`
	}
	if err := json.NewDecoder(candidates.Body).Decode(&candidatePayload); err != nil {
		t.Fatalf("decode persisted amazon candidates: %v", err)
	}
	if len(candidatePayload.Candidates) != 1 {
		t.Fatalf("expected one persisted amazon candidate, got %+v", candidatePayload.Candidates)
	}
	if got, _ := candidatePayload.Candidates[0]["source"].(string); got != "amazon" {
		t.Fatalf("expected persisted amazon source, got %+v", candidatePayload.Candidates[0])
	}

	reloaded := doRequest(t, a, http.MethodGet, "/api/scanner/query-sets", nil, nil)
	if reloaded.Code != http.StatusOK {
		t.Fatalf("reload query sets expected 200, got %d body=%s", reloaded.Code, reloaded.Body.String())
	}
	var querySetPayload struct {
		QuerySets []map[string]any `json:"query_sets"`
	}
	if err := json.NewDecoder(reloaded.Body).Decode(&querySetPayload); err != nil {
		t.Fatalf("decode query sets after amazon run: %v", err)
	}
	if len(querySetPayload.QuerySets) != 1 {
		t.Fatalf("expected one query set after amazon run, got %+v", querySetPayload.QuerySets)
	}
	latest := querySetPayload.QuerySets[0]
	if got, _ := latest["last_run_status"].(string); got != "succeeded" {
		t.Fatalf("expected amazon run to persist succeeded snapshot, got %+v", latest)
	}
	if got, ok := latest["last_candidate_count"].(float64); !ok || int(got) != 1 {
		t.Fatalf("expected amazon run to persist candidate count=1, got %+v", latest)
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

func TestWave4EbayRunPersistsSavedSearchCandidates(t *testing.T) {
	t.Parallel()

	ebayStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("expected bearer token header, got %q", got)
		}
		if got := r.Header.Get("X-EBAY-C-MARKETPLACE-ID"); got != "EBAY_AU" {
			t.Fatalf("expected EBAY_AU marketplace header, got %q", got)
		}
		if got := r.URL.Query().Get("q"); got != "afx" {
			t.Fatalf("expected query q=afx, got %q", got)
		}
		if got := r.URL.Query().Get("limit"); got != "37" {
			t.Fatalf("expected provider-run Browse limit from integration.ebay.items_per_page, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"ebay-1001","title":"AFX Slot Car eBay Listing","price":{"value":"29.95","currency":"AUD"},"itemWebUrl":"https://www.ebay.com.au/itm/ebay-1001","image":{"imageUrl":"https://example.test/ebay-1001.jpg"},"seller":{"username":"seller-one"},"estimatedAvailabilities":[{"estimatedAvailabilityStatus":"LIMITED_STOCK","estimatedAvailableQuantity":2}]}]}`))
	}))
	defer ebayStub.Close()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wave4Ebay"}`), map[string]string{"Content-Type": "application/json"})
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
	saveSettings := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/profiles/"+p.ID+"/settings",
		strings.NewReader(`{"settings":{"ebay_base_url":"`+ebayStub.URL+`","ebay_bearer_token":"test-token","ebay_marketplace":"EBAY_AU","integration.ebay.items_per_page":"37"}}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save ebay settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}

	createQuery := doRequest(t, a, http.MethodPost, "/api/scanner/query-sets", strings.NewReader(`{"name":"eBay Saved Search","keywords":["afx"],"provider_scope":["ebay"],"region":"AU","enabled":true}`), map[string]string{"Content-Type": "application/json"})
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

	run := doRequest(t, a, http.MethodPost, "/api/providers/ebay/run", strings.NewReader(`{"query_set_id":"`+querySetID+`"}`), map[string]string{"Content-Type": "application/json"})
	if run.Code != http.StatusOK {
		t.Fatalf("ebay run expected 200, got %d body=%s", run.Code, run.Body.String())
	}
	var runPayload struct {
		Provider   string           `json:"provider"`
		Candidates []map[string]any `json:"candidates"`
		Run        map[string]any   `json:"run"`
	}
	if err := json.NewDecoder(run.Body).Decode(&runPayload); err != nil {
		t.Fatalf("decode ebay run payload: %v", err)
	}
	if runPayload.Provider != "ebay" {
		t.Fatalf("expected provider ebay, got %+v", runPayload)
	}
	if len(runPayload.Candidates) != 1 {
		t.Fatalf("expected one persisted ebay candidate in run payload, got %+v", runPayload.Candidates)
	}
	if got, _ := runPayload.Candidates[0]["source"].(string); got != "ebay" {
		t.Fatalf("expected persisted ebay source, got %+v", runPayload.Candidates[0])
	}
	if got, _ := runPayload.Candidates[0]["stock_state"].(string); got != "low_stock" {
		t.Fatalf("expected low_stock persistence, got %+v", runPayload.Candidates[0])
	}
	if got, _ := runPayload.Candidates[0]["observed_currency"].(string); got != "AUD" {
		t.Fatalf("expected observed_currency AUD in ebay run payload, got %+v", runPayload.Candidates[0])
	}
	if got, ok := runPayload.Run["saved"].(float64); !ok || int(got) != 1 {
		t.Fatalf("expected run saved=1, got %+v", runPayload.Run)
	}
	if got, ok := runPayload.Run["items_per_page_requested"].(float64); !ok || int(got) != 37 {
		t.Fatalf("expected provider-run items_per_page_requested=37, got %+v", runPayload.Run)
	}
	if got, ok := runPayload.Run["items_per_page_effective"].(float64); !ok || int(got) != 37 {
		t.Fatalf("expected provider-run items_per_page_effective=37, got %+v", runPayload.Run)
	}

	reloaded := doRequest(t, a, http.MethodGet, "/api/scanner/query-sets", nil, nil)
	if reloaded.Code != http.StatusOK {
		t.Fatalf("reload query sets expected 200, got %d body=%s", reloaded.Code, reloaded.Body.String())
	}
	var querySetPayload struct {
		QuerySets []map[string]any `json:"query_sets"`
	}
	if err := json.NewDecoder(reloaded.Body).Decode(&querySetPayload); err != nil {
		t.Fatalf("decode query sets after ebay run: %v", err)
	}
	if len(querySetPayload.QuerySets) != 1 {
		t.Fatalf("expected one query set after ebay run, got %+v", querySetPayload.QuerySets)
	}
	latest := querySetPayload.QuerySets[0]
	if got, _ := latest["last_run_status"].(string); got != "succeeded" {
		t.Fatalf("expected ebay run to persist succeeded snapshot, got %+v", latest)
	}
	if got, ok := latest["last_candidate_count"].(float64); !ok || int(got) != 1 {
		t.Fatalf("expected ebay run to persist candidate count=1, got %+v", latest)
	}
}

func TestWave4EbayRunRejectsNonEbayScopedQuerySet(t *testing.T) {
	t.Parallel()

	ebayStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("eBay provider route must not call Browse for non-eBay scoped query sets")
	}))
	defer ebayStub.Close()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wave4EbayScopeGuard"}`), map[string]string{"Content-Type": "application/json"})
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
	saveSettings := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/profiles/"+p.ID+"/settings",
		strings.NewReader(`{"settings":{"ebay_base_url":"`+ebayStub.URL+`","ebay_bearer_token":"test-token","ebay_marketplace":"EBAY_AU"}}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save ebay settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}

	createQuery := doRequest(t, a, http.MethodPost, "/api/scanner/query-sets", strings.NewReader(`{"name":"Amazon Only Saved Search","keywords":["afx"],"provider_scope":["amazon"],"region":"AU","enabled":true}`), map[string]string{"Content-Type": "application/json"})
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

	run := doRequest(t, a, http.MethodPost, "/api/providers/ebay/run", strings.NewReader(`{"query_set_id":"`+querySetID+`"}`), map[string]string{"Content-Type": "application/json"})
	if run.Code != http.StatusBadRequest {
		t.Fatalf("eBay run for non-eBay scope expected 400, got %d body=%s", run.Code, run.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(run.Body).Decode(&payload); err != nil {
		t.Fatalf("decode scope guard payload: %v", err)
	}
	if got, _ := payload["error"].(string); got != "query_set_not_scoped_to_ebay" {
		t.Fatalf("expected query_set_not_scoped_to_ebay, got %+v", payload)
	}
	if got, _ := payload["provider"].(string); got != "ebay" {
		t.Fatalf("expected provider ebay in scope guard payload, got %+v", payload)
	}
	if got, _ := payload["query_set_id"].(string); got != querySetID {
		t.Fatalf("expected query_set_id %q in scope guard payload, got %+v", querySetID, payload)
	}
}

func TestWave4EbayRunRejectsInvalidConfiguredItemsPerPage(t *testing.T) {
	t.Parallel()

	ebayStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("eBay provider route must not call Browse when configured items_per_page is invalid")
	}))
	defer ebayStub.Close()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wave4EbayInvalidPageSize"}`), map[string]string{"Content-Type": "application/json"})
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
	saveSettings := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/profiles/"+p.ID+"/settings",
		strings.NewReader(`{"settings":{"ebay_base_url":"`+ebayStub.URL+`","ebay_bearer_token":"test-token","ebay_marketplace":"EBAY_AU","integration.ebay.items_per_page":"many"}}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save ebay settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}

	createQuery := doRequest(t, a, http.MethodPost, "/api/scanner/query-sets", strings.NewReader(`{"name":"eBay Invalid Page Size","keywords":["afx"],"provider_scope":["ebay"],"region":"AU","enabled":true}`), map[string]string{"Content-Type": "application/json"})
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

	run := doRequest(t, a, http.MethodPost, "/api/providers/ebay/run", strings.NewReader(`{"query_set_id":"`+querySetID+`"}`), map[string]string{"Content-Type": "application/json"})
	if run.Code != http.StatusBadRequest {
		t.Fatalf("eBay run with invalid items_per_page expected 400, got %d body=%s", run.Code, run.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(run.Body).Decode(&payload); err != nil {
		t.Fatalf("decode invalid page-size payload: %v", err)
	}
	if got, _ := payload["error"].(string); got != "invalid_ebay_items_per_page" {
		t.Fatalf("expected invalid_ebay_items_per_page, got %+v", payload)
	}
	if got, _ := payload["provider"].(string); got != "ebay" {
		t.Fatalf("expected provider ebay in invalid page-size payload, got %+v", payload)
	}
	if got, _ := payload["query_set_id"].(string); got != querySetID {
		t.Fatalf("expected query_set_id %q in invalid page-size payload, got %+v", querySetID, payload)
	}
	if got, _ := payload["setting"].(string); got != "integration.ebay.items_per_page" {
		t.Fatalf("expected setting integration.ebay.items_per_page, got %+v", payload)
	}
	if got, _ := payload["next_action"].(string); got != "update_ebay_items_per_page" {
		t.Fatalf("expected next_action update_ebay_items_per_page, got %+v", payload)
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
	for _, field := range []string{"listing_id", "title", "url", "seller", "first_seen", "last_seen", "stock_state", "observed_currency"} {
		if _, ok := first[field]; !ok {
			t.Fatalf("candidate missing %q: %+v", field, first)
		}
	}
}

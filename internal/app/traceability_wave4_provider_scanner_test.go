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
	if got, _ := payload["error_code"].(string); got != "query_set_not_scoped_to_ebay" {
		t.Fatalf("expected matching error_code query_set_not_scoped_to_ebay, got %+v", payload)
	}
	if got, _ := payload["provider"].(string); got != "ebay" {
		t.Fatalf("expected provider ebay in scope guard payload, got %+v", payload)
	}
	if got, _ := payload["query_set_id"].(string); got != querySetID {
		t.Fatalf("expected query_set_id %q in scope guard payload, got %+v", querySetID, payload)
	}
	if got, _ := payload["message"].(string); got == "" {
		t.Fatalf("expected actionable message in scope guard payload, got %+v", payload)
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
	if got, _ := payload["error_code"].(string); got != "invalid_ebay_items_per_page" {
		t.Fatalf("expected matching error_code invalid_ebay_items_per_page, got %+v", payload)
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
	if got, _ := payload["message"].(string); got == "" {
		t.Fatalf("expected actionable message in invalid page-size payload, got %+v", payload)
	}
}

func TestWave4EbayRunInvalidQuerySetReturnsActionableClientEnvelope(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wave4EbayInvalidQuerySetEnvelope"}`), map[string]string{"Content-Type": "application/json"})
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

	run := doRequest(t, a, http.MethodPost, "/api/providers/ebay/run", strings.NewReader(`{"query_set_id":"missing-ebay-query"}`), map[string]string{"Content-Type": "application/json"})
	if run.Code != http.StatusBadRequest {
		t.Fatalf("eBay run with invalid query set expected 400, got %d body=%s", run.Code, run.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(run.Body).Decode(&payload); err != nil {
		t.Fatalf("decode invalid query-set payload: %v", err)
	}
	if got, _ := payload["error"].(string); got != "invalid_query_set_id" {
		t.Fatalf("expected invalid_query_set_id, got %+v", payload)
	}
	if got, _ := payload["error_code"].(string); got != "invalid_query_set_id" {
		t.Fatalf("expected matching error_code invalid_query_set_id, got %+v", payload)
	}
	if got, _ := payload["provider"].(string); got != "ebay" {
		t.Fatalf("expected provider ebay in invalid query-set payload, got %+v", payload)
	}
	if got, _ := payload["query_set_id"].(string); got != "missing-ebay-query" {
		t.Fatalf("expected parsed query_set_id in invalid query-set payload, got %+v", payload)
	}
	if got, _ := payload["next_action"].(string); got != "select_existing_ebay_query_set" {
		t.Fatalf("expected next_action select_existing_ebay_query_set, got %+v", payload)
	}
	if got, _ := payload["message"].(string); got == "" {
		t.Fatalf("expected actionable message in invalid query-set payload, got %+v", payload)
	}
}

func TestWave4EbayRunMissingQuerySetReturnsActionableClientEnvelope(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wave4EbayMissingQuerySetEnvelope"}`), map[string]string{"Content-Type": "application/json"})
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

	run := doRequest(t, a, http.MethodPost, "/api/providers/ebay/run", strings.NewReader(`{"query_set_id":"   "}`), map[string]string{"Content-Type": "application/json"})
	if run.Code != http.StatusBadRequest {
		t.Fatalf("eBay run with missing query set expected 400, got %d body=%s", run.Code, run.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(run.Body).Decode(&payload); err != nil {
		t.Fatalf("decode missing query-set payload: %v", err)
	}
	if got, _ := payload["error"].(string); got != "missing_query_set_id" {
		t.Fatalf("expected missing_query_set_id, got %+v", payload)
	}
	if got, _ := payload["error_code"].(string); got != "missing_query_set_id" {
		t.Fatalf("expected matching error_code missing_query_set_id, got %+v", payload)
	}
	if got, _ := payload["provider"].(string); got != "ebay" {
		t.Fatalf("expected provider ebay in missing query-set payload, got %+v", payload)
	}
	if got, _ := payload["query_set_id"].(string); got != "" {
		t.Fatalf("expected trimmed empty query_set_id in missing query-set payload, got %+v", payload)
	}
	if got, _ := payload["next_action"].(string); got != "select_existing_ebay_query_set" {
		t.Fatalf("expected next_action select_existing_ebay_query_set, got %+v", payload)
	}
	if got, _ := payload["message"].(string); got == "" {
		t.Fatalf("expected actionable message in missing query-set payload, got %+v", payload)
	}
}

func TestWave4EbayRunBootstrapErrorsReturnActionableClientEnvelopes(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	invalidJSON := doRequest(t, a, http.MethodPost, "/api/providers/ebay/run", strings.NewReader(`{"query_set_id":`), map[string]string{"Content-Type": "application/json"})
	if invalidJSON.Code != http.StatusBadRequest {
		t.Fatalf("eBay run invalid JSON expected 400, got %d body=%s", invalidJSON.Code, invalidJSON.Body.String())
	}
	var invalidPayload map[string]any
	if err := json.NewDecoder(invalidJSON.Body).Decode(&invalidPayload); err != nil {
		t.Fatalf("decode invalid JSON payload: %v", err)
	}
	if got, _ := invalidPayload["error"].(string); got != "invalid_json" {
		t.Fatalf("expected invalid_json, got %+v", invalidPayload)
	}
	if got, _ := invalidPayload["error_code"].(string); got != "invalid_json" {
		t.Fatalf("expected matching error_code invalid_json, got %+v", invalidPayload)
	}
	if got, _ := invalidPayload["provider"].(string); got != "ebay" {
		t.Fatalf("expected provider ebay in invalid JSON payload, got %+v", invalidPayload)
	}
	if got, _ := invalidPayload["query_set_id"].(string); got != "" {
		t.Fatalf("expected empty query_set_id in invalid JSON payload, got %+v", invalidPayload)
	}
	if got, _ := invalidPayload["next_action"].(string); got != "select_existing_ebay_query_set" {
		t.Fatalf("expected next_action select_existing_ebay_query_set, got %+v", invalidPayload)
	}
	if got, _ := invalidPayload["message"].(string); got == "" {
		t.Fatalf("expected actionable message in invalid JSON payload, got %+v", invalidPayload)
	}

	trailingJSON := doRequest(t, a, http.MethodPost, "/api/providers/ebay/run", strings.NewReader(`{"query_set_id":"ebay-q-1"}{"query_set_id":"shadow"}`), map[string]string{"Content-Type": "application/json"})
	if trailingJSON.Code != http.StatusBadRequest {
		t.Fatalf("eBay run with trailing JSON expected 400, got %d body=%s", trailingJSON.Code, trailingJSON.Body.String())
	}
	var trailingPayload map[string]any
	if err := json.NewDecoder(trailingJSON.Body).Decode(&trailingPayload); err != nil {
		t.Fatalf("decode trailing JSON payload: %v", err)
	}
	if got, _ := trailingPayload["error"].(string); got != "invalid_json" {
		t.Fatalf("expected invalid_json for trailing JSON, got %+v", trailingPayload)
	}
	if got, _ := trailingPayload["error_code"].(string); got != "invalid_json" {
		t.Fatalf("expected matching error_code invalid_json for trailing JSON, got %+v", trailingPayload)
	}
	if got, _ := trailingPayload["provider"].(string); got != "ebay" {
		t.Fatalf("expected provider ebay in trailing JSON payload, got %+v", trailingPayload)
	}
	if got, _ := trailingPayload["query_set_id"].(string); got != "" {
		t.Fatalf("expected empty query_set_id in trailing JSON payload, got %+v", trailingPayload)
	}
	if got, _ := trailingPayload["next_action"].(string); got != "select_existing_ebay_query_set" {
		t.Fatalf("expected next_action select_existing_ebay_query_set, got %+v", trailingPayload)
	}

	missingProfile := doRequest(t, a, http.MethodPost, "/api/providers/ebay/run", strings.NewReader(`{"query_set_id":"ebay-q-1"}`), map[string]string{"Content-Type": "application/json"})
	if missingProfile.Code != http.StatusBadRequest {
		t.Fatalf("eBay run without active profile expected 400, got %d body=%s", missingProfile.Code, missingProfile.Body.String())
	}
	var profilePayload map[string]any
	if err := json.NewDecoder(missingProfile.Body).Decode(&profilePayload); err != nil {
		t.Fatalf("decode missing profile payload: %v", err)
	}
	if got, _ := profilePayload["error"].(string); got != "active_profile_not_set" {
		t.Fatalf("expected active_profile_not_set, got %+v", profilePayload)
	}
	if got, _ := profilePayload["error_code"].(string); got != "active_profile_not_set" {
		t.Fatalf("expected matching error_code active_profile_not_set, got %+v", profilePayload)
	}
	if got, _ := profilePayload["provider"].(string); got != "ebay" {
		t.Fatalf("expected provider ebay in missing profile payload, got %+v", profilePayload)
	}
	if got, _ := profilePayload["query_set_id"].(string); got != "ebay-q-1" {
		t.Fatalf("expected parsed query_set_id in missing profile payload, got %+v", profilePayload)
	}
	if got, _ := profilePayload["next_action"].(string); got != "select_active_profile" {
		t.Fatalf("expected next_action select_active_profile, got %+v", profilePayload)
	}
	if got, _ := profilePayload["message"].(string); got == "" {
		t.Fatalf("expected actionable message in missing profile payload, got %+v", profilePayload)
	}
}

func TestWave4EbayRunRejectsUnsupportedMethodsWithAllowHeader(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	run := doRequest(t, a, http.MethodGet, "/api/providers/ebay/run", nil, nil)
	if run.Code != http.StatusMethodNotAllowed {
		t.Fatalf("eBay run unsupported method expected 405, got %d body=%s", run.Code, run.Body.String())
	}
	if got := run.Header().Get("Allow"); got != http.MethodPost {
		t.Fatalf("expected Allow header POST for eBay provider run method error, got %q", got)
	}
	var payload map[string]any
	if err := json.NewDecoder(run.Body).Decode(&payload); err != nil {
		t.Fatalf("decode method error payload: %v", err)
	}
	if got, _ := payload["error"].(string); got != "method_not_allowed" {
		t.Fatalf("expected method_not_allowed payload, got %+v", payload)
	}
	if got, _ := payload["error_code"].(string); got != "method_not_allowed" {
		t.Fatalf("expected matching error_code method_not_allowed, got %+v", payload)
	}
	if got, _ := payload["provider"].(string); got != "ebay" {
		t.Fatalf("expected provider ebay in method error payload, got %+v", payload)
	}
	if got, _ := payload["allowed_method"].(string); got != http.MethodPost {
		t.Fatalf("expected allowed_method POST in method error payload, got %+v", payload)
	}
	if got, _ := payload["next_action"].(string); got != "retry_with_post" {
		t.Fatalf("expected next_action retry_with_post, got %+v", payload)
	}
	if got, _ := payload["message"].(string); got == "" {
		t.Fatalf("expected actionable message in method error payload, got %+v", payload)
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
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|100|0","title":"AFX Mega G+","price":{"value":"45.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/100","seller":{"username":"seller1"}}]}`))
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
	if querySetID == "" {
		t.Fatal("expected query set id")
	}

	createAmazonQuery := doRequest(t, a, http.MethodPost, "/api/scanner/query-sets", strings.NewReader(`{"name":"Amazon scheduled skip","keywords":["afx amazon"],"provider_scope":["amazon"],"schedule_cron":"*/15 * * * *","enabled":true,"rate_limit_rps":2}`), map[string]string{"Content-Type": "application/json"})
	if createAmazonQuery.Code != http.StatusCreated {
		t.Fatalf("create amazon-scoped query set status=%d body=%s", createAmazonQuery.Code, createAmazonQuery.Body.String())
	}
	var amazonQS map[string]any
	if err := json.NewDecoder(createAmazonQuery.Body).Decode(&amazonQS); err != nil {
		t.Fatalf("decode amazon query set: %v", err)
	}
	amazonQuerySetID, _ := amazonQS["id"].(string)
	if amazonQuerySetID == "" {
		t.Fatal("expected amazon query set id")
	}

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
	if got, ok := summary["query_sets_executed"].(float64); !ok || int(got) != 1 {
		t.Fatalf("expected scheduled route to execute only the eBay-scoped query, got %+v", summary)
	}
	if got, ok := summary["candidates_collected"].(float64); !ok || int(got) != 1 {
		t.Fatalf("expected one eBay candidate from scoped scheduled run, got %+v", summary)
	}
	if got, ok := summary["failures"].(float64); !ok || int(got) != 0 {
		t.Fatalf("expected skipped non-eBay scheduled query not to count as failure, got %+v", summary)
	}

	var amazonRuns int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM scanner_runs WHERE query_set_id = ?`, amazonQuerySetID).Scan(&amazonRuns); err != nil {
		t.Fatalf("query amazon scheduled run count: %v", err)
	}
	if amazonRuns != 0 {
		t.Fatalf("expected amazon-scoped scheduled query to be skipped by eBay scheduled route, got %d runs", amazonRuns)
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

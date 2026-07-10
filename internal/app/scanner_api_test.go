package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestScannerQuerySetsAndProviderHealthEndpoints(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/scanner/query-sets", strings.NewReader(`{"name":"Q1","keywords":["afx"],"exclusions":["broken"],"max_price":100,"region":"US","condition":"used","schedule_cron":"*/15 * * * *","enabled":true}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", create.Code, create.Body.String())
	}
	list := doRequest(t, a, http.MethodGet, "/api/scanner/query-sets", nil, nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list query sets status=%d body=%s", list.Code, list.Body.String())
	}
	var payload struct {
		QuerySets []map[string]any `json:"query_sets"`
	}
	if err := json.NewDecoder(list.Body).Decode(&payload); err != nil {
		t.Fatalf("decode list payload: %v", err)
	}
	if len(payload.QuerySets) != 1 {
		t.Fatalf("expected one query set, got %d", len(payload.QuerySets))
	}
	if got, _ := payload.QuerySets[0]["next_run_at"].(string); strings.TrimSpace(got) == "" {
		t.Fatalf("expected scheduled query set to expose computed next_run_at, got %+v", payload.QuerySets[0])
	}
	health := doRequest(t, a, http.MethodGet, "/api/provider/health?provider=ebay", nil, nil)
	if health.Code != http.StatusOK {
		t.Fatalf("provider health status=%d body=%s", health.Code, health.Body.String())
	}
	var healthPayload map[string]any
	if err := json.NewDecoder(health.Body).Decode(&healthPayload); err != nil {
		t.Fatalf("decode provider health payload: %v", err)
	}
	for key, expected := range map[string]any{
		"provider": "ebay",
		"status":   "unknown",
		"state":    "disabled",
		"category": "not_checked",
		"label":    "Not checked yet",
	} {
		if got := healthPayload[key]; got != expected {
			t.Fatalf("provider health %s=%v, want %v in %+v", key, got, expected, healthPayload)
		}
	}
	if _, ok := healthPayload["last_error"]; !ok {
		t.Fatalf("provider health missing last_error alias: %+v", healthPayload)
	}
	if _, ok := healthPayload["retry_after_seconds"]; !ok {
		t.Fatalf("provider health missing retry_after_seconds: %+v", healthPayload)
	}
}

func TestProviderHealthEndpointExposesMarketWatchTaxonomy(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		status     string
		message    string
		retryAfter int
		category   string
		label      string
	}{
		{name: "healthy", status: "ok", message: "Provider ready", category: "healthy", label: "Connected / healthy"},
		{name: "setup", status: "auth_required", message: "missing token", category: "needs_setup", label: "Needs setup"},
		{name: "reauth", status: "error", message: "Provider credentials expired", category: "needs_reauthentication", label: "Needs re-authentication"},
		{name: "rate", status: "error", message: "rate limit reached", retryAfter: 120, category: "rate_limited", label: "Rate limited"},
		{name: "unavailable", status: "error", message: "upstream unavailable", category: "provider_unavailable", label: "Provider unavailable"},
		{name: "partial", status: "partial_failure", message: "Amazon succeeded; eBay partial failure", category: "partially_failed", label: "Partially failed"},
		{name: "failed", status: "error", message: "Browse failed", category: "failed", label: "Failed"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := newTestApp(t)
			if _, err := a.db.Exec(`INSERT INTO provider_health(provider, status, message, retry_after_seconds, updated_at) VALUES ('ebay', ?, ?, ?, CURRENT_TIMESTAMP)`, tc.status, tc.message, tc.retryAfter); err != nil {
				t.Fatalf("seed provider health: %v", err)
			}

			resp := doRequest(t, a, http.MethodGet, "/api/provider/health?provider=ebay", nil, nil)
			if resp.Code != http.StatusOK {
				t.Fatalf("provider health status=%d body=%s", resp.Code, resp.Body.String())
			}
			var payload map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
				t.Fatalf("decode provider health payload: %v", err)
			}
			if payload["category"] != tc.category || payload["label"] != tc.label {
				t.Fatalf("unexpected health taxonomy for %s: got category=%v label=%v payload=%+v", tc.name, payload["category"], payload["label"], payload)
			}
			if got, _ := payload["guidance"].(string); strings.TrimSpace(got) == "" {
				t.Fatalf("expected taxonomy guidance for %s, got %+v", tc.name, payload)
			}
		})
	}
}

func TestProviderHealthEndpointKeepsUnrelatedProvidersIsolated(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	if _, err := a.db.Exec(`INSERT INTO provider_health(provider, status, message, retry_after_seconds, updated_at) VALUES ('bonzaslotcars', 'error', 'temporary provider failure', 0, CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("seed non-ebay provider health: %v", err)
	}

	bonza := doRequest(t, a, http.MethodGet, "/api/provider/health?provider=bonzaslotcars", nil, nil)
	if bonza.Code != http.StatusOK {
		t.Fatalf("bonza provider health status=%d body=%s", bonza.Code, bonza.Body.String())
	}
	var bonzaPayload map[string]any
	if err := json.NewDecoder(bonza.Body).Decode(&bonzaPayload); err != nil {
		t.Fatalf("decode bonza provider health payload: %v", err)
	}
	for key, expected := range map[string]any{
		"provider":    "bonzaslotcars",
		"status":      "error",
		"state":       "degraded",
		"category":    "failed",
		"label":       "Failed",
		"last_error":  "temporary provider failure",
		"next_action": "review_provider_status",
	} {
		if got := bonzaPayload[key]; got != expected {
			t.Fatalf("bonza provider health %s=%v, want %v in %+v", key, got, expected, bonzaPayload)
		}
	}

	ebay := doRequest(t, a, http.MethodGet, "/api/provider/health?provider=ebay", nil, nil)
	if ebay.Code != http.StatusOK {
		t.Fatalf("ebay provider health status=%d body=%s", ebay.Code, ebay.Body.String())
	}
	var ebayPayload map[string]any
	if err := json.NewDecoder(ebay.Body).Decode(&ebayPayload); err != nil {
		t.Fatalf("decode ebay provider health payload: %v", err)
	}
	for key, expected := range map[string]any{
		"provider": "ebay",
		"status":   "unknown",
		"state":    "disabled",
		"category": "not_checked",
		"label":    "Not checked yet",
	} {
		if got := ebayPayload[key]; got != expected {
			t.Fatalf("ebay provider health %s=%v, want %v in %+v", key, got, expected, ebayPayload)
		}
	}
	if got := ebayPayload["last_error"]; got != nil {
		t.Fatalf("unrelated non-ebay failure must not poison ebay last_error, got %+v", ebayPayload)
	}
}

func TestScannerRetryFailuresEndpoint(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/scanner/query-sets", strings.NewReader(`{"name":"Q1","keywords":["afx"],"enabled":true}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", create.Code, create.Body.String())
	}
	var created map[string]any
	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatalf("decode create payload: %v", err)
	}
	querySetID, _ := created["id"].(string)
	if querySetID == "" {
		t.Fatal("expected query set id")
	}

	retry := doRequest(t, a, http.MethodPost, "/api/scanner/failures/retry", strings.NewReader(`{"query_set_id":"`+querySetID+`"}`), map[string]string{"Content-Type": "application/json"})
	if retry.Code != http.StatusOK {
		t.Fatalf("retry failures status=%d body=%s", retry.Code, retry.Body.String())
	}
}

func TestScannerCandidatesResultInboxFiltersPaginationAndLifecycleAPI(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Result Inbox API"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	activate := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if activate.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", activate.Code, activate.Body.String())
	}
	if _, err := a.db.Exec(`INSERT INTO scanner_query_sets(id, profile_id, name, keywords_json, exclusions_json, provider_scope_json) VALUES ('result-inbox-api-q', ?, 'Result Inbox API Watch', '["afx"]', '[]', '["ebay","bonzaslotcars"]')`, p.ID); err != nil {
		t.Fatalf("seed query set: %v", err)
	}
	for _, seed := range []struct {
		id        string
		listingID string
		source    string
		status    string
	}{
		{"result-inbox-api-c1", "API-EBAY-1", "ebay", "new"},
		{"result-inbox-api-c2", "API-EBAY-2", "ebay", "seen"},
		{"result-inbox-api-c3", "API-BONZA-1", "bonzaslotcars", "new"},
	} {
		if _, err := a.db.Exec(`INSERT INTO scanner_candidates(id, profile_id, query_set_id, listing_id, title, price, observed_currency, shipping, url, image, seller, first_seen, last_seen, status, source, stock_state, stock_count) VALUES (?, ?, 'result-inbox-api-q', ?, ?, 10, 'AUD', 2, ?, '', 'seller', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, 'in_stock', 1)`, seed.id, p.ID, seed.listingID, seed.listingID+" title", "https://market.test/"+seed.listingID, seed.status, seed.source); err != nil {
			t.Fatalf("seed candidate %s: %v", seed.id, err)
		}
	}

	list := doRequest(t, a, http.MethodGet, "/api/scanner/candidates?query_set_id=result-inbox-api-q&provider=ebay&page=1&page_size=1", nil, nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list candidates status=%d body=%s", list.Code, list.Body.String())
	}
	var payload struct {
		Candidates []map[string]any `json:"candidates"`
		Total      int              `json:"total"`
		Page       int              `json:"page"`
		PageSize   int              `json:"page_size"`
	}
	if err := json.NewDecoder(list.Body).Decode(&payload); err != nil {
		t.Fatalf("decode candidates payload: %v", err)
	}
	if payload.Total != 2 || payload.Page != 1 || payload.PageSize != 1 || len(payload.Candidates) != 1 {
		t.Fatalf("expected provider-filtered paginated payload, got %+v", payload)
	}

	patch := doRequest(t, a, http.MethodPatch, "/api/scanner/candidates/result-inbox-api-c1", strings.NewReader(`{"status":"dismissed"}`), map[string]string{"Content-Type": "application/json"})
	if patch.Code != http.StatusOK {
		t.Fatalf("patch candidate status=%d body=%s", patch.Code, patch.Body.String())
	}
	var updated map[string]any
	if err := json.NewDecoder(patch.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated candidate: %v", err)
	}
	if got, _ := updated["status"].(string); got != "dismissed" {
		t.Fatalf("expected dismissed updated candidate, got %+v", updated)
	}
	patchHistory, ok := updated["decision_history"].([]any)
	if !ok || len(patchHistory) != 1 {
		t.Fatalf("expected patch response to include one decision history record, got %+v", updated["decision_history"])
	}
	firstHistory, ok := patchHistory[0].(map[string]any)
	if !ok {
		t.Fatalf("expected decision history object, got %+v", patchHistory[0])
	}
	if firstHistory["from_status"] != "new" || firstHistory["to_status"] != "dismissed" {
		t.Fatalf("expected decision history new -> dismissed, got %+v", firstHistory)
	}

	filtered := doRequest(t, a, http.MethodGet, "/api/scanner/candidates?query_set_id=result-inbox-api-q&status=dismissed&provider=ebay", nil, nil)
	if filtered.Code != http.StatusOK {
		t.Fatalf("filtered candidates status=%d body=%s", filtered.Code, filtered.Body.String())
	}
	payload = struct {
		Candidates []map[string]any `json:"candidates"`
		Total      int              `json:"total"`
		Page       int              `json:"page"`
		PageSize   int              `json:"page_size"`
	}{}
	if err := json.NewDecoder(filtered.Body).Decode(&payload); err != nil {
		t.Fatalf("decode filtered payload: %v", err)
	}
	if payload.Total != 1 || len(payload.Candidates) != 1 || payload.Candidates[0]["id"] != "result-inbox-api-c1" {
		t.Fatalf("expected dismissed provider-filtered candidate after patch, got %+v", payload)
	}
	listHistory, ok := payload.Candidates[0]["decision_history"].([]any)
	if !ok || len(listHistory) != 1 {
		t.Fatalf("expected list response to include decision history, got %+v", payload.Candidates[0])
	}
}

func TestScannerRunsEndpointListsPersistedRunHistory(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Run History API"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	activate := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if activate.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", activate.Code, activate.Body.String())
	}
	if _, err := a.db.Exec(`INSERT INTO scanner_query_sets(id, profile_id, name, keywords_json, exclusions_json, provider_scope_json) VALUES ('run-history-q', ?, 'Run History Watch', '["afx"]', '[]', '["ebay"]')`, p.ID); err != nil {
		t.Fatalf("seed query set: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO scanner_runs(id, profile_id, query_set_id, provider, trigger_type, started_at, finished_at, status, result_count, new_result_count, error_category, error_message, retry_guidance) VALUES ('run-history-1', ?, 'run-history-q', 'ebay', 'manual', '2026-06-29T00:01:00Z', '2026-06-29T00:02:00Z', 'failed', 0, 0, 'auth', 'Provider credentials expired', 'Reconnect eBay before retrying.')`, p.ID); err != nil {
		t.Fatalf("seed scanner run: %v", err)
	}

	resp := doRequest(t, a, http.MethodGet, "/api/scanner/runs?query_set_id=run-history-q", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("scanner runs status=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload struct {
		Runs []map[string]any `json:"runs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode scanner runs payload: %v", err)
	}
	if len(payload.Runs) != 1 {
		t.Fatalf("expected one run history record, got %+v", payload)
	}
	run := payload.Runs[0]
	for key, expected := range map[string]any{
		"id":               "run-history-1",
		"provider":         "ebay",
		"trigger_type":     "manual",
		"status":           "failed",
		"result_count":     float64(0),
		"new_result_count": float64(0),
		"error_message":    "Provider credentials expired",
		"retry_guidance":   "Reconnect eBay before retrying.",
		"next_action":      "check_provider_health_and_credentials",
	} {
		if got := run[key]; got != expected {
			t.Fatalf("expected run %s=%v, got %v in %+v", key, expected, got, run)
		}
	}
}

func TestScannerFailuresRejectsUnsupportedMethodsWithGuidance(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodPost, "/api/scanner/failures", strings.NewReader(`{}`), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected method not allowed, got %d body=%s", resp.Code, resp.Body.String())
	}
	if got := resp.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("expected Allow GET, got %q", got)
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode method error payload: %v body=%s", err, resp.Body.String())
	}
	for key, expected := range map[string]any{
		"error":          "method_not_allowed",
		"error_code":     "method_not_allowed",
		"provider":       "ebay",
		"message":        "Use GET to list scanner failure snapshots before retrying eBay saved-search failures.",
		"next_action":    "retry_with_get",
		"allowed_method": http.MethodGet,
	} {
		if got := payload[key]; got != expected {
			t.Fatalf("expected %s=%v, got %v in %+v", key, expected, got, payload)
		}
	}
}

func TestScannerQuerySetUpdateAndDeleteEndpoints(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/scanner/query-sets", strings.NewReader(`{"name":"Q1","keywords":["afx"],"enabled":true}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", create.Code, create.Body.String())
	}
	var created map[string]any
	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatalf("decode create payload: %v", err)
	}
	querySetID, _ := created["id"].(string)
	if querySetID == "" {
		t.Fatal("expected query set id")
	}

	update := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/scanner/query-sets/"+querySetID,
		strings.NewReader(`{"name":"Q1 Updated","keywords":["afx","camaro"],"enabled":false}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if update.Code != http.StatusOK {
		t.Fatalf("update query set status=%d body=%s", update.Code, update.Body.String())
	}
	var updated map[string]any
	if err := json.NewDecoder(update.Body).Decode(&updated); err != nil {
		t.Fatalf("decode update payload: %v", err)
	}
	if got, _ := updated["name"].(string); got != "Q1 Updated" {
		t.Fatalf("expected updated name, got %q", got)
	}

	remove := doRequest(t, a, http.MethodDelete, "/api/scanner/query-sets/"+querySetID, nil, nil)
	if remove.Code != http.StatusNoContent {
		t.Fatalf("delete query set status=%d body=%s", remove.Code, remove.Body.String())
	}

	list := doRequest(t, a, http.MethodGet, "/api/scanner/query-sets", nil, nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list query sets status=%d body=%s", list.Code, list.Body.String())
	}
	var payload struct {
		QuerySets []map[string]any `json:"query_sets"`
	}
	if err := json.NewDecoder(list.Body).Decode(&payload); err != nil {
		t.Fatalf("decode list payload: %v", err)
	}
	if len(payload.QuerySets) != 0 {
		t.Fatalf("expected no query sets after delete, got %d", len(payload.QuerySets))
	}
}

func TestScannerRunItemsPerPageSummaryAppliesSafeCap(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profile := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/profiles",
		strings.NewReader(`{"name":"items-per-page-profile"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if profile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", profile.Code, profile.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(profile.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile payload: %v", err)
	}
	activate := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/profiles/active",
		strings.NewReader(`{"profile_id":"`+p.ID+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if activate.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", activate.Code, activate.Body.String())
	}
	ebayStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET to ebay stub, got %s", r.Method)
		}
		if r.URL.Path != "/buy/browse/v1/item_summary/search" {
			t.Fatalf("unexpected ebay stub path: %s", r.URL.Path)
		}
		if got := strings.TrimSpace(r.Header.Get("Authorization")); got != "Bearer test-token" {
			t.Fatalf("unexpected auth header: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"I-1","title":"AFX Camaro","price":{"value":"49.95","currency":"AUD"},"itemWebUrl":"https://example/item/1","image":{"imageUrl":"https://example/image/1.jpg"},"seller":{"username":"seller-a"}}]}`))
	}))
	defer ebayStub.Close()

	saveSettings := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/profiles/"+p.ID+"/settings",
		strings.NewReader(`{"settings":{"ebay_base_url":"`+ebayStub.URL+`","ebay_bearer_token":"test-token","ebay_marketplace":"EBAY_AU","integration.bonzaslotcars.items_per_page":"200"}}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save ebay settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}

	create := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/scanner/query-sets",
		strings.NewReader(`{"name":"Bonza AFX","keywords":["afx"],"provider_scope":["bonzaslotcars"],"enabled":true}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if create.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", create.Code, create.Body.String())
	}
	var created map[string]any
	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatalf("decode create payload: %v", err)
	}
	querySetID, _ := created["id"].(string)
	if querySetID == "" {
		t.Fatal("expected query set id")
	}

	run := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/scanner/run",
		strings.NewReader(`{"query_set_id":"`+querySetID+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if run.Code != http.StatusOK {
		t.Fatalf("run query set status=%d body=%s", run.Code, run.Body.String())
	}
	var summary map[string]any
	if err := json.NewDecoder(run.Body).Decode(&summary); err != nil {
		t.Fatalf("decode run payload: %v", err)
	}

	if got, ok := summary["items_per_page_requested"].(float64); !ok || int(got) != 200 {
		t.Fatalf("expected items_per_page_requested=200, got %#v", summary["items_per_page_requested"])
	}
	if got, ok := summary["items_per_page_effective"].(float64); !ok || int(got) != 36 {
		t.Fatalf("expected items_per_page_effective=36, got %#v", summary["items_per_page_effective"])
	}
	if got, ok := summary["observed_page_size"].(float64); !ok || int(got) != 36 {
		t.Fatalf("expected observed_page_size=36, got %#v", summary["observed_page_size"])
	}
	if got, ok := summary["page_count"].(float64); !ok || int(got) != 1 {
		t.Fatalf("expected page_count=1, got %#v", summary["page_count"])
	}
	warning, _ := summary["items_per_page_warning"].(string)
	if warning == "" {
		t.Fatalf("expected items_per_page_warning in run summary, got %#v", summary["items_per_page_warning"])
	}

	list := doRequest(t, a, http.MethodGet, "/api/scanner/query-sets", nil, nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list query sets after run status=%d body=%s", list.Code, list.Body.String())
	}
	var querySetPayload map[string][]map[string]any
	if err := json.NewDecoder(list.Body).Decode(&querySetPayload); err != nil {
		t.Fatalf("decode query set list after run: %v", err)
	}
	if len(querySetPayload["query_sets"]) != 1 {
		t.Fatalf("expected one query set after run, got %d", len(querySetPayload["query_sets"]))
	}
	listed := querySetPayload["query_sets"][0]
	if got, _ := listed["last_run_status"].(string); got != "succeeded" {
		t.Fatalf("expected persisted last_run_status=succeeded, got %#v", listed["last_run_status"])
	}
	if got, ok := listed["last_candidate_count"].(float64); !ok || int(got) != 1 {
		t.Fatalf("expected persisted last_candidate_count=1, got %#v", listed["last_candidate_count"])
	}
	if got, _ := listed["last_run_at"].(string); strings.TrimSpace(got) == "" {
		t.Fatalf("expected persisted last_run_at, got %#v", listed["last_run_at"])
	}
}

func TestScannerRunMapsEbayAuthFailureToProviderErrorCode(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profile := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/profiles",
		strings.NewReader(`{"name":"ebay-auth-profile"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if profile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", profile.Code, profile.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(profile.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile payload: %v", err)
	}
	activate := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/profiles/active",
		strings.NewReader(`{"profile_id":"`+p.ID+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if activate.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", activate.Code, activate.Body.String())
	}
	ebayStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errors":[{"message":"invalid token"}]}`))
	}))
	defer ebayStub.Close()

	saveSettings := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/profiles/"+p.ID+"/settings",
		strings.NewReader(`{"settings":{"ebay_base_url":"`+ebayStub.URL+`","ebay_bearer_token":"expired-token","ebay_marketplace":"EBAY_AU"}}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save ebay settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}

	create := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/scanner/query-sets",
		strings.NewReader(`{"name":"eBay Auth Check","keywords":["pokemon"],"provider_scope":["ebay"],"enabled":true}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if create.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", create.Code, create.Body.String())
	}
	var created map[string]any
	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatalf("decode create payload: %v", err)
	}
	querySetID, _ := created["id"].(string)

	run := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/scanner/run",
		strings.NewReader(`{"query_set_id":"`+querySetID+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if run.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 auth failure, got %d body=%s", run.Code, run.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(run.Body).Decode(&payload); err != nil {
		t.Fatalf("decode run error payload: %v", err)
	}
	if payload["error_code"] != "PROVIDER_AUTH_INVALID" || payload["provider"] != "ebay" {
		t.Fatalf("unexpected provider auth payload: %+v", payload)
	}
}

func TestEbayProviderRunPreservesForbiddenAuthStatus(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profile := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/profiles",
		strings.NewReader(`{"name":"ebay-forbidden-auth-profile"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if profile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", profile.Code, profile.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(profile.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile payload: %v", err)
	}
	activate := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/profiles/active",
		strings.NewReader(`{"profile_id":"`+p.ID+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if activate.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", activate.Code, activate.Body.String())
	}
	ebayStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"errors":[{"errorId":1100,"domain":"ACCESS","category":"REQUEST","message":"scope denied","longMessage":"Browse scope missing"}]}`))
	}))
	defer ebayStub.Close()

	saveSettings := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/profiles/"+p.ID+"/settings",
		strings.NewReader(`{"settings":{"ebay_base_url":"`+ebayStub.URL+`","ebay_bearer_token":"scope-denied-token","ebay_marketplace":"EBAY_AU"}}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save ebay settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}

	create := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/scanner/query-sets",
		strings.NewReader(`{"name":"eBay Forbidden Auth Check","keywords":["pokemon"],"provider_scope":["ebay"],"enabled":true}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if create.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", create.Code, create.Body.String())
	}
	var created map[string]any
	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatalf("decode create payload: %v", err)
	}
	querySetID, _ := created["id"].(string)

	run := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/providers/ebay/run",
		strings.NewReader(`{"query_set_id":"`+querySetID+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if run.Code != http.StatusForbidden {
		t.Fatalf("expected 403 auth failure, got %d body=%s", run.Code, run.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(run.Body).Decode(&payload); err != nil {
		t.Fatalf("decode run error payload: %v", err)
	}
	if payload["error_code"] != "PROVIDER_AUTH_INVALID" || payload["provider"] != "ebay" {
		t.Fatalf("unexpected provider auth payload: %+v", payload)
	}
	if payload["next_action"] != "review_provider_credentials_and_health" {
		t.Fatalf("expected credential review next action, got %+v", payload)
	}
	message, _ := payload["message"].(string)
	for _, want := range []string{"1100", "ACCESS", "REQUEST", "scope denied", "Browse scope missing"} {
		if !strings.Contains(message, want) {
			t.Fatalf("expected provider auth message to preserve %q, got %q", want, message)
		}
	}
}

func TestEbayProviderRunMapsBrowseFailureToProviderHealthGuidance(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profile := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/profiles",
		strings.NewReader(`{"name":"ebay-search-failure-profile"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if profile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", profile.Code, profile.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(profile.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile payload: %v", err)
	}
	activate := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/profiles/active",
		strings.NewReader(`{"profile_id":"`+p.ID+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if activate.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", activate.Code, activate.Body.String())
	}
	ebayStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", "120")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"errors":[{"errorId":12001,"domain":"API_BROWSE","category":"REQUEST","message":"rate limit reached","longMessage":"Retry after the provider window resets"}]}`))
	}))
	defer ebayStub.Close()

	saveSettings := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/profiles/"+p.ID+"/settings",
		strings.NewReader(`{"settings":{"ebay_base_url":"`+ebayStub.URL+`","ebay_bearer_token":"valid-token","ebay_marketplace":"EBAY_AU"}}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save ebay settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}

	create := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/scanner/query-sets",
		strings.NewReader(`{"name":"eBay Browse Failure","keywords":["afx"],"provider_scope":["ebay"],"enabled":true}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if create.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", create.Code, create.Body.String())
	}
	var created map[string]any
	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatalf("decode create payload: %v", err)
	}
	querySetID, _ := created["id"].(string)

	run := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/providers/ebay/run",
		strings.NewReader(`{"query_set_id":"`+querySetID+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if run.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 search failure, got %d body=%s", run.Code, run.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(run.Body).Decode(&payload); err != nil {
		t.Fatalf("decode ebay run error payload: %v", err)
	}
	if payload["error_code"] != "PROVIDER_SEARCH_FAILED" || payload["provider"] != "ebay" {
		t.Fatalf("unexpected provider search payload: %+v", payload)
	}
	if payload["next_action"] != "check_provider_health_and_credentials" {
		t.Fatalf("expected provider-health next action, got %+v", payload)
	}
	if got, ok := payload["retry_after_seconds"].(float64); !ok || int(got) != 120 {
		t.Fatalf("expected retry_after_seconds=120, got %+v", payload)
	}
	message, _ := payload["message"].(string)
	for _, want := range []string{"12001", "API_BROWSE", "REQUEST", "rate limit reached", "Retry after the provider window resets"} {
		if !strings.Contains(message, want) {
			t.Fatalf("expected message to preserve %q, got %q", want, message)
		}
	}

	health := doRequest(t, a, http.MethodGet, "/api/provider/health?provider=ebay", nil, nil)
	if health.Code != http.StatusOK {
		t.Fatalf("provider health status=%d body=%s", health.Code, health.Body.String())
	}
	var healthPayload map[string]any
	if err := json.NewDecoder(health.Body).Decode(&healthPayload); err != nil {
		t.Fatalf("decode provider health payload: %v", err)
	}
	if healthPayload["status"] != "rate_limited" || healthPayload["category"] != "rate_limited" || healthPayload["label"] != "Rate limited" {
		t.Fatalf("expected persisted rate-limited provider health taxonomy, got %+v", healthPayload)
	}
	if healthPayload["state"] != "degraded" || healthPayload["last_error"] == nil {
		t.Fatalf("expected degraded provider health with last_error, got %+v", healthPayload)
	}
	if healthPayload["next_action"] != "check_provider_health_and_credentials" {
		t.Fatalf("expected degraded provider health next action, got %+v", healthPayload)
	}
	if got, ok := healthPayload["retry_after_seconds"].(float64); !ok || int(got) != 120 {
		t.Fatalf("expected provider health retry_after_seconds=120, got %+v", healthPayload)
	}

	inbox := doRequest(t, a, http.MethodGet, "/api/chat/inbox?profile_id="+p.ID, nil, nil)
	if inbox.Code != http.StatusOK {
		t.Fatalf("inbox status=%d body=%s", inbox.Code, inbox.Body.String())
	}
	var inboxPayload struct {
		Items []struct {
			Source   string         `json:"source"`
			Status   string         `json:"status"`
			Title    string         `json:"title"`
			Summary  string         `json:"summary"`
			Metadata map[string]any `json:"metadata"`
		} `json:"items"`
	}
	if err := json.NewDecoder(inbox.Body).Decode(&inboxPayload); err != nil {
		t.Fatalf("decode inbox payload: %v", err)
	}
	if len(inboxPayload.Items) != 1 {
		t.Fatalf("expected one provider workflow Inbox event, got %+v", inboxPayload.Items)
	}
	item := inboxPayload.Items[0]
	if item.Source != "provider_workflow" || item.Status != "unread" || item.Title != "eBay workflow failed" {
		t.Fatalf("unexpected provider workflow Inbox item shell: %+v", item)
	}
	for _, want := range []string{"12001", "API_BROWSE", "rate limit reached"} {
		if !strings.Contains(item.Summary, want) {
			t.Fatalf("expected Inbox summary to preserve provider failure detail %q, got %q", want, item.Summary)
		}
	}
	expectedMetadata := map[string]string{
		"provider_id":           "ebay",
		"provider_display_name": "eBay",
		"workflow_action_id":    "market_watch.run",
		"required_action_code":  "check_provider_health_and_credentials",
		"category":              "integration_workflow",
		"severity":              "error",
		"target_route":          "/integrations",
		"query_set_id":          querySetID,
		"provider_error_code":   "PROVIDER_SEARCH_FAILED",
		"health_impact":         "updates_provider_health",
	}
	for key, want := range expectedMetadata {
		if got := fmt.Sprintf("%v", item.Metadata[key]); got != want {
			t.Fatalf("Inbox metadata[%s] got %q want %q; metadata=%+v", key, got, want, item.Metadata)
		}
	}
	if got, ok := item.Metadata["retry_after_seconds"].(float64); !ok || int(got) != 120 {
		t.Fatalf("expected Inbox retry_after_seconds=120, got metadata=%+v", item.Metadata)
	}

	repeat := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/providers/ebay/run",
		strings.NewReader(`{"query_set_id":"`+querySetID+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if repeat.Code != http.StatusTooManyRequests {
		t.Fatalf("expected repeat 429 search failure, got %d body=%s", repeat.Code, repeat.Body.String())
	}
	repeatedInbox := doRequest(t, a, http.MethodGet, "/api/chat/inbox?profile_id="+p.ID, nil, nil)
	if repeatedInbox.Code != http.StatusOK {
		t.Fatalf("repeated inbox status=%d body=%s", repeatedInbox.Code, repeatedInbox.Body.String())
	}
	var repeatedPayload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(repeatedInbox.Body).Decode(&repeatedPayload); err != nil {
		t.Fatalf("decode repeated inbox payload: %v", err)
	}
	if len(repeatedPayload.Items) != 1 {
		t.Fatalf("expected repeated provider workflow failures to coalesce into one Inbox item, got %+v", repeatedPayload.Items)
	}
}

func TestEbayProviderRunMapsBlankSafeKeywordsToQueryGuidance(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profile := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/profiles",
		strings.NewReader(`{"name":"ebay-query-validation-profile"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if profile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", profile.Code, profile.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(profile.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile payload: %v", err)
	}
	activate := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/profiles/active",
		strings.NewReader(`{"profile_id":"`+p.ID+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if activate.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", activate.Code, activate.Body.String())
	}
	saveSettings := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/profiles/"+p.ID+"/settings",
		strings.NewReader(`{"settings":{"ebay_bearer_token":"valid-token","ebay_marketplace":"EBAY_AU"}}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save ebay settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}

	create := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/scanner/query-sets",
		strings.NewReader(`{"name":"eBay Blank Keywords","keywords":["placeholder"],"provider_scope":["ebay"],"enabled":true}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if create.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", create.Code, create.Body.String())
	}
	var created map[string]any
	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatalf("decode create payload: %v", err)
	}
	querySetID, _ := created["id"].(string)
	if _, err := a.db.Exec(`UPDATE scanner_query_sets SET keywords_json = '[" ","bad%0Akeyword"]' WHERE id = ?`, querySetID); err != nil {
		t.Fatalf("force blank safe keywords: %v", err)
	}

	run := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/providers/ebay/run",
		strings.NewReader(`{"query_set_id":"`+querySetID+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if run.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 query validation failure, got %d body=%s", run.Code, run.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(run.Body).Decode(&payload); err != nil {
		t.Fatalf("decode ebay query validation payload: %v", err)
	}
	for key, expected := range map[string]any{
		"error":        "failed_to_run_ebay_provider",
		"error_code":   "PROVIDER_QUERY_INVALID",
		"provider":     "ebay",
		"query_set_id": querySetID,
		"next_action":  "edit_ebay_query_criteria",
	} {
		if got := payload[key]; got != expected {
			t.Fatalf("query validation payload %s=%v, want %v in %+v", key, got, expected, payload)
		}
	}
	if message, _ := payload["message"].(string); !strings.Contains(message, "keywords are required") {
		t.Fatalf("expected actionable keyword message, got %+v", payload)
	}
}

func TestScannerRunMapsEbayBrowseRetryAfterToProviderErrorEnvelope(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profile := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/profiles",
		strings.NewReader(`{"name":"ebay-scanner-retry-after-profile"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if profile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", profile.Code, profile.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(profile.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile payload: %v", err)
	}
	activate := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/profiles/active",
		strings.NewReader(`{"profile_id":"`+p.ID+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if activate.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", activate.Code, activate.Body.String())
	}
	ebayStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", "45")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"errors":[{"errorId":12002,"domain":"API_BROWSE","category":"REQUEST","message":"rate limit reached","longMessage":"Retry after the provider window resets"}]}`))
	}))
	defer ebayStub.Close()

	saveSettings := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/profiles/"+p.ID+"/settings",
		strings.NewReader(`{"settings":{"ebay_base_url":"`+ebayStub.URL+`","ebay_bearer_token":"valid-token","ebay_marketplace":"EBAY_AU"}}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save ebay settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}

	create := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/scanner/query-sets",
		strings.NewReader(`{"name":"eBay Scanner Retry After","keywords":["afx"],"provider_scope":["ebay"],"enabled":true}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if create.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", create.Code, create.Body.String())
	}
	var created map[string]any
	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatalf("decode create payload: %v", err)
	}
	querySetID, _ := created["id"].(string)

	run := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/scanner/run",
		strings.NewReader(`{"query_set_id":"`+querySetID+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if run.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 search failure, got %d body=%s", run.Code, run.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(run.Body).Decode(&payload); err != nil {
		t.Fatalf("decode scanner run error payload: %v", err)
	}
	if payload["error_code"] != "PROVIDER_SEARCH_FAILED" || payload["provider"] != "ebay" {
		t.Fatalf("unexpected scanner provider error payload: %+v", payload)
	}
	if payload["next_action"] != "check_provider_health_and_credentials" {
		t.Fatalf("expected provider-health next action, got %+v", payload)
	}
	if got, ok := payload["retry_after_seconds"].(float64); !ok || int(got) != 45 {
		t.Fatalf("expected retry_after_seconds=45, got %+v", payload)
	}
	message, _ := payload["message"].(string)
	for _, want := range []string{"12002", "API_BROWSE", "REQUEST", "rate limit reached", "Retry after the provider window resets"} {
		if !strings.Contains(message, want) {
			t.Fatalf("expected scanner run message to preserve %q, got %q", want, message)
		}
	}
}

func TestScannerRecognitionReviewApplyRequiresConfirmationAndDoesNotMutate(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createScannerApplyProfile(t, a, "scanner-review-confirmation-profile")
	body := `{
		"target":"inventory",
		"confirmed":false,
		"candidates":[
			{"id":"SCAN-001","title":"Scanned Card","confidence":0.91,"source":"camera","provenance":"ocr-v1","media_id":"media-1","media_url":"https://example.test/scan-1.jpg"}
		]
	}`
	resp := doRequest(t, a, http.MethodPost, "/api/scanner/recognition-review/apply", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusConflict {
		t.Fatalf("expected confirmation-required conflict, got %d body=%s", resp.Code, resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"scanner_review_confirmation_required"`) || !strings.Contains(resp.Body.String(), `"confirm_before_create":true`) {
		t.Fatalf("expected confirmation-required review payload, body=%s", resp.Body.String())
	}
	items := doRequest(t, a, http.MethodGet, "/api/items", nil, nil)
	if items.Code != http.StatusOK {
		t.Fatalf("items status=%d body=%s", items.Code, items.Body.String())
	}
	if strings.Contains(items.Body.String(), "SCAN-001") {
		t.Fatalf("unconfirmed scanner apply must not create items, body=%s", items.Body.String())
	}
}

func TestScannerRecognitionReviewApplyCreatesWishlistItemWithEvidence(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createScannerApplyProfile(t, a, "scanner-review-wishlist-profile")
	body := `{
		"target":"wishlist",
		"confirmed":true,
		"candidates":[
			{"id":"SCAN-WISH-001","title":"Wishlist Scanner Card","confidence":0.72,"source":"camera","provenance":"ocr-v1","media_id":"media-wish-1","media_url":"https://example.test/wish-scan.jpg"},
			{"id":"SCAN-WISH-ALT","title":"Wishlist Scanner Alternate","confidence":0.64,"source":"catalog","provenance":"matcher-v2","override_id":"SCAN-WISH-ALT","override_note":"selected exact parallel"}
		]
	}`
	resp := doRequest(t, a, http.MethodPost, "/api/scanner/recognition-review/apply", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected created scanner review apply, got %d body=%s", resp.Code, resp.Body.String())
	}
	var payload struct {
		Target string `json:"target"`
		Item   struct {
			ID         string   `json:"id"`
			PartNumber string   `json:"part_number"`
			Title      string   `json:"title"`
			Status     string   `json:"status"`
			Notes      string   `json:"notes"`
			SourceURLs []string `json:"source_urls"`
		} `json:"item"`
		WishlistEntry struct {
			ID     string `json:"id"`
			ItemID string `json:"item_id"`
			Owned  bool   `json:"owned"`
		} `json:"wishlist_entry"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode scanner apply payload: %v", err)
	}
	if payload.Target != "wishlist" || payload.Item.ID == "" || payload.WishlistEntry.ID == "" || payload.WishlistEntry.ItemID != payload.Item.ID {
		t.Fatalf("expected wishlist item and entry, got %+v", payload)
	}
	if payload.Item.PartNumber != "SCAN-WISH-ALT" || payload.Item.Title != "Wishlist Scanner Alternate" || payload.Item.Status != "wishlist" {
		t.Fatalf("expected manual override wishlist item, got %+v", payload.Item)
	}
	if payload.WishlistEntry.Owned {
		t.Fatalf("scanner wishlist apply must not mark wishlist entry owned, got %+v", payload.WishlistEntry)
	}
	for _, want := range []string{"manual_override=true", "media_id=media-wish-1", "provenance=camera|ocr-v1|catalog|matcher-v2"} {
		if !strings.Contains(payload.Item.Notes, want) {
			t.Fatalf("expected item notes to contain %q, got %q", want, payload.Item.Notes)
		}
	}
	if len(payload.Item.SourceURLs) != 1 || payload.Item.SourceURLs[0] != "https://example.test/wish-scan.jpg" {
		t.Fatalf("expected scanner media URL source evidence, got %+v", payload.Item.SourceURLs)
	}
	items := doRequest(t, a, http.MethodGet, "/api/items?status=wishlist", nil, nil)
	if items.Code != http.StatusOK || !strings.Contains(items.Body.String(), "SCAN-WISH-ALT") {
		t.Fatalf("expected created wishlist item to reload, status=%d body=%s", items.Code, items.Body.String())
	}
}

func createScannerApplyProfile(t *testing.T, a *App, name string) string {
	t.Helper()
	profile := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/profiles",
		strings.NewReader(`{"name":"`+name+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if profile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", profile.Code, profile.Body.String())
	}
	var payload struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(profile.Body).Decode(&payload); err != nil {
		t.Fatalf("decode profile payload: %v", err)
	}
	activate := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/profiles/active",
		strings.NewReader(`{"profile_id":"`+payload.ID+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if activate.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", activate.Code, activate.Body.String())
	}
	return payload.ID
}

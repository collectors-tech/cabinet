package app

import (
	"encoding/json"
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
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"I-1","title":"AFX Camaro","price":{"value":"49.95"},"itemWebUrl":"https://example/item/1","image":{"imageUrl":"https://example/image/1.jpg"},"seller":{"username":"seller-a"}}]}`))
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
	if healthPayload["state"] != "degraded" || healthPayload["last_error"] == nil {
		t.Fatalf("expected degraded provider health with last_error, got %+v", healthPayload)
	}
	if got, ok := healthPayload["retry_after_seconds"].(float64); !ok || int(got) != 120 {
		t.Fatalf("expected provider health retry_after_seconds=120, got %+v", healthPayload)
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

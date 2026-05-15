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

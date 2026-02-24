package app

import (
	"encoding/json"
	"net/http"
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

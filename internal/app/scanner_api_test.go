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

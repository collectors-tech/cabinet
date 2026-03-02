package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestDefaultSiteSearchProviderScopeContract(t *testing.T) {
	a := newTestApp(t)

	create := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/scanner/query-sets",
		strings.NewReader(`{"name":"default-scope-contract","keywords":["slot car"],"exclusions":[],"max_price":120,"region":"AU","condition":"used","source":""}`),
		map[string]string{"Content-Type": "application/json"},
	)
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
		t.Fatalf("decode query sets: %v", err)
	}
	if len(payload.QuerySets) == 0 {
		t.Fatal("expected query_sets payload")
	}

	first := payload.QuerySets[0]
	rawScope, ok := first["provider_scope"]
	if !ok {
		t.Fatalf("expected provider_scope field in query set payload, got %+v", first)
	}
	providerScope, ok := rawScope.([]any)
	if !ok {
		t.Fatalf("expected provider_scope array, got %T", rawScope)
	}
	if len(providerScope) < 2 {
		t.Fatalf("expected default provider_scope entries, got %+v", providerScope)
	}
	if providerScope[0] != "ebay" {
		t.Fatalf("expected default provider_scope to start with ebay, got %+v", providerScope)
	}
	if providerScope[1] != "amazon" {
		t.Fatalf("expected default provider_scope to include amazon, got %+v", providerScope)
	}
	querySetID, _ := first["id"].(string)
	if querySetID == "" {
		t.Fatalf("expected query set id in payload, got %+v", first)
	}

	update := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/scanner/query-sets/"+querySetID,
		strings.NewReader(`{"name":"default-scope-contract","keywords":["slot car"],"exclusions":[],"provider_scope":["ebay","mrtoys"],"max_price":120,"region":"AU","condition":"used","enabled":true}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if update.Code != http.StatusOK {
		t.Fatalf("update query set status=%d body=%s", update.Code, update.Body.String())
	}

	getAfterUpdate := doRequest(t, a, http.MethodGet, "/api/scanner/query-sets", nil, nil)
	if getAfterUpdate.Code != http.StatusOK {
		t.Fatalf("list query sets after update status=%d body=%s", getAfterUpdate.Code, getAfterUpdate.Body.String())
	}
	var updatedPayload struct {
		QuerySets []map[string]any `json:"query_sets"`
	}
	if err := json.NewDecoder(getAfterUpdate.Body).Decode(&updatedPayload); err != nil {
		t.Fatalf("decode updated query sets: %v", err)
	}
	if len(updatedPayload.QuerySets) == 0 {
		t.Fatal("expected updated query set payload")
	}
	updatedScope, ok := updatedPayload.QuerySets[0]["provider_scope"].([]any)
	if !ok {
		t.Fatalf("expected updated provider_scope array, got %+v", updatedPayload.QuerySets[0]["provider_scope"])
	}
	if len(updatedScope) != 2 || updatedScope[0] != "ebay" || updatedScope[1] != "mrtoys" {
		t.Fatalf("expected persisted provider_scope override [ebay mrtoys], got %+v", updatedScope)
	}
}

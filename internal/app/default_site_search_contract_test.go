package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestDefaultSiteSearchScheduledRefreshPersistsRunSnapshot(t *testing.T) {
	t.Parallel()

	ebayStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET to ebay stub, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"itemSummaries\":[{\"itemId\":\"scheduled-1\",\"title\":\"Scheduled AFX Camaro\",\"price\":{\"value\":\"42.50\"},\"itemWebUrl\":\"https://example.test/scheduled-1\",\"seller\":{\"username\":\"seller-a\"}}]}"))
	}))
	defer ebayStub.Close()

	a := newTestApp(t)
	profile := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/profiles",
		strings.NewReader("{\"name\":\"default-search-scheduled-snapshot\"}"),
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
		strings.NewReader("{\"profile_id\":\""+p.ID+"\"}"),
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
		strings.NewReader("{\"settings\":{\"ebay_base_url\":\""+ebayStub.URL+"\",\"ebay_bearer_token\":\"test-token\",\"ebay_marketplace\":\"EBAY_AU\"}}"),
		map[string]string{"Content-Type": "application/json"},
	)
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}

	create := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/scanner/query-sets",
		strings.NewReader("{\"name\":\"Scheduled AFX\",\"keywords\":[\"afx\"],\"provider_scope\":[\"ebay\"],\"schedule_cron\":\"*/15 * * * *\",\"enabled\":true}"),
		map[string]string{"Content-Type": "application/json"},
	)
	if create.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", create.Code, create.Body.String())
	}
	var created map[string]any
	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatalf("decode created query set: %v", err)
	}
	querySetID, _ := created["id"].(string)
	if querySetID == "" {
		t.Fatalf("expected query set id, got %+v", created)
	}

	scheduled := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/scanner/run/scheduled",
		strings.NewReader("{}"),
		map[string]string{"Content-Type": "application/json"},
	)
	if scheduled.Code != http.StatusOK {
		t.Fatalf("scheduled refresh status=%d body=%s", scheduled.Code, scheduled.Body.String())
	}
	var runSummary map[string]any
	if err := json.NewDecoder(scheduled.Body).Decode(&runSummary); err != nil {
		t.Fatalf("decode scheduled refresh summary: %v", err)
	}
	if got, ok := runSummary["query_sets_executed"].(float64); !ok || int(got) != 1 {
		t.Fatalf("expected query_sets_executed=1, got %#v", runSummary["query_sets_executed"])
	}
	if got, ok := runSummary["candidates_collected"].(float64); !ok || int(got) != 1 {
		t.Fatalf("expected candidates_collected=1, got %#v", runSummary["candidates_collected"])
	}

	reloaded := doRequest(t, a, http.MethodGet, "/api/scanner/query-sets", nil, nil)
	if reloaded.Code != http.StatusOK {
		t.Fatalf("reload query sets status=%d body=%s", reloaded.Code, reloaded.Body.String())
	}
	var payload struct {
		QuerySets []map[string]any `json:"query_sets"`
	}
	if err := json.NewDecoder(reloaded.Body).Decode(&payload); err != nil {
		t.Fatalf("decode reloaded query sets: %v", err)
	}
	if len(payload.QuerySets) != 1 {
		t.Fatalf("expected one reloaded query set, got %d", len(payload.QuerySets))
	}
	latest := payload.QuerySets[0]
	if got, _ := latest["id"].(string); got != querySetID {
		t.Fatalf("expected reloaded query set id %q, got %#v", querySetID, latest["id"])
	}
	if got, _ := latest["last_run_status"].(string); got != "succeeded" {
		t.Fatalf("expected persisted last_run_status=succeeded after scheduled refresh, got %#v", latest["last_run_status"])
	}
	if got, ok := latest["last_candidate_count"].(float64); !ok || int(got) != 1 {
		t.Fatalf("expected persisted last_candidate_count=1 after scheduled refresh, got %#v", latest["last_candidate_count"])
	}
	if got, _ := latest["last_run_at"].(string); strings.TrimSpace(got) == "" {
		t.Fatalf("expected persisted last_run_at after scheduled refresh, got %#v", latest["last_run_at"])
	}
}

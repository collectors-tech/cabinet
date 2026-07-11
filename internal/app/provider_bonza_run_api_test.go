package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBonzaRunAggregatesPagesAndEnrichesWatchedStock(t *testing.T) {
	t.Parallel()

	requests := map[string]int{}
	bonza := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests[r.URL.Path]++
		switch {
		case strings.HasPrefix(r.URL.Path, "/wp-json/wc/store/v1/products"):
			page := r.URL.Query().Get("page")
			w.Header().Set("Content-Type", "application/json")
			if page == "2" {
				_, _ = w.Write([]byte(`[
{"id":1001,"name":"AFX Camaro Variant","permalink":"` + r.Host + `/product/1001","is_in_stock":true,"low_stock_remaining":null},
{"id":1003,"name":"AFX Mustang","permalink":"` + r.Host + `/product/1003","is_in_stock":false,"low_stock_remaining":null}
]`))
				return
			}
			w.Header().Set("X-WP-TotalPages", "2")
			_, _ = w.Write([]byte(`[
{"id":1001,"name":"AFX Camaro","permalink":"` + r.Host + `/product/1001","is_in_stock":true,"low_stock_remaining":null},
{"id":1002,"name":"AFX Thunderjet","permalink":"` + r.Host + `/product/1002","is_in_stock":null,"low_stock_remaining":null}
]`))
		case strings.HasPrefix(r.URL.Path, "/product/1002"):
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html><body><span>Only 2 left in stock</span></body></html>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer bonza.Close()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"BonzaProfile"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	activate := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if activate.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", activate.Code, activate.Body.String())
	}

	settingsBody := fmt.Sprintf(`{"settings":{"integration.bonzaslotcars.base_url":"%s","integration.bonzaslotcars.items_per_page":"36"}}`, bonza.URL)
	saveSettings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(settingsBody), map[string]string{"Content-Type": "application/json"})
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}

	createQuery := doRequest(t, a, http.MethodPost, "/api/scanner/query-sets", strings.NewReader(`{"name":"AFX","keywords":["AFX"],"provider_scope":["bonzaslotcars"],"enabled":true}`), map[string]string{"Content-Type": "application/json"})
	if createQuery.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", createQuery.Code, createQuery.Body.String())
	}
	var qs struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createQuery.Body).Decode(&qs); err != nil {
		t.Fatalf("decode query set: %v", err)
	}

	run := doRequest(t, a, http.MethodPost, "/api/providers/bonza/run", strings.NewReader(`{"query_set_id":"`+qs.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if run.Code != http.StatusOK {
		t.Fatalf("bonza run status=%d body=%s", run.Code, run.Body.String())
	}

	var payload struct {
		QuerySetID       string           `json:"query_set_id"`
		PageCount        int              `json:"page_count"`
		ObservedPageSize int              `json:"observed_page_size"`
		Candidates       []map[string]any `json:"candidates"`
		Run              map[string]any   `json:"run"`
		RunSummary       map[string]any   `json:"run_summary"`
	}
	if err := json.NewDecoder(run.Body).Decode(&payload); err != nil {
		t.Fatalf("decode bonza payload: %v", err)
	}
	if payload.QuerySetID != qs.ID {
		t.Fatalf("query_set_id mismatch: got=%s want=%s", payload.QuerySetID, qs.ID)
	}
	if payload.PageCount != 2 {
		t.Fatalf("expected page_count=2, got=%d", payload.PageCount)
	}
	if payload.ObservedPageSize == 0 {
		t.Fatalf("expected observed_page_size > 0, got=%d", payload.ObservedPageSize)
	}
	if len(payload.Candidates) != 3 {
		t.Fatalf("expected 3 deduplicated candidates, got=%d payload=%+v", len(payload.Candidates), payload.Candidates)
	}
	if saved, ok := payload.Run["saved"].(float64); !ok || int(saved) != 3 {
		t.Fatalf("expected persisted run saved=3, got %+v", payload.Run)
	}

	var enriched bool
	for _, candidate := range payload.Candidates {
		listingID, _ := candidate["listing_id"].(string)
		if listingID == "bonza-1002" {
			state, _ := candidate["stock_state"].(string)
			source, _ := candidate["source"].(string)
			if source != "bonzaslotcars" {
				t.Fatalf("expected persisted bonza source, got %+v", candidate)
			}
			if state == "low_stock" {
				enriched = true
			}
		}
	}
	if !enriched {
		t.Fatalf("expected fallback detail stock enrichment for bonza-1002, payload=%+v", payload.Candidates)
	}

	reloaded := doRequest(t, a, http.MethodGet, "/api/scanner/query-sets", nil, nil)
	if reloaded.Code != http.StatusOK {
		t.Fatalf("reload query sets status=%d body=%s", reloaded.Code, reloaded.Body.String())
	}
	var querySetPayload struct {
		QuerySets []map[string]any `json:"query_sets"`
	}
	if err := json.NewDecoder(reloaded.Body).Decode(&querySetPayload); err != nil {
		t.Fatalf("decode reloaded query sets: %v", err)
	}
	if len(querySetPayload.QuerySets) != 1 {
		t.Fatalf("expected one reloaded query set, got %+v", querySetPayload.QuerySets)
	}
	latest := querySetPayload.QuerySets[0]
	if got, _ := latest["last_run_status"].(string); got != "succeeded" {
		t.Fatalf("expected bonza run to persist succeeded snapshot, got %+v", latest)
	}
	if got, ok := latest["last_candidate_count"].(float64); !ok || int(got) != 3 {
		t.Fatalf("expected bonza run to persist candidate count=3, got %+v", latest)
	}
}

func TestBonzaRunRecordsLiveProviderProofForBetaRegistry(t *testing.T) {
	t.Parallel()

	bonza := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/wp-json/wc/store/v1/products") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-WP-TotalPages", "1")
		_, _ = w.Write([]byte(`[
{"id":2001,"name":"AFX Mega-G+ Mustang","permalink":"` + r.Host + `/product/2001","prices":{"price":"7495","currency_code":"AUD"},"is_in_stock":true,"low_stock_remaining":null}
]`))
	}))
	defer bonza.Close()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"BonzaLiveProofProfile"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	activate := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if activate.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", activate.Code, activate.Body.String())
	}

	settingsBody := fmt.Sprintf(`{"settings":{"integration.bonzaslotcars.base_url":"%s","integration.bonzaslotcars.items_per_page":"36"}}`, bonza.URL)
	saveSettings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(settingsBody), map[string]string{"Content-Type": "application/json"})
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}
	createQuery := doRequest(t, a, http.MethodPost, "/api/scanner/query-sets", strings.NewReader(`{"name":"Bonza AFX Live Proof","keywords":["AFX"],"provider_scope":["bonzaslotcars"],"enabled":true}`), map[string]string{"Content-Type": "application/json"})
	if createQuery.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", createQuery.Code, createQuery.Body.String())
	}
	var qs struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createQuery.Body).Decode(&qs); err != nil {
		t.Fatalf("decode query set: %v", err)
	}

	run := doRequest(t, a, http.MethodPost, "/api/providers/bonza/run", strings.NewReader(`{"query_set_id":"`+qs.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if run.Code != http.StatusOK {
		t.Fatalf("bonza run status=%d body=%s", run.Code, run.Body.String())
	}

	registry := doRequest(t, a, http.MethodGet, "/api/providers/registry", nil, nil)
	if registry.Code != http.StatusOK {
		t.Fatalf("registry status=%d body=%s", registry.Code, registry.Body.String())
	}
	var payload struct {
		Providers []map[string]any `json:"providers"`
	}
	if err := json.NewDecoder(registry.Body).Decode(&payload); err != nil {
		t.Fatalf("decode registry payload: %v", err)
	}
	bonzaProvider := findRegistryProvider(payload.Providers, "au-webshop-bonzaslotcars-com-au")
	if bonzaProvider == nil {
		t.Fatalf("Bonza provider missing from registry payload: %+v", payload.Providers)
	}
	if got := fmt.Sprintf("%v", bonzaProvider["beta_release_status"]); got != "available_live_validated" {
		t.Fatalf("Bonza beta release status got %q want available_live_validated: %+v", got, bonzaProvider)
	}
	if got := fmt.Sprintf("%v", bonzaProvider["live_evidence_state"]); got != "validated" {
		t.Fatalf("Bonza live evidence state got %q want validated: %+v", got, bonzaProvider)
	}
	action := findRegistryAction(anySlice(bonzaProvider["actions"]), "market_watch.run")
	if action == nil {
		t.Fatalf("Bonza provider missing market_watch.run action after live proof: %+v", bonzaProvider)
	}
	if got := fmt.Sprintf("%v", action["availability_state"]); got != "available" {
		t.Fatalf("Bonza market_watch.run availability got %q want available: %+v", got, action)
	}
	if action["next_action"] != nil {
		t.Fatalf("Bonza market_watch.run next_action got %+v want nil after live proof: %+v", action["next_action"], action)
	}
}

func TestBonzaRunFailureCreatesProviderWorkflowInboxEvent(t *testing.T) {
	t.Parallel()

	bonza := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream unavailable", http.StatusServiceUnavailable)
	}))
	defer bonza.Close()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"BonzaFailureProfile"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	activate := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if activate.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", activate.Code, activate.Body.String())
	}
	settingsBody := fmt.Sprintf(`{"settings":{"integration.bonzaslotcars.base_url":"%s","integration.bonzaslotcars.items_per_page":"36"}}`, bonza.URL)
	saveSettings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(settingsBody), map[string]string{"Content-Type": "application/json"})
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}
	createQuery := doRequest(t, a, http.MethodPost, "/api/scanner/query-sets", strings.NewReader(`{"name":"AFX","keywords":["AFX"],"provider_scope":["bonzaslotcars"],"enabled":true}`), map[string]string{"Content-Type": "application/json"})
	if createQuery.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", createQuery.Code, createQuery.Body.String())
	}
	var qs struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createQuery.Body).Decode(&qs); err != nil {
		t.Fatalf("decode query set: %v", err)
	}

	run := doRequest(t, a, http.MethodPost, "/api/providers/bonza/run", strings.NewReader(`{"query_set_id":"`+qs.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if run.Code != http.StatusBadRequest {
		t.Fatalf("expected failed Bonza run status=400, got=%d body=%s", run.Code, run.Body.String())
	}

	inbox := doRequest(t, a, http.MethodGet, "/api/chat/inbox?profile_id="+profile.ID, nil, nil)
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
	if item.Source != "provider_workflow" || item.Status != "unread" || item.Title != "bonzaslotcars.com.au workflow failed" {
		t.Fatalf("unexpected provider workflow Inbox item shell: %+v", item)
	}
	for _, want := range []string{"status 503", "bonza"} {
		if !strings.Contains(strings.ToLower(item.Summary), want) {
			t.Fatalf("expected Inbox summary to preserve provider failure detail %q, got %q", want, item.Summary)
		}
	}
	expectedMetadata := map[string]string{
		"provider_id":           "au-webshop-bonzaslotcars-com-au",
		"provider_display_name": "bonzaslotcars.com.au",
		"workflow_action_id":    "market_watch.run",
		"required_action_code":  "check_provider_health_and_retry",
		"category":              "integration_workflow",
		"severity":              "error",
		"target_route":          "/integrations",
		"query_set_id":          qs.ID,
		"provider_error_code":   "FAILED_TO_RUN_BONZA",
		"health_impact":         "updates_provider_health",
		"base_url":              bonza.URL,
	}
	for key, want := range expectedMetadata {
		if got := fmt.Sprintf("%v", item.Metadata[key]); got != want {
			t.Fatalf("Inbox metadata[%s] got %q want %q; metadata=%+v", key, got, want, item.Metadata)
		}
	}

	repeat := doRequest(t, a, http.MethodPost, "/api/providers/bonza/run", strings.NewReader(`{"query_set_id":"`+qs.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if repeat.Code != http.StatusBadRequest {
		t.Fatalf("expected repeat failed Bonza run status=400, got=%d body=%s", repeat.Code, repeat.Body.String())
	}
	repeatedInbox := doRequest(t, a, http.MethodGet, "/api/chat/inbox?profile_id="+profile.ID, nil, nil)
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
		t.Fatalf("expected repeated Bonza provider workflow failures to coalesce into one Inbox item, got %+v", repeatedPayload.Items)
	}
}

package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBigCommerceRegistryExposesFamilyAndActiveMode(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"BigCommerceRegistryProfile"}`), map[string]string{"Content-Type": "application/json"})
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

	withToken := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/profiles/"+profile.ID+"/settings",
		strings.NewReader(`{"settings":{"integration.au_webshop_voglers_com_au.token":"token-123"}}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if withToken.Code != http.StatusOK {
		t.Fatalf("save settings with token status=%d body=%s", withToken.Code, withToken.Body.String())
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
	var voglers map[string]any
	for _, provider := range payload.Providers {
		if strings.EqualFold(fmt.Sprintf("%v", provider["base_domain"]), "voglers.com.au") {
			voglers = provider
			break
		}
	}
	if voglers == nil {
		t.Fatalf("expected voglers.com.au in registry payload: %+v", payload.Providers)
	}
	if !strings.EqualFold(fmt.Sprintf("%v", voglers["api_family"]), "bigcommerce") {
		t.Fatalf("expected api_family=bigcommerce, got=%v", voglers["api_family"])
	}
	if !strings.EqualFold(fmt.Sprintf("%v", voglers["active_mode"]), "token_enabled") {
		t.Fatalf("expected active_mode=token_enabled with token present, got=%v", voglers["active_mode"])
	}
}

func TestBigCommerceRunStorefrontModePersistsCandidatesAndSnapshot(t *testing.T) {
	t.Parallel()

	storefrontServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/products/search" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "products":[
    {"id":"bc-101","name":"AFX Mega G+ Camaro","url":"https://voglers.com.au/p/bc-101","price":"59.95","image":"https://voglers.com.au/img/bc-101.jpg"}
  ],
  "meta":{"total":1}
}`))
	}))
	defer storefrontServer.Close()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"BigCommerceStorefrontProfile"}`), map[string]string{"Content-Type": "application/json"})
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
	createQuery := doRequest(t, a, http.MethodPost, "/api/scanner/query-sets", strings.NewReader(`{"name":"AFX","keywords":["AFX"],"provider_scope":["voglers.com.au"],"enabled":true}`), map[string]string{"Content-Type": "application/json"})
	if createQuery.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", createQuery.Code, createQuery.Body.String())
	}
	var qs struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createQuery.Body).Decode(&qs); err != nil {
		t.Fatalf("decode query set: %v", err)
	}

	run := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/providers/bigcommerce/run",
		strings.NewReader(fmt.Sprintf(`{"query_set_id":"%s","provider_domain":"voglers.com.au","search_url":"%s/products/search"}`, qs.ID, storefrontServer.URL)),
		map[string]string{"Content-Type": "application/json"},
	)
	if run.Code != http.StatusOK {
		t.Fatalf("bigcommerce storefront run status=%d body=%s", run.Code, run.Body.String())
	}
	var payload struct {
		AuthMode         string           `json:"auth_mode"`
		DataDepthSource  string           `json:"data_depth_source"`
		CapabilityLimits []string         `json:"capability_limits"`
		CandidateCount   int              `json:"candidate_count"`
		Candidates       []map[string]any `json:"candidates"`
	}
	if err := json.NewDecoder(run.Body).Decode(&payload); err != nil {
		t.Fatalf("decode run payload: %v", err)
	}
	if payload.AuthMode != "storefront_public" {
		t.Fatalf("expected auth_mode=storefront_public, got=%q", payload.AuthMode)
	}
	if payload.DataDepthSource != "storefront_public" {
		t.Fatalf("expected data_depth_source=storefront_public, got=%q", payload.DataDepthSource)
	}
	if len(payload.CapabilityLimits) == 0 {
		t.Fatal("expected capability_limits in storefront mode")
	}
	if len(payload.Candidates) != 1 {
		t.Fatalf("expected one candidate, got=%d", len(payload.Candidates))
	}
	if payload.CandidateCount != 1 {
		t.Fatalf("expected candidate_count=1, got=%d", payload.CandidateCount)
	}
	if got, _ := payload.Candidates[0]["source"].(string); got != "voglers.com.au" {
		t.Fatalf("expected persisted BigCommerce candidate source=voglers.com.au, got %#v", payload.Candidates[0]["source"])
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
	var latest map[string]any
	for _, item := range querySetPayload.QuerySets {
		if fmt.Sprintf("%v", item["id"]) == qs.ID {
			latest = item
			break
		}
	}
	if latest == nil {
		t.Fatalf("expected query set %s in reloaded response: %+v", qs.ID, querySetPayload.QuerySets)
	}
	if got, _ := latest["last_run_status"].(string); got != "succeeded" {
		t.Fatalf("expected persisted last_run_status=succeeded, got %#v", latest["last_run_status"])
	}
	if got, ok := latest["last_candidate_count"].(float64); !ok || int(got) != 1 {
		t.Fatalf("expected persisted last_candidate_count=1, got %#v", latest["last_candidate_count"])
	}
}

func TestBigCommerceRunTokenModeUnlocksRicherDepth(t *testing.T) {
	t.Parallel()

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphql" {
			http.NotFound(w, r)
			return
		}
		if strings.TrimSpace(r.Header.Get("X-Auth-Token")) == "" {
			http.Error(w, `{"error":"missing_token"}`, http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "data":{
    "site":{
      "search":{
        "searchProducts":{
          "products":[
            {
              "entityId":102,
              "name":"AFX Mega G+ Corvette",
              "path":"/p/bc-102",
              "prices":{"price":{"value":62.95}},
              "inventory":{"isInStock":true,"aggregated":{"availableToSell":7}}
            }
          ]
        }
      }
    }
  }
}`))
	}))
	defer tokenServer.Close()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"BigCommerceTokenProfile"}`), map[string]string{"Content-Type": "application/json"})
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
	saveSettings := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/profiles/"+profile.ID+"/settings",
		strings.NewReader(`{"settings":{"integration.au_webshop_voglers_com_au.token":"token-abc"}}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}
	createQuery := doRequest(t, a, http.MethodPost, "/api/scanner/query-sets", strings.NewReader(`{"name":"AFX","keywords":["AFX"],"provider_scope":["voglers.com.au"],"enabled":true}`), map[string]string{"Content-Type": "application/json"})
	if createQuery.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", createQuery.Code, createQuery.Body.String())
	}
	var qs struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createQuery.Body).Decode(&qs); err != nil {
		t.Fatalf("decode query set: %v", err)
	}

	run := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/providers/bigcommerce/run",
		strings.NewReader(fmt.Sprintf(`{"query_set_id":"%s","provider_domain":"voglers.com.au","graphql_url":"%s/graphql"}`, qs.ID, tokenServer.URL)),
		map[string]string{"Content-Type": "application/json"},
	)
	if run.Code != http.StatusOK {
		t.Fatalf("bigcommerce token run status=%d body=%s", run.Code, run.Body.String())
	}
	var payload struct {
		AuthMode         string           `json:"auth_mode"`
		DataDepthSource  string           `json:"data_depth_source"`
		CapabilityLimits []string         `json:"capability_limits"`
		Candidates       []map[string]any `json:"candidates"`
	}
	if err := json.NewDecoder(run.Body).Decode(&payload); err != nil {
		t.Fatalf("decode run payload: %v", err)
	}
	if payload.AuthMode != "token_enabled" {
		t.Fatalf("expected auth_mode=token_enabled, got=%q", payload.AuthMode)
	}
	if payload.DataDepthSource != "token_enabled" {
		t.Fatalf("expected data_depth_source=token_enabled, got=%q", payload.DataDepthSource)
	}
	if len(payload.CapabilityLimits) != 0 {
		t.Fatalf("expected no capability limits in token mode, got=%v", payload.CapabilityLimits)
	}
	if len(payload.Candidates) != 1 {
		t.Fatalf("expected one candidate, got=%d", len(payload.Candidates))
	}
}

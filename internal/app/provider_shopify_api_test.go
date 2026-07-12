package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestShopifyRunPersistsCandidatesAndLiveProviderProof(t *testing.T) {
	t.Parallel()

	server := shoppingFixtureServer(t)
	a := newTestApp(t)

	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"ShopifyRunProfile"}`), map[string]string{"Content-Type": "application/json"})
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
	createQuery := doRequest(t, a, http.MethodPost, "/api/scanner/query-sets", strings.NewReader(`{"name":"Tamiya","keywords":["tamiya"],"provider_scope":["andrewshobbies.com.au"],"enabled":true}`), map[string]string{"Content-Type": "application/json"})
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
		"/api/providers/shopify/run",
		strings.NewReader(fmt.Sprintf(`{"query_set_id":"%s","provider_domain":"andrewshobbies.com.au","search_url":"%s/shopify/products.json"}`, qs.ID, server.URL)),
		map[string]string{"Content-Type": "application/json"},
	)
	if run.Code != http.StatusOK {
		t.Fatalf("shopify run status=%d body=%s", run.Code, run.Body.String())
	}
	var payload struct {
		ProviderScope    string           `json:"provider_scope"`
		AuthMode         string           `json:"auth_mode"`
		DataDepthSource  string           `json:"data_depth_source"`
		CapabilityLimits []string         `json:"capability_limits"`
		CandidateCount   int              `json:"candidate_count"`
		Candidates       []map[string]any `json:"candidates"`
	}
	if err := json.NewDecoder(run.Body).Decode(&payload); err != nil {
		t.Fatalf("decode Shopify run payload: %v", err)
	}
	if payload.ProviderScope != "andrewshobbies" {
		t.Fatalf("expected provider_scope=andrewshobbies, got %q", payload.ProviderScope)
	}
	if payload.AuthMode != "storefront_public" || payload.DataDepthSource != "shopify_products_json" {
		t.Fatalf("expected public Shopify products JSON mode, got %+v", payload)
	}
	if !containsShopifyString(payload.CapabilityLimits, "no_login_cart_checkout_or_private_api") {
		t.Fatalf("expected Shopify guardrail capability limit, got %+v", payload.CapabilityLimits)
	}
	if payload.CandidateCount != 1 || len(payload.Candidates) != 1 {
		t.Fatalf("expected one persisted Shopify candidate, got %+v", payload)
	}
	if got, _ := payload.Candidates[0]["source"].(string); got != "andrewshobbies" {
		t.Fatalf("expected persisted Shopify source=andrewshobbies, got %+v", payload.Candidates[0])
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

	registry := doRequest(t, a, http.MethodGet, "/api/providers/registry", nil, nil)
	if registry.Code != http.StatusOK {
		t.Fatalf("registry status=%d body=%s", registry.Code, registry.Body.String())
	}
	var registryPayload struct {
		Providers []map[string]any `json:"providers"`
	}
	if err := json.NewDecoder(registry.Body).Decode(&registryPayload); err != nil {
		t.Fatalf("decode registry payload: %v", err)
	}
	andrew := findRegistryProvider(registryPayload.Providers, "au-webshop-andrewshobbies-com-au")
	if andrew == nil {
		t.Fatalf("Andrew's Hobbies provider missing from registry payload: %+v", registryPayload.Providers)
	}
	if got := fmt.Sprintf("%v", andrew["beta_release_status"]); got != "available_live_validated" {
		t.Fatalf("expected Shopify registry live proof to upgrade beta status, got %q provider=%+v", got, andrew)
	}
	health, _ := andrew["health"].(map[string]any)
	if got := fmt.Sprintf("%v", health["status"]); got != "ok" {
		t.Fatalf("expected Shopify provider health=ok after run, got %+v", health)
	}
}

func containsShopifyString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

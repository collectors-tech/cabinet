package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAIRegistryExposesMethodAwareSetupNeededState(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"OpenAIRegistryProfile"}`), map[string]string{"Content-Type": "application/json"})
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
	openai := findRegistryProvider(payload.Providers, "openai")
	if openai == nil {
		t.Fatalf("expected openai provider in registry payload: %+v", payload.Providers)
	}
	if got := fmt.Sprintf("%v", openai["auth_mode"]); got != "hybrid" {
		t.Fatalf("expected hybrid auth mode, got %q", got)
	}
	if got := fmt.Sprintf("%v", openai["state"]); got != "needs_config" {
		t.Fatalf("expected setup-needed state before proof, got %q", got)
	}
	methods, ok := openai["auth_methods"].(map[string]any)
	if !ok {
		t.Fatalf("expected auth_methods map, got %#v", openai["auth_methods"])
	}
	browser, ok := methods["browser_auth"].(map[string]any)
	if !ok {
		t.Fatalf("expected browser_auth method, got %#v", methods["browser_auth"])
	}
	if got := fmt.Sprintf("%v", browser["state"]); got != "setup_needed" {
		t.Fatalf("expected browser auth setup_needed until proof, got %q", got)
	}
	if connected, _ := browser["connected"].(bool); connected {
		t.Fatalf("browser auth must not be connected without a verified artifact/callback")
	}
}

func TestOpenAIRegistryUsesPersistedActiveMethodWithoutBrowserNavigationProof(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"OpenAIConfiguredProfile"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	settings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(`{"settings":{"openai.active_auth_method":"browser_auth","openai.browser_auth_state":"pending","openai.browser_auth_artifact_present":"false"}}`), map[string]string{"Content-Type": "application/json"})
	if settings.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", settings.Code, settings.Body.String())
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
	openai := findRegistryProvider(payload.Providers, "openai")
	if got := fmt.Sprintf("%v", openai["state"]); got != "needs_config" {
		t.Fatalf("OpenAI registry state must stay setup-needed without Browser Auth proof, got %q", got)
	}
	methods := openai["auth_methods"].(map[string]any)
	browser := methods["browser_auth"].(map[string]any)
	if got := fmt.Sprintf("%v", browser["state"]); got != "pending" {
		t.Fatalf("expected browser auth pending state, got %q", got)
	}
	if connected, _ := browser["connected"].(bool); connected {
		t.Fatalf("browser auth must not be connected when active method is browser_auth but proof is pending")
	}
}

func TestOpenAIProviderHealthReflectsProfileReadiness(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"OpenAIHealthProfile"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})

	missing := doRequest(t, a, http.MethodGet, "/api/provider/health?provider=openai", nil, nil)
	if missing.Code != http.StatusOK {
		t.Fatalf("missing health status=%d body=%s", missing.Code, missing.Body.String())
	}
	var missingPayload map[string]any
	if err := json.NewDecoder(missing.Body).Decode(&missingPayload); err != nil {
		t.Fatalf("decode missing health: %v", err)
	}
	if missingPayload["status"] != "needs_config" || missingPayload["next_action"] != "connect_openai_api_key_or_browser_auth" {
		t.Fatalf("expected setup-needed OpenAI health, got %+v", missingPayload)
	}

	settings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(`{"settings":{"openai.active_auth_method":"api_key"}}`), map[string]string{"Content-Type": "application/json"})
	if settings.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", settings.Code, settings.Body.String())
	}
	apiKeyMissing := doRequest(t, a, http.MethodGet, "/api/provider/health?provider=openai", nil, nil)
	var apiKeyMissingPayload map[string]any
	if err := json.NewDecoder(apiKeyMissing.Body).Decode(&apiKeyMissingPayload); err != nil {
		t.Fatalf("decode api key missing health: %v", err)
	}
	if apiKeyMissingPayload["status"] != "needs_config" || apiKeyMissingPayload["code"] != "OPENAI_API_KEY_MISSING" {
		t.Fatalf("expected missing api-key health, got %+v", apiKeyMissingPayload)
	}

	secret := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/secrets", strings.NewReader(`{"key":"openai_api_key","value":"sk-test-openai"}`), map[string]string{"Content-Type": "application/json"})
	if secret.Code != http.StatusOK {
		t.Fatalf("secret status=%d body=%s", secret.Code, secret.Body.String())
	}
	ready := doRequest(t, a, http.MethodGet, "/api/provider/health?provider=openai", nil, nil)
	var readyPayload map[string]any
	if err := json.NewDecoder(ready.Body).Decode(&readyPayload); err != nil {
		t.Fatalf("decode ready health: %v", err)
	}
	if readyPayload["status"] != "ready" || readyPayload["auth_method"] != "api_key" || readyPayload["credential_present"] != true {
		t.Fatalf("expected API-key ready health without secret value, got %+v", readyPayload)
	}
	if _, leaked := readyPayload["secret"]; leaked {
		t.Fatalf("OpenAI health payload must not expose secret material: %+v", readyPayload)
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
	openai := findRegistryProvider(registryPayload.Providers, "openai")
	methods := openai["auth_methods"].(map[string]any)
	apiKey := methods["api_key"].(map[string]any)
	if openai["state"] != "ready" || apiKey["connected"] != true || apiKey["credential_present"] != true {
		t.Fatalf("expected registry to become ready only after API-key secret is stored, got provider=%+v api_key=%+v", openai, apiKey)
	}

	browserSettings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(`{"settings":{"openai.active_auth_method":"browser_auth","openai.browser_auth_state":"pending","openai.browser_auth_artifact_present":"false"}}`), map[string]string{"Content-Type": "application/json"})
	if browserSettings.Code != http.StatusOK {
		t.Fatalf("browser settings status=%d body=%s", browserSettings.Code, browserSettings.Body.String())
	}
	browser := doRequest(t, a, http.MethodGet, "/api/provider/health?provider=openai", nil, nil)
	var browserPayload map[string]any
	if err := json.NewDecoder(browser.Body).Decode(&browserPayload); err != nil {
		t.Fatalf("decode browser health: %v", err)
	}
	if browserPayload["status"] != "needs_config" || browserPayload["code"] != "OPENAI_BROWSER_AUTH_PROOF_REQUIRED" {
		t.Fatalf("expected browser proof-required health, got %+v", browserPayload)
	}
}

func TestOpenAIProviderTestReturnsAuditableConnectivityEvidence(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"OpenAIProviderTestProfile"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	missingSetup := doRequest(t, a, http.MethodPost, "/api/provider/test", strings.NewReader(`{"provider":"openai","profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if missingSetup.Code != http.StatusBadRequest {
		t.Fatalf("expected setup-needed status, got %d body=%s", missingSetup.Code, missingSetup.Body.String())
	}
	var missingPayload map[string]any
	if err := json.NewDecoder(missingSetup.Body).Decode(&missingPayload); err != nil {
		t.Fatalf("decode missing setup payload: %v", err)
	}
	if missingPayload["status"] != "needs_config" || missingPayload["code"] != "OPENAI_AUTH_METHOD_REQUIRED" || missingPayload["provider_test_passed"] != false {
		t.Fatalf("expected explicit setup-needed provider-test evidence, got %+v", missingPayload)
	}

	var seenAuth string
	failProviderTest := true
	openaiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/models":
			if failProviderTest {
				http.Error(w, `{"error":"bad_upstream"}`, http.StatusUnauthorized)
				return
			}
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o-mini"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer openaiServer.Close()

	settings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(`{"settings":{"openai.active_auth_method":"api_key","openai_base_url":"`+openaiServer.URL+`"}}`), map[string]string{"Content-Type": "application/json"})
	if settings.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", settings.Code, settings.Body.String())
	}
	secret := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/secrets", strings.NewReader(`{"key":"openai_api_key","value":"sk-provider-test"}`), map[string]string{"Content-Type": "application/json"})
	if secret.Code != http.StatusOK {
		t.Fatalf("secret status=%d body=%s", secret.Code, secret.Body.String())
	}

	failed := doRequest(t, a, http.MethodPost, "/api/provider/test", strings.NewReader(`{"provider":"openai","profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if failed.Code != http.StatusBadRequest {
		t.Fatalf("expected failed provider-test status, got %d body=%s", failed.Code, failed.Body.String())
	}
	var failedPayload map[string]any
	if err := json.NewDecoder(failed.Body).Decode(&failedPayload); err != nil {
		t.Fatalf("decode failed provider-test payload: %v", err)
	}
	if failedPayload["status"] != "failed" || failedPayload["code"] != "OPENAI_PROVIDER_TEST_FAILED" || failedPayload["provider_test_passed"] != false || failedPayload["credential_present"] != true {
		t.Fatalf("expected truthful failed provider-test evidence, got %+v", failedPayload)
	}
	if strings.Contains(failed.Body.String(), "sk-provider-test") {
		t.Fatalf("provider-test response must not leak API key, body=%s", failed.Body.String())
	}

	failProviderTest = false
	passed := doRequest(t, a, http.MethodPost, "/api/provider/test", strings.NewReader(`{"provider":"openai","profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if passed.Code != http.StatusOK {
		t.Fatalf("expected passed provider-test status, got %d body=%s", passed.Code, passed.Body.String())
	}
	var passedPayload map[string]any
	if err := json.NewDecoder(passed.Body).Decode(&passedPayload); err != nil {
		t.Fatalf("decode passed provider-test payload: %v", err)
	}
	if passedPayload["status"] != "ready" || passedPayload["code"] != "OPENAI_PROVIDER_TEST_PASSED" || passedPayload["provider_test_passed"] != true || passedPayload["auth_method"] != "api_key" {
		t.Fatalf("expected ready provider-test evidence, got %+v", passedPayload)
	}
	if seenAuth != "Bearer sk-provider-test" {
		t.Fatalf("expected provider test to call OpenAI-compatible endpoint with bearer auth, got %q", seenAuth)
	}
	if strings.Contains(passed.Body.String(), "sk-provider-test") {
		t.Fatalf("provider-test success response must not leak API key, body=%s", passed.Body.String())
	}
}

func TestOpenAIBrowserAuthProviderTestRequiresVerifiedProof(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"OpenAIBrowserAuthProviderTestProfile"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})

	settings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(`{"settings":{"openai.active_auth_method":"browser_auth","openai.browser_auth_state":"connected","openai.browser_auth_artifact_present":"true"}}`), map[string]string{"Content-Type": "application/json"})
	if settings.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", settings.Code, settings.Body.String())
	}
	missingProof := doRequest(t, a, http.MethodPost, "/api/provider/test", strings.NewReader(`{"provider":"openai","profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if missingProof.Code != http.StatusBadRequest {
		t.Fatalf("expected missing browser proof status, got %d body=%s", missingProof.Code, missingProof.Body.String())
	}
	var missingPayload map[string]any
	if err := json.NewDecoder(missingProof.Body).Decode(&missingPayload); err != nil {
		t.Fatalf("decode missing browser proof payload: %v", err)
	}
	if missingPayload["status"] != "needs_config" || missingPayload["code"] != "OPENAI_BROWSER_AUTH_PROVIDER_TEST_PROOF_REQUIRED" || missingPayload["provider_test_passed"] != false || missingPayload["credential_present"] != true {
		t.Fatalf("expected setup-needed browser provider-test proof evidence, got %+v", missingPayload)
	}

	failedSettings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(`{"settings":{"openai.browser_auth_provider_test_state":"failed","openai.browser_auth_provider_test_artifact_id":"browser-proof-failed","openai.browser_auth_provider_test_message":"adapter rejected browser session"}}`), map[string]string{"Content-Type": "application/json"})
	if failedSettings.Code != http.StatusOK {
		t.Fatalf("failed proof settings status=%d body=%s", failedSettings.Code, failedSettings.Body.String())
	}
	failedProof := doRequest(t, a, http.MethodPost, "/api/provider/test", strings.NewReader(`{"provider":"openai","profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if failedProof.Code != http.StatusBadRequest {
		t.Fatalf("expected failed browser proof status, got %d body=%s", failedProof.Code, failedProof.Body.String())
	}
	var failedPayload map[string]any
	if err := json.NewDecoder(failedProof.Body).Decode(&failedPayload); err != nil {
		t.Fatalf("decode failed browser proof payload: %v", err)
	}
	if failedPayload["status"] != "failed" || failedPayload["code"] != "OPENAI_BROWSER_AUTH_PROVIDER_TEST_FAILED" || failedPayload["provider_test_passed"] != false || failedPayload["provider_test_artifact_id"] != "browser-proof-failed" {
		t.Fatalf("expected failed browser provider-test evidence, got %+v", failedPayload)
	}

	passedSettings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(`{"settings":{"openai.browser_auth_provider_test_state":"passed","openai.browser_auth_provider_test_artifact_id":"browser-proof-pass","openai.browser_auth_provider_test_message":"browser auth adapter proof accepted"}}`), map[string]string{"Content-Type": "application/json"})
	if passedSettings.Code != http.StatusOK {
		t.Fatalf("passed proof settings status=%d body=%s", passedSettings.Code, passedSettings.Body.String())
	}
	passedProof := doRequest(t, a, http.MethodPost, "/api/provider/test", strings.NewReader(`{"provider":"openai","profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if passedProof.Code != http.StatusOK {
		t.Fatalf("expected passed browser proof status, got %d body=%s", passedProof.Code, passedProof.Body.String())
	}
	var passedPayload map[string]any
	if err := json.NewDecoder(passedProof.Body).Decode(&passedPayload); err != nil {
		t.Fatalf("decode passed browser proof payload: %v", err)
	}
	if passedPayload["status"] != "ready" || passedPayload["code"] != "OPENAI_BROWSER_AUTH_PROVIDER_TEST_PASSED" || passedPayload["provider_test_passed"] != true || passedPayload["auth_method"] != "browser_auth" {
		t.Fatalf("expected ready browser provider-test evidence, got %+v", passedPayload)
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
	openai := findRegistryProvider(registryPayload.Providers, "openai")
	methods := openai["auth_methods"].(map[string]any)
	browser := methods["browser_auth"].(map[string]any)
	if openai["state"] != "ready" || browser["connected"] != true || browser["provider_test_passed"] != true || browser["provider_test_artifact_id"] != "browser-proof-pass" {
		t.Fatalf("expected registry to report browser auth ready only after passed proof, got provider=%+v browser=%+v", openai, browser)
	}
	if strings.Contains(passedProof.Body.String(), "sk-") {
		t.Fatalf("browser provider-test response must not expose secret-like material, body=%s", passedProof.Body.String())
	}
}

func TestEbayRegistryExposesSellerOperationCapabilityStatuses(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
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
	ebay := findRegistryProvider(payload.Providers, "ebay")
	if ebay == nil {
		t.Fatalf("expected ebay provider in registry payload: %+v", payload.Providers)
	}
	statuses, ok := ebay["seller_operations"].([]any)
	if !ok {
		t.Fatalf("expected seller_operations array, got %#v", ebay["seller_operations"])
	}
	if len(statuses) != 5 {
		t.Fatalf("expected five seller operation statuses, got %d: %+v", len(statuses), statuses)
	}
	for _, raw := range statuses {
		status, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("expected seller operation status object, got %#v", raw)
		}
		if status["read_available"].(bool) || status["write_available"].(bool) || status["confirmation_required"].(bool) {
			t.Fatalf("default registry status must not expose unverified seller operation availability: %+v", status)
		}
		if got := fmt.Sprintf("%v", status["blocker"]); got != "ebay_api_capability_not_verified" {
			t.Fatalf("expected capability-not-verified blocker, got %+v", status)
		}
	}
}

func TestTelegramRegistryExposesCaptureChannelSetupState(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"TelegramRegistryProfile"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})

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
	telegram := findRegistryProvider(payload.Providers, "telegram")
	if telegram == nil {
		t.Fatalf("expected telegram provider in registry payload: %+v", payload.Providers)
	}
	if got := fmt.Sprintf("%v", telegram["integration_mode"]); got != "assistant_capture_channel" {
		t.Fatalf("expected assistant capture channel mode, got %q", got)
	}
	if got := fmt.Sprintf("%v", telegram["state"]); got != "needs_config" {
		t.Fatalf("expected Telegram setup-needed state before sender/chat authorization, got %q", got)
	}
	capabilities, ok := telegram["capabilities"].(map[string]any)
	if !ok {
		t.Fatalf("expected capabilities map, got %#v", telegram["capabilities"])
	}
	if capabilities["assistant"] != true || capabilities["media_capture"] != true {
		t.Fatalf("expected Telegram assistant and media capture capabilities, got %+v", capabilities)
	}
	authMethods, ok := telegram["auth_methods"].(map[string]any)
	if !ok {
		t.Fatalf("expected auth_methods map, got %#v", telegram["auth_methods"])
	}
	senderChat, ok := authMethods["sender_chat"].(map[string]any)
	if !ok || senderChat["connected"].(bool) {
		t.Fatalf("expected sender/chat method to start disconnected, got %#v", authMethods["sender_chat"])
	}

	settings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(`{"settings":{"telegram.catalog_capture.sender_id":"12345","telegram.catalog_capture.chat_id":"-5235769556"}}`), map[string]string{"Content-Type": "application/json"})
	if settings.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", settings.Code, settings.Body.String())
	}
	registry = doRequest(t, a, http.MethodGet, "/api/providers/registry", nil, nil)
	if registry.Code != http.StatusOK {
		t.Fatalf("registry after settings status=%d body=%s", registry.Code, registry.Body.String())
	}
	payload = struct {
		Providers []map[string]any `json:"providers"`
	}{}
	if err := json.NewDecoder(registry.Body).Decode(&payload); err != nil {
		t.Fatalf("decode registry payload after settings: %v", err)
	}
	telegram = findRegistryProvider(payload.Providers, "telegram")
	if got := fmt.Sprintf("%v", telegram["state"]); got != "ready" {
		t.Fatalf("expected Telegram ready state after sender/chat authorization, got %q", got)
	}
	authMethods = telegram["auth_methods"].(map[string]any)
	senderChat = authMethods["sender_chat"].(map[string]any)
	if connected, _ := senderChat["connected"].(bool); !connected {
		t.Fatalf("expected sender/chat method connected after settings, got %+v", senderChat)
	}
}

func findRegistryProvider(providers []map[string]any, id string) map[string]any {
	for _, provider := range providers {
		if fmt.Sprintf("%v", provider["provider_id"]) == id {
			return provider
		}
	}
	return nil
}

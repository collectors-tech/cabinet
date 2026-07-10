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
	setup, ok := openai["setup_status"].(map[string]any)
	if !ok {
		t.Fatalf("expected OpenAI setup_status map, got %#v", openai["setup_status"])
	}
	health, ok := openai["health"].(map[string]any)
	if !ok {
		t.Fatalf("expected OpenAI registry health map, got %#v", openai["health"])
	}
	actions, ok := openai["actions"].([]any)
	if !ok {
		t.Fatalf("expected OpenAI registry actions list, got %#v", openai["actions"])
	}
	for _, actionID := range []string{"assistant.chat", "assistant.image_help", "assistant.content_generation"} {
		action := findRegistryAction(actions, actionID)
		if action == nil {
			t.Fatalf("OpenAI registry missing assistant action %q in %+v", actionID, actions)
		}
		if action["availability_state"] != "setup_needed" || action["next_action"] != "connect_openai_api_key_or_browser_auth" {
			t.Fatalf("expected setup-needed assistant action %q, got %+v", actionID, action)
		}
	}
	if health["status"] != "needs_config" || health["state"] != "provider_setup_required" || health["next_action"] != "connect_openai_api_key_or_browser_auth" {
		t.Fatalf("expected setup-needed OpenAI registry health, got %+v", health)
	}
	for key, want := range map[string]string{
		"auth_mode":                   "hybrid",
		"active_auth_method":          "none",
		"api_key_state":               "missing",
		"browser_auth_state":          "setup_needed",
		"browser_auth_artifact_state": "missing",
		"validation_status":           "setup_needed",
		"assistant_default_provider":  "openai",
		"assistant_default_model":     "gpt-4o-mini",
		"workflow_state":              "provider_setup_required",
		"next_action":                 "connect_openai_api_key_or_browser_auth",
	} {
		if got := fmt.Sprintf("%v", setup[key]); got != want {
			t.Fatalf("setup_status[%s] got %q want %q; setup=%+v", key, got, want, setup)
		}
	}
}

func TestOpenAIRegistryProjectsAssistantMigrationContract(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"OpenAIAssistantRegistryMigration"}`), map[string]string{"Content-Type": "application/json"})
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
	settings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(`{"settings":{"openai.active_auth_method":"api_key","assistant_default_provider":"openai","assistant_default_model":"gpt-4.1-mini"}}`), map[string]string{"Content-Type": "application/json"})
	if settings.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", settings.Code, settings.Body.String())
	}
	secret := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/secrets", strings.NewReader(`{"key":"openai_api_key","value":"sk-registry-migration"}`), map[string]string{"Content-Type": "application/json"})
	if secret.Code != http.StatusOK {
		t.Fatalf("secret status=%d body=%s", secret.Code, secret.Body.String())
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
	if openai["state"] != "ready" || openai["provider_type"] != "assistant" || openai["provider_category"] != "chat/AI" || openai["config_schema_ref"] != "integrations/openai/auth" {
		t.Fatalf("OpenAI assistant provider registry metadata drifted: %+v", openai)
	}
	capabilities := openai["capabilities"].(map[string]any)
	for _, capability := range []string{"assistant", "image_help", "content_generation", "health"} {
		if capabilities[capability] != true {
			t.Fatalf("OpenAI capability %q missing from registry payload: %+v", capability, capabilities)
		}
	}
	workflowRefs := openai["workflow_refs"].([]any)
	for _, workflow := range []string{"assistant.chat", "assistant.image_help", "assistant.content_generation"} {
		if !containsAnyString(workflowRefs, workflow) {
			t.Fatalf("OpenAI registry missing assistant workflow %q in %+v", workflow, workflowRefs)
		}
	}
	actions := openai["actions"].([]any)
	for _, actionID := range []string{"assistant.chat", "assistant.image_help", "assistant.content_generation"} {
		action := findRegistryAction(actions, actionID)
		if action == nil {
			t.Fatalf("OpenAI registry missing assistant action %q in %+v", actionID, actions)
		}
		if action["workflow_ref"] != actionID || action["capability_category"] != "assistant" || action["execution_mode"] != "provider_workflow" {
			t.Fatalf("OpenAI assistant action metadata drifted for %q: %+v", actionID, action)
		}
		if action["availability_state"] != "available" || action["next_action"] != nil {
			t.Fatalf("expected ready assistant action %q, got %+v", actionID, action)
		}
	}
	setup := openai["setup_status"].(map[string]any)
	health := openai["health"].(map[string]any)
	if health["status"] != "ready" || health["state"] != "assistant_workflows_ready" || health["auth_method"] != "api_key" || health["next_action"] != "run_openai_test" {
		t.Fatalf("expected ready OpenAI registry health to mirror assistant setup readiness, got %+v", health)
	}
	for key, want := range map[string]string{
		"active_auth_method":         "api_key",
		"api_key_state":              "stored",
		"validation_status":          "ready",
		"assistant_default_provider": "openai",
		"assistant_default_model":    "gpt-4.1-mini",
		"workflow_state":             "assistant_workflows_ready",
		"next_action":                "run_openai_test",
	} {
		if got := fmt.Sprintf("%v", setup[key]); got != want {
			t.Fatalf("setup_status[%s] got %q want %q; setup=%+v", key, got, want, setup)
		}
	}
	if strings.Contains(registry.Body.String(), "sk-registry-migration") {
		t.Fatalf("OpenAI registry response must not leak API key: %s", registry.Body.String())
	}
}

func TestOpenAIRegistryExposesSchemaDrivenSetupFields(t *testing.T) {
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
	openai := findRegistryProvider(payload.Providers, "openai")
	if openai == nil {
		t.Fatalf("expected openai provider in registry payload: %+v", payload.Providers)
	}
	schema, ok := openai["setup_schema"].(map[string]any)
	if !ok {
		t.Fatalf("expected OpenAI setup_schema map, got %#v", openai["setup_schema"])
	}
	if schema["schema_ref"] != "integrations/openai/auth" || schema["persistence_scope"] != "active_profile" {
		t.Fatalf("OpenAI setup schema identity drifted: %+v", schema)
	}
	fields, ok := schema["fields"].([]any)
	if !ok {
		t.Fatalf("expected OpenAI setup schema fields list, got %#v", schema["fields"])
	}
	for _, want := range []struct {
		key         string
		fieldType   string
		persistence string
		writeOnly   bool
		required    bool
	}{
		{key: "openai.active_auth_method", fieldType: "select", persistence: "profile_settings", required: true},
		{key: "assistant_default_model", fieldType: "select", persistence: "profile_settings", required: true},
		{key: "openai_api_key", fieldType: "secret", persistence: "profile_secrets", writeOnly: true},
		{key: "openai.browser_auth_artifact_present", fieldType: "browser-auth-status", persistence: "profile_settings"},
	} {
		field := findSetupSchemaField(fields, want.key)
		if field == nil {
			t.Fatalf("OpenAI setup_schema missing field %q in %+v", want.key, fields)
		}
		if field["type"] != want.fieldType || field["persistence"] != want.persistence {
			t.Fatalf("OpenAI setup field %q metadata drifted: %+v", want.key, field)
		}
		if got, _ := field["write_only"].(bool); got != want.writeOnly {
			t.Fatalf("OpenAI setup field %q write_only got %v want %v: %+v", want.key, got, want.writeOnly, field)
		}
		if got, _ := field["required"].(bool); got != want.required {
			t.Fatalf("OpenAI setup field %q required got %v want %v: %+v", want.key, got, want.required, field)
		}
	}
	if strings.Contains(registry.Body.String(), "sk-") {
		t.Fatalf("OpenAI setup_schema must not expose API-key values: %s", registry.Body.String())
	}
}

func TestAssistantPlaceholderProvidersAreDisabledRegistryEntries(t *testing.T) {
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

	for _, providerID := range []string{"anthropic", "google"} {
		provider := findRegistryProvider(payload.Providers, providerID)
		if provider == nil {
			t.Fatalf("expected disabled assistant placeholder provider %q in registry payload: %+v", providerID, payload.Providers)
		}
		if provider["provider_category"] != "chat/AI" || provider["provider_type"] != "assistant" {
			t.Fatalf("expected %s to remain categorized as a disabled assistant provider, got %+v", providerID, provider)
		}
		if provider["state"] != "disabled" || provider["api_available"] != false || provider["active_mode"] != "disabled_placeholder" || provider["api_support_profile"] != "placeholder_disabled" {
			t.Fatalf("expected %s placeholder to be visibly disabled, got %+v", providerID, provider)
		}
		if provider["auth_requirement"] != "not_supported" || provider["auth_mode"] != "none" || provider["config_schema_ref"] != "integrations/assistant/placeholder" {
			t.Fatalf("expected %s placeholder setup metadata to block credential entry, got %+v", providerID, provider)
		}
		capabilities, ok := provider["capabilities"].(map[string]any)
		if !ok {
			t.Fatalf("expected %s capabilities map, got %#v", providerID, provider["capabilities"])
		}
		if capabilities["assistant"] != false || capabilities["image_help"] != false || capabilities["content_generation"] != false || capabilities["health"] != true {
			t.Fatalf("expected %s placeholder to expose disabled assistant capabilities and health metadata, got %+v", providerID, capabilities)
		}
		workflowRefs, ok := provider["workflow_refs"].([]any)
		if !ok || len(workflowRefs) != 0 {
			t.Fatalf("expected %s placeholder to expose no workflow refs until adapter support exists, got %#v", providerID, provider["workflow_refs"])
		}
		if strings.Contains(strings.ToLower(registry.Body.String()), providerID+"_api_key") {
			t.Fatalf("%s placeholder registry response must not advertise credential keys, body=%s", providerID, registry.Body.String())
		}
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

func TestEbayRegistryExposesSetupReadinessWithoutCredentialLeak(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"EbayRegistrySetupProfile"}`), map[string]string{"Content-Type": "application/json"})
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
	settings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(`{"settings":{"ebay_bearer_token":"secret-ebay-token","ebay_marketplace":"EBAY_AU","ebay_base_url":"https://api.ebay.example"}}`), map[string]string{"Content-Type": "application/json"})
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
	ebay := findRegistryProvider(payload.Providers, "ebay")
	if ebay == nil {
		t.Fatalf("expected ebay provider in registry payload: %+v", payload.Providers)
	}
	if fmt.Sprintf("%v", ebay["has_token"]) != "true" {
		t.Fatalf("expected ebay registry token presence flag, got %+v", ebay)
	}
	setup, ok := ebay["setup_status"].(map[string]any)
	if !ok {
		t.Fatalf("expected setup_status map, got %#v", ebay["setup_status"])
	}
	for key, want := range map[string]string{
		"auth_mode":         "api_key",
		"marketplace":       "EBAY_AU",
		"token_state":       "stored",
		"health_state":      "ready",
		"validation_status": "ready",
		"next_action":       "run_ebay_query_sets_from_market_watch",
	} {
		if got := fmt.Sprintf("%v", setup[key]); got != want {
			t.Fatalf("setup_status[%s] got %q want %q; setup=%+v", key, got, want, setup)
		}
	}
	if strings.Contains(registry.Body.String(), "secret-ebay-token") {
		t.Fatalf("registry response must not leak eBay bearer token: %s", registry.Body.String())
	}
}

func TestEbayRegistrySetupStatusReflectsDegradedProviderHealth(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"EbayRegistryDegradedProfile"}`), map[string]string{"Content-Type": "application/json"})
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
	settings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(`{"settings":{"ebay_bearer_token":"secret-ebay-token","ebay_marketplace":"EBAY_AU"}}`), map[string]string{"Content-Type": "application/json"})
	if settings.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", settings.Code, settings.Body.String())
	}
	if _, err := a.db.Exec(`INSERT INTO provider_health(provider, status, message, retry_after_seconds, updated_at) VALUES ('ebay', 'error', 'Browse rate limited', 120, CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("seed degraded eBay provider health: %v", err)
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
	ebay := findRegistryProvider(payload.Providers, "ebay")
	if ebay == nil {
		t.Fatalf("expected ebay provider in registry payload: %+v", payload.Providers)
	}
	setup, ok := ebay["setup_status"].(map[string]any)
	if !ok {
		t.Fatalf("expected setup_status map, got %#v", ebay["setup_status"])
	}
	health, ok := ebay["health"].(map[string]any)
	if !ok {
		t.Fatalf("expected health map, got %#v", ebay["health"])
	}
	for key, want := range map[string]string{
		"status":      "error",
		"state":       "degraded",
		"message":     "Browse rate limited",
		"last_error":  "Browse rate limited",
		"next_action": "check_provider_health_and_credentials",
	} {
		if got := fmt.Sprintf("%v", health[key]); got != want {
			t.Fatalf("health[%s] got %q want %q; health=%+v", key, got, want, health)
		}
	}
	if got, ok := health["retry_after_seconds"].(float64); !ok || int(got) != 120 {
		t.Fatalf("health retry_after_seconds got %v want 120; health=%+v", health["retry_after_seconds"], health)
	}
	if health["last_checked_at"] == nil {
		t.Fatalf("health missing last_checked_at for registry surfacing: %+v", health)
	}
	for key, want := range map[string]string{
		"token_state":       "stored",
		"marketplace":       "EBAY_AU",
		"validation_status": "degraded",
		"health_state":      "degraded",
		"next_action":       "check_provider_health_and_credentials",
	} {
		if got := fmt.Sprintf("%v", setup[key]); got != want {
			t.Fatalf("setup_status[%s] got %q want %q; setup=%+v", key, got, want, setup)
		}
	}
	if strings.Contains(registry.Body.String(), "secret-ebay-token") {
		t.Fatalf("registry response must not leak eBay bearer token: %s", registry.Body.String())
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
	if got := fmt.Sprintf("%v", telegram["state"]); got != "needs_config" {
		t.Fatalf("expected Telegram to stay setup-needed until bot/webhook proof exists, got %q", got)
	}
	authMethods = telegram["auth_methods"].(map[string]any)
	senderChat = authMethods["sender_chat"].(map[string]any)
	if connected, _ := senderChat["connected"].(bool); !connected {
		t.Fatalf("expected sender/chat method connected after settings, got %+v", senderChat)
	}
	setupStatus, ok := telegram["setup_status"].(map[string]any)
	if !ok {
		t.Fatalf("expected Telegram setup_status map, got %#v", telegram["setup_status"])
	}
	for key, want := range map[string]string{
		"sender_chat_state": "authorized",
		"bot_token_state":   "missing",
		"webhook_state":     "pending",
		"runtime_proof":     "pending_live_channel_check",
		"next_action":       "store_bot_token_secret",
	} {
		if got := fmt.Sprintf("%v", setupStatus[key]); got != want {
			t.Fatalf("setup_status[%s] got %q want %q; setup=%+v", key, got, want, setupStatus)
		}
	}

	health := doRequest(t, a, http.MethodGet, "/api/provider/health?provider=telegram", nil, nil)
	if health.Code != http.StatusOK {
		t.Fatalf("telegram health status=%d body=%s", health.Code, health.Body.String())
	}
	if !strings.Contains(health.Body.String(), `"credential_returned":false`) ||
		!strings.Contains(health.Body.String(), `"code":"TELEGRAM_BOT_TOKEN_REQUIRED"`) ||
		strings.Contains(strings.ToLower(health.Body.String()), "token-secret") {
		t.Fatalf("expected non-secret Telegram diagnostics requiring bot token, body=%s", health.Body.String())
	}

	secret := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/secrets", strings.NewReader(`{"key":"telegram_bot_token","value":"token-secret"}`), map[string]string{"Content-Type": "application/json"})
	if secret.Code != http.StatusOK {
		t.Fatalf("telegram secret status=%d body=%s", secret.Code, secret.Body.String())
	}
	settings = doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(`{"settings":{"telegram.catalog_capture.sender_id":"12345","telegram.catalog_capture.chat_id":"-5235769556","telegram.webhook_configured":"true"}}`), map[string]string{"Content-Type": "application/json"})
	if settings.Code != http.StatusOK {
		t.Fatalf("full settings status=%d body=%s", settings.Code, settings.Body.String())
	}
	registry = doRequest(t, a, http.MethodGet, "/api/providers/registry", nil, nil)
	if registry.Code != http.StatusOK {
		t.Fatalf("registry after full setup status=%d body=%s", registry.Code, registry.Body.String())
	}
	payload = struct {
		Providers []map[string]any `json:"providers"`
	}{}
	if err := json.NewDecoder(registry.Body).Decode(&payload); err != nil {
		t.Fatalf("decode registry payload after full setup: %v", err)
	}
	telegram = findRegistryProvider(payload.Providers, "telegram")
	if got := fmt.Sprintf("%v", telegram["state"]); got != "ready" {
		t.Fatalf("expected Telegram ready state after full non-secret setup proof, got %q", got)
	}
	authMethods = telegram["auth_methods"].(map[string]any)
	botToken, ok := authMethods["bot_token"].(map[string]any)
	if !ok || botToken["credential_present"] != true {
		t.Fatalf("expected Telegram bot token presence flag only, got %+v", authMethods["bot_token"])
	}
	if strings.Contains(registry.Body.String(), "token-secret") {
		t.Fatalf("registry response must not leak Telegram token material: %s", registry.Body.String())
	}
	health = doRequest(t, a, http.MethodGet, "/api/provider/health?provider=telegram", nil, nil)
	if health.Code != http.StatusOK {
		t.Fatalf("telegram health after full setup status=%d body=%s", health.Code, health.Body.String())
	}
	if !strings.Contains(health.Body.String(), `"code":"TELEGRAM_CHANNEL_READY"`) ||
		!strings.Contains(health.Body.String(), `"credential_returned":false`) ||
		strings.Contains(health.Body.String(), "token-secret") {
		t.Fatalf("expected ready non-secret Telegram health, body=%s", health.Body.String())
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

func findRegistryAction(actions []any, id string) map[string]any {
	for _, raw := range actions {
		action, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if fmt.Sprintf("%v", action["action_id"]) == id {
			return action
		}
	}
	return nil
}

func findSetupSchemaField(fields []any, key string) map[string]any {
	for _, raw := range fields {
		field, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if fmt.Sprintf("%v", field["key"]) == key {
			return field
		}
	}
	return nil
}

func containsAnyString(values []any, want string) bool {
	for _, value := range values {
		if fmt.Sprintf("%v", value) == want {
			return true
		}
	}
	return false
}

package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestIntegrationMigrationRegistryKeepsLegacyProvidersVisibleAndRedacted(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"IntegrationMigrationRegression"}`), map[string]string{"Content-Type": "application/json"})
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

	settings := doRequest(
		t,
		a,
		http.MethodPut,
		"/api/profiles/"+profile.ID+"/settings",
		strings.NewReader(`{"settings":{"ebay_bearer_token":"secret-ebay-migration-token","ebay_marketplace":"EBAY_AU","ebay_base_url":"https://api.ebay.example","integration.au_webshops.domains":"bonzaslotcars.com.au,legacy.example.test","integration.au_webshop_legacy_example_test.enabled":"true","integration.au_webshop_legacy_example_test.token":"secret-legacy-provider-token","telegram.catalog_capture.sender_id":"12345","telegram.catalog_capture.chat_id":"-100987654321"}}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if settings.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", settings.Code, settings.Body.String())
	}
	if _, err := a.db.Exec(`INSERT INTO app_state(key, value, updated_at) VALUES('provider.amazon.mode','disabled',CURRENT_TIMESTAMP) ON CONFLICT(key) DO UPDATE SET value='disabled', updated_at=CURRENT_TIMESTAMP`); err != nil {
		t.Fatalf("disable amazon migration mode: %v", err)
	}

	registry := doRequest(t, a, http.MethodGet, "/api/providers/registry", nil, nil)
	if registry.Code != http.StatusOK {
		t.Fatalf("registry status=%d body=%s", registry.Code, registry.Body.String())
	}
	raw := registry.Body.String()
	for _, secret := range []string{"secret-ebay-migration-token", "secret-legacy-provider-token"} {
		if strings.Contains(raw, secret) {
			t.Fatalf("registry response must not leak migrated credential %q: %s", secret, raw)
		}
	}

	var payload struct {
		Providers []map[string]any `json:"providers"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("decode registry payload: %v", err)
	}

	for _, id := range []string{
		"openai",
		"telegram",
		"ebay",
		"amazon",
		"au-webshop-bonzaslotcars-com-au",
		"au-webshop-legacy-example-test",
	} {
		if provider := findRegistryProvider(payload.Providers, id); provider == nil {
			t.Fatalf("migrated registry must keep provider id %q visible; providers=%+v", id, payload.Providers)
		}
	}

	ebay := findRegistryProvider(payload.Providers, "ebay")
	if ebay["has_token"] != true {
		t.Fatalf("expected eBay token presence flag without token value, got %+v", ebay)
	}
	setup, ok := ebay["setup_status"].(map[string]any)
	if !ok {
		t.Fatalf("expected eBay setup_status object, got %#v", ebay["setup_status"])
	}
	if got := fmt.Sprintf("%v", setup["token_state"]); got != "stored" {
		t.Fatalf("expected migrated eBay token_state=stored, got setup=%+v", setup)
	}

	telegram := findRegistryProvider(payload.Providers, "telegram")
	authMethods, ok := telegram["auth_methods"].(map[string]any)
	if !ok {
		t.Fatalf("expected Telegram auth methods after migration, got %+v", telegram)
	}
	senderChat, ok := authMethods["sender_chat"].(map[string]any)
	if !ok || senderChat["connected"] != true {
		t.Fatalf("expected migrated Telegram sender/chat authorization, got %+v", authMethods["sender_chat"])
	}

	amazon := findRegistryProvider(payload.Providers, "amazon")
	if got := fmt.Sprintf("%v", amazon["state"]); got != "disabled" {
		t.Fatalf("unsupported migrated Amazon provider must stay visible as disabled, got %+v", amazon)
	}
	if got := strings.TrimSpace(fmt.Sprintf("%v", amazon["setup_instructions"])); got == "" {
		t.Fatalf("disabled Amazon provider needs visible setup/unsupported guidance, got %+v", amazon)
	}

	legacy := findRegistryProvider(payload.Providers, "au-webshop-legacy-example-test")
	if got := fmt.Sprintf("%v", legacy["state"]); got != "ready" {
		t.Fatalf("legacy AU webshop provider should migrate into ready registry entry, got %+v", legacy)
	}
	if legacy["has_token"] != true {
		t.Fatalf("legacy provider should expose only token presence after migration, got %+v", legacy)
	}
}

func TestProviderRegistryProjectsCanonicalManifestCategories(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"ProviderManifestProjection"}`), map[string]string{"Content-Type": "application/json"})
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

	expectManifestProjection := map[string]struct {
		category  string
		typ       string
		schemaRef string
		workflow  string
	}{
		"openai":                          {category: "chat/AI", typ: "assistant", schemaRef: "integrations/openai/auth", workflow: "assistant.chat"},
		"telegram":                        {category: "notification", typ: "messaging", schemaRef: "integrations/telegram/channel", workflow: "telegram.agent_text"},
		"ebay":                            {category: "marketplace", typ: "marketplace", schemaRef: "integrations/ebay/setup", workflow: "market_watch.run"},
		"au-webshop-bonzaslotcars-com-au": {category: "storefront/source matcher", typ: "retailer", schemaRef: "integrations/au-webshop/setup", workflow: "provider.family_detect"},
	}
	for id, want := range expectManifestProjection {
		provider := findRegistryProvider(payload.Providers, id)
		if provider == nil {
			t.Fatalf("provider %q missing from registry payload: %+v", id, payload.Providers)
		}
		if got := fmt.Sprintf("%v", provider["provider_category"]); got != want.category {
			t.Fatalf("provider %s category got %q want %q: %+v", id, got, want.category, provider)
		}
		if got := fmt.Sprintf("%v", provider["provider_type"]); got != want.typ {
			t.Fatalf("provider %s type got %q want %q: %+v", id, got, want.typ, provider)
		}
		if got := fmt.Sprintf("%v", provider["config_schema_ref"]); got != want.schemaRef {
			t.Fatalf("provider %s schema ref got %q want %q: %+v", id, got, want.schemaRef, provider)
		}
		workflowRefs, ok := provider["workflow_refs"].([]any)
		if !ok {
			t.Fatalf("provider %s workflow_refs got %T: %+v", id, provider["workflow_refs"], provider)
		}
		foundWorkflow := false
		for _, ref := range workflowRefs {
			if fmt.Sprintf("%v", ref) == want.workflow {
				foundWorkflow = true
				break
			}
		}
		if !foundWorkflow {
			t.Fatalf("provider %s missing workflow %q in %+v", id, want.workflow, workflowRefs)
		}
	}
}

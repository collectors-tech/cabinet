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

func TestProviderRegistryProjectsMarketWatchProviderScopes(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"ProviderManifestMarketWatchProjection"}`), map[string]string{"Content-Type": "application/json"})
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

	for providerID, wantScope := range map[string]string{
		"ebay":                               "ebay",
		"amazon":                             "amazon",
		"au-webshop-bonzaslotcars-com-au":    "bonzaslotcars",
		"au-webshop-frontlinehobbies-com-au": "frontlinehobbies",
		"au-webshop-hobbytechtoys-com-au":    "hobbytechtoys",
		"au-webshop-voglers-com-au":          "voglers",
		"au-webshop-mrtoys-com-au":           "mrtoys",
	} {
		provider := findRegistryProvider(payload.Providers, providerID)
		if provider == nil {
			t.Fatalf("provider %q missing from registry payload: %+v", providerID, payload.Providers)
		}
		if got := fmt.Sprintf("%v", provider["market_watch_scope"]); got != wantScope {
			t.Fatalf("provider %s market_watch_scope got %q want %q: %+v", providerID, got, wantScope, provider)
		}
		workflowRefs, ok := provider["workflow_refs"].([]any)
		if !ok {
			t.Fatalf("provider %s workflow_refs got %T: %+v", providerID, provider["workflow_refs"], provider)
		}
		foundMarketWatchWorkflow := false
		for _, ref := range workflowRefs {
			if fmt.Sprintf("%v", ref) == "market_watch.run" {
				foundMarketWatchWorkflow = true
				break
			}
		}
		if !foundMarketWatchWorkflow {
			t.Fatalf("provider %s must advertise market_watch.run for UI provider projection: %+v", providerID, workflowRefs)
		}
	}
}

func TestBetaMarketWatchProviderRegistryFailsClosedWithoutLiveProof(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"BetaProviderFailClosed"}`), map[string]string{"Content-Type": "application/json"})
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

	ebay := findRegistryProvider(payload.Providers, "ebay")
	if ebay == nil {
		t.Fatalf("eBay provider missing from registry payload: %+v", payload.Providers)
	}
	if got := fmt.Sprintf("%v", ebay["state"]); got != "needs_config" {
		t.Fatalf("eBay must not appear ready without credentials/live proof, got state=%q provider=%+v", got, ebay)
	}
	if got := fmt.Sprintf("%v", ebay["beta_release_status"]); got != "setup_required" {
		t.Fatalf("eBay beta release status got %q want setup_required: %+v", got, ebay)
	}
	if got := fmt.Sprintf("%v", ebay["live_evidence_state"]); got != "missing_credentials" {
		t.Fatalf("eBay live evidence state got %q want missing_credentials: %+v", got, ebay)
	}
	assertRegistryActionAvailability(t, ebay, "market_watch.run", "setup_needed", "connect_ebay_credentials_and_run_live_provider_proof")

	amazon := findRegistryProvider(payload.Providers, "amazon")
	if amazon == nil {
		t.Fatalf("Amazon provider missing from registry payload: %+v", payload.Providers)
	}
	if got := fmt.Sprintf("%v", amazon["beta_release_status"]); got != "disabled_unsupported" {
		t.Fatalf("Amazon beta release status got %q want disabled_unsupported: %+v", got, amazon)
	}
	assertRegistryActionAvailability(t, amazon, "market_watch.run", "disabled", "choose_supported_beta_market_watch_provider")

	bonza := findRegistryProvider(payload.Providers, "au-webshop-bonzaslotcars-com-au")
	if bonza == nil {
		t.Fatalf("Bonza provider missing from registry payload: %+v", payload.Providers)
	}
	if got := fmt.Sprintf("%v", bonza["beta_release_status"]); got != "manual_url_capture_only" {
		t.Fatalf("Bonza beta release status got %q want manual_url_capture_only: %+v", got, bonza)
	}

	for _, id := range []string{
		"au-webshop-frontlinehobbies-com-au",
		"au-webshop-hobbytechtoys-com-au",
		"au-webshop-voglers-com-au",
		"au-webshop-mrtoys-com-au",
	} {
		provider := findRegistryProvider(payload.Providers, id)
		if provider == nil {
			t.Fatalf("provider %q missing from registry payload: %+v", id, payload.Providers)
		}
		if got := fmt.Sprintf("%v", provider["beta_release_status"]); got != "beta_limited" {
			t.Fatalf("provider %s beta release status got %q want beta_limited: %+v", id, got, provider)
		}
		if got := fmt.Sprintf("%v", provider["live_evidence_state"]); got != "public_provider_probe_required" {
			t.Fatalf("provider %s live evidence state got %q want public_provider_probe_required: %+v", id, got, provider)
		}
	}
}

func TestProviderRegistryProjectsWorkflowActionRegistryMetadata(t *testing.T) {
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
		t.Fatalf("expected eBay provider in registry payload: %+v", payload.Providers)
	}
	actions, ok := ebay["actions"].([]any)
	if !ok {
		t.Fatalf("expected eBay registry actions list, got %#v", ebay["actions"])
	}
	for _, want := range []struct {
		id                   string
		sideEffect           string
		confirmationRequired bool
		scheduleSupport      string
	}{
		{id: "market_watch.run", sideEffect: "preview_only", scheduleSupport: "manual_and_scheduled"},
		{id: "ebay.seller_operations", sideEffect: "write", confirmationRequired: true, scheduleSupport: "manual"},
		{id: "ebay.listing_lifecycle", sideEffect: "destructive", confirmationRequired: true, scheduleSupport: "manual"},
	} {
		action := findRegistryAction(actions, want.id)
		if action == nil {
			t.Fatalf("eBay registry missing workflow action %q in %+v", want.id, actions)
		}
		if action["workflow_ref"] != want.id || action["side_effect_level"] != want.sideEffect || action["classification"] != want.sideEffect {
			t.Fatalf("workflow action %q safety metadata drifted: %+v", want.id, action)
		}
		if got, _ := action["confirmation_required"].(bool); got != want.confirmationRequired {
			t.Fatalf("workflow action %q confirmation_required got %v want %v: %+v", want.id, got, want.confirmationRequired, action)
		}
		if action["schedule_support"] != want.scheduleSupport || action["health_impact"] == "" {
			t.Fatalf("workflow action %q missing schedule/health contract: %+v", want.id, action)
		}
		capabilities, ok := action["capabilities"].([]any)
		if !ok || len(capabilities) == 0 {
			t.Fatalf("workflow action %q must expose capability list, got %#v", want.id, action["capabilities"])
		}
		inboxEvents, ok := action["inbox_events"].([]any)
		if !ok || !containsAnyString(inboxEvents, "required_action") {
			t.Fatalf("workflow action %q must expose required-action Inbox event metadata, got %#v", want.id, action["inbox_events"])
		}
	}

	telegram := findRegistryProvider(payload.Providers, "telegram")
	if telegram == nil {
		t.Fatalf("expected Telegram provider in registry payload: %+v", payload.Providers)
	}
	telegramActions, ok := telegram["actions"].([]any)
	if !ok {
		t.Fatalf("expected Telegram registry actions list, got %#v", telegram["actions"])
	}
	capture := findRegistryAction(telegramActions, "telegram.catalog_capture")
	if capture == nil || capture["requires_auth"] != true || capture["requires_secrets"] != true || capture["execution_mode"] != "provider_workflow" {
		t.Fatalf("Telegram catalog capture workflow registry metadata drifted: %+v", capture)
	}
}

func assertRegistryActionAvailability(t *testing.T, provider map[string]any, actionID, availability, nextAction string) {
	t.Helper()

	actions, ok := provider["actions"].([]any)
	if !ok {
		t.Fatalf("provider %v actions got %T: %+v", provider["provider_id"], provider["actions"], provider)
	}
	for _, rawAction := range actions {
		action, ok := rawAction.(map[string]any)
		if !ok {
			continue
		}
		if fmt.Sprintf("%v", action["action_id"]) != actionID {
			continue
		}
		if got := fmt.Sprintf("%v", action["availability_state"]); got != availability {
			t.Fatalf("provider %v action %s availability got %q want %q: %+v", provider["provider_id"], actionID, got, availability, action)
		}
		if got := fmt.Sprintf("%v", action["next_action"]); got != nextAction {
			t.Fatalf("provider %v action %s next_action got %q want %q: %+v", provider["provider_id"], actionID, got, nextAction, action)
		}
		return
	}
	t.Fatalf("provider %v missing action %s: %+v", provider["provider_id"], actionID, actions)
}

func TestProviderRegistryProjectsConfigSchemaShapes(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"ProviderConfigSchemaShapes"}`), map[string]string{"Content-Type": "application/json"})
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
		strings.NewReader(`{"settings":{"ebay_bearer_token":"secret-ebay-schema-token","telegram.bot_token":"secret-telegram-schema-token"}}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if settings.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", settings.Code, settings.Body.String())
	}

	registry := doRequest(t, a, http.MethodGet, "/api/providers/registry", nil, nil)
	if registry.Code != http.StatusOK {
		t.Fatalf("registry status=%d body=%s", registry.Code, registry.Body.String())
	}
	raw := registry.Body.String()
	for _, secret := range []string{"secret-ebay-schema-token", "secret-telegram-schema-token"} {
		if strings.Contains(raw, secret) {
			t.Fatalf("registry setup schemas must not leak secret value %q: %s", secret, raw)
		}
	}

	var payload struct {
		Providers []map[string]any `json:"providers"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("decode registry payload: %v", err)
	}

	for providerID, want := range map[string]struct {
		schemaRef        string
		fieldKey         string
		fieldType        string
		fieldPersistence string
		writeOnly        bool
	}{
		"ebay":                            {schemaRef: "integrations/ebay/setup", fieldKey: "ebay_bearer_token", fieldType: "secret", fieldPersistence: "profile_secrets", writeOnly: true},
		"telegram":                        {schemaRef: "integrations/telegram/channel", fieldKey: "telegram.bot_token", fieldType: "secret", fieldPersistence: "profile_secrets", writeOnly: true},
		"au-webshop-bonzaslotcars-com-au": {schemaRef: "integrations/au-webshop/setup", fieldKey: "crawl_interval_minutes", fieldType: "number", fieldPersistence: "profile_settings"},
	} {
		provider := findRegistryProvider(payload.Providers, providerID)
		if provider == nil {
			t.Fatalf("provider %q missing from registry payload: %+v", providerID, payload.Providers)
		}
		schema, ok := provider["setup_schema"].(map[string]any)
		if !ok {
			t.Fatalf("provider %s missing setup_schema map: %+v", providerID, provider)
		}
		if got := fmt.Sprintf("%v", schema["schema_ref"]); got != want.schemaRef {
			t.Fatalf("provider %s schema_ref got %q want %q: %+v", providerID, got, want.schemaRef, schema)
		}
		if got := fmt.Sprintf("%v", schema["validate_action"]); got == "" {
			t.Fatalf("provider %s setup_schema must expose validate/test action: %+v", providerID, schema)
		}
		fields, ok := schema["fields"].([]any)
		if !ok || len(fields) == 0 {
			t.Fatalf("provider %s setup_schema fields got %#v", providerID, schema["fields"])
		}
		field := findSetupSchemaField(fields, want.fieldKey)
		if field == nil {
			t.Fatalf("provider %s setup_schema missing field %q in %+v", providerID, want.fieldKey, fields)
		}
		if got := fmt.Sprintf("%v", field["type"]); got != want.fieldType {
			t.Fatalf("provider %s field %s type got %q want %q: %+v", providerID, want.fieldKey, got, want.fieldType, field)
		}
		if got := fmt.Sprintf("%v", field["persistence"]); got != want.fieldPersistence {
			t.Fatalf("provider %s field %s persistence got %q want %q: %+v", providerID, want.fieldKey, got, want.fieldPersistence, field)
		}
		if got, _ := field["write_only"].(bool); got != want.writeOnly {
			t.Fatalf("provider %s field %s write_only got %v want %v: %+v", providerID, want.fieldKey, got, want.writeOnly, field)
		}
	}
}

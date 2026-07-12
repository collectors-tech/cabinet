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
		"au-webshop-acercmodels-com":         "acercmodels",
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

func TestMarketplaceProvidersExposeIssue1480RegistryContract(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	if _, err := a.db.Exec(`INSERT INTO app_state(key, value, updated_at) VALUES('provider.amazon.mode','program_api',CURRENT_TIMESTAMP) ON CONFLICT(key) DO UPDATE SET value='program_api', updated_at=CURRENT_TIMESTAMP`); err != nil {
		t.Fatalf("enable amazon program API mode: %v", err)
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

	for providerID, want := range map[string]struct {
		displayName     string
		authMode        string
		configSchemaRef string
		workflows       []string
		capabilities    []string
	}{
		"ebay": {
			displayName:     "eBay",
			authMode:        "api_key",
			configSchemaRef: "integrations/ebay/setup",
			workflows:       []string{"market_watch.run", "ebay.buyer_interest", "ebay.seller_operations", "ebay.listing_lifecycle"},
			capabilities:    []string{"search", "import", "scanner_source_matching", "pricing", "price_checks", "order_reconciliation", "purchase_reconciliation", "listing_lookup", "seller_operations", "listing_lifecycle", "health"},
		},
		"amazon": {
			displayName:     "Amazon",
			authMode:        "hybrid",
			configSchemaRef: "integrations/amazon/setup",
			workflows:       []string{"market_watch.run"},
			capabilities:    []string{"search", "import", "scanner_source_matching", "pricing", "price_checks", "order_reconciliation", "purchase_reconciliation", "listing_lookup", "health"},
		},
	} {
		provider := findRegistryProvider(payload.Providers, providerID)
		if provider == nil {
			t.Fatalf("marketplace provider %q missing from registry payload: %+v", providerID, payload.Providers)
		}
		for field, value := range map[string]string{
			"display_name":       want.displayName,
			"provider_category":  "marketplace",
			"provider_type":      "marketplace",
			"auth_mode":          want.authMode,
			"config_schema_ref":  want.configSchemaRef,
			"market_watch_scope": providerID,
		} {
			if got := fmt.Sprintf("%v", provider[field]); got != value {
				t.Fatalf("provider %s field %s got %q want %q: %+v", providerID, field, got, value, provider)
			}
		}
		setupSchema, ok := provider["setup_schema"].(map[string]any)
		if !ok {
			t.Fatalf("provider %s missing setup schema payload: %+v", providerID, provider)
		}
		if got := fmt.Sprintf("%v", setupSchema["schema_ref"]); got != want.configSchemaRef {
			t.Fatalf("provider %s setup schema ref got %q want %q: %+v", providerID, got, want.configSchemaRef, setupSchema)
		}
		workflowRefs := anySlice(provider["workflow_refs"])
		for _, workflow := range want.workflows {
			if !containsAnyString(workflowRefs, workflow) {
				t.Fatalf("provider %s missing workflow ref %q in %+v", providerID, workflow, workflowRefs)
			}
		}
		actions, ok := provider["actions"].([]any)
		if !ok {
			t.Fatalf("provider %s missing action metadata list: %+v", providerID, provider)
		}
		for _, workflow := range want.workflows {
			if action := findRegistryAction(actions, workflow); action == nil {
				t.Fatalf("provider %s missing action metadata for %q in %+v", providerID, workflow, actions)
			}
		}
		capabilities, ok := provider["capabilities"].(map[string]any)
		if !ok {
			t.Fatalf("provider %s capabilities got %T: %+v", providerID, provider["capabilities"], provider)
		}
		for _, capability := range want.capabilities {
			if capabilities[capability] != true {
				t.Fatalf("provider %s capability %s must be true for #1480 migration: %+v", providerID, capability, capabilities)
			}
		}
	}
}

func TestAcerLightspeedProviderRegistryMetadata(t *testing.T) {
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

	acer := findRegistryProvider(payload.Providers, "au-webshop-acercmodels-com")
	if acer == nil {
		t.Fatalf("Acer Lightspeed provider missing from registry payload: %+v", payload.Providers)
	}
	for field, want := range map[string]string{
		"base_domain":         "acercmodels.com",
		"market_watch_scope":  "acercmodels",
		"provider_category":   "storefront/source matcher",
		"provider_type":       "retailer",
		"adapter_type":        "lightspeed-storefront",
		"api_family":          "lightspeed",
		"api_support_profile": "lightspeed_storefront_v1",
		"active_mode":         "lightspeed_catalog",
		"integration_mode":    "storefront_access",
	} {
		if got := fmt.Sprintf("%v", acer[field]); got != want {
			t.Fatalf("Acer registry field %s got %q want %q: %+v", field, got, want, acer)
		}
	}
	capabilities, ok := acer["capabilities"].(map[string]any)
	if !ok {
		t.Fatalf("Acer capabilities got %T: %+v", acer["capabilities"], acer)
	}
	for _, key := range []string{"search", "stock_observation", "pricing", "health"} {
		if capabilities[key] != true {
			t.Fatalf("Acer capability %s must be true for source matching: %+v", key, capabilities)
		}
	}
	actions, ok := acer["actions"].([]any)
	if !ok {
		t.Fatalf("Acer actions got %T: %+v", acer["actions"], acer)
	}
	foundMarketWatch := false
	for _, raw := range actions {
		action, ok := raw.(map[string]any)
		if !ok || fmt.Sprintf("%v", action["action_id"]) != "market_watch.run" {
			continue
		}
		foundMarketWatch = true
		if got := fmt.Sprintf("%v", action["availability_state"]); got != "available" {
			t.Fatalf("Acer market_watch.run availability got %q want available: %+v", got, action)
		}
	}
	if !foundMarketWatch {
		t.Fatalf("Acer registry missing market_watch.run action: %+v", actions)
	}
}

func TestHobbyShopProviderRegistryAdapterMatrix(t *testing.T) {
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

	for providerID, want := range map[string]struct {
		domain         string
		adapterType    string
		apiFamily      string
		supportProfile string
		activeMode     string
		scope          string
	}{
		"au-webshop-acercmodels-com":         {domain: "acercmodels.com", adapterType: "lightspeed-storefront", apiFamily: "lightspeed", supportProfile: "lightspeed_storefront_v1", activeMode: "lightspeed_catalog", scope: "acercmodels"},
		"au-webshop-andrewshobbies-com-au":   {domain: "andrewshobbies.com.au", adapterType: "shopify-storefront", apiFamily: "shopify", supportProfile: "shopify_storefront_candidate", activeMode: "shopify_storefront_catalog", scope: "andrewshobbies"},
		"au-webshop-frontlinehobbies-com-au": {domain: "frontlinehobbies.com.au", adapterType: "generic-structured-storefront", apiFamily: "algolia", supportProfile: "algolia_runtime_v1", activeMode: "algolia_runtime", scope: "frontlinehobbies"},
		"au-webshop-hobbyco-com-au":          {domain: "hobbyco.com.au", adapterType: "generic-storefront-crawler", apiFamily: "web_ingestion", supportProfile: "html_fallback", activeMode: "web_ingestion", scope: "hobbyco"},
		"au-webshop-hobbytechtoys-com-au":    {domain: "hobbytechtoys.com.au", adapterType: "shopify-boost-storefront", apiFamily: "boost_shopify", supportProfile: "boost_v2", activeMode: "boost_api", scope: "hobbytechtoys"},
		"au-webshop-metrohobbies-com-au":     {domain: "metrohobbies.com.au", adapterType: "shopify-storefront", apiFamily: "shopify", supportProfile: "shopify_storefront_candidate", activeMode: "shopify_storefront_catalog", scope: "metrohobbies"},
		"au-webshop-mrtoys-com-au":           {domain: "mrtoys.com.au", adapterType: "generic-storefront-crawler", apiFamily: "doofinder", supportProfile: "doofinder_hashid_v1", activeMode: "hashid_search", scope: "mrtoys"},
		"au-webshop-voglers-com-au":          {domain: "voglers.com.au", adapterType: "bigcommerce-storefront", apiFamily: "bigcommerce", supportProfile: "bigcommerce_storefront_v1", activeMode: "storefront_public", scope: "voglers"},
		"au-webshop-bonzaslotcars-com-au":    {domain: "bonzaslotcars.com.au", adapterType: "woocommerce-store-api", apiFamily: "woo_store_api", supportProfile: "store_v1", activeMode: "store_api_first", scope: "bonzaslotcars"},
	} {
		provider := findRegistryProvider(payload.Providers, providerID)
		if provider == nil {
			t.Fatalf("provider %q missing from registry payload: %+v", providerID, payload.Providers)
		}
		for field, value := range map[string]string{
			"base_domain":         want.domain,
			"adapter_type":        want.adapterType,
			"api_family":          want.apiFamily,
			"api_support_profile": want.supportProfile,
			"active_mode":         want.activeMode,
			"market_watch_scope":  want.scope,
			"auth_mode":           "none",
		} {
			if got := fmt.Sprintf("%v", provider[field]); got != value {
				t.Fatalf("provider %s field %s got %q want %q: %+v", providerID, field, got, value, provider)
			}
		}
		capabilities, ok := provider["capabilities"].(map[string]any)
		if !ok {
			t.Fatalf("provider %s capabilities got %T: %+v", providerID, provider["capabilities"], provider)
		}
		for _, key := range []string{"search", "stock_observation", "pricing", "health"} {
			if capabilities[key] != true {
				t.Fatalf("provider %s capability %s must be true for source matching: %+v", providerID, key, capabilities)
			}
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

func TestHobbytechRegistryExposesPartsFinderDiscoveryContract(t *testing.T) {
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

	hobbytech := findRegistryProvider(payload.Providers, "au-webshop-hobbytechtoys-com-au")
	if hobbytech == nil {
		t.Fatalf("expected Hobbytech provider in registry payload: %+v", payload.Providers)
	}
	for field, want := range map[string]string{
		"adapter_type":        "shopify-boost-storefront",
		"api_family":          "boost_shopify",
		"parts_finder_state":  "public_page_discovery",
		"parts_finder_path":   "/pages/parts-finder",
		"auth_mode":           "none",
		"integration_mode":    "web_ingestion",
		"market_watch_scope":  "hobbytechtoys",
		"api_support_profile": "boost_v2",
	} {
		if got := fmt.Sprintf("%v", hobbytech[field]); got != want {
			t.Fatalf("Hobbytech registry field %s got %q want %q: %+v", field, got, want, hobbytech)
		}
	}
	capabilities, ok := hobbytech["capabilities"].(map[string]any)
	if !ok {
		t.Fatalf("Hobbytech capabilities got %T: %+v", hobbytech["capabilities"], hobbytech)
	}
	for _, capability := range []string{"search", "pricing", "stock_observation", "health", "parts_finder"} {
		if capabilities[capability] != true {
			t.Fatalf("Hobbytech capability %s must be true for parts-finder source matching: %+v", capability, capabilities)
		}
	}
	discovery, ok := hobbytech["parts_finder_discovery"].(map[string]any)
	if !ok {
		t.Fatalf("Hobbytech parts_finder_discovery got %T: %+v", hobbytech["parts_finder_discovery"], hobbytech)
	}
	for field, want := range map[string]string{
		"platform":              "shopify_page_plus_boost_search",
		"robots_scope":          "public_page_allowed_search_query_disallowed",
		"safe_workflow":         "catalogue_source_matching_only",
		"manual_capture_action": "provider_product_url_ingest",
	} {
		if got := fmt.Sprintf("%v", discovery[field]); got != want {
			t.Fatalf("Hobbytech parts-finder discovery field %s got %q want %q: %+v", field, got, want, discovery)
		}
	}
	blocked, ok := discovery["blocked_actions"].([]any)
	if !ok {
		t.Fatalf("Hobbytech blocked_actions got %T: %+v", discovery["blocked_actions"], discovery)
	}
	for _, blockedAction := range []string{"login", "cart", "checkout", "payment", "purchase"} {
		if !containsAnyString(blocked, blockedAction) {
			t.Fatalf("Hobbytech blocked_actions missing %q: %+v", blockedAction, blocked)
		}
	}
	actions, ok := hobbytech["actions"].([]any)
	if !ok {
		t.Fatalf("Hobbytech actions got %T: %+v", hobbytech["actions"], hobbytech)
	}
	action := findRegistryAction(actions, "hobbytech.parts_finder")
	if action == nil {
		t.Fatalf("Hobbytech parts-finder action missing: %+v", actions)
	}
	for field, want := range map[string]string{
		"type":               "storefront_parts_finder",
		"side_effect_level":  "preview_only",
		"execution_mode":     "provider_workflow",
		"schedule_support":   "manual",
		"availability_state": "available",
	} {
		if got := fmt.Sprintf("%v", action[field]); got != want {
			t.Fatalf("Hobbytech parts-finder action field %s got %q want %q: %+v", field, got, want, action)
		}
	}
	if action["requires_auth"] != false || action["requires_secrets"] != false || action["confirmation_required"] != false {
		t.Fatalf("Hobbytech parts-finder action must stay credential-free and preview-only: %+v", action)
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

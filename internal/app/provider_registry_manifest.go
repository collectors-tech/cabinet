package app

import "strings"

type integrationProviderManifest struct {
	ProviderID        string
	DisplayName       string
	BaseDomain        string
	MarketWatchScope  string
	ProviderCategory  string
	ProviderType      string
	AdapterType       string
	APIFamily         string
	APISupportProfile string
	ActiveMode        string
	IntegrationMode   string
	APIAvailable      bool
	AuthRequirement   string
	AuthMode          string
	ConfigSchemaRef   string
	WorkflowRefs      []string
	CapabilityFlags   map[string]bool
	SetupInstructions string
}

type integrationConfigSchemaDefinition struct {
	SchemaRef        string
	PersistenceScope string
	SubmitTarget     string
	SecretTarget     string
	Fields           []integrationConfigSchemaField
	ValidateAction   string
}

type integrationConfigSchemaField struct {
	Key              string
	Label            string
	Type             string
	Required         bool
	WriteOnly        bool
	ReadOnly         bool
	Persistence      string
	SecretKey        string
	Default          any
	Placeholder      string
	HelperText       string
	ValidationRules  []string
	Options          []integrationConfigSchemaOption
	Condition        map[string]string
	DocumentationURL string
}

type integrationConfigSchemaOption struct {
	Value string
	Label string
}

type integrationWorkflowActionDefinition struct {
	ID                   string
	Label                string
	Description          string
	Type                 string
	InputSchema          string
	OutputSchema         string
	RequiresAuth         bool
	RequiresSecrets      bool
	Capabilities         []string
	SideEffectLevel      string
	ConfirmationRequired bool
	ScheduleSupport      string
	InboxEvents          []string
	HealthImpact         string
	ExecutionMode        string
}

func (m integrationProviderManifest) payload() map[string]any {
	workflowRefs := append([]string{}, m.WorkflowRefs...)
	payload := map[string]any{
		"provider_id":         m.ProviderID,
		"display_name":        m.DisplayName,
		"base_domain":         m.BaseDomain,
		"provider_category":   m.ProviderCategory,
		"provider_type":       m.ProviderType,
		"adapter_type":        m.AdapterType,
		"api_family":          m.APIFamily,
		"api_support_profile": m.APISupportProfile,
		"active_mode":         m.ActiveMode,
		"integration_mode":    m.IntegrationMode,
		"api_available":       m.APIAvailable,
		"auth_requirement":    m.AuthRequirement,
		"auth_mode":           m.AuthMode,
		"capabilities":        copyCapabilityFlags(m.CapabilityFlags),
		"workflow_refs":       workflowRefs,
		"actions":             workflowActionsForRefs(workflowRefs, "available", nil),
		"setup_instructions":  m.SetupInstructions,
	}
	if strings.TrimSpace(m.ConfigSchemaRef) != "" {
		payload["config_schema_ref"] = m.ConfigSchemaRef
	}
	if strings.TrimSpace(m.MarketWatchScope) != "" {
		payload["market_watch_scope"] = strings.TrimSpace(m.MarketWatchScope)
	}
	return payload
}

func providerConfigSchemaForRef(ref string) (map[string]any, bool) {
	definition, ok := integrationConfigSchemaDefinitions()[strings.TrimSpace(ref)]
	if !ok {
		return nil, false
	}
	return definition.payload(), true
}

func (d integrationConfigSchemaDefinition) payload() map[string]any {
	fields := make([]map[string]any, 0, len(d.Fields))
	for _, field := range d.Fields {
		fields = append(fields, field.payload())
	}
	payload := map[string]any{
		"schema_ref":        d.SchemaRef,
		"persistence_scope": d.PersistenceScope,
		"submit_target":     d.SubmitTarget,
		"fields":            fields,
	}
	if strings.TrimSpace(d.SecretTarget) != "" {
		payload["secret_target"] = d.SecretTarget
	}
	if strings.TrimSpace(d.ValidateAction) != "" {
		payload["validate_action"] = d.ValidateAction
	}
	return payload
}

func (f integrationConfigSchemaField) payload() map[string]any {
	payload := map[string]any{
		"key":         f.Key,
		"label":       f.Label,
		"type":        f.Type,
		"required":    f.Required,
		"write_only":  f.WriteOnly,
		"persistence": f.Persistence,
	}
	if f.ReadOnly {
		payload["read_only"] = true
	}
	if strings.TrimSpace(f.SecretKey) != "" {
		payload["secret_key"] = f.SecretKey
	}
	if f.Default != nil {
		payload["default"] = f.Default
	}
	if strings.TrimSpace(f.Placeholder) != "" {
		payload["placeholder"] = f.Placeholder
	}
	if strings.TrimSpace(f.HelperText) != "" {
		payload["helper_text"] = f.HelperText
	}
	if len(f.ValidationRules) > 0 {
		rules := append([]string{}, f.ValidationRules...)
		payload["validation_rules"] = rules
	}
	if len(f.Options) > 0 {
		options := make([]map[string]string, 0, len(f.Options))
		for _, option := range f.Options {
			options = append(options, map[string]string{"value": option.Value, "label": option.Label})
		}
		payload["options"] = options
	}
	if len(f.Condition) > 0 {
		condition := map[string]string{}
		for k, v := range f.Condition {
			condition[k] = v
		}
		payload["condition"] = condition
	}
	if strings.TrimSpace(f.DocumentationURL) != "" {
		payload["documentation_url"] = f.DocumentationURL
	}
	return payload
}

func integrationConfigSchemaDefinitions() map[string]integrationConfigSchemaDefinition {
	profileSettingsTarget := "/api/profiles/:profileId/settings"
	profileSecretsTarget := "/api/profiles/:profileId/secrets"
	return map[string]integrationConfigSchemaDefinition{
		"integrations/openai/auth": {
			SchemaRef: "integrations/openai/auth", PersistenceScope: "active_profile", SubmitTarget: profileSettingsTarget, SecretTarget: profileSecretsTarget, ValidateAction: "provider.test",
			Fields: []integrationConfigSchemaField{
				{Key: "openai.active_auth_method", Label: "Connection method", Type: "select", Required: true, Persistence: "profile_settings", Default: "api_key", Options: []integrationConfigSchemaOption{{Value: "api_key", Label: "API key"}, {Value: "browser_auth", Label: "Browser Auth"}}},
				{Key: "assistant_default_model", Label: "Default assistant model", Type: "select", Required: true, Persistence: "profile_settings", Default: "gpt-4o-mini", Options: []integrationConfigSchemaOption{{Value: "gpt-4o-mini", Label: "GPT-4o mini"}, {Value: "gpt-4.1-mini", Label: "GPT-4.1 mini"}, {Value: "gpt-5.3-codex", Label: "GPT-5.3 Codex"}}},
				{Key: "openai_api_key", Label: "API key", Type: "secret", WriteOnly: true, Persistence: "profile_secrets", SecretKey: "openai_api_key", Condition: map[string]string{"openai.active_auth_method": "api_key"}, HelperText: "Stored through the profile secrets path and never returned in registry payloads."},
				{Key: "openai.browser_auth_artifact_present", Label: "Browser Auth proof", Type: "browser-auth-status", ReadOnly: true, Persistence: "profile_settings", Condition: map[string]string{"openai.active_auth_method": "browser_auth"}, HelperText: "Cabinet requires verified auth artifact and provider-test proof before marking Browser Auth ready."},
			},
		},
		"integrations/telegram/channel": {
			SchemaRef: "integrations/telegram/channel", PersistenceScope: "active_profile", SubmitTarget: profileSettingsTarget, SecretTarget: profileSecretsTarget, ValidateAction: "provider.test",
			Fields: []integrationConfigSchemaField{
				{Key: "telegram.catalog_capture.sender_id", Label: "Sender ID", Type: "text", Required: true, Persistence: "profile_settings", Placeholder: "123456789", ValidationRules: []string{"numeric_string"}},
				{Key: "telegram.catalog_capture.chat_id", Label: "Chat ID", Type: "text", Required: true, Persistence: "profile_settings", Placeholder: "-1001234567890", ValidationRules: []string{"telegram_chat_id"}},
				{Key: "telegram.bot_token", Label: "Bot token", Type: "secret", Required: true, WriteOnly: true, Persistence: "profile_secrets", SecretKey: "telegram_bot_token"},
				{Key: "telegram.webhook_route", Label: "Webhook route", Type: "url", Required: false, Persistence: "profile_settings", ValidationRules: []string{"url"}},
			},
		},
		"integrations/ebay/setup": {
			SchemaRef: "integrations/ebay/setup", PersistenceScope: "active_profile", SubmitTarget: profileSettingsTarget, SecretTarget: profileSecretsTarget, ValidateAction: "provider.test",
			Fields: []integrationConfigSchemaField{
				{Key: "ebay_marketplace", Label: "Marketplace", Type: "select", Required: true, Persistence: "profile_settings", Default: "EBAY_AU", Options: []integrationConfigSchemaOption{{Value: "EBAY_AU", Label: "Australia"}, {Value: "EBAY_US", Label: "United States"}, {Value: "EBAY_GB", Label: "United Kingdom"}}},
				{Key: "ebay_base_url", Label: "API base URL", Type: "url", Required: false, Persistence: "profile_settings", Placeholder: "https://api.ebay.com", ValidationRules: []string{"url"}},
				{Key: "ebay_bearer_token", Label: "Bearer token", Type: "secret", Required: true, WriteOnly: true, Persistence: "profile_secrets", SecretKey: "ebay_bearer_token"},
			},
		},
		"integrations/amazon/setup": {
			SchemaRef: "integrations/amazon/setup", PersistenceScope: "active_profile", SubmitTarget: profileSettingsTarget, SecretTarget: profileSecretsTarget, ValidateAction: "provider.test",
			Fields: []integrationConfigSchemaField{
				{Key: "amazon_access_mode", Label: "Access mode", Type: "select", Required: true, Persistence: "profile_settings", Default: "program_api", Options: []integrationConfigSchemaOption{{Value: "program_api", Label: "Program API"}, {Value: "disabled", Label: "Disabled"}}},
				{Key: "amazon_partner_tag", Label: "Partner tag", Type: "text", Required: false, Persistence: "profile_settings"},
				{Key: "amazon_oauth_status", Label: "OAuth status", Type: "oauth-connect", Required: false, ReadOnly: true, Persistence: "profile_settings"},
			},
		},
		"integrations/au-webshop/setup": {
			SchemaRef: "integrations/au-webshop/setup", PersistenceScope: "active_profile", SubmitTarget: profileSettingsTarget, ValidateAction: "provider.family_detect",
			Fields: []integrationConfigSchemaField{
				{Key: "base_domain", Label: "Store domain", Type: "text", Required: true, ReadOnly: true, Persistence: "provider_manifest", ValidationRules: []string{"domain"}},
				{Key: "provider_family", Label: "Provider family", Type: "select", Required: false, Persistence: "profile_settings", Options: []integrationConfigSchemaOption{{Value: "auto", Label: "Auto-detect"}, {Value: "woo_store_api", Label: "Woo Store API"}, {Value: "bigcommerce", Label: "BigCommerce"}, {Value: "algolia", Label: "Algolia"}, {Value: "boost_shopify", Label: "Boost Shopify"}, {Value: "web_ingestion", Label: "Web ingestion"}}},
				{Key: "crawl_interval_minutes", Label: "Polling interval", Type: "number", Required: false, Persistence: "profile_settings", Default: 1440, ValidationRules: []string{"min:60", "max:10080"}},
			},
		},
		"integrations/assistant/placeholder": {
			SchemaRef: "integrations/assistant/placeholder", PersistenceScope: "none", SubmitTarget: "", Fields: []integrationConfigSchemaField{
				{Key: "adapter_status", Label: "Adapter status", Type: "text", ReadOnly: true, Persistence: "provider_manifest", Default: "not_supported"},
			},
		},
	}
}

func (d integrationWorkflowActionDefinition) payload(availability string, nextAction any) map[string]any {
	capabilities := append([]string{}, d.Capabilities...)
	inboxEvents := append([]string{}, d.InboxEvents...)
	capabilityCategory := ""
	if len(capabilities) > 0 {
		capabilityCategory = capabilities[0]
	}
	if strings.TrimSpace(availability) == "" {
		availability = "available"
	}
	return map[string]any{
		"action_id":             d.ID,
		"workflow_ref":          d.ID,
		"label":                 d.Label,
		"description":           d.Description,
		"type":                  d.Type,
		"input_schema":          d.InputSchema,
		"output_schema":         d.OutputSchema,
		"requires_auth":         d.RequiresAuth,
		"requires_secrets":      d.RequiresSecrets,
		"capabilities":          capabilities,
		"capability_category":   capabilityCategory,
		"execution_mode":        d.ExecutionMode,
		"classification":        d.SideEffectLevel,
		"side_effect_level":     d.SideEffectLevel,
		"confirmation_required": d.ConfirmationRequired,
		"schedule_support":      d.ScheduleSupport,
		"inbox_events":          inboxEvents,
		"health_impact":         d.HealthImpact,
		"availability_state":    availability,
		"next_action":           nextAction,
	}
}

func workflowActionsForRefs(refs []string, availability string, nextAction any) []map[string]any {
	actions := make([]map[string]any, 0, len(refs))
	for _, ref := range refs {
		definition, ok := integrationWorkflowActionDefinitions()[strings.TrimSpace(ref)]
		if !ok {
			continue
		}
		actions = append(actions, definition.payload(availability, nextAction))
	}
	return actions
}

func integrationWorkflowActionDefinitions() map[string]integrationWorkflowActionDefinition {
	commonInboxEvents := []string{"workflow_failed", "required_action"}
	return map[string]integrationWorkflowActionDefinition{
		"assistant.chat": {
			ID: "assistant.chat", Label: "Assistant chat", Description: "Run a provider-backed assistant chat turn with profile and thread context.", Type: "assistant_chat",
			InputSchema: "assistant.chat.request.v1", OutputSchema: "assistant.chat.response.v1", RequiresAuth: true, RequiresSecrets: true, Capabilities: []string{"assistant"},
			SideEffectLevel: "read_only", ScheduleSupport: "none", InboxEvents: commonInboxEvents, HealthImpact: "requires_ready_provider", ExecutionMode: "provider_workflow",
		},
		"assistant.image_help": {
			ID: "assistant.image_help", Label: "Image help", Description: "Prepare assistant image analysis or processing guidance without mutating media by default.", Type: "assistant_media",
			InputSchema: "assistant.image.request.v1", OutputSchema: "assistant.image.preview.v1", RequiresAuth: true, RequiresSecrets: true, Capabilities: []string{"assistant", "image_help"},
			SideEffectLevel: "preview_only", ScheduleSupport: "none", InboxEvents: commonInboxEvents, HealthImpact: "requires_ready_provider", ExecutionMode: "provider_workflow",
		},
		"assistant.content_generation": {
			ID: "assistant.content_generation", Label: "Content generation", Description: "Generate catalog or listing copy as a preview before any downstream write.", Type: "assistant_content",
			InputSchema: "assistant.content.request.v1", OutputSchema: "assistant.content.preview.v1", RequiresAuth: true, RequiresSecrets: true, Capabilities: []string{"assistant", "content_generation"},
			SideEffectLevel: "preview_only", ScheduleSupport: "none", InboxEvents: commonInboxEvents, HealthImpact: "requires_ready_provider", ExecutionMode: "provider_workflow",
		},
		"telegram.catalog_capture": {
			ID: "telegram.catalog_capture", Label: "Catalog capture", Description: "Accept Telegram media capture into a governed preview before catalog changes are applied.", Type: "notification_capture",
			InputSchema: "telegram.catalog_capture.request.v1", OutputSchema: "telegram.catalog_capture.preview.v1", RequiresAuth: true, RequiresSecrets: true, Capabilities: []string{"media_capture", "assistant"},
			SideEffectLevel: "preview_only", ConfirmationRequired: true, ScheduleSupport: "event_driven", InboxEvents: []string{"workflow_failed", "required_action", "confirmation_pending"}, HealthImpact: "requires_channel_authorization", ExecutionMode: "provider_workflow",
		},
		"telegram.agent_text": {
			ID: "telegram.agent_text", Label: "Agent text intake", Description: "Route authorised Telegram text into governed agent skill previews and callbacks.", Type: "notification_agent_text",
			InputSchema: "telegram.agent_text.request.v1", OutputSchema: "telegram.agent_text.result.v1", RequiresAuth: true, RequiresSecrets: true, Capabilities: []string{"text_capture", "assistant"},
			SideEffectLevel: "write", ConfirmationRequired: true, ScheduleSupport: "event_driven", InboxEvents: []string{"workflow_failed", "required_action", "confirmation_pending"}, HealthImpact: "requires_channel_authorization", ExecutionMode: "provider_workflow",
		},
		"market_watch.run": {
			ID: "market_watch.run", Label: "Run Market Watch", Description: "Fetch or ingest provider search results into Cabinet's reviewable result inbox.", Type: "market_watch_scan",
			InputSchema: "market_watch.run.request.v1", OutputSchema: "market_watch.run.result_inbox.v1", RequiresAuth: false, RequiresSecrets: false, Capabilities: []string{"search", "pricing", "stock_observation"},
			SideEffectLevel: "preview_only", ScheduleSupport: "manual_and_scheduled", InboxEvents: []string{"workflow_failed", "required_action", "result_inbox_updated"}, HealthImpact: "updates_provider_health", ExecutionMode: "provider_workflow",
		},
		"provider.family_detect": {
			ID: "provider.family_detect", Label: "Detect provider family", Description: "Inspect a storefront and classify the supported adapter family for setup guidance.", Type: "provider_diagnostics",
			InputSchema: "provider.family_detect.request.v1", OutputSchema: "provider.family_detect.result.v1", RequiresAuth: false, RequiresSecrets: false, Capabilities: []string{"health", "search"},
			SideEffectLevel: "read_only", ScheduleSupport: "manual", InboxEvents: commonInboxEvents, HealthImpact: "updates_provider_health", ExecutionMode: "local_workflow",
		},
		"ebay.buyer_interest": {
			ID: "ebay.buyer_interest", Label: "Buyer interest review", Description: "Review watched eBay candidates and buyer-interest handoff state before downstream action.", Type: "marketplace_review",
			InputSchema: "ebay.buyer_interest.request.v1", OutputSchema: "ebay.buyer_interest.result.v1", RequiresAuth: true, RequiresSecrets: true, Capabilities: []string{"search", "pricing"},
			SideEffectLevel: "preview_only", ScheduleSupport: "manual", InboxEvents: []string{"workflow_failed", "required_action", "result_inbox_updated"}, HealthImpact: "updates_provider_health", ExecutionMode: "provider_workflow",
		},
		"ebay.seller_operations": {
			ID: "ebay.seller_operations", Label: "Seller operations", Description: "Preview seller messages, orders, fulfilment, and offers with explicit confirmation gates for writes.", Type: "marketplace_seller_ops",
			InputSchema: "ebay.seller_operations.request.v1", OutputSchema: "ebay.seller_operations.preview.v1", RequiresAuth: true, RequiresSecrets: true, Capabilities: []string{"seller_operations"},
			SideEffectLevel: "write", ConfirmationRequired: true, ScheduleSupport: "manual", InboxEvents: []string{"workflow_failed", "required_action", "confirmation_pending"}, HealthImpact: "updates_provider_health", ExecutionMode: "provider_workflow",
		},
		"ebay.listing_lifecycle": {
			ID: "ebay.listing_lifecycle", Label: "Listing lifecycle", Description: "Draft, publish, revise, end, or relist marketplace listings behind preview and confirmation gates.", Type: "marketplace_listing_lifecycle",
			InputSchema: "ebay.listing_lifecycle.request.v1", OutputSchema: "ebay.listing_lifecycle.preview.v1", RequiresAuth: true, RequiresSecrets: true, Capabilities: []string{"listing_lifecycle"},
			SideEffectLevel: "destructive", ConfirmationRequired: true, ScheduleSupport: "manual", InboxEvents: []string{"workflow_failed", "required_action", "confirmation_pending"}, HealthImpact: "updates_provider_health", ExecutionMode: "provider_workflow",
		},
	}
}

func copyCapabilityFlags(flags map[string]bool) map[string]bool {
	out := make(map[string]bool, len(flags))
	for k, v := range flags {
		out[k] = v
	}
	return out
}

func coreIntegrationProviderManifests(amazonMode string) []integrationProviderManifest {
	return []integrationProviderManifest{
		{
			ProviderID:        "openai",
			DisplayName:       "OpenAI / ChatGPT",
			BaseDomain:        "platform.openai.com",
			MarketWatchScope:  "",
			ProviderCategory:  "chat/AI",
			ProviderType:      "assistant",
			APIFamily:         "ai_provider",
			APISupportProfile: "browser_auth_or_api_key",
			ActiveMode:        "browser_auth_or_api_key",
			IntegrationMode:   "assistant_workflows",
			APIAvailable:      true,
			AuthRequirement:   "browser_auth_or_api_key",
			AuthMode:          "hybrid",
			ConfigSchemaRef:   "integrations/openai/auth",
			WorkflowRefs:      []string{"assistant.chat", "assistant.image_help", "assistant.content_generation"},
			CapabilityFlags: map[string]bool{
				"search":             false,
				"stock_observation":  false,
				"pricing":            false,
				"health":             true,
				"assistant":          true,
				"image_help":         true,
				"content_generation": true,
			},
			SetupInstructions: "Configure OpenAI with Browser Auth or an API key. Browser Auth stays setup-needed until Cabinet verifies an auth artifact/callback; navigation alone is never connected proof.",
		},
		{
			ProviderID:        "anthropic",
			DisplayName:       "Anthropic / Claude",
			BaseDomain:        "console.anthropic.com",
			MarketWatchScope:  "",
			ProviderCategory:  "chat/AI",
			ProviderType:      "assistant",
			APIFamily:         "ai_provider",
			APISupportProfile: "placeholder_disabled",
			ActiveMode:        "disabled_placeholder",
			IntegrationMode:   "assistant_workflows_disabled",
			APIAvailable:      false,
			AuthRequirement:   "not_supported",
			AuthMode:          "none",
			ConfigSchemaRef:   "integrations/assistant/placeholder",
			WorkflowRefs:      []string{},
			CapabilityFlags: map[string]bool{
				"search":             false,
				"stock_observation":  false,
				"pricing":            false,
				"health":             true,
				"assistant":          false,
				"image_help":         false,
				"content_generation": false,
			},
			SetupInstructions: "Anthropic is present only as a disabled assistant runtime placeholder until Cabinet ships a supported provider adapter, setup schema, health check, and workflow mapping.",
		},
		{
			ProviderID:        "google",
			DisplayName:       "Google / Gemini",
			BaseDomain:        "ai.google.dev",
			MarketWatchScope:  "",
			ProviderCategory:  "chat/AI",
			ProviderType:      "assistant",
			APIFamily:         "ai_provider",
			APISupportProfile: "placeholder_disabled",
			ActiveMode:        "disabled_placeholder",
			IntegrationMode:   "assistant_workflows_disabled",
			APIAvailable:      false,
			AuthRequirement:   "not_supported",
			AuthMode:          "none",
			ConfigSchemaRef:   "integrations/assistant/placeholder",
			WorkflowRefs:      []string{},
			CapabilityFlags: map[string]bool{
				"search":             false,
				"stock_observation":  false,
				"pricing":            false,
				"health":             true,
				"assistant":          false,
				"image_help":         false,
				"content_generation": false,
			},
			SetupInstructions: "Google Gemini is present only as a disabled assistant runtime placeholder until Cabinet ships a supported provider adapter, setup schema, health check, and workflow mapping.",
		},
		{
			ProviderID:        "telegram",
			DisplayName:       "Telegram",
			BaseDomain:        "telegram.org",
			MarketWatchScope:  "",
			ProviderCategory:  "notification",
			ProviderType:      "messaging",
			APIFamily:         "messaging_channel",
			APISupportProfile: "bot_webhook_sender_chat_v1",
			ActiveMode:        "sender_chat_authorization",
			IntegrationMode:   "assistant_capture_channel",
			APIAvailable:      true,
			AuthRequirement:   "sender_chat_authorization",
			AuthMode:          "sender_chat",
			ConfigSchemaRef:   "integrations/telegram/channel",
			WorkflowRefs:      []string{"telegram.catalog_capture", "telegram.agent_text"},
			CapabilityFlags: map[string]bool{
				"search":        false,
				"health":        true,
				"assistant":     true,
				"media_capture": true,
				"text_capture":  true,
			},
			SetupInstructions: "Configure Telegram sender/chat authorization, bot token secret, and webhook routing proof before running governed preview-before-apply channel intake.",
		},
		{
			ProviderID:        "ebay",
			DisplayName:       "eBay",
			BaseDomain:        "ebay.com",
			MarketWatchScope:  "ebay",
			ProviderCategory:  "marketplace",
			ProviderType:      "marketplace",
			APIFamily:         "official_api",
			APISupportProfile: "rest_v1",
			ActiveMode:        "official_api",
			IntegrationMode:   "official_api",
			APIAvailable:      true,
			AuthRequirement:   "api_key",
			AuthMode:          "api_key",
			ConfigSchemaRef:   "integrations/ebay/setup",
			WorkflowRefs:      []string{"market_watch.run", "ebay.buyer_interest", "ebay.seller_operations", "ebay.listing_lifecycle"},
			CapabilityFlags: map[string]bool{
				"search":                  true,
				"import":                  true,
				"scanner_source_matching": true,
				"stock_observation":       false,
				"pricing":                 true,
				"price_checks":            true,
				"order_reconciliation":    true,
				"purchase_reconciliation": true,
				"listing_lookup":          true,
				"seller_operations":       true,
				"listing_lifecycle":       true,
				"health":                  true,
			},
			SetupInstructions: "Add eBay API token and marketplace, validate health, then run scanner query sets.",
		},
		{
			ProviderID:        "amazon",
			DisplayName:       "Amazon",
			BaseDomain:        "amazon.com",
			MarketWatchScope:  "amazon",
			ProviderCategory:  "marketplace",
			ProviderType:      "marketplace",
			APIFamily:         "official_api",
			APISupportProfile: "program_api_v1",
			ActiveMode:        map[bool]string{true: "program_api", false: "disabled"}[amazonMode == "program_api"],
			IntegrationMode:   amazonMode,
			APIAvailable:      amazonMode == "program_api",
			AuthRequirement:   "oauth",
			AuthMode:          "hybrid",
			ConfigSchemaRef:   "integrations/amazon/setup",
			WorkflowRefs:      []string{"market_watch.run"},
			CapabilityFlags: map[string]bool{
				"search":                  amazonMode == "program_api",
				"import":                  amazonMode == "program_api",
				"scanner_source_matching": amazonMode == "program_api",
				"stock_observation":       false,
				"pricing":                 amazonMode == "program_api",
				"price_checks":            amazonMode == "program_api",
				"order_reconciliation":    true,
				"purchase_reconciliation": true,
				"listing_lookup":          amazonMode == "program_api",
				"health":                  true,
			},
			SetupInstructions: "Configure Amazon credentials and eligibility mode before running provider scans.",
		},
	}
}

func auWebshopProviderManifest(domain string) integrationProviderManifest {
	apiFamily := "web_ingestion"
	activeMode := "web_ingestion"
	integrationMode := "web_ingestion"
	supportProfile := "html_fallback"
	adapterType := "generic-storefront-crawler"
	if domain == "voglers.com.au" {
		apiFamily = "bigcommerce"
		activeMode = "storefront_public"
		integrationMode = "storefront_access"
		supportProfile = "bigcommerce_storefront_v1"
		adapterType = "bigcommerce-storefront"
	}
	if domain == "mrtoys.com.au" {
		apiFamily = "doofinder"
		activeMode = "hashid_search"
		integrationMode = "api_family_search"
		supportProfile = "doofinder_hashid_v1"
		adapterType = "generic-storefront-crawler"
	}
	if domain == "bonzaslotcars.com.au" {
		apiFamily = "woo_store_api"
		activeMode = "store_api_first"
		supportProfile = "store_v1"
		adapterType = "woocommerce-store-api"
	}
	if domain == "frontlinehobbies.com.au" {
		apiFamily = "algolia"
		activeMode = "algolia_runtime"
		supportProfile = "algolia_runtime_v1"
		adapterType = "generic-structured-storefront"
	}
	if domain == "hobbytechtoys.com.au" {
		apiFamily = "boost_shopify"
		activeMode = "boost_api"
		supportProfile = "boost_v2"
		adapterType = "shopify-boost-storefront"
	}
	if domain == "andrewshobbies.com.au" || domain == "metrohobbies.com.au" {
		apiFamily = "shopify"
		activeMode = "shopify_storefront_catalog"
		integrationMode = "storefront_access"
		supportProfile = "shopify_storefront_candidate"
		adapterType = "shopify-storefront"
	}
	if domain == "acercmodels.com" {
		apiFamily = "lightspeed"
		activeMode = "lightspeed_catalog"
		integrationMode = "storefront_access"
		supportProfile = "lightspeed_storefront_v1"
		adapterType = "lightspeed-storefront"
	}
	return integrationProviderManifest{
		ProviderID:        "au-webshop-" + strings.ReplaceAll(domain, ".", "-"),
		DisplayName:       domain,
		BaseDomain:        domain,
		MarketWatchScope:  marketWatchScopeForAUWebshopDomain(domain),
		ProviderCategory:  "storefront/source matcher",
		ProviderType:      "retailer",
		AdapterType:       adapterType,
		APIFamily:         apiFamily,
		APISupportProfile: supportProfile,
		ActiveMode:        activeMode,
		IntegrationMode:   integrationMode,
		APIAvailable:      domain == "voglers.com.au" || domain == "mrtoys.com.au",
		AuthRequirement:   "none",
		AuthMode:          "none",
		ConfigSchemaRef:   "integrations/au-webshop/setup",
		WorkflowRefs:      []string{"market_watch.run", "provider.family_detect"},
		CapabilityFlags: map[string]bool{
			"search":            true,
			"stock_observation": true,
			"pricing":           true,
			"health":            true,
		},
		SetupInstructions: "Webshop ingestion uses crawl parsing and does not require API credentials.",
	}
}

func marketWatchScopeForAUWebshopDomain(domain string) string {
	switch normalizeProviderDomain(domain) {
	case "bonzaslotcars.com.au":
		return "bonzaslotcars"
	case "frontlinehobbies.com.au":
		return "frontlinehobbies"
	case "hobbytechtoys.com.au":
		return "hobbytechtoys"
	case "andrewshobbies.com.au":
		return "andrewshobbies"
	case "voglers.com.au":
		return "voglers"
	case "acercmodels.com":
		return "acercmodels"
	case "mrtoys.com.au":
		return "mrtoys"
	case "hobbyco.com.au":
		return "hobbyco"
	case "metrohobbies.com.au":
		return "metrohobbies"
	default:
		return normalizeProviderDomain(domain)
	}
}

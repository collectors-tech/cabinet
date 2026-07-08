package app

import "strings"

type integrationProviderManifest struct {
	ProviderID        string
	DisplayName       string
	BaseDomain        string
	ProviderCategory  string
	ProviderType      string
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

func (m integrationProviderManifest) payload() map[string]any {
	payload := map[string]any{
		"provider_id":         m.ProviderID,
		"display_name":        m.DisplayName,
		"base_domain":         m.BaseDomain,
		"provider_category":   m.ProviderCategory,
		"provider_type":       m.ProviderType,
		"api_family":          m.APIFamily,
		"api_support_profile": m.APISupportProfile,
		"active_mode":         m.ActiveMode,
		"integration_mode":    m.IntegrationMode,
		"api_available":       m.APIAvailable,
		"auth_requirement":    m.AuthRequirement,
		"auth_mode":           m.AuthMode,
		"capabilities":        copyCapabilityFlags(m.CapabilityFlags),
		"workflow_refs":       append([]string(nil), m.WorkflowRefs...),
		"setup_instructions":  m.SetupInstructions,
	}
	if strings.TrimSpace(m.ConfigSchemaRef) != "" {
		payload["config_schema_ref"] = m.ConfigSchemaRef
	}
	return payload
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
			ProviderID:        "telegram",
			DisplayName:       "Telegram",
			BaseDomain:        "telegram.org",
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
				"search":            true,
				"stock_observation": false,
				"pricing":           true,
				"health":            true,
			},
			SetupInstructions: "Add eBay API token and marketplace, validate health, then run scanner query sets.",
		},
		{
			ProviderID:        "amazon",
			DisplayName:       "Amazon",
			BaseDomain:        "amazon.com",
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
				"search":            amazonMode == "program_api",
				"stock_observation": false,
				"pricing":           amazonMode == "program_api",
				"health":            true,
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
	if domain == "voglers.com.au" {
		apiFamily = "bigcommerce"
		activeMode = "storefront_public"
		integrationMode = "storefront_access"
		supportProfile = "bigcommerce_storefront_v1"
	}
	if domain == "mrtoys.com.au" {
		apiFamily = "doofinder"
		activeMode = "hashid_search"
		integrationMode = "api_family_search"
		supportProfile = "doofinder_hashid_v1"
	}
	if domain == "bonzaslotcars.com.au" {
		apiFamily = "woo_store_api"
		activeMode = "store_api_first"
		supportProfile = "store_v1"
	}
	if domain == "frontlinehobbies.com.au" {
		apiFamily = "algolia"
		activeMode = "algolia_runtime"
		supportProfile = "algolia_runtime_v1"
	}
	if domain == "hobbytechtoys.com.au" {
		apiFamily = "boost_shopify"
		activeMode = "boost_api"
		supportProfile = "boost_v2"
	}
	return integrationProviderManifest{
		ProviderID:        "au-webshop-" + strings.ReplaceAll(domain, ".", "-"),
		DisplayName:       domain,
		BaseDomain:        domain,
		ProviderCategory:  "storefront/source matcher",
		ProviderType:      "retailer",
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

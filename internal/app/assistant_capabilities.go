package app

import "github.com/collectors-tech/cabinet/internal/chat"

type assistantCapability struct {
	ID               string                      `json:"id"`
	Name             string                      `json:"name"`
	Description      string                      `json:"description"`
	Group            string                      `json:"group"`
	Mode             string                      `json:"mode"`
	PermissionState  string                      `json:"permission_state"`
	Requires         []string                    `json:"requires"`
	ProviderRequires []string                    `json:"provider_requires,omitempty"`
	InputSchema      string                      `json:"input_schema"`
	PreviewShape     string                      `json:"preview_shape"`
	ApplyBehavior    string                      `json:"apply_behavior"`
	AuditBehavior    string                      `json:"audit_behavior"`
	ResultLink       string                      `json:"result_link"`
	Targets          []assistantCapabilityTarget `json:"targets,omitempty"`
	Unavailable      bool                        `json:"unavailable"`
}

type assistantCapabilityTarget struct {
	ID      string   `json:"id"`
	Label   string   `json:"label"`
	Route   string   `json:"route"`
	Aliases []string `json:"aliases,omitempty"`
}

func assistantCapabilityRegistry() []assistantCapability {
	execution := assistantExecutionCapabilitiesByID()
	return []assistantCapability{
		withExecutionMetadata(assistantCapability{ID: "navigate.open_surface", Name: "Open Cabinet surface", Description: "Navigate the active Cabinet workspace to a known safe surface such as Media without mutating records.", Group: "app-control", Requires: []string{"profile", "workspace", "thread", "known_surface"}, Targets: assistantOpenSurfaceTargets()}, execution),
		withExecutionMetadata(assistantCapability{ID: "update_open_item_title", Name: "Update open item title", Description: "Preview a title edit for the currently open inventory item before applying the confirmed mutation.", Group: "app-control", Requires: []string{"profile", "workspace", "thread", "open_item"}}, execution),
		withExecutionMetadata(assistantCapability{ID: "inventory.item.create", Name: "Create inventory item", Description: "Draft a new catalog item from chat context, attachments, or structured user input.", Group: "inventory", Requires: []string{"profile", "workspace", "thread"}}, execution),
		withExecutionMetadata(assistantCapability{ID: "inventory.item.update", Name: "Update inventory item", Description: "Preview edits to an existing item before applying user-confirmed changes.", Group: "inventory", Requires: []string{"profile", "workspace", "thread", "selected_item"}}, execution),
		withExecutionMetadata(assistantCapability{ID: "collections.item.assign", Name: "Assign item to collection", Description: "Prepare a collection assignment preview and require confirmation before collection membership changes.", Group: "collections", Requires: []string{"profile", "workspace", "thread", "selected_item"}}, execution),
		withExecutionMetadata(assistantCapability{ID: "wishlist.entry.create", Name: "Create wishlist entry", Description: "Draft a wishlist entry and require confirmation before persistence.", Group: "wishlist", Requires: []string{"profile", "workspace", "thread"}}, execution),
		withExecutionMetadata(assistantCapability{ID: "settings.profile.read", Name: "Read profile settings", Description: "Summarize active profile and provider defaults without changing settings.", Group: "settings", Requires: []string{"profile", "workspace", "thread"}}, execution),
		withExecutionMetadata(assistantCapability{ID: "data.import.dry-run", Name: "Dry-run data import", Description: "Validate an import payload and present effects before any apply operation is allowed.", Group: "data", Requires: []string{"profile", "workspace", "thread", "import_file"}}, execution),
		withExecutionMetadata(assistantCapability{ID: "integrations.provider.run", Name: "Run integration provider", Description: "Expose provider execution as setup-needed until credentials and provider health are verified.", Group: "integrations", Requires: []string{"profile", "workspace", "thread", "connected_provider"}}, execution),
		{ID: "image_analyze", Name: "Analyze image evidence", Description: "Read image contents and metadata from approved Cabinet media to return source-linked findings without mutating media or item records.", Group: "media", Mode: "unavailable", PermissionState: "setup-needed", Requires: []string{"profile", "workspace", "thread", "media_id"}, ProviderRequires: []string{"openai", "verified_api_key_or_browser_auth", "provider_test_passed", "media_read_access"}, InputSchema: "assistant.image_analyze.v1", PreviewShape: "image_analysis_preview_with_sources", ApplyBehavior: "preview_only_no_mutation", AuditBehavior: "workflow_run_provider_trace_media_source_and_thread_message", ResultLink: "/media", Unavailable: true},
		{ID: "image_process", Name: "Process image variant", Description: "Prepare approved processed image variants while preserving original media, provenance, and reviewable source evidence.", Group: "media", Mode: "unavailable", PermissionState: "setup-needed", Requires: []string{"profile", "workspace", "thread", "media_id", "processing_intent"}, ProviderRequires: []string{"openai", "verified_api_key_or_browser_auth", "provider_test_passed", "media_write_access"}, InputSchema: "assistant.image_process.v1", PreviewShape: "image_process_variant_preview", ApplyBehavior: "requires_explicit_confirmation", AuditBehavior: "workflow_run_provider_trace_confirmation_original_media_and_variant_link", ResultLink: "/media", Unavailable: true},
		{ID: "content_generate", Name: "Generate catalog content", Description: "Draft catalog descriptions, condition notes, and enrichment copy from approved Cabinet item context.", Group: "assistant", Mode: "unavailable", PermissionState: "setup-needed", Requires: []string{"profile", "workspace", "thread", "approved_item_context"}, ProviderRequires: []string{"openai", "verified_api_key_or_browser_auth", "provider_test_passed"}, InputSchema: "assistant.content_generate.v1", PreviewShape: "catalog_content_draft_preview", ApplyBehavior: "preview_only_no_mutation", AuditBehavior: "workflow_run_provider_trace_and_thread_message", ResultLink: "/inventory", Unavailable: true},
		{ID: "listing_draft_generate", Name: "Generate listing draft", Description: "Create marketplace-ready listing draft content with provider constraints and source attribution.", Group: "assistant", Mode: "unavailable", PermissionState: "setup-needed", Requires: []string{"profile", "workspace", "thread", "approved_item_context", "target_marketplace"}, ProviderRequires: []string{"openai", "verified_api_key_or_browser_auth", "provider_test_passed"}, InputSchema: "assistant.listing_draft_generate.v1", PreviewShape: "listing_draft_preview_with_sources", ApplyBehavior: "requires_explicit_confirmation", AuditBehavior: "workflow_run_provider_trace_confirmation_and_result_link", ResultLink: "/integrations/ebay", Unavailable: true},
	}
}

func assistantExecutionCapabilitiesByID() map[string]chat.ActionCapability {
	out := map[string]chat.ActionCapability{}
	for _, capability := range chat.ActionCapabilityRegistry() {
		out[capability.ID] = capability
	}
	return out
}

func withExecutionMetadata(capability assistantCapability, execution map[string]chat.ActionCapability) assistantCapability {
	metadata, ok := execution[capability.ID]
	if !ok {
		return capability
	}
	capability.Mode = metadata.Mode
	capability.PermissionState = metadata.PermissionState
	capability.InputSchema = metadata.InputSchema
	capability.PreviewShape = metadata.PreviewShape
	capability.ApplyBehavior = metadata.ApplyBehavior
	capability.AuditBehavior = metadata.AuditBehavior
	capability.ResultLink = metadata.ResultLink
	capability.Unavailable = metadata.Unavailable
	return capability
}

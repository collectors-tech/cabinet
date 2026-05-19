package app

type assistantCapability struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Group           string   `json:"group"`
	Mode            string   `json:"mode"`
	PermissionState string   `json:"permission_state"`
	Requires        []string `json:"requires"`
	InputSchema     string   `json:"input_schema"`
	PreviewShape    string   `json:"preview_shape"`
	ApplyBehavior   string   `json:"apply_behavior"`
	AuditBehavior   string   `json:"audit_behavior"`
	ResultLink      string   `json:"result_link"`
	Unavailable     bool     `json:"unavailable"`
}

func assistantCapabilityRegistry() []assistantCapability {
	return []assistantCapability{
		{ID: "inventory.item.create", Name: "Create inventory item", Description: "Draft a new catalog item from chat context, attachments, or structured user input.", Group: "inventory", Mode: "confirm-required", PermissionState: "available", Requires: []string{"profile", "workspace", "thread"}, InputSchema: "inventory.item.create.v1", PreviewShape: "chat_action_preview", ApplyBehavior: "requires_explicit_confirmation", AuditBehavior: "thread_message_and_inbox_handoff", ResultLink: "/inventory"},
		{ID: "inventory.item.update", Name: "Update inventory item", Description: "Preview edits to an existing item before applying user-confirmed changes.", Group: "inventory", Mode: "confirm-required", PermissionState: "available", Requires: []string{"profile", "workspace", "thread", "selected_item"}, InputSchema: "inventory.item.update.v1", PreviewShape: "chat_action_preview", ApplyBehavior: "requires_explicit_confirmation", AuditBehavior: "thread_message_and_inbox_handoff", ResultLink: "/inventory"},
		{ID: "collections.item.assign", Name: "Assign item to collection", Description: "Prepare a collection assignment preview without mutating collection membership directly.", Group: "collections", Mode: "preview-only", PermissionState: "preview-only", Requires: []string{"profile", "workspace", "thread", "selected_item"}, InputSchema: "collections.item.assign.v1", PreviewShape: "chat_action_preview", ApplyBehavior: "manual_review_required", AuditBehavior: "thread_message_and_inbox_handoff", ResultLink: "/collections"},
		{ID: "wishlist.entry.create", Name: "Create wishlist entry", Description: "Draft a wishlist entry and require confirmation before persistence.", Group: "wishlist", Mode: "confirm-required", PermissionState: "available", Requires: []string{"profile", "workspace", "thread"}, InputSchema: "wishlist.entry.create.v1", PreviewShape: "chat_action_preview", ApplyBehavior: "requires_explicit_confirmation", AuditBehavior: "thread_message_and_inbox_handoff", ResultLink: "/wishlist"},
		{ID: "settings.profile.read", Name: "Read profile settings", Description: "Summarize active profile and provider defaults without changing settings.", Group: "settings", Mode: "read-only", PermissionState: "available", Requires: []string{"profile", "workspace", "thread"}, InputSchema: "settings.profile.read.v1", PreviewShape: "assistant_summary", ApplyBehavior: "not_applicable", AuditBehavior: "thread_message", ResultLink: "/settings"},
		{ID: "data.import.dry-run", Name: "Dry-run data import", Description: "Validate an import payload and present effects before any apply operation is allowed.", Group: "data", Mode: "preview-only", PermissionState: "preview-only", Requires: []string{"profile", "workspace", "thread", "import_file"}, InputSchema: "data.import.dry_run.v1", PreviewShape: "import_dry_run_summary", ApplyBehavior: "separate_confirmed_apply_required", AuditBehavior: "thread_message_and_inbox_handoff", ResultLink: "/settings/storage"},
		{ID: "integrations.provider.run", Name: "Run integration provider", Description: "Expose provider execution as setup-needed until credentials and provider health are verified.", Group: "integrations", Mode: "unavailable", PermissionState: "setup-needed", Requires: []string{"profile", "workspace", "thread", "connected_provider"}, InputSchema: "integrations.provider.run.v1", PreviewShape: "provider_run_preview", ApplyBehavior: "unavailable_until_provider_connected", AuditBehavior: "thread_message_and_inbox_handoff", ResultLink: "/integrations", Unavailable: true},
	}
}

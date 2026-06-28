package chat

import (
	"fmt"
	"strings"
)

type ActionCapability struct {
	ID              string
	Mode            string
	PermissionState string
	InputSchema     string
	PreviewShape    string
	ApplyBehavior   string
	AuditBehavior   string
	ResultLink      string
	Action          string
	ActionAliases   []string
	Unavailable     bool
	Guidance        string
}

func ActionCapabilityRegistry() []ActionCapability {
	return []ActionCapability{
		{ID: "inventory.item.create", Mode: "confirm-required", PermissionState: "available", InputSchema: "inventory.item.create.v1", PreviewShape: "chat_action_preview", ApplyBehavior: "requires_explicit_confirmation", AuditBehavior: "thread_message_and_inbox_handoff", ResultLink: "/inventory", Action: "create_inventory_item", ActionAliases: []string{"create_item_stub"}},
		{ID: "inventory.item.update", Mode: "confirm-required", PermissionState: "available", InputSchema: "inventory.item.update.v1", PreviewShape: "chat_action_preview", ApplyBehavior: "requires_explicit_confirmation", AuditBehavior: "thread_message_and_inbox_handoff", ResultLink: "/inventory", Action: "update_inventory_item"},
		{ID: "update_open_item_title", Mode: "confirm-required", PermissionState: "available", InputSchema: "agent.update_open_item_title.v1", PreviewShape: "chat_action_preview", ApplyBehavior: "requires_explicit_confirmation", AuditBehavior: "thread_message_and_inbox_handoff", ResultLink: "/inventory", Action: "update_open_item_title"},
		{ID: "wishlist.entry.create", Mode: "confirm-required", PermissionState: "available", InputSchema: "wishlist.entry.create.v1", PreviewShape: "chat_action_preview", ApplyBehavior: "requires_explicit_confirmation", AuditBehavior: "thread_message_and_inbox_handoff", ResultLink: "/wishlist", Action: "create_wishlist_entry"},
		{ID: "collections.item.assign", Mode: "confirm-required", PermissionState: "available", InputSchema: "collections.item.assign.v1", PreviewShape: "chat_action_preview", ApplyBehavior: "requires_explicit_confirmation", AuditBehavior: "thread_message_and_inbox_handoff", ResultLink: "/collections", Action: "assign_collection_item"},
		{ID: "navigate.open_surface", Mode: "preview-only", PermissionState: "available", InputSchema: "agent.navigate.open_surface.v1", PreviewShape: "route_navigation_preview", ApplyBehavior: "client_route_open_no_mutation", AuditBehavior: "workflow_run_and_thread_message", ResultLink: "/media", Guidance: "Route opening is handled by the app-control planner and does not use the mutation preview endpoint."},
		{ID: "settings.profile.read", Mode: "read-only", PermissionState: "available", InputSchema: "settings.profile.read.v1", PreviewShape: "assistant_summary", ApplyBehavior: "not_applicable", AuditBehavior: "thread_message", ResultLink: "/settings", Guidance: "Read-only capabilities do not use mutation preview/apply endpoints."},
		{ID: "data.import.dry-run", Mode: "preview-only", PermissionState: "preview-only", InputSchema: "data.import.dry_run.v1", PreviewShape: "import_dry_run_summary", ApplyBehavior: "separate_confirmed_apply_required", AuditBehavior: "thread_message_and_inbox_handoff", ResultLink: "/settings/storage", Guidance: "Data import dry-runs use their dedicated import preview endpoint before any confirmed apply."},
		{ID: "integrations.provider.run", Mode: "unavailable", PermissionState: "setup-needed", InputSchema: "integrations.provider.run.v1", PreviewShape: "provider_run_preview", ApplyBehavior: "unavailable_until_provider_connected", AuditBehavior: "thread_message_and_inbox_handoff", ResultLink: "/integrations", Unavailable: true, Guidance: "setup needed: connect and verify the provider before running this capability."},
	}
}

func resolveActionCapability(capabilityID, action string) (ActionCapability, error) {
	capabilityID = strings.TrimSpace(capabilityID)
	action = strings.TrimSpace(action)
	if capabilityID != "" {
		for _, capability := range ActionCapabilityRegistry() {
			if capability.ID == capabilityID {
				return validateExecutableCapability(capability)
			}
		}
		return ActionCapability{}, fmt.Errorf("unsupported capability: %s. Query /api/chat/capabilities for available Cabinet actions and setup guidance", capabilityID)
	}
	if action != "" {
		for _, capability := range ActionCapabilityRegistry() {
			if capability.Action == action {
				return validateExecutableCapability(capability)
			}
			for _, alias := range capability.ActionAliases {
				if alias == action {
					return validateExecutableCapability(capability)
				}
			}
		}
		return ActionCapability{}, fmt.Errorf("unsupported action: %s. Query /api/chat/capabilities for available Cabinet actions and setup guidance", action)
	}
	return ActionCapability{}, fmt.Errorf("capability_id or action is required")
}

func validateExecutableCapability(capability ActionCapability) (ActionCapability, error) {
	if capability.Unavailable || capability.Mode == "unavailable" {
		return ActionCapability{}, fmt.Errorf("capability %s is unavailable: %s", capability.ID, capability.Guidance)
	}
	if strings.TrimSpace(capability.Action) == "" {
		return ActionCapability{}, fmt.Errorf("capability %s is %s: %s", capability.ID, capability.Mode, capability.Guidance)
	}
	return capability, nil
}

func capabilityForAction(action string) ActionCapability {
	action = strings.TrimSpace(action)
	for _, capability := range ActionCapabilityRegistry() {
		if capability.Action == action {
			return capability
		}
		for _, alias := range capability.ActionAliases {
			if alias == action {
				return capability
			}
		}
	}
	return ActionCapability{}
}

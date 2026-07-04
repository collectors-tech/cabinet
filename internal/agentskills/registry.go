package agentskills

import (
	"errors"
	"slices"
	"strings"

	"github.com/collectors-tech/cabinet/internal/chat"
)

type SourceType string

const (
	SourceBuiltIn SourceType = "built-in"
	SourceArchive SourceType = "archive"
)

type Status string

const (
	StatusAvailable              Status = "available"
	StatusPreviewOnly            Status = "preview-only"
	StatusRequiresImplementation Status = "requires-implementation"
	StatusDisabled               Status = "disabled"
	StatusInvalid                Status = "invalid"
)

type SafetyLevel string

const (
	SafetyReadOnly        SafetyLevel = "read-only"
	SafetyPreviewOnly     SafetyLevel = "preview-only"
	SafetyConfirmRequired SafetyLevel = "confirm-required"
	SafetyDestructive     SafetyLevel = "destructive"
)

type PermissionDeclaration struct {
	LocalRead       bool `json:"local_read"`
	LocalWrite      bool `json:"local_write"`
	ExternalRead    bool `json:"external_read"`
	ExternalWrite   bool `json:"external_write"`
	SecretAccess    bool `json:"secret_access"`
	Destructive     bool `json:"destructive"`
	RequiresConfirm bool `json:"requires_confirm"`
}

type Skill struct {
	ID                   string                `json:"id"`
	Version              string                `json:"version"`
	DisplayName          string                `json:"display_name"`
	Description          string                `json:"description"`
	Category             string                `json:"category"`
	Source               SourceType            `json:"source"`
	Status               Status                `json:"status"`
	SafetyLevel          SafetyLevel           `json:"safety_level"`
	RequiredContext      []string              `json:"required_context"`
	RequiredActions      []string              `json:"required_actions,omitempty"`
	RequiredProviders    []string              `json:"required_providers,omitempty"`
	Capabilities         []string              `json:"capabilities,omitempty"`
	GuidedWorkflows      []string              `json:"guided_workflows,omitempty"`
	UITargets            []string              `json:"ui_targets,omitempty"`
	IntegrationWorkflows []string              `json:"integration_workflows,omitempty"`
	ShellCommands        []string              `json:"shell_commands,omitempty"`
	InputSchemaRefs      []string              `json:"input_schema_refs,omitempty"`
	OutputSchemaRefs     []string              `json:"output_schema_refs,omitempty"`
	Permissions          PermissionDeclaration `json:"permissions"`
	AuditBehavior        string                `json:"audit_behavior"`
	Provenance           string                `json:"provenance"`
	BuiltIn              bool                  `json:"built_in"`
	Removable            bool                  `json:"removable"`
	Enabled              bool                  `json:"enabled"`
	Executable           bool                  `json:"executable"`
	NextAction           string                `json:"next_action,omitempty"`
	ValidationWarnings   []string              `json:"validation_warnings,omitempty"`
	ValidationErrors     []string              `json:"validation_errors,omitempty"`
}

type Registry struct {
	builtIns []Skill
	imported []Skill
}

func NewRegistry(imported []Skill) Registry {
	return Registry{
		builtIns: builtInSkills(),
		imported: normalizeImported(imported),
	}
}

func (r Registry) List() []Skill {
	out := make([]Skill, 0, len(r.builtIns)+len(r.imported))
	out = append(out, r.builtIns...)
	builtInIDs := map[string]struct{}{}
	for _, skill := range r.builtIns {
		builtInIDs[skill.ID] = struct{}{}
	}
	for _, skill := range r.imported {
		if _, exists := builtInIDs[skill.ID]; exists {
			continue
		}
		out = append(out, deriveExecutionState(skill))
	}
	return out
}

func (r Registry) Resolve(id string) (Skill, bool) {
	id = strings.TrimSpace(id)
	for _, skill := range r.List() {
		if skill.ID == id {
			return skill, true
		}
	}
	return Skill{}, false
}

func (r Registry) ValidateImportedSkill(skill Skill) error {
	id := strings.TrimSpace(skill.ID)
	if id == "" {
		return errors.New("skill id is required")
	}
	for _, builtIn := range r.builtIns {
		if builtIn.ID == id {
			return errors.New("imported skill cannot override built-in skill id")
		}
	}
	return nil
}

func builtInSkills() []Skill {
	return []Skill{
		builtIn("cabinet.navigate.open_surface", "Open Cabinet surface", "Navigate to a known Cabinet surface without mutating records.", "navigation", SafetyPreviewOnly, []string{"profile", "workspace", "thread", "known_surface"}, []string{"navigate.open_surface"}, nil, nil),
		builtIn("cabinet.inventory.create_item", "Create inventory item", "Draft an inventory item and require explicit confirmation before persistence.", "inventory", SafetyConfirmRequired, []string{"profile", "workspace", "thread"}, []string{"inventory.item.create"}, nil, nil),
		builtIn("cabinet.inventory.update_item", "Update inventory item", "Preview edits to an existing inventory item before applying confirmed changes.", "inventory", SafetyConfirmRequired, []string{"profile", "workspace", "thread", "selected_item"}, []string{"inventory.item.update", "update_open_item_title"}, nil, nil),
		builtIn("cabinet.wishlist.create_entry", "Create wishlist entry", "Draft a wishlist entry and require confirmation before persistence.", "wishlist", SafetyConfirmRequired, []string{"profile", "workspace", "thread", "wanted_item_details"}, []string{"wishlist.entry.create"}, []string{"wishlist.entry.create"}, []string{"wishlist.create.button", "wishlist.entry.form", "wishlist.entry.save"}),
		builtIn("cabinet.collection.assign_item", "Assign item to collection", "Prepare a collection assignment preview before collection membership changes.", "collections", SafetyConfirmRequired, []string{"profile", "workspace", "thread", "selected_item", "collection"}, []string{"collections.item.assign"}, nil, nil),
		builtIn("cabinet.guided.inventory.update_item", "Guided inventory item update", "Guide an inventory item update through route focus, target highlight, preview, and confirmation.", "guided-workflows", SafetyConfirmRequired, []string{"profile", "thread", "target_inventory_item", "editable_field"}, []string{"inventory.item.update"}, []string{"inventory.item.update"}, []string{"inventory.item.row", "inventory.item.editor.title", "inventory.item.editor.save"}),
		builtIn("cabinet.chat.action_timeline.view", "View chat Action Timeline", "Read assistant workflow and action timeline evidence for the active thread.", "chat", SafetyReadOnly, []string{"profile", "thread"}, nil, nil, nil),
		builtIn("cabinet.inbox.search_notifications", "Search Inbox notifications", "Search and filter Inbox notifications without mutating review state.", "inbox", SafetyReadOnly, []string{"profile", "workspace"}, nil, nil, []string{"inbox.list", "inbox.search"}),
		builtIn("cabinet.inbox.summarise_unhandled", "Summarise unhandled Inbox", "Summarise unhandled Inbox items without marking them handled.", "inbox", SafetyReadOnly, []string{"profile", "workspace"}, nil, nil, []string{"inbox.summary", "inbox.unhandled"}),
		builtIn("cabinet.inbox.open_notification", "Open Inbox notification", "Open a selected Inbox notification and route to its related Cabinet surface.", "inbox", SafetyPreviewOnly, []string{"profile", "workspace", "selected_notification"}, []string{"navigate.open_surface"}, nil, []string{"inbox.notification.card", "inbox.notification.open"}),
		builtIn("cabinet.inbox.mark_handled", "Mark Inbox item handled", "Preview and confirm marking a selected Inbox item as handled.", "inbox", SafetyConfirmRequired, []string{"profile", "workspace", "selected_notification"}, nil, nil, []string{"inbox.notification.card", "inbox.notification.mark_handled"}),
		builtIn("cabinet.inbox.archive_or_hide", "Archive or hide Inbox item", "Preview and confirm archiving or hiding a selected Inbox item.", "inbox", SafetyConfirmRequired, []string{"profile", "workspace", "selected_notification"}, nil, nil, []string{"inbox.notification.card", "inbox.notification.archive"}),
		builtIn("cabinet.inbox.route_to_surface", "Route Inbox item to surface", "Open the Cabinet surface connected to an Inbox item without applying a state change.", "inbox", SafetyPreviewOnly, []string{"profile", "workspace", "selected_notification"}, []string{"navigate.open_surface"}, nil, []string{"inbox.notification.route", "app.surface.target"}),
		builtIn("cabinet.users.search", "Search users", "Search workspace users without changing roles, status, or invitations.", "users", SafetyReadOnly, []string{"profile", "workspace", "admin_session"}, nil, nil, []string{"users.table", "users.search"}),
		builtIn("cabinet.users.invite_user", "Invite user", "Prepare a user invitation draft that requires explicit confirmation before sending or persistence.", "users", SafetyConfirmRequired, []string{"profile", "workspace", "admin_session", "target_email", "target_role"}, nil, nil, []string{"users.invite.form", "users.invite.submit"}),
		builtIn("cabinet.users.resend_invitation", "Resend invitation", "Preview resending an invitation to an explicitly selected invited user.", "users", SafetyConfirmRequired, []string{"profile", "workspace", "admin_session", "target_user"}, nil, nil, []string{"users.row.invitation", "users.invite.resend"}),
		builtIn("cabinet.users.update_role", "Update user role", "Preview a role change with protected owner/admin safeguards before applying.", "users", SafetyConfirmRequired, []string{"profile", "workspace", "admin_session", "target_user", "target_role"}, nil, nil, []string{"users.row.role", "users.role.editor"}),
		builtIn("cabinet.users.activate_or_deactivate", "Activate or deactivate user", "Preview activation state changes with protected owner/admin safeguards before applying.", "users", SafetyConfirmRequired, []string{"profile", "workspace", "admin_session", "target_user", "target_status"}, nil, nil, []string{"users.row.status", "users.status.editor"}),
		builtIn("cabinet.users.remove_user", "Remove user", "Require destructive confirmation before removing a non-protected workspace user.", "users", SafetyDestructive, []string{"profile", "workspace", "admin_session", "target_user"}, nil, nil, []string{"users.row.remove", "users.remove.confirmation"}),
	}
}

func builtIn(id, displayName, description, category string, safety SafetyLevel, context, capabilities, guidedWorkflows, uiTargets []string) Skill {
	status := StatusAvailable
	nextAction := ""
	executable := true
	if id == "cabinet.guided.inventory.update_item" {
		status = StatusRequiresImplementation
		nextAction = "Complete and validate issue #1513 before advertising this guided skill as executable."
		executable = false
	}
	if strings.HasPrefix(id, "cabinet.inbox.mark_") || id == "cabinet.inbox.archive_or_hide" {
		status = StatusRequiresImplementation
		nextAction = "Bind Inbox preview/confirm/apply handlers before advertising this Inbox state-change skill as executable."
		executable = false
	}
	if strings.HasPrefix(id, "cabinet.users.") && id != "cabinet.users.search" {
		status = StatusRequiresImplementation
		nextAction = "Bind Users admin preview/confirm/apply handlers and protected owner/admin enforcement before advertising this admin skill as executable."
		executable = false
	}
	return deriveExecutionState(Skill{
		ID:              id,
		Version:         "1.0.0",
		DisplayName:     displayName,
		Description:     description,
		Category:        category,
		Source:          SourceBuiltIn,
		Status:          status,
		SafetyLevel:     safety,
		RequiredContext: append([]string{}, context...),
		Capabilities:    append([]string{}, capabilities...),
		GuidedWorkflows: append([]string{}, guidedWorkflows...),
		UITargets:       append([]string{}, uiTargets...),
		Permissions: PermissionDeclaration{
			LocalRead:       true,
			LocalWrite:      safety == SafetyConfirmRequired || safety == SafetyDestructive,
			Destructive:     safety == SafetyDestructive,
			RequiresConfirm: safety == SafetyConfirmRequired || safety == SafetyDestructive,
		},
		AuditBehavior: "thread_message_and_action_timeline",
		Provenance:    "cabinet built-in agent capability registry",
		BuiltIn:       true,
		Removable:     false,
		Enabled:       true,
		Executable:    executable,
		NextAction:    nextAction,
	})
}

func normalizeImported(imported []Skill) []Skill {
	out := make([]Skill, 0, len(imported))
	for _, skill := range imported {
		skill.ID = strings.TrimSpace(skill.ID)
		if skill.ID == "" {
			continue
		}
		skill.Source = SourceArchive
		skill.BuiltIn = false
		skill.Removable = true
		if !skill.Enabled && skill.Status == StatusAvailable {
			skill.Status = StatusDisabled
		}
		out = append(out, skill)
	}
	return out
}

func deriveExecutionState(skill Skill) Skill {
	if skill.Status == "" {
		skill.Status = StatusAvailable
	}
	if skill.SafetyLevel == "" {
		skill.SafetyLevel = SafetyPreviewOnly
	}
	if skill.Source == "" {
		skill.Source = SourceArchive
	}
	if skill.Source == SourceBuiltIn {
		skill.BuiltIn = true
		skill.Removable = false
		skill.Enabled = true
	}
	if skill.Status == StatusDisabled || skill.Status == StatusInvalid || skill.Status == StatusRequiresImplementation {
		skill.Executable = false
	}
	if skill.Status == StatusAvailable || skill.Status == StatusPreviewOnly {
		skill.Executable = skill.Enabled || skill.Source == SourceBuiltIn
	}
	if skill.Status == StatusDisabled && skill.NextAction == "" {
		skill.NextAction = "Enable this imported skill before execution."
	}
	if skill.Status == StatusInvalid && skill.NextAction == "" {
		skill.NextAction = "Repair or remove this imported skill before execution."
	}
	return skill
}

func capabilityIDs() []string {
	out := make([]string, 0)
	for _, capability := range chat.ActionCapabilityRegistry() {
		out = append(out, capability.ID)
	}
	slices.Sort(out)
	return out
}

func guidedWorkflowIDs() []string {
	out := make([]string, 0)
	for _, workflow := range chat.GuidedWorkflowRegistry() {
		out = append(out, workflow.ID)
	}
	slices.Sort(out)
	return out
}

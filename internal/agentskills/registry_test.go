package agentskills

import (
	"slices"
	"testing"
)

func TestSkillRegistryListsAndResolvesBuiltInAndImportedSkills(t *testing.T) {
	t.Parallel()

	registry := NewRegistry([]Skill{{
		ID:              "local.archive.read_only",
		Version:         "0.1.0",
		DisplayName:     "Local read-only archive",
		Description:     "Read local skill fixture metadata.",
		Category:        "testing",
		Status:          StatusAvailable,
		SafetyLevel:     SafetyReadOnly,
		RequiredContext: []string{"profile"},
		Enabled:         true,
	}})

	skills := registry.List()
	if len(skills) < 20 {
		t.Fatalf("expected built-in plus imported skills, got %d", len(skills))
	}
	for _, id := range []string{
		"cabinet.navigate.open_surface",
		"cabinet.inventory.create_item",
		"cabinet.inventory.update_item",
		"cabinet.wishlist.create_entry",
		"cabinet.collection.assign_item",
		"cabinet.guided.inventory.update_item",
		"cabinet.chat.action_timeline.view",
		"cabinet.inbox.search_notifications",
		"cabinet.inbox.summarise_unhandled",
		"cabinet.inbox.open_notification",
		"cabinet.inbox.mark_handled",
		"cabinet.inbox.archive_or_hide",
		"cabinet.inbox.route_to_surface",
		"cabinet.users.search",
		"cabinet.users.invite_user",
		"cabinet.users.resend_invitation",
		"cabinet.users.update_role",
		"cabinet.users.activate_or_deactivate",
		"cabinet.users.remove_user",
		"local.archive.read_only",
	} {
		if !containsSkill(skills, id) {
			t.Fatalf("expected registry to contain %s", id)
		}
	}

	resolved, ok := registry.Resolve("cabinet.inventory.update_item")
	if !ok {
		t.Fatalf("expected to resolve built-in inventory update skill")
	}
	if resolved.Source != SourceBuiltIn || !resolved.BuiltIn || resolved.Removable {
		t.Fatalf("expected immutable built-in source metadata, got %+v", resolved)
	}
	if !slices.Contains(resolved.Capabilities, "inventory.item.update") || !slices.Contains(resolved.Capabilities, "update_open_item_title") {
		t.Fatalf("expected inventory update capability bindings, got %+v", resolved.Capabilities)
	}
	if resolved.SafetyLevel != SafetyConfirmRequired || !resolved.Permissions.RequiresConfirm || !resolved.Permissions.LocalWrite {
		t.Fatalf("expected confirm-required local write permission metadata, got %+v", resolved)
	}
}

func TestInboxAndUsersAdminSkillsExposeSafetyAndExecutionBoundaries(t *testing.T) {
	t.Parallel()

	registry := NewRegistry(nil)

	inboxSearch, ok := registry.Resolve("cabinet.inbox.search_notifications")
	if !ok {
		t.Fatalf("expected inbox search skill")
	}
	if inboxSearch.SafetyLevel != SafetyReadOnly || inboxSearch.Permissions.LocalWrite || !inboxSearch.Executable {
		t.Fatalf("inbox search should be executable read-only metadata, got %+v", inboxSearch)
	}

	inboxMutation, ok := registry.Resolve("cabinet.inbox.mark_handled")
	if !ok {
		t.Fatalf("expected inbox mark handled skill")
	}
	if inboxMutation.SafetyLevel != SafetyConfirmRequired || !inboxMutation.Permissions.RequiresConfirm || !inboxMutation.Permissions.LocalWrite {
		t.Fatalf("inbox mutation should declare confirm-required local write safety, got %+v", inboxMutation)
	}
	if inboxMutation.Status != StatusAvailable || !inboxMutation.Executable || inboxMutation.NextAction != "" {
		t.Fatalf("inbox mutation should be executable after handler binding, got %+v", inboxMutation)
	}

	userSearch, ok := registry.Resolve("cabinet.users.search")
	if !ok {
		t.Fatalf("expected users search skill")
	}
	if userSearch.SafetyLevel != SafetyReadOnly || !slices.Contains(userSearch.RequiredContext, "admin_session") || !userSearch.Executable {
		t.Fatalf("users search should be executable read-only admin metadata, got %+v", userSearch)
	}

	removeUser, ok := registry.Resolve("cabinet.users.remove_user")
	if !ok {
		t.Fatalf("expected remove user skill")
	}
	if removeUser.SafetyLevel != SafetyDestructive || !removeUser.Permissions.Destructive || !removeUser.Permissions.RequiresConfirm {
		t.Fatalf("remove user should declare destructive confirmation safety, got %+v", removeUser)
	}
	if removeUser.Status != StatusAvailable || !removeUser.Executable || removeUser.NextAction != "" {
		t.Fatalf("remove user should be executable after protected admin handlers are bound, got %+v", removeUser)
	}
}

func TestSkillPreviewBlocksUnboundInboxAndUsersMutations(t *testing.T) {
	t.Parallel()

	registry := NewRegistry(nil)

	inboxMissingTarget, err := registry.Preview(PreviewRequest{SkillID: "cabinet.inbox.mark_handled"})
	if err != nil {
		t.Fatalf("preview inbox skill: %v", err)
	}
	if inboxMissingTarget.Allowed || inboxMissingTarget.MutationApplied || inboxMissingTarget.Blocker != "inbox_notification_target_required" {
		t.Fatalf("expected missing Inbox target blocker without mutation, got %+v", inboxMissingTarget)
	}

	inboxTargeted, err := registry.Preview(PreviewRequest{
		SkillID:    "cabinet.inbox.archive_or_hide",
		Parameters: map[string]any{"notification_id": "notice-1", "action": "hide"},
	})
	if err != nil {
		t.Fatalf("preview inbox targeted skill: %v", err)
	}
	if inboxTargeted.Allowed || !inboxTargeted.ConfirmationRequired || inboxTargeted.Blocker != "confirmation_required" {
		t.Fatalf("expected Inbox confirmation blocker, got %+v", inboxTargeted)
	}

	usersMissingTarget, err := registry.Preview(PreviewRequest{SkillID: "cabinet.users.update_role"})
	if err != nil {
		t.Fatalf("preview users skill: %v", err)
	}
	if usersMissingTarget.Allowed || usersMissingTarget.Blocker != "users_admin_target_required" {
		t.Fatalf("expected users admin missing target blocker, got %+v", usersMissingTarget)
	}

	protectedRemove, err := registry.Preview(PreviewRequest{
		SkillID: "cabinet.users.remove_user",
		Parameters: map[string]any{
			"target_user":         "owner@example.test",
			"target_role_current": "owner",
		},
	})
	if err != nil {
		t.Fatalf("preview remove user skill: %v", err)
	}
	if protectedRemove.Allowed || protectedRemove.MutationApplied || protectedRemove.Blocker != "users_admin_protected_owner_remove_blocked" {
		t.Fatalf("expected protected owner removal blocker without mutation, got %+v", protectedRemove)
	}
}

func TestImportedSkillCannotOverrideBuiltInSkillID(t *testing.T) {
	t.Parallel()

	registry := NewRegistry([]Skill{{
		ID:          "cabinet.navigate.open_surface",
		DisplayName: "Unsafe override",
		Enabled:     true,
	}})

	skills := registry.List()
	var count int
	for _, skill := range skills {
		if skill.ID == "cabinet.navigate.open_surface" {
			count++
			if skill.Source != SourceBuiltIn || !skill.BuiltIn || skill.Removable {
				t.Fatalf("override changed built-in metadata: %+v", skill)
			}
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one effective built-in navigate skill, got %d", count)
	}
	if err := registry.ValidateImportedSkill(Skill{ID: "cabinet.navigate.open_surface"}); err == nil {
		t.Fatalf("expected duplicate built-in id validation error")
	}
}

func TestSkillStatusAndSafetyDerivation(t *testing.T) {
	t.Parallel()

	registry := NewRegistry(nil)
	guided, ok := registry.Resolve("cabinet.guided.inventory.update_item")
	if !ok {
		t.Fatalf("expected guided inventory skill")
	}
	if guided.Status != StatusRequiresImplementation || guided.Executable {
		t.Fatalf("guided inventory update must not be executable before #1513, got %+v", guided)
	}
	if !slices.Contains(guided.GuidedWorkflows, "inventory.item.update") {
		t.Fatalf("expected guided workflow binding, got %+v", guided.GuidedWorkflows)
	}
	if guided.NextAction == "" {
		t.Fatalf("expected blocked skill to expose next action")
	}

	readOnly, ok := registry.Resolve("cabinet.chat.action_timeline.view")
	if !ok {
		t.Fatalf("expected action timeline skill")
	}
	if readOnly.SafetyLevel != SafetyReadOnly || readOnly.Permissions.LocalWrite || readOnly.Permissions.RequiresConfirm {
		t.Fatalf("expected read-only no-write timeline skill, got %+v", readOnly)
	}
}

func TestProfileScopedInstalledSkillEnableDisableAndInvalidState(t *testing.T) {
	t.Parallel()

	registry := NewRegistry([]Skill{
		{
			ID:          "local.archive.disabled_writer",
			Version:     "0.1.0",
			DisplayName: "Disabled writer",
			Status:      StatusAvailable,
			SafetyLevel: SafetyConfirmRequired,
			Enabled:     false,
		},
		{
			ID:          "local.archive.invalid_reader",
			Version:     "0.1.0",
			DisplayName: "Invalid reader",
			Status:      StatusInvalid,
			SafetyLevel: SafetyReadOnly,
			Enabled:     true,
		},
	})

	disabled, ok := registry.Resolve("local.archive.disabled_writer")
	if !ok {
		t.Fatalf("expected disabled imported skill")
	}
	if disabled.Status != StatusDisabled || disabled.Executable || !disabled.Removable || disabled.BuiltIn {
		t.Fatalf("disabled imported skill should stay visible and non-executable, got %+v", disabled)
	}
	if disabled.NextAction == "" {
		t.Fatalf("expected disabled imported skill guidance")
	}

	invalid, ok := registry.Resolve("local.archive.invalid_reader")
	if !ok {
		t.Fatalf("expected invalid imported skill")
	}
	if invalid.Status != StatusInvalid || invalid.Executable {
		t.Fatalf("invalid imported skill should not be executable, got %+v", invalid)
	}
}

func containsSkill(skills []Skill, id string) bool {
	for _, skill := range skills {
		if skill.ID == id {
			return true
		}
	}
	return false
}

package chat

import "strings"

type GuidedWorkflowMode string

const (
	GuidedWorkflowModeExplain GuidedWorkflowMode = "explain"
	GuidedWorkflowModeShowMe  GuidedWorkflowMode = "show_me"
	GuidedWorkflowModeWithMe  GuidedWorkflowMode = "do_it_with_me"
	GuidedWorkflowModeForMe   GuidedWorkflowMode = "do_it_for_me"
)

type GuidedWorkflowCommand string

const (
	GuidedCommandNavigateOpenSurface GuidedWorkflowCommand = "navigate.open_surface"
	GuidedCommandHighlightTarget     GuidedWorkflowCommand = "ui.highlight_target"
	GuidedCommandWaitForUserAction   GuidedWorkflowCommand = "ui.wait_for_user_action"
	GuidedCommandPreviewAction       GuidedWorkflowCommand = "chat.action.preview"
	GuidedCommandConfirmApply        GuidedWorkflowCommand = "chat.action.confirm_apply"
)

type GuidedWorkflowStep struct {
	ID            string                `json:"id"`
	Title         string                `json:"title"`
	Instruction   string                `json:"instruction"`
	Route         string                `json:"route"`
	UITargetID    string                `json:"ui_target_id,omitempty"`
	Command       GuidedWorkflowCommand `json:"command"`
	ExpectedState string                `json:"expected_state"`
	OnFailure     string                `json:"on_failure"`
}

type GuidedWorkflowRecipe struct {
	ID                   string                  `json:"id"`
	Title                string                  `json:"title"`
	Description          string                  `json:"description"`
	IntentPhrases        []string                `json:"intent_phrases"`
	ModeSupport          []GuidedWorkflowMode    `json:"mode_support"`
	Preconditions        []string                `json:"preconditions"`
	RequiredContext      []string                `json:"required_context"`
	RouteSequence        []string                `json:"route_sequence"`
	Steps                []GuidedWorkflowStep    `json:"steps"`
	UITargets            []string                `json:"ui_targets"`
	AllowedCommands      []GuidedWorkflowCommand `json:"allowed_commands"`
	MutationBoundaries   string                  `json:"mutation_boundaries"`
	CompletionCriteria   string                  `json:"completion_criteria"`
	AuditDestination     string                  `json:"audit_destination"`
	FallbackInstructions string                  `json:"fallback_instructions"`
}

type GuidedWorkflowContext struct {
	Route             string   `json:"route,omitempty"`
	HasTargetItem     bool     `json:"has_target_item,omitempty"`
	EditableFields    []string `json:"editable_fields,omitempty"`
	Capabilities      []string `json:"capabilities,omitempty"`
	AvailableTargets  []string `json:"available_targets,omitempty"`
	ActiveProfileID   string   `json:"active_profile_id,omitempty"`
	ActiveThreadID    string   `json:"active_thread_id,omitempty"`
	ActiveWorkspaceID string   `json:"active_workspace_id,omitempty"`
}

type GuidedWorkflowMatchInput struct {
	Request string                `json:"request"`
	Context GuidedWorkflowContext `json:"context"`
}

type GuidedWorkflowMatch struct {
	Matched        bool                 `json:"matched"`
	Recipe         GuidedWorkflowRecipe `json:"recipe,omitempty"`
	FollowUpPrompt string               `json:"follow_up_prompt,omitempty"`
	UnsafeActions  []string             `json:"unsafe_actions,omitempty"`
}

func GuidedWorkflowRegistry() []GuidedWorkflowRecipe {
	return []GuidedWorkflowRecipe{
		inventoryItemUpdateRecipe(),
		wishlistEntryCreateRecipe(),
	}
}

func MatchGuidedWorkflowRecipe(input GuidedWorkflowMatchInput) GuidedWorkflowMatch {
	normalized := normalizeWorkflowIntent(input.Request)
	if normalized == "" {
		return guidedWorkflowFollowUp("I need a supported Cabinet workflow and enough context before I can guide it safely.")
	}
	if matchesInventoryUpdateIntent(normalized) {
		if strings.TrimSpace(input.Context.ActiveProfileID) == "" || strings.TrimSpace(input.Context.ActiveThreadID) == "" {
			return guidedWorkflowFollowUp("I need an active profile and assistant thread before starting an inventory update walkthrough.")
		}
		if !input.Context.HasTargetItem {
			return guidedWorkflowFollowUp("Please open or select the inventory item before I build update steps.")
		}
		if len(input.Context.EditableFields) == 0 {
			return guidedWorkflowFollowUp("Tell me which editable item field you want to update before I create guided steps.")
		}
		for _, required := range []string{
			string(GuidedCommandNavigateOpenSurface),
			string(GuidedCommandHighlightTarget),
			string(GuidedCommandPreviewAction),
			string(GuidedCommandConfirmApply),
		} {
			if !containsString(input.Context.Capabilities, required) {
				return guidedWorkflowFollowUp("This walkthrough needs safe preview and confirmation tools before it can run.")
			}
		}
		for _, required := range []string{"inventory.item.row", "inventory.item.editor.title", "inventory.item.editor.save"} {
			if !containsString(input.Context.AvailableTargets, required) {
				return guidedWorkflowFollowUp("This walkthrough needs stable inventory UI targets before it can run.")
			}
		}
		return GuidedWorkflowMatch{Matched: true, Recipe: inventoryItemUpdateRecipe()}
	}
	if matchesWishlistCreateIntent(normalized) {
		return GuidedWorkflowMatch{Matched: true, Recipe: wishlistEntryCreateRecipe()}
	}
	return guidedWorkflowFollowUp("I need a supported Cabinet workflow such as updating an inventory item or creating a wishlist entry.")
}

func inventoryItemUpdateRecipe() GuidedWorkflowRecipe {
	return GuidedWorkflowRecipe{
		ID:          "inventory.item.update",
		Title:       "Update an inventory item",
		Description: "Guide a user through selecting an inventory item, focusing an editable field, previewing the update, and confirming before persistence.",
		IntentPhrases: []string{
			"show me how to update an item",
			"help me edit an inventory item",
			"walk me through changing an item title",
		},
		ModeSupport: []GuidedWorkflowMode{
			GuidedWorkflowModeExplain,
			GuidedWorkflowModeShowMe,
			GuidedWorkflowModeWithMe,
			GuidedWorkflowModeForMe,
		},
		Preconditions:   []string{"authenticated_profile", "assistant_thread_open", "inventory_surface_available"},
		RequiredContext: []string{"active_profile", "assistant_thread", "target_inventory_item", "editable_field"},
		RouteSequence:   []string{"/inventory"},
		Steps: []GuidedWorkflowStep{
			{
				ID:            "open-inventory",
				Title:         "Open Inventory",
				Instruction:   "Open the Inventory surface while preserving the assistant thread.",
				Route:         "/inventory",
				Command:       GuidedCommandNavigateOpenSurface,
				ExpectedState: "Inventory route is visible and side-panel Chat remains open.",
				OnFailure:     "Keep the current route and ask the user to retry or open Inventory manually.",
			},
			{
				ID:            "select-item",
				Title:         "Select item",
				Instruction:   "Highlight the item row that will be edited.",
				Route:         "/inventory",
				UITargetID:    "inventory.item.row",
				Command:       GuidedCommandHighlightTarget,
				ExpectedState: "The target inventory item is selected or visibly called out.",
				OnFailure:     "Ask the user to select a target item before continuing.",
			},
			{
				ID:            "focus-editable-field",
				Title:         "Focus editable field",
				Instruction:   "Highlight the editable title field and collect the intended value.",
				Route:         "/inventory",
				UITargetID:    "inventory.item.editor.title",
				Command:       GuidedCommandHighlightTarget,
				ExpectedState: "The editable title field is visible and ready for a draft value.",
				OnFailure:     "Report the missing target and avoid creating a mutation preview.",
			},
			{
				ID:            "preview-change",
				Title:         "Preview change",
				Instruction:   "Create a chat action preview for the intended field update.",
				Route:         "/inventory",
				UITargetID:    "inventory.item.editor.title",
				Command:       GuidedCommandPreviewAction,
				ExpectedState: "A preview records item id, field, old value, new value, and pending confirmation.",
				OnFailure:     "Leave the item unchanged and show retry guidance.",
			},
			{
				ID:            "confirm-apply",
				Title:         "Confirm apply",
				Instruction:   "Wait for explicit confirmation before saving the update.",
				Route:         "/inventory",
				UITargetID:    "inventory.item.editor.save",
				Command:       GuidedCommandConfirmApply,
				ExpectedState: "Confirmed apply persists the item update and records audit evidence.",
				OnFailure:     "Cancel or pause without mutating the item.",
			},
		},
		UITargets: []string{
			"inventory.item.row",
			"inventory.item.editor.title",
			"inventory.item.editor.save",
		},
		AllowedCommands: []GuidedWorkflowCommand{
			GuidedCommandNavigateOpenSurface,
			GuidedCommandHighlightTarget,
			GuidedCommandWaitForUserAction,
			GuidedCommandPreviewAction,
			GuidedCommandConfirmApply,
		},
		MutationBoundaries:   "preview_before_apply_explicit_confirmation",
		CompletionCriteria:   "persisted item update for the active profile plus visible Action Timeline result",
		AuditDestination:     "assistant_workflow_runs and assistant thread action timeline",
		FallbackInstructions: "If item, field, target, preview, or confirmation context is missing, ask a focused follow-up and do not invent UI actions.",
	}
}

func wishlistEntryCreateRecipe() GuidedWorkflowRecipe {
	return GuidedWorkflowRecipe{
		ID:                   "wishlist.entry.create",
		Title:                "Create a wishlist entry",
		Description:          "Guide a user through drafting a wishlist entry and confirming before persistence.",
		IntentPhrases:        []string{"help me add a wishlist entry", "show me how to create a wish list item"},
		ModeSupport:          []GuidedWorkflowMode{GuidedWorkflowModeExplain, GuidedWorkflowModeShowMe, GuidedWorkflowModeWithMe},
		Preconditions:        []string{"authenticated_profile", "assistant_thread_open"},
		RequiredContext:      []string{"active_profile", "assistant_thread", "wanted_item_details"},
		RouteSequence:        []string{"/wishlist"},
		UITargets:            []string{"wishlist.create.button", "wishlist.entry.form", "wishlist.entry.save"},
		AllowedCommands:      []GuidedWorkflowCommand{GuidedCommandNavigateOpenSurface, GuidedCommandHighlightTarget, GuidedCommandPreviewAction, GuidedCommandConfirmApply},
		MutationBoundaries:   "preview_before_apply_explicit_confirmation",
		CompletionCriteria:   "confirmed wishlist entry creation is visible in Wishlist for the active profile",
		AuditDestination:     "assistant_workflow_runs and assistant thread action timeline",
		FallbackInstructions: "Ask for the wanted item details before creating a preview.",
		Steps: []GuidedWorkflowStep{
			{
				ID:            "open-wishlist",
				Title:         "Open Wishlist",
				Instruction:   "Open the Wishlist surface while preserving the assistant thread.",
				Route:         "/wishlist",
				Command:       GuidedCommandNavigateOpenSurface,
				ExpectedState: "Wishlist route is visible and side-panel Chat remains open.",
				OnFailure:     "Keep the current route and ask the user to retry or open Wishlist manually.",
			},
			{
				ID:            "preview-entry",
				Title:         "Preview entry",
				Instruction:   "Create a preview from the wanted item details.",
				Route:         "/wishlist",
				UITargetID:    "wishlist.entry.form",
				Command:       GuidedCommandPreviewAction,
				ExpectedState: "A wishlist preview is pending explicit confirmation.",
				OnFailure:     "Leave Wishlist unchanged and ask for missing details.",
			},
		},
	}
}

func matchesInventoryUpdateIntent(normalized string) bool {
	return strings.Contains(normalized, "update") && strings.Contains(normalized, "item") ||
		strings.Contains(normalized, "edit") && strings.Contains(normalized, "inventory") ||
		strings.Contains(normalized, "change") && strings.Contains(normalized, "item title")
}

func matchesWishlistCreateIntent(normalized string) bool {
	return (strings.Contains(normalized, "wishlist") || strings.Contains(normalized, "wish list")) &&
		(strings.Contains(normalized, "create") || strings.Contains(normalized, "add"))
}

func normalizeWorkflowIntent(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(value))), " ")
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == needle {
			return true
		}
	}
	return false
}

func guidedWorkflowFollowUp(prompt string) GuidedWorkflowMatch {
	return GuidedWorkflowMatch{FollowUpPrompt: prompt, UnsafeActions: []string{}}
}

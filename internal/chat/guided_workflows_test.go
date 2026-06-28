package chat

import (
	"slices"
	"strings"
	"testing"
)

func TestGuidedWorkflowRegistryMatchesInventoryItemUpdateRecipe(t *testing.T) {
	t.Parallel()

	match := MatchGuidedWorkflowRecipe(GuidedWorkflowMatchInput{
		Request: "show me how to update an item",
		Context: GuidedWorkflowContext{
			Route:             "/inventory",
			HasTargetItem:     true,
			EditableFields:    []string{"title", "condition"},
			Capabilities:      []string{"navigate.open_surface", "ui.highlight_target", "chat.action.preview", "chat.action.confirm_apply"},
			AvailableTargets:  []string{"inventory.item.row", "inventory.item.editor.title", "inventory.item.editor.save"},
			ActiveProfileID:   "profile-1",
			ActiveThreadID:    "thread-1",
			ActiveWorkspaceID: "inventory",
		},
	})

	if !match.Matched {
		t.Fatalf("expected inventory update match, got %+v", match)
	}
	if match.FollowUpPrompt != "" {
		t.Fatalf("matched recipe must not ask a follow-up, got %q", match.FollowUpPrompt)
	}
	recipe := match.Recipe
	if recipe.ID != "inventory.item.update" || recipe.Title == "" {
		t.Fatalf("expected inventory.item.update recipe, got %+v", recipe)
	}
	if !slices.Contains(recipe.IntentPhrases, "show me how to update an item") {
		t.Fatalf("expected recipe to expose matching intent phrase, got %+v", recipe.IntentPhrases)
	}
	for _, mode := range []GuidedWorkflowMode{"explain", "show_me", "do_it_with_me", "do_it_for_me"} {
		if !slices.Contains(recipe.ModeSupport, mode) {
			t.Fatalf("expected inventory recipe to support mode %s, got %+v", mode, recipe.ModeSupport)
		}
	}
	for _, required := range []string{"active_profile", "assistant_thread", "target_inventory_item", "editable_field"} {
		if !slices.Contains(recipe.RequiredContext, required) {
			t.Fatalf("expected required context %q, got %+v", required, recipe.RequiredContext)
		}
	}
	for _, command := range []GuidedWorkflowCommand{"navigate.open_surface", "ui.highlight_target", "chat.action.preview", "chat.action.confirm_apply"} {
		if !slices.Contains(recipe.AllowedCommands, command) {
			t.Fatalf("expected allowed command %q, got %+v", command, recipe.AllowedCommands)
		}
	}
	for _, target := range []string{"inventory.item.row", "inventory.item.editor.title", "inventory.item.editor.save"} {
		if !slices.Contains(recipe.UITargets, target) {
			t.Fatalf("expected UI target %q, got %+v", target, recipe.UITargets)
		}
	}
	if len(recipe.Steps) < 5 {
		t.Fatalf("expected ordered walkthrough steps, got %+v", recipe.Steps)
	}
	assertStep(t, recipe.Steps[0], "open-inventory", "navigate.open_surface", "/inventory", "")
	assertStep(t, recipe.Steps[1], "select-item", "ui.highlight_target", "/inventory", "inventory.item.row")
	assertStep(t, recipe.Steps[2], "focus-editable-field", "ui.highlight_target", "/inventory", "inventory.item.editor.title")
	assertStep(t, recipe.Steps[3], "preview-change", "chat.action.preview", "/inventory", "inventory.item.editor.title")
	assertStep(t, recipe.Steps[4], "confirm-apply", "chat.action.confirm_apply", "/inventory", "inventory.item.editor.save")
	if recipe.MutationBoundaries != "preview_before_apply_explicit_confirmation" {
		t.Fatalf("expected confirmation mutation boundary, got %q", recipe.MutationBoundaries)
	}
	if !strings.Contains(recipe.CompletionCriteria, "persisted item update") || !strings.Contains(recipe.AuditDestination, "assistant_workflow_runs") {
		t.Fatalf("expected persistence and audit destinations, got criteria=%q audit=%q", recipe.CompletionCriteria, recipe.AuditDestination)
	}

	registry := GuidedWorkflowRegistry()
	if len(registry) < 2 {
		t.Fatalf("expected at least two recipes for extensibility, got %+v", registry)
	}
	if !slices.ContainsFunc(registry, func(candidate GuidedWorkflowRecipe) bool {
		return candidate.ID == "wishlist.entry.create"
	}) {
		t.Fatalf("expected a non-inventory example recipe, got %+v", registry)
	}
}

func TestGuidedWorkflowRegistryAsksFollowUpForUnderSpecifiedRequests(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name    string
		request string
		context GuidedWorkflowContext
		want    string
	}{
		{
			name:    "unknown workflow",
			request: "make Cabinet organize my hobby room",
			context: GuidedWorkflowContext{ActiveProfileID: "profile-1", ActiveThreadID: "thread-1"},
			want:    "I need a supported Cabinet workflow",
		},
		{
			name:    "missing item",
			request: "show me how to update an item",
			context: GuidedWorkflowContext{
				ActiveProfileID:  "profile-1",
				ActiveThreadID:   "thread-1",
				EditableFields:   []string{"title"},
				Capabilities:     []string{"navigate.open_surface", "ui.highlight_target", "chat.action.preview", "chat.action.confirm_apply"},
				AvailableTargets: []string{"inventory.item.editor.title"},
			},
			want: "open or select the inventory item",
		},
		{
			name:    "missing capability",
			request: "show me how to update an item",
			context: GuidedWorkflowContext{
				ActiveProfileID:  "profile-1",
				ActiveThreadID:   "thread-1",
				HasTargetItem:    true,
				EditableFields:   []string{"title"},
				Capabilities:     []string{"navigate.open_surface"},
				AvailableTargets: []string{"inventory.item.row", "inventory.item.editor.title", "inventory.item.editor.save"},
			},
			want: "safe preview and confirmation tools",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			match := MatchGuidedWorkflowRecipe(GuidedWorkflowMatchInput{
				Request: tt.request,
				Context: tt.context,
			})
			if match.Matched {
				t.Fatalf("expected follow-up instead of recipe, got %+v", match)
			}
			if !strings.Contains(match.FollowUpPrompt, tt.want) {
				t.Fatalf("expected follow-up prompt to include %q, got %q", tt.want, match.FollowUpPrompt)
			}
			if len(match.UnsafeActions) != 0 {
				t.Fatalf("follow-up path must not invent unsafe actions, got %+v", match.UnsafeActions)
			}
		})
	}
}

func assertStep(t *testing.T, step GuidedWorkflowStep, id string, command GuidedWorkflowCommand, route, target string) {
	t.Helper()
	if step.ID != id || step.Command != command || step.Route != route || step.UITargetID != target {
		t.Fatalf("expected step id=%s command=%s route=%s target=%s, got %+v", id, command, route, target, step)
	}
	if step.Title == "" || step.Instruction == "" || step.ExpectedState == "" || step.OnFailure == "" {
		t.Fatalf("step must include user and validation guidance, got %+v", step)
	}
}

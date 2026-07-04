package app

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"
)

type apiSkillPayload struct {
	ID              string   `json:"id"`
	Source          string   `json:"source"`
	Status          string   `json:"status"`
	SafetyLevel     string   `json:"safety_level"`
	RequiredContext []string `json:"required_context"`
	Capabilities    []string `json:"capabilities"`
	GuidedWorkflows []string `json:"guided_workflows"`
	BuiltIn         bool     `json:"built_in"`
	Removable       bool     `json:"removable"`
	Executable      bool     `json:"executable"`
	NextAction      string   `json:"next_action"`
	Permissions     struct {
		LocalWrite      bool `json:"local_write"`
		Destructive     bool `json:"destructive"`
		RequiresConfirm bool `json:"requires_confirm"`
	} `json:"permissions"`
}

func TestAgentSkillRegistryAPIExposesGovernedSkillMetadata(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodGet, "/api/agent/skills?profile_id=test-profile", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("skills status=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload struct {
		ProfileID string            `json:"profile_id"`
		Skills    []apiSkillPayload `json:"skills"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode skills: %v", err)
	}
	if payload.ProfileID != "test-profile" {
		t.Fatalf("expected profile echo, got %q", payload.ProfileID)
	}
	inventory := findAPISkill(payload.Skills, "cabinet.inventory.update_item")
	if inventory == nil {
		t.Fatalf("missing inventory update skill")
	}
	if inventory.Source != "built-in" || !inventory.BuiltIn || inventory.Removable {
		t.Fatalf("expected immutable built-in metadata, got %+v", inventory)
	}
	if inventory.SafetyLevel != "confirm-required" || !inventory.Permissions.LocalWrite || !inventory.Permissions.RequiresConfirm {
		t.Fatalf("expected confirm-required write safety, got %+v", inventory)
	}
	if !slices.Contains(inventory.Capabilities, "inventory.item.update") {
		t.Fatalf("expected capability binding, got %+v", inventory.Capabilities)
	}

	guided := findAPISkill(payload.Skills, "cabinet.guided.inventory.update_item")
	if guided == nil {
		t.Fatalf("missing guided inventory update skill")
	}
	if guided.Status != "requires-implementation" || guided.Executable {
		t.Fatalf("guided skill must remain non-executable before #1513, got %+v", guided)
	}
	if !slices.Contains(guided.GuidedWorkflows, "inventory.item.update") || guided.NextAction == "" {
		t.Fatalf("expected guided workflow binding and next action, got %+v", guided)
	}

	inboxSearch := findAPISkill(payload.Skills, "cabinet.inbox.search_notifications")
	if inboxSearch == nil {
		t.Fatalf("missing Inbox search skill")
	}
	if inboxSearch.SafetyLevel != "read-only" || inboxSearch.Permissions.LocalWrite || !inboxSearch.Executable {
		t.Fatalf("expected executable read-only Inbox search metadata, got %+v", inboxSearch)
	}

	inboxMutation := findAPISkill(payload.Skills, "cabinet.inbox.mark_handled")
	if inboxMutation == nil {
		t.Fatalf("missing Inbox mark handled skill")
	}
	if inboxMutation.Status != "requires-implementation" || inboxMutation.Executable || !inboxMutation.Permissions.RequiresConfirm {
		t.Fatalf("expected Inbox mutation to stay confirmation-gated and non-executable, got %+v", inboxMutation)
	}

	userSearch := findAPISkill(payload.Skills, "cabinet.users.search")
	if userSearch == nil {
		t.Fatalf("missing Users search skill")
	}
	if userSearch.SafetyLevel != "read-only" || !slices.Contains(userSearch.RequiredContext, "admin_session") || !userSearch.Executable {
		t.Fatalf("expected executable read-only Users search metadata, got %+v", userSearch)
	}

	removeUser := findAPISkill(payload.Skills, "cabinet.users.remove_user")
	if removeUser == nil {
		t.Fatalf("missing remove user skill")
	}
	if removeUser.SafetyLevel != "destructive" || !removeUser.Permissions.Destructive || removeUser.Executable {
		t.Fatalf("expected destructive non-executable remove user metadata, got %+v", removeUser)
	}
}

func TestAgentSkillPreviewAPIBlocksUnsafeAdminMutation(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"test-profile",
		"skill_id":"cabinet.users.remove_user",
		"confirm":true,
		"parameters":{
			"target_user":"owner@example.test",
			"target_role_current":"owner"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusOK {
		t.Fatalf("preview status=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload struct {
		SkillID              string `json:"skill_id"`
		Allowed              bool   `json:"allowed"`
		PreviewOnly          bool   `json:"preview_only"`
		MutationApplied      bool   `json:"mutation_applied"`
		ConfirmationRequired bool   `json:"confirmation_required"`
		Blocker              string `json:"blocker"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode preview: %v", err)
	}
	if payload.SkillID != "cabinet.users.remove_user" || payload.Allowed || !payload.PreviewOnly || payload.MutationApplied {
		t.Fatalf("expected preview-only blocked admin mutation, got %+v", payload)
	}
	if !payload.ConfirmationRequired || payload.Blocker != "users_admin_protected_owner_remove_blocked" {
		t.Fatalf("expected protected owner confirmation blocker, got %+v", payload)
	}
}

func findAPISkill(skills []apiSkillPayload, id string) *apiSkillPayload {
	for i := range skills {
		if skills[i].ID == id {
			return &skills[i]
		}
	}
	return nil
}

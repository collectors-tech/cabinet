package app

import (
	"context"
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
	if inboxMutation.Status != "available" || !inboxMutation.Executable || !inboxMutation.Permissions.RequiresConfirm {
		t.Fatalf("expected Inbox mutation to be executable and confirmation-gated, got %+v", inboxMutation)
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
	if removeUser.SafetyLevel != "destructive" || !removeUser.Permissions.Destructive || !removeUser.Executable {
		t.Fatalf("expected destructive executable remove user metadata, got %+v", removeUser)
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

func TestAgentSkillApplyAPIConfirmsInboxMutation(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Inbox Apply"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	record := doRequest(t, a, http.MethodPost, "/api/chat/inbox", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"records":[{
			"local_history_id":"agent-skill-apply-1",
			"title":"Agent skill handoff",
			"summary":"Needs review"
		}]
	}`), map[string]string{"Content-Type": "application/json"})
	if record.Code != http.StatusCreated {
		t.Fatalf("create inbox record status=%d body=%s", record.Code, record.Body.String())
	}
	var recordPayload struct {
		Items []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"items"`
	}
	if err := json.NewDecoder(record.Body).Decode(&recordPayload); err != nil {
		t.Fatalf("decode inbox record: %v", err)
	}
	if len(recordPayload.Items) != 1 {
		t.Fatalf("expected one inbox item, got %+v", recordPayload.Items)
	}

	preview := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inbox.mark_handled",
		"parameters":{"notification_id":"`+recordPayload.Items[0].ID+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if preview.Code != http.StatusOK {
		t.Fatalf("preview inbox status=%d body=%s", preview.Code, preview.Body.String())
	}
	if !strings.Contains(preview.Body.String(), `"mutation_applied":false`) || !strings.Contains(preview.Body.String(), `"blocker":"confirmation_required"`) {
		t.Fatalf("preview must stay non-mutating and confirmation-gated, body=%s", preview.Body.String())
	}

	apply := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inbox.mark_handled",
		"confirm":true,
		"parameters":{"notification_id":"`+recordPayload.Items[0].ID+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if apply.Code != http.StatusOK {
		t.Fatalf("apply inbox status=%d body=%s", apply.Code, apply.Body.String())
	}
	if !strings.Contains(apply.Body.String(), `"mutation_applied":true`) || !strings.Contains(apply.Body.String(), `"status":"read"`) {
		t.Fatalf("expected confirmed apply to mark inbox item read, body=%s", apply.Body.String())
	}
}

func TestAgentSkillApplyAPIConfirmsUsersMutationAndProtectsOwner(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Users Apply"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	invite := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.users.invite_user",
		"confirm":true,
		"parameters":{"target_email":"agent_skill_invite@example.test","target_role":"view"}
	}`), map[string]string{"Content-Type": "application/json"})
	if invite.Code != http.StatusOK {
		t.Fatalf("invite apply status=%d body=%s", invite.Code, invite.Body.String())
	}
	if !strings.Contains(invite.Body.String(), `"mutation_applied":true`) || !strings.Contains(invite.Body.String(), `"email":"agent_skill_invite@example.test"`) || !strings.Contains(invite.Body.String(), `"status":"invited"`) {
		t.Fatalf("expected invite apply result, body=%s", invite.Body.String())
	}

	users, err := listRuntimeUsers(context.Background(), a.db, p.ID)
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	var ownerID string
	for _, user := range users {
		if strings.HasPrefix(user.Username, "owner_") {
			ownerID = user.ID
		}
	}
	if ownerID == "" {
		t.Fatalf("expected seeded owner user, got %+v", users)
	}
	removeOwner := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.users.remove_user",
		"confirm":true,
		"parameters":{"target_user":"`+ownerID+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if removeOwner.Code != http.StatusBadRequest {
		t.Fatalf("remove protected owner status=%d body=%s", removeOwner.Code, removeOwner.Body.String())
	}
	if !strings.Contains(removeOwner.Body.String(), `"mutation_applied":false`) || !strings.Contains(removeOwner.Body.String(), `"blocker":"users_admin_protected_owner_change_blocked"`) {
		t.Fatalf("expected protected owner blocker, body=%s", removeOwner.Body.String())
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

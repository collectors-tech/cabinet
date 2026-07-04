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

func TestAgentSkillAPIPropagatesInvocationSourceContext(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Source Context"}`), map[string]string{"Content-Type": "application/json"})
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
			"local_history_id":"agent-skill-source-context",
			"title":"Agent skill source context",
			"summary":"Needs sourced review"
		}]
	}`), map[string]string{"Content-Type": "application/json"})
	if record.Code != http.StatusCreated {
		t.Fatalf("create inbox record status=%d body=%s", record.Code, record.Body.String())
	}
	var recordPayload struct {
		Items []struct {
			ID string `json:"id"`
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
		"source_surface":"inbox.notification.card",
		"source_channel":"in-app",
		"source_thread_id":"thread-source-context",
		"source_message_id":"message-source-context",
		"parameters":{"notification_id":"`+recordPayload.Items[0].ID+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if preview.Code != http.StatusOK {
		t.Fatalf("preview source context status=%d body=%s", preview.Code, preview.Body.String())
	}
	if !strings.Contains(preview.Body.String(), `"source_surface":"inbox.notification.card"`) ||
		!strings.Contains(preview.Body.String(), `"source_channel":"in-app"`) ||
		!strings.Contains(preview.Body.String(), `"source_thread_id":"thread-source-context"`) ||
		!strings.Contains(preview.Body.String(), `"source_message_id":"message-source-context"`) ||
		!strings.Contains(preview.Body.String(), `"mutation_applied":false`) {
		t.Fatalf("expected preview to retain source context without mutation, body=%s", preview.Body.String())
	}

	apply := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inbox.mark_handled",
		"confirm":true,
		"source_surface":"inbox.notification.card",
		"source_channel":"telegram",
		"source_thread_id":"tg-chat-42",
		"source_message_id":"tg-message-99",
		"parameters":{"notification_id":"`+recordPayload.Items[0].ID+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if apply.Code != http.StatusOK {
		t.Fatalf("apply source context status=%d body=%s", apply.Code, apply.Body.String())
	}
	if !strings.Contains(apply.Body.String(), `"source_surface":"inbox.notification.card"`) ||
		!strings.Contains(apply.Body.String(), `"source_channel":"telegram"`) ||
		!strings.Contains(apply.Body.String(), `"source_thread_id":"tg-chat-42"`) ||
		!strings.Contains(apply.Body.String(), `"source_message_id":"tg-message-99"`) ||
		!strings.Contains(apply.Body.String(), `"mutation_applied":true`) ||
		!strings.Contains(apply.Body.String(), `"status":"read"`) {
		t.Fatalf("expected confirmed apply to retain channel source context, body=%s", apply.Body.String())
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
		},{
			"local_history_id":"agent-skill-apply-2",
			"title":"Agent skill archive",
			"summary":"Can be archived"
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
	if len(recordPayload.Items) != 2 {
		t.Fatalf("expected two inbox items, got %+v", recordPayload.Items)
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

	archive := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inbox.archive_or_hide",
		"confirm":true,
		"parameters":{"notification_id":"`+recordPayload.Items[1].ID+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if archive.Code != http.StatusOK {
		t.Fatalf("archive inbox status=%d body=%s", archive.Code, archive.Body.String())
	}
	if !strings.Contains(archive.Body.String(), `"mutation_applied":true`) || !strings.Contains(archive.Body.String(), `"status":"archived"`) {
		t.Fatalf("expected confirmed apply to archive inbox item, body=%s", archive.Body.String())
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
	var invitedID string
	for _, user := range users {
		if user.Email == "agent_skill_invite@example.test" {
			invitedID = user.ID
		}
	}
	if invitedID == "" {
		t.Fatalf("expected invited user, got %+v", users)
	}

	updateRole := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.users.update_role",
		"confirm":true,
		"parameters":{"target_user":"`+invitedID+`","target_role":"admin"}
	}`), map[string]string{"Content-Type": "application/json"})
	if updateRole.Code != http.StatusOK {
		t.Fatalf("update role apply status=%d body=%s", updateRole.Code, updateRole.Body.String())
	}
	if !strings.Contains(updateRole.Body.String(), `"mutation_applied":true`) || !strings.Contains(updateRole.Body.String(), `"role":"admin"`) {
		t.Fatalf("expected confirmed role update result, body=%s", updateRole.Body.String())
	}

	deactivate := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.users.activate_or_deactivate",
		"confirm":true,
		"parameters":{"target_user":"`+invitedID+`","target_status":"inactive"}
	}`), map[string]string{"Content-Type": "application/json"})
	if deactivate.Code != http.StatusOK {
		t.Fatalf("deactivate apply status=%d body=%s", deactivate.Code, deactivate.Body.String())
	}
	if !strings.Contains(deactivate.Body.String(), `"mutation_applied":true`) || !strings.Contains(deactivate.Body.String(), `"status":"inactive"`) {
		t.Fatalf("expected confirmed deactivate result, body=%s", deactivate.Body.String())
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

	removeInvited := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.users.remove_user",
		"confirm":true,
		"parameters":{"target_user":"`+invitedID+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if removeInvited.Code != http.StatusOK {
		t.Fatalf("remove invited user status=%d body=%s", removeInvited.Code, removeInvited.Body.String())
	}
	if !strings.Contains(removeInvited.Body.String(), `"mutation_applied":true`) || !strings.Contains(removeInvited.Body.String(), `"removed_user_id":"`+invitedID+`"`) {
		t.Fatalf("expected confirmed remove user result, body=%s", removeInvited.Body.String())
	}
	users, err = listRuntimeUsers(context.Background(), a.db, p.ID)
	if err != nil {
		t.Fatalf("list users after remove: %v", err)
	}
	for _, user := range users {
		if user.ID == invitedID {
			t.Fatalf("expected invited user to be removed, got %+v", users)
		}
	}
}

func TestAgentSkillApplyAPIHandlesIntegrationsAndSettingsSkills(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Integrations Apply"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	testConnection := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.integrations.test_connection",
		"parameters":{"provider_id":"ebay","provider_secret":"must-not-leak"}
	}`), map[string]string{"Content-Type": "application/json"})
	if testConnection.Code != http.StatusOK {
		t.Fatalf("test connection apply status=%d body=%s", testConnection.Code, testConnection.Body.String())
	}
	if !strings.Contains(testConnection.Body.String(), `"mutation_applied":false`) ||
		!strings.Contains(testConnection.Body.String(), `"operation":"integrations.provider.test_connection"`) ||
		!strings.Contains(testConnection.Body.String(), `"connection_status":"setup_needed"`) ||
		strings.Contains(testConnection.Body.String(), "must-not-leak") {
		t.Fatalf("expected non-mutating provider test without secret leak, body=%s", testConnection.Body.String())
	}

	configure := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.integrations.configure_provider",
		"confirm":true,
		"parameters":{"provider_id":"ebay","provider_secret":"must-not-leak","setup_step":"oauth"}
	}`), map[string]string{"Content-Type": "application/json"})
	if configure.Code != http.StatusOK {
		t.Fatalf("configure provider apply status=%d body=%s", configure.Code, configure.Body.String())
	}
	if !strings.Contains(configure.Body.String(), `"mutation_applied":true`) ||
		!strings.Contains(configure.Body.String(), `"operation":"integrations.provider.configure"`) ||
		!strings.Contains(configure.Body.String(), `"secret_redacted":true`) ||
		strings.Contains(configure.Body.String(), "must-not-leak") {
		t.Fatalf("expected confirmed provider configure result without secret leak, body=%s", configure.Body.String())
	}

	appearance := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.settings.update_appearance",
		"confirm":true,
		"parameters":{"setting_key":"theme","setting_scope":"appearance","setting_value":"dark"}
	}`), map[string]string{"Content-Type": "application/json"})
	if appearance.Code != http.StatusOK {
		t.Fatalf("appearance apply status=%d body=%s", appearance.Code, appearance.Body.String())
	}
	if !strings.Contains(appearance.Body.String(), `"mutation_applied":true`) ||
		!strings.Contains(appearance.Body.String(), `"operation":"settings.appearance.update"`) ||
		!strings.Contains(appearance.Body.String(), `"setting_key":"theme"`) {
		t.Fatalf("expected confirmed appearance setting result, body=%s", appearance.Body.String())
	}

	storageStatus := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.storage.show_status"
	}`), map[string]string{"Content-Type": "application/json"})
	if storageStatus.Code != http.StatusOK {
		t.Fatalf("storage status apply status=%d body=%s", storageStatus.Code, storageStatus.Body.String())
	}
	if !strings.Contains(storageStatus.Body.String(), `"mutation_applied":false`) ||
		!strings.Contains(storageStatus.Body.String(), `"operation":"storage.status.show"`) ||
		!strings.Contains(storageStatus.Body.String(), `"read_only":true`) {
		t.Fatalf("expected read-only storage status result, body=%s", storageStatus.Body.String())
	}
}

func TestAgentSkillApplyAPIRequiresConfirmationAndRejectsUnknownSkill(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Apply Guard"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	cancel := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.users.invite_user",
		"confirm":false,
		"parameters":{"target_email":"agent_skill_cancel@example.test","target_role":"view"}
	}`), map[string]string{"Content-Type": "application/json"})
	if cancel.Code != http.StatusConflict {
		t.Fatalf("cancelled apply status=%d body=%s", cancel.Code, cancel.Body.String())
	}
	if !strings.Contains(cancel.Body.String(), `"error":"confirmation_required"`) {
		t.Fatalf("expected confirmation_required on cancelled apply, body=%s", cancel.Body.String())
	}
	users, err := listRuntimeUsers(context.Background(), a.db, p.ID)
	if err != nil {
		t.Fatalf("list users after cancelled apply: %v", err)
	}
	for _, user := range users {
		if user.Email == "agent_skill_cancel@example.test" {
			t.Fatalf("cancelled apply must not create user, got %+v", users)
		}
	}

	unknown := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.users.unsupported",
		"confirm":true,
		"parameters":{"target_user":"missing"}
	}`), map[string]string{"Content-Type": "application/json"})
	if unknown.Code != http.StatusNotFound {
		t.Fatalf("unknown apply status=%d body=%s", unknown.Code, unknown.Body.String())
	}
	if !strings.Contains(unknown.Body.String(), `"error":"skill_not_found"`) {
		t.Fatalf("expected skill_not_found on unknown skill, body=%s", unknown.Body.String())
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

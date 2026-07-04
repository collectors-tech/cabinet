package app

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/profile"
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

	runWatch := findAPISkill(payload.Skills, "cabinet.market_watch.run_watch")
	if runWatch == nil {
		t.Fatalf("missing Market Watch run skill")
	}
	if runWatch.SafetyLevel != "confirm-required" || !runWatch.Permissions.RequiresConfirm || !runWatch.Executable {
		t.Fatalf("expected confirmation-gated Market Watch run metadata, got %+v", runWatch)
	}

	addPurchaseLine := findAPISkill(payload.Skills, "cabinet.purchases.add_line_item")
	if addPurchaseLine == nil {
		t.Fatalf("missing Purchases add line item skill")
	}
	if addPurchaseLine.SafetyLevel != "confirm-required" || !addPurchaseLine.Permissions.RequiresConfirm || !slices.Contains(addPurchaseLine.RequiredContext, "target_order") {
		t.Fatalf("expected confirmation-gated purchase line item metadata, got %+v", addPurchaseLine)
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
	if _, err := a.db.Exec(`INSERT INTO provider_health(provider, status, message, retry_after_seconds, updated_at) VALUES ('ebay', 'error', 'credentials expired; refresh token required', 0, CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("seed provider health: %v", err)
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
		!strings.Contains(testConnection.Body.String(), `"connection_status":"needs_reauthentication"`) ||
		!strings.Contains(testConnection.Body.String(), `"provider_health"`) ||
		!strings.Contains(testConnection.Body.String(), `"next_action":"check_provider_health_and_credentials"`) ||
		strings.Contains(testConnection.Body.String(), "must-not-leak") {
		t.Fatalf("expected non-mutating provider health test without secret leak, body=%s", testConnection.Body.String())
	}

	configure := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.integrations.configure_provider",
		"confirm":true,
		"source_surface":"settings.integrations.provider.card",
		"source_channel":"telegram",
		"source_thread_id":"integration-thread-1",
		"source_message_id":"integration-message-1",
		"parameters":{"provider_id":"ebay","provider_secret":"must-not-leak","setup_step":"oauth"}
	}`), map[string]string{"Content-Type": "application/json"})
	if configure.Code != http.StatusOK {
		t.Fatalf("configure provider apply status=%d body=%s", configure.Code, configure.Body.String())
	}
	if !strings.Contains(configure.Body.String(), `"mutation_applied":true`) ||
		!strings.Contains(configure.Body.String(), `"source_surface":"settings.integrations.provider.card"`) ||
		!strings.Contains(configure.Body.String(), `"source_channel":"telegram"`) ||
		!strings.Contains(configure.Body.String(), `"source_thread_id":"integration-thread-1"`) ||
		!strings.Contains(configure.Body.String(), `"source_message_id":"integration-message-1"`) ||
		!strings.Contains(configure.Body.String(), `"operation":"integrations.provider.configure"`) ||
		!strings.Contains(configure.Body.String(), `"secret_redacted":true`) ||
		!strings.Contains(configure.Body.String(), `"secret_persisted":true`) ||
		strings.Contains(configure.Body.String(), "must-not-leak") {
		t.Fatalf("expected confirmed provider configure result without secret leak, body=%s", configure.Body.String())
	}
	assertProfileSetting(t, a, p.ID, "integration.ebay.enabled", "true")
	assertProfileSetting(t, a, p.ID, "integration.ebay.setup_step", "oauth")

	repair := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.integrations.repair_provider",
		"confirm":true,
		"parameters":{"provider_name":"ebay"}
	}`), map[string]string{"Content-Type": "application/json"})
	if repair.Code != http.StatusOK {
		t.Fatalf("repair provider apply status=%d body=%s", repair.Code, repair.Body.String())
	}
	if !strings.Contains(repair.Body.String(), `"mutation_applied":true`) ||
		!strings.Contains(repair.Body.String(), `"operation":"integrations.provider.repair"`) ||
		!strings.Contains(repair.Body.String(), `"external_write_claimed":false`) {
		t.Fatalf("expected confirmed provider repair result without external write claim, body=%s", repair.Body.String())
	}

	disable := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.integrations.disable_provider",
		"confirm":true,
		"parameters":{"provider_id":"ebay"}
	}`), map[string]string{"Content-Type": "application/json"})
	if disable.Code != http.StatusOK {
		t.Fatalf("disable provider apply status=%d body=%s", disable.Code, disable.Body.String())
	}
	if !strings.Contains(disable.Body.String(), `"mutation_applied":true`) ||
		!strings.Contains(disable.Body.String(), `"operation":"integrations.provider.disable"`) ||
		!strings.Contains(disable.Body.String(), `"external_write_claimed":false`) ||
		!strings.Contains(disable.Body.String(), `"settings_persisted":["`) {
		t.Fatalf("expected confirmed provider disable result without external write claim, body=%s", disable.Body.String())
	}
	assertProfileSetting(t, a, p.ID, "integration.ebay.enabled", "false")

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
		!strings.Contains(appearance.Body.String(), `"setting_key":"theme"`) ||
		!strings.Contains(appearance.Body.String(), `"settings_persisted":["`) {
		t.Fatalf("expected confirmed appearance setting result, body=%s", appearance.Body.String())
	}
	assertProfileSetting(t, a, p.ID, "theme", "dark")

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

	configureBackup := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.storage.configure_backup",
		"confirm":true,
		"parameters":{"backup_target":"backups/nightly"}
	}`), map[string]string{"Content-Type": "application/json"})
	if configureBackup.Code != http.StatusOK {
		t.Fatalf("configure backup apply status=%d body=%s", configureBackup.Code, configureBackup.Body.String())
	}
	if !strings.Contains(configureBackup.Body.String(), `"mutation_applied":true`) ||
		!strings.Contains(configureBackup.Body.String(), `"operation":"storage.backup.configure"`) ||
		!strings.Contains(configureBackup.Body.String(), `"settings_persisted":["`) {
		t.Fatalf("expected confirmed backup settings persistence result, body=%s", configureBackup.Body.String())
	}
	assertProfileSetting(t, a, p.ID, "storage.backup_target", "backups/nightly")

	exportBundle := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.data.export_bundle",
		"parameters":{"export_scope":"profile"}
	}`), map[string]string{"Content-Type": "application/json"})
	if exportBundle.Code != http.StatusOK {
		t.Fatalf("export bundle apply status=%d body=%s", exportBundle.Code, exportBundle.Body.String())
	}
	if !strings.Contains(exportBundle.Body.String(), `"mutation_applied":false`) ||
		!strings.Contains(exportBundle.Body.String(), `"operation":"data.export.bundle"`) ||
		!strings.Contains(exportBundle.Body.String(), `"read_only":true`) {
		t.Fatalf("expected read-only export bundle result, body=%s", exportBundle.Body.String())
	}

	importFile := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.data.import_file",
		"confirm":true,
		"parameters":{"file_path":"imports/profile.json"}
	}`), map[string]string{"Content-Type": "application/json"})
	if importFile.Code != http.StatusOK {
		t.Fatalf("import file apply status=%d body=%s", importFile.Code, importFile.Body.String())
	}
	if !strings.Contains(importFile.Body.String(), `"mutation_applied":true`) ||
		!strings.Contains(importFile.Body.String(), `"operation":"data.import.file"`) ||
		!strings.Contains(importFile.Body.String(), `"impact":"import_preview_confirmed"`) {
		t.Fatalf("expected confirmed import preview result, body=%s", importFile.Body.String())
	}

	restore := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.data.restore_backup",
		"confirm":true,
		"parameters":{"backup_path":"backups/cabinet-backup.zip"}
	}`), map[string]string{"Content-Type": "application/json"})
	if restore.Code != http.StatusOK {
		t.Fatalf("restore backup apply status=%d body=%s", restore.Code, restore.Body.String())
	}
	if !strings.Contains(restore.Body.String(), `"mutation_applied":true`) ||
		!strings.Contains(restore.Body.String(), `"operation":"data.backup.restore"`) ||
		!strings.Contains(restore.Body.String(), `"destructive_confirmation":true`) {
		t.Fatalf("expected destructive restore confirmation result, body=%s", restore.Body.String())
	}
}

func TestAgentSkillApplyAPICapturesStubbedProviderWritePathEvidence(t *testing.T) {
	t.Parallel()

	const providerSecret = "issue-1780-provider-secret-must-not-leak"

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Provider Evidence"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO provider_health(provider, status, message, retry_after_seconds, updated_at) VALUES ('openai', 'auth_missing', 'missing credential: configure provider API key', 0, CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("seed provider health: %v", err)
	}

	testConnection := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.integrations.test_connection",
		"source_surface":"settings.integrations.provider.card",
		"source_channel":"in-app",
		"parameters":{"provider_id":"openai","provider_secret":"`+providerSecret+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if testConnection.Code != http.StatusOK {
		t.Fatalf("test connection status=%d body=%s", testConnection.Code, testConnection.Body.String())
	}
	requireBodyOmitsSecret(t, testConnection.Body.String(), providerSecret)
	for _, want := range []string{
		`"mutation_applied":false`,
		`"operation":"integrations.provider.test_connection"`,
		`"connection_status":"needs_setup"`,
		`"guidance":"Save provider credentials and marketplace setup before running Market Watch."`,
		`"next_action":"review_provider_status"`,
	} {
		if !strings.Contains(testConnection.Body.String(), want) {
			t.Fatalf("provider readiness response missing %s: body=%s", want, testConnection.Body.String())
		}
	}

	configure := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.integrations.configure_provider",
		"confirm":true,
		"source_surface":"settings.integrations.provider.card",
		"source_channel":"in-app",
		"parameters":{
			"provider_id":"openai",
			"provider_secret":"`+providerSecret+`",
			"setup_step":"api-key",
			"base_url":"https://api.openai.test",
			"marketplace":"global",
			"items_per_page":"25"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if configure.Code != http.StatusOK {
		t.Fatalf("configure provider status=%d body=%s", configure.Code, configure.Body.String())
	}
	configureBody := configure.Body.String()
	requireBodyOmitsSecret(t, configureBody, providerSecret)
	for _, want := range []string{
		`"mutation_applied":true`,
		`"operation":"integrations.provider.configure"`,
		`"secret_redacted":true`,
		`"secret_persisted":true`,
		`"external_write_claimed":false`,
		`"source_surface":"settings.integrations.provider.card"`,
		`"source_channel":"in-app"`,
		`"next_action":"Run provider health validation from Integrations before routing live provider workflows."`,
	} {
		if !strings.Contains(configureBody, want) {
			t.Fatalf("configure provider response missing %s: body=%s", want, configureBody)
		}
	}
	assertProfileSetting(t, a, p.ID, "integration.openai.enabled", "true")
	assertProfileSetting(t, a, p.ID, "integration.openai.setup_step", "api-key")
	assertProfileSetting(t, a, p.ID, "integration.openai.base_url", "https://api.openai.test")
	assertProfileSetting(t, a, p.ID, "integration.openai.marketplace", "global")
	assertProfileSetting(t, a, p.ID, "integration.openai.items_per_page", "25")
	assertProfileSecret(t, a, p.ID, "integration.openai.token", providerSecret)

	repair := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.integrations.repair_provider",
		"confirm":true,
		"parameters":{"provider_id":"openai"}
	}`), map[string]string{"Content-Type": "application/json"})
	if repair.Code != http.StatusOK {
		t.Fatalf("repair provider status=%d body=%s", repair.Code, repair.Body.String())
	}
	requireBodyOmitsSecret(t, repair.Body.String(), providerSecret)
	if !strings.Contains(repair.Body.String(), `"operation":"integrations.provider.repair"`) ||
		!strings.Contains(repair.Body.String(), `"next_action":"Run a provider health check after reviewing repaired setup steps."`) ||
		!strings.Contains(repair.Body.String(), `"external_write_claimed":false`) {
		t.Fatalf("expected actionable repair result without external write claim, body=%s", repair.Body.String())
	}
	assertProfileSetting(t, a, p.ID, "integration.openai.repair_status", "reviewed")

	disable := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.integrations.disable_provider",
		"confirm":true,
		"parameters":{"provider_id":"openai"}
	}`), map[string]string{"Content-Type": "application/json"})
	if disable.Code != http.StatusOK {
		t.Fatalf("disable provider status=%d body=%s", disable.Code, disable.Body.String())
	}
	requireBodyOmitsSecret(t, disable.Body.String(), providerSecret)
	if !strings.Contains(disable.Body.String(), `"operation":"integrations.provider.disable"`) ||
		!strings.Contains(disable.Body.String(), `"next_action":"Confirm provider disabled state in Integrations before routing provider-backed workflows."`) ||
		!strings.Contains(disable.Body.String(), `"external_write_claimed":false`) {
		t.Fatalf("expected actionable disable result without external write claim, body=%s", disable.Body.String())
	}
	assertProfileSetting(t, a, p.ID, "integration.openai.enabled", "false")
}

func TestAgentSkillApplyAPIHandlesMarketWatchAndPurchasesSkills(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Market Watch Purchases"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO provider_health(provider, status, message, retry_after_seconds, updated_at) VALUES ('ebay', 'auth_missing', 'provider credentials required before live watch run', 0, CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("seed provider health: %v", err)
	}

	missingProvider := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.market_watch.run_watch",
		"parameters":{"watch_id":"watch-42"}
	}`), map[string]string{"Content-Type": "application/json"})
	if missingProvider.Code != http.StatusOK {
		t.Fatalf("market watch preview status=%d body=%s", missingProvider.Code, missingProvider.Body.String())
	}
	if !strings.Contains(missingProvider.Body.String(), `"blocker":"market_watch_provider_required"`) ||
		!strings.Contains(missingProvider.Body.String(), `"mutation_applied":false`) {
		t.Fatalf("expected provider blocker without mutation, body=%s", missingProvider.Body.String())
	}

	createWatch := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.market_watch.create_saved_watch",
		"confirm":true,
		"parameters":{"provider_id":"ebay","watch_name":"Agent boxed kits","watch_query":"boxed model kit","region":"AU","enabled":true}
	}`), map[string]string{"Content-Type": "application/json"})
	if createWatch.Code != http.StatusOK {
		t.Fatalf("create saved watch status=%d body=%s", createWatch.Code, createWatch.Body.String())
	}
	for _, want := range []string{
		`"operation":"market_watch.watch.create"`,
		`"watch_persisted":true`,
		`"watch_query":"boxed model kit"`,
		`"provider_scope":["ebay"]`,
	} {
		if !strings.Contains(createWatch.Body.String(), want) {
			t.Fatalf("create watch response missing %s: body=%s", want, createWatch.Body.String())
		}
	}
	var watchID string
	if err := a.db.QueryRow(`SELECT id FROM scanner_query_sets WHERE profile_id = ? AND name = 'Agent boxed kits'`, p.ID).Scan(&watchID); err != nil {
		t.Fatalf("load created saved watch: %v", err)
	}

	updateWatch := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.market_watch.update_saved_watch",
		"confirm":true,
		"parameters":{"provider_id":"ebay","watch_id":"`+watchID+`","watch_name":"Agent boxed kits under 100","keywords":["boxed","kit"],"max_price":100}
	}`), map[string]string{"Content-Type": "application/json"})
	if updateWatch.Code != http.StatusOK {
		t.Fatalf("update saved watch status=%d body=%s", updateWatch.Code, updateWatch.Body.String())
	}
	if !strings.Contains(updateWatch.Body.String(), `"operation":"market_watch.watch.update"`) ||
		!strings.Contains(updateWatch.Body.String(), `"watch_persisted":true`) ||
		!strings.Contains(updateWatch.Body.String(), `"max_price":100`) {
		t.Fatalf("expected persisted saved watch update evidence, body=%s", updateWatch.Body.String())
	}

	searchWatches := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.market_watch.search_watches",
		"parameters":{"provider_id":"ebay","query":"under 100"}
	}`), map[string]string{"Content-Type": "application/json"})
	if searchWatches.Code != http.StatusOK {
		t.Fatalf("search saved watches status=%d body=%s", searchWatches.Code, searchWatches.Body.String())
	}
	if !strings.Contains(searchWatches.Body.String(), `"mutation_applied":false`) ||
		!strings.Contains(searchWatches.Body.String(), `"operation":"market_watch.watch.search"`) ||
		!strings.Contains(searchWatches.Body.String(), `"name":"Agent boxed kits under 100"`) {
		t.Fatalf("expected read-only saved watch reload evidence, body=%s", searchWatches.Body.String())
	}

	runWatch := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.market_watch.run_watch",
		"confirm":true,
		"source_surface":"market_watch.saved_watch.row",
		"source_channel":"in-app",
		"parameters":{"provider_id":"ebay","watch_id":"`+watchID+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if runWatch.Code != http.StatusOK {
		t.Fatalf("run watch status=%d body=%s", runWatch.Code, runWatch.Body.String())
	}
	for _, want := range []string{
		`"mutation_applied":true`,
		`"operation":"market_watch.watch.run"`,
		`"provider_id":"ebay"`,
		`"watch_id":"` + watchID + `"`,
		`"provider_health"`,
		`"saved_watch"`,
		`"external_write_claimed":false`,
		`"source_surface":"market_watch.saved_watch.row"`,
	} {
		if !strings.Contains(runWatch.Body.String(), want) {
			t.Fatalf("run watch response missing %s: body=%s", want, runWatch.Body.String())
		}
	}

	handoff := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.market_watch.handoff_result",
		"confirm":true,
		"parameters":{"provider_id":"ebay","result_id":"result-9","destination":"purchases","source_url":"https://example.test/listing/9"}
	}`), map[string]string{"Content-Type": "application/json"})
	if handoff.Code != http.StatusOK {
		t.Fatalf("handoff status=%d body=%s", handoff.Code, handoff.Body.String())
	}
	if !strings.Contains(handoff.Body.String(), `"provenance_preserved":true`) ||
		!strings.Contains(handoff.Body.String(), `"result_id":"result-9"`) ||
		!strings.Contains(handoff.Body.String(), `"destination":"purchases"`) {
		t.Fatalf("expected provenance-preserving handoff result, body=%s", handoff.Body.String())
	}

	missingOrder := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.purchases.add_line_item",
		"parameters":{"item_id":"item-99"}
	}`), map[string]string{"Content-Type": "application/json"})
	if missingOrder.Code != http.StatusOK {
		t.Fatalf("purchase preview status=%d body=%s", missingOrder.Code, missingOrder.Body.String())
	}
	if !strings.Contains(missingOrder.Body.String(), `"blocker":"purchases_order_required"`) {
		t.Fatalf("expected missing order blocker, body=%s", missingOrder.Body.String())
	}

	if _, err := a.db.Exec(`INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, status) VALUES ('item-99', ?, 'AFX', 'Slot Cars', 'AGENT-PURCHASE-99', 'Agent purchase line item', 'active')`, p.ID); err != nil {
		t.Fatalf("seed purchase target item: %v", err)
	}

	addLine := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.purchases.add_line_item",
		"confirm":true,
		"parameters":{
			"order_id":"order-1",
			"item_id":"item-99",
			"source":"market_watch",
			"result_id":"result-9",
			"source_url":"https://example.test/listing/9",
			"seller":"seller-one",
			"tracking":"TRACK-99",
			"quantity":2,
			"amount":42.5,
			"currency":"AUD"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if addLine.Code != http.StatusOK {
		t.Fatalf("add line item status=%d body=%s", addLine.Code, addLine.Body.String())
	}
	if !strings.Contains(addLine.Body.String(), `"operation":"purchases.order.add_line_item"`) ||
		!strings.Contains(addLine.Body.String(), `"order_id":"order-1"`) ||
		!strings.Contains(addLine.Body.String(), `"item_id":"item-99"`) ||
		!strings.Contains(addLine.Body.String(), `"purchase_persisted":true`) ||
		!strings.Contains(addLine.Body.String(), `"provenance_preserved":true`) ||
		!strings.Contains(addLine.Body.String(), `"lifecycle_entry_id":"`) ||
		!strings.Contains(addLine.Body.String(), `"expected_arrival_id":"`) {
		t.Fatalf("expected purchase add line item preview/apply evidence, body=%s", addLine.Body.String())
	}

	searchOrders := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.purchases.search_orders",
		"parameters":{"query":"order-1","status":"all"}
	}`), map[string]string{"Content-Type": "application/json"})
	if searchOrders.Code != http.StatusOK {
		t.Fatalf("search purchase orders status=%d body=%s", searchOrders.Code, searchOrders.Body.String())
	}
	for _, want := range []string{
		`"mutation_applied":false`,
		`"operation":"purchases.orders.search"`,
		`"order_id":"order-1"`,
		`"source":"market_watch"`,
		`"seller":"seller-one"`,
		`"tracking":"TRACK-99"`,
		`"line_item_count":1`,
		`"expected_arrival_id":"`,
	} {
		if !strings.Contains(searchOrders.Body.String(), want) {
			t.Fatalf("purchase order search response missing %s: body=%s", want, searchOrders.Body.String())
		}
	}

	receiveLine := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.purchases.receive_line_item",
		"confirm":true,
		"parameters":{"order_id":"order-1","line_item_id":"item-99","delivered_on":"2026-07-05","notes":"received by agent skill"}
	}`), map[string]string{"Content-Type": "application/json"})
	if receiveLine.Code != http.StatusOK {
		t.Fatalf("receive line item status=%d body=%s", receiveLine.Code, receiveLine.Body.String())
	}
	for _, want := range []string{
		`"operation":"purchases.line_item.receive"`,
		`"purchase_persisted":true`,
		`"arrival_status":"delivered"`,
		`"delivered_on":"2026-07-05"`,
	} {
		if !strings.Contains(receiveLine.Body.String(), want) {
			t.Fatalf("receive line response missing %s: body=%s", want, receiveLine.Body.String())
		}
	}

	reconcile := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.purchases.reconcile_item",
		"confirm":true,
		"parameters":{"order_id":"order-1","item_id":"item-99","instance_id":"instance-99","notes":"reconciled by agent skill"}
	}`), map[string]string{"Content-Type": "application/json"})
	if reconcile.Code != http.StatusOK {
		t.Fatalf("reconcile item status=%d body=%s", reconcile.Code, reconcile.Body.String())
	}
	for _, want := range []string{
		`"operation":"purchases.item.reconcile"`,
		`"purchase_persisted":true`,
		`"reconciliation_persisted":true`,
		`"arrival_status":"reconciled"`,
		`"reconciled_instance_id":"instance-99"`,
	} {
		if !strings.Contains(reconcile.Body.String(), want) {
			t.Fatalf("reconcile item response missing %s: body=%s", want, reconcile.Body.String())
		}
	}

	searchReceived := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.purchases.search_orders",
		"parameters":{"query":"order-1","status":"received"}
	}`), map[string]string{"Content-Type": "application/json"})
	if searchReceived.Code != http.StatusOK {
		t.Fatalf("search received purchase orders status=%d body=%s", searchReceived.Code, searchReceived.Body.String())
	}
	if !strings.Contains(searchReceived.Body.String(), `"status":"received"`) ||
		!strings.Contains(searchReceived.Body.String(), `"received_count":1`) ||
		!strings.Contains(searchReceived.Body.String(), `"status":"reconciled"`) {
		t.Fatalf("expected received/reconciled purchase order search evidence, body=%s", searchReceived.Body.String())
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

func assertProfileSetting(t *testing.T, a *App, profileID, key, want string) {
	t.Helper()
	var got string
	if err := a.db.QueryRow(`SELECT value FROM profile_settings WHERE profile_id = ? AND key = ?`, profileID, key).Scan(&got); err != nil {
		t.Fatalf("read profile setting %s: %v", key, err)
	}
	if got != want {
		t.Fatalf("profile setting %s = %q, want %q", key, got, want)
	}
}

func assertProfileSecret(t *testing.T, a *App, profileID, key, want string) {
	t.Helper()
	got, err := profile.NewRepository(a.db).GetSecret(context.Background(), profileID, key)
	if err != nil {
		t.Fatalf("read profile secret %s: %v", key, err)
	}
	if got != want {
		t.Fatalf("profile secret %s = %q, want %q", key, got, want)
	}
}

func requireBodyOmitsSecret(t *testing.T, body, secret string) {
	t.Helper()
	if strings.Contains(body, secret) {
		t.Fatalf("response leaked secret %q: body=%s", secret, body)
	}
}

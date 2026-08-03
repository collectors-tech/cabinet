package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/collectors-tech/cabinet/internal/update"
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
	Enabled         bool     `json:"enabled"`
	Executable      bool     `json:"executable"`
	NextAction      string   `json:"next_action"`
	Provenance      string   `json:"provenance"`
	Permissions     struct {
		LocalWrite      bool `json:"local_write"`
		Destructive     bool `json:"destructive"`
		RequiresConfirm bool `json:"requires_confirm"`
	} `json:"permissions"`
}

type apiCapabilityExplanationPayload struct {
	SkillID           string   `json:"skill_id"`
	DisplayName       string   `json:"display_name"`
	Status            string   `json:"status"`
	SafetyLevel       string   `json:"safety_level"`
	CapabilityState   string   `json:"capability_state"`
	ExecutionBoundary string   `json:"execution_boundary"`
	RequiredContext   []string `json:"required_context"`
	RequiredSetup     []string `json:"required_setup"`
	Authority         struct {
		Decision   string `json:"decision"`
		Blocker    string `json:"blocker"`
		NextAction string `json:"next_action"`
	} `json:"authority"`
	NextAction string `json:"next_action"`
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
	inventorySearch := findAPISkill(payload.Skills, "cabinet.inventory.search_items")
	if inventorySearch == nil {
		t.Fatalf("missing inventory search skill")
	}
	if inventorySearch.SafetyLevel != "read-only" || inventorySearch.Permissions.LocalWrite || !inventorySearch.Executable {
		t.Fatalf("expected executable read-only Inventory search metadata, got %+v", inventorySearch)
	}
	inventoryAttach := findAPISkill(payload.Skills, "cabinet.inventory.attach_media")
	if inventoryAttach == nil {
		t.Fatalf("missing inventory attach media skill")
	}
	if inventoryAttach.SafetyLevel != "confirm-required" || !inventoryAttach.Permissions.RequiresConfirm || !slices.Contains(inventoryAttach.RequiredContext, "selected_media") {
		t.Fatalf("expected confirmation-gated Inventory media metadata, got %+v", inventoryAttach)
	}
	inventoryAssign := findAPISkill(payload.Skills, "cabinet.inventory.assign_to_collection")
	if inventoryAssign == nil {
		t.Fatalf("missing inventory assign to collection skill")
	}
	if inventoryAssign.SafetyLevel != "confirm-required" || !inventoryAssign.Permissions.RequiresConfirm || !slices.Contains(inventoryAssign.RequiredContext, "collection") {
		t.Fatalf("expected confirmation-gated Inventory collection metadata, got %+v", inventoryAssign)
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

	wishlistSearch := findAPISkill(payload.Skills, "cabinet.wishlist.search_entries")
	if wishlistSearch == nil {
		t.Fatalf("missing Wishlist search skill")
	}
	if wishlistSearch.SafetyLevel != "read-only" || wishlistSearch.Permissions.LocalWrite || !wishlistSearch.Executable {
		t.Fatalf("expected executable read-only Wishlist search metadata, got %+v", wishlistSearch)
	}

	wishlistPurchased := findAPISkill(payload.Skills, "cabinet.wishlist.mark_purchased")
	if wishlistPurchased == nil {
		t.Fatalf("missing Wishlist mark purchased skill")
	}
	if wishlistPurchased.SafetyLevel != "confirm-required" || !wishlistPurchased.Permissions.RequiresConfirm || !slices.Contains(wishlistPurchased.RequiredContext, "purchase_details") {
		t.Fatalf("expected confirmation-gated Wishlist purchased metadata, got %+v", wishlistPurchased)
	}

	collectionsSearch := findAPISkill(payload.Skills, "cabinet.collections.search")
	if collectionsSearch == nil {
		t.Fatalf("missing Collections search skill")
	}
	if collectionsSearch.SafetyLevel != "read-only" || collectionsSearch.Permissions.LocalWrite || !collectionsSearch.Executable {
		t.Fatalf("expected executable read-only Collections search metadata, got %+v", collectionsSearch)
	}

	collectionDelete := findAPISkill(payload.Skills, "cabinet.collections.soft_delete")
	if collectionDelete == nil {
		t.Fatalf("missing Collections soft delete skill")
	}
	if collectionDelete.SafetyLevel != "confirm-required" || !collectionDelete.Permissions.RequiresConfirm || !slices.Contains(collectionDelete.RequiredContext, "collection") {
		t.Fatalf("expected confirmation-gated Collections delete metadata, got %+v", collectionDelete)
	}
}

func TestAgentCapabilityExplanationDerivesFromRegistryAndProfileAuthority(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Capability Explanation"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	root := writeAgentSkillImportFixture(t, validAgentSkillImportManifest(""))
	importResp := doRequest(t, a, http.MethodPost, "/api/agent/skills/import", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"source_type":"folder",
		"path":`+strconv.Quote(root)+`
	}`), map[string]string{"Content-Type": "application/json"})
	if importResp.Code != http.StatusOK {
		t.Fatalf("import skill status=%d body=%s", importResp.Code, importResp.Body.String())
	}
	disableResp := doRequest(t, a, http.MethodPost, "/api/agent/skills/state", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.example.api_imported_reader",
		"enabled":false
	}`), map[string]string{"Content-Type": "application/json"})
	if disableResp.Code != http.StatusOK {
		t.Fatalf("disable imported skill status=%d body=%s", disableResp.Code, disableResp.Body.String())
	}

	repo := profile.NewRepository(a.db)
	if _, err := repo.PutAgentAuthorityPolicy(context.Background(), p.ID, profile.AgentAuthorityPolicy{
		Mode:                  profile.AgentAuthorityModeReadOnly,
		ExternalWriteApproved: false,
	}); err != nil {
		t.Fatalf("set read-only authority policy: %v", err)
	}

	readonly := doRequest(t, a, http.MethodGet, "/api/agent/capabilities?profile_id="+p.ID, nil, nil)
	if readonly.Code != http.StatusOK {
		t.Fatalf("capabilities status=%d body=%s", readonly.Code, readonly.Body.String())
	}
	var locked struct {
		ProfileID      string                            `json:"profile_id"`
		AuthorityMode  string                            `json:"authority_mode"`
		CapabilityHelp []apiCapabilityExplanationPayload `json:"capabilities"`
	}
	if err := json.NewDecoder(readonly.Body).Decode(&locked); err != nil {
		t.Fatalf("decode capabilities: %v", err)
	}
	if locked.ProfileID != p.ID || locked.AuthorityMode != string(profile.AgentAuthorityModeReadOnly) {
		t.Fatalf("expected profile and read-only authority echo, got %+v", locked)
	}
	inventorySearch := findCapabilityExplanation(locked.CapabilityHelp, "cabinet.inventory.search_items")
	if inventorySearch == nil || inventorySearch.CapabilityState != "available" || inventorySearch.ExecutionBoundary != "read_only" || inventorySearch.Authority.Decision != "allowed" {
		t.Fatalf("expected read-only inventory skill to be available and allowed, got %+v", inventorySearch)
	}
	inventoryCreate := findCapabilityExplanation(locked.CapabilityHelp, "cabinet.inventory.create_item")
	if inventoryCreate == nil || inventoryCreate.CapabilityState != "blocked_by_policy" || inventoryCreate.Authority.Blocker != "agent_authority_read_only" || inventoryCreate.ExecutionBoundary != "preview_then_confirm" {
		t.Fatalf("expected local-write inventory skill to be policy-blocked in read-only mode, got %+v", inventoryCreate)
	}
	externalWrite := findCapabilityExplanation(locked.CapabilityHelp, "cabinet.integrations.configure_provider")
	if externalWrite == nil || externalWrite.ExecutionBoundary != "external_write_confirmation" || externalWrite.CapabilityState != "blocked_by_policy" || len(externalWrite.RequiredSetup) == 0 {
		t.Fatalf("expected external-write setup skill to show setup and policy blocker, got %+v", externalWrite)
	}
	unimplemented := findCapabilityExplanation(locked.CapabilityHelp, "cabinet.guided.inventory.update_item")
	if unimplemented == nil || unimplemented.CapabilityState != "unavailable" || unimplemented.Status != "requires-implementation" {
		t.Fatalf("expected unimplemented guided skill to remain visible as unavailable, got %+v", unimplemented)
	}
	disabled := findCapabilityExplanation(locked.CapabilityHelp, "cabinet.example.api_imported_reader")
	if disabled == nil || disabled.CapabilityState != "disabled" || disabled.NextAction == "" {
		t.Fatalf("expected disabled imported skill explanation, got %+v", disabled)
	}
	if strings.Contains(readonly.Body.String(), root) || strings.Contains(readonly.Body.String(), "api_key") || strings.Contains(readonly.Body.String(), "preview_id") {
		t.Fatalf("capability explanation leaked local path or hidden values: %s", readonly.Body.String())
	}

	if _, err := repo.PutAgentAuthorityPolicy(context.Background(), p.ID, profile.AgentAuthorityPolicy{
		Mode:                  profile.AgentAuthorityModeAskBeforeLocalChanges,
		ExternalWriteApproved: false,
	}); err != nil {
		t.Fatalf("set ask-before authority policy: %v", err)
	}
	unlocked := doRequest(t, a, http.MethodGet, "/api/agent/capabilities?profile_id="+p.ID, nil, nil)
	if unlocked.Code != http.StatusOK {
		t.Fatalf("capabilities after policy update status=%d body=%s", unlocked.Code, unlocked.Body.String())
	}
	var changed struct {
		CapabilityHelp []apiCapabilityExplanationPayload `json:"capabilities"`
	}
	if err := json.NewDecoder(unlocked.Body).Decode(&changed); err != nil {
		t.Fatalf("decode changed capabilities: %v", err)
	}
	changedCreate := findCapabilityExplanation(changed.CapabilityHelp, "cabinet.inventory.create_item")
	if changedCreate == nil || changedCreate.CapabilityState != "confirm_required" || changedCreate.Authority.Blocker != "confirmation_required" {
		t.Fatalf("expected authority mode change to make local write previewable with confirmation, got %+v", changedCreate)
	}
}

func TestAgentSkillPreviewNormalizesAgentContextEnvelope(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Context"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	resp := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inventory.update_item",
		"agent_context":{
			"profile_id":"`+p.ID+`",
			"workspace_id":"workspace-agent-skill",
			"route_id":"/inventory/item/item-agent-ctx-001",
			"surface_id":"inventory.detail",
			"selected_record":{"type":"inventory_item","id":"item-agent-ctx-001"},
			"thread_id":"thread-agent-skill",
			"intent_text":"rename this item",
			"source_channel":"in-app",
			"permission_state":"ask_before_local_changes",
			"setup_state":"ready",
			"workflow_run_id":"workflow-agent-skill",
			"audit_id":"audit-agent-skill"
		},
		"parameters":{"title":"Updated title"}
	}`), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusOK {
		t.Fatalf("preview status=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload struct {
		SourceSurface  string         `json:"source_surface"`
		SourceChannel  string         `json:"source_channel"`
		SourceThreadID string         `json:"source_thread_id"`
		Blocker        string         `json:"blocker"`
		Target         map[string]any `json:"target"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode preview: %v", err)
	}
	if payload.SourceSurface != "inventory.detail" || payload.SourceChannel != "in-app" || payload.SourceThreadID != "thread-agent-skill" {
		t.Fatalf("expected skill preview source fields from agent context, got %+v", payload)
	}
	if payload.Blocker != "confirmation_required" {
		t.Fatalf("expected hydrated selected item to advance to confirmation, got %+v", payload)
	}
	if payload.Target["item_id"] != "item-agent-ctx-001" || payload.Target["title"] != "Updated title" {
		t.Fatalf("expected selected record and params in preview target, got %+v", payload.Target)
	}
	if strings.Contains(resp.Body.String(), "audit-agent-skill") {
		t.Fatalf("preview response must not expose audit-only context ids: %s", resp.Body.String())
	}
}

func TestAgentSkillPreviewAndApplyClarifyMissingAgentContext(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Clarification"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	requestBody := `{
		"profile_id":"` + p.ID + `",
		"skill_id":"cabinet.inventory.update_item",
		"agent_context":{
			"profile_id":"` + p.ID + `",
			"workspace_id":"workspace-agent-skill",
			"thread_id":"thread-agent-skill",
			"source_channel":"in-app",
			"setup_state":"setup_needed"
		},
		"parameters":{"title":"Updated title"}
	}`

	resp := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(requestBody), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusConflict {
		t.Fatalf("preview status=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload struct {
		Error          string            `json:"error"`
		MissingContext []string          `json:"missing_context"`
		NextAction     string            `json:"next_action"`
		Clarification  map[string]string `json:"clarification"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode clarification: %v", err)
	}
	if payload.Error != "missing_context" || !slices.Contains(payload.MissingContext, "selected_item") ||
		!slices.Contains(payload.MissingContext, "route") || !slices.Contains(payload.MissingContext, "setup_state") {
		t.Fatalf("expected selected item, route, and setup clarification, got %+v", payload)
	}
	if payload.Clarification["selected_item"] == "" || payload.Clarification["route"] == "" || payload.Clarification["setup_state"] == "" || payload.NextAction == "" {
		t.Fatalf("expected actionable clarification guidance, got %+v", payload)
	}
	if strings.Contains(resp.Body.String(), "direct-api") || strings.Contains(resp.Body.String(), "audit") {
		t.Fatalf("clarification must not invent direct-api targets or leak audit context: %s", resp.Body.String())
	}

	applyResp := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(requestBody), map[string]string{"Content-Type": "application/json"})
	if applyResp.Code != http.StatusConflict {
		t.Fatalf("apply status=%d body=%s", applyResp.Code, applyResp.Body.String())
	}
	var applyPayload struct {
		Error          string            `json:"error"`
		MissingContext []string          `json:"missing_context"`
		NextAction     string            `json:"next_action"`
		Clarification  map[string]string `json:"clarification"`
	}
	if err := json.NewDecoder(applyResp.Body).Decode(&applyPayload); err != nil {
		t.Fatalf("decode apply clarification: %v", err)
	}
	if applyPayload.Error != "missing_context" || !slices.Contains(applyPayload.MissingContext, "selected_item") ||
		!slices.Contains(applyPayload.MissingContext, "route") || !slices.Contains(applyPayload.MissingContext, "setup_state") {
		t.Fatalf("expected apply selected item, route, and setup clarification, got %+v", applyPayload)
	}
	if applyPayload.Clarification["selected_item"] == "" || applyPayload.Clarification["route"] == "" || applyPayload.Clarification["setup_state"] == "" || applyPayload.NextAction == "" {
		t.Fatalf("expected actionable apply clarification guidance, got %+v", applyPayload)
	}
	if strings.Contains(applyResp.Body.String(), "direct-api") || strings.Contains(applyResp.Body.String(), "audit") {
		t.Fatalf("apply clarification must not invent direct-api targets or leak audit context: %s", applyResp.Body.String())
	}
}

func TestAgentSkillImportAPIInstallsLocalFolderDisabledAndListsMetadata(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	root := writeAgentSkillImportFixture(t, validAgentSkillImportManifest(`{
		"id": "cabinet.example.api_imported_writer",
		"safetyLevel": "confirm-required",
		"status": "preview-only",
		"capabilities": ["inventory.item.update"],
		"guidedWorkflows": ["inventory.item.update"],
		"uiTargets": ["inventory.item.title"],
		"permissions": {
			"cabinetReads": ["inventory.item"],
			"cabinetWrites": ["inventory.item"],
			"externalReads": [],
			"externalWrites": [],
			"secretAccess": false,
			"destructive": false
		},
		"audit": {
			"actionTimeline": "records selected inventory item",
			"requiresConfirmation": true
		}
	}`))

	resp := doRequest(t, a, http.MethodPost, "/api/agent/skills/import", strings.NewReader(`{
		"profile_id":"profile-a",
		"source_type":"folder",
		"path":`+strconv.Quote(root)+`
	}`), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusOK {
		t.Fatalf("import status=%d body=%s", resp.Code, resp.Body.String())
	}
	var importPayload struct {
		State          string `json:"state"`
		Provenance     string `json:"provenance"`
		InstalledState struct {
			ProfileID string `json:"profile_id"`
			SkillID   string `json:"skill_id"`
			Status    string `json:"status"`
			Enabled   bool   `json:"enabled"`
		} `json:"installed_state"`
		Skill apiSkillPayload `json:"skill"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&importPayload); err != nil {
		t.Fatalf("decode import: %v", err)
	}
	if importPayload.State != "installed-disabled" || importPayload.InstalledState.Enabled || importPayload.InstalledState.Status != "disabled" {
		t.Fatalf("expected disabled import result, got %+v", importPayload)
	}
	if importPayload.Provenance != root || importPayload.Skill.Provenance != root {
		t.Fatalf("expected source provenance, got result=%q skill=%q", importPayload.Provenance, importPayload.Skill.Provenance)
	}

	listResp := doRequest(t, a, http.MethodGet, "/api/agent/skills?profile_id=profile-a", nil, nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listResp.Code, listResp.Body.String())
	}
	var listPayload struct {
		Skills []apiSkillPayload `json:"skills"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	imported := findAPISkill(listPayload.Skills, "cabinet.example.api_imported_writer")
	if imported == nil {
		t.Fatalf("expected imported skill in profile registry")
	}
	if imported.Source != "archive" || imported.Status != "disabled" || imported.Executable || imported.Enabled || !imported.Removable || imported.BuiltIn {
		t.Fatalf("expected disabled removable archive metadata, got %+v", imported)
	}
	if imported.Provenance != root || imported.NextAction == "" {
		t.Fatalf("expected provenance and next action, got %+v", imported)
	}
}

func TestAgentSkillImportAPIPersistsInstalledMetadataAcrossRestart(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	cfg := config.Config{
		Addr:           "127.0.0.1:0",
		DataDir:        base,
		DBPath:         filepath.Join(base, "cabinet.db"),
		UpdateChannel:  update.ChannelStable,
		WebAuthnRPID:   "127.0.0.1",
		WebAuthnOrigin: "http://127.0.0.1:8080",
		WebAuthnName:   "Cabinet Test",
		BackupInterval: 60,
	}
	a := newTestAppWithConfig(t, cfg)
	root := writeAgentSkillImportFixture(t, validAgentSkillImportManifest(`{
		"id": "cabinet.example.api_imported_restart",
		"safetyLevel": "confirm-required",
		"status": "preview-only",
		"capabilities": ["inventory.item.update"],
		"guidedWorkflows": ["inventory.item.update"],
		"uiTargets": ["inventory.item.title"],
		"permissions": {
			"cabinetReads": ["inventory.item"],
			"cabinetWrites": ["inventory.item"],
			"externalReads": [],
			"externalWrites": [],
			"secretAccess": false,
			"destructive": false
		},
		"audit": {
			"actionTimeline": "records restart-safe local import metadata",
			"requiresConfirmation": true
		}
	}`))

	resp := doRequest(t, a, http.MethodPost, "/api/agent/skills/import", strings.NewReader(`{
		"profile_id":"profile-a",
		"source_type":"folder",
		"path":`+strconv.Quote(root)+`
	}`), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusOK {
		t.Fatalf("import status=%d body=%s", resp.Code, resp.Body.String())
	}
	if err := a.close(); err != nil {
		t.Fatalf("close app before restart: %v", err)
	}

	restarted := newTestAppWithConfig(t, cfg)
	listResp := doRequest(t, restarted, http.MethodGet, "/api/agent/skills?profile_id=profile-a", nil, nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list after restart status=%d body=%s", listResp.Code, listResp.Body.String())
	}
	var listPayload struct {
		Skills []apiSkillPayload `json:"skills"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode restarted list: %v", err)
	}
	imported := findAPISkill(listPayload.Skills, "cabinet.example.api_imported_restart")
	if imported == nil {
		t.Fatalf("expected imported skill after app restart")
	}
	if imported.Status != "disabled" || imported.Executable || imported.Enabled || imported.Provenance != root {
		t.Fatalf("expected durable disabled archive metadata after restart, got %+v", imported)
	}
}

func TestAgentSkillStateAPIEnablesAndDisablesImportedSkill(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	root := writeAgentSkillImportFixture(t, validAgentSkillImportManifest(`{
		"id": "cabinet.example.api_imported_toggle",
		"safetyLevel": "confirm-required",
		"status": "preview-only",
		"capabilities": ["inventory.item.update"],
		"guidedWorkflows": ["inventory.item.update"],
		"uiTargets": ["inventory.item.title"],
		"permissions": {
			"cabinetReads": ["inventory.item"],
			"cabinetWrites": ["inventory.item"],
			"externalReads": [],
			"externalWrites": [],
			"secretAccess": false,
			"destructive": false
		},
		"audit": {
			"actionTimeline": "records imported skill state changes",
			"requiresConfirmation": true
		}
	}`))

	importResp := doRequest(t, a, http.MethodPost, "/api/agent/skills/import", strings.NewReader(`{
		"profile_id":"profile-a",
		"source_type":"folder",
		"path":`+strconv.Quote(root)+`
	}`), map[string]string{"Content-Type": "application/json"})
	if importResp.Code != http.StatusOK {
		t.Fatalf("import status=%d body=%s", importResp.Code, importResp.Body.String())
	}

	enableResp := doRequest(t, a, http.MethodPost, "/api/agent/skills/state", strings.NewReader(`{
		"profile_id":"profile-a",
		"skill_id":"cabinet.example.api_imported_toggle",
		"enabled":true
	}`), map[string]string{"Content-Type": "application/json"})
	if enableResp.Code != http.StatusOK {
		t.Fatalf("enable status=%d body=%s", enableResp.Code, enableResp.Body.String())
	}
	var enablePayload struct {
		State struct {
			ProfileID string `json:"profile_id"`
			SkillID   string `json:"skill_id"`
			Status    string `json:"status"`
			Enabled   bool   `json:"enabled"`
		} `json:"state"`
		Skill apiSkillPayload `json:"skill"`
	}
	if err := json.NewDecoder(enableResp.Body).Decode(&enablePayload); err != nil {
		t.Fatalf("decode enable: %v", err)
	}
	if enablePayload.State.ProfileID != "profile-a" || enablePayload.State.SkillID != "cabinet.example.api_imported_toggle" || !enablePayload.State.Enabled {
		t.Fatalf("expected enabled profile state, got %+v", enablePayload.State)
	}
	if !enablePayload.Skill.Enabled || enablePayload.Skill.Status != "available" || !enablePayload.Skill.Executable {
		t.Fatalf("expected enabled executable imported skill, got %+v", enablePayload.Skill)
	}

	disableResp := doRequest(t, a, http.MethodPost, "/api/agent/skills/state", strings.NewReader(`{
		"profile_id":"profile-a",
		"skill_id":"cabinet.example.api_imported_toggle",
		"enabled":false
	}`), map[string]string{"Content-Type": "application/json"})
	if disableResp.Code != http.StatusOK {
		t.Fatalf("disable status=%d body=%s", disableResp.Code, disableResp.Body.String())
	}
	listResp := doRequest(t, a, http.MethodGet, "/api/agent/skills?profile_id=profile-a", nil, nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listResp.Code, listResp.Body.String())
	}
	var listPayload struct {
		Skills []apiSkillPayload `json:"skills"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	imported := findAPISkill(listPayload.Skills, "cabinet.example.api_imported_toggle")
	if imported == nil {
		t.Fatalf("expected imported skill after state change")
	}
	if imported.Enabled || imported.Status != "disabled" || imported.Executable {
		t.Fatalf("expected disabled non-executable imported skill, got %+v", imported)
	}
}

func TestAgentSkillStateAPIBlocksBuiltInAndHighRiskWithoutConfirmation(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	builtInResp := doRequest(t, a, http.MethodPost, "/api/agent/skills/state", strings.NewReader(`{
		"profile_id":"profile-a",
		"skill_id":"cabinet.inventory.search_items",
		"enabled":false
	}`), map[string]string{"Content-Type": "application/json"})
	if builtInResp.Code != http.StatusConflict || !strings.Contains(builtInResp.Body.String(), "built_in_skill_state_locked") {
		t.Fatalf("expected built-in lock conflict, status=%d body=%s", builtInResp.Code, builtInResp.Body.String())
	}

	root := writeAgentSkillImportFixture(t, validAgentSkillImportManifest(`{
		"id": "cabinet.example.api_imported_external",
		"safetyLevel": "external-write",
		"status": "preview-only",
		"capabilities": ["inventory.item.update"],
		"permissions": {
			"cabinetReads": ["inventory.item"],
			"cabinetWrites": ["inventory.item"],
			"externalReads": [],
			"externalWrites": ["provider.configure"],
			"secretAccess": false,
			"destructive": false
		},
		"audit": {
			"actionTimeline": "records external write confirmation",
			"requiresConfirmation": true
		}
	}`))
	importResp := doRequest(t, a, http.MethodPost, "/api/agent/skills/import", strings.NewReader(`{
		"profile_id":"profile-a",
		"source_type":"folder",
		"path":`+strconv.Quote(root)+`
	}`), map[string]string{"Content-Type": "application/json"})
	if importResp.Code != http.StatusOK {
		t.Fatalf("import status=%d body=%s", importResp.Code, importResp.Body.String())
	}
	withoutConfirm := doRequest(t, a, http.MethodPost, "/api/agent/skills/state", strings.NewReader(`{
		"profile_id":"profile-a",
		"skill_id":"cabinet.example.api_imported_external",
		"enabled":true
	}`), map[string]string{"Content-Type": "application/json"})
	if withoutConfirm.Code != http.StatusConflict || !strings.Contains(withoutConfirm.Body.String(), "strong_confirmation_required") {
		t.Fatalf("expected strong confirmation conflict, status=%d body=%s", withoutConfirm.Code, withoutConfirm.Body.String())
	}
	withConfirm := doRequest(t, a, http.MethodPost, "/api/agent/skills/state", strings.NewReader(`{
		"profile_id":"profile-a",
		"skill_id":"cabinet.example.api_imported_external",
		"enabled":true,
		"confirm":true
	}`), map[string]string{"Content-Type": "application/json"})
	if withConfirm.Code != http.StatusOK {
		t.Fatalf("confirmed enable status=%d body=%s", withConfirm.Code, withConfirm.Body.String())
	}
}

func TestAgentSkillImportAPIRejectsInvalidFolderWithoutListingSkill(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# missing manifest\n"), 0o644); err != nil {
		t.Fatalf("write invalid fixture: %v", err)
	}

	resp := doRequest(t, a, http.MethodPost, "/api/agent/skills/import", strings.NewReader(`{
		"profile_id":"profile-a",
		"source_type":"folder",
		"path":`+strconv.Quote(root)+`
	}`), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid import status=%d body=%s", resp.Code, resp.Body.String())
	}
	var importPayload struct {
		State  string   `json:"state"`
		Errors []string `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&importPayload); err != nil {
		t.Fatalf("decode invalid import: %v", err)
	}
	if importPayload.State != "blocked-invalid-manifest" || !slices.ContainsFunc(importPayload.Errors, func(value string) bool {
		return strings.Contains(value, "manifest")
	}) {
		t.Fatalf("expected manifest validation errors, got %+v", importPayload)
	}

	listResp := doRequest(t, a, http.MethodGet, "/api/agent/skills?profile_id=profile-a", nil, nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listResp.Code, listResp.Body.String())
	}
	var listPayload struct {
		Skills []apiSkillPayload `json:"skills"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if imported := findAPISkill(listPayload.Skills, "cabinet.example.invalid_import"); imported != nil {
		t.Fatalf("invalid import must not be listed, got %+v", imported)
	}
}

func TestAgentSkillPreviewAPIBlocksUnsafeAdminMutation(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Unsafe Guard"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	resp := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
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

func TestAgentSkillDirectAPIGatesPreviewAndApplyWithProfileAuthorityPolicy(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Authority"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	repo := profile.NewRepository(a.db)
	if _, err := repo.PutAgentAuthorityPolicy(context.Background(), p.ID, profile.AgentAuthorityPolicy{
		Mode: profile.AgentAuthorityModeReadOnly,
	}); err != nil {
		t.Fatalf("set read-only authority policy: %v", err)
	}

	readOnlySearch := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inventory.search_items",
		"parameters":{"query":"Authority","workspace_id":"workspace-authority"}
	}`), map[string]string{"Content-Type": "application/json"})
	if readOnlySearch.Code != http.StatusOK {
		t.Fatalf("read-only search status=%d body=%s", readOnlySearch.Code, readOnlySearch.Body.String())
	}
	if !strings.Contains(readOnlySearch.Body.String(), `"read_only":true`) ||
		!strings.Contains(readOnlySearch.Body.String(), `"mutation_applied":false`) {
		t.Fatalf("expected read-only direct skill to remain executable, body=%s", readOnlySearch.Body.String())
	}

	blockedPreview := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inventory.create_item",
		"confirm":true,
		"source_thread_id":"thread-authority",
		"parameters":{"title":"Blocked authority item","part_number":"AUTH-1","workspace_id":"workspace-authority"}
	}`), map[string]string{"Content-Type": "application/json"})
	if blockedPreview.Code != http.StatusConflict {
		t.Fatalf("blocked preview status=%d body=%s", blockedPreview.Code, blockedPreview.Body.String())
	}
	if !strings.Contains(blockedPreview.Body.String(), `"error":"agent_authority_read_only"`) ||
		!strings.Contains(blockedPreview.Body.String(), `"entry_point":"direct-api"`) {
		t.Fatalf("expected read-only authority blocker on crafted preview, body=%s", blockedPreview.Body.String())
	}

	blockedApply := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inventory.create_item",
		"confirm":true,
		"source_thread_id":"thread-authority",
		"parameters":{"title":"Blocked authority item","part_number":"AUTH-1","workspace_id":"workspace-authority"}
	}`), map[string]string{"Content-Type": "application/json"})
	if blockedApply.Code != http.StatusConflict {
		t.Fatalf("blocked apply status=%d body=%s", blockedApply.Code, blockedApply.Body.String())
	}
	if !strings.Contains(blockedApply.Body.String(), `"error":"agent_authority_read_only"`) ||
		!strings.Contains(blockedApply.Body.String(), `"skill_id":"cabinet.inventory.create_item"`) {
		t.Fatalf("expected read-only authority blocker on crafted apply, body=%s", blockedApply.Body.String())
	}
	var itemCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM canonical_items WHERE profile_id = ? AND part_number = 'AUTH-1'`, p.ID).Scan(&itemCount); err != nil {
		t.Fatalf("count blocked authority item: %v", err)
	}
	if itemCount != 0 {
		t.Fatalf("read-only authority apply must not create inventory item, got %d", itemCount)
	}

	rows, err := a.db.Query(`
		SELECT source, after_json
		FROM audit_events
		WHERE entity_type = 'profile_agent_authority_decision'
			AND entity_id = ?
		ORDER BY created_at ASC, id ASC
	`, p.ID)
	if err != nil {
		t.Fatalf("query authority decision audit: %v", err)
	}
	defer rows.Close()
	var decisions []map[string]any
	for rows.Next() {
		var source string
		var afterRaw string
		if err := rows.Scan(&source, &afterRaw); err != nil {
			t.Fatalf("scan authority decision audit: %v", err)
		}
		if source != "direct-api" {
			t.Fatalf("expected direct-api audit source, got %q", source)
		}
		var decision map[string]any
		if err := json.Unmarshal([]byte(afterRaw), &decision); err != nil {
			t.Fatalf("unmarshal authority decision audit: %v", err)
		}
		decisions = append(decisions, decision)
		if strings.Contains(afterRaw, "Blocked authority item") || strings.Contains(afterRaw, "AUTH-1") {
			t.Fatalf("authority decision audit must not store raw payload values: %s", afterRaw)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate authority decision audit: %v", err)
	}
	if len(decisions) != 4 {
		t.Fatalf("expected four direct API authority decision audit rows, got %d: %+v", len(decisions), decisions)
	}
	var allowedSearch int
	var appliedSearch int
	var blockedCreate int
	var createPayloadRef map[string]any
	for _, decision := range decisions {
		if decision["skill_id"] == "cabinet.inventory.search_items" && decision["outcome"] == "apply_allowed" {
			allowedSearch++
		}
		if decision["skill_id"] == "cabinet.inventory.search_items" && decision["outcome"] == "applied" {
			appliedSearch++
		}
		if decision["skill_id"] == "cabinet.inventory.create_item" &&
			decision["outcome"] == "blocked" &&
			decision["blocker"] == "agent_authority_read_only" {
			blockedCreate++
			if ref, ok := decision["payload_ref"].(map[string]any); ok {
				createPayloadRef = ref
			}
		}
	}
	if allowedSearch != 1 {
		t.Fatalf("expected one allowed read-only search audit, got %d in %+v", allowedSearch, decisions)
	}
	if appliedSearch != 1 {
		t.Fatalf("expected one applied read-only search audit, got %d in %+v", appliedSearch, decisions)
	}
	if blockedCreate != 2 {
		t.Fatalf("expected two blocked create-item audits, got %d in %+v", blockedCreate, decisions)
	}
	if createPayloadRef == nil || createPayloadRef["parameter_count"] == nil {
		t.Fatalf("expected redacted payload reference on create-item authority audit, got %+v", decisions)
	}
}

func TestAgentSkillDirectAPIRecordsGovernedTimelineEvidence(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Timeline"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	const threadID = "thread-agent-skill-timeline-1981"
	const messageID = "message-agent-skill-timeline-1981"
	previewMutation := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inventory.create_item",
		"source_surface":"chats.main",
		"source_channel":"in-app",
		"source_thread_id":"`+threadID+`",
		"source_message_id":"`+messageID+`",
		"parameters":{"part_number":"TIMELINE-1981","title":"Timeline Preview Item","brand":"AFX","category":"Slot Cars"}
	}`), map[string]string{"Content-Type": "application/json"})
	if previewMutation.Code != http.StatusOK {
		t.Fatalf("preview mutation status=%d body=%s", previewMutation.Code, previewMutation.Body.String())
	}
	if !strings.Contains(previewMutation.Body.String(), `"confirmation_required":true`) ||
		!strings.Contains(previewMutation.Body.String(), `"mutation_applied":false`) ||
		!strings.Contains(previewMutation.Body.String(), `"source_thread_id":"`+threadID+`"`) {
		t.Fatalf("expected preview-required non-mutating source evidence, body=%s", previewMutation.Body.String())
	}
	var itemCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM canonical_items WHERE profile_id = ? AND part_number = 'TIMELINE-1981'`, p.ID).Scan(&itemCount); err != nil {
		t.Fatalf("count items after preview: %v", err)
	}
	if itemCount != 0 {
		t.Fatalf("preview must not mutate inventory before confirmation, got %d items", itemCount)
	}

	applyMutation := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inventory.create_item",
		"confirm":true,
		"source_surface":"chats.main",
		"source_channel":"in-app",
		"source_thread_id":"`+threadID+`",
		"source_message_id":"`+messageID+`",
		"parameters":{"part_number":"TIMELINE-1981","title":"Timeline Preview Item","brand":"AFX","category":"Slot Cars"}
	}`), map[string]string{"Content-Type": "application/json"})
	if applyMutation.Code != http.StatusOK {
		t.Fatalf("apply mutation status=%d body=%s", applyMutation.Code, applyMutation.Body.String())
	}
	if !strings.Contains(applyMutation.Body.String(), `"mutation_applied":true`) ||
		!strings.Contains(applyMutation.Body.String(), `"source_channel":"in-app"`) {
		t.Fatalf("expected confirmed apply evidence with source context, body=%s", applyMutation.Body.String())
	}

	readOnly := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inventory.search_items",
		"source_surface":"chats.main",
		"source_channel":"in-app",
		"source_thread_id":"`+threadID+`",
		"source_message_id":"message-agent-skill-readonly-1981",
		"parameters":{"query":"TIMELINE-1981"}
	}`), map[string]string{"Content-Type": "application/json"})
	if readOnly.Code != http.StatusOK {
		t.Fatalf("read-only apply status=%d body=%s", readOnly.Code, readOnly.Body.String())
	}
	if !strings.Contains(readOnly.Body.String(), `"read_only":true`) ||
		!strings.Contains(readOnly.Body.String(), `"mutation_applied":false`) ||
		strings.Contains(readOnly.Body.String(), `"confirmation_required":true`) {
		t.Fatalf("expected read-only non-mutating execution without confirmation token, body=%s", readOnly.Body.String())
	}

	navigationPreview := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.navigate.open_surface",
		"source_surface":"chats.main",
		"source_channel":"in-app",
		"source_thread_id":"`+threadID+`",
		"source_message_id":"message-agent-skill-shell-command-2005",
		"parameters":{"workspace":"default","known_surface":"inventory","route":"/inventory"}
	}`), map[string]string{"Content-Type": "application/json"})
	if navigationPreview.Code != http.StatusOK {
		t.Fatalf("navigation preview status=%d body=%s", navigationPreview.Code, navigationPreview.Body.String())
	}
	if !strings.Contains(navigationPreview.Body.String(), `"preview_only":true`) ||
		strings.Contains(navigationPreview.Body.String(), `"mutation_applied":true`) {
		t.Fatalf("expected preview-only shell command dispatch without mutation, body=%s", navigationPreview.Body.String())
	}

	providerReadinessPreview := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.integrations.test_connection",
		"source_surface":"chats.main",
		"source_channel":"in-app",
		"source_thread_id":"`+threadID+`",
		"source_message_id":"message-agent-skill-provider-readiness-2007",
		"parameters":{"provider":"openai","readiness_check":"setup_status"}
	}`), map[string]string{"Content-Type": "application/json"})
	if providerReadinessPreview.Code != http.StatusOK {
		t.Fatalf("provider-readiness preview status=%d body=%s", providerReadinessPreview.Code, providerReadinessPreview.Body.String())
	}
	if !strings.Contains(providerReadinessPreview.Body.String(), `"preview_only":true`) ||
		strings.Contains(providerReadinessPreview.Body.String(), `"mutation_applied":true`) ||
		strings.Contains(providerReadinessPreview.Body.String(), "test-secret") {
		t.Fatalf("expected preview-only provider-readiness dispatch without mutation or secret leakage, body=%s", providerReadinessPreview.Body.String())
	}

	runs := doRequest(t, a, http.MethodGet, "/api/chat/workflow-runs?profile_id="+p.ID+"&thread_id="+threadID, nil, nil)
	if runs.Code != http.StatusOK {
		t.Fatalf("workflow runs status=%d body=%s", runs.Code, runs.Body.String())
	}
	for _, want := range []string{
		`"workflow_id":"agent-skill-direct-preview"`,
		`"workflow_id":"agent-skill-direct-apply"`,
		`"capability_id":"cabinet.inventory.create_item"`,
		`"capability_id":"cabinet.inventory.search_items"`,
		`"source_channel":"in-app"`,
		`"source_thread_id":"` + threadID + `"`,
		`"source_message_id":"` + messageID + `"`,
		`"confirmation_state":"preview_required"`,
		`"confirmation_state":"confirmed"`,
		`"confirmation_state":"not_required"`,
		`"authority_outcome":"apply_allowed"`,
		`"mutation_applied":true`,
		`"mutation_applied":false`,
		`"ui_targets":["inventory.table","inventory.item.detail","inventory.item.editor"]`,
		`"capability_id":"cabinet.navigate.open_surface"`,
		`"source_message_id":"message-agent-skill-shell-command-2005"`,
		`"shell_commands":["navigate.open_surface"]`,
		`"shell_command_ids":["navigate.open_surface"]`,
		`"capability_id":"cabinet.integrations.test_connection"`,
		`"source_message_id":"message-agent-skill-provider-readiness-2007"`,
		`"provider_readiness_ids":["provider-registry"]`,
		`"provider_ids":["openai"]`,
		`"dispatch_boundary":"provider_readiness_registry"`,
		`"dispatch_outcome":"preview_only_no_mutation"`,
	} {
		if !strings.Contains(runs.Body.String(), want) {
			t.Fatalf("workflow timeline evidence missing %s: body=%s", want, runs.Body.String())
		}
	}
}

func TestAgentSkillApplyAPIHandlesDashboardActivitySummary(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createA := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Dashboard A"}`), map[string]string{"Content-Type": "application/json"})
	if createA.Code != http.StatusCreated {
		t.Fatalf("create profile a status=%d body=%s", createA.Code, createA.Body.String())
	}
	var profileA struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createA.Body).Decode(&profileA); err != nil {
		t.Fatalf("decode profile a: %v", err)
	}
	createB := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Dashboard B"}`), map[string]string{"Content-Type": "application/json"})
	if createB.Code != http.StatusCreated {
		t.Fatalf("create profile b status=%d body=%s", createB.Code, createB.Body.String())
	}
	var profileB struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createB.Body).Decode(&profileB); err != nil {
		t.Fatalf("decode profile b: %v", err)
	}

	if _, err := a.db.Exec(`
		INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, created_at)
		VALUES
			('dash-skill-item-a', ?, 'AFX', 'Slot', 'DSA-1', 'Dashboard Skill A Camaro', '2026-06-01T10:00:00Z'),
			('dash-skill-item-b', ?, 'AFX', 'Slot', 'DSB-1', 'Dashboard Skill B Porsche', '2026-06-02T10:00:00Z');
		INSERT INTO instances(id, item_id, condition, status, quantity, storage_location, acquisition_price, acquisition_date, notes)
		VALUES
			('dash-skill-inst-a','dash-skill-item-a','used','loose',2,'shelf',15,'',''),
			('dash-skill-inst-b','dash-skill-item-b','used','loose',9,'case',25,'','');
		INSERT INTO scanner_query_sets(id, profile_id, name, keywords_json, exclusions_json)
		VALUES ('dash-skill-query-a', ?, 'A', '["dsa"]', '[]'),('dash-skill-query-b', ?, 'B', '["dsb"]', '[]');
		INSERT INTO scanner_candidates(id, profile_id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source, stock_state, stock_count)
		VALUES
			('dash-skill-cand-a', ?, 'dash-skill-query-a', 'DSA-L1', 'Dashboard Skill A DSA-1', 20, 0, 'http://a.example', '', 'seller-a', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'low_stock', 2),
			('dash-skill-cand-b', ?, 'dash-skill-query-b', 'DSB-L1', 'Dashboard Skill B DSB-1', 20, 0, 'http://b.example', '', 'seller-b', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'low_stock', 2);
		INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at)
		VALUES
			('dash-skill-cand-a', '', 'not_in_collection', 0, 1, 'DSA-1', CURRENT_TIMESTAMP),
			('dash-skill-cand-b', '', 'not_in_collection', 0, 1, 'DSB-1', CURRENT_TIMESTAMP);
		INSERT INTO wishlist_entries(id, profile_id, item_id, target_price, priority, notes, highlight_hit)
		VALUES
			('dash-skill-wish-a', ?, 'dash-skill-item-a', 30, 'high', '', 1),
			('dash-skill-wish-b', ?, 'dash-skill-item-b', 30, 'high', '', 1);
		INSERT INTO price_snapshots(id, item_id, snapshot_date, source, min_price, median_price, latest_price, stock_count)
		VALUES
			('dash-skill-price-a1','dash-skill-item-a','2026-02-20','ebay',15,15,15,0),
			('dash-skill-price-a2','dash-skill-item-a','2026-02-21','ebay',12,12,12,4),
			('dash-skill-price-b1','dash-skill-item-b','2026-02-20','ebay',25,25,25,0),
			('dash-skill-price-b2','dash-skill-item-b','2026-02-21','ebay',22,22,22,6)
	`, profileA.ID, profileB.ID, profileA.ID, profileB.ID, profileA.ID, profileB.ID, profileA.ID, profileB.ID); err != nil {
		t.Fatalf("seed dashboard skill data: %v", err)
	}

	resp := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+profileA.ID+`",
		"skill_id":"cabinet.dashboard.summarise_activity",
		"source_surface":"dashboard.home",
		"source_channel":"web",
		"source_thread_id":"dashboard-thread-1942",
		"parameters":{"workspace_id":"workspace-dashboard","window":"last_7_days"}
	}`), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusOK {
		t.Fatalf("dashboard summary status=%d body=%s", resp.Code, resp.Body.String())
	}
	body := resp.Body.String()
	for _, want := range []string{
		`"skill_id":"cabinet.dashboard.summarise_activity"`,
		`"mutation_applied":false`,
		`"operation":"dashboard.activity.summary"`,
		`"read_only":true`,
		`"source_surface":"dashboard.home"`,
		`"requested_window":"last_7_days"`,
		`"evidence_backed":false`,
		`"snapshot_only":true`,
		`"total_items":1`,
		`"total_instances":2`,
		`"estimated_value":30`,
		`"new_discoveries"`,
		`"price_drops"`,
		`"restocks"`,
		`"Dashboard Skill A Camaro"`,
		`"item_id":"dash-skill-item-a"`,
		`"destination_links"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("dashboard summary response missing %s: body=%s", want, body)
		}
	}
	if strings.Contains(body, "Dashboard Skill B Porsche") || strings.Contains(body, "dash-skill-item-b") {
		t.Fatalf("dashboard summary leaked another profile, body=%s", body)
	}
	var auditCount int
	if err := a.db.QueryRow(`
		SELECT COUNT(1)
		FROM audit_events
		WHERE entity_type = 'profile_agent_authority_decision'
			AND entity_id = ?
			AND source = 'direct-api'
			AND json_extract(after_json, '$.outcome') = 'applied'
	`, profileA.ID).Scan(&auditCount); err != nil {
		t.Fatalf("count dashboard skill authority audit: %v", err)
	}
	if auditCount != 1 {
		t.Fatalf("expected one non-secret direct authority audit row, got %d", auditCount)
	}
}

func TestAgentSkillApplyAPIHandlesEmptyDashboardActivitySummary(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Empty Dashboard"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	resp := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.dashboard.summarise_activity",
		"parameters":{"workspace_id":"workspace-empty-dashboard"}
	}`), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusOK {
		t.Fatalf("empty dashboard summary status=%d body=%s", resp.Code, resp.Body.String())
	}
	body := resp.Body.String()
	for _, want := range []string{
		`"operation":"dashboard.activity.summary"`,
		`"read_only":true`,
		`"mutation_applied":false`,
		`"nothing_needs_attention":true`,
		`"total_items":0`,
		`"total_instances":0`,
		`"recent_items":[]`,
		`"destination_links"`,
		`"route":"/discoveries"`,
		`"route":"/collections"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("empty dashboard summary response missing %s: body=%s", want, body)
		}
	}
}

func TestAgentSkillApplyAPIHandlesPartialAndUnavailableDashboardActivitySummary(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Partial Dashboard"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if _, err := a.db.Exec(`
		INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, created_at)
		VALUES ('dash-partial-item', ?, 'AFX', 'Slot', 'PARTIAL-1', 'Partial Dashboard Item', '2026-06-03T10:00:00Z');
		INSERT INTO instances(id, item_id, condition, status, quantity, storage_location, acquisition_price, acquisition_date, notes)
		VALUES ('dash-partial-inst', 'dash-partial-item', 'used', 'loose', 1, 'shelf', 12, '', '');
	`, p.ID); err != nil {
		t.Fatalf("seed partial dashboard profile: %v", err)
	}
	if _, err := a.db.Exec(`
		ALTER TABLE canonical_items RENAME TO canonical_items_full;
		CREATE VIEW canonical_items AS
			SELECT id, profile_id, brand, category, part_number, title, created_at
			FROM canonical_items_full;
	`); err != nil {
		t.Fatalf("make recent-item dependency partial: %v", err)
	}

	partial := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.dashboard.summarise_activity",
		"parameters":{"workspace_id":"workspace-partial-dashboard","window":"today"}
	}`), map[string]string{"Content-Type": "application/json"})
	if partial.Code != http.StatusOK {
		t.Fatalf("partial dashboard summary status=%d body=%s", partial.Code, partial.Body.String())
	}
	partialBody := partial.Body.String()
	for _, want := range []string{
		`"operation":"dashboard.activity.summary"`,
		`"mutation_applied":false`,
		`"status":"partial"`,
		`"recent_items_unavailable":true`,
		`"total_items":1`,
		`"requested_window":"today"`,
		`"Recent Dashboard item identifiers are unavailable`,
	} {
		if !strings.Contains(partialBody, want) {
			t.Fatalf("partial dashboard summary missing %s: body=%s", want, partialBody)
		}
	}
	if strings.Contains(partialBody, "no such column") || strings.Contains(partialBody, "canonical_items") {
		t.Fatalf("partial dashboard summary leaked storage error details, body=%s", partialBody)
	}
	if _, err := a.db.Exec(`DROP VIEW canonical_items; ALTER TABLE canonical_items_full RENAME TO canonical_items; ALTER TABLE scanner_matches RENAME TO scanner_matches_unavailable;`); err != nil {
		t.Fatalf("make dashboard dependency unavailable: %v", err)
	}

	unavailable := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.dashboard.summarise_activity",
		"parameters":{"workspace_id":"workspace-unavailable-dashboard"}
	}`), map[string]string{"Content-Type": "application/json"})
	if unavailable.Code != http.StatusOK {
		t.Fatalf("unavailable dashboard summary status=%d body=%s", unavailable.Code, unavailable.Body.String())
	}
	unavailableBody := unavailable.Body.String()
	for _, want := range []string{
		`"operation":"dashboard.activity.summary"`,
		`"mutation_applied":false`,
		`"status":"unavailable"`,
		`"reason":"dashboard_summary_unavailable"`,
		`"collection":{"unavailable":true}`,
		`"recent_items":[]`,
		`"Dashboard activity data is unavailable right now`,
		`"route":"/dashboard"`,
	} {
		if !strings.Contains(unavailableBody, want) {
			t.Fatalf("unavailable dashboard summary missing %s: body=%s", want, unavailableBody)
		}
	}
	if strings.Contains(unavailableBody, "no such table") || strings.Contains(unavailableBody, "scanner_matches") {
		t.Fatalf("unavailable dashboard summary leaked storage error details, body=%s", unavailableBody)
	}
}

func TestAgentSkillPreviewAPIBlocksWishlistAndCollectionMissingContext(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Context Guard"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	wishlist := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.wishlist.mark_purchased",
		"parameters":{"purchase_url":"https://example.test/order"}
	}`), map[string]string{"Content-Type": "application/json"})
	if wishlist.Code != http.StatusOK {
		t.Fatalf("wishlist preview status=%d body=%s", wishlist.Code, wishlist.Body.String())
	}
	if !strings.Contains(wishlist.Body.String(), `"blocker":"wishlist_entry_required"`) ||
		!strings.Contains(wishlist.Body.String(), `"mutation_applied":false`) {
		t.Fatalf("expected missing wishlist entry blocker, body=%s", wishlist.Body.String())
	}

	allItems := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.collections.soft_delete",
		"parameters":{"collection_name":"All Items"}
	}`), map[string]string{"Content-Type": "application/json"})
	if allItems.Code != http.StatusOK {
		t.Fatalf("collections All Items preview status=%d body=%s", allItems.Code, allItems.Body.String())
	}
	if !strings.Contains(allItems.Body.String(), `"blocker":"collections_all_items_protected"`) ||
		!strings.Contains(allItems.Body.String(), `"mutation_applied":false`) {
		t.Fatalf("expected protected All Items blocker, body=%s", allItems.Body.String())
	}

	nonEmptyDelete := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.collections.soft_delete",
		"parameters":{"collection_name":"Display Case","has_items":true}
	}`), map[string]string{"Content-Type": "application/json"})
	if nonEmptyDelete.Code != http.StatusOK {
		t.Fatalf("collections non-empty delete preview status=%d body=%s", nonEmptyDelete.Code, nonEmptyDelete.Body.String())
	}
	if !strings.Contains(nonEmptyDelete.Body.String(), `"blocker":"collections_delete_destination_required"`) {
		t.Fatalf("expected missing destination blocker, body=%s", nonEmptyDelete.Body.String())
	}
}

func TestAgentSkillApplyAPIHandlesInventorySkills(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Inventory"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	createItem := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inventory.create_item",
		"confirm":true,
		"source_surface":"chats.main",
		"source_channel":"in-app",
		"source_thread_id":"thread-inventory-1707",
		"source_message_id":"message-inventory-1707",
		"parameters":{
			"part_number":"INV-AGENT-1707",
			"title":"Agent Inventory Runtime Item",
			"brand":"AFX",
			"category":"Slot Cars",
			"notes":"created by inventory agent skill",
			"source_url":"https://example.test/inventory/1707"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if createItem.Code != http.StatusOK {
		t.Fatalf("create inventory skill status=%d body=%s", createItem.Code, createItem.Body.String())
	}
	for _, want := range []string{
		`"skill_id":"cabinet.inventory.create_item"`,
		`"source_channel":"in-app"`,
		`"mutation_applied":true`,
		`"operation":"inventory.item.create"`,
		`"inventory_persisted":true`,
		`"part_number":"INV-AGENT-1707"`,
	} {
		if !strings.Contains(createItem.Body.String(), want) {
			t.Fatalf("create inventory response missing %s: body=%s", want, createItem.Body.String())
		}
	}
	var created struct {
		Target struct {
			ItemID string `json:"item_id"`
		} `json:"target"`
	}
	if err := json.NewDecoder(createItem.Body).Decode(&created); err != nil {
		t.Fatalf("decode created inventory response: %v", err)
	}
	if created.Target.ItemID == "" {
		t.Fatalf("expected created item id, body=%s", createItem.Body.String())
	}
	var itemCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM canonical_items WHERE profile_id = ? AND id = ? AND part_number = 'INV-AGENT-1707'`, p.ID, created.Target.ItemID).Scan(&itemCount); err != nil {
		t.Fatalf("count created inventory item: %v", err)
	}
	if itemCount != 1 {
		t.Fatalf("expected one persisted inventory item, got %d", itemCount)
	}

	searchItems := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inventory.search_items",
		"parameters":{"query":"Runtime Item"}
	}`), map[string]string{"Content-Type": "application/json"})
	if searchItems.Code != http.StatusOK {
		t.Fatalf("search inventory skill status=%d body=%s", searchItems.Code, searchItems.Body.String())
	}
	if !strings.Contains(searchItems.Body.String(), `"mutation_applied":false`) ||
		!strings.Contains(searchItems.Body.String(), `"operation":"inventory.item.search"`) ||
		!strings.Contains(searchItems.Body.String(), `"id":"`+created.Target.ItemID+`"`) {
		t.Fatalf("expected read-only inventory search evidence, body=%s", searchItems.Body.String())
	}

	updateItem := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inventory.update_item",
		"confirm":true,
		"parameters":{"item_id":"`+created.Target.ItemID+`","title":"Agent Inventory Runtime Item Updated","status":"active","priority":"high","notes":"updated by inventory agent skill"}
	}`), map[string]string{"Content-Type": "application/json"})
	if updateItem.Code != http.StatusOK {
		t.Fatalf("update inventory skill status=%d body=%s", updateItem.Code, updateItem.Body.String())
	}
	if !strings.Contains(updateItem.Body.String(), `"operation":"inventory.item.update"`) ||
		!strings.Contains(updateItem.Body.String(), `"inventory_persisted":true`) ||
		!strings.Contains(updateItem.Body.String(), `"title":"Agent Inventory Runtime Item Updated"`) {
		t.Fatalf("expected inventory update persistence evidence, body=%s", updateItem.Body.String())
	}

	assignCollection := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inventory.assign_to_collection",
		"confirm":true,
		"parameters":{"item_id":"`+created.Target.ItemID+`","collection_name":"Display Case"}
	}`), map[string]string{"Content-Type": "application/json"})
	if assignCollection.Code != http.StatusOK {
		t.Fatalf("assign collection status=%d body=%s", assignCollection.Code, assignCollection.Body.String())
	}
	if !strings.Contains(assignCollection.Body.String(), `"operation":"inventory.collection.assign"`) ||
		!strings.Contains(assignCollection.Body.String(), `"collection_persisted":true`) ||
		!strings.Contains(assignCollection.Body.String(), `"collection_name":"Display Case"`) {
		t.Fatalf("expected inventory collection assignment evidence, body=%s", assignCollection.Body.String())
	}

	importMedia := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.media.upload_or_import",
		"confirm":true,
		"parameters":{"source_url":"https://example.test/media/1707.jpg","filename":"inventory-1707.jpg","title":"Inventory 1707 media"}
	}`), map[string]string{"Content-Type": "application/json"})
	if importMedia.Code != http.StatusOK {
		t.Fatalf("import media status=%d body=%s", importMedia.Code, importMedia.Body.String())
	}
	var imported struct {
		Target struct {
			MediaID string `json:"media_id"`
		} `json:"target"`
	}
	if err := json.NewDecoder(importMedia.Body).Decode(&imported); err != nil {
		t.Fatalf("decode imported media response: %v", err)
	}
	if imported.Target.MediaID == "" {
		t.Fatalf("expected imported media id, body=%s", importMedia.Body.String())
	}

	attachMedia := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inventory.attach_media",
		"confirm":true,
		"parameters":{"item_id":"`+created.Target.ItemID+`","media_id":"`+imported.Target.MediaID+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if attachMedia.Code != http.StatusOK {
		t.Fatalf("attach media status=%d body=%s", attachMedia.Code, attachMedia.Body.String())
	}
	if !strings.Contains(attachMedia.Body.String(), `"operation":"inventory.media.attach"`) ||
		!strings.Contains(attachMedia.Body.String(), `"attachment_persisted":true`) ||
		!strings.Contains(attachMedia.Body.String(), `"provenance_preserved":true`) {
		t.Fatalf("expected inventory media attachment evidence, body=%s", attachMedia.Body.String())
	}
	var linkCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM media_asset_links WHERE profile_id = ? AND asset_id = ? AND target_type = 'inventory' AND target_id = ?`, p.ID, imported.Target.MediaID, created.Target.ItemID).Scan(&linkCount); err != nil {
		t.Fatalf("count inventory media link: %v", err)
	}
	if linkCount != 1 {
		t.Fatalf("expected one persisted inventory media link, got %d", linkCount)
	}
}

func TestAgentSkillApplyAPIHandlesWishlistSkills(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Wishlist"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	createWishlist := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.wishlist.create_entry",
		"confirm":true,
		"source_surface":"telegram.wishlist.intent",
		"source_channel":"telegram",
		"source_thread_id":"tg-wishlist-thread-1708",
		"source_message_id":"tg-wishlist-message-1708",
		"parameters":{
			"part_number":"WISH-AGENT-1708",
			"title":"Agent Wishlist Runtime Item",
			"brand":"AFX",
			"category":"Slot Cars",
			"target_price":42,
			"priority":"high",
			"needed_quantity":3,
			"notes":"agent wishlist runtime create",
			"source_url":"https://example.test/wishlist/1708"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if createWishlist.Code != http.StatusOK {
		t.Fatalf("create wishlist skill status=%d body=%s", createWishlist.Code, createWishlist.Body.String())
	}
	for _, want := range []string{
		`"skill_id":"cabinet.wishlist.create_entry"`,
		`"source_channel":"telegram"`,
		`"mutation_applied":true`,
		`"operation":"wishlist.entry.create"`,
		`"wishlist_persisted":true`,
		`"created_item":true`,
	} {
		if !strings.Contains(createWishlist.Body.String(), want) {
			t.Fatalf("create wishlist response missing %s: body=%s", want, createWishlist.Body.String())
		}
	}
	var created struct {
		Target struct {
			ItemID          string `json:"item_id"`
			WishlistEntryID string `json:"wishlist_entry_id"`
		} `json:"target"`
	}
	if err := json.NewDecoder(createWishlist.Body).Decode(&created); err != nil {
		t.Fatalf("decode created wishlist skill response: %v", err)
	}
	if created.Target.ItemID == "" || created.Target.WishlistEntryID == "" {
		t.Fatalf("expected created item and wishlist ids, got %+v body=%s", created.Target, createWishlist.Body.String())
	}
	var wishlistCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM wishlist_entries WHERE profile_id = ? AND id = ? AND item_id = ? AND priority = 'high'`, p.ID, created.Target.WishlistEntryID, created.Target.ItemID).Scan(&wishlistCount); err != nil {
		t.Fatalf("count created wishlist entry: %v", err)
	}
	if wishlistCount != 1 {
		t.Fatalf("expected one persisted wishlist entry, got %d", wishlistCount)
	}

	searchWishlist := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.wishlist.search_entries",
		"parameters":{"query":"runtime create"}
	}`), map[string]string{"Content-Type": "application/json"})
	if searchWishlist.Code != http.StatusOK {
		t.Fatalf("search wishlist skill status=%d body=%s", searchWishlist.Code, searchWishlist.Body.String())
	}
	if !strings.Contains(searchWishlist.Body.String(), `"mutation_applied":false`) ||
		!strings.Contains(searchWishlist.Body.String(), `"operation":"wishlist.entry.search"`) ||
		!strings.Contains(searchWishlist.Body.String(), `"id":"`+created.Target.WishlistEntryID+`"`) {
		t.Fatalf("expected read-only wishlist search evidence, body=%s", searchWishlist.Body.String())
	}

	updateWishlist := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.wishlist.update_entry",
		"confirm":true,
		"parameters":{"wishlist_entry_id":"`+created.Target.WishlistEntryID+`","priority":"medium","notes":"agent wishlist runtime updated","target_price":37}
	}`), map[string]string{"Content-Type": "application/json"})
	if updateWishlist.Code != http.StatusOK {
		t.Fatalf("update wishlist skill status=%d body=%s", updateWishlist.Code, updateWishlist.Body.String())
	}
	if !strings.Contains(updateWishlist.Body.String(), `"operation":"wishlist.entry.update"`) ||
		!strings.Contains(updateWishlist.Body.String(), `"wishlist_persisted":true`) ||
		!strings.Contains(updateWishlist.Body.String(), `"priority":"medium"`) {
		t.Fatalf("expected wishlist update persistence evidence, body=%s", updateWishlist.Body.String())
	}

	if _, err := a.db.Exec(`INSERT INTO instances(id, item_id, condition, status, quantity, notes) VALUES ('agent-wishlist-existing-instance', ?, 'loose', 'loose', 2, 'existing inventory')`, created.Target.ItemID); err != nil {
		t.Fatalf("seed existing wishlist inventory: %v", err)
	}
	markPurchasedBody := `{
		"profile_id":"` + p.ID + `",
		"skill_id":"cabinet.wishlist.mark_purchased",
		"confirm":true,
		"parameters":{
			"wishlist_entry_id":"` + created.Target.WishlistEntryID + `",
			"price_paid":31.5,
			"purchase_url":"https://example.test/order/1708",
			"purchase_date":"2026-07-05",
			"purchase_condition":"sealed",
			"quantity":3
		}
	}`
	markPurchased := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(markPurchasedBody), map[string]string{"Content-Type": "application/json"})
	if markPurchased.Code != http.StatusOK {
		t.Fatalf("mark purchased skill status=%d body=%s", markPurchased.Code, markPurchased.Body.String())
	}
	if !strings.Contains(markPurchased.Body.String(), `"operation":"wishlist.entry.mark_purchased"`) ||
		!strings.Contains(markPurchased.Body.String(), `"purchase_sync_provenance":true`) ||
		!strings.Contains(markPurchased.Body.String(), `"inventory_quantity_sync_safe":true`) {
		t.Fatalf("expected wishlist purchase sync evidence, body=%s", markPurchased.Body.String())
	}
	repeatPurchased := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(markPurchasedBody), map[string]string{"Content-Type": "application/json"})
	if repeatPurchased.Code != http.StatusOK {
		t.Fatalf("repeat mark purchased skill status=%d body=%s", repeatPurchased.Code, repeatPurchased.Body.String())
	}
	var quantity int
	var lifecycleCount int
	if err := a.db.QueryRow(`SELECT quantity FROM instances WHERE id = 'agent-wishlist-existing-instance'`).Scan(&quantity); err != nil {
		t.Fatalf("load synced wishlist inventory instance: %v", err)
	}
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM commerce_lifecycle_entries WHERE profile_id = ? AND source = 'wishlist' AND external_ref = ?`, p.ID, created.Target.WishlistEntryID).Scan(&lifecycleCount); err != nil {
		t.Fatalf("count wishlist purchase lifecycle: %v", err)
	}
	if quantity != 5 || lifecycleCount != 1 {
		t.Fatalf("expected one quantity sync and one lifecycle after repeated purchase apply, quantity=%d lifecycle=%d", quantity, lifecycleCount)
	}

	softDelete := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.wishlist.soft_delete_entry",
		"confirm":true,
		"parameters":{"wishlist_entry_id":"`+created.Target.WishlistEntryID+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if softDelete.Code != http.StatusOK {
		t.Fatalf("soft-delete wishlist skill status=%d body=%s", softDelete.Code, softDelete.Body.String())
	}
	if !strings.Contains(softDelete.Body.String(), `"operation":"wishlist.entry.soft_delete"`) ||
		!strings.Contains(softDelete.Body.String(), `"wishlist_deleted":true`) {
		t.Fatalf("expected wishlist soft-delete evidence, body=%s", softDelete.Body.String())
	}

	restore := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.wishlist.restore_entry",
		"confirm":true,
		"parameters":{"wishlist_entry_id":"`+created.Target.WishlistEntryID+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if restore.Code != http.StatusOK {
		t.Fatalf("restore wishlist skill status=%d body=%s", restore.Code, restore.Body.String())
	}
	if !strings.Contains(restore.Body.String(), `"operation":"wishlist.entry.restore"`) ||
		!strings.Contains(restore.Body.String(), `"wishlist_deleted":false`) {
		t.Fatalf("expected wishlist restore evidence, body=%s", restore.Body.String())
	}
}

func TestAgentSkillApplyAPIHandlesCollectionsSkills(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Collections"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, status) VALUES ('agent-collection-item-1', ?, 'AFX', 'Slot Cars', 'COLL-1708', 'Agent Collections Item', 'active')`, p.ID); err != nil {
		t.Fatalf("seed collection item: %v", err)
	}

	createCollection := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.collections.create",
		"confirm":true,
		"source_surface":"telegram.collections.intent",
		"source_channel":"telegram",
		"source_thread_id":"tg-collections-thread-1708",
		"source_message_id":"tg-collections-message-1708",
		"parameters":{"collection_name":"Display Case"}
	}`), map[string]string{"Content-Type": "application/json"})
	if createCollection.Code != http.StatusOK {
		t.Fatalf("create collection skill status=%d body=%s", createCollection.Code, createCollection.Body.String())
	}
	for _, want := range []string{
		`"skill_id":"cabinet.collections.create"`,
		`"source_channel":"telegram"`,
		`"mutation_applied":true`,
		`"operation":"collections.create"`,
		`"collection_persisted":true`,
		`"collection_name":"Display Case"`,
	} {
		if !strings.Contains(createCollection.Body.String(), want) {
			t.Fatalf("create collection response missing %s: body=%s", want, createCollection.Body.String())
		}
	}

	assignItem := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.collection.assign_item",
		"confirm":true,
		"parameters":{"collection_name":"Display Case","item_id":"agent-collection-item-1"}
	}`), map[string]string{"Content-Type": "application/json"})
	if assignItem.Code != http.StatusOK {
		t.Fatalf("assign collection item status=%d body=%s", assignItem.Code, assignItem.Body.String())
	}
	if !strings.Contains(assignItem.Body.String(), `"operation":"collections.item.assign"`) ||
		!strings.Contains(assignItem.Body.String(), `"item_id":"agent-collection-item-1"`) ||
		!strings.Contains(assignItem.Body.String(), `"collection_persisted":true`) {
		t.Fatalf("expected collection assignment evidence, body=%s", assignItem.Body.String())
	}

	updateCollection := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.collections.update_metadata",
		"confirm":true,
		"parameters":{"collection_name":"Display Case","new_collection_name":"Showcase Shelf"}
	}`), map[string]string{"Content-Type": "application/json"})
	if updateCollection.Code != http.StatusOK {
		t.Fatalf("update collection status=%d body=%s", updateCollection.Code, updateCollection.Body.String())
	}
	if !strings.Contains(updateCollection.Body.String(), `"operation":"collections.update_metadata"`) ||
		!strings.Contains(updateCollection.Body.String(), `"collection_name":"Showcase Shelf"`) {
		t.Fatalf("expected collection metadata update evidence, body=%s", updateCollection.Body.String())
	}

	searchCollections := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.collections.search",
		"parameters":{"query":"Showcase"}
	}`), map[string]string{"Content-Type": "application/json"})
	if searchCollections.Code != http.StatusOK {
		t.Fatalf("search collections status=%d body=%s", searchCollections.Code, searchCollections.Body.String())
	}
	if !strings.Contains(searchCollections.Body.String(), `"mutation_applied":false`) ||
		!strings.Contains(searchCollections.Body.String(), `"operation":"collections.search"`) ||
		!strings.Contains(searchCollections.Body.String(), `"Showcase Shelf"`) {
		t.Fatalf("expected read-only collection search evidence, body=%s", searchCollections.Body.String())
	}

	moveOnDelete := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.collections.move_items_on_delete",
		"confirm":true,
		"parameters":{"collection_name":"Showcase Shelf","destination_collection":"Archive Shelf"}
	}`), map[string]string{"Content-Type": "application/json"})
	if moveOnDelete.Code != http.StatusOK {
		t.Fatalf("move collection items status=%d body=%s", moveOnDelete.Code, moveOnDelete.Body.String())
	}
	if !strings.Contains(moveOnDelete.Body.String(), `"operation":"collections.items.move_on_delete"`) ||
		!strings.Contains(moveOnDelete.Body.String(), `"collection_deleted":true`) ||
		!strings.Contains(moveOnDelete.Body.String(), `"destination_collection":"Archive Shelf"`) ||
		!strings.Contains(moveOnDelete.Body.String(), `"moved_item_count":1`) {
		t.Fatalf("expected collection move-on-delete evidence, body=%s", moveOnDelete.Body.String())
	}

	var workspaceRaw string
	if err := a.db.QueryRow(`SELECT value FROM profile_settings WHERE profile_id = ? AND key = 'collections.workspace.v1'`, p.ID).Scan(&workspaceRaw); err != nil {
		t.Fatalf("load collection workspace setting: %v", err)
	}
	if !strings.Contains(workspaceRaw, "Archive Shelf") ||
		strings.Contains(workspaceRaw, "Showcase Shelf") ||
		!strings.Contains(workspaceRaw, `"collectionName":"Archive Shelf"`) {
		t.Fatalf("expected persisted collection move state, got %s", workspaceRaw)
	}

	softDelete := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.collections.soft_delete",
		"confirm":true,
		"parameters":{"collection_name":"Archive Shelf","remove_items":true}
	}`), map[string]string{"Content-Type": "application/json"})
	if softDelete.Code != http.StatusOK {
		t.Fatalf("soft-delete collection status=%d body=%s", softDelete.Code, softDelete.Body.String())
	}
	if !strings.Contains(softDelete.Body.String(), `"operation":"collections.soft_delete"`) ||
		!strings.Contains(softDelete.Body.String(), `"collection_deleted":true`) ||
		!strings.Contains(softDelete.Body.String(), `"removed_item_count":1`) {
		t.Fatalf("expected collection soft-delete evidence, body=%s", softDelete.Body.String())
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

func TestAgentSkillInboxReviewContextClarifiesMissingOrStaleNotification(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Inbox Context Clarification"}`), map[string]string{"Content-Type": "application/json"})
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
			"local_history_id":"agent-skill-inbox-context-1987",
			"title":"Agent skill Inbox context",
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
		"source_thread_id":"thread-inbox-context-1987",
		"source_message_id":"message-inbox-context-1987",
		"agent_context":{
			"profile_id":"`+p.ID+`",
			"workspace_id":"`+p.ID+`",
			"route_id":"/inbox",
			"surface_id":"inbox.notification.card",
			"source_channel":"in-app",
			"thread_id":"thread-inbox-context-1987",
			"source_thread_id":"source-thread-1987",
			"source_message_id":"source-message-1987",
			"selected_notification":{"id":"`+recordPayload.Items[0].ID+`","source":"assistant_handoff"},
			"setup_state":"ready"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if preview.Code != http.StatusOK {
		t.Fatalf("preview from Inbox agent_context status=%d body=%s", preview.Code, preview.Body.String())
	}
	if !strings.Contains(preview.Body.String(), `"source_surface":"inbox.notification.card"`) ||
		!strings.Contains(preview.Body.String(), `"source_thread_id":"thread-inbox-context-1987"`) ||
		!strings.Contains(preview.Body.String(), `"mutation_applied":false`) {
		t.Fatalf("expected preview to use canonical Inbox agent context without mutation, body=%s", preview.Body.String())
	}

	missing := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inbox.mark_handled",
		"confirm":true,
		"source_surface":"inbox.notification.card",
		"source_channel":"in-app",
		"source_thread_id":"thread-inbox-context-1987",
		"source_message_id":"message-inbox-context-1987",
		"agent_context":{
			"profile_id":"`+p.ID+`",
			"workspace_id":"`+p.ID+`",
			"route_id":"/inbox",
			"surface_id":"inbox.notification.card",
			"source_channel":"in-app",
			"thread_id":"thread-inbox-context-1987",
			"setup_state":"ready"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if missing.Code != http.StatusConflict ||
		!strings.Contains(missing.Body.String(), `"error":"missing_context"`) ||
		!strings.Contains(missing.Body.String(), `"selected_notification"`) {
		t.Fatalf("expected missing Inbox notification context clarification, status=%d body=%s", missing.Code, missing.Body.String())
	}

	stale := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.inbox.mark_handled",
		"source_surface":"inbox.notification.card",
		"source_channel":"in-app",
		"source_thread_id":"thread-inbox-context-1987",
		"source_message_id":"message-inbox-context-1987",
		"agent_context":{
			"profile_id":"`+p.ID+`",
			"workspace_id":"`+p.ID+`",
			"route_id":"/inbox",
			"surface_id":"inbox.notification.card",
			"source_channel":"in-app",
			"thread_id":"thread-inbox-context-1987",
			"selected_notification":{"id":"stale-notification-1987","source":"assistant_handoff"},
			"setup_state":"ready"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if stale.Code != http.StatusConflict ||
		!strings.Contains(stale.Body.String(), `"error":"missing_context"`) ||
		!strings.Contains(stale.Body.String(), `"stale_selected_notification"`) ||
		!strings.Contains(stale.Body.String(), `"stale-notification-1987"`) {
		t.Fatalf("expected stale Inbox notification context clarification, status=%d body=%s", stale.Code, stale.Body.String())
	}
}

func TestAgentSkillAPIPropagatesExternalChannelContextForMarketWatchAndPurchases(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill External Channel Context"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	marketWatchPreview := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.market_watch.handoff_result",
		"source_surface":"telegram.market_watch.result",
		"source_channel":"telegram",
		"source_thread_id":"tg-chat-1710",
		"source_message_id":"tg-message-market-watch-1710",
		"parameters":{"provider_id":"ebay","result_id":"result-telegram-1710","destination":"wishlist","source_url":"https://example.test/listing/telegram-1710"}
	}`), map[string]string{"Content-Type": "application/json"})
	if marketWatchPreview.Code != http.StatusOK {
		t.Fatalf("market watch external preview status=%d body=%s", marketWatchPreview.Code, marketWatchPreview.Body.String())
	}
	for _, want := range []string{
		`"skill_id":"cabinet.market_watch.handoff_result"`,
		`"source_surface":"telegram.market_watch.result"`,
		`"source_channel":"telegram"`,
		`"source_thread_id":"tg-chat-1710"`,
		`"source_message_id":"tg-message-market-watch-1710"`,
		`"mutation_applied":false`,
		`"confirmation_required":true`,
	} {
		if !strings.Contains(marketWatchPreview.Body.String(), want) {
			t.Fatalf("market watch external preview missing %s: body=%s", want, marketWatchPreview.Body.String())
		}
	}

	purchasesApply := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.purchases.create_order",
		"confirm":true,
		"source_surface":"telegram.purchase.capture",
		"source_channel":"telegram",
		"source_thread_id":"tg-chat-1710",
		"source_message_id":"tg-message-purchase-1710",
		"parameters":{
			"purchase_source":"telegram",
			"order_id":"telegram-order-1710",
			"title":"Telegram purchase order item",
			"part_number":"TG-1710",
			"source_url":"https://example.test/orders/telegram-1710",
			"quantity":1,
			"amount":88,
			"currency":"AUD"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if purchasesApply.Code != http.StatusOK {
		t.Fatalf("purchases external apply status=%d body=%s", purchasesApply.Code, purchasesApply.Body.String())
	}
	for _, want := range []string{
		`"skill_id":"cabinet.purchases.create_order"`,
		`"source_surface":"telegram.purchase.capture"`,
		`"source_channel":"telegram"`,
		`"source_thread_id":"tg-chat-1710"`,
		`"source_message_id":"tg-message-purchase-1710"`,
		`"mutation_applied":true`,
		`"operation":"purchases.order.create"`,
		`"order_id":"telegram-order-1710"`,
		`"purchase_persisted":true`,
		`"provenance_preserved":true`,
	} {
		if !strings.Contains(purchasesApply.Body.String(), want) {
			t.Fatalf("purchases external apply missing %s: body=%s", want, purchasesApply.Body.String())
		}
	}
	var purchaseCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM commerce_lifecycle_entries WHERE profile_id = ? AND external_ref = 'telegram-order-1710' AND source = 'telegram'`, p.ID).Scan(&purchaseCount); err != nil {
		t.Fatalf("count external purchase order evidence: %v", err)
	}
	if purchaseCount != 1 {
		t.Fatalf("expected one persisted external purchase order, got %d", purchaseCount)
	}
}

func TestAgentSkillApplyAPIHandlesMediaSkills(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Media"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if _, err := a.db.Exec(`
		INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, status)
		VALUES ('media-item-1', ?, 'AFX', 'Slot Cars', 'MEDIA-1', 'Agent media target item', 'active');
		INSERT INTO chat_threads(id, profile_id, title)
		VALUES ('media-thread-1', ?, 'Agent media uploads');
		INSERT INTO chat_attachments(id, profile_id, thread_id, filename, mime_type, size_bytes, stored_path)
		VALUES ('media-attach-1', ?, 'media-thread-1', 'loose-reference.jpg', 'image/jpeg', 123, 'https://example.test/media/loose-reference.jpg');
	`, p.ID, p.ID, p.ID); err != nil {
		t.Fatalf("seed media skill data: %v", err)
	}

	search := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.media.search",
		"parameters":{"query":"loose-reference"}
	}`), map[string]string{"Content-Type": "application/json"})
	if search.Code != http.StatusOK {
		t.Fatalf("media search status=%d body=%s", search.Code, search.Body.String())
	}
	for _, want := range []string{
		`"mutation_applied":false`,
		`"operation":"media.search"`,
		`"read_only":true`,
		`"filename":"loose-reference.jpg"`,
	} {
		if !strings.Contains(search.Body.String(), want) {
			t.Fatalf("media search response missing %s: body=%s", want, search.Body.String())
		}
	}

	upload := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.media.upload_or_import",
		"confirm":true,
		"source_surface":"telegram.media.upload",
		"source_channel":"telegram",
		"source_thread_id":"tg-media-thread-1709",
		"source_message_id":"tg-media-message-1709",
		"parameters":{
			"source_url":"https://example.test/media/imported-reference.jpg",
			"filename":"imported-reference.jpg",
			"title":"Imported agent reference",
			"notes":"imported by agent skill"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if upload.Code != http.StatusOK {
		t.Fatalf("media upload/import status=%d body=%s", upload.Code, upload.Body.String())
	}
	for _, want := range []string{
		`"mutation_applied":true`,
		`"operation":"media.upload_or_import"`,
		`"media_persisted":true`,
		`"provenance_preserved":true`,
		`"source_url":"https://example.test/media/imported-reference.jpg"`,
		`"source_channel":"telegram"`,
	} {
		if !strings.Contains(upload.Body.String(), want) {
			t.Fatalf("media upload response missing %s: body=%s", want, upload.Body.String())
		}
	}
	var uploadPayload struct {
		Target struct {
			MediaID string `json:"media_id"`
		} `json:"target"`
	}
	if err := json.NewDecoder(upload.Body).Decode(&uploadPayload); err != nil {
		t.Fatalf("decode upload response: %v", err)
	}
	if uploadPayload.Target.MediaID == "" {
		t.Fatalf("expected uploaded media id, body=%s", upload.Body.String())
	}
	var importedCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM chat_attachments WHERE profile_id = ? AND id = ? AND stored_path = 'https://example.test/media/imported-reference.jpg'`, p.ID, uploadPayload.Target.MediaID).Scan(&importedCount); err != nil {
		t.Fatalf("count imported media attachment: %v", err)
	}
	if importedCount != 1 {
		t.Fatalf("expected one persisted imported media attachment, got %d", importedCount)
	}

	attach := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.media.attach_to_item",
		"confirm":true,
		"parameters":{"media_id":"`+uploadPayload.Target.MediaID+`","item_id":"media-item-1"}
	}`), map[string]string{"Content-Type": "application/json"})
	if attach.Code != http.StatusOK {
		t.Fatalf("media attach status=%d body=%s", attach.Code, attach.Body.String())
	}
	if !strings.Contains(attach.Body.String(), `"operation":"media.attach_to_item"`) ||
		!strings.Contains(attach.Body.String(), `"attachment_persisted":true`) ||
		!strings.Contains(attach.Body.String(), `"provenance_preserved":true`) {
		t.Fatalf("expected media attach persistence evidence, body=%s", attach.Body.String())
	}
	var linkCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM media_asset_links WHERE profile_id = ? AND asset_id = ? AND target_type = 'inventory' AND target_id = 'media-item-1'`, p.ID, uploadPayload.Target.MediaID).Scan(&linkCount); err != nil {
		t.Fatalf("count media attachment link: %v", err)
	}
	if linkCount != 1 {
		t.Fatalf("expected one persisted media link, got %d", linkCount)
	}

	updateNotes := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.media.update_notes",
		"confirm":true,
		"parameters":{"media_id":"`+uploadPayload.Target.MediaID+`","notes":"agent skill updated notes"}
	}`), map[string]string{"Content-Type": "application/json"})
	if updateNotes.Code != http.StatusOK {
		t.Fatalf("media update notes status=%d body=%s", updateNotes.Code, updateNotes.Body.String())
	}
	if !strings.Contains(updateNotes.Body.String(), `"operation":"media.update_notes"`) ||
		!strings.Contains(updateNotes.Body.String(), `"metadata_persisted":true`) ||
		!strings.Contains(updateNotes.Body.String(), `"notes":"agent skill updated notes"`) {
		t.Fatalf("expected media notes persistence evidence, body=%s", updateNotes.Body.String())
	}

	detach := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.media.detach_from_item",
		"confirm":true,
		"parameters":{"media_id":"`+uploadPayload.Target.MediaID+`","item_id":"media-item-1"}
	}`), map[string]string{"Content-Type": "application/json"})
	if detach.Code != http.StatusOK {
		t.Fatalf("media detach status=%d body=%s", detach.Code, detach.Body.String())
	}
	if !strings.Contains(detach.Body.String(), `"operation":"media.detach_from_item"`) ||
		!strings.Contains(detach.Body.String(), `"detachment_persisted":true`) {
		t.Fatalf("expected media detach persistence evidence, body=%s", detach.Body.String())
	}
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM media_asset_links WHERE profile_id = ? AND asset_id = ? AND target_type = 'inventory' AND target_id = 'media-item-1'`, p.ID, uploadPayload.Target.MediaID).Scan(&linkCount); err != nil {
		t.Fatalf("count media link after detach: %v", err)
	}
	if linkCount != 0 {
		t.Fatalf("expected media link to be detached, got %d", linkCount)
	}
}

func TestAgentSkillApplyAPIHandlesDiscoveriesSkills(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Discoveries"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if _, err := a.db.Exec(`
		INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, status)
		VALUES ('discoveries-item-1', ?, 'AFX', 'Slot Cars', 'DISC-ITEM-1', 'Agent discovery wishlist target', 'active');
		INSERT INTO scanner_query_sets(id, profile_id, name, keywords_json, exclusions_json, provider_scope_json)
		VALUES ('discoveries-q1', ?, 'Agent discoveries saved search', '["afx"]', '[]', '["ebay"]');
		INSERT INTO scanner_candidates(id, profile_id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source, stock_state, stock_count)
		VALUES
			('disc-review', ?, 'discoveries-q1', 'LIST-REVIEW', 'Agent review discovery', 14, 0, 'https://example.test/review', '', 'seller-review', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'in_stock', 3),
			('disc-dismiss', ?, 'discoveries-q1', 'LIST-DISMISS', 'Agent dismiss discovery', 15, 0, 'https://example.test/dismiss', '', 'seller-dismiss', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'in_stock', 2),
			('disc-wishlist', ?, 'discoveries-q1', 'LIST-WISH', 'Agent wishlist discovery', 16, 0, 'https://example.test/wish', '', 'seller-wish', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'low_stock', 1),
			('disc-purchase', ?, 'discoveries-q1', 'LIST-PURCHASE', 'Agent purchase discovery', 17, 0, 'https://example.test/purchase', '', 'seller-purchase', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'in_stock', 5),
			('disc-inventory', ?, 'discoveries-q1', 'LIST-INVENTORY', 'Agent inventory discovery', 18, 0, 'https://example.test/inventory', '', 'seller-inventory', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'in_stock', 6);
		INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at)
		VALUES
			('disc-review', '', 'not_in_collection', 0.8, 1, 'DISC-REVIEW', CURRENT_TIMESTAMP),
			('disc-dismiss', '', 'not_in_collection', 0.8, 1, 'DISC-DISMISS', CURRENT_TIMESTAMP),
			('disc-wishlist', 'discoveries-item-1', 'not_in_collection', 0.9, 0, 'DISC-WISH', CURRENT_TIMESTAMP),
			('disc-purchase', '', 'not_in_collection', 0.9, 0, 'DISC-PURCHASE', CURRENT_TIMESTAMP),
			('disc-inventory', '', 'not_in_collection', 0.9, 0, 'DISC-INVENTORY', CURRENT_TIMESTAMP);
	`, p.ID, p.ID, p.ID, p.ID, p.ID, p.ID, p.ID, p.ID); err != nil {
		t.Fatalf("seed discovery skill data: %v", err)
	}

	search := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.discoveries.search",
		"parameters":{"provider_id":"ebay","query":"wishlist"}
	}`), map[string]string{"Content-Type": "application/json"})
	if search.Code != http.StatusOK {
		t.Fatalf("discoveries search status=%d body=%s", search.Code, search.Body.String())
	}
	for _, want := range []string{
		`"mutation_applied":false`,
		`"operation":"discoveries.search"`,
		`"read_only":true`,
		`"candidate_id":"disc-wishlist"`,
	} {
		if !strings.Contains(search.Body.String(), want) {
			t.Fatalf("discoveries search response missing %s: body=%s", want, search.Body.String())
		}
	}

	review := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.discoveries.review_result",
		"parameters":{"provider_id":"ebay","result_id":"disc-review"}
	}`), map[string]string{"Content-Type": "application/json"})
	if review.Code != http.StatusOK {
		t.Fatalf("discoveries review status=%d body=%s", review.Code, review.Body.String())
	}
	if !strings.Contains(review.Body.String(), `"operation":"discoveries.review_result"`) ||
		!strings.Contains(review.Body.String(), `"mutation_applied":false`) ||
		!strings.Contains(review.Body.String(), `"source_result_url":"https://example.test/review"`) {
		t.Fatalf("expected read-only discovery review evidence, body=%s", review.Body.String())
	}

	dismiss := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.discoveries.dismiss_result",
		"confirm":true,
		"parameters":{"provider_id":"ebay","result_id":"disc-dismiss","notes":"not relevant for collection"}
	}`), map[string]string{"Content-Type": "application/json"})
	if dismiss.Code != http.StatusOK {
		t.Fatalf("discoveries dismiss status=%d body=%s", dismiss.Code, dismiss.Body.String())
	}
	if !strings.Contains(dismiss.Body.String(), `"operation":"discoveries.dismiss_result"`) ||
		!strings.Contains(dismiss.Body.String(), `"action":"ignore"`) ||
		!strings.Contains(dismiss.Body.String(), `"discovery_persisted":true`) ||
		!strings.Contains(dismiss.Body.String(), `"provenance_preserved":true`) {
		t.Fatalf("expected discovery dismiss persistence evidence, body=%s", dismiss.Body.String())
	}
	var dismissedCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM ignored_candidates WHERE candidate_id = 'disc-dismiss'`).Scan(&dismissedCount); err != nil {
		t.Fatalf("count dismissed discovery: %v", err)
	}
	if dismissedCount != 1 {
		t.Fatalf("expected one ignored discovery marker, got %d", dismissedCount)
	}

	wishlist := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.discoveries.send_to_wishlist",
		"confirm":true,
		"source_surface":"telegram.discoveries.result",
		"source_channel":"telegram",
		"source_thread_id":"tg-discovery-thread-1709",
		"source_message_id":"tg-discovery-message-1709",
		"parameters":{"provider_id":"ebay","result_id":"disc-wishlist","notes":"promote from agent skill"}
	}`), map[string]string{"Content-Type": "application/json"})
	if wishlist.Code != http.StatusOK {
		t.Fatalf("discoveries wishlist status=%d body=%s", wishlist.Code, wishlist.Body.String())
	}
	for _, want := range []string{
		`"mutation_applied":true`,
		`"operation":"discoveries.send_to_wishlist"`,
		`"action":"add_to_wishlist"`,
		`"source_channel":"telegram"`,
		`"discovery_persisted":true`,
		`"provenance_preserved":true`,
	} {
		if !strings.Contains(wishlist.Body.String(), want) {
			t.Fatalf("discoveries wishlist response missing %s: body=%s", want, wishlist.Body.String())
		}
	}
	var wishlistCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM wishlist_entries WHERE profile_id = ? AND item_id = 'discoveries-item-1' AND highlight_hit = 1`, p.ID).Scan(&wishlistCount); err != nil {
		t.Fatalf("count discovery wishlist handoff: %v", err)
	}
	if wishlistCount != 1 {
		t.Fatalf("expected one persisted wishlist handoff, got %d", wishlistCount)
	}

	purchase := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.discoveries.create_purchase",
		"confirm":true,
		"parameters":{"provider_id":"ebay","result_id":"disc-purchase","quantity":2,"notes":"purchase candidate from agent skill"}
	}`), map[string]string{"Content-Type": "application/json"})
	if purchase.Code != http.StatusOK {
		t.Fatalf("discoveries purchase status=%d body=%s", purchase.Code, purchase.Body.String())
	}
	if !strings.Contains(purchase.Body.String(), `"operation":"discoveries.create_purchase"`) ||
		!strings.Contains(purchase.Body.String(), `"action":"mark_purchased"`) ||
		!strings.Contains(purchase.Body.String(), `"discovery_persisted":true`) {
		t.Fatalf("expected discovery purchase persistence evidence, body=%s", purchase.Body.String())
	}
	var purchaseCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM commerce_lifecycle_entries WHERE profile_id = ? AND source = 'market_watch' AND external_ref = 'LIST-PURCHASE'`, p.ID).Scan(&purchaseCount); err != nil {
		t.Fatalf("count discovery purchase handoff: %v", err)
	}
	if purchaseCount != 1 {
		t.Fatalf("expected one persisted purchase handoff, got %d", purchaseCount)
	}

	inventory := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.discoveries.create_or_update_inventory_candidate",
		"confirm":true,
		"parameters":{"provider_id":"ebay","result_id":"disc-inventory","notes":"inventory candidate from agent skill"}
	}`), map[string]string{"Content-Type": "application/json"})
	if inventory.Code != http.StatusOK {
		t.Fatalf("discoveries inventory status=%d body=%s", inventory.Code, inventory.Body.String())
	}
	if !strings.Contains(inventory.Body.String(), `"operation":"discoveries.create_or_update_inventory_candidate"`) ||
		!strings.Contains(inventory.Body.String(), `"action":"create_item"`) ||
		!strings.Contains(inventory.Body.String(), `"discovery_persisted":true`) {
		t.Fatalf("expected discovery inventory persistence evidence, body=%s", inventory.Body.String())
	}
	var inventoryCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM canonical_items WHERE profile_id = ? AND part_number = 'DISC-INVENTORY' AND title = 'Agent inventory discovery'`, p.ID).Scan(&inventoryCount); err != nil {
		t.Fatalf("count discovery inventory candidate: %v", err)
	}
	if inventoryCount != 1 {
		t.Fatalf("expected one persisted inventory candidate item, got %d", inventoryCount)
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
	if _, err := profile.NewRepository(a.db).PutAgentAuthorityPolicy(context.Background(), p.ID, profile.AgentAuthorityPolicy{
		Mode:                  profile.AgentAuthorityModeApprovedExternalActions,
		ExternalWriteApproved: true,
	}); err != nil {
		t.Fatalf("approve external agent authority: %v", err)
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

func TestAgentSkillApplyAPIHandlesDataImportRestorePersistenceEvidence(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Data Restore"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	fixtureDir := t.TempDir()
	importPath := filepath.Join(fixtureDir, "profile-import-private.json")
	importPayload := []byte(`{"profile_settings":{"display_currency":"AUD","profile_private_note":"Sydney secure vault - private"},"items":[{"part_number":"IMP-2023-001","title":"Imported private kit"}]}`)
	if err := os.WriteFile(importPath, importPayload, 0o600); err != nil {
		t.Fatalf("write import fixture: %v", err)
	}
	backupPath := filepath.Join(fixtureDir, "cabinet-restore-private.zip")
	if err := os.WriteFile(backupPath, []byte("fixture-backup-bytes"), 0o600); err != nil {
		t.Fatalf("write restore fixture: %v", err)
	}

	importFile := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.data.import_file",
		"confirm":true,
		"parameters":{"file_path":"`+strings.ReplaceAll(importPath, `\`, `\\`)+`","import_note":"private note must not leak"}
	}`), map[string]string{"Content-Type": "application/json"})
	if importFile.Code != http.StatusOK {
		t.Fatalf("import file apply status=%d body=%s", importFile.Code, importFile.Body.String())
	}
	importBody := importFile.Body.String()
	for _, required := range []string{
		`"mutation_applied":true`,
		`"operation":"data.import.file"`,
		`"import_persisted":true`,
		`"profile_scope":"` + p.ID + `"`,
		`"imported_item_count":1`,
		`"settings_persisted":["display_currency"]`,
		`"source_path_redacted":true`,
		`"raw_payload_redacted":true`,
	} {
		if !strings.Contains(importBody, required) {
			t.Fatalf("expected import persistence evidence %q, body=%s", required, importBody)
		}
	}
	for _, forbidden := range []string{importPath, filepath.Base(importPath), "Sydney secure vault", "Imported private kit", "private note must not leak"} {
		if strings.Contains(importBody, forbidden) {
			t.Fatalf("import response leaked %q, body=%s", forbidden, importBody)
		}
	}
	assertProfileSetting(t, a, p.ID, "display_currency", "AUD")

	restore := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.data.restore_backup",
		"confirm":true,
		"parameters":{"backup_path":"`+strings.ReplaceAll(backupPath, `\`, `\\`)+`","confirmation_phrase":"Restore profile `+p.ID+` from selected backup"}
	}`), map[string]string{"Content-Type": "application/json"})
	if restore.Code != http.StatusOK {
		t.Fatalf("restore backup apply status=%d body=%s", restore.Code, restore.Body.String())
	}
	restoreBody := restore.Body.String()
	for _, required := range []string{
		`"mutation_applied":true`,
		`"operation":"data.backup.restore"`,
		`"destructive_confirmation":true`,
		`"restore_drill_verified":true`,
		`"profile_isolated":true`,
		`"integrity_check":"ok"`,
		`"selected_backup_redacted":true`,
	} {
		if !strings.Contains(restoreBody, required) {
			t.Fatalf("expected restore drill evidence %q, body=%s", required, restoreBody)
		}
	}
	for _, forbidden := range []string{backupPath, filepath.Base(backupPath), "fixture-backup-bytes"} {
		if strings.Contains(restoreBody, forbidden) {
			t.Fatalf("restore response leaked %q, body=%s", forbidden, restoreBody)
		}
	}
}

func TestAgentSkillApplyAPIHandlesSettingsProfilePersistenceEvidence(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Settings Profile"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	preview := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.settings.update_profile",
		"source_surface":"settings.profile.form",
		"source_channel":"in-app",
		"source_thread_id":"settings-profile-thread",
		"source_message_id":"settings-profile-message",
		"parameters":{
			"settings_profile":{
				"display_currency":"AUD",
				"telegram.catalog_capture.sender_id":"987654321",
				"profile_private_note":"Sydney secure vault - private"
			}
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if preview.Code != http.StatusOK {
		t.Fatalf("preview status=%d body=%s", preview.Code, preview.Body.String())
	}
	if !strings.Contains(preview.Body.String(), `"confirmation_required":true`) ||
		!strings.Contains(preview.Body.String(), `"mutation_applied":false`) ||
		!strings.Contains(preview.Body.String(), `"source_surface":"settings.profile.form"`) ||
		!strings.Contains(preview.Body.String(), `"source_channel":"in-app"`) {
		t.Fatalf("expected profile settings preview boundary with source context, body=%s", preview.Body.String())
	}
	var previewPersistedCount int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM profile_settings WHERE profile_id = ? AND key IN ('display_currency', 'telegram.catalog_capture.sender_id', 'profile_private_note')`, p.ID).Scan(&previewPersistedCount); err != nil {
		t.Fatalf("count preview profile settings: %v", err)
	}
	if previewPersistedCount != 0 {
		t.Fatalf("preview must not persist profile settings, count=%d", previewPersistedCount)
	}

	apply := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.settings.update_profile",
		"confirm":true,
		"source_surface":"settings.profile.form",
		"source_channel":"in-app",
		"source_thread_id":"settings-profile-thread",
		"source_message_id":"settings-profile-message",
		"parameters":{
			"settings_profile":{
				"display_currency":"AUD",
				"telegram.catalog_capture.sender_id":"987654321",
				"profile_private_note":"Sydney secure vault - private"
			}
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if apply.Code != http.StatusOK {
		t.Fatalf("apply status=%d body=%s", apply.Code, apply.Body.String())
	}
	for _, want := range []string{
		`"mutation_applied":true`,
		`"operation":"settings.profile.update"`,
		`"source_surface":"settings.profile.form"`,
		`"source_channel":"in-app"`,
		`"source_thread_id":"settings-profile-thread"`,
		`"source_message_id":"settings-profile-message"`,
		`"settings_persisted":["`,
		`"display_currency"`,
		`"telegram.catalog_capture.sender_id"`,
		`"profile_private_note"`,
	} {
		if !strings.Contains(apply.Body.String(), want) {
			t.Fatalf("profile setting apply response missing %s: body=%s", want, apply.Body.String())
		}
	}
	if strings.Contains(apply.Body.String(), "Sydney secure vault - private") {
		t.Fatalf("profile setting apply response must not expose raw setting values: body=%s", apply.Body.String())
	}
	assertProfileSetting(t, a, p.ID, "display_currency", "AUD")
	assertProfileSetting(t, a, p.ID, "telegram.catalog_capture.sender_id", "987654321")
	assertProfileSetting(t, a, p.ID, "profile_private_note", "Sydney secure vault - private")
}

func TestAgentSkillApplyAPIHandlesSettingsAccountPersistenceEvidence(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Skill Settings Account"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	preview := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.settings.update_account",
		"source_surface":"settings.account.form",
		"source_channel":"in-app",
		"source_thread_id":"settings-account-thread",
		"source_message_id":"settings-account-message",
		"parameters":{
			"settings_account":{
				"account.display_name":"Cabinet Account",
				"account.default_language":"en-AU",
				"account_private_note":"Do not echo account secret"
			}
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if preview.Code != http.StatusOK {
		t.Fatalf("preview status=%d body=%s", preview.Code, preview.Body.String())
	}
	if !strings.Contains(preview.Body.String(), `"confirmation_required":true`) ||
		!strings.Contains(preview.Body.String(), `"mutation_applied":false`) ||
		!strings.Contains(preview.Body.String(), `"source_surface":"settings.account.form"`) ||
		!strings.Contains(preview.Body.String(), `"source_channel":"in-app"`) {
		t.Fatalf("expected account settings preview boundary with source context, body=%s", preview.Body.String())
	}
	var previewPersistedCount int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM profile_settings WHERE profile_id = ? AND key IN ('account.display_name', 'account.default_language', 'account_private_note')`, p.ID).Scan(&previewPersistedCount); err != nil {
		t.Fatalf("count preview account settings: %v", err)
	}
	if previewPersistedCount != 0 {
		t.Fatalf("preview must not persist account settings, count=%d", previewPersistedCount)
	}

	apply := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.settings.update_account",
		"confirm":true,
		"source_surface":"settings.account.form",
		"source_channel":"in-app",
		"source_thread_id":"settings-account-thread",
		"source_message_id":"settings-account-message",
		"parameters":{
			"settings_account":{
				"account.display_name":"Cabinet Account",
				"account.default_language":"en-AU",
				"account_private_note":"Do not echo account secret"
			}
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if apply.Code != http.StatusOK {
		t.Fatalf("apply status=%d body=%s", apply.Code, apply.Body.String())
	}
	for _, want := range []string{
		`"mutation_applied":true`,
		`"operation":"settings.account.update"`,
		`"source_surface":"settings.account.form"`,
		`"source_channel":"in-app"`,
		`"source_thread_id":"settings-account-thread"`,
		`"source_message_id":"settings-account-message"`,
		`"settings_persisted":["`,
		`"account.display_name"`,
		`"account.default_language"`,
		`"account_private_note"`,
	} {
		if !strings.Contains(apply.Body.String(), want) {
			t.Fatalf("account setting apply response missing %s: body=%s", want, apply.Body.String())
		}
	}
	if strings.Contains(apply.Body.String(), "Do not echo account secret") ||
		strings.Contains(apply.Body.String(), "Cabinet Account") ||
		strings.Contains(apply.Body.String(), "en-AU") {
		t.Fatalf("account setting apply response must not expose raw setting values: body=%s", apply.Body.String())
	}
	assertProfileSetting(t, a, p.ID, "account.display_name", "Cabinet Account")
	assertProfileSetting(t, a, p.ID, "account.default_language", "en-AU")
	assertProfileSetting(t, a, p.ID, "account_private_note", "Do not echo account secret")
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
	if _, err := profile.NewRepository(a.db).PutAgentAuthorityPolicy(context.Background(), p.ID, profile.AgentAuthorityPolicy{
		Mode:                  profile.AgentAuthorityModeApprovedExternalActions,
		ExternalWriteApproved: true,
	}); err != nil {
		t.Fatalf("approve external agent authority: %v", err)
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

	ebayStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer agent-skill-token" {
			t.Fatalf("expected bearer token header, got %q", got)
		}
		if got := r.Header.Get("X-EBAY-C-MARKETPLACE-ID"); got != "EBAY_AU" {
			t.Fatalf("expected EBAY_AU marketplace header, got %q", got)
		}
		if got := r.URL.Query().Get("q"); got != "boxed kit" {
			t.Fatalf("expected query q=boxed kit, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"agent-skill-ebay-1","title":"Agent skill eBay live provider result","price":{"value":"88.50","currency":"AUD"},"itemWebUrl":"https://www.ebay.com.au/itm/agent-skill-ebay-1","image":{"imageUrl":"https://example.test/agent-skill-ebay-1.jpg"},"seller":{"username":"agent-seller"},"estimatedAvailabilities":[{"estimatedAvailabilityStatus":"IN_STOCK","estimatedAvailableQuantity":4}]}]}`))
	}))
	defer ebayStub.Close()

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
	if _, err := profile.NewRepository(a.db).PutAgentAuthorityPolicy(context.Background(), p.ID, profile.AgentAuthorityPolicy{
		Mode:                  profile.AgentAuthorityModeApprovedExternalActions,
		ExternalWriteApproved: true,
	}); err != nil {
		t.Fatalf("approve external agent authority: %v", err)
	}
	if err := profile.NewRepository(a.db).PutSettings(context.Background(), p.ID, map[string]string{
		"ebay_base_url":                   ebayStub.URL,
		"ebay_bearer_token":               "agent-skill-token",
		"ebay_marketplace":                "EBAY_AU",
		"integration.ebay.items_per_page": "11",
	}); err != nil {
		t.Fatalf("save ebay provider settings: %v", err)
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
	if _, err := a.db.Exec(`UPDATE provider_health SET status = 'ok', message = '', retry_after_seconds = 0, updated_at = CURRENT_TIMESTAMP WHERE provider = 'ebay'`); err != nil {
		t.Fatalf("mark provider health ready: %v", err)
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
		`"live_provider_dispatched":true`,
		`"status":"confirmed_provider_run"`,
		`"candidate_count":1`,
		`"title":"Agent skill eBay live provider result"`,
		`"items_per_page_requested":11`,
		`"source_surface":"market_watch.saved_watch.row"`,
	} {
		if !strings.Contains(runWatch.Body.String(), want) {
			t.Fatalf("run watch response missing %s: body=%s", want, runWatch.Body.String())
		}
	}
	var providerRunCount, providerCandidateCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM scanner_runs WHERE profile_id = ? AND query_set_id = ? AND provider = 'ebay' AND status = 'succeeded'`, p.ID, watchID).Scan(&providerRunCount); err != nil {
		t.Fatalf("count agent skill provider runs: %v", err)
	}
	if providerRunCount != 1 {
		t.Fatalf("expected one persisted provider run, got %d", providerRunCount)
	}
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM scanner_candidates WHERE profile_id = ? AND query_set_id = ? AND source = 'ebay' AND listing_id = 'agent-skill-ebay-1'`, p.ID, watchID).Scan(&providerCandidateCount); err != nil {
		t.Fatalf("count agent skill provider candidates: %v", err)
	}
	if providerCandidateCount != 1 {
		t.Fatalf("expected one persisted provider candidate, got %d", providerCandidateCount)
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
		!strings.Contains(handoff.Body.String(), `"destination":"purchases"`) ||
		!strings.Contains(handoff.Body.String(), `"handoff_persisted":true`) ||
		!strings.Contains(handoff.Body.String(), `"lifecycle_entry_id":"`) ||
		!strings.Contains(handoff.Body.String(), `"expected_arrival_id":"`) {
		t.Fatalf("expected provenance-preserving handoff result, body=%s", handoff.Body.String())
	}
	var handoffPurchaseCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM commerce_lifecycle_entries WHERE profile_id = ? AND source = 'market_watch' AND external_ref = 'result-9'`, p.ID).Scan(&handoffPurchaseCount); err != nil {
		t.Fatalf("count market watch purchase handoff: %v", err)
	}
	if handoffPurchaseCount != 1 {
		t.Fatalf("expected one persisted purchase handoff, got %d", handoffPurchaseCount)
	}

	wishlistHandoff := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.market_watch.handoff_result",
		"confirm":true,
		"parameters":{"provider_id":"ebay","result_id":"result-10","destination":"wishlist","title":"Wishlist handoff result","source_url":"https://example.test/listing/10","target_price":55,"priority":"high"}
	}`), map[string]string{"Content-Type": "application/json"})
	if wishlistHandoff.Code != http.StatusOK {
		t.Fatalf("wishlist handoff status=%d body=%s", wishlistHandoff.Code, wishlistHandoff.Body.String())
	}
	if !strings.Contains(wishlistHandoff.Body.String(), `"destination_applied":"wishlist"`) ||
		!strings.Contains(wishlistHandoff.Body.String(), `"wishlist_entry_id":"`) ||
		!strings.Contains(wishlistHandoff.Body.String(), `"handoff_persisted":true`) {
		t.Fatalf("expected wishlist handoff persistence evidence, body=%s", wishlistHandoff.Body.String())
	}
	var wishlistHandoffCount int
	if err := a.db.QueryRow(`
		SELECT COUNT(1)
		FROM wishlist_entries w
		JOIN canonical_items i ON i.id = w.item_id
		WHERE w.profile_id = ? AND i.title = 'Wishlist handoff result' AND w.priority = 'high'
	`, p.ID).Scan(&wishlistHandoffCount); err != nil {
		t.Fatalf("count market watch wishlist handoff: %v", err)
	}
	if wishlistHandoffCount != 1 {
		t.Fatalf("expected one persisted wishlist handoff, got %d", wishlistHandoffCount)
	}

	inventoryHandoff := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.market_watch.handoff_result",
		"confirm":true,
		"parameters":{"provider_id":"ebay","result_id":"result-11","destination":"inventory","title":"Inventory handoff result","source_url":"https://example.test/listing/11","condition":"sealed","quantity":3}
	}`), map[string]string{"Content-Type": "application/json"})
	if inventoryHandoff.Code != http.StatusOK {
		t.Fatalf("inventory handoff status=%d body=%s", inventoryHandoff.Code, inventoryHandoff.Body.String())
	}
	if !strings.Contains(inventoryHandoff.Body.String(), `"destination_applied":"inventory"`) ||
		!strings.Contains(inventoryHandoff.Body.String(), `"instance_id":"`) ||
		!strings.Contains(inventoryHandoff.Body.String(), `"handoff_persisted":true`) {
		t.Fatalf("expected inventory handoff persistence evidence, body=%s", inventoryHandoff.Body.String())
	}
	var inventoryHandoffCount int
	if err := a.db.QueryRow(`
		SELECT COUNT(1)
		FROM instances inst
		JOIN canonical_items i ON i.id = inst.item_id
		WHERE i.profile_id = ? AND i.title = 'Inventory handoff result' AND inst.condition = 'sealed' AND inst.quantity = 3
	`, p.ID).Scan(&inventoryHandoffCount); err != nil {
		t.Fatalf("count market watch inventory handoff: %v", err)
	}
	if inventoryHandoffCount != 1 {
		t.Fatalf("expected one persisted inventory handoff, got %d", inventoryHandoffCount)
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

	createOrder := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.purchases.create_order",
		"confirm":true,
		"parameters":{
			"purchase_source":"agent_manual",
			"order_id":"agent-order-1",
			"title":"Agent created purchase order item",
			"part_number":"AGENT-CREATE-1",
			"source_url":"https://example.test/order/1",
			"quantity":1,
			"amount":77,
			"currency":"AUD"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if createOrder.Code != http.StatusOK {
		t.Fatalf("create order status=%d body=%s", createOrder.Code, createOrder.Body.String())
	}
	for _, want := range []string{
		`"operation":"purchases.order.create"`,
		`"order_id":"agent-order-1"`,
		`"created_item":true`,
		`"purchase_persisted":true`,
		`"provenance_preserved":true`,
		`"expected_arrival_id":"`,
	} {
		if !strings.Contains(createOrder.Body.String(), want) {
			t.Fatalf("create order response missing %s: body=%s", want, createOrder.Body.String())
		}
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

	receiveOrder := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"skill_id":"cabinet.purchases.receive_order",
		"confirm":true,
		"parameters":{"order_id":"agent-order-1","delivered_on":"2026-07-06","notes":"bulk receive by agent skill"}
	}`), map[string]string{"Content-Type": "application/json"})
	if receiveOrder.Code != http.StatusOK {
		t.Fatalf("receive order status=%d body=%s", receiveOrder.Code, receiveOrder.Body.String())
	}
	for _, want := range []string{
		`"operation":"purchases.order.receive"`,
		`"order_id":"agent-order-1"`,
		`"purchase_persisted":true`,
		`"received_count":1`,
		`"status":"delivered"`,
	} {
		if !strings.Contains(receiveOrder.Body.String(), want) {
			t.Fatalf("receive order response missing %s: body=%s", want, receiveOrder.Body.String())
		}
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

func findCapabilityExplanation(capabilities []apiCapabilityExplanationPayload, id string) *apiCapabilityExplanationPayload {
	for i := range capabilities {
		if capabilities[i].SkillID == id {
			return &capabilities[i]
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

func writeAgentSkillImportFixture(t *testing.T, manifest string) string {
	t.Helper()
	root := t.TempDir()
	writeAgentSkillImportFile(t, root, "cabinet.skill.json", manifest)
	writeAgentSkillImportFile(t, root, "README.md", "# Test skill\n")
	return root
}

func writeAgentSkillImportFile(t *testing.T, root, path, content string) {
	t.Helper()
	fullPath := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir fixture path: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture file %s: %v", path, err)
	}
}

func validAgentSkillImportManifest(overrides string) string {
	base := `{
		"schema": "https://collectors.tech/cabinet/schemas/agent-skill.v1.json",
		"id": "cabinet.example.api_imported_reader",
		"version": "1.0.0",
		"displayName": "API imported reader",
		"description": "Reads Cabinet data without changing state.",
		"category": "inventory",
		"source": {"type": "archive"},
		"safetyLevel": "read-only",
		"status": "available",
		"modes": ["in-app", "assistant"],
		"capabilities": ["navigate.open_surface"],
		"guidedWorkflows": [],
		"uiTargets": ["inventory.item.title"],
		"integrationRequirements": [],
		"permissions": {
			"cabinetReads": ["inventory.help"],
			"cabinetWrites": [],
			"externalReads": [],
			"externalWrites": [],
			"secretAccess": false,
			"destructive": false
		},
		"compatibility": {"cabinetMinVersion": "0.1.0", "schemaVersion": "v1"},
		"audit": {
			"actionTimeline": "records local import metadata",
			"requiresConfirmation": false
		}
	}`
	if strings.TrimSpace(overrides) == "" {
		return base
	}
	return strings.TrimSuffix(strings.TrimSpace(base), "}") + "," + strings.TrimPrefix(strings.TrimSpace(overrides), "{")
}

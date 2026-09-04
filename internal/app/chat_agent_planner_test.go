package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/agentskills"
	"github.com/collectors-tech/cabinet/internal/ai"
	"github.com/collectors-tech/cabinet/internal/chat"
	"github.com/collectors-tech/cabinet/internal/collection"
	"github.com/collectors-tech/cabinet/internal/commerce"
	"github.com/collectors-tech/cabinet/internal/discovery"
	"github.com/collectors-tech/cabinet/internal/media"
	"github.com/collectors-tech/cabinet/internal/scanner"
	"github.com/collectors-tech/cabinet/internal/wishlist"
)

type captureAssistantProvider struct {
	req          ai.AssistantTurnRequest
	responseText string
	err          error
	errorClass   string
	nextAction   string
}

func (p *captureAssistantProvider) Name() string {
	return "fake"
}

func (p *captureAssistantProvider) RunAssistantTurn(_ context.Context, req ai.AssistantTurnRequest) (ai.AssistantTurnResponse, error) {
	p.req = req
	if p.err != nil {
		return ai.AssistantTurnResponse{
			Provider:        "fake",
			Model:           "fake-planner-model",
			ErrorClass:      p.errorClass,
			SetupNextAction: p.nextAction,
			Metadata: map[string]string{
				"network":       "disabled",
				"test_provider": "true",
				"error_class":   "provider_runtime_error",
			},
		}, p.err
	}
	text := strings.TrimSpace(p.responseText)
	if text == "" {
		text = `{"decision":"select_skill","skill_id":"cabinet.inventory.search_items","parameters":{"part_number":"AFX-123"},"message":"Searching inventory for AFX-123."}`
	}
	return ai.AssistantTurnResponse{
		Provider: "fake",
		Model:    "fake-planner-model",
		Text:     text,
		Metadata: map[string]string{
			"network":       "disabled",
			"test_provider": "true",
		},
	}, nil
}

func TestChatAgentPlannerRecognisesDiscoveryAndAcquisitionLanguage(t *testing.T) {
	t.Parallel()

	for _, request := range []string{
		"Find eBay discoveries for AFX slot cars",
		"Review this discovery result",
		"Dismiss this result",
		"Create a saved watch for this provider",
		"Run my eBay market watch",
		"Test the eBay connection",
		"Send this listing to my wishlist",
		"Create a purchase from this discovery",
		"Receive purchase order PO-2083",
		"Explain the eBay provider setup",
	} {
		if !chatMessageNeedsNaturalLanguageAgentPlanning(request) {
			t.Fatalf("expected acquisition request to require provider planning: %q", request)
		}
	}
}

func TestChatAgentPlannerReadSummaryCoverageTracksExecutableReadRegistry(t *testing.T) {
	t.Parallel()

	summaryFixtures := map[string]struct {
		execution map[string]any
		kind      string
	}{
		"cabinet.chat.action_timeline.view": {
			execution: map[string]any{"timeline_entries": []map[string]any{{"workflow_run_id": "run-1", "capability_id": "cabinet.inventory.search_items", "status": "completed", "operation": "inventory.item.search"}}},
			kind:      "chat_action_timeline",
		},
		"cabinet.inventory.search_items": {
			execution: map[string]any{"items": []collection.Item{{ID: "item-1", PartNumber: "AFX-1", Title: "Visible item", Status: "active", Category: "Slot", Brand: "AFX"}}},
			kind:      "inventory_items",
		},
		"cabinet.dashboard.summarise_activity": {
			execution: map[string]any{"attention_signals": []map[string]any{{"id": "signal-1", "label": "Needs review", "count": 1}}},
			kind:      "dashboard_activity",
		},
		"cabinet.storage.show_status": {
			execution: map[string]any{"storage_status": "ready", "backup_status": "configured"},
			kind:      "storage_status",
		},
		"cabinet.wishlist.search_entries": {
			execution: map[string]any{"entries": []wishlist.Entry{{ID: "wish-1", ItemID: "item-1", Priority: "high", HighlightHit: true}}},
			kind:      "wishlist_entries",
		},
		"cabinet.collections.search": {
			execution: map[string]any{"collections": []string{"Ready collection"}},
			kind:      "collections",
		},
		"cabinet.integrations.search_providers": {
			execution: map[string]any{"providers": []map[string]any{{"id": "openai", "status": "ready", "setup_required": false}}},
			kind:      "integration_providers",
		},
		"cabinet.integrations.explain_required_setup": {
			execution: map[string]any{"provider_id": "openai", "setup_required": true, "next_action": "Add credentials"},
			kind:      "integration_setup_explanation",
		},
		"cabinet.integrations.test_connection": {
			execution: map[string]any{"provider_id": "openai", "connection_status": "ready", "next_action": "none"},
			kind:      "integration_connection_status",
		},
		"cabinet.data.export_bundle": {
			execution: map[string]any{"status": "ready", "export_scope": "all"},
			kind:      "data_export_bundle",
		},
		"cabinet.maintenance.run_safe_check": {
			execution: map[string]any{"status": "healthy", "maintenance_check": "storage", "check_level": "safe"},
			kind:      "maintenance_safe_check",
		},
		"cabinet.inbox.search_notifications": {
			execution: map[string]any{"items": []map[string]any{{"id": "inbox-1", "title": "Review", "status": "unread", "source": "notification_history"}}},
			kind:      "inbox_notifications",
		},
		"cabinet.inbox.summarise_unhandled": {
			execution: map[string]any{"items": []map[string]any{{"id": "inbox-1", "title": "Review", "status": "unread", "source": "notification_history"}}},
			kind:      "inbox_unhandled",
		},
		"cabinet.users.search": {
			execution: map[string]any{"users": []map[string]any{{"id": "user-1", "display_name": "Visible User", "status": "active", "role": "admin"}}},
			kind:      "workspace_users",
		},
		"cabinet.media.search": {
			execution: map[string]any{"assets": []media.WorkspaceAsset{{ID: "asset-1", Title: "Visible asset", LinkageState: "unlinked", Source: "upload"}}},
			kind:      "media_assets",
		},
		"cabinet.media.review_unlinked": {
			execution: map[string]any{"assets": []media.WorkspaceAsset{{ID: "asset-1", Title: "Visible asset", LinkageState: "unlinked", Source: "upload"}}},
			kind:      "unlinked_media_assets",
		},
		"cabinet.discoveries.search": {
			execution: map[string]any{"items": []discovery.Item{{CandidateID: "discovery-1", Title: "Visible discovery", Status: "new", SourceProvider: "ebay"}}},
			kind:      "discovery_results",
		},
		"cabinet.discoveries.review_result": {
			execution: map[string]any{"item": discovery.Item{CandidateID: "discovery-1", Title: "Visible discovery", Status: "new", SourceProvider: "ebay"}},
			kind:      "discovery_results",
		},
		"cabinet.purchases.search_orders": {
			execution: map[string]any{"orders": []commerce.PurchaseOrder{{OrderID: "po-1", Status: "open", Source: "manual", LineItemCount: 1}}},
			kind:      "purchase_orders",
		},
		"cabinet.purchases.review_purchase": {
			execution: map[string]any{"order_id": "po-1", "review_status": "needs_attention"},
			kind:      "purchase_review",
		},
		"cabinet.market_watch.search_watches": {
			execution: map[string]any{"watches": []scanner.QuerySet{{ID: "watch-1", Name: "AFX watch", Enabled: true, ProviderScope: []string{"ebay"}, LastCandidateCount: 1}}},
			kind:      "market_watch_watches",
		},
		"cabinet.market_watch.review_results": {
			execution: map[string]any{"candidates": []scanner.Candidate{{ID: "candidate-1", Title: "Visible candidate", Status: "new", Source: "ebay", StockState: "available"}}},
			kind:      "market_watch_results",
		},
	}

	registry := agentskills.NewRegistry(nil)
	for _, skill := range registry.List() {
		if skill.SafetyLevel != agentskills.SafetyReadOnly && !plannerSafePreviewExecutionSkill(skill) {
			continue
		}
		if !skill.Executable {
			continue
		}
		fixture, ok := summaryFixtures[skill.ID]
		if !ok {
			t.Fatalf("executable read skill %s lacks a Chat result_summary fixture and contract mapping", skill.ID)
		}
		summary := plannerAgentReadResultSummary(skill.ID, fixture.execution)
		if summary == nil {
			t.Fatalf("executable read skill %s did not produce a Chat result_summary", skill.ID)
		}
		if summary.Kind != fixture.kind {
			t.Fatalf("executable read skill %s summary kind=%q, want %q", skill.ID, summary.Kind, fixture.kind)
		}
	}
	for skillID := range summaryFixtures {
		skill, ok := registry.Resolve(skillID)
		if !ok {
			t.Fatalf("summary fixture references unknown Agent skill %s", skillID)
		}
		if skill.SafetyLevel != agentskills.SafetyReadOnly && !plannerSafePreviewExecutionSkill(skill) {
			t.Fatalf("summary fixture %s is not an executable read-summary skill", skillID)
		}
	}
}

func TestChatAgentPlannerPreservesProviderSetupFailureTaxonomy(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Provider Truth"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != 201 {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profile.ID+`","title":"Provider Truth"}`), map[string]string{"Content-Type": "application/json"})
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{
			err:        errors.New("secret provider detail"),
			errorClass: "missing_credentials",
			nextAction: "configure_openai_api_key",
		}),
		agentskills.NewRegistry(nil),
		profile.ID,
		thread.ID,
		"Find eBay discoveries for AFX slot cars",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id": profile.ID, "thread_id": thread.ID, "route_id": "/chats",
				"surface_id": "chats.main", "source_channel": "in-app",
			},
		},
		"message-provider-truth",
	)
	if !handled {
		t.Fatal("expected acquisition request to enter provider planner")
	}
	errPayload, _ := result["error"].(map[string]any)
	if errPayload["code"] != "assistant_provider_missing_credentials" || result["setup_next_action"] != "configure_openai_api_key" {
		t.Fatalf("expected provider setup taxonomy, got %+v", result)
	}
	serialized, _ := json.Marshal(result)
	if strings.Contains(string(serialized), "secret provider detail") || result["preview_result"] != nil || result["execution_result"] != nil {
		t.Fatalf("provider failure must stay redacted and mutation-free: %s", serialized)
	}
	if !strings.Contains(string(serialized), `"state":"setup_required"`) ||
		!strings.Contains(string(serialized), `"kind":"open_setup"`) ||
		!strings.Contains(string(serialized), `"route":"/integrations?provider=openai"`) {
		t.Fatalf("provider configuration failure must expose normalized setup guidance: %s", serialized)
	}
}

func TestChatAgentPlannerPreviewsDiscoveryHandoffWithoutMutation(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Discovery Preview"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != 201 {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profile.ID+`","title":"Discovery Preview"}`), map[string]string{"Content-Type": "application/json"})
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.discoveries.send_to_wishlist","parameters":{"provider_id":"ebay","result_id":"disc-2083","source_url":"https://example.test/listing/2083","provider_secret":"must-not-leak","provider_context":{"access_token":"nested-secret-must-not-leak","listing_id":"listing-2083"}},"message":"I prepared the discovery handoff for review."}`}),
		agentskills.NewRegistry(nil),
		profile.ID,
		thread.ID,
		"Send this discovery result to my wishlist",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id": profile.ID, "thread_id": thread.ID, "route_id": "/discoveries",
				"surface_id": "discoveries.result.card", "source_channel": "in-app",
			},
		},
		"message-discovery-preview",
	)
	if !handled {
		t.Fatal("expected discovery handoff to enter provider planner")
	}
	preview, ok := result["preview_result"].(map[string]any)
	if !ok || preview["skill_id"] != "cabinet.discoveries.send_to_wishlist" ||
		preview["preview_kind"] != "agent_skill" || preview["confirmation_required"] != true ||
		preview["mutation_applied"] != false || preview["apply_endpoint"] != "/api/agent/skills/apply" {
		t.Fatalf("expected governed Agent Skill preview contract, got %+v", result)
	}
	target, _ := preview["target"].(map[string]any)
	if target["provider_id"] != "ebay" || target["result_id"] != "disc-2083" || target["source_url"] != "https://example.test/listing/2083" {
		t.Fatalf("expected provider/listing provenance in preview target, got %+v", target)
	}
	if result["confirmation_state"] != "preview_required" {
		t.Fatalf("expected explicit confirmation state, got %+v", result)
	}
	parameters, _ := preview["parameters"].(map[string]any)
	providerContext, _ := parameters["provider_context"].(map[string]any)
	if _, exists := providerContext["access_token"]; exists || providerContext["listing_id"] != "listing-2083" {
		t.Fatalf("expected nested provider context to preserve provenance but omit secrets, got %+v", parameters)
	}
	serialized, _ := json.Marshal(result)
	if strings.Contains(string(serialized), "must-not-leak") {
		t.Fatalf("preview leaked provider secret: %s", serialized)
	}
}

func TestChatAgentPlannerUsesDeterministicFakeProviderSkillSelection(t *testing.T) {
	t.Parallel()

	provider := &captureAssistantProvider{}
	registry := agentskills.NewRegistry(nil)
	disabledCreate, ok := registry.Resolve("cabinet.inventory.create_item")
	if !ok {
		t.Fatal("missing built-in inventory create skill")
	}
	disabledCreate.Enabled = false
	disabledCreate.Status = agentskills.StatusDisabled

	selection, err := planChatAgentSkill(context.Background(), provider, chatAgentPlannerInput{
		ProfileID: "profile-planner",
		ThreadID:  "thread-planner",
		Intent:    "Find the item with part number AFX-123",
		Model:     "fake-planner-model",
		AgentContext: map[string]any{
			"profile_id":     "profile-planner",
			"thread_id":      "thread-planner",
			"route_id":       "/chats",
			"surface_id":     "chats.main",
			"source_channel": "in-app",
			"intent_text":    "Find the item with part number AFX-123",
		},
		Skills: []agentskills.Skill{
			mustResolveSkill(t, registry, "cabinet.inventory.search_items"),
			disabledCreate,
		},
	})
	if err != nil {
		t.Fatalf("planChatAgentSkill() error = %v", err)
	}
	if selection.Decision != "select_skill" || selection.SkillID != "cabinet.inventory.search_items" {
		t.Fatalf("expected read skill selection, got %+v", selection)
	}
	if selection.Parameters["part_number"] != "AFX-123" {
		t.Fatalf("expected structured parameters from provider, got %+v", selection.Parameters)
	}
	if selection.ProviderTrace["provider"] != "fake" ||
		selection.ProviderTrace["network"] != "disabled" ||
		selection.ProviderTrace["governed_dispatch_owner"] != "cabinet" {
		t.Fatalf("expected redacted fake-provider trace, got %+v", selection.ProviderTrace)
	}

	var exposed struct {
		Skills []struct {
			ID                 string         `json:"id"`
			Status             string         `json:"status"`
			InputSchema        map[string]any `json:"input_schema"`
			SafetyDeclaration  map[string]any `json:"safety_declaration"`
			RequiredActions    []string       `json:"required_actions"`
			InputSchemaRefs    []string       `json:"input_schema_refs"`
			RequiredContext    []string       `json:"required_context"`
			ConfirmationNeeded bool           `json:"confirmation_required"`
		} `json:"skills"`
	}
	rawSkills, err := json.Marshal(provider.req.Context)
	if err != nil {
		t.Fatalf("marshal provider context: %v", err)
	}
	if err := json.Unmarshal(rawSkills, &exposed); err != nil {
		t.Fatalf("decode provider context skills: %v", err)
	}
	ids := make([]string, 0, len(exposed.Skills))
	for _, skill := range exposed.Skills {
		ids = append(ids, skill.ID)
		if skill.ID == "cabinet.inventory.create_item" {
			t.Fatalf("disabled skill was exposed to provider: %+v", exposed.Skills)
		}
		if skill.ID == "cabinet.inventory.search_items" {
			if skill.InputSchema["type"] != "object" || skill.InputSchema["additionalProperties"] != false {
				t.Fatalf("planner skill missing closed JSON input schema: %+v", skill)
			}
			properties, ok := skill.InputSchema["properties"].(map[string]any)
			if !ok || properties["query"] == nil {
				t.Fatalf("planner read skill schema missing query property: %+v", skill.InputSchema)
			}
			if skill.SafetyDeclaration["side_effect_level"] != "read-only" ||
				skill.SafetyDeclaration["local_read"] != true ||
				skill.SafetyDeclaration["local_write"] != false ||
				skill.ConfirmationNeeded {
				t.Fatalf("planner read skill missing explicit read-only safety declaration: %+v", skill)
			}
			if !slices.Contains(skill.RequiredActions, "inventory.item.search") || len(skill.RequiredContext) == 0 {
				t.Fatalf("planner skill missing required actions/context: %+v", skill)
			}
		}
	}
	if !slices.Contains(ids, "cabinet.inventory.search_items") {
		t.Fatalf("available read skill was not exposed to provider: %+v", exposed.Skills)
	}
	if provider.req.Context["agent_context"] == nil {
		t.Fatalf("provider request missing canonical agent context: %+v", provider.req.Context)
	}
}

func TestChatAgentPlannerNormalizesFriendlyIntegrationConfiguration(t *testing.T) {
	t.Parallel()

	registry := agentskills.NewRegistry(nil)
	provider := &captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.integrations.configure_provider","parameters":{"provider_name":"Voglers","catalogue":"public catalogue"},"message":"Prepared for review."}`}
	selection, err := planChatAgentSkill(context.Background(), provider, chatAgentPlannerInput{
		ProfileID: "profile-integration-config",
		ThreadID:  "thread-integration-config",
		Intent:    "Configure Voglers for its public catalogue.",
		Model:     "fake-planner-model",
		AgentContext: map[string]any{
			"profile_id": "profile-integration-config", "thread_id": "thread-integration-config",
			"route_id": "/integrations", "surface_id": "integrations.main", "source_channel": "in-app",
		},
		Skills: []agentskills.Skill{mustResolveSkill(t, registry, "cabinet.integrations.configure_provider")},
	})
	if err != nil {
		t.Fatalf("plan integration configuration: %v", err)
	}
	if selection.SkillID != "cabinet.integrations.configure_provider" {
		t.Fatalf("unexpected skill selection: %+v", selection)
	}
	for key, want := range map[string]any{
		"provider_id": "voglers", "setup_payload": "public_catalogue", "setup_step": "public_catalogue", "marketplace": "public",
	} {
		if got := selection.Parameters[key]; got != want {
			t.Fatalf("normalized parameter %s=%v, want %v; all=%+v", key, got, want, selection.Parameters)
		}
	}
	if _, exists := selection.Parameters["provider_name"]; exists {
		t.Fatalf("friendly provider name must be normalized away: %+v", selection.Parameters)
	}
	if _, exists := selection.Parameters["catalogue"]; exists {
		t.Fatalf("friendly catalogue must be normalized away: %+v", selection.Parameters)
	}

	exposedSkills, _ := provider.req.Context["skills"].([]map[string]any)
	if len(exposedSkills) != 1 {
		t.Fatalf("expected one exposed skill, got %+v", exposedSkills)
	}
	schema, _ := exposedSkills[0]["input_schema"].(map[string]any)
	properties, _ := schema["properties"].(map[string]any)
	if properties["provider_id"] == nil || properties["setup_payload"] == nil || properties["provider_secret"] == nil {
		t.Fatalf("integration planner schema missing provider/setup/optional-secret inputs: %+v", schema)
	}
	required, _ := schema["required"].([]string)
	if !slices.Contains(required, "provider_id") || !slices.Contains(required, "setup_payload") || slices.Contains(required, "provider_secret") {
		t.Fatalf("integration schema required fields are unsafe or incomplete: %+v", schema)
	}
}

func TestNormalizeIntegrationPlannerParametersCanonicalizesLiveBrowserAuthProse(t *testing.T) {
	t.Parallel()

	profileID := "c5f12408-10b1-4532-8b46-1fbdead1442f"
	parameters := normalizeIntegrationPlannerParameters("cabinet.integrations.configure_provider", map[string]any{
		"provider_id":     "voglers",
		"setup_payload":   "Configure Voglers for its public catalogue and enable it for profile " + profileID + "; do not use or request an API key or secret.",
		"provider_secret": "  ",
		"api_key":         nil,
	})
	for key, want := range map[string]any{
		"provider_id": "voglers", "setup_payload": "public_catalogue", "setup_step": "public_catalogue", "marketplace": "public",
	} {
		if got := parameters[key]; got != want {
			t.Fatalf("normalized live Browser Auth parameter %s=%v, want %v; all=%+v", key, got, want, parameters)
		}
	}
	serialized := fmt.Sprint(parameters)
	for _, forbidden := range []string{profileID, "api key", "secret"} {
		if strings.Contains(strings.ToLower(serialized), forbidden) {
			t.Fatalf("live prompt prose reached normalized provider parameters (%s): %+v", forbidden, parameters)
		}
	}
	if containsSensitiveAgentSkillPreviewValue(parameters) {
		t.Fatalf("empty optional secret fields must not route a Browser Auth preview through secure storage: %+v", parameters)
	}
	nonEmptySecret := normalizeIntegrationPlannerParameters("cabinet.integrations.configure_provider", map[string]any{
		"provider_id":     "voglers",
		"setup_payload":   "public_catalogue",
		"provider_secret": "must-remain-governed",
	})
	if !containsSensitiveAgentSkillPreviewValue(nonEmptySecret) || nonEmptySecret["provider_secret"] != "must-remain-governed" {
		t.Fatalf("non-empty provider secret was not retained for governed secure handling: %+v", nonEmptySecret)
	}

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Live Browser Auth Canonicalization"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if _, _, err := persistAgentProviderSettings(context.Background(), a.db, profile.ID, "voglers", parameters); err != nil {
		t.Fatalf("persist canonical Browser Auth provider settings: %v", err)
	}
	assertProfileSetting(t, a, profile.ID, "integration.voglers.enabled", "true")
	assertProfileSetting(t, a, profile.ID, "integration.voglers.setup_step", "public_catalogue")
	assertProfileSetting(t, a, profile.ID, "integration.voglers.marketplace", "public")
}

func TestChatAgentPlannerNormalizesSchemaRefWrappedProviderParameters(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Browser Auth Planner Parameters"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	threadResp := doRequest(t, a, http.MethodPost, "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profile.ID+`","title":"Browser Auth Planner Parameters"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	provider := &captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.wishlist.create_entry","parameters":{"wishlist.entry.create":"{\"part_number\":\"CHAT-WISH-BROWSER-AUTH\",\"title\":\"Browser Auth Wishlist\",\"target_price\":37,\"currency\":\"AUD\"}"},"message":"I prepared the wishlist preview."}`}
	result, handled := dispatchChatAgentProviderPlanner(
		context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(provider),
		agentskills.NewRegistry(nil),
		profile.ID,
		thread.ID,
		"Prepare a wishlist entry named Browser Auth Wishlist with target price AUD 37.",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "gpt-5.6-luna"},
			"agent_context": map[string]any{
				"profile_id": profile.ID, "thread_id": thread.ID, "route_id": "/chats/",
				"surface_id": "chats.main", "source_channel": "in-app", "setup_state": "ready",
				"permission_state": "ask_before_local_changes",
			},
		},
		"message-browser-auth-parameters",
	)
	if !handled {
		t.Fatal("expected Browser Auth wishlist request to enter provider planner")
	}
	exposedSkills, _ := provider.req.Context["skills"].([]map[string]any)
	var wishlistSchema map[string]any
	for _, exposedSkill := range exposedSkills {
		if exposedSkill["id"] == "cabinet.wishlist.create_entry" {
			wishlistSchema, _ = exposedSkill["input_schema"].(map[string]any)
			break
		}
	}
	properties, _ := wishlistSchema["properties"].(map[string]any)
	required, _ := wishlistSchema["required"].([]string)
	if properties["part_number"] == nil || properties["title"] == nil || properties["target_price"] == nil || properties["currency"] == nil ||
		!slices.Contains(required, "part_number") || !slices.Contains(required, "title") || slices.Contains(required, "target_price") {
		t.Fatalf("Browser Auth planner schema must expose required identity and optional wishlist planning fields, got %+v", wishlistSchema)
	}
	preview, ok := result["preview_result"].(map[string]any)
	if !ok || preview["blocker"] != "confirmation_required" {
		t.Fatalf("expected executable wishlist preview from schema-ref parameters, got %+v", result)
	}
	parameters, _ := preview["parameters"].(map[string]any)
	if parameters["part_number"] != "CHAT-WISH-BROWSER-AUTH" || parameters["title"] != "Browser Auth Wishlist" || fmt.Sprint(parameters["target_price"]) != "37" || parameters["currency"] != "AUD" {
		t.Fatalf("expected flattened Browser Auth parameters, got %+v", parameters)
	}
	if _, wrapped := parameters["wishlist.entry.create"]; wrapped {
		t.Fatalf("schema-ref wrapper must not reach the durable preview: %+v", parameters)
	}
	applyRequest, _ := preview["apply_request"].(map[string]any)
	applyBody, err := json.Marshal(applyRequest)
	if err != nil {
		t.Fatalf("marshal generated apply request: %v", err)
	}
	apply := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(string(applyBody)), map[string]string{"Content-Type": "application/json"})
	if apply.Code != http.StatusOK {
		t.Fatalf("generated Browser Auth apply status=%d body=%s", apply.Code, apply.Body.String())
	}
	for _, token := range []string{`"skill_id":"cabinet.wishlist.create_entry"`, `"mutation_applied":true`, `"operation":"wishlist.entry.create"`, `"currency":"AUD"`} {
		if !strings.Contains(apply.Body.String(), token) {
			t.Fatalf("generated Browser Auth apply response missing %s: %s", token, apply.Body.String())
		}
	}

	incomplete, handled := dispatchChatAgentProviderPlanner(
		context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.wishlist.create_entry","parameters":{"wishlist.entry.create":"{\"name\":\"Incomplete Browser Auth Wishlist\",\"target_price\":37,\"currency\":\"AUD\"}"},"message":"I prepared the wishlist preview."}`}),
		agentskills.NewRegistry(nil),
		profile.ID,
		thread.ID,
		"Prepare a wishlist entry named Incomplete Browser Auth Wishlist with target price AUD 37.",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "gpt-5.6-luna"},
			"agent_context": map[string]any{
				"profile_id": profile.ID, "thread_id": thread.ID, "route_id": "/chats/",
				"surface_id": "chats.main", "source_channel": "in-app", "setup_state": "ready",
				"permission_state": "ask_before_local_changes",
			},
		},
		"message-browser-auth-incomplete-parameters",
	)
	if !handled {
		t.Fatal("expected incomplete Browser Auth wishlist request to enter provider planner")
	}
	if incomplete["preview_result"] != nil || incomplete["decision"] != "clarify" {
		t.Fatalf("incomplete Browser Auth parameters must ask for context without an actionable preview: %+v", incomplete)
	}
	errPayload, _ := incomplete["error"].(map[string]any)
	if errPayload["code"] != "missing_context" {
		t.Fatalf("incomplete Browser Auth parameters must expose missing_context, got %+v", incomplete)
	}
}

func TestChatAgentPlannerRejectsOrClarifiesUnsafeSelections(t *testing.T) {
	t.Parallel()

	registry := agentskills.NewRegistry(nil)
	tests := []struct {
		name         string
		responseText string
		wantDecision string
		wantCode     string
	}{
		{
			name:         "disabled unavailable skill",
			responseText: `{"decision":"select_skill","skill_id":"cabinet.inventory.create_item","parameters":{"title":"Bad mutation"},"message":"Creating it now."}`,
			wantDecision: "reject",
			wantCode:     "skill_unavailable",
		},
		{
			name:         "missing selected skill",
			responseText: `{"decision":"select_skill","parameters":{"query":"AFX"},"message":"I will search."}`,
			wantDecision: "clarify",
			wantCode:     "skill_required",
		},
		{
			name:         "unsupported decision",
			responseText: `{"decision":"apply_now","skill_id":"cabinet.inventory.search_items","parameters":{"query":"AFX"},"message":"Done."}`,
			wantDecision: "reject",
			wantCode:     "unsupported_planner_decision",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			provider := &captureAssistantProvider{responseText: tt.responseText}
			selection, err := planChatAgentSkill(context.Background(), provider, chatAgentPlannerInput{
				ProfileID: "profile-planner",
				ThreadID:  "thread-planner",
				Intent:    "Find the item with part number AFX-123",
				Skills: []agentskills.Skill{
					mustResolveSkill(t, registry, "cabinet.inventory.search_items"),
				},
			})
			if err != nil {
				t.Fatalf("planChatAgentSkill() error = %v", err)
			}
			if selection.Decision != tt.wantDecision || selection.ErrorCode != tt.wantCode {
				t.Fatalf("expected decision/code %s/%s, got %+v", tt.wantDecision, tt.wantCode, selection)
			}
			if selection.Message == "" || strings.Contains(strings.ToLower(selection.Message), "done") {
				t.Fatalf("expected truthful non-completion guidance, got %+v", selection)
			}
		})
	}
}

func TestChatAgentPlannerDispatchInvokesProviderNeutralRuntime(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Runtime"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != 201 {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+p.ID+`","title":"Planner Runtime"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != 201 {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	provider := &captureAssistantProvider{}
	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(provider),
		agentskills.NewRegistry(nil),
		p.ID,
		thread.ID,
		"Find the item with part number AFX-123",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     p.ID,
				"thread_id":      thread.ID,
				"route_id":       "/chats",
				"surface_id":     "chats.main",
				"source_channel": "in-app",
				"intent_text":    "Find the item with part number AFX-123",
			},
		},
		"message-provider-runtime",
	)
	if !handled {
		t.Fatal("expected natural-language planner dispatch to handle item lookup")
	}
	if result["skill_id"] != "cabinet.inventory.search_items" {
		t.Fatalf("expected provider-selected inventory search skill, got %+v", result)
	}
	if provider.req.ProfileID != p.ID || provider.req.ThreadID != thread.ID || provider.req.Model != "fake-planner-model" {
		t.Fatalf("provider-neutral runtime request missing chat context: %+v", provider.req)
	}
	if provider.req.Metadata["entry_point"] != "chat.agent_planner" || provider.req.Metadata["governed_dispatch_owner"] != "cabinet" {
		t.Fatalf("provider-neutral request missing governed metadata: %+v", provider.req.Metadata)
	}
}

func TestChatAgentPlannerExecutesReadOnlySelectionWithProfileIsolation(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createA := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Read A"}`), map[string]string{"Content-Type": "application/json"})
	createB := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Read B"}`), map[string]string{"Content-Type": "application/json"})
	if createA.Code != 201 || createB.Code != 201 {
		t.Fatalf("create profile statuses=%d/%d bodies=%s / %s", createA.Code, createB.Code, createA.Body.String(), createB.Body.String())
	}
	var profileA, profileB struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createA.Body).Decode(&profileA); err != nil {
		t.Fatalf("decode profile A: %v", err)
	}
	if err := json.NewDecoder(createB.Body).Decode(&profileB); err != nil {
		t.Fatalf("decode profile B: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, status, priority, notes, source_urls_json) VALUES (?, ?, 'AFX', 'Slot', 'PLAN-READ-1', 'Visible Planner Item', 'active', 'medium', 'sk-read-result-must-not-leak', '["https://private.example.test/item"]')`, "planner-read-visible", profileA.ID); err != nil {
		t.Fatalf("seed profile A item: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, status, priority) VALUES (?, ?, 'AFX', 'Slot', 'PLAN-READ-HIDDEN', 'Hidden Planner Item', 'active', 'medium')`, "planner-read-hidden", profileB.ID); err != nil {
		t.Fatalf("seed profile B item: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profileA.ID+`","title":"Planner Read"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != 201 {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.inventory.search_items","parameters":{"query":"PLAN-READ-1"},"message":"Searching inventory for PLAN-READ-1."}`}),
		agentskills.NewRegistry(nil),
		profileA.ID,
		thread.ID,
		"Find the item with part number PLAN-READ-1",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profileA.ID,
				"thread_id":      thread.ID,
				"route_id":       "/chats",
				"surface_id":     "chats.main",
				"source_channel": "in-app",
				"intent_text":    "Find the item with part number PLAN-READ-1",
			},
		},
		"message-read-only",
	)
	if !handled {
		t.Fatal("expected planner dispatch to handle read-only selection")
	}
	execution, ok := result["execution_result"].(map[string]any)
	if !ok || execution["read_only"] != true || execution["profile_id"] != profileA.ID {
		t.Fatalf("expected read-only execution result for profile A, got %+v", result)
	}
	body, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	if !strings.Contains(string(body), "Visible Planner Item") || strings.Contains(string(body), "Hidden Planner Item") {
		t.Fatalf("read-only planner result must be grounded and profile-isolated, body=%s", string(body))
	}
	threadMessage, ok := result["thread_message"].(chat.Message)
	if !ok {
		t.Fatalf("planner result missing trusted assistant thread message: %+v", result)
	}
	agentResponseJSON, err := json.Marshal(threadMessage.Context["agent_response"])
	if err != nil {
		t.Fatalf("marshal persisted Agent response: %v", err)
	}
	var agentResponse chat.AgentResponse
	if err := json.Unmarshal(agentResponseJSON, &agentResponse); err != nil {
		t.Fatalf("decode persisted Agent response: %v", err)
	}
	if agentResponse.ResultSummary == nil {
		t.Fatalf("assistant response missing server-owned read result summary: %+v", threadMessage.Context)
	}
	if agentResponse.ResultSummary.Total != 1 || len(agentResponse.ResultSummary.Items) != 1 {
		t.Fatalf("read result summary = %+v, want one bounded item", agentResponse.ResultSummary)
	}
	item := agentResponse.ResultSummary.Items[0]
	if item.PartNumber != "PLAN-READ-1" || item.Title != "Visible Planner Item" {
		t.Fatalf("read result summary item = %+v, want exact part number and title", item)
	}
	responseJSON, err := json.Marshal(agentResponse)
	if err != nil {
		t.Fatalf("marshal normalized Agent response: %v", err)
	}
	for _, forbidden := range []string{"sk-read-result-must-not-leak", "private.example.test", "Hidden Planner Item"} {
		if strings.Contains(string(responseJSON), forbidden) {
			t.Fatalf("normalized read result leaked %q: %s", forbidden, responseJSON)
		}
	}
}

func TestChatAgentPlannerRoutesDashboardActivitySummaryFromMainChat(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createA := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Dashboard A"}`), map[string]string{"Content-Type": "application/json"})
	createB := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Dashboard B"}`), map[string]string{"Content-Type": "application/json"})
	if createA.Code != 201 || createB.Code != 201 {
		t.Fatalf("create profile statuses=%d/%d bodies=%s / %s", createA.Code, createB.Code, createA.Body.String(), createB.Body.String())
	}
	var profileA, profileB struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createA.Body).Decode(&profileA); err != nil {
		t.Fatalf("decode profile A: %v", err)
	}
	if err := json.NewDecoder(createB.Body).Decode(&profileB); err != nil {
		t.Fatalf("decode profile B: %v", err)
	}
	if _, err := a.db.Exec(`
		INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, created_at)
		VALUES
			('planner-dash-item-a', ?, 'AFX', 'Slot', 'PDA-1', 'Planner Dashboard A Camaro', '2026-06-01T10:00:00Z'),
			('planner-dash-item-b', ?, 'AFX', 'Slot', 'PDB-1', 'Planner Dashboard B Porsche', '2026-06-02T10:00:00Z');
		INSERT INTO instances(id, item_id, condition, status, quantity, storage_location, acquisition_price, acquisition_date, notes)
		VALUES
			('planner-dash-inst-a','planner-dash-item-a','used','loose',2,'shelf',15,'',''),
			('planner-dash-inst-b','planner-dash-item-b','used','loose',7,'case',25,'','');
		INSERT INTO scanner_query_sets(id, profile_id, name, keywords_json, exclusions_json)
		VALUES ('planner-dash-query-a', ?, 'A', '["pda"]', '[]'),('planner-dash-query-b', ?, 'B', '["pdb"]', '[]');
		INSERT INTO scanner_candidates(id, profile_id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source, stock_state, stock_count)
		VALUES
			('planner-dash-cand-a', ?, 'planner-dash-query-a', 'PDA-L1', 'Planner Dashboard A PDA-1', 20, 0, 'http://a.example', '', 'seller-a', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'low_stock', 2),
			('planner-dash-cand-b', ?, 'planner-dash-query-b', 'PDB-L1', 'Planner Dashboard B PDB-1', 20, 0, 'http://b.example', '', 'seller-b', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'low_stock', 2);
		INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at)
		VALUES
			('planner-dash-cand-a', '', 'not_in_collection', 0, 1, 'PDA-1', CURRENT_TIMESTAMP),
			('planner-dash-cand-b', '', 'not_in_collection', 0, 1, 'PDB-1', CURRENT_TIMESTAMP);
		INSERT INTO wishlist_entries(id, profile_id, item_id, target_price, priority, notes, highlight_hit)
		VALUES
			('planner-dash-wish-a', ?, 'planner-dash-item-a', 30, 'high', '', 1),
			('planner-dash-wish-b', ?, 'planner-dash-item-b', 30, 'high', '', 1);
		INSERT INTO price_snapshots(id, item_id, snapshot_date, source, min_price, median_price, latest_price, stock_count)
		VALUES
			('planner-dash-price-a1','planner-dash-item-a','2026-02-20','ebay',15,15,15,0),
			('planner-dash-price-a2','planner-dash-item-a','2026-02-21','ebay',12,12,12,4),
			('planner-dash-price-b1','planner-dash-item-b','2026-02-20','ebay',25,25,25,0),
			('planner-dash-price-b2','planner-dash-item-b','2026-02-21','ebay',22,22,22,6)
	`, profileA.ID, profileB.ID, profileA.ID, profileB.ID, profileA.ID, profileB.ID, profileA.ID, profileB.ID); err != nil {
		t.Fatalf("seed dashboard planner data: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profileA.ID+`","title":"Planner Dashboard"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != 201 {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.dashboard.summarise_activity","parameters":{"window":"today","provider_secret":"sk-planner-dashboard-secret"},"message":"Summarising the current Dashboard snapshot."}`}),
		agentskills.NewRegistry(nil),
		profileA.ID,
		thread.ID,
		"Summarise what changed on my Dashboard today",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profileA.ID,
				"workspace_id":   "workspace-planner",
				"thread_id":      thread.ID,
				"route_id":       "/chats",
				"surface_id":     "chats.main",
				"source_channel": "in-app",
				"setup_state":    "ready",
				"intent_text":    "Summarise what changed on my Dashboard today",
			},
		},
		"message-dashboard-summary",
	)
	if !handled {
		t.Fatal("expected main Chat planner dispatch to handle Dashboard summary request")
	}
	if result["skill_id"] != "cabinet.dashboard.summarise_activity" || result["preview_result"] != nil {
		t.Fatalf("expected read-only Dashboard skill execution without preview, got %+v", result)
	}
	execution, ok := result["execution_result"].(map[string]any)
	if !ok || execution["read_only"] != true || execution["profile_id"] != profileA.ID || execution["mutation_applied"] == true {
		t.Fatalf("expected read-only profile-scoped Dashboard execution result, got %+v", result)
	}
	window, _ := execution["time_window"].(map[string]any)
	if window["requested_window"] != "today" || window["snapshot_only"] != true || window["evidence_backed"] != false {
		t.Fatalf("expected truthful snapshot-only window caveat, got %+v", execution)
	}
	evidence, _ := result["evidence"].(map[string]any)
	tokenState, _ := evidence["preview_apply_token_state"].(map[string]any)
	contextSummary, _ := evidence["context"].(map[string]any)
	if evidence["entry_point"] != "chat.agent_planner" ||
		evidence["selected_skill"] != "cabinet.dashboard.summarise_activity" ||
		evidence["raw_provider_payload_kept"] != false ||
		tokenState["confirmation_state"] != "not_required" ||
		tokenState["apply_state"] != "not_applicable" ||
		tokenState["mutation_applied"] != false ||
		contextSummary["surface_id"] != "chats.main" ||
		contextSummary["route_id"] != "/chats" ||
		contextSummary["thread_id"] != thread.ID {
		t.Fatalf("expected governed main Chat planner evidence without apply token, evidence=%+v", evidence)
	}
	body, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal dashboard planner result: %v", err)
	}
	bodyText := string(body)
	for _, want := range []string{
		"Planner Dashboard A Camaro",
		`"operation":"dashboard.activity.summary"`,
		`"snapshot_only":true`,
		`"new_discoveries"`,
		`"price_drops"`,
		`"restocks"`,
	} {
		if !strings.Contains(bodyText, want) {
			t.Fatalf("dashboard planner response missing %s: body=%s", want, bodyText)
		}
	}
	if strings.Contains(bodyText, "Planner Dashboard B Porsche") ||
		strings.Contains(bodyText, "sk-planner-dashboard-secret") ||
		strings.Contains(bodyText, "preview_id") {
		t.Fatalf("dashboard planner response leaked wrong profile, secret, or preview token: body=%s", bodyText)
	}
	threadMessage, ok := result["thread_message"].(chat.Message)
	if !ok {
		t.Fatalf("dashboard planner result missing trusted assistant thread message: %+v", result)
	}
	agentResponseJSON, err := json.Marshal(threadMessage.Context["agent_response"])
	if err != nil {
		t.Fatalf("marshal dashboard Agent response: %v", err)
	}
	var agentResponse chat.AgentResponse
	if err := json.Unmarshal(agentResponseJSON, &agentResponse); err != nil {
		t.Fatalf("decode dashboard Agent response: %v", err)
	}
	if agentResponse.ResultSummary == nil || agentResponse.ResultSummary.Kind != "dashboard_activity" {
		t.Fatalf("dashboard response missing typed server-owned summary: %+v", agentResponse)
	}
	if len(agentResponse.ResultSummary.Metrics) == 0 || len(agentResponse.ResultSummary.Items) == 0 {
		t.Fatalf("dashboard summary must include bounded signals and recent records: %+v", agentResponse.ResultSummary)
	}
	if agentResponse.ResultSummary.Items[0].Title != "Planner Dashboard A Camaro" {
		t.Fatalf("dashboard summary recent item = %+v, want profile A record", agentResponse.ResultSummary.Items[0])
	}
	summaryJSON, err := json.Marshal(agentResponse.ResultSummary)
	if err != nil {
		t.Fatalf("marshal dashboard result summary: %v", err)
	}
	for _, want := range []string{"New discoveries", "Wishlist hits", "Price drops"} {
		if !strings.Contains(string(summaryJSON), want) {
			t.Fatalf("dashboard result summary missing metric %q: %s", want, summaryJSON)
		}
	}
	for _, forbidden := range []string{"Planner Dashboard B Porsche", "sk-planner-dashboard-secret", "http://a.example", "seller-a"} {
		if strings.Contains(string(summaryJSON), forbidden) {
			t.Fatalf("dashboard result summary leaked %q: %s", forbidden, summaryJSON)
		}
	}
}

func TestChatAgentPlannerRoutesStorageStatusSummaryFromMainChat(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Storage"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != 201 {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profile.ID+`","title":"Planner Storage"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != 201 {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.storage.show_status","parameters":{"provider_secret":"sk-planner-storage-secret"},"message":"Checking storage status."}`}),
		agentskills.NewRegistry(nil),
		profile.ID,
		thread.ID,
		"Show my storage and backup status",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profile.ID,
				"workspace_id":   "workspace-planner",
				"thread_id":      thread.ID,
				"route_id":       "/chats",
				"surface_id":     "chats.main",
				"source_channel": "in-app",
				"setup_state":    "ready",
				"intent_text":    "Show my storage and backup status",
			},
		},
		"message-storage-status-summary",
	)
	if !handled {
		t.Fatal("expected main Chat planner dispatch to handle Storage status request")
	}
	if result["skill_id"] != "cabinet.storage.show_status" || result["preview_result"] != nil {
		t.Fatalf("expected read-only Storage skill execution without preview, got %+v", result)
	}
	execution, ok := result["execution_result"].(map[string]any)
	if !ok || execution["read_only"] != true || execution["profile_id"] != profile.ID || execution["mutation_applied"] == true {
		t.Fatalf("expected read-only profile-scoped Storage execution result, got %+v", result)
	}
	body, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal storage planner result: %v", err)
	}
	bodyText := string(body)
	for _, want := range []string{
		`"operation":"storage.status.show"`,
		`"storage_status":"available"`,
		`"backup_status":"not_configured"`,
	} {
		if !strings.Contains(bodyText, want) {
			t.Fatalf("storage planner response missing %s: body=%s", want, bodyText)
		}
	}
	if strings.Contains(bodyText, "sk-planner-storage-secret") || strings.Contains(bodyText, "preview_id") {
		t.Fatalf("storage planner response leaked secret or preview token: body=%s", bodyText)
	}
	threadMessage, ok := result["thread_message"].(chat.Message)
	if !ok {
		t.Fatalf("storage planner result missing trusted assistant thread message: %+v", result)
	}
	agentResponseJSON, err := json.Marshal(threadMessage.Context["agent_response"])
	if err != nil {
		t.Fatalf("marshal storage Agent response: %v", err)
	}
	var agentResponse chat.AgentResponse
	if err := json.Unmarshal(agentResponseJSON, &agentResponse); err != nil {
		t.Fatalf("decode storage Agent response: %v", err)
	}
	if agentResponse.ResultSummary == nil || agentResponse.ResultSummary.Kind != "storage_status" {
		t.Fatalf("storage response missing typed server-owned summary: %+v", agentResponse)
	}
	if agentResponse.ResultSummary.Total != 1 || len(agentResponse.ResultSummary.Items) != 1 {
		t.Fatalf("storage summary must expose one bounded status record: %+v", agentResponse.ResultSummary)
	}
	summaryJSON, err := json.Marshal(agentResponse.ResultSummary)
	if err != nil {
		t.Fatalf("marshal storage result summary: %v", err)
	}
	for _, want := range []string{"Storage status", "available", "Backup: not_configured"} {
		if !strings.Contains(string(summaryJSON), want) {
			t.Fatalf("storage result summary missing %q: %s", want, summaryJSON)
		}
	}
	for _, forbidden := range []string{"sk-planner-storage-secret", "preview_id", "backup_path", "backup_target"} {
		if strings.Contains(string(summaryJSON), forbidden) {
			t.Fatalf("storage result summary leaked %q: %s", forbidden, summaryJSON)
		}
	}
}

func TestChatAgentPlannerRoutesDataExportSummaryFromMainChat(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Data Export"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != 201 {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profile.ID+`","title":"Planner Data Export"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != 201 {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.data.export_bundle","parameters":{"export_scope":"all","provider_secret":"sk-planner-export-secret","backup_path":"C:/private/cabinet.db"},"message":"Preparing an export summary."}`}),
		agentskills.NewRegistry(nil),
		profile.ID,
		thread.ID,
		"Export all Cabinet data",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profile.ID,
				"workspace_id":   "workspace-planner",
				"thread_id":      thread.ID,
				"route_id":       "/settings/storage",
				"surface_id":     "settings.data.export",
				"source_channel": "in-app",
				"setup_state":    "ready",
				"intent_text":    "Export all Cabinet data",
			},
		},
		"message-data-export-summary",
	)
	if !handled {
		t.Fatal("expected main Chat planner dispatch to handle Data export request")
	}
	if result["skill_id"] != "cabinet.data.export_bundle" || result["preview_result"] != nil {
		t.Fatalf("expected non-mutating Data export execution without preview, got %+v", result)
	}
	execution, ok := result["execution_result"].(map[string]any)
	if !ok || execution["read_only"] != true || execution["profile_id"] != profile.ID || execution["mutation_applied"] == true {
		t.Fatalf("expected profile-scoped non-mutating Data export execution result, got %+v", result)
	}
	body, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal data export planner result: %v", err)
	}
	bodyText := string(body)
	for _, want := range []string{
		`"operation":"data.export.bundle"`,
		`"export_scope":"all"`,
		`"status":"ready"`,
	} {
		if !strings.Contains(bodyText, want) {
			t.Fatalf("data export planner response missing %s: body=%s", want, bodyText)
		}
	}
	for _, forbidden := range []string{"sk-planner-export-secret", "preview_id", "backup_path", "C:/private", "export_artifact_path", "configuration_payload"} {
		if strings.Contains(bodyText, forbidden) {
			t.Fatalf("data export planner response leaked %q: body=%s", forbidden, bodyText)
		}
	}
	threadMessage, ok := result["thread_message"].(chat.Message)
	if !ok {
		t.Fatalf("data export planner result missing trusted assistant thread message: %+v", result)
	}
	agentResponseJSON, err := json.Marshal(threadMessage.Context["agent_response"])
	if err != nil {
		t.Fatalf("marshal data export Agent response: %v", err)
	}
	var agentResponse chat.AgentResponse
	if err := json.Unmarshal(agentResponseJSON, &agentResponse); err != nil {
		t.Fatalf("decode data export Agent response: %v", err)
	}
	if agentResponse.State != chat.AgentResponseReadResult || agentResponse.Preview != nil || agentResponse.NextAction != nil {
		t.Fatalf("data export response must be a read result without preview/apply controls: %+v", agentResponse)
	}
	if agentResponse.ResultSummary == nil || agentResponse.ResultSummary.Kind != "data_export_bundle" {
		t.Fatalf("data export response missing typed server-owned summary: %+v", agentResponse)
	}
	if agentResponse.ResultSummary.Total != 1 || len(agentResponse.ResultSummary.Items) != 1 {
		t.Fatalf("data export summary must expose one bounded readiness record: %+v", agentResponse.ResultSummary)
	}
	summaryJSON, err := json.Marshal(agentResponse.ResultSummary)
	if err != nil {
		t.Fatalf("marshal data export result summary: %v", err)
	}
	for _, want := range []string{"Data export bundle", "ready", "Scope: all"} {
		if !strings.Contains(string(summaryJSON), want) {
			t.Fatalf("data export result summary missing %q: %s", want, summaryJSON)
		}
	}
	for _, forbidden := range []string{"sk-planner-export-secret", "preview_id", "backup_path", "C:/private", "raw_payload", "export_artifact_path", "configuration_payload", "apply"} {
		if strings.Contains(string(summaryJSON), forbidden) {
			t.Fatalf("data export result summary leaked %q: %s", forbidden, summaryJSON)
		}
	}
}

func TestChatAgentPlannerRoutesMaintenanceSafeCheckSummaryFromMainChat(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Maintenance"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != 201 {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profile.ID+`","title":"Planner Maintenance"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != 201 {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.maintenance.run_safe_check","parameters":{"maintenance_check":"database","check_level":"safe","provider_secret":"sk-planner-maintenance-secret","backup_path":"C:/private/cabinet.db"},"message":"Checking Cabinet maintenance health."}`}),
		agentskills.NewRegistry(nil),
		profile.ID,
		thread.ID,
		"Run a safe Cabinet maintenance health check",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profile.ID,
				"workspace_id":   "workspace-planner",
				"thread_id":      thread.ID,
				"route_id":       "/settings/storage",
				"surface_id":     "settings.maintenance",
				"source_channel": "in-app",
				"setup_state":    "ready",
				"intent_text":    "Run a safe Cabinet maintenance health check",
			},
		},
		"message-maintenance-summary",
	)
	if !handled {
		t.Fatal("expected main Chat planner dispatch to handle Maintenance safe-check request")
	}
	if result["skill_id"] != "cabinet.maintenance.run_safe_check" || result["preview_result"] != nil {
		t.Fatalf("expected non-mutating Maintenance execution without preview, got %+v", result)
	}
	execution, ok := result["execution_result"].(map[string]any)
	if !ok || execution["read_only"] != true || execution["profile_id"] != profile.ID || execution["mutation_applied"] == true {
		t.Fatalf("expected profile-scoped non-mutating Maintenance execution result, got %+v", result)
	}
	body, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal maintenance planner result: %v", err)
	}
	bodyText := string(body)
	for _, want := range []string{
		`"operation":"maintenance.safe_check"`,
		`"maintenance_check":"database"`,
		`"check_level":"safe"`,
		`"status":"healthy"`,
	} {
		if !strings.Contains(bodyText, want) {
			t.Fatalf("maintenance planner response missing %s: body=%s", want, bodyText)
		}
	}
	if strings.Contains(bodyText, "sk-planner-maintenance-secret") || strings.Contains(bodyText, "preview_id") {
		t.Fatalf("maintenance planner response leaked secret or preview token: body=%s", bodyText)
	}
	threadMessage, ok := result["thread_message"].(chat.Message)
	if !ok {
		t.Fatalf("maintenance planner result missing trusted assistant thread message: %+v", result)
	}
	agentResponseJSON, err := json.Marshal(threadMessage.Context["agent_response"])
	if err != nil {
		t.Fatalf("marshal maintenance Agent response: %v", err)
	}
	var agentResponse chat.AgentResponse
	if err := json.Unmarshal(agentResponseJSON, &agentResponse); err != nil {
		t.Fatalf("decode maintenance Agent response: %v", err)
	}
	if agentResponse.ResultSummary == nil || agentResponse.ResultSummary.Kind != "maintenance_safe_check" {
		t.Fatalf("maintenance response missing typed server-owned summary: %+v", agentResponse)
	}
	if agentResponse.ResultSummary.Total != 1 || len(agentResponse.ResultSummary.Items) != 1 {
		t.Fatalf("maintenance summary must expose one bounded health record: %+v", agentResponse.ResultSummary)
	}
	summaryJSON, err := json.Marshal(agentResponse.ResultSummary)
	if err != nil {
		t.Fatalf("marshal maintenance result summary: %v", err)
	}
	for _, want := range []string{"Maintenance safe check", "healthy", "Check: database / safe"} {
		if !strings.Contains(string(summaryJSON), want) {
			t.Fatalf("maintenance result summary missing %q: %s", want, summaryJSON)
		}
	}
	for _, forbidden := range []string{"sk-planner-maintenance-secret", "preview_id", "backup_path", "C:/private", "raw_payload"} {
		if strings.Contains(string(summaryJSON), forbidden) {
			t.Fatalf("maintenance result summary leaked %q: %s", forbidden, summaryJSON)
		}
	}
}

func TestChatAgentPlannerRoutesWishlistSearchSummaryFromMainChat(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createA := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Wishlist A"}`), map[string]string{"Content-Type": "application/json"})
	createB := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Wishlist B"}`), map[string]string{"Content-Type": "application/json"})
	if createA.Code != 201 || createB.Code != 201 {
		t.Fatalf("create profiles status A=%d B=%d body A=%s body B=%s", createA.Code, createB.Code, createA.Body.String(), createB.Body.String())
	}
	var profileA, profileB struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createA.Body).Decode(&profileA); err != nil {
		t.Fatalf("decode profile A: %v", err)
	}
	if err := json.NewDecoder(createB.Body).Decode(&profileB); err != nil {
		t.Fatalf("decode profile B: %v", err)
	}
	if _, err := a.db.Exec(`
		INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, status)
		VALUES
			('planner-wishlist-item-a', ?, 'AFX', 'Slot', 'PWA-1', 'Planner Wishlist A Camaro', 'wishlist'),
			('planner-wishlist-item-b', ?, 'AFX', 'Slot', 'PWB-1', 'Planner Wishlist B Porsche', 'wishlist');
		INSERT INTO wishlist_entries(id, profile_id, item_id, target_price, currency, priority, notes, highlight_hit, below_target_now, purchase_url)
		VALUES
			('planner-wishlist-entry-a', ?, 'planner-wishlist-item-a', 30, 'AUD', 'high', 'private wishlist note must not render', 1, 1, 'https://private.example/wishlist-a'),
			('planner-wishlist-entry-b', ?, 'planner-wishlist-item-b', 30, 'AUD', 'high', 'other profile note', 1, 1, 'https://private.example/wishlist-b');
	`, profileA.ID, profileB.ID, profileA.ID, profileB.ID); err != nil {
		t.Fatalf("seed wishlist planner data: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profileA.ID+`","title":"Planner Wishlist"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != 201 {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.wishlist.search_entries","parameters":{"query":"planner-wishlist-entry","provider_secret":"sk-planner-wishlist-secret"},"message":"Searching wishlist entries."}`}),
		agentskills.NewRegistry(nil),
		profileA.ID,
		thread.ID,
		"Find my wishlist entries for planner wishlist",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profileA.ID,
				"workspace_id":   "workspace-planner",
				"thread_id":      thread.ID,
				"route_id":       "/chats",
				"surface_id":     "chats.main",
				"source_channel": "in-app",
				"setup_state":    "ready",
				"intent_text":    "Find my wishlist entries for planner wishlist",
			},
		},
		"message-wishlist-summary",
	)
	if !handled {
		t.Fatal("expected main Chat planner dispatch to handle Wishlist search request")
	}
	if result["skill_id"] != "cabinet.wishlist.search_entries" || result["preview_result"] != nil {
		t.Fatalf("expected read-only Wishlist skill execution without preview, got %+v", result)
	}
	execution, ok := result["execution_result"].(map[string]any)
	if !ok || execution["read_only"] != true || execution["profile_id"] != profileA.ID || execution["mutation_applied"] == true {
		t.Fatalf("expected read-only profile-scoped Wishlist execution result, got %+v", result)
	}
	threadMessage, ok := result["thread_message"].(chat.Message)
	if !ok {
		t.Fatalf("wishlist planner result missing trusted assistant thread message: %+v", result)
	}
	agentResponseJSON, err := json.Marshal(threadMessage.Context["agent_response"])
	if err != nil {
		t.Fatalf("marshal wishlist Agent response: %v", err)
	}
	var agentResponse chat.AgentResponse
	if err := json.Unmarshal(agentResponseJSON, &agentResponse); err != nil {
		t.Fatalf("decode wishlist Agent response: %v", err)
	}
	if agentResponse.ResultSummary == nil || agentResponse.ResultSummary.Kind != "wishlist_entries" {
		t.Fatalf("wishlist response missing typed server-owned summary: %+v", agentResponse)
	}
	if agentResponse.ResultSummary.Total != 1 || len(agentResponse.ResultSummary.Items) != 1 {
		t.Fatalf("wishlist summary must expose one bounded active-profile entry: %+v", agentResponse.ResultSummary)
	}
	item := agentResponse.ResultSummary.Items[0]
	if item.ID != "planner-wishlist-entry-a" || item.Title != "planner-wishlist-item-a" || item.Status != "below_target" || item.Category != "high" {
		t.Fatalf("wishlist summary item = %+v, want bounded active-profile facts", item)
	}
	summaryJSON, err := json.Marshal(agentResponse.ResultSummary)
	if err != nil {
		t.Fatalf("marshal wishlist result summary: %v", err)
	}
	for _, forbidden := range []string{"planner-wishlist-entry-b", "planner-wishlist-item-b", "private wishlist note", "other profile note", "https://private.example", "sk-planner-wishlist-secret", "preview_id"} {
		if strings.Contains(string(summaryJSON), forbidden) {
			t.Fatalf("wishlist result summary leaked %q: %s", forbidden, summaryJSON)
		}
	}
}

func TestChatAgentPlannerRoutesCollectionsSearchSummaryFromMainChat(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createA := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Collections A"}`), map[string]string{"Content-Type": "application/json"})
	createB := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Collections B"}`), map[string]string{"Content-Type": "application/json"})
	if createA.Code != 201 || createB.Code != 201 {
		t.Fatalf("create profiles status A=%d B=%d body A=%s body B=%s", createA.Code, createB.Code, createA.Body.String(), createB.Body.String())
	}
	var profileA, profileB struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createA.Body).Decode(&profileA); err != nil {
		t.Fatalf("decode profile A: %v", err)
	}
	if err := json.NewDecoder(createB.Body).Decode(&profileB); err != nil {
		t.Fatalf("decode profile B: %v", err)
	}
	workspaceA := `{"collections":["All Items","Planner Shelf A"],"activeCollection":"Planner Shelf A","items":[{"id":"planner-collection-item-a","name":"Planner Collection A Camaro","detail":"private collection detail must not render","collectionName":"Planner Shelf A"}]}`
	workspaceB := `{"collections":["All Items","Planner Shelf B"],"activeCollection":"Planner Shelf B","items":[{"id":"planner-collection-item-b","name":"Planner Collection B Porsche","detail":"other profile private detail","collectionName":"Planner Shelf B"}]}`
	if _, err := a.db.Exec(`
		INSERT INTO profile_settings(profile_id, key, value)
		VALUES
			(?, 'collections.workspace.v1', ?),
			(?, 'collections.workspace.v1', ?)
	`, profileA.ID, workspaceA, profileB.ID, workspaceB); err != nil {
		t.Fatalf("seed collections workspace state: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profileA.ID+`","title":"Planner Collections"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != 201 {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.collections.search","parameters":{"query":"planner","provider_secret":"sk-planner-collections-secret"},"message":"Searching collections."}`}),
		agentskills.NewRegistry(nil),
		profileA.ID,
		thread.ID,
		"Find my planner collections",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profileA.ID,
				"workspace_id":   "workspace-planner",
				"thread_id":      thread.ID,
				"route_id":       "/chats",
				"surface_id":     "chats.main",
				"source_channel": "in-app",
				"setup_state":    "ready",
				"intent_text":    "Find my planner collections",
			},
		},
		"message-collections-summary",
	)
	if !handled {
		t.Fatal("expected main Chat planner dispatch to handle Collections search request")
	}
	if result["skill_id"] != "cabinet.collections.search" || result["preview_result"] != nil {
		t.Fatalf("expected read-only Collections skill execution without preview, got %+v", result)
	}
	execution, ok := result["execution_result"].(map[string]any)
	if !ok || execution["read_only"] != true || execution["profile_id"] != profileA.ID || execution["mutation_applied"] == true {
		t.Fatalf("expected read-only profile-scoped Collections execution result, got %+v", result)
	}
	threadMessage, ok := result["thread_message"].(chat.Message)
	if !ok {
		t.Fatalf("collections planner result missing trusted assistant thread message: %+v", result)
	}
	agentResponseJSON, err := json.Marshal(threadMessage.Context["agent_response"])
	if err != nil {
		t.Fatalf("marshal collections Agent response: %v", err)
	}
	var agentResponse chat.AgentResponse
	if err := json.Unmarshal(agentResponseJSON, &agentResponse); err != nil {
		t.Fatalf("decode collections Agent response: %v", err)
	}
	if agentResponse.ResultSummary == nil || agentResponse.ResultSummary.Kind != "collections" {
		t.Fatalf("collections response missing typed server-owned summary: %+v", agentResponse)
	}
	if agentResponse.ResultSummary.Total != 2 || len(agentResponse.ResultSummary.Items) != 2 {
		t.Fatalf("collections summary must expose bounded active-profile records: %+v", agentResponse.ResultSummary)
	}
	summaryJSON, err := json.Marshal(agentResponse.ResultSummary)
	if err != nil {
		t.Fatalf("marshal collections result summary: %v", err)
	}
	for _, want := range []string{"Planner Shelf A", "Planner Collection A Camaro", "planner-collection-item-a", "assigned"} {
		if !strings.Contains(string(summaryJSON), want) {
			t.Fatalf("collections result summary missing %q: %s", want, summaryJSON)
		}
	}
	for _, forbidden := range []string{"Planner Shelf B", "Planner Collection B Porsche", "private collection detail", "other profile private detail", "sk-planner-collections-secret", "preview_id"} {
		if strings.Contains(string(summaryJSON), forbidden) {
			t.Fatalf("collections result summary leaked %q: %s", forbidden, summaryJSON)
		}
	}
}

func TestChatAgentPlannerRoutesIntegrationProviderSearchSummaryFromMainChat(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Integrations"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != 201 {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profile.ID+`","title":"Planner Integrations"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != 201 {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.integrations.search_providers","parameters":{"query":"provider","provider_secret":"sk-planner-integration-secret"},"message":"Searching integration providers."}`}),
		agentskills.NewRegistry(nil),
		profile.ID,
		thread.ID,
		"Find available integration providers",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profile.ID,
				"workspace_id":   "workspace-planner",
				"thread_id":      thread.ID,
				"route_id":       "/chats",
				"surface_id":     "chats.main",
				"source_channel": "in-app",
				"setup_state":    "ready",
				"intent_text":    "Find available integration providers",
			},
		},
		"message-integrations-summary",
	)
	if !handled {
		t.Fatal("expected main Chat planner dispatch to handle Integrations provider search request")
	}
	if result["skill_id"] != "cabinet.integrations.search_providers" || result["preview_result"] != nil {
		t.Fatalf("expected read-only Integrations skill execution without preview, got %+v", result)
	}
	execution, ok := result["execution_result"].(map[string]any)
	if !ok || execution["read_only"] != true || execution["mutation_applied"] == true {
		t.Fatalf("expected read-only Integrations execution result, got %+v", result)
	}
	body, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal integrations planner result: %v", err)
	}
	bodyText := string(body)
	for _, want := range []string{
		`"operation":"integrations.provider.search"`,
		`"id":"provider-registry"`,
		`"status":"available"`,
		`"setup_required":true`,
	} {
		if !strings.Contains(bodyText, want) {
			t.Fatalf("integrations planner response missing %s: body=%s", want, bodyText)
		}
	}
	if strings.Contains(bodyText, "sk-planner-integration-secret") || strings.Contains(bodyText, "preview_id") {
		t.Fatalf("integrations planner response leaked secret or preview token: body=%s", bodyText)
	}
	threadMessage, ok := result["thread_message"].(chat.Message)
	if !ok {
		t.Fatalf("integrations planner result missing trusted assistant thread message: %+v", result)
	}
	agentResponseJSON, err := json.Marshal(threadMessage.Context["agent_response"])
	if err != nil {
		t.Fatalf("marshal integrations Agent response: %v", err)
	}
	var agentResponse chat.AgentResponse
	if err := json.Unmarshal(agentResponseJSON, &agentResponse); err != nil {
		t.Fatalf("decode integrations Agent response: %v", err)
	}
	if agentResponse.ResultSummary == nil || agentResponse.ResultSummary.Kind != "integration_providers" {
		t.Fatalf("integrations response missing typed server-owned summary: %+v", agentResponse)
	}
	if agentResponse.ResultSummary.Total != 1 || len(agentResponse.ResultSummary.Items) != 1 {
		t.Fatalf("integrations summary must expose one bounded provider record: %+v", agentResponse.ResultSummary)
	}
	item := agentResponse.ResultSummary.Items[0]
	if item.ID != "provider-registry" || item.Title != "provider-registry" || item.Status != "available" || item.Category != "Setup required" {
		t.Fatalf("integrations summary item = %+v, want bounded provider facts", item)
	}
	summaryJSON, err := json.Marshal(agentResponse.ResultSummary)
	if err != nil {
		t.Fatalf("marshal integrations result summary: %v", err)
	}
	for _, forbidden := range []string{"sk-planner-integration-secret", "preview_id", "secret_redacted", "external_write_claimed", "raw_provider_payload"} {
		if strings.Contains(string(summaryJSON), forbidden) {
			t.Fatalf("integrations result summary leaked %q: %s", forbidden, summaryJSON)
		}
	}
}

func TestChatAgentPlannerRoutesIntegrationConnectionStatusSummaryFromMainChat(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Integration Connection"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != 201 {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO provider_health(provider, status, message, retry_after_seconds, updated_at) VALUES ('openai', 'auth_missing', 'missing credential: configure provider API key', 0, CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("seed provider health: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profile.ID+`","title":"Planner Integration Connection"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != 201 {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.integrations.test_connection","parameters":{"provider_id":"openai","api_key":"sk-planner-connection-secret","provider_health":{"message":"raw provider health must not render"}},"message":"Test OpenAI connection."}`}),
		agentskills.NewRegistry(nil),
		profile.ID,
		thread.ID,
		"Test the OpenAI provider connection",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profile.ID,
				"workspace_id":   "workspace-planner",
				"thread_id":      thread.ID,
				"route_id":       "/chats",
				"surface_id":     "chats.main",
				"source_channel": "in-app",
				"setup_state":    "ready",
				"intent_text":    "Test the OpenAI provider connection",
			},
		},
		"message-integration-connection-summary",
	)
	if !handled {
		t.Fatal("expected main Chat planner dispatch to handle Integrations connection status request")
	}
	if result["skill_id"] != "cabinet.integrations.test_connection" || result["preview_result"] != nil {
		t.Fatalf("expected non-mutating Integrations test connection without preview, got %+v", result)
	}
	execution, ok := result["execution_result"].(map[string]any)
	if !ok || execution["read_only"] != true || execution["mutation_applied"] == true {
		t.Fatalf("expected read-only Integrations connection execution result, got %+v", result)
	}
	body, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal integrations connection planner result: %v", err)
	}
	bodyText := string(body)
	for _, want := range []string{
		`"operation":"integrations.provider.test_connection"`,
		`"provider_id":"openai"`,
		`"connection_status":"needs_setup"`,
		`"next_action":"review_provider_status"`,
	} {
		if !strings.Contains(bodyText, want) {
			t.Fatalf("integrations connection planner response missing %s: body=%s", want, bodyText)
		}
	}
	if strings.Contains(bodyText, "sk-planner-connection-secret") || strings.Contains(bodyText, "preview_id") {
		t.Fatalf("integrations connection planner response leaked secret or preview token: body=%s", bodyText)
	}
	threadMessage, ok := result["thread_message"].(chat.Message)
	if !ok {
		t.Fatalf("integrations connection planner result missing trusted assistant thread message: %+v", result)
	}
	agentResponseJSON, err := json.Marshal(threadMessage.Context["agent_response"])
	if err != nil {
		t.Fatalf("marshal integrations connection Agent response: %v", err)
	}
	var agentResponse chat.AgentResponse
	if err := json.Unmarshal(agentResponseJSON, &agentResponse); err != nil {
		t.Fatalf("decode integrations connection Agent response: %v", err)
	}
	if agentResponse.ResultSummary == nil || agentResponse.ResultSummary.Kind != "integration_connection_status" {
		t.Fatalf("integrations connection response missing typed server-owned summary: %+v", agentResponse)
	}
	if agentResponse.ResultSummary.Total != 1 || len(agentResponse.ResultSummary.Items) != 1 {
		t.Fatalf("integrations connection summary must expose one bounded provider status record: %+v", agentResponse.ResultSummary)
	}
	item := agentResponse.ResultSummary.Items[0]
	if item.ID != "openai" || item.Title != "openai" || item.Status != "needs_setup" || item.Category != "review_provider_status" {
		t.Fatalf("integrations connection summary item = %+v, want bounded connection facts", item)
	}
	summaryJSON, err := json.Marshal(agentResponse.ResultSummary)
	if err != nil {
		t.Fatalf("marshal integrations connection result summary: %v", err)
	}
	for _, forbidden := range []string{"sk-planner-connection-secret", "missing credential", "raw provider health", "preview_id", "secret_redacted", "external_write_claimed", "provider_health", "raw_provider_payload"} {
		if strings.Contains(string(summaryJSON), forbidden) {
			t.Fatalf("integrations connection result summary leaked %q: %s", forbidden, summaryJSON)
		}
	}
}

func TestChatAgentPlannerRoutesIntegrationSetupExplanationSummaryFromMainChat(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Integration Setup"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != 201 {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profile.ID+`","title":"Planner Integration Setup"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != 201 {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.integrations.explain_required_setup","parameters":{"provider_id":"openai","api_key":"sk-planner-setup-secret","setup_payload":{"api_key":"sk-nested-secret"}},"message":"Explain provider setup."}`}),
		agentskills.NewRegistry(nil),
		profile.ID,
		thread.ID,
		"Explain what setup OpenAI needs",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profile.ID,
				"workspace_id":   "workspace-planner",
				"thread_id":      thread.ID,
				"route_id":       "/chats",
				"surface_id":     "chats.main",
				"source_channel": "in-app",
				"setup_state":    "ready",
				"intent_text":    "Explain what setup OpenAI needs",
			},
		},
		"message-integration-setup-summary",
	)
	if !handled {
		t.Fatal("expected main Chat planner dispatch to handle Integrations setup explanation request")
	}
	if result["skill_id"] != "cabinet.integrations.explain_required_setup" || result["preview_result"] != nil {
		t.Fatalf("expected read-only Integrations setup explanation without preview, got %+v", result)
	}
	execution, ok := result["execution_result"].(map[string]any)
	if !ok || execution["read_only"] != true || execution["mutation_applied"] == true {
		t.Fatalf("expected read-only Integrations setup execution result, got %+v", result)
	}
	body, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal integrations setup planner result: %v", err)
	}
	bodyText := string(body)
	for _, want := range []string{
		`"operation":"integrations.provider.explain_setup"`,
		`"provider_id":"openai"`,
		`"setup_required":true`,
		`"next_action":"Open Integrations settings for provider-specific credential and permission setup."`,
	} {
		if !strings.Contains(bodyText, want) {
			t.Fatalf("integrations setup planner response missing %s: body=%s", want, bodyText)
		}
	}
	if strings.Contains(bodyText, "sk-planner-setup-secret") || strings.Contains(bodyText, "sk-nested-secret") || strings.Contains(bodyText, "preview_id") {
		t.Fatalf("integrations setup planner response leaked secret or preview token: body=%s", bodyText)
	}
	threadMessage, ok := result["thread_message"].(chat.Message)
	if !ok {
		t.Fatalf("integrations setup planner result missing trusted assistant thread message: %+v", result)
	}
	agentResponseJSON, err := json.Marshal(threadMessage.Context["agent_response"])
	if err != nil {
		t.Fatalf("marshal integrations setup Agent response: %v", err)
	}
	var agentResponse chat.AgentResponse
	if err := json.Unmarshal(agentResponseJSON, &agentResponse); err != nil {
		t.Fatalf("decode integrations setup Agent response: %v", err)
	}
	if agentResponse.ResultSummary == nil || agentResponse.ResultSummary.Kind != "integration_setup_explanation" {
		t.Fatalf("integrations setup response missing typed server-owned summary: %+v", agentResponse)
	}
	if agentResponse.ResultSummary.Total != 1 || len(agentResponse.ResultSummary.Items) != 1 {
		t.Fatalf("integrations setup summary must expose one bounded provider setup record: %+v", agentResponse.ResultSummary)
	}
	item := agentResponse.ResultSummary.Items[0]
	if item.ID != "openai" || item.Title != "openai" || item.Status != "setup_required" || item.Category != "Open Integrations settings for provider-specific credential and permission setup." {
		t.Fatalf("integrations setup summary item = %+v, want bounded setup facts", item)
	}
	summaryJSON, err := json.Marshal(agentResponse.ResultSummary)
	if err != nil {
		t.Fatalf("marshal integrations setup result summary: %v", err)
	}
	for _, forbidden := range []string{"sk-planner-setup-secret", "sk-nested-secret", "preview_id", "secret_redacted", "external_write_claimed", "setup_payload", "raw_provider_payload"} {
		if strings.Contains(string(summaryJSON), forbidden) {
			t.Fatalf("integrations setup result summary leaked %q: %s", forbidden, summaryJSON)
		}
	}
}

func TestChatAgentPlannerRoutesMediaSearchSummaryFromMainChat(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createA := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Media A"}`), map[string]string{"Content-Type": "application/json"})
	if createA.Code != 201 {
		t.Fatalf("create profile A status=%d body=%s", createA.Code, createA.Body.String())
	}
	var profileA struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createA.Body).Decode(&profileA); err != nil {
		t.Fatalf("decode profile A: %v", err)
	}
	createB := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Media B"}`), map[string]string{"Content-Type": "application/json"})
	if createB.Code != 201 {
		t.Fatalf("create profile B status=%d body=%s", createB.Code, createB.Body.String())
	}
	var profileB struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createB.Body).Decode(&profileB); err != nil {
		t.Fatalf("decode profile B: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profileA.ID+`","title":"Planner Media"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != 201 {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}
	if _, err := a.db.Exec(`
		INSERT INTO chat_threads(id, profile_id, title)
		VALUES ('planner-media-thread-a', ?, 'Media A'), ('planner-media-thread-b', ?, 'Media B');
		INSERT INTO chat_attachments(id, profile_id, thread_id, filename, mime_type, size_bytes, stored_path)
		VALUES
			('planner-media-visible', ?, 'planner-media-thread-a', 'loose-reference.jpg', 'image/jpeg', 123, 'https://example.test/private/loose-reference.jpg'),
			('planner-media-hidden', ?, 'planner-media-thread-b', 'hidden-reference.jpg', 'image/jpeg', 456, 'https://example.test/private/hidden-reference.jpg');
	`, profileA.ID, profileB.ID, profileA.ID, profileB.ID); err != nil {
		t.Fatalf("seed media planner data: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.media.search","parameters":{"query":"loose-reference","provider_secret":"sk-planner-media-secret"},"message":"Searching media assets."}`}),
		agentskills.NewRegistry(nil),
		profileA.ID,
		thread.ID,
		"Find media asset loose-reference",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profileA.ID,
				"workspace_id":   "workspace-planner-media",
				"thread_id":      thread.ID,
				"route_id":       "/media",
				"surface_id":     "media.workspace",
				"source_channel": "in-app",
				"setup_state":    "ready",
			},
		},
		"message-media-summary",
	)
	if !handled {
		t.Fatal("expected main Chat planner dispatch to handle Media search request")
	}
	if result["skill_id"] != "cabinet.media.search" || result["preview_result"] != nil {
		t.Fatalf("expected read-only Media skill execution without preview, got %+v", result)
	}
	execution, ok := result["execution_result"].(map[string]any)
	if !ok || execution["read_only"] != true || execution["mutation_applied"] == true || execution["profile_id"] != profileA.ID {
		t.Fatalf("expected read-only Media execution result, got %+v", result)
	}
	body, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal media planner result: %v", err)
	}
	bodyText := string(body)
	if !strings.Contains(bodyText, `"operation":"media.search"`) || !strings.Contains(bodyText, `"filename":"loose-reference.jpg"`) {
		t.Fatalf("media planner response missing expected active-profile asset: body=%s", bodyText)
	}
	if strings.Contains(bodyText, "hidden-reference.jpg") || strings.Contains(bodyText, "sk-planner-media-secret") {
		t.Fatalf("media planner response leaked cross-profile record or secret: body=%s", bodyText)
	}
	threadMessage, ok := result["thread_message"].(chat.Message)
	if !ok {
		t.Fatalf("media planner result missing trusted assistant thread message: %+v", result)
	}
	agentResponseJSON, err := json.Marshal(threadMessage.Context["agent_response"])
	if err != nil {
		t.Fatalf("marshal media Agent response: %v", err)
	}
	var agentResponse chat.AgentResponse
	if err := json.Unmarshal(agentResponseJSON, &agentResponse); err != nil {
		t.Fatalf("decode media Agent response: %v", err)
	}
	if agentResponse.ResultSummary == nil || agentResponse.ResultSummary.Kind != "media_assets" {
		t.Fatalf("media response missing typed server-owned summary: %+v", agentResponse)
	}
	if agentResponse.ResultSummary.Total != 1 || len(agentResponse.ResultSummary.Items) != 1 {
		t.Fatalf("media summary must expose one bounded active-profile media asset: %+v", agentResponse.ResultSummary)
	}
	item := agentResponse.ResultSummary.Items[0]
	if item.ID != "planner-media-visible" || item.Title != "loose-reference.jpg" || item.Status != "unlinked" || item.Category != "Chat attachment" {
		t.Fatalf("media summary item = %+v, want bounded media id/title/linkage/source facts", item)
	}
	summaryJSON, err := json.Marshal(agentResponse.ResultSummary)
	if err != nil {
		t.Fatalf("marshal media result summary: %v", err)
	}
	for _, forbidden := range []string{"https://example.test/private", "hidden-reference.jpg", "sk-planner-media-secret", "provider_secret", "execution_result", "mutation_applied", "stored_path"} {
		if strings.Contains(string(summaryJSON), forbidden) {
			t.Fatalf("media read result summary leaked %q: %s", forbidden, summaryJSON)
		}
	}
}

func TestChatAgentPlannerRoutesUnlinkedMediaReviewSummaryFromMainChat(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createA := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Unlinked Media A"}`), map[string]string{"Content-Type": "application/json"})
	createB := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Unlinked Media B"}`), map[string]string{"Content-Type": "application/json"})
	if createA.Code != http.StatusCreated || createB.Code != http.StatusCreated {
		t.Fatalf("create profile statuses=%d/%d bodies=%s / %s", createA.Code, createB.Code, createA.Body.String(), createB.Body.String())
	}
	var profileA, profileB struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createA.Body).Decode(&profileA); err != nil {
		t.Fatalf("decode profile A: %v", err)
	}
	if err := json.NewDecoder(createB.Body).Decode(&profileB); err != nil {
		t.Fatalf("decode profile B: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profileA.ID+`","title":"Planner Unlinked Media"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}
	if _, err := a.db.Exec(`
		INSERT INTO chat_threads(id, profile_id, title)
		VALUES ('planner-unlinked-media-thread-a', ?, 'Unlinked Media A'), ('planner-unlinked-media-thread-b', ?, 'Unlinked Media B');
		INSERT INTO chat_attachments(id, profile_id, thread_id, filename, mime_type, size_bytes, stored_path)
		VALUES
			('planner-unlinked-media-visible', ?, 'planner-unlinked-media-thread-a', 'orphan-reference.jpg', 'image/jpeg', 321, 'https://example.test/private/orphan-reference.jpg'),
			('planner-unlinked-media-hidden', ?, 'planner-unlinked-media-thread-b', 'hidden-orphan.jpg', 'image/jpeg', 654, 'https://example.test/private/hidden-orphan.jpg');
	`, profileA.ID, profileB.ID, profileA.ID, profileB.ID); err != nil {
		t.Fatalf("seed unlinked media planner data: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.media.review_unlinked","parameters":{"query":"orphan-reference","provider_secret":"sk-planner-unlinked-media-secret"},"message":"Reviewing unlinked media."}`}),
		agentskills.NewRegistry(nil),
		profileA.ID,
		thread.ID,
		"Review unlinked media orphan-reference",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profileA.ID,
				"workspace_id":   "workspace-planner-unlinked-media",
				"thread_id":      thread.ID,
				"route_id":       "/media",
				"surface_id":     "media.workspace.unlinked",
				"source_channel": "in-app",
				"setup_state":    "ready",
			},
		},
		"message-unlinked-media-summary",
	)
	if !handled {
		t.Fatal("expected main Chat planner dispatch to handle unlinked Media review request")
	}
	if result["skill_id"] != "cabinet.media.review_unlinked" || result["preview_result"] != nil {
		t.Fatalf("expected read-only unlinked Media execution without preview, got %+v", result)
	}
	execution, ok := result["execution_result"].(map[string]any)
	if !ok || execution["read_only"] != true || execution["mutation_applied"] == true || execution["profile_id"] != profileA.ID {
		t.Fatalf("expected read-only unlinked Media execution result, got %+v", result)
	}
	body, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal unlinked media planner result: %v", err)
	}
	bodyText := string(body)
	if !strings.Contains(bodyText, `"operation":"media.review_unlinked"`) || !strings.Contains(bodyText, `"filename":"orphan-reference.jpg"`) {
		t.Fatalf("unlinked media planner response missing expected active-profile asset: body=%s", bodyText)
	}
	if strings.Contains(bodyText, "hidden-orphan.jpg") || strings.Contains(bodyText, "sk-planner-unlinked-media-secret") {
		t.Fatalf("unlinked media planner response leaked cross-profile record or secret: body=%s", bodyText)
	}
	threadMessage, ok := result["thread_message"].(chat.Message)
	if !ok {
		t.Fatalf("unlinked media planner result missing trusted assistant thread message: %+v", result)
	}
	agentResponseJSON, err := json.Marshal(threadMessage.Context["agent_response"])
	if err != nil {
		t.Fatalf("marshal unlinked media Agent response: %v", err)
	}
	var agentResponse chat.AgentResponse
	if err := json.Unmarshal(agentResponseJSON, &agentResponse); err != nil {
		t.Fatalf("decode unlinked media Agent response: %v", err)
	}
	if agentResponse.ResultSummary == nil || agentResponse.ResultSummary.Kind != "unlinked_media_assets" {
		t.Fatalf("unlinked media response missing typed server-owned summary: %+v", agentResponse)
	}
	if agentResponse.ResultSummary.Total != 1 || len(agentResponse.ResultSummary.Items) != 1 {
		t.Fatalf("unlinked media summary must expose one bounded active-profile media asset: %+v", agentResponse.ResultSummary)
	}
	item := agentResponse.ResultSummary.Items[0]
	if item.ID != "planner-unlinked-media-visible" || item.Title != "orphan-reference.jpg" || item.Status != "unlinked" || item.Category != "Chat attachment" {
		t.Fatalf("unlinked media summary item = %+v, want bounded media id/title/linkage/source facts", item)
	}
	summaryJSON, err := json.Marshal(agentResponse.ResultSummary)
	if err != nil {
		t.Fatalf("marshal unlinked media result summary: %v", err)
	}
	for _, forbidden := range []string{"https://example.test/private", "hidden-orphan.jpg", "sk-planner-unlinked-media-secret", "provider_secret", "execution_result", "mutation_applied", "stored_path"} {
		if strings.Contains(string(summaryJSON), forbidden) {
			t.Fatalf("unlinked media read result summary leaked %q: %s", forbidden, summaryJSON)
		}
	}
}

func TestChatAgentPlannerRoutesDiscoveriesSearchSummaryFromMainChat(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createA := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Discoveries A"}`), map[string]string{"Content-Type": "application/json"})
	createB := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Discoveries B"}`), map[string]string{"Content-Type": "application/json"})
	if createA.Code != http.StatusCreated || createB.Code != http.StatusCreated {
		t.Fatalf("create profile statuses=%d/%d bodies=%s / %s", createA.Code, createB.Code, createA.Body.String(), createB.Body.String())
	}
	var profileA, profileB struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createA.Body).Decode(&profileA); err != nil {
		t.Fatalf("decode profile A: %v", err)
	}
	if err := json.NewDecoder(createB.Body).Decode(&profileB); err != nil {
		t.Fatalf("decode profile B: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profileA.ID+`","title":"Planner Discoveries"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}
	if _, err := a.db.Exec(`
		INSERT INTO scanner_query_sets(id, profile_id, name, keywords_json, provider_scope_json, enabled)
		VALUES
			('planner-discovery-visible-watch', ?, 'Visible discovery watch', '["visible"]', '["ebay"]', 1),
			('planner-discovery-hidden-watch', ?, 'Hidden discovery watch', '["visible"]', '["ebay"]', 1);
		INSERT INTO scanner_candidates(id, profile_id, query_set_id, listing_id, title, price, observed_currency, shipping, url, image, seller, first_seen, last_seen, status, source, stock_state, stock_count, reviewer_notes, source_result_url)
		VALUES
			('planner-discovery-visible', ?, 'planner-discovery-visible-watch', 'DISC-VISIBLE-1', 'Visible planner discovery', 21, 'AUD', 3, 'https://example.test/private-visible-listing', 'https://example.test/private-visible-image.jpg', 'private-visible-seller', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'in_stock', 2, 'private reviewer note', 'https://example.test/private-source-result'),
			('planner-discovery-hidden', ?, 'planner-discovery-hidden-watch', 'DISC-HIDDEN-1', 'Hidden planner discovery', 99, 'AUD', 4, 'https://example.test/private-hidden-listing', '', 'private-hidden-seller', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'in_stock', 1, 'hidden reviewer note', 'https://example.test/private-hidden-source');
		INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at)
		VALUES
			('planner-discovery-visible', '', 'not_in_collection', 0.88, 1, 'PRIVATE-PART-VISIBLE', CURRENT_TIMESTAMP),
			('planner-discovery-hidden', '', 'not_in_collection', 0.77, 1, 'PRIVATE-PART-HIDDEN', CURRENT_TIMESTAMP);
	`, profileA.ID, profileB.ID, profileA.ID, profileB.ID); err != nil {
		t.Fatalf("seed discovery planner data: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.discoveries.search","parameters":{"provider_id":"ebay","query":"Visible planner","provider_secret":"sk-planner-discovery-secret"},"message":"Searching discovery results."}`}),
		agentskills.NewRegistry(nil),
		profileA.ID,
		thread.ID,
		"Find discovery result Visible planner",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profileA.ID,
				"workspace_id":   "workspace-planner-discoveries",
				"thread_id":      thread.ID,
				"route_id":       "/discovery",
				"surface_id":     "discoveries.workspace",
				"source_channel": "in-app",
				"setup_state":    "ready",
			},
		},
		"message-discovery-summary",
	)
	if !handled {
		t.Fatal("expected main Chat planner dispatch to handle Discoveries search request")
	}
	if result["skill_id"] != "cabinet.discoveries.search" || result["preview_result"] != nil {
		t.Fatalf("expected read-only Discoveries skill execution without preview, got %+v", result)
	}
	execution, ok := result["execution_result"].(map[string]any)
	if !ok || execution["read_only"] != true || execution["mutation_applied"] == true || execution["profile_id"] != profileA.ID {
		t.Fatalf("expected read-only Discoveries execution result, got %+v", result)
	}
	body, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal discoveries planner result: %v", err)
	}
	bodyText := string(body)
	if !strings.Contains(bodyText, `"operation":"discoveries.search"`) || !strings.Contains(bodyText, `"candidate_id":"planner-discovery-visible"`) {
		t.Fatalf("discoveries planner response missing expected active-profile discovery result: body=%s", bodyText)
	}
	if strings.Contains(bodyText, "planner-discovery-hidden") || strings.Contains(bodyText, "sk-planner-discovery-secret") {
		t.Fatalf("discoveries planner response leaked cross-profile record or secret: body=%s", bodyText)
	}
	threadMessage, ok := result["thread_message"].(chat.Message)
	if !ok {
		t.Fatalf("discoveries planner result missing trusted assistant thread message: %+v", result)
	}
	agentResponseJSON, err := json.Marshal(threadMessage.Context["agent_response"])
	if err != nil {
		t.Fatalf("marshal discoveries Agent response: %v", err)
	}
	var agentResponse chat.AgentResponse
	if err := json.Unmarshal(agentResponseJSON, &agentResponse); err != nil {
		t.Fatalf("decode discoveries Agent response: %v", err)
	}
	if agentResponse.ResultSummary == nil || agentResponse.ResultSummary.Kind != "discovery_results" {
		t.Fatalf("discoveries response missing typed server-owned summary: %+v", agentResponse)
	}
	if agentResponse.ResultSummary.Total != 1 || len(agentResponse.ResultSummary.Items) != 1 {
		t.Fatalf("discoveries summary must expose one bounded active-profile discovery result: %+v", agentResponse.ResultSummary)
	}
	item := agentResponse.ResultSummary.Items[0]
	if item.ID != "planner-discovery-visible" || item.Title != "Visible planner discovery" || item.Status != "new" || item.Category != "ebay / needs review" {
		t.Fatalf("discoveries summary item = %+v, want bounded candidate id/title/status/provider facts", item)
	}
	summaryJSON, err := json.Marshal(agentResponse.ResultSummary)
	if err != nil {
		t.Fatalf("marshal discoveries result summary: %v", err)
	}
	for _, forbidden := range []string{"https://example.test/private", "private-visible-seller", "private reviewer note", "PRIVATE-PART", "planner-discovery-hidden", "sk-planner-discovery-secret", "provider_secret", "execution_result", "mutation_applied", "source_result_url", "price", "shipping", "confidence"} {
		if strings.Contains(string(summaryJSON), forbidden) {
			t.Fatalf("discoveries read result summary leaked %q: %s", forbidden, summaryJSON)
		}
	}
}

func TestChatAgentPlannerRoutesDiscoveriesReviewResultSummaryFromMainChat(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createA := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Discovery Review A"}`), map[string]string{"Content-Type": "application/json"})
	createB := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Discovery Review B"}`), map[string]string{"Content-Type": "application/json"})
	if createA.Code != http.StatusCreated || createB.Code != http.StatusCreated {
		t.Fatalf("create profile statuses=%d/%d bodies=%s / %s", createA.Code, createB.Code, createA.Body.String(), createB.Body.String())
	}
	var profileA, profileB struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createA.Body).Decode(&profileA); err != nil {
		t.Fatalf("decode profile A: %v", err)
	}
	if err := json.NewDecoder(createB.Body).Decode(&profileB); err != nil {
		t.Fatalf("decode profile B: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profileA.ID+`","title":"Planner Discovery Review"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}
	if _, err := a.db.Exec(`
		INSERT INTO scanner_query_sets(id, profile_id, name, keywords_json, provider_scope_json, enabled)
		VALUES
			('planner-discovery-review-visible-watch', ?, 'Visible discovery review watch', '["visible"]', '["ebay"]', 1),
			('planner-discovery-review-hidden-watch', ?, 'Hidden discovery review watch', '["visible"]', '["ebay"]', 1);
		INSERT INTO scanner_candidates(id, profile_id, query_set_id, listing_id, title, price, observed_currency, shipping, url, image, seller, first_seen, last_seen, status, source, stock_state, stock_count, reviewer_notes, source_result_url)
		VALUES
			('planner-discovery-review-visible', ?, 'planner-discovery-review-visible-watch', 'DISC-REVIEW-VISIBLE-1', 'Visible planner discovery review', 31, 'AUD', 5, 'https://example.test/private-review-listing', 'https://example.test/private-review-image.jpg', 'private-review-seller', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'in_stock', 2, 'private review note', 'https://example.test/private-review-source'),
			('planner-discovery-review-hidden', ?, 'planner-discovery-review-hidden-watch', 'DISC-REVIEW-HIDDEN-1', 'Hidden planner discovery review', 88, 'AUD', 6, 'https://example.test/private-hidden-review-listing', '', 'private-hidden-review-seller', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'in_stock', 1, 'hidden review note', 'https://example.test/private-hidden-review-source');
		INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at)
		VALUES
			('planner-discovery-review-visible', '', 'not_in_collection', 0.91, 1, 'PRIVATE-REVIEW-PART', CURRENT_TIMESTAMP),
			('planner-discovery-review-hidden', '', 'not_in_collection', 0.66, 1, 'PRIVATE-HIDDEN-REVIEW-PART', CURRENT_TIMESTAMP);
	`, profileA.ID, profileB.ID, profileA.ID, profileB.ID); err != nil {
		t.Fatalf("seed discovery review planner data: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.discoveries.review_result","parameters":{"provider_id":"ebay","result_id":"planner-discovery-review-visible","provider_secret":"sk-planner-discovery-review-secret","source_result_url":"https://example.test/private-review-source"},"message":"Reviewing discovery result."}`}),
		agentskills.NewRegistry(nil),
		profileA.ID,
		thread.ID,
		"Review discovery result planner-discovery-review-visible",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profileA.ID,
				"workspace_id":   "workspace-planner-discovery-review",
				"thread_id":      thread.ID,
				"route_id":       "/discovery",
				"surface_id":     "discoveries.workspace",
				"source_channel": "in-app",
				"setup_state":    "ready",
			},
		},
		"message-discovery-review-summary",
	)
	if !handled {
		t.Fatal("expected main Chat planner dispatch to handle Discoveries review request")
	}
	if result["skill_id"] != "cabinet.discoveries.review_result" || result["preview_result"] != nil {
		t.Fatalf("expected read-only Discoveries review skill execution without preview, got %+v", result)
	}
	execution, ok := result["execution_result"].(map[string]any)
	if !ok || execution["operation"] != "discoveries.review_result" || execution["read_only"] != true || execution["mutation_applied"] == true || execution["profile_id"] != profileA.ID {
		t.Fatalf("expected read-only Discoveries review execution result, got %+v", result)
	}
	body, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal discovery review planner result: %v", err)
	}
	bodyText := string(body)
	if !strings.Contains(bodyText, `"candidate_id":"planner-discovery-review-visible"`) || !strings.Contains(bodyText, `"provenance_preserved":true`) {
		t.Fatalf("discovery review planner response missing expected active-profile review result: body=%s", bodyText)
	}
	if strings.Contains(bodyText, "planner-discovery-review-hidden") || strings.Contains(bodyText, "sk-planner-discovery-review-secret") {
		t.Fatalf("discovery review planner response leaked cross-profile record or secret: body=%s", bodyText)
	}
	threadMessage, ok := result["thread_message"].(chat.Message)
	if !ok {
		t.Fatalf("discovery review planner result missing trusted assistant thread message: %+v", result)
	}
	agentResponseJSON, err := json.Marshal(threadMessage.Context["agent_response"])
	if err != nil {
		t.Fatalf("marshal discovery review Agent response: %v", err)
	}
	var agentResponse chat.AgentResponse
	if err := json.Unmarshal(agentResponseJSON, &agentResponse); err != nil {
		t.Fatalf("decode discovery review Agent response: %v", err)
	}
	if agentResponse.ResultSummary == nil || agentResponse.ResultSummary.Kind != "discovery_results" {
		t.Fatalf("discovery review response missing typed server-owned summary: %+v", agentResponse)
	}
	if agentResponse.ResultSummary.Total != 1 || len(agentResponse.ResultSummary.Items) != 1 {
		t.Fatalf("discovery review summary must expose one bounded active-profile result: %+v", agentResponse.ResultSummary)
	}
	item := agentResponse.ResultSummary.Items[0]
	if item.ID != "planner-discovery-review-visible" || item.Title != "Visible planner discovery review" || item.Status != "new" || item.Category != "ebay / needs review" {
		t.Fatalf("discovery review summary item = %+v, want bounded candidate id/title/status/provider facts", item)
	}
	summaryJSON, err := json.Marshal(agentResponse.ResultSummary)
	if err != nil {
		t.Fatalf("marshal discovery review result summary: %v", err)
	}
	for _, forbidden := range []string{"https://example.test/private", "private-review-seller", "private review note", "PRIVATE-REVIEW-PART", "planner-discovery-review-hidden", "sk-planner-discovery-review-secret", "provider_secret", "execution_result", "mutation_applied", "source_result_url", "price", "shipping", "confidence"} {
		if strings.Contains(string(summaryJSON), forbidden) {
			t.Fatalf("discovery review read result summary leaked %q: %s", forbidden, summaryJSON)
		}
	}
}

func TestChatAgentPlannerRoutesPurchaseOrdersSummaryFromMainChat(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createA := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Purchases A"}`), map[string]string{"Content-Type": "application/json"})
	createB := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Purchases B"}`), map[string]string{"Content-Type": "application/json"})
	if createA.Code != http.StatusCreated || createB.Code != http.StatusCreated {
		t.Fatalf("create profile statuses=%d/%d bodies=%s / %s", createA.Code, createB.Code, createA.Body.String(), createB.Body.String())
	}
	var profileA, profileB struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createA.Body).Decode(&profileA); err != nil {
		t.Fatalf("decode profile A: %v", err)
	}
	if err := json.NewDecoder(createB.Body).Decode(&profileB); err != nil {
		t.Fatalf("decode profile B: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profileA.ID+`","title":"Planner Purchases"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}
	if _, err := a.db.Exec(`
		INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, status)
		VALUES
			('planner-purchase-visible-item', ?, 'AFX', 'Slot Cars', 'PO-VISIBLE-1', 'Visible purchase line', 'active'),
			('planner-purchase-hidden-item', ?, 'AFX', 'Slot Cars', 'PO-HIDDEN-1', 'Hidden purchase line', 'active');
		INSERT INTO commerce_lifecycle_entries(id, profile_id, item_id, state, source, external_ref, quantity, amount, currency, notes)
		VALUES
			('planner-purchase-visible-entry', ?, 'planner-purchase-visible-item', 'purchase', 'market_watch', 'order-visible-1', 2, 44.50, 'AUD', 'seller=private-seller tracking=TRACK-PRIVATE notes=keep-private'),
			('planner-purchase-hidden-entry', ?, 'planner-purchase-hidden-item', 'purchase', 'market_watch', 'order-hidden-1', 1, 12.00, 'AUD', 'seller=hidden-seller tracking=TRACK-HIDDEN');
		INSERT INTO expected_arrivals(id, profile_id, item_id, lifecycle_entry_id, source, external_ref, quantity, amount, currency, status, notes)
		VALUES
			('planner-purchase-visible-arrival', ?, 'planner-purchase-visible-item', 'planner-purchase-visible-entry', 'market_watch', 'order-visible-1', 2, 44.50, 'AUD', 'expected', 'private arrival note'),
			('planner-purchase-hidden-arrival', ?, 'planner-purchase-hidden-item', 'planner-purchase-hidden-entry', 'market_watch', 'order-hidden-1', 1, 12.00, 'AUD', 'expected', 'hidden arrival note');
	`, profileA.ID, profileB.ID, profileA.ID, profileB.ID, profileA.ID, profileB.ID); err != nil {
		t.Fatalf("seed purchase planner data: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.purchases.search_orders","parameters":{"query":"order-visible","status":"all","provider_secret":"sk-planner-purchase-secret"},"message":"Searching purchase orders."}`}),
		agentskills.NewRegistry(nil),
		profileA.ID,
		thread.ID,
		"Find purchase order order-visible",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profileA.ID,
				"workspace_id":   "workspace-planner-purchases",
				"thread_id":      thread.ID,
				"route_id":       "/purchases",
				"surface_id":     "purchases.workspace",
				"source_channel": "in-app",
				"setup_state":    "ready",
			},
		},
		"message-purchase-summary",
	)
	if !handled {
		t.Fatal("expected main Chat planner dispatch to handle Purchase order search request")
	}
	if result["skill_id"] != "cabinet.purchases.search_orders" || result["preview_result"] != nil {
		t.Fatalf("expected read-only Purchase order skill execution without preview, got %+v", result)
	}
	execution, ok := result["execution_result"].(map[string]any)
	if !ok || execution["read_only"] != true || execution["mutation_applied"] == true || execution["profile_id"] != profileA.ID {
		t.Fatalf("expected read-only Purchase order execution result, got %+v", result)
	}
	body, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal purchase planner result: %v", err)
	}
	bodyText := string(body)
	if !strings.Contains(bodyText, `"operation":"purchases.orders.search"`) || !strings.Contains(bodyText, `"order_id":"order-visible-1"`) {
		t.Fatalf("purchase planner response missing expected active-profile order: body=%s", bodyText)
	}
	if strings.Contains(bodyText, "order-hidden-1") || strings.Contains(bodyText, "sk-planner-purchase-secret") {
		t.Fatalf("purchase planner response leaked cross-profile record or secret: body=%s", bodyText)
	}
	threadMessage, ok := result["thread_message"].(chat.Message)
	if !ok {
		t.Fatalf("purchase planner result missing trusted assistant thread message: %+v", result)
	}
	agentResponseJSON, err := json.Marshal(threadMessage.Context["agent_response"])
	if err != nil {
		t.Fatalf("marshal purchase Agent response: %v", err)
	}
	var agentResponse chat.AgentResponse
	if err := json.Unmarshal(agentResponseJSON, &agentResponse); err != nil {
		t.Fatalf("decode purchase Agent response: %v", err)
	}
	if agentResponse.ResultSummary == nil || agentResponse.ResultSummary.Kind != "purchase_orders" {
		t.Fatalf("purchase response missing typed server-owned summary: %+v", agentResponse)
	}
	if agentResponse.ResultSummary.Total != 1 || len(agentResponse.ResultSummary.Items) != 1 {
		t.Fatalf("purchase summary must expose one bounded active-profile purchase order: %+v", agentResponse.ResultSummary)
	}
	item := agentResponse.ResultSummary.Items[0]
	if item.ID != "order-visible-1" || item.Title != "order-visible-1" || item.Status != "active" || item.Category != "market_watch / 1 line items" {
		t.Fatalf("purchase summary item = %+v, want bounded order id/status/source/line-count facts", item)
	}
	summaryJSON, err := json.Marshal(agentResponse.ResultSummary)
	if err != nil {
		t.Fatalf("marshal purchase result summary: %v", err)
	}
	for _, forbidden := range []string{"private-seller", "TRACK-PRIVATE", "hidden-seller", "order-hidden-1", "sk-planner-purchase-secret", "provider_secret", "execution_result", "mutation_applied", "expected_arrival_id", "amount", "currency"} {
		if strings.Contains(string(summaryJSON), forbidden) {
			t.Fatalf("purchase read result summary leaked %q: %s", forbidden, summaryJSON)
		}
	}
}

func TestChatAgentPlannerRoutesPurchaseReviewSummaryFromMainChat(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Purchase Review"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profile.ID+`","title":"Planner Purchase Review"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.purchases.review_purchase","parameters":{"order_id":"purchase-review-visible-order","review_status":"needs_attention","provider_secret":"sk-planner-purchase-review-secret","amount":"99.95","tracking":"TRACK-PRIVATE"},"message":"Reviewing purchase evidence."}`}),
		agentskills.NewRegistry(nil),
		profile.ID,
		thread.ID,
		"Review purchase order purchase-review-visible-order",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profile.ID,
				"workspace_id":   "workspace-planner-purchase-review",
				"thread_id":      thread.ID,
				"route_id":       "/purchases",
				"surface_id":     "purchases.workspace",
				"source_channel": "in-app",
				"setup_state":    "ready",
			},
		},
		"message-purchase-review-summary",
	)
	if !handled {
		t.Fatal("expected main Chat planner dispatch to handle Purchase review request")
	}
	if result["skill_id"] != "cabinet.purchases.review_purchase" || result["preview_result"] != nil {
		t.Fatalf("expected read-only Purchase review skill execution without preview, got %+v", result)
	}
	execution, ok := result["execution_result"].(map[string]any)
	if !ok || execution["operation"] != "purchases.order.review" || execution["read_only"] != true || execution["mutation_applied"] == true || execution["profile_id"] != profile.ID {
		t.Fatalf("expected read-only Purchase review execution result, got %+v", result)
	}
	threadMessage, ok := result["thread_message"].(chat.Message)
	if !ok {
		t.Fatalf("purchase review planner result missing trusted assistant thread message: %+v", result)
	}
	agentResponseJSON, err := json.Marshal(threadMessage.Context["agent_response"])
	if err != nil {
		t.Fatalf("marshal purchase review Agent response: %v", err)
	}
	var agentResponse chat.AgentResponse
	if err := json.Unmarshal(agentResponseJSON, &agentResponse); err != nil {
		t.Fatalf("decode purchase review Agent response: %v", err)
	}
	if agentResponse.ResultSummary == nil || agentResponse.ResultSummary.Kind != "purchase_review" {
		t.Fatalf("purchase review response missing typed server-owned summary: %+v", agentResponse)
	}
	if agentResponse.ResultSummary.Total != 1 || len(agentResponse.ResultSummary.Items) != 1 {
		t.Fatalf("purchase review summary must expose one bounded active-profile review: %+v", agentResponse.ResultSummary)
	}
	item := agentResponse.ResultSummary.Items[0]
	if item.ID != "purchase-review-visible-order" || item.Title != "purchase-review-visible-order" || item.Status != "needs_attention" || item.Category != "Purchase review" {
		t.Fatalf("purchase review summary item = %+v, want bounded order id/review status facts", item)
	}
	summaryJSON, err := json.Marshal(agentResponse.ResultSummary)
	if err != nil {
		t.Fatalf("marshal purchase review result summary: %v", err)
	}
	for _, forbidden := range []string{"sk-planner-purchase-review-secret", "provider_secret", "execution_result", "mutation_applied", "amount", "99.95", "tracking", "TRACK-PRIVATE"} {
		if strings.Contains(string(summaryJSON), forbidden) {
			t.Fatalf("purchase review read result summary leaked %q: %s", forbidden, summaryJSON)
		}
	}
}

func TestChatAgentPlannerRoutesMarketWatchSearchSummaryFromMainChat(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createA := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Market Watch A"}`), map[string]string{"Content-Type": "application/json"})
	createB := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Market Watch B"}`), map[string]string{"Content-Type": "application/json"})
	if createA.Code != http.StatusCreated || createB.Code != http.StatusCreated {
		t.Fatalf("create profile statuses=%d/%d bodies=%s / %s", createA.Code, createB.Code, createA.Body.String(), createB.Body.String())
	}
	var profileA, profileB struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createA.Body).Decode(&profileA); err != nil {
		t.Fatalf("decode profile A: %v", err)
	}
	if err := json.NewDecoder(createB.Body).Decode(&profileB); err != nil {
		t.Fatalf("decode profile B: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profileA.ID+`","title":"Planner Market Watch"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}
	scannerSvc := scanner.NewService(a.db)
	visible, err := scannerSvc.CreateQuerySetForProfile(context.Background(), profileA.ID, scanner.QuerySet{
		Name:               "Planner A Slot Search",
		Keywords:           []string{"AFX", "secret-keyword"},
		Exclusions:         []string{"private-exclusion"},
		ProviderScope:      []string{"ebay"},
		Enabled:            true,
		LastCandidateCount: 3,
	})
	if err != nil {
		t.Fatalf("seed visible market watch: %v", err)
	}
	if _, err := scannerSvc.CreateQuerySetForProfile(context.Background(), profileB.ID, scanner.QuerySet{
		Name:          "Planner B Hidden Watch",
		Keywords:      []string{"AFX"},
		ProviderScope: []string{"ebay"},
		Enabled:       true,
	}); err != nil {
		t.Fatalf("seed hidden market watch: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.market_watch.search_watches","parameters":{"provider_id":"ebay","query":"Planner A","provider_secret":"sk-planner-market-watch-secret"},"message":"Searching saved watches."}`}),
		agentskills.NewRegistry(nil),
		profileA.ID,
		thread.ID,
		"Find Market Watch saved watches for Planner A on ebay",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profileA.ID,
				"workspace_id":   "workspace-planner-market-watch",
				"thread_id":      thread.ID,
				"route_id":       "/scanner",
				"surface_id":     "market-watch.workspace",
				"source_channel": "in-app",
				"setup_state":    "ready",
			},
		},
		"message-market-watch-summary",
	)
	if !handled {
		t.Fatal("expected main Chat planner dispatch to handle Market Watch search request")
	}
	if result["skill_id"] != "cabinet.market_watch.search_watches" || result["preview_result"] != nil {
		t.Fatalf("expected read-only Market Watch skill execution without preview, got %+v", result)
	}
	execution, ok := result["execution_result"].(map[string]any)
	if !ok || execution["read_only"] != true || execution["mutation_applied"] == true {
		t.Fatalf("expected read-only Market Watch execution result, got %+v", result)
	}
	body, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal market watch planner result: %v", err)
	}
	bodyText := string(body)
	if !strings.Contains(bodyText, `"operation":"market_watch.watch.search"`) || !strings.Contains(bodyText, visible.ID) {
		t.Fatalf("market watch planner response missing expected active-profile watch: body=%s", bodyText)
	}
	if strings.Contains(bodyText, "Planner B Hidden Watch") || strings.Contains(bodyText, "sk-planner-market-watch-secret") {
		t.Fatalf("market watch planner response leaked cross-profile record or secret: body=%s", bodyText)
	}
	threadMessage, ok := result["thread_message"].(chat.Message)
	if !ok {
		t.Fatalf("market watch planner result missing trusted assistant thread message: %+v", result)
	}
	agentResponseJSON, err := json.Marshal(threadMessage.Context["agent_response"])
	if err != nil {
		t.Fatalf("marshal market watch Agent response: %v", err)
	}
	var agentResponse chat.AgentResponse
	if err := json.Unmarshal(agentResponseJSON, &agentResponse); err != nil {
		t.Fatalf("decode market watch Agent response: %v", err)
	}
	if agentResponse.ResultSummary == nil || agentResponse.ResultSummary.Kind != "market_watch_watches" {
		t.Fatalf("market watch response missing typed server-owned summary: %+v", agentResponse)
	}
	if agentResponse.ResultSummary.Total != 1 || len(agentResponse.ResultSummary.Items) != 1 {
		t.Fatalf("market watch summary must expose one bounded active-profile watch: %+v", agentResponse.ResultSummary)
	}
	item := agentResponse.ResultSummary.Items[0]
	if item.ID != visible.ID || item.Title != "Planner A Slot Search" || item.Status != "enabled" || item.Category != "Market Watch / 1 providers" {
		t.Fatalf("market watch summary item = %+v, want bounded saved-watch id/name/status/provider-count facts", item)
	}
	summaryJSON, err := json.Marshal(agentResponse.ResultSummary)
	if err != nil {
		t.Fatalf("marshal market watch result summary: %v", err)
	}
	for _, forbidden := range []string{"secret-keyword", "private-exclusion", "Planner B Hidden Watch", "sk-planner-market-watch-secret", "provider_secret", "execution_result", "mutation_applied", "provider_health", "external_write_claimed"} {
		if strings.Contains(string(summaryJSON), forbidden) {
			t.Fatalf("market watch read result summary leaked %q: %s", forbidden, summaryJSON)
		}
	}
}

func TestChatAgentPlannerRoutesMarketWatchResultsSummaryFromMainChat(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createA := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Market Watch Results A"}`), map[string]string{"Content-Type": "application/json"})
	createB := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Market Watch Results B"}`), map[string]string{"Content-Type": "application/json"})
	if createA.Code != http.StatusCreated || createB.Code != http.StatusCreated {
		t.Fatalf("create profile statuses=%d/%d bodies=%s / %s", createA.Code, createB.Code, createA.Body.String(), createB.Body.String())
	}
	var profileA, profileB struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createA.Body).Decode(&profileA); err != nil {
		t.Fatalf("decode profile A: %v", err)
	}
	if err := json.NewDecoder(createB.Body).Decode(&profileB); err != nil {
		t.Fatalf("decode profile B: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profileA.ID+`","title":"Planner Market Watch Results"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}
	if _, err := a.db.Exec(`
		INSERT INTO scanner_query_sets(id, profile_id, name, keywords_json, provider_scope_json, enabled)
		VALUES
			('planner-market-results-visible-watch', ?, 'Visible Market results watch', '["visible"]', '["ebay"]', 1),
			('planner-market-results-hidden-watch', ?, 'Hidden Market results watch', '["visible"]', '["ebay"]', 1);
		INSERT INTO scanner_candidates(id, profile_id, query_set_id, listing_id, title, price, observed_currency, shipping, url, image, seller, first_seen, last_seen, status, source, stock_state, stock_count, reviewer_notes, source_result_url)
		VALUES
			('planner-market-result-visible', ?, 'planner-market-results-visible-watch', 'MW-VISIBLE-1', 'Visible planner market result', 55, 'AUD', 9, 'https://example.test/private-visible-market', 'https://example.test/private-visible-market.jpg', 'private-market-seller', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'in_stock', 2, 'private market note', 'https://example.test/private-source-market'),
			('planner-market-result-hidden', ?, 'planner-market-results-hidden-watch', 'MW-HIDDEN-1', 'Hidden planner market result', 77, 'AUD', 4, 'https://example.test/private-hidden-market', '', 'private-hidden-market-seller', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'in_stock', 1, 'hidden market note', 'https://example.test/private-hidden-source-market');
	`, profileA.ID, profileB.ID, profileA.ID, profileB.ID); err != nil {
		t.Fatalf("seed market watch result planner data: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.market_watch.review_results","parameters":{"provider_id":"ebay","watch_id":"planner-market-results-visible-watch","provider_secret":"sk-planner-market-result-secret"},"message":"Reviewing Market Watch results."}`}),
		agentskills.NewRegistry(nil),
		profileA.ID,
		thread.ID,
		"Review Market Watch results for planner-market-results-visible-watch on ebay",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profileA.ID,
				"workspace_id":   "workspace-planner-market-watch-results",
				"thread_id":      thread.ID,
				"route_id":       "/scanner",
				"surface_id":     "market-watch.workspace",
				"source_channel": "in-app",
				"setup_state":    "ready",
			},
		},
		"message-market-watch-results-summary",
	)
	if !handled {
		t.Fatal("expected main Chat planner dispatch to handle Market Watch results request")
	}
	if result["skill_id"] != "cabinet.market_watch.review_results" || result["preview_result"] != nil {
		t.Fatalf("expected read-only Market Watch results skill execution without preview, got %+v", result)
	}
	execution, ok := result["execution_result"].(map[string]any)
	if !ok || execution["read_only"] != true || execution["mutation_applied"] == true {
		t.Fatalf("expected read-only Market Watch results execution result, got %+v", result)
	}
	body, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal market watch results planner result: %v", err)
	}
	bodyText := string(body)
	if !strings.Contains(bodyText, `"operation":"market_watch.results.review"`) || !strings.Contains(bodyText, `"id":"planner-market-result-visible"`) {
		t.Fatalf("market watch results planner response missing expected active-profile result: body=%s", bodyText)
	}
	if strings.Contains(bodyText, "planner-market-result-hidden") || strings.Contains(bodyText, "sk-planner-market-result-secret") {
		t.Fatalf("market watch results planner response leaked cross-profile record or secret: body=%s", bodyText)
	}
	threadMessage, ok := result["thread_message"].(chat.Message)
	if !ok {
		t.Fatalf("market watch results planner result missing trusted assistant thread message: %+v", result)
	}
	agentResponseJSON, err := json.Marshal(threadMessage.Context["agent_response"])
	if err != nil {
		t.Fatalf("marshal market watch results Agent response: %v", err)
	}
	var agentResponse chat.AgentResponse
	if err := json.Unmarshal(agentResponseJSON, &agentResponse); err != nil {
		t.Fatalf("decode market watch results Agent response: %v", err)
	}
	if agentResponse.ResultSummary == nil || agentResponse.ResultSummary.Kind != "market_watch_results" {
		t.Fatalf("market watch results response missing typed server-owned summary: %+v", agentResponse)
	}
	if agentResponse.ResultSummary.Total != 1 || len(agentResponse.ResultSummary.Items) != 1 {
		t.Fatalf("market watch results summary must expose one bounded active-profile result: %+v", agentResponse.ResultSummary)
	}
	item := agentResponse.ResultSummary.Items[0]
	if item.ID != "planner-market-result-visible" || item.Title != "Visible planner market result" || item.Status != "new" || item.Category != "ebay / in_stock" {
		t.Fatalf("market watch results summary item = %+v, want bounded result id/title/status/provider facts", item)
	}
	summaryJSON, err := json.Marshal(agentResponse.ResultSummary)
	if err != nil {
		t.Fatalf("marshal market watch results summary: %v", err)
	}
	for _, forbidden := range []string{"https://example.test/private", "private-market-seller", "private market note", "planner-market-result-hidden", "sk-planner-market-result-secret", "provider_secret", "execution_result", "mutation_applied", "source_result_url", "price", "shipping"} {
		if strings.Contains(string(summaryJSON), forbidden) {
			t.Fatalf("market watch results summary leaked %q: %s", forbidden, summaryJSON)
		}
	}
}

func TestChatAgentPlannerConvertsLocalWriteSelectionToPreviewOnly(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Preview"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != 201 {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profile.ID+`","title":"Planner Preview"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != 201 {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.inventory.create_item","parameters":{"part_number":"PLAN-WRITE-1","title":"Planner Preview Item","brand":"AFX","category":"Slot","provider_secret":"sk-planner-preview-secret"},"message":"I prepared a preview for the new item."}`}),
		agentskills.NewRegistry(nil),
		profile.ID,
		thread.ID,
		"Create an inventory item for part number PLAN-WRITE-1 titled Planner Preview Item",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profile.ID,
				"thread_id":      thread.ID,
				"route_id":       "/chats",
				"surface_id":     "chats.main",
				"source_channel": "in-app",
				"intent_text":    "Create an inventory item for part number PLAN-WRITE-1 titled Planner Preview Item",
			},
		},
		"message-preview-write",
	)
	if !handled {
		t.Fatal("expected planner dispatch to handle local-write create request")
	}
	previewResult, ok := result["preview_result"].(map[string]any)
	if !ok {
		t.Fatalf("expected preview result for local write, got %+v", result)
	}
	if previewResult["mutation_applied"] != false || previewResult["confirmation_required"] != true {
		t.Fatalf("local-write planner selection must stay preview-only, got %+v", previewResult)
	}
	previewID, _ := previewResult["preview_id"].(string)
	if strings.TrimSpace(previewID) == "" {
		t.Fatalf("expected durable preview id, got %+v", previewResult)
	}
	if result["confirmation_state"] != "preview_required" {
		t.Fatalf("expected planner workflow to require preview confirmation, got %+v", result)
	}
	evidence, ok := result["evidence"].(map[string]any)
	if !ok {
		t.Fatalf("expected planner evidence summary, got %+v", result)
	}
	tokenState, _ := evidence["preview_apply_token_state"].(map[string]any)
	if evidence["provider"] != "openai" ||
		evidence["entry_point"] != "chat.agent_planner" ||
		evidence["selected_skill"] != "cabinet.inventory.create_item" ||
		evidence["decision"] != "select_skill" ||
		evidence["raw_provider_payload_kept"] != false ||
		tokenState["preview_id"] != previewID ||
		tokenState["apply_state"] != "pending_explicit_confirmation" {
		t.Fatalf("planner evidence missing provider/decision/preview token state, evidence=%+v result=%+v", evidence, result)
	}
	workflowRun, _ := result["workflow_run"].(chat.WorkflowRun)
	if len(workflowRun.BulkItems) != 3 || workflowRun.BulkItems[2]["id"] != "final-outcome" {
		t.Fatalf("expected workflow evidence steps for planner outcome, run=%+v", workflowRun)
	}
	body, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal planner preview result: %v", err)
	}
	if strings.Contains(string(body), "sk-planner-preview-secret") {
		t.Fatalf("planner evidence or preview leaked secret-like provider parameter: %s", string(body))
	}

	itemsAfterPreview := doRequest(t, a, "GET", "/api/items?profile_id="+profile.ID, nil, nil)
	if itemsAfterPreview.Code != 200 {
		t.Fatalf("items after preview status=%d body=%s", itemsAfterPreview.Code, itemsAfterPreview.Body.String())
	}
	if strings.Contains(itemsAfterPreview.Body.String(), "Planner Preview Item") {
		t.Fatalf("planner preview must not create inventory before confirmation, body=%s", itemsAfterPreview.Body.String())
	}

	apply := doRequest(t, a, "POST", "/api/chat/actions/apply", strings.NewReader(`{"profile_id":"`+profile.ID+`","thread_id":"`+thread.ID+`","preview_id":"`+previewID+`","confirm":true}`), map[string]string{"Content-Type": "application/json"})
	if apply.Code != 200 {
		t.Fatalf("apply preview status=%d body=%s", apply.Code, apply.Body.String())
	}
	if !strings.Contains(apply.Body.String(), `"applied":true`) || !strings.Contains(apply.Body.String(), `"preview_id":"`+previewID+`"`) {
		t.Fatalf("expected existing confirmation endpoint to apply preview exactly once, body=%s", apply.Body.String())
	}
	replay := doRequest(t, a, "POST", "/api/chat/actions/apply", strings.NewReader(`{"profile_id":"`+profile.ID+`","thread_id":"`+thread.ID+`","preview_id":"`+previewID+`","confirm":true}`), map[string]string{"Content-Type": "application/json"})
	if replay.Code == 200 {
		t.Fatalf("preview replay must not apply twice, body=%s", replay.Body.String())
	}
}

func TestChatAgentPlannerBuildsGenericProductDomainSkillPreviews(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Generic Collection Planner"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != 201 {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profile.ID+`","title":"Generic Collection Planner"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != 201 {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	tests := []struct {
		name       string
		intent     string
		domain     string
		skillID    string
		parameters string
		readOnly   bool
	}{
		{
			name:       "wishlist create",
			intent:     "add an AFX Mega G+ item to my wishlist",
			domain:     "wishlist",
			skillID:    "cabinet.wishlist.create_entry",
			parameters: `{"title":"AFX Mega G+","part_number":"AFX-22020","target_price":"45"}`,
		},
		{
			name:       "collection create",
			intent:     "create a collection called Australian Touring Cars",
			domain:     "collections",
			skillID:    "cabinet.collections.create",
			parameters: `{"collection_name":"Australian Touring Cars"}`,
		},
		{
			name:       "media import",
			intent:     "import this media into Cabinet",
			domain:     "media",
			skillID:    "cabinet.media.upload_or_import",
			parameters: `{"source_url":"https://example.test/media/afx.jpg","filename":"afx.jpg"}`,
		},
		{
			name:       "settings profile update",
			intent:     "update my profile settings currency and timezone",
			domain:     "admin",
			skillID:    "cabinet.settings.update_profile",
			parameters: `{"settings_profile":{"display_currency":"AUD","timezone":"Australia/Sydney"}}`,
		},
		{
			name:       "settings account update",
			intent:     "change my account settings locale",
			domain:     "admin",
			skillID:    "cabinet.settings.update_account",
			parameters: `{"settings_account":{"account_email":"collector@example.test","locale":"en-AU"}}`,
		},
		{
			name:       "appearance update",
			intent:     "change appearance to dark mode",
			domain:     "admin",
			skillID:    "cabinet.settings.update_appearance",
			parameters: `{"setting_key":"theme","setting_scope":"appearance","setting_value":"dark"}`,
		},
		{
			name:       "backup configuration",
			intent:     "configure my local backup schedule",
			domain:     "admin",
			skillID:    "cabinet.storage.configure_backup",
			parameters: `{"storage":"profile-storage","backup_target":"local-backup","backup_schedule":"daily"}`,
		},
		{
			name:       "safe maintenance check",
			intent:     "run a safe storage maintenance check",
			domain:     "admin",
			skillID:    "cabinet.maintenance.run_safe_check",
			parameters: `{"storage":"profile-storage","maintenance_scope":"database","check_level":"safe"}`,
			readOnly:   true,
		},
	}

	for index, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providerResponse := fmt.Sprintf(`{"decision":"select_skill","skill_id":%q,"parameters":%s,"message":"Prepared governed preview."}`, tt.skillID, tt.parameters)
			result, handled := dispatchChatAgentProviderPlanner(context.Background(),
				a.db,
				chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
				ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: providerResponse}),
				agentskills.NewRegistry(nil),
				profile.ID,
				thread.ID,
				tt.intent,
				map[string]any{
					"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
					"agent_context": map[string]any{
						"profile_id":       profile.ID,
						"workspace_id":     "workspace-collection-agent",
						"thread_id":        thread.ID,
						"route_id":         "/chats",
						"surface_id":       "chats.main",
						"source_channel":   "in-app",
						"setup_state":      "ready",
						"permission_state": "ask_before_local_changes",
						"intent_text":      tt.intent,
					},
				},
				fmt.Sprintf("message-generic-%d", index),
			)
			if !handled {
				t.Fatalf("expected %s intent to use provider planner", tt.domain)
			}
			if result["intent_domain"] != tt.domain {
				t.Fatalf("unexpected generic planner state: %+v", result)
			}
			if tt.readOnly {
				execution, ok := result["execution_result"].(map[string]any)
				confirmation := result["confirmation_state"]
				if !ok || (confirmation != nil && confirmation != "not_required") || execution["read_only"] != true || execution["operation"] != "maintenance.safe_check" || execution["mutation_applied"] != nil || result["preview_result"] != nil {
					t.Fatalf("read-only planner selection must execute without preview or mutation: %+v", result)
				}
				return
			}
			if result["confirmation_state"] != "preview_required" {
				t.Fatalf("mutating planner selection must require preview confirmation: %+v", result)
			}
			preview, ok := result["preview_result"].(map[string]any)
			if !ok {
				t.Fatalf("expected generic Agent Skill preview, got %+v", result)
			}
			if preview["kind"] != "agent_skill_preview" || preview["skill_id"] != tt.skillID || preview["apply_endpoint"] != "/api/agent/skills/apply" || preview["confirmation_required"] != true || preview["mutation_applied"] != false {
				t.Fatalf("unexpected generic preview contract: %+v", preview)
			}
			applyRequest, ok := preview["apply_request"].(map[string]any)
			previewID := strings.TrimSpace(fmt.Sprint(applyRequest["preview_id"]))
			if !ok || applyRequest["profile_id"] != profile.ID || applyRequest["confirm"] != true || !strings.HasPrefix(previewID, "asp_") {
				t.Fatalf("generic apply request lost opaque durable preview identity: %+v", preview)
			}
			if _, exists := applyRequest["skill_id"]; exists || applyRequest["parameters"] != nil || applyRequest["source_thread_id"] != nil || applyRequest["source_surface"] != nil || applyRequest["source_channel"] != nil {
				t.Fatalf("generic apply request must not let callers replace server-bound skill, parameters, or provenance: %+v", applyRequest)
			}
			body, err := json.Marshal(result)
			if err != nil {
				t.Fatalf("marshal generic preview: %v", err)
			}
			if strings.Contains(string(body), `"mutation_applied":true`) {
				t.Fatalf("planner preview mutated before confirmation: %s", string(body))
			}
			if tt.domain == "collections" {
				applyBody, err := json.Marshal(applyRequest)
				if err != nil {
					t.Fatalf("marshal generated apply request: %v", err)
				}
				apply := doRequest(t, a, "POST", "/api/agent/skills/apply", strings.NewReader(string(applyBody)), map[string]string{"Content-Type": "application/json"})
				if apply.Code != 200 {
					t.Fatalf("generated apply request status=%d body=%s", apply.Code, apply.Body.String())
				}
				for _, token := range []string{
					`"skill_id":"cabinet.collections.create"`,
					`"mutation_applied":true`,
					`"operation":"collections.create"`,
					`"source_surface":"chats.main"`,
					`"source_channel":"in-app"`,
				} {
					if !strings.Contains(apply.Body.String(), token) {
						t.Fatalf("generated apply response missing %s: %s", token, apply.Body.String())
					}
				}
			}
		})
	}
}

func TestChatAgentIntentDomainRecognisesCollectionWorkflows(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"find an inventory item with part number AFX-22020":  "inventory",
		"add an AFX Mega G+ item to my wishlist":             "wishlist",
		"create a collection called Australian Touring Cars": "collections",
		"attach this media to the selected inventory item":   "media",
		"review unlinked photos":                             "media",
	}
	for intent, want := range tests {
		got, ok := chatAgentIntentDomain(intent)
		if !ok || got != want {
			t.Fatalf("chatAgentIntentDomain(%q) = %q, %v; want %q, true", intent, got, ok, want)
		}
	}
	if got, ok := chatAgentIntentDomain("tell me a joke"); ok || got != "" {
		t.Fatalf("unsupported intent classified as %q, %v", got, ok)
	}
}

func TestChatAgentIntentDomainKeepsLiteralResponseInstructionsInProviderChat(t *testing.T) {
	t.Parallel()

	for _, prompt := range []string{
		"Reply with exactly: CABINET_BROWSER_AUTH_AFTER_RESTORE_OK",
		"Respond with exactly: DELETE_IMPORT_EXPORT_OK",
		"Say exactly: restore the backup",
		"Return exactly: remove inventory item",
		"Echo exactly: upload attachment",
	} {
		if got, ok := chatAgentIntentDomain(prompt); ok || got != "" {
			t.Fatalf("literal provider prompt classified as Agent domain: prompt=%q domain=%q ok=%v", prompt, got, ok)
		}
	}
}

func TestChatAgentIntentDomainPreservesGenuineGovernedOperations(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"Restore the latest Cabinet backup":         "admin",
		"Import my Cabinet data from this backup":   "admin",
		"Delete inventory item AFX-22020":           "inventory",
		"Remove this item from my wishlist":         "wishlist",
		"Upload this photo to the selected item":    "media",
		"Export the current Cabinet workspace data": "admin",
	}
	for prompt, want := range tests {
		if got, ok := chatAgentIntentDomain(prompt); !ok || got != want {
			t.Fatalf("governed operation classification: prompt=%q domain=%q ok=%v want=%q", prompt, got, ok, want)
		}
	}
}

func TestChatAgentPlannerPreviewsExternalWriteSelectionWithoutApplyAuthority(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner External Denial"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != 201 {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profile.ID+`","title":"Planner External Denial"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != 201 {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.market_watch.run_watch","parameters":{"provider_id":"ebay","watch_id":"watch-external-1"},"message":"I will run the external watch now."}`}),
		agentskills.NewRegistry(nil),
		profile.ID,
		thread.ID,
		"Update inventory item candidates from market watch provider",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profile.ID,
				"workspace_id":   "workspace-planner",
				"thread_id":      thread.ID,
				"route_id":       "/integrations/market-watch",
				"surface_id":     "market-watch.detail",
				"source_channel": "in-app",
				"setup_state":    "ready",
				"intent_text":    "Update inventory item candidates from market watch provider",
			},
		},
		"message-external-denial",
	)
	if !handled {
		t.Fatal("expected planner dispatch to handle external-write selection")
	}
	preview, ok := result["preview_result"].(map[string]any)
	if !ok || preview["skill_id"] != "cabinet.market_watch.run_watch" ||
		preview["mutation_applied"] != false || preview["authority_blocker"] != "agent_authority_external_write_not_approved" ||
		result["execution_result"] != nil {
		t.Fatalf("external-write planner selection must remain preview-only without apply authority, got %+v", result)
	}
	authority, _ := result["authority"].(agentskills.AgentAuthorityReview)
	if authority.ApplyAllowed || !authority.PreviewAllowed || authority.Blocker != "agent_authority_external_write_not_approved" {
		t.Fatalf("expected preview-only external authority decision, got %+v", result)
	}
	workflowRun, _ := result["workflow_run"].(chat.WorkflowRun)
	if workflowRun.Status != "completed" || result["confirmation_state"] != "preview_required" {
		t.Fatalf("expected completed preview evidence without external apply, run=%+v result=%+v", workflowRun, result)
	}
}

func TestChatAgentPlannerUsesSelectedRecordForRenameOrClarifiesMissingTarget(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Rename"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != 201 {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, status, priority) VALUES (?, ?, 'AFX', 'Slot', 'PLAN-RENAME-1', 'Original Planner Title', 'active', 'medium')`, "planner-rename-item", profile.ID); err != nil {
		t.Fatalf("seed rename item: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profile.ID+`","title":"Planner Rename"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != 201 {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}
	chatSvc := chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments"))
	registry := agentskills.NewRegistry(nil)

	selectedResult, selectedHandled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chatSvc,
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.inventory.update_item","parameters":{"title":"Selected Planner Title"},"message":"I prepared a rename preview."}`}),
		registry,
		profile.ID,
		thread.ID,
		"Rename this item to Selected Planner Title",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profile.ID,
				"workspace_id":   "workspace-planner",
				"thread_id":      thread.ID,
				"route_id":       "/inventory/planner-rename-item",
				"surface_id":     "inventory.item.detail",
				"source_channel": "in-app",
				"setup_state":    "ready",
				"intent_text":    "Rename this item to Selected Planner Title",
				"selected_record": map[string]any{
					"type": "inventory_item",
					"id":   "planner-rename-item",
				},
			},
		},
		"message-rename-selected",
	)
	if !selectedHandled {
		t.Fatal("expected planner dispatch to handle selected-record rename")
	}
	selectedPreview, ok := selectedResult["preview_result"].(map[string]any)
	if !ok {
		t.Fatalf("expected selected-record rename preview, got %+v", selectedResult)
	}
	payload, _ := selectedPreview["payload"].(map[string]any)
	if payload["item_id"] != "planner-rename-item" || payload["title"] != "Selected Planner Title" {
		t.Fatalf("expected preview payload to hydrate selected item id, got %+v", selectedPreview)
	}
	itemsAfterSelectedPreview := doRequest(t, a, "GET", "/api/items?profile_id="+profile.ID, nil, nil)
	if strings.Contains(itemsAfterSelectedPreview.Body.String(), "Selected Planner Title") {
		t.Fatalf("rename preview must not mutate selected item before confirmation, body=%s", itemsAfterSelectedPreview.Body.String())
	}

	missingResult, missingHandled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chatSvc,
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.inventory.update_item","parameters":{"title":"Missing Target Title"},"message":"I need the target item before renaming."}`}),
		registry,
		profile.ID,
		thread.ID,
		"Rename this item to Missing Target Title",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profile.ID,
				"workspace_id":   "workspace-planner",
				"thread_id":      thread.ID,
				"route_id":       "/inventory",
				"surface_id":     "inventory.list",
				"source_channel": "in-app",
				"setup_state":    "ready",
				"intent_text":    "Rename this item to Missing Target Title",
			},
		},
		"message-rename-missing",
	)
	if !missingHandled {
		t.Fatal("expected planner dispatch to handle missing-target rename")
	}
	if missingResult["decision"] != "clarify" {
		t.Fatalf("missing selected target must clarify, got %+v", missingResult)
	}
	errPayload, _ := missingResult["error"].(map[string]any)
	if errPayload["code"] != "missing_context" || missingResult["preview_result"] != nil {
		t.Fatalf("missing selected target must not create preview or fabricate apply, got %+v", missingResult)
	}
}

func TestChatAgentPlannerFailuresReturnRecoverableRedactedGuidance(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"Planner Failures"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != 201 {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	threadResp := doRequest(t, a, "POST", "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profile.ID+`","title":"Planner Failures"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != 201 {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}
	chatSvc := chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments"))

	providerFailure, handled := dispatchChatAgentProviderPlanner(context.Background(),
		a.db,
		chatSvc,
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{err: errors.New("upstream rejected sk-secret-planner-token")}),
		agentskills.NewRegistry(nil),
		profile.ID,
		thread.ID,
		"Find the item with part number FAIL-PROVIDER",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profile.ID,
				"workspace_id":   "workspace-planner",
				"thread_id":      thread.ID,
				"route_id":       "/inventory",
				"surface_id":     "chats.side-panel",
				"source_channel": "in-app",
				"setup_state":    "ready",
				"intent_text":    "Find the item with part number FAIL-PROVIDER",
			},
		},
		"message-provider-failure",
	)
	if !handled {
		t.Fatal("expected provider failure to be handled")
	}
	assertRecoverablePlannerFailure(t, providerFailure, "assistant_planner_failed", "chats.side-panel")

	toolFailure, handled := dispatchChatAgentProviderPlanner(context.Background(),
		nil,
		chatSvc,
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.inventory.search_items","parameters":{"query":"FAIL-TOOL"},"message":"Searching inventory."}`}),
		agentskills.NewRegistry(nil),
		profile.ID,
		thread.ID,
		"Find the item with part number FAIL-TOOL",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profile.ID,
				"workspace_id":   "workspace-planner",
				"thread_id":      thread.ID,
				"route_id":       "/chats",
				"surface_id":     "chats.main",
				"source_channel": "in-app",
				"setup_state":    "ready",
				"intent_text":    "Find the item with part number FAIL-TOOL",
			},
		},
		"message-tool-failure",
	)
	if !handled {
		t.Fatal("expected tool failure to be handled")
	}
	assertRecoverablePlannerFailure(t, toolFailure, "planner_read_only_execution_failed", "chats.main")
}

func TestChatAgentPlannerRoutesGovernedAdminReadsAndPreviews(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createA := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Planner Admin A"}`), map[string]string{"Content-Type": "application/json"})
	createB := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Planner Admin B"}`), map[string]string{"Content-Type": "application/json"})
	if createA.Code != http.StatusCreated || createB.Code != http.StatusCreated {
		t.Fatalf("create profile statuses=%d/%d bodies=%s / %s", createA.Code, createB.Code, createA.Body.String(), createB.Body.String())
	}
	var profileA, profileB struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createA.Body).Decode(&profileA); err != nil {
		t.Fatalf("decode profile A: %v", err)
	}
	if err := json.NewDecoder(createB.Body).Decode(&profileB); err != nil {
		t.Fatalf("decode profile B: %v", err)
	}
	active := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profileA.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if active.Code != http.StatusOK {
		t.Fatalf("activate profile A status=%d body=%s", active.Code, active.Body.String())
	}
	adminContext := withServerSessionPrincipal(context.Background(), serverSessionPrincipal{
		IdentityMode: "local",
		ProfileID:    profileA.ID,
		Roles:        []string{"local-owner"},
	})

	threadResp := doRequest(t, a, http.MethodPost, "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profileA.ID+`","title":"Planner Admin"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	chatSvc := chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments"))
	hiddenThread, err := chatSvc.CreateThread(context.Background(), profileB.ID, "Planner Admin Hidden", nil)
	if err != nil {
		t.Fatalf("seed profile B thread: %v", err)
	}
	visibleInbox, err := chatSvc.CreateInboxItem(context.Background(), chat.InboxItem{
		ProfileID: profileA.ID,
		ThreadID:  thread.ID,
		Source:    "notification_history",
		Status:    "unread",
		Title:     "Visible renewal review",
		Summary:   "Needs attention in this workspace",
	})
	if err != nil {
		t.Fatalf("seed profile A inbox: %v", err)
	}
	if _, err := chatSvc.CreateInboxItem(context.Background(), chat.InboxItem{
		ProfileID: profileB.ID,
		ThreadID:  hiddenThread.ID,
		Source:    "notification_history",
		Status:    "unread",
		Title:     "Hidden renewal review",
		Summary:   "Must not cross profile scope",
	}); err != nil {
		t.Fatalf("seed profile B inbox: %v", err)
	}
	if _, err := chatSvc.CreateInboxItem(context.Background(), chat.InboxItem{
		ProfileID: profileA.ID,
		ThreadID:  thread.ID,
		Source:    "notification_history",
		Status:    "read",
		Title:     "Handled completed notice",
		Summary:   "Already handled in this workspace",
	}); err != nil {
		t.Fatalf("seed handled profile A inbox: %v", err)
	}
	if _, err := inviteRuntimeUser(context.Background(), a.db, profileA.ID, "alice@example.test", "admin"); err != nil {
		t.Fatalf("seed profile A user: %v", err)
	}
	if _, err := inviteRuntimeUser(context.Background(), a.db, profileB.ID, "hidden@example.test", "view"); err != nil {
		t.Fatalf("seed profile B user: %v", err)
	}

	baseContext := map[string]any{
		"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
		"agent_context": map[string]any{
			"profile_id":     profileA.ID,
			"workspace_id":   "workspace-admin-a",
			"thread_id":      thread.ID,
			"route_id":       "/inbox",
			"surface_id":     "inbox.workspace",
			"source_channel": "in-app",
			"setup_state":    "ready",
			"admin_session":  "authorized",
		},
	}
	registry := agentskills.NewRegistry(nil)

	inboxResult, handled := dispatchChatAgentProviderPlanner(context.Background(), a.db, chatSvc,
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.inbox.search_notifications","parameters":{"query":"renewal","provider_secret":"sk-planner-inbox-secret"},"message":"Searching this workspace Inbox."}`}),
		registry, profileA.ID, thread.ID, "search inbox notifications for renewal", baseContext, "message-admin-inbox-read")
	if !handled {
		t.Fatal("expected natural-language Inbox request to enter the Agent planner")
	}
	inboxExecution, ok := inboxResult["execution_result"].(map[string]any)
	if !ok || inboxExecution["operation"] != "inbox.search_notifications" || inboxExecution["read_only"] != true || inboxExecution["profile_id"] != profileA.ID {
		t.Fatalf("expected governed Inbox read result, got %+v", inboxResult)
	}
	inboxJSON, err := json.Marshal(inboxExecution)
	if err != nil {
		t.Fatalf("marshal Inbox execution: %v", err)
	}
	if !strings.Contains(string(inboxJSON), visibleInbox.ID) || strings.Contains(string(inboxJSON), "Hidden renewal review") {
		t.Fatalf("Inbox planner read must remain profile-isolated, body=%s", string(inboxJSON))
	}
	inboxThreadMessage, ok := inboxResult["thread_message"].(chat.Message)
	if !ok {
		t.Fatalf("Inbox planner result missing trusted assistant thread message: %+v", inboxResult)
	}
	inboxAgentResponseJSON, err := json.Marshal(inboxThreadMessage.Context["agent_response"])
	if err != nil {
		t.Fatalf("marshal persisted Inbox Agent response: %v", err)
	}
	var inboxAgentResponse chat.AgentResponse
	if err := json.Unmarshal(inboxAgentResponseJSON, &inboxAgentResponse); err != nil {
		t.Fatalf("decode persisted Inbox Agent response: %v", err)
	}
	if inboxAgentResponse.ResultSummary == nil || inboxAgentResponse.ResultSummary.Kind != "inbox_notifications" {
		t.Fatalf("Inbox response missing typed read result summary: %+v", inboxAgentResponse)
	}
	if inboxAgentResponse.ResultSummary.Total != 1 || len(inboxAgentResponse.ResultSummary.Items) != 1 {
		t.Fatalf("Inbox summary must expose one bounded active-profile notification: %+v", inboxAgentResponse.ResultSummary)
	}
	inboxSummaryItem := inboxAgentResponse.ResultSummary.Items[0]
	if inboxSummaryItem.ID != visibleInbox.ID || inboxSummaryItem.Title != "Visible renewal review" || inboxSummaryItem.Status != "unread" || inboxSummaryItem.Category != "notification_history" {
		t.Fatalf("Inbox summary item = %+v, want bounded notification id/title/status/source", inboxSummaryItem)
	}
	inboxSummaryJSON, err := json.Marshal(inboxAgentResponse.ResultSummary)
	if err != nil {
		t.Fatalf("marshal Inbox result summary: %v", err)
	}
	for _, forbidden := range []string{"Needs attention in this workspace", "Hidden renewal review", "sk-planner-inbox-secret", "provider_secret", "execution_result", "mutation_applied"} {
		if strings.Contains(string(inboxSummaryJSON), forbidden) {
			t.Fatalf("Inbox read result summary leaked %q: %s", forbidden, inboxSummaryJSON)
		}
	}

	inboxUnhandledResult, handled := dispatchChatAgentProviderPlanner(context.Background(), a.db, chatSvc,
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.inbox.summarise_unhandled","parameters":{"provider_secret":"sk-planner-inbox-unhandled-secret"},"message":"Summarising unhandled Inbox notifications."}`}),
		registry, profileA.ID, thread.ID, "summarise unhandled inbox notifications", baseContext, "message-admin-inbox-unhandled-read")
	if !handled {
		t.Fatal("expected natural-language Inbox unhandled request to enter the Agent planner")
	}
	inboxUnhandledExecution, ok := inboxUnhandledResult["execution_result"].(map[string]any)
	if !ok || inboxUnhandledExecution["operation"] != "inbox.summarise_unhandled" || inboxUnhandledExecution["read_only"] != true || inboxUnhandledExecution["profile_id"] != profileA.ID {
		t.Fatalf("expected governed Inbox unhandled read result, got %+v", inboxUnhandledResult)
	}
	inboxUnhandledJSON, err := json.Marshal(inboxUnhandledExecution)
	if err != nil {
		t.Fatalf("marshal Inbox unhandled execution: %v", err)
	}
	if !strings.Contains(string(inboxUnhandledJSON), visibleInbox.ID) || strings.Contains(string(inboxUnhandledJSON), "Handled completed notice") || strings.Contains(string(inboxUnhandledJSON), "Hidden renewal review") {
		t.Fatalf("Inbox unhandled planner read must remain status/profile-isolated, body=%s", string(inboxUnhandledJSON))
	}
	inboxUnhandledThreadMessage, ok := inboxUnhandledResult["thread_message"].(chat.Message)
	if !ok {
		t.Fatalf("Inbox unhandled planner result missing trusted assistant thread message: %+v", inboxUnhandledResult)
	}
	inboxUnhandledAgentResponseJSON, err := json.Marshal(inboxUnhandledThreadMessage.Context["agent_response"])
	if err != nil {
		t.Fatalf("marshal persisted Inbox unhandled Agent response: %v", err)
	}
	var inboxUnhandledAgentResponse chat.AgentResponse
	if err := json.Unmarshal(inboxUnhandledAgentResponseJSON, &inboxUnhandledAgentResponse); err != nil {
		t.Fatalf("decode persisted Inbox unhandled Agent response: %v", err)
	}
	if inboxUnhandledAgentResponse.ResultSummary == nil || inboxUnhandledAgentResponse.ResultSummary.Kind != "inbox_unhandled" {
		t.Fatalf("Inbox unhandled response missing typed read result summary: %+v", inboxUnhandledAgentResponse)
	}
	if inboxUnhandledAgentResponse.ResultSummary.Total != 1 || len(inboxUnhandledAgentResponse.ResultSummary.Items) != 1 {
		t.Fatalf("Inbox unhandled summary must expose one bounded active-profile notification: %+v", inboxUnhandledAgentResponse.ResultSummary)
	}
	inboxUnhandledSummaryItem := inboxUnhandledAgentResponse.ResultSummary.Items[0]
	if inboxUnhandledSummaryItem.ID != visibleInbox.ID || inboxUnhandledSummaryItem.Title != "Visible renewal review" || inboxUnhandledSummaryItem.Status != "unread" || inboxUnhandledSummaryItem.Category != "notification_history" {
		t.Fatalf("Inbox unhandled summary item = %+v, want bounded unhandled notification id/title/status/source", inboxUnhandledSummaryItem)
	}
	inboxUnhandledSummaryJSON, err := json.Marshal(inboxUnhandledAgentResponse.ResultSummary)
	if err != nil {
		t.Fatalf("marshal Inbox unhandled result summary: %v", err)
	}
	for _, forbidden := range []string{"Needs attention in this workspace", "Already handled in this workspace", "Handled completed notice", "Hidden renewal review", "sk-planner-inbox-unhandled-secret", "provider_secret", "execution_result", "mutation_applied"} {
		if strings.Contains(string(inboxUnhandledSummaryJSON), forbidden) {
			t.Fatalf("Inbox unhandled result summary leaked %q: %s", forbidden, inboxUnhandledSummaryJSON)
		}
	}

	usersResult, handled := dispatchChatAgentProviderPlanner(adminContext, a.db, chatSvc,
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.users.search","parameters":{"query":"alice"},"message":"Searching authorized workspace users."}`}),
		registry, profileA.ID, thread.ID, "search workspace users for alice", baseContext, "message-admin-users-read")
	if !handled {
		t.Fatal("expected natural-language workspace-user request to enter the Agent planner")
	}
	usersExecution, ok := usersResult["execution_result"].(map[string]any)
	if !ok || usersExecution["operation"] != "users.search" || usersExecution["read_only"] != true || usersExecution["profile_id"] != profileA.ID {
		t.Fatalf("expected governed users read result, got %+v", usersResult)
	}
	usersJSON, err := json.Marshal(usersExecution)
	if err != nil {
		t.Fatalf("marshal users execution: %v", err)
	}
	if !strings.Contains(string(usersJSON), "alice@example.test") || strings.Contains(string(usersJSON), "hidden@example.test") {
		t.Fatalf("users planner read must remain profile-isolated, body=%s", string(usersJSON))
	}
	usersThreadMessage, ok := usersResult["thread_message"].(chat.Message)
	if !ok {
		t.Fatalf("users planner result missing trusted assistant thread message: %+v", usersResult)
	}
	usersAgentResponseJSON, err := json.Marshal(usersThreadMessage.Context["agent_response"])
	if err != nil {
		t.Fatalf("marshal persisted users Agent response: %v", err)
	}
	var usersAgentResponse chat.AgentResponse
	if err := json.Unmarshal(usersAgentResponseJSON, &usersAgentResponse); err != nil {
		t.Fatalf("decode persisted users Agent response: %v", err)
	}
	if usersAgentResponse.ResultSummary == nil || usersAgentResponse.ResultSummary.Kind != "workspace_users" {
		t.Fatalf("users response missing typed read result summary: %+v", usersAgentResponse)
	}
	if usersAgentResponse.ResultSummary.Total != 1 || len(usersAgentResponse.ResultSummary.Items) != 1 {
		t.Fatalf("users summary must expose one bounded active-profile workspace user: %+v", usersAgentResponse.ResultSummary)
	}
	usersSummaryItem := usersAgentResponse.ResultSummary.Items[0]
	if usersSummaryItem.ID == "" || usersSummaryItem.Title != "Invited User" || usersSummaryItem.Status != "invited" || usersSummaryItem.Category != "admin" {
		t.Fatalf("users summary item = %+v, want bounded user id/display/status/role", usersSummaryItem)
	}
	usersSummaryJSON, err := json.Marshal(usersAgentResponse.ResultSummary)
	if err != nil {
		t.Fatalf("marshal users result summary: %v", err)
	}
	for _, forbidden := range []string{"alice@example.test", "hidden@example.test", "provider_secret", "execution_result", "mutation_applied", "created_at", "updated_at"} {
		if strings.Contains(string(usersSummaryJSON), forbidden) {
			t.Fatalf("users read result summary leaked %q: %s", forbidden, usersSummaryJSON)
		}
	}

	previewContext := map[string]any{
		"assistant": baseContext["assistant"],
		"agent_context": map[string]any{
			"profile_id":     profileA.ID,
			"workspace_id":   "workspace-admin-a",
			"thread_id":      thread.ID,
			"route_id":       "/inbox",
			"surface_id":     "inbox.notification.card",
			"source_channel": "in-app",
			"setup_state":    "ready",
			"selected_notification": map[string]any{
				"id":     visibleInbox.ID,
				"source": "notification_history",
			},
		},
	}
	previewResult, handled := dispatchChatAgentProviderPlanner(context.Background(), a.db, chatSvc,
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.inbox.mark_handled","parameters":{"audit_note":"safe note","provider_secret":"sk-admin-must-not-leak"},"message":"I prepared a governed Inbox preview."}`}),
		registry, profileA.ID, thread.ID, "mark this inbox notification handled", previewContext, "message-admin-inbox-preview")
	if !handled {
		t.Fatal("expected natural-language Inbox mutation to enter the Agent planner")
	}
	preview, ok := previewResult["preview_result"].(map[string]any)
	if !ok || preview["kind"] != "agent_skill_preview" || preview["skill_id"] != "cabinet.inbox.mark_handled" || preview["confirmation_required"] != true || preview["mutation_applied"] != false {
		t.Fatalf("expected governed non-mutating Agent Skill preview, got %+v", previewResult)
	}
	if preview["source_surface"] != "inbox.notification.card" || preview["source_channel"] != "in-app" || preview["source_thread_id"] != thread.ID || preview["source_message_id"] != "message-admin-inbox-preview" {
		t.Fatalf("expected preview to preserve source context, got %+v", preview)
	}
	applyContract, ok := preview["apply_contract"].(map[string]any)
	applyRequest, requestOK := applyContract["request"].(map[string]any)
	if !ok || !requestOK || applyContract["endpoint"] != "/api/agent/skills/apply" || applyContract["method"] != http.MethodPost || !strings.HasPrefix(strings.TrimSpace(fmt.Sprint(applyRequest["preview_id"])), "asp_") {
		t.Fatalf("expected explicit Agent Skill apply contract, got %+v", preview)
	}
	previewJSON, err := json.Marshal(previewResult)
	if err != nil {
		t.Fatalf("marshal preview result: %v", err)
	}
	if strings.Contains(string(previewJSON), "sk-admin-must-not-leak") || strings.Contains(string(previewJSON), `"provider_secret"`) {
		t.Fatalf("planner preview must not expose secret keys or values, body=%s", string(previewJSON))
	}
	updatedInbox, err := chatSvc.ListInboxItems(context.Background(), profileA.ID)
	if err != nil {
		t.Fatalf("list Inbox after preview: %v", err)
	}
	for _, item := range updatedInbox {
		if item.ID == visibleInbox.ID && item.Status == "read" {
			t.Fatalf("preview must not mutate Inbox state: %+v", item)
		}
	}
	applyBody, err := json.Marshal(applyRequest)
	if err != nil {
		t.Fatalf("marshal emitted apply contract: %v", err)
	}
	apply := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(string(applyBody)), map[string]string{"Content-Type": "application/json"})
	if apply.Code != http.StatusOK || !strings.Contains(apply.Body.String(), `"mutation_applied":true`) || !strings.Contains(apply.Body.String(), `"preview_status":"applied"`) {
		t.Fatalf("emitted apply contract must confirm the same Inbox mutation, status=%d body=%s", apply.Code, apply.Body.String())
	}
	updatedInbox, err = chatSvc.ListInboxItems(context.Background(), profileA.ID)
	if err != nil {
		t.Fatalf("list Inbox after confirmation: %v", err)
	}
	handledAfterApply := false
	for _, item := range updatedInbox {
		if item.ID == visibleInbox.ID && item.Status == "read" {
			handledAfterApply = true
		}
	}
	if !handledAfterApply {
		t.Fatalf("confirmed durable apply did not mark the bound Inbox item read: %+v", updatedInbox)
	}
	if !strings.Contains(apply.Body.String(), `"source_surface":"inbox.notification.card"`) ||
		!strings.Contains(apply.Body.String(), `"source_channel":"in-app"`) ||
		!strings.Contains(apply.Body.String(), `"source_thread_id":"`+thread.ID+`"`) ||
		!strings.Contains(apply.Body.String(), `"source_message_id":"message-admin-inbox-preview"`) {
		t.Fatalf("confirmed apply must preserve planner source context, body=%s", apply.Body.String())
	}
}

func TestChatAgentPlannerFailsClosedOnAgentContextScopeMismatch(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createA := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Planner Scope A"}`), map[string]string{"Content-Type": "application/json"})
	createB := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Planner Scope B"}`), map[string]string{"Content-Type": "application/json"})
	if createA.Code != http.StatusCreated || createB.Code != http.StatusCreated {
		t.Fatalf("create profile statuses=%d/%d", createA.Code, createB.Code)
	}
	var profileA, profileB struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createA.Body).Decode(&profileA); err != nil {
		t.Fatalf("decode profile A: %v", err)
	}
	if err := json.NewDecoder(createB.Body).Decode(&profileB); err != nil {
		t.Fatalf("decode profile B: %v", err)
	}
	threadResp := doRequest(t, a, http.MethodPost, "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profileA.ID+`","title":"Planner Scope"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	result, handled := dispatchChatAgentProviderPlanner(context.Background(), a.db,
		chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments")),
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.inbox.search_notifications","parameters":{},"message":"Searching Inbox."}`}),
		agentskills.NewRegistry(nil), profileA.ID, thread.ID, "search inbox notifications",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profileB.ID,
				"workspace_id":   "workspace-scope-b",
				"thread_id":      "thread-from-another-scope",
				"route_id":       "/inbox",
				"surface_id":     "inbox.workspace",
				"source_channel": "telegram",
				"setup_state":    "ready",
			},
		}, "message-admin-scope-mismatch")
	if !handled {
		t.Fatal("expected mismatched admin request to be handled as a governed rejection")
	}
	errPayload, ok := result["error"].(map[string]any)
	if !ok || errPayload["code"] != "agent_context_scope_mismatch" || result["decision"] != "reject" || result["recoverable"] != true {
		t.Fatalf("expected recoverable fail-closed scope mismatch, got %+v", result)
	}
	if result["execution_result"] != nil || result["preview_result"] != nil {
		t.Fatalf("scope mismatch must not execute or preview a skill, got %+v", result)
	}
}

func TestChatAgentPlannerPreservesSettingsAndDataBoundaries(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Planner Settings Data"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var profilePayload struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&profilePayload); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	threadResp := doRequest(t, a, http.MethodPost, "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profilePayload.ID+`","title":"Planner Settings Data"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	chatSvc := chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments"))
	registry := agentskills.NewRegistry(nil)
	contextEnvelope := func(routeID, surfaceID string) map[string]any {
		return map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profilePayload.ID,
				"workspace_id":   "workspace-settings-data",
				"thread_id":      thread.ID,
				"route_id":       routeID,
				"surface_id":     surfaceID,
				"source_channel": "in-app",
				"setup_state":    "ready",
			},
		}
	}

	appearanceResult, handled := dispatchChatAgentProviderPlanner(context.Background(), a.db, chatSvc,
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.settings.update_appearance","parameters":{"setting_key":"theme","setting_scope":"appearance","setting_value":"dark","provider_secret":"sk-settings-must-not-leak"},"message":"I prepared an appearance preview."}`}),
		registry, profilePayload.ID, thread.ID, "update appearance theme to dark", contextEnvelope("/settings/appearance", "settings.appearance.form"), "message-settings-preview")
	if !handled {
		t.Fatal("expected appearance update to enter the Agent planner")
	}
	appearancePreview, ok := appearanceResult["preview_result"].(map[string]any)
	if !ok || appearancePreview["kind"] != "agent_skill_preview" || appearancePreview["mutation_applied"] != false {
		t.Fatalf("expected non-mutating appearance preview, got %+v", appearanceResult)
	}
	var settingsCount int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM profile_settings WHERE profile_id = ? AND key = 'theme'`, profilePayload.ID).Scan(&settingsCount); err != nil {
		t.Fatalf("count theme before confirmation: %v", err)
	}
	if settingsCount != 0 {
		t.Fatalf("appearance preview must not persist theme, count=%d", settingsCount)
	}
	appearanceApply, ok := appearancePreview["apply_contract"].(map[string]any)
	appearanceApplyRequest, requestOK := appearanceApply["request"].(map[string]any)
	if !ok || !requestOK {
		t.Fatalf("appearance preview missing apply contract: %+v", appearancePreview)
	}
	applyBody, err := json.Marshal(appearanceApplyRequest)
	if err != nil {
		t.Fatalf("marshal appearance apply: %v", err)
	}
	if strings.Contains(string(applyBody), "sk-settings-must-not-leak") || strings.Contains(string(applyBody), "provider_secret") {
		t.Fatalf("appearance apply contract leaked provider secret evidence: %s", string(applyBody))
	}
	apply := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(string(applyBody)), map[string]string{"Content-Type": "application/json"})
	if apply.Code != http.StatusOK || !strings.Contains(apply.Body.String(), `"mutation_applied":true`) || !strings.Contains(apply.Body.String(), `"operation":"settings.appearance.update"`) {
		t.Fatalf("confirmed appearance apply failed, status=%d body=%s", apply.Code, apply.Body.String())
	}
	var theme string
	if err := a.db.QueryRow(`SELECT value FROM profile_settings WHERE profile_id = ? AND key = 'theme'`, profilePayload.ID).Scan(&theme); err != nil {
		t.Fatalf("read confirmed theme: %v", err)
	}
	if theme != "dark" {
		t.Fatalf("expected confirmed dark theme, got %q", theme)
	}

	exportResult, handled := dispatchChatAgentProviderPlanner(context.Background(), a.db, chatSvc,
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.data.export_bundle","parameters":{"export_scope":"all"},"message":"Preparing a scoped export summary."}`}),
		registry, profilePayload.ID, thread.ID, "export all Cabinet data", contextEnvelope("/settings/storage", "settings.data.export"), "message-data-export")
	if !handled {
		t.Fatal("expected data export to enter the Agent planner")
	}
	exportExecution, ok := exportResult["execution_result"].(map[string]any)
	if !ok || exportExecution["operation"] != "data.export.bundle" || exportExecution["read_only"] != true || exportExecution["profile_id"] != profilePayload.ID || exportExecution["export_scope"] != "all" {
		t.Fatalf("expected profile-scoped non-mutating export preparation, got %+v", exportResult)
	}
	if exportResult["preview_result"] != nil || exportExecution["mutation_applied"] == true {
		t.Fatalf("export preparation must not use a mutation preview or apply data changes, got %+v", exportResult)
	}

	restoreResult, handled := dispatchChatAgentProviderPlanner(context.Background(), a.db, chatSvc,
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.data.restore_backup","parameters":{"backup_path":"C:/backups/cabinet.db"},"message":"Restoring the backup."}`}),
		registry, profilePayload.ID, thread.ID, "restore Cabinet from this backup", contextEnvelope("/settings/storage", "settings.data.restore"), "message-data-restore")
	if !handled {
		t.Fatal("expected restore request to enter the Agent planner")
	}
	restorePreview, ok := restoreResult["preview_result"].(map[string]any)
	if !ok || restoreResult["decision"] != "select_skill" || restoreResult["confirmation_state"] != "preview_required" || restorePreview["kind"] != "agent_skill_preview" {
		t.Fatalf("destructive restore must create a durable non-mutating preview, got %+v", restoreResult)
	}
	if restorePreview["strong_confirmation_required"] != true || restorePreview["strong_confirmation_endpoint"] != "/api/agent/skills/confirm-destructive" || restorePreview["mutation_applied"] != false {
		t.Fatalf("destructive restore preview lost the dedicated strong-confirmation boundary, got %+v", restorePreview)
	}
	if restoreResult["execution_result"] != nil {
		t.Fatalf("restore request must not execute before strong confirmation, got %+v", restoreResult)
	}
}

func TestChatAgentPlannerRoutesChatActionTimelineSummaryFromMainChat(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Planner Timeline"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	threadResp := doRequest(t, a, http.MethodPost, "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profile.ID+`","title":"Planner Timeline"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	createRun := doRequest(t, a, http.MethodPost, "/api/chat/workflow-runs", strings.NewReader(`{
		"profile_id":"`+profile.ID+`",
		"workflow_id":"chat.app_control.dispatch",
		"capability_id":"cabinet.inventory.search_items",
		"source_channel":"in-app",
		"source_thread_id":"`+thread.ID+`",
		"source_message_id":"message-planner-timeline-source",
		"confirmation_state":"not_required",
		"input":{"query":"private raw planner prompt"},
		"provider_trace":{"api_key":"sk-planner-timeline-secret","preview_id":"preview-planner-timeline-secret"},
		"bulk_items":[{"id":"timeline-step","label":"Search Inventory"}]
	}`), map[string]string{"Content-Type": "application/json"})
	if createRun.Code != http.StatusCreated {
		t.Fatalf("create workflow run status=%d body=%s", createRun.Code, createRun.Body.String())
	}
	var run struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createRun.Body).Decode(&run); err != nil {
		t.Fatalf("decode workflow run: %v", err)
	}
	completeRun := doRequest(t, a, http.MethodPatch, "/api/chat/workflow-runs/"+run.ID, strings.NewReader(`{
		"profile_id":"`+profile.ID+`",
		"status":"completed",
		"confirmation_state":"not_required",
		"result":{"operation":"inventory.item.search","authority_outcome":"apply_allowed","mutation_applied":false,"preview_id":"preview-planner-timeline-secret"}
	}`), map[string]string{"Content-Type": "application/json"})
	if completeRun.Code != http.StatusOK {
		t.Fatalf("complete workflow run status=%d body=%s", completeRun.Code, completeRun.Body.String())
	}

	chatSvc := chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments"))
	result, handled := dispatchChatAgentProviderPlanner(context.Background(), a.db, chatSvc,
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.chat.action_timeline.view","parameters":{"thread_id":"` + thread.ID + `","provider_secret":"sk-planner-timeline-secret"},"message":"Reading this Chat thread action timeline."}`}),
		agentskills.NewRegistry(nil),
		profile.ID,
		thread.ID,
		"show the governed action timeline for this chat",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "fake-planner-model"},
			"agent_context": map[string]any{
				"profile_id":     profile.ID,
				"workspace_id":   "workspace-planner-timeline",
				"thread_id":      thread.ID,
				"route_id":       "/chats",
				"surface_id":     "chats.main",
				"source_channel": "in-app",
				"setup_state":    "ready",
			},
		},
		"message-planner-timeline-request")
	if !handled {
		t.Fatal("expected natural-language timeline request to enter the Agent planner")
	}
	execution, ok := result["execution_result"].(map[string]any)
	if !ok || execution["operation"] != "chat.action_timeline.view" || execution["read_only"] != true || execution["thread_id"] != thread.ID {
		t.Fatalf("expected governed Chat action timeline read result, got %+v", result)
	}
	threadMessage, ok := result["thread_message"].(chat.Message)
	if !ok {
		t.Fatalf("timeline planner result missing trusted assistant thread message: %+v", result)
	}
	agentResponseJSON, err := json.Marshal(threadMessage.Context["agent_response"])
	if err != nil {
		t.Fatalf("marshal persisted timeline Agent response: %v", err)
	}
	var agentResponse chat.AgentResponse
	if err := json.Unmarshal(agentResponseJSON, &agentResponse); err != nil {
		t.Fatalf("decode persisted timeline Agent response: %v", err)
	}
	if agentResponse.ResultSummary == nil || agentResponse.ResultSummary.Kind != "chat_action_timeline" {
		t.Fatalf("timeline response missing typed read result summary: %+v", agentResponse)
	}
	if agentResponse.ResultSummary.Total != 1 || len(agentResponse.ResultSummary.Items) != 1 {
		t.Fatalf("timeline summary must expose one bounded active-thread entry: %+v", agentResponse.ResultSummary)
	}
	item := agentResponse.ResultSummary.Items[0]
	if item.ID != run.ID || item.Title != "cabinet.inventory.search_items" || item.Status != "completed" || item.Category != "inventory.item.search" {
		t.Fatalf("timeline summary item = %+v, want bounded run/capability/status/operation", item)
	}
	summaryJSON, err := json.Marshal(agentResponse.ResultSummary)
	if err != nil {
		t.Fatalf("marshal timeline result summary: %v", err)
	}
	for _, forbidden := range []string{"private raw planner prompt", "sk-planner-timeline-secret", "preview-planner-timeline-secret", "provider_trace", "bulk_items", "source_message_id", "authority_outcome", "mutation_applied", "created_at", "updated_at"} {
		if strings.Contains(string(summaryJSON), forbidden) {
			t.Fatalf("timeline read result summary leaked %q: %s", forbidden, summaryJSON)
		}
	}
}

func assertRecoverablePlannerFailure(t *testing.T, result map[string]any, wantCode, wantSurface string) {
	t.Helper()
	errPayload, _ := result["error"].(map[string]any)
	if errPayload["code"] != wantCode {
		t.Fatalf("expected error code %s, got %+v", wantCode, result)
	}
	if result["recoverable"] != true || strings.TrimSpace(fmt.Sprint(result["next_action"])) == "" {
		t.Fatalf("expected recoverable next-action guidance, got %+v", result)
	}
	if result["source_surface"] != wantSurface {
		t.Fatalf("expected source surface %s, got %+v", wantSurface, result)
	}
	if result["workflow_run"] == nil || result["thread_message"] == nil {
		t.Fatalf("expected persisted workflow and assistant message evidence, got %+v", result)
	}
	body, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal failure result: %v", err)
	}
	if strings.Contains(string(body), "sk-secret") {
		t.Fatalf("recoverable planner failure leaked secret-like provider detail: %s", string(body))
	}
}

func mustResolveSkill(t *testing.T, registry agentskills.Registry, id string) agentskills.Skill {
	t.Helper()
	skill, ok := registry.Resolve(id)
	if !ok {
		t.Fatalf("missing skill %s", id)
	}
	return skill
}

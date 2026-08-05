package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/agentskills"
	"github.com/collectors-tech/cabinet/internal/ai"
	"github.com/collectors-tech/cabinet/internal/chat"
)

type captureAssistantProvider struct {
	req          ai.AssistantTurnRequest
	responseText string
	err          error
}

func (p *captureAssistantProvider) Name() string {
	return "fake"
}

func (p *captureAssistantProvider) RunAssistantTurn(_ context.Context, req ai.AssistantTurnRequest) (ai.AssistantTurnResponse, error) {
	p.req = req
	if p.err != nil {
		return ai.AssistantTurnResponse{
			Provider: "fake",
			Model:    "fake-planner-model",
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
	if _, err := a.db.Exec(`INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, status, priority) VALUES (?, ?, 'AFX', 'Slot', 'PLAN-READ-1', 'Visible Planner Item', 'active', 'medium')`, "planner-read-visible", profileA.ID); err != nil {
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

func TestChatAgentPlannerDeniesExternalWriteSelectionWithoutApproval(t *testing.T) {
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
	if result["decision"] != "reject" || result["recoverable"] != true || result["preview_result"] != nil || result["execution_result"] != nil {
		t.Fatalf("external-write planner selection must be denied without preview/apply, got %+v", result)
	}
	errPayload, _ := result["error"].(map[string]any)
	if errPayload["code"] != "agent_authority_external_write_not_approved" {
		t.Fatalf("expected external-write authority denial, got %+v", result)
	}
	workflowRun, _ := result["workflow_run"].(chat.WorkflowRun)
	if workflowRun.Status != "failed" || workflowRun.Error["code"] != "agent_authority_external_write_not_approved" {
		t.Fatalf("expected failed workflow evidence for denied external write, run=%+v result=%+v", workflowRun, result)
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

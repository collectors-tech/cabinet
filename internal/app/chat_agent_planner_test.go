package app

import (
	"context"
	"encoding/json"
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
}

func (p *captureAssistantProvider) Name() string {
	return "fake"
}

func (p *captureAssistantProvider) RunAssistantTurn(_ context.Context, req ai.AssistantTurnRequest) (ai.AssistantTurnResponse, error) {
	p.req = req
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

func mustResolveSkill(t *testing.T, registry agentskills.Registry, id string) agentskills.Skill {
	t.Helper()
	skill, ok := registry.Resolve(id)
	if !ok {
		t.Fatalf("missing skill %s", id)
	}
	return skill
}

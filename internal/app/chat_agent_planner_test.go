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
	req ai.AssistantTurnRequest
}

func (p *captureAssistantProvider) Name() string {
	return "fake"
}

func (p *captureAssistantProvider) RunAssistantTurn(_ context.Context, req ai.AssistantTurnRequest) (ai.AssistantTurnResponse, error) {
	p.req = req
	return ai.AssistantTurnResponse{
		Provider: "fake",
		Model:    "fake-planner-model",
		Text:     `{"decision":"select_skill","skill_id":"cabinet.inventory.search_items","parameters":{"part_number":"AFX-123"},"message":"Searching inventory for AFX-123."}`,
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
			ID     string `json:"id"`
			Status string `json:"status"`
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
	}
	if !slices.Contains(ids, "cabinet.inventory.search_items") {
		t.Fatalf("available read skill was not exposed to provider: %+v", exposed.Skills)
	}
	if provider.req.Context["agent_context"] == nil {
		t.Fatalf("provider request missing canonical agent context: %+v", provider.req.Context)
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

func mustResolveSkill(t *testing.T, registry agentskills.Registry, id string) agentskills.Skill {
	t.Helper()
	skill, ok := registry.Resolve(id)
	if !ok {
		t.Fatalf("missing skill %s", id)
	}
	return skill
}

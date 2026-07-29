package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/collectors-tech/cabinet/internal/agentskills"
	"github.com/collectors-tech/cabinet/internal/ai"
)

type chatAgentPlannerInput struct {
	ProfileID     string
	ThreadID      string
	Intent        string
	Provider      string
	Model         string
	AgentContext  map[string]any
	Skills        []agentskills.Skill
	SourceChannel string
}

type chatAgentSkillSelection struct {
	Decision      string         `json:"decision"`
	SkillID       string         `json:"skill_id"`
	Parameters    map[string]any `json:"parameters"`
	Message       string         `json:"message"`
	ProviderTrace map[string]string
}

func planChatAgentSkill(ctx context.Context, provider ai.AssistantTurnProvider, input chatAgentPlannerInput) (chatAgentSkillSelection, error) {
	if provider == nil {
		return chatAgentSkillSelection{}, errors.New("assistant provider is required")
	}
	profileID := strings.TrimSpace(input.ProfileID)
	threadID := strings.TrimSpace(input.ThreadID)
	intent := strings.TrimSpace(input.Intent)
	if profileID == "" || threadID == "" || intent == "" {
		return chatAgentSkillSelection{}, errors.New("profile_id, thread_id and intent are required")
	}

	exposedSkills := plannerSkillMetadata(input.Skills)
	if len(exposedSkills) == 0 {
		return chatAgentSkillSelection{}, errors.New("no enabled agent skills available for planning")
	}

	resp, err := provider.RunAssistantTurn(ctx, ai.AssistantTurnRequest{
		ProfileID: profileID,
		ThreadID:  threadID,
		Provider:  strings.TrimSpace(input.Provider),
		Model:     strings.TrimSpace(input.Model),
		Messages: []ai.AssistantTurnMessage{
			{Role: "system", Content: "Return only structured Cabinet planner JSON. Cabinet owns all tool dispatch and mutation authority."},
			{Role: "user", Content: intent},
		},
		Context: map[string]any{
			"agent_context": input.AgentContext,
			"skills":        exposedSkills,
		},
		Metadata: map[string]string{
			"entry_point":              "chat.agent_planner",
			"governed_dispatch_owner":  "cabinet",
			"cabinet_tool_authority":   "none",
			"raw_provider_tools_given": "false",
		},
	})
	trace := plannerProviderTrace(resp)
	if err != nil {
		return chatAgentSkillSelection{ProviderTrace: trace}, err
	}

	var selection chatAgentSkillSelection
	if err := json.Unmarshal([]byte(strings.TrimSpace(resp.Text)), &selection); err != nil {
		return chatAgentSkillSelection{ProviderTrace: trace}, fmt.Errorf("assistant planner returned invalid structured selection: %w", err)
	}
	selection.Decision = strings.TrimSpace(selection.Decision)
	selection.SkillID = strings.TrimSpace(selection.SkillID)
	if selection.Parameters == nil {
		selection.Parameters = map[string]any{}
	}
	selection.ProviderTrace = trace
	if selection.Decision == "" {
		selection.Decision = "clarify"
	}
	if selection.Decision == "select_skill" && !plannerSkillExposed(selection.SkillID, exposedSkills) {
		return chatAgentSkillSelection{ProviderTrace: trace}, fmt.Errorf("assistant planner selected unavailable skill %q", selection.SkillID)
	}
	return selection, nil
}

func plannerSkillMetadata(skills []agentskills.Skill) []map[string]any {
	out := make([]map[string]any, 0, len(skills))
	for _, skill := range skills {
		if !skill.Enabled || !skill.Executable || skill.Status == agentskills.StatusDisabled || skill.Status == agentskills.StatusInvalid {
			continue
		}
		if skill.Status != agentskills.StatusAvailable && skill.Status != agentskills.StatusPreviewOnly {
			continue
		}
		out = append(out, map[string]any{
			"id":                    skill.ID,
			"display_name":          skill.DisplayName,
			"description":           skill.Description,
			"status":                string(skill.Status),
			"safety_level":          string(skill.SafetyLevel),
			"required_context":      skill.RequiredContext,
			"input_schema_refs":     skill.InputSchemaRefs,
			"permissions":           skill.Permissions,
			"audit_behavior":        skill.AuditBehavior,
			"confirmation_required": skill.Permissions.RequiresConfirm,
		})
	}
	return out
}

func plannerSkillExposed(skillID string, exposed []map[string]any) bool {
	for _, skill := range exposed {
		if strings.TrimSpace(fmt.Sprint(skill["id"])) == skillID {
			return true
		}
	}
	return false
}

func plannerProviderTrace(resp ai.AssistantTurnResponse) map[string]string {
	trace := map[string]string{
		"provider":                strings.TrimSpace(resp.Provider),
		"model":                   strings.TrimSpace(resp.Model),
		"governed_dispatch_owner": "cabinet",
		"cabinet_tool_authority":  "none",
	}
	for key, value := range resp.Metadata {
		switch key {
		case "network", "test_provider", "live_provider", "integration_id", "active_auth_method", "error_class":
			if strings.TrimSpace(value) != "" {
				trace[key] = strings.TrimSpace(value)
			}
		}
	}
	return trace
}

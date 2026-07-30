package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/collectors-tech/cabinet/internal/agentskills"
	"github.com/collectors-tech/cabinet/internal/ai"
	"github.com/collectors-tech/cabinet/internal/chat"
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
			"safety_declaration":    plannerSkillSafetyDeclaration(skill),
			"required_context":      skill.RequiredContext,
			"required_actions":      plannerSkillRequiredActions(skill),
			"capabilities":          skill.Capabilities,
			"input_schema_refs":     skill.InputSchemaRefs,
			"input_schema":          plannerSkillInputSchema(skill),
			"permissions":           skill.Permissions,
			"audit_behavior":        skill.AuditBehavior,
			"confirmation_required": skill.Permissions.RequiresConfirm,
		})
	}
	return out
}

func plannerSkillRequiredActions(skill agentskills.Skill) []string {
	if len(skill.RequiredActions) > 0 {
		return append([]string{}, skill.RequiredActions...)
	}
	return append([]string{}, skill.Capabilities...)
}

func plannerSkillInputSchema(skill agentskills.Skill) map[string]any {
	properties := map[string]any{}
	required := []string{}
	for _, ref := range skill.InputSchemaRefs {
		key := strings.TrimSpace(ref)
		if key == "" {
			continue
		}
		properties[key] = plannerSchemaPropertyForKey(key)
		required = append(required, key)
	}
	if len(properties) == 0 && skill.SafetyLevel == agentskills.SafetyReadOnly && strings.Contains(skill.ID, ".search") {
		properties["query"] = map[string]any{
			"type":        "string",
			"description": "Optional user search text, identifier, part number, or filter term.",
		}
	}
	schema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties":           properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func plannerSchemaPropertyForKey(key string) map[string]any {
	lower := strings.ToLower(key)
	property := map[string]any{"type": "string", "description": "Structured planner input for " + key + "."}
	switch {
	case strings.Contains(lower, "id"):
		property["description"] = "Stable Cabinet identifier for " + key + "."
	case strings.Contains(lower, "price") || strings.Contains(lower, "quantity") || strings.Contains(lower, "count"):
		property["type"] = []string{"number", "string"}
		property["description"] = "Numeric or formatted value for " + key + "."
	case strings.Contains(lower, "confirm"):
		property["type"] = "boolean"
		property["description"] = "Explicit confirmation flag for " + key + "."
	}
	return property
}

func plannerSkillSafetyDeclaration(skill agentskills.Skill) map[string]any {
	return map[string]any{
		"side_effect_level":      string(skill.SafetyLevel),
		"local_read":             skill.Permissions.LocalRead,
		"local_write":            skill.Permissions.LocalWrite,
		"external_read":          skill.Permissions.ExternalRead,
		"external_write":         skill.Permissions.ExternalWrite,
		"secret_access":          skill.Permissions.SecretAccess,
		"destructive":            skill.Permissions.Destructive,
		"confirmation_required":  skill.Permissions.RequiresConfirm,
		"preview_only":           skill.SafetyLevel == agentskills.SafetyPreviewOnly,
		"cabinet_dispatch_owner": "cabinet",
	}
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

func dispatchChatAgentProviderPlanner(ctx context.Context, chatSvc *chat.Service, providers *ai.AssistantProviderRegistry, registry agentskills.Registry, profileID, threadID, content string, envelope map[string]any, sourceMessageID string) (map[string]any, bool) {
	if !chatMessageNeedsNaturalLanguageAgentPlanning(content) {
		return nil, false
	}
	assistantContext, _ := envelope["assistant"].(map[string]any)
	providerID := strings.ToLower(strings.TrimSpace(fmt.Sprint(assistantContext["provider"])))
	if providerID == "" || providerID == "<nil>" {
		providerID = "openai"
	}
	provider, ok := providers.Provider(providerID)
	if !ok {
		return map[string]any{
			"error":       map[string]any{"code": "assistant_provider_unavailable", "message": "The selected assistant provider is not available for Chat planning."},
			"next_action": "Choose a configured assistant provider before retrying this natural-language request.",
		}, true
	}
	model := strings.TrimSpace(fmt.Sprint(assistantContext["model"]))
	if model == "<nil>" {
		model = ""
	}
	selection, err := planChatAgentSkill(ctx, provider, chatAgentPlannerInput{
		ProfileID:    profileID,
		ThreadID:     threadID,
		Intent:       content,
		Provider:     providerID,
		Model:        model,
		AgentContext: agentContextEvidence(envelope),
		Skills:       registry.List(),
	})
	result := map[string]any{
		"mode":           "provider_planner",
		"provider":       providerID,
		"source_msg_id":  sourceMessageID,
		"provider_trace": selection.ProviderTrace,
	}
	status := "completed"
	confirmationState := "not_required"
	var runError map[string]any
	if err != nil {
		status = "failed"
		runError = map[string]any{"code": "assistant_planner_failed", "message": "Assistant planning did not return a usable governed skill selection."}
		result["error"] = runError
		result["next_action"] = "Review assistant provider setup and retry the request."
	} else {
		result["decision"] = selection.Decision
		result["skill_id"] = selection.SkillID
		result["parameters"] = selection.Parameters
		result["message"] = selection.Message
	}
	if chatSvc != nil {
		run, runErr := chatSvc.CreateWorkflowRun(ctx, chat.CreateWorkflowRunInput{
			ProfileID:         profileID,
			WorkflowID:        "chat.agent_planner.dispatch",
			CapabilityID:      "assistant.agent_planner",
			SourceChannel:     "in_app_chat",
			SourceThreadID:    threadID,
			SourceMessageID:   sourceMessageID,
			Input:             map[string]any{"content": content, "agent_context": agentContextEvidence(envelope)},
			ProviderTrace:     stringMapToAny(selection.ProviderTrace),
			ConfirmationState: confirmationState,
		})
		if runErr == nil {
			updated, updateErr := chatSvc.UpdateWorkflowRun(ctx, chat.UpdateWorkflowRunInput{
				ProfileID:         profileID,
				RunID:             run.ID,
				Status:            status,
				Result:            result,
				Error:             runError,
				ProviderTrace:     stringMapToAny(selection.ProviderTrace),
				ConfirmationState: confirmationState,
			})
			if updateErr == nil {
				result["workflow_run"] = updated
			}
		}
		messageText := strings.TrimSpace(selection.Message)
		if messageText == "" {
			if err != nil {
				messageText = "I could not complete provider-backed planning for that request. Review assistant provider setup, then retry."
			} else {
				messageText = "I selected a governed Cabinet skill for this request. Cabinet still controls dispatch and confirmation."
			}
		}
		if assistantMessage, messageErr := chatSvc.CreateMessage(ctx, profileID, threadID, "assistant", messageText, map[string]any{"agent_planner": result}); messageErr == nil {
			result["thread_message"] = assistantMessage
		}
	}
	return result, true
}

func chatMessageNeedsNaturalLanguageAgentPlanning(content string) bool {
	normalized := normalizePlannerText(content)
	if normalized == "" {
		return false
	}
	return (strings.Contains(normalized, "find") ||
		strings.Contains(normalized, "search") ||
		strings.Contains(normalized, "look up") ||
		strings.Contains(normalized, "lookup")) &&
		(strings.Contains(normalized, "item") ||
			strings.Contains(normalized, "inventory") ||
			strings.Contains(normalized, "wishlist") ||
			strings.Contains(normalized, "part number"))
}

func stringMapToAny(values map[string]string) map[string]any {
	out := make(map[string]any, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}

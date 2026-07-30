package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
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
	ErrorCode     string         `json:"error_code"`
	NextAction    string         `json:"next_action"`
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
	return normalizeChatAgentSkillSelection(selection, exposedSkills), nil
}

func normalizeChatAgentSkillSelection(selection chatAgentSkillSelection, exposedSkills []map[string]any) chatAgentSkillSelection {
	switch selection.Decision {
	case "select_skill":
		if selection.SkillID == "" {
			return plannerClarification("skill_required", "I need a supported Cabinet skill target before planning that request.", "Choose a supported Cabinet skill or ask a more specific Cabinet question.", selection)
		}
		if !plannerSkillExposed(selection.SkillID, exposedSkills) {
			return plannerRejection("skill_unavailable", "That Cabinet skill is disabled, unavailable, or not exposed for this request.", "Choose an enabled Cabinet skill or adjust Agent Skill settings before retrying.", selection)
		}
		if strings.TrimSpace(selection.Message) == "" {
			selection.Message = "I selected a governed Cabinet skill for this request. Cabinet still controls dispatch and confirmation."
		}
		return selection
	case "clarify":
		if selection.ErrorCode == "" {
			selection.ErrorCode = "clarification_required"
		}
		if strings.TrimSpace(selection.Message) == "" {
			selection.Message = "I need one more detail before selecting a Cabinet skill."
		}
		if strings.TrimSpace(selection.NextAction) == "" {
			selection.NextAction = "Provide the missing identifier, target, or intent and retry the request."
		}
		return selection
	case "reject", "unsupported":
		return plannerRejection("unsupported_request", "Cabinet cannot safely plan that request as a supported skill selection.", "Ask for a supported Cabinet read or previewable local action.", selection)
	default:
		return plannerRejection("unsupported_planner_decision", "Cabinet rejected an unsupported planner decision before any work was applied.", "Retry with a supported Cabinet request; no action was completed.", selection)
	}
}

func plannerClarification(code, message, nextAction string, selection chatAgentSkillSelection) chatAgentSkillSelection {
	selection.Decision = "clarify"
	selection.ErrorCode = code
	selection.Message = message
	selection.NextAction = nextAction
	return selection
}

func plannerRejection(code, message, nextAction string, selection chatAgentSkillSelection) chatAgentSkillSelection {
	selection.Decision = "reject"
	selection.ErrorCode = code
	selection.Message = message
	selection.NextAction = nextAction
	selection.Parameters = map[string]any{}
	return selection
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

func dispatchChatAgentProviderPlanner(ctx context.Context, conn *sql.DB, chatSvc *chat.Service, providers *ai.AssistantProviderRegistry, registry agentskills.Registry, profileID, threadID, content string, envelope map[string]any, sourceMessageID string) (map[string]any, bool) {
	if !chatMessageNeedsNaturalLanguageAgentPlanning(content) {
		return nil, false
	}
	assistantContext, _ := envelope["assistant"].(map[string]any)
	providerID := strings.ToLower(strings.TrimSpace(fmt.Sprint(assistantContext["provider"])))
	if providerID == "" || providerID == "<nil>" {
		providerID = "openai"
	}
	agentCtx := agentContextEvidence(envelope)
	provider, ok := providers.Provider(providerID)
	if !ok {
		return map[string]any{
			"mode":           "provider_planner",
			"provider":       providerID,
			"source_msg_id":  sourceMessageID,
			"source_surface": strings.TrimSpace(fmt.Sprint(agentCtx["surface_id"])),
			"source_channel": strings.TrimSpace(fmt.Sprint(agentCtx["source_channel"])),
			"recoverable":    true,
			"error":          map[string]any{"code": "assistant_provider_unavailable", "message": "The selected assistant provider is not available for Chat planning."},
			"next_action":    "Choose a configured assistant provider before retrying this natural-language request.",
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
		AgentContext: agentCtx,
		Skills:       registry.List(),
	})
	result := map[string]any{
		"mode":           "provider_planner",
		"provider":       providerID,
		"source_msg_id":  sourceMessageID,
		"source_surface": strings.TrimSpace(fmt.Sprint(agentCtx["surface_id"])),
		"source_channel": strings.TrimSpace(fmt.Sprint(agentCtx["source_channel"])),
		"provider_trace": selection.ProviderTrace,
	}
	status := "completed"
	confirmationState := "not_required"
	var runError map[string]any
	if err != nil {
		status = "failed"
		runError = map[string]any{"code": "assistant_planner_failed", "message": "Assistant planning did not return a usable governed skill selection."}
		result["recoverable"] = true
		result["error"] = runError
		result["next_action"] = "Review assistant provider setup and retry the request."
	} else {
		result["decision"] = selection.Decision
		result["skill_id"] = selection.SkillID
		result["parameters"] = redactPlannerEvidenceMap(selection.Parameters)
		result["message"] = selection.Message
		if selection.ErrorCode != "" {
			result["error"] = map[string]any{"code": selection.ErrorCode, "message": selection.Message}
		}
		if selection.NextAction != "" {
			result["next_action"] = selection.NextAction
		}
		if executionResult, authority, execErr := executeReadOnlyPlannerSelection(ctx, conn, chatSvc, registry, profileID, threadID, selection, envelope, sourceMessageID); execErr != nil {
			status = "failed"
			runError = map[string]any{"code": "planner_read_only_execution_failed", "message": "Cabinet could not execute the selected read-only skill."}
			result["recoverable"] = true
			result["error"] = runError
			result["next_action"] = "Retry after checking the selected skill requirements and active profile context."
			if authority.SkillID != "" {
				result["authority"] = authority
			}
		} else if executionResult != nil {
			result["execution_result"] = executionResult
			result["authority"] = authority
		} else if previewResult, authority, previewErr := previewLocalWritePlannerSelection(ctx, chatSvc, registry, profileID, threadID, selection, envelope, sourceMessageID); previewErr != nil {
			status = "failed"
			runError = map[string]any{"code": "planner_preview_failed", "message": "Cabinet could not create a confirmation preview for the selected skill."}
			result["recoverable"] = true
			if detail := strings.TrimSpace(previewErr.Error()); detail != "" {
				runError["detail"] = detail
			}
			result["error"] = runError
			result["next_action"] = "Review the selected skill requirements and retry before confirming any local change."
			if authority.SkillID != "" {
				result["authority"] = authority
			}
		} else if previewResult != nil {
			if previewResult["decision"] == "clarify" {
				result["decision"] = "clarify"
				result["message"] = previewResult["message"]
				result["error"] = previewResult["error"]
				result["next_action"] = previewResult["next_action"]
				result["clarification"] = previewResult["clarification"]
				result["missing_context"] = previewResult["missing_context"]
			} else {
				confirmationState = "preview_required"
				result["preview_result"] = previewResult
				result["authority"] = authority
				result["confirmation_state"] = confirmationState
			}
		} else if denialResult, authority := denyPlannerSelectionWithoutSupportedDispatch(registry, profileID, threadID, selection, envelope, sourceMessageID); denialResult != nil {
			status = "failed"
			result["decision"] = denialResult["decision"]
			result["message"] = denialResult["message"]
			result["error"] = denialResult["error"]
			result["next_action"] = denialResult["next_action"]
			result["recoverable"] = true
			if authority.SkillID != "" {
				result["authority"] = authority
			}
			if errPayload, _ := denialResult["error"].(map[string]any); errPayload != nil {
				runError = errPayload
			}
		}
	}
	evidence := plannerWorkflowEvidence(providerID, selection, agentCtx, confirmationState, status, result, runError)
	result["evidence"] = evidence
	if chatSvc != nil {
		run, runErr := chatSvc.CreateWorkflowRun(ctx, chat.CreateWorkflowRunInput{
			ProfileID:         profileID,
			WorkflowID:        "chat.agent_planner.dispatch",
			CapabilityID:      "assistant.agent_planner",
			SourceChannel:     "in_app_chat",
			SourceThreadID:    threadID,
			SourceMessageID:   sourceMessageID,
			Input:             map[string]any{"content": content, "agent_context": agentCtx, "evidence": evidence},
			ProviderTrace:     stringMapToAny(selection.ProviderTrace),
			ConfirmationState: confirmationState,
			BulkItems:         plannerWorkflowEvidenceSteps(evidence),
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
				BulkItems:         plannerWorkflowEvidenceSteps(evidence),
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

func dispatchChatAgentCapabilityExplanation(ctx context.Context, chatSvc *chat.Service, profileID, threadID, content string, envelope map[string]any, sourceMessageID string, explanation agentCapabilityExplanationResponse) (map[string]any, bool) {
	if !chatMessageRequestsAgentCapabilityExplanation(content) {
		return nil, false
	}
	agentCtx := agentContextEvidence(envelope)
	summary := summarizeAgentCapabilityExplanation(explanation.Capabilities)
	result := map[string]any{
		"mode":           "capability_explanation",
		"profile_id":     profileID,
		"source_msg_id":  sourceMessageID,
		"source_surface": strings.TrimSpace(fmt.Sprint(agentCtx["surface_id"])),
		"source_channel": strings.TrimSpace(fmt.Sprint(agentCtx["source_channel"])),
		"capability_id":  "assistant.agent_capability_explanation",
		"summary":        summary,
		"explanation":    explanation,
		"evidence": map[string]any{
			"entry_point":              "chat.agent_capability_explanation",
			"capability_id":            "assistant.agent_capability_explanation",
			"source_surface":           strings.TrimSpace(fmt.Sprint(agentCtx["surface_id"])),
			"source_channel":           strings.TrimSpace(fmt.Sprint(agentCtx["source_channel"])),
			"agent_context":            agentCtx,
			"governed_dispatch_owner":  "cabinet",
			"cabinet_tool_authority":   "none",
			"raw_provider_tools_given": "false",
		},
	}
	if chatSvc != nil {
		run, runErr := chatSvc.CreateWorkflowRun(ctx, chat.CreateWorkflowRunInput{
			ProfileID:         profileID,
			WorkflowID:        "chat.agent_capability_explanation",
			CapabilityID:      "assistant.agent_capability_explanation",
			SourceChannel:     "in_app_chat",
			SourceThreadID:    threadID,
			SourceMessageID:   sourceMessageID,
			Input:             map[string]any{"content": content, "agent_context": agentCtx},
			ConfirmationState: "not_required",
		})
		if runErr == nil {
			updated, updateErr := chatSvc.UpdateWorkflowRun(ctx, chat.UpdateWorkflowRunInput{
				ProfileID:         profileID,
				RunID:             run.ID,
				Status:            "completed",
				Result:            result,
				ConfirmationState: "not_required",
			})
			if updateErr == nil {
				result["workflow_run"] = updated
			}
		}
		messageText := "Cabinet Agent can explain available skills, setup needs, and confirmation boundaries from the active profile."
		if assistantMessage, messageErr := chatSvc.CreateMessage(ctx, profileID, threadID, "assistant", messageText, map[string]any{"agent_capabilities": result}); messageErr == nil {
			result["thread_message"] = assistantMessage
		}
	}
	return result, true
}

func summarizeAgentCapabilityExplanation(capabilities []agentCapabilityExplanation) map[string]any {
	counts := map[string]int{}
	for _, capability := range capabilities {
		counts[capability.CapabilityState]++
	}
	return map[string]any{
		"total":             len(capabilities),
		"available":         counts["available"],
		"confirm_required":  counts["confirm_required"],
		"blocked_by_policy": counts["blocked_by_policy"],
		"setup_required":    counts["setup_required"],
		"disabled":          counts["disabled"],
		"unavailable":       counts["unavailable"],
	}
}

func executeReadOnlyPlannerSelection(ctx context.Context, conn *sql.DB, chatSvc *chat.Service, registry agentskills.Registry, profileID, threadID string, selection chatAgentSkillSelection, envelope map[string]any, sourceMessageID string) (map[string]any, agentskills.AgentAuthorityReview, error) {
	if selection.Decision != "select_skill" || selection.ErrorCode != "" {
		return nil, agentskills.AgentAuthorityReview{}, nil
	}
	skill, ok := registry.Resolve(selection.SkillID)
	if !ok || skill.SafetyLevel != agentskills.SafetyReadOnly {
		return nil, agentskills.AgentAuthorityReview{}, nil
	}
	agentCtx := agentContextEvidence(envelope)
	req := agentskills.PreviewRequest{
		SkillID:         selection.SkillID,
		ProfileID:       profileID,
		Confirm:         false,
		SourceSurface:   strings.TrimSpace(fmt.Sprint(agentCtx["surface_id"])),
		SourceChannel:   strings.TrimSpace(fmt.Sprint(agentCtx["source_channel"])),
		SourceThreadID:  threadID,
		SourceMessageID: sourceMessageID,
		AgentContext:    agentCtx,
		Parameters:      selection.Parameters,
	}
	req = normalizeAgentSkillContextRequest(req)
	hydratePlannerSelectedRecordParams(req.Parameters, agentCtx)
	if plannerSkillNeedsSelectedTarget(skill) {
		if clarification, ok := agentSkillContextClarification(registry, req); ok {
			return map[string]any{
				"decision":        "clarify",
				"message":         "I need the target Cabinet record before previewing that change.",
				"error":           map[string]any{"code": "missing_context", "message": "Cabinet needs selected-record context before creating this preview."},
				"missing_context": clarification["missing_context"],
				"clarification":   clarification["clarification"],
				"next_action":     clarification["next_action"],
			}, agentskills.AgentAuthorityReview{}, nil
		}
	}
	review, err := registry.ReviewAuthority(agentSkillAuthorityRequest(req), agentskills.AgentAuthorityPolicy{
		ProfileID:  profileID,
		Mode:       agentskills.AgentAuthorityAskBeforeLocalChanges,
		EntryPoint: "chat.agent_planner",
	})
	if err != nil {
		return nil, review, err
	}
	if !review.ApplyAllowed {
		return nil, review, fmt.Errorf("read-only planner selection not allowed: %s", review.Blocker)
	}
	result, blocker, err := applyAgentSkill(ctx, conn, chatSvc, selection.SkillID, profileID, selection.Parameters)
	if err != nil {
		if blocker != "" {
			review.Blocker = blocker
		}
		return nil, review, err
	}
	return result, review, nil
}

func previewLocalWritePlannerSelection(ctx context.Context, chatSvc *chat.Service, registry agentskills.Registry, profileID, threadID string, selection chatAgentSkillSelection, envelope map[string]any, sourceMessageID string) (map[string]any, agentskills.AgentAuthorityReview, error) {
	if chatSvc == nil || selection.Decision != "select_skill" || selection.ErrorCode != "" {
		return nil, agentskills.AgentAuthorityReview{}, nil
	}
	skill, ok := registry.Resolve(selection.SkillID)
	if !ok || !skill.Permissions.LocalWrite || skill.Permissions.ExternalWrite || skill.Permissions.Destructive {
		return nil, agentskills.AgentAuthorityReview{}, nil
	}
	action := plannerChatActionForSkill(selection.SkillID)
	if action == "" {
		return nil, agentskills.AgentAuthorityReview{}, nil
	}
	agentCtx := agentContextEvidence(envelope)
	req := agentskills.PreviewRequest{
		SkillID:         selection.SkillID,
		ProfileID:       profileID,
		Confirm:         false,
		SourceSurface:   strings.TrimSpace(fmt.Sprint(agentCtx["surface_id"])),
		SourceChannel:   strings.TrimSpace(fmt.Sprint(agentCtx["source_channel"])),
		SourceThreadID:  threadID,
		SourceMessageID: sourceMessageID,
		AgentContext:    agentCtx,
		Parameters:      selection.Parameters,
	}
	req = normalizeAgentSkillContextRequest(req)
	hydratePlannerSelectedRecordParams(req.Parameters, agentCtx)
	if plannerSkillNeedsSelectedTarget(skill) {
		if clarification, ok := agentSkillContextClarification(registry, req); ok {
			return map[string]any{
				"decision":        "clarify",
				"message":         "I need the target Cabinet record before previewing that change.",
				"error":           map[string]any{"code": "missing_context", "message": "Cabinet needs selected-record context before creating this preview."},
				"missing_context": clarification["missing_context"],
				"clarification":   clarification["clarification"],
				"next_action":     clarification["next_action"],
			}, agentskills.AgentAuthorityReview{}, nil
		}
	}
	review, err := registry.ReviewAuthority(agentSkillAuthorityRequest(req), agentskills.AgentAuthorityPolicy{
		ProfileID:  profileID,
		Mode:       agentskills.AgentAuthorityAskBeforeLocalChanges,
		EntryPoint: "chat.agent_planner",
	})
	if err != nil {
		return nil, review, err
	}
	if !review.PreviewAllowed {
		return nil, review, fmt.Errorf("planner preview not allowed: %s", review.Blocker)
	}
	preview, err := chatSvc.PreviewAction(ctx, chat.PreviewActionInput{
		ProfileID: profileID,
		ThreadID:  threadID,
		Action:    action,
		Payload:   plannerNonSecretActionPayload(req.Parameters),
	})
	if err != nil {
		return nil, review, fmt.Errorf("create chat action preview for %s: %w", action, err)
	}
	return map[string]any{
		"preview_id":            preview.ID,
		"action":                preview.Action,
		"capability_id":         preview.CapabilityID,
		"status":                preview.Status,
		"payload":               preview.Payload,
		"confirmation_required": true,
		"mutation_applied":      false,
		"next_action":           "Review the preview and confirm through the existing Chat confirmation endpoint before Cabinet applies this local change.",
	}, review, nil
}

func denyPlannerSelectionWithoutSupportedDispatch(registry agentskills.Registry, profileID, threadID string, selection chatAgentSkillSelection, envelope map[string]any, sourceMessageID string) (map[string]any, agentskills.AgentAuthorityReview) {
	if selection.Decision != "select_skill" || selection.ErrorCode != "" {
		return nil, agentskills.AgentAuthorityReview{}
	}
	skill, ok := registry.Resolve(selection.SkillID)
	if !ok || skill.SafetyLevel == agentskills.SafetyReadOnly {
		return nil, agentskills.AgentAuthorityReview{}
	}
	if skill.Permissions.LocalWrite && !skill.Permissions.ExternalWrite && !skill.Permissions.Destructive && plannerChatActionForSkill(selection.SkillID) != "" {
		return nil, agentskills.AgentAuthorityReview{}
	}
	agentCtx := agentContextEvidence(envelope)
	req := agentskills.PreviewRequest{
		SkillID:         selection.SkillID,
		ProfileID:       profileID,
		Confirm:         false,
		SourceSurface:   strings.TrimSpace(fmt.Sprint(agentCtx["surface_id"])),
		SourceChannel:   strings.TrimSpace(fmt.Sprint(agentCtx["source_channel"])),
		SourceThreadID:  threadID,
		SourceMessageID: sourceMessageID,
		AgentContext:    agentCtx,
		Parameters:      selection.Parameters,
	}
	req = normalizeAgentSkillContextRequest(req)
	hydratePlannerSelectedRecordParams(req.Parameters, agentCtx)
	review, err := registry.ReviewAuthority(agentSkillAuthorityRequest(req), agentskills.AgentAuthorityPolicy{
		ProfileID:  profileID,
		Mode:       agentskills.AgentAuthorityAskBeforeLocalChanges,
		EntryPoint: "chat.agent_planner",
	})
	blocker := "planner_skill_dispatch_not_supported"
	nextAction := "Choose a supported Cabinet read or previewable local action before retrying this planner request."
	if err != nil {
		blocker = "planner_authority_review_failed"
		nextAction = "Retry after checking Agent Skill configuration and active profile context."
	} else {
		if strings.TrimSpace(review.Blocker) != "" {
			blocker = strings.TrimSpace(review.Blocker)
		}
		if strings.TrimSpace(review.NextAction) != "" {
			nextAction = strings.TrimSpace(review.NextAction)
		}
	}
	return map[string]any{
		"decision":    "reject",
		"message":     "Cabinet blocked this planner selection before any unsupported, external, or destructive action was previewed.",
		"error":       map[string]any{"code": blocker, "message": "The selected Agent Skill is not approved for this Chat planner dispatch path."},
		"next_action": nextAction,
	}, review
}

func plannerWorkflowEvidence(providerID string, selection chatAgentSkillSelection, agentCtx map[string]any, confirmationState, status string, result map[string]any, runError map[string]any) map[string]any {
	previewID := ""
	action := ""
	mutationApplied := false
	if previewResult, _ := result["preview_result"].(map[string]any); previewResult != nil {
		previewID = strings.TrimSpace(fmt.Sprint(previewResult["preview_id"]))
		action = strings.TrimSpace(fmt.Sprint(previewResult["action"]))
		if applied, ok := previewResult["mutation_applied"].(bool); ok {
			mutationApplied = applied
		}
	}
	if executionResult, _ := result["execution_result"].(map[string]any); executionResult != nil {
		if applied, ok := executionResult["mutation_applied"].(bool); ok {
			mutationApplied = applied
		}
	}
	errCode := ""
	if runError != nil {
		errCode = strings.TrimSpace(fmt.Sprint(runError["code"]))
	}
	if errCode == "" {
		if errPayload, _ := result["error"].(map[string]any); errPayload != nil {
			errCode = strings.TrimSpace(fmt.Sprint(errPayload["code"]))
		}
	}
	tokenState := map[string]any{
		"confirmation_state": confirmationState,
		"apply_state":        "not_applicable",
		"mutation_applied":   mutationApplied,
	}
	if previewID != "" && previewID != "<nil>" {
		tokenState["preview_id"] = previewID
		tokenState["action"] = action
		tokenState["apply_state"] = "pending_explicit_confirmation"
		tokenState["apply_endpoint"] = "/api/chat/actions/apply"
	}
	evidence := map[string]any{
		"provider":                  providerID,
		"entry_point":               "chat.agent_planner",
		"governed_dispatch_owner":   "cabinet",
		"raw_provider_payload_kept": false,
		"selected_skill":            strings.TrimSpace(selection.SkillID),
		"decision":                  strings.TrimSpace(selection.Decision),
		"context":                   plannerContextEvidenceSummary(agentCtx),
		"parameters":                plannerParameterEvidence(selection.Parameters),
		"preview_apply_token_state": tokenState,
		"final_outcome": map[string]any{
			"status":           status,
			"error_code":       errCode,
			"recoverable":      result["recoverable"] == true,
			"confirmation":     confirmationState,
			"mutation_applied": mutationApplied,
		},
	}
	if selection.ProviderTrace != nil {
		evidence["provider_trace"] = stringMapToAny(selection.ProviderTrace)
	}
	return evidence
}

func plannerWorkflowEvidenceSteps(evidence map[string]any) []map[string]any {
	if len(evidence) == 0 {
		return nil
	}
	steps := []map[string]any{
		{
			"id":      "provider-planner",
			"command": "assistant.provider.plan",
			"status":  "completed",
			"evidence": map[string]any{
				"provider":    evidence["provider"],
				"entry_point": evidence["entry_point"],
				"decision":    evidence["decision"],
			},
		},
		{
			"id":      "cabinet-authority",
			"command": "agent_skill.authority_review",
			"status":  "completed",
			"evidence": map[string]any{
				"selected_skill": evidence["selected_skill"],
				"context":        evidence["context"],
			},
		},
		{
			"id":      "final-outcome",
			"command": "chat.agent_planner.outcome",
			"status":  "completed",
			"evidence": map[string]any{
				"preview_apply_token_state": evidence["preview_apply_token_state"],
				"final_outcome":             evidence["final_outcome"],
			},
		},
	}
	return steps
}

func plannerContextEvidenceSummary(agentCtx map[string]any) map[string]any {
	out := map[string]any{}
	for _, key := range []string{"profile_id", "workspace_id", "thread_id", "route_id", "surface_id", "source_channel", "setup_state", "permission_state", "workflow_run_id"} {
		value := strings.TrimSpace(fmt.Sprint(agentCtx[key]))
		if value != "" && value != "<nil>" {
			out[key] = value
		}
	}
	if selected, _ := agentCtx["selected_record"].(map[string]any); selected != nil {
		summary := map[string]any{}
		for _, key := range []string{"type", "id"} {
			value := strings.TrimSpace(fmt.Sprint(selected[key]))
			if value != "" && value != "<nil>" {
				summary[key] = value
			}
		}
		if len(summary) > 0 {
			out["selected_record"] = summary
		}
	}
	return out
}

func plannerParameterEvidence(params map[string]any) map[string]any {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return map[string]any{
		"keys":            keys,
		"values":          redactPlannerEvidenceMap(params),
		"secret_redacted": containsSecretParameter(params),
	}
}

func redactPlannerEvidenceMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for key, value := range in {
		if plannerSensitiveEvidenceKey(key) {
			out[key] = "[redacted]"
			continue
		}
		out[key] = redactPlannerEvidenceValue(value)
	}
	return out
}

func plannerNonSecretActionPayload(in map[string]any) map[string]any {
	if len(in) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for key, value := range in {
		if plannerSensitiveEvidenceKey(key) {
			continue
		}
		out[key] = value
	}
	return copyActionPayload(out)
}

func redactPlannerEvidenceValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return redactPlannerEvidenceMap(typed)
	case []any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, redactPlannerEvidenceValue(item))
		}
		return out
	case string:
		lower := strings.ToLower(typed)
		if strings.Contains(lower, "sk-") || strings.Contains(lower, "secret") || strings.Contains(lower, "bearer ") {
			return "[redacted]"
		}
		return typed
	default:
		return value
	}
}

func plannerSensitiveEvidenceKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	return strings.Contains(normalized, "secret") ||
		strings.Contains(normalized, "token") ||
		strings.Contains(normalized, "api_key") ||
		strings.Contains(normalized, "password")
}

func hydratePlannerSelectedRecordParams(params map[string]any, agentCtx map[string]any) {
	if params == nil {
		return
	}
	selected, _ := agentCtx["selected_record"].(map[string]any)
	selectedType := strings.TrimSpace(fmt.Sprint(selected["type"]))
	selectedID := strings.TrimSpace(fmt.Sprint(selected["id"]))
	if selectedID != "" && selectedID != "<nil>" {
		putAgentContextParamIfMissing(params, "selected_record_type", selectedType)
		putAgentContextParamIfMissing(params, "selected_record_id", selectedID)
		switch selectedType {
		case "inventory_item", "item":
			putAgentContextParamIfMissing(params, "item_id", selectedID)
		case "media", "media_asset":
			putAgentContextParamIfMissing(params, "media_id", selectedID)
		case "wishlist_entry":
			putAgentContextParamIfMissing(params, "wishlist_entry_id", selectedID)
		case "inbox_notification", "notification":
			putAgentContextParamIfMissing(params, "selected_notification", selectedID)
		case "collection":
			putAgentContextParamIfMissing(params, "collection_name", selectedID)
		}
	}
	if !agentSkillHasAnyParam(params, "item_id") {
		routeID := strings.Trim(strings.TrimSpace(fmt.Sprint(agentCtx["route_id"])), "/")
		parts := strings.Split(routeID, "/")
		if len(parts) >= 2 && parts[0] == "inventory" && strings.TrimSpace(parts[len(parts)-1]) != "" {
			putAgentContextParamIfMissing(params, "item_id", parts[len(parts)-1])
		}
	}
}

func plannerSkillNeedsSelectedTarget(skill agentskills.Skill) bool {
	for _, contextName := range skill.RequiredContext {
		switch strings.TrimSpace(contextName) {
		case "selected_item", "target_item":
			return true
		}
	}
	return false
}

func plannerChatActionForSkill(skillID string) string {
	switch strings.TrimSpace(skillID) {
	case "cabinet.inventory.create_item":
		return "create_inventory_item"
	case "cabinet.inventory.update_item":
		return "update_inventory_item"
	default:
		return ""
	}
}

func chatMessageNeedsNaturalLanguageAgentPlanning(content string) bool {
	normalized := normalizePlannerText(content)
	if normalized == "" {
		return false
	}
	readIntent := (strings.Contains(normalized, "find") ||
		strings.Contains(normalized, "search") ||
		strings.Contains(normalized, "look up") ||
		strings.Contains(normalized, "lookup")) &&
		(strings.Contains(normalized, "item") ||
			strings.Contains(normalized, "inventory") ||
			strings.Contains(normalized, "wishlist") ||
			strings.Contains(normalized, "part number"))
	writeIntent := (strings.Contains(normalized, "create") ||
		strings.Contains(normalized, "add") ||
		strings.Contains(normalized, "rename") ||
		strings.Contains(normalized, "update")) &&
		(strings.Contains(normalized, "item") ||
			strings.Contains(normalized, "inventory") ||
			strings.Contains(normalized, "part number"))
	return readIntent || writeIntent
}

func chatMessageRequestsAgentCapabilityExplanation(content string) bool {
	normalized := normalizePlannerText(content)
	if normalized == "" {
		return false
	}
	return (strings.Contains(normalized, "what can") && (strings.Contains(normalized, "agent") || strings.Contains(normalized, "you do"))) ||
		strings.Contains(normalized, "available skill") ||
		strings.Contains(normalized, "available capabilities") ||
		strings.Contains(normalized, "setup state") ||
		strings.Contains(normalized, "setup required")
}

func stringMapToAny(values map[string]string) map[string]any {
	out := make(map[string]any, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}

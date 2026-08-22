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
	"github.com/collectors-tech/cabinet/internal/collection"
	"github.com/collectors-tech/cabinet/internal/commerce"
	"github.com/collectors-tech/cabinet/internal/discovery"
	"github.com/collectors-tech/cabinet/internal/media"
	"github.com/collectors-tech/cabinet/internal/scanner"
	"github.com/collectors-tech/cabinet/internal/wishlist"
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
	Decision                string         `json:"decision"`
	SkillID                 string         `json:"skill_id"`
	Parameters              map[string]any `json:"parameters"`
	Message                 string         `json:"message"`
	ErrorCode               string         `json:"error_code"`
	NextAction              string         `json:"next_action"`
	ProviderTrace           map[string]string
	ProviderErrorClass      string `json:"-"`
	ProviderSetupNextAction string `json:"-"`
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
		return chatAgentSkillSelection{
			ProviderTrace:           trace,
			ProviderErrorClass:      strings.TrimSpace(resp.ErrorClass),
			ProviderSetupNextAction: strings.TrimSpace(resp.SetupNextAction),
		}, err
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
		selection.Parameters = normalizePlannerSchemaRefParameters(selection.SkillID, selection.Parameters, exposedSkills)
		selection.Parameters = normalizeIntegrationPlannerParameters(selection.SkillID, selection.Parameters)
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

func normalizeIntegrationPlannerParameters(skillID string, parameters map[string]any) map[string]any {
	if skillID != "cabinet.integrations.configure_provider" || len(parameters) == 0 {
		return parameters
	}

	normalized := make(map[string]any, len(parameters)+4)
	for key, value := range parameters {
		if isSensitiveAgentSkillPreviewKey(key) && emptyPlannerOptionalSecretValue(value) {
			continue
		}
		normalized[key] = value
	}
	providerID := strings.ToLower(strings.TrimSpace(fmt.Sprint(normalized["provider_id"])))
	if providerID == "" || providerID == "<nil>" {
		providerID = strings.ToLower(strings.TrimSpace(fmt.Sprint(normalized["provider_name"])))
	}
	providerID = strings.ReplaceAll(providerID, " ", "-")
	if providerID != "" && providerID != "<nil>" {
		normalized["provider_id"] = providerID
	}
	catalogue := strings.ToLower(strings.TrimSpace(fmt.Sprint(normalized["catalogue"])))
	if catalogue == "<nil>" {
		catalogue = ""
	}
	if integrationParameterMentionsPublicCatalogue(catalogue) ||
		integrationParameterMentionsPublicCatalogue(normalized["setup_payload"]) ||
		integrationParameterMentionsPublicCatalogue(normalized["setup_step"]) {
		normalized["setup_payload"] = "public_catalogue"
		normalized["setup_step"] = "public_catalogue"
		normalized["marketplace"] = "public"
		delete(normalized, "provider_name")
		delete(normalized, "catalogue")
		return normalized
	}
	catalogueName := strings.TrimSpace(strings.TrimSuffix(catalogue, " catalogue"))
	setupPayload := strings.ToLower(strings.TrimSpace(fmt.Sprint(normalized["setup_payload"])))
	if setupPayload == "" || setupPayload == "<nil>" {
		setupPayload = catalogueName
		if setupPayload != "" {
			setupPayload += "_catalogue"
		}
	}
	setupPayload = strings.ReplaceAll(setupPayload, " ", "_")
	if setupPayload != "" {
		normalized["setup_payload"] = setupPayload
		if strings.TrimSpace(fmt.Sprint(normalized["setup_step"])) == "" || fmt.Sprint(normalized["setup_step"]) == "<nil>" {
			normalized["setup_step"] = setupPayload
		}
	}
	if catalogueName != "" && (strings.TrimSpace(fmt.Sprint(normalized["marketplace"])) == "" || fmt.Sprint(normalized["marketplace"]) == "<nil>") {
		normalized["marketplace"] = catalogueName
	}
	delete(normalized, "provider_name")
	delete(normalized, "catalogue")
	return normalized
}

func emptyPlannerOptionalSecretValue(value any) bool {
	if value == nil {
		return true
	}
	text, ok := value.(string)
	return ok && strings.TrimSpace(text) == ""
}

func integrationParameterMentionsPublicCatalogue(value any) bool {
	text := strings.ToLower(strings.TrimSpace(fmt.Sprint(value)))
	if text == "" || text == "<nil>" {
		return false
	}
	text = strings.NewReplacer("_", " ", "-", " ").Replace(text)
	text = strings.Join(strings.Fields(text), " ")
	return strings.Contains(text, "public catalogue") || strings.Contains(text, "public catalog")
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
			"id":                         skill.ID,
			"display_name":               skill.DisplayName,
			"description":                skill.Description,
			"status":                     string(skill.Status),
			"safety_level":               string(skill.SafetyLevel),
			"safety_declaration":         plannerSkillSafetyDeclaration(skill),
			"required_context":           skill.RequiredContext,
			"required_actions":           plannerSkillRequiredActions(skill),
			"capabilities":               skill.Capabilities,
			"input_schema_refs":          skill.InputSchemaRefs,
			"optional_input_schema_refs": skill.OptionalInputSchemaRefs,
			"input_schema":               plannerSkillInputSchema(skill),
			"permissions":                skill.Permissions,
			"audit_behavior":             skill.AuditBehavior,
			"confirmation_required":      skill.Permissions.RequiresConfirm,
		})
	}
	return out
}

func plannerSkillRequiredActions(skill agentskills.Skill) []string {
	if len(skill.RequiredActions) > 0 {
		return append([]string{}, skill.RequiredActions...)
	}
	if len(skill.IntegrationWorkflows) > 0 {
		return append([]string{}, skill.IntegrationWorkflows...)
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
	for _, ref := range skill.OptionalInputSchemaRefs {
		key := strings.TrimSpace(ref)
		if key == "" {
			continue
		}
		if _, exists := properties[key]; !exists {
			properties[key] = plannerSchemaPropertyForKey(key)
		}
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

func normalizePlannerSchemaRefParameters(skillID string, parameters map[string]any, exposed []map[string]any) map[string]any {
	if len(parameters) == 0 {
		return map[string]any{}
	}

	normalized := make(map[string]any, len(parameters))
	for key, value := range parameters {
		normalized[key] = value
	}
	for _, skill := range exposed {
		if strings.TrimSpace(fmt.Sprint(skill["id"])) != strings.TrimSpace(skillID) {
			continue
		}
		refs, _ := skill["input_schema_refs"].([]string)
		if optional, ok := skill["optional_input_schema_refs"].([]string); ok {
			refs = append(append([]string{}, refs...), optional...)
		}
		if actions, ok := skill["required_actions"].([]string); ok {
			refs = append(append([]string{}, refs...), actions...)
		}
		for _, ref := range refs {
			ref = strings.TrimSpace(ref)
			wrapped, exists := normalized[ref]
			if ref == "" || !exists {
				continue
			}
			decoded := map[string]any{}
			switch value := wrapped.(type) {
			case string:
				if err := json.Unmarshal([]byte(strings.TrimSpace(value)), &decoded); err != nil {
					continue
				}
			case map[string]any:
				for key, nestedValue := range value {
					decoded[key] = nestedValue
				}
			default:
				continue
			}
			if len(decoded) == 0 {
				continue
			}
			delete(normalized, ref)
			for key, nestedValue := range decoded {
				if _, explicit := normalized[key]; !explicit {
					normalized[key] = nestedValue
				}
			}
		}
		break
	}
	return normalized
}

func plannerProviderTrace(resp ai.AssistantTurnResponse) map[string]string {
	trace := map[string]string{
		"provider":                strings.TrimSpace(resp.Provider),
		"model":                   strings.TrimSpace(resp.Model),
		"governed_dispatch_owner": "cabinet",
		"cabinet_tool_authority":  "none",
	}
	if errorClass := strings.TrimSpace(resp.ErrorClass); errorClass != "" {
		trace["error_class"] = errorClass
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
	intentDomain, needsPlanning := chatAgentIntentDomain(content)
	if !needsPlanning {
		return nil, false
	}
	assistantContext, _ := envelope["assistant"].(map[string]any)
	providerID := strings.ToLower(strings.TrimSpace(fmt.Sprint(assistantContext["provider"])))
	if providerID == "" || providerID == "<nil>" {
		providerID = "openai"
	}
	agentCtx := withoutClientAssertedAdminAuthority(agentContextEvidence(envelope))
	provider, ok := providers.Provider(providerID)
	if !ok {
		result := map[string]any{
			"mode":           "provider_planner",
			"provider":       providerID,
			"intent_domain":  intentDomain,
			"source_msg_id":  sourceMessageID,
			"source_surface": strings.TrimSpace(fmt.Sprint(agentCtx["surface_id"])),
			"source_channel": strings.TrimSpace(fmt.Sprint(agentCtx["source_channel"])),
			"recoverable":    true,
			"error":          map[string]any{"code": "assistant_provider_unavailable", "message": "The selected assistant provider is not available for Chat planning."},
			"next_action":    "Choose a configured assistant provider before retrying this natural-language request.",
		}
		response, _ := chat.NewAgentResponse(
			chat.AgentResponseProviderUnavailable,
			"The selected assistant provider is unavailable. No Cabinet action was completed.",
			content,
			"",
			"Cabinet Agent",
			strings.TrimSpace(fmt.Sprint(agentCtx["surface_id"])),
			strings.TrimSpace(fmt.Sprint(agentCtx["source_channel"])),
		)
		if chatSvc != nil {
			if assistantMessage, messageErr := chatSvc.CreateMessage(ctx, profileID, threadID, "assistant", response.Message, map[string]any{
				"agent_planner":  result,
				"agent_response": response,
			}); messageErr == nil {
				result["thread_message"] = assistantMessage
			}
		}
		return result, true
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
		Skills:       agentPlannerSkillsForSession(ctx, conn, profileID, registry.List()),
	})
	result := map[string]any{
		"mode":           "provider_planner",
		"provider":       providerID,
		"intent_domain":  intentDomain,
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
		failureCode := plannerProviderFailureCode(selection.ProviderErrorClass)
		runError = map[string]any{"code": failureCode, "message": "Assistant planning did not return a usable governed skill selection."}
		result["recoverable"] = true
		result["error"] = runError
		if selection.ProviderSetupNextAction != "" {
			result["setup_next_action"] = selection.ProviderSetupNextAction
			result["next_action"] = plannerProviderSetupGuidance(selection.ProviderSetupNextAction)
		} else {
			result["next_action"] = "Review assistant provider setup and retry the request."
		}
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
		if mismatch := plannerAgentContextScopeMismatch(profileID, threadID, agentCtx); len(mismatch) > 0 {
			status = "failed"
			runError = map[string]any{"code": "agent_context_scope_mismatch", "message": "Cabinet rejected Agent context that did not match the active profile or thread.", "mismatched_fields": mismatch}
			selection.Decision = "reject"
			selection.Message = "I could not use that Agent context because it did not match the active Cabinet profile or thread. Reopen Agent from the intended workspace and retry."
			result["decision"] = selection.Decision
			result["message"] = selection.Message
			result["recoverable"] = true
			result["error"] = runError
			result["next_action"] = "Reopen Agent from the intended Cabinet profile and thread before retrying this request."
		} else if executionResult, authority, execErr := executeReadOnlyPlannerSelection(ctx, conn, chatSvc, registry, profileID, threadID, selection, envelope, sourceMessageID); execErr != nil {
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
		} else if previewResult, authority, previewErr := previewLocalWritePlannerSelection(ctx, conn, chatSvc, registry, profileID, threadID, selection, envelope, sourceMessageID); previewErr != nil {
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
		sourceChannel := strings.TrimSpace(fmt.Sprint(agentCtx["source_channel"]))
		if sourceChannel == "" || sourceChannel == "<nil>" {
			sourceChannel = "in_app_chat"
		}
		run, runErr := chatSvc.CreateWorkflowRun(ctx, chat.CreateWorkflowRunInput{
			ProfileID:         profileID,
			WorkflowID:        "chat.agent_planner.dispatch",
			CapabilityID:      "assistant.agent_planner",
			SourceChannel:     sourceChannel,
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
		agentResponse := plannerAgentResponse(registry, selection, result, status, content, agentCtx)
		if assistantMessage, messageErr := chatSvc.CreateMessage(ctx, profileID, threadID, "assistant", agentResponse.Message, map[string]any{
			"agent_planner":  result,
			"agent_response": agentResponse,
		}); messageErr == nil {
			result["thread_message"] = assistantMessage
		}
	}
	return result, true
}

func plannerAgentResponse(registry agentskills.Registry, selection chatAgentSkillSelection, result map[string]any, status, originalIntent string, agentCtx map[string]any) chat.AgentResponse {
	state := chat.AgentResponseUnsupported
	message := strings.TrimSpace(selection.Message)
	var readResultSummary *chat.AgentResponseResultSummary
	if message == "" {
		message = "Cabinet cannot safely complete that Agent request."
	}
	if status == "failed" {
		state = chat.AgentResponseRetryableFailure
		message = "Cabinet could not complete that Agent request. No action was completed."
	}
	if selection.Decision == "clarify" || strings.EqualFold(strings.TrimSpace(fmt.Sprint(result["decision"])), "clarify") {
		state = chat.AgentResponseClarificationRequired
		message = strings.TrimSpace(fmt.Sprint(result["message"]))
	}
	if selection.Decision == "reject" || selection.Decision == "unsupported" {
		state = chat.AgentResponseUnsupported
	}
	if execution, _ := result["execution_result"].(map[string]any); execution != nil {
		state = chat.AgentResponseReadResult
		readResultSummary = plannerAgentReadResultSummary(selection.SkillID, execution)
		message = strings.TrimSpace(selection.Message)
		if readResultSummary != nil {
			message = plannerAgentReadResultMessage(readResultSummary)
		} else if message == "" {
			message = "Cabinet completed the governed read-only Agent request."
		}
	}
	if _, ok := result["preview_result"].(map[string]any); ok {
		state = chat.AgentResponsePreviewRequired
		message = "Cabinet prepared a preview. Review it before applying any local change."
	}
	if errPayload, _ := result["error"].(map[string]any); errPayload != nil {
		code := strings.TrimSpace(fmt.Sprint(errPayload["code"]))
		switch {
		case plannerProviderSetupRequired(result):
			state = chat.AgentResponseSetupRequired
			message = "The assistant provider needs setup before Cabinet can plan that request. No action was completed."
		case strings.Contains(code, "authority") || strings.Contains(code, "policy") || strings.Contains(code, "dispatch_not_supported"):
			state = chat.AgentResponseAuthorityBlocked
			message = "Cabinet blocked this Agent request at the authority boundary. No action was completed."
		case strings.Contains(code, "preview"):
			state = chat.AgentResponsePreviewFailed
			message = "Cabinet could not create a safe preview. No action was completed."
		case strings.Contains(code, "provider") || strings.Contains(code, "assistant_planner"):
			state = chat.AgentResponseProviderUnavailable
			message = "The assistant provider is unavailable. No Cabinet action was completed."
		case strings.Contains(code, "unsupported") || strings.Contains(code, "skill_unavailable"):
			state = chat.AgentResponseUnsupported
			message = "Cabinet does not support that Agent request. No action was completed."
		}
	}
	skillName := strings.TrimSpace(selection.SkillID)
	if skill, ok := registry.Resolve(selection.SkillID); ok && strings.TrimSpace(skill.DisplayName) != "" {
		skillName = strings.TrimSpace(skill.DisplayName)
	}
	response, responseErr := chat.NewAgentResponse(
		state,
		message,
		originalIntent,
		selection.SkillID,
		skillName,
		strings.TrimSpace(fmt.Sprint(agentCtx["surface_id"])),
		strings.TrimSpace(fmt.Sprint(agentCtx["source_channel"])),
	)
	if responseErr != nil {
		response, _ = chat.NewAgentResponse(chat.AgentResponseUnsupported, "Cabinet rejected an invalid Agent response before any action was completed.", "", selection.SkillID, skillName, strings.TrimSpace(fmt.Sprint(agentCtx["surface_id"])), strings.TrimSpace(fmt.Sprint(agentCtx["source_channel"])))
	}
	if preview, _ := result["preview_result"].(map[string]any); preview != nil {
		previewPayload, _ := preview["payload"].(map[string]any)
		response.Preview = &chat.AgentResponsePreview{
			ID:      strings.TrimSpace(fmt.Sprint(preview["preview_id"])),
			Action:  strings.TrimSpace(fmt.Sprint(preview["action"])),
			Status:  strings.TrimSpace(fmt.Sprint(preview["status"])),
			Payload: plannerNonSecretActionPayload(previewPayload),
		}
	}
	response.ResultSummary = readResultSummary
	return response
}

func plannerAgentReadResultMessage(summary *chat.AgentResponseResultSummary) string {
	if summary == nil {
		return ""
	}
	switch summary.Kind {
	case "inventory_items":
		switch summary.Total {
		case 0:
			return "Cabinet found no matching inventory items."
		case 1:
			return "Cabinet found 1 matching inventory item."
		default:
			return fmt.Sprintf("Cabinet found %d matching inventory items.", summary.Total)
		}
	case "dashboard_activity":
		if summary.Total == 0 {
			return "Cabinet found no Dashboard activity records available for that read-only request."
		}
		return "Cabinet summarised the current Dashboard snapshot with bounded attention signals and recent records."
	case "storage_status":
		return "Cabinet read the current storage and backup status without changing settings."
	case "wishlist_entries":
		switch summary.Total {
		case 0:
			return "Cabinet found no matching wishlist entries."
		case 1:
			return "Cabinet found 1 matching wishlist entry."
		default:
			return fmt.Sprintf("Cabinet found %d matching wishlist entries.", summary.Total)
		}
	case "collections":
		switch summary.Total {
		case 0:
			return "Cabinet found no matching collections."
		case 1:
			return "Cabinet found 1 matching collection record."
		default:
			return fmt.Sprintf("Cabinet found %d matching collection records.", summary.Total)
		}
	case "integration_providers":
		switch summary.Total {
		case 0:
			return "Cabinet found no matching integration providers."
		case 1:
			return "Cabinet found 1 matching integration provider."
		default:
			return fmt.Sprintf("Cabinet found %d matching integration providers.", summary.Total)
		}
	case "data_export_bundle":
		return "Cabinet prepared a bounded data export readiness summary without creating or changing data."
	case "maintenance_safe_check":
		return "Cabinet completed a bounded maintenance safe-check summary without changing data."
	case "inbox_notifications":
		switch summary.Total {
		case 0:
			return "Cabinet found no matching Inbox notifications."
		case 1:
			return "Cabinet found 1 matching Inbox notification."
		default:
			return fmt.Sprintf("Cabinet found %d matching Inbox notifications.", summary.Total)
		}
	case "inbox_unhandled":
		switch summary.Total {
		case 0:
			return "Cabinet found no unhandled Inbox notifications."
		case 1:
			return "Cabinet found 1 unhandled Inbox notification."
		default:
			return fmt.Sprintf("Cabinet found %d unhandled Inbox notifications.", summary.Total)
		}
	case "workspace_users":
		switch summary.Total {
		case 0:
			return "Cabinet found no matching workspace users."
		case 1:
			return "Cabinet found 1 matching workspace user."
		default:
			return fmt.Sprintf("Cabinet found %d matching workspace users.", summary.Total)
		}
	case "media_assets":
		switch summary.Total {
		case 0:
			return "Cabinet found no matching media assets."
		case 1:
			return "Cabinet found 1 matching media asset."
		default:
			return fmt.Sprintf("Cabinet found %d matching media assets.", summary.Total)
		}
	case "purchase_orders":
		switch summary.Total {
		case 0:
			return "Cabinet found no matching purchase orders."
		case 1:
			return "Cabinet found 1 matching purchase order."
		default:
			return fmt.Sprintf("Cabinet found %d matching purchase orders.", summary.Total)
		}
	case "market_watch_watches":
		switch summary.Total {
		case 0:
			return "Cabinet found no matching Market Watch saved watches."
		case 1:
			return "Cabinet found 1 matching Market Watch saved watch."
		default:
			return fmt.Sprintf("Cabinet found %d matching Market Watch saved watches.", summary.Total)
		}
	case "discovery_results":
		switch summary.Total {
		case 0:
			return "Cabinet found no matching discovery results."
		case 1:
			return "Cabinet found 1 matching discovery result."
		default:
			return fmt.Sprintf("Cabinet found %d matching discovery results.", summary.Total)
		}
	case "chat_action_timeline":
		switch summary.Total {
		case 0:
			return "Cabinet found no governed action timeline entries for this Chat thread."
		case 1:
			return "Cabinet found 1 governed action timeline entry for this Chat thread."
		default:
			return fmt.Sprintf("Cabinet found %d governed action timeline entries for this Chat thread.", summary.Total)
		}
	default:
		return "Cabinet completed the governed read-only Agent request."
	}
}

const (
	plannerAgentReadResultItemLimit = 5
	plannerAgentReadResultTextLimit = 160
)

func plannerAgentReadResultSummary(skillID string, execution map[string]any) *chat.AgentResponseResultSummary {
	if execution == nil {
		return nil
	}
	switch strings.TrimSpace(skillID) {
	case "cabinet.chat.action_timeline.view":
		return plannerAgentChatActionTimelineReadResultSummary(execution)
	case "cabinet.inventory.search_items":
		return plannerAgentInventoryReadResultSummary(execution)
	case "cabinet.dashboard.summarise_activity":
		return plannerAgentDashboardReadResultSummary(execution)
	case "cabinet.storage.show_status":
		return plannerAgentStorageStatusReadResultSummary(execution)
	case "cabinet.wishlist.search_entries":
		return plannerAgentWishlistReadResultSummary(execution)
	case "cabinet.collections.search":
		return plannerAgentCollectionsReadResultSummary(execution)
	case "cabinet.integrations.search_providers":
		return plannerAgentIntegrationProvidersReadResultSummary(execution)
	case "cabinet.data.export_bundle":
		return plannerAgentDataExportBundleReadResultSummary(execution)
	case "cabinet.maintenance.run_safe_check":
		return plannerAgentMaintenanceSafeCheckReadResultSummary(execution)
	case "cabinet.inbox.search_notifications":
		return plannerAgentInboxNotificationsReadResultSummary(execution, "inbox_notifications")
	case "cabinet.inbox.summarise_unhandled":
		return plannerAgentInboxNotificationsReadResultSummary(execution, "inbox_unhandled")
	case "cabinet.users.search":
		return plannerAgentWorkspaceUsersReadResultSummary(execution)
	case "cabinet.media.search", "cabinet.media.review_unlinked":
		return plannerAgentMediaAssetsReadResultSummary(execution)
	case "cabinet.purchases.search_orders":
		return plannerAgentPurchaseOrdersReadResultSummary(execution)
	case "cabinet.market_watch.search_watches":
		return plannerAgentMarketWatchWatchesReadResultSummary(execution)
	case "cabinet.discoveries.search", "cabinet.discoveries.review_result":
		return plannerAgentDiscoveryResultsReadResultSummary(execution)
	default:
		return nil
	}
}

func plannerAgentChatActionTimelineReadResultSummary(execution map[string]any) *chat.AgentResponseResultSummary {
	rawEntries := plannerReadResultMapSlice(execution["timeline_entries"])
	if len(rawEntries) == 0 {
		if count := plannerReadResultInt(execution["timeline_entry_count"]); count == 0 {
			return &chat.AgentResponseResultSummary{Kind: "chat_action_timeline", Total: 0}
		}
		return nil
	}
	total := len(rawEntries)
	if _, ok := execution["timeline_entry_count"]; ok {
		total = plannerReadResultInt(execution["timeline_entry_count"])
	}
	limit := len(rawEntries)
	if limit > plannerAgentReadResultItemLimit {
		limit = plannerAgentReadResultItemLimit
	}
	items := make([]chat.AgentResponseResultItem, 0, limit)
	for _, entry := range rawEntries[:limit] {
		id := plannerBoundedReadResultText(plannerReadResultString(entry["workflow_run_id"]))
		title := plannerBoundedReadResultText(plannerReadResultString(entry["capability_id"]))
		status := plannerBoundedReadResultText(plannerReadResultString(entry["status"]))
		category := plannerBoundedReadResultText(plannerReadResultString(entry["confirmation_state"]))
		if operation := plannerBoundedReadResultText(plannerReadResultString(entry["operation"])); operation != "" {
			category = operation
		}
		if id == "" && title == "" && status == "" && category == "" {
			continue
		}
		items = append(items, chat.AgentResponseResultItem{
			ID:       id,
			Title:    title,
			Status:   status,
			Category: category,
		})
	}
	return &chat.AgentResponseResultSummary{
		Kind:  "chat_action_timeline",
		Total: total,
		Items: items,
	}
}

func plannerAgentInventoryReadResultSummary(execution map[string]any) *chat.AgentResponseResultSummary {
	rawItems, ok := execution["items"].([]collection.Item)
	if !ok {
		return nil
	}
	total := len(rawItems)
	if value, ok := execution["total"].(int); ok && value >= 0 {
		total = value
	}
	limit := len(rawItems)
	if limit > plannerAgentReadResultItemLimit {
		limit = plannerAgentReadResultItemLimit
	}
	items := make([]chat.AgentResponseResultItem, 0, limit)
	for _, item := range rawItems[:limit] {
		items = append(items, chat.AgentResponseResultItem{
			ID:         plannerBoundedReadResultText(item.ID),
			PartNumber: plannerBoundedReadResultText(item.PartNumber),
			Title:      plannerBoundedReadResultText(item.Title),
			Status:     plannerBoundedReadResultText(item.Status),
			Category:   plannerBoundedReadResultText(item.Category),
			Brand:      plannerBoundedReadResultText(item.Brand),
		})
	}
	return &chat.AgentResponseResultSummary{
		Kind:  "inventory_items",
		Total: total,
		Items: items,
	}
}

func plannerAgentDashboardReadResultSummary(execution map[string]any) *chat.AgentResponseResultSummary {
	signals := plannerReadResultMapSlice(execution["attention_signals"])
	recentItems := plannerReadResultMapSlice(execution["recent_items"])
	if len(signals) == 0 && len(recentItems) == 0 {
		if unavailable, _ := plannerReadResultBool(plannerReadResultMapValue(execution["dependency_state"], "status"), "unavailable"); unavailable {
			return &chat.AgentResponseResultSummary{Kind: "dashboard_activity", Total: 0}
		}
		return nil
	}

	limit := len(signals)
	if limit > plannerAgentReadResultItemLimit {
		limit = plannerAgentReadResultItemLimit
	}
	metrics := make([]chat.AgentResponseResultMetric, 0, limit)
	for _, signal := range signals[:limit] {
		metrics = append(metrics, chat.AgentResponseResultMetric{
			ID:    plannerBoundedReadResultText(plannerReadResultString(signal["id"])),
			Label: plannerBoundedReadResultText(plannerReadResultString(signal["label"])),
			Value: plannerReadResultInt(signal["count"]),
			Route: plannerBoundedReadResultText(plannerReadResultString(signal["destination_link"])),
		})
	}

	itemLimit := len(recentItems)
	if itemLimit > plannerAgentReadResultItemLimit {
		itemLimit = plannerAgentReadResultItemLimit
	}
	items := make([]chat.AgentResponseResultItem, 0, itemLimit)
	for _, item := range recentItems[:itemLimit] {
		items = append(items, chat.AgentResponseResultItem{
			ID:       plannerBoundedReadResultText(plannerReadResultString(item["item_id"])),
			Title:    plannerBoundedReadResultText(plannerReadResultString(item["title"])),
			Category: "Recent Dashboard item",
		})
	}

	return &chat.AgentResponseResultSummary{
		Kind:    "dashboard_activity",
		Total:   len(metrics) + len(items),
		Items:   items,
		Metrics: metrics,
	}
}

func plannerAgentStorageStatusReadResultSummary(execution map[string]any) *chat.AgentResponseResultSummary {
	storageStatus := plannerBoundedReadResultText(plannerReadResultString(execution["storage_status"]))
	backupStatus := plannerBoundedReadResultText(plannerReadResultString(execution["backup_status"]))
	if storageStatus == "" && backupStatus == "" {
		return nil
	}
	item := chat.AgentResponseResultItem{
		ID:       "storage-status",
		Title:    "Storage status",
		Status:   storageStatus,
		Category: "Backup: " + backupStatus,
	}
	if backupStatus == "" {
		item.Category = "Backup status unavailable"
	}
	return &chat.AgentResponseResultSummary{
		Kind:  "storage_status",
		Total: 1,
		Items: []chat.AgentResponseResultItem{item},
	}
}

func plannerAgentWishlistReadResultSummary(execution map[string]any) *chat.AgentResponseResultSummary {
	rawEntries, ok := execution["entries"].([]wishlist.Entry)
	if !ok {
		return nil
	}
	total := len(rawEntries)
	if value, ok := execution["total"].(int); ok && value >= 0 {
		total = value
	}
	limit := len(rawEntries)
	if limit > plannerAgentReadResultItemLimit {
		limit = plannerAgentReadResultItemLimit
	}
	items := make([]chat.AgentResponseResultItem, 0, limit)
	for _, entry := range rawEntries[:limit] {
		items = append(items, chat.AgentResponseResultItem{
			ID:       plannerBoundedReadResultText(entry.ID),
			Title:    plannerBoundedReadResultText(entry.ItemID),
			Status:   plannerWishlistEntryStatus(entry),
			Category: plannerBoundedReadResultText(entry.Priority),
		})
	}
	return &chat.AgentResponseResultSummary{
		Kind:  "wishlist_entries",
		Total: total,
		Items: items,
	}
}

func plannerWishlistEntryStatus(entry wishlist.Entry) string {
	switch {
	case entry.Deleted:
		return "deleted"
	case entry.Delivered:
		return "delivered"
	case entry.Owned:
		return "purchased"
	case entry.BelowTargetNow:
		return "below_target"
	case entry.HighlightHit:
		return "highlighted"
	default:
		return "watching"
	}
}

func plannerAgentCollectionsReadResultSummary(execution map[string]any) *chat.AgentResponseResultSummary {
	collections := plannerReadResultStringSlice(execution["collections"])
	workspaceItems, _ := execution["items"].([]agentCollectionsWorkspaceItem)
	if len(collections) == 0 && len(workspaceItems) == 0 {
		return &chat.AgentResponseResultSummary{Kind: "collections", Total: 0}
	}
	total := len(collections) + len(workspaceItems)
	if value, ok := execution["total"].(int); ok && value >= 0 {
		total = value + len(workspaceItems)
	}

	items := make([]chat.AgentResponseResultItem, 0, plannerAgentReadResultItemLimit)
	for _, name := range collections {
		if len(items) >= plannerAgentReadResultItemLimit {
			break
		}
		name = plannerBoundedReadResultText(name)
		if name == "" {
			continue
		}
		items = append(items, chat.AgentResponseResultItem{
			ID:       plannerBoundedReadResultText("collection:" + name),
			Title:    name,
			Status:   "available",
			Category: "Collection",
		})
	}
	for _, item := range workspaceItems {
		if len(items) >= plannerAgentReadResultItemLimit {
			break
		}
		items = append(items, chat.AgentResponseResultItem{
			ID:       plannerBoundedReadResultText(item.ID),
			Title:    plannerBoundedReadResultText(item.Name),
			Status:   "assigned",
			Category: plannerBoundedReadResultText(item.CollectionName),
		})
	}

	return &chat.AgentResponseResultSummary{
		Kind:  "collections",
		Total: total,
		Items: items,
	}
}

func plannerAgentIntegrationProvidersReadResultSummary(execution map[string]any) *chat.AgentResponseResultSummary {
	providers := plannerReadResultMapSlice(execution["providers"])
	if len(providers) == 0 {
		return &chat.AgentResponseResultSummary{Kind: "integration_providers", Total: 0}
	}
	total := len(providers)
	if value, ok := execution["total"].(int); ok && value >= 0 {
		total = value
	}
	limit := len(providers)
	if limit > plannerAgentReadResultItemLimit {
		limit = plannerAgentReadResultItemLimit
	}
	items := make([]chat.AgentResponseResultItem, 0, limit)
	for _, provider := range providers[:limit] {
		id := plannerBoundedReadResultText(plannerReadResultString(provider["id"]))
		status := plannerBoundedReadResultText(plannerReadResultString(provider["status"]))
		setupRequired, setupKnown := plannerReadResultBool(provider["setup_required"], "true")
		category := "Integration provider"
		if setupKnown {
			if setupRequired {
				category = "Setup required"
			} else {
				category = "Ready"
			}
		}
		if id == "" && status == "" {
			continue
		}
		items = append(items, chat.AgentResponseResultItem{
			ID:       id,
			Title:    id,
			Status:   status,
			Category: category,
		})
	}
	return &chat.AgentResponseResultSummary{
		Kind:  "integration_providers",
		Total: total,
		Items: items,
	}
}

func plannerAgentDataExportBundleReadResultSummary(execution map[string]any) *chat.AgentResponseResultSummary {
	status := plannerBoundedReadResultText(plannerReadResultString(execution["status"]))
	exportScope := plannerBoundedReadResultText(plannerReadResultString(execution["export_scope"]))
	if status == "" && exportScope == "" {
		return nil
	}
	if exportScope == "" {
		exportScope = "default"
	}
	if status == "" {
		status = "ready"
	}
	return &chat.AgentResponseResultSummary{
		Kind:  "data_export_bundle",
		Total: 1,
		Items: []chat.AgentResponseResultItem{
			{
				ID:       "data-export-bundle",
				Title:    "Data export bundle",
				Status:   status,
				Category: "Scope: " + exportScope,
			},
		},
	}
}

func plannerAgentMaintenanceSafeCheckReadResultSummary(execution map[string]any) *chat.AgentResponseResultSummary {
	status := plannerBoundedReadResultText(plannerReadResultString(execution["status"]))
	maintenanceCheck := plannerBoundedReadResultText(plannerReadResultString(execution["maintenance_check"]))
	checkLevel := plannerBoundedReadResultText(plannerReadResultString(execution["check_level"]))
	if status == "" && maintenanceCheck == "" && checkLevel == "" {
		return nil
	}
	if status == "" {
		status = "checked"
	}
	if maintenanceCheck == "" {
		maintenanceCheck = "safe"
	}
	category := "Check: " + maintenanceCheck
	if checkLevel != "" {
		category = category + " / " + checkLevel
	}
	return &chat.AgentResponseResultSummary{
		Kind:  "maintenance_safe_check",
		Total: 1,
		Items: []chat.AgentResponseResultItem{
			{
				ID:       "maintenance-safe-check",
				Title:    "Maintenance safe check",
				Status:   status,
				Category: category,
			},
		},
	}
}

func plannerAgentInboxNotificationsReadResultSummary(execution map[string]any, kind string) *chat.AgentResponseResultSummary {
	if kind == "" {
		kind = "inbox_notifications"
	}
	rawItems := plannerReadResultMapSlice(execution["items"])
	if len(rawItems) == 0 {
		if count := plannerReadResultInt(execution["item_count"]); count == 0 {
			return &chat.AgentResponseResultSummary{Kind: kind, Total: 0}
		}
		return nil
	}
	total := len(rawItems)
	if _, ok := execution["item_count"]; ok {
		value := plannerReadResultInt(execution["item_count"])
		total = value
	}
	limit := len(rawItems)
	if limit > plannerAgentReadResultItemLimit {
		limit = plannerAgentReadResultItemLimit
	}
	items := make([]chat.AgentResponseResultItem, 0, limit)
	for _, rawItem := range rawItems[:limit] {
		id := plannerBoundedReadResultText(plannerReadResultString(rawItem["id"]))
		title := plannerBoundedReadResultText(plannerReadResultString(rawItem["title"]))
		status := plannerBoundedReadResultText(plannerReadResultString(rawItem["status"]))
		source := plannerBoundedReadResultText(plannerReadResultString(rawItem["source"]))
		if id == "" && title == "" && status == "" && source == "" {
			continue
		}
		items = append(items, chat.AgentResponseResultItem{
			ID:       id,
			Title:    title,
			Status:   status,
			Category: source,
		})
	}
	return &chat.AgentResponseResultSummary{
		Kind:  kind,
		Total: total,
		Items: items,
	}
}

func plannerAgentWorkspaceUsersReadResultSummary(execution map[string]any) *chat.AgentResponseResultSummary {
	rawUsers := plannerReadResultMapSlice(execution["users"])
	if len(rawUsers) == 0 {
		if count := plannerReadResultInt(execution["user_count"]); count == 0 {
			return &chat.AgentResponseResultSummary{Kind: "workspace_users", Total: 0}
		}
		return nil
	}
	total := len(rawUsers)
	if _, ok := execution["user_count"]; ok {
		total = plannerReadResultInt(execution["user_count"])
	}
	limit := len(rawUsers)
	if limit > plannerAgentReadResultItemLimit {
		limit = plannerAgentReadResultItemLimit
	}
	items := make([]chat.AgentResponseResultItem, 0, limit)
	for _, rawUser := range rawUsers[:limit] {
		id := plannerBoundedReadResultText(plannerReadResultString(rawUser["id"]))
		displayName := plannerBoundedReadResultText(plannerReadResultString(rawUser["display_name"]))
		username := plannerBoundedReadResultText(plannerReadResultString(rawUser["username"]))
		status := plannerBoundedReadResultText(plannerReadResultString(rawUser["status"]))
		role := plannerBoundedReadResultText(plannerReadResultString(rawUser["role"]))
		title := displayName
		if title == "" {
			title = username
		}
		if id == "" && title == "" && status == "" && role == "" {
			continue
		}
		items = append(items, chat.AgentResponseResultItem{
			ID:       id,
			Title:    title,
			Status:   status,
			Category: role,
		})
	}
	return &chat.AgentResponseResultSummary{
		Kind:  "workspace_users",
		Total: total,
		Items: items,
	}
}

func plannerAgentMediaAssetsReadResultSummary(execution map[string]any) *chat.AgentResponseResultSummary {
	rawAssets, ok := execution["assets"].([]media.WorkspaceAsset)
	if !ok {
		return nil
	}
	total := len(rawAssets)
	if value, ok := execution["total"].(int); ok && value >= 0 {
		total = value
	}
	limit := len(rawAssets)
	if limit > plannerAgentReadResultItemLimit {
		limit = plannerAgentReadResultItemLimit
	}
	items := make([]chat.AgentResponseResultItem, 0, limit)
	for _, asset := range rawAssets[:limit] {
		title := plannerBoundedReadResultText(asset.Title)
		if title == "" {
			title = plannerBoundedReadResultText(asset.Filename)
		}
		status := plannerBoundedReadResultText(asset.LinkageState)
		if status == "" {
			status = plannerBoundedReadResultText(asset.AnalysisStatus)
		}
		category := plannerBoundedReadResultText(asset.Source)
		if category == "" {
			category = "Media asset"
		}
		items = append(items, chat.AgentResponseResultItem{
			ID:       plannerBoundedReadResultText(asset.ID),
			Title:    title,
			Status:   status,
			Category: category,
		})
	}
	return &chat.AgentResponseResultSummary{
		Kind:  "media_assets",
		Total: total,
		Items: items,
	}
}

func plannerAgentPurchaseOrdersReadResultSummary(execution map[string]any) *chat.AgentResponseResultSummary {
	rawOrders, ok := execution["orders"].([]commerce.PurchaseOrder)
	if !ok {
		return nil
	}
	total := len(rawOrders)
	if value, ok := execution["total"].(int); ok && value >= 0 {
		total = value
	}
	limit := len(rawOrders)
	if limit > plannerAgentReadResultItemLimit {
		limit = plannerAgentReadResultItemLimit
	}
	items := make([]chat.AgentResponseResultItem, 0, limit)
	for _, order := range rawOrders[:limit] {
		category := plannerBoundedReadResultText(order.Source)
		if category == "" {
			category = "Purchase order"
		}
		if order.LineItemCount > 0 {
			category = fmt.Sprintf("%s / %d line items", category, order.LineItemCount)
		}
		items = append(items, chat.AgentResponseResultItem{
			ID:       plannerBoundedReadResultText(order.OrderID),
			Title:    plannerBoundedReadResultText(order.OrderID),
			Status:   plannerBoundedReadResultText(order.Status),
			Category: plannerBoundedReadResultText(category),
		})
	}
	return &chat.AgentResponseResultSummary{
		Kind:  "purchase_orders",
		Total: total,
		Items: items,
	}
}

func plannerAgentMarketWatchWatchesReadResultSummary(execution map[string]any) *chat.AgentResponseResultSummary {
	rawWatches, ok := execution["watches"].([]scanner.QuerySet)
	if !ok {
		return nil
	}
	total := len(rawWatches)
	if value, ok := execution["total"].(int); ok && value >= 0 {
		total = value
	}
	limit := len(rawWatches)
	if limit > plannerAgentReadResultItemLimit {
		limit = plannerAgentReadResultItemLimit
	}
	items := make([]chat.AgentResponseResultItem, 0, limit)
	for _, watch := range rawWatches[:limit] {
		status := "paused"
		if watch.Enabled {
			status = "enabled"
		}
		category := "Market Watch"
		if len(watch.ProviderScope) > 0 {
			category = fmt.Sprintf("%s / %d providers", category, len(watch.ProviderScope))
		}
		if watch.LastCandidateCount > 0 {
			category = fmt.Sprintf("%s / %d last results", category, watch.LastCandidateCount)
		}
		items = append(items, chat.AgentResponseResultItem{
			ID:       plannerBoundedReadResultText(watch.ID),
			Title:    plannerBoundedReadResultText(watch.Name),
			Status:   plannerBoundedReadResultText(status),
			Category: plannerBoundedReadResultText(category),
		})
	}
	return &chat.AgentResponseResultSummary{
		Kind:  "market_watch_watches",
		Total: total,
		Items: items,
	}
}

func plannerAgentDiscoveryResultsReadResultSummary(execution map[string]any) *chat.AgentResponseResultSummary {
	rawItems, ok := execution["items"].([]discovery.Item)
	if !ok {
		if item, ok := execution["item"].(discovery.Item); ok {
			rawItems = []discovery.Item{item}
		} else {
			return nil
		}
	}
	total := len(rawItems)
	if value, ok := execution["total"].(int); ok && value >= 0 {
		total = value
	}
	limit := len(rawItems)
	if limit > plannerAgentReadResultItemLimit {
		limit = plannerAgentReadResultItemLimit
	}
	items := make([]chat.AgentResponseResultItem, 0, limit)
	for _, result := range rawItems[:limit] {
		status := plannerBoundedReadResultText(result.TriageStatus)
		if status == "" {
			status = plannerBoundedReadResultText(result.Status)
		}
		category := plannerBoundedReadResultText(result.SourceProvider)
		if category == "" {
			category = "Discovery result"
		}
		if result.NeedsReview {
			category = plannerBoundedReadResultText(category + " / needs review")
		}
		items = append(items, chat.AgentResponseResultItem{
			ID:       plannerBoundedReadResultText(result.CandidateID),
			Title:    plannerBoundedReadResultText(result.Title),
			Status:   status,
			Category: category,
		})
	}
	return &chat.AgentResponseResultSummary{
		Kind:  "discovery_results",
		Total: total,
		Items: items,
	}
}

func plannerBoundedReadResultText(value string) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) > plannerAgentReadResultTextLimit {
		runes = runes[:plannerAgentReadResultTextLimit]
	}
	return string(runes)
}

func plannerReadResultMapSlice(value any) []map[string]any {
	typed, ok := value.([]map[string]any)
	if ok {
		return typed
	}
	raw, ok := value.([]any)
	if !ok {
		return nil
	}
	items := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		if mapped, ok := item.(map[string]any); ok {
			items = append(items, mapped)
		}
	}
	return items
}

func plannerReadResultStringSlice(value any) []string {
	raw, ok := value.([]string)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func plannerReadResultString(value any) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func plannerReadResultInt(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		parsed, err := typed.Int64()
		if err == nil {
			return int(parsed)
		}
	}
	return 0
}

func plannerReadResultMapValue(value any, key string) any {
	if mapped, ok := value.(map[string]any); ok {
		return mapped[key]
	}
	return nil
}

func plannerReadResultBool(value any, expected string) (bool, bool) {
	text := strings.TrimSpace(fmt.Sprint(value))
	if text == "" || text == "<nil>" {
		return false, false
	}
	return strings.EqualFold(text, expected), true
}

func plannerProviderSetupRequired(result map[string]any) bool {
	nextAction := strings.ToLower(strings.TrimSpace(fmt.Sprint(result["setup_next_action"])))
	for _, prefix := range []string{"configure_", "choose_", "review_"} {
		if strings.HasPrefix(nextAction, prefix) {
			return true
		}
	}
	return false
}

func plannerAgentContextScopeMismatch(profileID, threadID string, agentCtx map[string]any) []string {
	mismatches := []string{}
	if contextProfileID := strings.TrimSpace(fmt.Sprint(agentCtx["profile_id"])); contextProfileID != "" && contextProfileID != "<nil>" && contextProfileID != strings.TrimSpace(profileID) {
		mismatches = append(mismatches, "profile_id")
	}
	if contextThreadID := strings.TrimSpace(fmt.Sprint(agentCtx["thread_id"])); contextThreadID != "" && contextThreadID != "<nil>" && contextThreadID != strings.TrimSpace(threadID) {
		mismatches = append(mismatches, "thread_id")
	}
	return mismatches
}

func agentPlannerSkillsForSession(ctx context.Context, conn *sql.DB, profileID string, skills []agentskills.Skill) []agentskills.Skill {
	if _, err := resolveAgentAdminAuthority(ctx, conn, profileID); err == nil {
		return skills
	}
	filtered := make([]agentskills.Skill, 0, len(skills))
	for _, skill := range skills {
		if !strings.HasPrefix(strings.TrimSpace(skill.ID), "cabinet.users.") {
			filtered = append(filtered, skill)
		}
	}
	return filtered
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
		agentResponse, _ := chat.NewAgentResponse(
			chat.AgentResponseReadResult,
			messageText,
			content,
			"assistant.agent_capability_explanation",
			"Cabinet Agent capabilities",
			strings.TrimSpace(fmt.Sprint(agentCtx["surface_id"])),
			strings.TrimSpace(fmt.Sprint(agentCtx["source_channel"])),
		)
		if assistantMessage, messageErr := chatSvc.CreateMessage(ctx, profileID, threadID, "assistant", messageText, map[string]any{
			"agent_capabilities": result,
			"agent_response":     agentResponse,
		}); messageErr == nil {
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
	if !ok || (skill.SafetyLevel != agentskills.SafetyReadOnly && !plannerSafePreviewExecutionSkill(skill)) {
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
	var adminAuthority agentAdminAuthority
	var err error
	req, adminAuthority, err = authorizeAgentUsersRequest(ctx, conn, req)
	if err != nil {
		return nil, agentskills.AgentAuthorityReview{}, err
	}
	if adminAuthority.UserID != "" {
		ctx = withAgentAdminAuthority(ctx, adminAuthority)
	}
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
	result, blocker, err := applyAgentSkill(ctx, conn, chatSvc, selection.SkillID, profileID, req.Parameters)
	if err != nil {
		if blocker != "" {
			review.Blocker = blocker
		}
		return nil, review, err
	}
	return result, review, nil
}

func plannerSafePreviewExecutionSkill(skill agentskills.Skill) bool {
	return skill.ID == "cabinet.data.export_bundle" &&
		skill.SafetyLevel == agentskills.SafetyPreviewOnly &&
		!skill.Permissions.LocalWrite &&
		!skill.Permissions.ExternalWrite &&
		!skill.Permissions.Destructive
}

func previewLocalWritePlannerSelection(ctx context.Context, conn *sql.DB, chatSvc *chat.Service, registry agentskills.Registry, profileID, threadID string, selection chatAgentSkillSelection, envelope map[string]any, sourceMessageID string) (map[string]any, agentskills.AgentAuthorityReview, error) {
	if chatSvc == nil || selection.Decision != "select_skill" || selection.ErrorCode != "" {
		return nil, agentskills.AgentAuthorityReview{}, nil
	}
	skill, ok := registry.Resolve(selection.SkillID)
	if !ok || (!skill.Permissions.LocalWrite && !skill.Permissions.ExternalWrite && !skill.Permissions.Destructive) {
		return nil, agentskills.AgentAuthorityReview{}, nil
	}
	action := plannerChatActionForSkill(selection.SkillID)
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
	var adminAuthority agentAdminAuthority
	var err error
	req, adminAuthority, err = authorizeAgentUsersRequest(ctx, conn, req)
	if err != nil {
		return nil, agentskills.AgentAuthorityReview{}, err
	}
	if adminAuthority.UserID != "" {
		ctx = withAgentAdminAuthority(ctx, adminAuthority)
	}
	if plannerSkillNeedsSelectedTarget(skill) || plannerSkillRequiresStrictAgentContext(skill) {
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
	if action == "" || strings.EqualFold(strings.TrimSpace(req.SourceChannel), "telegram") {
		preview, err := registry.Preview(req)
		if err != nil {
			return nil, review, fmt.Errorf("create Agent Skill preview for %s: %w", selection.SkillID, err)
		}
		if clarification, blocked := plannerPreviewBlockerClarification(preview); blocked {
			return clarification, review, nil
		}
		record, err := createDurableAgentSkillPreview(ctx, conn, req, preview)
		if err != nil {
			return nil, review, fmt.Errorf("persist Agent Skill preview for %s: %w", selection.SkillID, err)
		}
		preview = bindDurableAgentSkillPreviewResponse(preview, record)
		return genericPlannerAgentSkillPreview(preview, req, review), review, nil
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

func plannerPreviewBlockerClarification(preview agentskills.PreviewResponse) (map[string]any, bool) {
	blocker := strings.TrimSpace(preview.Blocker)
	if blocker != "wishlist_item_context_required" {
		return nil, false
	}
	nextAction := strings.TrimSpace(preview.NextAction)
	if nextAction == "" {
		nextAction = "Provide the missing Cabinet record or requested details, then retry before applying any change."
	}
	return map[string]any{
		"decision":        "clarify",
		"message":         "I need the required Cabinet details before creating an actionable preview.",
		"error":           map[string]any{"code": "missing_context", "message": "Cabinet rejected an incomplete Agent Skill preview before any mutation."},
		"missing_context": []string{blocker},
		"clarification":   map[string]string{blocker: nextAction},
		"next_action":     nextAction,
	}, true
}

func genericPlannerAgentSkillPreview(preview agentskills.PreviewResponse, req agentskills.PreviewRequest, review agentskills.AgentAuthorityReview) map[string]any {
	applyRequest := map[string]any{
		"profile_id": strings.TrimSpace(req.ProfileID),
		"preview_id": strings.TrimSpace(preview.PreviewID),
		"confirm":    true,
	}
	cancelRequest := map[string]any{
		"profile_id": strings.TrimSpace(req.ProfileID),
		"preview_id": strings.TrimSpace(preview.PreviewID),
	}
	strongConfirmationRequest := map[string]any{
		"profile_id": strings.TrimSpace(req.ProfileID),
		"preview_id": strings.TrimSpace(preview.PreviewID),
	}
	return map[string]any{
		"kind":                         "agent_skill_preview",
		"preview_kind":                 "agent_skill",
		"preview_id":                   strings.TrimSpace(preview.PreviewID),
		"preview_status":               strings.TrimSpace(preview.PreviewStatus),
		"expires_at":                   strings.TrimSpace(preview.ExpiresAt),
		"skill_id":                     strings.TrimSpace(req.SkillID),
		"status":                       strings.TrimSpace(preview.PreviewStatus),
		"safety_level":                 preview.SafetyLevel,
		"allowed":                      preview.Allowed,
		"preview_only":                 true,
		"confirmation_required":        preview.ConfirmationRequired,
		"strong_confirmation_required": preview.StrongConfirmationRequired,
		"strong_confirmation_endpoint": preview.StrongConfirmationEndpoint,
		"strong_confirmation_request":  strongConfirmationRequest,
		"mutation_applied":             false,
		"blocker":                      preview.Blocker,
		"next_action":                  preview.NextAction,
		"target":                       redactPlannerEvidenceMap(preview.Target),
		"parameters":                   plannerNonSecretActionPayload(req.Parameters),
		"authority_blocker":            review.Blocker,
		"apply_endpoint":               "/api/agent/skills/apply",
		"cancel_endpoint":              "/api/agent/skills/cancel",
		"retrieval_endpoint":           "/api/agent/skills/preview",
		"source_surface":               strings.TrimSpace(req.SourceSurface),
		"source_channel":               strings.TrimSpace(req.SourceChannel),
		"source_thread_id":             strings.TrimSpace(req.SourceThreadID),
		"source_message_id":            strings.TrimSpace(req.SourceMessageID),
		"apply_contract": map[string]any{
			"endpoint": "/api/agent/skills/apply",
			"method":   "POST",
			"request":  applyRequest,
		},
		"apply_request": applyRequest,
		"cancel_contract": map[string]any{
			"endpoint": "/api/agent/skills/cancel",
			"method":   "POST",
			"request":  cancelRequest,
		},
		"cancel_request": cancelRequest,
		"retrieval_request": map[string]any{
			"profile_id": strings.TrimSpace(req.ProfileID),
			"preview_id": strings.TrimSpace(preview.PreviewID),
		},
	}
}

func plannerSkillRequiresStrictAgentContext(skill agentskills.Skill) bool {
	return skill.Category == "inbox" || skill.Category == "users" || skill.Category == "settings" || skill.Category == "storage" || skill.Category == "data-management"
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
	if previewResult, _ := result["preview_result"].(map[string]any); previewResult != nil && previewResult["kind"] == "agent_skill_preview" {
		tokenState["preview_id"] = previewID
		tokenState["skill_id"] = previewResult["skill_id"]
		tokenState["action"] = "agent_skill.apply"
		tokenState["apply_state"] = "pending_explicit_confirmation"
		tokenState["apply_endpoint"] = "/api/agent/skills/apply"
		tokenState["cancel_endpoint"] = "/api/agent/skills/cancel"
		tokenState["authority_revalidated_on_apply"] = true
	} else if previewID != "" && previewID != "<nil>" {
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
		if plannerSensitiveEvidenceKey(key) {
			continue
		}
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
		if safeValue, ok := plannerNonSecretActionValue(value); ok {
			out[key] = safeValue
		}
	}
	return out
}

func plannerNonSecretActionValue(value any) (any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		return plannerNonSecretActionPayload(typed), true
	case []any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			if safeItem, ok := plannerNonSecretActionValue(item); ok {
				out = append(out, safeItem)
			}
		}
		return out, true
	case string:
		lower := strings.ToLower(typed)
		if strings.Contains(lower, "sk-") || strings.Contains(lower, "secret") || strings.Contains(lower, "bearer ") {
			return nil, false
		}
		return typed, true
	default:
		return value, true
	}
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
	_, ok := chatAgentIntentDomain(content)
	return ok
}

func chatMessageRequestsLiteralProviderResponse(normalized string) bool {
	for _, prefix := range []string{
		"reply with exactly:",
		"respond with exactly:",
		"say exactly:",
		"return exactly:",
		"echo exactly:",
	} {
		if strings.HasPrefix(normalized, prefix) && strings.TrimSpace(strings.TrimPrefix(normalized, prefix)) != "" {
			return true
		}
	}
	return false
}

func chatAgentIntentDomain(content string) (string, bool) {
	normalized := normalizePlannerText(content)
	if normalized == "" {
		return "", false
	}
	if chatMessageRequestsLiteralProviderResponse(normalized) {
		return "", false
	}
	dashboardSummaryIntent := (strings.Contains(normalized, "summarise") ||
		strings.Contains(normalized, "summarize") ||
		strings.Contains(normalized, "summary") ||
		strings.Contains(normalized, "what changed")) &&
		strings.Contains(normalized, "dashboard")
	if dashboardSummaryIntent {
		return "dashboard", true
	}
	actionIntent := strings.Contains(normalized, "find") ||
		strings.Contains(normalized, "search") ||
		strings.Contains(normalized, "look up") ||
		strings.Contains(normalized, "lookup") ||
		strings.Contains(normalized, "summarise") ||
		strings.Contains(normalized, "summarize") ||
		strings.Contains(normalized, "show") ||
		strings.Contains(normalized, "list") ||
		strings.Contains(normalized, "review") ||
		strings.Contains(normalized, "inspect") ||
		strings.Contains(normalized, "create") ||
		strings.Contains(normalized, "add") ||
		strings.Contains(normalized, "rename") ||
		strings.Contains(normalized, "update") ||
		strings.Contains(normalized, "change") ||
		strings.Contains(normalized, "mark") ||
		strings.Contains(normalized, "archive") ||
		strings.Contains(normalized, "route") ||
		strings.Contains(normalized, "invite") ||
		strings.Contains(normalized, "configure") ||
		strings.Contains(normalized, "explain") ||
		strings.Contains(normalized, "test") ||
		strings.Contains(normalized, "check") ||
		strings.Contains(normalized, "connect") ||
		strings.Contains(normalized, "repair") ||
		strings.Contains(normalized, "disable") ||
		strings.Contains(normalized, "run") ||
		strings.Contains(normalized, "dismiss") ||
		strings.Contains(normalized, "send") ||
		strings.Contains(normalized, "handoff") ||
		strings.Contains(normalized, "receive") ||
		strings.Contains(normalized, "reconcile") ||
		strings.Contains(normalized, "purchase") ||
		strings.Contains(normalized, "hide") ||
		strings.Contains(normalized, "delete") ||
		strings.Contains(normalized, "move") ||
		strings.Contains(normalized, "assign") ||
		strings.Contains(normalized, "upload") ||
		strings.Contains(normalized, "import") ||
		strings.Contains(normalized, "export") ||
		strings.Contains(normalized, "attach") ||
		strings.Contains(normalized, "annotate") ||
		strings.Contains(normalized, "note") ||
		strings.Contains(normalized, "restore") ||
		strings.Contains(normalized, "remove")
	actionIntent = actionIntent || strings.Contains(normalized, "make")
	if !actionIntent {
		return "", false
	}
	for _, candidate := range []struct {
		domain string
		terms  []string
	}{
		{domain: "media", terms: []string{"media", "photo", "image", "attachment", "unlinked"}},
		{domain: "acquisition", terms: []string{"integration", "provider", "connection", "market watch", "saved watch", "watch", "discover", "result", "listing", "purchase", "order", "ebay"}},
		{domain: "wishlist", terms: []string{"wishlist", "wish list", "wanted item"}},
		{domain: "collections", terms: []string{"collection", "collections"}},
		{domain: "admin", terms: []string{"inbox", "notification", "workspace user", "invite user", "user role", "user-", " admin", "profile setting", "account setting", "appearance", "storage", "backup", "import", "export", "restore"}},
		{domain: "inventory", terms: []string{"inventory", "item", "part number"}},
	} {
		for _, term := range candidate.terms {
			if strings.Contains(normalized, term) {
				return candidate.domain, true
			}
		}
	}
	return "general", true
}

func plannerProviderFailureCode(errorClass string) string {
	class := strings.ToLower(strings.TrimSpace(errorClass))
	switch class {
	case "missing_credentials", "missing_context", "unsupported_model", "rate_limit", "rate_limited", "timeout", "cancelled", "transport_failure", "provider_failure", "unhealthy_provider", "adapter_unavailable", "login_needed", "partial_result":
		return "assistant_provider_" + class
	default:
		return "assistant_planner_failed"
	}
}

func plannerProviderSetupGuidance(nextAction string) string {
	switch strings.TrimSpace(nextAction) {
	case "configure_openai_api_key", "configure_openai_provider":
		return "Configure and test the OpenAI API key in Integrations, then retry this request."
	case "choose_supported_openai_model":
		return "Choose a supported OpenAI model in Integrations, then retry this request."
	default:
		return "Review assistant provider setup and retry the request."
	}
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

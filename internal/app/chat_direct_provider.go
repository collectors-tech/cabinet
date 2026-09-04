package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/collectors-tech/cabinet/internal/ai"
	"github.com/collectors-tech/cabinet/internal/chat"
)

const directAssistantHistoryLimit = 50

func dispatchChatDirectAssistantProvider(
	ctx context.Context,
	chatSvc *chat.Service,
	providers *ai.AssistantProviderRegistry,
	profileID, threadID, content string,
	envelope map[string]any,
	sourceMessageID string,
) (map[string]any, bool) {
	assistantContext, _ := envelope["assistant"].(map[string]any)
	sourceSurface, sourceChannel := directAssistantSource(envelope)
	providerID := strings.ToLower(strings.TrimSpace(fmt.Sprint(assistantContext["provider"])))
	if providerID == "" || providerID == "<nil>" {
		return nil, false
	}
	if providers == nil {
		return persistDirectAssistantProviderFailure(
			ctx, chatSvc, profileID, threadID, content, sourceMessageID, providerID, "",
			"provider_registry_unavailable", "retry_when_provider_registry_is_available", sourceSurface, sourceChannel, nil,
		), true
	}
	provider, ok := providers.Provider(providerID)
	if !ok || provider == nil {
		return persistDirectAssistantProviderFailure(
			ctx, chatSvc, profileID, threadID, content, sourceMessageID, providerID, "",
			"unsupported_provider", "choose_supported_assistant_provider", sourceSurface, sourceChannel, nil,
		), true
	}
	model := strings.TrimSpace(fmt.Sprint(assistantContext["model"]))
	if model == "<nil>" {
		model = ""
	}

	messages := directAssistantTurnMessages(ctx, chatSvc, profileID, threadID, content)
	response, runErr := provider.RunAssistantTurn(ctx, ai.AssistantTurnRequest{
		ProfileID: profileID,
		ThreadID:  threadID,
		Provider:  providerID,
		Model:     model,
		Messages:  messages,
		Context: map[string]any{
			"agent_context": agentContextEvidence(envelope),
		},
		Metadata: map[string]string{
			"entry_point":              "chat.direct_conversation",
			"governed_dispatch_owner":  "cabinet",
			"cabinet_tool_authority":   "none",
			"raw_provider_tools_given": "false",
		},
	})
	resolvedProvider := strings.TrimSpace(response.Provider)
	if resolvedProvider == "" {
		resolvedProvider = providerID
	}
	resolvedModel := strings.TrimSpace(response.Model)
	if resolvedModel == "" {
		resolvedModel = model
	}
	trace := directAssistantProviderTrace(response)

	if runErr != nil || strings.TrimSpace(response.Text) == "" {
		return persistDirectAssistantProviderFailure(
			ctx,
			chatSvc,
			profileID,
			threadID,
			content,
			sourceMessageID,
			resolvedProvider,
			resolvedModel,
			response.ErrorClass,
			response.SetupNextAction,
			sourceSurface,
			sourceChannel,
			trace,
		), true
	}

	result := map[string]any{
		"mode":           "provider",
		"source":         "assistant_provider",
		"provider":       resolvedProvider,
		"model":          resolvedModel,
		"source_msg_id":  sourceMessageID,
		"provider_trace": trace,
	}
	if chatSvc != nil {
		assistantMessage, err := chatSvc.CreateMessage(ctx, profileID, threadID, "assistant", strings.TrimSpace(response.Text), map[string]any{
			"assistant_response": result,
		})
		if err == nil {
			result["thread_message"] = assistantMessage
		} else {
			result["error"] = map[string]any{"code": "assistant_response_persistence_failed"}
		}
	}
	return result, true
}

func directAssistantTurnMessages(ctx context.Context, chatSvc *chat.Service, profileID, threadID, fallbackContent string) []ai.AssistantTurnMessage {
	if chatSvc == nil {
		return []ai.AssistantTurnMessage{{Role: "user", Content: strings.TrimSpace(fallbackContent)}}
	}
	threadMessages, err := chatSvc.ListMessages(ctx, profileID, threadID)
	if err != nil || len(threadMessages) == 0 {
		return []ai.AssistantTurnMessage{{Role: "user", Content: strings.TrimSpace(fallbackContent)}}
	}
	start := 0
	if len(threadMessages) > directAssistantHistoryLimit {
		start = len(threadMessages) - directAssistantHistoryLimit
	}
	out := make([]ai.AssistantTurnMessage, 0, len(threadMessages)-start)
	for _, message := range threadMessages[start:] {
		role := strings.ToLower(strings.TrimSpace(message.Role))
		if role != "user" && role != "assistant" && role != "system" {
			continue
		}
		if text := strings.TrimSpace(message.Content); text != "" {
			out = append(out, ai.AssistantTurnMessage{Role: role, Content: text})
		}
	}
	if len(out) == 0 {
		out = append(out, ai.AssistantTurnMessage{Role: "user", Content: strings.TrimSpace(fallbackContent)})
	}
	return out
}

func directAssistantProviderTrace(response ai.AssistantTurnResponse) map[string]string {
	trace := map[string]string{
		"entry_point":             "chat.direct_conversation",
		"governed_dispatch_owner": "cabinet",
		"cabinet_tool_authority":  "none",
	}
	for key, value := range response.Metadata {
		switch key {
		case "active_auth_method", "integration_id", "live_provider", "network", "test_provider":
			if value = strings.TrimSpace(value); value != "" {
				trace[key] = value
			}
		}
	}
	return trace
}

func persistDirectAssistantProviderFailure(
	ctx context.Context,
	chatSvc *chat.Service,
	profileID, threadID, originalIntent, sourceMessageID, providerID, model, errorClass, setupNextAction, sourceSurface, sourceChannel string,
	trace map[string]string,
) map[string]any {
	errorClass = strings.TrimSpace(errorClass)
	if errorClass == "" {
		errorClass = "provider_failure"
	}
	state := chat.AgentResponseRetryableFailure
	message := "The assistant provider could not complete that reply. No Cabinet action was completed."
	if errorClass == "missing_credentials" || errorClass == "unsupported_model" || strings.TrimSpace(setupNextAction) != "" {
		state = chat.AgentResponseSetupRequired
		message = "The assistant provider needs setup before Cabinet can reply. No Cabinet action was completed."
	}
	agentResponse, _ := chat.NewAgentResponse(
		state,
		message,
		originalIntent,
		"",
		"Cabinet Agent",
		sourceSurface,
		sourceChannel,
	)
	result := map[string]any{
		"mode":              "provider_failure",
		"source":            "assistant_provider",
		"provider":          providerID,
		"model":             model,
		"source_msg_id":     sourceMessageID,
		"recoverable":       true,
		"error":             map[string]any{"code": errorClass},
		"setup_next_action": strings.TrimSpace(setupNextAction),
		"provider_trace":    trace,
	}
	if chatSvc != nil {
		assistantMessage, err := chatSvc.CreateMessage(ctx, profileID, threadID, "assistant", agentResponse.Message, map[string]any{
			"assistant_response": result,
			"agent_response":     agentResponse,
		})
		if err == nil {
			result["thread_message"] = assistantMessage
		}
	}
	return result
}

func directAssistantSource(envelope map[string]any) (string, string) {
	evidence := agentContextEvidence(envelope)
	surface := strings.TrimSpace(fmt.Sprint(evidence["surface_id"]))
	if surface == "" || surface == "<nil>" {
		surface = "chats.main"
	}
	channel := strings.TrimSpace(fmt.Sprint(evidence["source_channel"]))
	if channel == "" || channel == "<nil>" {
		channel = "in-app"
	}
	return surface, channel
}

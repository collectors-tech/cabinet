package chat

import (
	"fmt"
	"strings"
)

type AgentResponseState string

const (
	AgentResponseReadResult            AgentResponseState = "read_result"
	AgentResponseClarificationRequired AgentResponseState = "clarification_required"
	AgentResponseSetupRequired         AgentResponseState = "setup_required"
	AgentResponseAuthorityBlocked      AgentResponseState = "authority_blocked"
	AgentResponseUnsupported           AgentResponseState = "unsupported"
	AgentResponseProviderUnavailable   AgentResponseState = "provider_unavailable"
	AgentResponseRetryableFailure      AgentResponseState = "retryable_failure"
	AgentResponsePreviewRequired       AgentResponseState = "preview_required"
	AgentResponsePreviewExpired        AgentResponseState = "preview_expired"
	AgentResponsePreviewFailed         AgentResponseState = "preview_failed"
	AgentResponsePreviewStaleTarget    AgentResponseState = "preview_stale_target"
	AgentResponseCancelled             AgentResponseState = "cancelled"
	AgentResponseApplied               AgentResponseState = "applied"
)

type AgentResponseSkill struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AgentResponseSource struct {
	Surface string `json:"surface"`
	Channel string `json:"channel"`
}

type AgentResponseAction struct {
	Kind  string `json:"kind"`
	Label string `json:"label"`
	Route string `json:"route,omitempty"`
}

type AgentResponsePreview struct {
	ID      string         `json:"id,omitempty"`
	Action  string         `json:"action,omitempty"`
	Status  string         `json:"status,omitempty"`
	Payload map[string]any `json:"payload,omitempty"`
}

type AgentResponseResultItem struct {
	ID         string `json:"id,omitempty"`
	PartNumber string `json:"part_number,omitempty"`
	Title      string `json:"title,omitempty"`
	Status     string `json:"status,omitempty"`
	Category   string `json:"category,omitempty"`
	Brand      string `json:"brand,omitempty"`
}

type AgentResponseResultSummary struct {
	Kind  string                    `json:"kind"`
	Total int                       `json:"total"`
	Items []AgentResponseResultItem `json:"items"`
}

// AgentResponse is the server-owned, presentation-neutral contract used by
// every Chat surface. Clients render this value; they do not infer success by
// scanning earlier planner, preview, or capability messages.
type AgentResponse struct {
	State          AgentResponseState          `json:"state"`
	Outcome        string                      `json:"outcome"`
	Title          string                      `json:"title"`
	Message        string                      `json:"message"`
	Retryable      bool                        `json:"retryable"`
	OriginalIntent string                      `json:"original_intent,omitempty"`
	Skill          AgentResponseSkill          `json:"skill"`
	Source         AgentResponseSource         `json:"source"`
	NextAction     *AgentResponseAction        `json:"next_action,omitempty"`
	Preview        *AgentResponsePreview       `json:"preview,omitempty"`
	ResultSummary  *AgentResponseResultSummary `json:"result_summary,omitempty"`
}

func NewAgentResponse(state AgentResponseState, message, originalIntent, skillID, skillName, surface, channel string) (AgentResponse, error) {
	response := AgentResponse{
		State:          state,
		Message:        strings.TrimSpace(message),
		OriginalIntent: strings.TrimSpace(originalIntent),
		Skill: AgentResponseSkill{
			ID:   strings.TrimSpace(skillID),
			Name: strings.TrimSpace(skillName),
		},
		Source: AgentResponseSource{
			Surface: strings.TrimSpace(surface),
			Channel: strings.TrimSpace(channel),
		},
	}
	if response.Skill.Name == "" {
		response.Skill.Name = response.Skill.ID
	}
	if response.Source.Surface == "" {
		response.Source.Surface = "chats.main"
	}
	if response.Source.Channel == "" {
		response.Source.Channel = "in-app"
	}

	switch state {
	case AgentResponseReadResult:
		response.Outcome, response.Title = "success", "Read result"
	case AgentResponseClarificationRequired:
		response.Outcome, response.Title = "needs_input", "Clarification required"
		response.NextAction = &AgentResponseAction{Kind: "provide_details", Label: "Provide details"}
	case AgentResponseSetupRequired:
		response.Outcome, response.Title = "blocked", "Setup required"
		response.NextAction = &AgentResponseAction{Kind: "open_setup", Label: "Open setup", Route: "/integrations?provider=openai"}
	case AgentResponseAuthorityBlocked:
		response.Outcome, response.Title = "blocked", "Authority blocked"
		response.NextAction = &AgentResponseAction{Kind: "review_authority", Label: "Review authority", Route: "/settings/skills"}
	case AgentResponseUnsupported:
		response.Outcome, response.Title = "blocked", "Unsupported"
		response.NextAction = &AgentResponseAction{Kind: "new_request", Label: "Start a new request"}
	case AgentResponseProviderUnavailable:
		response.Outcome, response.Title, response.Retryable = "failed", "Provider unavailable", true
		response.NextAction = &AgentResponseAction{Kind: "retry", Label: "Retry"}
	case AgentResponseRetryableFailure:
		response.Outcome, response.Title, response.Retryable = "failed", "Retryable failure", true
		response.NextAction = &AgentResponseAction{Kind: "retry", Label: "Retry"}
	case AgentResponsePreviewRequired:
		response.Outcome, response.Title = "preview", "Preview required"
		response.NextAction = &AgentResponseAction{Kind: "apply", Label: "Apply"}
	case AgentResponsePreviewExpired:
		response.Outcome, response.Title = "failed", "Preview expired"
		response.NextAction = &AgentResponseAction{Kind: "new_preview", Label: "Create a new preview"}
	case AgentResponsePreviewFailed:
		response.Outcome, response.Title, response.Retryable = "failed", "Preview failed", true
		response.NextAction = &AgentResponseAction{Kind: "retry", Label: "Retry"}
	case AgentResponsePreviewStaleTarget:
		response.Outcome, response.Title = "failed", "Preview stale target"
		response.NextAction = &AgentResponseAction{Kind: "refresh_target", Label: "Refresh target"}
	case AgentResponseCancelled:
		response.Outcome, response.Title = "cancelled", "Cancelled"
		response.NextAction = &AgentResponseAction{Kind: "new_request", Label: "Start a new request"}
	case AgentResponseApplied:
		response.Outcome, response.Title = "applied", "Applied"
	default:
		return AgentResponse{}, fmt.Errorf("unsupported agent response state: %s", state)
	}
	if response.Message == "" {
		response.Message = response.Title + "."
	}
	if response.Retryable && response.OriginalIntent == "" {
		return AgentResponse{}, fmt.Errorf("retryable agent response requires original_intent")
	}
	return response, nil
}

func AgentResponseContext(response AgentResponse) map[string]any {
	return map[string]any{"agent_response": response}
}

package chat

import "testing"

func TestAgentResponseStateMatrix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		state      AgentResponseState
		outcome    string
		retryable  bool
		actionKind string
	}{
		{AgentResponseReadResult, "success", false, ""},
		{AgentResponseClarificationRequired, "needs_input", false, "provide_details"},
		{AgentResponseSetupRequired, "blocked", false, "open_setup"},
		{AgentResponseAuthorityBlocked, "blocked", false, "review_authority"},
		{AgentResponseUnsupported, "blocked", false, "new_request"},
		{AgentResponseProviderUnavailable, "failed", true, "retry"},
		{AgentResponseRetryableFailure, "failed", true, "retry"},
		{AgentResponsePreviewRequired, "preview", false, "apply"},
		{AgentResponsePreviewExpired, "failed", false, "new_preview"},
		{AgentResponsePreviewFailed, "failed", true, "retry"},
		{AgentResponsePreviewStaleTarget, "failed", false, "refresh_target"},
		{AgentResponseCancelled, "cancelled", false, "new_request"},
		{AgentResponseApplied, "applied", false, ""},
	}
	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			response, err := NewAgentResponse(tt.state, "truthful response", "bounded original intent", "cabinet.inventory.search_items", "Cabinet Inventory Search", "chats.main", "in-app")
			if err != nil {
				t.Fatalf("NewAgentResponse(%s) error = %v", tt.state, err)
			}
			if response.Outcome != tt.outcome || response.Retryable != tt.retryable {
				t.Fatalf("state %s normalized to outcome=%s retryable=%v", tt.state, response.Outcome, response.Retryable)
			}
			actionKind := ""
			if response.NextAction != nil {
				actionKind = response.NextAction.Kind
			}
			if actionKind != tt.actionKind {
				t.Fatalf("state %s action kind = %q, want %q", tt.state, actionKind, tt.actionKind)
			}
			if (response.Outcome == "failed" || response.Outcome == "blocked") && actionKind == "apply" {
				t.Fatalf("state %s exposed false-success apply action", tt.state)
			}
			if response.Skill.Name != "Cabinet Inventory Search" || response.Source.Surface != "chats.main" || response.Source.Channel != "in-app" {
				t.Fatalf("state %s lost governed provenance: %+v", tt.state, response)
			}
		})
	}
}

func TestAgentResponseRetryRequiresBoundedOriginalIntent(t *testing.T) {
	t.Parallel()
	if _, err := NewAgentResponse(AgentResponseRetryableFailure, "failed", "", "skill", "Skill", "chats.main", "in-app"); err == nil {
		t.Fatal("retryable state accepted without bounded original intent")
	}
}

func TestAgentResponseSetupRequiredRoutesToAssistantProviderConfiguration(t *testing.T) {
	t.Parallel()
	response, err := NewAgentResponse(
		AgentResponseSetupRequired,
		"assistant provider setup is required",
		"configure Cabinet Chat",
		"assistant.agent_planner",
		"Cabinet Agent Planner",
		"chats.main",
		"in-app",
	)
	if err != nil {
		t.Fatalf("NewAgentResponse(setup_required) error = %v", err)
	}
	if response.NextAction == nil {
		t.Fatal("setup-required response omitted its setup action")
	}
	if response.NextAction.Kind != "open_setup" || response.NextAction.Route != "/integrations?provider=openai" {
		t.Fatalf("setup-required action = %+v, want the OpenAI provider configuration deep link", response.NextAction)
	}
}

func TestAgentResponseRejectsUnknownState(t *testing.T) {
	t.Parallel()
	if _, err := NewAgentResponse(AgentResponseState("completed_somehow"), "unsafe", "", "skill", "Skill", "chats.main", "in-app"); err == nil {
		t.Fatal("unknown response state was accepted")
	}
}

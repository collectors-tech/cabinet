package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type AssistantTurnMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AssistantTurnRequest struct {
	ProfileID string                 `json:"profile_id"`
	ThreadID  string                 `json:"thread_id"`
	Provider  string                 `json:"provider"`
	Model     string                 `json:"model"`
	Messages  []AssistantTurnMessage `json:"messages"`
	Context   map[string]any         `json:"context,omitempty"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
}

type AssistantTurnResponse struct {
	Provider        string            `json:"provider"`
	Model           string            `json:"model"`
	Text            string            `json:"text"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	ErrorClass      string            `json:"error_class,omitempty"`
	SetupNextAction string            `json:"setup_next_action,omitempty"`
}

type AssistantTurnProvider interface {
	Name() string
	RunAssistantTurn(ctx context.Context, req AssistantTurnRequest) (AssistantTurnResponse, error)
}

type AssistantProviderSetup struct {
	ProviderID        string
	Enabled           bool
	ActiveAuthMethod  string
	DefaultModel      string
	APIKeySecretRef   string
	HealthState       string
	IntegrationMode   string
	IntegrationID     string
	ConfigSchemaRef   string
	WorkflowReference string
}

type AssistantProviderSetupResolver interface {
	ResolveAssistantProviderSetup(ctx context.Context, profileID, providerID string) (AssistantProviderSetup, error)
	GetAssistantProviderSecret(ctx context.Context, profileID, secretRef string) (string, error)
}

type FakeAssistantProvider struct {
	provider string
	model    string
}

func NewFakeAssistantProvider() *FakeAssistantProvider {
	return &FakeAssistantProvider{
		provider: "fake",
		model:    "fake-assistant-model",
	}
}

func (p *FakeAssistantProvider) Name() string {
	if p == nil || strings.TrimSpace(p.provider) == "" {
		return "fake"
	}
	return p.provider
}

func (p *FakeAssistantProvider) RunAssistantTurn(ctx context.Context, req AssistantTurnRequest) (AssistantTurnResponse, error) {
	if err := ctx.Err(); err != nil {
		return AssistantTurnResponse{}, err
	}
	if strings.TrimSpace(req.ProfileID) == "" {
		return AssistantTurnResponse{}, errors.New("assistant turn profile_id is required")
	}
	if strings.TrimSpace(req.ThreadID) == "" {
		return AssistantTurnResponse{}, errors.New("assistant turn thread_id is required")
	}
	model := strings.TrimSpace(req.Model)
	if model == "" && p != nil {
		model = p.model
	}
	if model == "" {
		model = "fake-assistant-model"
	}
	prompt := lastAssistantUserMessage(req.Messages)
	if prompt == "" {
		prompt = "empty prompt"
	}
	return AssistantTurnResponse{
		Provider: p.Name(),
		Model:    model,
		Text:     fmt.Sprintf("Fake assistant response for %s/%s: %s", strings.TrimSpace(req.ProfileID), strings.TrimSpace(req.ThreadID), prompt),
		Metadata: map[string]string{
			"network":       "disabled",
			"test_provider": "true",
		},
	}, nil
}

func lastAssistantUserMessage(messages []AssistantTurnMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if strings.EqualFold(strings.TrimSpace(messages[i].Role), "user") {
			return strings.TrimSpace(messages[i].Content)
		}
	}
	return ""
}

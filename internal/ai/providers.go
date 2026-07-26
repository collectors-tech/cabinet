package ai

import (
	"context"
	"fmt"
	"strings"
)

type Provider interface {
	Name() string
	TestConnectivity(ctx context.Context, apiKey string) error
	SuggestFromTitle(ctx context.Context, apiKey, title string) (Suggestion, error)
	SuggestFromPhoto(ctx context.Context, apiKey, imageURL string) (Suggestion, error)
}

type PlaceholderProvider struct {
	name string
}

func (p PlaceholderProvider) Name() string { return p.name }
func (p PlaceholderProvider) TestConnectivity(context.Context, string) error {
	return fmt.Errorf("%s provider is disabled", p.name)
}
func (p PlaceholderProvider) SuggestFromTitle(context.Context, string, string) (Suggestion, error) {
	return Suggestion{}, fmt.Errorf("%s provider is disabled", p.name)
}
func (p PlaceholderProvider) SuggestFromPhoto(context.Context, string, string) (Suggestion, error) {
	return Suggestion{}, fmt.Errorf("%s provider is disabled", p.name)
}

type Registry struct {
	providers map[string]Provider
}

func NewRegistry(openAI *Service) *Registry {
	out := &Registry{providers: map[string]Provider{}}
	out.providers["openai"] = openAI
	out.providers["anthropic"] = PlaceholderProvider{name: "anthropic"}
	out.providers["google"] = PlaceholderProvider{name: "google"}
	return out
}

func (r *Registry) Provider(name string) (Provider, bool) {
	p, ok := r.providers[strings.ToLower(strings.TrimSpace(name))]
	return p, ok
}

type AssistantProviderRegistry struct {
	providers map[string]AssistantTurnProvider
}

func NewAssistantProviderRegistry(openAI AssistantTurnProvider) *AssistantProviderRegistry {
	out := &AssistantProviderRegistry{providers: map[string]AssistantTurnProvider{}}
	if openAI != nil {
		out.providers["openai"] = openAI
	}
	out.providers["anthropic"] = PlaceholderAssistantProvider{name: "anthropic"}
	out.providers["google"] = PlaceholderAssistantProvider{name: "google"}
	return out
}

func (r *AssistantProviderRegistry) Provider(name string) (AssistantTurnProvider, bool) {
	if r == nil {
		return nil, false
	}
	p, ok := r.providers[strings.ToLower(strings.TrimSpace(name))]
	return p, ok
}

type PlaceholderAssistantProvider struct {
	name string
}

func (p PlaceholderAssistantProvider) Name() string {
	if name := strings.TrimSpace(p.name); name != "" {
		return name
	}
	return "placeholder"
}

func (p PlaceholderAssistantProvider) RunAssistantTurn(context.Context, AssistantTurnRequest) (AssistantTurnResponse, error) {
	name := p.Name()
	return AssistantTurnResponse{
		Provider:        name,
		ErrorClass:      "adapter_unavailable",
		SetupNextAction: "wait_for_supported_assistant_provider_adapter",
		Metadata: map[string]string{
			"live_provider":      "false",
			"adapter_supported":  "false",
			"placeholder_status": "disabled",
		},
	}, fmt.Errorf("%s assistant provider adapter is not supported", name)
}

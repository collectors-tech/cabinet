package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/collectors-tech/cabinet/internal/ai"
	"github.com/collectors-tech/cabinet/internal/profile"
)

type profileAssistantProviderSetupResolver struct {
	profiles *profile.Repository
}

func newProfileAssistantProviderSetupResolver(profiles *profile.Repository) profileAssistantProviderSetupResolver {
	return profileAssistantProviderSetupResolver{profiles: profiles}
}

func (r profileAssistantProviderSetupResolver) ResolveAssistantProviderSetup(ctx context.Context, profileID, providerID string) (ai.AssistantProviderSetup, error) {
	providerID = strings.ToLower(strings.TrimSpace(providerID))
	if providerID != "openai" {
		return ai.AssistantProviderSetup{}, fmt.Errorf("assistant provider %q is not supported", providerID)
	}
	if r.profiles == nil {
		return ai.AssistantProviderSetup{}, fmt.Errorf("profile repository is required")
	}
	settings, err := r.profiles.GetSettings(ctx, profileID)
	if err != nil {
		return ai.AssistantProviderSetup{}, err
	}
	instances, err := r.profiles.ListIntegrationInstances(ctx, profileID)
	if err != nil {
		return ai.AssistantProviderSetup{}, err
	}
	instance, ok := selectOpenAIAssistantIntegrationInstance(instances)
	if !ok {
		return ai.AssistantProviderSetup{}, fmt.Errorf("openai assistant integration instance is required")
	}
	activeMethod := firstNonEmpty(instance.Config["openai.active_auth_method"], settings["openai.active_auth_method"], settings["openai_active_auth_method"])
	defaultModel := firstNonEmpty(instance.Config["assistant_default_model"], settings["assistant_default_model"], "gpt-4o-mini")
	baseURL := firstNonEmpty(instance.Config["openai_base_url"], instance.Config["base_url"], settings["openai_base_url"], settings["integration.openai.base_url"])
	return ai.AssistantProviderSetup{
		ProviderID:        "openai",
		Enabled:           instance.Enabled,
		ActiveAuthMethod:  activeMethod,
		DefaultModel:      defaultModel,
		SupportedModels:   openAIAssistantSupportedModels(),
		BaseURL:           baseURL,
		APIKeySecretRef:   strings.TrimSpace(instance.SecretRefs["openai_api_key"]),
		HealthState:       firstNonEmpty(instance.HealthState, "unknown"),
		IntegrationMode:   "assistant_workflows",
		IntegrationID:     strings.TrimSpace(instance.ID),
		ConfigSchemaRef:   "integrations/openai/auth",
		WorkflowReference: "assistant.chat",
	}, nil
}

func (r profileAssistantProviderSetupResolver) GetAssistantProviderSecret(ctx context.Context, profileID, secretRef string) (string, error) {
	if r.profiles == nil {
		return "", fmt.Errorf("profile repository is required")
	}
	secretRef = strings.TrimSpace(secretRef)
	if secretRef == "" {
		return "", fmt.Errorf("assistant provider secret ref is required")
	}
	return r.profiles.GetSecret(ctx, profileID, secretRef)
}

func selectOpenAIAssistantIntegrationInstance(instances []profile.IntegrationInstance) (profile.IntegrationInstance, bool) {
	for _, instance := range instances {
		if instance.Enabled && strings.EqualFold(strings.TrimSpace(instance.ProviderID), "openai") {
			return instance, true
		}
	}
	return profile.IntegrationInstance{}, false
}

func openAIAssistantSupportedModels() []string {
	schema, ok := providerConfigSchemaForRef("integrations/openai/auth")
	if !ok {
		return []string{"gpt-4o-mini"}
	}
	for _, field := range assistantSetupSchemaFields(schema["fields"]) {
		if strings.TrimSpace(fmt.Sprintf("%v", field["key"])) != "assistant_default_model" {
			continue
		}
		models := []string{}
		for _, option := range assistantSetupSchemaOptions(field["options"]) {
			if value := strings.TrimSpace(option); value != "" {
				models = append(models, value)
			}
		}
		if len(models) > 0 {
			return models
		}
	}
	return []string{"gpt-4o-mini"}
}

func assistantSetupSchemaFields(value any) []map[string]any {
	switch typed := value.(type) {
	case []map[string]any:
		return typed
	case []any:
		fields := make([]map[string]any, 0, len(typed))
		for _, raw := range typed {
			if field, ok := raw.(map[string]any); ok {
				fields = append(fields, field)
			}
		}
		return fields
	default:
		return nil
	}
}

func assistantSetupSchemaOptions(value any) []string {
	switch typed := value.(type) {
	case []map[string]string:
		options := make([]string, 0, len(typed))
		for _, option := range typed {
			options = append(options, option["value"])
		}
		return options
	case []map[string]any:
		options := make([]string, 0, len(typed))
		for _, option := range typed {
			options = append(options, fmt.Sprintf("%v", option["value"]))
		}
		return options
	case []any:
		options := []string{}
		for _, raw := range typed {
			switch option := raw.(type) {
			case map[string]any:
				options = append(options, fmt.Sprintf("%v", option["value"]))
			case map[string]string:
				options = append(options, option["value"])
			}
		}
		return options
	default:
		return nil
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

package app

import (
	"context"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/profile"
)

func TestProfileAssistantProviderResolverSelectsOpenAIIntegrationInstance(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileID := createIntegrationInstanceProfile(t, a, "Assistant Provider Resolver")
	repo := profile.NewRepository(a.db)
	if err := repo.PutSettings(context.Background(), profileID, map[string]string{
		"openai.active_auth_method":     "api_key",
		"assistant_default_provider":    "openai",
		"assistant_default_model":       "gpt-4o-mini",
		"openai.api_key_secret_present": "true",
	}); err != nil {
		t.Fatalf("put settings: %v", err)
	}
	enabled := true
	instance, err := repo.UpsertIntegrationInstance(context.Background(), profileID, profile.IntegrationInstancePatch{
		ProviderID:  "openai",
		DisplayName: stringPtr("OpenAI runtime"),
		Enabled:     &enabled,
		Config: map[string]string{
			"assistant_default_model": "gpt-4.1-mini",
		},
		Secrets: map[string]string{
			"openai_api_key": "sk-runtime-boundary-secret",
		},
		AuthState:   stringPtr("configured"),
		HealthState: stringPtr("ready"),
	})
	if err != nil {
		t.Fatalf("upsert integration instance: %v", err)
	}

	resolver := newProfileAssistantProviderSetupResolver(repo)
	setup, err := resolver.ResolveAssistantProviderSetup(context.Background(), profileID, "openai")
	if err != nil {
		t.Fatalf("resolve setup: %v", err)
	}

	if setup.ProviderID != "openai" || setup.IntegrationID != instance.ID || !setup.Enabled {
		t.Fatalf("expected enabled OpenAI integration instance, got %+v", setup)
	}
	if setup.ActiveAuthMethod != "api_key" || setup.DefaultModel != "gpt-4.1-mini" || setup.HealthState != "ready" {
		t.Fatalf("expected setup to merge profile auth and instance model/health, got %+v", setup)
	}
	if setup.APIKeySecretRef != instance.SecretRefs["openai_api_key"] || setup.APIKeySecretRef == "" {
		t.Fatalf("expected setup to expose only instance secret ref, got setup=%+v refs=%+v", setup, instance.SecretRefs)
	}
	if strings.Contains(setup.APIKeySecretRef, "sk-runtime-boundary-secret") {
		t.Fatalf("setup leaked secret material: %+v", setup)
	}
	secret, err := resolver.GetAssistantProviderSecret(context.Background(), profileID, setup.APIKeySecretRef)
	if err != nil {
		t.Fatalf("resolve secret: %v", err)
	}
	if secret != "sk-runtime-boundary-secret" {
		t.Fatalf("expected resolver to fetch profile secret by ref, got %q", secret)
	}
}

func TestProfileAssistantProviderResolverRejectsUnavailableOpenAIInstance(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileID := createIntegrationInstanceProfile(t, a, "Assistant Provider Resolver Missing")
	resolver := newProfileAssistantProviderSetupResolver(profile.NewRepository(a.db))
	if _, err := resolver.ResolveAssistantProviderSetup(context.Background(), profileID, "openai"); err == nil {
		t.Fatal("expected missing OpenAI integration instance error")
	}
	if _, err := resolver.ResolveAssistantProviderSetup(context.Background(), profileID, "anthropic"); err == nil {
		t.Fatal("expected unsupported assistant provider error")
	}
}

func stringPtr(value string) *string {
	return &value
}

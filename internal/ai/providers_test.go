package ai

import (
	"context"
	"testing"
)

func TestProviderRegistryIncludesPlaceholders(t *testing.T) {
	t.Parallel()
	reg := NewRegistry(NewService(Config{BaseURL: "http://example"}))
	if _, ok := reg.Provider("openai"); !ok {
		t.Fatal("expected openai provider")
	}
	p, ok := reg.Provider("anthropic")
	if !ok {
		t.Fatal("expected anthropic placeholder provider")
	}
	if err := p.TestConnectivity(context.Background(), "x"); err == nil {
		t.Fatal("expected disabled placeholder error")
	}
}

func TestAssistantProviderRegistryKeepsPlaceholdersUnavailable(t *testing.T) {
	t.Parallel()

	reg := NewAssistantProviderRegistry(NewOpenAIAssistantProvider(NewService(Config{BaseURL: "http://example"}), nil))
	for _, providerID := range []string{"anthropic", "google"} {
		provider, ok := reg.Provider(providerID)
		if !ok {
			t.Fatalf("expected %s assistant placeholder provider", providerID)
		}
		if provider.Name() != providerID {
			t.Fatalf("expected %s provider name, got %q", providerID, provider.Name())
		}
		resp, err := provider.RunAssistantTurn(context.Background(), AssistantTurnRequest{
			ProfileID: "profile-1",
			ThreadID:  "thread-1",
			Provider:  providerID,
			Messages:  []AssistantTurnMessage{{Role: "user", Content: "hello"}},
		})
		if err == nil {
			t.Fatalf("expected %s placeholder to reject assistant turns", providerID)
		}
		if resp.Provider != providerID || resp.ErrorClass != "adapter_unavailable" || resp.SetupNextAction != "wait_for_supported_assistant_provider_adapter" {
			t.Fatalf("expected %s unavailable guidance without OpenAI fallback, got resp=%+v err=%v", providerID, resp, err)
		}
		if resp.Metadata["fallback_provider"] != "" || resp.Metadata["live_provider"] != "false" {
			t.Fatalf("expected %s placeholder to avoid fallback/live readiness claims, got %+v", providerID, resp.Metadata)
		}
	}
}

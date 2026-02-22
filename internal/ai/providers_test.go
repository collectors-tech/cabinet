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

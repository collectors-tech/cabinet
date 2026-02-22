package scanner

import (
	"context"
	"testing"
)

type noOpProvider struct{}

func (noOpProvider) Search(context.Context, QuerySet) ([]CandidateInput, error) {
	return []CandidateInput{}, nil
}

func TestProviderRegistrySupportsAdditionalSources(t *testing.T) {
	t.Parallel()
	reg := NewProviderRegistry()
	reg.Register("ebay", func() Provider { return noOpProvider{} })
	p, err := reg.Provider("ebay")
	if err != nil {
		t.Fatalf("Provider(ebay) error = %v", err)
	}
	if _, err := p.Search(context.Background(), QuerySet{}); err != nil {
		t.Fatalf("search ebay provider error = %v", err)
	}
	disabled, err := reg.Provider("mercari")
	if err != nil {
		t.Fatalf("Provider(mercari) error = %v", err)
	}
	if _, err := disabled.Search(context.Background(), QuerySet{}); err == nil {
		t.Fatal("expected disabled provider error")
	}
}

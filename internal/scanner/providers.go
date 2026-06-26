package scanner

import (
	"context"
	"fmt"
	"strings"
)

type ProviderFactory func() Provider

type ProviderRegistry struct {
	factories map[string]ProviderFactory
}

func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{factories: map[string]ProviderFactory{}}
}

func (r *ProviderRegistry) Register(name string, factory ProviderFactory) {
	r.factories[strings.ToLower(strings.TrimSpace(name))] = factory
}

func (r *ProviderRegistry) Provider(name string) (Provider, error) {
	f, ok := r.factories[strings.ToLower(strings.TrimSpace(name))]
	if !ok {
		return DisabledProvider{name: name}, nil
	}
	return f(), nil
}

type DisabledProvider struct {
	name string
}

func (p DisabledProvider) ProviderID() string {
	return p.name
}

func (p DisabledProvider) Search(context.Context, QuerySet) ([]CandidateInput, error) {
	return nil, fmt.Errorf("%s provider is disabled", p.name)
}

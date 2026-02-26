package app

import (
	"os"
	"strings"
	"testing"
)

func TestProviderSpecsExistAndRegistryLinksThem(t *testing.T) {
	t.Parallel()

	requiredSpecs := []string{
		"../../openspec/specs/provider-registry/spec.md",
		"../../openspec/specs/provider-ebay/spec.md",
		"../../openspec/specs/provider-amazon/spec.md",
		"../../openspec/specs/provider-au-webshops/spec.md",
	}

	for _, path := range requiredSpecs {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("missing provider spec: %s (%v)", path, err)
		}
	}

	registryBytes, err := os.ReadFile("../../openspec/specs/provider-registry/spec.md")
	if err != nil {
		t.Fatalf("read provider registry spec: %v", err)
	}
	registry := string(registryBytes)

	requiredRefs := []string{
		"provider-ebay",
		"provider-amazon",
		"provider-au-webshops",
		"bonzaslotcars.com.au",
		"frontlinehobbies.com.au",
		"hobbytechtoys.com.au",
		"andrewshobbies.com.au",
		"voglers.com.au",
		"acercmodels.com",
		"mrtoys.com.au",
	}
	for _, token := range requiredRefs {
		if !strings.Contains(registry, token) {
			t.Fatalf("provider registry missing token: %s", token)
		}
	}
}


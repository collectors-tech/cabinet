package app

import (
	"os"
	"strings"
	"testing"
)

func TestProviderSpecsExistAndRegistryLinksThem(t *testing.T) {
	t.Parallel()

	requiredSpecs := []string{
		"../../openspec/specs/integrations/provider-registry/spec.md",
		"../../openspec/specs/integrations/provider-ebay/spec.md",
		"../../openspec/specs/integrations/provider-amazon/spec.md",
		"../../openspec/specs/integrations/provider-au-webshops/spec.md",
	}

	for _, path := range requiredSpecs {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("missing provider spec: %s (%v)", path, err)
		}
	}

	registryBytes, err := os.ReadFile("../../openspec/specs/integrations/provider-registry/spec.md")
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

func TestProviderFamilyContractsAreIndexedAndLinkedFromAUProviderSpec(t *testing.T) {
	t.Parallel()

	familySpecPath := "../../openspec/specs/integrations/provider-api-families/spec.md"
	if _, err := os.Stat(familySpecPath); err != nil {
		t.Fatalf("missing provider family spec: %s (%v)", familySpecPath, err)
	}
	readmeBytes, err := os.ReadFile("../../openspec/specs/integrations/README.md")
	if err != nil {
		t.Fatalf("read integrations README: %v", err)
	}
	readme := string(readmeBytes)
	if !strings.Contains(readme, "provider-api-families") {
		t.Fatalf("integrations README missing provider-api-families index reference")
	}

	auSpecBytes, err := os.ReadFile("../../openspec/specs/integrations/provider-au-webshops/spec.md")
	if err != nil {
		t.Fatalf("read AU webshop provider spec: %v", err)
	}
	auSpec := string(auSpecBytes)
	requiredTokens := []string{
		"provider-api-families/spec.md",
		"PROVIDER-FAMILY-001",
		"PROVIDER-FAMILY-002",
		"PROVIDER-FAMILY-003",
		"PROVIDER-FAMILY-004",
	}
	for _, token := range requiredTokens {
		if !strings.Contains(auSpec, token) {
			t.Fatalf("AU webshop provider spec missing provider-family token: %s", token)
		}
	}
}

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

func TestIntegrationRegistryOpenSpecCoversIssue1469Contracts(t *testing.T) {
	t.Parallel()

	registryBytes, err := os.ReadFile("../../openspec/specs/integrations/provider-registry/spec.md")
	if err != nil {
		t.Fatalf("read provider registry spec: %v", err)
	}
	registry := string(registryBytes)

	requiredSpecTokens := []string{
		"INTEGRATION-027",
		"provider manifests",
		"setup schema",
		"Schema-driven Add Integration setup",
		"write-only field metadata",
		"INTEGRATION-028",
		"workflow/action metadata",
		"remote writes that require explicit confirmation",
		"INTEGRATION-029",
		"profile-scoped integration instance",
		"required-action code",
		"INTEGRATION-030",
		"Notification Inbox event",
		"Provider setup schemas and Add Integration form rendering",
	}
	for _, token := range requiredSpecTokens {
		if !strings.Contains(registry, token) {
			t.Fatalf("provider registry spec missing #1469 coverage token: %s", token)
		}
	}

	traceabilityBytes, err := os.ReadFile("../../openspec/traceability.md")
	if err != nil {
		t.Fatalf("read OpenSpec traceability: %v", err)
	}
	traceability := string(traceabilityBytes)

	requiredTraceabilityTokens := []string{
		"`INTEGRATION-027`",
		"`INTEGRATION-028`",
		"`INTEGRATION-029`",
		"`INTEGRATION-030`",
		"#1469",
		"targeted Cypress Add Integration provider-selection/setup-schema rendering",
		"Go status/inbox API tests",
	}
	for _, token := range requiredTraceabilityTokens {
		if !strings.Contains(traceability, token) {
			t.Fatalf("traceability missing #1469 coverage token: %s", token)
		}
	}
}

func TestIntegrationRegistryOpenSpecCoversIssue1463ConsumerContract(t *testing.T) {
	t.Parallel()

	registryBytes, err := os.ReadFile("../../openspec/specs/integrations/provider-registry/spec.md")
	if err != nil {
		t.Fatalf("read provider registry spec: %v", err)
	}
	registry := string(registryBytes)

	requiredSpecTokens := []string{
		"INTEGRATION-063",
		"canonical registry definition",
		"/api/providers/registry",
		"/api/providers/:id/*",
		"Add Integration UI list",
		"Market Watch provider projection",
		"marketplace",
		"storefront/source matcher",
		"browser-auth",
		"chat/AI",
		"notification",
		"workflow/local",
		"config form requirements",
		"health/diagnostics",
		"matching/import/export support",
		"browser-auth/external-login behavior",
	}
	for _, token := range requiredSpecTokens {
		if !strings.Contains(registry, token) {
			t.Fatalf("provider registry spec missing #1463 consumer/category token: %s", token)
		}
	}

	traceabilityBytes, err := os.ReadFile("../../openspec/traceability.md")
	if err != nil {
		t.Fatalf("read OpenSpec traceability: %v", err)
	}
	traceability := string(traceabilityBytes)

	requiredTraceabilityTokens := []string{
		"`INTEGRATION-063`",
		"#1463",
		"canonical provider registry consumers",
		"marketplace, storefront/source matcher, browser-auth, chat/AI, notification, and workflow/local provider categories",
		"TestIntegrationRegistryOpenSpecCoversIssue1463ConsumerContract",
	}
	for _, token := range requiredTraceabilityTokens {
		if !strings.Contains(traceability, token) {
			t.Fatalf("traceability missing #1463 consumer/category token: %s", token)
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

func TestBigCommerceVoglersIssue1497Traceability(t *testing.T) {
	t.Parallel()

	familySpecBytes, err := os.ReadFile("../../openspec/specs/integrations/provider-api-families/spec.md")
	if err != nil {
		t.Fatalf("read provider family spec: %v", err)
	}
	familySpec := string(familySpecBytes)
	for _, token := range []string{
		"PROVIDER-FAMILY-006",
		"BigCommerce public/storefront-access run",
		"BigCommerce token-enabled run",
		"storefront-accessible endpoints/content paths",
		"capability limits",
	} {
		if !strings.Contains(familySpec, token) {
			t.Fatalf("provider family spec missing #1497 BigCommerce token: %s", token)
		}
	}

	traceabilityBytes, err := os.ReadFile("../../openspec/traceability.md")
	if err != nil {
		t.Fatalf("read OpenSpec traceability: %v", err)
	}
	traceability := string(traceabilityBytes)
	for _, token := range []string{
		"`PROVIDER-FAMILY-006`",
		"#1497",
		"Voglers",
		"bigcommerce_storefront_success.json",
		"bigcommerce_graphql_stock_success.json",
		"TestShoppingProviderFixturesNormalizeSharedCandidateShape",
		"TestShoppingProviderFixturesPreserveAvailabilitySignals",
	} {
		if !strings.Contains(traceability, token) {
			t.Fatalf("traceability missing #1497 BigCommerce coverage token: %s", token)
		}
	}
}

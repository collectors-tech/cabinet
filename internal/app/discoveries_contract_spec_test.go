package app

import (
	"os"
	"strings"
	"testing"
)

func TestDiscoveriesPurposeAndHandoffSpecContracts(t *testing.T) {
	t.Parallel()

	domainBytes, err := os.ReadFile("../../openspec/specs/dashboard/discovery/spec.md")
	if err != nil {
		t.Fatalf("read discovery domain spec: %v", err)
	}
	screenBytes, err := os.ReadFile("../../openspec/specs/dashboard/ui-screen-discover/spec.md")
	if err != nil {
		t.Fatalf("read discovery screen spec: %v", err)
	}
	traceBytes, err := os.ReadFile("../../openspec/traceability.md")
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	combined := string(domainBytes) + "\n" + string(screenBytes) + "\n" + string(traceBytes)
	requiredTokens := []string{
		"DISCOVERY-003",
		"DISCOVERY-004",
		"UI-SCREEN-DISCOVER-005",
		"candidate-item triage inbox",
		"source/provider identifier",
		"source query or run identifier",
		"source-result link",
		"`new`, `reviewing`, `wishlisted`, `inventory_candidate`, `purchase_candidate`, `ignored`, or `archived`",
		"Wishlist promotion MUST create or link a Wishlist entry without claiming ownership",
		"Market Watch and provider workflows own query",
		"planned Cypress `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts`",
	}
	for _, token := range requiredTokens {
		if !strings.Contains(combined, token) {
			t.Fatalf("Discoveries contract spec missing token %q", token)
		}
	}
}

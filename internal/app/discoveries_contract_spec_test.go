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
		"DISCOVERY-005",
		"DISCOVERY-006",
		"UI-SCREEN-DISCOVER-005",
		"UI-SCREEN-DISCOVER-007",
		"candidate-item triage inbox",
		"source/provider identifier",
		"source query or run identifier",
		"source-result link",
		"`new`, `reviewing`, `wishlisted`, `inventory_candidate`, `purchase_candidate`, `ignored`, or `archived`",
		"`new`, `reviewing`, `wishlisted`, `purchase_candidate`, `inventory_candidate`, `ignored`, or `archived`",
		"Wishlist promotion MUST create or link a Wishlist entry without claiming ownership",
		"Market Watch and provider workflows own query",
		"found-opportunity dashboard",
		"best deals, wishlist matches, new findings, Market Watch outputs, and provider/store attention",
		"other public or shared inventories",
		"Market Watch query controls from Discoveries output review",
		"Wishlist price-match candidate",
		"Market Watch output candidate",
		"private collector inventory",
		"private collector inventory records, private notes, storage locations, unpublished collection values",
		"Cypress `ui.web/cypress/e2e/dashboard/ui-screen-discover/spec.cy.ts` (`UI-SCREEN-DISCOVER-005 renders candidate provenance and destination actions`)",
	}
	for _, token := range requiredTokens {
		if !strings.Contains(combined, token) {
			t.Fatalf("Discoveries contract spec missing token %q", token)
		}
	}
}

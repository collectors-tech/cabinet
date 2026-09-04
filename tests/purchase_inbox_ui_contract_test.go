package tests

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPurchaseInboxUIContract(t *testing.T) {
	t.Parallel()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file location: runtime.Caller failed")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), ".."))
	featurePath := filepath.Join(repoRoot, "ui.web", "src", "features", "purchases", "index.tsx")
	purchasesRoutePath := filepath.Join(repoRoot, "ui.web", "src", "routes", "_authenticated", "purchases", "index.tsx")
	cypressPath := filepath.Join(repoRoot, "ui.web", "cypress", "e2e", "purchases", "purchase-inbox", "spec.cy.ts")

	feature := readContractFile(t, featurePath)
	purchasesRoute := readContractFile(t, purchasesRoutePath)
	cypressSpec := readContractFile(t, cypressPath)

	requiredFeatureSnippets := []string{
		"/api/integrations/ebay/purchase-inbox/reviews",
		"purchase-inbox-empty-state",
		"purchase-inbox-loading-state",
		"purchase-inbox-error-state",
		"purchase-inbox-ready-state",
		"purchase-inbox-confirm-dialog",
		"purchases-header-title",
		"purchases-page-icon",
		"purchases-global-header-actions",
		"requires_confirmation",
		"Confirmation required",
		"purchases-add-csv-preview",
		"purchases-add-email-preview",
		"purchases-add-csv-confirm",
		"purchases-add-email-confirm",
		"/api/commerce/lifecycle",
		"purchases-row-persistence",
		"purchases-row-purchase-date",
		"purchases-row-delivery",
		"purchases-row-order-link",
		"capturedReviewsOpen",
		"sourceMatchesOpen",
		"purchase-review-tools",
		"purchases-split-pane",
		"purchases-detail-pane",
		"purchases-order-detail",
		"purchases-item-detail",
	}
	for _, snippet := range requiredFeatureSnippets {
		if !strings.Contains(feature, snippet) {
			t.Fatalf("Purchases UI missing %q in %s", snippet, featurePath)
		}
	}
	forbiddenFeatureSnippets := []string{
		"data-testid='purchases-source-matches-toggle'",
		"data-testid='purchase-inbox-load-reviews'",
		"aria-label='Review source matches'",
		"aria-label='Review captured purchases'",
	}
	for _, snippet := range forbiddenFeatureSnippets {
		if strings.Contains(feature, snippet) {
			t.Fatalf("Purchases UI must not expose rejected header action %q in %s", snippet, featurePath)
		}
	}

	if !strings.Contains(purchasesRoute, "@/features/purchases") || !strings.Contains(purchasesRoute, "Purchases") {
		t.Fatalf("/purchases route must render Purchases feature, got %s", purchasesRoute)
	}

	requiredCypressSnippets := []string{
		"COMMERCE-RECONCILIATION-006",
		"purchases-global-header-actions",
		"COMMERCE-RECONCILIATION-009",
		"COMMERCE-RECONCILIATION-010",
		"COMMERCE-RECONCILIATION-011",
		"COMMERCE-RECONCILIATION-014",
		"purchases-detail-pane",
		"path: '/purchases'",
		"EBAY-PURCHASE-CAPTURE-006",
		"purchase-inbox-empty-state",
		"purchase-inbox-error-state",
		"purchase-inbox-confirm-action",
		"forwarder-package-inbox",
		"should('not.exist')",
	}
	for _, snippet := range requiredCypressSnippets {
		if !strings.Contains(cypressSpec, snippet) {
			t.Fatalf("Purchases Cypress contract missing %q in %s", snippet, cypressPath)
		}
	}
}

func readContractFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

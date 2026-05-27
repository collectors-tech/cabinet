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
	routePath := filepath.Join(repoRoot, "ui.web", "src", "routes", "_authenticated", "inbox", "index.tsx")
	cypressPath := filepath.Join(repoRoot, "ui.web", "cypress", "e2e", "purchases", "purchase-inbox", "spec.cy.ts")

	feature := readContractFile(t, featurePath)
	route := readContractFile(t, routePath)
	cypressSpec := readContractFile(t, cypressPath)

	requiredFeatureSnippets := []string{
		"/api/integrations/ebay/purchase-inbox/reviews",
		"purchase-inbox-empty-state",
		"purchase-inbox-loading-state",
		"purchase-inbox-error-state",
		"purchase-inbox-ready-state",
		"purchase-inbox-confirm-dialog",
		"requires_confirmation",
		"Confirmation required",
	}
	for _, snippet := range requiredFeatureSnippets {
		if !strings.Contains(feature, snippet) {
			t.Fatalf("Purchase Inbox UI missing %q in %s", snippet, featurePath)
		}
	}

	if !strings.Contains(route, "@/features/purchases") || !strings.Contains(route, "Purchases") {
		t.Fatalf("/inbox route must render Purchase Inbox feature, got %s", route)
	}

	requiredCypressSnippets := []string{
		"EBAY-PURCHASE-CAPTURE-006",
		"purchase-inbox-empty-state",
		"purchase-inbox-error-state",
		"purchase-inbox-confirm-action",
	}
	for _, snippet := range requiredCypressSnippets {
		if !strings.Contains(cypressSpec, snippet) {
			t.Fatalf("Purchase Inbox Cypress contract missing %q in %s", snippet, cypressPath)
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

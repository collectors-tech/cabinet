package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEbayBuyerInterestTraceabilityImplemented(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	rows := map[string]string{}
	for _, line := range strings.Split(string(raw), "\n") {
		for _, id := range []string{"INTEGRATION-025", "INTEGRATION-026"} {
			if strings.HasPrefix(line, "| `"+id+"` ") {
				rows[id] = line
			}
		}
	}

	requiredByID := map[string][]string{
		"INTEGRATION-025": {
			"internal/ebay/buyer_interest.go",
			"/api/providers/ebay/buyer-interest/preview",
			"/api/providers/ebay/buyer-interest/import",
			"TestMapBuyerInterestPreservesProvenanceAndDestinations",
			"TestEbayBuyerInterestImportPersistsWishlistAndDiscovery",
			"TestOpenAPIDocumentsEbayBuyerInterestContract",
			"INTEGRATION-025 + INTEGRATION-026: previews and imports eBay buyer-interest sync without remote write-back claims",
			"| implemented |",
		},
		"INTEGRATION-026": {
			"write-back capability/blocker fields",
			"TestMapBuyerInterestRequiresVerifiedAPIForWriteBack",
			"TestEbayBuyerInterestPreviewMapsDestinationsAndWriteBackCapability",
			"TestEbayBuyerInterestPreviewRejectsEmptyItems",
			"TestOpenAPIDocumentsEbayBuyerInterestContract",
			"INTEGRATION-025 + INTEGRATION-026: previews and imports eBay buyer-interest sync without remote write-back claims",
			"| implemented |",
		},
	}

	for id, requiredFragments := range requiredByID {
		row := rows[id]
		if row == "" {
			t.Fatalf("expected traceability row for %s", id)
		}
		for _, fragment := range requiredFragments {
			if !strings.Contains(row, fragment) {
				t.Fatalf("expected %s traceability row to include %q; row: %s", id, fragment, row)
			}
		}
	}
}

func TestEbayProviderTraceabilityImplemented(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	rows := map[string]string{}
	for _, line := range strings.Split(string(raw), "\n") {
		for _, id := range []string{"INTEGRATION-005", "INTEGRATION-006", "INTEGRATION-007"} {
			if strings.HasPrefix(line, "| `"+id+"` ") {
				rows[id] = line
			}
		}
	}

	requiredByID := map[string][]string{
		"INTEGRATION-005": {
			"internal/ebay/provider.go",
			"/api/providers/ebay/run",
			"/api/scanner/query-sets",
			"TestOpenAPIDocumentsEbayProviderRunContract",
			"TestOpenAPIDocumentsEbaySavedSearchHandoffContract",
			"INTEGRATION-005 + #827 manages eBay saved-query create edit schedule and delete lifecycle",
			"TestEbayProviderTraceabilityImplemented",
			"| implemented |",
		},
		"INTEGRATION-006": {
			"/api/provider/health?provider=ebay",
			"TestOpenAPIDocumentsEbayProviderHealthContract",
			"INTEGRATION-006 + #1289: displays eBay provider-health readiness aliases and recovery guidance",
			"UI-SCREEN-SCANNER-002 exposes provider health and failure retry",
			"TestEbayProviderTraceabilityImplemented",
			"| implemented |",
		},
		"INTEGRATION-007": {
			"stock_state",
			"stock_count",
			"TestProviderSearchNormalizesCandidates",
			"TestOpenAPIDocumentsEbaySavedSearchHandoffContract",
			"TestEbayProviderTraceabilityImplemented",
			"| implemented |",
		},
	}

	for id, requiredFragments := range requiredByID {
		row := rows[id]
		if row == "" {
			t.Fatalf("expected traceability row for %s", id)
		}
		for _, fragment := range requiredFragments {
			if !strings.Contains(row, fragment) {
				t.Fatalf("expected %s traceability row to include %q; row: %s", id, fragment, row)
			}
		}
	}
}

func TestEbaySellerOperationsTraceabilityImplemented(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	var row string
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, "| `INTEGRATION-027` ") || strings.HasPrefix(line, "| INTEGRATION-027 ") {
			row = line
			break
		}
	}
	if row == "" {
		t.Fatalf("expected traceability row for INTEGRATION-027")
	}

	requiredFragments := []string{
		"seller operation capability-gated states",
		"UI/API preview/execute workflows",
		"local read-result rendering",
		"TestSellerOperationStatusesDefaultToBlocked",
		"TestPreviewSellerOperationActionRequiresConfirmationForConfirmedWrites",
		"TestExecuteSellerOperationActionCompletesReadOnlySyncLocally",
		"TestSellerOperationReadResultsExposePerOperationModels",
		"TestExecuteSellerOperationActionRefusesRemoteWriteWithoutAdapter",
		"TestEbaySellerOperationPreviewBlocksUnverifiedWrite",
		"TestEbaySellerOperationExecuteRejectsRemoteWriteWithoutAdapter",
		"TestOpenAPIDocumentsEbaySellerOperationsContract",
		"TestIntegrationsEbaySellerOperationsPanelContract",
		"INTEGRATION-027 + #842: previews and executes seller operation sync without remote write claims",
		"TestEbaySellerOperationsTraceabilityImplemented",
		"| implemented |",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(row, fragment) {
			t.Fatalf("expected INTEGRATION-027 traceability row to include %q; row: %s", fragment, row)
		}
	}
}

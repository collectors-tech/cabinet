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
			"/api/providers/registry",
			"/api/scanner/query-sets",
			"trimmed Browse string metadata",
			"trimmed Browse price value",
			"first-parseable shipping-cost normalization",
			"invalid JSON/profile/settings bootstrap client-error diagnostics",
			"missing query-set client-error diagnostics",
			"TestOpenAPIDocumentsEbayProviderRunContract",
			"TestOpenAPIDocumentsEbaySavedSearchHandoffContract",
			"TestEbayRegistryExposesSetupReadinessWithoutCredentialLeak",
			"TestProviderSearchTrimsBrowseStringMetadata",
			"TestProviderSearchTrimsBrowsePriceValue",
			"TestProviderSearchUsesFirstParseableShippingCost",
			"eBay setup UI status panel renders registry `setup_status` readiness",
			"INTEGRATION-005 + #827 manages eBay saved-query create edit schedule and delete lifecycle",
			"omitted provider-scope updates preserving existing eBay scope",
			"TestUpdateQuerySetPreservesProviderScopeWhenOmitted",
			"TestEbayProviderTraceabilityImplemented",
			"| implemented |",
		},
		"INTEGRATION-006": {
			"/api/provider/health?provider=ebay",
			"latest positive eBay Browse `Retry-After` timing",
			"scanner failure list reusable schemas",
			"scanner failure list method-error diagnostics",
			"scanner failure retry request/accepted response fields",
			"TestScannerFailuresRejectsUnsupportedMethodsWithGuidance",
			"TestEbayProviderRunMapsBrowseFailureToProviderHealthGuidance",
			"TestOpenAPIDocumentsEbayProviderHealthContract",
			"TestOpenAPIDocumentsScannerFailureRetryGuidance",
			"TestOpenAPIDocumentsEbayScannerFailureRetryContract",
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
			"INTEGRATION-005 + INTEGRATION-007 + #827 surfaces eBay provider run pagination metadata, observed-currency output, and stock state",
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

func TestEbayScannerFailureListTraceabilityImplemented(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	var row string
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, "| `INTEGRATION-006` ") || strings.HasPrefix(line, "| INTEGRATION-006 ") {
			row = line
			break
		}
	}
	if row == "" {
		t.Fatalf("expected traceability row for INTEGRATION-006")
	}

	requiredFragments := []string{
		"ScannerFailuresResponse",
		"ScannerFailure",
		"query set id, provider, message, raw failure reason, failure timestamps, retry guidance, and next action",
		"TestOpenAPIDocumentsScannerFailureRetryGuidance",
		"TestOpenAPIDocumentsEbayScannerFailureRetryContract",
		"UI-SCREEN-SCANNER-002 exposes provider health and failure retry",
		"TestEbayScannerFailureListTraceabilityImplemented",
		"| implemented |",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(row, fragment) {
			t.Fatalf("expected INTEGRATION-006 traceability row to include %q; row: %s", fragment, row)
		}
	}
}

func TestEbayProviderSearchFailureTraceabilityImplemented(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	var row string
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, "| `INTEGRATION-005` ") || strings.HasPrefix(line, "| INTEGRATION-005 ") {
			row = line
			break
		}
	}
	if row == "" {
		t.Fatalf("expected traceability row for INTEGRATION-005")
	}

	requiredFragments := []string{
		"Browse request headers/query criteria",
		"blank keyword/exclusion trimming",
		"structured auth and non-auth Browse error payload preservation",
		"positive integer `Retry-After` timing",
		"trimmed Browse string metadata",
		"trimmed Browse price value",
		"first-parseable shipping-cost normalization",
		"`/api/scanner/run` provider-error envelope",
		"`/api/providers/ebay/run` saved-search candidate persistence",
		"`/api/providers/registry` eBay setup readiness contract",
		"setup_status.base_url_set",
		"degraded provider-health next action",
		"eBay setup UI status panel renders registry `setup_status` readiness",
		"Wishlist/Inventory provenance handoff",
		"TestProviderSearchTrimsBlankCriteriaBeforeBrowseRequest",
		"TestProviderSearchUsesFirstParseableShippingCost",
		"TestWave4EbayRunMissingQuerySetReturnsActionableClientEnvelope",
		"TestProviderSearchPreservesStructuredBrowseErrorPayload",
		"TestScannerRunMapsEbayBrowseRetryAfterToProviderErrorEnvelope",
		"TestEbayProviderRunMapsBrowseFailureToProviderHealthGuidance",
		"TestOpenAPIDocumentsEbayScannerRunSearchErrorEnvelope",
		"TestProviderSearchTrimsBrowseStringMetadata",
		"TestProviderSearchTrimsBrowsePriceValue",
		"TestOpenAPIDocumentsEbayRegistrySetupReadinessContract",
		"TestEbayRegistrySetupStatusReflectsDegradedProviderHealth",
		"TestWave4EbayRunBootstrapErrorsReturnActionableClientEnvelopes",
		"INTEGRATION-005 + #827 surfaces eBay invalid query-set diagnostics",
		"INTEGRATION-005 + #827 surfaces eBay provider run method diagnostics",
		"method-error diagnostics include provider, next_action, and allowed_method",
		"INTEGRATION-005 + #827 surfaces eBay provider run pagination metadata",
		"INTEGRATION-005 + #827 surfaces eBay retry-after provider run diagnostics",
		"INTEGRATION-005 + #827 manages eBay saved-query create edit schedule and delete lifecycle",
		"omitted provider-scope updates preserving existing eBay scope",
		"TestUpdateQuerySetPreservesProviderScopeWhenOmitted",
		"INTEGRATION-005 + UI-SCREEN-MARKET-WATCH-009 + UI-SCREEN-MARKET-WATCH-010 preserves eBay output handoff response provenance",
		"TestEbayProviderSearchFailureTraceabilityImplemented",
		"| implemented |",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(row, fragment) {
			t.Fatalf("expected INTEGRATION-005 traceability row to include %q; row: %s", fragment, row)
		}
	}
}

func TestEbayProviderRunMessageDiagnosticsTraceabilityImplemented(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	var row string
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, "| `INTEGRATION-005` ") || strings.HasPrefix(line, "| INTEGRATION-005 ") {
			row = line
			break
		}
	}
	if row == "" {
		t.Fatalf("expected traceability row for INTEGRATION-005")
	}

	requiredFragments := []string{
		"Market Watch actionable provider-run `message` diagnostics",
		"INTEGRATION-005 + #827 surfaces eBay provider-run actionable message diagnostics",
		"TestEbayProviderRunMessageDiagnosticsTraceabilityImplemented",
		"| implemented |",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(row, fragment) {
			t.Fatalf("expected INTEGRATION-005 traceability row to include %q; row: %s", fragment, row)
		}
	}
}

func TestEbaySetupDocsTraceabilityImplemented(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	var row string
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, "| `INTEGRATION-005` ") || strings.HasPrefix(line, "| INTEGRATION-005 ") {
			row = line
			break
		}
	}
	if row == "" {
		t.Fatalf("expected traceability row for INTEGRATION-005")
	}

	requiredFragments := []string{
		"Help Center Integrations guide documents eBay bearer-token setup, marketplace/region, base URL override state, validation, Market Watch run path, auth/search diagnostics, and live-credential limitations",
		"TestIntegrationsGuideDocumentsEbaySetupWorkflow",
		"TestEbaySetupDocsTraceabilityImplemented",
		"| implemented |",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(row, fragment) {
			t.Fatalf("expected INTEGRATION-005 traceability row to include %q; row: %s", fragment, row)
		}
	}
}

func TestEbayDefaultSiteSearchHandoffTraceabilityImplemented(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	var row string
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, "| `DEFAULT-SITE-SEARCH-006` ") || strings.HasPrefix(line, "| DEFAULT-SITE-SEARCH-006 ") {
			row = line
			break
		}
	}
	if row == "" {
		t.Fatalf("expected traceability row for DEFAULT-SITE-SEARCH-006")
	}

	requiredFragments := []string{
		"eBay source/query provenance",
		"Discoveries, Wishlist, and Inventory handoff",
		"source provider, query-set id, query name, saved provider scope, and listing URL",
		"OpenAPI candidate/action provenance contracts",
		"TestOpenAPIDocumentsEbaySavedSearchHandoffContract",
		"TestApplySavedSearchActionsRetainAuditProvenance",
		"DEFAULT-SITE-SEARCH-006 hands off saved-search output to discoveries wishlist and inventory flows",
		"TestEbayDefaultSiteSearchHandoffTraceabilityImplemented",
		"| implemented |",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(row, fragment) {
			t.Fatalf("expected DEFAULT-SITE-SEARCH-006 traceability row to include %q; row: %s", fragment, row)
		}
	}
}

func TestEbaySavedQueryLifecycleTraceabilityImplemented(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	var row string
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, "| `INTEGRATION-005` ") || strings.HasPrefix(line, "| INTEGRATION-005 ") {
			row = line
			break
		}
	}
	if row == "" {
		t.Fatalf("expected traceability row for INTEGRATION-005")
	}

	requiredFragments := []string{
		"eBay saved-query create/edit/schedule/delete lifecycle",
		"provider_scope=[\"ebay\"]",
		"schedule_cron, enabled, rate_limit_rps, and latest-run hydration metadata",
		"omitted provider-scope updates preserving existing eBay scope",
		"TestOpenAPIDocumentsEbayQuerySetContract",
		"TestUpdateQuerySetPreservesProviderScopeWhenOmitted",
		"TestDefaultSiteSearchScheduledRefreshPersistsRunSnapshot",
		"INTEGRATION-005 + #827 manages eBay saved-query create edit schedule and delete lifecycle",
		"TestEbaySavedQueryLifecycleTraceabilityImplemented",
		"| implemented |",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(row, fragment) {
			t.Fatalf("expected INTEGRATION-005 traceability row to include %q; row: %s", fragment, row)
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

func TestEbayListingLifecycleTraceabilityImplemented(t *testing.T) {
	t.Parallel()

	traceabilityPath := filepath.Join("..", "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	var row string
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, "| `INTEGRATION-028` ") || strings.HasPrefix(line, "| INTEGRATION-028 ") {
			row = line
			break
		}
	}
	if row == "" {
		t.Fatalf("expected traceability row for INTEGRATION-028")
	}

	requiredFragments := []string{
		"seller listing lifecycle draft/publish/revise/end/relist safety-gated command contract",
		"API preview/execute wiring",
		"OpenAPI parity",
		"integration UI confirm-before-write workflow",
		"TestPreviewSellerListingLifecycleCommandsGateRemoteWrites",
		"TestExecuteSellerListingLifecycleCommandUsesMockedEbayResponses",
		"TestExecuteSellerListingLifecycleCommandBlocksUnconfirmedWrites",
		"TestEbayListingLifecyclePreviewExposesLocalDraftOnly",
		"TestEbayListingLifecycleExecuteAllowsLocalDraftAndBlocksRemoteAdapter",
		"TestOpenAPIDocumentsEbayListingLifecycleContract",
		"INTEGRATION-028: previews listing lifecycle commands and executes local drafts without remote write claims",
		"TestEbayListingLifecycleTraceabilityImplemented",
		"| implemented |",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(row, fragment) {
			t.Fatalf("expected INTEGRATION-028 traceability row to include %q; row: %s", fragment, row)
		}
	}
}

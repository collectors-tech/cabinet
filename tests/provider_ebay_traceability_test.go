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
			"fieldgroups=EXTENDED",
			"unsafe Browse base URL override fallback",
			"query-bearing or fragment-bearing overrides",
			"unsafe keyword/exclusion control-text omission",
			"unsafe keyword/exclusion Unicode format/control-text omission",
			"trimmed Browse string metadata",
			"malformed marketplace fallback",
			"malformed saved-query region filter omission",
			"local emitted-candidate page-size cap",
			"trimmed Browse price value",
			"unparseable Browse price skip",
			"non-finite saved-query max-price guard",
			"max-price Browse currency mismatch skip",
			"decimal-comma Browse price and shipping amount rejection",
			"non-plain-decimal Browse price skip",
			"over-precision Browse price skip",
			"non-positive Browse price skip",
			"non-finite Browse price skip",
			"blank Browse price currency skip",
			"malformed Browse price currency skip",
			"non-ASCII Browse price currency skip",
			"incomplete Browse item summary skip",
			"embedded-whitespace Browse listing id skip",
			"raw-control, encoded-control, or Unicode format/control Browse listing id/title skip",
			"invalid duplicate Browse listing id fallthrough",
			"case-insensitive comparison",
			"non-web, userinfo-bearing, raw-whitespace including Unicode whitespace, raw-control, or encoded-control Browse item URL skip",
			"non-web, userinfo-bearing, raw-whitespace including Unicode whitespace, raw-control, or encoded-control Browse image URL drop",
			"blank Browse seller fallback",
			"raw-control, encoded-control, or Unicode format/control Browse seller fallback",
			"embedded-whitespace Browse seller fallback",
			"first-parseable shipping-cost normalization",
			"non-plain-decimal shipping-cost skip",
			"over-precision shipping-cost skip",
			"blank-currency shipping-cost skip",
			"mismatched-currency shipping-cost skip",
			"non-positive shipping-cost skip",
			"non-finite shipping-cost skip",
			"max-price currency, item location, and condition Browse filters",
			"invalid JSON/profile/settings bootstrap client-error diagnostics",
			"missing query-set client-error diagnostics",
			"distinct `403` rejected-credential auth envelope documentation and runtime status preservation",
			"structured upstream error diagnostic whitespace compaction, length bounding, and control-byte omission",
			"upstream error body read cap plus bounded whitespace-compacted non-JSON/plain-text upstream body diagnostics",
			"malformed successful Browse response provider-error diagnostics",
			"trailing Browse response JSON/data rejection",
			"TestOpenAPIDocumentsEbayProviderRunContract",
			"TestOpenAPIDocumentsEbaySavedSearchHandoffContract",
			"TestEbayRegistryExposesSetupReadinessWithoutCredentialLeak",
			"TestNewProviderFallsBackFromUnsafeBaseURLOverrides",
			"TestProviderSearchCapsEmittedCandidatesToEffectiveBrowseLimit",
			"TestProviderSearchOmitsUnsafeQueryTextBeforeBrowseRequest",
			"TestProviderSearchRejectsOnlyUnsafeKeywords",
			"TestProviderSearchTrimsBrowseStringMetadata",
			"TestProviderSearchTrimsBrowsePriceValue",
			"TestProviderSearchSkipsUnparseableBrowsePrices",
			"TestProviderSearchOmitsNonFiniteMaxPriceBrowseFilter",
			"TestProviderSearchSkipsMaxPriceResultsWithMismatchedCurrency",
			"TestProviderSearchSkipsBrowsePricesAboveSavedQueryMax",
			"TestProviderSearchSkipsDecimalCommaBrowsePrices",
			"TestProviderSearchSkipsNonPlainDecimalBrowsePrices",
			"TestProviderSearchSkipsOverPrecisionBrowsePrices",
			"TestProviderSearchSkipsNonPositiveBrowsePrices",
			"TestProviderSearchSkipsNonFiniteBrowsePrices",
			"TestProviderSearchSkipsBlankBrowsePriceCurrency",
			"TestProviderSearchSkipsMalformedBrowsePriceCurrency",
			"TestProviderSearchSkipsIncompleteBrowseItemSummaries",
			"TestProviderSearchSkipsBrowseListingIDsWithEmbeddedWhitespace",
			"TestProviderSearchSkipsBrowseTextFieldsWithEncodedControlCharacters",
			"TestProviderSearchSkipsBrowseTextFieldsWithUnicodeFormatCharacters",
			"TestProviderSearchSkipsCaseVariantDuplicateBrowseListingIDs",
			"TestProviderSearchUsesFirstValidDuplicateBrowseListingID",
			"TestProviderSearchSkipsNonWebBrowseItemURLs",
			"TestProviderSearchSkipsBrowseItemURLsWithUserinfo",
			"TestProviderSearchSkipsBrowseItemURLsWithEncodedControlCharacters",
			"TestProviderSearchSkipsBrowseItemURLsWithRawControlCharacters",
			"TestProviderSearchSkipsBrowseItemURLsWithRawWhitespace",
			"TestProviderSearchSkipsBrowseItemURLsWithUnicodeWhitespace",
			"TestProviderSearchDropsNonWebBrowseImageURLs",
			"TestProviderSearchDropsBrowseImageURLsWithUserinfo",
			"TestProviderSearchDropsBrowseImageURLsWithEncodedControlCharacters",
			"TestProviderSearchDropsBrowseImageURLsWithRawControlCharacters",
			"TestProviderSearchDropsBrowseImageURLsWithRawWhitespace",
			"TestProviderSearchDropsBrowseImageURLsWithUnicodeWhitespace",
			"TestProviderSearchFallsBackBlankBrowseSeller",
			"TestProviderSearchFallsBackWhitespaceBrowseSeller",
			"TestProviderSearchUsesFirstParseableShippingCost",
			"TestProviderSearchSkipsDecimalCommaShippingCost",
			"TestProviderSearchSkipsNonPlainDecimalShippingCost",
			"TestProviderSearchSkipsOverPrecisionShippingCost",
			"TestProviderSearchSkipsBlankShippingCurrency",
			"TestProviderSearchSkipsMismatchedShippingCurrency",
			"TestProviderSearchSkipsNonPositiveShippingCost",
			"TestProviderSearchSkipsNonFiniteShippingCost",
			"TestProviderSearchClassifiesMalformedBrowsePayload",
			"TestProviderSearchRejectsBrowsePayloadWithTrailingData",
			"TestProviderSearchBoundsStructuredAuthErrorPayloadFields",
			"TestProviderSearchOmitsUnicodeFormatBrowseErrorPayloadFields",
			"TestProviderSearchBoundsStructuredBrowseErrorPayloadFields",
			"TestProviderSearchBuildsBrowseFiltersFromSavedQueryCriteria",
			"TestProviderSearchFallsBackFromMalformedMarketplace",
			"TestEbayProviderRunPreservesForbiddenAuthStatus",
			"eBay setup UI status panel renders registry `setup_status` readiness",
			"INTEGRATION-005 + #827 manages eBay saved-query create edit schedule and delete lifecycle",
			"omitted provider-scope updates preserving existing eBay scope",
			"TestUpdateQuerySetPreservesProviderScopeWhenOmitted",
			"TestEbayProviderTraceabilityImplemented",
			"| implemented |",
		},
		"INTEGRATION-006": {
			"/api/provider/health?provider=ebay",
			"latest positive eBay Browse `Retry-After` timing, including HTTP-date retry values",
			"provider health and scanner failure snapshots remain provider-scoped",
			"scanner failure list reusable schemas",
			"scanner failure list method-error diagnostics",
			"scanner failure retry request/accepted response fields",
			"TestProviderHealthEndpointKeepsUnrelatedProvidersIsolated",
			"TestRunNowRecordsProviderHealthForExecutingProvider",
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
			"fieldgroups=EXTENDED",
			"first meaningful availability entry",
			"unsafe availability status text",
			"Unicode control/format text",
			"negative eBay availability quantities",
			"unrecognized positive-quantity availability statuses",
			"TestProviderSearchNormalizesCandidates",
			"TestProviderSearchUsesFirstMeaningfulAvailability",
			"TestProviderSearchIgnoresUnsafeAvailabilityStatusText",
			"TestProviderSearchIgnoresUnicodeFormatAvailabilityStatusText",
			"TestProviderSearchIgnoresNegativeAvailabilityQuantity",
			"TestProviderSearchDoesNotInferUnknownAvailabilityAsInStock",
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
		"fieldgroups=EXTENDED",
		"unsafe Browse base URL override fallback",
		"query-bearing or fragment-bearing overrides",
		"blank keyword/exclusion trimming",
		"unsafe keyword/exclusion control-text omission",
		"malformed marketplace fallback",
		"malformed saved-query region filter omission",
		"structured auth and non-auth Browse error payload preservation",
		"structured upstream error diagnostic whitespace compaction, length bounding, and control-byte omission plus Unicode format/control omission",
		"upstream error body read cap plus bounded whitespace-compacted non-JSON/plain-text upstream body diagnostics",
		"malformed successful Browse response provider-error diagnostics",
		"trailing Browse response JSON/data rejection",
		"positive integer or HTTP-date `Retry-After` timing",
		"TestProviderSearchPreservesHTTPDateRetryAfter",
		"TestNewProviderFallsBackFromUnsafeBaseURLOverrides",
		"trimmed Browse string metadata",
		"trimmed Browse price value",
		"non-finite saved-query max-price guard",
		"max-price Browse currency mismatch skip",
		"non-positive Browse price skip",
		"non-plain-decimal Browse price skip",
		"over-precision Browse price skip",
		"blank Browse price currency skip",
		"malformed Browse price currency skip",
		"non-ASCII Browse price currency skip",
		"embedded-whitespace Browse listing id skip",
		"raw-control, encoded-control, or Unicode format/control Browse listing id/title skip",
		"invalid duplicate Browse listing id fallthrough",
		"case-insensitive comparison",
		"non-web, userinfo-bearing, raw-whitespace including Unicode whitespace, raw-control, or encoded-control Browse item URL skip",
		"non-web, userinfo-bearing, raw-whitespace including Unicode whitespace, raw-control, or encoded-control Browse image URL drop",
		"blank Browse seller fallback",
		"raw-control, encoded-control, or Unicode format/control Browse seller fallback",
		"embedded-whitespace Browse seller fallback",
		"first-parseable shipping-cost normalization",
		"non-plain-decimal shipping-cost skip",
		"over-precision shipping-cost skip",
		"blank-currency shipping-cost skip",
		"mismatched-currency shipping-cost skip",
		"non-positive shipping-cost skip",
		"max-price currency, item location, and condition Browse filters",
		"`/api/scanner/run` provider-error envelope",
		"`/api/providers/ebay/run` saved-search candidate persistence",
		"`/api/providers/registry` eBay setup readiness contract",
		"setup_status.base_url_set",
		"degraded provider-health next action",
		"eBay setup UI status panel renders registry `setup_status` readiness",
		"Wishlist/Inventory provenance handoff",
		"TestProviderSearchTrimsBlankCriteriaBeforeBrowseRequest",
		"TestProviderSearchOmitsUnsafeQueryTextBeforeBrowseRequest",
		"TestProviderSearchRejectsOnlyUnsafeKeywords",
		"TestProviderSearchFallsBackFromMalformedMarketplace",
		"TestProviderSearchUsesFirstParseableShippingCost",
		"TestProviderSearchSkipsNonPlainDecimalShippingCost",
		"TestProviderSearchSkipsOverPrecisionShippingCost",
		"TestProviderSearchSkipsBlankShippingCurrency",
		"TestProviderSearchSkipsNonPositiveShippingCost",
		"TestProviderSearchBuildsBrowseFiltersFromSavedQueryCriteria",
		"TestWave4EbayRunMissingQuerySetReturnsActionableClientEnvelope",
		"TestProviderSearchPreservesStructuredBrowseErrorPayload",
		"TestProviderSearchSanitizesStructuredBrowseErrorPayload",
		"TestProviderSearchOmitsUnicodeFormatBrowseErrorPayloadFields",
		"TestProviderSearchBoundsStructuredBrowseErrorPayloadFields",
		"TestProviderSearchPreservesPlainTextBrowseErrorPayload",
		"TestEbayErrorMessageCapsBodyReadBeforeDiagnostics",
		"TestProviderSearchClassifiesMalformedBrowsePayload",
		"TestProviderSearchRejectsBrowsePayloadWithTrailingData",
		"TestScannerRunMapsEbayBrowseRetryAfterToProviderErrorEnvelope",
		"TestEbayProviderRunMapsBrowseFailureToProviderHealthGuidance",
		"TestOpenAPIDocumentsEbayScannerRunSearchErrorEnvelope",
		"TestProviderSearchTrimsBrowseStringMetadata",
		"TestProviderSearchTrimsBrowsePriceValue",
		"TestProviderSearchSkipsBrowsePricesAboveSavedQueryMax",
		"TestProviderSearchOmitsNonFiniteMaxPriceBrowseFilter",
		"TestProviderSearchSkipsMaxPriceResultsWithMismatchedCurrency",
		"TestProviderSearchSkipsNonPlainDecimalBrowsePrices",
		"TestProviderSearchSkipsOverPrecisionBrowsePrices",
		"TestProviderSearchSkipsNonPositiveBrowsePrices",
		"TestProviderSearchSkipsBlankBrowsePriceCurrency",
		"TestProviderSearchSkipsMalformedBrowsePriceCurrency",
		"TestProviderSearchSkipsBrowseTextFieldsWithRawControlCharacters",
		"TestProviderSearchSkipsBrowseListingIDsWithEmbeddedWhitespace",
		"TestProviderSearchSkipsBrowseTextFieldsWithEncodedControlCharacters",
		"TestProviderSearchSkipsBrowseTextFieldsWithUnicodeFormatCharacters",
		"TestProviderSearchSkipsCaseVariantDuplicateBrowseListingIDs",
		"TestProviderSearchUsesFirstValidDuplicateBrowseListingID",
		"TestProviderSearchSkipsNonWebBrowseItemURLs",
		"TestProviderSearchSkipsBrowseItemURLsWithUserinfo",
		"TestProviderSearchSkipsBrowseItemURLsWithEncodedControlCharacters",
		"TestProviderSearchSkipsBrowseItemURLsWithRawControlCharacters",
		"TestProviderSearchSkipsBrowseItemURLsWithRawWhitespace",
		"TestProviderSearchSkipsBrowseItemURLsWithUnicodeWhitespace",
		"TestProviderSearchDropsNonWebBrowseImageURLs",
		"TestProviderSearchDropsBrowseImageURLsWithUserinfo",
		"TestProviderSearchDropsBrowseImageURLsWithEncodedControlCharacters",
		"TestProviderSearchDropsBrowseImageURLsWithRawControlCharacters",
		"TestProviderSearchDropsBrowseImageURLsWithRawWhitespace",
		"TestProviderSearchDropsBrowseImageURLsWithUnicodeWhitespace",
		"TestProviderSearchFallsBackBlankBrowseSeller",
		"TestProviderSearchFallsBackRawControlBrowseSeller",
		"TestProviderSearchFallsBackWhitespaceBrowseSeller",
		"TestProviderSearchPreservesPlainTextAuthErrorPayload",
		"TestProviderSearchBoundsStructuredAuthErrorPayloadFields",
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
		"setup page-size validation diagnostics that route to setup correction instead of credential review",
		"invalid query-set diagnostics that route to saved-query selection instead of credential review",
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

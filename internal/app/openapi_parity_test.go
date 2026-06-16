package app

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestOpenAPIDocumentsRuntimeEndpoints(t *testing.T) {
	t.Parallel()

	specPath, raw := readOpenAPISpec(t)

	requiredPaths := []string{
		"/api/onboarding/sample-data",
		"/api/scanner/run/scheduled",
		"/api/scanner/query-sets/{querySetID}",
		"/api/scanner/failures",
		"/api/scanner/failures/retry",
		"/api/provider/health",
		"/api/settings/reset-ignore-rules",
		"/api/wishlist/convert-owned",
		"/api/wishlist/hits",
		"/api/pricing/snapshot/run",
		"/api/pricing/graph",
		"/api/pricing/by-source",
		"/api/pricing/stats",
		"/api/pricing/trend",
		"/api/pricing/history/export",
		"/api/logs/debug",
		"/api/ai/toggle",
		"/api/ai/suggest/photo",
		"/api/telegram/catalog-captures",
		"/api/telegram/webhook/catalog-captures",
		"/api/integrations/ebay/purchase-inbox/reviews",
		"/api/integrations/ebay/purchase-inbox/actions",
		"/api/forwarding/packages",
		"/api/forwarding/packages/import-csv",
		"/api/forwarding/packages/import-email",
		"/api/forwarding/package-links",
		"/api/forwarding/package-match-suggestions",
		"/api/providers/ebay/seller-operations/preview",
		"/api/providers/ebay/seller-operations/execute",
		"/api/providers/ebay/listing-lifecycle/preview",
		"/api/providers/ebay/listing-lifecycle/execute",
		"/api/providers/ebay/run",
		"/api/chat/threads",
		"/api/chat/messages",
		"/api/chat/attachments",
		"/api/chat/actions/preview",
		"/api/chat/actions/apply",
		"/api/chat/actions/cancel",
		"/api/data/export/json",
		"/api/data/export/csv/items",
		"/api/data/import/json/dry-run",
		"/api/data/import/json/apply",
		"/api/data/import/csv/dry-run",
		"/api/data/import/csv/apply",
		"/api/data/reindex",
		"/api/data/rebuild-thumbnails",
		"/api/data/repair",
		"/api/backup/run",
		"/api/backup/list",
		"/api/backup/restore",
		"/api/auth/webauthn/register/begin",
		"/api/auth/webauthn/register/finish",
		"/api/auth/requirements",
		"/api/auth/recovery/passphrase",
		"/api/auth/recovery/reset/begin",
		"/api/auth/webauthn/login/begin",
		"/api/auth/webauthn/login/finish",
		"/api/auth/session/validate",
		"/api/auth/session/lock",
		"/api/profiles/{profileID}/saved-filters",
		"/api/profiles/{profileID}/storage",
		"/api/profiles/{profileID}/license",
		"/api/items/bulk-edit",
		"/api/items/{itemID}",
		"/api/items/{itemID}/instances/{instanceID}",
		"/api/items/{itemID}/barcodes",
		"/api/items/{itemID}/photos",
		"/api/items/{itemID}/photos/reorder",
		"/api/items/{itemID}/photos/{photoID}",
		"/api/items/{itemID}/photos/{photoID}/primary",
		"/api/items/{itemID}/photos/{photoID}/rotate",
		"/api/items/{itemID}/photos/{photoID}/file",
		"/api/items/{itemID}/photos-rebuild",
		"/api/barcodes/{barcode}",
		"/api/barcodes/{barcode}/external-search",
	}

	for _, endpoint := range requiredPaths {
		pathPattern := regexp.MustCompile(fmt.Sprintf(`(?m)^  %s:\r?$`, regexp.QuoteMeta(endpoint)))
		if !pathPattern.Match(raw) {
			t.Fatalf("openapi missing %s path in %s", endpoint, specPath)
		}
	}
}

func TestOpenAPIDocumentsEbayListingLifecycleContract(t *testing.T) {
	t.Parallel()

	specPath, raw := readOpenAPISpec(t)
	for path, operationID := range map[string]string{
		"/api/providers/ebay/listing-lifecycle/preview": "previewEbayListingLifecycle",
		"/api/providers/ebay/listing-lifecycle/execute": "executeEbayListingLifecycle",
	} {
		section, ok := openAPIPathSection(raw, path)
		if !ok {
			t.Fatalf("openapi missing %s path in %s", path, specPath)
		}
		for _, token := range []string{
			"operationId: " + operationID,
			"$ref: \"#/components/schemas/EbayListingLifecycleRequest\"",
		} {
			if !strings.Contains(section, token) {
				t.Fatalf("openapi %s section missing %q:\n%s", path, token, section)
			}
		}
	}

	requestSchema, ok := openAPIComponentSection(raw, "EbayListingLifecycleRequest")
	if !ok {
		t.Fatalf("openapi missing EbayListingLifecycleRequest schema in %s", specPath)
	}
	for _, token := range []string{
		"command:",
		"enum: [draft, publish, revise, end, relist]",
		"capability:",
		"enum: [unsupported, draft_only, confirmed_api]",
		"confirmed: { type: boolean }",
		"item_id: { type: string }",
		"draft_id: { type: string }",
		"listing_id: { type: string }",
		"title: { type: string }",
	} {
		if !strings.Contains(requestSchema, token) {
			t.Fatalf("openapi EbayListingLifecycleRequest schema missing %q:\n%s", token, requestSchema)
		}
	}

	previewSchema, ok := openAPIComponentSection(raw, "EbayListingLifecyclePreview")
	if !ok {
		t.Fatalf("openapi missing EbayListingLifecyclePreview schema in %s", specPath)
	}
	for _, token := range []string{
		"allowed: { type: boolean }",
		"local_only: { type: boolean }",
		"remote_write: { type: boolean }",
		"confirmation_required: { type: boolean }",
		"blocker: { type: string }",
	} {
		if !strings.Contains(previewSchema, token) {
			t.Fatalf("openapi EbayListingLifecyclePreview schema missing %q:\n%s", token, previewSchema)
		}
	}

	executionSchema, ok := openAPIComponentSection(raw, "EbayListingLifecycleExecution")
	if !ok {
		t.Fatalf("openapi missing EbayListingLifecycleExecution schema in %s", specPath)
	}
	for _, token := range []string{
		"executed: { type: boolean }",
		"status: { type: string }",
		"response:",
		"$ref: \"#/components/schemas/EbayListingLifecycleResponse\"",
	} {
		if !strings.Contains(executionSchema, token) {
			t.Fatalf("openapi EbayListingLifecycleExecution schema missing %q:\n%s", token, executionSchema)
		}
	}

	responseSchema, ok := openAPIComponentSection(raw, "EbayListingLifecycleResponse")
	if !ok {
		t.Fatalf("openapi missing EbayListingLifecycleResponse schema in %s", specPath)
	}
	for _, token := range []string{
		"provider: { type: string }",
		"command: { type: string }",
		"draft_id: { type: string }",
		"listing_id: { type: string }",
		"status: { type: string }",
	} {
		if !strings.Contains(responseSchema, token) {
			t.Fatalf("openapi EbayListingLifecycleResponse schema missing %q:\n%s", token, responseSchema)
		}
	}
}

func TestOpenAPIDocumentsScannerFailureRetryGuidance(t *testing.T) {
	t.Parallel()

	specPath, raw := readOpenAPISpec(t)
	section, ok := openAPIPathSection(raw, "/api/scanner/failures")
	if !ok {
		t.Fatalf("openapi missing /api/scanner/failures path in %s", specPath)
	}

	requiredTokens := []string{
		"failures:",
		"retry_guidance:",
		"next_action:",
		"check_provider_health_and_credentials",
	}
	for _, token := range requiredTokens {
		if !strings.Contains(section, token) {
			t.Fatalf("openapi scanner failures contract missing %q:\n%s", token, section)
		}
	}
}

func TestOpenAPIDocumentsScannerCandidateObservedCurrency(t *testing.T) {
	t.Parallel()

	specPath, raw := readOpenAPISpec(t)
	candidateSchema, ok := openAPIComponentSection(raw, "Candidate")
	if !ok {
		t.Fatalf("openapi missing Candidate schema in %s", specPath)
	}
	for _, token := range []string{
		"observed_currency: { type: string",
		"Normalized listing price currency",
	} {
		if !strings.Contains(candidateSchema, token) {
			t.Fatalf("openapi Candidate schema missing %q:\n%s", token, candidateSchema)
		}
	}
}

func TestOpenAPIOperationsDeclareClientErrorResponses(t *testing.T) {
	t.Parallel()

	specPath, raw := readOpenAPISpec(t)
	lines := regexp.MustCompile(`\r?\n`).Split(string(raw), -1)
	operationPattern := regexp.MustCompile(`^    (get|post|put|patch|delete):\s*$`)
	responsePattern := regexp.MustCompile(`^        ["']?([0-9]{3}|default)["']?:\s*$`)

	var missing []string
	for idx := 0; idx < len(lines); idx++ {
		if !operationPattern.MatchString(lines[idx]) {
			continue
		}

		operationLine := idx + 1
		operation := lines[idx]
		has4XX := false
		for scan := idx + 1; scan < len(lines); scan++ {
			line := lines[scan]
			if line != "" && len(line)-len(strings.TrimLeft(line, " ")) <= 4 {
				break
			}
			match := responsePattern.FindStringSubmatch(line)
			if len(match) == 2 && strings.HasPrefix(match[1], "4") {
				has4XX = true
				break
			}
		}
		if !has4XX {
			missing = append(missing, fmt.Sprintf("line %d %s", operationLine, strings.TrimSpace(operation)))
		}
	}

	if len(missing) > 0 {
		t.Fatalf("openapi operations missing 4XX responses in %s: %v", specPath, missing)
	}
}

func TestOpenAPIDocumentsForwarderMatchSuggestionQueryContract(t *testing.T) {
	t.Parallel()

	specPath, raw := readOpenAPISpec(t)
	section, ok := openAPIPathSection(raw, "/api/forwarding/package-match-suggestions")
	if !ok {
		t.Fatalf("openapi missing /api/forwarding/package-match-suggestions path in %s", specPath)
	}

	requiredTokens := []string{
		"operationId: listForwarderPackageMatchSuggestions",
		"name: package_id",
		"name: confidence_label",
		"enum: [high, medium, low]",
		"confidence_filter:",
		"suggestions:",
		"summary:",
		"count: { type: integer }",
		"high_confidence: { type: integer }",
		"medium_confidence: { type: integer }",
		"low_confidence: { type: integer }",
		"scoped_packages: { type: integer }",
		"$ref: \"#/components/schemas/ForwarderPackageMatchSuggestion\"",
	}
	for _, token := range requiredTokens {
		if !strings.Contains(section, token) {
			t.Fatalf("openapi %s section missing %q:\n%s", "/api/forwarding/package-match-suggestions", token, section)
		}
	}

	suggestionSchema, ok := openAPIComponentSection(raw, "ForwarderPackageMatchSuggestion")
	if !ok {
		t.Fatalf("openapi missing ForwarderPackageMatchSuggestion schema in %s", specPath)
	}
	for _, token := range []string{
		"package_id: { type: string }",
		"confidence_score: { type: number }",
		"confidence_label:",
		"explanation:",
		"signals:",
		"audit_trail:",
	} {
		if !strings.Contains(suggestionSchema, token) {
			t.Fatalf("openapi ForwarderPackageMatchSuggestion schema missing %q:\n%s", token, suggestionSchema)
		}
	}
}

func TestOpenAPIDocumentsEbaySellerOperationsContract(t *testing.T) {
	t.Parallel()

	specPath, raw := readOpenAPISpec(t)
	for path, operationID := range map[string]string{
		"/api/providers/ebay/seller-operations/preview": "previewEbaySellerOperation",
		"/api/providers/ebay/seller-operations/execute": "executeEbaySellerOperation",
	} {
		section, ok := openAPIPathSection(raw, path)
		if !ok {
			t.Fatalf("openapi missing %s path in %s", path, specPath)
		}
		for _, token := range []string{
			"operationId: " + operationID,
			"$ref: \"#/components/schemas/EbaySellerOperationRequest\"",
		} {
			if !strings.Contains(section, token) {
				t.Fatalf("openapi %s section missing %q:\n%s", path, token, section)
			}
		}
	}

	requestSchema, ok := openAPIComponentSection(raw, "EbaySellerOperationRequest")
	if !ok {
		t.Fatalf("openapi missing EbaySellerOperationRequest schema in %s", specPath)
	}
	for _, token := range []string{
		"operation:",
		"enum: [messages, notifications, sold_orders, fulfilment, fulfillment, offers, orders]",
		"action: { type: string }",
		"capability:",
		"enum: [unverified, read_only, confirmed_api]",
		"reference_id: { type: string }",
		"confirmed: { type: boolean }",
	} {
		if !strings.Contains(requestSchema, token) {
			t.Fatalf("openapi EbaySellerOperationRequest schema missing %q:\n%s", token, requestSchema)
		}
	}

	previewSchema, ok := openAPIComponentSection(raw, "EbaySellerOperationPreview")
	if !ok {
		t.Fatalf("openapi missing EbaySellerOperationPreview schema in %s", specPath)
	}
	for _, token := range []string{
		"read_available: { type: boolean }",
		"write_available: { type: boolean }",
		"allowed: { type: boolean }",
		"remote_write: { type: boolean }",
		"confirmation_required: { type: boolean }",
		"blocker: { type: string }",
	} {
		if !strings.Contains(previewSchema, token) {
			t.Fatalf("openapi EbaySellerOperationPreview schema missing %q:\n%s", token, previewSchema)
		}
	}

	executionSchema, ok := openAPIComponentSection(raw, "EbaySellerOperationExecution")
	if !ok {
		t.Fatalf("openapi missing EbaySellerOperationExecution schema in %s", specPath)
	}
	for _, token := range []string{
		"executed: { type: boolean }",
		"local_only: { type: boolean }",
		"result:",
		"$ref: \"#/components/schemas/EbaySellerOperationReadResult\"",
	} {
		if !strings.Contains(executionSchema, token) {
			t.Fatalf("openapi EbaySellerOperationExecution schema missing %q:\n%s", token, executionSchema)
		}
	}

	resultSchema, ok := openAPIComponentSection(raw, "EbaySellerOperationReadResult")
	if !ok {
		t.Fatalf("openapi missing EbaySellerOperationReadResult schema in %s", specPath)
	}
	for _, token := range []string{
		"source: { type: string }",
		"records:",
		"kind: { type: string }",
		"status: { type: string }",
		"summary:",
	} {
		if !strings.Contains(resultSchema, token) {
			t.Fatalf("openapi EbaySellerOperationReadResult schema missing %q:\n%s", token, resultSchema)
		}
	}
}

func TestOpenAPIDocumentsEbayScannerRunAuthErrorEnvelope(t *testing.T) {
	t.Parallel()

	specPath, raw := readOpenAPISpec(t)
	section, ok := openAPIPathSection(raw, "/api/scanner/run")
	if !ok {
		t.Fatalf("openapi missing /api/scanner/run path in %s", specPath)
	}
	for _, token := range []string{
		"Runs the saved search and maps eBay provider auth and search failures to actionable provider error envelopes.",
		`"401":`,
		"$ref: \"#/components/schemas/ProviderAuthErrorResponse\"",
	} {
		if !strings.Contains(section, token) {
			t.Fatalf("openapi /api/scanner/run section missing %q:\n%s", token, section)
		}
	}

	authErrorSchema, ok := openAPIComponentSection(raw, "ProviderAuthErrorResponse")
	if !ok {
		t.Fatalf("openapi missing ProviderAuthErrorResponse schema in %s", specPath)
	}
	for _, token := range []string{
		"error: { type: string, enum: [failed_to_run_scanner] }",
		"error_code:",
		"enum: [PROVIDER_AUTH_MISSING, PROVIDER_AUTH_INVALID]",
		"provider: { type: string, enum: [ebay] }",
		"query_set_id: { type: string }",
		"next_action: { type: string, enum: [review_provider_credentials_and_health] }",
	} {
		if !strings.Contains(authErrorSchema, token) {
			t.Fatalf("openapi ProviderAuthErrorResponse schema missing %q:\n%s", token, authErrorSchema)
		}
	}
}

func TestOpenAPIDocumentsEbayScannerRunSearchErrorEnvelope(t *testing.T) {
	t.Parallel()

	specPath, raw := readOpenAPISpec(t)
	section, ok := openAPIPathSection(raw, "/api/scanner/run")
	if !ok {
		t.Fatalf("openapi missing /api/scanner/run path in %s", specPath)
	}
	for _, token := range []string{
		"maps eBay provider auth and search failures to actionable provider error envelopes",
		`"429":`,
		"$ref: \"#/components/schemas/ProviderSearchErrorResponse\"",
	} {
		if !strings.Contains(section, token) {
			t.Fatalf("openapi /api/scanner/run section missing %q:\n%s", token, section)
		}
	}

	searchErrorSchema, ok := openAPIComponentSection(raw, "ProviderSearchErrorResponse")
	if !ok {
		t.Fatalf("openapi missing ProviderSearchErrorResponse schema in %s", specPath)
	}
	for _, token := range []string{
		"error: { type: string, enum: [failed_to_run_scanner] }",
		"error_code: { type: string, enum: [PROVIDER_SEARCH_FAILED] }",
		"provider: { type: string, enum: [ebay] }",
		"message: { type: string, description: Structured upstream Browse failure details preserved for diagnostics. }",
		"next_action: { type: string, enum: [check_provider_health_and_credentials] }",
		"retry_after_seconds:",
		"Upstream eBay Retry-After seconds when Browse rate limiting provides retry timing.",
		"query_set_id: { type: string }",
	} {
		if !strings.Contains(searchErrorSchema, token) {
			t.Fatalf("openapi ProviderSearchErrorResponse schema missing %q:\n%s", token, searchErrorSchema)
		}
	}
}

func TestOpenAPIDocumentsEbayProviderHealthContract(t *testing.T) {
	t.Parallel()

	specPath, raw := readOpenAPISpec(t)
	section, ok := openAPIPathSection(raw, "/api/provider/health")
	if !ok {
		t.Fatalf("openapi missing /api/provider/health path in %s", specPath)
	}
	for _, token := range []string{
		"summary: Provider health status",
		"ProviderHealthResponse",
		"name: provider",
		"description: Provider id such as ebay.",
	} {
		if !strings.Contains(section, token) {
			t.Fatalf("openapi /api/provider/health section missing %q:\n%s", token, section)
		}
	}

	schema, ok := openAPIComponentSection(raw, "ProviderHealthResponse")
	if !ok {
		t.Fatalf("openapi missing ProviderHealthResponse schema in %s", specPath)
	}
	for _, token := range []string{
		"required: [provider, status, state]",
		"provider: { type: string, example: ebay }",
		"status:",
		"enum: [ok, error, unknown]",
		"state:",
		"enum: [ready, degraded, disabled]",
		"message: { type: string }",
		"last_error:",
		"retry_after_seconds:",
		"updated_at:",
	} {
		if !strings.Contains(schema, token) {
			t.Fatalf("openapi ProviderHealthResponse schema missing %q:\n%s", token, schema)
		}
	}
}

func TestOpenAPIDocumentsEbayProviderRunContract(t *testing.T) {
	t.Parallel()

	specPath, raw := readOpenAPISpec(t)
	section, ok := openAPIPathSection(raw, "/api/providers/ebay/run")
	if !ok {
		t.Fatalf("openapi missing /api/providers/ebay/run path in %s", specPath)
	}
	for _, token := range []string{
		"summary: Run eBay saved-search provider",
		"Runs an eBay-scoped saved search with active-profile credentials, persists normalized candidates, and returns the hydrated run snapshot.",
		"required: [query_set_id]",
		"$ref: \"#/components/schemas/EbayProviderRunResponse\"",
		`"401":`,
		"$ref: \"#/components/schemas/EbayProviderRunAuthErrorResponse\"",
		`"429":`,
		"$ref: \"#/components/schemas/EbayProviderRunSearchErrorResponse\"",
	} {
		if !strings.Contains(section, token) {
			t.Fatalf("openapi /api/providers/ebay/run section missing %q:\n%s", token, section)
		}
	}

	runSchema, ok := openAPIComponentSection(raw, "EbayProviderRunResponse")
	if !ok {
		t.Fatalf("openapi missing EbayProviderRunResponse schema in %s", specPath)
	}
	for _, token := range []string{
		"required: [query_set_id, provider, candidates, run]",
		"query_set_id: { type: string }",
		"provider: { type: string, enum: [ebay] }",
		"$ref: \"#/components/schemas/Candidate\"",
		"saved: { type: integer }",
		"attempts: { type: integer }",
	} {
		if !strings.Contains(runSchema, token) {
			t.Fatalf("openapi EbayProviderRunResponse schema missing %q:\n%s", token, runSchema)
		}
	}

	authErrorSchema, ok := openAPIComponentSection(raw, "EbayProviderRunAuthErrorResponse")
	if !ok {
		t.Fatalf("openapi missing EbayProviderRunAuthErrorResponse schema in %s", specPath)
	}
	for _, token := range []string{
		"error: { type: string, enum: [failed_to_run_ebay_provider] }",
		"enum: [PROVIDER_AUTH_MISSING, PROVIDER_AUTH_INVALID]",
		"provider: { type: string, enum: [ebay] }",
		"next_action: { type: string, enum: [review_provider_credentials_and_health] }",
		"query_set_id: { type: string }",
	} {
		if !strings.Contains(authErrorSchema, token) {
			t.Fatalf("openapi EbayProviderRunAuthErrorResponse schema missing %q:\n%s", token, authErrorSchema)
		}
	}

	searchErrorSchema, ok := openAPIComponentSection(raw, "EbayProviderRunSearchErrorResponse")
	if !ok {
		t.Fatalf("openapi missing EbayProviderRunSearchErrorResponse schema in %s", specPath)
	}
	for _, token := range []string{
		"error: { type: string, enum: [failed_to_run_ebay_provider] }",
		"error_code: { type: string, enum: [PROVIDER_SEARCH_FAILED] }",
		"provider: { type: string, enum: [ebay] }",
		"message: { type: string, description: Structured upstream Browse failure details preserved for diagnostics. }",
		"next_action: { type: string, enum: [check_provider_health_and_credentials] }",
		"retry_after_seconds:",
		"Upstream eBay Retry-After seconds when Browse rate limiting provides retry timing.",
		"query_set_id: { type: string }",
	} {
		if !strings.Contains(searchErrorSchema, token) {
			t.Fatalf("openapi EbayProviderRunSearchErrorResponse schema missing %q:\n%s", token, searchErrorSchema)
		}
	}
}

func TestOpenAPIDocumentsEbayQuerySetContract(t *testing.T) {
	t.Parallel()

	specPath, raw := readOpenAPISpec(t)
	listSection, ok := openAPIPathSection(raw, "/api/scanner/query-sets")
	if !ok {
		t.Fatalf("openapi missing /api/scanner/query-sets path in %s", specPath)
	}
	for _, token := range []string{
		"Lists active-profile query sets including provider scope, schedule controls, rate-limit settings, and hydrated latest-run metadata used by Market Watch.",
		"Creates a provider-scoped saved search. eBay Market Watch query sets use `provider_scope=[\"ebay\"]` with optional pagination, schedule, and rate-limit controls.",
		"$ref: \"#/components/schemas/ScannerQuerySetInput\"",
		"$ref: \"#/components/schemas/ScannerQuerySet\"",
	} {
		if !strings.Contains(listSection, token) {
			t.Fatalf("openapi /api/scanner/query-sets section missing %q:\n%s", token, listSection)
		}
	}

	updateSection, ok := openAPIPathSection(raw, "/api/scanner/query-sets/{querySetID}")
	if !ok {
		t.Fatalf("openapi missing /api/scanner/query-sets/{querySetID} path in %s", specPath)
	}
	if !strings.Contains(updateSection, "Updates saved-search filters, provider scope, pagination, schedule, enabled state, and rate-limit controls for an active-profile query set.") {
		t.Fatalf("openapi /api/scanner/query-sets/{querySetID} section missing update description:\n%s", updateSection)
	}

	inputSchema, ok := openAPIComponentSection(raw, "ScannerQuerySetInput")
	if !ok {
		t.Fatalf("openapi missing ScannerQuerySetInput schema in %s", specPath)
	}
	for _, token := range []string{
		"provider_scope:",
		"Provider ids to run for this saved search. eBay-only Market Watch query sets use `ebay`.",
		"enum: [ebay, amazon, bonzaslotcars, frontlinehobbies, hobbytechtoys, mrtoys.com.au]",
		"items_per_page:",
		"Requested page size for provider-backed runs; runtime applies provider-safe caps in run summaries.",
		"schedule_cron: { type: string }",
		"enabled: { type: boolean }",
		"rate_limit_rps: { type: integer }",
	} {
		if !strings.Contains(inputSchema, token) {
			t.Fatalf("openapi ScannerQuerySetInput schema missing %q:\n%s", token, inputSchema)
		}
	}

	querySetSchema, ok := openAPIComponentSection(raw, "ScannerQuerySet")
	if !ok {
		t.Fatalf("openapi missing ScannerQuerySet schema in %s", specPath)
	}
	for _, token := range []string{
		"last_run_status:",
		"enum: [never, succeeded, failed]",
		"last_run_at: { type: string }",
		"last_run_message: { type: string }",
		"last_candidate_count: { type: integer }",
	} {
		if !strings.Contains(querySetSchema, token) {
			t.Fatalf("openapi ScannerQuerySet schema missing %q:\n%s", token, querySetSchema)
		}
	}
}

func TestOpenAPIDocumentsEbaySavedSearchHandoffContract(t *testing.T) {
	t.Parallel()

	specPath, raw := readOpenAPISpec(t)
	candidatesSection, ok := openAPIPathSection(raw, "/api/scanner/candidates")
	if !ok {
		t.Fatalf("openapi missing /api/scanner/candidates path in %s", specPath)
	}
	for _, token := range []string{
		"Lists saved-search candidates for output inspection and handoff, including provider/query provenance used by eBay Market Watch handoff actions.",
		"name: query_set_id",
		"$ref: \"#/components/schemas/Candidate\"",
	} {
		if !strings.Contains(candidatesSection, token) {
			t.Fatalf("openapi /api/scanner/candidates section missing %q:\n%s", token, candidatesSection)
		}
	}

	actionSection, ok := openAPIPathSection(raw, "/api/discovery/action")
	if !ok {
		t.Fatalf("openapi missing /api/discovery/action path in %s", specPath)
	}
	for _, token := range []string{
		"Applies saved-search output handoff actions for Discoveries, Wishlist, or Inventory while preserving eBay provider/query provenance in the durable discovery action audit.",
		"enum: [ignore, add_to_wishlist, track_price, create_item]",
		"source: { type: string, enum: [market_watch] }",
		"source_provider: { type: string, enum: [ebay] }",
		"provider_scope:",
		"$ref: \"#/components/schemas/DiscoveryActionResponse\"",
	} {
		if !strings.Contains(actionSection, token) {
			t.Fatalf("openapi /api/discovery/action section missing %q:\n%s", token, actionSection)
		}
	}

	actionResponseSchema, ok := openAPIComponentSection(raw, "DiscoveryActionResponse")
	if !ok {
		t.Fatalf("openapi missing DiscoveryActionResponse schema in %s", specPath)
	}
	for _, token := range []string{
		"required: [ok, action, candidate_id, audit]",
		"source_provider: { type: string, enum: [ebay] }",
		"query_set_id: { type: string }",
		"query_name: { type: string }",
		"provider_scope:",
		"source_result_url: { type: string }",
		"observed_currency: { type: string }",
	} {
		if !strings.Contains(actionResponseSchema, token) {
			t.Fatalf("openapi DiscoveryActionResponse schema missing %q:\n%s", token, actionResponseSchema)
		}
	}

	candidateSchema, ok := openAPIComponentSection(raw, "Candidate")
	if !ok {
		t.Fatalf("openapi missing Candidate schema in %s", specPath)
	}
	for _, token := range []string{
		"query_set_id: { type: string }",
		"listing_id: { type: string }",
		"url: { type: string }",
		"seller: { type: string }",
		"source: { type: string, description: Source provider id such as ebay. }",
	} {
		if !strings.Contains(candidateSchema, token) {
			t.Fatalf("openapi Candidate schema missing %q:\n%s", token, candidateSchema)
		}
	}
}

func TestOpenAPIDocumentsForwarderPackageLinkDecisionContract(t *testing.T) {
	t.Parallel()

	specPath, raw := readOpenAPISpec(t)
	section, ok := openAPIPathSection(raw, "/api/forwarding/package-links")
	if !ok {
		t.Fatalf("openapi missing /api/forwarding/package-links path in %s", specPath)
	}

	for _, token := range []string{
		"operationId: listForwarderPackageLinks",
		"name: package_id",
		"links:",
		"events:",
		"summary:",
		"count: { type: integer }",
		"events: { type: integer }",
		"$ref: \"#/components/schemas/ForwarderPackageLink\"",
		"$ref: \"#/components/schemas/ForwarderPackageLinkEvent\"",
		"operationId: linkForwarderPackage",
		"$ref: \"#/components/schemas/ForwarderPackageLinkRequest\"",
		"operationId: unlinkForwarderPackage",
		"$ref: \"#/components/schemas/ForwarderPackageUnlinkRequest\"",
		"enum: [forwarder_package_reconciliation_unlink]",
	} {
		if !strings.Contains(section, token) {
			t.Fatalf("openapi %s section missing %q:\n%s", "/api/forwarding/package-links", token, section)
		}
	}

	requestSchema, ok := openAPIComponentSection(raw, "ForwarderPackageLinkRequest")
	if !ok {
		t.Fatalf("openapi missing ForwarderPackageLinkRequest schema in %s", specPath)
	}
	for _, token := range []string{
		"package_id: { type: string }",
		"item_id: { type: string }",
		"lifecycle_entry_id: { type: string }",
		"expected_arrival_id: { type: string }",
		"enum: [confirmed, override]",
		"override: { type: boolean }",
		"audit_trail:",
	} {
		if !strings.Contains(requestSchema, token) {
			t.Fatalf("openapi ForwarderPackageLinkRequest schema missing %q:\n%s", token, requestSchema)
		}
	}

	unlinkSchema, ok := openAPIComponentSection(raw, "ForwarderPackageUnlinkRequest")
	if !ok {
		t.Fatalf("openapi missing ForwarderPackageUnlinkRequest schema in %s", specPath)
	}
	for _, token := range []string{
		"package_id: { type: string }",
		"source: { type: string }",
		"notes: { type: string }",
		"actor: { type: string }",
		"audit_trail:",
	} {
		if !strings.Contains(unlinkSchema, token) {
			t.Fatalf("openapi ForwarderPackageUnlinkRequest schema missing %q:\n%s", token, unlinkSchema)
		}
	}

	linkSchema, ok := openAPIComponentSection(raw, "ForwarderPackageLink")
	if !ok {
		t.Fatalf("openapi missing ForwarderPackageLink schema in %s", specPath)
	}
	for _, token := range []string{
		"package_id: { type: string }",
		"item_id: { type: string }",
		"expected_arrival_id: { type: string }",
		"decision: { type: string }",
		"audit_trail:",
		"updated_at: { type: string }",
	} {
		if !strings.Contains(linkSchema, token) {
			t.Fatalf("openapi ForwarderPackageLink schema missing %q:\n%s", token, linkSchema)
		}
	}

	eventSchema, ok := openAPIComponentSection(raw, "ForwarderPackageLinkEvent")
	if !ok {
		t.Fatalf("openapi missing ForwarderPackageLinkEvent schema in %s", specPath)
	}
	for _, token := range []string{
		"action:",
		"enum: [confirmed, override, unlinked]",
		"previous_item_id: { type: string }",
		"previous_expected_arrival_id: { type: string }",
		"audit_trail:",
		"created_at: { type: string }",
	} {
		if !strings.Contains(eventSchema, token) {
			t.Fatalf("openapi ForwarderPackageLinkEvent schema missing %q:\n%s", token, eventSchema)
		}
	}
}

func TestOpenAPIDocumentsForwarderActiveProfileIsolation(t *testing.T) {
	t.Parallel()

	specPath, raw := readOpenAPISpec(t)

	linkSection, ok := openAPIPathSection(raw, "/api/forwarding/package-links")
	if !ok {
		t.Fatalf("openapi missing /api/forwarding/package-links path in %s", specPath)
	}
	for _, token := range []string{
		"Returns only active-profile package reconciliation links",
		"cross-profile evidence is rejected without creating links or audit events",
		"Active-profile forwarder package id whose current link should be unlinked",
	} {
		if !strings.Contains(linkSection, token) {
			t.Fatalf("openapi /api/forwarding/package-links section missing active-profile token %q:\n%s", token, linkSection)
		}
	}

	suggestionSection, ok := openAPIPathSection(raw, "/api/forwarding/package-match-suggestions")
	if !ok {
		t.Fatalf("openapi missing /api/forwarding/package-match-suggestions path in %s", specPath)
	}
	for _, token := range []string{
		"derived only from active-profile forwarder packages and purchase-arrival evidence",
		"package_id from another profile returns an empty scoped result with zero summary counts",
		"Optional confidence bucket filter applied after active-profile scoping",
	} {
		if !strings.Contains(suggestionSection, token) {
			t.Fatalf("openapi /api/forwarding/package-match-suggestions section missing active-profile token %q:\n%s", token, suggestionSection)
		}
	}

	requestSchema, ok := openAPIComponentSection(raw, "ForwarderPackageLinkRequest")
	if !ok {
		t.Fatalf("openapi missing ForwarderPackageLinkRequest schema in %s", specPath)
	}
	if !strings.Contains(requestSchema, "outside the active profile are rejected without link or audit-event mutation") {
		t.Fatalf("openapi ForwarderPackageLinkRequest schema missing active-profile rejection description:\n%s", requestSchema)
	}

	suggestionSchema, ok := openAPIComponentSection(raw, "ForwarderPackageMatchSuggestion")
	if !ok {
		t.Fatalf("openapi missing ForwarderPackageMatchSuggestion schema in %s", specPath)
	}
	if !strings.Contains(suggestionSchema, "Non-mutating active-profile package-to-purchase-arrival suggestion") {
		t.Fatalf("openapi ForwarderPackageMatchSuggestion schema missing active-profile suggestion description:\n%s", suggestionSchema)
	}
}

func openAPIPathSection(raw []byte, path string) (string, bool) {
	return indentedYAMLSection(raw, "  "+path+":", 2)
}

func openAPIComponentSection(raw []byte, component string) (string, bool) {
	return indentedYAMLSection(raw, "    "+component+":", 4)
}

func indentedYAMLSection(raw []byte, marker string, indent int) (string, bool) {
	lines := regexp.MustCompile(`\r?\n`).Split(string(raw), -1)
	for idx, line := range lines {
		if line != marker {
			continue
		}
		end := len(lines)
		for scan := idx + 1; scan < len(lines); scan++ {
			next := lines[scan]
			if next == "" {
				continue
			}
			if len(next)-len(strings.TrimLeft(next, " ")) <= indent {
				end = scan
				break
			}
		}
		return strings.Join(lines[idx:end], "\n"), true
	}
	return "", false
}

func readOpenAPISpec(t *testing.T) (string, []byte) {
	t.Helper()

	root, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	specPath := filepath.Clean(filepath.Join(root, "..", "..", "docs", "api", "openapi.yaml"))
	raw, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read openapi: %v", err)
	}
	return specPath, raw
}

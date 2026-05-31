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

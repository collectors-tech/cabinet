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
		"/api/chat/threads",
		"/api/chat/messages",
		"/api/chat/attachments",
		"/api/chat/actions/preview",
		"/api/chat/actions/apply",
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

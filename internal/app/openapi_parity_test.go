package app

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestOpenAPIDocumentsRuntimeEndpoints(t *testing.T) {
	t.Parallel()

	root, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	specPath := filepath.Clean(filepath.Join(root, "..", "..", "docs", "api", "openapi.yaml"))
	raw, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read openapi: %v", err)
	}

	requiredPaths := []string{
		"/api/onboarding/sample-data",
		"/api/scanner/run/scheduled",
		"/api/scanner/query-sets/{querySetID}",
		"/api/scanner/failures",
		"/api/scanner/failures/retry",
		"/api/provider/health",
		"/api/settings/reset-ignore-rules",
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
		"/api/items/{itemID}/photos/{photoID}/file",
		"/api/items/{itemID}/photos-rebuild",
		"/api/barcodes/{barcode}",
		"/api/barcodes/{barcode}/external-search",
	}

	for _, endpoint := range requiredPaths {
		pathPattern := regexp.MustCompile(fmt.Sprintf(`(?m)^  %s:$`, regexp.QuoteMeta(endpoint)))
		if !pathPattern.Match(raw) {
			t.Fatalf("openapi missing %s path in %s", endpoint, specPath)
		}
	}
}

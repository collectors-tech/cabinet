package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompanionDiscoveryWishlistHandoffIsBoundToReleaseEvidence(t *testing.T) {
	t.Parallel()

	required := map[string][]string{
		filepath.Join("..", "openspec", "specs", "integrations", "browser-companion", "spec.md"): {
			"Requirement INTEGRATION-080", "not_in_collection", "add_to_wishlist", "MUST NOT mark the item owned",
			"terminal status MUST be `succeeded`", "historical Browser Companion row persisted as `completed` MUST be reported as `succeeded`",
		},
		filepath.Join("..", "openspec", "traceability.md"): {
			"`INTEGRATION-080`", "#2065", "#1944/#1945", "TestCompanionProviderCaptureIsReviewableAndPersistsWishlistHandoff",
			"`PROVIDER-FAMILY-011`", "#2054", "#2032/#2064", "canonical `succeeded` Market Watch history",
		},
		filepath.Join("..", "openspec", "migration", "beta-packaged-core-workflow-acceptance.md"): {
			"A user-present real Frontline search persists an observation, appears through `GET /api/discovery/not-in-collection`, accepts reviewed `add_to_wishlist`, and persists exactly one linked Wishlist row visible through `GET /api/wishlist`.",
			"A user-present real Bonza search after normal browser interaction persists an observation, appears through `GET /api/discovery/not-in-collection`, accepts reviewed `add_to_wishlist`, and persists exactly one linked Wishlist row visible through `GET /api/wishlist`.",
			"A stalled or unavailable Frontline request returns within the bounded provider timeout, records no candidates or false success, and leaves the next provider run usable.",
			"A stalled or unavailable Bonza request returns within the bounded provider timeout, records no candidates or false success, and leaves the next provider run usable.",
			"#2054 / `PROVIDER-FAMILY-011`",
		},
	}
	for path, fragments := range required {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, fragment := range fragments {
			if !strings.Contains(string(raw), fragment) {
				t.Fatalf("%s missing %q", path, fragment)
			}
		}
	}
}

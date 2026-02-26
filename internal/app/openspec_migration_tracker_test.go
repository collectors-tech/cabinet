package app

import (
	"os"
	"strings"
	"testing"
)

func TestShopProvidersHistoryMigrationTrackerEntry(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../openspec/migration/docs-history-migration-todo.md")
	if err != nil {
		t.Fatalf("read migration tracker: %v", err)
	}
	src := string(b)

	required := []string{
		"82294546bf0b715fe49394e1c5a885d3045294d2:docs/SHOP_PROVIDERS.md",
		"INTEGRATION-001..015",
		"OPS-001",
		"OpenAPI changed: `no`",
		"TestProviderSpecsExistAndRegistryLinksThem",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("migration tracker missing token: %s", token)
		}
	}
}


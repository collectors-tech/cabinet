package app

import (
	"os"
	"strings"
	"testing"
)

func TestOpenSpecMigrationCatalogIsFinalized(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../docs/OPENSPEC_MIGRATION_CATALOG.md")
	if err != nil {
		t.Fatalf("read migration catalog: %v", err)
	}
	src := string(b)

	if strings.Contains(src, "PARTIAL") {
		t.Fatalf("migration catalog still contains PARTIAL status markers")
	}
	if strings.Contains(src, "PENDING") {
		t.Fatalf("migration catalog still contains PENDING status markers")
	}
	if !strings.Contains(src, "Tracking issue: `#171`") {
		t.Fatalf("migration catalog must reference lock-in issue #171")
	}
}

func TestLegacyDocsStatusAndSourceOfTruthContract(t *testing.T) {
	t.Parallel()

	read := func(path string) string {
		t.Helper()
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		return string(b)
	}

	readme := read("../../docs/README.md")
	if !strings.Contains(readme, "OpenSpec is the normative source of requirements") {
		t.Fatalf("docs/README.md must declare OpenSpec as normative requirements source")
	}

	legacy := read("../../docs/LEGACY_DOCS_STATUS.md")
	requiredLegacyTokens := []string{
		"docs/FULL_FEATURE_LIST.md",
		"docs/SPEC.md",
		"docs/USE_CASES_AND_SCENARIOS.md",
		"docs/ui-spec/02-SCREEN-SPECS.md",
		"docs/ui-spec/05-TEST-MATRIX-UI.md",
		"MIGRATED_TO_OPENSPEC",
	}
	for _, token := range requiredLegacyTokens {
		if !strings.Contains(legacy, token) {
			t.Fatalf("legacy doc status missing token: %s", token)
		}
	}
}

package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUIFoundationComponentsSpecHasTestabilityArtifacts(t *testing.T) {
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}

	specPath := filepath.Join(repoRoot, "openspec", "specs", "general", "ui-foundation-components", "spec.md")
	specBytes, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	spec := string(specBytes)

	requiredSections := []string{
		"## Acceptance Criteria",
		"## Success Criteria",
		"## E2E and Integration Test Mapping Requirements",
	}
	for _, section := range requiredSections {
		if !strings.Contains(spec, section) {
			t.Fatalf("missing required section %q in %s", section, specPath)
		}
	}

	e2ePath := filepath.Join(repoRoot, "ui.web", "cypress", "e2e", "general", "ui-foundation-components", "spec.cy.ts")
	if _, err := os.Stat(e2ePath); err != nil {
		t.Fatalf("missing component E2E mapping file %s: %v", e2ePath, err)
	}
}

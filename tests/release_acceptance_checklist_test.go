package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackagedCoreWorkflowAcceptanceChecklistCoversIssue1869(t *testing.T) {
	t.Parallel()

	repoRoot := resolveRepoRoot(t)
	checklistPath := filepath.Join(repoRoot, "openspec", "migration", "beta-packaged-core-workflow-acceptance.md")
	raw, err := os.ReadFile(checklistPath)
	if err != nil {
		t.Fatalf("read packaged acceptance checklist: %v", err)
	}
	content := string(raw)

	requiredFragments := []string{
		"Issue: #1869",
		"Windows portable beta package from #1868",
		"#1864 approval required",
		"Cabinet package SHA-256",
		"Cabinet source commit SHA",
		"Browser Companion package SHA-256",
		"/api/runtime.app_version",
		"Fresh start and onboarding/profile setup",
		"Inventory item can be created, edited, searched, filtered, reloaded",
		"Media can be attached, marked primary",
		"Wishlist item can be created, reprioritised, status-updated, and marked purchased into Inventory",
		"Collection can be created/edited, receive/move an item, soft-delete safely, and protect All Items",
		"Data export and backup",
		"#1937 media migration evidence records discovered, migrated, already-migrated, duplicate, skipped, failed, and orphan counts",
		"Backup restore into an isolated target preserves core record counts and relationships",
		"Voglers",
		"Hobbytech",
		"Frontline",
		"Bonza",
		"Pair to Cabinet through #2033",
		"canonical asset manifest/layout",
		"Discovery review can hand an item to Wishlist or Inventory",
		"One failed provider and one invalid import/restore input",
		"Persistence is verified after reload and application restart",
		"Active-profile isolation",
		"No raw translation keys, placeholder security claims, or unsigned-installer claims",
		"Every failure creates or links a focused GitHub issue",
		"rerun after release-blocking fixes",
		"uses the packaged binary, not a dev server",
		"does not require test-only hooks",
		"does not merge `develop` into `main`",
		"final approval follows packaged acceptance",
	}
	for _, fragment := range requiredFragments {
		if !strings.Contains(content, fragment) {
			t.Fatalf("packaged acceptance checklist missing %q", fragment)
		}
	}
}

func TestPackagedAcceptanceTraceabilityStaysBoundToOpenSpec(t *testing.T) {
	t.Parallel()

	repoRoot := resolveRepoRoot(t)
	traceabilityPath := filepath.Join(repoRoot, "openspec", "traceability.md")
	raw, err := os.ReadFile(traceabilityPath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}
	traceability := string(raw)

	requiredFragments := []string{
		"`RUNTIME-CORE-020`",
		"#1869",
		"openspec/migration/beta-packaged-core-workflow-acceptance.md",
		"TestPackagedCoreWorkflowAcceptanceChecklistCoversIssue1869",
		"TestPackagedAcceptanceTraceabilityStaysBoundToOpenSpec",
		"partial",
	}
	for _, fragment := range requiredFragments {
		if !strings.Contains(traceability, fragment) {
			t.Fatalf("traceability missing %q", fragment)
		}
	}
}

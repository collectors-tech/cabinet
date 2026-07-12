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
		"Package SHA-256",
		"Source commit SHA",
		"/api/runtime.app_version",
		"Fresh start and onboarding/profile setup",
		"Inventory item can be created, edited, searched, filtered, reloaded",
		"Media can be attached, marked primary",
		"Wishlist item can be created, reprioritised, status-updated, and marked purchased into Inventory",
		"Collection can be created/edited, receive/move an item, soft-delete safely, and protect All Items",
		"Data export and backup",
		"Backup restore into an isolated target preserves core record counts and relationships",
		"A saved Market Watch can run against the chosen live beta provider",
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

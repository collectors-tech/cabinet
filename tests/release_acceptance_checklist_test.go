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
		"/api/runtime.build_revision",
		"equals the Cabinet manifest `source_commit`",
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

func TestPackagedAcceptanceRecorderIsResumableFailClosedAndNonPublishing(t *testing.T) {
	t.Parallel()

	repoRoot := resolveRepoRoot(t)
	for relativePath, required := range map[string][]string{
		filepath.Join("openspec", "migration", "beta-packaged-core-workflow-acceptance.md"): {
			"scripts/record-beta-acceptance.mjs",
			"not_run`, `blocked`, `pass`, or `fail`",
			"candidate fingerprint",
			"stale candidate",
			"operator-confirmed",
			"JSON and Markdown",
		},
		filepath.Join("openspec", "specs", "general", "runtime-core", "spec.md"): {
			"resumable evidence recorder",
			"exact-candidate fingerprint",
			"MUST NOT auto-pass",
			"redact credentials, bearer and cookie material, private page content, and sensitive local paths",
		},
		filepath.Join("openspec", "traceability.md"): {
			"#2048",
			"acceptance-evidence-recorder.test.mjs",
			"TestPackagedAcceptanceRecorderIsResumableFailClosedAndNonPublishing",
		},
		"package.json": {
			"acceptance:record",
			"scripts/record-beta-acceptance.mjs",
		},
	} {
		raw, err := os.ReadFile(filepath.Join(repoRoot, relativePath))
		if err != nil {
			t.Fatalf("read %s: %v", relativePath, err)
		}
		content := string(raw)
		for _, fragment := range required {
			if !strings.Contains(content, fragment) {
				t.Errorf("%s missing acceptance-recorder contract %q", relativePath, fragment)
			}
		}
	}
}

func TestSecondPCGAAcceptancePlanIsExecutableAndFailClosed(t *testing.T) {
	t.Parallel()

	repoRoot := resolveRepoRoot(t)
	planPath := filepath.Join(repoRoot, "openspec", "migration", "cabinet-1.0-ga-second-pc-test-plan.md")
	raw, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("read second-PC GA acceptance plan: %v", err)
	}
	content := string(raw)

	requiredFragments := []string{
		"Cabinet 1.0 GA Second-PC Acceptance Test Plan",
		"Issue: #1869",
		"Do not start the acceptance run",
		"exact candidate bundle from #1868",
		"Cabinet portable ZIP",
		"Chrome Companion ZIP",
		"Edge Companion ZIP",
		"Get-FileHash",
		"/api/runtime",
		"scripts/record-beta-acceptance.mjs init",
		"scripts/record-beta-acceptance.mjs record",
		"IDENTITY-01..11",
		"COLLECTOR-01..10",
		"PROVIDER-01..15",
		"CROSS-01..06",
		"FAILURE-01..05",
		"SHORTCUT-01..04",
		"separate Chrome and Edge evidence packs",
		"owner-approved GA scope",
		"all 51 rows",
		"Candidate invalidation",
		"Do not record credentials, tokens, cookies",
		"fail_with_blockers",
		"#1867",
		"#1864",
		"does not publish",
		"does not promote `develop` to `main`",
	}
	for _, fragment := range requiredFragments {
		if !strings.Contains(content, fragment) {
			t.Errorf("second-PC GA acceptance plan missing %q", fragment)
		}
	}

	for _, relativePath := range []string{
		filepath.Join("openspec", "migration", "beta-packaged-core-workflow-acceptance.md"),
		filepath.Join("openspec", "traceability.md"),
	} {
		linked, err := os.ReadFile(filepath.Join(repoRoot, relativePath))
		if err != nil {
			t.Fatalf("read %s: %v", relativePath, err)
		}
		if !strings.Contains(string(linked), "openspec/migration/cabinet-1.0-ga-second-pc-test-plan.md") {
			t.Errorf("%s does not link the second-PC GA acceptance plan", relativePath)
		}
	}
}

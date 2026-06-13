package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsolatedCypressHarnessDocsCoverPrerequisitesAndFallbacks(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join("..", "docs", "testing", "isolated-cypress-harness.md"))
	if err != nil {
		t.Fatalf("read isolated Cypress harness docs: %v", err)
	}
	content := string(raw)

	requiredFragments := []string{
		"# Isolated Cypress Harness",
		"## Local Prerequisites",
		"Docker Desktop",
		"docker build -t cabinet:e2e .",
		"pwsh -NoLogo -NoProfile -File .\\scripts\\run-cypress-matrix.ps1",
		"-UseContainerImage",
		"-ApiContractSmoke",
		"-RequireE2EHooks",
		".work-agent\\logs\\cypress-matrix\\<run-id>\\matrix.summary.json",
		"## Fallback When Docker Is Unavailable",
		"Do not reuse stale shared desktop runtimes as isolated-lane proof.",
		"## Failure Stages",
		"container_image",
		"container_start",
		"runtime_health",
		"cypress",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(content, fragment) {
			t.Fatalf("isolated Cypress harness docs missing required fragment %q", fragment)
		}
	}
}

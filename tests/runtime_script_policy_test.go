package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCypressScriptPrefersProjectBinRuntimePath(t *testing.T) {
	t.Parallel()

	scriptPath := filepath.Join("..", "cypress.ps1")
	raw, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read cypress script: %v", err)
	}
	content := string(raw)
	if !strings.Contains(content, `Join-Path $repoRoot "bin/cabinet.exe"`) {
		t.Fatalf("expected cypress runtime resolution to prefer project-local bin executable")
	}
	if !strings.Contains(content, "Runtime executable resolved") {
		t.Fatalf("expected cypress script to log resolved runtime executable path")
	}
}

func TestCypressScriptRejectsEphemeralTempRuntimePathByDefault(t *testing.T) {
	t.Parallel()

	scriptPath := filepath.Join("..", "cypress.ps1")
	raw, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read cypress script: %v", err)
	}
	content := string(raw)
	if !strings.Contains(content, "$AllowTempRuntimePath") {
		t.Fatalf("expected explicit allow flag for temp runtime executable path")
	}
	if !strings.Contains(content, "ephemeral temp/template runtime path was rejected") {
		t.Fatalf("expected script to reject temp/template runtime path by default")
	}
}

func TestCypressScriptDisablesBrowserAutoOpenForManagedRuns(t *testing.T) {
	t.Parallel()

	scriptPath := filepath.Join("..", "cypress.ps1")
	raw, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read cypress script: %v", err)
	}
	content := string(raw)
	if !strings.Contains(content, "--no-open-browser") {
		t.Fatalf("expected managed cypress runtime startup to pass --no-open-browser")
	}
}

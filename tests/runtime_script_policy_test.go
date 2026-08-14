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

func TestCypressScriptFailsOnStaleRuntimeAppVersion(t *testing.T) {
	t.Parallel()

	scriptPath := filepath.Join("..", "cypress.ps1")
	raw, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read cypress script: %v", err)
	}
	content := string(raw)
	for _, snippet := range []string{
		"[switch]$AllowStaleRuntimeVersion",
		"Assert-RuntimeBuildIdentityMatchesSourceCommit",
		"Runtime build revision mismatch",
		"/api/runtime build_revision=",
		"app_version=$appVersion",
		"Test-AppVersionMatchesSourceCommit",
		"allow_stale_runtime_version",
		"-AllowStaleRuntimeVersion",
	} {
		if !strings.Contains(content, snippet) {
			t.Fatalf("expected cypress script to include stale runtime guard snippet %q", snippet)
		}
	}
}

func TestCypressScriptSupportsFixedPackSpecLists(t *testing.T) {
	t.Parallel()

	scriptPath := filepath.Join("..", "cypress.ps1")
	raw, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read cypress script: %v", err)
	}
	content := string(raw)
	for _, snippet := range []string{
		"Resolve-CypressSpecArgument",
		"$specValue -split \",\"",
		"Resolve-Path $candidate",
		"$resolvedSpecs -join \",\"",
		"Missing Cypress spec: no spec paths were provided.",
	} {
		if !strings.Contains(content, snippet) {
			t.Fatalf("expected cypress script to support comma-separated fixed pack specs with snippet %q", snippet)
		}
	}
}

func TestCypressScriptResetsManagedRuntimeDataDir(t *testing.T) {
	t.Parallel()

	scriptPath := filepath.Join("..", "cypress.ps1")
	raw, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read cypress script: %v", err)
	}
	content := string(raw)
	for _, snippet := range []string{
		"Reset-CypressRuntimeDataDir",
		"Clearing managed Cypress runtime data dir",
		"Remove-Item -LiteralPath $runtimeDataDir -Recurse -Force",
	} {
		if !strings.Contains(content, snippet) {
			t.Fatalf("expected cypress script to reset managed runtime data dir with snippet %q", snippet)
		}
	}
}

func TestCypressScriptRecordsFailClosedExecutionTimeoutEvidence(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile(filepath.Join("..", "cypress.ps1"))
	if err != nil {
		t.Fatalf("read cypress script: %v", err)
	}
	content := string(raw)
	for _, snippet := range []string{
		"[int]$ExecutionTimeoutSec = 300", "Invoke-CypressOwnedProcess",
		"scripts\\run-cypress.mjs", "$args | Select-Object -Skip 1", "-ArgumentList $watchdogArgs", "execution_timeout_sec",
		"Cypress application data isolated", "Join-Path $e2eDataDir \"cypress-appdata\"", "$env:APPDATA = $cypressAppDataDir",
		"execution_timed_out", "runner_phase", "cypress_root_pid",
		"cypress_child_pids", "cypress_process_tree", "cypress_elapsed_ms", "last_cypress_output",
		"cypress_cleanup_result", "runtime_revision", "runtime_port",
	} {
		if !strings.Contains(content, snippet) {
			t.Fatalf("expected Cypress timeout summary contract snippet %q", snippet)
		}
	}
	for _, forbidden := range []string{"Get-Process cypress", "Get-Process chrome", "Get-Process electron"} {
		if strings.Contains(strings.ToLower(content), strings.ToLower(forbidden)) {
			t.Fatalf("Cypress timeout cleanup must not enumerate unrelated processes with %q", forbidden)
		}
	}
	watchdogRaw, err := os.ReadFile(filepath.Join("..", "scripts", "lib", "cypress-process-watchdog.ps1"))
	if err != nil {
		t.Fatalf("read Cypress watchdog helper: %v", err)
	}
	watchdog := string(watchdogRaw)
	for _, snippet := range []string{"Get-CypressProcessTreeSnapshot", "ConvertTo-CypressRedactedCommandLine", "process_tree = @($processTree)"} {
		if !strings.Contains(watchdog, snippet) {
			t.Fatalf("expected Cypress process-tree diagnostic contract snippet %q", snippet)
		}
	}
}

func TestCabinetBuildScriptInjectsExplicitRuntimeRevision(t *testing.T) {
	t.Parallel()

	scriptPath := filepath.Join("..", "scripts", "build-cabinet.ps1")
	raw, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read build script: %v", err)
	}
	content := string(raw)
	for _, snippet := range []string{
		"git -C $repoRoot rev-parse HEAD",
		"github.com/collectors-tech/cabinet/internal/app.buildRevision",
		"github.com/collectors-tech/cabinet/internal/app.buildDate",
		"-ldflags",
	} {
		if !strings.Contains(content, snippet) {
			t.Fatalf("expected build script to include explicit revision stamping snippet %q", snippet)
		}
	}
}

func TestRuntimeMetadataUsesExplicitRevisionFallback(t *testing.T) {
	t.Parallel()

	appPath := filepath.Join("..", "internal", "app", "app.go")
	raw, err := os.ReadFile(appPath)
	if err != nil {
		t.Fatalf("read app runtime metadata: %v", err)
	}
	content := string(raw)
	for _, snippet := range []string{
		"buildRevision",
		"buildDate",
		`version = "rev-" + short`,
	} {
		if !strings.Contains(content, snippet) {
			t.Fatalf("expected runtime metadata to include explicit revision fallback snippet %q", snippet)
		}
	}
}

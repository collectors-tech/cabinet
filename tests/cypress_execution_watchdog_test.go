package tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCypressExecutionWatchdogUsesWindowsCommandShim(t *testing.T) {
	scriptPath := filepath.Join("..", "cypress.ps1")
	raw, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read cypress script: %v", err)
	}
	content := string(raw)
	for _, snippet := range []string{
		"Resolve-CypressRunnerCommand",
		`if ($IsWindows) {`,
		`return "npx.cmd"`,
		`$cypressCommand = Resolve-CypressRunnerCommand`,
		`Start-Process -FilePath $cypressCommand`,
	} {
		if !strings.Contains(content, snippet) {
			t.Fatalf("expected cypress script to resolve the Windows npx.cmd command shim with snippet %q", snippet)
		}
	}
	if strings.Contains(content, `Start-Process -FilePath "npx"`) {
		t.Fatalf("cypress watchdog still launches extensionless npx, which resolves to a POSIX shim on some Windows hosts")
	}
}

func TestCypressExecutionWatchdogFailsClosedOnHungRunner(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("cypress.ps1 watchdog execution contract is Windows PowerShell-specific")
	}

	repoRoot := resolveRepoRoot(t)
	tempDir := t.TempDir()
	logDir := filepath.Join(tempDir, "logs")
	fakeBin := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(fakeBin, 0o755); err != nil {
		t.Fatalf("create fake bin: %v", err)
	}
	npxPath := filepath.Join(fakeBin, "npx.cmd")
	npxScript := strings.Join([]string{
		"@echo off",
		"powershell -NoLogo -NoProfile -Command \"Write-Output 'fake Cypress process alive before run start'; Start-Sleep -Seconds 30\"",
		"",
	}, "\r\n")
	if err := os.WriteFile(npxPath, []byte(npxScript), 0o755); err != nil {
		t.Fatalf("write fake npx: %v", err)
	}

	cmd := exec.Command(
		"pwsh",
		"-NoLogo",
		"-NoProfile",
		"-File", filepath.Join(repoRoot, "cypress.ps1"),
		"-Spec", "cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts",
		"-Browser", "chrome",
		"-NoServer",
		"-SkipDependencyPrep",
		"-SkipRuntimeBuild",
		"-RuntimeExecutablePath", filepath.Join(repoRoot, "cypress.ps1"),
		"-AllowTempRuntimePath",
		"-ExecutionTimeoutSec", "2",
		"-LogDir", logDir,
		"-LogName", "watchdog-hang-fixture",
	)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(), "PATH="+fakeBin+";"+os.Getenv("PATH"))
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected cypress watchdog command to fail closed, output:\n%s", string(output))
	}
	if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 124 {
		t.Fatalf("expected exit code 124, got err=%v output:\n%s", err, string(output))
	}

	matches, err := filepath.Glob(filepath.Join(logDir, "*watchdog-hang-fixture.summary.json"))
	if err != nil {
		t.Fatalf("glob summary: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected one watchdog summary, got %d in %s; output:\n%s", len(matches), logDir, string(output))
	}
	raw, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("read summary: %v", err)
	}
	var summary struct {
		ExitCode             int      `json:"exit_code"`
		RunnerPhase          string   `json:"runner_phase"`
		CypressElapsedMs     int64    `json:"cypress_elapsed_ms"`
		CypressLastOutput    []string `json:"cypress_last_output"`
		CypressCleanupResult string   `json:"cypress_cleanup_result"`
		ExecutionTimeoutSec  int      `json:"execution_timeout_sec"`
	}
	if err := json.Unmarshal(raw, &summary); err != nil {
		t.Fatalf("parse summary: %v\n%s", err, string(raw))
	}
	if summary.ExitCode != 124 || summary.RunnerPhase != "execution_timeout" || summary.ExecutionTimeoutSec != 2 {
		t.Fatalf("unexpected timeout summary: %+v", summary)
	}
	if summary.CypressElapsedMs < 1500 || summary.CypressElapsedMs > 10000 {
		t.Fatalf("expected bounded elapsed time around the 2s timeout, got %d ms", summary.CypressElapsedMs)
	}
	if !strings.Contains(summary.CypressCleanupResult, "stopped_pids=") {
		t.Fatalf("expected cleanup evidence, got %q", summary.CypressCleanupResult)
	}
	if strings.Contains(strings.Join(summary.CypressLastOutput, "\n"), "Run Starting") {
		t.Fatalf("hang fixture should not fake Cypress Run Starting output: %+v", summary.CypressLastOutput)
	}
}

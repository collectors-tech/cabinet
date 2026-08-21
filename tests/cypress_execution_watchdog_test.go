package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

type cypressWatchdogResult struct {
	TimedOut    bool   `json:"timed_out"`
	ExitCode    int    `json:"exit_code"`
	RunnerPhase string `json:"runner_phase"`
	RootPID     int    `json:"root_pid"`
	ChildPIDs   []int  `json:"child_pids"`
	ProcessTree []struct {
		PID         int    `json:"pid"`
		ParentPID   int    `json:"parent_pid"`
		Name        string `json:"name"`
		CommandLine string `json:"command_line"`
	} `json:"process_tree"`
	ElapsedMS     int64    `json:"elapsed_ms"`
	LastOutput    []string `json:"last_output"`
	CleanupResult string   `json:"cleanup_result"`
}

func TestCypressExecutionWatchdogFailsClosedAndPreservesUnrelatedProcess(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows PowerShell workflow")
	}
	repoRoot := resolveRepoRoot(t)
	watchdogPath := filepath.Join(repoRoot, "scripts", "lib", "cypress-process-watchdog.ps1")
	tempDir := t.TempDir()
	fixturePath := filepath.Join(tempDir, "hanging-cypress-fixture.cmd")
	harnessPath := filepath.Join(tempDir, "watchdog-harness.ps1")
	stdoutPath := filepath.Join(tempDir, "fixture.stdout.log")
	stderrPath := filepath.Join(tempDir, "fixture.stderr.log")
	fixture := "@echo off\r\nstart \"\" /b pwsh -NoLogo -NoProfile -Command \"Start-Sleep -Seconds 120\"\r\necho fixture-ready owned_child_started\r\necho token=super-secret-fixture-value\r\nping 127.0.0.1 -n 120 >nul\r\n"
	if err := os.WriteFile(fixturePath, []byte(fixture), 0600); err != nil {
		t.Fatal(err)
	}
	harness := fmt.Sprintf(". %q\n$result = Invoke-CypressOwnedProcess -FilePath \"cmd.exe\" -ArgumentList @(\"/d\", \"/s\", \"/c\", %q, \"--token=super-secret-command-line\") -WorkingDirectory %q -TimeoutSec 1 -StandardOutputPath %q -StandardErrorPath %q -OutputTailLineCount 20\n$result | ConvertTo-Json -Depth 6 -Compress\n", watchdogPath, fixturePath, tempDir, stdoutPath, stderrPath)
	if err := os.WriteFile(harnessPath, []byte(harness), 0600); err != nil {
		t.Fatal(err)
	}
	unrelated := exec.Command("pwsh", "-NoLogo", "-NoProfile", "-Command", "Start-Sleep -Seconds 120")
	if err := unrelated.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = unrelated.Process.Kill(); _, _ = unrelated.Process.Wait() })
	cmd := exec.Command("pwsh", "-NoLogo", "-NoProfile", "-File", harnessPath)
	output, err := runWatchdogCommand(cmd, 15*time.Second)
	if err != nil {
		t.Fatalf("watchdog harness: %v\n%s", err, output)
	}
	var result cypressWatchdogResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
		t.Fatalf("parse: %v\n%s", err, output)
	}
	if !result.TimedOut || result.ExitCode != 124 || result.RunnerPhase != "execution_timeout" {
		t.Fatalf("timeout outcome: %+v", result)
	}
	if result.RootPID <= 0 || len(result.ChildPIDs) == 0 {
		t.Fatalf("owned tree missing: %+v", result)
	}
	if len(result.ProcessTree) == 0 {
		t.Fatalf("process-tree diagnostics missing: %+v", result)
	}
	sawRedactedCommandLine := false
	for _, process := range result.ProcessTree {
		if process.PID <= 0 {
			t.Fatalf("invalid process-tree diagnostics: %+v", process)
		}
		if strings.Contains(process.CommandLine, "super-secret-command-line") {
			t.Fatalf("process-tree diagnostics leaked command-line secret: %+v", process)
		}
		if strings.Contains(process.CommandLine, "token=[REDACTED]") {
			sawRedactedCommandLine = true
		}
	}
	if !sawRedactedCommandLine {
		t.Fatalf("process-tree diagnostics did not retain redacted command evidence: %+v", result.ProcessTree)
	}
	if result.ElapsedMS < 900 || result.ElapsedMS > 10000 {
		t.Fatalf("elapsed outside bound: %+v", result)
	}
	tail := strings.Join(result.LastOutput, "\n")
	if !strings.Contains(tail, "fixture-ready owned_child_started") || strings.Contains(tail, "super-secret-fixture-value") || !strings.Contains(tail, "token=[REDACTED]") {
		t.Fatalf("unsafe output: %+v", result)
	}
	if !strings.Contains(result.CleanupResult, "owned_process_tree_stopped") {
		t.Fatalf("cleanup: %+v", result)
	}
	source, err := os.ReadFile(watchdogPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"ObservedChildProcessIds",
		"Test-CypressObservedCleanupCandidate",
		"current command-line evidence",
		"Cypress\\\\cy\\\\production\\\\browsers",
	} {
		if !strings.Contains(string(source), required) {
			t.Fatalf("watchdog cleanup missing guarded observed-process contract %q", required)
		}
	}
	for _, pid := range append([]int{result.RootPID}, result.ChildPIDs...) {
		assertWatchdogProcessStopped(t, pid)
	}
	if unrelated.ProcessState != nil || !watchdogProcessExists(unrelated.Process.Pid) {
		t.Fatalf("unrelated process %d stopped", unrelated.Process.Pid)
	}
}

func TestCypressExecutionWatchdogUsesSanitizedNodeShim(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows executable resolution contract")
	}

	repoRoot := resolveRepoRoot(t)
	content, err := os.ReadFile(filepath.Join(repoRoot, "cypress.ps1"))
	if err != nil {
		t.Fatalf("read cypress.ps1: %v", err)
	}
	script := string(content)
	if !strings.Contains(script, `scripts\run-cypress.mjs`) {
		t.Fatalf("cypress.ps1 must use the environment-sanitizing Cypress launcher")
	}
	if !strings.Contains(script, `@($args | Select-Object -Skip 1)`) || !strings.Contains(script, `-FilePath $nodeCommand`) {
		t.Fatalf("cypress.ps1 must pass CLI arguments after the Cypress subcommand through the owned Node process")
	}
	for _, required := range []string{`Join-Path $e2eDataDir "cypress-appdata"`, `$env:APPDATA = $cypressAppDataDir`, `$env:APPDATA = $originalAppData`} {
		if !strings.Contains(script, required) {
			t.Fatalf("cypress.ps1 must isolate and restore Cypress application data with %q", required)
		}
	}
}

func TestCypressE2EHookProbeDoesNotMisclassifyBoundedResetWorkAsUnavailable(t *testing.T) {
	repoRoot := resolveRepoRoot(t)
	raw, err := os.ReadFile(filepath.Join(repoRoot, "cypress.ps1"))
	if err != nil {
		t.Fatalf("read cypress.ps1: %v", err)
	}
	script := string(raw)
	for _, required := range []string{
		`[int]$StartupTimeoutSec = 90`,
		`$e2eHooksProbeTimeoutSec = [Math]::Min([Math]::Max($StartupTimeoutSec, 3), 20)`,
		`function Test-E2EHooks([string]$url, [int]$timeoutSec)`,
		`-TimeoutSec $timeoutSec -SkipHttpErrorCheck`,
		`$script:LastE2EHooksProbeDiagnostic`,
		`timed out after $timeoutSec seconds`,
		`HTTP $($res.StatusCode)`,
		`Test-E2EHooks $BaseUrl $e2eHooksProbeTimeoutSec`,
		`$script:LastE2EHooksProbeDiagnostic`,
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("cypress E2E hook probe is missing bounded reset diagnostic contract %q", required)
		}
	}
	if strings.Contains(script, `api/test/reset" -Method Post -Body "{}" -ContentType "application/json" -UseBasicParsing -TimeoutSec 2`) {
		t.Fatal("E2E reset probe must not classify legitimate reset work over two seconds as unavailable")
	}
	if !strings.Contains(script, `), 20)`) {
		t.Fatal("E2E reset probe budget must exceed the three-attempt SQLite contention envelope")
	}
}

func runWatchdogCommand(cmd *exec.Cmd, timeout time.Duration) (string, error) {
	var output strings.Builder
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Start(); err != nil {
		return output.String(), err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		return output.String(), err
	case <-time.After(timeout):
		_ = cmd.Process.Kill()
		<-done
		return output.String(), fmt.Errorf("command timed out after %s", timeout)
	}
}
func assertWatchdogProcessStopped(t *testing.T, pid int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if !watchdogProcessExists(pid) {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("owned process %d still exists", pid)
}
func watchdogProcessExists(pid int) bool {
	return exec.Command("powershell", "-NoLogo", "-NoProfile", "-Command", "if (Get-Process -Id "+strconv.Itoa(pid)+" -ErrorAction SilentlyContinue) { exit 0 }; exit 1").Run() == nil
}

package tests

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

var sharedWatchdogPowerShellHost *watchdogPowerShellHost

func TestMain(m *testing.M) {
	if runtime.GOOS != "windows" {
		os.Exit(m.Run())
	}
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Fprintln(os.Stderr, "resolve Cypress watchdog test path")
		os.Exit(1)
	}
	hostTempDir, err := os.MkdirTemp("", "cabinet-watchdog-host-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	watchdogPath := filepath.Join(filepath.Dir(thisFile), "..", "scripts", "lib", "cypress-process-watchdog.ps1")
	hostScriptPath := filepath.Join(hostTempDir, "watchdog-host.ps1")
	sharedWatchdogPowerShellHost, err = startWatchdogPowerShellHost(watchdogPath, hostScriptPath, 15*time.Second)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		_ = os.RemoveAll(hostTempDir)
		os.Exit(1)
	}
	exitCode := m.Run()
	sharedWatchdogPowerShellHost.stop()
	_ = os.RemoveAll(hostTempDir)
	os.Exit(exitCode)
}

func TestCypressExecutionWatchdogFailsClosedAndPreservesUnrelatedProcess(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows PowerShell workflow")
	}
	repoRoot := resolveRepoRoot(t)
	watchdogPath := filepath.Join(repoRoot, "scripts", "lib", "cypress-process-watchdog.ps1")
	tempDir := t.TempDir()
	harnessPath := filepath.Join(tempDir, "watchdog-harness.ps1")
	stdoutPath := filepath.Join(tempDir, "fixture.stdout.log")
	stderrPath := filepath.Join(tempDir, "fixture.stderr.log")
	fixtureCommand := `echo --token=super-secret-command-line >nul & echo fixture-ready owned_child_started & echo token=super-secret-fixture-value & ping 127.0.0.1 -n 120 >nul`
	harness := fmt.Sprintf("$result = Invoke-CypressOwnedProcess -FilePath \"cmd.exe\" -ArgumentList @(\"/d\", \"/s\", \"/c\", '%s') -WorkingDirectory %q -TimeoutSec 1 -StandardOutputPath %q -StandardErrorPath %q -OutputTailLineCount 20\n$result | ConvertTo-Json -Depth 6 -Compress\n", fixtureCommand, tempDir, stdoutPath, stderrPath)
	if err := os.WriteFile(harnessPath, []byte(harness), 0600); err != nil {
		t.Fatal(err)
	}
	unrelated := exec.Command("ping", "127.0.0.1", "-n", "120")
	if err := unrelated.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = unrelated.Process.Kill(); _, _ = unrelated.Process.Wait() })
	// cypress.ps1 invokes the watchdog from an already-running PowerShell host.
	// Warm only that host before measuring the watchdog operation so unrelated
	// PowerShell startup contention cannot consume the unchanged 15-second guard.
	if sharedWatchdogPowerShellHost == nil {
		t.Fatal("PowerShell watchdog host unavailable")
	}
	output, err := sharedWatchdogPowerShellHost.invokeScriptWithinDeadline(harnessPath, 15*time.Second)
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
		assertWatchdogProcessStopped(t, pid, result)
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

const watchdogPowerShellReady = "__CABINET_WATCHDOG_POWERSHELL_READY__"

type watchdogPowerShellHost struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	lines  chan string
	done   chan error
	stderr bytes.Buffer
	dead   bool
}

func startWatchdogPowerShellHost(watchdogPath string, hostScriptPath string, timeout time.Duration) (*watchdogPowerShellHost, error) {
	hostScript := fmt.Sprintf("$ErrorActionPreference = 'Stop'\n. '%s'\n[Console]::Out.WriteLine('%s')\nwhile (($invocationPath = [Console]::In.ReadLine()) -ne $null) {\n  if ($invocationPath -eq '__CABINET_WATCHDOG_POWERSHELL_EXIT__') { break }\n  . $invocationPath\n}\n", escapePowerShellSingleQuoted(watchdogPath), watchdogPowerShellReady)
	if err := os.WriteFile(hostScriptPath, []byte(hostScript), 0600); err != nil {
		return nil, err
	}
	cmd := exec.Command("powershell", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", hostScriptPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, err
	}
	host := &watchdogPowerShellHost{cmd: cmd, stdin: stdin, lines: make(chan string, 8), done: make(chan error, 1)}
	cmd.Stderr = &host.stderr
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		return nil, err
	}
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			host.lines <- scanner.Text()
		}
		close(host.lines)
	}()
	go func() { host.done <- cmd.Wait() }()
	line, err := host.readLineWithin(timeout)
	if err != nil || line != watchdogPowerShellReady {
		host.stop()
		return nil, fmt.Errorf("initialize PowerShell watchdog host: ready=%q err=%v\n%s", line, err, host.stderr.String())
	}
	return host, nil
}

func (host *watchdogPowerShellHost) invokeScriptWithinDeadline(scriptPath string, timeout time.Duration) (string, error) {
	if _, err := io.WriteString(host.stdin, scriptPath+"\n"); err != nil {
		return "", err
	}
	line, err := host.readLineWithin(timeout)
	if err != nil {
		host.stop()
		return line + host.stderr.String(), err
	}
	return line, nil
}

func (host *watchdogPowerShellHost) readLineWithin(timeout time.Duration) (string, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case line, ok := <-host.lines:
		if !ok {
			return "", fmt.Errorf("PowerShell host output closed")
		}
		return line, nil
	case err := <-host.done:
		host.dead = true
		if err != nil {
			return "", fmt.Errorf("PowerShell host exited: %w", err)
		}
		return "", fmt.Errorf("PowerShell host exited")
	case <-timer.C:
		return "", fmt.Errorf("command timed out after %s", timeout)
	}
}

func (host *watchdogPowerShellHost) stop() {
	if host == nil || host.dead {
		return
	}
	_, _ = io.WriteString(host.stdin, "__CABINET_WATCHDOG_POWERSHELL_EXIT__\n")
	_ = host.stdin.Close()
	select {
	case <-host.done:
		host.dead = true
	case <-time.After(5 * time.Second):
		_ = host.cmd.Process.Kill()
		<-host.done
		host.dead = true
	}
}

func escapePowerShellSingleQuoted(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func assertWatchdogProcessStopped(t *testing.T, pid int, result cypressWatchdogResult) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if !watchdogProcessExists(pid) {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("owned process %d still exists: %+v", pid, result)
}

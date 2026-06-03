package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestApiContractSmokeScriptWritesMachineReadableSummary(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell harness validation is exercised on the Windows QA lane")
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/healthz":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		case "/api/runtime":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"mode":"test","version":"mock"}`))
		case "/api/openapi.yaml":
			w.Header().Set("Content-Type", "application/yaml")
			_, _ = w.Write([]byte("openapi: 3.0.0\ninfo:\n  title: Cabinet Mock\n"))
		case "/sign-in":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<html><body>Sign in</body></html>"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	repoRoot := filepath.Dir(currentFileDir(t))
	logRoot := t.TempDir()
	runID := "go-api-contract-smoke"
	cmd := exec.Command("pwsh", "-NoLogo", "-NoProfile", "-File", filepath.Join(repoRoot, "scripts", "run-api-contract-smoke.ps1"), "-BaseUrl", srv.URL, "-LogRoot", logRoot, "-RunId", runID)
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("api contract smoke script failed: %v\n%s", err, string(output))
	}

	summaryPath := filepath.Join(logRoot, runID, "api-contract-smoke.summary.json")
	raw, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("read summary: %v", err)
	}

	var summary struct {
		ExitCode    int `json:"exit_code"`
		CheckCount  int `json:"check_count"`
		FailedCount int `json:"failed_count"`
	}
	if err := json.Unmarshal(raw, &summary); err != nil {
		t.Fatalf("decode summary: %v\n%s", err, string(raw))
	}
	if summary.ExitCode != 0 || summary.CheckCount != 4 || summary.FailedCount != 0 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

func TestApiContractSmokeScriptWritesFailureSummary(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell harness validation is exercised on the Windows QA lane")
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/healthz":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		case "/api/runtime":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`not-json`))
		case "/api/openapi.yaml":
			w.Header().Set("Content-Type", "application/yaml")
			_, _ = w.Write([]byte("openapi: 3.0.0\ninfo:\n  title: Cabinet Mock\n"))
		case "/sign-in":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<html><body>Sign in</body></html>"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	repoRoot := filepath.Dir(currentFileDir(t))
	logRoot := t.TempDir()
	runID := "go-api-contract-smoke-failure"
	cmd := exec.Command("pwsh", "-NoLogo", "-NoProfile", "-File", filepath.Join(repoRoot, "scripts", "run-api-contract-smoke.ps1"), "-BaseUrl", srv.URL, "-LogRoot", logRoot, "-RunId", runID)
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("api contract smoke script unexpectedly passed\n%s", string(output))
	}

	summaryPath := filepath.Join(logRoot, runID, "api-contract-smoke.summary.json")
	raw, readErr := os.ReadFile(summaryPath)
	if readErr != nil {
		t.Fatalf("read failure summary: %v\n%s", readErr, string(output))
	}

	var summary struct {
		ExitCode    int `json:"exit_code"`
		CheckCount  int `json:"check_count"`
		FailedCount int `json:"failed_count"`
		Checks      []struct {
			Name   string `json:"name"`
			Passed bool   `json:"passed"`
			Error  string `json:"error"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(raw, &summary); err != nil {
		t.Fatalf("decode failure summary: %v\n%s", err, string(raw))
	}
	if summary.ExitCode != 1 || summary.CheckCount != 4 || summary.FailedCount != 1 {
		t.Fatalf("unexpected failure summary: %+v", summary)
	}

	var runtimeCheckError string
	for _, check := range summary.Checks {
		if check.Name == "runtime API" {
			if check.Passed {
				t.Fatalf("runtime API check unexpectedly passed: %+v", check)
			}
			runtimeCheckError = check.Error
			break
		}
	}
	if !strings.Contains(runtimeCheckError, "not valid JSON") {
		t.Fatalf("runtime API failure was not diagnostic: %q", runtimeCheckError)
	}
}

func TestApiContractSmokeScriptRequiresE2EHooksWhenRequested(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell harness validation is exercised on the Windows QA lane")
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/healthz":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		case "/api/runtime":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"mode":"test","version":"mock"}`))
		case "/api/openapi.yaml":
			w.Header().Set("Content-Type", "application/yaml")
			_, _ = w.Write([]byte("openapi: 3.0.0\ninfo:\n  title: Cabinet Mock\n"))
		case "/sign-in":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<html><body>Sign in</body></html>"))
		case "/api/test/reset":
			if r.Method != http.MethodPost {
				t.Fatalf("reset hook used method %s, want POST", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ok":true}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	repoRoot := filepath.Dir(currentFileDir(t))
	logRoot := t.TempDir()
	runID := "go-api-contract-smoke-e2e-hooks"
	cmd := exec.Command("pwsh", "-NoLogo", "-NoProfile", "-File", filepath.Join(repoRoot, "scripts", "run-api-contract-smoke.ps1"), "-BaseUrl", srv.URL, "-LogRoot", logRoot, "-RunId", runID, "-RequireE2EHooks")
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("api contract smoke script with E2E hooks failed: %v\n%s", err, string(output))
	}

	summaryPath := filepath.Join(logRoot, runID, "api-contract-smoke.summary.json")
	raw, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("read E2E hooks summary: %v", err)
	}

	var summary struct {
		ExitCode        int  `json:"exit_code"`
		CheckCount      int  `json:"check_count"`
		FailedCount     int  `json:"failed_count"`
		RequireE2EHooks bool `json:"require_e2e_hooks"`
		Checks          []struct {
			Name   string `json:"name"`
			Method string `json:"method"`
			Path   string `json:"path"`
			Passed bool   `json:"passed"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(raw, &summary); err != nil {
		t.Fatalf("decode E2E hooks summary: %v\n%s", err, string(raw))
	}
	if summary.ExitCode != 0 || summary.CheckCount != 5 || summary.FailedCount != 0 || !summary.RequireE2EHooks {
		t.Fatalf("unexpected E2E hooks summary: %+v", summary)
	}

	var sawResetHook bool
	for _, check := range summary.Checks {
		if check.Name == "E2E reset hook" {
			sawResetHook = true
			if check.Method != http.MethodPost || check.Path != "/api/test/reset" || !check.Passed {
				t.Fatalf("unexpected reset hook check: %+v", check)
			}
			break
		}
	}
	if !sawResetHook {
		t.Fatalf("summary did not include E2E reset hook check: %+v", summary.Checks)
	}
}

func TestCypressHarnessCanRunApiContractSmokeBeforeBrowserSpec(t *testing.T) {
	repoRoot := filepath.Dir(currentFileDir(t))
	cypressPath := filepath.Join(repoRoot, "cypress.ps1")
	raw, err := os.ReadFile(cypressPath)
	if err != nil {
		t.Fatalf("read cypress.ps1: %v", err)
	}
	content := string(raw)
	for _, snippet := range []string{
		"[switch]$ApiContractSmoke",
		"scripts\\run-api-contract-smoke.ps1",
		"Running API contract smoke preflight.",
		"API contract smoke summary:",
		"api_contract_smoke_summary_path",
		"API contract smoke preflight failed",
	} {
		if !strings.Contains(content, snippet) {
			t.Fatalf("cypress.ps1 missing API smoke preflight snippet %q", snippet)
		}
	}
}

func currentFileDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(filename)
}

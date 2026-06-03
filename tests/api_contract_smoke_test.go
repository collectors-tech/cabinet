package tests

import (
	"encoding/json"
	"fmt"
	"net"
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
		ExitCode     int    `json:"exit_code"`
		CheckCount   int    `json:"check_count"`
		FailedCount  int    `json:"failed_count"`
		SourceCommit string `json:"source_commit"`
	}
	if err := json.Unmarshal(raw, &summary); err != nil {
		t.Fatalf("decode summary: %v\n%s", err, string(raw))
	}
	if summary.ExitCode != 0 || summary.CheckCount != 4 || summary.FailedCount != 0 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	expectedCommit := currentGitCommit(t, repoRoot)
	if summary.SourceCommit != expectedCommit {
		t.Fatalf("summary source_commit = %q, want %q", summary.SourceCommit, expectedCommit)
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

func TestApiContractSmokeScriptDiagnosesHealthzFailure(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell harness validation is exercised on the Windows QA lane")
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/healthz":
			http.Error(w, "stale runtime", http.StatusServiceUnavailable)
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
	runID := "go-api-contract-smoke-healthz-failure"
	cmd := exec.Command("pwsh", "-NoLogo", "-NoProfile", "-File", filepath.Join(repoRoot, "scripts", "run-api-contract-smoke.ps1"), "-BaseUrl", srv.URL, "-LogRoot", logRoot, "-RunId", runID)
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("api contract smoke script unexpectedly passed\n%s", string(output))
	}

	summaryPath := filepath.Join(logRoot, runID, "api-contract-smoke.summary.json")
	raw, readErr := os.ReadFile(summaryPath)
	if readErr != nil {
		t.Fatalf("read healthz failure summary: %v\n%s", readErr, string(output))
	}

	var summary struct {
		ExitCode    int `json:"exit_code"`
		CheckCount  int `json:"check_count"`
		FailedCount int `json:"failed_count"`
		Checks      []struct {
			Name   string `json:"name"`
			Status int    `json:"status"`
			Passed bool   `json:"passed"`
			Error  string `json:"error"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(raw, &summary); err != nil {
		t.Fatalf("decode healthz failure summary: %v\n%s", err, string(raw))
	}
	if summary.ExitCode != 1 || summary.CheckCount != 4 || summary.FailedCount != 1 {
		t.Fatalf("unexpected healthz failure summary: %+v", summary)
	}

	var healthzCheck struct {
		Name   string `json:"name"`
		Status int    `json:"status"`
		Passed bool   `json:"passed"`
		Error  string `json:"error"`
	}
	for _, check := range summary.Checks {
		if check.Name == "healthz" {
			healthzCheck = check
			break
		}
	}
	if healthzCheck.Name == "" {
		t.Fatalf("summary did not include healthz check: %+v", summary.Checks)
	}
	if healthzCheck.Passed || healthzCheck.Status != http.StatusServiceUnavailable {
		t.Fatalf("healthz check did not record stale runtime status: %+v", healthzCheck)
	}
	if !strings.Contains(healthzCheck.Error, "503") {
		t.Fatalf("healthz failure was not diagnostic: %q", healthzCheck.Error)
	}
}

func TestApiContractSmokeScriptDiagnosesDeadRuntime(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell harness validation is exercised on the Windows QA lane")
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve dead runtime port: %v", err)
	}
	deadBaseURL := fmt.Sprintf("http://%s", listener.Addr().String())
	if err := listener.Close(); err != nil {
		t.Fatalf("release dead runtime port: %v", err)
	}

	repoRoot := filepath.Dir(currentFileDir(t))
	logRoot := t.TempDir()
	runID := "go-api-contract-smoke-dead-runtime"
	cmd := exec.Command("pwsh", "-NoLogo", "-NoProfile", "-File", filepath.Join(repoRoot, "scripts", "run-api-contract-smoke.ps1"), "-BaseUrl", deadBaseURL, "-LogRoot", logRoot, "-RunId", runID, "-TimeoutSec", "1")
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("api contract smoke script unexpectedly passed against dead runtime\n%s", string(output))
	}

	summaryPath := filepath.Join(logRoot, runID, "api-contract-smoke.summary.json")
	raw, readErr := os.ReadFile(summaryPath)
	if readErr != nil {
		t.Fatalf("read dead runtime summary: %v\n%s", readErr, string(output))
	}

	var summary struct {
		ExitCode    int `json:"exit_code"`
		CheckCount  int `json:"check_count"`
		FailedCount int `json:"failed_count"`
		Checks      []struct {
			Name   string `json:"name"`
			Status int    `json:"status"`
			Passed bool   `json:"passed"`
			Error  string `json:"error"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(raw, &summary); err != nil {
		t.Fatalf("decode dead runtime summary: %v\n%s", err, string(raw))
	}
	if summary.ExitCode != 1 || summary.CheckCount != 4 || summary.FailedCount != 4 {
		t.Fatalf("unexpected dead runtime summary: %+v", summary)
	}

	var healthzCheck struct {
		Name   string `json:"name"`
		Status int    `json:"status"`
		Passed bool   `json:"passed"`
		Error  string `json:"error"`
	}
	for _, check := range summary.Checks {
		if check.Name == "healthz" {
			healthzCheck = check
			break
		}
	}
	if healthzCheck.Name == "" {
		t.Fatalf("summary did not include healthz check: %+v", summary.Checks)
	}
	if healthzCheck.Passed || healthzCheck.Status != 0 {
		t.Fatalf("healthz check did not record dead runtime failure: %+v", healthzCheck)
	}
	lowerError := strings.ToLower(healthzCheck.Error)
	if !strings.Contains(lowerError, "connection") && !strings.Contains(lowerError, "refused") && !strings.Contains(lowerError, "timeout") {
		t.Fatalf("dead runtime failure was not diagnostic: %q", healthzCheck.Error)
	}
}

func TestApiContractSmokeScriptDiagnosesSignInContentMismatch(t *testing.T) {
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
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"error":"wrong route"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	repoRoot := filepath.Dir(currentFileDir(t))
	logRoot := t.TempDir()
	runID := "go-api-contract-smoke-sign-in-content-failure"
	cmd := exec.Command("pwsh", "-NoLogo", "-NoProfile", "-File", filepath.Join(repoRoot, "scripts", "run-api-contract-smoke.ps1"), "-BaseUrl", srv.URL, "-LogRoot", logRoot, "-RunId", runID)
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("api contract smoke script unexpectedly passed\n%s", string(output))
	}

	summaryPath := filepath.Join(logRoot, runID, "api-contract-smoke.summary.json")
	raw, readErr := os.ReadFile(summaryPath)
	if readErr != nil {
		t.Fatalf("read sign-in content failure summary: %v\n%s", readErr, string(output))
	}

	var summary struct {
		ExitCode    int `json:"exit_code"`
		CheckCount  int `json:"check_count"`
		FailedCount int `json:"failed_count"`
		Checks      []struct {
			Name   string `json:"name"`
			Status int    `json:"status"`
			Passed bool   `json:"passed"`
			Error  string `json:"error"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(raw, &summary); err != nil {
		t.Fatalf("decode sign-in content failure summary: %v\n%s", err, string(raw))
	}
	if summary.ExitCode != 1 || summary.CheckCount != 4 || summary.FailedCount != 1 {
		t.Fatalf("unexpected sign-in content failure summary: %+v", summary)
	}

	var signInCheck struct {
		Name   string `json:"name"`
		Status int    `json:"status"`
		Passed bool   `json:"passed"`
		Error  string `json:"error"`
	}
	for _, check := range summary.Checks {
		if check.Name == "sign-in route" {
			signInCheck = check
			break
		}
	}
	if signInCheck.Name == "" {
		t.Fatalf("summary did not include sign-in route check: %+v", summary.Checks)
	}
	if signInCheck.Passed || signInCheck.Status != http.StatusOK {
		t.Fatalf("sign-in check did not record route content mismatch: %+v", signInCheck)
	}
	if !strings.Contains(signInCheck.Error, "required marker") {
		t.Fatalf("sign-in content failure was not diagnostic: %q", signInCheck.Error)
	}
}

func TestApiContractSmokeScriptDiagnosesOpenAPIContentTypeMismatch(t *testing.T) {
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
			w.Header().Set("Content-Type", "text/plain")
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
	runID := "go-api-contract-smoke-openapi-content-type-failure"
	cmd := exec.Command("pwsh", "-NoLogo", "-NoProfile", "-File", filepath.Join(repoRoot, "scripts", "run-api-contract-smoke.ps1"), "-BaseUrl", srv.URL, "-LogRoot", logRoot, "-RunId", runID)
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("api contract smoke script unexpectedly passed\n%s", string(output))
	}

	summaryPath := filepath.Join(logRoot, runID, "api-contract-smoke.summary.json")
	raw, readErr := os.ReadFile(summaryPath)
	if readErr != nil {
		t.Fatalf("read OpenAPI content-type failure summary: %v\n%s", readErr, string(output))
	}

	var summary struct {
		ExitCode    int `json:"exit_code"`
		CheckCount  int `json:"check_count"`
		FailedCount int `json:"failed_count"`
		Checks      []struct {
			Name   string `json:"name"`
			Status int    `json:"status"`
			Passed bool   `json:"passed"`
			Error  string `json:"error"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(raw, &summary); err != nil {
		t.Fatalf("decode OpenAPI content-type failure summary: %v\n%s", err, string(raw))
	}
	if summary.ExitCode != 1 || summary.CheckCount != 4 || summary.FailedCount != 1 {
		t.Fatalf("unexpected OpenAPI content-type failure summary: %+v", summary)
	}

	var openAPICheck struct {
		Name   string `json:"name"`
		Status int    `json:"status"`
		Passed bool   `json:"passed"`
		Error  string `json:"error"`
	}
	for _, check := range summary.Checks {
		if check.Name == "OpenAPI YAML" {
			openAPICheck = check
			break
		}
	}
	if openAPICheck.Name == "" {
		t.Fatalf("summary did not include OpenAPI YAML check: %+v", summary.Checks)
	}
	if openAPICheck.Passed || openAPICheck.Status != http.StatusOK {
		t.Fatalf("OpenAPI check did not record content-type mismatch: %+v", openAPICheck)
	}
	if !strings.Contains(openAPICheck.Error, "content type") {
		t.Fatalf("OpenAPI content-type failure was not diagnostic: %q", openAPICheck.Error)
	}
}

func TestApiContractSmokeScriptDiagnosesOpenAPIBodyMismatch(t *testing.T) {
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
			_, _ = w.Write([]byte("info:\n  title: Wrong Runtime\n"))
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
	runID := "go-api-contract-smoke-openapi-body-failure"
	cmd := exec.Command("pwsh", "-NoLogo", "-NoProfile", "-File", filepath.Join(repoRoot, "scripts", "run-api-contract-smoke.ps1"), "-BaseUrl", srv.URL, "-LogRoot", logRoot, "-RunId", runID)
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("api contract smoke script unexpectedly passed\n%s", string(output))
	}

	summaryPath := filepath.Join(logRoot, runID, "api-contract-smoke.summary.json")
	raw, readErr := os.ReadFile(summaryPath)
	if readErr != nil {
		t.Fatalf("read OpenAPI body failure summary: %v\n%s", readErr, string(output))
	}

	var summary struct {
		ExitCode    int `json:"exit_code"`
		CheckCount  int `json:"check_count"`
		FailedCount int `json:"failed_count"`
		Checks      []struct {
			Name   string `json:"name"`
			Status int    `json:"status"`
			Passed bool   `json:"passed"`
			Error  string `json:"error"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(raw, &summary); err != nil {
		t.Fatalf("decode OpenAPI body failure summary: %v\n%s", err, string(raw))
	}
	if summary.ExitCode != 1 || summary.CheckCount != 4 || summary.FailedCount != 1 {
		t.Fatalf("unexpected OpenAPI body failure summary: %+v", summary)
	}

	var openAPICheck struct {
		Name   string `json:"name"`
		Status int    `json:"status"`
		Passed bool   `json:"passed"`
		Error  string `json:"error"`
	}
	for _, check := range summary.Checks {
		if check.Name == "OpenAPI YAML" {
			openAPICheck = check
			break
		}
	}
	if openAPICheck.Name == "" {
		t.Fatalf("summary did not include OpenAPI YAML check: %+v", summary.Checks)
	}
	if openAPICheck.Passed || openAPICheck.Status != http.StatusOK {
		t.Fatalf("OpenAPI check did not record body mismatch: %+v", openAPICheck)
	}
	if !strings.Contains(openAPICheck.Error, "required marker") {
		t.Fatalf("OpenAPI body failure was not diagnostic: %q", openAPICheck.Error)
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

func TestApiContractSmokeScriptDiagnosesMissingE2EHooksWhenRequired(t *testing.T) {
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
	runID := "go-api-contract-smoke-missing-e2e-hooks"
	cmd := exec.Command("pwsh", "-NoLogo", "-NoProfile", "-File", filepath.Join(repoRoot, "scripts", "run-api-contract-smoke.ps1"), "-BaseUrl", srv.URL, "-LogRoot", logRoot, "-RunId", runID, "-RequireE2EHooks")
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("api contract smoke script unexpectedly passed without E2E hooks\n%s", string(output))
	}

	summaryPath := filepath.Join(logRoot, runID, "api-contract-smoke.summary.json")
	raw, readErr := os.ReadFile(summaryPath)
	if readErr != nil {
		t.Fatalf("read missing E2E hooks summary: %v\n%s", readErr, string(output))
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
			Status int    `json:"status"`
			Passed bool   `json:"passed"`
			Error  string `json:"error"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(raw, &summary); err != nil {
		t.Fatalf("decode missing E2E hooks summary: %v\n%s", err, string(raw))
	}
	if summary.ExitCode != 1 || summary.CheckCount != 5 || summary.FailedCount != 1 || !summary.RequireE2EHooks {
		t.Fatalf("unexpected missing E2E hooks summary: %+v", summary)
	}

	var resetHookCheck struct {
		Name   string `json:"name"`
		Method string `json:"method"`
		Path   string `json:"path"`
		Status int    `json:"status"`
		Passed bool   `json:"passed"`
		Error  string `json:"error"`
	}
	for _, check := range summary.Checks {
		if check.Name == "E2E reset hook" {
			resetHookCheck = check
			break
		}
	}
	if resetHookCheck.Name == "" {
		t.Fatalf("summary did not include E2E reset hook check: %+v", summary.Checks)
	}
	if resetHookCheck.Passed || resetHookCheck.Method != http.MethodPost || resetHookCheck.Path != "/api/test/reset" || resetHookCheck.Status != http.StatusNotFound {
		t.Fatalf("reset hook check did not record missing hook failure: %+v", resetHookCheck)
	}
	if !strings.Contains(resetHookCheck.Error, "404") {
		t.Fatalf("missing E2E hook failure was not diagnostic: %q", resetHookCheck.Error)
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

func currentGitCommit(t *testing.T, repoRoot string) string {
	t.Helper()
	cmd := exec.Command("git", "-C", repoRoot, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("git rev-parse HEAD failed: %v", err)
	}
	return strings.TrimSpace(string(output))
}

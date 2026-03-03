package app

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeSetupStatusAndCompleteContract(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	setupPath := runtimeSetupConfigPath(a.cfg)
	_ = os.Remove(setupPath)

	statusMissing := doRequest(t, a, http.MethodGet, "/api/runtime/setup-status", nil, nil)
	if statusMissing.Code != http.StatusOK {
		t.Fatalf("setup-status missing expected 200, got %d body=%s", statusMissing.Code, statusMissing.Body.String())
	}
	var missingPayload map[string]any
	if err := json.Unmarshal(statusMissing.Body.Bytes(), &missingPayload); err != nil {
		t.Fatalf("decode missing payload: %v", err)
	}
	if missingPayload["setup_required"] != true {
		t.Fatalf("expected setup_required=true when config missing, got %v", missingPayload["setup_required"])
	}

	completeReq := map[string]any{
		"instance_name":          "Primary",
		"profile_key":            "primary",
		"auth_mode":              "local",
		"runtime_port_mode":      "auto",
		"bootstrap_workspace":    "Local Workspace",
		"bootstrap_database_ref": "Primary DB",
	}
	completeReqJSON, err := json.Marshal(completeReq)
	if err != nil {
		t.Fatalf("marshal complete request: %v", err)
	}
	complete := doRequest(t, a, http.MethodPost, "/api/runtime/setup-complete", strings.NewReader(string(completeReqJSON)), map[string]string{"Content-Type": "application/json"})
	if complete.Code != http.StatusOK {
		t.Fatalf("setup-complete expected 200, got %d body=%s", complete.Code, complete.Body.String())
	}
	if _, err := os.Stat(setupPath); err != nil {
		t.Fatalf("expected setup config at %s: %v", setupPath, err)
	}
	rawConfig, err := os.ReadFile(setupPath)
	if err != nil {
		t.Fatalf("read setup config: %v", err)
	}
	var configPayload map[string]any
	if err := json.Unmarshal(rawConfig, &configPayload); err != nil {
		t.Fatalf("decode setup config: %v", err)
	}
	requiredSections := []string{"instance", "storage", "runtime", "auth", "bootstrap", "features", "meta"}
	for _, section := range requiredSections {
		if _, ok := configPayload[section]; !ok {
			t.Fatalf("expected section %s in setup config", section)
		}
	}
	authPayload, ok := configPayload["auth"].(map[string]any)
	if !ok {
		t.Fatalf("expected auth object in setup config")
	}
	if strings.TrimSpace(authPayload["mode"].(string)) != "local" {
		t.Fatalf("expected local auth mode in setup config")
	}

	statusPresent := doRequest(t, a, http.MethodGet, "/api/runtime/setup-status", nil, nil)
	if statusPresent.Code != http.StatusOK {
		t.Fatalf("setup-status present expected 200, got %d body=%s", statusPresent.Code, statusPresent.Body.String())
	}
	var presentPayload map[string]any
	if err := json.Unmarshal(statusPresent.Body.Bytes(), &presentPayload); err != nil {
		t.Fatalf("decode present payload: %v", err)
	}
	if presentPayload["setup_required"] != false {
		t.Fatalf("expected setup_required=false when config exists, got %v", presentPayload["setup_required"])
	}
}

func TestRuntimeSetupStatusPathUsesDataDirCabinetJSON(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	expected := filepath.Join(a.cfg.DataDir, "cabinet.json")
	if got := runtimeSetupConfigPath(a.cfg); got != expected {
		t.Fatalf("setup config path mismatch: got %s want %s", got, expected)
	}
}

func TestRuntimeSetupCompleteRequiresClerkPublishableKey(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	setupPath := runtimeSetupConfigPath(a.cfg)
	_ = os.Remove(setupPath)

	completeReq := map[string]any{
		"instance_name":          "Primary",
		"profile_key":            "primary",
		"auth_mode":              "clerk",
		"runtime_port_mode":      "auto",
		"bootstrap_workspace":    "Local Workspace",
		"bootstrap_database_ref": "Primary DB",
	}

	completeReqJSON, err := json.Marshal(completeReq)
	if err != nil {
		t.Fatalf("marshal complete request: %v", err)
	}
	resp := doRequest(t, a, http.MethodPost, "/api/runtime/setup-complete", strings.NewReader(string(completeReqJSON)), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing clerk key, got %d body=%s", resp.Code, resp.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode validation payload: %v", err)
	}
	if payload["error_code"] != "SETUP_CLERK_PUBLISHABLE_KEY_REQUIRED" {
		t.Fatalf("unexpected error code: %v", payload["error_code"])
	}
	if _, err := os.Stat(setupPath); !os.IsNotExist(err) {
		t.Fatalf("setup file should not be created on validation failure")
	}
}

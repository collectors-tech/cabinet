package app

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
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

	complete := doRequest(t, a, http.MethodPost, "/api/runtime/setup-complete", nil, nil)
	if complete.Code != http.StatusOK {
		t.Fatalf("setup-complete expected 200, got %d body=%s", complete.Code, complete.Body.String())
	}
	if _, err := os.Stat(setupPath); err != nil {
		t.Fatalf("expected setup config at %s: %v", setupPath, err)
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


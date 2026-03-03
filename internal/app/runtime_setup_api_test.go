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
	var completePayload map[string]any
	if err := json.Unmarshal(complete.Body.Bytes(), &completePayload); err != nil {
		t.Fatalf("decode complete payload: %v", err)
	}
	if completePayload["ok"] != true {
		t.Fatalf("expected ok=true in setup-complete payload")
	}
	if strings.TrimSpace(asString(completePayload["config_path"])) == "" {
		t.Fatalf("expected config_path in setup-complete payload")
	}
	if strings.TrimSpace(asString(completePayload["data_dir"])) == "" {
		t.Fatalf("expected data_dir in setup-complete payload")
	}
	if strings.TrimSpace(asString(completePayload["media_dir"])) == "" {
		t.Fatalf("expected media_dir in setup-complete payload")
	}
	runtimeURL := strings.TrimSpace(asString(completePayload["runtime_url"]))
	if runtimeURL == "" {
		t.Fatalf("expected runtime_url in setup-complete payload")
	}
	if !strings.HasPrefix(runtimeURL, "http://") {
		t.Fatalf("expected runtime_url to start with http://, got %s", runtimeURL)
	}
	if asFloat64(completePayload["runtime_port"]) <= 0 {
		t.Fatalf("expected runtime_port > 0 in setup-complete payload, got %v", completePayload["runtime_port"])
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
	metaPayload, ok := configPayload["meta"].(map[string]any)
	if !ok {
		t.Fatalf("expected meta object in setup config")
	}
	currentURL := strings.TrimSpace(asString(metaPayload["currentUrl"]))
	if currentURL == "" {
		t.Fatalf("expected meta.currentUrl in setup config")
	}
	if currentURL != runtimeURL {
		t.Fatalf("expected meta.currentUrl=%s, got %s", runtimeURL, currentURL)
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

func TestRuntimeSetupSyncCurrentURLUpdatesConfigMetadata(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	payload := runtimeSetupConfigFile{
		Version: 1,
		Instance: runtimeSetupInstanceConfig{
			Name:    "Primary",
			Profile: "primary",
		},
		Storage: runtimeSetupStorageConfig{
			DataDir:  filepath.Join(a.cfg.DataDir, "profiles", "primary"),
			MediaDir: filepath.Join(a.cfg.DataDir, "profiles", "primary", "media"),
		},
		Runtime: runtimeSetupRuntimeConfig{
			PortMode:    "auto",
			ResolvedURL: "http://127.0.0.1:17880",
		},
		Auth: runtimeSetupAuthConfig{
			Mode: "local",
			Clerk: runtimeSetupClerkAuthConfig{
				PublishableKey: "",
				Enabled:        false,
			},
		},
		Bootstrap: runtimeSetupBootstrapConfig{
			Workspace:       "Local Workspace",
			DatabaseProfile: "Primary DB",
		},
		Features: runtimeSetupFeaturesConfig{
			Chat:      true,
			Providers: true,
			Scanner:   true,
		},
		Meta: runtimeSetupMetaConfig{
			CreatedAt:     "2026-01-01T00:00:00Z",
			UpdatedAt:     "2026-01-01T00:00:00Z",
			WizardVersion: "1",
			CurrentURL:    "http://old-host:1000",
		},
	}
	if err := writeRuntimeSetupConfig(a.cfg, payload); err != nil {
		t.Fatalf("writeRuntimeSetupConfig() error = %v", err)
	}
	if err := syncRuntimeSetupCurrentURL(a.cfg); err != nil {
		t.Fatalf("syncRuntimeSetupCurrentURL() error = %v", err)
	}
	raw, err := os.ReadFile(runtimeSetupConfigPath(a.cfg))
	if err != nil {
		t.Fatalf("read setup config: %v", err)
	}
	var updated runtimeSetupConfigFile
	if err := json.Unmarshal(raw, &updated); err != nil {
		t.Fatalf("decode setup config: %v", err)
	}
	expected := payload.Runtime.ResolvedURL
	if strings.TrimSpace(updated.Meta.CurrentURL) != expected {
		t.Fatalf("expected meta.currentUrl=%s, got %s", expected, strings.TrimSpace(updated.Meta.CurrentURL))
	}
}

func TestRuntimePIDFileContainsPIDOnly(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	pidPath := runtimePIDPath(a.cfg)
	if err := writeRuntimePIDFile(a.cfg, 4242); err != nil {
		t.Fatalf("writeRuntimePIDFile() error = %v", err)
	}
	raw, err := os.ReadFile(pidPath)
	if err != nil {
		t.Fatalf("read pid file: %v", err)
	}
	content := strings.TrimSpace(string(raw))
	if content != "4242" {
		t.Fatalf("expected pid-only file content 4242, got %q", content)
	}
	if strings.Contains(content, "{") || strings.Contains(content, ":") {
		t.Fatalf("pid file should contain pid only, got %q", content)
	}
	if err := removeRuntimePIDFile(a.cfg); err != nil {
		t.Fatalf("removeRuntimePIDFile() error = %v", err)
	}
	if _, err := os.Stat(pidPath); !os.IsNotExist(err) {
		t.Fatalf("expected pid file removed, stat err=%v", err)
	}
}

func asString(value any) string {
	if value == nil {
		return ""
	}
	if v, ok := value.(string); ok {
		return v
	}
	return ""
}

func asFloat64(value any) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0
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

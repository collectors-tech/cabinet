package app

import (
	"encoding/json"
	"fmt"
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
	if _, ok := completePayload["local_login_username"]; ok {
		t.Fatalf("setup completion must not expose a simulated local username: %+v", completePayload)
	}
	if _, ok := completePayload["local_login_password"]; ok {
		t.Fatalf("setup completion must not expose a simulated local password: %+v", completePayload)
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
			Mode:    "local",
			Zitadel: runtimeSetupZitadelAuthConfig{},
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

func TestRuntimeSetupCompleteRejectsUnsupportedAuthMode(t *testing.T) {
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
		t.Fatalf("expected 400 for unsupported auth mode, got %d body=%s", resp.Code, resp.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode validation payload: %v", err)
	}
	if payload["error_code"] != "SETUP_AUTH_MODE_INVALID" {
		t.Fatalf("unexpected error code: %v", payload["error_code"])
	}
	if _, err := os.Stat(setupPath); !os.IsNotExist(err) {
		t.Fatalf("setup file should not be created on validation failure")
	}
}

func TestRuntimeSetupImportExistingConfigContract(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	setupPath := runtimeSetupConfigPath(a.cfg)
	_ = os.Remove(setupPath)

	payload, err := buildRuntimeSetupConfig(a.cfg, runtimeSetupRequest{
		InstanceName: "Imported Instance",
		ProfileKey:   "imported-profile",
		AuthMode:     "local",
	})
	if err != nil {
		t.Fatalf("buildRuntimeSetupConfig() error = %v", err)
	}
	sourcePath := filepath.Join(a.cfg.DataDir, "import-source.json")
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatalf("marshal import payload: %v", err)
	}
	raw = append(raw, '\n')
	if err := os.WriteFile(sourcePath, raw, 0o644); err != nil {
		t.Fatalf("write import source: %v", err)
	}

	resp := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/runtime/setup-import",
		strings.NewReader(fmt.Sprintf(`{"source_path":%q}`, sourcePath)),
		map[string]string{"Content-Type": "application/json"},
	)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected setup-import 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	var importPayload map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &importPayload); err != nil {
		t.Fatalf("decode import payload: %v", err)
	}
	if importPayload["ok"] != true {
		t.Fatalf("expected ok=true from setup-import payload")
	}
	if importPayload["setup_required"] != false {
		t.Fatalf("expected setup_required=false from setup-import payload")
	}
	if strings.TrimSpace(asString(importPayload["config_path"])) == "" {
		t.Fatalf("expected config_path in setup-import payload")
	}
	status := doRequest(t, a, http.MethodGet, "/api/runtime/setup-status", nil, nil)
	if status.Code != http.StatusOK {
		t.Fatalf("setup-status after import expected 200, got %d body=%s", status.Code, status.Body.String())
	}
	var statusPayload map[string]any
	if err := json.Unmarshal(status.Body.Bytes(), &statusPayload); err != nil {
		t.Fatalf("decode setup-status payload: %v", err)
	}
	if statusPayload["setup_required"] != false {
		t.Fatalf("expected setup_required=false after successful import, got %v", statusPayload["setup_required"])
	}

	invalid := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/runtime/setup-import",
		strings.NewReader(`{"source_path":""}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("expected setup-import 400 for missing source path, got %d body=%s", invalid.Code, invalid.Body.String())
	}
	var invalidPayload map[string]any
	if err := json.Unmarshal(invalid.Body.Bytes(), &invalidPayload); err != nil {
		t.Fatalf("decode invalid setup-import payload: %v", err)
	}
	if invalidPayload["error_code"] != "SETUP_IMPORT_SOURCE_PATH_REQUIRED" {
		t.Fatalf("unexpected setup-import error code: %v", invalidPayload["error_code"])
	}
}

func TestRuntimeSetupImportRejectsUnsupportedAuthSubkey(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	setupPath := runtimeSetupConfigPath(a.cfg)
	_ = os.Remove(setupPath)

	sourcePath := filepath.Join(a.cfg.DataDir, "clerk-shaped-import.json")
	raw := []byte(`{
  "version": 1,
  "instance": {
    "name": "Clerk Shaped",
    "profile": "clerk-shaped"
  },
  "storage": {
    "dataDir": "C:/cabinet/data",
    "mediaDir": "C:/cabinet/data/media",
    "portableMode": false
  },
  "runtime": {
    "portMode": "auto",
    "port": null,
    "resolvedUrl": "http://127.0.0.1:17880"
  },
  "auth": {
    "mode": "zitadel",
    "clerk": {
      "enabled": false,
      "publishableKey": "pk_test_retired"
    }
  },
  "bootstrap": {
    "workspace": "Local Workspace",
    "databaseProfile": "Primary DB"
  },
  "features": {
    "chat": true,
    "providers": true,
    "scanner": true
  },
  "meta": {
    "createdAt": "2026-01-01T00:00:00Z",
    "updatedAt": "2026-01-01T00:00:00Z",
    "wizardVersion": "1",
    "currentUrl": "http://127.0.0.1:17880"
  }
}`)
	if err := os.WriteFile(sourcePath, raw, 0o644); err != nil {
		t.Fatalf("write clerk-shaped import source: %v", err)
	}

	resp := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/runtime/setup-import",
		strings.NewReader(fmt.Sprintf(`{"source_path":%q}`, sourcePath)),
		map[string]string{"Content-Type": "application/json"},
	)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected setup-import 400 for unsupported auth config subkey, got %d body=%s", resp.Code, resp.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode unsupported auth config import payload: %v", err)
	}
	if payload["error_code"] != "SETUP_IMPORT_FAILED" {
		t.Fatalf("unexpected setup-import error code: %v", payload["error_code"])
	}
	if !strings.Contains(strings.ToLower(asString(payload["message"])), "auth.clerk") {
		t.Fatalf("expected auth.clerk rejection message, got %v", payload["message"])
	}
	if _, err := os.Stat(setupPath); !os.IsNotExist(err) {
		t.Fatalf("setup file should not be created from Clerk-shaped import")
	}
}

func TestRuntimeSetupCompleteDerivesProfileKeyWhenBlank(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	setupPath := runtimeSetupConfigPath(a.cfg)
	_ = os.Remove(setupPath)

	completeReq := map[string]any{
		"instance_name":          "My Fancy Instance",
		"profile_key":            "",
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
	rawConfig, err := os.ReadFile(setupPath)
	if err != nil {
		t.Fatalf("read setup config: %v", err)
	}
	var configPayload map[string]any
	if err := json.Unmarshal(rawConfig, &configPayload); err != nil {
		t.Fatalf("decode setup config: %v", err)
	}
	instance, ok := configPayload["instance"].(map[string]any)
	if !ok {
		t.Fatalf("expected instance object in setup config")
	}
	if strings.TrimSpace(asString(instance["profile"])) != "my-fancy-instance" {
		t.Fatalf("expected derived profile key my-fancy-instance, got %v", instance["profile"])
	}
}

func TestRuntimeSetupStorageValidateContract(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	missing := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/runtime/setup-storage-validate",
		strings.NewReader(`{"data_dir":""}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if missing.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing storage path, got %d body=%s", missing.Code, missing.Body.String())
	}
	var missingPayload map[string]any
	if err := json.Unmarshal(missing.Body.Bytes(), &missingPayload); err != nil {
		t.Fatalf("decode missing storage payload: %v", err)
	}
	if missingPayload["error_code"] != "SETUP_STORAGE_PATH_REQUIRED" {
		t.Fatalf("unexpected storage error code: %v", missingPayload["error_code"])
	}

	validDir := filepath.Join(a.cfg.DataDir, "custom-storage")
	valid := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/runtime/setup-storage-validate",
		strings.NewReader(fmt.Sprintf(`{"data_dir":%q}`, validDir)),
		map[string]string{"Content-Type": "application/json"},
	)
	if valid.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid storage path, got %d body=%s", valid.Code, valid.Body.String())
	}
	var validPayload map[string]any
	if err := json.Unmarshal(valid.Body.Bytes(), &validPayload); err != nil {
		t.Fatalf("decode valid storage payload: %v", err)
	}
	if validPayload["writable"] != true {
		t.Fatalf("expected writable=true, got %v", validPayload["writable"])
	}
	if strings.TrimSpace(asString(validPayload["free_space_status"])) == "" {
		t.Fatalf("expected free_space_status in storage validation payload")
	}
}

func TestRuntimeSetupCompletePersistsSelectedStoragePath(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	setupPath := runtimeSetupConfigPath(a.cfg)
	_ = os.Remove(setupPath)

	customDir := filepath.Join(a.cfg.DataDir, "persisted-storage")
	completeReq := map[string]any{
		"instance_name":          "Storage Persist",
		"profile_key":            "",
		"storage_mode":           "custom",
		"storage_data_dir":       customDir,
		"portable_mode":          true,
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
	rawConfig, err := os.ReadFile(setupPath)
	if err != nil {
		t.Fatalf("read setup config: %v", err)
	}
	var configPayload map[string]any
	if err := json.Unmarshal(rawConfig, &configPayload); err != nil {
		t.Fatalf("decode setup config: %v", err)
	}
	storage, ok := configPayload["storage"].(map[string]any)
	if !ok {
		t.Fatalf("expected storage object in setup config")
	}
	if strings.TrimSpace(asString(storage["dataDir"])) != customDir {
		t.Fatalf("expected storage.dataDir=%s, got %v", customDir, storage["dataDir"])
	}
	if strings.TrimSpace(asString(storage["mediaDir"])) != filepath.Join(customDir, "media") {
		t.Fatalf("expected storage.mediaDir=%s, got %v", filepath.Join(customDir, "media"), storage["mediaDir"])
	}
}

func TestRuntimeSetupCompletePersistsFixedPortRuntime(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	setupPath := runtimeSetupConfigPath(a.cfg)
	_ = os.Remove(setupPath)

	completeReq := map[string]any{
		"instance_name":          "Runtime Persist",
		"profile_key":            "",
		"storage_mode":           "exe_local",
		"auth_mode":              "local",
		"runtime_port_mode":      "fixed",
		"runtime_fixed_port":     18999,
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
	rawConfig, err := os.ReadFile(setupPath)
	if err != nil {
		t.Fatalf("read setup config: %v", err)
	}
	var configPayload map[string]any
	if err := json.Unmarshal(rawConfig, &configPayload); err != nil {
		t.Fatalf("decode setup config: %v", err)
	}
	runtimePayload, ok := configPayload["runtime"].(map[string]any)
	if !ok {
		t.Fatalf("expected runtime object in setup config")
	}
	if strings.TrimSpace(asString(runtimePayload["portMode"])) != "fixed" {
		t.Fatalf("expected runtime.portMode=fixed, got %v", runtimePayload["portMode"])
	}
	if asFloat64(runtimePayload["port"]) != 18999 {
		t.Fatalf("expected runtime.port=18999, got %v", runtimePayload["port"])
	}
	if !strings.HasSuffix(strings.TrimSpace(asString(runtimePayload["resolvedUrl"])), ":18999") {
		t.Fatalf("expected runtime.resolvedUrl to end with :18999, got %v", runtimePayload["resolvedUrl"])
	}
}

func TestRuntimeSetupCompletePersistsZitadelAuthConfiguration(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	setupPath := runtimeSetupConfigPath(a.cfg)
	_ = os.Remove(setupPath)

	completeReq := map[string]any{
		"instance_name":          "ZITADEL Persist",
		"profile_key":            "",
		"storage_mode":           "exe_local",
		"auth_mode":              "zitadel",
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
	rawConfig, err := os.ReadFile(setupPath)
	if err != nil {
		t.Fatalf("read setup config: %v", err)
	}
	var configPayload map[string]any
	if err := json.Unmarshal(rawConfig, &configPayload); err != nil {
		t.Fatalf("decode setup config: %v", err)
	}
	authPayload, ok := configPayload["auth"].(map[string]any)
	if !ok {
		t.Fatalf("expected auth object in setup config")
	}
	if strings.TrimSpace(asString(authPayload["mode"])) != "zitadel" {
		t.Fatalf("expected auth.mode=zitadel, got %v", authPayload["mode"])
	}
	if _, ok := authPayload["clerk"]; ok {
		t.Fatalf("auth.clerk must not be persisted in setup config: %+v", authPayload["clerk"])
	}
	if _, ok := authPayload["zitadel"].(map[string]any); !ok {
		t.Fatalf("expected auth.zitadel object in setup config")
	}
}

func TestRuntimeSetupCompletePersistsFeatureToggles(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	setupPath := runtimeSetupConfigPath(a.cfg)
	_ = os.Remove(setupPath)

	completeReq := map[string]any{
		"instance_name":          "Feature Toggle Persist",
		"profile_key":            "",
		"storage_mode":           "exe_local",
		"auth_mode":              "local",
		"runtime_port_mode":      "auto",
		"feature_chat":           false,
		"feature_providers":      true,
		"feature_scanner":        false,
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
	rawConfig, err := os.ReadFile(setupPath)
	if err != nil {
		t.Fatalf("read setup config: %v", err)
	}
	var configPayload map[string]any
	if err := json.Unmarshal(rawConfig, &configPayload); err != nil {
		t.Fatalf("decode setup config: %v", err)
	}
	features, ok := configPayload["features"].(map[string]any)
	if !ok {
		t.Fatalf("expected features object in setup config")
	}
	if features["chat"] != false {
		t.Fatalf("expected features.chat=false, got %v", features["chat"])
	}
	if features["providers"] != true {
		t.Fatalf("expected features.providers=true, got %v", features["providers"])
	}
	if features["scanner"] != false {
		t.Fatalf("expected features.scanner=false, got %v", features["scanner"])
	}
}

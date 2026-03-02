package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/update"
)

func TestRuntimeLANModeExposesBindMetadataAndHealth(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Addr:           "0.0.0.0:17880",
		Host:           "0.0.0.0",
		Port:           17880,
		BindMode:       "lan",
		DataDir:        t.TempDir(),
		DBPath:         t.TempDir() + "/cabinet.db",
		UpdateChannel:  update.ChannelStable,
		WebAuthnRPID:   "127.0.0.1",
		WebAuthnOrigin: "http://127.0.0.1:17880",
		WebAuthnName:   "Cabinet Test",
		BackupInterval: 60,
	}
	a := newTestAppWithConfig(t, cfg)

	runtimeResp := doRequest(t, a, http.MethodGet, "/api/runtime", nil, nil)
	if runtimeResp.Code != http.StatusOK {
		t.Fatalf("runtime status=%d body=%s", runtimeResp.Code, runtimeResp.Body.String())
	}
	if !strings.Contains(runtimeResp.Body.String(), `"bind_mode":"lan"`) {
		t.Fatalf("expected bind_mode lan in runtime payload, got %s", runtimeResp.Body.String())
	}
	if !strings.Contains(runtimeResp.Body.String(), `"runtime_host":"0.0.0.0"`) {
		t.Fatalf("expected runtime_host 0.0.0.0, got %s", runtimeResp.Body.String())
	}

	healthResp := doRequest(t, a, http.MethodGet, "/healthz", nil, nil)
	if healthResp.Code != http.StatusOK {
		t.Fatalf("health status=%d body=%s", healthResp.Code, healthResp.Body.String())
	}
}

func TestRuntimeLANModeUnauthorizedProtectedEndpointDoesNotMutateProfileData(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Addr:           "0.0.0.0:17880",
		Host:           "0.0.0.0",
		Port:           17880,
		BindMode:       "lan",
		DataDir:        t.TempDir(),
		DBPath:         t.TempDir() + "/cabinet.db",
		UpdateChannel:  update.ChannelStable,
		WebAuthnRPID:   "127.0.0.1",
		WebAuthnOrigin: "http://127.0.0.1:17880",
		WebAuthnName:   "Cabinet Test",
		BackupInterval: 60,
	}
	a := newTestAppWithConfig(t, cfg)

	before := doRequest(t, a, http.MethodGet, "/api/profiles", nil, nil)
	if before.Code != http.StatusOK {
		t.Fatalf("profiles before status=%d body=%s", before.Code, before.Body.String())
	}

	unauthorized := doRequest(t, a, http.MethodPost, "/api/auth/session/validate", strings.NewReader(`{"session_token":"missing"}`), map[string]string{"Content-Type": "application/json"})
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthorized validate in LAN mode, got %d body=%s", unauthorized.Code, unauthorized.Body.String())
	}

	after := doRequest(t, a, http.MethodGet, "/api/profiles", nil, nil)
	if after.Code != http.StatusOK {
		t.Fatalf("profiles after status=%d body=%s", after.Code, after.Body.String())
	}

	var beforePayload struct {
		Profiles []map[string]any `json:"profiles"`
	}
	if err := json.Unmarshal(before.Body.Bytes(), &beforePayload); err != nil {
		t.Fatalf("decode profiles before: %v", err)
	}
	var afterPayload struct {
		Profiles []map[string]any `json:"profiles"`
	}
	if err := json.Unmarshal(after.Body.Bytes(), &afterPayload); err != nil {
		t.Fatalf("decode profiles after: %v", err)
	}
	if len(beforePayload.Profiles) != len(afterPayload.Profiles) {
		t.Fatalf("expected no profile mutation from unauthorized request, before=%d after=%d", len(beforePayload.Profiles), len(afterPayload.Profiles))
	}
}

package app

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/profile"
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

	unauthorizedRead := doRequest(t, a, http.MethodGet, "/api/profiles", nil, nil)
	if unauthorizedRead.Code != http.StatusUnauthorized && unauthorizedRead.Code != http.StatusForbidden {
		t.Fatalf("expected unauthorized LAN profile read to fail closed, got %d body=%s", unauthorizedRead.Code, unauthorizedRead.Body.String())
	}
	if strings.Contains(unauthorizedRead.Body.String(), `"profiles"`) {
		t.Fatalf("unauthorized LAN profile read exposed protected data: %s", unauthorizedRead.Body.String())
	}

	unauthorizedMutation := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"LAN-UNAUTH-001","title":"Must not persist"}`), map[string]string{"Content-Type": "application/json"})
	if unauthorizedMutation.Code != http.StatusUnauthorized && unauthorizedMutation.Code != http.StatusForbidden {
		t.Fatalf("expected unauthorized LAN item mutation to fail closed, got %d body=%s", unauthorizedMutation.Code, unauthorizedMutation.Body.String())
	}
	var itemCount int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM canonical_items WHERE part_number = 'LAN-UNAUTH-001'`).Scan(&itemCount); err != nil {
		t.Fatalf("count unauthorized LAN items: %v", err)
	}
	if itemCount != 0 {
		t.Fatalf("unauthorized LAN request persisted %d protected items", itemCount)
	}

	profiles := profile.NewRepository(a.db)
	registrationPending, err := profiles.Create(context.Background(), "LAN registration pending")
	if err != nil {
		t.Fatalf("seed LAN registration-pending profile: %v", err)
	}
	if err := profiles.SetActiveProfile(context.Background(), registrationPending.ID); err != nil {
		t.Fatalf("activate LAN registration-pending profile: %v", err)
	}
	registrationBypass := doRequest(t, a, http.MethodGet, "/api/profiles", nil, nil)
	if registrationBypass.Code != http.StatusUnauthorized && registrationBypass.Code != http.StatusForbidden {
		t.Fatalf("registration-pending LAN profile must not bypass auth, got %d body=%s", registrationBypass.Code, registrationBypass.Body.String())
	}
	companionManagement := doCompanionManagementRequest(t, a, http.MethodGet, "/api/companion/pairing/requests", nil, nil)
	if companionManagement.Code != http.StatusUnauthorized && companionManagement.Code != http.StatusForbidden {
		t.Fatalf("registration-pending LAN profile must not expose Companion management, got %d body=%s", companionManagement.Code, companionManagement.Body.String())
	}
}

func TestRuntimeLANModeValidUnlockedSessionCanReadAndMutateProtectedData(t *testing.T) {
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
	profiles := profile.NewRepository(a.db)
	active, err := profiles.Create(context.Background(), "LAN unlocked profile")
	if err != nil {
		t.Fatalf("seed LAN profile: %v", err)
	}
	if err := profiles.SetActiveProfile(context.Background(), active.ID); err != nil {
		t.Fatalf("activate LAN profile: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO webauthn_credentials(id, profile_id, credential_json) VALUES('cred-lan-unlocked', ?, '{}')`, active.ID); err != nil {
		t.Fatalf("seed LAN credential: %v", err)
	}
	token, err := a.authService.CreateUnlockedSession(active.ID)
	if err != nil {
		t.Fatalf("create LAN unlocked session: %v", err)
	}
	headers := map[string]string{"X-Cabinet-Session": token}
	read := doRequest(t, a, http.MethodGet, "/api/profiles", nil, headers)
	if read.Code != http.StatusOK {
		t.Fatalf("authorized LAN profile read status=%d body=%s", read.Code, read.Body.String())
	}
	headers["Content-Type"] = "application/json"
	mutation := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"LAN-AUTH-001","title":"Authorized LAN item"}`), headers)
	if mutation.Code != http.StatusCreated {
		t.Fatalf("authorized LAN item mutation status=%d body=%s", mutation.Code, mutation.Body.String())
	}
}

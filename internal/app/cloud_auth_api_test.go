package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/config"
)

func TestZitadelModeRetiresUnverifiedLegacyCloudBootstrap(t *testing.T) {
	t.Parallel()

	boundary := newZitadelAuthBoundary(zitadelAuthConfig{
		IdentityMode: "zitadel",
		Issuer:       "https://identity.example.test",
		ClientID:     "cabinet-client",
		Audience:     "cabinet-project",
		PublicOrigin: "https://cabinet.example.test",
	}, nil)
	recorder := httptest.NewRecorder()
	if !rejectLegacyCloudBootstrapInZitadel(recorder, boundary) {
		t.Fatal("ZITADEL mode must reject caller-supplied legacy entitlement claims")
	}
	if recorder.Code != http.StatusGone || !strings.Contains(recorder.Body.String(), "legacy_cloud_bootstrap_disabled") {
		t.Fatalf("retired ZITADEL bootstrap status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	localRecorder := httptest.NewRecorder()
	if rejectLegacyCloudBootstrapInZitadel(localRecorder, nil) {
		t.Fatal("local legacy compatibility path should remain explicit outside ZITADEL mode")
	}
}

func TestZitadelAndLANModesDisableCredentialFreeCompanionManagement(t *testing.T) {
	t.Parallel()

	boundary := newZitadelAuthBoundary(zitadelAuthConfig{
		IdentityMode: "zitadel",
		Issuer:       "https://identity.example.test",
		ClientID:     "cabinet-client",
		Audience:     "cabinet-project",
		PublicOrigin: "https://cabinet.example.test",
	}, nil)
	if credentialFreeLocalCompanionManagementAllowed(config.Config{BindMode: "local"}, boundary) {
		t.Fatal("ZITADEL mode must require its authenticated session for Companion management")
	}
	if credentialFreeLocalCompanionManagementAllowed(config.Config{BindMode: "lan"}, nil) {
		t.Fatal("LAN mode must require an unlocked session for Companion management")
	}
	if !credentialFreeLocalCompanionManagementAllowed(config.Config{BindMode: "local"}, nil) {
		t.Fatal("credential-free local mode must keep Companion management usable")
	}
}

func TestCloudSessionBootstrapReturnsEntitlement(t *testing.T) {
	a := newTestApp(t)

	// header.payload.sig where payload is {"sub":"user_123","email":"owner@example.com","plan":"pro"}
	token := "e30.eyJzdWIiOiJ1c2VyXzEyMyIsImVtYWlsIjoib3duZXJAZXhhbXBsZS5jb20iLCJwbGFuIjoicHJvIn0.e30"
	resp := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/auth/cloud/session/bootstrap",
		strings.NewReader(`{"provider":"zitadel","token":"`+token+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)

	if resp.Code != http.StatusOK {
		t.Fatalf("bootstrap status=%d body=%s", resp.Code, resp.Body.String())
	}
	body := resp.Body.String()
	for _, needle := range []string{`"provider":"zitadel"`, `"user_id":"user_123"`, `"email":"owner@example.com"`, `"plan":"pro"`, `"ai_assist"`} {
		if !strings.Contains(body, needle) {
			t.Fatalf("expected %q in body %s", needle, body)
		}
	}
}

func TestCloudSessionBootstrapRejectsUnsupportedProvider(t *testing.T) {
	a := newTestApp(t)

	token := "e30.eyJzdWIiOiJ1c2VyXzEyMyIsImVtYWlsIjoib3duZXJAZXhhbXBsZS5jb20iLCJwbGFuIjoicHJvIn0.e30"
	resp := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/auth/cloud/session/bootstrap",
		strings.NewReader(`{"provider":"clerk","token":"`+token+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("unsupported provider expected 400, got %d body=%s", resp.Code, resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"error":"unsupported_provider"`) {
		t.Fatalf("expected unsupported_provider error, got %s", resp.Body.String())
	}
}

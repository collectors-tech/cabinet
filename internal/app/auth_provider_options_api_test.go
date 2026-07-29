package app

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestAuthProviderOptionsDefaults(t *testing.T) {
	t.Setenv("CABINET_AUTH_IDENTITY_MODE", "")
	t.Setenv("VITE_CLERK_PUBLISHABLE_KEY", "")
	t.Setenv("CABINET_AUTH_PROVIDER_GOOGLE_ENABLED", "")
	t.Setenv("CABINET_AUTH_PROVIDER_APPLE_ENABLED", "")
	t.Setenv("CABINET_AUTH_PROVIDER_MICROSOFT_ENABLED", "")

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodGet, "/api/auth/provider-options", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("provider-options expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode provider-options payload: %v", err)
	}
	if payload["identity_mode"] != "local" {
		t.Fatalf("expected local identity mode, got %v", payload["identity_mode"])
	}
	providers, ok := payload["providers"].([]any)
	if !ok || len(providers) < 3 {
		t.Fatalf("expected provider options array with 3 entries")
	}
}

func TestAuthProviderOptionsRejectsUnsupportedIdentityEnv(t *testing.T) {
	t.Setenv("CABINET_AUTH_IDENTITY_MODE", "clerk")
	t.Setenv("VITE_CLERK_PUBLISHABLE_KEY", "pk_test_123")
	t.Setenv("CABINET_AUTH_PROVIDER_APPLE_ENABLED", "false")

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodGet, "/api/auth/provider-options", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("provider-options expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var payload struct {
		IdentityMode string `json:"identity_mode"`
		Providers    []struct {
			ID      string `json:"id"`
			Enabled bool   `json:"enabled"`
		} `json:"providers"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode provider-options payload: %v", err)
	}
	if payload.IdentityMode != "local" {
		t.Fatalf("expected unsupported identity mode to fall back to local, got %s", payload.IdentityMode)
	}
	var rawPayload map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &rawPayload); err != nil {
		t.Fatalf("decode provider-options raw payload: %v", err)
	}
	if _, ok := rawPayload["clerk_configured"]; ok {
		t.Fatalf("provider options must not expose unsupported clerk_configured compatibility field: %s", resp.Body.String())
	}
	foundApple := false
	for _, provider := range payload.Providers {
		if provider.ID == "apple" {
			foundApple = true
			if provider.Enabled {
				t.Fatalf("expected apple provider disabled from env override")
			}
		}
	}
	if !foundApple {
		t.Fatalf("expected apple provider in provider options")
	}
}

func TestAuthProviderOptionsReportsConfiguredZitadel(t *testing.T) {
	t.Setenv("CABINET_AUTH_IDENTITY_MODE", "zitadel")
	t.Setenv("CABINET_ZITADEL_ISSUER", "https://identity.example.com")
	t.Setenv("CABINET_ZITADEL_CLIENT_ID", "cabinet-client")
	t.Setenv("CABINET_ZITADEL_AUDIENCE", "cabinet-project")
	t.Setenv("CABINET_PUBLIC_ORIGIN", "https://cabinet.example.com")

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodGet, "/api/auth/provider-options", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("provider-options expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	for _, fragment := range []string{`"identity_mode":"zitadel"`, `"zitadel_configured":true`, `"zitadel_login_path":"/api/auth/zitadel/login"`} {
		if !strings.Contains(resp.Body.String(), fragment) {
			t.Fatalf("configured ZITADEL response missing %s: %s", fragment, resp.Body.String())
		}
	}
	setup := doRequest(t, a, http.MethodGet, "/api/runtime/setup-status", nil, nil)
	if setup.Code != http.StatusOK || !strings.Contains(setup.Body.String(), `"setup_required":false`) {
		t.Fatalf("configured remote identity must bypass the unauthenticated local setup wizard: status=%d body=%s", setup.Code, setup.Body.String())
	}
}

func TestAuthProviderOptionsE2EOverrideHook(t *testing.T) {
	if err := os.Setenv("CABINET_E2E_MODE", "1"); err != nil {
		t.Fatalf("set CABINET_E2E_MODE: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("CABINET_E2E_MODE") })

	a := newTestApp(t)
	override := `{"identity_mode":"zitadel","providers":[{"id":"google","label":"Google","enabled":true},{"id":"apple","label":"Apple","enabled":false},{"id":"microsoft","label":"Microsoft","enabled":true}]}`
	hookResp := doRequest(t, a, http.MethodPost, "/api/test/auth/provider-options", strings.NewReader(override), map[string]string{"Content-Type": "application/json"})
	if hookResp.Code != http.StatusOK {
		t.Fatalf("provider-options override hook expected 200, got %d body=%s", hookResp.Code, hookResp.Body.String())
	}

	resp := doRequest(t, a, http.MethodGet, "/api/auth/provider-options", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("provider-options expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"identity_mode":"zitadel"`) {
		t.Fatalf("expected zitadel identity mode in overridden response: %s", resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"id":"apple","label":"Apple","enabled":false`) {
		t.Fatalf("expected overridden apple provider state in response: %s", resp.Body.String())
	}
}

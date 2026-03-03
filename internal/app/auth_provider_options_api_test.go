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

func TestAuthProviderOptionsRespectsEnv(t *testing.T) {
	t.Setenv("CABINET_AUTH_IDENTITY_MODE", "clerk")
	t.Setenv("VITE_CLERK_PUBLISHABLE_KEY", "pk_test_123")
	t.Setenv("CABINET_AUTH_PROVIDER_APPLE_ENABLED", "false")

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodGet, "/api/auth/provider-options", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("provider-options expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var payload struct {
		IdentityMode    string `json:"identity_mode"`
		ClerkConfigured bool   `json:"clerk_configured"`
		Providers       []struct {
			ID      string `json:"id"`
			Enabled bool   `json:"enabled"`
		} `json:"providers"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode provider-options payload: %v", err)
	}
	if payload.IdentityMode != "clerk" {
		t.Fatalf("expected clerk identity mode, got %s", payload.IdentityMode)
	}
	if !payload.ClerkConfigured {
		t.Fatalf("expected clerk configured=true when publishable key is set")
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

func TestAuthProviderOptionsE2EOverrideHook(t *testing.T) {
	if err := os.Setenv("CABINET_E2E_MODE", "1"); err != nil {
		t.Fatalf("set CABINET_E2E_MODE: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("CABINET_E2E_MODE") })

	a := newTestApp(t)
	override := `{"identity_mode":"clerk","providers":[{"id":"google","label":"Google","enabled":true},{"id":"apple","label":"Apple","enabled":false},{"id":"microsoft","label":"Microsoft","enabled":true}]}`
	hookResp := doRequest(t, a, http.MethodPost, "/api/test/auth/provider-options", strings.NewReader(override), map[string]string{"Content-Type": "application/json"})
	if hookResp.Code != http.StatusOK {
		t.Fatalf("provider-options override hook expected 200, got %d body=%s", hookResp.Code, hookResp.Body.String())
	}

	resp := doRequest(t, a, http.MethodGet, "/api/auth/provider-options", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("provider-options expected 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"identity_mode":"clerk"`) {
		t.Fatalf("expected clerk identity mode in overridden response: %s", resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), `"id":"apple","label":"Apple","enabled":false`) {
		t.Fatalf("expected overridden apple provider state in response: %s", resp.Body.String())
	}
}

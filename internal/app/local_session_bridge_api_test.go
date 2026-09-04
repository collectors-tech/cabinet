package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLocalSessionBootstrapIssuesProfileBoundCredentialOnlyAtTrustedLoopbackOrigin(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileID := createAndActivateLocalSessionTestProfile(t, a, "Local session bridge")
	body := `{"profile_id":"` + profileID + `"}`

	missingOrigin := doRequest(t, a, http.MethodPost, "/api/auth/local/session", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if missingOrigin.Code != http.StatusForbidden {
		t.Fatalf("origin-less browser session bootstrap status=%d body=%s", missingOrigin.Code, missingOrigin.Body.String())
	}

	response := doRequest(t, a, http.MethodPost, "/api/auth/local/session", strings.NewReader(body), map[string]string{
		"Content-Type":   "application/json",
		"Origin":         "http://127.0.0.1:8080",
		"Sec-Fetch-Site": "same-origin",
	})
	if response.Code != http.StatusOK {
		t.Fatalf("same-origin local session bootstrap status=%d body=%s", response.Code, response.Body.String())
	}
	var payload struct {
		OK           bool   `json:"ok"`
		SessionToken string `json:"session_token"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode local session bootstrap: %v", err)
	}
	if !payload.OK || len(payload.SessionToken) < 40 {
		t.Fatalf("expected opaque server session, got %+v", payload)
	}
	boundProfileID, err := a.authService.ValidateUnlockedSessionProfile(payload.SessionToken)
	if err != nil || boundProfileID != profileID {
		t.Fatalf("session profile binding=%q err=%v want=%q", boundProfileID, err, profileID)
	}
	if strings.Contains(response.Body.String(), profileID) {
		t.Fatalf("session response should not duplicate profile data beside the opaque credential: %s", response.Body.String())
	}
}

func TestLocalSessionBootstrapFailsForWrongProfileRegisteredProfileAndLAN(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	activeProfileID := createAndActivateLocalSessionTestProfile(t, a, "Active local profile")
	otherProfileID := createLocalSessionTestProfile(t, a, "Other local profile")
	headers := map[string]string{
		"Content-Type":   "application/json",
		"Origin":         "http://127.0.0.1:8080",
		"Sec-Fetch-Site": "same-origin",
	}

	wrongProfile := doRequest(t, a, http.MethodPost, "/api/auth/local/session", strings.NewReader(`{"profile_id":"`+otherProfileID+`"}`), headers)
	if wrongProfile.Code != http.StatusForbidden || strings.Contains(wrongProfile.Body.String(), activeProfileID) {
		t.Fatalf("wrong-profile local session status=%d body=%s", wrongProfile.Code, wrongProfile.Body.String())
	}

	if _, err := a.db.Exec(`INSERT INTO webauthn_credentials(id, profile_id, credential_json) VALUES('local-session-passkey', ?, '{}')`, activeProfileID); err != nil {
		t.Fatalf("seed registered profile: %v", err)
	}
	registered := doRequest(t, a, http.MethodPost, "/api/auth/local/session", strings.NewReader(`{"profile_id":"`+activeProfileID+`"}`), headers)
	if registered.Code != http.StatusConflict || !strings.Contains(registered.Body.String(), "passkey_authentication_required") {
		t.Fatalf("registered profile local session status=%d body=%s", registered.Code, registered.Body.String())
	}

	lanCfg := a.cfg
	lanCfg.BindMode = "lan"
	trustedRequest := httptest.NewRequest(http.MethodPost, "/api/auth/local/session", nil)
	trustedRequest.Host = "127.0.0.1:8080"
	trustedRequest.Header.Set("Origin", "http://127.0.0.1:8080")
	trustedRequest.Header.Set("Sec-Fetch-Site", "same-origin")
	if localSessionBootstrapAllowed(lanCfg, nil, trustedRequest) {
		t.Fatal("LAN mode must never allow credential-free local session bootstrap")
	}
	zitadel := newZitadelAuthBoundary(zitadelAuthConfig{IdentityMode: "zitadel"}, nil)
	if localSessionBootstrapAllowed(a.cfg, zitadel, trustedRequest) {
		t.Fatal("ZITADEL mode must never allow credential-free local session bootstrap")
	}
}

func TestLocalSessionBootstrapIsProtectedByOriginAndHostBoundary(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileID := createAndActivateLocalSessionTestProfile(t, a, "Local boundary profile")
	body := `{"profile_id":"` + profileID + `"}`

	hostileOrigin := doRequest(t, a, http.MethodPost, "/api/auth/local/session", strings.NewReader(body), map[string]string{
		"Content-Type":   "application/json",
		"Origin":         "https://attacker.invalid",
		"Sec-Fetch-Site": "cross-site",
	})
	if hostileOrigin.Code != http.StatusForbidden || !strings.Contains(hostileOrigin.Body.String(), "cross_site_request_blocked") {
		t.Fatalf("hostile origin status=%d body=%s", hostileOrigin.Code, hostileOrigin.Body.String())
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/local/session", strings.NewReader(body))
	req.Host = "attacker.invalid"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://attacker.invalid")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	recorder := httptest.NewRecorder()
	a.srv.Handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusForbidden || !strings.Contains(recorder.Body.String(), "untrusted_host") {
		t.Fatalf("hostile host status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if strings.Contains(hostileOrigin.Body.String(), "session_token") || strings.Contains(recorder.Body.String(), "session_token") {
		t.Fatal("rejected boundary request exposed a session credential field")
	}
}

func TestLocalSessionLockAcceptsHeaderAndInvalidatesCredential(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileID := createAndActivateLocalSessionTestProfile(t, a, "Local session lock")
	token, err := a.authService.CreateUnlockedSession(profileID)
	if err != nil {
		t.Fatalf("create local session: %v", err)
	}
	locked := doRequest(t, a, http.MethodPost, "/api/auth/session/lock", strings.NewReader(`{}`), map[string]string{
		"Content-Type":      "application/json",
		"X-Cabinet-Session": token,
	})
	if locked.Code != http.StatusOK || strings.Contains(locked.Body.String(), token) {
		t.Fatalf("header session lock status=%d body=%s", locked.Code, locked.Body.String())
	}
	if err := a.authService.ValidateUnlockedSession(token); err == nil {
		t.Fatal("server session remained valid after header-based lock")
	}
}

func createAndActivateLocalSessionTestProfile(t *testing.T, a *App, name string) string {
	t.Helper()
	profileID := createLocalSessionTestProfile(t, a, name)
	response := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profileID+`"}`), map[string]string{"Content-Type": "application/json"})
	if response.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", response.Code, response.Body.String())
	}
	return profileID
}

func createLocalSessionTestProfile(t *testing.T, a *App, name string) string {
	t.Helper()
	response := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"`+name+`"}`), map[string]string{"Content-Type": "application/json"})
	if response.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", response.Code, response.Body.String())
	}
	var payload struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	return payload.ID
}

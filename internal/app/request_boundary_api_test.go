package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/config"
)

func TestLocalRequestBoundaryRejectsCrossSiteSimpleMutation(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	req := httptest.NewRequest(http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Cross-site profile"}`))
	req.Host = "127.0.0.1:8080"
	req.Header.Set("Origin", "https://attacker.example")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()
	a.srv.Handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("cross-site simple mutation status=%d body=%s", rr.Code, rr.Body.String())
	}
	var count int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM profiles WHERE name = 'Cross-site profile'`).Scan(&count); err != nil {
		t.Fatalf("count cross-site profiles: %v", err)
	}
	if count != 0 {
		t.Fatalf("cross-site mutation persisted %d profiles", count)
	}
}

func TestLocalRequestBoundaryRejectsUntrustedHostAndAllowsSameOriginUI(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	hostile := httptest.NewRequest(http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Host injection profile"}`))
	hostile.Host = "attacker.example"
	hostile.Header.Set("Content-Type", "application/json")
	hostileRecorder := httptest.NewRecorder()
	a.srv.Handler.ServeHTTP(hostileRecorder, hostile)
	if hostileRecorder.Code != http.StatusForbidden {
		t.Fatalf("untrusted host status=%d body=%s", hostileRecorder.Code, hostileRecorder.Body.String())
	}

	sameOrigin := httptest.NewRequest(http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Same-origin profile"}`))
	sameOrigin.Host = "127.0.0.1:8080"
	sameOrigin.Header.Set("Origin", "http://127.0.0.1:8080")
	sameOrigin.Header.Set("Sec-Fetch-Site", "same-origin")
	sameOrigin.Header.Set("Content-Type", "application/json")
	sameOriginRecorder := httptest.NewRecorder()
	a.srv.Handler.ServeHTTP(sameOriginRecorder, sameOrigin)
	if sameOriginRecorder.Code != http.StatusCreated {
		t.Fatalf("same-origin Cabinet mutation status=%d body=%s", sameOriginRecorder.Code, sameOriginRecorder.Body.String())
	}

	cli := httptest.NewRequest(http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"CLI profile"}`))
	cli.Host = "127.0.0.1:8080"
	cli.Header.Set("Content-Type", "application/json")
	cliRecorder := httptest.NewRecorder()
	a.srv.Handler.ServeHTTP(cliRecorder, cli)
	if cliRecorder.Code != http.StatusCreated {
		t.Fatalf("origin-less trusted-host CLI mutation status=%d body=%s", cliRecorder.Code, cliRecorder.Body.String())
	}
}

func TestPublicAPIAllowlistIsMinimalAndMethodAware(t *testing.T) {
	t.Parallel()

	cases := []struct {
		method string
		path   string
		public bool
	}{
		{method: http.MethodGet, path: "/api/runtime", public: true},
		{method: http.MethodGet, path: "/api/runtime/setup-status", public: true},
		{method: http.MethodGet, path: "/api/auth/provider-options", public: true},
		{method: http.MethodPost, path: "/api/auth/webauthn/login/begin", public: true},
		{method: http.MethodPost, path: "/api/auth/webauthn/login/finish", public: true},
		{method: http.MethodPost, path: "/api/auth/session/validate", public: true},
		{method: http.MethodGet, path: "/api/auth/zitadel/login", public: true},
		{method: http.MethodPost, path: "/api/auth/zitadel/refresh", public: true},
		{method: http.MethodPost, path: "/api/auth/zitadel/login", public: false},
		{method: http.MethodGet, path: "/api/auth/zitadel/refresh", public: false},
		{method: http.MethodGet, path: "/api/auth/zitadel/future-route", public: false},
		{method: http.MethodPost, path: "/api/runtime/setup-complete", public: false},
		{method: http.MethodPost, path: "/api/runtime/setup-import", public: false},
		{method: http.MethodPost, path: "/api/runtime/setup-storage-validate", public: false},
		{method: http.MethodPost, path: "/api/profiles", public: false},
		{method: http.MethodPost, path: "/api/auth/webauthn/register/begin", public: false},
		{method: http.MethodPost, path: "/api/auth/recovery/passphrase", public: false},
		{method: http.MethodGet, path: "/api/profiles", public: false},
		{method: http.MethodGet, path: "/api/items", public: false},
		{method: http.MethodPost, path: "/api/items", public: false},
		{method: http.MethodGet, path: "/api/auth/cloud/session/effective", public: false},
	}
	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if got := isPublicAPIRequest(req); got != tc.public {
				t.Fatalf("isPublicAPIRequest(%s %s)=%v want=%v", tc.method, tc.path, got, tc.public)
			}
		})
	}
}

func TestInitialSetupBootstrapRequestIsMinimalAndStateAware(t *testing.T) {
	t.Parallel()

	cfg := config.Config{DataDir: t.TempDir()}
	cases := []struct {
		method string
		path   string
		want   bool
	}{
		{method: http.MethodPost, path: "/api/runtime/setup-complete", want: true},
		{method: http.MethodPost, path: "/api/runtime/setup-import", want: true},
		{method: http.MethodPost, path: "/api/runtime/setup-storage-validate", want: true},
		{method: http.MethodGet, path: "/api/runtime/setup-complete", want: false},
		{method: http.MethodPost, path: "/api/runtime/setup-status", want: false},
		{method: http.MethodPost, path: "/api/profiles", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if got := isInitialSetupBootstrapRequest(cfg, req); got != tc.want {
				t.Fatalf("isInitialSetupBootstrapRequest(%s %s)=%v want=%v", tc.method, tc.path, got, tc.want)
			}
		})
	}

	if err := writeRuntimeSetupConfig(cfg, runtimeSetupConfigFile{}); err != nil {
		t.Fatalf("write setup config: %v", err)
	}
	for _, path := range []string{
		"/api/runtime/setup-complete",
		"/api/runtime/setup-import",
		"/api/runtime/setup-storage-validate",
	} {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		if isInitialSetupBootstrapRequest(cfg, req) {
			t.Fatalf("completed setup unexpectedly retained bootstrap exception for %s", path)
		}
	}
}

func TestInitialSetupCompletionBypassesIncompleteZitadelBoundaryOnlyDuringBootstrap(t *testing.T) {
	t.Setenv("CABINET_AUTH_IDENTITY_MODE", "zitadel")
	t.Setenv("CABINET_ZITADEL_ISSUER", "")
	t.Setenv("CABINET_ZITADEL_CLIENT_ID", "")
	t.Setenv("CABINET_ZITADEL_AUDIENCE", "")
	t.Setenv("CABINET_PUBLIC_ORIGIN", "")

	a := newTestApp(t)
	if !runtimeSetupRequired(a.cfg) {
		t.Fatal("test precondition failed: initial setup must be required")
	}

	body := `{
		"instance_name":"Boundary Setup",
		"profile_key":"boundary-setup",
		"auth_mode":"local",
		"runtime_port_mode":"auto",
		"bootstrap_workspace":"Local Workspace",
		"bootstrap_database_ref":"Primary DB"
	}`
	first := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/runtime/setup-complete",
		strings.NewReader(body),
		map[string]string{"Content-Type": "application/json"},
	)
	if first.Code != http.StatusOK {
		t.Fatalf("initial setup completion status=%d body=%s", first.Code, first.Body.String())
	}
	if runtimeSetupRequired(a.cfg) {
		t.Fatal("valid initial setup did not persist configuration")
	}

	second := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/runtime/setup-complete",
		strings.NewReader(body),
		map[string]string{"Content-Type": "application/json"},
	)
	if second.Code != http.StatusUnauthorized {
		t.Fatalf("post-setup request status=%d want=%d body=%s", second.Code, http.StatusUnauthorized, second.Body.String())
	}
}

func TestInitialSetupBootstrapStillRejectsCrossSiteMutation(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	req := httptest.NewRequest(http.MethodPost, "/api/runtime/setup-complete", strings.NewReader(`{"instance_name":"Hostile"}`))
	req.Host = "127.0.0.1:8080"
	req.Header.Set("Origin", "https://attacker.example")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	a.srv.Handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("cross-site setup mutation status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	if !runtimeSetupRequired(a.cfg) {
		t.Fatal("cross-site setup mutation persisted configuration")
	}
}

func TestE2EModeBypassesLocalUnlockOnlyWhenEnabled(t *testing.T) {
	t.Parallel()

	resetReq := httptest.NewRequest(http.MethodPost, "/api/test/reset", nil)
	if got := requiresUnlockedSession(config.Config{}, resetReq); !got {
		t.Fatalf("E2E hook should remain locked when E2E hooks are disabled")
	}
	if got := requiresUnlockedSession(config.Config{EnableE2EHooks: true}, resetReq); got {
		t.Fatalf("E2E hook should bypass local unlock when E2E hooks are enabled")
	}

	itemReq := httptest.NewRequest(http.MethodPost, "/api/items", nil)
	if got := requiresUnlockedSession(config.Config{}, itemReq); !got {
		t.Fatalf("normal mutation should require local unlock when E2E hooks are disabled")
	}
	if got := requiresUnlockedSession(config.Config{EnableE2EHooks: true}, itemReq); got {
		t.Fatalf("E2E mode should bypass local unlock for fixture-driven API mutations")
	}
}

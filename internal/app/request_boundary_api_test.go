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

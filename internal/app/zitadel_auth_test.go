package app

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type fakeOIDCProvider struct {
	t          *testing.T
	server     *httptest.Server
	key        *rsa.PrivateKey
	kid        string
	clientID   string
	audience   string
	requiredRole string

	mu               sync.Mutex
	nonce            string
	expectedChallenge string
	lastVerifier      string
	refreshCount      int
}

func newFakeOIDCProvider(t *testing.T) *fakeOIDCProvider {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate fake OIDC key: %v", err)
	}
	provider := &fakeOIDCProvider{
		t: t, key: key, kid: "cabinet-test-key", clientID: "cabinet-test-client",
		audience: "cabinet-test-project", requiredRole: "cabinet.user",
	}
	provider.server = httptest.NewServer(http.HandlerFunc(provider.serveHTTP))
	t.Cleanup(provider.server.Close)
	return provider
}

func (provider *fakeOIDCProvider) serveHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/.well-known/openid-configuration":
		writeFakeJSON(w, map[string]any{
			"issuer": provider.server.URL,
			"authorization_endpoint": provider.server.URL + "/oauth/v2/authorize",
			"token_endpoint": provider.server.URL + "/oauth/v2/token",
			"jwks_uri": provider.server.URL + "/oauth/v2/keys",
			"userinfo_endpoint": provider.server.URL + "/oidc/v1/userinfo",
			"end_session_endpoint": provider.server.URL + "/oidc/v1/end_session",
		})
	case "/oauth/v2/keys":
		writeFakeJSON(w, map[string]any{"keys": []any{provider.jwk(provider.key, provider.kid)}})
	case "/oauth/v2/token":
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		provider.mu.Lock()
		provider.lastVerifier = r.Form.Get("code_verifier")
		if r.Form.Get("grant_type") == "refresh_token" {
			provider.refreshCount++
		}
		nonce := provider.nonce
		provider.mu.Unlock()
		writeFakeJSON(w, map[string]any{
			"access_token": "server-only-access-token",
			"id_token": provider.signToken(provider.validClaims(nonce), provider.key, provider.kid),
			"refresh_token": "server-only-refresh-token",
			"token_type": "Bearer",
			"expires_in": 3600,
		})
	case "/oidc/v1/userinfo":
		if r.Header.Get("Authorization") != "Bearer server-only-access-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		writeFakeJSON(w, map[string]any{"sub": "user-123", "email": "owner@example.com", "name": "Cabinet Owner"})
	default:
		http.NotFound(w, r)
	}
}

func (provider *fakeOIDCProvider) boundary() *zitadelAuthBoundary {
	return newZitadelAuthBoundary(zitadelAuthConfig{
		IdentityMode: "zitadel",
		Issuer: provider.server.URL,
		ClientID: provider.clientID,
		ProjectID: provider.audience,
		Audience: provider.audience,
		PublicOrigin: "http://cabinet.test",
		Scopes: []string{"openid", "profile", "email", "offline_access", zitadelRoleClaim},
		RequiredRoles: []string{provider.requiredRole, "cabinet.admin"},
		AllowInsecureHTTP: true,
	}, provider.server.Client())
}

func (provider *fakeOIDCProvider) validClaims(nonce string) jwt.MapClaims {
	return jwt.MapClaims{
		"iss": provider.server.URL,
		"sub": "user-123",
		"aud": []string{provider.clientID, provider.audience},
		"azp": provider.clientID,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Add(-time.Minute).Unix(),
		"nonce": nonce,
		zitadelRoleClaim: map[string]any{provider.requiredRole: map[string]any{"org-1": "cabinet.test"}},
	}
}

func (provider *fakeOIDCProvider) signToken(claims jwt.MapClaims, key *rsa.PrivateKey, kid string) string {
	provider.t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	raw, err := token.SignedString(key)
	if err != nil {
		provider.t.Fatalf("sign fake ID token: %v", err)
	}
	return raw
}

func (provider *fakeOIDCProvider) jwk(key *rsa.PrivateKey, kid string) map[string]any {
	exponent := big.NewInt(int64(key.PublicKey.E)).Bytes()
	return map[string]any{
		"kty": "RSA", "kid": kid, "alg": "RS256", "use": "sig",
		"n": base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
		"e": base64.RawURLEncoding.EncodeToString(exponent),
	}
}

func writeFakeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}

func TestZitadelAuthorizationCodePKCECallbackAndSession(t *testing.T) {
	provider := newFakeOIDCProvider(t)
	boundary := provider.boundary()
	mux := http.NewServeMux()
	registerZitadelAuthRoutes(mux, boundary)

	loginRequest := httptest.NewRequest(http.MethodGet, "/api/auth/zitadel/login?return_to=%2Fsettings%2Fdisplay", nil)
	loginResponse := httptest.NewRecorder()
	mux.ServeHTTP(loginResponse, loginRequest)
	if loginResponse.Code != http.StatusFound {
		t.Fatalf("login start status=%d body=%s", loginResponse.Code, loginResponse.Body.String())
	}
	authorizeURL, err := url.Parse(loginResponse.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse authorization redirect: %v", err)
	}
	query := authorizeURL.Query()
	for key, expected := range map[string]string{
		"client_id": provider.clientID,
		"response_type": "code",
		"code_challenge_method": "S256",
	} {
		if query.Get(key) != expected {
			t.Fatalf("expected %s=%q, got %q", key, expected, query.Get(key))
		}
	}
	if query.Get("state") == "" || query.Get("nonce") == "" || query.Get("code_challenge") == "" {
		t.Fatalf("state, nonce and code challenge must be present")
	}
	provider.mu.Lock()
	provider.nonce = query.Get("nonce")
	provider.expectedChallenge = query.Get("code_challenge")
	provider.mu.Unlock()

	transactionCookie := findResponseCookie(t, loginResponse.Result(), zitadelTransactionCookieName)
	callbackRequest := httptest.NewRequest(http.MethodGet, "/api/auth/zitadel/callback?code=valid-code&state="+url.QueryEscape(query.Get("state")), nil)
	callbackRequest.AddCookie(transactionCookie)
	callbackResponse := httptest.NewRecorder()
	mux.ServeHTTP(callbackResponse, callbackRequest)
	if callbackResponse.Code != http.StatusFound || callbackResponse.Header().Get("Location") != "/settings/display" {
		t.Fatalf("callback status=%d location=%q body=%s", callbackResponse.Code, callbackResponse.Header().Get("Location"), callbackResponse.Body.String())
	}
	provider.mu.Lock()
	verifier := provider.lastVerifier
	expectedChallenge := provider.expectedChallenge
	provider.mu.Unlock()
	challenge := sha256.Sum256([]byte(verifier))
	if base64.RawURLEncoding.EncodeToString(challenge[:]) != expectedChallenge {
		t.Fatalf("token exchange did not use the original PKCE verifier")
	}

	sessionCookie := findResponseCookie(t, callbackResponse.Result(), zitadelSessionCookieName)
	if !sessionCookie.HttpOnly || sessionCookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected HTTP-only SameSite=Lax application session cookie: %#v", sessionCookie)
	}
	sessionRequest := httptest.NewRequest(http.MethodGet, "/api/auth/zitadel/session", nil)
	sessionRequest.AddCookie(sessionCookie)
	sessionResponse := httptest.NewRecorder()
	mux.ServeHTTP(sessionResponse, sessionRequest)
	if sessionResponse.Code != http.StatusOK {
		t.Fatalf("session status=%d body=%s", sessionResponse.Code, sessionResponse.Body.String())
	}
	for _, forbidden := range []string{"server-only-access-token", "server-only-refresh-token", "id_token"} {
		if strings.Contains(sessionResponse.Body.String(), forbidden) {
			t.Fatalf("browser session response leaked %q", forbidden)
		}
	}
	for _, required := range []string{"owner@example.com", "cabinet.user", "user-123"} {
		if !strings.Contains(sessionResponse.Body.String(), required) {
			t.Fatalf("session response missing %q: %s", required, sessionResponse.Body.String())
		}
	}
}

func TestZitadelCallbackConsumesStateExactlyOnce(t *testing.T) {
	provider := newFakeOIDCProvider(t)
	boundary := provider.boundary()
	mux := http.NewServeMux()
	registerZitadelAuthRoutes(mux, boundary)

	loginResponse := httptest.NewRecorder()
	mux.ServeHTTP(loginResponse, httptest.NewRequest(http.MethodGet, "/api/auth/zitadel/login", nil))
	authorizeURL, _ := url.Parse(loginResponse.Header().Get("Location"))
	provider.mu.Lock()
	provider.nonce = authorizeURL.Query().Get("nonce")
	provider.mu.Unlock()
	cookie := findResponseCookie(t, loginResponse.Result(), zitadelTransactionCookieName)
	callbackPath := "/api/auth/zitadel/callback?code=valid-code&state="+url.QueryEscape(authorizeURL.Query().Get("state"))

	first := httptest.NewRecorder()
	firstRequest := httptest.NewRequest(http.MethodGet, callbackPath, nil)
	firstRequest.AddCookie(cookie)
	mux.ServeHTTP(first, firstRequest)
	if first.Code != http.StatusFound || strings.Contains(first.Header().Get("Location"), "auth_error") {
		t.Fatalf("expected first callback to succeed")
	}
	second := httptest.NewRecorder()
	secondRequest := httptest.NewRequest(http.MethodGet, callbackPath, nil)
	secondRequest.AddCookie(cookie)
	mux.ServeHTTP(second, secondRequest)
	if second.Code != http.StatusFound || !strings.Contains(second.Header().Get("Location"), "auth_error=invalid_identity_callback") {
		t.Fatalf("expected replayed callback to fail closed: status=%d location=%q", second.Code, second.Header().Get("Location"))
	}
}

func TestZitadelRegistrationIntentUsesBrandedProviderRegistration(t *testing.T) {
	provider := newFakeOIDCProvider(t)
	mux := http.NewServeMux()
	registerZitadelAuthRoutes(mux, provider.boundary())
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/auth/zitadel/login?intent=register", nil))
	redirectURL, err := url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse registration redirect: %v", err)
	}
	if redirectURL.Query().Get("prompt") != "create" {
		t.Fatalf("expected ZITADEL prompt=create registration handoff")
	}
}

func TestZitadelIDTokenValidationDeniesInvalidClaims(t *testing.T) {
	provider := newFakeOIDCProvider(t)
	boundary := provider.boundary()
	discovery, err := boundary.discover(t.Context())
	if err != nil {
		t.Fatalf("discover fake provider: %v", err)
	}
	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate alternate key: %v", err)
	}

	tests := []struct {
		name string
		mutate func(jwt.MapClaims)
		key *rsa.PrivateKey
		kid string
	}{
		{name: "issuer", mutate: func(c jwt.MapClaims) { c["iss"] = "https://wrong.example" }},
		{name: "audience", mutate: func(c jwt.MapClaims) { c["aud"] = []string{provider.clientID, "wrong-audience"} }},
		{name: "authorised party", mutate: func(c jwt.MapClaims) { c["azp"] = "wrong-client" }},
		{name: "missing authorised party", mutate: func(c jwt.MapClaims) { delete(c, "azp") }},
		{name: "nonce", mutate: func(c jwt.MapClaims) { c["nonce"] = "wrong-nonce" }},
		{name: "expiry", mutate: func(c jwt.MapClaims) { c["exp"] = time.Now().Add(-2 * time.Minute).Unix() }},
		{name: "role", mutate: func(c jwt.MapClaims) { c[zitadelRoleClaim] = map[string]any{"unrelated.role": map[string]any{"org": "example"}} }},
		{name: "signature", mutate: func(c jwt.MapClaims) {}, key: otherKey},
		{name: "unknown key", mutate: func(c jwt.MapClaims) {}, kid: "unknown-key"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			claims := provider.validClaims("expected-nonce")
			test.mutate(claims)
			key := test.key
			if key == nil {
				key = provider.key
			}
			kid := test.kid
			if kid == "" {
				kid = provider.kid
			}
			raw := provider.signToken(claims, key, kid)
			if _, err := boundary.validateIDToken(t.Context(), discovery, raw, "expected-nonce"); err == nil {
				t.Fatalf("expected %s token to be denied", test.name)
			}
		})
	}
}

func TestZitadelSessionRefreshAndLogoutKeepTokensServerSide(t *testing.T) {
	provider := newFakeOIDCProvider(t)
	boundary := provider.boundary()
	sessionID := "opaque-session-id"
	boundary.sessions[sessionID] = zitadelSession{
		Identity: zitadelIdentity{Subject: "user-123", Roles: []string{"cabinet.user"}, ExpiresAt: time.Now().Add(-time.Minute)},
		RefreshToken: "server-only-refresh-token",
		ExpiresAt: time.Now().Add(-time.Minute),
	}
	mux := http.NewServeMux()
	registerZitadelAuthRoutes(mux, boundary)

	refreshRequest := httptest.NewRequest(http.MethodPost, "/api/auth/zitadel/refresh", nil)
	refreshRequest.AddCookie(&http.Cookie{Name: zitadelSessionCookieName, Value: sessionID})
	refreshResponse := httptest.NewRecorder()
	mux.ServeHTTP(refreshResponse, refreshRequest)
	if refreshResponse.Code != http.StatusOK {
		t.Fatalf("refresh status=%d body=%s", refreshResponse.Code, refreshResponse.Body.String())
	}
	if strings.Contains(refreshResponse.Body.String(), "server-only") {
		t.Fatalf("refresh response leaked provider token")
	}

	logoutRequest := httptest.NewRequest(http.MethodPost, "/api/auth/zitadel/logout", nil)
	logoutRequest.AddCookie(&http.Cookie{Name: zitadelSessionCookieName, Value: sessionID})
	logoutResponse := httptest.NewRecorder()
	mux.ServeHTTP(logoutResponse, logoutRequest)
	if logoutResponse.Code != http.StatusOK {
		t.Fatalf("logout status=%d body=%s", logoutResponse.Code, logoutResponse.Body.String())
	}
	if strings.Contains(logoutResponse.Body.String(), "server-only") || strings.Contains(logoutResponse.Body.String(), "id_token_hint") {
		t.Fatalf("logout response leaked a provider token: %s", logoutResponse.Body.String())
	}
	if _, ok := boundary.sessions[sessionID]; ok {
		t.Fatalf("logout did not revoke the Cabinet application session")
	}
	cleared := findResponseCookie(t, logoutResponse.Result(), zitadelSessionCookieName)
	if cleared.MaxAge >= 0 {
		t.Fatalf("logout did not expire the session cookie")
	}
}

func TestZitadelStateChangingRoutesRejectForeignOrigins(t *testing.T) {
	provider := newFakeOIDCProvider(t)
	boundary := provider.boundary()
	boundary.sessions["opaque-session-id"] = zitadelSession{
		Identity:     zitadelIdentity{Subject: "user-123", Roles: []string{"cabinet.user"}},
		RefreshToken: "server-only-refresh-token",
		ExpiresAt:    time.Now().Add(time.Hour),
	}
	mux := http.NewServeMux()
	registerZitadelAuthRoutes(mux, boundary)

	for _, path := range []string{"/api/auth/zitadel/refresh", "/api/auth/zitadel/logout"} {
		request := httptest.NewRequest(http.MethodPost, path, nil)
		request.Header.Set("Origin", "https://attacker.example")
		request.AddCookie(&http.Cookie{Name: zitadelSessionCookieName, Value: "opaque-session-id"})
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		if response.Code != http.StatusForbidden {
			t.Fatalf("%s accepted a foreign origin: status=%d", path, response.Code)
		}
	}
}

func TestZitadelRemoteAPIAndAdminRoleBoundaries(t *testing.T) {
	boundary := newZitadelAuthBoundary(zitadelAuthConfig{IdentityMode: "zitadel"}, nil)
	request := httptest.NewRequest(http.MethodGet, "/api/inventory", nil)
	if !requiresZitadelSession(boundary, request) {
		t.Fatalf("remote application API must require a ZITADEL session")
	}
	for _, publicPath := range []string{"/api/runtime", "/api/runtime/setup-status", "/api/openapi.yaml", "/api/auth/provider-options", "/api/auth/zitadel/session"} {
		if requiresZitadelSession(boundary, httptest.NewRequest(http.MethodGet, publicPath, nil)) {
			t.Fatalf("public identity bootstrap path %s unexpectedly required middleware session", publicPath)
		}
	}
	if !requiresZitadelAdminRole(httptest.NewRequest(http.MethodGet, "/api/users/invite", nil)) {
		t.Fatalf("user administration must require cabinet.admin")
	}
	if (zitadelIdentity{Roles: []string{"cabinet.user"}}).hasRole("cabinet.admin") {
		t.Fatalf("cabinet.user must not inherit cabinet.admin")
	}
}

func TestZitadelCookiesAreSecureForHTTPSOrigins(t *testing.T) {
	boundary := newZitadelAuthBoundary(zitadelAuthConfig{PublicOrigin: "https://cabinet.example.com"}, nil)
	recorder := httptest.NewRecorder()
	boundary.setCookie(recorder, zitadelSessionCookieName, "opaque", time.Now().Add(time.Hour), "/")
	cookie := findResponseCookie(t, recorder.Result(), zitadelSessionCookieName)
	if !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("unexpected HTTPS session cookie attributes: %#v", cookie)
	}
}

func TestSafeCabinetReturnToRejectsExternalRedirects(t *testing.T) {
	for input, expected := range map[string]string{
		"/settings/display?tab=theme": "/settings/display?tab=theme",
		"https://evil.example/steal": "/dashboard",
		"//evil.example/steal": "/dashboard",
		"dashboard": "/dashboard",
	} {
		if got := safeCabinetReturnTo(input); got != expected {
			t.Fatalf("safeCabinetReturnTo(%q)=%q, want %q", input, got, expected)
		}
	}
}

func findResponseCookie(t *testing.T, response *http.Response, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range response.Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("response missing cookie %s; headers=%v", name, response.Header)
	return nil
}

func TestZitadelJWKEncodingIsStable(t *testing.T) {
	provider := newFakeOIDCProvider(t)
	jwk := provider.jwk(provider.key, provider.kid)
	if fmt.Sprint(jwk["kid"]) != provider.kid || strings.TrimSpace(fmt.Sprint(jwk["n"])) == "" {
		t.Fatalf("invalid fake JWK")
	}
}

package app

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	zitadelSessionCookieName     = "cabinet_oidc_session"
	zitadelTransactionCookieName = "cabinet_oidc_transaction"
	zitadelRoleClaim              = "urn:zitadel:iam:org:project:roles"
	zitadelLoginTTL               = 10 * time.Minute
	zitadelDefaultSessionTTL      = 8 * time.Hour
)

type zitadelAuthConfig struct {
	IdentityMode      string
	Issuer            string
	ClientID          string
	ClientSecret      string
	ProjectID         string
	Audience          string
	PublicOrigin      string
	Scopes            []string
	RequiredRoles     []string
	AllowInsecureHTTP bool
}

type oidcDiscoveryDocument struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
	UserInfoEndpoint      string `json:"userinfo_endpoint"`
	EndSessionEndpoint    string `json:"end_session_endpoint"`
}

type oidcTokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type zitadelIdentity struct {
	Subject   string   `json:"subject"`
	Email     string   `json:"email,omitempty"`
	Name      string   `json:"name,omitempty"`
	Roles     []string `json:"roles"`
	ExpiresAt time.Time `json:"-"`
}

func (identity zitadelIdentity) hasRole(role string) bool {
	role = strings.TrimSpace(strings.ToLower(role))
	for _, candidate := range identity.Roles {
		if strings.TrimSpace(strings.ToLower(candidate)) == role {
			return true
		}
	}
	return false
}

type zitadelLoginTransaction struct {
	State        string
	Nonce        string
	CodeVerifier string
	ReturnTo     string
	ExpiresAt    time.Time
}

type zitadelSession struct {
	Identity     zitadelIdentity
	AccessToken  string
	IDToken      string
	RefreshToken string
	ExpiresAt    time.Time
}

type zitadelAuthBoundary struct {
	config zitadelAuthConfig
	client *http.Client
	now    func() time.Time

	mu           sync.Mutex
	transactions map[string]zitadelLoginTransaction
	sessions     map[string]zitadelSession
}

func newZitadelAuthFromEnv() *zitadelAuthBoundary {
	issuer := strings.TrimRight(strings.TrimSpace(os.Getenv("CABINET_ZITADEL_ISSUER")), "/")
	clientID := strings.TrimSpace(os.Getenv("CABINET_ZITADEL_CLIENT_ID"))
	audience := strings.TrimSpace(os.Getenv("CABINET_ZITADEL_AUDIENCE"))
	if audience == "" {
		audience = clientID
	}
	requiredRoles := splitConfiguredValues(os.Getenv("CABINET_ZITADEL_REQUIRED_ROLES"))
	if len(requiredRoles) == 0 {
		requiredRoles = []string{"cabinet.user", "cabinet.demo", "cabinet.admin"}
	}
	scopes := splitConfiguredValues(os.Getenv("CABINET_ZITADEL_SCOPES"))
	if len(scopes) == 0 {
		scopes = []string{"openid", "profile", "email", "offline_access", zitadelRoleClaim}
	}
	projectID := strings.TrimSpace(os.Getenv("CABINET_ZITADEL_PROJECT_ID"))
	if projectID != "" {
		scopes = append(scopes, "urn:zitadel:iam:org:project:id:"+projectID+":aud")
	}

	return newZitadelAuthBoundary(zitadelAuthConfig{
		IdentityMode:      strings.ToLower(strings.TrimSpace(os.Getenv("CABINET_AUTH_IDENTITY_MODE"))),
		Issuer:            issuer,
		ClientID:          clientID,
		ClientSecret:      strings.TrimSpace(os.Getenv("CABINET_ZITADEL_CLIENT_SECRET")),
		ProjectID:         projectID,
		Audience:          audience,
		PublicOrigin:      strings.TrimRight(strings.TrimSpace(os.Getenv("CABINET_PUBLIC_ORIGIN")), "/"),
		Scopes:            uniqueStrings(scopes),
		RequiredRoles:     uniqueStrings(requiredRoles),
		AllowInsecureHTTP: envEnabled("CABINET_ZITADEL_ALLOW_INSECURE_HTTP"),
	}, &http.Client{Timeout: 10 * time.Second})
}

func newZitadelAuthBoundary(config zitadelAuthConfig, client *http.Client) *zitadelAuthBoundary {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &zitadelAuthBoundary{
		config:       config,
		client:       client,
		now:          time.Now,
		transactions: make(map[string]zitadelLoginTransaction),
		sessions:     make(map[string]zitadelSession),
	}
}

func (boundary *zitadelAuthBoundary) enabled() bool {
	return boundary != nil && strings.EqualFold(boundary.config.IdentityMode, "zitadel")
}

func (boundary *zitadelAuthBoundary) configured() bool {
	return boundary.enabled() && boundary.config.Issuer != "" && boundary.config.ClientID != "" && boundary.config.Audience != "" && boundary.config.PublicOrigin != ""
}

func (boundary *zitadelAuthBoundary) validateConfig() error {
	if !boundary.enabled() {
		return errors.New("zitadel identity mode is disabled")
	}
	if !boundary.configured() {
		return errors.New("zitadel identity configuration is incomplete")
	}
	issuer, err := url.Parse(boundary.config.Issuer)
	if err != nil || issuer.Host == "" || issuer.User != nil || issuer.RawQuery != "" || issuer.Fragment != "" {
		return errors.New("zitadel issuer is invalid")
	}
	publicOrigin, err := url.Parse(boundary.config.PublicOrigin)
	if err != nil || publicOrigin.Host == "" || publicOrigin.User != nil ||
		(publicOrigin.Path != "" && publicOrigin.Path != "/") || publicOrigin.RawQuery != "" || publicOrigin.Fragment != "" {
		return errors.New("Cabinet public origin is invalid")
	}
	if !boundary.config.AllowInsecureHTTP && (issuer.Scheme != "https" || publicOrigin.Scheme != "https") {
		return errors.New("zitadel and Cabinet origins must use HTTPS")
	}
	if boundary.config.AllowInsecureHTTP &&
		((issuer.Scheme != "https" && issuer.Scheme != "http") ||
			(publicOrigin.Scheme != "https" && publicOrigin.Scheme != "http")) {
		return errors.New("zitadel and Cabinet origin schemes are invalid")
	}
	return nil
}

func registerZitadelAuthRoutes(mux *http.ServeMux, boundary *zitadelAuthBoundary) {
	mux.HandleFunc("/api/auth/zitadel/login", boundary.handleLogin)
	mux.HandleFunc("/api/auth/zitadel/callback", boundary.handleCallback)
	mux.HandleFunc("/api/auth/zitadel/session", boundary.handleSession)
	mux.HandleFunc("/api/auth/zitadel/refresh", boundary.handleRefresh)
	mux.HandleFunc("/api/auth/zitadel/logout", boundary.handleLogout)
}

func requiresZitadelSession(boundary *zitadelAuthBoundary, r *http.Request) bool {
	if boundary == nil || !boundary.enabled() || !strings.HasPrefix(r.URL.Path, "/api/") {
		return false
	}
	if r.URL.Path == "/api/runtime" || r.URL.Path == "/api/runtime/setup-status" || r.URL.Path == "/api/openapi.yaml" || r.URL.Path == "/api/auth/provider-options" {
		return false
	}
	if companionSelfAuthenticatedPath(r.URL.Path) {
		return false
	}
	return !strings.HasPrefix(r.URL.Path, "/api/auth/zitadel/")
}

func requiresZitadelAdminRole(r *http.Request) bool {
	return r.URL.Path == "/api/users" || strings.HasPrefix(r.URL.Path, "/api/users/")
}

func (boundary *zitadelAuthBoundary) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeZitadelError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	if err := boundary.validateConfig(); err != nil {
		writeZitadelError(w, http.StatusServiceUnavailable, "identity_unavailable")
		return
	}
	discovery, err := boundary.discover(r.Context())
	if err != nil {
		writeZitadelError(w, http.StatusServiceUnavailable, "identity_unavailable")
		return
	}
	state, err := secureRandomValue(32)
	if err != nil {
		writeZitadelError(w, http.StatusInternalServerError, "identity_start_failed")
		return
	}
	nonce, err := secureRandomValue(32)
	if err != nil {
		writeZitadelError(w, http.StatusInternalServerError, "identity_start_failed")
		return
	}
	verifier, err := secureRandomValue(64)
	if err != nil {
		writeZitadelError(w, http.StatusInternalServerError, "identity_start_failed")
		return
	}
	transaction := zitadelLoginTransaction{
		State:        state,
		Nonce:        nonce,
		CodeVerifier: verifier,
		ReturnTo:     safeCabinetReturnTo(r.URL.Query().Get("return_to")),
		ExpiresAt:    boundary.now().Add(zitadelLoginTTL),
	}
	boundary.mu.Lock()
	boundary.pruneLocked()
	boundary.transactions[state] = transaction
	boundary.mu.Unlock()
	boundary.setCookie(w, zitadelTransactionCookieName, state, transaction.ExpiresAt, "/api/auth/zitadel")

	authorizeURL, err := url.Parse(discovery.AuthorizationEndpoint)
	if err != nil {
		writeZitadelError(w, http.StatusServiceUnavailable, "identity_unavailable")
		return
	}
	challengeBytes := sha256.Sum256([]byte(verifier))
	query := authorizeURL.Query()
	query.Set("client_id", boundary.config.ClientID)
	query.Set("redirect_uri", boundary.callbackURL())
	query.Set("response_type", "code")
	query.Set("scope", strings.Join(boundary.config.Scopes, " "))
	query.Set("state", state)
	query.Set("nonce", nonce)
	query.Set("code_challenge", base64.RawURLEncoding.EncodeToString(challengeBytes[:]))
	query.Set("code_challenge_method", "S256")
	switch strings.ToLower(strings.TrimSpace(r.URL.Query().Get("intent"))) {
	case "register":
		query.Set("prompt", "create")
	case "recover":
		query.Set("prompt", "login")
	}
	authorizeURL.RawQuery = query.Encode()
	http.Redirect(w, r, authorizeURL.String(), http.StatusFound)
}

func (boundary *zitadelAuthBoundary) handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeZitadelError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	if err := boundary.validateConfig(); err != nil {
		boundary.redirectAuthError(w, r, "identity_unavailable")
		return
	}
	if r.URL.Query().Get("error") != "" {
		boundary.redirectAuthError(w, r, "identity_denied")
		return
	}
	state := strings.TrimSpace(r.URL.Query().Get("state"))
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	cookie, err := r.Cookie(zitadelTransactionCookieName)
	if err != nil || state == "" || code == "" || cookie.Value != state {
		boundary.redirectAuthError(w, r, "invalid_identity_callback")
		return
	}
	boundary.mu.Lock()
	transaction, ok := boundary.transactions[state]
	delete(boundary.transactions, state)
	boundary.mu.Unlock()
	boundary.clearCookie(w, zitadelTransactionCookieName, "/api/auth/zitadel")
	if !ok || transaction.ExpiresAt.Before(boundary.now()) {
		boundary.redirectAuthError(w, r, "invalid_identity_callback")
		return
	}

	discovery, err := boundary.discover(r.Context())
	if err != nil {
		boundary.redirectAuthError(w, r, "identity_unavailable")
		return
	}
	tokens, err := boundary.exchangeCode(r.Context(), discovery, code, transaction.CodeVerifier)
	if err != nil || tokens.IDToken == "" {
		boundary.redirectAuthError(w, r, "invalid_identity_token")
		return
	}
	identity, err := boundary.validateIDToken(r.Context(), discovery, tokens.IDToken, transaction.Nonce)
	if err != nil {
		boundary.redirectAuthError(w, r, "invalid_identity_token")
		return
	}
	if tokens.AccessToken != "" {
		identity = boundary.enrichIdentity(r.Context(), discovery, tokens.AccessToken, identity)
	}
	expiresAt := identity.ExpiresAt
	if tokens.ExpiresIn > 0 {
		accessExpiry := boundary.now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
		if expiresAt.IsZero() || accessExpiry.Before(expiresAt) {
			expiresAt = accessExpiry
		}
	}
	if expiresAt.IsZero() || expiresAt.After(boundary.now().Add(zitadelDefaultSessionTTL)) {
		expiresAt = boundary.now().Add(zitadelDefaultSessionTTL)
	}
	sessionID, err := secureRandomValue(48)
	if err != nil {
		boundary.redirectAuthError(w, r, "identity_session_failed")
		return
	}
	boundary.mu.Lock()
	boundary.pruneLocked()
	boundary.sessions[sessionID] = zitadelSession{
		Identity: identity, AccessToken: tokens.AccessToken, IDToken: tokens.IDToken,
		RefreshToken: tokens.RefreshToken, ExpiresAt: expiresAt,
	}
	boundary.mu.Unlock()
	boundary.setCookie(w, zitadelSessionCookieName, sessionID, expiresAt, "/")
	http.Redirect(w, r, transaction.ReturnTo, http.StatusFound)
}

func (boundary *zitadelAuthBoundary) handleSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeZitadelError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	session, err := boundary.validateZitadelRequestSession(r)
	if err != nil {
		writeZitadelError(w, http.StatusUnauthorized, "authentication_required")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"authenticated": true,
		"identity_mode": "zitadel",
		"user": map[string]any{
			"subject": session.Identity.Subject,
			"email": session.Identity.Email,
			"name": session.Identity.Name,
			"roles": session.Identity.Roles,
		},
		"expires_at": session.ExpiresAt.UTC().Format(time.RFC3339),
	})
}

func (boundary *zitadelAuthBoundary) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeZitadelError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	if !boundary.allowsStateChangingRequest(r) {
		writeZitadelError(w, http.StatusForbidden, "origin_denied")
		return
	}
	sessionID, session, ok := boundary.requestSession(r, false)
	if !ok || session.RefreshToken == "" {
		writeZitadelError(w, http.StatusUnauthorized, "refresh_unavailable")
		return
	}
	discovery, err := boundary.discover(r.Context())
	if err != nil {
		writeZitadelError(w, http.StatusServiceUnavailable, "identity_unavailable")
		return
	}
	tokens, err := boundary.exchangeRefreshToken(r.Context(), discovery, session.RefreshToken)
	if err != nil {
		boundary.deleteSession(sessionID)
		boundary.clearCookie(w, zitadelSessionCookieName, "/")
		writeZitadelError(w, http.StatusUnauthorized, "refresh_failed")
		return
	}
	if tokens.IDToken != "" {
		identity, validateErr := boundary.validateIDToken(r.Context(), discovery, tokens.IDToken, "")
		if validateErr != nil || identity.Subject != session.Identity.Subject {
			boundary.deleteSession(sessionID)
			boundary.clearCookie(w, zitadelSessionCookieName, "/")
			writeZitadelError(w, http.StatusUnauthorized, "refresh_failed")
			return
		}
		session.Identity = identity
		session.IDToken = tokens.IDToken
	}
	if tokens.AccessToken != "" {
		session.AccessToken = tokens.AccessToken
	}
	if tokens.RefreshToken != "" {
		session.RefreshToken = tokens.RefreshToken
	}
	session.ExpiresAt = boundary.now().Add(zitadelDefaultSessionTTL)
	if tokens.ExpiresIn > 0 {
		session.ExpiresAt = boundary.now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
	}
	boundary.mu.Lock()
	boundary.sessions[sessionID] = session
	boundary.mu.Unlock()
	boundary.setCookie(w, zitadelSessionCookieName, sessionID, session.ExpiresAt, "/")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"refreshed": true, "expires_at": session.ExpiresAt.UTC().Format(time.RFC3339)})
}

func (boundary *zitadelAuthBoundary) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeZitadelError(w, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	if !boundary.allowsStateChangingRequest(r) {
		writeZitadelError(w, http.StatusForbidden, "origin_denied")
		return
	}
	sessionID, _, _ := boundary.requestSession(r, false)
	if sessionID != "" {
		boundary.deleteSession(sessionID)
	}
	boundary.clearCookie(w, zitadelSessionCookieName, "/")
	logoutURL := boundary.config.PublicOrigin + "/sign-in"
	if discovery, err := boundary.discover(r.Context()); err == nil && discovery.EndSessionEndpoint != "" {
		if parsed, parseErr := url.Parse(discovery.EndSessionEndpoint); parseErr == nil {
			query := parsed.Query()
			query.Set("client_id", boundary.config.ClientID)
			query.Set("post_logout_redirect_uri", logoutURL)
			parsed.RawQuery = query.Encode()
			logoutURL = parsed.String()
		}
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"logged_out": true, "provider_logout_url": logoutURL})
}

func (boundary *zitadelAuthBoundary) validateZitadelRequestSession(r *http.Request) (zitadelSession, error) {
	_, session, ok := boundary.requestSession(r, true)
	if !ok {
		return zitadelSession{}, errors.New("valid Cabinet application session required")
	}
	return session, nil
}

func (boundary *zitadelAuthBoundary) requestSession(r *http.Request, requireCurrent bool) (string, zitadelSession, bool) {
	if !boundary.enabled() {
		return "", zitadelSession{}, false
	}
	cookie, err := r.Cookie(zitadelSessionCookieName)
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		return "", zitadelSession{}, false
	}
	boundary.mu.Lock()
	session, ok := boundary.sessions[cookie.Value]
	if requireCurrent && ok && !session.ExpiresAt.After(boundary.now()) {
		delete(boundary.sessions, cookie.Value)
		ok = false
	}
	boundary.mu.Unlock()
	return cookie.Value, session, ok
}

func (boundary *zitadelAuthBoundary) discover(ctx context.Context) (oidcDiscoveryDocument, error) {
	if err := boundary.validateConfig(); err != nil {
		return oidcDiscoveryDocument{}, err
	}
	var document oidcDiscoveryDocument
	if err := boundary.getJSON(ctx, boundary.config.Issuer+"/.well-known/openid-configuration", "", &document); err != nil {
		return document, err
	}
	if strings.TrimRight(document.Issuer, "/") != boundary.config.Issuer {
		return document, errors.New("OIDC discovery issuer mismatch")
	}
	if document.AuthorizationEndpoint == "" || document.TokenEndpoint == "" || document.JWKSURI == "" {
		return document, errors.New("OIDC discovery document is incomplete")
	}
	for _, endpoint := range []string{document.AuthorizationEndpoint, document.TokenEndpoint, document.JWKSURI, document.UserInfoEndpoint, document.EndSessionEndpoint} {
		if endpoint == "" {
			continue
		}
		if err := boundary.validateProviderEndpoint(endpoint); err != nil {
			return document, err
		}
	}
	return document, nil
}

func (boundary *zitadelAuthBoundary) validateProviderEndpoint(raw string) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" || parsed.User != nil {
		return errors.New("OIDC discovery endpoint is invalid")
	}
	issuer, _ := url.Parse(boundary.config.Issuer)
	if !strings.EqualFold(parsed.Host, issuer.Host) {
		return errors.New("OIDC discovery endpoint host mismatch")
	}
	if !boundary.config.AllowInsecureHTTP && parsed.Scheme != "https" {
		return errors.New("OIDC discovery endpoint must use HTTPS")
	}
	if boundary.config.AllowInsecureHTTP && parsed.Scheme != "https" && parsed.Scheme != "http" {
		return errors.New("OIDC discovery endpoint scheme is invalid")
	}
	return nil
}

func (boundary *zitadelAuthBoundary) allowsStateChangingRequest(r *http.Request) bool {
	origin := strings.TrimRight(strings.TrimSpace(r.Header.Get("Origin")), "/")
	if origin == "" {
		return true
	}
	return origin == boundary.config.PublicOrigin
}

func (boundary *zitadelAuthBoundary) exchangeCode(ctx context.Context, discovery oidcDiscoveryDocument, code, verifier string) (oidcTokenResponse, error) {
	values := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {boundary.config.ClientID},
		"code":          {code},
		"redirect_uri":  {boundary.callbackURL()},
		"code_verifier": {verifier},
	}
	return boundary.exchangeTokens(ctx, discovery.TokenEndpoint, values)
}

func (boundary *zitadelAuthBoundary) exchangeRefreshToken(ctx context.Context, discovery oidcDiscoveryDocument, refreshToken string) (oidcTokenResponse, error) {
	values := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {boundary.config.ClientID},
		"refresh_token": {refreshToken},
	}
	return boundary.exchangeTokens(ctx, discovery.TokenEndpoint, values)
}

func (boundary *zitadelAuthBoundary) exchangeTokens(ctx context.Context, endpoint string, values url.Values) (oidcTokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return oidcTokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	if boundary.config.ClientSecret != "" {
		req.SetBasicAuth(boundary.config.ClientID, boundary.config.ClientSecret)
	}
	resp, err := boundary.client.Do(req)
	if err != nil {
		return oidcTokenResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return oidcTokenResponse{}, fmt.Errorf("OIDC token endpoint returned %d", resp.StatusCode)
	}
	var tokens oidcTokenResponse
	if err := decodeLimitedJSON(resp.Body, &tokens); err != nil {
		return tokens, err
	}
	return tokens, nil
}

func (boundary *zitadelAuthBoundary) validateIDToken(ctx context.Context, discovery oidcDiscoveryDocument, rawToken, expectedNonce string) (zitadelIdentity, error) {
	keys, err := boundary.fetchSigningKeys(ctx, discovery.JWKSURI)
	if err != nil {
		return zitadelIdentity{}, err
	}
	token, err := jwt.Parse(rawToken, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, errors.New("unexpected ID token signing algorithm")
		}
		kid, _ := token.Header["kid"].(string)
		key, ok := keys[kid]
		if !ok {
			return nil, errors.New("unknown ID token signing key")
		}
		return key, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}), jwt.WithIssuer(boundary.config.Issuer), jwt.WithExpirationRequired(), jwt.WithLeeway(30*time.Second))
	if err != nil || !token.Valid {
		return zitadelIdentity{}, errors.New("invalid ID token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return zitadelIdentity{}, errors.New("invalid ID token claims")
	}
	audience, err := claims.GetAudience()
	if err != nil || !claimStringsContain(audience, boundary.config.ClientID) || !claimStringsContain(audience, boundary.config.Audience) {
		return zitadelIdentity{}, errors.New("ID token audience mismatch")
	}
	if azp := strings.TrimSpace(claimAsString(map[string]any(claims), "azp")); azp != boundary.config.ClientID {
		return zitadelIdentity{}, errors.New("ID token authorised party mismatch")
	}
	if expectedNonce != "" && claimAsString(map[string]any(claims), "nonce") != expectedNonce {
		return zitadelIdentity{}, errors.New("ID token nonce mismatch")
	}
	subject, err := claims.GetSubject()
	if err != nil || strings.TrimSpace(subject) == "" {
		return zitadelIdentity{}, errors.New("ID token subject missing")
	}
	expiration, err := claims.GetExpirationTime()
	if err != nil || expiration == nil || !expiration.After(boundary.now().Add(-30*time.Second)) {
		return zitadelIdentity{}, errors.New("ID token expired")
	}
	roles := extractZitadelRoles(map[string]any(claims))
	if !hasAnyRole(roles, boundary.config.RequiredRoles) {
		return zitadelIdentity{}, errors.New("required Cabinet role missing")
	}
	return zitadelIdentity{
		Subject: subject,
		Email: claimAsString(map[string]any(claims), "email"),
		Name: claimAsString(map[string]any(claims), "name"),
		Roles: roles,
		ExpiresAt: expiration.Time,
	}, nil
}

func (boundary *zitadelAuthBoundary) fetchSigningKeys(ctx context.Context, jwksURI string) (map[string]*rsa.PublicKey, error) {
	var payload struct {
		Keys []struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			Alg string `json:"alg"`
			Use string `json:"use"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := boundary.getJSON(ctx, jwksURI, "", &payload); err != nil {
		return nil, err
	}
	keys := make(map[string]*rsa.PublicKey)
	for _, encoded := range payload.Keys {
		if encoded.Kty != "RSA" || encoded.Kid == "" || (encoded.Alg != "" && encoded.Alg != "RS256") || (encoded.Use != "" && encoded.Use != "sig") {
			continue
		}
		nBytes, nErr := base64.RawURLEncoding.DecodeString(encoded.N)
		eBytes, eErr := base64.RawURLEncoding.DecodeString(encoded.E)
		if nErr != nil || eErr != nil || len(nBytes) == 0 || len(eBytes) == 0 {
			continue
		}
		exponent := 0
		for _, value := range eBytes {
			exponent = exponent<<8 + int(value)
		}
		if exponent < 3 {
			continue
		}
		keys[encoded.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: exponent}
	}
	if len(keys) == 0 {
		return nil, errors.New("OIDC signing keys unavailable")
	}
	return keys, nil
}

func (boundary *zitadelAuthBoundary) enrichIdentity(ctx context.Context, discovery oidcDiscoveryDocument, accessToken string, identity zitadelIdentity) zitadelIdentity {
	if discovery.UserInfoEndpoint == "" {
		return identity
	}
	var claims map[string]any
	if err := boundary.getJSON(ctx, discovery.UserInfoEndpoint, accessToken, &claims); err != nil {
		return identity
	}
	if claimAsString(claims, "sub") != identity.Subject {
		return identity
	}
	if email := strings.TrimSpace(claimAsString(claims, "email")); email != "" {
		identity.Email = email
	}
	if name := strings.TrimSpace(claimAsString(claims, "name")); name != "" {
		identity.Name = name
	}
	return identity
}

func (boundary *zitadelAuthBoundary) getJSON(ctx context.Context, endpoint, bearer string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := boundary.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("OIDC endpoint returned %d", resp.StatusCode)
	}
	return decodeLimitedJSON(resp.Body, target)
}

func decodeLimitedJSON(reader io.Reader, target any) error {
	decoder := json.NewDecoder(io.LimitReader(reader, 1<<20))
	return decoder.Decode(target)
}

func (boundary *zitadelAuthBoundary) callbackURL() string {
	return boundary.config.PublicOrigin + "/api/auth/zitadel/callback"
}

func (boundary *zitadelAuthBoundary) setCookie(w http.ResponseWriter, name, value string, expires time.Time, path string) {
	secure := strings.HasPrefix(strings.ToLower(boundary.config.PublicOrigin), "https://")
	http.SetCookie(w, &http.Cookie{
		Name: name, Value: value, Path: path, Expires: expires, MaxAge: int(expires.Sub(boundary.now()).Seconds()),
		HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode,
	})
}

func (boundary *zitadelAuthBoundary) clearCookie(w http.ResponseWriter, name, path string) {
	secure := strings.HasPrefix(strings.ToLower(boundary.config.PublicOrigin), "https://")
	http.SetCookie(w, &http.Cookie{
		Name: name, Value: "", Path: path, Expires: time.Unix(1, 0), MaxAge: -1,
		HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode,
	})
}

func (boundary *zitadelAuthBoundary) redirectAuthError(w http.ResponseWriter, r *http.Request, code string) {
	destination := &url.URL{Path: "/sign-in"}
	if configured, err := url.Parse(boundary.config.PublicOrigin + "/sign-in"); err == nil && configured != nil {
		destination = configured
	}
	query := destination.Query()
	query.Set("auth_error", code)
	destination.RawQuery = query.Encode()
	http.Redirect(w, r, destination.String(), http.StatusFound)
}

func (boundary *zitadelAuthBoundary) pruneLocked() {
	now := boundary.now()
	for key, transaction := range boundary.transactions {
		if !transaction.ExpiresAt.After(now) {
			delete(boundary.transactions, key)
		}
	}
	for key, session := range boundary.sessions {
		if session.ExpiresAt.Add(24 * time.Hour).Before(now) {
			delete(boundary.sessions, key)
		}
	}
}

func (boundary *zitadelAuthBoundary) deleteSession(sessionID string) {
	boundary.mu.Lock()
	delete(boundary.sessions, sessionID)
	boundary.mu.Unlock()
}

func extractZitadelRoles(claims map[string]any) []string {
	roles := make(map[string]struct{})
	for key, value := range claims {
		if key != zitadelRoleClaim && !(strings.HasPrefix(key, "urn:zitadel:iam:org:project:") && strings.HasSuffix(key, ":roles")) {
			continue
		}
		extractRoleKeys(value, roles)
	}
	result := make([]string, 0, len(roles))
	for role := range roles {
		result = append(result, role)
	}
	sort.Strings(result)
	return result
}

func extractRoleKeys(value any, roles map[string]struct{}) {
	switch typed := value.(type) {
	case map[string]any:
		for key := range typed {
			key = strings.TrimSpace(strings.ToLower(key))
			if key != "" {
				roles[key] = struct{}{}
			}
		}
	case []any:
		for _, entry := range typed {
			extractRoleKeys(entry, roles)
		}
	case []string:
		for _, role := range typed {
			role = strings.TrimSpace(strings.ToLower(role))
			if role != "" {
				roles[role] = struct{}{}
			}
		}
	case string:
		for _, role := range splitConfiguredValues(typed) {
			roles[strings.ToLower(role)] = struct{}{}
		}
	}
}

func hasAnyRole(actual, required []string) bool {
	if len(required) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(actual))
	for _, role := range actual {
		set[strings.ToLower(strings.TrimSpace(role))] = struct{}{}
	}
	for _, role := range required {
		if _, ok := set[strings.ToLower(strings.TrimSpace(role))]; ok {
			return true
		}
	}
	return false
}

func claimStringsContain(values jwt.ClaimStrings, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func secureRandomValue(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func safeCabinetReturnTo(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "/dashboard"
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.IsAbs() || parsed.Host != "" || !strings.HasPrefix(parsed.Path, "/") || strings.HasPrefix(parsed.Path, "//") {
		return "/dashboard"
	}
	return parsed.String()
}

func splitConfiguredValues(raw string) []string {
	return strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == ' ' || r == '\n' || r == '\t' })
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func envEnabled(key string) bool {
	value, _ := strconv.ParseBool(strings.TrimSpace(os.Getenv(key)))
	return value
}

func writeZitadelError(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code})
}

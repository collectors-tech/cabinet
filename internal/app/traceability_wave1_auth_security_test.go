package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestWave1AuthRequirementsReflectCredentialState(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wave1Auth"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	before := doRequest(t, a, http.MethodGet, "/api/auth/requirements?profile_id="+p.ID, nil, nil)
	if before.Code != http.StatusOK {
		t.Fatalf("requirements before status=%d body=%s", before.Code, before.Body.String())
	}
	var beforePayload map[string]any
	if err := json.NewDecoder(before.Body).Decode(&beforePayload); err != nil {
		t.Fatalf("decode requirements before: %v", err)
	}
	if beforePayload["requires_registration"] != true {
		t.Fatalf("expected requires_registration=true before credential, got %+v", beforePayload)
	}

	if _, err := a.db.Exec(`INSERT INTO webauthn_credentials(id, profile_id, credential_json) VALUES('cred-wave1-auth', ?, '{}')`, p.ID); err != nil {
		t.Fatalf("seed credential: %v", err)
	}

	after := doRequest(t, a, http.MethodGet, "/api/auth/requirements?profile_id="+p.ID, nil, nil)
	if after.Code != http.StatusOK {
		t.Fatalf("requirements after status=%d body=%s", after.Code, after.Body.String())
	}
	var afterPayload map[string]any
	if err := json.NewDecoder(after.Body).Decode(&afterPayload); err != nil {
		t.Fatalf("decode requirements after: %v", err)
	}
	if afterPayload["requires_registration"] != false {
		t.Fatalf("expected requires_registration=false after credential, got %+v", afterPayload)
	}
}

func TestWave1AuthRecoveryResetBeginSuccess(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wave1Recovery"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	setRecovery := doRequest(t, a, http.MethodPost, "/api/auth/recovery/passphrase", strings.NewReader(`{"profile_id":"`+p.ID+`","passphrase":"hunter2"}`), map[string]string{"Content-Type": "application/json"})
	if setRecovery.Code != http.StatusOK {
		t.Fatalf("set recovery status=%d body=%s", setRecovery.Code, setRecovery.Body.String())
	}

	begin := doRequest(t, a, http.MethodPost, "/api/auth/recovery/reset/begin", strings.NewReader(`{"profile_id":"`+p.ID+`","passphrase":"hunter2"}`), map[string]string{"Content-Type": "application/json"})
	if begin.Code != http.StatusOK {
		t.Fatalf("begin recovery reset status=%d body=%s", begin.Code, begin.Body.String())
	}

	var payload map[string]any
	if err := json.NewDecoder(begin.Body).Decode(&payload); err != nil {
		t.Fatalf("decode recovery reset payload: %v", err)
	}
	if strings.TrimSpace(stringifyAny(payload["session_id"])) == "" {
		t.Fatalf("expected non-empty session_id in recovery reset payload: %+v", payload)
	}
}

func TestWave1ErrorEnvelopeContracts(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	invalidMethod := doRequest(t, a, http.MethodPost, "/api/auth/requirements", strings.NewReader(`{}`), map[string]string{"Content-Type": "application/json"})
	if invalidMethod.Code != http.StatusMethodNotAllowed {
		t.Fatalf("auth requirements invalid method expected 405, got %d body=%s", invalidMethod.Code, invalidMethod.Body.String())
	}
	if !strings.Contains(invalidMethod.Body.String(), `"error":"method_not_allowed"`) {
		t.Fatalf("expected method_not_allowed envelope, got %s", invalidMethod.Body.String())
	}

	invalidJSON := doRequest(t, a, http.MethodPost, "/api/auth/session/validate", strings.NewReader(`{"session_token"`), map[string]string{"Content-Type": "application/json"})
	if invalidJSON.Code != http.StatusBadRequest {
		t.Fatalf("invalid json expected 400, got %d body=%s", invalidJSON.Code, invalidJSON.Body.String())
	}
	if !strings.Contains(invalidJSON.Body.String(), `"error":"invalid_json"`) {
		t.Fatalf("expected invalid_json envelope, got %s", invalidJSON.Body.String())
	}

	scannerInvalid := doRequest(t, a, http.MethodPost, "/api/scanner/failures/retry", strings.NewReader(`{"query_set_id":"missing"}`), map[string]string{"Content-Type": "application/json"})
	if scannerInvalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid scanner retry expected 400, got %d body=%s", scannerInvalid.Code, scannerInvalid.Body.String())
	}
	if !strings.Contains(scannerInvalid.Body.String(), `"error":"invalid_query_set_id"`) {
		t.Fatalf("expected invalid_query_set_id envelope, got %s", scannerInvalid.Body.String())
	}
}

func TestWave1CloudBootstrapContractAndInvalidTokenHandling(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	validToken := "e30.eyJzdWIiOiJ1c2VyX3dhdmUxIiwiZW1haWwiOiJ3YXZlMUBleGFtcGxlLmNvbSIsInBsYW4iOiJwcm8ifQ.e30"
	valid := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/auth/cloud/session/bootstrap",
		strings.NewReader(`{"provider":"zitadel","token":"`+validToken+`"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if valid.Code != http.StatusOK {
		t.Fatalf("bootstrap valid status=%d body=%s", valid.Code, valid.Body.String())
	}
	for _, needle := range []string{`"provider":"zitadel"`, `"user_id":"user_wave1"`, `"plan":"pro"`, `"features"`} {
		if !strings.Contains(valid.Body.String(), needle) {
			t.Fatalf("expected %q in bootstrap body: %s", needle, valid.Body.String())
		}
	}
	if !strings.Contains(valid.Body.String(), `"entitlement_source":"billing"`) {
		t.Fatalf("expected entitlement_source=billing in bootstrap body: %s", valid.Body.String())
	}

	invalid := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/auth/cloud/session/bootstrap",
		strings.NewReader(`{"provider":"zitadel","token":"not-a-jwt"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if invalid.Code != http.StatusUnauthorized {
		t.Fatalf("bootstrap invalid token expected 401, got %d body=%s", invalid.Code, invalid.Body.String())
	}
	if !strings.Contains(invalid.Body.String(), `"error":"invalid_token"`) {
		t.Fatalf("expected invalid_token envelope, got %s", invalid.Body.String())
	}
}

func TestWave1RuntimeHealthAndRecoveryContracts(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	healthz := doRequest(t, a, http.MethodGet, "/healthz", nil, nil)
	if healthz.Code != http.StatusOK {
		t.Fatalf("healthz status=%d body=%s", healthz.Code, healthz.Body.String())
	}
	if strings.TrimSpace(healthz.Body.String()) != "ok" {
		t.Fatalf("healthz expected body=ok, got %q", strings.TrimSpace(healthz.Body.String()))
	}

	runtimeResp := doRequest(t, a, http.MethodGet, "/api/runtime", nil, nil)
	if runtimeResp.Code != http.StatusOK {
		t.Fatalf("runtime status=%d body=%s", runtimeResp.Code, runtimeResp.Body.String())
	}
	var runtimePayload map[string]any
	if err := json.NewDecoder(runtimeResp.Body).Decode(&runtimePayload); err != nil {
		t.Fatalf("decode runtime payload: %v", err)
	}
	for _, field := range []string{"app_version", "build_date", "update_channel"} {
		if _, ok := runtimePayload[field]; !ok {
			t.Fatalf("runtime payload missing %q: %+v", field, runtimePayload)
		}
	}

	recovery := doRequest(t, a, http.MethodGet, "/api/runtime/recovery", nil, nil)
	if recovery.Code != http.StatusOK {
		t.Fatalf("recovery status=%d body=%s", recovery.Code, recovery.Body.String())
	}
	var recoveryPayload map[string]any
	if err := json.NewDecoder(recovery.Body).Decode(&recoveryPayload); err != nil {
		t.Fatalf("decode recovery payload: %v", err)
	}
	if _, ok := recoveryPayload["recovery_required"]; !ok {
		t.Fatalf("expected recovery_required key in payload: %+v", recoveryPayload)
	}
}

func stringifyAny(v any) string {
	switch x := v.(type) {
	case string:
		return x
	default:
		return ""
	}
}

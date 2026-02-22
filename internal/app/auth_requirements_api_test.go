package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestAuthRequirementsAndRecoveryPassphraseEndpoints(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"P1"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	reqs := doRequest(t, a, http.MethodGet, "/api/auth/requirements?profile_id="+p.ID, nil, nil)
	if reqs.Code != http.StatusOK {
		t.Fatalf("requirements status=%d body=%s", reqs.Code, reqs.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(reqs.Body).Decode(&payload); err != nil {
		t.Fatalf("decode requirements: %v", err)
	}
	if payload["requires_registration"] != true {
		t.Fatalf("expected requires_registration=true, got %+v", payload)
	}

	setRecovery := doRequest(t, a, http.MethodPost, "/api/auth/recovery/passphrase", strings.NewReader(`{"profile_id":"`+p.ID+`","passphrase":"hunter2"}`), map[string]string{"Content-Type": "application/json"})
	if setRecovery.Code != http.StatusOK {
		t.Fatalf("set recovery status=%d body=%s", setRecovery.Code, setRecovery.Body.String())
	}
	resetBeginWrong := doRequest(t, a, http.MethodPost, "/api/auth/recovery/reset/begin", strings.NewReader(`{"profile_id":"`+p.ID+`","passphrase":"wrong"}`), map[string]string{"Content-Type": "application/json"})
	if resetBeginWrong.Code != http.StatusBadRequest {
		t.Fatalf("reset begin wrong passphrase expected 400, got %d", resetBeginWrong.Code)
	}

	validateMissing := doRequest(t, a, http.MethodPost, "/api/auth/session/validate", strings.NewReader(`{"session_token":"missing"}`), map[string]string{"Content-Type": "application/json"})
	if validateMissing.Code != http.StatusUnauthorized {
		t.Fatalf("validate missing session expected 401, got %d", validateMissing.Code)
	}
}

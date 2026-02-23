package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestProtectedWriteBlockedWhenRegisteredProfileIsLocked(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Locked Profile"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}

	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	activate := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if activate.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", activate.Code, activate.Body.String())
	}

	_, err := a.db.Exec(`INSERT INTO webauthn_credentials(id, profile_id, credential_json) VALUES('cred-lock-1', ?, '{}')`, profile.ID)
	if err != nil {
		t.Fatalf("seed credential: %v", err)
	}

	createItem := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"LOCK-001","title":"Blocked Item","brand":"AFX","category":"General"}`), map[string]string{"Content-Type": "application/json"})
	if createItem.Code != http.StatusLocked {
		t.Fatalf("create item while locked expected 423, got %d body=%s", createItem.Code, createItem.Body.String())
	}
}

func TestProtectedWriteAllowedWhenRegisteredProfileHasUnlockedSession(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Unlocked Profile"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}

	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	activate := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if activate.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", activate.Code, activate.Body.String())
	}

	_, err := a.db.Exec(`INSERT INTO webauthn_credentials(id, profile_id, credential_json) VALUES('cred-lock-2', ?, '{}')`, profile.ID)
	if err != nil {
		t.Fatalf("seed credential: %v", err)
	}
	if _, err := a.authService.CreateUnlockedSession(profile.ID); err != nil {
		t.Fatalf("create unlocked session: %v", err)
	}

	createItem := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"LOCK-002","title":"Allowed Item","brand":"AFX","category":"General"}`), map[string]string{"Content-Type": "application/json"})
	if createItem.Code != http.StatusCreated {
		t.Fatalf("create item with unlocked session expected 201, got %d body=%s", createItem.Code, createItem.Body.String())
	}
}

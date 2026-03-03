package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestUsersAPIListCreateInviteUpdateDelete(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Users Profile"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if strings.TrimSpace(profile.ID) == "" {
		t.Fatal("expected profile id")
	}
	setActive := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if setActive.Code != http.StatusOK {
		t.Fatalf("set active profile status=%d body=%s", setActive.Code, setActive.Body.String())
	}

	list := doRequest(t, a, http.MethodGet, "/api/users", nil, nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list users status=%d body=%s", list.Code, list.Body.String())
	}
	var listPayload struct {
		Users []runtimeUser `json:"users"`
	}
	if err := json.NewDecoder(list.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode users payload: %v", err)
	}
	if len(listPayload.Users) == 0 {
		t.Fatalf("expected default owner user")
	}

	createUser := doRequest(t, a, http.MethodPost, "/api/users", strings.NewReader(`{"firstName":"Test","lastName":"User","username":"api_test_user","email":"api_test_user@example.com","phoneNumber":"+6100000000","role":"admin"}`), map[string]string{"Content-Type": "application/json"})
	if createUser.Code != http.StatusCreated {
		t.Fatalf("create user status=%d body=%s", createUser.Code, createUser.Body.String())
	}
	var created runtimeUser
	if err := json.NewDecoder(createUser.Body).Decode(&created); err != nil {
		t.Fatalf("decode created user: %v", err)
	}
	if created.Role != "admin" {
		t.Fatalf("expected role admin, got %s", created.Role)
	}

	invite := doRequest(t, a, http.MethodPost, "/api/users/invite", strings.NewReader(`{"email":"invite_test_user@example.com","role":"view","desc":"hello"}`), map[string]string{"Content-Type": "application/json"})
	if invite.Code != http.StatusCreated {
		t.Fatalf("invite user status=%d body=%s", invite.Code, invite.Body.String())
	}
	var invited runtimeUser
	if err := json.NewDecoder(invite.Body).Decode(&invited); err != nil {
		t.Fatalf("decode invited user: %v", err)
	}
	if invited.Status != "invited" {
		t.Fatalf("expected invited status, got %s", invited.Status)
	}

	update := doRequest(t, a, http.MethodPut, "/api/users/"+created.ID, strings.NewReader(`{"role":"view","status":"inactive"}`), map[string]string{"Content-Type": "application/json"})
	if update.Code != http.StatusOK {
		t.Fatalf("update user status=%d body=%s", update.Code, update.Body.String())
	}
	var updated runtimeUser
	if err := json.NewDecoder(update.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated user: %v", err)
	}
	if updated.Role != "view" || updated.Status != "inactive" {
		t.Fatalf("expected role=view,status=inactive got role=%s status=%s", updated.Role, updated.Status)
	}

	remove := doRequest(t, a, http.MethodDelete, "/api/users/"+created.ID, nil, nil)
	if remove.Code != http.StatusNoContent {
		t.Fatalf("delete user status=%d body=%s", remove.Code, remove.Body.String())
	}
}

func TestUsersAPIListWithoutActiveProfileFallsBackToDefaultScope(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	list := doRequest(t, a, http.MethodGet, "/api/users", nil, nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list users without active profile status=%d body=%s", list.Code, list.Body.String())
	}
	var payload struct {
		Users []runtimeUser `json:"users"`
	}
	if err := json.NewDecoder(list.Body).Decode(&payload); err != nil {
		t.Fatalf("decode users payload: %v", err)
	}
	if len(payload.Users) == 0 {
		t.Fatal("expected default scoped owner user when active profile is missing")
	}
}

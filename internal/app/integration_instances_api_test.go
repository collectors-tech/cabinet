package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestIntegrationInstancesPersistStateAndRedactSecrets(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileID := createIntegrationInstanceProfile(t, a, "Integration Instance Persistence")

	create := doRequest(t, a, http.MethodPost, "/api/profiles/"+profileID+"/integration-instances", strings.NewReader(`{
		"provider_id":"openai",
		"display_name":"OpenAI workspace",
		"enabled":true,
		"config":{"assistant_default_model":"gpt-4o-mini","openai_api_key":"must-not-persist-in-config"},
		"secrets":{"openai_api_key":"sk-instance-secret"},
		"auth_state":"configured",
		"health_state":"unknown",
		"required_action":{"code":"validate_provider","message":"Run provider test","workflow":"assistant.chat","guidance":"Validate credentials before workflow execution"}
	}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create instance status=%d body=%s", create.Code, create.Body.String())
	}
	if strings.Contains(create.Body.String(), "sk-instance-secret") || strings.Contains(create.Body.String(), "must-not-persist-in-config") {
		t.Fatalf("instance response leaked secret material: %s", create.Body.String())
	}
	var created struct {
		Instance map[string]any `json:"instance"`
	}
	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatalf("decode created instance: %v", err)
	}
	instanceID := fmt.Sprintf("%v", created.Instance["id"])
	if instanceID == "" {
		t.Fatalf("created instance needs stable id: %+v", created.Instance)
	}
	if created.Instance["enabled"] != true || created.Instance["provider_id"] != "openai" {
		t.Fatalf("created instance did not preserve provider/enabled state: %+v", created.Instance)
	}
	secretRefs, ok := created.Instance["secret_refs"].(map[string]any)
	if !ok || secretRefs["openai_api_key"] == "" {
		t.Fatalf("created instance should expose secret refs only, got %+v", created.Instance["secret_refs"])
	}
	requiredAction, ok := created.Instance["required_action"].(map[string]any)
	if !ok || requiredAction["code"] != "validate_provider" {
		t.Fatalf("created instance should preserve required-action state, got %+v", created.Instance["required_action"])
	}

	disable := doRequest(t, a, http.MethodPut, "/api/profiles/"+profileID+"/integration-instances", strings.NewReader(`{
		"id":"`+instanceID+`",
		"enabled":false,
		"health_state":"failed",
		"last_error":"credential validation failed",
		"required_action":{"code":"repair_credentials","message":"Reconnect provider","workflow":"assistant.chat","guidance":"Update API key and rerun health check"}
	}`), map[string]string{"Content-Type": "application/json"})
	if disable.Code != http.StatusOK {
		t.Fatalf("disable instance status=%d body=%s", disable.Code, disable.Body.String())
	}

	list := doRequest(t, a, http.MethodGet, "/api/profiles/"+profileID+"/integration-instances", nil, nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list instances status=%d body=%s", list.Code, list.Body.String())
	}
	if strings.Contains(list.Body.String(), "sk-instance-secret") || strings.Contains(list.Body.String(), "must-not-persist-in-config") {
		t.Fatalf("list response leaked secret material: %s", list.Body.String())
	}
	var listed struct {
		Instances []map[string]any `json:"instances"`
	}
	if err := json.NewDecoder(list.Body).Decode(&listed); err != nil {
		t.Fatalf("decode listed instances: %v", err)
	}
	if len(listed.Instances) != 1 {
		t.Fatalf("expected one persisted instance, got %+v", listed.Instances)
	}
	got := listed.Instances[0]
	if got["enabled"] != false || got["health_state"] != "failed" || got["last_error"] != "credential validation failed" {
		t.Fatalf("disabled health/error state not persisted: %+v", got)
	}
	if action, ok := got["required_action"].(map[string]any); !ok || action["code"] != "repair_credentials" {
		t.Fatalf("repair required-action state not persisted: %+v", got["required_action"])
	}

	deleteResp := doRequest(t, a, http.MethodDelete, "/api/profiles/"+profileID+"/integration-instances?id="+instanceID, nil, nil)
	if deleteResp.Code != http.StatusNoContent {
		t.Fatalf("delete instance status=%d body=%s", deleteResp.Code, deleteResp.Body.String())
	}
	empty := doRequest(t, a, http.MethodGet, "/api/profiles/"+profileID+"/integration-instances", nil, nil)
	if empty.Code != http.StatusOK {
		t.Fatalf("list after delete status=%d body=%s", empty.Code, empty.Body.String())
	}
	if strings.Contains(empty.Body.String(), instanceID) {
		t.Fatalf("deleted instance still listed: %s", empty.Body.String())
	}
}

func createIntegrationInstanceProfile(t *testing.T, a *App, name string) string {
	t.Helper()
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"`+name+`"}`), map[string]string{"Content-Type": "application/json"})
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
	return profile.ID
}

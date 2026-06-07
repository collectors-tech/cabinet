package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestOpenAIRegistryExposesMethodAwareSetupNeededState(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"OpenAIRegistryProfile"}`), map[string]string{"Content-Type": "application/json"})
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

	registry := doRequest(t, a, http.MethodGet, "/api/providers/registry", nil, nil)
	if registry.Code != http.StatusOK {
		t.Fatalf("registry status=%d body=%s", registry.Code, registry.Body.String())
	}
	var payload struct {
		Providers []map[string]any `json:"providers"`
	}
	if err := json.NewDecoder(registry.Body).Decode(&payload); err != nil {
		t.Fatalf("decode registry payload: %v", err)
	}
	openai := findRegistryProvider(payload.Providers, "openai")
	if openai == nil {
		t.Fatalf("expected openai provider in registry payload: %+v", payload.Providers)
	}
	if got := fmt.Sprintf("%v", openai["auth_mode"]); got != "hybrid" {
		t.Fatalf("expected hybrid auth mode, got %q", got)
	}
	if got := fmt.Sprintf("%v", openai["state"]); got != "needs_config" {
		t.Fatalf("expected setup-needed state before proof, got %q", got)
	}
	methods, ok := openai["auth_methods"].(map[string]any)
	if !ok {
		t.Fatalf("expected auth_methods map, got %#v", openai["auth_methods"])
	}
	browser, ok := methods["browser_auth"].(map[string]any)
	if !ok {
		t.Fatalf("expected browser_auth method, got %#v", methods["browser_auth"])
	}
	if got := fmt.Sprintf("%v", browser["state"]); got != "setup_needed" {
		t.Fatalf("expected browser auth setup_needed until proof, got %q", got)
	}
	if connected, _ := browser["connected"].(bool); connected {
		t.Fatalf("browser auth must not be connected without a verified artifact/callback")
	}
}

func TestOpenAIRegistryUsesPersistedActiveMethodWithoutBrowserNavigationProof(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"OpenAIConfiguredProfile"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	settings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(`{"settings":{"openai.active_auth_method":"browser_auth","openai.browser_auth_state":"pending","openai.browser_auth_artifact_present":"false"}}`), map[string]string{"Content-Type": "application/json"})
	if settings.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", settings.Code, settings.Body.String())
	}

	registry := doRequest(t, a, http.MethodGet, "/api/providers/registry", nil, nil)
	if registry.Code != http.StatusOK {
		t.Fatalf("registry status=%d body=%s", registry.Code, registry.Body.String())
	}
	var payload struct {
		Providers []map[string]any `json:"providers"`
	}
	if err := json.NewDecoder(registry.Body).Decode(&payload); err != nil {
		t.Fatalf("decode registry payload: %v", err)
	}
	openai := findRegistryProvider(payload.Providers, "openai")
	methods := openai["auth_methods"].(map[string]any)
	browser := methods["browser_auth"].(map[string]any)
	if got := fmt.Sprintf("%v", browser["state"]); got != "pending" {
		t.Fatalf("expected browser auth pending state, got %q", got)
	}
	if connected, _ := browser["connected"].(bool); connected {
		t.Fatalf("browser auth must not be connected when active method is browser_auth but proof is pending")
	}
}

func TestEbayRegistryExposesSellerOperationCapabilityStatuses(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	registry := doRequest(t, a, http.MethodGet, "/api/providers/registry", nil, nil)
	if registry.Code != http.StatusOK {
		t.Fatalf("registry status=%d body=%s", registry.Code, registry.Body.String())
	}
	var payload struct {
		Providers []map[string]any `json:"providers"`
	}
	if err := json.NewDecoder(registry.Body).Decode(&payload); err != nil {
		t.Fatalf("decode registry payload: %v", err)
	}
	ebay := findRegistryProvider(payload.Providers, "ebay")
	if ebay == nil {
		t.Fatalf("expected ebay provider in registry payload: %+v", payload.Providers)
	}
	statuses, ok := ebay["seller_operations"].([]any)
	if !ok {
		t.Fatalf("expected seller_operations array, got %#v", ebay["seller_operations"])
	}
	if len(statuses) != 5 {
		t.Fatalf("expected five seller operation statuses, got %d: %+v", len(statuses), statuses)
	}
	for _, raw := range statuses {
		status, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("expected seller operation status object, got %#v", raw)
		}
		if status["read_available"].(bool) || status["write_available"].(bool) || status["confirmation_required"].(bool) {
			t.Fatalf("default registry status must not expose unverified seller operation availability: %+v", status)
		}
		if got := fmt.Sprintf("%v", status["blocker"]); got != "ebay_api_capability_not_verified" {
			t.Fatalf("expected capability-not-verified blocker, got %+v", status)
		}
	}
}

func TestTelegramRegistryExposesCaptureChannelSetupState(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"TelegramRegistryProfile"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})

	registry := doRequest(t, a, http.MethodGet, "/api/providers/registry", nil, nil)
	if registry.Code != http.StatusOK {
		t.Fatalf("registry status=%d body=%s", registry.Code, registry.Body.String())
	}
	var payload struct {
		Providers []map[string]any `json:"providers"`
	}
	if err := json.NewDecoder(registry.Body).Decode(&payload); err != nil {
		t.Fatalf("decode registry payload: %v", err)
	}
	telegram := findRegistryProvider(payload.Providers, "telegram")
	if telegram == nil {
		t.Fatalf("expected telegram provider in registry payload: %+v", payload.Providers)
	}
	if got := fmt.Sprintf("%v", telegram["integration_mode"]); got != "assistant_capture_channel" {
		t.Fatalf("expected assistant capture channel mode, got %q", got)
	}
	if got := fmt.Sprintf("%v", telegram["state"]); got != "needs_config" {
		t.Fatalf("expected Telegram setup-needed state before sender/chat authorization, got %q", got)
	}
	capabilities, ok := telegram["capabilities"].(map[string]any)
	if !ok {
		t.Fatalf("expected capabilities map, got %#v", telegram["capabilities"])
	}
	if capabilities["assistant"] != true || capabilities["media_capture"] != true {
		t.Fatalf("expected Telegram assistant and media capture capabilities, got %+v", capabilities)
	}
	authMethods, ok := telegram["auth_methods"].(map[string]any)
	if !ok {
		t.Fatalf("expected auth_methods map, got %#v", telegram["auth_methods"])
	}
	senderChat, ok := authMethods["sender_chat"].(map[string]any)
	if !ok || senderChat["connected"].(bool) {
		t.Fatalf("expected sender/chat method to start disconnected, got %#v", authMethods["sender_chat"])
	}

	settings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(`{"settings":{"telegram.catalog_capture.sender_id":"12345","telegram.catalog_capture.chat_id":"-5235769556"}}`), map[string]string{"Content-Type": "application/json"})
	if settings.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", settings.Code, settings.Body.String())
	}
	registry = doRequest(t, a, http.MethodGet, "/api/providers/registry", nil, nil)
	if registry.Code != http.StatusOK {
		t.Fatalf("registry after settings status=%d body=%s", registry.Code, registry.Body.String())
	}
	payload = struct {
		Providers []map[string]any `json:"providers"`
	}{}
	if err := json.NewDecoder(registry.Body).Decode(&payload); err != nil {
		t.Fatalf("decode registry payload after settings: %v", err)
	}
	telegram = findRegistryProvider(payload.Providers, "telegram")
	if got := fmt.Sprintf("%v", telegram["state"]); got != "ready" {
		t.Fatalf("expected Telegram ready state after sender/chat authorization, got %q", got)
	}
	authMethods = telegram["auth_methods"].(map[string]any)
	senderChat = authMethods["sender_chat"].(map[string]any)
	if connected, _ := senderChat["connected"].(bool); !connected {
		t.Fatalf("expected sender/chat method connected after settings, got %+v", senderChat)
	}
}

func findRegistryProvider(providers []map[string]any, id string) map[string]any {
	for _, provider := range providers {
		if fmt.Sprintf("%v", provider["provider_id"]) == id {
			return provider
		}
	}
	return nil
}

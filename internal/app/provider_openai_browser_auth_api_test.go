package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/ai"
	"github.com/collectors-tech/cabinet/internal/profile"
)

func TestOpenAIBrowserAuthConnectBindsVerifiedChatGPTLoginToActiveProfile(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	repo := profile.NewRepository(a.db)
	created, err := repo.Create(context.Background(), "Browser Auth Profile")
	if err != nil {
		t.Fatalf("create profile: %v", err)
	}
	if err := repo.SetActiveProfile(context.Background(), created.ID); err != nil {
		t.Fatalf("set active profile: %v", err)
	}
	runner := &fakeOpenAIBrowserAuthRuntime{
		status: ai.BrowserAuthStatus{State: "connected", Authenticated: true, Method: "chatgpt"},
	}
	mux := http.NewServeMux()
	registerOpenAIBrowserAuthRoutes(mux, repo, runner)

	req := httptest.NewRequest(http.MethodPost, "/api/providers/openai/browser-auth", strings.NewReader(`{"profile_id":"`+created.ID+`"}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	mux.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("connect status=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["state"] != "connected" || payload["authenticated"] != true || payload["profile_connected"] != true || payload["recommended"] != true {
		t.Fatalf("unexpected browser auth payload: %+v", payload)
	}
	settings, err := repo.GetSettings(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	for key, want := range map[string]string{
		"openai.active_auth_method":            "browser_auth",
		"openai_active_auth_method":            "browser_auth",
		"openai.browser_auth_state":            "connected",
		"openai.browser_auth_artifact_present": "true",
		"assistant_default_provider":           "openai",
		"assistant_default_model":              "gpt-5.6-luna",
		"integration.openai.enabled":           "true",
	} {
		if got := settings[key]; got != want {
			t.Fatalf("setting %s=%q, want %q", key, got, want)
		}
	}
	instances, err := repo.ListIntegrationInstances(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("list integration instances: %v", err)
	}
	if len(instances) != 1 || instances[0].ProviderID != "openai" || !instances[0].Enabled || instances[0].Config["openai.active_auth_method"] != "browser_auth" {
		t.Fatalf("expected enabled browser-auth OpenAI instance, got %+v", instances)
	}
	if len(instances[0].SecretRefs) != 0 {
		t.Fatalf("browser auth must not create API-key secret refs: %+v", instances[0].SecretRefs)
	}
}

func TestOpenAIBrowserAuthStatusAndDisconnectRemainNonSecretAndProfileScoped(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	repo := profile.NewRepository(a.db)
	created, err := repo.Create(context.Background(), "Browser Auth Status")
	if err != nil {
		t.Fatalf("create profile: %v", err)
	}
	runner := &fakeOpenAIBrowserAuthRuntime{
		status: ai.BrowserAuthStatus{State: "signed_out", Authenticated: false, Method: "chatgpt"},
	}
	mux := http.NewServeMux()
	registerOpenAIBrowserAuthRoutes(mux, repo, runner)
	if err := repo.PutSecret(context.Background(), created.ID, "openai_api_key", "sk-browser-disconnect-fallback"); err != nil {
		t.Fatalf("store fallback API key: %v", err)
	}
	if err := repo.PutSettings(context.Background(), created.ID, map[string]string{
		"openai.active_auth_method":            "browser_auth",
		"openai.browser_auth_state":            "connected",
		"openai.browser_auth_artifact_present": "true",
	}); err != nil {
		t.Fatalf("store browser auth settings: %v", err)
	}

	statusReq := httptest.NewRequest(http.MethodGet, "/api/providers/openai/browser-auth?profile_id="+created.ID, nil)
	statusResp := httptest.NewRecorder()
	mux.ServeHTTP(statusResp, statusReq)
	if statusResp.Code != http.StatusOK || !strings.Contains(statusResp.Body.String(), `"state":"signed_out"`) || strings.Contains(strings.ToLower(statusResp.Body.String()), "token") {
		t.Fatalf("unexpected safe status response: status=%d body=%s", statusResp.Code, statusResp.Body.String())
	}

	disconnectReq := httptest.NewRequest(http.MethodDelete, "/api/providers/openai/browser-auth?profile_id="+created.ID, nil)
	disconnectResp := httptest.NewRecorder()
	mux.ServeHTTP(disconnectResp, disconnectReq)
	if disconnectResp.Code != http.StatusOK || !strings.Contains(disconnectResp.Body.String(), `"global_login_preserved":true`) {
		t.Fatalf("unexpected disconnect response: status=%d body=%s", disconnectResp.Code, disconnectResp.Body.String())
	}
	if runner.logoutCalls != 0 {
		t.Fatalf("profile disconnect must not log the user out of Codex globally, got %d logout calls", runner.logoutCalls)
	}
	settings, err := repo.GetSettings(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("get post-disconnect settings: %v", err)
	}
	if settings["openai.active_auth_method"] != "api_key" || settings["integration.openai.enabled"] != "true" {
		t.Fatalf("disconnect must preserve API-key fallback readiness, got %+v", settings)
	}
}

type fakeOpenAIBrowserAuthRuntime struct {
	status      ai.BrowserAuthStatus
	startCalls  int
	logoutCalls int
}

func (r *fakeOpenAIBrowserAuthRuntime) Status(context.Context) (ai.BrowserAuthStatus, error) {
	return r.status, nil
}

func (r *fakeOpenAIBrowserAuthRuntime) StartLogin(context.Context) (ai.BrowserAuthStatus, error) {
	r.startCalls++
	return r.status, nil
}

func (r *fakeOpenAIBrowserAuthRuntime) RunAssistantTurn(context.Context, ai.BrowserAuthTurnRequest) (string, error) {
	return "browser auth test response", nil
}

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/collectors-tech/cabinet/internal/ai"
	"github.com/collectors-tech/cabinet/internal/profile"
)

type openAIBrowserAuthResponse struct {
	Provider             string `json:"provider"`
	AuthMethod           string `json:"auth_method"`
	State                string `json:"state"`
	Authenticated        bool   `json:"authenticated"`
	ProfileConnected     bool   `json:"profile_connected"`
	Recommended          bool   `json:"recommended"`
	Message              string `json:"message"`
	GlobalLoginPreserved bool   `json:"global_login_preserved,omitempty"`
}

func registerOpenAIBrowserAuthRoutes(mux *http.ServeMux, profiles *profile.Repository, browserAuth ai.BrowserAuthRuntime) {
	mux.HandleFunc("/api/providers/openai/browser-auth", func(w http.ResponseWriter, r *http.Request) {
		if profiles == nil || browserAuth == nil {
			http.Error(w, `{"error":"openai_browser_auth_unavailable"}`, http.StatusServiceUnavailable)
			return
		}
		profileID, err := openAIBrowserAuthProfileID(r.Context(), profiles, r)
		if err != nil {
			http.Error(w, `{"error":"active_profile_required"}`, http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			status, statusErr := browserAuth.Status(r.Context())
			if statusErr != nil {
				writeOpenAIBrowserAuthJSON(w, http.StatusServiceUnavailable, openAIBrowserAuthResponse{
					Provider: "openai", AuthMethod: "browser_auth", State: "unavailable", Recommended: true,
					Message: "Cabinet could not check ChatGPT sign-in. Retry from this PC.",
				})
				return
			}
			connected := openAIBrowserAuthProfileConnected(r.Context(), profiles, profileID)
			writeOpenAIBrowserAuthJSON(w, http.StatusOK, openAIBrowserAuthResponse{
				Provider: "openai", AuthMethod: "browser_auth", State: status.State,
				Authenticated: status.Authenticated, ProfileConnected: connected, Recommended: true,
				Message: status.Message,
			})
		case http.MethodPost:
			status, startErr := browserAuth.StartLogin(r.Context())
			if startErr != nil {
				writeOpenAIBrowserAuthJSON(w, http.StatusServiceUnavailable, openAIBrowserAuthResponse{
					Provider: "openai", AuthMethod: "browser_auth", State: "unavailable", Recommended: true,
					Message: "Cabinet could not open ChatGPT sign-in. Retry from this PC.",
				})
				return
			}
			if !status.Authenticated {
				code := http.StatusAccepted
				if status.State == "runtime_missing" || status.State == "unavailable" {
					code = http.StatusServiceUnavailable
				}
				writeOpenAIBrowserAuthJSON(w, code, openAIBrowserAuthResponse{
					Provider: "openai", AuthMethod: "browser_auth", State: status.State,
					Authenticated: false, ProfileConnected: false, Recommended: true, Message: status.Message,
				})
				return
			}
			if err := verifyAndBindOpenAIBrowserAuth(r.Context(), profiles, browserAuth, profileID); err != nil {
				writeOpenAIBrowserAuthJSON(w, http.StatusServiceUnavailable, openAIBrowserAuthResponse{
					Provider: "openai", AuthMethod: "browser_auth", State: "verification_failed",
					Authenticated: true, ProfileConnected: false, Recommended: true,
					Message: "ChatGPT is signed in, but Cabinet could not complete a test response. Retry the connection.",
				})
				return
			}
			writeOpenAIBrowserAuthJSON(w, http.StatusOK, openAIBrowserAuthResponse{
				Provider: "openai", AuthMethod: "browser_auth", State: "connected",
				Authenticated: true, ProfileConnected: true, Recommended: true,
				Message: "ChatGPT is connected to Cabinet and ready for Chat.",
			})
		case http.MethodDelete:
			if err := disconnectOpenAIBrowserAuthProfile(r.Context(), profiles, profileID); err != nil {
				http.Error(w, `{"error":"openai_browser_auth_disconnect_failed"}`, http.StatusInternalServerError)
				return
			}
			writeOpenAIBrowserAuthJSON(w, http.StatusOK, openAIBrowserAuthResponse{
				Provider: "openai", AuthMethod: "browser_auth", State: "disconnected",
				Authenticated: false, ProfileConnected: false, Recommended: true,
				Message: "ChatGPT is disconnected from this Cabinet profile.", GlobalLoginPreserved: true,
			})
		default:
			w.Header().Set("Allow", "GET, POST, DELETE")
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		}
	})
}

func openAIBrowserAuthProfileID(ctx context.Context, profiles *profile.Repository, r *http.Request) (string, error) {
	profileID := strings.TrimSpace(r.URL.Query().Get("profile_id"))
	if r.Method == http.MethodPost && strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		var request struct {
			ProfileID string `json:"profile_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			return "", err
		}
		profileID = strings.TrimSpace(request.ProfileID)
	}
	if profileID == "" {
		active, err := profiles.GetActiveProfile(ctx)
		if err != nil {
			return "", err
		}
		profileID = strings.TrimSpace(active.ID)
	}
	if profileID == "" {
		return "", fmt.Errorf("active profile is required")
	}
	if _, err := profiles.GetByID(ctx, profileID); err != nil {
		return "", err
	}
	return profileID, nil
}

func verifyAndBindOpenAIBrowserAuth(ctx context.Context, profiles *profile.Repository, browserAuth ai.BrowserAuthRuntime, profileID string) error {
	const model = "gpt-5.6-luna"
	if _, err := browserAuth.RunAssistantTurn(ctx, ai.BrowserAuthTurnRequest{
		ProfileID: profileID,
		ThreadID:  "provider-connection-test",
		Model:     model,
		Messages: []ai.AssistantTurnMessage{{
			Role: "user", Content: "Reply briefly that Cabinet Chat is connected. Do not perform any action.",
		}},
		Context: map[string]any{"surface": "integrations", "purpose": "connection_test"},
	}); err != nil {
		_ = profiles.PutSettings(ctx, profileID, map[string]string{
			"openai.browser_auth_provider_test_state":   "failed",
			"openai.browser_auth_provider_test_message": "Cabinet could not complete the browser-auth connection test.",
		})
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	settings := map[string]string{
		"openai.active_auth_method":                     "browser_auth",
		"openai_active_auth_method":                     "browser_auth",
		"openai.browser_auth_state":                     "connected",
		"openai.browser_auth_artifact_present":          "true",
		"openai.browser_auth_provider_test_state":       "passed",
		"openai.browser_auth_provider_test_artifact_id": "local_codex_chatgpt_test",
		"openai.browser_auth_provider_test_message":     "ChatGPT browser-auth runtime test passed.",
		"assistant_default_provider":                    "openai",
		"assistant_default_model":                       model,
		"integration.openai.enabled":                    "true",
	}
	if err := profiles.PutSettings(ctx, profileID, settings); err != nil {
		return err
	}
	instances, err := profiles.ListIntegrationInstances(ctx, profileID)
	if err != nil {
		return err
	}
	instanceID := ""
	for _, instance := range instances {
		if strings.EqualFold(strings.TrimSpace(instance.ProviderID), "openai") {
			instanceID = instance.ID
			break
		}
	}
	displayName, enabled, authState, healthState := "OpenAI / ChatGPT", true, "connected", "ready"
	_, err = profiles.UpsertIntegrationInstance(ctx, profileID, profile.IntegrationInstancePatch{
		ID: instanceID, ProviderID: "openai", DisplayName: &displayName, Enabled: &enabled,
		Config: map[string]string{
			"openai.active_auth_method": "browser_auth",
			"assistant_default_model":   model,
		},
		AuthState: &authState, HealthState: &healthState, LastCheckedAt: &now, LastSuccessAt: &now,
	})
	return err
}

func openAIBrowserAuthProfileConnected(ctx context.Context, profiles *profile.Repository, profileID string) bool {
	settings, err := profiles.GetSettings(ctx, profileID)
	if err != nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(settings["openai.active_auth_method"]), "browser_auth") &&
		strings.EqualFold(strings.TrimSpace(settings["openai.browser_auth_state"]), "connected") &&
		strings.EqualFold(strings.TrimSpace(settings["openai.browser_auth_artifact_present"]), "true") &&
		strings.EqualFold(strings.TrimSpace(settings["openai.browser_auth_provider_test_state"]), "passed")
}

func disconnectOpenAIBrowserAuthProfile(ctx context.Context, profiles *profile.Repository, profileID string) error {
	apiKeyRemains := false
	if apiKey, secretErr := profiles.GetSecret(ctx, profileID, "openai_api_key"); secretErr == nil && strings.TrimSpace(apiKey) != "" {
		apiKeyRemains = true
	}
	nextMethod, enabledSetting := "", "false"
	if apiKeyRemains {
		nextMethod, enabledSetting = "api_key", "true"
	}
	if err := profiles.PutSettings(ctx, profileID, map[string]string{
		"openai.active_auth_method":                     nextMethod,
		"openai_active_auth_method":                     nextMethod,
		"openai.browser_auth_state":                     "disconnected",
		"openai.browser_auth_artifact_present":          "false",
		"openai.browser_auth_provider_test_state":       "not_run",
		"openai.browser_auth_provider_test_artifact_id": "",
		"openai.browser_auth_provider_test_message":     "",
		"integration.openai.enabled":                    enabledSetting,
	}); err != nil {
		return err
	}
	instances, err := profiles.ListIntegrationInstances(ctx, profileID)
	if err != nil {
		return err
	}
	for _, instance := range instances {
		if !strings.EqualFold(strings.TrimSpace(instance.ProviderID), "openai") {
			continue
		}
		enabled, authState, healthState := apiKeyRemains, nextMethod, "needs_config"
		_, err := profiles.UpsertIntegrationInstance(ctx, profileID, profile.IntegrationInstancePatch{
			ID: instance.ID, ProviderID: "openai", Enabled: &enabled,
			Config:    map[string]string{"openai.active_auth_method": nextMethod},
			AuthState: &authState, HealthState: &healthState,
		})
		return err
	}
	return nil
}

func writeOpenAIBrowserAuthJSON(w http.ResponseWriter, status int, response openAIBrowserAuthResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}

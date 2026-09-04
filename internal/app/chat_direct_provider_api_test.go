package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestChatMessageOrdinaryConversationUsesSelectedAssistantProvider(t *testing.T) {
	t.Setenv("CABINET_E2E_MODE", "1")
	a := newTestApp(t)

	profileResponse := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Direct Provider Chat"}`), map[string]string{"Content-Type": "application/json"})
	if profileResponse.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", profileResponse.Code, profileResponse.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(profileResponse.Body.Bytes(), &profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	threadResponse := doRequest(t, a, http.MethodPost, "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profile.ID+`","title":"Ordinary provider chat"}`), map[string]string{"Content-Type": "application/json"})
	if threadResponse.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", threadResponse.Code, threadResponse.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(threadResponse.Body.Bytes(), &thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	messageBody := `{
		"profile_id":"` + profile.ID + `",
		"thread_id":"` + thread.ID + `",
		"role":"user",
		"content":"Tell me something helpful about my Cabinet workspace.",
		"context":{"assistant":{"provider":"fake","model":"cabinet-e2e-direct"}}
	}`
	messageResponse := doRequest(t, a, http.MethodPost, "/api/chat/messages", strings.NewReader(messageBody), map[string]string{"Content-Type": "application/json"})
	if messageResponse.Code != http.StatusCreated {
		t.Fatalf("create message status=%d body=%s", messageResponse.Code, messageResponse.Body.String())
	}
	var response struct {
		AssistantResponse struct {
			Mode          string `json:"mode"`
			Provider      string `json:"provider"`
			Model         string `json:"model"`
			ThreadMessage struct {
				Content string         `json:"content"`
				Context map[string]any `json:"context_json"`
			} `json:"thread_message"`
		} `json:"assistant_response"`
	}
	if err := json.Unmarshal(messageResponse.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode message response: %v", err)
	}
	if response.AssistantResponse.Mode != "provider" || response.AssistantResponse.Provider != "fake" || response.AssistantResponse.Model != "cabinet-e2e-direct" {
		t.Fatalf("expected selected provider response provenance, got %+v body=%s", response.AssistantResponse, messageResponse.Body.String())
	}
	if response.AssistantResponse.ThreadMessage.Content != "E2E direct provider response" {
		t.Fatalf("expected provider response content, got %q", response.AssistantResponse.ThreadMessage.Content)
	}

	messagesResponse := doRequest(t, a, http.MethodGet, "/api/chat/messages?profile_id="+profile.ID+"&thread_id="+thread.ID, nil, nil)
	if messagesResponse.Code != http.StatusOK {
		t.Fatalf("list messages status=%d body=%s", messagesResponse.Code, messagesResponse.Body.String())
	}
	if body := messagesResponse.Body.String(); !strings.Contains(body, `"source":"assistant_provider"`) || strings.Contains(body, "deterministic_chat_fallback") {
		t.Fatalf("expected persisted provider provenance without deterministic fallback, body=%s", body)
	}
}

func TestChatMessageOrdinaryConversationFailsClosedWhenSelectedProviderIsUnavailable(t *testing.T) {
	a := newTestApp(t)

	profileResponse := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Unavailable Provider Chat"}`), map[string]string{"Content-Type": "application/json"})
	if profileResponse.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", profileResponse.Code, profileResponse.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(profileResponse.Body.Bytes(), &profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	threadResponse := doRequest(t, a, http.MethodPost, "/api/chat/threads", strings.NewReader(`{"profile_id":"`+profile.ID+`","title":"Unavailable provider chat"}`), map[string]string{"Content-Type": "application/json"})
	if threadResponse.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", threadResponse.Code, threadResponse.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(threadResponse.Body.Bytes(), &thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	messageBody := `{
		"profile_id":"` + profile.ID + `",
		"thread_id":"` + thread.ID + `",
		"role":"user",
		"content":"Hello there.",
		"context":{"assistant":{"provider":"anthropic","model":"placeholder-model"}}
	}`
	messageResponse := doRequest(t, a, http.MethodPost, "/api/chat/messages", strings.NewReader(messageBody), map[string]string{"Content-Type": "application/json"})
	if messageResponse.Code != http.StatusCreated {
		t.Fatalf("create message status=%d body=%s", messageResponse.Code, messageResponse.Body.String())
	}
	body := messageResponse.Body.String()
	if !strings.Contains(body, `"mode":"provider_failure"`) ||
		!strings.Contains(body, `"provider":"anthropic"`) ||
		!strings.Contains(body, `"state":"setup_required"`) ||
		!strings.Contains(body, `"code":"adapter_unavailable"`) ||
		!strings.Contains(body, `"setup_next_action":"wait_for_supported_assistant_provider_adapter"`) {
		t.Fatalf("expected explicit selected-provider setup failure, body=%s", body)
	}
	if strings.Contains(body, "deterministic_chat_fallback") || strings.Contains(body, "I can help with Cabinet inventory") {
		t.Fatalf("provider failure must not be presented as a deterministic success, body=%s", body)
	}
}

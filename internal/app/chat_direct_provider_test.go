package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/ai"
	"github.com/collectors-tech/cabinet/internal/chat"
)

type directCaptureAssistantProvider struct {
	req ai.AssistantTurnRequest
}

func (p *directCaptureAssistantProvider) Name() string { return "openai" }

func (p *directCaptureAssistantProvider) RunAssistantTurn(_ context.Context, req ai.AssistantTurnRequest) (ai.AssistantTurnResponse, error) {
	p.req = req
	return ai.AssistantTurnResponse{
		Provider: "openai",
		Model:    req.Model,
		Text:     "Bounded direct provider reply",
		Metadata: map[string]string{
			"active_auth_method": "browser_auth",
			"integration_id":     "integration-openai",
			"secret_value":       "must-not-persist",
		},
	}, nil
}

func TestDirectAssistantProviderReceivesBoundedConversationWithoutCabinetAuthority(t *testing.T) {
	a := newTestApp(t)
	chatSvc := chat.NewService(a.db, filepath.Join(a.cfg.DataDir, "chat-attachments"))
	profileResponse := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Direct provider boundary"}`), map[string]string{"Content-Type": "application/json"})
	if profileResponse.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", profileResponse.Code, profileResponse.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(profileResponse.Body.Bytes(), &profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	thread, err := chatSvc.CreateThread(context.Background(), profile.ID, "Bounded conversation", nil)
	if err != nil {
		t.Fatalf("create thread: %v", err)
	}
	for index := 0; index < 55; index++ {
		role := "user"
		if index%2 == 1 {
			role = "assistant"
		}
		if _, err := chatSvc.CreateMessage(context.Background(), profile.ID, thread.ID, role, fmt.Sprintf("message-%02d", index), nil); err != nil {
			t.Fatalf("create history message %d: %v", index, err)
		}
	}

	provider := &directCaptureAssistantProvider{}
	result, handled := dispatchChatDirectAssistantProvider(
		context.Background(),
		chatSvc,
		ai.NewAssistantProviderRegistry(provider),
		profile.ID,
		thread.ID,
		"message-54",
		map[string]any{
			"assistant": map[string]any{"provider": "openai", "model": "gpt-test"},
			"route":     map[string]any{"pathname": "/chats/", "client_secret": "must-not-cross-provider-boundary"},
			"profile":   map[string]any{"id": profile.ID, "private_note": "must-not-cross-provider-boundary"},
			"agent_context": map[string]any{
				"profile_id": profile.ID,
				"thread_id":  thread.ID,
			},
		},
		"source-message",
	)
	if !handled || result["mode"] != "provider" {
		t.Fatalf("expected handled provider response, got handled=%v result=%+v", handled, result)
	}
	if len(provider.req.Messages) != directAssistantHistoryLimit {
		t.Fatalf("provider history length=%d want=%d", len(provider.req.Messages), directAssistantHistoryLimit)
	}
	if provider.req.Messages[0].Content != "message-05" || provider.req.Messages[len(provider.req.Messages)-1].Content != "message-54" {
		t.Fatalf("unexpected bounded ordered history: first=%+v last=%+v", provider.req.Messages[0], provider.req.Messages[len(provider.req.Messages)-1])
	}
	if provider.req.Metadata["cabinet_tool_authority"] != "none" ||
		provider.req.Metadata["raw_provider_tools_given"] != "false" ||
		provider.req.Metadata["entry_point"] != "chat.direct_conversation" {
		t.Fatalf("unexpected direct provider authority metadata: %+v", provider.req.Metadata)
	}
	if len(provider.req.Context) != 1 || provider.req.Context["agent_context"] == nil {
		t.Fatalf("provider context must contain only normalized agent_context evidence: %+v", provider.req.Context)
	}
	contextJSON, _ := json.Marshal(provider.req.Context)
	if strings.Contains(string(contextJSON), "must-not-cross-provider-boundary") {
		t.Fatalf("provider context leaked arbitrary client context: %s", contextJSON)
	}
	trace, _ := result["provider_trace"].(map[string]string)
	if trace["active_auth_method"] != "browser_auth" || trace["integration_id"] != "integration-openai" {
		t.Fatalf("expected allowlisted provider trace, got %+v", trace)
	}
	if _, leaked := trace["secret_value"]; leaked {
		t.Fatalf("provider trace leaked non-allowlisted metadata: %+v", trace)
	}
}

package ai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestConnectivityAndSuggest(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"part_number\":\"P-1\",\"brand\":\"AFX\",\"title\":\"AFX P-1\",\"confidence\":0.91}"}}]}`))
	}))
	defer srv.Close()

	svc := NewService(Config{BaseURL: srv.URL})
	if err := svc.TestConnectivity(context.Background(), "sk-test"); err != nil {
		t.Fatalf("TestConnectivity() error = %v", err)
	}
	out, err := svc.SuggestFromTitle(context.Background(), "sk-test", "AFX P-1 slot car")
	if err != nil {
		t.Fatalf("SuggestFromTitle() error = %v", err)
	}
	if out.PartNumber == "" || out.Confidence <= 0 {
		t.Fatalf("unexpected AI output: %+v", out)
	}
	if !out.RequiresConfirmation {
		t.Fatal("expected requires confirmation true")
	}
}

func TestProposeOperationStructuredPayload(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o-mini"}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"action\":\"create_wishlist_entry\",\"target\":\"wishlist\",\"payload\":{\"part_number\":\"CP-233\",\"title\":\"Copilot Wishlist\"},\"confidence\":0.88}"}}]}`))
	}))
	defer srv.Close()

	svc := NewService(Config{BaseURL: srv.URL})
	proposal, err := svc.ProposeOperation(context.Background(), "sk-test", "Create wishlist entry for CP-233")
	if err != nil {
		t.Fatalf("ProposeOperation() error = %v", err)
	}
	if proposal.Action != "create_wishlist_entry" {
		t.Fatalf("expected action create_wishlist_entry, got %q", proposal.Action)
	}
	if proposal.Target != "wishlist" {
		t.Fatalf("expected target wishlist, got %q", proposal.Target)
	}
	if proposal.Payload["part_number"] != "CP-233" {
		t.Fatalf("expected payload part_number CP-233, got %+v", proposal.Payload)
	}
	if !proposal.RequiresConfirmation {
		t.Fatal("expected requires confirmation true")
	}
}

func TestFakeAssistantProviderReturnsProviderNeutralTurn(t *testing.T) {
	t.Parallel()

	provider := NewFakeAssistantProvider()
	req := AssistantTurnRequest{
		ProfileID: "profile-1",
		ThreadID:  "thread-1",
		Provider:  "fake",
		Model:     "fake-assistant-model",
		Messages: []AssistantTurnMessage{
			{Role: "system", Content: "stay deterministic"},
			{Role: "user", Content: "summarize setup state"},
		},
		Context: map[string]any{
			"route": map[string]any{"pathname": "/integrations"},
		},
	}

	first, err := provider.RunAssistantTurn(context.Background(), req)
	if err != nil {
		t.Fatalf("RunAssistantTurn() error = %v", err)
	}
	second, err := provider.RunAssistantTurn(context.Background(), req)
	if err != nil {
		t.Fatalf("RunAssistantTurn() second error = %v", err)
	}

	if first.Provider != "fake" || first.Model != "fake-assistant-model" {
		t.Fatalf("expected fake provider/model metadata, got %+v", first)
	}
	if first.Text == "" || first.Text != second.Text {
		t.Fatalf("expected deterministic fake turn text, first=%q second=%q", first.Text, second.Text)
	}
	if first.Metadata["test_provider"] != "true" || first.Metadata["network"] != "disabled" {
		t.Fatalf("expected explicit no-network test metadata, got %+v", first.Metadata)
	}
	if first.Metadata["openai_ready"] != "" || first.Metadata["anthropic_ready"] != "" || first.Metadata["google_ready"] != "" {
		t.Fatalf("fake provider must not mark live providers ready, got %+v", first.Metadata)
	}
}

func TestFakeAssistantProviderRejectsMissingTurnContext(t *testing.T) {
	t.Parallel()

	_, err := NewFakeAssistantProvider().RunAssistantTurn(context.Background(), AssistantTurnRequest{
		ProfileID: "profile-1",
		Provider:  "fake",
		Messages:  []AssistantTurnMessage{{Role: "user", Content: "hello"}},
	})
	if err == nil {
		t.Fatal("expected missing thread context error")
	}
}

func TestOpenAIAssistantProviderUsesProfileSecretBoundary(t *testing.T) {
	t.Parallel()

	var sawAuthorization string
	var sawModel string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected OpenAI path %s", r.URL.Path)
		}
		sawAuthorization = r.Header.Get("Authorization")
		var body struct {
			Model    string                 `json:"model"`
			Messages []AssistantTurnMessage `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		sawModel = body.Model
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ready from OpenAI boundary"}}]}`))
	}))
	defer srv.Close()

	resolver := fakeAssistantSetupResolver{
		setup: AssistantProviderSetup{
			ProviderID:        "openai",
			Enabled:           true,
			ActiveAuthMethod:  "api_key",
			DefaultModel:      "gpt-4.1-mini",
			APIKeySecretRef:   "integration.inst_123.openai_api_key",
			HealthState:       "ready",
			IntegrationMode:   "assistant_workflows",
			IntegrationID:     "inst_123",
			ConfigSchemaRef:   "integrations/openai/auth",
			WorkflowReference: "assistant.chat",
		},
		secret: "sk-boundary-secret",
	}
	provider := NewOpenAIAssistantProvider(NewService(Config{BaseURL: srv.URL}), resolver)
	resp, err := provider.RunAssistantTurn(context.Background(), AssistantTurnRequest{
		ProfileID: "profile-1",
		ThreadID:  "thread-1",
		Messages:  []AssistantTurnMessage{{Role: "user", Content: "hello"}},
	})
	if err != nil {
		t.Fatalf("RunAssistantTurn() error = %v", err)
	}

	if sawAuthorization != "Bearer sk-boundary-secret" {
		t.Fatalf("expected adapter to fetch the profile secret for Authorization, got %q", sawAuthorization)
	}
	if sawModel != "gpt-4.1-mini" {
		t.Fatalf("expected default model from setup boundary, got %q", sawModel)
	}
	if resp.Provider != "openai" || resp.Model != "gpt-4.1-mini" || resp.Text != "ready from OpenAI boundary" {
		t.Fatalf("unexpected OpenAI assistant response: %+v", resp)
	}
	if resp.Metadata["integration_id"] != "inst_123" || resp.Metadata["active_auth_method"] != "api_key" || resp.Metadata["secret_ref"] != "" {
		t.Fatalf("expected non-secret setup metadata only, got %+v", resp.Metadata)
	}
	if strings.Contains(resp.Text, "sk-boundary-secret") {
		t.Fatalf("assistant response leaked secret: %+v", resp)
	}
}

func TestOpenAIAssistantProviderDoesNotReceiveCabinetToolAuthority(t *testing.T) {
	t.Parallel()

	var rawBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&rawBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"planner text only"}}]}`))
	}))
	defer srv.Close()

	resp, err := newReadyOpenAITestProvider(srv.URL).RunAssistantTurn(context.Background(), AssistantTurnRequest{
		ProfileID: "profile-1",
		ThreadID:  "thread-1",
		Messages:  []AssistantTurnMessage{{Role: "user", Content: "draft a wishlist update"}},
		Context: map[string]any{
			"db":         "sqlite://cabinet.db",
			"filesystem": `C:\Users\maxbarrass\.openclaw\workspace-cabinet-developer`,
			"skills":     []string{"wishlist.apply", "inventory.delete"},
			"tools":      []string{"app_control.apply"},
		},
		Metadata: map[string]string{
			"capability_id": "wishlist.apply",
		},
	})
	if err != nil {
		t.Fatalf("RunAssistantTurn() error = %v", err)
	}
	for _, forbidden := range []string{"context", "metadata", "tools", "skills", "db", "filesystem", "capability_id"} {
		if _, ok := rawBody[forbidden]; ok {
			t.Fatalf("OpenAI request body exposed Cabinet authority field %q: %+v", forbidden, rawBody)
		}
	}
	if resp.Metadata["cabinet_tool_authority"] != "none" ||
		resp.Metadata["cabinet_database_access"] != "none" ||
		resp.Metadata["cabinet_filesystem_access"] != "none" ||
		resp.Metadata["governed_dispatch_owner"] != "cabinet" {
		t.Fatalf("expected explicit no-authority provider metadata, got %+v", resp.Metadata)
	}
}

func TestOpenAIAssistantProviderReportsMissingProfileSecretWithoutNetwork(t *testing.T) {
	t.Parallel()

	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		t.Fatalf("OpenAI adapter must not call network without a profile secret")
	}))
	defer srv.Close()

	provider := NewOpenAIAssistantProvider(NewService(Config{BaseURL: srv.URL}), fakeAssistantSetupResolver{
		setup: AssistantProviderSetup{
			ProviderID:       "openai",
			Enabled:          true,
			ActiveAuthMethod: "api_key",
			DefaultModel:     "gpt-4o-mini",
			APIKeySecretRef:  "integration.inst_123.openai_api_key",
			HealthState:      "ready",
			IntegrationID:    "inst_123",
		},
	})
	resp, err := provider.RunAssistantTurn(context.Background(), AssistantTurnRequest{
		ProfileID: "profile-1",
		ThreadID:  "thread-1",
		Messages:  []AssistantTurnMessage{{Role: "user", Content: "hello"}},
	})
	if err == nil {
		t.Fatal("expected missing credential error")
	}
	if called {
		t.Fatal("network was called despite missing profile secret")
	}
	if resp.ErrorClass != "missing_credentials" || resp.SetupNextAction != "configure_openai_api_key" {
		t.Fatalf("expected redacted missing credential guidance, got resp=%+v err=%v", resp, err)
	}
	if strings.Contains(err.Error(), "integration.inst_123.openai_api_key") || strings.Contains(err.Error(), "sk-") {
		t.Fatalf("missing credential error leaked secret detail: %v", err)
	}
}

func TestOpenAIAssistantProviderRejectsUnsupportedModelBeforeSecretOrNetwork(t *testing.T) {
	t.Parallel()

	var networkCalled bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		networkCalled = true
		t.Fatalf("OpenAI adapter must not call network for unsupported model")
	}))
	defer srv.Close()

	var secretCalls int
	provider := NewOpenAIAssistantProvider(NewService(Config{BaseURL: srv.URL}), countedAssistantSetupResolver{
		setup: AssistantProviderSetup{
			ProviderID:       "openai",
			Enabled:          true,
			ActiveAuthMethod: "api_key",
			DefaultModel:     "gpt-4o-mini",
			SupportedModels:  []string{"gpt-4o-mini", "gpt-4.1-mini"},
			APIKeySecretRef:  "integration.inst_123.openai_api_key",
			HealthState:      "ready",
			IntegrationID:    "inst_123",
		},
		secret:      "sk-boundary-secret",
		secretCalls: &secretCalls,
	})
	resp, err := provider.RunAssistantTurn(context.Background(), AssistantTurnRequest{
		ProfileID: "profile-1",
		ThreadID:  "thread-1",
		Model:     "gpt-9-experimental",
		Messages:  []AssistantTurnMessage{{Role: "user", Content: "hello"}},
	})
	if err == nil {
		t.Fatal("expected unsupported model error")
	}
	if networkCalled || secretCalls != 0 {
		t.Fatalf("unsupported model must stop before secret/network, network=%v secretCalls=%d", networkCalled, secretCalls)
	}
	if resp.ErrorClass != "unsupported_model" || resp.SetupNextAction != "choose_supported_openai_model" || resp.Model != "gpt-9-experimental" {
		t.Fatalf("expected redacted unsupported model guidance, got resp=%+v err=%v", resp, err)
	}
	if strings.Contains(err.Error(), "gpt-9-experimental") || strings.Contains(err.Error(), "sk-") {
		t.Fatalf("unsupported model error leaked request/secret detail: %v", err)
	}
}

func TestOpenAIAssistantProviderClassifiesProviderFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		status         int
		delay          time.Duration
		timeout        time.Duration
		wantClass      string
		wantNextAction string
	}{
		{
			name:           "rate limit",
			status:         http.StatusTooManyRequests,
			wantClass:      "rate_limit",
			wantNextAction: "wait_for_openai_rate_limit",
		},
		{
			name:           "provider failure",
			status:         http.StatusInternalServerError,
			wantClass:      "provider_failure",
			wantNextAction: "retry_openai_assistant_turn",
		},
		{
			name:           "timeout",
			status:         http.StatusOK,
			delay:          50 * time.Millisecond,
			timeout:        5 * time.Millisecond,
			wantClass:      "timeout",
			wantNextAction: "retry_openai_assistant_turn_after_timeout",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.delay > 0 {
					time.Sleep(tc.delay)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.status < 300 {
					_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
				} else {
					_, _ = w.Write([]byte(`{"error":{"message":"provider failed with sk-live-sensitive detail"}}`))
				}
			}))
			defer srv.Close()

			ctx := context.Background()
			if tc.timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tc.timeout)
				defer cancel()
			}
			resp, err := newReadyOpenAITestProvider(srv.URL).RunAssistantTurn(ctx, AssistantTurnRequest{
				ProfileID: "profile-1",
				ThreadID:  "thread-1",
				Messages:  []AssistantTurnMessage{{Role: "user", Content: "hello"}},
			})
			if err == nil {
				t.Fatal("expected classified provider error")
			}
			if resp.ErrorClass != tc.wantClass || resp.SetupNextAction != tc.wantNextAction {
				t.Fatalf("expected %s/%s, got resp=%+v err=%v", tc.wantClass, tc.wantNextAction, resp, err)
			}
			if strings.Contains(err.Error(), "sk-live-sensitive") {
				t.Fatalf("classified provider error leaked raw provider payload: %v", err)
			}
		})
	}
}

func TestOpenAIAssistantProviderClassifiesTransportFailure(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("closed test server should not receive requests")
	}))
	baseURL := srv.URL
	srv.Close()

	resp, err := newReadyOpenAITestProvider(baseURL).RunAssistantTurn(context.Background(), AssistantTurnRequest{
		ProfileID: "profile-1",
		ThreadID:  "thread-1",
		Messages:  []AssistantTurnMessage{{Role: "user", Content: "hello"}},
	})
	if err == nil {
		t.Fatal("expected transport error")
	}
	if resp.ErrorClass != "transport_failure" || resp.SetupNextAction != "retry_when_openai_network_is_available" {
		t.Fatalf("expected transport failure guidance, got resp=%+v err=%v", resp, err)
	}
	if strings.Contains(err.Error(), "sk-boundary-secret") {
		t.Fatalf("transport error leaked secret detail: %v", err)
	}
}

func TestOpenAIAssistantProviderClassifiesCancelledTurn(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	resp, err := newReadyOpenAITestProvider("https://example.invalid").RunAssistantTurn(ctx, AssistantTurnRequest{
		ProfileID: "profile-1",
		ThreadID:  "thread-1",
		Messages:  []AssistantTurnMessage{{Role: "user", Content: "hello"}},
	})
	if err == nil {
		t.Fatal("expected cancellation error")
	}
	if resp.ErrorClass != "cancelled" {
		t.Fatalf("expected cancelled error class, got resp=%+v err=%v", resp, err)
	}
}

func TestOpenAIAssistantProviderReturnsRedactedRuntimeErrors(t *testing.T) {
	t.Parallel()

	const (
		secretValue = "sk-runtime-redaction-secret"
		rawPath     = `C:\Users\maxbarrass\.openclaw\workspace-cabinet-developer\secret-openai-dump.json`
		rawPayload  = `{"error":{"message":"provider leaked ` + secretValue + ` from ` + rawPath + `"}}`
	)

	service := NewService(Config{BaseURL: "https://openai.example"})
	service.client = &http.Client{Transport: failingRoundTripper{err: errors.New(rawPayload)}}
	resp, err := NewOpenAIAssistantProvider(service, fakeAssistantSetupResolver{
		setup: AssistantProviderSetup{
			ProviderID:       "openai",
			Enabled:          true,
			ActiveAuthMethod: "api_key",
			DefaultModel:     "gpt-4o-mini",
			SupportedModels:  []string{"gpt-4o-mini"},
			APIKeySecretRef:  "integration.inst_123.openai_api_key",
			HealthState:      "ready",
			IntegrationID:    "inst_123",
		},
		secret: secretValue,
	}).RunAssistantTurn(context.Background(), AssistantTurnRequest{
		ProfileID: "profile-1",
		ThreadID:  "thread-1",
		Messages:  []AssistantTurnMessage{{Role: "user", Content: "hello"}},
	})
	if err == nil {
		t.Fatal("expected redacted runtime error")
	}
	if resp.ErrorClass != "transport_failure" || resp.SetupNextAction != "retry_when_openai_network_is_available" {
		t.Fatalf("expected transport failure guidance, got resp=%+v err=%v", resp, err)
	}
	encoded, marshalErr := json.Marshal(resp)
	if marshalErr != nil {
		t.Fatalf("marshal response: %v", marshalErr)
	}
	for _, leaked := range []string{secretValue, "integration.inst_123.openai_api_key", rawPath, rawPayload, "secret-openai-dump"} {
		if strings.Contains(err.Error(), leaked) || strings.Contains(string(encoded), leaked) {
			t.Fatalf("assistant provider runtime evidence leaked %q: err=%v resp=%s", leaked, err, string(encoded))
		}
	}
}

type fakeAssistantSetupResolver struct {
	setup  AssistantProviderSetup
	secret string
}

func (r fakeAssistantSetupResolver) ResolveAssistantProviderSetup(context.Context, string, string) (AssistantProviderSetup, error) {
	return r.setup, nil
}

func (r fakeAssistantSetupResolver) GetAssistantProviderSecret(context.Context, string, string) (string, error) {
	if strings.TrimSpace(r.secret) == "" {
		return "", errFakeSecretMissing{}
	}
	return r.secret, nil
}

type errFakeSecretMissing struct{}

func (errFakeSecretMissing) Error() string { return "secret not found" }

type countedAssistantSetupResolver struct {
	setup       AssistantProviderSetup
	secret      string
	secretCalls *int
}

func (r countedAssistantSetupResolver) ResolveAssistantProviderSetup(context.Context, string, string) (AssistantProviderSetup, error) {
	return r.setup, nil
}

func (r countedAssistantSetupResolver) GetAssistantProviderSecret(context.Context, string, string) (string, error) {
	if r.secretCalls != nil {
		(*r.secretCalls)++
	}
	return r.secret, nil
}

func newReadyOpenAITestProvider(baseURL string) *OpenAIAssistantProvider {
	return NewOpenAIAssistantProvider(NewService(Config{BaseURL: baseURL}), fakeAssistantSetupResolver{
		setup: AssistantProviderSetup{
			ProviderID:       "openai",
			Enabled:          true,
			ActiveAuthMethod: "api_key",
			DefaultModel:     "gpt-4o-mini",
			SupportedModels:  []string{"gpt-4o-mini"},
			APIKeySecretRef:  "integration.inst_123.openai_api_key",
			HealthState:      "ready",
			IntegrationID:    "inst_123",
		},
		secret: "sk-boundary-secret",
	})
}

type failingRoundTripper struct {
	err error
}

func (rt failingRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, rt.err
}

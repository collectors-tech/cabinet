package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

type Config struct {
	BaseURL string
}

type Suggestion struct {
	PartNumber           string  `json:"part_number"`
	Brand                string  `json:"brand"`
	Title                string  `json:"title"`
	Confidence           float64 `json:"confidence"`
	RequiresConfirmation bool    `json:"requires_confirmation"`
}

type OperationProposal struct {
	Action               string         `json:"action"`
	Target               string         `json:"target"`
	Payload              map[string]any `json:"payload"`
	Confidence           float64        `json:"confidence"`
	RequiresConfirmation bool           `json:"requires_confirmation"`
}

type Service struct {
	baseURL string
	client  *http.Client
}

func (s *Service) Name() string {
	return "openai"
}

type OpenAIAssistantProvider struct {
	service     *Service
	resolver    AssistantProviderSetupResolver
	browserAuth BrowserAuthRuntime
}

func NewOpenAIAssistantProvider(service *Service, resolver AssistantProviderSetupResolver, browserAuth ...BrowserAuthRuntime) *OpenAIAssistantProvider {
	if service == nil {
		service = NewService(Config{})
	}
	var runtime BrowserAuthRuntime
	if len(browserAuth) > 0 {
		runtime = browserAuth[0]
	}
	return &OpenAIAssistantProvider{service: service, resolver: resolver, browserAuth: runtime}
}

func (p *OpenAIAssistantProvider) Name() string {
	return "openai"
}

func (p *OpenAIAssistantProvider) RunAssistantTurn(ctx context.Context, req AssistantTurnRequest) (AssistantTurnResponse, error) {
	if err := ctx.Err(); err != nil {
		return AssistantTurnResponse{Provider: "openai", ErrorClass: "cancelled"}, err
	}
	if strings.TrimSpace(req.ProfileID) == "" {
		return AssistantTurnResponse{Provider: "openai", ErrorClass: "missing_context", SetupNextAction: "select_active_profile"}, fmt.Errorf("profile_id is required")
	}
	if strings.TrimSpace(req.ThreadID) == "" {
		return AssistantTurnResponse{Provider: "openai", ErrorClass: "missing_context", SetupNextAction: "open_assistant_thread"}, fmt.Errorf("thread_id is required")
	}
	if p == nil || p.resolver == nil {
		return AssistantTurnResponse{Provider: "openai", ErrorClass: "unhealthy_provider", SetupNextAction: "configure_openai_provider"}, fmt.Errorf("assistant provider setup resolver is required")
	}
	setup, err := p.resolver.ResolveAssistantProviderSetup(ctx, req.ProfileID, "openai")
	if err != nil {
		return AssistantTurnResponse{Provider: "openai", ErrorClass: "unhealthy_provider", SetupNextAction: "configure_openai_provider"}, fmt.Errorf("openai assistant setup unavailable")
	}
	respBase := AssistantTurnResponse{
		Provider: "openai",
		Model:    assistantTurnModel(req.Model, setup.DefaultModel),
		Metadata: openAIAssistantSetupMetadata(setup),
	}
	if !assistantModelSupported(respBase.Model, setup.SupportedModels) {
		respBase.ErrorClass = "unsupported_model"
		respBase.SetupNextAction = "choose_supported_openai_model"
		return respBase, fmt.Errorf("openai assistant model is not supported")
	}
	if strings.EqualFold(strings.TrimSpace(setup.ActiveAuthMethod), "browser_auth") {
		if !setup.Enabled || p.browserAuth == nil {
			respBase.ErrorClass = "missing_credentials"
			respBase.SetupNextAction = "connect_openai_browser_auth"
			return respBase, fmt.Errorf("openai browser auth is required")
		}
		status, err := p.browserAuth.Status(ctx)
		if err != nil || !status.Authenticated {
			respBase.ErrorClass = "missing_credentials"
			respBase.SetupNextAction = "connect_openai_browser_auth"
			return respBase, fmt.Errorf("openai browser auth is required")
		}
		text, err := p.browserAuth.RunAssistantTurn(ctx, BrowserAuthTurnRequest{
			ProfileID: req.ProfileID,
			ThreadID:  req.ThreadID,
			Model:     respBase.Model,
			Messages:  req.Messages,
			Context:   req.Context,
		})
		if err != nil {
			respBase.ErrorClass = classifyAssistantProviderError(err)
			respBase.SetupNextAction = assistantProviderFailureNextAction(respBase.ErrorClass)
			return respBase, assistantProviderRedactedError(respBase.ErrorClass)
		}
		respBase.Text = text
		return respBase, nil
	}
	if !setup.Enabled || strings.TrimSpace(setup.APIKeySecretRef) == "" || !strings.EqualFold(strings.TrimSpace(setup.ActiveAuthMethod), "api_key") {
		respBase.ErrorClass = "missing_credentials"
		respBase.SetupNextAction = "configure_openai_api_key"
		return respBase, fmt.Errorf("openai api key is required")
	}
	apiKey, err := p.resolver.GetAssistantProviderSecret(ctx, req.ProfileID, setup.APIKeySecretRef)
	if err != nil || strings.TrimSpace(apiKey) == "" {
		respBase.ErrorClass = "missing_credentials"
		respBase.SetupNextAction = "configure_openai_api_key"
		return respBase, fmt.Errorf("openai api key is required")
	}
	service := p.service
	if baseURL := strings.TrimSpace(setup.BaseURL); baseURL != "" {
		service = NewService(Config{BaseURL: baseURL})
	}
	text, err := service.runAssistantChatCompletion(ctx, apiKey, respBase.Model, req.Messages)
	if err != nil {
		respBase.ErrorClass = classifyAssistantProviderError(err)
		respBase.SetupNextAction = assistantProviderFailureNextAction(respBase.ErrorClass)
		return respBase, assistantProviderRedactedError(respBase.ErrorClass)
	}
	respBase.Text = text
	return respBase, nil
}

func NewService(cfg Config) *Service {
	base := strings.TrimSpace(cfg.BaseURL)
	if base == "" {
		base = "https://api.openai.com"
	}
	return &Service{
		baseURL: strings.TrimRight(base, "/"),
		client:  &http.Client{Timeout: 20 * time.Second},
	}
}

func (s *Service) TestConnectivity(ctx context.Context, apiKey string) error {
	if strings.TrimSpace(apiKey) == "" {
		return fmt.Errorf("api key is required")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"/v1/models", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("connectivity failed: status %d", resp.StatusCode)
	}
	return nil
}

func (s *Service) SuggestFromTitle(ctx context.Context, apiKey, title string) (Suggestion, error) {
	if strings.TrimSpace(apiKey) == "" {
		return Suggestion{}, fmt.Errorf("api key is required")
	}
	prompt := "Extract part_number, brand, title and confidence from: " + title + ". Return JSON only."
	reqBody := map[string]any{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "system", "content": "You are a strict JSON extraction engine."},
			{"role": "user", "content": prompt},
		},
	}
	raw, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/v1/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return Suggestion{}, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return Suggestion{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return Suggestion{}, fmt.Errorf("suggest failed: status %d", resp.StatusCode)
	}

	var payload struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Suggestion{}, err
	}
	if len(payload.Choices) == 0 {
		return Suggestion{}, fmt.Errorf("no AI choices")
	}
	content := strings.TrimSpace(payload.Choices[0].Message.Content)
	content = strings.Trim(content, "`")
	content = strings.TrimPrefix(content, "json")
	content = strings.TrimSpace(content)
	var out Suggestion
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return Suggestion{}, err
	}
	out.RequiresConfirmation = true
	return out, nil
}

func (s *Service) runAssistantChatCompletion(ctx context.Context, apiKey, model string, messages []AssistantTurnMessage) (string, error) {
	if strings.TrimSpace(apiKey) == "" {
		return "", fmt.Errorf("api key is required")
	}
	if strings.TrimSpace(model) == "" {
		model = "gpt-4o-mini"
	}
	reqBody := map[string]any{
		"model":    model,
		"messages": messages,
	}
	raw, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/v1/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return "", classifyAssistantProviderTransportError(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", assistantProviderHTTPError{statusCode: resp.StatusCode}
	}
	var payload struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if len(payload.Choices) == 0 {
		return "", fmt.Errorf("no AI choices")
	}
	return strings.TrimSpace(payload.Choices[0].Message.Content), nil
}

type assistantProviderHTTPError struct {
	statusCode int
}

func (e assistantProviderHTTPError) Error() string {
	return fmt.Sprintf("assistant provider request failed: status %d", e.statusCode)
}

type assistantProviderClassifiedError struct {
	class string
	err   error
}

func (e assistantProviderClassifiedError) Error() string {
	if e.err == nil {
		return e.class
	}
	return e.err.Error()
}

func (e assistantProviderClassifiedError) Unwrap() error {
	return e.err
}

func classifyAssistantProviderTransportError(err error) error {
	if err == nil {
		return nil
	}
	class := "transport_failure"
	switch {
	case errors.Is(err, context.Canceled):
		class = "cancelled"
	case errors.Is(err, context.DeadlineExceeded):
		class = "timeout"
	default:
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			class = "timeout"
		}
	}
	return assistantProviderClassifiedError{class: class, err: err}
}

func classifyAssistantProviderError(err error) string {
	if err == nil {
		return ""
	}
	var classified assistantProviderClassifiedError
	if errors.As(err, &classified) && strings.TrimSpace(classified.class) != "" {
		return classified.class
	}
	var httpErr assistantProviderHTTPError
	if errors.As(err, &httpErr) {
		if httpErr.statusCode == http.StatusTooManyRequests {
			return "rate_limit"
		}
		return "provider_failure"
	}
	switch {
	case errors.Is(err, context.Canceled):
		return "cancelled"
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	default:
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return "timeout"
		}
		return "provider_failure"
	}
}

func assistantProviderFailureNextAction(class string) string {
	switch class {
	case "cancelled":
		return "retry_openai_assistant_turn"
	case "timeout":
		return "retry_openai_assistant_turn_after_timeout"
	case "rate_limit":
		return "wait_for_openai_rate_limit"
	case "transport_failure":
		return "retry_when_openai_network_is_available"
	default:
		return "retry_openai_assistant_turn"
	}
}

func assistantProviderRedactedError(class string) error {
	switch strings.TrimSpace(class) {
	case "cancelled":
		return fmt.Errorf("openai assistant turn cancelled")
	case "timeout":
		return fmt.Errorf("openai assistant turn timed out")
	case "rate_limit":
		return fmt.Errorf("openai assistant provider rate limited the request")
	case "transport_failure":
		return fmt.Errorf("openai assistant provider transport failed")
	default:
		return fmt.Errorf("openai assistant provider failed")
	}
}

func assistantTurnModel(requested, fallback string) string {
	if model := strings.TrimSpace(requested); model != "" {
		return model
	}
	if model := strings.TrimSpace(fallback); model != "" {
		return model
	}
	return "gpt-4o-mini"
}

func assistantModelSupported(model string, supported []string) bool {
	model = strings.TrimSpace(model)
	if model == "" {
		return false
	}
	if len(supported) == 0 {
		supported = []string{"gpt-5.6-luna", "gpt-4o-mini", "gpt-4.1-mini"}
	}
	for _, candidate := range supported {
		if strings.EqualFold(model, strings.TrimSpace(candidate)) {
			return true
		}
	}
	return false
}

func openAIAssistantSetupMetadata(setup AssistantProviderSetup) map[string]string {
	metadata := map[string]string{
		"provider_id":               "openai",
		"active_auth_method":        strings.TrimSpace(setup.ActiveAuthMethod),
		"health_state":              strings.TrimSpace(setup.HealthState),
		"integration_mode":          strings.TrimSpace(setup.IntegrationMode),
		"integration_id":            strings.TrimSpace(setup.IntegrationID),
		"base_domain":               providerBaseDomain(setup.BaseURL),
		"config_schema_ref":         strings.TrimSpace(setup.ConfigSchemaRef),
		"workflow_ref":              strings.TrimSpace(setup.WorkflowReference),
		"cabinet_tool_authority":    "none",
		"cabinet_database_access":   "none",
		"cabinet_filesystem_access": "none",
		"governed_dispatch_owner":   "cabinet",
	}
	for key, value := range metadata {
		if value == "" {
			delete(metadata, key)
		}
	}
	return metadata
}

func providerBaseDomain(baseURL string) string {
	trimmed := strings.TrimSpace(baseURL)
	if trimmed == "" {
		return ""
	}
	trimmed = strings.TrimPrefix(trimmed, "https://")
	trimmed = strings.TrimPrefix(trimmed, "http://")
	if slash := strings.Index(trimmed, "/"); slash >= 0 {
		trimmed = trimmed[:slash]
	}
	return trimmed
}

func (s *Service) SuggestFromPhoto(ctx context.Context, apiKey, imageURL string) (Suggestion, error) {
	// v1 behavior: route through title extraction prompt with photo URL context
	return s.SuggestFromTitle(ctx, apiKey, "Photo URL: "+imageURL)
}

func (s *Service) ProposeOperation(ctx context.Context, apiKey, requestText string) (OperationProposal, error) {
	if strings.TrimSpace(apiKey) == "" {
		return OperationProposal{}, fmt.Errorf("api key is required")
	}
	prompt := "Return JSON only with keys action, target, payload, confidence for this request: " + strings.TrimSpace(requestText)
	reqBody := map[string]any{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "system", "content": "You are a strict structured operation planner for inventory and wishlist updates."},
			{"role": "user", "content": prompt},
		},
	}
	raw, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/v1/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return OperationProposal{}, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return OperationProposal{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return OperationProposal{}, fmt.Errorf("propose operation failed: status %d", resp.StatusCode)
	}
	var payload struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return OperationProposal{}, err
	}
	if len(payload.Choices) == 0 {
		return OperationProposal{}, fmt.Errorf("no AI choices")
	}
	content := strings.TrimSpace(payload.Choices[0].Message.Content)
	content = strings.Trim(content, "`")
	content = strings.TrimPrefix(content, "json")
	content = strings.TrimSpace(content)
	var out OperationProposal
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return OperationProposal{}, err
	}
	if strings.TrimSpace(out.Action) == "" || strings.TrimSpace(out.Target) == "" {
		return OperationProposal{}, fmt.Errorf("invalid operation proposal")
	}
	if out.Payload == nil {
		out.Payload = map[string]any{}
	}
	out.RequiresConfirmation = true
	return out, nil
}

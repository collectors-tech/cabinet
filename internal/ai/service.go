package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	service  *Service
	resolver AssistantProviderSetupResolver
}

func NewOpenAIAssistantProvider(service *Service, resolver AssistantProviderSetupResolver) *OpenAIAssistantProvider {
	if service == nil {
		service = NewService(Config{})
	}
	return &OpenAIAssistantProvider{service: service, resolver: resolver}
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
	text, err := p.service.runAssistantChatCompletion(ctx, apiKey, respBase.Model, req.Messages)
	if err != nil {
		respBase.ErrorClass = "provider_failure"
		respBase.SetupNextAction = "retry_openai_assistant_turn"
		return respBase, err
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
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("assistant turn failed: status %d", resp.StatusCode)
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
		supported = []string{"gpt-4o-mini", "gpt-4.1-mini", "gpt-5.3-codex"}
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
		"provider_id":        "openai",
		"active_auth_method": strings.TrimSpace(setup.ActiveAuthMethod),
		"health_state":       strings.TrimSpace(setup.HealthState),
		"integration_mode":   strings.TrimSpace(setup.IntegrationMode),
		"integration_id":     strings.TrimSpace(setup.IntegrationID),
		"config_schema_ref":  strings.TrimSpace(setup.ConfigSchemaRef),
		"workflow_ref":       strings.TrimSpace(setup.WorkflowReference),
	}
	for key, value := range metadata {
		if value == "" {
			delete(metadata, key)
		}
	}
	return metadata
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

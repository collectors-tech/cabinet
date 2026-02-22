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

type Service struct {
	baseURL string
	client  *http.Client
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

func (s *Service) SuggestFromPhoto(ctx context.Context, apiKey, imageURL string) (Suggestion, error) {
	// v1 behavior: route through title extraction prompt with photo URL context
	return s.SuggestFromTitle(ctx, apiKey, "Photo URL: "+imageURL)
}

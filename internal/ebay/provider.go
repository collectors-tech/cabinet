package ebay

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/collectors-tech/cabinet/internal/scanner"
)

type ProviderConfig struct {
	BaseURL      string
	BearerToken  string
	Marketplace  string
	HealthWindow int
}

type Provider struct {
	baseURL     string
	bearerToken string
	marketplace string
	client      *http.Client
}

func NewProvider(cfg ProviderConfig) *Provider {
	base := strings.TrimSpace(cfg.BaseURL)
	if base == "" {
		base = "https://api.ebay.com"
	}
	market := strings.TrimSpace(cfg.Marketplace)
	if market == "" {
		market = "EBAY_US"
	}
	return &Provider{
		baseURL:     strings.TrimRight(base, "/"),
		bearerToken: strings.TrimSpace(cfg.BearerToken),
		marketplace: market,
		client:      &http.Client{Timeout: 20 * time.Second},
	}
}

func (p *Provider) Search(ctx context.Context, q scanner.QuerySet) ([]scanner.CandidateInput, error) {
	if p.bearerToken == "" {
		return nil, fmt.Errorf("missing ebay bearer token")
	}
	terms := strings.Join(q.Keywords, " ")
	if terms == "" {
		return nil, fmt.Errorf("keywords are required")
	}
	v := url.Values{}
	v.Set("q", terms)
	if q.MaxPrice > 0 {
		v.Set("filter", fmt.Sprintf("price:[..%s]", strconv.FormatFloat(q.MaxPrice, 'f', 2, 64)))
	}
	if len(q.Exclusions) > 0 {
		v.Set("exclude", strings.Join(q.Exclusions, ","))
	}
	u := p.baseURL + "/buy/browse/v1/item_summary/search?" + v.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.bearerToken)
	req.Header.Set("X-EBAY-C-MARKETPLACE-ID", p.marketplace)
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request ebay: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ebay search status: %d", resp.StatusCode)
	}

	var payload struct {
		ItemSummaries []struct {
			ItemID string `json:"itemId"`
			Title  string `json:"title"`
			Price  struct {
				Value string `json:"value"`
			} `json:"price"`
			ItemWebURL string `json:"itemWebUrl"`
			Image      struct {
				ImageURL string `json:"imageUrl"`
			} `json:"image"`
			Seller struct {
				Username string `json:"username"`
			} `json:"seller"`
		} `json:"itemSummaries"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode ebay response: %w", err)
	}

	out := make([]scanner.CandidateInput, 0, len(payload.ItemSummaries))
	for _, it := range payload.ItemSummaries {
		price, _ := strconv.ParseFloat(it.Price.Value, 64)
		out = append(out, scanner.CandidateInput{
			ListingID: it.ItemID,
			Title:     it.Title,
			Price:     price,
			URL:       it.ItemWebURL,
			Image:     it.Image.ImageURL,
			Seller:    it.Seller.Username,
			Source:    "ebay",
		})
	}
	return out, nil
}

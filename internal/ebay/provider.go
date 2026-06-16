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

type ProviderError struct {
	StatusCode int
	ErrorCode  string
	Message    string
}

func (e *ProviderError) Error() string {
	if e == nil {
		return ""
	}
	if strings.TrimSpace(e.Message) != "" {
		return e.Message
	}
	if strings.TrimSpace(e.ErrorCode) != "" {
		return e.ErrorCode
	}
	return "ebay provider error"
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
		return nil, &ProviderError{StatusCode: http.StatusUnauthorized, ErrorCode: "PROVIDER_AUTH_MISSING", Message: "missing ebay bearer token"}
	}
	keywords := compactSearchTerms(q.Keywords)
	terms := strings.Join(keywords, " ")
	if terms == "" {
		return nil, fmt.Errorf("keywords are required")
	}
	v := url.Values{}
	v.Set("q", terms)
	if q.MaxPrice > 0 {
		v.Set("filter", fmt.Sprintf("price:[..%s]", strconv.FormatFloat(q.MaxPrice, 'f', 2, 64)))
	}
	exclusions := compactSearchTerms(q.Exclusions)
	if len(exclusions) > 0 {
		v.Set("exclude", strings.Join(exclusions, ","))
	}
	if q.ItemsPerPage > 0 {
		v.Set("limit", strconv.Itoa(q.ItemsPerPage))
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
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, &ProviderError{
			StatusCode: http.StatusUnauthorized,
			ErrorCode:  "PROVIDER_AUTH_INVALID",
			Message:    ebayErrorMessage(resp, fmt.Sprintf("ebay credentials rejected with status %d", resp.StatusCode)),
		}
	}
	if resp.StatusCode >= 300 {
		return nil, &ProviderError{StatusCode: resp.StatusCode, ErrorCode: "PROVIDER_SEARCH_FAILED", Message: browseErrorMessage(resp)}
	}

	var payload struct {
		ItemSummaries []struct {
			ItemID string `json:"itemId"`
			Title  string `json:"title"`
			Price  struct {
				Value    string `json:"value"`
				Currency string `json:"currency"`
			} `json:"price"`
			ItemWebURL string `json:"itemWebUrl"`
			Image      struct {
				ImageURL string `json:"imageUrl"`
			} `json:"image"`
			Seller struct {
				Username string `json:"username"`
			} `json:"seller"`
			EstimatedAvailabilities []struct {
				Status   string `json:"estimatedAvailabilityStatus"`
				Quantity int    `json:"estimatedAvailableQuantity"`
			} `json:"estimatedAvailabilities"`
		} `json:"itemSummaries"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode ebay response: %w", err)
	}

	out := make([]scanner.CandidateInput, 0, len(payload.ItemSummaries))
	for _, it := range payload.ItemSummaries {
		price, _ := strconv.ParseFloat(it.Price.Value, 64)
		stockState, stockCount := normalizeAvailability(it.EstimatedAvailabilities)
		out = append(out, scanner.CandidateInput{
			ListingID:  it.ItemID,
			Title:      it.Title,
			Price:      price,
			Currency:   strings.ToUpper(strings.TrimSpace(it.Price.Currency)),
			URL:        it.ItemWebURL,
			Image:      it.Image.ImageURL,
			Seller:     it.Seller.Username,
			Source:     "ebay",
			StockState: stockState,
			StockCount: stockCount,
		})
	}
	return out, nil
}

func browseErrorMessage(resp *http.Response) string {
	return ebayErrorMessage(resp, fmt.Sprintf("ebay search status: %d", resp.StatusCode))
}

func ebayErrorMessage(resp *http.Response, statusMessage string) string {
	var payload struct {
		Errors []struct {
			ErrorID     int    `json:"errorId"`
			Domain      string `json:"domain"`
			Category    string `json:"category"`
			Message     string `json:"message"`
			LongMessage string `json:"longMessage"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil || len(payload.Errors) == 0 {
		return statusMessage
	}
	parts := make([]string, 0, len(payload.Errors))
	for _, ebayErr := range payload.Errors {
		details := []string{}
		if ebayErr.ErrorID != 0 {
			details = append(details, strconv.Itoa(ebayErr.ErrorID))
		}
		if domain := strings.TrimSpace(ebayErr.Domain); domain != "" {
			details = append(details, domain)
		}
		if category := strings.TrimSpace(ebayErr.Category); category != "" {
			details = append(details, category)
		}
		if message := strings.TrimSpace(ebayErr.Message); message != "" {
			details = append(details, message)
		}
		if longMessage := strings.TrimSpace(ebayErr.LongMessage); longMessage != "" {
			details = append(details, longMessage)
		}
		if len(details) > 0 {
			parts = append(parts, strings.Join(details, " | "))
		}
	}
	if len(parts) == 0 {
		return statusMessage
	}
	return statusMessage + ": " + strings.Join(parts, "; ")
}

func compactSearchTerms(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		term := strings.TrimSpace(value)
		if term == "" {
			continue
		}
		out = append(out, term)
	}
	return out
}

func normalizeAvailability(items []struct {
	Status   string `json:"estimatedAvailabilityStatus"`
	Quantity int    `json:"estimatedAvailableQuantity"`
}) (string, int) {
	if len(items) == 0 {
		return "", -1
	}
	first := items[0]
	status := strings.ToUpper(strings.TrimSpace(first.Status))
	count := first.Quantity
	switch status {
	case "IN_STOCK", "AVAILABLE", "LIMITED_STOCK":
		if count == 0 {
			return "in_stock", -1
		}
		if count > 0 && count <= 3 {
			return "low_stock", count
		}
		return "in_stock", count
	case "OUT_OF_STOCK", "SOLD_OUT":
		return "out_of_stock", 0
	default:
		if count > 0 {
			return "in_stock", count
		}
		return "unknown", -1
	}
}

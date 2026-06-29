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
	StatusCode        int
	ErrorCode         string
	Message           string
	RetryAfterSeconds int
}

const browseMaxLimit = 200

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

func (e *ProviderError) RetryAfter() int {
	if e == nil {
		return 0
	}
	return e.RetryAfterSeconds
}

func (e *ProviderError) ProviderStatusCode() int {
	if e == nil {
		return 0
	}
	return e.StatusCode
}

func (e *ProviderError) ProviderErrorCode() string {
	if e == nil {
		return ""
	}
	return e.ErrorCode
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
	market = strings.ToUpper(market)
	return &Provider{
		baseURL:     strings.TrimRight(base, "/"),
		bearerToken: strings.TrimSpace(cfg.BearerToken),
		marketplace: market,
		client:      &http.Client{Timeout: 20 * time.Second},
	}
}

func (p *Provider) ProviderID() string {
	return "ebay"
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
	filters := []string{}
	if q.MaxPrice > 0 {
		filters = append(filters,
			fmt.Sprintf("price:[..%s]", strconv.FormatFloat(q.MaxPrice, 'f', 2, 64)),
			"priceCurrency:"+browseCurrency(q.Region, p.marketplace),
		)
	}
	if country := browseCountry(q.Region); country != "" {
		filters = append(filters, "itemLocationCountry:"+country)
	}
	if condition := browseCondition(q.Condition); condition != "" {
		filters = append(filters, "conditions:{"+condition+"}")
	}
	if len(filters) > 0 {
		v.Set("filter", strings.Join(filters, ","))
	}
	exclusions := compactSearchTerms(q.Exclusions)
	if len(exclusions) > 0 {
		v.Set("exclude", strings.Join(exclusions, ","))
	}
	if limit := browseLimit(q.ItemsPerPage); limit > 0 {
		v.Set("limit", strconv.Itoa(limit))
	}
	v.Set("fieldgroups", "EXTENDED")
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
			StatusCode:        resp.StatusCode,
			ErrorCode:         "PROVIDER_AUTH_INVALID",
			Message:           ebayErrorMessage(resp, fmt.Sprintf("ebay credentials rejected with status %d", resp.StatusCode)),
			RetryAfterSeconds: retryAfterSeconds(resp),
		}
	}
	if resp.StatusCode >= 300 {
		return nil, &ProviderError{
			StatusCode:        resp.StatusCode,
			ErrorCode:         "PROVIDER_SEARCH_FAILED",
			Message:           browseErrorMessage(resp),
			RetryAfterSeconds: retryAfterSeconds(resp),
		}
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
			ShippingOptions []struct {
				ShippingCost struct {
					Value    string `json:"value"`
					Currency string `json:"currency"`
				} `json:"shippingCost"`
			} `json:"shippingOptions"`
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
	seenListingIDs := map[string]struct{}{}
	for _, it := range payload.ItemSummaries {
		listingID := strings.TrimSpace(it.ItemID)
		title := strings.TrimSpace(it.Title)
		itemURL := strings.TrimSpace(it.ItemWebURL)
		if listingID == "" || title == "" || itemURL == "" || !isWebURL(itemURL) {
			continue
		}
		if _, seen := seenListingIDs[listingID]; seen {
			continue
		}
		price, err := strconv.ParseFloat(strings.TrimSpace(it.Price.Value), 64)
		if err != nil || price <= 0 {
			continue
		}
		currency := strings.ToUpper(strings.TrimSpace(it.Price.Currency))
		if currency == "" {
			continue
		}
		seenListingIDs[listingID] = struct{}{}
		shipping := normalizeShippingCost(it.ShippingOptions)
		stockState, stockCount := normalizeAvailability(it.EstimatedAvailabilities)
		out = append(out, scanner.CandidateInput{
			ListingID:  listingID,
			Title:      title,
			Price:      price,
			Currency:   currency,
			Shipping:   shipping,
			URL:        itemURL,
			Image:      normalizeOptionalWebURL(it.Image.ImageURL),
			Seller:     normalizeSellerUsername(it.Seller.Username),
			Source:     "ebay",
			StockState: stockState,
			StockCount: stockCount,
		})
	}
	return out, nil
}

func isWebURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	scheme := strings.ToLower(parsed.Scheme)
	return parsed.Host != "" && (scheme == "http" || scheme == "https")
}

func normalizeOptionalWebURL(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" || !isWebURL(value) {
		return ""
	}
	return value
}

func normalizeSellerUsername(raw string) string {
	seller := strings.TrimSpace(raw)
	if seller == "" {
		return "ebay"
	}
	return seller
}

func normalizeShippingCost(options []struct {
	ShippingCost struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"shippingCost"`
}) float64 {
	for _, option := range options {
		raw := strings.TrimSpace(option.ShippingCost.Value)
		if raw == "" {
			continue
		}
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil || value <= 0 {
			continue
		}
		return value
	}
	return 0
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

func retryAfterSeconds(resp *http.Response) int {
	if resp == nil {
		return 0
	}
	raw := strings.TrimSpace(resp.Header.Get("Retry-After"))
	if raw == "" {
		return 0
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return 0
	}
	return seconds
}

func browseCountry(region string) string {
	region = strings.ToUpper(strings.TrimSpace(region))
	if len(region) == 2 {
		return region
	}
	return ""
}

func browseCondition(condition string) string {
	switch strings.ToLower(strings.TrimSpace(condition)) {
	case "new", "mint", "sealed":
		return "NEW"
	case "used", "loose", "opened":
		return "USED"
	case "unspecified":
		return "UNSPECIFIED"
	default:
		return ""
	}
}

func browseCurrency(region, marketplace string) string {
	marketplace = strings.ToUpper(strings.TrimSpace(marketplace))
	if after, ok := strings.CutPrefix(marketplace, "EBAY_"); ok {
		if currency := currencyForCountry(after); currency != "" {
			return currency
		}
	}
	if country := browseCountry(region); country != "" {
		if currency := currencyForCountry(country); currency != "" {
			return currency
		}
	}
	return "USD"
}

func browseLimit(itemsPerPage int) int {
	if itemsPerPage <= 0 {
		return 0
	}
	if itemsPerPage > browseMaxLimit {
		return browseMaxLimit
	}
	return itemsPerPage
}

func currencyForCountry(country string) string {
	switch strings.ToUpper(strings.TrimSpace(country)) {
	case "AU":
		return "AUD"
	case "CA":
		return "CAD"
	case "GB", "UK":
		return "GBP"
	case "DE", "FR", "IT", "ES", "IE", "NL", "AT", "BE":
		return "EUR"
	case "US":
		return "USD"
	default:
		return ""
	}
}

func normalizeAvailability(items []struct {
	Status   string `json:"estimatedAvailabilityStatus"`
	Quantity int    `json:"estimatedAvailableQuantity"`
}) (string, int) {
	if len(items) == 0 {
		return "", -1
	}
	for _, item := range items {
		status, count := normalizeAvailabilityEntry(item.Status, item.Quantity)
		if status != "unknown" || count >= 0 {
			return status, count
		}
	}
	return "unknown", -1
}

func normalizeAvailabilityEntry(rawStatus string, count int) (string, int) {
	status := strings.ToUpper(strings.TrimSpace(rawStatus))
	if count < 0 {
		count = -1
	}
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

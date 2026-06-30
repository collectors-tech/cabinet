package ebay

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
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
	base := normalizeProviderBaseURL(cfg.BaseURL)
	market := strings.TrimSpace(cfg.Marketplace)
	if market == "" {
		market = "EBAY_US"
	}
	market = normalizeMarketplaceID(market)
	return &Provider{
		baseURL:     strings.TrimRight(base, "/"),
		bearerToken: strings.TrimSpace(cfg.BearerToken),
		marketplace: market,
		client:      &http.Client{Timeout: 20 * time.Second},
	}
}

func normalizeProviderBaseURL(raw string) string {
	value := strings.TrimRight(strings.TrimSpace(raw), "/")
	if value == "" || containsRawControlByte(value) || containsEncodedControlByte(value) {
		return "https://api.ebay.com"
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "https://api.ebay.com"
	}
	scheme := strings.ToLower(parsed.Scheme)
	if parsed.Host == "" || parsed.User != nil || (scheme != "http" && scheme != "https") {
		return "https://api.ebay.com"
	}
	return value
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
		return nil, &ProviderError{
			StatusCode: http.StatusBadGateway,
			ErrorCode:  "PROVIDER_SEARCH_FAILED",
			Message:    fmt.Sprintf("decode ebay Browse response: %v", err),
		}
	}

	out := make([]scanner.CandidateInput, 0, len(payload.ItemSummaries))
	seenListingIDs := map[string]struct{}{}
	for _, it := range payload.ItemSummaries {
		listingID := normalizeRequiredText(it.ItemID)
		title := normalizeRequiredText(it.Title)
		itemURL := strings.TrimSpace(it.ItemWebURL)
		if listingID == "" || title == "" || itemURL == "" || !isWebURL(itemURL) {
			continue
		}
		if _, seen := seenListingIDs[listingID]; seen {
			continue
		}
		price, err := parseBrowseAmount(it.Price.Value)
		if err != nil || price <= 0 {
			continue
		}
		if q.MaxPrice > 0 && price > q.MaxPrice {
			continue
		}
		currency, ok := normalizeCurrencyCode(it.Price.Currency)
		if !ok {
			continue
		}
		seenListingIDs[listingID] = struct{}{}
		shipping := normalizeShippingCost(it.ShippingOptions, currency)
		stockState, stockCount := normalizeAvailability(it.EstimatedAvailabilities)
		out = append(out, scanner.CandidateInput{
			ListingID:  listingID,
			Title:      title,
			Price:      price,
			Currency:   currency,
			Shipping:   shipping,
			URL:        itemURL,
			Image:      normalizeOptionalWebURL(it.Image.ImageURL),
			Seller:     normalizeOptionalText(it.Seller.Username, "ebay"),
			Source:     "ebay",
			StockState: stockState,
			StockCount: stockCount,
		})
	}
	return out, nil
}

func isWebURL(raw string) bool {
	if containsRawControlByte(raw) || containsEncodedControlByte(raw) {
		return false
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	scheme := strings.ToLower(parsed.Scheme)
	return parsed.Host != "" && parsed.User == nil && (scheme == "http" || scheme == "https")
}

func containsRawControlByte(raw string) bool {
	for i := 0; i < len(raw); i++ {
		if raw[i] < 0x20 || raw[i] == 0x7f {
			return true
		}
	}
	return false
}

func containsEncodedControlByte(raw string) bool {
	for i := 0; i+2 < len(raw); i++ {
		if raw[i] != '%' {
			continue
		}
		value, err := strconv.ParseUint(raw[i+1:i+3], 16, 8)
		if err == nil && (value < 0x20 || value == 0x7f) {
			return true
		}
	}
	return false
}

func normalizeOptionalWebURL(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" || !isWebURL(value) {
		return ""
	}
	return value
}

func normalizeRequiredText(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" || containsRawControlByte(value) || containsEncodedControlByte(value) {
		return ""
	}
	return value
}

func normalizeOptionalText(raw, fallback string) string {
	value := normalizeRequiredText(raw)
	if value == "" {
		return fallback
	}
	return value
}

func normalizeCurrencyCode(raw string) (string, bool) {
	value := strings.ToUpper(strings.TrimSpace(raw))
	if !isCurrencyCode(value) {
		return "", false
	}
	return value, true
}

func isCurrencyCode(value string) bool {
	if len(value) != 3 {
		return false
	}
	for _, r := range value {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func normalizeShippingCost(options []struct {
	ShippingCost struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"shippingCost"`
}, priceCurrency string) float64 {
	priceCurrency, ok := normalizeCurrencyCode(priceCurrency)
	if !ok {
		return 0
	}
	for _, option := range options {
		raw := strings.TrimSpace(option.ShippingCost.Value)
		if raw == "" {
			continue
		}
		shippingCurrency, ok := normalizeCurrencyCode(option.ShippingCost.Currency)
		if !ok || shippingCurrency != priceCurrency {
			continue
		}
		value, err := parseBrowseAmount(raw)
		if err != nil || value <= 0 {
			continue
		}
		return value
	}
	return 0
}

func parseBrowseAmount(raw string) (float64, error) {
	value := strings.TrimSpace(raw)
	if strings.HasPrefix(value, "+") || strings.ContainsAny(value, "eE") {
		return 0, fmt.Errorf("amount must use plain decimal syntax")
	}
	if !hasValidCommaGrouping(value) {
		return 0, fmt.Errorf("invalid comma-grouped amount")
	}
	if !hasPlainDecimalDigits(value) {
		return 0, fmt.Errorf("amount must include required decimal digits")
	}
	if !hasCurrencyScale(value) {
		return 0, fmt.Errorf("amount must use currency scale")
	}
	amount, err := strconv.ParseFloat(strings.ReplaceAll(value, ",", ""), 64)
	if err != nil {
		return 0, err
	}
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0, fmt.Errorf("amount must be finite")
	}
	return amount, nil
}

func hasValidCommaGrouping(value string) bool {
	if !strings.Contains(value, ",") {
		return true
	}
	integerPart := value
	if decimalIndex := strings.IndexByte(value, '.'); decimalIndex >= 0 {
		integerPart = value[:decimalIndex]
		if strings.Contains(value[decimalIndex+1:], ",") {
			return false
		}
	}
	integerPart = strings.TrimPrefix(integerPart, "-")
	groups := strings.Split(integerPart, ",")
	if len(groups[0]) == 0 || len(groups[0]) > 3 {
		return false
	}
	for _, group := range groups[1:] {
		if len(group) != 3 {
			return false
		}
	}
	return true
}

func hasPlainDecimalDigits(value string) bool {
	normalized := strings.ReplaceAll(value, ",", "")
	normalized = strings.TrimPrefix(normalized, "-")
	if normalized == "" || strings.Count(normalized, ".") > 1 {
		return false
	}
	parts := strings.Split(normalized, ".")
	for _, part := range parts {
		if part == "" {
			return false
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return false
			}
		}
	}
	return true
}

func hasCurrencyScale(value string) bool {
	normalized := strings.ReplaceAll(value, ",", "")
	if decimalIndex := strings.IndexByte(normalized, '.'); decimalIndex >= 0 {
		return len(normalized[decimalIndex+1:]) <= 2
	}
	return true
}

func browseErrorMessage(resp *http.Response) string {
	return ebayErrorMessage(resp, fmt.Sprintf("ebay search status: %d", resp.StatusCode))
}

func ebayErrorMessage(resp *http.Response, statusMessage string) string {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return statusMessage
	}
	var payload struct {
		Errors []struct {
			ErrorID     int    `json:"errorId"`
			Domain      string `json:"domain"`
			Category    string `json:"category"`
			Message     string `json:"message"`
			LongMessage string `json:"longMessage"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &payload); err != nil || len(payload.Errors) == 0 {
		return appendRawProviderBody(statusMessage, body)
	}
	parts := make([]string, 0, len(payload.Errors))
	for _, ebayErr := range payload.Errors {
		details := []string{}
		if ebayErr.ErrorID != 0 {
			details = append(details, strconv.Itoa(ebayErr.ErrorID))
		}
		if domain := normalizeProviderDiagnosticField(ebayErr.Domain); domain != "" {
			details = append(details, domain)
		}
		if category := normalizeProviderDiagnosticField(ebayErr.Category); category != "" {
			details = append(details, category)
		}
		if message := normalizeProviderDiagnosticField(ebayErr.Message); message != "" {
			details = append(details, message)
		}
		if longMessage := normalizeProviderDiagnosticField(ebayErr.LongMessage); longMessage != "" {
			details = append(details, longMessage)
		}
		if len(details) > 0 {
			parts = append(parts, strings.Join(details, " | "))
		}
	}
	if len(parts) == 0 {
		return appendRawProviderBody(statusMessage, body)
	}
	return statusMessage + ": " + strings.Join(parts, "; ")
}

func appendRawProviderBody(statusMessage string, body []byte) string {
	raw := normalizeProviderDiagnosticField(string(body))
	if raw == "" {
		return statusMessage
	}
	const maxProviderBodyDetail = 500
	if len(raw) > maxProviderBodyDetail {
		raw = raw[:maxProviderBodyDetail] + "..."
	}
	return statusMessage + ": " + raw
}

func normalizeProviderDiagnosticField(raw string) string {
	value := strings.Join(strings.Fields(strings.TrimSpace(raw)), " ")
	if value == "" || containsRawControlByte(value) || containsEncodedControlByte(value) {
		return ""
	}
	return value
}

func compactSearchTerms(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		term := strings.TrimSpace(value)
		if term == "" || containsRawControlByte(term) || containsEncodedControlByte(term) {
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
		retryAt, parseErr := http.ParseTime(raw)
		if parseErr != nil {
			return 0
		}
		seconds = int(time.Until(retryAt).Seconds())
		if seconds <= 0 {
			return 0
		}
	}
	return seconds
}

func browseCountry(region string) string {
	region = strings.ToUpper(strings.TrimSpace(region))
	if isCountryCode(region) {
		return region
	}
	return ""
}

func normalizeMarketplaceID(marketplace string) string {
	marketplace = strings.ToUpper(strings.TrimSpace(marketplace))
	if !isMarketplaceID(marketplace) {
		return "EBAY_US"
	}
	return marketplace
}

func isMarketplaceID(marketplace string) bool {
	if !strings.HasPrefix(marketplace, "EBAY_") || len(marketplace) <= len("EBAY_") {
		return false
	}
	for _, r := range marketplace {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return false
	}
	return true
}

func isCountryCode(country string) bool {
	if len(country) != 2 {
		return false
	}
	for _, r := range country {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
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
		return "unknown", -1
	}
}

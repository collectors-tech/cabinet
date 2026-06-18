package ebay

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/scanner"
)

func TestProviderSearchNormalizesCandidates(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected Browse search GET, got %s", r.Method)
		}
		if r.URL.Path != "/buy/browse/v1/item_summary/search" {
			t.Errorf("expected Browse search path, got %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token" {
			t.Errorf("expected bearer auth header, got %q", got)
		}
		if got := r.Header.Get("X-EBAY-C-MARKETPLACE-ID"); got != "EBAY_US" {
			t.Errorf("expected marketplace header EBAY_US, got %q", got)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("expected Accept application/json, got %q", got)
		}
		if got := r.URL.Query().Get("q"); got != "AFX P-1" {
			t.Errorf("expected joined search keywords, got %q", got)
		}
		if got := r.URL.Query().Get("filter"); got != "price:[..100.00]" {
			t.Errorf("expected max-price filter, got %q", got)
		}
		if got := r.URL.Query().Get("exclude"); got != "broken" {
			t.Errorf("expected exclusion query, got %q", got)
		}
		if got := r.URL.Query().Get("limit"); got != "12" {
			t.Errorf("expected Browse limit from effective items_per_page, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|123|0","title":"AFX P-1","price":{"value":"45.00","currency":"USD"},"itemWebUrl":"https://ebay/item/123","image":{"imageUrl":"https://img/123.jpg"},"seller":{"username":"seller1"},"estimatedAvailabilities":[{"estimatedAvailabilityStatus":"LIMITED_STOCK","estimatedAvailableQuantity":2}]}]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{
		BaseURL:      srv.URL,
		BearerToken:  "token",
		Marketplace:  "EBAY_US",
		HealthWindow: 10,
	})
	items, err := p.Search(context.Background(), scanner.QuerySet{
		Keywords:     []string{"AFX", "P-1"},
		Exclusions:   []string{"broken"},
		MaxPrice:     100,
		Region:       "US",
		Condition:    "used",
		ItemsPerPage: 12,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(items) != 1 || items[0].ListingID == "" || items[0].Title == "" || items[0].URL == "" {
		t.Fatalf("unexpected normalized items: %+v", items)
	}
	if items[0].Currency != "USD" {
		t.Fatalf("expected currency USD, got %q", items[0].Currency)
	}
	if items[0].Image != "https://img/123.jpg" {
		t.Fatalf("expected image URL from Browse payload, got %q", items[0].Image)
	}
	if items[0].Seller != "seller1" {
		t.Fatalf("expected seller username seller1, got %q", items[0].Seller)
	}
	if items[0].Source != "ebay" {
		t.Fatalf("expected source ebay, got %q", items[0].Source)
	}
	if items[0].StockState != "low_stock" || items[0].StockCount != 2 {
		t.Fatalf("expected low_stock/2, got %+v", items[0])
	}
}

func TestProviderSearchTrimsBlankCriteriaBeforeBrowseRequest(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("q"); got != "AFX P-1" {
			t.Errorf("expected blank keywords to be removed before joining, got %q", got)
		}
		if got := r.URL.Query().Get("exclude"); got != "broken,rust" {
			t.Errorf("expected blank exclusions to be removed before joining, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: "token", Marketplace: "EBAY_AU"})
	_, err := p.Search(context.Background(), scanner.QuerySet{
		Keywords:   []string{" AFX ", "", "   ", "P-1"},
		Exclusions: []string{" broken ", "", "rust"},
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
}

func TestProviderSearchRejectsOnlyBlankKeywords(t *testing.T) {
	t.Parallel()

	p := NewProvider(ProviderConfig{BaseURL: "https://example.invalid", BearerToken: "token", Marketplace: "EBAY_AU"})
	_, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{" ", "\t"}})
	if err == nil {
		t.Fatal("expected blank keyword validation error")
	}
	if !strings.Contains(err.Error(), "keywords are required") {
		t.Fatalf("expected keywords validation error, got %v", err)
	}
}

func TestProviderSearchNormalizesSparseCandidateMetadata(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|456|0","title":"No Currency Slot Car","price":{"value":"12.50","currency":" aud "},"itemWebUrl":"https://ebay/item/456","seller":{"username":"seller2"},"estimatedAvailabilities":[{"estimatedAvailabilityStatus":"OUT_OF_STOCK","estimatedAvailableQuantity":5}]}]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{
		BaseURL:     srv.URL,
		BearerToken: "token",
		Marketplace: "EBAY_AU",
	})
	items, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"slot", "car"}})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one normalized item, got %+v", items)
	}
	got := items[0]
	if got.Currency != "AUD" {
		t.Fatalf("expected currency trim/uppercase to AUD, got %q", got.Currency)
	}
	if strings.TrimSpace(got.Image) != "" {
		t.Fatalf("expected missing Browse image to remain empty, got %q", got.Image)
	}
	if got.StockState != "out_of_stock" || got.StockCount != 0 {
		t.Fatalf("expected out_of_stock/0, got %+v", got)
	}
}

func TestProviderSearchTrimsBrowseStringMetadata(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":" v1|trimmed|0 ","title":"  Trimmed Slot Car  ","price":{"value":"12.50","currency":" aud "},"itemWebUrl":" https://ebay/item/trimmed ","image":{"imageUrl":" https://img/trimmed.jpg "},"seller":{"username":" seller-trimmed "}}]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{
		BaseURL:     srv.URL,
		BearerToken: "token",
		Marketplace: "EBAY_AU",
	})
	items, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"slot", "car"}})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one normalized item, got %+v", items)
	}
	got := items[0]
	if got.ListingID != "v1|trimmed|0" {
		t.Fatalf("expected trimmed listing id, got %q", got.ListingID)
	}
	if got.Title != "Trimmed Slot Car" {
		t.Fatalf("expected trimmed title, got %q", got.Title)
	}
	if got.Currency != "AUD" {
		t.Fatalf("expected trimmed uppercase currency AUD, got %q", got.Currency)
	}
	if got.URL != "https://ebay/item/trimmed" {
		t.Fatalf("expected trimmed item URL, got %q", got.URL)
	}
	if got.Image != "https://img/trimmed.jpg" {
		t.Fatalf("expected trimmed image URL, got %q", got.Image)
	}
	if got.Seller != "seller-trimmed" {
		t.Fatalf("expected trimmed seller username, got %q", got.Seller)
	}
}

func TestProviderSearchTrimsBrowsePriceValue(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|price-trim|0","title":"Price Trim Slot Car","price":{"value":" 37.50 ","currency":"AUD"},"itemWebUrl":"https://ebay/item/price-trim","seller":{"username":"seller-price"}}]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{
		BaseURL:     srv.URL,
		BearerToken: "token",
		Marketplace: "EBAY_AU",
	})
	items, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"slot", "car"}})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one normalized item, got %+v", items)
	}
	if items[0].Price != 37.50 {
		t.Fatalf("expected trimmed Browse price 37.50, got %+v", items[0])
	}
}

func TestProviderSearchNormalizesShippingCost(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|789|0","title":"Shipping Cost Slot Car","price":{"value":"22.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/789","seller":{"username":"seller3"},"shippingOptions":[{"shippingCost":{"value":"8.75","currency":"AUD"}}]}]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{
		BaseURL:     srv.URL,
		BearerToken: "token",
		Marketplace: "EBAY_AU",
	})
	items, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"slot", "car"}})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one normalized item, got %+v", items)
	}
	if items[0].Shipping != 8.75 {
		t.Fatalf("expected shipping cost 8.75, got %+v", items[0])
	}
}

func TestProviderSearchUsesFirstParseableShippingCost(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|790|0","title":"Fallback Shipping Slot Car","price":{"value":"22.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/790","seller":{"username":"seller3"},"shippingOptions":[{"shippingCost":{"value":"","currency":"AUD"}},{"shippingCost":{"value":"not-a-price","currency":"AUD"}},{"shippingCost":{"value":"11.25","currency":"AUD"}}]}]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{
		BaseURL:     srv.URL,
		BearerToken: "token",
		Marketplace: "EBAY_AU",
	})
	items, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"slot", "car"}})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one normalized item, got %+v", items)
	}
	if items[0].Shipping != 11.25 {
		t.Fatalf("expected first parseable shipping cost 11.25, got %+v", items[0])
	}
}

func TestProviderSearchReturnsActionableAuthError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errors":[{"errorId":1100,"message":"Access token invalid"}]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: "expired-token", Marketplace: "EBAY_AU"})
	_, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"pokemon"}})
	if err == nil {
		t.Fatal("expected auth error")
	}
	var providerErr *ProviderError
	if !errors.As(err, &providerErr) {
		t.Fatalf("expected ProviderError, got %T %v", err, err)
	}
	if providerErr.StatusCode != http.StatusUnauthorized || providerErr.ErrorCode != "PROVIDER_AUTH_INVALID" {
		t.Fatalf("unexpected provider error: %+v", providerErr)
	}
}

func TestProviderSearchPreservesStructuredAuthErrorPayload(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"errors":[{"errorId":1100,"domain":"ACCESS","category":"REQUEST","message":"Access token invalid","longMessage":"Token scope is missing Browse access"}]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: "expired-token", Marketplace: "EBAY_AU"})
	_, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"pokemon"}})
	if err == nil {
		t.Fatal("expected structured auth error")
	}
	var providerErr *ProviderError
	if !errors.As(err, &providerErr) {
		t.Fatalf("expected ProviderError, got %T %v", err, err)
	}
	if providerErr.StatusCode != http.StatusUnauthorized || providerErr.ErrorCode != "PROVIDER_AUTH_INVALID" {
		t.Fatalf("unexpected provider error: %+v", providerErr)
	}
	for _, want := range []string{"1100", "ACCESS", "REQUEST", "Access token invalid", "Token scope is missing Browse access"} {
		if !strings.Contains(providerErr.Message, want) {
			t.Fatalf("expected provider auth error message to preserve %q, got %q", want, providerErr.Message)
		}
	}
}

func TestProviderSearchPreservesStructuredBrowseErrorPayload(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", "90")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"errors":[{"errorId":12001,"domain":"API_BROWSE","category":"REQUEST","message":"Rate limit exceeded","longMessage":"Try again later"}]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: "token", Marketplace: "EBAY_AU"})
	_, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"pokemon"}})
	if err == nil {
		t.Fatal("expected structured Browse error")
	}
	var providerErr *ProviderError
	if !errors.As(err, &providerErr) {
		t.Fatalf("expected ProviderError, got %T %v", err, err)
	}
	if providerErr.StatusCode != http.StatusTooManyRequests || providerErr.ErrorCode != "PROVIDER_SEARCH_FAILED" {
		t.Fatalf("unexpected provider error: %+v", providerErr)
	}
	if providerErr.RetryAfterSeconds != 90 {
		t.Fatalf("expected retry-after seconds to be preserved, got %+v", providerErr)
	}
	for _, want := range []string{"12001", "API_BROWSE", "REQUEST", "Rate limit exceeded", "Try again later"} {
		if !strings.Contains(providerErr.Message, want) {
			t.Fatalf("expected provider error message to preserve %q, got %q", want, providerErr.Message)
		}
	}
}

func TestProviderSearchRequiresBearerToken(t *testing.T) {
	t.Parallel()

	p := NewProvider(ProviderConfig{BaseURL: "https://example.invalid", Marketplace: "EBAY_AU"})
	_, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"pokemon"}})
	if err == nil {
		t.Fatal("expected missing auth error")
	}
	var providerErr *ProviderError
	if !errors.As(err, &providerErr) {
		t.Fatalf("expected ProviderError, got %T %v", err, err)
	}
	if providerErr.StatusCode != http.StatusUnauthorized || providerErr.ErrorCode != "PROVIDER_AUTH_MISSING" {
		t.Fatalf("unexpected provider error: %+v", providerErr)
	}
}

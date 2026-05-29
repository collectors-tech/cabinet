package ebay

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/collectors-tech/cabinet/internal/scanner"
)

func TestProviderSearchNormalizesCandidates(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		Keywords:   []string{"AFX", "P-1"},
		Exclusions: []string{"broken"},
		MaxPrice:   100,
		Region:     "US",
		Condition:  "used",
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
	if items[0].StockState != "low_stock" || items[0].StockCount != 2 {
		t.Fatalf("expected low_stock/2, got %+v", items[0])
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

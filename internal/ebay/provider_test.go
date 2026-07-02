package ebay

import (
	"context"
	"errors"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
		wantFilter := "price:[..100.00],priceCurrency:USD,itemLocationCountry:US,conditions:{USED}"
		if got := r.URL.Query().Get("filter"); got != wantFilter {
			t.Errorf("expected Browse filters %q, got %q", wantFilter, got)
		}
		if got := r.URL.Query().Get("exclude"); got != "broken" {
			t.Errorf("expected exclusion query, got %q", got)
		}
		if got := r.URL.Query().Get("limit"); got != "12" {
			t.Errorf("expected Browse limit from effective items_per_page, got %q", got)
		}
		if got := r.URL.Query().Get("fieldgroups"); got != "EXTENDED" {
			t.Errorf("expected Browse EXTENDED fieldgroup for availability metadata, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|123|0","title":"AFX P-1","price":{"value":"45.00","currency":"USD"},"itemWebUrl":"https://ebay/item/123","image":{"imageUrl":"https://img/123.jpg"},"seller":{"username":"seller1"},"itemCreationDate":"2026-06-30T05:04:03.123Z","itemEndDate":"2026-07-07T15:04:03+10:00","estimatedAvailabilities":[{"estimatedAvailabilityStatus":"LIMITED_STOCK","estimatedAvailableQuantity":2}]}]}`))
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
	if items[0].ListingCreatedAt != "2026-06-30T05:04:03Z" {
		t.Fatalf("expected normalized listing creation timestamp, got %q", items[0].ListingCreatedAt)
	}
	if items[0].ListingUpdatedAt != "2026-07-07T05:04:03Z" {
		t.Fatalf("expected normalized listing update/end timestamp, got %q", items[0].ListingUpdatedAt)
	}
}

func TestProviderSearchDropsUnsafeBrowseTimestamps(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|unsafe-time|0","title":"Unsafe Time Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/unsafe-time","itemCreationDate":"2026-06-30T05:04:03Z%0A","itemEndDate":"not-a-time","seller":{"username":"seller-time"}},{"itemId":"v1|encoded-unsafe-time|0","title":"Encoded Unsafe Time Slot Car","price":{"value":"13.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/encoded-unsafe-time","itemCreationDate":"2026-06-30T05:04:03Z%E2%80%AE","itemEndDate":"2026-07-07T15:04:03%25E2%2580%25AFZ","seller":{"username":"seller-time"}}]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: "token", Marketplace: "EBAY_AU"})
	items, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"slot", "car"}})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected otherwise valid candidate to survive, got %+v", items)
	}
	for _, item := range items {
		if item.ListingCreatedAt != "" || item.ListingUpdatedAt != "" {
			t.Fatalf("expected unsafe/malformed Browse timestamps to be dropped, got %+v", item)
		}
	}
}

func TestProviderSearchBuildsBrowseFiltersFromSavedQueryCriteria(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		marketplace string
		query       scanner.QuerySet
		wantFilter  string
	}{
		{
			name:        "marketplace drives currency while region drives location and condition",
			marketplace: "EBAY_AU",
			query: scanner.QuerySet{
				Keywords:  []string{"slot car"},
				MaxPrice:  75,
				Region:    " au ",
				Condition: "sealed",
			},
			wantFilter: "price:[..75.00],priceCurrency:AUD,itemLocationCountry:AU,conditions:{NEW}",
		},
		{
			name:        "marketplace supplies price currency when region is blank",
			marketplace: "EBAY_GB",
			query: scanner.QuerySet{
				Keywords: []string{"slot car"},
				MaxPrice: 75,
			},
			wantFilter: "price:[..75.00],priceCurrency:GBP",
		},
		{
			name:        "marketplace currency wins when location country differs",
			marketplace: "EBAY_US",
			query: scanner.QuerySet{
				Keywords: []string{"slot car"},
				MaxPrice: 75,
				Region:   "AU",
			},
			wantFilter: "price:[..75.00],priceCurrency:USD,itemLocationCountry:AU",
		},
		{
			name:        "unmapped condition is ignored instead of sending unsupported value",
			marketplace: "EBAY_US",
			query: scanner.QuerySet{
				Keywords:  []string{"slot car"},
				Region:    "US",
				Condition: "collector grade",
			},
			wantFilter: "itemLocationCountry:US",
		},
		{
			name:        "malformed region is ignored instead of sending unsupported country filter",
			marketplace: "EBAY_US",
			query: scanner.QuerySet{
				Keywords: []string{"slot car"},
				MaxPrice: 75,
				Region:   "U1",
			},
			wantFilter: "price:[..75.00],priceCurrency:USD",
		},
		{
			name:        "non-ascii region is ignored instead of sending unsupported country filter",
			marketplace: "EBAY_US",
			query: scanner.QuerySet{
				Keywords: []string{"slot car"},
				MaxPrice: 75,
				Region:   "ÅU",
			},
			wantFilter: "price:[..75.00],priceCurrency:USD",
		},
		{
			name:        "unsafe saved-query region is ignored before trim can make it valid",
			marketplace: "EBAY_US",
			query: scanner.QuerySet{
				Keywords: []string{"slot car"},
				MaxPrice: 75,
				Region:   "\nUS",
			},
			wantFilter: "price:[..75.00],priceCurrency:USD",
		},
		{
			name:        "unsafe saved-query condition is ignored before trim can make it valid",
			marketplace: "EBAY_AU",
			query: scanner.QuerySet{
				Keywords:  []string{"slot car"},
				Region:    "AU",
				Condition: "used\n",
			},
			wantFilter: "itemLocationCountry:AU",
		},
		{
			name:        "encoded unsafe saved-query region and condition are omitted",
			marketplace: "EBAY_US",
			query: scanner.QuerySet{
				Keywords:  []string{"slot car"},
				Region:    "US%0A",
				Condition: "used%2520",
			},
			wantFilter: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("filter"); got != tt.wantFilter {
					t.Errorf("expected Browse filters %q, got %q", tt.wantFilter, got)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"itemSummaries":[]}`))
			}))
			defer srv.Close()

			p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: "token", Marketplace: tt.marketplace})
			if _, err := p.Search(context.Background(), tt.query); err != nil {
				t.Fatalf("Search() error = %v", err)
			}
		})
	}
}

func TestProviderSearchSkipsBrowsePricesAboveSavedQueryMax(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantFilter := "price:[..25.00],priceCurrency:AUD"
		if got := r.URL.Query().Get("filter"); got != wantFilter {
			t.Errorf("expected Browse max-price filter %q, got %q", wantFilter, got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|over-max|0","title":"Over Max Slot Car","price":{"value":"25.01","currency":"AUD"},"itemWebUrl":"https://ebay/item/over-max","seller":{"username":"seller-price"}},{"itemId":"v1|at-max|0","title":"At Max Slot Car","price":{"value":"25.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/at-max","seller":{"username":"seller-price"}},{"itemId":"v1|under-max|0","title":"Under Max Slot Car","price":{"value":"24.99","currency":"AUD"},"itemWebUrl":"https://ebay/item/under-max","seller":{"username":"seller-price"}}]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{
		BaseURL:     srv.URL,
		BearerToken: "token",
		Marketplace: "EBAY_AU",
	})
	items, err := p.Search(context.Background(), scanner.QuerySet{
		Keywords: []string{"slot", "car"},
		MaxPrice: 25,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected only at/under-max items, got %+v", items)
	}
	for _, item := range items {
		if item.ListingID == "v1|over-max|0" || item.Price > 25 {
			t.Fatalf("expected over-threshold candidate to be skipped, got %+v", items)
		}
	}
}

func TestProviderSearchSkipsMaxPriceResultsWithMismatchedCurrency(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantFilter := "price:[..50.00],priceCurrency:AUD"
		if got := r.URL.Query().Get("filter"); got != wantFilter {
			t.Errorf("expected Browse max-price filter %q, got %q", wantFilter, got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|usd-under-max|0","title":"USD Under Max Slot Car","price":{"value":"40.00","currency":"USD"},"itemWebUrl":"https://ebay/item/usd-under-max","seller":{"username":"seller-price"}},{"itemId":"v1|aud-under-max|0","title":"AUD Under Max Slot Car","price":{"value":"45.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/aud-under-max","seller":{"username":"seller-price"}}]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{
		BaseURL:     srv.URL,
		BearerToken: "token",
		Marketplace: "EBAY_AU",
	})
	items, err := p.Search(context.Background(), scanner.QuerySet{
		Keywords: []string{"slot", "car"},
		MaxPrice: 50,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected only matching-currency max-price item, got %+v", items)
	}
	if items[0].ListingID != "v1|aud-under-max|0" || items[0].Currency != "AUD" {
		t.Fatalf("expected AUD matching-currency candidate to survive, got %+v", items[0])
	}
}

func TestProviderSearchOmitsNonFiniteMaxPriceBrowseFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		maxPrice float64
	}{
		{name: "positive infinity", maxPrice: math.Inf(1)},
		{name: "not a number", maxPrice: math.NaN()},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("filter"); got != "" {
					t.Errorf("expected non-finite max-price criteria to omit Browse filter, got %q", got)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|finite-guard|0","title":"Finite Guard Slot Car","price":{"value":"18.50","currency":"AUD"},"itemWebUrl":"https://ebay/item/finite-guard","seller":{"username":"seller-price"}}]}`))
			}))
			defer srv.Close()

			p := NewProvider(ProviderConfig{
				BaseURL:     srv.URL,
				BearerToken: "token",
				Marketplace: "EBAY_AU",
			})
			items, err := p.Search(context.Background(), scanner.QuerySet{
				Keywords: []string{"slot", "car"},
				MaxPrice: tt.maxPrice,
			})
			if err != nil {
				t.Fatalf("Search() error = %v", err)
			}
			if len(items) != 1 || items[0].ListingID != "v1|finite-guard|0" {
				t.Fatalf("expected valid Browse item to survive without non-finite max-price filter, got %+v", items)
			}
		})
	}
}

func TestProviderSearchCanonicalizesConfiguredMarketplace(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-EBAY-C-MARKETPLACE-ID"); got != "EBAY_AU" {
			t.Errorf("expected canonical marketplace header EBAY_AU, got %q", got)
		}
		wantFilter := "price:[..75.00],priceCurrency:AUD"
		if got := r.URL.Query().Get("filter"); got != wantFilter {
			t.Errorf("expected Browse filters %q, got %q", wantFilter, got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: "token", Marketplace: " ebay_au "})
	if _, err := p.Search(context.Background(), scanner.QuerySet{
		Keywords: []string{"slot", "car"},
		MaxPrice: 75,
	}); err != nil {
		t.Fatalf("Search() error = %v", err)
	}
}

func TestProviderSearchFallsBackFromMalformedMarketplace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		marketplace string
	}{
		{name: "control characters", marketplace: "EBAY_AU\r\nX-Injected: 1"},
		{name: "leading control character before valid marketplace", marketplace: "\nEBAY_AU"},
		{name: "encoded control before valid marketplace", marketplace: "%0AEBAY_AU"},
		{name: "unicode format control before valid marketplace", marketplace: string(rune(0x202e)) + "EBAY_AU"},
		{name: "encoded unicode format control before valid marketplace", marketplace: "%E2%80%AEEBAY_AU"},
		{name: "non-ascii", marketplace: "EBAY_ÅU"},
		{name: "missing prefix", marketplace: "AU"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("X-EBAY-C-MARKETPLACE-ID"); got != "EBAY_US" {
					t.Errorf("expected malformed marketplace to fall back to EBAY_US, got %q", got)
				}
				wantFilter := "price:[..75.00],priceCurrency:USD"
				if got := r.URL.Query().Get("filter"); got != wantFilter {
					t.Errorf("expected fallback Browse filters %q, got %q", wantFilter, got)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"itemSummaries":[]}`))
			}))
			defer srv.Close()

			p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: "token", Marketplace: tt.marketplace})
			if _, err := p.Search(context.Background(), scanner.QuerySet{
				Keywords: []string{"slot", "car"},
				MaxPrice: 75,
			}); err != nil {
				t.Fatalf("Search() error = %v", err)
			}
		})
	}
}

func TestNewProviderFallsBackFromUnsafeBaseURLOverrides(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		baseURL string
	}{
		{name: "non-web scheme", baseURL: "javascript:alert(1)"},
		{name: "relative URL", baseURL: "/buy/browse"},
		{name: "embedded userinfo", baseURL: "https://token@api.ebay.com"},
		{name: "query string", baseURL: "https://api.ebay.com?environment=sandbox"},
		{name: "fragment", baseURL: "https://api.ebay.com#browse"},
		{name: "raw control byte", baseURL: "https://api.ebay.com" + string(rune(0x7f))},
		{name: "encoded control byte", baseURL: "https://api.ebay.com/%0A"},
		{name: "encoded space", baseURL: "https://api.ebay.com/custom%20path"},
		{name: "encoded unicode whitespace", baseURL: "https://api.ebay.com/custom%E2%80%AFpath"},
		{name: "encoded unicode format control", baseURL: "https://api.ebay.com/custom%E2%80%AEpath"},
		{name: "malformed percent escape", baseURL: "https://api.ebay.com/custom%ZZpath"},
		{name: "raw whitespace", baseURL: "https://api.ebay.com/custom path"},
		{name: "unicode whitespace", baseURL: "https://api.ebay.com/custom" + string(rune(0x00a0)) + "path"},
		{name: "raw unicode format control", baseURL: "https://api.ebay.com/custom" + string(rune(0x202e)) + "path"},
		{name: "parse failure", baseURL: "http://%zz"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := NewProvider(ProviderConfig{BaseURL: tt.baseURL, BearerToken: "token", Marketplace: "EBAY_AU"})
			if p.baseURL != "https://api.ebay.com" {
				t.Fatalf("expected unsafe base URL override to fall back to production eBay API, got %q", p.baseURL)
			}
		})
	}
}

func TestNewProviderTrimsValidBaseURLOverride(t *testing.T) {
	t.Parallel()

	p := NewProvider(ProviderConfig{BaseURL: " http://127.0.0.1:4567/custom/// ", BearerToken: "token", Marketplace: "EBAY_AU"})
	if p.baseURL != "http://127.0.0.1:4567/custom" {
		t.Fatalf("expected valid controlled base URL override to be trimmed, got %q", p.baseURL)
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

func TestProviderSearchOmitsUnsafeQueryTextBeforeBrowseRequest(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("q"); got != "AFX P-1" {
			t.Errorf("expected unsafe keywords to be omitted before joining, got %q", got)
		}
		if got := r.URL.Query().Get("exclude"); got != "broken" {
			t.Errorf("expected unsafe exclusions to be omitted before joining, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: "token", Marketplace: "EBAY_AU"})
	_, err := p.Search(context.Background(), scanner.QuerySet{
		Keywords:   []string{"AFX", "bad\nkeyword", "P%0A-1", "P-1"},
		Exclusions: []string{"broken", "skip%7Fthis", "skip\rthis"},
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
}

func TestProviderSearchOmitsEncodedUnicodeQueryTextBeforeBrowseRequest(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("q"); got != "AFX P-1" {
			t.Errorf("expected percent-encoded unsafe Unicode keywords to be omitted before joining, got %q", got)
		}
		if got := r.URL.Query().Get("exclude"); got != "broken" {
			t.Errorf("expected percent-encoded unsafe Unicode exclusions to be omitted before joining, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: "token", Marketplace: "EBAY_AU"})
	_, err := p.Search(context.Background(), scanner.QuerySet{
		Keywords:   []string{"AFX", "bad%E2%80%AEkeyword", "wide%E2%80%AFgap", "P-1"},
		Exclusions: []string{"broken", "skip%E2%81%A6this"},
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
}

func TestProviderSearchOmitsEncodedWhitespaceQueryTextBeforeBrowseRequest(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("q"); got != "AFX P-1" {
			t.Errorf("expected percent-encoded whitespace keywords to be omitted before joining, got %q", got)
		}
		if got := r.URL.Query().Get("exclude"); got != "broken" {
			t.Errorf("expected percent-encoded whitespace exclusions to be omitted before joining, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: "token", Marketplace: "EBAY_AU"})
	_, err := p.Search(context.Background(), scanner.QuerySet{
		Keywords:   []string{"AFX", "bad%20keyword", "wide%C2%A0gap", "P-1"},
		Exclusions: []string{"broken", "skip%09this"},
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
}

func TestProviderSearchOmitsNestedEncodedUnsafeQueryTextBeforeBrowseRequest(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("q"); got != "AFX P-1" {
			t.Errorf("expected nested percent-encoded unsafe keywords to be omitted before joining, got %q", got)
		}
		if got := r.URL.Query().Get("exclude"); got != "broken" {
			t.Errorf("expected nested percent-encoded unsafe exclusions to be omitted before joining, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: "token", Marketplace: "EBAY_AU"})
	_, err := p.Search(context.Background(), scanner.QuerySet{
		Keywords:   []string{"AFX", "bad%250Akeyword", "wide%25E2%2580%25AFgap", "P-1"},
		Exclusions: []string{"broken", "skip%2509this", "skip%25E2%2580%25AEthis"},
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
}

func TestProviderSearchOmitsOversizedQueryTextBeforeBrowseRequest(t *testing.T) {
	t.Parallel()

	oversized := strings.Repeat("slot", 200)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("q"); got != "AFX P-1" {
			t.Errorf("expected oversized keywords to be omitted before joining, got %q", got)
		}
		if got := r.URL.Query().Get("exclude"); got != "broken" {
			t.Errorf("expected oversized exclusions to be omitted before joining, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: "token", Marketplace: "EBAY_AU"})
	_, err := p.Search(context.Background(), scanner.QuerySet{
		Keywords:   []string{"AFX", oversized, "P-1"},
		Exclusions: []string{"broken", oversized},
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
}

func TestProviderSearchRejectsOnlyUnsafeKeywords(t *testing.T) {
	t.Parallel()

	p := NewProvider(ProviderConfig{BaseURL: "https://example.invalid", BearerToken: "token", Marketplace: "EBAY_AU"})
	_, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"bad\nkeyword", "bad%0Dkeyword"}})
	if err == nil {
		t.Fatal("expected unsafe keyword validation error")
	}
	if !strings.Contains(err.Error(), "keywords are required") {
		t.Fatalf("expected keywords validation error, got %v", err)
	}
}

func TestProviderSearchRejectsOnlyOversizedKeywords(t *testing.T) {
	t.Parallel()

	p := NewProvider(ProviderConfig{BaseURL: "https://example.invalid", BearerToken: "token", Marketplace: "EBAY_AU"})
	_, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{strings.Repeat("slot", 200)}})
	if err == nil {
		t.Fatal("expected oversized keyword validation error")
	}
	if !strings.Contains(err.Error(), "keywords are required") {
		t.Fatalf("expected keywords validation error, got %v", err)
	}
}

func TestProviderSearchCapsBrowseLimit(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("limit"); got != "200" {
			t.Errorf("expected direct provider Browse limit to be capped at 200, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: "token", Marketplace: "EBAY_AU"})
	_, err := p.Search(context.Background(), scanner.QuerySet{
		Keywords:     []string{"slot", "car"},
		ItemsPerPage: 500,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
}

func TestProviderSearchCapsEmittedCandidatesToEffectiveBrowseLimit(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("limit"); got != "2" {
			t.Errorf("expected Browse limit 2, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|first|0","title":"First Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/first","seller":{"username":"seller-limit"}},{"itemId":"v1|second|0","title":"Second Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/second","seller":{"username":"seller-limit"}},{"itemId":"v1|third|0","title":"Third Slot Car","price":{"value":"13.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/third","seller":{"username":"seller-limit"}}]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: "token", Marketplace: "EBAY_AU"})
	items, err := p.Search(context.Background(), scanner.QuerySet{
		Keywords:     []string{"slot", "car"},
		ItemsPerPage: 2,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected oversized Browse payload to be capped to 2 emitted candidates, got %+v", items)
	}
	if items[0].ListingID != "v1|first|0" || items[1].ListingID != "v1|second|0" {
		t.Fatalf("expected first two valid candidates to survive local page-size cap, got %+v", items)
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

func TestProviderSearchParsesCommaGroupedBrowsePriceValue(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|comma-price|0","title":"Comma Price Slot Car","price":{"value":" 1,237.50 ","currency":"AUD"},"itemWebUrl":"https://ebay/item/comma-price","seller":{"username":"seller-price"}}]}`))
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
	if items[0].Price != 1237.50 {
		t.Fatalf("expected comma-grouped Browse price 1237.50, got %+v", items[0])
	}
}

func TestProviderSearchSkipsUnparseableBrowsePrices(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|bad-price|0","title":"Bad Price Slot Car","price":{"value":"not-a-price","currency":"AUD"},"itemWebUrl":"https://ebay/item/bad-price","seller":{"username":"seller-price"}},{"itemId":"v1|bad-comma-price|0","title":"Bad Comma Price Slot Car","price":{"value":"12,34.50","currency":"AUD"},"itemWebUrl":"https://ebay/item/bad-comma-price","seller":{"username":"seller-price"}},{"itemId":"v1|good-price|0","title":"Good Price Slot Car","price":{"value":"37.50","currency":"AUD"},"itemWebUrl":"https://ebay/item/good-price","seller":{"username":"seller-price"}}]}`))
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
		t.Fatalf("expected only the parseable-price item, got %+v", items)
	}
	if items[0].ListingID != "v1|good-price|0" || items[0].Price != 37.50 {
		t.Fatalf("expected good-price candidate to survive normalization, got %+v", items[0])
	}
}

func TestProviderSearchSkipsDecimalCommaBrowsePrices(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|decimal-comma-price|0","title":"Decimal Comma Price Slot Car","price":{"value":"123.4,5","currency":"AUD"},"itemWebUrl":"https://ebay/item/decimal-comma-price","seller":{"username":"seller-price"}},{"itemId":"v1|good-price|0","title":"Good Price Slot Car","price":{"value":"123.45","currency":"AUD"},"itemWebUrl":"https://ebay/item/good-price","seller":{"username":"seller-price"}}]}`))
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
		t.Fatalf("expected only the valid price item, got %+v", items)
	}
	if items[0].ListingID != "v1|good-price|0" || items[0].Price != 123.45 {
		t.Fatalf("expected good-price candidate to survive normalization, got %+v", items[0])
	}
}

func TestProviderSearchSkipsNonPositiveBrowsePrices(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|zero-price|0","title":"Zero Price Slot Car","price":{"value":"0","currency":"AUD"},"itemWebUrl":"https://ebay/item/zero-price","seller":{"username":"seller-price"}},{"itemId":"v1|negative-price|0","title":"Negative Price Slot Car","price":{"value":"-1.50","currency":"AUD"},"itemWebUrl":"https://ebay/item/negative-price","seller":{"username":"seller-price"}},{"itemId":"v1|positive-price|0","title":"Positive Price Slot Car","price":{"value":"1.50","currency":"AUD"},"itemWebUrl":"https://ebay/item/positive-price","seller":{"username":"seller-price"}}]}`))
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
		t.Fatalf("expected only the positive-price item, got %+v", items)
	}
	if items[0].ListingID != "v1|positive-price|0" || items[0].Price != 1.50 {
		t.Fatalf("expected positive-price candidate to survive normalization, got %+v", items[0])
	}
}

func TestProviderSearchSkipsNonFiniteBrowsePrices(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|nan-price|0","title":"NaN Price Slot Car","price":{"value":"NaN","currency":"AUD"},"itemWebUrl":"https://ebay/item/nan-price","seller":{"username":"seller-price"}},{"itemId":"v1|infinite-price|0","title":"Infinite Price Slot Car","price":{"value":"Infinity","currency":"AUD"},"itemWebUrl":"https://ebay/item/infinite-price","seller":{"username":"seller-price"}},{"itemId":"v1|finite-price|0","title":"Finite Price Slot Car","price":{"value":"18.50","currency":"AUD"},"itemWebUrl":"https://ebay/item/finite-price","seller":{"username":"seller-price"}}]}`))
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
		t.Fatalf("expected only the finite-price item, got %+v", items)
	}
	if items[0].ListingID != "v1|finite-price|0" || items[0].Price != 18.50 {
		t.Fatalf("expected finite-price candidate to survive normalization, got %+v", items[0])
	}
}

func TestProviderSearchSkipsNonPlainDecimalBrowsePrices(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|exponent-price|0","title":"Exponent Price Slot Car","price":{"value":"1e3","currency":"AUD"},"itemWebUrl":"https://ebay/item/exponent-price","seller":{"username":"seller-price"}},{"itemId":"v1|plus-price|0","title":"Plus Price Slot Car","price":{"value":"+12.50","currency":"AUD"},"itemWebUrl":"https://ebay/item/plus-price","seller":{"username":"seller-price"}},{"itemId":"v1|plain-price|0","title":"Plain Price Slot Car","price":{"value":"12.50","currency":"AUD"},"itemWebUrl":"https://ebay/item/plain-price","seller":{"username":"seller-price"}}]}`))
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
		t.Fatalf("expected only the plain decimal price item, got %+v", items)
	}
	if items[0].ListingID != "v1|plain-price|0" || items[0].Price != 12.50 {
		t.Fatalf("expected plain-price candidate to survive normalization, got %+v", items[0])
	}
}

func TestProviderSearchSkipsPartialDecimalBrowsePrices(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|leading-dot-price|0","title":"Leading Dot Price Slot Car","price":{"value":".50","currency":"AUD"},"itemWebUrl":"https://ebay/item/leading-dot-price","seller":{"username":"seller-price"}},{"itemId":"v1|trailing-dot-price|0","title":"Trailing Dot Price Slot Car","price":{"value":"12.","currency":"AUD"},"itemWebUrl":"https://ebay/item/trailing-dot-price","seller":{"username":"seller-price"}},{"itemId":"v1|strict-price|0","title":"Strict Price Slot Car","price":{"value":"12.50","currency":"AUD"},"itemWebUrl":"https://ebay/item/strict-price","seller":{"username":"seller-price"}}]}`))
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
		t.Fatalf("expected only the strict decimal price item, got %+v", items)
	}
	if items[0].ListingID != "v1|strict-price|0" || items[0].Price != 12.50 {
		t.Fatalf("expected strict-price candidate to survive normalization, got %+v", items[0])
	}
}

func TestProviderSearchSkipsOverPrecisionBrowsePrices(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|three-decimal-price|0","title":"Three Decimal Price Slot Car","price":{"value":"12.345","currency":"AUD"},"itemWebUrl":"https://ebay/item/three-decimal-price","seller":{"username":"seller-price"}},{"itemId":"v1|valid-scale-price|0","title":"Valid Scale Price Slot Car","price":{"value":"12.30","currency":"AUD"},"itemWebUrl":"https://ebay/item/valid-scale-price","seller":{"username":"seller-price"}}]}`))
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
		t.Fatalf("expected only the two-decimal price item, got %+v", items)
	}
	if items[0].ListingID != "v1|valid-scale-price|0" || items[0].Price != 12.30 {
		t.Fatalf("expected valid-scale-price candidate to survive normalization, got %+v", items[0])
	}
}

func TestProviderSearchSkipsBlankBrowsePriceCurrency(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|missing-currency|0","title":"Missing Currency Slot Car","price":{"value":"11.00","currency":" "},"itemWebUrl":"https://ebay/item/missing-currency","seller":{"username":"seller-currency"}},{"itemId":"v1|required-currency|0","title":"Required Currency Slot Car","price":{"value":"14.00","currency":" aud "},"itemWebUrl":"https://ebay/item/required-currency","seller":{"username":"seller-currency"}}]}`))
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
		t.Fatalf("expected only candidate with normalized currency to survive, got %+v", items)
	}
	if items[0].ListingID != "v1|required-currency|0" || items[0].Currency != "AUD" {
		t.Fatalf("expected required-currency candidate to survive with normalized AUD currency, got %+v", items[0])
	}
}

func TestProviderSearchSkipsMalformedBrowsePriceCurrency(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|symbol-currency|0","title":"Symbol Currency Slot Car","price":{"value":"11.00","currency":"AU$"},"itemWebUrl":"https://ebay/item/symbol-currency","seller":{"username":"seller-currency"}},{"itemId":"v1|long-currency|0","title":"Long Currency Slot Car","price":{"value":"12.00","currency":"AUDD"},"itemWebUrl":"https://ebay/item/long-currency","seller":{"username":"seller-currency"}},{"itemId":"v1|digit-currency|0","title":"Digit Currency Slot Car","price":{"value":"13.00","currency":"AU1"},"itemWebUrl":"https://ebay/item/digit-currency","seller":{"username":"seller-currency"}},{"itemId":"v1|non-ascii-currency|0","title":"Non ASCII Currency Slot Car","price":{"value":"13.50","currency":"ＡＵＤ"},"itemWebUrl":"https://ebay/item/non-ascii-currency","seller":{"username":"seller-currency"}},{"itemId":"v1|required-currency|0","title":"Required Currency Slot Car","price":{"value":"14.00","currency":" aud "},"itemWebUrl":"https://ebay/item/required-currency","seller":{"username":"seller-currency"}}]}`))
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
		t.Fatalf("expected only candidate with a valid currency code to survive, got %+v", items)
	}
	if items[0].ListingID != "v1|required-currency|0" || items[0].Currency != "AUD" {
		t.Fatalf("expected required-currency candidate to survive with normalized AUD currency, got %+v", items[0])
	}
}

func TestProviderSearchSkipsIncompleteBrowseItemSummaries(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"","title":"Missing ID Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/missing-id","seller":{"username":"seller-required"}},{"itemId":"v1|missing-title|0","title":"   ","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/missing-title","seller":{"username":"seller-required"}},{"itemId":"v1|missing-url|0","title":"Missing URL Slot Car","price":{"value":"13.00","currency":"AUD"},"itemWebUrl":" ","seller":{"username":"seller-required"}},{"itemId":"v1|required-fields|0","title":"Required Fields Slot Car","price":{"value":"14.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/required-fields","seller":{"username":"seller-required"}}]}`))
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
		t.Fatalf("expected only complete item summary to survive, got %+v", items)
	}
	if items[0].ListingID != "v1|required-fields|0" || items[0].Title != "Required Fields Slot Car" || items[0].URL != "https://ebay/item/required-fields" {
		t.Fatalf("expected required-fields candidate to survive normalization, got %+v", items[0])
	}
}

func TestProviderSearchSkipsBrowseTextFieldsWithRawControlCharacters(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|raw` + string(rune(0x7f)) + `id|0","title":"Raw ID Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/raw-id","seller":{"username":"seller-text"}},{"itemId":"v1|raw-title|0","title":"Raw` + string(rune(0x7f)) + ` Title Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/raw-title","seller":{"username":"seller-text"}},{"itemId":"v1|valid-text|0","title":"Valid Text Slot Car","price":{"value":"13.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/valid-text","seller":{"username":"seller-text"}}]}`))
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
		t.Fatalf("expected only candidate with safe required text fields to survive, got %+v", items)
	}
	if items[0].ListingID != "v1|valid-text|0" || items[0].Title != "Valid Text Slot Car" {
		t.Fatalf("expected valid-text candidate to survive text control-byte guard, got %+v", items[0])
	}
}

func TestProviderSearchSkipsBrowseListingIDsWithEmbeddedWhitespace(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|space id|0","title":"Space ID Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/space-id","seller":{"username":"seller-id"}},{"itemId":"v1|unicode` + string(rune(0x00a0)) + `id|0","title":"Unicode Space ID Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/unicode-space-id","seller":{"username":"seller-id"}},{"itemId":" v1|valid-id|0 ","title":"Valid ID Slot Car","price":{"value":"13.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/valid-id","seller":{"username":"seller-id"}}]}`))
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
		t.Fatalf("expected only listing id without embedded whitespace to survive, got %+v", items)
	}
	if items[0].ListingID != "v1|valid-id|0" {
		t.Fatalf("expected trim-only listing id to survive whitespace guard, got %+v", items[0])
	}
}

func TestProviderSearchSkipsBrowseTextFieldsWithEncodedControlCharacters(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|encoded%0Aid|0","title":"Encoded ID Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/encoded-id","seller":{"username":"seller-text"}},{"itemId":"v1|encoded-title|0","title":"Encoded%7F Title Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/encoded-title","seller":{"username":"seller-text"}},{"itemId":"v1|encoded-seller|0","title":"Encoded Seller Slot Car","price":{"value":"13.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/encoded-seller","seller":{"username":"seller%00text"}},{"itemId":"v1|valid-text|0","title":"Valid Text Slot Car","price":{"value":"14.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/valid-text","seller":{"username":"seller-text"}}]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{
		BaseURL:     srv.URL,
		BearerToken: "token",
		Marketplace: "EBAY_AU",
	})
	items, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"afx"}})
	if err != nil {
		t.Fatalf("search ebay: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected encoded-control listing id/title to be skipped while seller fallback survives, got %+v", items)
	}
	if items[0].ListingID != "v1|encoded-seller|0" || items[0].Seller != "ebay" {
		t.Fatalf("expected encoded-control seller to fall back to ebay, got %+v", items[0])
	}
	if items[1].ListingID != "v1|valid-text|0" || items[1].Seller != "seller-text" {
		t.Fatalf("expected safe text candidate to survive, got %+v", items[1])
	}
}

func TestProviderSearchSkipsBrowseTextFieldsWithUnicodeFormatCharacters(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|format` + string(rune(0x202e)) + `id|0","title":"Format ID Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/format-id","seller":{"username":"seller-text"}},{"itemId":"v1|format-title|0","title":"Format` + string(rune(0x2066)) + ` Title Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/format-title","seller":{"username":"seller-text"}},{"itemId":"v1|format-seller|0","title":"Format Seller Slot Car","price":{"value":"13.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/format-seller","seller":{"username":"seller` + string(rune(0x202e)) + `text"}},{"itemId":"v1|safe-unicode|0","title":"Pokémon Märklin Slot Car","price":{"value":"14.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/safe-unicode","seller":{"username":"seller-text"}}]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{
		BaseURL:     srv.URL,
		BearerToken: "token",
		Marketplace: "EBAY_AU",
	})
	items, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"afx"}})
	if err != nil {
		t.Fatalf("search ebay: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected Unicode-format listing id/title to be skipped while seller fallback survives, got %+v", items)
	}
	if items[0].ListingID != "v1|format-seller|0" || items[0].Seller != "ebay" {
		t.Fatalf("expected Unicode-format seller to fall back to ebay, got %+v", items[0])
	}
	if items[1].ListingID != "v1|safe-unicode|0" || items[1].Title != "Pokémon Märklin Slot Car" || items[1].Seller != "seller-text" {
		t.Fatalf("expected safe Unicode text candidate to survive, got %+v", items[1])
	}
}

func TestProviderSearchSkipsBrowseTextFieldsWithEncodedUnicodeText(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|encoded%E2%80%AEid|0","title":"Encoded ID Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/encoded-id","seller":{"username":"seller-text"}},{"itemId":"v1|encoded-title|0","title":"Encoded%E2%80%AFTitle Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/encoded-title","seller":{"username":"seller-text"}},{"itemId":"v1|encoded-seller|0","title":"Encoded Seller Slot Car","price":{"value":"13.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/encoded-seller","seller":{"username":"seller%E2%81%A6text"}},{"itemId":"v1|safe-encoded-text|0","title":"Safe Encoded Text Slot Car","price":{"value":"14.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/safe-encoded-text","seller":{"username":"seller-text"}}]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{
		BaseURL:     srv.URL,
		BearerToken: "token",
		Marketplace: "EBAY_AU",
	})
	items, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"afx"}})
	if err != nil {
		t.Fatalf("search ebay: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected encoded-Unicode listing id/title to be skipped while seller fallback survives, got %+v", items)
	}
	if items[0].ListingID != "v1|encoded-seller|0" || items[0].Seller != "ebay" {
		t.Fatalf("expected encoded-Unicode seller to fall back to ebay, got %+v", items[0])
	}
	if items[1].ListingID != "v1|safe-encoded-text|0" || items[1].Seller != "seller-text" {
		t.Fatalf("expected safe text candidate to survive, got %+v", items[1])
	}
}

func TestProviderSearchSkipsBrowseTextFieldsWithEncodedWhitespaceText(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|encoded%20id|0","title":"Encoded ID Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/encoded-id","seller":{"username":"seller-text"}},{"itemId":"v1|encoded-title|0","title":"Encoded%09Title Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/encoded-title","seller":{"username":"seller-text"}},{"itemId":"v1|encoded-seller|0","title":"Encoded Seller Slot Car","price":{"value":"13.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/encoded-seller","seller":{"username":"seller%20text"}},{"itemId":"v1|safe-text|0","title":"Safe Text Slot Car","price":{"value":"14.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/safe-text","seller":{"username":"seller-text"}}]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{
		BaseURL:     srv.URL,
		BearerToken: "token",
		Marketplace: "EBAY_AU",
	})
	items, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"afx"}})
	if err != nil {
		t.Fatalf("search ebay: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected encoded-whitespace listing id/title to be skipped while seller fallback survives, got %+v", items)
	}
	if items[0].ListingID != "v1|encoded-seller|0" || items[0].Seller != "ebay" {
		t.Fatalf("expected encoded-whitespace seller to fall back to ebay, got %+v", items[0])
	}
	if items[1].ListingID != "v1|safe-text|0" || items[1].Seller != "seller-text" {
		t.Fatalf("expected safe text candidate to survive, got %+v", items[1])
	}
}

func TestProviderSearchSkipsBrowseTextFieldsWithNestedEncodedUnsafeText(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|nested%250Aid|0","title":"Nested ID Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/nested-id","seller":{"username":"seller-text"}},{"itemId":"v1|nested-title|0","title":"Nested%25E2%2580%25AFTitle Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/nested-title","seller":{"username":"seller-text"}},{"itemId":"v1|nested-seller|0","title":"Nested Seller Slot Car","price":{"value":"13.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/nested-seller","seller":{"username":"seller%2509text"}},{"itemId":"v1|safe-nested-text|0","title":"Safe Nested Text Slot Car","price":{"value":"14.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/safe-nested-text","seller":{"username":"seller-text"}}]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{
		BaseURL:     srv.URL,
		BearerToken: "token",
		Marketplace: "EBAY_AU",
	})
	items, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"afx"}})
	if err != nil {
		t.Fatalf("search ebay: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected nested-encoded listing id/title to be skipped while seller fallback survives, got %+v", items)
	}
	if items[0].ListingID != "v1|nested-seller|0" || items[0].Seller != "ebay" {
		t.Fatalf("expected nested-encoded seller to fall back to ebay, got %+v", items[0])
	}
	if items[1].ListingID != "v1|safe-nested-text|0" || items[1].Seller != "seller-text" {
		t.Fatalf("expected safe text candidate to survive, got %+v", items[1])
	}
}

func TestProviderSearchSkipsOversizedBrowseRequiredText(t *testing.T) {
	t.Parallel()

	oversizedListingID := strings.Repeat("listing-", 90)
	oversizedTitle := strings.Repeat("title ", 120)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"` + oversizedListingID + `","title":"Oversized ID Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/oversized-id","seller":{"username":"seller-text"}},{"itemId":"v1|oversized-title|0","title":"` + oversizedTitle + `","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/oversized-title","seller":{"username":"seller-text"}},{"itemId":"v1|valid-sized-text|0","title":"Valid Sized Slot Car","price":{"value":"13.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/valid-sized-text","seller":{"username":"seller-text"}}]}`))
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
	if len(items) != 1 || items[0].ListingID != "v1|valid-sized-text|0" {
		t.Fatalf("expected oversized required text candidates to be skipped, got %+v", items)
	}
}

func TestProviderSearchFallsBackOversizedBrowseSeller(t *testing.T) {
	t.Parallel()

	oversizedSeller := strings.Repeat("seller-", 90)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|oversized-seller|0","title":"Oversized Seller Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/oversized-seller","seller":{"username":"` + oversizedSeller + `"}},{"itemId":"v1|safe-seller|0","title":"Safe Seller Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/safe-seller","seller":{"username":"seller-safe"}}]}`))
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
	if len(items) != 2 {
		t.Fatalf("expected valid candidates to survive oversized seller fallback, got %+v", items)
	}
	if items[0].ListingID != "v1|oversized-seller|0" || items[0].Seller != "ebay" {
		t.Fatalf("expected oversized seller to fall back to ebay, got %+v", items[0])
	}
	if items[1].ListingID != "v1|safe-seller|0" || items[1].Seller != "seller-safe" {
		t.Fatalf("expected safe seller to survive, got %+v", items[1])
	}
}

func TestProviderSearchSkipsDuplicateBrowseListingIDs(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":" v1|duplicate|0 ","title":"First Duplicate Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/duplicate-first","seller":{"username":"seller-duplicate"}},{"itemId":"v1|duplicate|0","title":"Second Duplicate Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/duplicate-second","seller":{"username":"seller-duplicate"}},{"itemId":"v1|unique|0","title":"Unique Slot Car","price":{"value":"13.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/unique","seller":{"username":"seller-unique"}}]}`))
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
	if len(items) != 2 {
		t.Fatalf("expected duplicate listing id to be emitted once, got %+v", items)
	}
	if items[0].ListingID != "v1|duplicate|0" || items[0].Title != "First Duplicate Slot Car" || items[0].URL != "https://ebay/item/duplicate-first" {
		t.Fatalf("expected first normalized duplicate listing to survive, got %+v", items[0])
	}
	if items[1].ListingID != "v1|unique|0" {
		t.Fatalf("expected unique listing to survive after duplicate, got %+v", items[1])
	}
}

func TestProviderSearchSkipsCaseVariantDuplicateBrowseListingIDs(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":" v1|Case-Duplicate|0 ","title":"First Case Duplicate Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/case-duplicate-first","seller":{"username":"seller-duplicate"}},{"itemId":"V1|case-duplicate|0","title":"Second Case Duplicate Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/case-duplicate-second","seller":{"username":"seller-duplicate"}},{"itemId":"v1|unique-case|0","title":"Unique Case Slot Car","price":{"value":"13.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/unique-case","seller":{"username":"seller-unique"}}]}`))
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
	if len(items) != 2 {
		t.Fatalf("expected case-variant duplicate listing id to be emitted once, got %+v", items)
	}
	if items[0].ListingID != "v1|Case-Duplicate|0" || items[0].Title != "First Case Duplicate Slot Car" || items[0].URL != "https://ebay/item/case-duplicate-first" {
		t.Fatalf("expected first normalized case-variant duplicate listing to survive, got %+v", items[0])
	}
	if items[1].ListingID != "v1|unique-case|0" {
		t.Fatalf("expected unique listing to survive after case-variant duplicate, got %+v", items[1])
	}
}

func TestProviderSearchUsesFirstValidDuplicateBrowseListingID(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":" v1|duplicate-fallthrough|0 ","title":"Invalid Duplicate Price Slot Car","price":{"value":"not-a-price","currency":"AUD"},"itemWebUrl":"https://ebay/item/duplicate-invalid","seller":{"username":"seller-duplicate"}},{"itemId":"v1|duplicate-fallthrough|0","title":"Valid Duplicate Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/duplicate-valid","seller":{"username":"seller-duplicate"}},{"itemId":" v1|duplicate-fallthrough|0 ","title":"Later Duplicate Slot Car","price":{"value":"13.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/duplicate-later","seller":{"username":"seller-duplicate"}}]}`))
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
		t.Fatalf("expected first valid duplicate listing id to be emitted once, got %+v", items)
	}
	if items[0].ListingID != "v1|duplicate-fallthrough|0" || items[0].Title != "Valid Duplicate Slot Car" || items[0].URL != "https://ebay/item/duplicate-valid" {
		t.Fatalf("expected invalid duplicate to fall through to first valid duplicate candidate, got %+v", items[0])
	}
}

func TestProviderSearchSkipsNonWebBrowseItemURLs(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|relative-url|0","title":"Relative URL Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"/itm/relative-url","seller":{"username":"seller-url"}},{"itemId":"v1|javascript-url|0","title":"Script URL Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"javascript:alert(1)","seller":{"username":"seller-url"}},{"itemId":"v1|valid-url|0","title":"Valid URL Slot Car","price":{"value":"13.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/valid-url","seller":{"username":"seller-url"}}]}`))
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
		t.Fatalf("expected only valid web URL item to survive, got %+v", items)
	}
	if items[0].ListingID != "v1|valid-url|0" || items[0].URL != "https://www.ebay.com/itm/valid-url" {
		t.Fatalf("expected valid-url candidate to survive normalization, got %+v", items[0])
	}
}

func TestProviderSearchSkipsBrowseItemURLsWithUserinfo(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|userinfo-url|0","title":"Userinfo URL Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://token@www.ebay.com/itm/userinfo-url","seller":{"username":"seller-url"}},{"itemId":"v1|valid-url|0","title":"Valid URL Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/valid-url","seller":{"username":"seller-url"}}]}`))
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
		t.Fatalf("expected only URL without userinfo to survive, got %+v", items)
	}
	if items[0].ListingID != "v1|valid-url|0" || items[0].URL != "https://www.ebay.com/itm/valid-url" {
		t.Fatalf("expected valid-url candidate to survive userinfo guard, got %+v", items[0])
	}
}

func TestProviderSearchSkipsBrowseItemURLsWithEncodedControlCharacters(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|encoded-newline-url|0","title":"Encoded Newline URL Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/encoded%0Anewline","seller":{"username":"seller-url"}},{"itemId":"v1|encoded-nul-url|0","title":"Encoded NUL URL Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/encoded%00nul","seller":{"username":"seller-url"}},{"itemId":"v1|valid-url|0","title":"Valid URL Slot Car","price":{"value":"13.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/valid-url","seller":{"username":"seller-url"}}]}`))
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
	if len(items) != 1 || items[0].ListingID != "v1|valid-url|0" {
		t.Fatalf("expected encoded-control item URLs to be skipped, got %+v", items)
	}
}

func TestProviderSearchSkipsBrowseItemURLsWithEncodedSpaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|encoded-space-url|0","title":"Encoded Space URL Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/encoded%20space","seller":{"username":"seller-url"}},{"itemId":"v1|valid-url|0","title":"Valid URL Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/valid-url","seller":{"username":"seller-url"}}]}`))
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
	if len(items) != 1 || items[0].ListingID != "v1|valid-url|0" {
		t.Fatalf("expected encoded-space item URL to be skipped, got %+v", items)
	}
}

func TestProviderSearchSkipsBrowseItemURLsWithMalformedEscapes(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|malformed-escape-url|0","title":"Malformed Escape URL Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/malformed%ZZescape","seller":{"username":"seller-url"}},{"itemId":"v1|valid-url|0","title":"Valid URL Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/valid-url","seller":{"username":"seller-url"}}]}`))
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
	if len(items) != 1 || items[0].ListingID != "v1|valid-url|0" {
		t.Fatalf("expected malformed-escape item URL to be skipped, got %+v", items)
	}
}

func TestProviderSearchSkipsBrowseItemURLsWithEncodedUnicodeURLText(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|encoded-narrow-space-url|0","title":"Encoded Unicode Space URL Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/encoded%E2%80%AFspace","seller":{"username":"seller-url"}},{"itemId":"v1|encoded-format-url|0","title":"Encoded Format URL Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/encoded%E2%80%AEformat","seller":{"username":"seller-url"}},{"itemId":"v1|valid-url|0","title":"Valid URL Slot Car","price":{"value":"13.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/valid-url","seller":{"username":"seller-url"}}]}`))
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
	if len(items) != 1 || items[0].ListingID != "v1|valid-url|0" {
		t.Fatalf("expected encoded Unicode URL text item URLs to be skipped, got %+v", items)
	}
}

func TestProviderSearchSkipsBrowseItemURLsWithRawUnicodeFormatCharacters(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|raw-format-url|0","title":"Raw Format URL Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/raw` + string(rune(0x202e)) + `format","seller":{"username":"seller-url"}},{"itemId":"v1|valid-url|0","title":"Valid URL Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/valid-url","seller":{"username":"seller-url"}}]}`))
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
	if len(items) != 1 || items[0].ListingID != "v1|valid-url|0" {
		t.Fatalf("expected raw Unicode format item URL to be skipped, got %+v", items)
	}
}

func TestProviderSearchSkipsBrowseItemURLsWithRawControlCharacters(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|raw-del-url|0","title":"Raw DEL URL Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/raw` + string(rune(0x7f)) + `del","seller":{"username":"seller-url"}},{"itemId":"v1|valid-url|0","title":"Valid URL Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/valid-url","seller":{"username":"seller-url"}}]}`))
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
	if len(items) != 1 || items[0].ListingID != "v1|valid-url|0" {
		t.Fatalf("expected raw-control item URL to be skipped, got %+v", items)
	}
}

func TestProviderSearchSkipsBrowseItemURLsWithRawWhitespace(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|space-url|0","title":"Raw Space URL Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/raw space","seller":{"username":"seller-url"}},{"itemId":"v1|valid-url|0","title":"Valid URL Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/valid-url","seller":{"username":"seller-url"}}]}`))
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
	if len(items) != 1 || items[0].ListingID != "v1|valid-url|0" {
		t.Fatalf("expected raw-whitespace item URL to be skipped, got %+v", items)
	}
}

func TestProviderSearchSkipsBrowseItemURLsWithUnicodeWhitespace(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|unicode-space-url|0","title":"Unicode Space URL Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/raw` + string(rune(0x00a0)) + `space","seller":{"username":"seller-url"}},{"itemId":"v1|valid-url|0","title":"Valid URL Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/valid-url","seller":{"username":"seller-url"}}]}`))
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
	if len(items) != 1 || items[0].ListingID != "v1|valid-url|0" {
		t.Fatalf("expected unicode-whitespace item URL to be skipped, got %+v", items)
	}
}

func TestProviderSearchSkipsOversizedBrowseItemURLs(t *testing.T) {
	t.Parallel()

	oversizedPath := strings.Repeat("oversized-url-segment/", 120)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|oversized-url|0","title":"Oversized URL Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/` + oversizedPath + `","seller":{"username":"seller-url"}},{"itemId":"v1|valid-url|0","title":"Valid URL Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/valid-url","seller":{"username":"seller-url"}}]}`))
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
	if len(items) != 1 || items[0].ListingID != "v1|valid-url|0" {
		t.Fatalf("expected oversized item URL to be skipped, got %+v", items)
	}
}

func TestProviderSearchDropsNonWebBrowseImageURLs(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|relative-image|0","title":"Relative Image Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/relative-image","image":{"imageUrl":"/image/relative.jpg"},"seller":{"username":"seller-image"}},{"itemId":"v1|script-image|0","title":"Script Image Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/script-image","image":{"imageUrl":"javascript:alert(1)"},"seller":{"username":"seller-image"}},{"itemId":"v1|valid-image|0","title":"Valid Image Slot Car","price":{"value":"13.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/valid-image","image":{"imageUrl":" https://i.ebayimg.com/images/valid.jpg "},"seller":{"username":"seller-image"}}]}`))
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
	if len(items) != 3 {
		t.Fatalf("expected all candidates to survive when only optional images are invalid, got %+v", items)
	}
	for _, item := range items {
		switch item.ListingID {
		case "v1|relative-image|0", "v1|script-image|0":
			if item.Image != "" {
				t.Fatalf("expected invalid optional image URL to be dropped for %+v", item)
			}
		case "v1|valid-image|0":
			if item.Image != "https://i.ebayimg.com/images/valid.jpg" {
				t.Fatalf("expected valid image URL to be trimmed and preserved, got %+v", item)
			}
		default:
			t.Fatalf("unexpected candidate %+v", item)
		}
	}
}

func TestProviderSearchDropsBrowseImageURLsWithUserinfo(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|userinfo-image|0","title":"Userinfo Image Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/userinfo-image","image":{"imageUrl":"https://token@i.ebayimg.com/images/userinfo.jpg"},"seller":{"username":"seller-image"}},{"itemId":"v1|valid-image|0","title":"Valid Image Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/valid-image","image":{"imageUrl":"https://i.ebayimg.com/images/valid.jpg"},"seller":{"username":"seller-image"}}]}`))
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
	if len(items) != 2 {
		t.Fatalf("expected candidates to survive when only optional image userinfo is invalid, got %+v", items)
	}
	for _, item := range items {
		switch item.ListingID {
		case "v1|userinfo-image|0":
			if item.Image != "" {
				t.Fatalf("expected image URL with userinfo to be dropped, got %+v", item)
			}
		case "v1|valid-image|0":
			if item.Image != "https://i.ebayimg.com/images/valid.jpg" {
				t.Fatalf("expected valid image URL to be preserved, got %+v", item)
			}
		default:
			t.Fatalf("unexpected candidate %+v", item)
		}
	}
}

func TestProviderSearchDropsBrowseImageURLsWithEncodedControlCharacters(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|encoded-newline-image|0","title":"Encoded Newline Image Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/encoded-newline-image","image":{"imageUrl":"https://i.ebayimg.com/images/encoded%0Anewline.jpg"},"seller":{"username":"seller-image"}},{"itemId":"v1|encoded-nul-image|0","title":"Encoded NUL Image Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/encoded-nul-image","image":{"imageUrl":"https://i.ebayimg.com/images/encoded%00nul.jpg"},"seller":{"username":"seller-image"}},{"itemId":"v1|valid-image|0","title":"Valid Image Slot Car","price":{"value":"13.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/valid-image","image":{"imageUrl":"https://i.ebayimg.com/images/valid.jpg"},"seller":{"username":"seller-image"}}]}`))
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
	if len(items) != 3 {
		t.Fatalf("expected all candidates with valid item URLs to survive, got %+v", items)
	}
	for _, item := range items {
		switch item.ListingID {
		case "v1|encoded-newline-image|0", "v1|encoded-nul-image|0":
			if item.Image != "" {
				t.Fatalf("expected encoded-control image URL to be dropped, got %+v", item)
			}
		case "v1|valid-image|0":
			if item.Image != "https://i.ebayimg.com/images/valid.jpg" {
				t.Fatalf("expected valid image URL to be preserved, got %+v", item)
			}
		default:
			t.Fatalf("unexpected candidate %+v", item)
		}
	}
}

func TestProviderSearchDropsBrowseImageURLsWithEncodedSpaces(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|encoded-space-image|0","title":"Encoded Space Image Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/encoded-space-image","image":{"imageUrl":"https://i.ebayimg.com/images/encoded%20space.jpg"},"thumbnailImages":[{"imageUrl":"https://i.ebayimg.com/images/valid-thumb.jpg"}],"seller":{"username":"seller-image"}},{"itemId":"v1|valid-image|0","title":"Valid Image Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/valid-image","image":{"imageUrl":"https://i.ebayimg.com/images/valid.jpg"},"seller":{"username":"seller-image"}}]}`))
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
	if len(items) != 2 {
		t.Fatalf("expected candidates with valid item URLs to survive, got %+v", items)
	}
	for _, item := range items {
		switch item.ListingID {
		case "v1|encoded-space-image|0":
			if item.Image != "https://i.ebayimg.com/images/valid-thumb.jpg" {
				t.Fatalf("expected encoded-space primary image URL to fall back to safe thumbnail, got %+v", item)
			}
		case "v1|valid-image|0":
			if item.Image != "https://i.ebayimg.com/images/valid.jpg" {
				t.Fatalf("expected valid image URL to be preserved, got %+v", item)
			}
		default:
			t.Fatalf("unexpected candidate %+v", item)
		}
	}
}

func TestProviderSearchDropsBrowseImageURLsWithMalformedEscapes(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|malformed-image-url|0","title":"Malformed Image URL Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/malformed-image-url","image":{"imageUrl":"https://i.ebayimg.com/images/malformed%ZZimage.jpg"},"thumbnailImages":[{"imageUrl":"https://i.ebayimg.com/images/valid-thumb.jpg"}],"seller":{"username":"seller-image"}}]}`))
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
		t.Fatalf("expected otherwise valid candidate to survive, got %+v", items)
	}
	if items[0].Image != "https://i.ebayimg.com/images/valid-thumb.jpg" {
		t.Fatalf("expected malformed-escape primary image URL to fall back to safe thumbnail, got %+v", items[0])
	}
}

func TestProviderSearchDropsBrowseImageURLsWithEncodedUnicodeURLText(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|encoded-narrow-space-image|0","title":"Encoded Unicode Space Image Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/encoded-narrow-space-image","image":{"imageUrl":"https://i.ebayimg.com/images/encoded%E2%80%AFspace.jpg"},"seller":{"username":"seller-image"}},{"itemId":"v1|encoded-format-image|0","title":"Encoded Format Image Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/encoded-format-image","image":{"imageUrl":"https://i.ebayimg.com/images/encoded%E2%80%AEformat.jpg"},"thumbnailImages":[{"imageUrl":"https://i.ebayimg.com/images/valid-thumb.jpg"}],"seller":{"username":"seller-image"}},{"itemId":"v1|valid-image|0","title":"Valid Image Slot Car","price":{"value":"13.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/valid-image","image":{"imageUrl":"https://i.ebayimg.com/images/valid.jpg"},"seller":{"username":"seller-image"}}]}`))
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
	if len(items) != 3 {
		t.Fatalf("expected candidates with valid item URLs to survive, got %+v", items)
	}
	for _, item := range items {
		switch item.ListingID {
		case "v1|encoded-narrow-space-image|0":
			if item.Image != "" {
				t.Fatalf("expected encoded Unicode whitespace image URL to be dropped, got %+v", item)
			}
		case "v1|encoded-format-image|0":
			if item.Image != "https://i.ebayimg.com/images/valid-thumb.jpg" {
				t.Fatalf("expected encoded Unicode format image URL to fall back to safe thumbnail, got %+v", item)
			}
		case "v1|valid-image|0":
			if item.Image != "https://i.ebayimg.com/images/valid.jpg" {
				t.Fatalf("expected valid image URL to be preserved, got %+v", item)
			}
		default:
			t.Fatalf("unexpected candidate %+v", item)
		}
	}
}

func TestProviderSearchDropsBrowseImageURLsWithRawUnicodeFormatCharacters(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|raw-format-image|0","title":"Raw Format Image Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/raw-format-image","image":{"imageUrl":"https://i.ebayimg.com/images/raw` + string(rune(0x202e)) + `format.jpg"},"thumbnailImages":[{"imageUrl":"https://i.ebayimg.com/images/valid-thumb.jpg"}],"seller":{"username":"seller-image"}},{"itemId":"v1|valid-image|0","title":"Valid Image Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/valid-image","image":{"imageUrl":"https://i.ebayimg.com/images/valid.jpg"},"seller":{"username":"seller-image"}}]}`))
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
	if len(items) != 2 {
		t.Fatalf("expected candidates with valid item URLs to survive, got %+v", items)
	}
	for _, item := range items {
		switch item.ListingID {
		case "v1|raw-format-image|0":
			if item.Image != "https://i.ebayimg.com/images/valid-thumb.jpg" {
				t.Fatalf("expected raw Unicode format image URL to fall back to safe thumbnail, got %+v", item)
			}
		case "v1|valid-image|0":
			if item.Image != "https://i.ebayimg.com/images/valid.jpg" {
				t.Fatalf("expected valid image URL to be preserved, got %+v", item)
			}
		default:
			t.Fatalf("unexpected candidate %+v", item)
		}
	}
}

func TestProviderSearchDropsBrowseImageURLsWithRawControlCharacters(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|raw-del-image|0","title":"Raw DEL Image Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/raw-del-image","image":{"imageUrl":"https://i.ebayimg.com/images/raw` + string(rune(0x7f)) + `del.jpg"},"seller":{"username":"seller-image"}},{"itemId":"v1|valid-image|0","title":"Valid Image Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/valid-image","image":{"imageUrl":"https://i.ebayimg.com/images/valid.jpg"},"seller":{"username":"seller-image"}}]}`))
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
	if len(items) != 2 {
		t.Fatalf("expected candidates with valid item URLs to survive, got %+v", items)
	}
	for _, item := range items {
		switch item.ListingID {
		case "v1|raw-del-image|0":
			if item.Image != "" {
				t.Fatalf("expected raw-control image URL to be dropped, got %+v", item)
			}
		case "v1|valid-image|0":
			if item.Image != "https://i.ebayimg.com/images/valid.jpg" {
				t.Fatalf("expected valid image URL to be preserved, got %+v", item)
			}
		default:
			t.Fatalf("unexpected candidate %+v", item)
		}
	}
}

func TestProviderSearchDropsBrowseImageURLsWithRawWhitespace(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|space-image|0","title":"Raw Space Image Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/space-image","image":{"imageUrl":"https://i.ebayimg.com/images/raw space.jpg"},"seller":{"username":"seller-image"}},{"itemId":"v1|valid-image|0","title":"Valid Image Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/valid-image","image":{"imageUrl":"https://i.ebayimg.com/images/valid.jpg"},"seller":{"username":"seller-image"}}]}`))
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
	if len(items) != 2 {
		t.Fatalf("expected candidates with valid item URLs to survive, got %+v", items)
	}
	for _, item := range items {
		switch item.ListingID {
		case "v1|space-image|0":
			if item.Image != "" {
				t.Fatalf("expected raw-whitespace image URL to be dropped, got %+v", item)
			}
		case "v1|valid-image|0":
			if item.Image != "https://i.ebayimg.com/images/valid.jpg" {
				t.Fatalf("expected valid image URL to be preserved, got %+v", item)
			}
		default:
			t.Fatalf("unexpected candidate %+v", item)
		}
	}
}

func TestProviderSearchDropsBrowseImageURLsWithUnicodeWhitespace(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|unicode-space-image|0","title":"Unicode Space Image Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/unicode-space-image","image":{"imageUrl":"https://i.ebayimg.com/images/raw` + string(rune(0x00a0)) + `space.jpg"},"seller":{"username":"seller-image"}},{"itemId":"v1|valid-image|0","title":"Valid Image Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/valid-image","image":{"imageUrl":"https://i.ebayimg.com/images/valid.jpg"},"seller":{"username":"seller-image"}}]}`))
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
	if len(items) != 2 {
		t.Fatalf("expected candidates with valid item URLs to survive, got %+v", items)
	}
	for _, item := range items {
		switch item.ListingID {
		case "v1|unicode-space-image|0":
			if item.Image != "" {
				t.Fatalf("expected unicode-whitespace image URL to be dropped, got %+v", item)
			}
		case "v1|valid-image|0":
			if item.Image != "https://i.ebayimg.com/images/valid.jpg" {
				t.Fatalf("expected valid image URL to be preserved, got %+v", item)
			}
		default:
			t.Fatalf("unexpected candidate %+v", item)
		}
	}
}

func TestProviderSearchDropsOversizedBrowseImageURLs(t *testing.T) {
	t.Parallel()

	oversizedPath := strings.Repeat("oversized-image-segment/", 120)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|oversized-image|0","title":"Oversized Image Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/oversized-image","image":{"imageUrl":"https://i.ebayimg.com/images/` + oversizedPath + `.jpg"},"seller":{"username":"seller-image"}},{"itemId":"v1|valid-image|0","title":"Valid Image Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/valid-image","image":{"imageUrl":"https://i.ebayimg.com/images/valid.jpg"},"seller":{"username":"seller-image"}}]}`))
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
	if len(items) != 2 {
		t.Fatalf("expected candidates with valid item URLs to survive, got %+v", items)
	}
	for _, item := range items {
		switch item.ListingID {
		case "v1|oversized-image|0":
			if item.Image != "" {
				t.Fatalf("expected oversized image URL to be dropped, got %+v", item)
			}
		case "v1|valid-image|0":
			if item.Image != "https://i.ebayimg.com/images/valid.jpg" {
				t.Fatalf("expected valid image URL to be preserved, got %+v", item)
			}
		default:
			t.Fatalf("unexpected candidate %+v", item)
		}
	}
}

func TestProviderSearchUsesFirstSafeAlternateBrowseImageURL(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|thumb-image|0","title":"Thumbnail Image Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/thumb-image","image":{"imageUrl":"javascript:alert(1)"},"thumbnailImages":[{"imageUrl":"/relative-thumb.jpg"},{"imageUrl":" https://i.ebayimg.com/images/thumb.jpg "}],"additionalImages":[{"imageUrl":"https://i.ebayimg.com/images/additional.jpg"}],"seller":{"username":"seller-image"}},{"itemId":"v1|additional-image|0","title":"Additional Image Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/additional-image","thumbnailImages":[{"imageUrl":"https://token@i.ebayimg.com/images/unsafe.jpg"}],"additionalImages":[{"imageUrl":"https://i.ebayimg.com/images/additional-valid.jpg"}],"seller":{"username":"seller-image"}},{"itemId":"v1|primary-image|0","title":"Primary Image Slot Car","price":{"value":"13.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/primary-image","image":{"imageUrl":"https://i.ebayimg.com/images/primary.jpg"},"thumbnailImages":[{"imageUrl":"https://i.ebayimg.com/images/thumb-ignored.jpg"}],"seller":{"username":"seller-image"}}]}`))
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
	if len(items) != 3 {
		t.Fatalf("expected all candidates to survive alternate image normalization, got %+v", items)
	}
	for _, item := range items {
		switch item.ListingID {
		case "v1|thumb-image|0":
			if item.Image != "https://i.ebayimg.com/images/thumb.jpg" {
				t.Fatalf("expected first safe thumbnail image URL, got %+v", item)
			}
		case "v1|additional-image|0":
			if item.Image != "https://i.ebayimg.com/images/additional-valid.jpg" {
				t.Fatalf("expected first safe additional image URL, got %+v", item)
			}
		case "v1|primary-image|0":
			if item.Image != "https://i.ebayimg.com/images/primary.jpg" {
				t.Fatalf("expected primary image URL to win, got %+v", item)
			}
		default:
			t.Fatalf("unexpected candidate %+v", item)
		}
	}
}

func TestProviderSearchFallsBackBlankBrowseSeller(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|blank-seller|0","title":"Blank Seller Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/blank-seller","seller":{"username":"   "}},{"itemId":"v1|named-seller|0","title":"Named Seller Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/named-seller","seller":{"username":" seller-name "}}]}`))
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
	if len(items) != 2 {
		t.Fatalf("expected both otherwise valid candidates to survive seller fallback, got %+v", items)
	}
	for _, item := range items {
		switch item.ListingID {
		case "v1|blank-seller|0":
			if item.Seller != "ebay" {
				t.Fatalf("expected blank seller to fall back to ebay, got %+v", item)
			}
		case "v1|named-seller|0":
			if item.Seller != "seller-name" {
				t.Fatalf("expected seller username to be trimmed, got %+v", item)
			}
		default:
			t.Fatalf("unexpected candidate %+v", item)
		}
	}
}

func TestProviderSearchFallsBackRawControlBrowseSeller(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|raw-seller|0","title":"Raw Seller Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/raw-seller","seller":{"username":"seller` + string(rune(0x7f)) + `name"}},{"itemId":"v1|named-seller|0","title":"Named Seller Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/named-seller","seller":{"username":" seller-name "}}]}`))
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
	if len(items) != 2 {
		t.Fatalf("expected otherwise valid candidates to survive seller control-byte fallback, got %+v", items)
	}
	for _, item := range items {
		switch item.ListingID {
		case "v1|raw-seller|0":
			if item.Seller != "ebay" {
				t.Fatalf("expected raw-control seller to fall back to ebay, got %+v", item)
			}
		case "v1|named-seller|0":
			if item.Seller != "seller-name" {
				t.Fatalf("expected safe seller username to be trimmed, got %+v", item)
			}
		default:
			t.Fatalf("unexpected candidate %+v", item)
		}
	}
}

func TestProviderSearchFallsBackWhitespaceBrowseSeller(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|ascii-space-seller|0","title":"ASCII Space Seller Slot Car","price":{"value":"11.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/ascii-space-seller","seller":{"username":"seller name"}},{"itemId":"v1|unicode-space-seller|0","title":"Unicode Space Seller Slot Car","price":{"value":"12.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/unicode-space-seller","seller":{"username":"seller` + string(rune(0x00a0)) + `name"}},{"itemId":"v1|named-seller|0","title":"Named Seller Slot Car","price":{"value":"13.00","currency":"AUD"},"itemWebUrl":"https://www.ebay.com/itm/named-seller","seller":{"username":" seller-name "}}]}`))
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
	if len(items) != 3 {
		t.Fatalf("expected otherwise valid candidates to survive seller whitespace fallback, got %+v", items)
	}
	for _, item := range items {
		switch item.ListingID {
		case "v1|ascii-space-seller|0", "v1|unicode-space-seller|0":
			if item.Seller != "ebay" {
				t.Fatalf("expected whitespace seller to fall back to ebay, got %+v", item)
			}
		case "v1|named-seller|0":
			if item.Seller != "seller-name" {
				t.Fatalf("expected safe seller username to be trimmed, got %+v", item)
			}
		default:
			t.Fatalf("unexpected candidate %+v", item)
		}
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

func TestProviderSearchParsesCommaGroupedShippingCost(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|comma-shipping|0","title":"Comma Shipping Slot Car","price":{"value":"122.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/comma-shipping","seller":{"username":"seller3"},"shippingOptions":[{"shippingCost":{"value":" 1,008.75 ","currency":"AUD"}}]}]}`))
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
	if items[0].Shipping != 1008.75 {
		t.Fatalf("expected comma-grouped shipping cost 1008.75, got %+v", items[0])
	}
}

func TestProviderSearchUsesFirstParseableShippingCost(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|790|0","title":"Fallback Shipping Slot Car","price":{"value":"22.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/790","seller":{"username":"seller3"},"shippingOptions":[{"shippingCost":{"value":"","currency":"AUD"}},{"shippingCost":{"value":"not-a-price","currency":"AUD"}},{"shippingCost":{"value":"12,34.25","currency":"AUD"}},{"shippingCost":{"value":"11.25","currency":"AUD"}}]}]}`))
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

func TestProviderSearchSkipsDecimalCommaShippingCost(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|decimal-comma-shipping|0","title":"Decimal Comma Shipping Slot Car","price":{"value":"22.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/decimal-comma-shipping","seller":{"username":"seller3"},"shippingOptions":[{"shippingCost":{"value":"12.3,4","currency":"AUD"}},{"shippingCost":{"value":"13.25","currency":"AUD"}}]},{"itemId":"v1|decimal-comma-shipping-zero|0","title":"Decimal Comma Shipping Zero Slot Car","price":{"value":"23.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/decimal-comma-shipping-zero","seller":{"username":"seller3"},"shippingOptions":[{"shippingCost":{"value":"4.2,5","currency":"AUD"}}]}]}`))
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
	if len(items) != 2 {
		t.Fatalf("expected both candidates to survive decimal-comma shipping normalization, got %+v", items)
	}
	for _, item := range items {
		switch item.ListingID {
		case "v1|decimal-comma-shipping|0":
			if item.Shipping != 13.25 {
				t.Fatalf("expected decimal-comma shipping option to fall through to 13.25, got %+v", item)
			}
		case "v1|decimal-comma-shipping-zero|0":
			if item.Shipping != 0 {
				t.Fatalf("expected decimal-comma shipping-only option to fall back to zero, got %+v", item)
			}
		default:
			t.Fatalf("unexpected candidate %+v", item)
		}
	}
}

func TestProviderSearchSkipsNonFiniteShippingCost(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|finite-shipping|0","title":"Finite Shipping Slot Car","price":{"value":"22.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/finite-shipping","seller":{"username":"seller3"},"shippingOptions":[{"shippingCost":{"value":"NaN","currency":"AUD"}},{"shippingCost":{"value":"Infinity","currency":"AUD"}},{"shippingCost":{"value":"13.25","currency":"AUD"}}]}]}`))
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
	if items[0].Shipping != 13.25 {
		t.Fatalf("expected non-finite shipping options to fall through to 13.25, got %+v", items[0])
	}
}

func TestProviderSearchSkipsNonPlainDecimalShippingCost(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|plain-shipping|0","title":"Plain Shipping Slot Car","price":{"value":"22.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/plain-shipping","seller":{"username":"seller3"},"shippingOptions":[{"shippingCost":{"value":"1e2","currency":"AUD"}},{"shippingCost":{"value":"+8.75","currency":"AUD"}},{"shippingCost":{"value":"9.25","currency":"AUD"}}]}]}`))
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
	if items[0].Shipping != 9.25 {
		t.Fatalf("expected non-plain decimal shipping options to fall through to 9.25, got %+v", items[0])
	}
}

func TestProviderSearchSkipsPartialDecimalShippingCost(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|strict-shipping|0","title":"Strict Shipping Slot Car","price":{"value":"22.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/strict-shipping","seller":{"username":"seller3"},"shippingOptions":[{"shippingCost":{"value":".75","currency":"AUD"}},{"shippingCost":{"value":"8.","currency":"AUD"}},{"shippingCost":{"value":"9.25","currency":"AUD"}}]}]}`))
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
	if items[0].Shipping != 9.25 {
		t.Fatalf("expected partial decimal shipping options to fall through to 9.25, got %+v", items[0])
	}
}

func TestProviderSearchSkipsOverPrecisionShippingCost(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|scale-shipping|0","title":"Scale Shipping Slot Car","price":{"value":"22.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/scale-shipping","seller":{"username":"seller3"},"shippingOptions":[{"shippingCost":{"value":"8.755","currency":"AUD"}},{"shippingCost":{"value":"9.25","currency":"AUD"}}]}]}`))
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
	if items[0].Shipping != 9.25 {
		t.Fatalf("expected over-precision shipping option to fall through to 9.25, got %+v", items[0])
	}
}

func TestProviderSearchSkipsNonPositiveShippingCost(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|791|0","title":"Positive Shipping Slot Car","price":{"value":"22.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/791","seller":{"username":"seller3"},"shippingOptions":[{"shippingCost":{"value":"0","currency":"AUD"}},{"shippingCost":{"value":"-1.25","currency":"AUD"}},{"shippingCost":{"value":"8.50","currency":"AUD"}}]}]}`))
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
	if items[0].Shipping != 8.50 {
		t.Fatalf("expected first positive shipping cost 8.50, got %+v", items[0])
	}
}

func TestProviderSearchSkipsBlankShippingCurrency(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|shipping-currency|0","title":"Shipping Currency Slot Car","price":{"value":"22.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/shipping-currency","seller":{"username":"seller3"},"shippingOptions":[{"shippingCost":{"value":"7.25","currency":"  "}},{"shippingCost":{"value":"9.50","currency":"AUD"}}]},{"itemId":"v1|shipping-currency-zero|0","title":"No Shipping Currency Slot Car","price":{"value":"23.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/shipping-currency-zero","seller":{"username":"seller3"},"shippingOptions":[{"shippingCost":{"value":"4.25","currency":""}}]}]}`))
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
	if len(items) != 2 {
		t.Fatalf("expected both candidates to survive shipping currency normalization, got %+v", items)
	}
	for _, item := range items {
		switch item.ListingID {
		case "v1|shipping-currency|0":
			if item.Shipping != 9.50 {
				t.Fatalf("expected blank-currency shipping option to fall through to 9.50, got %+v", item)
			}
		case "v1|shipping-currency-zero|0":
			if item.Shipping != 0 {
				t.Fatalf("expected blank-currency shipping-only option to fall back to zero, got %+v", item)
			}
		default:
			t.Fatalf("unexpected candidate %+v", item)
		}
	}
}

func TestProviderSearchSkipsMismatchedShippingCurrency(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|shipping-currency-match|0","title":"Shipping Currency Match Slot Car","price":{"value":"22.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/shipping-currency-match","seller":{"username":"seller3"},"shippingOptions":[{"shippingCost":{"value":"7.25","currency":"USD"}},{"shippingCost":{"value":"9.50","currency":" aud "}}]},{"itemId":"v1|shipping-currency-mismatch|0","title":"Shipping Currency Mismatch Slot Car","price":{"value":"23.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/shipping-currency-mismatch","seller":{"username":"seller3"},"shippingOptions":[{"shippingCost":{"value":"4.25","currency":"USD"}}]}]}`))
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
	if len(items) != 2 {
		t.Fatalf("expected both candidates to survive shipping currency-match normalization, got %+v", items)
	}
	for _, item := range items {
		switch item.ListingID {
		case "v1|shipping-currency-match|0":
			if item.Shipping != 9.50 {
				t.Fatalf("expected mismatched-currency shipping option to fall through to 9.50 AUD, got %+v", item)
			}
		case "v1|shipping-currency-mismatch|0":
			if item.Shipping != 0 {
				t.Fatalf("expected only mismatched-currency shipping options to fall back to zero, got %+v", item)
			}
		default:
			t.Fatalf("unexpected candidate %+v", item)
		}
	}
}

func TestProviderSearchUsesFirstMeaningfulAvailability(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|stock-fallback|0","title":"Availability Fallback Slot Car","price":{"value":"22.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/stock-fallback","seller":{"username":"seller-stock"},"estimatedAvailabilities":[{"estimatedAvailabilityStatus":" ","estimatedAvailableQuantity":0},{"estimatedAvailabilityStatus":"LIMITED_STOCK","estimatedAvailableQuantity":2}]}]}`))
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
	if items[0].StockState != "low_stock" || items[0].StockCount != 2 {
		t.Fatalf("expected first meaningful availability low_stock/2, got %+v", items[0])
	}
}

func TestProviderSearchIgnoresUnsafeAvailabilityStatusText(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|safe-stock|0","title":"Safe Stock Slot Car","price":{"value":"22.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/safe-stock","seller":{"username":"seller-stock"},"estimatedAvailabilities":[{"estimatedAvailabilityStatus":"IN_STOCK` + string(rune(0x7f)) + `","estimatedAvailableQuantity":7},{"estimatedAvailabilityStatus":"LIMITED_STOCK%0A","estimatedAvailableQuantity":4},{"estimatedAvailabilityStatus":"LIMITED_STOCK","estimatedAvailableQuantity":2}]}]}`))
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
	if items[0].StockState != "low_stock" || items[0].StockCount != 2 {
		t.Fatalf("expected unsafe availability statuses to be ignored before low_stock/2, got %+v", items[0])
	}
}

func TestProviderSearchIgnoresUnicodeFormatAvailabilityStatusText(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|unicode-stock|0","title":"Unicode Stock Slot Car","price":{"value":"22.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/unicode-stock","seller":{"username":"seller-stock"},"estimatedAvailabilities":[{"estimatedAvailabilityStatus":"IN_STOCK` + string(rune(0x202e)) + `","estimatedAvailableQuantity":7},{"estimatedAvailabilityStatus":"LIMITED_STOCK` + string(rune(0x2066)) + `","estimatedAvailableQuantity":4},{"estimatedAvailabilityStatus":"LIMITED_STOCK","estimatedAvailableQuantity":2}]}]}`))
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
		t.Fatalf("expected unicode-format availability status item to survive, got %+v", items)
	}
	if items[0].StockState != "low_stock" || items[0].StockCount != 2 {
		t.Fatalf("expected Unicode-format availability statuses to be ignored before low_stock/2, got %+v", items[0])
	}
}

func TestProviderSearchIgnoresEncodedUnsafeAvailabilityStatusText(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|encoded-stock|0","title":"Encoded Stock Slot Car","price":{"value":"22.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/encoded-stock","seller":{"username":"seller-stock"},"estimatedAvailabilities":[{"estimatedAvailabilityStatus":"IN%20STOCK","estimatedAvailableQuantity":7},{"estimatedAvailabilityStatus":"LIMITED%E2%80%AFSTOCK","estimatedAvailableQuantity":4},{"estimatedAvailabilityStatus":"LIMITED%E2%80%AESTOCK","estimatedAvailableQuantity":3},{"estimatedAvailabilityStatus":"LIMITED_STOCK","estimatedAvailableQuantity":2}]}]}`))
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
		t.Fatalf("expected encoded-unsafe availability status item to survive, got %+v", items)
	}
	if items[0].StockState != "low_stock" || items[0].StockCount != 2 {
		t.Fatalf("expected encoded-unsafe availability statuses to be ignored before low_stock/2, got %+v", items[0])
	}
}

func TestProviderSearchIgnoresNestedEncodedUnsafeAvailabilityStatusText(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|nested-encoded-stock|0","title":"Nested Encoded Stock Slot Car","price":{"value":"22.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/nested-encoded-stock","seller":{"username":"seller-stock"},"estimatedAvailabilities":[{"estimatedAvailabilityStatus":"IN%2520STOCK","estimatedAvailableQuantity":7},{"estimatedAvailabilityStatus":"LIMITED%25E2%2580%25AFSTOCK","estimatedAvailableQuantity":4},{"estimatedAvailabilityStatus":"LIMITED%25E2%2580%25AESTOCK","estimatedAvailableQuantity":3},{"estimatedAvailabilityStatus":"LIMITED_STOCK","estimatedAvailableQuantity":2}]}]}`))
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
		t.Fatalf("expected nested-encoded unsafe availability status item to survive, got %+v", items)
	}
	if items[0].StockState != "low_stock" || items[0].StockCount != 2 {
		t.Fatalf("expected nested-encoded unsafe availability statuses to be ignored before low_stock/2, got %+v", items[0])
	}
}

func TestProviderSearchIgnoresNegativeAvailabilityQuantity(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|negative-stock|0","title":"Negative Stock Slot Car","price":{"value":"22.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/negative-stock","seller":{"username":"seller-stock"},"estimatedAvailabilities":[{"estimatedAvailabilityStatus":"IN_STOCK","estimatedAvailableQuantity":-4}]}]}`))
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
	if items[0].StockState != "in_stock" || items[0].StockCount != -1 {
		t.Fatalf("expected negative availability quantity to be ignored as in_stock/-1, got %+v", items[0])
	}
}

func TestProviderSearchIgnoresOversizedAvailabilityQuantity(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|oversized-stock|0","title":"Oversized Stock Slot Car","price":{"value":"22.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/oversized-stock","seller":{"username":"seller-stock"},"estimatedAvailabilities":[{"estimatedAvailabilityStatus":"IN_STOCK","estimatedAvailableQuantity":100001}]}]}`))
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
	if items[0].StockState != "in_stock" || items[0].StockCount != -1 {
		t.Fatalf("expected oversized availability quantity to be ignored as in_stock/-1, got %+v", items[0])
	}
}

func TestProviderSearchDoesNotInferUnknownAvailabilityAsInStock(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|unknown-stock|0","title":"Unknown Stock Slot Car","price":{"value":"22.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/unknown-stock","seller":{"username":"seller-stock"},"estimatedAvailabilities":[{"estimatedAvailabilityStatus":"BACKORDERED","estimatedAvailableQuantity":6}]}]}`))
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
	if items[0].StockState != "unknown" || items[0].StockCount != -1 {
		t.Fatalf("expected unrecognized availability status to remain unknown/-1, got %+v", items[0])
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
	if providerErr.StatusCode != http.StatusForbidden || providerErr.ErrorCode != "PROVIDER_AUTH_INVALID" {
		t.Fatalf("unexpected provider error: %+v", providerErr)
	}
	for _, want := range []string{"1100", "ACCESS", "REQUEST", "Access token invalid", "Token scope is missing Browse access"} {
		if !strings.Contains(providerErr.Message, want) {
			t.Fatalf("expected provider auth error message to preserve %q, got %q", want, providerErr.Message)
		}
	}
}

func TestProviderSearchSanitizesStructuredAuthErrorPayload(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"errors":[{"errorId":1100,"domain":"ACCESS\nTOKEN","category":"REQUEST\tSCOPE","message":"Access token\ninvalid","longMessage":"unsafe%0Adetail"}]}`))
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
	for _, want := range []string{"1100", "ACCESS TOKEN", "REQUEST SCOPE", "Access token invalid"} {
		if !strings.Contains(providerErr.Message, want) {
			t.Fatalf("expected sanitized provider auth error message to preserve %q, got %q", want, providerErr.Message)
		}
	}
	for _, unwanted := range []string{"\n", "\t", "%0A", "unsafe"} {
		if strings.Contains(providerErr.Message, unwanted) {
			t.Fatalf("expected sanitized provider auth error message to omit %q, got %q", unwanted, providerErr.Message)
		}
	}
}

func TestProviderSearchBoundsStructuredAuthErrorPayloadFields(t *testing.T) {
	t.Parallel()

	oversized := strings.Repeat("scope-detail-", 80)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errors":[{"errorId":1100,"domain":"ACCESS","category":"REQUEST","message":"Access token invalid","longMessage":"` + oversized + `"}]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: "bad", Marketplace: "EBAY_AU"})
	_, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"pokemon"}})
	if err == nil {
		t.Fatal("expected structured auth error")
	}
	var providerErr *ProviderError
	if !errors.As(err, &providerErr) {
		t.Fatalf("expected ProviderError, got %T %v", err, err)
	}
	for _, want := range []string{"1100", "ACCESS", "REQUEST", "Access token invalid", "..."} {
		if !strings.Contains(providerErr.Message, want) {
			t.Fatalf("expected bounded provider auth error message to preserve %q, got %q", want, providerErr.Message)
		}
	}
	if strings.Contains(providerErr.Message, strings.Repeat("scope-detail-", 20)) {
		t.Fatalf("expected structured provider auth detail to be bounded, got %q", providerErr.Message)
	}
}

func TestProviderSearchPreservesPlainTextAuthErrorPayload(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("OAuth scope\n\nmissing\tfor   Browse API"))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: "expired-token", Marketplace: "EBAY_AU"})
	_, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"pokemon"}})
	if err == nil {
		t.Fatal("expected plain-text auth error")
	}
	var providerErr *ProviderError
	if !errors.As(err, &providerErr) {
		t.Fatalf("expected ProviderError, got %T %v", err, err)
	}
	if providerErr.StatusCode != http.StatusForbidden || providerErr.ErrorCode != "PROVIDER_AUTH_INVALID" {
		t.Fatalf("unexpected provider error: %+v", providerErr)
	}
	for _, want := range []string{"ebay credentials rejected with status 403", "OAuth scope missing for Browse API"} {
		if !strings.Contains(providerErr.Message, want) {
			t.Fatalf("expected provider auth error message to preserve %q, got %q", want, providerErr.Message)
		}
	}
	for _, unwanted := range []string{"\n", "\t", "  "} {
		if strings.Contains(providerErr.Message, unwanted) {
			t.Fatalf("expected provider auth error message to compact %q, got %q", unwanted, providerErr.Message)
		}
	}
}

func TestEbayErrorMessageCapsBodyReadBeforeDiagnostics(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusBadGateway,
		Body: &errorAfterReader{
			data: []byte("upstream Browse gateway timeout " + strings.Repeat("detail ", 900)),
		},
	}
	got := ebayErrorMessage(resp, "ebay search status: 502")

	if !strings.Contains(got, "ebay search status: 502") || !strings.Contains(got, "upstream Browse gateway timeout") {
		t.Fatalf("expected bounded body diagnostic to preserve status and prefix detail, got %q", got)
	}
}

type errorAfterReader struct {
	data   []byte
	offset int
}

func (r *errorAfterReader) Read(p []byte) (int, error) {
	if r.offset >= maxProviderErrorBodyRead {
		return 0, errors.New("unexpected unbounded follow-up read")
	}
	remaining := maxProviderErrorBodyRead - r.offset
	if remaining > len(r.data)-r.offset {
		remaining = len(r.data) - r.offset
	}
	if remaining > len(p) {
		remaining = len(p)
	}
	n := copy(p, r.data[r.offset:r.offset+remaining])
	r.offset += n
	if r.offset >= len(r.data) {
		return n, io.EOF
	}
	return n, nil
}

func (r *errorAfterReader) Close() error {
	return nil
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

func TestProviderSearchSanitizesStructuredBrowseErrorPayload(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"errors":[{"errorId":12001,"domain":"API_BROWSE\nSEARCH","category":"REQUEST\tLIMIT","message":"Rate limit\nexceeded","longMessage":"drop%7Fdetail"}]}`))
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
	for _, want := range []string{"12001", "API_BROWSE SEARCH", "REQUEST LIMIT", "Rate limit exceeded"} {
		if !strings.Contains(providerErr.Message, want) {
			t.Fatalf("expected sanitized provider error message to preserve %q, got %q", want, providerErr.Message)
		}
	}
	for _, unwanted := range []string{"\n", "\t", "%7F", "drop"} {
		if strings.Contains(providerErr.Message, unwanted) {
			t.Fatalf("expected sanitized provider error message to omit %q, got %q", unwanted, providerErr.Message)
		}
	}
}

func TestProviderSearchOmitsUnicodeFormatBrowseErrorPayloadFields(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"errors":[{"errorId":13000,"domain":"API_BROWSE","category":"APPLICATION` + string(rune(0x202e)) + `","message":"Browse gateway timeout","longMessage":"Retry later` + string(rune(0x2066)) + `"}]}`))
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
	for _, want := range []string{"13000", "API_BROWSE", "Browse gateway timeout"} {
		if !strings.Contains(providerErr.Message, want) {
			t.Fatalf("expected provider error message to preserve %q, got %q", want, providerErr.Message)
		}
	}
	for _, unwanted := range []string{"APPLICATION", "Retry later", string(rune(0x202e)), string(rune(0x2066))} {
		if strings.Contains(providerErr.Message, unwanted) {
			t.Fatalf("expected Unicode-format diagnostic field to be omitted for %q, got %q", unwanted, providerErr.Message)
		}
	}
}

func TestProviderSearchOmitsEncodedUnicodeBrowseErrorPayloadFields(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"errors":[{"errorId":13001,"domain":"API_BROWSE","category":"APPLICATION%E2%80%AE","message":"Browse gateway timeout","longMessage":"Retry%E2%80%AFlater"}]}`))
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
	for _, want := range []string{"13001", "API_BROWSE", "Browse gateway timeout"} {
		if !strings.Contains(providerErr.Message, want) {
			t.Fatalf("expected provider error message to preserve %q, got %q", want, providerErr.Message)
		}
	}
	for _, unwanted := range []string{"APPLICATION", "Retry", "%E2%80%AE", "%E2%80%AF"} {
		if strings.Contains(providerErr.Message, unwanted) {
			t.Fatalf("expected encoded-Unicode diagnostic field to be omitted for %q, got %q", unwanted, providerErr.Message)
		}
	}
}

func TestProviderSearchOmitsNestedEncodedBrowseErrorPayloadFields(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"errors":[{"errorId":13002,"domain":"API_BROWSE","category":"APPLICATION%250A","message":"Browse gateway timeout","longMessage":"Retry%25E2%2580%25AFlater"}]}`))
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
	for _, want := range []string{"13002", "API_BROWSE", "Browse gateway timeout"} {
		if !strings.Contains(providerErr.Message, want) {
			t.Fatalf("expected provider error message to preserve %q, got %q", want, providerErr.Message)
		}
	}
	for _, unwanted := range []string{"APPLICATION", "Retry", "%250A", "%25E2%2580%25AF"} {
		if strings.Contains(providerErr.Message, unwanted) {
			t.Fatalf("expected nested-encoded diagnostic field to be omitted for %q, got %q", unwanted, providerErr.Message)
		}
	}
}

func TestProviderSearchBoundsStructuredBrowseErrorPayloadFields(t *testing.T) {
	t.Parallel()

	oversized := strings.Repeat("browse-detail-", 80)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"errors":[{"errorId":13000,"domain":"API_BROWSE","category":"APPLICATION","message":"Browse gateway timeout","longMessage":"` + oversized + `"}]}`))
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
	for _, want := range []string{"13000", "API_BROWSE", "APPLICATION", "Browse gateway timeout", "..."} {
		if !strings.Contains(providerErr.Message, want) {
			t.Fatalf("expected bounded provider search error message to preserve %q, got %q", want, providerErr.Message)
		}
	}
	if strings.Contains(providerErr.Message, strings.Repeat("browse-detail-", 20)) {
		t.Fatalf("expected structured provider search detail to be bounded, got %q", providerErr.Message)
	}
}

func TestProviderSearchPreservesHTTPDateRetryAfter(t *testing.T) {
	t.Parallel()

	retryAt := time.Now().UTC().Add(2 * time.Minute)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", retryAt.Format(http.TimeFormat))
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"errors":[{"errorId":12001,"message":"Rate limit exceeded"}]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: "token", Marketplace: "EBAY_AU"})
	_, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"pokemon"}})
	if err == nil {
		t.Fatal("expected HTTP-date retry-after Browse error")
	}
	var providerErr *ProviderError
	if !errors.As(err, &providerErr) {
		t.Fatalf("expected ProviderError, got %T %v", err, err)
	}
	if providerErr.StatusCode != http.StatusTooManyRequests || providerErr.ErrorCode != "PROVIDER_SEARCH_FAILED" {
		t.Fatalf("unexpected provider error: %+v", providerErr)
	}
	if providerErr.RetryAfterSeconds < 30 || providerErr.RetryAfterSeconds > 180 {
		t.Fatalf("expected HTTP-date retry-after seconds near two minutes, got %+v", providerErr)
	}
}

func TestProviderSearchPreservesPlainTextBrowseErrorPayload(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("upstream Browse\n  gateway\t timeout"))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: "token", Marketplace: "EBAY_AU"})
	_, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"pokemon"}})
	if err == nil {
		t.Fatal("expected plain-text Browse error")
	}
	var providerErr *ProviderError
	if !errors.As(err, &providerErr) {
		t.Fatalf("expected ProviderError, got %T %v", err, err)
	}
	if providerErr.StatusCode != http.StatusBadGateway || providerErr.ErrorCode != "PROVIDER_SEARCH_FAILED" {
		t.Fatalf("unexpected provider error: %+v", providerErr)
	}
	for _, want := range []string{"ebay search status: 502", "upstream Browse gateway timeout"} {
		if !strings.Contains(providerErr.Message, want) {
			t.Fatalf("expected provider error message to preserve %q, got %q", want, providerErr.Message)
		}
	}
	for _, unwanted := range []string{"\n", "\t", "  "} {
		if strings.Contains(providerErr.Message, unwanted) {
			t.Fatalf("expected provider error message to compact %q, got %q", unwanted, providerErr.Message)
		}
	}
}

func TestProviderSearchClassifiesMalformedBrowsePayload(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: "token", Marketplace: "EBAY_AU"})
	_, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"pokemon"}})
	if err == nil {
		t.Fatal("expected malformed Browse payload error")
	}
	var providerErr *ProviderError
	if !errors.As(err, &providerErr) {
		t.Fatalf("expected ProviderError, got %T %v", err, err)
	}
	if providerErr.StatusCode != http.StatusBadGateway || providerErr.ErrorCode != "PROVIDER_SEARCH_FAILED" {
		t.Fatalf("unexpected provider error: %+v", providerErr)
	}
	for _, want := range []string{"decode ebay Browse response", "unexpected EOF"} {
		if !strings.Contains(providerErr.Message, want) {
			t.Fatalf("expected malformed payload message to preserve %q, got %q", want, providerErr.Message)
		}
	}
}

func TestProviderSearchRejectsBrowsePayloadWithTrailingData(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|trailing|0","title":"Trailing Slot Car","price":{"value":"22.00","currency":"AUD"},"itemWebUrl":"https://ebay/item/trailing","seller":{"username":"seller-trailing"}}]}{"itemSummaries":[]}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: "token", Marketplace: "EBAY_AU"})
	_, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"slot", "car"}})
	if err == nil {
		t.Fatal("expected trailing Browse payload error")
	}
	var providerErr *ProviderError
	if !errors.As(err, &providerErr) {
		t.Fatalf("expected ProviderError, got %T %v", err, err)
	}
	if providerErr.StatusCode != http.StatusBadGateway || providerErr.ErrorCode != "PROVIDER_SEARCH_FAILED" {
		t.Fatalf("unexpected provider error: %+v", providerErr)
	}
	if !strings.Contains(providerErr.Message, "trailing data after JSON object") {
		t.Fatalf("expected trailing data diagnostic, got %q", providerErr.Message)
	}
}

func TestProviderSearchRejectsOversizedSuccessfulBrowsePayload(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[],"warnings":"` + strings.Repeat("x", 2*1024*1024+1) + `"}`))
	}))
	defer srv.Close()

	p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: "token", Marketplace: "EBAY_AU"})
	_, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"slot", "car"}})
	if err == nil {
		t.Fatal("expected oversized Browse payload error")
	}
	var providerErr *ProviderError
	if !errors.As(err, &providerErr) {
		t.Fatalf("expected ProviderError, got %T %v", err, err)
	}
	if providerErr.StatusCode != http.StatusBadGateway || providerErr.ErrorCode != "PROVIDER_SEARCH_FAILED" {
		t.Fatalf("unexpected provider error: %+v", providerErr)
	}
	if !strings.Contains(providerErr.Message, "exceeded maximum response size") {
		t.Fatalf("expected oversized payload diagnostic, got %q", providerErr.Message)
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

func TestProviderSearchRejectsUnsafeBearerTokenBeforeBrowseRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		token string
	}{
		{name: "raw control byte", token: "valid-prefix\ninjected"},
		{name: "encoded control byte", token: "valid-prefix%0Ainjected"},
		{name: "unicode format control", token: "valid-prefix" + string(rune(0x202e)) + "injected"},
		{name: "encoded unicode format control", token: "valid-prefix%E2%80%AEinjected"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Fatalf("unsafe bearer token must be rejected before Browse request")
			}))
			defer srv.Close()

			p := NewProvider(ProviderConfig{BaseURL: srv.URL, BearerToken: tt.token, Marketplace: "EBAY_AU"})
			_, err := p.Search(context.Background(), scanner.QuerySet{Keywords: []string{"slot", "car"}})
			if err == nil {
				t.Fatal("expected unsafe bearer token error")
			}
			var providerErr *ProviderError
			if !errors.As(err, &providerErr) {
				t.Fatalf("expected ProviderError, got %T %v", err, err)
			}
			if providerErr.StatusCode != http.StatusUnauthorized || providerErr.ErrorCode != "PROVIDER_AUTH_INVALID" {
				t.Fatalf("unexpected provider error: %+v", providerErr)
			}
			if !strings.Contains(providerErr.Message, "invalid ebay bearer token format") {
				t.Fatalf("expected actionable token-format message, got %q", providerErr.Message)
			}
		})
	}
}

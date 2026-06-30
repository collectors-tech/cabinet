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
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|symbol-currency|0","title":"Symbol Currency Slot Car","price":{"value":"11.00","currency":"AU$"},"itemWebUrl":"https://ebay/item/symbol-currency","seller":{"username":"seller-currency"}},{"itemId":"v1|long-currency|0","title":"Long Currency Slot Car","price":{"value":"12.00","currency":"AUDD"},"itemWebUrl":"https://ebay/item/long-currency","seller":{"username":"seller-currency"}},{"itemId":"v1|digit-currency|0","title":"Digit Currency Slot Car","price":{"value":"13.00","currency":"AU1"},"itemWebUrl":"https://ebay/item/digit-currency","seller":{"username":"seller-currency"}},{"itemId":"v1|required-currency|0","title":"Required Currency Slot Car","price":{"value":"14.00","currency":" aud "},"itemWebUrl":"https://ebay/item/required-currency","seller":{"username":"seller-currency"}}]}`))
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

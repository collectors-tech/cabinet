package ebay

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/collectors-tech/cabinet/internal/scanner"
)

func TestProviderSearchNormalizesCandidates(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"itemSummaries":[{"itemId":"v1|123|0","title":"AFX P-1","price":{"value":"45.00"},"itemWebUrl":"https://ebay/item/123","image":{"imageUrl":"https://img/123.jpg"},"seller":{"username":"seller1"}}]}`))
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
}

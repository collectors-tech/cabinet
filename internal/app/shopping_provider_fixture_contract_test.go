package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/scanner"
)

func TestShoppingProviderFixturesNormalizeSharedCandidateShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		providerID string
		run        func(t *testing.T, serverURL string, client *http.Client) []scanner.CandidateInput
	}{
		{
			name:       "hobbytech boost product fixture",
			providerID: "hobbytechtoys",
			run: func(t *testing.T, serverURL string, client *http.Client) []scanner.CandidateInput {
				t.Helper()
				candidates, pageCount, err := runHobbytechSearch(
					context.Background(),
					client,
					scanner.QuerySet{Name: "AFX", Keywords: []string{"AFX"}},
					hobbytechBoostConfig{
						Shop:      "hobbytechtoys.com.au",
						SID:       "fixture-sid",
						Template:  "search",
						Widget:    "main",
						SearchURL: serverURL + "/hobbytech/search",
					},
					24,
				)
				if err != nil {
					t.Fatalf("runHobbytechSearch() error = %v", err)
				}
				if pageCount != 1 {
					t.Fatalf("expected one fixture page, got %d", pageCount)
				}
				return hobbytechCandidatesForScanner(candidates)
			},
		},
		{
			name:       "bigcommerce storefront product fixture",
			providerID: "voglers.com.au",
			run: func(t *testing.T, serverURL string, client *http.Client) []scanner.CandidateInput {
				t.Helper()
				candidates, err := runBigCommerceStorefrontSearch(
					context.Background(),
					client,
					serverURL+"/bigcommerce/products/search",
					"Scalextric",
					1,
					24,
					"voglers.com.au",
				)
				if err != nil {
					t.Fatalf("runBigCommerceStorefrontSearch() error = %v", err)
				}
				return bigCommerceCandidatesForScanner(candidates, "voglers.com.au")
			},
		},
		{
			name:       "doofinder product fixture",
			providerID: "mrtoys.com.au",
			run: func(t *testing.T, serverURL string, client *http.Client) []scanner.CandidateInput {
				t.Helper()
				candidates, total, err := runDoofinderSearch(
					context.Background(),
					client,
					serverURL+"/doofinder/search",
					"Carrera",
					1,
					24,
					"fixture-hashid",
					"https://www.mrtoys.com.au",
					"mrtoys.com.au",
				)
				if err != nil {
					t.Fatalf("runDoofinderSearch() error = %v", err)
				}
				if total != 1 {
					t.Fatalf("expected total=1 from fixture, got %d", total)
				}
				return doofinderCandidatesForScanner(candidates, "mrtoys.com.au")
			},
		},
	}

	server := shoppingFixtureServer(t)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			candidates := tt.run(t, server.URL, server.Client())
			if len(candidates) != 1 {
				t.Fatalf("expected one normalized candidate, got %+v", candidates)
			}
			assertSharedShoppingCandidateShape(t, tt.providerID, candidates[0])
		})
	}
}

func TestShoppingProviderFixturesRejectMissingRequiredFields(t *testing.T) {
	t.Parallel()

	server := shoppingFixtureServer(t)
	candidates, pageCount, err := runHobbytechSearch(
		context.Background(),
		server.Client(),
		scanner.QuerySet{Name: "Missing fields", Keywords: []string{"missing"}},
		hobbytechBoostConfig{
			Shop:      "hobbytechtoys.com.au",
			SID:       "fixture-sid",
			SearchURL: server.URL + "/missing-fields/search",
		},
		24,
	)
	if err != nil {
		t.Fatalf("runHobbytechSearch() missing-field fixture error = %v", err)
	}
	if pageCount != 1 {
		t.Fatalf("expected one missing-field fixture page, got %d", pageCount)
	}
	if normalized := hobbytechCandidatesForScanner(candidates); len(normalized) != 0 {
		t.Fatalf("expected missing title/url records to be rejected, got %+v", normalized)
	}
}

func shoppingFixtureServer(t *testing.T) *httptest.Server {
	t.Helper()

	fixtures := map[string]string{
		"/hobbytech/search":            readShoppingFixture(t, "hobbytech_success.json"),
		"/bigcommerce/products/search": readShoppingFixture(t, "bigcommerce_storefront_success.json"),
		"/doofinder/search":            readShoppingFixture(t, "doofinder_success.json"),
		"/missing-fields/search":       readShoppingFixture(t, "missing_required_fields.json"),
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := fixtures[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		if strings.Contains(r.URL.Path, "doofinder") && strings.TrimSpace(r.Header.Get("Origin")) == "" {
			http.Error(w, `{"error":"missing_origin"}`, http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server
}

func readShoppingFixture(t *testing.T, name string) string {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("testdata", "shopping_provider_fixtures", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return string(raw)
}

func assertSharedShoppingCandidateShape(t *testing.T, providerID string, candidate scanner.CandidateInput) {
	t.Helper()

	if strings.TrimSpace(candidate.ListingID) == "" {
		t.Fatalf("listing_id is required: %+v", candidate)
	}
	if strings.TrimSpace(candidate.Title) == "" {
		t.Fatalf("title is required: %+v", candidate)
	}
	if strings.TrimSpace(candidate.URL) == "" {
		t.Fatalf("source URL is required: %+v", candidate)
	}
	if strings.TrimSpace(candidate.Source) != providerID {
		t.Fatalf("expected provider source %q, got %+v", providerID, candidate)
	}
	if strings.TrimSpace(candidate.Seller) == "" {
		t.Fatalf("seller/provider domain is required: %+v", candidate)
	}
	if candidate.Price <= 0 {
		t.Fatalf("price must be normalized to a positive value: %+v", candidate)
	}
}

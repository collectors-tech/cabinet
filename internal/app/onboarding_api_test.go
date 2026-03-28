package app

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"testing"
)

func TestOnboardingSampleDataEndpointIsIdempotent(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"P1"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	setActive := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if setActive.Code != http.StatusOK {
		t.Fatalf("set active profile status=%d body=%s", setActive.Code, setActive.Body.String())
	}

	firstSeed := doRequest(t, a, http.MethodPost, "/api/onboarding/sample-data", nil, nil)
	if firstSeed.Code != http.StatusOK {
		t.Fatalf("first sample seed status=%d body=%s", firstSeed.Code, firstSeed.Body.String())
	}
	var firstPayload struct {
		CreatedItems          int  `json:"created_items"`
		CreatedWishlist       int  `json:"created_wishlist_entries"`
		TotalItems            int  `json:"total_items"`
		TotalWishlist         int  `json:"total_wishlist_entries"`
		AlreadySeededForProfile bool `json:"already_seeded_for_profile"`
	}
	if err := json.NewDecoder(firstSeed.Body).Decode(&firstPayload); err != nil {
		t.Fatalf("decode first seed payload: %v", err)
	}
	if firstPayload.CreatedItems == 0 {
		t.Fatalf("expected created_items > 0, got %+v", firstPayload)
	}
	if firstPayload.AlreadySeededForProfile {
		t.Fatalf("expected already_seeded_for_profile=false on first run, got %+v", firstPayload)
	}

	secondSeed := doRequest(t, a, http.MethodPost, "/api/onboarding/sample-data", nil, nil)
	if secondSeed.Code != http.StatusOK {
		t.Fatalf("second sample seed status=%d body=%s", secondSeed.Code, secondSeed.Body.String())
	}
	var secondPayload struct {
		CreatedItems            int  `json:"created_items"`
		CreatedWishlist         int  `json:"created_wishlist_entries"`
		TotalItems              int  `json:"total_items"`
		TotalWishlist           int  `json:"total_wishlist_entries"`
		AlreadySeededForProfile bool `json:"already_seeded_for_profile"`
	}
	if err := json.NewDecoder(secondSeed.Body).Decode(&secondPayload); err != nil {
		t.Fatalf("decode second seed payload: %v", err)
	}
	if secondPayload.CreatedItems != 0 {
		t.Fatalf("expected no new items on second run, got %+v", secondPayload)
	}
	if !secondPayload.AlreadySeededForProfile {
		t.Fatalf("expected already_seeded_for_profile=true on second run, got %+v", secondPayload)
	}
	if secondPayload.TotalItems != firstPayload.TotalItems {
		t.Fatalf("expected stable total_items across reruns, first=%d second=%d", firstPayload.TotalItems, secondPayload.TotalItems)
	}

	itemsResp := doRequest(t, a, http.MethodGet, "/api/items", nil, nil)
	if itemsResp.Code != http.StatusOK {
		t.Fatalf("list items status=%d body=%s", itemsResp.Code, itemsResp.Body.String())
	}
	var itemsPayload struct {
		Items []struct {
			Category string `json:"category"`
		} `json:"items"`
	}
	if err := json.NewDecoder(itemsResp.Body).Decode(&itemsPayload); err != nil {
		t.Fatalf("decode items payload: %v", err)
	}
	categories := make(map[string]struct{})
	for _, item := range itemsPayload.Items {
		if strings.TrimSpace(item.Category) != "" {
			categories[strings.TrimSpace(item.Category)] = struct{}{}
		}
	}
	if len(categories) < 6 {
		keys := make([]string, 0, len(categories))
		for category := range categories {
			keys = append(keys, category)
		}
		sort.Strings(keys)
		t.Fatalf("expected representative seeded category coverage (>=6 categories), got %d: %v", len(categories), keys)
	}
	for _, required := range []string{"Diecast", "Slot Car", "Trading Card", "Action Figure", "Comic", "Model Kit"} {
		if _, ok := categories[required]; !ok {
			keys := make([]string, 0, len(categories))
			for category := range categories {
				keys = append(keys, category)
			}
			sort.Strings(keys)
			t.Fatalf("expected seeded category %q, got %v", required, keys)
		}
	}
	if firstPayload.TotalItems < 6 {
		t.Fatalf("expected at least 6 seeded items for representative sample content, got %+v", firstPayload)
	}
	if firstPayload.TotalWishlist < 3 {
		t.Fatalf("expected seeded wishlist coverage, got %+v", firstPayload)
	}
}


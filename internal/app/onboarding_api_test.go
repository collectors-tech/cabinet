package app

import (
	"encoding/json"
	"net/http"
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
}


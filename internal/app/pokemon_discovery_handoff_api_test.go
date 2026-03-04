package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestPokemonDiscoveryHandoffRetainsMarketplaceMetadata(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Pokemon Discovery"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	setActive := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if setActive.Code != http.StatusOK {
		t.Fatalf("set active profile status=%d body=%s", setActive.Code, setActive.Body.String())
	}

	if _, err := a.db.Exec(`INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title) VALUES ('poke-item-1', ?, 'Pokemon','Cards','POKE-001','Pokemon Card')`, profile.ID); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO scanner_query_sets(id, profile_id, name, keywords_json, exclusions_json) VALUES ('poke-qs-1', ?, 'Pokemon Query', '["pokemon"]', '[]')`, profile.ID); err != nil {
		t.Fatalf("seed query set: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO scanner_candidates(id, profile_id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source, stock_state, stock_count) VALUES ('poke-cand-1', ?, 'poke-qs-1', 'poke-listing-1', 'Pokemon Candidate', 33.25, 0, 'https://example.test/poke-listing-1', '', 'poke-seller', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'low_stock', 2)`, profile.ID); err != nil {
		t.Fatalf("seed candidate: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at) VALUES ('poke-cand-1', 'poke-item-1', 'not_in_collection', 0.9, 0, 'POKE-001', CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("seed match: %v", err)
	}

	action := doRequest(t, a, http.MethodPost, "/api/discovery/action", strings.NewReader(`{"candidate_id":"poke-cand-1","type":"add_to_wishlist"}`), map[string]string{"Content-Type": "application/json"})
	if action.Code != http.StatusOK {
		t.Fatalf("discovery action status=%d body=%s", action.Code, action.Body.String())
	}

	wishlistResp := doRequest(t, a, http.MethodGet, "/api/wishlist", nil, nil)
	if wishlistResp.Code != http.StatusOK {
		t.Fatalf("wishlist status=%d body=%s", wishlistResp.Code, wishlistResp.Body.String())
	}
	var wishlistPayload struct {
		Items []struct {
			ItemID string `json:"item_id"`
			Notes  string `json:"notes"`
		} `json:"items"`
	}
	if err := json.NewDecoder(wishlistResp.Body).Decode(&wishlistPayload); err != nil {
		t.Fatalf("decode wishlist payload: %v", err)
	}
	if len(wishlistPayload.Items) != 1 {
		t.Fatalf("expected one wishlist item, got %d", len(wishlistPayload.Items))
	}
	notes := wishlistPayload.Items[0].Notes
	parts := strings.Split(notes, "[discovery_metadata]")
	if len(parts) < 2 {
		t.Fatalf("expected discovery metadata marker in notes, got %q", notes)
	}
	var metadata map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(parts[len(parts)-1])), &metadata); err != nil {
		t.Fatalf("decode discovery metadata json: %v", err)
	}
	for _, field := range []string{"listing_url", "seller", "stock_signal", "observed_price"} {
		if _, ok := metadata[field]; !ok {
			t.Fatalf("metadata missing %q: %+v", field, metadata)
		}
	}
}

func TestPokemonDiscoveryHandoffRejectsMissingCandidateID(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodPost, "/api/discovery/action", strings.NewReader(`{"type":"add_to_wishlist"}`), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 status, got=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode error payload: %v", err)
	}
	if payload["error"] != "failed_to_apply_discovery_action" {
		t.Fatalf("expected failed_to_apply_discovery_action, got %+v", payload)
	}
}

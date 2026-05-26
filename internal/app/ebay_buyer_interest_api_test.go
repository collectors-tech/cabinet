package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestEbayBuyerInterestPreviewMapsDestinationsAndWriteBackCapability(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	body := `{
		"source_account":"buyer@example.test",
		"items":[
			{"listing_id":" v1|123|0 ","title":"AFX Mega-G+","url":"https://www.ebay.com/itm/123","state":"watchlist","observed_at":"2026-05-24T00:00:00Z"},
			{"listing_id":"v1|456|0","title":"AFX Mustang","url":"https://www.ebay.com/itm/456","state":"cart","write_back_capability":"api_confirmed"}
		]
	}`
	resp := doRequest(t, a, http.MethodPost, "/api/providers/ebay/buyer-interest/preview", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusOK {
		t.Fatalf("preview status=%d body=%s", resp.Code, resp.Body.String())
	}

	var payload struct {
		Provider string `json:"provider"`
		Mode     string `json:"mode"`
		Items    []struct {
			ListingID           string `json:"listing_id"`
			State               string `json:"state"`
			Destination         string `json:"destination"`
			SourceProvider      string `json:"source_provider"`
			SourceAccount       string `json:"source_account"`
			ProvenanceKey       string `json:"provenance_key"`
			WriteBackCapability string `json:"write_back_capability"`
			WriteBackAllowed    bool   `json:"write_back_allowed"`
			WriteBackBlocker    string `json:"write_back_blocker"`
			OwnedInventory      bool   `json:"owned_inventory"`
		} `json:"items"`
		Summary map[string]int `json:"summary"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode preview payload: %v", err)
	}
	if payload.Provider != "ebay" || payload.Mode != "preview" {
		t.Fatalf("unexpected provider/mode: %+v", payload)
	}
	if len(payload.Items) != 2 {
		t.Fatalf("expected two mapped interest items, got %+v", payload.Items)
	}
	watched := payload.Items[0]
	if watched.ListingID != "v1|123|0" || watched.State != "watched" || watched.Destination != "wishlist" {
		t.Fatalf("watched interest did not map to wishlist with normalized listing/state: %+v", watched)
	}
	if watched.SourceProvider != "ebay" || watched.SourceAccount != "buyer@example.test" {
		t.Fatalf("expected source provenance on watched interest, got %+v", watched)
	}
	if watched.ProvenanceKey != "ebay:buyer@example.test:v1|123|0:watched" {
		t.Fatalf("unexpected provenance key %q", watched.ProvenanceKey)
	}
	if watched.WriteBackAllowed || watched.WriteBackBlocker != "ebay_api_capability_not_verified" {
		t.Fatalf("default write-back must be blocked until capability is verified: %+v", watched)
	}
	if watched.OwnedInventory {
		t.Fatalf("buyer interest preview must not mark synced interest as owned inventory")
	}

	cartLike := payload.Items[1]
	if cartLike.State != "cart_like" || cartLike.Destination != "discovery" {
		t.Fatalf("cart-like interest did not map to discovery: %+v", cartLike)
	}
	if !cartLike.WriteBackAllowed || cartLike.WriteBackCapability != "api_confirmed" {
		t.Fatalf("verified capability should allow write-back: %+v", cartLike)
	}
	if payload.Summary["wishlist"] != 1 || payload.Summary["discovery"] != 1 || payload.Summary["write_back_allowed"] != 1 || payload.Summary["write_back_blocked"] != 1 {
		t.Fatalf("unexpected summary: %+v", payload.Summary)
	}
}

func TestEbayBuyerInterestPreviewRejectsEmptyItems(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodPost, "/api/providers/ebay/buyer-interest/preview", strings.NewReader(`{"items":[]}`), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected empty preview request to be rejected, status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestEbayBuyerInterestImportPersistsWishlistAndDiscovery(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	body := "{\"source_account\":\"buyer@example.test\",\"items\":[{\"listing_id\":\"v1|watch|0\",\"title\":\"AFX Watch Slot Car\",\"url\":\"https://www.ebay.com/itm/watch\",\"state\":\"watched\"},{\"listing_id\":\"v1|like|0\",\"title\":\"AFX Like Slot Car\",\"url\":\"https://www.ebay.com/itm/like\",\"state\":\"liked\"}]}"
	resp := doRequest(t, a, http.MethodPost, "/api/providers/ebay/buyer-interest/import", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusOK {
		t.Fatalf("import status=%d body=%s", resp.Code, resp.Body.String())
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode import payload: %v", err)
	}
	if payload["provider"] != "ebay" || payload["mode"] != "import" {
		t.Fatalf("unexpected provider/mode: %+v", payload)
	}
	items := payload["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("expected two imported items, got %+v", items)
	}
	summary := payload["summary"].(map[string]any)
	if summary["wishlist"].(float64) != 1 || summary["discovery"].(float64) != 1 || summary["write_back_blocked"].(float64) != 2 {
		t.Fatalf("unexpected import summary: %+v", summary)
	}

	wishlistItem := items[0].(map[string]any)
	if wishlistItem["destination"] != "wishlist" || strings.TrimSpace(wishlistItem["item_id"].(string)) == "" || strings.TrimSpace(wishlistItem["persisted_id"].(string)) == "" {
		t.Fatalf("wishlist import did not return persisted item and entry ids: %+v", wishlistItem)
	}
	if wishlistItem["owned_inventory"].(bool) {
		t.Fatalf("wishlist buyer-interest import must remain separate from owned inventory")
	}
	var status, notes string
	var owned int
	if err := a.db.QueryRow("SELECT i.status, w.owned, w.notes FROM canonical_items i JOIN wishlist_entries w ON w.item_id = i.id WHERE i.id = ?", wishlistItem["item_id"].(string)).Scan(&status, &owned, &notes); err != nil {
		t.Fatalf("load persisted wishlist interest: %v", err)
	}
	if status != "wishlist" || owned != 0 {
		t.Fatalf("wishlist interest persisted with status=%q owned=%d", status, owned)
	}
	if !strings.Contains(notes, "[ebay_buyer_interest]") || !strings.Contains(notes, wishlistItem["provenance_key"].(string)) {
		t.Fatalf("wishlist interest notes did not preserve provenance: %q", notes)
	}

	discoveryItem := items[1].(map[string]any)
	if discoveryItem["destination"] != "discovery" || strings.TrimSpace(discoveryItem["candidate_id"].(string)) == "" || discoveryItem["persisted_id"] != discoveryItem["candidate_id"] {
		t.Fatalf("discovery import did not return persisted candidate id: %+v", discoveryItem)
	}
	var candidateSource, matchState, actionPayload string
	if err := a.db.QueryRow("SELECT c.source, m.state, a.payload_json FROM scanner_candidates c JOIN scanner_matches m ON m.candidate_id = c.id JOIN discovery_actions a ON a.candidate_id = c.id WHERE c.id = ?", discoveryItem["candidate_id"].(string)).Scan(&candidateSource, &matchState, &actionPayload); err != nil {
		t.Fatalf("load persisted discovery interest: %v", err)
	}
	if candidateSource != "ebay_buyer_interest" || matchState != "not_in_collection" {
		t.Fatalf("discovery interest persisted with source=%q match_state=%q", candidateSource, matchState)
	}
	if !strings.Contains(actionPayload, discoveryItem["provenance_key"].(string)) || !strings.Contains(actionPayload, "\"write_back_allowed\":false") {
		t.Fatalf("discovery action did not preserve provenance and blocked write-back: %s", actionPayload)
	}
}

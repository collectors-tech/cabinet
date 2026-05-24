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

package ebay

import "testing"

func TestMapBuyerInterestPreservesProvenanceAndDestinations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		state       string
		expectState string
		expectDest  string
	}{
		{name: "watched maps to wishlist", state: "watchlist", expectState: InterestStateWatched, expectDest: InterestDestinationWishlist},
		{name: "saved maps to wishlist", state: "saved", expectState: InterestStateSaved, expectDest: InterestDestinationWishlist},
		{name: "liked maps to discovery", state: "liked", expectState: InterestStateLiked, expectDest: InterestDestinationDiscovery},
		{name: "cart-like maps to discovery", state: "cart", expectState: InterestStateCartLike, expectDest: InterestDestinationDiscovery},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := MapBuyerInterest(BuyerInterestInput{
				ListingID:     " v1|123|0 ",
				Title:         " AFX Mega-G+ ",
				URL:           " https://www.ebay.com/itm/123 ",
				State:         tt.state,
				SourceAccount: "buyer@example.test",
				ObservedAt:    "2026-05-24T00:00:00Z",
			})

			if got.State != tt.expectState {
				t.Fatalf("expected state %q, got %q", tt.expectState, got.State)
			}
			if got.Destination != tt.expectDest {
				t.Fatalf("expected destination %q, got %q", tt.expectDest, got.Destination)
			}
			if got.SourceProvider != "ebay" || got.SourceAccount != "buyer@example.test" {
				t.Fatalf("expected ebay source provenance, got %+v", got)
			}
			if got.ProvenanceKey != "ebay:buyer@example.test:v1|123|0:"+tt.expectState {
				t.Fatalf("unexpected provenance key %q", got.ProvenanceKey)
			}
			if got.OwnedInventory {
				t.Fatalf("buyer-interest sync must not mark imported interest as owned inventory")
			}
		})
	}
}

func TestMapBuyerInterestRequiresVerifiedAPIForWriteBack(t *testing.T) {
	t.Parallel()

	unsupported := MapBuyerInterest(BuyerInterestInput{
		ListingID: "v1|456|0",
		State:     InterestStateWatched,
	})
	if unsupported.WriteBackAllowed {
		t.Fatalf("expected default write-back to be blocked")
	}
	if unsupported.WriteBackCapability != InterestWriteBackUnsupported {
		t.Fatalf("expected unsupported capability, got %q", unsupported.WriteBackCapability)
	}
	if unsupported.WriteBackBlocker != "ebay_api_capability_not_verified" {
		t.Fatalf("expected capability blocker, got %q", unsupported.WriteBackBlocker)
	}

	confirmed := MapBuyerInterest(BuyerInterestInput{
		ListingID:           "v1|456|0",
		State:               InterestStateWatched,
		WriteBackCapability: InterestWriteBackAPIConfirmed,
	})
	if !confirmed.WriteBackAllowed {
		t.Fatalf("expected verified API capability to allow write-back")
	}
	if confirmed.WriteBackBlocker != "" {
		t.Fatalf("expected no blocker for confirmed capability, got %q", confirmed.WriteBackBlocker)
	}
}

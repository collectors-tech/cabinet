package ebay

import "strings"

const (
	InterestStateWatched  = "watched"
	InterestStateSaved    = "saved"
	InterestStateLiked    = "liked"
	InterestStateCartLike = "cart_like"
)

const (
	InterestDestinationWishlist   = "wishlist"
	InterestDestinationDiscovery  = "discovery"
	InterestWriteBackUnsupported  = "unsupported"
	InterestWriteBackUnavailable  = "unavailable"
	InterestWriteBackAPIConfirmed = "api_confirmed"
)

type BuyerInterestInput struct {
	ListingID           string `json:"listing_id"`
	Title               string `json:"title"`
	URL                 string `json:"url"`
	State               string `json:"state"`
	SourceAccount       string `json:"source_account"`
	ObservedAt          string `json:"observed_at"`
	WriteBackCapability string `json:"write_back_capability"`
}

type BuyerInterestMapping struct {
	ListingID           string `json:"listing_id"`
	Title               string `json:"title"`
	URL                 string `json:"url"`
	State               string `json:"state"`
	Destination         string `json:"destination"`
	SourceProvider      string `json:"source_provider"`
	SourceAccount       string `json:"source_account"`
	ObservedAt          string `json:"observed_at"`
	ProvenanceKey       string `json:"provenance_key"`
	WriteBackCapability string `json:"write_back_capability"`
	WriteBackAllowed    bool   `json:"write_back_allowed"`
	WriteBackBlocker    string `json:"write_back_blocker,omitempty"`
	OwnedInventory      bool   `json:"owned_inventory"`
}

func MapBuyerInterest(in BuyerInterestInput) BuyerInterestMapping {
	state := normalizeBuyerInterestState(in.State)
	capability := normalizeBuyerInterestWriteBack(in.WriteBackCapability)
	listingID := strings.TrimSpace(in.ListingID)
	account := strings.TrimSpace(in.SourceAccount)
	provider := "ebay"

	out := BuyerInterestMapping{
		ListingID:           listingID,
		Title:               strings.TrimSpace(in.Title),
		URL:                 strings.TrimSpace(in.URL),
		State:               state,
		Destination:         destinationForBuyerInterest(state),
		SourceProvider:      provider,
		SourceAccount:       account,
		ObservedAt:          strings.TrimSpace(in.ObservedAt),
		ProvenanceKey:       buyerInterestProvenanceKey(provider, account, listingID, state),
		WriteBackCapability: capability,
		WriteBackAllowed:    capability == InterestWriteBackAPIConfirmed,
		OwnedInventory:      false,
	}
	if !out.WriteBackAllowed {
		out.WriteBackBlocker = "ebay_api_capability_not_verified"
	}
	return out
}

func normalizeBuyerInterestState(state string) string {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case InterestStateWatched, "watch", "watchlist":
		return InterestStateWatched
	case InterestStateSaved, "save":
		return InterestStateSaved
	case InterestStateLiked, "like":
		return InterestStateLiked
	case InterestStateCartLike, "cart", "in_cart":
		return InterestStateCartLike
	default:
		return InterestStateSaved
	}
}

func normalizeBuyerInterestWriteBack(capability string) string {
	switch strings.ToLower(strings.TrimSpace(capability)) {
	case InterestWriteBackAPIConfirmed:
		return InterestWriteBackAPIConfirmed
	case InterestWriteBackUnavailable:
		return InterestWriteBackUnavailable
	default:
		return InterestWriteBackUnsupported
	}
}

func destinationForBuyerInterest(state string) string {
	switch state {
	case InterestStateWatched, InterestStateSaved:
		return InterestDestinationWishlist
	default:
		return InterestDestinationDiscovery
	}
}

func buyerInterestProvenanceKey(provider, account, listingID, state string) string {
	parts := []string{strings.TrimSpace(provider), strings.TrimSpace(account), strings.TrimSpace(listingID), strings.TrimSpace(state)}
	return strings.Join(parts, ":")
}

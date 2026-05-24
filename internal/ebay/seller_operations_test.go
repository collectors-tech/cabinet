package ebay

import "testing"

func TestSellerOperationStatusesDefaultToBlocked(t *testing.T) {
	t.Parallel()

	statuses := SellerOperationStatuses(nil)

	if len(statuses) != 5 {
		t.Fatalf("expected all seller operation statuses, got %d: %+v", len(statuses), statuses)
	}
	for _, status := range statuses {
		if status.ReadAvailable || status.WriteAvailable || status.ConfirmationRequired {
			t.Fatalf("expected default unsupported status to expose no availability, got %+v", status)
		}
		if status.Capability != SellerCapabilityUnsupported || status.Blocker != "ebay_api_capability_not_verified" {
			t.Fatalf("expected truthful unsupported blocker, got %+v", status)
		}
	}
}

func TestSellerOperationStatusesKeepReadOnlySellerDataNonMutating(t *testing.T) {
	t.Parallel()

	statuses := SellerOperationStatuses([]SellerOperationCapabilityInput{
		{Operation: "messages", Capability: "read_only"},
		{Operation: "orders", Capability: SellerCapabilityReadOnlySync},
		{Operation: "offers", Capability: SellerCapabilityUnavailable},
	})
	byOperation := sellerStatusesByOperation(statuses)

	for _, operation := range []string{SellerOperationMessages, SellerOperationSoldOrders} {
		status := byOperation[operation]
		if !status.ReadAvailable || status.WriteAvailable || status.ConfirmationRequired {
			t.Fatalf("expected %s to be read-only and non-mutating, got %+v", operation, status)
		}
		if status.Blocker != "ebay_write_capability_not_verified" {
			t.Fatalf("expected write blocker for %s, got %+v", operation, status)
		}
	}
	if got := byOperation[SellerOperationOffers]; got.ReadAvailable || got.WriteAvailable || got.Blocker != "ebay_api_capability_unavailable" {
		t.Fatalf("expected unavailable offers capability to stay blocked, got %+v", got)
	}
}

func TestSellerOperationStatusesRequireConfirmationForConfirmedWrites(t *testing.T) {
	t.Parallel()

	statuses := SellerOperationStatuses([]SellerOperationCapabilityInput{
		{Operation: "seller_notifications", Capability: SellerCapabilityConfirmedAPI},
		{Operation: "fulfilment", Capability: "confirmed_write"},
	})
	byOperation := sellerStatusesByOperation(statuses)

	for _, operation := range []string{SellerOperationNotifications, SellerOperationFulfillment} {
		status := byOperation[operation]
		if !status.ReadAvailable || !status.WriteAvailable || !status.ConfirmationRequired {
			t.Fatalf("expected %s confirmed API to require confirmation before writes, got %+v", operation, status)
		}
		if status.Blocker != "" {
			t.Fatalf("expected no blocker for confirmed %s capability, got %+v", operation, status)
		}
	}
}

func sellerStatusesByOperation(statuses []SellerOperationCapabilityStatus) map[string]SellerOperationCapabilityStatus {
	out := map[string]SellerOperationCapabilityStatus{}
	for _, status := range statuses {
		out[status.Operation] = status
	}
	return out
}

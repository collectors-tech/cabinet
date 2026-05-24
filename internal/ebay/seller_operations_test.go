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

func TestPreviewSellerOperationActionBlocksReadOnlyWrites(t *testing.T) {
	t.Parallel()

	preview := PreviewSellerOperationAction(SellerOperationActionRequest{
		Operation:  "messages",
		Capability: SellerCapabilityReadOnlySync,
		Action:     "reply",
		Confirmed:  true,
	})

	if preview.Allowed || preview.RemoteWrite {
		t.Fatalf("read-only seller operation must not allow remote writes: %+v", preview)
	}
	if !preview.ReadAvailable || preview.WriteAvailable || preview.Blocker != "ebay_write_capability_not_verified" {
		t.Fatalf("read-only write preview should preserve truthful capability blocker, got %+v", preview)
	}
	if preview.Action != "reply" || preview.Operation != SellerOperationMessages {
		t.Fatalf("preview did not normalize operation/action: %+v", preview)
	}
}

func TestPreviewSellerOperationActionRequiresConfirmationForConfirmedWrites(t *testing.T) {
	t.Parallel()

	unconfirmed := PreviewSellerOperationAction(SellerOperationActionRequest{
		Operation:    "fulfilment",
		Capability:   SellerCapabilityConfirmedAPI,
		Action:       "ship",
		ReferenceID:  "order-123",
		DraftPayload: "{\"tracking_number\":\"TRACK-1\"}",
	})
	if unconfirmed.Allowed || unconfirmed.RemoteWrite || unconfirmed.Blocker != "ebay_seller_action_confirmation_required" {
		t.Fatalf("confirmed API write must still require explicit confirmation, got %+v", unconfirmed)
	}

	confirmed := PreviewSellerOperationAction(SellerOperationActionRequest{
		Operation:   "fulfilment",
		Capability:  SellerCapabilityConfirmedAPI,
		Action:      "ship",
		Confirmed:   true,
		ReferenceID: "order-123",
	})
	if !confirmed.Allowed || !confirmed.RemoteWrite || !confirmed.ConfirmationRequired || confirmed.Blocker != "" {
		t.Fatalf("confirmed seller operation write should be allowed only after confirmation, got %+v", confirmed)
	}
	if confirmed.Operation != SellerOperationFulfillment || confirmed.Action != "fulfill" || confirmed.ReferenceID != "order-123" {
		t.Fatalf("confirmed preview did not normalize expected fields: %+v", confirmed)
	}
}

func TestPreviewSellerOperationActionAllowsReadOnlySyncWithoutRemoteWrite(t *testing.T) {
	t.Parallel()

	preview := PreviewSellerOperationAction(SellerOperationActionRequest{
		Operation:  "sold_order",
		Capability: SellerCapabilityReadOnlySync,
		Action:     "refresh",
	})

	if !preview.Allowed || preview.RemoteWrite || preview.Blocker != "" {
		t.Fatalf("read-only sync should be allowed without a remote write, got %+v", preview)
	}
	if preview.Operation != SellerOperationSoldOrders || preview.Action != "sync" {
		t.Fatalf("sync preview did not normalize operation/action: %+v", preview)
	}
}

func sellerStatusesByOperation(statuses []SellerOperationCapabilityStatus) map[string]SellerOperationCapabilityStatus {
	out := map[string]SellerOperationCapabilityStatus{}
	for _, status := range statuses {
		out[status.Operation] = status
	}
	return out
}

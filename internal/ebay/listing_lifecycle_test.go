package ebay

import "testing"

type recordingListingLifecycleClient struct {
	commands []SellerListingLifecycleCommandRequest
}

func (c *recordingListingLifecycleClient) ExecuteSellerListingLifecycleCommand(req SellerListingLifecycleCommandRequest) (SellerListingLifecycleCommandResponse, error) {
	c.commands = append(c.commands, req)
	return SellerListingLifecycleCommandResponse{
		Provider:  "ebay",
		Command:   req.Command,
		ListingID: firstNonEmpty(req.ListingID, "listing-mock-1"),
		DraftID:   firstNonEmpty(req.DraftID, "draft-mock-1"),
		Status:    "mocked_ebay_response",
	}, nil
}

func TestPreviewSellerListingLifecycleCommandsGateRemoteWrites(t *testing.T) {
	t.Parallel()

	draft := PreviewSellerListingLifecycleCommand(SellerListingLifecycleCommandRequest{
		Command:    "draft",
		Capability: SellerListingCapabilityDraftOnly,
		ItemID:     "item-1",
		Title:      "Cabinet listing draft",
	})
	if !draft.Allowed || !draft.LocalOnly || draft.RemoteWrite || draft.ConfirmationRequired || draft.Blocker != "" {
		t.Fatalf("draft creation should be allowed as local-only without remote write, got %+v", draft)
	}

	unconfirmedPublish := PreviewSellerListingLifecycleCommand(SellerListingLifecycleCommandRequest{
		Command:    "publish",
		Capability: SellerListingCapabilityConfirmedAPI,
		DraftID:    "draft-1",
	})
	if unconfirmedPublish.Allowed || !unconfirmedPublish.RemoteWrite || !unconfirmedPublish.ConfirmationRequired || unconfirmedPublish.Blocker != "ebay_listing_lifecycle_confirmation_required" {
		t.Fatalf("publish must require confirmation before remote write, got %+v", unconfirmedPublish)
	}

	confirmedPublish := PreviewSellerListingLifecycleCommand(SellerListingLifecycleCommandRequest{
		Command:    "publish",
		Capability: SellerListingCapabilityConfirmedAPI,
		DraftID:    "draft-1",
		Confirmed:  true,
	})
	if !confirmedPublish.Allowed || !confirmedPublish.RemoteWrite || !confirmedPublish.ConfirmationRequired || confirmedPublish.Blocker != "" {
		t.Fatalf("confirmed publish should be allowed for remote write preview, got %+v", confirmedPublish)
	}

	revise := PreviewSellerListingLifecycleCommand(SellerListingLifecycleCommandRequest{
		Command:    "revise",
		Capability: SellerListingCapabilityDraftOnly,
		ListingID:  "listing-1",
		Confirmed:  true,
	})
	if revise.Allowed || revise.RemoteWrite || revise.Blocker != "ebay_listing_write_capability_not_verified" {
		t.Fatalf("draft-only capability must block revise remote writes, got %+v", revise)
	}
}

func TestExecuteSellerListingLifecycleCommandUsesMockedEbayResponses(t *testing.T) {
	t.Parallel()

	client := &recordingListingLifecycleClient{}
	for _, command := range []string{"publish", "revise", "end", "relist"} {
		command := command
		t.Run(command, func(t *testing.T) {
			execution := ExecuteSellerListingLifecycleCommand(SellerListingLifecycleCommandRequest{
				Command:    command,
				Capability: SellerListingCapabilityConfirmedAPI,
				ListingID:  "listing-1",
				DraftID:    "draft-1",
				Confirmed:  true,
			}, client)
			if !execution.Allowed || !execution.Executed || !execution.RemoteWrite || execution.LocalOnly || execution.Blocker != "" {
				t.Fatalf("confirmed %s should execute through mocked eBay client, got %+v", command, execution)
			}
			if execution.Response == nil || execution.Response.Provider != "ebay" || execution.Response.Status != "mocked_ebay_response" {
				t.Fatalf("expected mocked eBay response for %s, got %+v", command, execution.Response)
			}
		})
	}
	if len(client.commands) != 4 {
		t.Fatalf("expected mocked client to receive four lifecycle commands, got %d", len(client.commands))
	}
}

func TestExecuteSellerListingLifecycleCommandBlocksUnconfirmedWrites(t *testing.T) {
	t.Parallel()

	client := &recordingListingLifecycleClient{}
	execution := ExecuteSellerListingLifecycleCommand(SellerListingLifecycleCommandRequest{
		Command:    "end",
		Capability: SellerListingCapabilityConfirmedAPI,
		ListingID:  "listing-1",
	}, client)
	if execution.Allowed || execution.Executed || !execution.RemoteWrite || execution.Blocker != "ebay_listing_lifecycle_confirmation_required" {
		t.Fatalf("unconfirmed lifecycle write should stay blocked, got %+v", execution)
	}
	if len(client.commands) != 0 {
		t.Fatalf("blocked lifecycle write must not call mocked eBay client, got %d calls", len(client.commands))
	}
}

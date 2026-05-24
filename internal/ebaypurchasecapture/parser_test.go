package ebaypurchasecapture

import "testing"

func TestParsePurchaseCardHTMLPreservesPurchasedCardMetadata(t *testing.T) {
	card := ParsePurchaseCardHTML(`
		<div class="order-card" input-listing-id="316046161178">
			<h3 class="title-heading"><a href="https://www.ebay.com.au/itm/316046161178?var=615514115326">-| DECK BUILDER Part 1 |- Build your deck</a></h3>
			<img src="https://i.ebayimg.com/images/g/example/s-l1600.jpg" />
			<div class="container-item-col__info-item-info-aspectValuesList">
				<div>Choose Single Or Playset: Single</div>
				<div>Card: Accompanying Flute TWM 142 (142/167)</div>
				<div>Quantity: 4</div>
			</div></div>
			<div class="container-item-col__info-item-info-additionalPrice"><div><span>AU $2.40</span></div></div>
		</div>
	`)

	if card.ListingTitle != "-| DECK BUILDER Part 1 |- Build your deck" {
		t.Fatalf("listing title = %q", card.ListingTitle)
	}
	if card.PurchasedIdentity != "Accompanying Flute TWM 142 (142/167)" {
		t.Fatalf("purchased identity = %q", card.PurchasedIdentity)
	}
	if got := card.Aspects["Choose Single Or Playset"]; got != "Single" {
		t.Fatalf("playset aspect = %q", got)
	}
	if card.Quantity != 4 {
		t.Fatalf("quantity = %d", card.Quantity)
	}
	if card.ListingID != "316046161178" {
		t.Fatalf("listing id = %q", card.ListingID)
	}
	if card.VariationID != "615514115326" {
		t.Fatalf("variation id = %q", card.VariationID)
	}
	if card.ItemPrice != "AU $2.40" {
		t.Fatalf("item price = %q", card.ItemPrice)
	}
}

func TestParsePurchaseCardHTMLCapturesSellerAndPassiveActions(t *testing.T) {
	card := ParsePurchaseCardHTML(`
		<div class="order-card">
			<a href="https://www.ebay.com.au/itm/316046161178?var=615514115326">View item</a>
			<a href="https://www.ebay.com.au/usr/nearmintormeta"><span>nearmintormeta</span></a>
			<textarea id="note-123" aria-label="Note" maxlength="250"></textarea>
			<button data-action-name="SAVE_NOTE" data-href="https://www.ebay.com.au/myb/SaveNote?itemId=316046161178&variationId=615514115326&transaction_id=10080684936020" aria-disabled="true">Save note</button>
			<button data-action-name="DELETE_NOTE" data-href="https://www.ebay.com.au/myb/DeleteNote?itemId=316046161178&variationId=615514115326&transaction_id=10080684936020">Delete note</button>
			<a data-action-name="LEAVE_FEEDBACK_FOR_SELLER" href="https://www.ebay.com.au/fdbk/leave_feedback?item_id=316046161178&transaction_id=10080684936020">Leave feedback</a>
			<a data-action-name="CONTACT_SELLER" href="https://www.ebay.com.au/help/contact_us?id=123&item_id=316046161178&transId=10080684936020">Contact seller</a>
			<a data-action-name="START_RETURN" href="https://www.ebay.com.au/ret/start?itemId=316046161178&transactionId=10080684936020">Start return</a>
			<button data-action-name="HIDE" data-href="https://www.ebay.com.au/mye/ajax/bulk_update/hide?orderId=20-14595-70928">Hide order</button>
		</div>
	`)

	if card.SellerUsername != "nearmintormeta" {
		t.Fatalf("seller username = %q", card.SellerUsername)
	}
	if card.SellerProfileURL != "https://www.ebay.com.au/usr/nearmintormeta" {
		t.Fatalf("seller profile = %q", card.SellerProfileURL)
	}
	if card.NoteCapability.TextareaID != "note-123" || card.NoteCapability.MaxLength != 250 {
		t.Fatalf("note capability = %+v", card.NoteCapability)
	}
	if card.TransactionID != "10080684936020" {
		t.Fatalf("transaction id = %q", card.TransactionID)
	}
	if card.OrderID != "20-14595-70928" {
		t.Fatalf("order id = %q", card.OrderID)
	}

	actions := map[string]Action{}
	for _, action := range card.Actions {
		actions[action.Name] = action
	}
	for _, name := range []string{"SAVE_NOTE", "DELETE_NOTE", "LEAVE_FEEDBACK_FOR_SELLER", "CONTACT_SELLER", "START_RETURN", "HIDE"} {
		if _, ok := actions[name]; !ok {
			t.Fatalf("missing passive action %s in %+v", name, card.Actions)
		}
	}
	if actions["SAVE_NOTE"].Enabled {
		t.Fatal("SAVE_NOTE should preserve disabled state from aria-disabled=true")
	}
	if actions["HIDE"].Metadata["orderId"] != "20-14595-70928" {
		t.Fatalf("hide metadata = %+v", actions["HIDE"].Metadata)
	}
	if len(card.NoteCapability.Actions) != 2 {
		t.Fatalf("note actions = %+v", card.NoteCapability.Actions)
	}
}

package ebaypurchasecapture

import "testing"

func TestGroupPurchaseCardsByOrderCreatesParentWithChildItems(t *testing.T) {
	cards := []PurchaseCard{
		{
			OrderID:           "20-14595-70928",
			ListingID:         "316046161178",
			VariationID:       "615514115326",
			TransactionID:     "10080684936020",
			PurchasedIdentity: "Accompanying Flute TWM 142 (142/167)",
			Quantity:          4,
			ItemPrice:         "AU $2.40",
			SellerUsername:    "nearmintormeta",
			OrderTotal:        "AU $8.10",
			Currency:          "AUD",
		},
		{
			OrderID:           "20-14595-70928",
			ListingID:         "316046161179",
			VariationID:       "615514115327",
			TransactionID:     "10080684936021",
			PurchasedIdentity: "Buddy-Buddy Poffin TEF 144 (144/162)",
			Quantity:          2,
			ItemPrice:         "AU $5.70",
			SellerUsername:    "nearmintormeta",
		},
	}

	orders := GroupPurchaseCardsByOrder(cards)

	if len(orders) != 1 {
		t.Fatalf("orders = %d", len(orders))
	}
	order := orders[0]
	if order.OrderID != "20-14595-70928" {
		t.Fatalf("order id = %q", order.OrderID)
	}
	if order.OrderTotal != "AU $8.10" || order.Currency != "AUD" {
		t.Fatalf("order totals = total %q currency %q", order.OrderTotal, order.Currency)
	}
	if len(order.SellerUsernames) != 1 || order.SellerUsernames[0] != "nearmintormeta" {
		t.Fatalf("sellers = %+v", order.SellerUsernames)
	}
	if len(order.Items) != 2 {
		t.Fatalf("items = %d", len(order.Items))
	}
	if order.Items[0].PurchasedIdentity != "Accompanying Flute TWM 142 (142/167)" {
		t.Fatalf("first child = %+v", order.Items[0])
	}
	if order.Items[1].PurchasedIdentity != "Buddy-Buddy Poffin TEF 144 (144/162)" {
		t.Fatalf("second child = %+v", order.Items[1])
	}
}

func TestGroupPurchaseCardsByOrderKeepsOneItemOrdersSeparate(t *testing.T) {
	cards := []PurchaseCard{
		{OrderID: "20-11111-11111", ListingID: "111", TransactionID: "tx-111", PurchasedIdentity: "First card", SellerUsername: "seller-a"},
		{OrderID: "20-22222-22222", ListingID: "222", TransactionID: "tx-222", PurchasedIdentity: "Second card", SellerUsername: "seller-b"},
	}

	orders := GroupPurchaseCardsByOrder(cards)

	if len(orders) != 2 {
		t.Fatalf("orders = %d", len(orders))
	}
	if orders[0].OrderID != "20-11111-11111" || orders[1].OrderID != "20-22222-22222" {
		t.Fatalf("order ids = %+v", orders)
	}
	if len(orders[0].Items) != 1 || len(orders[1].Items) != 1 {
		t.Fatalf("child item counts = %d/%d", len(orders[0].Items), len(orders[1].Items))
	}
}

func TestGroupPurchaseCardsByOrderMergesRepeatedCaptures(t *testing.T) {
	cards := []PurchaseCard{
		{
			OrderID:           "20-14595-70928",
			ListingID:         "316046161178",
			VariationID:       "615514115326",
			TransactionID:     "10080684936020",
			PurchasedIdentity: "Accompanying Flute TWM 142 (142/167)",
			Quantity:          4,
			SellerUsername:    "nearmintormeta",
		},
		{
			OrderID:           "20-14595-70928",
			ListingID:         "316046161178",
			VariationID:       "615514115326",
			TransactionID:     "10080684936020",
			PurchasedIdentity: "Accompanying Flute TWM 142 (142/167)",
			Quantity:          4,
			SellerUsername:    "nearmintormeta",
			ItemStatus:        "Delivered",
			TrackingStatus:    "Tracking provided",
		},
	}

	orders := GroupPurchaseCardsByOrder(cards)

	if len(orders) != 1 {
		t.Fatalf("orders = %d", len(orders))
	}
	if len(orders[0].Items) != 1 {
		t.Fatalf("repeated capture produced %d child items", len(orders[0].Items))
	}
	item := orders[0].Items[0]
	if item.ItemStatus != "Delivered" || item.TrackingStatus != "Tracking provided" {
		t.Fatalf("merged item status = %+v", item)
	}
}

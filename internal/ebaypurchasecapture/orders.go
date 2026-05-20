package ebaypurchasecapture

import "sort"

// PurchaseOrder is the parent record Cabinet uses for purchase inbox grouping.
// Order-level metadata lives here, while each captured item card remains a child
// record so reconciliation and landed-cost allocation can reason at both levels.
type PurchaseOrder struct {
	OrderID           string
	SellerUsernames   []string
	OrderTotal        string
	Currency          string
	Shipping          string
	Tax               string
	ImportCharges     string
	DestinationMarker string
	OrderStatus       string
	OrderDetailURL    string
	Items             []PurchaseCard
}

// GroupPurchaseCardsByOrder folds parsed eBay purchase-history cards into stable
// order parent records. Repeated captures merge into the same order and item key
// rather than duplicating child items.
func GroupPurchaseCardsByOrder(cards []PurchaseCard) []PurchaseOrder {
	orderByID := map[string]*PurchaseOrder{}
	itemKeysByOrder := map[string]map[string]int{}
	orderIDs := make([]string, 0)

	for _, card := range cards {
		orderID := firstNonEmpty(card.OrderID, fallbackOrderID(card))
		order, ok := orderByID[orderID]
		if !ok {
			order = &PurchaseOrder{OrderID: orderID}
			orderByID[orderID] = order
			itemKeysByOrder[orderID] = map[string]int{}
			orderIDs = append(orderIDs, orderID)
		}

		mergeOrderMetadata(order, card)
		key := purchaseItemKey(card)
		if existingIndex, exists := itemKeysByOrder[orderID][key]; exists {
			order.Items[existingIndex] = mergePurchaseCard(order.Items[existingIndex], card)
			continue
		}
		itemKeysByOrder[orderID][key] = len(order.Items)
		order.Items = append(order.Items, card)
	}

	sort.Strings(orderIDs)
	orders := make([]PurchaseOrder, 0, len(orderIDs))
	for _, orderID := range orderIDs {
		orders = append(orders, *orderByID[orderID])
	}
	return orders
}

func mergeOrderMetadata(order *PurchaseOrder, card PurchaseCard) {
	order.OrderTotal = firstNonEmpty(order.OrderTotal, card.OrderTotal)
	order.Currency = firstNonEmpty(order.Currency, card.Currency)
	order.Shipping = firstNonEmpty(order.Shipping, card.Shipping)
	order.Tax = firstNonEmpty(order.Tax, card.Tax)
	order.ImportCharges = firstNonEmpty(order.ImportCharges, card.ImportCharges)
	order.DestinationMarker = firstNonEmpty(order.DestinationMarker, card.DestinationMarker)
	order.OrderStatus = firstNonEmpty(order.OrderStatus, card.OrderStatus)
	order.OrderDetailURL = firstNonEmpty(order.OrderDetailURL, card.OrderDetailURL)
	if card.SellerUsername != "" && !containsString(order.SellerUsernames, card.SellerUsername) {
		order.SellerUsernames = append(order.SellerUsernames, card.SellerUsername)
		sort.Strings(order.SellerUsernames)
	}
}

func mergePurchaseCard(existing, incoming PurchaseCard) PurchaseCard {
	existing.ListingID = firstNonEmpty(existing.ListingID, incoming.ListingID)
	existing.VariationID = firstNonEmpty(existing.VariationID, incoming.VariationID)
	existing.TransactionID = firstNonEmpty(existing.TransactionID, incoming.TransactionID)
	existing.OrderID = firstNonEmpty(existing.OrderID, incoming.OrderID)
	existing.ListingTitle = firstNonEmpty(existing.ListingTitle, incoming.ListingTitle)
	existing.PurchasedIdentity = firstNonEmpty(existing.PurchasedIdentity, incoming.PurchasedIdentity)
	existing.ItemPrice = firstNonEmpty(existing.ItemPrice, incoming.ItemPrice)
	existing.ImageURL = firstNonEmpty(existing.ImageURL, incoming.ImageURL)
	existing.ItemURL = firstNonEmpty(existing.ItemURL, incoming.ItemURL)
	existing.SellerUsername = firstNonEmpty(existing.SellerUsername, incoming.SellerUsername)
	existing.SellerProfileURL = firstNonEmpty(existing.SellerProfileURL, incoming.SellerProfileURL)
	existing.ItemStatus = firstNonEmpty(existing.ItemStatus, incoming.ItemStatus)
	existing.TrackingStatus = firstNonEmpty(existing.TrackingStatus, incoming.TrackingStatus)
	if existing.Quantity == 0 {
		existing.Quantity = incoming.Quantity
	}
	if existing.Aspects == nil {
		existing.Aspects = map[string]string{}
	}
	for key, value := range incoming.Aspects {
		if existing.Aspects[key] == "" {
			existing.Aspects[key] = value
		}
	}
	if existing.NoteCapability.TextareaID == "" {
		existing.NoteCapability = incoming.NoteCapability
	}
	if len(existing.Actions) == 0 {
		existing.Actions = incoming.Actions
	}
	return existing
}

func purchaseItemKey(card PurchaseCard) string {
	return firstNonEmpty(
		card.TransactionID,
		joinKey(card.ListingID, card.VariationID),
		joinKey(card.ListingID, card.PurchasedIdentity),
		card.PurchasedIdentity,
		card.ListingTitle,
	)
}

func fallbackOrderID(card PurchaseCard) string {
	return "orderless:" + purchaseItemKey(card)
}

func joinKey(values ...string) string {
	out := ""
	for _, value := range values {
		if value == "" {
			continue
		}
		if out != "" {
			out += ":"
		}
		out += value
	}
	return out
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

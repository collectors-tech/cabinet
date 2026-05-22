package ebaypurchasecapture

import "sort"

// PurchaseOrder is the parent record Cabinet uses for purchase inbox grouping.
// Order-level metadata lives here, while each captured item card remains a child
// record so reconciliation and landed-cost allocation can reason at both levels.
type PurchaseOrder struct {
	OrderID           string         `json:"order_id,omitempty"`
	SellerUsernames   []string       `json:"seller_usernames,omitempty"`
	OrderTotal        string         `json:"order_total,omitempty"`
	Currency          string         `json:"currency,omitempty"`
	Shipping          string         `json:"shipping,omitempty"`
	Tax               string         `json:"tax,omitempty"`
	ImportCharges     string         `json:"import_charges,omitempty"`
	DestinationMarker string         `json:"destination_marker,omitempty"`
	OrderStatus       string         `json:"order_status,omitempty"`
	OrderDetailURL    string         `json:"order_detail_url,omitempty"`
	Items             []PurchaseCard `json:"items,omitempty"`
}

// PurchaseInboxReview is the review-ready representation used by the Purchase
// Inbox before any link or convert action mutates Cabinet inventory.
type PurchaseInboxReview struct {
	Order            PurchaseOrder             `json:"order"`
	Status           string                    `json:"status"`
	Items            []PurchaseInboxItemReview `json:"items"`
	SuggestedActions []PurchaseInboxAction     `json:"suggested_actions"`
}

type PurchaseInboxItemReview struct {
	Item             PurchaseCard          `json:"item"`
	Status           string                `json:"status"`
	MissingFields    []string              `json:"missing_fields"`
	SuggestedActions []PurchaseInboxAction `json:"suggested_actions"`
}

type PurchaseInboxAction struct {
	ID                   string `json:"id"`
	Label                string `json:"label"`
	Scope                string `json:"scope"`
	TargetKey            string `json:"target_key"`
	RequiresConfirmation bool   `json:"requires_confirmation"`
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

func BuildPurchaseInboxReviews(cards []PurchaseCard) []PurchaseInboxReview {
	orders := GroupPurchaseCardsByOrder(cards)
	reviews := make([]PurchaseInboxReview, 0, len(orders))
	for _, order := range orders {
		itemReviews := make([]PurchaseInboxItemReview, 0, len(order.Items))
		status := "ready"
		for _, item := range order.Items {
			itemReview := buildPurchaseInboxItemReview(item)
			if itemReview.Status == "needs_review" {
				status = "needs_review"
			}
			itemReviews = append(itemReviews, itemReview)
		}
		reviews = append(reviews, PurchaseInboxReview{
			Order:            order,
			Status:           status,
			Items:            itemReviews,
			SuggestedActions: purchaseInboxOrderActions(order, status),
		})
	}
	return reviews
}

func buildPurchaseInboxItemReview(item PurchaseCard) PurchaseInboxItemReview {
	missing := purchaseInboxMissingFields(item)
	status := "ready_to_link_or_convert"
	if len(missing) > 0 {
		status = "needs_review"
	}
	return PurchaseInboxItemReview{
		Item:             item,
		Status:           status,
		MissingFields:    missing,
		SuggestedActions: purchaseInboxItemActions(item, len(missing) == 0),
	}
}

func purchaseInboxMissingFields(item PurchaseCard) []string {
	missing := make([]string, 0, 3)
	if firstNonEmpty(item.PurchasedIdentity, item.ListingTitle) == "" {
		missing = append(missing, "purchased_identity")
	}
	if item.Quantity == 0 {
		missing = append(missing, "quantity")
	}
	if firstNonEmpty(item.ItemPrice, item.OrderTotal) == "" {
		missing = append(missing, "item_price")
	}
	return missing
}

func purchaseInboxOrderActions(order PurchaseOrder, status string) []PurchaseInboxAction {
	actions := []PurchaseInboxAction{{
		ID:                   "review_purchase_order",
		Label:                "Review purchase order",
		Scope:                "order",
		TargetKey:            order.OrderID,
		RequiresConfirmation: false,
	}}
	if status == "ready" {
		actions = append(actions, PurchaseInboxAction{
			ID:                   "mark_order_reviewed",
			Label:                "Mark order reviewed",
			Scope:                "order",
			TargetKey:            order.OrderID,
			RequiresConfirmation: true,
		})
	}
	return actions
}

func purchaseInboxItemActions(item PurchaseCard, ready bool) []PurchaseInboxAction {
	targetKey := purchaseItemKey(item)
	if !ready {
		return []PurchaseInboxAction{{
			ID:                   "complete_purchase_item_fields",
			Label:                "Complete missing purchase item fields",
			Scope:                "item",
			TargetKey:            targetKey,
			RequiresConfirmation: false,
		}}
	}
	return []PurchaseInboxAction{
		{
			ID:                   "link_existing_inventory_item",
			Label:                "Link existing inventory item",
			Scope:                "item",
			TargetKey:            targetKey,
			RequiresConfirmation: true,
		},
		{
			ID:                   "convert_to_inventory_item",
			Label:                "Convert to inventory item",
			Scope:                "item",
			TargetKey:            targetKey,
			RequiresConfirmation: true,
		},
	}
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
	return PurchaseItemKey(card)
}

// PurchaseItemKey returns the stable review/action key for a captured purchase
// item so API callers can confirm actions against the same target Cabinet shows
// in Purchase Inbox reviews.
func PurchaseItemKey(card PurchaseCard) string {
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

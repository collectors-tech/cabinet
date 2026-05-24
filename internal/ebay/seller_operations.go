package ebay

import "strings"

const (
	SellerCapabilityUnsupported  = "unsupported"
	SellerCapabilityUnavailable  = "unavailable"
	SellerCapabilityReadOnlySync = "read_only_sync"
	SellerCapabilityConfirmedAPI = "confirmed_api"
)

const (
	SellerOperationMessages      = "messages"
	SellerOperationNotifications = "notifications"
	SellerOperationSoldOrders    = "sold_orders"
	SellerOperationFulfillment   = "fulfillment"
	SellerOperationOffers        = "offers"
)

type SellerOperationCapabilityInput struct {
	Operation  string `json:"operation"`
	Capability string `json:"capability"`
}

type SellerOperationCapabilityStatus struct {
	Operation            string `json:"operation"`
	Capability           string `json:"capability"`
	ReadAvailable        bool   `json:"read_available"`
	WriteAvailable       bool   `json:"write_available"`
	ConfirmationRequired bool   `json:"confirmation_required"`
	Blocker              string `json:"blocker,omitempty"`
}

func SellerOperationStatuses(inputs []SellerOperationCapabilityInput) []SellerOperationCapabilityStatus {
	configured := map[string]string{}
	for _, in := range inputs {
		operation := normalizeSellerOperation(in.Operation)
		if operation == "" {
			continue
		}
		configured[operation] = normalizeSellerCapability(in.Capability)
	}

	operations := []string{
		SellerOperationMessages,
		SellerOperationNotifications,
		SellerOperationSoldOrders,
		SellerOperationFulfillment,
		SellerOperationOffers,
	}
	statuses := make([]SellerOperationCapabilityStatus, 0, len(operations))
	for _, operation := range operations {
		capability := configured[operation]
		if capability == "" {
			capability = SellerCapabilityUnsupported
		}
		statuses = append(statuses, sellerOperationStatus(operation, capability))
	}
	return statuses
}

func sellerOperationStatus(operation, capability string) SellerOperationCapabilityStatus {
	status := SellerOperationCapabilityStatus{
		Operation:  operation,
		Capability: normalizeSellerCapability(capability),
	}
	switch status.Capability {
	case SellerCapabilityConfirmedAPI:
		status.ReadAvailable = true
		status.WriteAvailable = true
		status.ConfirmationRequired = true
	case SellerCapabilityReadOnlySync:
		status.ReadAvailable = true
		status.Blocker = "ebay_write_capability_not_verified"
	case SellerCapabilityUnavailable:
		status.Blocker = "ebay_api_capability_unavailable"
	default:
		status.Capability = SellerCapabilityUnsupported
		status.Blocker = "ebay_api_capability_not_verified"
	}
	return status
}

func normalizeSellerOperation(operation string) string {
	switch strings.ToLower(strings.TrimSpace(operation)) {
	case SellerOperationMessages, "seller_messages", "message":
		return SellerOperationMessages
	case SellerOperationNotifications, "seller_notifications", "notification":
		return SellerOperationNotifications
	case SellerOperationSoldOrders, "orders", "sold_order", "order":
		return SellerOperationSoldOrders
	case SellerOperationFulfillment, "fulfilment", "shipping":
		return SellerOperationFulfillment
	case SellerOperationOffers, "negotiation", "seller_offers", "offer":
		return SellerOperationOffers
	default:
		return ""
	}
}

func normalizeSellerCapability(capability string) string {
	switch strings.ToLower(strings.TrimSpace(capability)) {
	case SellerCapabilityConfirmedAPI, "confirmed_write", "write_confirmed":
		return SellerCapabilityConfirmedAPI
	case SellerCapabilityReadOnlySync, "read_only", "sync_only":
		return SellerCapabilityReadOnlySync
	case SellerCapabilityUnavailable:
		return SellerCapabilityUnavailable
	default:
		return SellerCapabilityUnsupported
	}
}

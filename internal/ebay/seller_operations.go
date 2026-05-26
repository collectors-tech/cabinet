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

type SellerOperationActionRequest struct {
	Operation    string `json:"operation"`
	Capability   string `json:"capability"`
	Action       string `json:"action"`
	Confirmed    bool   `json:"confirmed"`
	ReferenceID  string `json:"reference_id,omitempty"`
	DraftPayload string `json:"draft_payload,omitempty"`
}

type SellerOperationActionPreview struct {
	Operation            string `json:"operation"`
	Action               string `json:"action"`
	ReferenceID          string `json:"reference_id,omitempty"`
	Capability           string `json:"capability"`
	ReadAvailable        bool   `json:"read_available"`
	WriteAvailable       bool   `json:"write_available"`
	ConfirmationRequired bool   `json:"confirmation_required"`
	Confirmed            bool   `json:"confirmed"`
	Allowed              bool   `json:"allowed"`
	RemoteWrite          bool   `json:"remote_write"`
	Blocker              string `json:"blocker,omitempty"`
}

type SellerOperationActionExecution struct {
	SellerOperationActionPreview
	Executed  bool                       `json:"executed"`
	LocalOnly bool                       `json:"local_only"`
	Status    string                     `json:"status"`
	Result    *SellerOperationReadResult `json:"result,omitempty"`
}

type SellerOperationReadResult struct {
	Operation string                      `json:"operation"`
	Source    string                      `json:"source"`
	Records   []SellerOperationReadRecord `json:"records"`
	Summary   map[string]int              `json:"summary"`
}

type SellerOperationReadRecord struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Status string `json:"status"`
	Title  string `json:"title"`
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

func PreviewSellerOperationAction(req SellerOperationActionRequest) SellerOperationActionPreview {
	status := sellerOperationStatus(req.Operation, req.Capability)
	preview := SellerOperationActionPreview{
		Operation:            status.Operation,
		Action:               normalizeSellerAction(req.Action),
		ReferenceID:          strings.TrimSpace(req.ReferenceID),
		Capability:           status.Capability,
		ReadAvailable:        status.ReadAvailable,
		WriteAvailable:       status.WriteAvailable,
		ConfirmationRequired: status.ConfirmationRequired,
		Confirmed:            req.Confirmed,
	}
	if preview.Operation == "" {
		preview.Blocker = "ebay_seller_operation_required"
		return preview
	}
	if preview.Action == "" {
		preview.Blocker = "ebay_seller_action_required"
		return preview
	}
	if preview.Action == "sync" {
		preview.Allowed = status.ReadAvailable
		preview.Blocker = status.Blocker
		if preview.Allowed {
			preview.Blocker = ""
		}
		return preview
	}
	if !status.WriteAvailable {
		preview.Blocker = status.Blocker
		if preview.Blocker == "" {
			preview.Blocker = "ebay_write_capability_not_verified"
		}
		return preview
	}
	if status.ConfirmationRequired && !req.Confirmed {
		preview.Blocker = "ebay_seller_action_confirmation_required"
		return preview
	}
	preview.Allowed = true
	preview.RemoteWrite = true
	return preview
}

func ExecuteSellerOperationAction(req SellerOperationActionRequest) SellerOperationActionExecution {
	preview := PreviewSellerOperationAction(req)
	execution := SellerOperationActionExecution{
		SellerOperationActionPreview: preview,
	}
	if preview.Operation == "" || preview.Action == "" {
		execution.Status = "invalid"
		return execution
	}
	if !preview.Allowed {
		execution.Status = "blocked"
		return execution
	}
	if preview.RemoteWrite {
		execution.Allowed = false
		execution.Blocker = "ebay_seller_remote_write_execution_not_configured"
		execution.Status = "blocked"
		return execution
	}
	execution.Executed = true
	execution.LocalOnly = true
	execution.Status = "read_only_sync_ready"
	result := SellerOperationReadResultFor(preview.Operation)
	execution.Result = &result
	return execution
}

func SellerOperationReadResultFor(operation string) SellerOperationReadResult {
	operation = normalizeSellerOperation(operation)
	result := SellerOperationReadResult{
		Operation: operation,
		Source:    "local_read_model",
		Summary:   map[string]int{},
	}
	switch operation {
	case SellerOperationMessages:
		result.Records = []SellerOperationReadRecord{{ID: "msg-local-1", Kind: "seller_message", Status: "unread", Title: "Buyer message awaiting review"}}
		result.Summary["unread_messages"] = 1
	case SellerOperationNotifications:
		result.Records = []SellerOperationReadRecord{{ID: "note-local-1", Kind: "seller_notification", Status: "unread", Title: "Seller notification awaiting review"}}
		result.Summary["unread_notifications"] = 1
	case SellerOperationSoldOrders:
		result.Records = []SellerOperationReadRecord{{ID: "order-local-1", Kind: "sold_order", Status: "awaiting_fulfillment", Title: "Sold order awaiting fulfilment"}}
		result.Summary["open_orders"] = 1
	case SellerOperationFulfillment:
		result.Records = []SellerOperationReadRecord{{ID: "fulfillment-local-1", Kind: "fulfillment", Status: "ready_to_ship", Title: "Shipment ready for tracking update"}}
		result.Summary["shipments_ready"] = 1
	case SellerOperationOffers:
		result.Records = []SellerOperationReadRecord{{ID: "offer-local-1", Kind: "seller_offer", Status: "pending", Title: "Offer awaiting seller decision"}}
		result.Summary["pending_offers"] = 1
	}
	result.Summary["records"] = len(result.Records)
	return result
}

func sellerOperationStatus(operation, capability string) SellerOperationCapabilityStatus {
	status := SellerOperationCapabilityStatus{
		Operation:  normalizeSellerOperation(operation),
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

func normalizeSellerAction(action string) string {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "sync", "read", "refresh", "preview":
		return "sync"
	case "reply", "send_message", "message_reply":
		return "reply"
	case "mark_read", "dismiss_notification", "acknowledge":
		return "acknowledge"
	case "fulfill", "fulfil", "ship", "update_fulfillment", "update_fulfilment":
		return "fulfill"
	case "send_offer", "counter_offer", "accept_offer", "decline_offer":
		return "offer"
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

package ebay

import "strings"

const (
	SellerListingCapabilityUnsupported  = "unsupported"
	SellerListingCapabilityDraftOnly    = "draft_only"
	SellerListingCapabilityConfirmedAPI = "confirmed_api"
)

const (
	SellerListingCommandDraft   = "draft"
	SellerListingCommandPublish = "publish"
	SellerListingCommandRevise  = "revise"
	SellerListingCommandEnd     = "end"
	SellerListingCommandRelist  = "relist"
)

type SellerListingLifecycleCommandRequest struct {
	Command    string `json:"command"`
	Capability string `json:"capability"`
	Confirmed  bool   `json:"confirmed"`
	ItemID     string `json:"item_id,omitempty"`
	DraftID    string `json:"draft_id,omitempty"`
	ListingID  string `json:"listing_id,omitempty"`
	Title      string `json:"title,omitempty"`
}

type SellerListingLifecycleCommandPreview struct {
	Command              string `json:"command"`
	Capability           string `json:"capability"`
	Confirmed            bool   `json:"confirmed"`
	ItemID               string `json:"item_id,omitempty"`
	DraftID              string `json:"draft_id,omitempty"`
	ListingID            string `json:"listing_id,omitempty"`
	Allowed              bool   `json:"allowed"`
	LocalOnly            bool   `json:"local_only"`
	RemoteWrite          bool   `json:"remote_write"`
	ConfirmationRequired bool   `json:"confirmation_required"`
	Blocker              string `json:"blocker,omitempty"`
}

type SellerListingLifecycleCommandExecution struct {
	SellerListingLifecycleCommandPreview
	Executed bool                                   `json:"executed"`
	Status   string                                 `json:"status"`
	Response *SellerListingLifecycleCommandResponse `json:"response,omitempty"`
}

type SellerListingLifecycleCommandResponse struct {
	Provider  string `json:"provider"`
	Command   string `json:"command"`
	DraftID   string `json:"draft_id,omitempty"`
	ListingID string `json:"listing_id,omitempty"`
	Status    string `json:"status"`
}

type SellerListingLifecycleClient interface {
	ExecuteSellerListingLifecycleCommand(SellerListingLifecycleCommandRequest) (SellerListingLifecycleCommandResponse, error)
}

func PreviewSellerListingLifecycleCommand(req SellerListingLifecycleCommandRequest) SellerListingLifecycleCommandPreview {
	command := normalizeSellerListingCommand(req.Command)
	capability := normalizeSellerListingCapability(req.Capability)
	preview := SellerListingLifecycleCommandPreview{
		Command:    command,
		Capability: capability,
		Confirmed:  req.Confirmed,
		ItemID:     strings.TrimSpace(req.ItemID),
		DraftID:    strings.TrimSpace(req.DraftID),
		ListingID:  strings.TrimSpace(req.ListingID),
	}
	if command == "" {
		preview.Blocker = "ebay_listing_lifecycle_command_required"
		return preview
	}
	if capability == SellerListingCapabilityUnsupported {
		preview.Blocker = "ebay_listing_capability_not_verified"
		return preview
	}
	if command == SellerListingCommandDraft {
		if preview.ItemID == "" || strings.TrimSpace(req.Title) == "" {
			preview.Blocker = "ebay_listing_draft_source_required"
			return preview
		}
		preview.Allowed = true
		preview.LocalOnly = true
		return preview
	}
	preview.ConfirmationRequired = true
	if capability != SellerListingCapabilityConfirmedAPI {
		preview.Blocker = "ebay_listing_write_capability_not_verified"
		return preview
	}
	preview.RemoteWrite = true
	if !req.Confirmed {
		preview.Blocker = "ebay_listing_lifecycle_confirmation_required"
		return preview
	}
	if command == SellerListingCommandPublish && preview.DraftID == "" {
		preview.Blocker = "ebay_listing_draft_required"
		return preview
	}
	if command != SellerListingCommandPublish && preview.ListingID == "" {
		preview.Blocker = "ebay_listing_id_required"
		return preview
	}
	preview.Allowed = true
	return preview
}

func ExecuteSellerListingLifecycleCommand(req SellerListingLifecycleCommandRequest, client SellerListingLifecycleClient) SellerListingLifecycleCommandExecution {
	preview := PreviewSellerListingLifecycleCommand(req)
	execution := SellerListingLifecycleCommandExecution{SellerListingLifecycleCommandPreview: preview}
	if preview.Command == "" {
		execution.Status = "invalid"
		return execution
	}
	if !preview.Allowed {
		execution.Status = "blocked"
		return execution
	}
	if preview.LocalOnly {
		execution.Executed = true
		execution.Status = "local_draft_ready"
		response := SellerListingLifecycleCommandResponse{
			Provider: "cabinet",
			Command:  preview.Command,
			DraftID:  firstNonEmpty(preview.DraftID, "draft-local-"+preview.ItemID),
			Status:   "local_draft_ready",
		}
		execution.Response = &response
		return execution
	}
	if client == nil {
		execution.Allowed = false
		execution.Blocker = "ebay_listing_lifecycle_adapter_required"
		execution.Status = "blocked"
		return execution
	}
	response, err := client.ExecuteSellerListingLifecycleCommand(req)
	if err != nil {
		execution.Allowed = false
		execution.Blocker = "ebay_listing_lifecycle_mock_failed"
		execution.Status = "blocked"
		return execution
	}
	execution.Executed = true
	execution.Status = firstNonEmpty(response.Status, "mocked_ebay_response")
	response.Provider = firstNonEmpty(response.Provider, "ebay")
	response.Command = preview.Command
	execution.Response = &response
	return execution
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func normalizeSellerListingCommand(command string) string {
	switch strings.ToLower(strings.TrimSpace(command)) {
	case SellerListingCommandDraft, "create_draft", "draft_listing":
		return SellerListingCommandDraft
	case SellerListingCommandPublish, "publish_listing":
		return SellerListingCommandPublish
	case SellerListingCommandRevise, "revise_listing", "update":
		return SellerListingCommandRevise
	case SellerListingCommandEnd, "end_listing":
		return SellerListingCommandEnd
	case SellerListingCommandRelist, "relist_listing":
		return SellerListingCommandRelist
	default:
		return ""
	}
}

func normalizeSellerListingCapability(capability string) string {
	switch strings.ToLower(strings.TrimSpace(capability)) {
	case SellerListingCapabilityDraftOnly, "local_draft":
		return SellerListingCapabilityDraftOnly
	case SellerListingCapabilityConfirmedAPI, "confirmed_write", "write_confirmed":
		return SellerListingCapabilityConfirmedAPI
	default:
		return SellerListingCapabilityUnsupported
	}
}

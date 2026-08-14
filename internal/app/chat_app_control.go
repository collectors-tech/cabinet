package app

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/collectors-tech/cabinet/internal/chat"
)

type chatAppControlIntent struct {
	CapabilityID      string
	Action            string
	Payload           map[string]any
	Route             string
	ConfirmationState string
	Message           string
	ErrorCode         string
	ErrorMessage      string
	SetupNeeded       bool
	GuidedWorkflow    map[string]any
}

func publicChatTrustedEvidenceKey(value any) (string, bool) {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			normalized := strings.ToLower(strings.TrimSpace(key))
			switch normalized {
			case "agent_planner",
				"agent_capabilities",
				"assistant_response",
				"assistant_handoff",
				"action_preview",
				"preview",
				"execution",
				"authority",
				"admin_session",
				"success_evidence":
				return key, true
			}
			if nestedKey, found := publicChatTrustedEvidenceKey(nested); found {
				return nestedKey, true
			}
		}
	case []any:
		for _, nested := range typed {
			if nestedKey, found := publicChatTrustedEvidenceKey(nested); found {
				return nestedKey, true
			}
		}
	}
	return "", false
}

func dispatchChatMessageAppControl(ctx context.Context, chatSvc *chat.Service, profileID, threadID, content string, envelope map[string]any, sourceMessageID string) (map[string]any, bool) {
	intent, ok := planChatMessageAppControl(content, envelope)
	if !ok {
		return nil, false
	}
	result := map[string]any{
		"capability_id": intent.CapabilityID,
		"policy":        "preview-before-apply",
	}
	if intent.Route != "" {
		result["route"] = intent.Route
	}
	if intent.SetupNeeded {
		result["setup_needed"] = true
	}
	if intent.GuidedWorkflow != nil {
		result["guided_workflow"] = intent.GuidedWorkflow
	}
	if intent.ErrorCode != "" {
		result["error"] = map[string]any{"code": intent.ErrorCode, "message": intent.ErrorMessage}
	}

	workflowInput := map[string]any{"content": content, "route": envelope["route"], "selection": envelope["selection"]}
	if contextEvidence := agentContextEvidence(envelope); len(contextEvidence) > 0 {
		workflowInput["agent_context"] = contextEvidence
	}
	run, runErr := chatSvc.CreateWorkflowRun(ctx, chat.CreateWorkflowRunInput{
		ProfileID:         profileID,
		WorkflowID:        "chat.app_control.dispatch",
		CapabilityID:      intent.CapabilityID,
		SourceChannel:     "in_app_chat",
		SourceThreadID:    threadID,
		SourceMessageID:   sourceMessageID,
		Input:             workflowInput,
		ProviderTrace:     map[string]any{"mode": "deterministic_app_control_planner", "live_provider": false},
		ConfirmationState: intent.ConfirmationState,
	})
	if runErr == nil {
		result["workflow_run"] = run
	}

	if intent.Action != "" {
		preview, previewErr := chatSvc.PreviewAction(ctx, chat.PreviewActionInput{
			ProfileID: profileID,
			ThreadID:  threadID,
			Action:    intent.Action,
			Payload:   intent.Payload,
		})
		if previewErr != nil {
			result["error"] = map[string]any{"code": "app_control_preview_failed", "message": previewErr.Error()}
			intent.Message = "I found the app-control request, but Cabinet could not create a safe preview."
		} else {
			result["preview"] = preview
			if runErr == nil {
				updated, updateErr := chatSvc.UpdateWorkflowRun(ctx, chat.UpdateWorkflowRunInput{
					ProfileID:         profileID,
					RunID:             run.ID,
					Status:            "completed",
					ProviderTrace:     map[string]any{"mode": "deterministic_app_control_planner", "live_provider": false},
					Result:            map[string]any{"preview_id": preview.ID, "action": preview.Action, "confirmation_required": true},
					ConfirmationState: "pending",
				})
				if updateErr == nil {
					result["workflow_run"] = updated
				}
			}
		}
	} else if runErr == nil {
		status := "completed"
		var runError map[string]any
		if intent.ErrorCode != "" {
			status = "failed"
			runError = map[string]any{"code": intent.ErrorCode, "message": intent.ErrorMessage}
		}
		updated, updateErr := chatSvc.UpdateWorkflowRun(ctx, chat.UpdateWorkflowRunInput{
			ProfileID:         profileID,
			RunID:             run.ID,
			Status:            status,
			ProviderTrace:     map[string]any{"mode": "deterministic_app_control_planner", "live_provider": false},
			Result:            map[string]any{"route": intent.Route, "setup_needed": intent.SetupNeeded, "guided_workflow": intent.GuidedWorkflow},
			Error:             runError,
			ConfirmationState: intent.ConfirmationState,
		})
		if updateErr == nil {
			result["workflow_run"] = updated
		}
	}

	state := chat.AgentResponseReadResult
	if intent.SetupNeeded {
		state = chat.AgentResponseSetupRequired
	} else if intent.ErrorCode != "" {
		state = chat.AgentResponseUnsupported
	} else if preview, _ := result["preview"].(chat.ActionPreview); strings.TrimSpace(preview.ID) != "" {
		state = chat.AgentResponsePreviewRequired
	}
	agentCtx := agentContextEvidence(envelope)
	agentResponse, _ := chat.NewAgentResponse(
		state,
		intent.Message,
		content,
		intent.CapabilityID,
		intent.CapabilityID,
		strings.TrimSpace(fmt.Sprint(agentCtx["surface_id"])),
		strings.TrimSpace(fmt.Sprint(agentCtx["source_channel"])),
	)
	if preview, _ := result["preview"].(chat.ActionPreview); strings.TrimSpace(preview.ID) != "" {
		agentResponse.Preview = &chat.AgentResponsePreview{
			ID: preview.ID, Action: preview.Action, Status: preview.Status, Payload: preview.Payload,
		}
	}
	if route := strings.TrimSpace(intent.Route); route != "" {
		agentResponse.NextAction = &chat.AgentResponseAction{Kind: "open_route", Label: labelForAgentResponseRoute(route), Route: route}
	}
	assistantMessage, assistantErr := chatSvc.CreateMessage(ctx, profileID, threadID, "assistant", agentResponse.Message, map[string]any{
		"app_control":    result,
		"agent_response": agentResponse,
	})
	if assistantErr == nil {
		result["thread_message"] = assistantMessage
	}
	return result, true
}

func labelForAgentResponseRoute(route string) string {
	normalized := strings.Trim(strings.TrimSpace(route), "/")
	if normalized == "" {
		return "Open Cabinet"
	}
	parts := strings.Split(normalized, "/")
	for index, part := range parts {
		if part == "" {
			continue
		}
		parts[index] = strings.ToUpper(part[:1]) + part[1:]
	}
	return "Open " + strings.Join(parts, " / ")
}

func planChatMessageAppControl(content string, envelope map[string]any) (chatAppControlIntent, bool) {
	normalized := normalizePlannerText(content)
	if normalized == "" {
		return chatAppControlIntent{}, false
	}
	if requestsGuidedInventoryUpdate(normalized) {
		recipe := chat.GuidedWorkflowRegistry()[0]
		return chatAppControlIntent{
			CapabilityID:      "guided.inventory.item.update",
			Route:             "/inventory",
			ConfirmationState: "not_required",
			Message:           "I can show the Inventory update walkthrough. I will open Inventory, highlight registered targets, and stop before any save or apply action.",
			GuidedWorkflow: map[string]any{
				"recipe_id":           recipe.ID,
				"title":               recipe.Title,
				"mode":                string(chat.GuidedWorkflowModeShowMe),
				"route":               "/inventory",
				"steps":               recipe.Steps,
				"mutation_boundary":   "show_me_never_mutates",
				"completion_criteria": "Inventory route and registered update targets were highlighted without applying changes.",
			},
		}, true
	}
	if route, label, ok := plannedOpenSurface(normalized); ok {
		return chatAppControlIntent{
			CapabilityID:      "navigate.open_surface",
			Route:             route,
			ConfirmationState: "not_required",
			Message:           fmt.Sprintf("I can open %s from this thread without changing Cabinet data.", label),
		}, true
	}
	if requestsOpenSurfaceRejection(normalized) {
		return chatAppControlIntent{
			CapabilityID:      "navigate.open_surface",
			ConfirmationState: "not_required",
			Message:           "I can only open known Cabinet surfaces from chat. Choose a listed Cabinet surface such as Dashboard, Inventory, Media, Integrations, Chats, Inbox, or Settings.",
			ErrorCode:         "unknown_surface",
			ErrorMessage:      "Unknown or unsafe Cabinet surface target.",
		}, true
	}
	if partNumber, title, ok := plannedCreateInventoryItem(content); ok {
		return chatAppControlIntent{
			CapabilityID:      "inventory.item.create",
			Action:            "create_inventory_item",
			Payload:           map[string]any{"part_number": partNumber, "title": title, "brand": "Unknown", "category": "General", "source": "chat_app_control"},
			ConfirmationState: "pending",
			Message:           fmt.Sprintf("I prepared a preview to create %s. Confirm before Cabinet saves anything.", partNumber),
		}, true
	}
	if title, ok := plannedRenameTitle(content); ok {
		itemID := selectedItemID(envelope)
		if itemID == "" {
			return chatAppControlIntent{
				CapabilityID:      "update_open_item_title",
				ConfirmationState: "not_required",
				Message:           "I can rename an item after you open or select the target inventory item.",
				SetupNeeded:       true,
			}, true
		}
		return chatAppControlIntent{
			CapabilityID: "update_open_item_title",
			Action:       "update_open_item_title",
			Payload: map[string]any{
				"item_id":            itemID,
				"title":              title,
				"field":              "title",
				"source_route":       routePath(envelope),
				"guided_workflow_id": "inventory.item.update",
				"guided_mode":        string(chat.GuidedWorkflowModeWithMe),
			},
			ConfirmationState: "pending",
			Message:           "I prepared a title-change preview for the open item. Confirm before Cabinet applies it.",
		}, true
	}
	if mentionsProviderBackedCapability(normalized) {
		return chatAppControlIntent{
			CapabilityID:      "content_generate",
			ConfirmationState: "not_required",
			Message:           "That assistant action needs provider setup before Cabinet can run it.",
			SetupNeeded:       true,
		}, true
	}
	return chatAppControlIntent{}, false
}

func agentContextEvidence(envelope map[string]any) map[string]any {
	rawAgentContext, _ := envelope["agent_context"].(map[string]any)
	if len(rawAgentContext) == 0 {
		return map[string]any{}
	}
	out := map[string]any{}
	for _, key := range []string{
		"profile_id",
		"workspace_id",
		"route_id",
		"surface_id",
		"thread_id",
		"intent_text",
		"source_channel",
		"permission_state",
		"setup_state",
		"workflow_run_id",
		"audit_id",
	} {
		if value := strings.TrimSpace(fmt.Sprint(rawAgentContext[key])); value != "" && value != "<nil>" {
			out[key] = value
		}
	}
	if selected, _ := rawAgentContext["selected_record"].(map[string]any); len(selected) > 0 {
		selectedEvidence := map[string]any{}
		for _, key := range []string{"type", "id"} {
			if value := strings.TrimSpace(fmt.Sprint(selected[key])); value != "" && value != "<nil>" {
				selectedEvidence[key] = value
			}
		}
		if len(selectedEvidence) > 0 {
			out["selected_record"] = selectedEvidence
		}
	}
	if selected, _ := rawAgentContext["selected_notification"].(map[string]any); len(selected) > 0 {
		selectedEvidence := map[string]any{}
		for _, key := range []string{"id", "source"} {
			if value := strings.TrimSpace(fmt.Sprint(selected[key])); value != "" && value != "<nil>" {
				selectedEvidence[key] = value
			}
		}
		if len(selectedEvidence) > 0 {
			out["selected_notification"] = selectedEvidence
		}
	}
	if adminSession := strings.ToLower(strings.TrimSpace(fmt.Sprint(rawAgentContext["admin_session"]))); adminSession == "authorized" || adminSession == "verified" {
		out["admin_session"] = "authorized"
	}
	if values := stringSliceEvidence(rawAgentContext["media_ids"]); len(values) > 0 {
		out["media_ids"] = values
	}
	if values := stringSliceEvidence(rawAgentContext["attachment_ids"]); len(values) > 0 {
		out["attachment_ids"] = values
	}
	return out
}

func stringSliceEvidence(raw any) []string {
	rawSlice, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(rawSlice))
	for _, entry := range rawSlice {
		if value := strings.TrimSpace(fmt.Sprint(entry)); value != "" && value != "<nil>" {
			out = append(out, value)
		}
	}
	return out
}

func requestsGuidedInventoryUpdate(normalized string) bool {
	return (strings.Contains(normalized, "show me") ||
		strings.Contains(normalized, "walk me through") ||
		strings.Contains(normalized, "guide me") ||
		strings.Contains(normalized, "how to")) &&
		(strings.Contains(normalized, "update an item") ||
			strings.Contains(normalized, "update item") ||
			strings.Contains(normalized, "edit an item") ||
			strings.Contains(normalized, "edit inventory"))
}

func chatMessageRequiresAssistantHandoff(content string) bool {
	normalized := normalizePlannerText(content)
	if normalized == "" {
		return false
	}
	handoffPhrases := []string{
		"follow up",
		"handoff",
		"queue",
		"background",
		"review this",
		"remind me",
		"notify me",
		"when ready",
	}
	for _, phrase := range handoffPhrases {
		if strings.Contains(normalized, phrase) {
			return true
		}
	}
	return false
}

func directAssistantChatResponse(content string) string {
	normalized := normalizePlannerText(content)
	switch {
	case normalized == "hello" || normalized == "hi" || strings.Contains(normalized, "hello cabinet"):
		return "I can help with Cabinet inventory, media, integrations, purchases, settings, and guided actions from this chat."
	default:
		return "I can help with Cabinet inventory, media, integrations, purchases, settings, and guided actions from this chat. Ask me to open a surface, prepare a safe preview, or queue a handoff when background follow-up is needed."
	}
}

func normalizePlannerText(content string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(content))), " ")
}

func plannedOpenSurface(normalized string) (string, string, bool) {
	if !requestsOpenSurface(normalized) {
		return "", "", false
	}
	for _, target := range assistantOpenSurfaceTargets() {
		for _, alias := range target.Aliases {
			if strings.Contains(normalized, alias) {
				return target.Route, target.Label, true
			}
		}
	}
	return "", "", false
}

func requestsOpenSurface(normalized string) bool {
	return strings.Contains(normalized, "open ") ||
		strings.Contains(normalized, "go to ") ||
		strings.Contains(normalized, "navigate to ") ||
		strings.Contains(normalized, "show ")
}

func requestsOpenSurfaceRejection(normalized string) bool {
	return strings.Contains(normalized, "open ") ||
		strings.Contains(normalized, "go to ") ||
		strings.Contains(normalized, "navigate to ")
}

func assistantOpenSurfaceTargets() []assistantCapabilityTarget {
	return []assistantCapabilityTarget{
		{ID: "dashboard", Label: "Dashboard", Route: "/dashboard", Aliases: []string{"dashboard", "home"}},
		{ID: "inventory", Label: "Inventory", Route: "/inventory", Aliases: []string{"inventory", "items"}},
		{ID: "wishlist", Label: "Wishlist", Route: "/wishlist", Aliases: []string{"wishlist", "wish list", "watchlist", "watch list", "saved views"}},
		{ID: "collections", Label: "Collections", Route: "/collections", Aliases: []string{"collections"}},
		{ID: "media", Label: "Media", Route: "/media", Aliases: []string{"media", "photos", "images"}},
		{ID: "discoveries", Label: "Discoveries", Route: "/discoveries", Aliases: []string{"discoveries", "discover"}},
		{ID: "scanner", Label: "Scanner", Route: "/scanner", Aliases: []string{"scanner", "manual capture", "capture"}},
		{ID: "market_watch", Label: "Market Watch", Route: "/scanner", Aliases: []string{"market watch", "market scanner"}},
		{ID: "purchases", Label: "Purchases", Route: "/purchases", Aliases: []string{"purchases", "orders"}},
		{ID: "integrations", Label: "Integrations", Route: "/integrations", Aliases: []string{"integrations", "apps", "providers"}},
		{ID: "chats", Label: "Chats", Route: "/chats", Aliases: []string{"chats", "chat", "assistant"}},
		{ID: "inbox", Label: "Inbox", Route: "/inbox", Aliases: []string{"inbox", "notifications"}},
		{ID: "settings_profile", Label: "Settings / Profile", Route: "/settings/profile", Aliases: []string{"settings profile", "profile settings"}},
		{ID: "settings_account", Label: "Settings / Account", Route: "/settings/account", Aliases: []string{"settings account", "account settings"}},
		{ID: "settings_appearance", Label: "Settings / Appearance", Route: "/settings/appearance", Aliases: []string{"settings appearance", "appearance settings", "theme settings"}},
		{ID: "settings_storage", Label: "Settings / Storage", Route: "/settings/storage", Aliases: []string{"settings storage", "storage settings"}},
		{ID: "settings", Label: "Settings", Route: "/settings", Aliases: []string{"settings"}},
	}
}

func plannedCreateInventoryItem(content string) (string, string, bool) {
	normalized := normalizePlannerText(content)
	if !strings.Contains(normalized, "create") || (!strings.Contains(normalized, "inventory item") && !strings.Contains(normalized, "item")) {
		return "", "", false
	}
	parts := strings.Fields(strings.TrimSpace(content))
	for i, part := range parts {
		clean := strings.Trim(part, ".,:;")
		if looksLikePartNumber(clean) && i+1 < len(parts) {
			title := strings.TrimSpace(strings.Join(parts[i+1:], " "))
			title = strings.Trim(title, ".,")
			if title != "" {
				return clean, title, true
			}
		}
	}
	return "", "", false
}

func plannedRenameTitle(content string) (string, bool) {
	normalized := normalizePlannerText(content)
	if !strings.Contains(normalized, "rename") && !strings.Contains(normalized, "title") {
		return "", false
	}
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\brename\b.+?\bto\s+(.+)$`),
		regexp.MustCompile(`(?i)\btitle\b.+?\bto\s+(.+)$`),
	}
	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(strings.TrimSpace(content))
		if len(matches) == 2 {
			title := strings.Trim(strings.TrimSpace(matches[1]), ` "'.,`)
			if title != "" {
				return title, true
			}
		}
	}
	return "", false
}

func mentionsProviderBackedCapability(normalized string) bool {
	return strings.Contains(normalized, "generate listing") ||
		strings.Contains(normalized, "catalog content") ||
		strings.Contains(normalized, "analyze image") ||
		strings.Contains(normalized, "process image")
}

func looksLikePartNumber(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) < 2 {
		return false
	}
	hasDigit := false
	hasLetter := false
	for _, r := range value {
		switch {
		case r >= '0' && r <= '9':
			hasDigit = true
		case r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z':
			hasLetter = true
		case r == '-' || r == '_' || r == '.':
		default:
			return false
		}
	}
	return hasDigit && hasLetter
}

func selectedItemID(envelope map[string]any) string {
	if id := stringFromNestedMap(envelope, "selection", "item_id"); id != "" {
		return id
	}
	if id := stringFromNestedMap(envelope, "selection", "active_item_id"); id != "" {
		return id
	}
	pathname := routePath(envelope)
	if strings.HasPrefix(pathname, "/inventory/") {
		return strings.Trim(strings.TrimPrefix(pathname, "/inventory/"), "/")
	}
	if rawSearch := stringFromNestedMap(envelope, "route", "search"); rawSearch != "" {
		values, err := url.ParseQuery(strings.TrimPrefix(rawSearch, "?"))
		if err == nil {
			return strings.TrimSpace(values.Get("item"))
		}
	}
	return ""
}

func routePath(envelope map[string]any) string {
	if path := stringFromNestedMap(envelope, "route", "pathname"); path != "" {
		return path
	}
	return "/"
}

func stringFromNestedMap(envelope map[string]any, parent, key string) string {
	rawParent, _ := envelope[parent].(map[string]any)
	if rawParent == nil {
		return ""
	}
	rawValue, _ := rawParent[key].(string)
	return strings.TrimSpace(rawValue)
}

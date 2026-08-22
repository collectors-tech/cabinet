package app

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/collectors-tech/cabinet/internal/ai"
)

// e2eSyntheticAgentProvider is registered only when Cabinet's explicit E2E
// hooks are enabled. It exercises the production planner and dispatcher with
// deterministic structured output and cannot make a network request.
type e2eSyntheticAgentProvider struct{}

func (e2eSyntheticAgentProvider) Name() string { return "fake" }

func (e2eSyntheticAgentProvider) RunAssistantTurn(ctx context.Context, req ai.AssistantTurnRequest) (ai.AssistantTurnResponse, error) {
	if err := ctx.Err(); err != nil {
		return ai.AssistantTurnResponse{}, err
	}
	if req.Metadata["entry_point"] == "chat.direct_conversation" {
		return ai.AssistantTurnResponse{
			Provider: "fake",
			Model:    strings.TrimSpace(req.Model),
			Text:     "E2E direct provider response",
			Metadata: map[string]string{
				"network":       "disabled",
				"test_provider": "true",
				"live_provider": "false",
			},
		}, nil
	}
	text := ""
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if strings.EqualFold(req.Messages[i].Role, "user") {
			text = strings.TrimSpace(req.Messages[i].Content)
			break
		}
	}
	normalized := strings.ToLower(text)
	plan := map[string]any{
		"decision":   "select_skill",
		"skill_id":   "cabinet.inventory.search_items",
		"parameters": map[string]any{},
		"message":    "Here is what Cabinet found.",
	}
	if strings.Contains(normalized, "tg-e2e-2086") && (strings.Contains(normalized, "add") || strings.Contains(normalized, "create")) {
		plan = map[string]any{
			"decision": "select_skill",
			"skill_id": "cabinet.inventory.create_item",
			"parameters": map[string]any{
				"part_number": "TG-E2E-2086",
				"title":       "Telegram E2E Preview",
				"category":    "Slot Cars",
			},
			"message": "I prepared the item for review.",
		}
	} else if strings.Contains(normalized, "agent-2097-synthetic") && strings.Contains(normalized, "wishlist") {
		plan = map[string]any{
			"decision":   "select_skill",
			"skill_id":   "cabinet.wishlist.create_entry",
			"parameters": map[string]any{"part_number": "AGENT-2097-SYNTHETIC", "title": syntheticPlannerTitle(text), "category": "Slot Cars"},
			"message":    "I prepared the wishlist entry for governed review.",
		}
	} else if strings.Contains(normalized, "chat-wish-2334") && strings.Contains(normalized, "wishlist") && (strings.Contains(normalized, "find") || strings.Contains(normalized, "search")) {
		plan = map[string]any{
			"decision":   "select_skill",
			"skill_id":   "cabinet.wishlist.search_entries",
			"parameters": map[string]any{"query": "CHAT-WISH-2334"},
			"message":    "Here are the matching Cabinet wishlist entries.",
		}
	} else if strings.Contains(normalized, "agent-2089-synthetic") && strings.Contains(normalized, "remove user") {
		plan = map[string]any{
			"decision":   "select_skill",
			"skill_id":   "cabinet.users.remove_user",
			"parameters": map[string]any{"target_user": syntheticPlannerValue(text, " target ")},
			"message":    "I prepared the exact user removal for strong confirmation.",
		}
	} else if strings.Contains(normalized, "agent-2185-synthetic") && strings.Contains(normalized, "configure provider") {
		plan = map[string]any{
			"decision": "select_skill",
			"skill_id": "cabinet.integrations.configure_provider",
			"parameters": map[string]any{
				"provider_id":     "voglers",
				"setup_payload":   "Configure Voglers for its public catalogue and enable it for profile e2e-profile-001; do not use or request an API key or secret.",
				"provider_secret": "",
				"api_key":         nil,
			},
			"message": "I prepared the Voglers public catalogue configuration for governed review.",
		}
	}
	planJSON, err := json.Marshal(plan)
	if err != nil {
		return ai.AssistantTurnResponse{}, err
	}
	return ai.AssistantTurnResponse{
		Provider: "fake",
		Model:    "cabinet-e2e-planner",
		Text:     string(planJSON),
		Metadata: map[string]string{
			"network":       "disabled",
			"test_provider": "true",
			"live_provider": "false",
		},
	}, nil
}

func syntheticPlannerTitle(text string) string {
	lower := strings.ToLower(text)
	const marker = " title "
	if index := strings.LastIndex(lower, marker); index >= 0 {
		if title := strings.TrimSpace(text[index+len(marker):]); title != "" {
			return title
		}
	}
	return "Synthetic Agent title"
}

func syntheticPlannerValue(text, marker string) string {
	lower := strings.ToLower(text)
	if index := strings.LastIndex(lower, marker); index >= 0 {
		return strings.TrimSpace(text[index+len(marker):])
	}
	return ""
}

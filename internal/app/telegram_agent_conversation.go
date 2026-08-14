package app

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/collectors-tech/cabinet/internal/agentcontext"
	"github.com/collectors-tech/cabinet/internal/agentskills"
	"github.com/collectors-tech/cabinet/internal/ai"
	"github.com/collectors-tech/cabinet/internal/chat"
	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/collectors-tech/cabinet/internal/telegramcapture"
	"github.com/google/uuid"
)

var (
	errTelegramPrivateChatRequired = errors.New("telegram Agent requires a paired private chat")
	errTelegramProfileMismatch     = errors.New("telegram Agent profile mismatch")
	errTelegramLegacyGrammar       = errors.New("telegram Agent accepts natural language only")
	errTelegramPreviewScope        = errors.New("telegram Agent preview does not belong to this sender and chat")
)

type telegramAgentConversationService struct {
	conn       *sql.DB
	profiles   *profile.Repository
	chat       *chat.Service
	registry   func(string) agentskills.Registry
	providers  *ai.AssistantProviderRegistry
	deliveryMu sync.Mutex
}

func newTelegramAgentConversationService(conn *sql.DB, profiles *profile.Repository, chatSvc *chat.Service, registry func(string) agentskills.Registry, providers *ai.AssistantProviderRegistry) *telegramAgentConversationService {
	return &telegramAgentConversationService{conn: conn, profiles: profiles, chat: chatSvc, registry: registry, providers: providers}
}

func (s *telegramAgentConversationService) HandleText(ctx context.Context, authorizedProfileID string, req telegramAgentTextRequest) (map[string]any, error) {
	profileID := strings.TrimSpace(authorizedProfileID)
	if profileID == "" || s == nil || s.conn == nil || s.chat == nil || s.registry == nil || s.providers == nil {
		return nil, telegramcapture.ErrUnauthorizedSender
	}
	if requested := strings.TrimSpace(req.ProfileID); requested != "" && requested != profileID {
		return nil, errTelegramProfileMismatch
	}
	if err := validateTelegramPrivatePeer(req.ChatType, req.SenderID, req.ChatID); err != nil {
		return nil, err
	}
	text := strings.TrimSpace(req.Text)
	if text == "" {
		return nil, errTelegramAgentTextNeedsClarification
	}
	if telegramRequestUsesLegacyGrammar(req) {
		return nil, errTelegramLegacyGrammar
	}
	deliveryID := telegramTextDeliveryID(req)
	if deliveryID == "" {
		return nil, errTelegramAgentTextNeedsClarification
	}
	fingerprint := telegramDeliveryFingerprint(profileID, req.SenderID, req.ChatID, req.Text)
	s.deliveryMu.Lock()
	defer s.deliveryMu.Unlock()
	if replay, found, err := s.claimDelivery(ctx, profileID, req.SenderID, req.ChatID, "text", deliveryID, fingerprint); err != nil {
		return nil, err
	} else if found {
		return replay, nil
	}
	completed := false
	defer func() {
		if !completed {
			s.failDelivery(context.Background(), profileID, req.SenderID, req.ChatID, "text", deliveryID)
		}
	}()
	thread, err := s.stableThread(ctx, profileID, req.SenderID, req.ChatID)
	if err != nil {
		return nil, err
	}
	providerID, model := s.assistantSelection(ctx, profileID)
	agentInput := map[string]any{
		"profile_id":     profileID,
		"thread_id":      thread.ID,
		"workspace_id":   telegramPeerWorkspace(req.SenderID, req.ChatID),
		"route_id":       "/chats",
		"surface_id":     "telegram.agent.conversation",
		"source_channel": "telegram",
		"setup_state":    "ready",
		"audit_id":       strings.TrimSpace(req.MessageID),
	}
	messageContext := agentcontext.WithEnvelope(map[string]any{
		"assistant":       map[string]any{"provider": providerID, "model": model, "source_channel": "telegram", "surface_id": "telegram.agent.conversation"},
		"route":           map[string]any{"pathname": "/chats"},
		"source_metadata": boundedTelegramSourceMetadata(req.SourceMetadata),
	}, agentcontext.NormalizeInput{
		ProfileID:    profileID,
		ThreadID:     thread.ID,
		IntentText:   text,
		Context:      map[string]any{"assistant": map[string]any{"provider": providerID, "model": model}, "route": map[string]any{"pathname": "/chats"}},
		AgentContext: agentInput,
	})
	message, err := s.chat.CreateMessage(ctx, profileID, thread.ID, "user", text, messageContext)
	if err != nil {
		return nil, err
	}
	planner, handled := dispatchChatAgentProviderPlanner(ctx, s.conn, s.chat, s.providers, s.registry(profileID), profileID, thread.ID, text, messageContext, message.ID)
	if !handled {
		return nil, errTelegramAgentTextNeedsClarification
	}

	if telegramPlannerNeedsInAppApproval(planner) {
		reply := telegramInAppApprovalReply(profileID, thread.ID)
		result := map[string]any{
			"profile_id": profileID, "thread": thread, "message": message,
			"preview_id": "", "confirmation_state": "in_app_approval_required", "telegram_reply": reply,
		}
		if err := s.completeDelivery(ctx, profileID, req.SenderID, req.ChatID, "text", deliveryID, result); err != nil {
			return nil, err
		}
		completed = true
		return result, nil
	}
	previewID := telegramPlannerPreviewID(planner)
	if previewID != "" && !strings.HasPrefix(previewID, "asp_") {
		return nil, fmt.Errorf("telegram planner returned a non-durable preview")
	}
	reply := telegramPlannerReply(profileID, thread.ID, previewID, planner)
	state := "not_required"
	if previewID != "" {
		state = "preview_required"
	}
	result := map[string]any{
		"profile_id":         profileID,
		"thread":             thread,
		"message":            message,
		"preview_id":         previewID,
		"confirmation_state": state,
		"telegram_reply":     reply,
	}
	if err := s.completeDelivery(ctx, profileID, req.SenderID, req.ChatID, "text", deliveryID, result); err != nil {
		return nil, err
	}
	completed = true
	return result, nil
}

func (s *telegramAgentConversationService) HandleCallback(ctx context.Context, authorizedProfileID string, req telegramAgentTextCallbackRequest) (map[string]any, error) {
	profileID := strings.TrimSpace(authorizedProfileID)
	if profileID == "" || s == nil || s.conn == nil || s.chat == nil || s.profiles == nil || s.registry == nil {
		return nil, telegramcapture.ErrUnauthorizedSender
	}
	if err := validateTelegramPrivatePeer(req.ChatType, req.SenderID, req.ChatID); err != nil {
		return nil, err
	}
	previewID, decision, ok := parseOpaqueTelegramPreviewCallback(req.CallbackData)
	if !ok {
		return nil, errTelegramLegacyGrammar
	}
	deliveryID := telegramCallbackDeliveryID(req)
	if deliveryID == "" {
		return nil, errTelegramLegacyGrammar
	}
	fingerprint := telegramDeliveryFingerprint(profileID, req.SenderID, req.ChatID, req.CallbackData)
	s.deliveryMu.Lock()
	defer s.deliveryMu.Unlock()
	if replay, found, err := s.claimDelivery(ctx, profileID, req.SenderID, req.ChatID, "callback", deliveryID, fingerprint); err != nil {
		return nil, err
	} else if found {
		return replay, nil
	}
	completed := false
	defer func() {
		if !completed {
			s.failDelivery(context.Background(), profileID, req.SenderID, req.ChatID, "callback", deliveryID)
		}
	}()
	record, err := getDurableAgentSkillPreview(ctx, s.conn, profileID, previewID)
	if err != nil {
		return nil, err
	}
	thread, err := s.lookupThread(ctx, profileID, req.SenderID, req.ChatID)
	if err != nil || record.SourceChannel != "telegram" || record.SourceSurface != "telegram.agent.conversation" || record.SourceThreadID != thread.ID {
		return nil, errTelegramPreviewScope
	}
	if strings.HasPrefix(record.SkillID, "cabinet.users.") {
		return nil, errTelegramPreviewScope
	}
	if record.Status != "previewed" {
		if record.Status == "applied" || record.Status == "cancelled" {
			result := telegramTerminalCallbackResult(profileID, thread.ID, record, true)
			if err := s.completeDelivery(ctx, profileID, req.SenderID, req.ChatID, "callback", deliveryID, result); err != nil {
				return nil, err
			}
			completed = true
			return result, nil
		}
		return nil, agentSkillPreviewStatusError(record.Status)
	}

	request := agentskills.PreviewRequest{ProfileID: profileID, PreviewID: previewID, Confirm: decision == "apply"}
	var preview agentskills.PreviewResponse
	if decision == "apply" {
		preview, err = applyDurableAgentSkillPreview(ctx, s.conn, s.chat, s.profiles, nil, s.registry(profileID), request)
	} else {
		preview, err = cancelDurableAgentSkillPreviewResponse(ctx, s.conn, s.chat, s.registry(profileID), request)
	}
	if err != nil {
		return nil, err
	}
	record.Status = preview.PreviewStatus
	record.Result = preview.Target
	result := telegramTerminalCallbackResult(profileID, thread.ID, record, false)
	if err := s.completeDelivery(ctx, profileID, req.SenderID, req.ChatID, "callback", deliveryID, result); err != nil {
		return nil, err
	}
	completed = true
	return result, nil
}

func (s *telegramAgentConversationService) stableThread(ctx context.Context, profileID, senderID, chatID string) (chat.Thread, error) {
	if thread, err := s.lookupThread(ctx, profileID, senderID, chatID); err == nil {
		return thread, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return chat.Thread{}, err
	}
	metadataJSON, err := json.Marshal(map[string]any{
		"kind": "telegram_agent_conversation", "source_channel": "telegram", "source_surface": "telegram.agent.conversation",
	})
	if err != nil {
		return chat.Thread{}, err
	}
	threadID := uuid.NewSHA1(uuid.NameSpaceURL, []byte("cabinet:telegram-agent:"+profileID+":"+strings.TrimSpace(senderID)+":"+strings.TrimSpace(chatID))).String()
	if _, err = s.conn.ExecContext(ctx, `
		INSERT INTO chat_threads(id, profile_id, title, metadata_json)
		VALUES(?, ?, 'Telegram Cabinet Agent', ?)
		ON CONFLICT(id) DO NOTHING
	`, threadID, profileID, string(metadataJSON)); err != nil {
		return chat.Thread{}, err
	}
	_, err = s.conn.ExecContext(ctx, `
		INSERT INTO telegram_agent_threads(profile_id, sender_id, chat_id, thread_id, created_at, updated_at)
		VALUES(?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(profile_id, sender_id, chat_id) DO NOTHING
	`, profileID, strings.TrimSpace(senderID), strings.TrimSpace(chatID), threadID)
	if err != nil {
		return chat.Thread{}, err
	}
	thread, err := s.lookupThread(ctx, profileID, senderID, chatID)
	if err == nil && thread.ID != threadID {
		_, _ = s.conn.ExecContext(ctx, `DELETE FROM chat_threads WHERE id = ? AND NOT EXISTS (SELECT 1 FROM telegram_agent_threads WHERE thread_id = ?)`, threadID, threadID)
	}
	return thread, err
}

func (s *telegramAgentConversationService) lookupThread(ctx context.Context, profileID, senderID, chatID string) (chat.Thread, error) {
	var threadID string
	err := s.conn.QueryRowContext(ctx, `
		SELECT thread_id FROM telegram_agent_threads
		WHERE profile_id = ? AND sender_id = ? AND chat_id = ?
	`, strings.TrimSpace(profileID), strings.TrimSpace(senderID), strings.TrimSpace(chatID)).Scan(&threadID)
	if err != nil {
		return chat.Thread{}, err
	}
	return s.chat.GetThread(ctx, profileID, threadID)
}

func (s *telegramAgentConversationService) claimDelivery(ctx context.Context, profileID, senderID, chatID, kind, deliveryID, fingerprint string) (map[string]any, bool, error) {
	profileID, senderID, chatID = strings.TrimSpace(profileID), strings.TrimSpace(senderID), strings.TrimSpace(chatID)
	kind, deliveryID = strings.TrimSpace(kind), strings.TrimSpace(deliveryID)
	var status, storedFingerprint, responseJSON string
	err := s.conn.QueryRowContext(ctx, `
		SELECT status, request_fingerprint, response_json
		FROM telegram_agent_deliveries
		WHERE profile_id = ? AND sender_id = ? AND chat_id = ? AND delivery_kind = ? AND delivery_id = ?
	`, profileID, senderID, chatID, kind, deliveryID).Scan(&status, &storedFingerprint, &responseJSON)
	if err == nil {
		if storedFingerprint != fingerprint {
			return nil, false, errors.New("telegram delivery identity was reused with different content")
		}
		if status == "failed" {
			res, updateErr := s.conn.ExecContext(ctx, `
				UPDATE telegram_agent_deliveries
				SET status = 'processing', response_json = '{}', updated_at = CURRENT_TIMESTAMP
				WHERE profile_id = ? AND sender_id = ? AND chat_id = ? AND delivery_kind = ? AND delivery_id = ? AND status = 'failed'
			`, profileID, senderID, chatID, kind, deliveryID)
			if updateErr != nil {
				return nil, false, updateErr
			}
			if changed, _ := res.RowsAffected(); changed == 1 {
				return nil, false, nil
			}
		}
		if status != "completed" {
			return nil, false, errors.New("telegram delivery is not safely replayable")
		}
		var result map[string]any
		if err := json.Unmarshal([]byte(responseJSON), &result); err != nil {
			return nil, false, errors.New("telegram delivery replay record is invalid")
		}
		result["duplicate"] = true
		return result, true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, false, err
	}
	_, err = s.conn.ExecContext(ctx, `
		INSERT INTO telegram_agent_deliveries(profile_id, sender_id, chat_id, delivery_kind, delivery_id, request_fingerprint, status, response_json, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, 'processing', '{}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, profileID, senderID, chatID, kind, deliveryID, fingerprint)
	if err != nil {
		return nil, false, err
	}
	return nil, false, nil
}

func (s *telegramAgentConversationService) completeDelivery(ctx context.Context, profileID, senderID, chatID, kind, deliveryID string, result map[string]any) error {
	encoded, err := json.Marshal(result)
	if err != nil {
		return err
	}
	res, err := s.conn.ExecContext(ctx, `
		UPDATE telegram_agent_deliveries
		SET status = 'completed', response_json = ?, updated_at = CURRENT_TIMESTAMP
		WHERE profile_id = ? AND sender_id = ? AND chat_id = ? AND delivery_kind = ? AND delivery_id = ? AND status = 'processing'
	`, string(encoded), strings.TrimSpace(profileID), strings.TrimSpace(senderID), strings.TrimSpace(chatID), strings.TrimSpace(kind), strings.TrimSpace(deliveryID))
	if err != nil {
		return err
	}
	if changed, _ := res.RowsAffected(); changed != 1 {
		return errors.New("telegram delivery completion claim was lost")
	}
	return nil
}

func (s *telegramAgentConversationService) failDelivery(ctx context.Context, profileID, senderID, chatID, kind, deliveryID string) {
	_, _ = s.conn.ExecContext(ctx, `
		UPDATE telegram_agent_deliveries SET status = 'failed', updated_at = CURRENT_TIMESTAMP
		WHERE profile_id = ? AND sender_id = ? AND chat_id = ? AND delivery_kind = ? AND delivery_id = ? AND status = 'processing'
	`, strings.TrimSpace(profileID), strings.TrimSpace(senderID), strings.TrimSpace(chatID), strings.TrimSpace(kind), strings.TrimSpace(deliveryID))
}

func telegramTextDeliveryID(req telegramAgentTextRequest) string {
	if value, ok := req.SourceMetadata["update_id"]; ok {
		if updateID := strings.TrimSpace(fmt.Sprint(value)); updateID != "" && updateID != "<nil>" && updateID != "0" {
			return "update:" + updateID
		}
	}
	if messageID := strings.TrimSpace(req.MessageID); messageID != "" {
		return "message:" + messageID
	}
	return ""
}

func telegramCallbackDeliveryID(req telegramAgentTextCallbackRequest) string {
	if callbackID := strings.TrimSpace(req.CallbackQueryID); callbackID != "" {
		return "callback:" + callbackID
	}
	if messageID := strings.TrimSpace(req.MessageID); messageID != "" {
		return "message:" + messageID
	}
	return ""
}

func telegramDeliveryFingerprint(values ...string) string {
	digest := sha256.Sum256([]byte(strings.Join(values, "\x00")))
	return hex.EncodeToString(digest[:])
}

func (s *telegramAgentConversationService) assistantSelection(ctx context.Context, profileID string) (string, string) {
	if _, ok := s.providers.Provider("fake"); ok {
		return "fake", "fake-planner-model"
	}
	providerID, model := "openai", ""
	if s.profiles != nil {
		if settings, err := s.profiles.GetSettings(ctx, profileID); err == nil {
			if selected := strings.ToLower(strings.TrimSpace(settings["assistant_default_provider"])); selected != "" {
				providerID = selected
			}
			model = strings.TrimSpace(settings["assistant_default_model"])
		}
	}
	return providerID, model
}

func validateTelegramPrivatePeer(chatType, senderID, chatID string) error {
	senderID, chatID = strings.TrimSpace(senderID), strings.TrimSpace(chatID)
	if !strings.EqualFold(strings.TrimSpace(chatType), "private") || senderID == "" || chatID == "" || senderID != chatID {
		return errTelegramPrivateChatRequired
	}
	return nil
}

func telegramRequestUsesLegacyGrammar(req telegramAgentTextRequest) bool {
	text := strings.ToLower(strings.TrimSpace(req.Text))
	return strings.TrimSpace(req.SkillID) != "" || req.Confirm || len(req.Parameters) > 0 || text == "/agent" || strings.HasPrefix(text, "/agent ") || strings.HasPrefix(text, "agent:") || strings.HasPrefix(text, "cabinet.")
}

func telegramPeerWorkspace(senderID, chatID string) string {
	return "telegram-private:" + strings.TrimSpace(senderID) + ":" + strings.TrimSpace(chatID)
}

func boundedTelegramSourceMetadata(in map[string]any) map[string]any {
	out := map[string]any{}
	if updateID, ok := in["update_id"]; ok {
		out["update_id"] = updateID
	}
	return out
}

func telegramPlannerPreviewID(planner map[string]any) string {
	preview, _ := planner["preview_result"].(map[string]any)
	value, _ := preview["preview_id"].(string)
	return strings.TrimSpace(value)
}

func telegramPlannerNeedsInAppApproval(planner map[string]any) bool {
	skillID := strings.TrimSpace(fmt.Sprint(planner["skill_id"]))
	if strings.HasPrefix(skillID, "cabinet.users.") {
		return true
	}
	// A selected read-only status or maintenance skill is safe to answer on the
	// paired channel. Only unresolved elevated intent remains an in-app handoff.
	return (skillID == "" || skillID == "<nil>") && strings.TrimSpace(fmt.Sprint(planner["intent_domain"])) == "admin"
}

func telegramPlannerReply(profileID, threadID, previewID string, planner map[string]any) telegramcapture.TelegramReply {
	text := strings.TrimSpace(fmt.Sprint(planner["message"]))
	if text == "" || text == "<nil>" {
		if errPayload, _ := planner["error"].(map[string]any); errPayload != nil {
			text = strings.TrimSpace(fmt.Sprint(errPayload["message"]))
		}
	}
	if text == "" || text == "<nil>" {
		text = "Cabinet could not plan that request yet. Open Cabinet to review Agent setup and retry."
	}
	reply := telegramcapture.TelegramReply{Text: text, ConfirmationState: "not_required"}
	if previewID == "" {
		return reply
	}
	reply.Text = text + " Nothing changes until you confirm."
	reply.ConfirmationState = "preview_required"
	reply.Actions = []string{"apply", "cancel"}
	reply.ActionButtons = []telegramcapture.TelegramReplyButton{
		{Label: "Apply", Kind: "callback", Action: "apply", CallbackData: previewID + ":apply"},
		{Label: "Cancel", Kind: "callback", Action: "cancel", CallbackData: previewID + ":cancel"},
	}
	return reply
}

func telegramInAppApprovalReply(profileID, threadID string) telegramcapture.TelegramReply {
	reviewURL := telegramAgentReviewURL(profileID, threadID, "")
	return telegramcapture.TelegramReply{
		Text:      "This request needs an authenticated Cabinet admin session. Open Cabinet to review it; Telegram cannot grant or carry admin authority.",
		ReviewURL: reviewURL, ConfirmationState: "in_app_approval_required", Actions: []string{"open_cabinet_review"},
		ActionButtons: []telegramcapture.TelegramReplyButton{{Label: "Open Cabinet", Kind: "url", Action: "open_cabinet_review", URL: reviewURL}},
	}
}

func parseOpaqueTelegramPreviewCallback(data string) (string, string, bool) {
	previewID, decision, ok := strings.Cut(strings.TrimSpace(data), ":")
	decision = strings.ToLower(strings.TrimSpace(decision))
	if !ok || !strings.HasPrefix(previewID, "asp_") || len(previewID) != len("asp_")+32 || (decision != "apply" && decision != "cancel") {
		return "", "", false
	}
	for _, r := range strings.TrimPrefix(previewID, "asp_") {
		if !strings.ContainsRune("0123456789abcdef", r) {
			return "", "", false
		}
	}
	return previewID, decision, true
}

func telegramTerminalCallbackResult(profileID, threadID string, record durableAgentSkillPreview, duplicate bool) map[string]any {
	state := record.Status
	text := "Cabinet cancelled the pending change. Nothing was changed."
	if state == "applied" {
		text = "Cabinet applied the confirmed change once."
	}
	reviewURL := telegramAgentReviewURL(profileID, threadID, record.ID)
	return map[string]any{
		"ok": true, "profile_id": profileID, "thread_id": threadID, "preview_id": record.ID,
		"confirmation_state": state, "duplicate": duplicate,
		"telegram_reply": telegramcapture.TelegramReply{
			Text: text, ReviewURL: reviewURL, ConfirmationState: state, Actions: []string{"open_cabinet_review"},
			ActionButtons: []telegramcapture.TelegramReplyButton{{Label: "Open Cabinet", Kind: "url", Action: "open_cabinet_review", URL: reviewURL}},
		},
	}
}

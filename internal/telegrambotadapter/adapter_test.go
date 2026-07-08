package telegrambotadapter

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/telegramcapture"
)

type fakeCabinetGateway struct {
	path     string
	body     any
	response telegramcapture.TelegramReply
	err      error
}

func (f *fakeCabinetGateway) PostJSON(_ context.Context, path string, body any, response any) (int, error) {
	f.path = path
	f.body = body
	out := response.(*cabinetTelegramReplyResponse)
	out.TelegramReply = f.response
	if f.err != nil {
		return 500, f.err
	}
	return 200, nil
}

func TestCabinetRequestFromUpdateRoutesWebhookMessage(t *testing.T) {
	t.Parallel()

	req, err := CabinetRequestFromUpdate(Update{
		UpdateID: 7002,
		Message: &WebhookMessage{
			MessageID: 43,
			From:      WebhookUser{ID: 12345},
			Chat:      WebhookChat{ID: -5235769556},
			Caption:   "Front image and barcode 4904810900016",
			Photo: []telegramcapture.WebhookPhotoSize{
				{FileID: "small-photo", FileUniqueID: "small", Width: 90, Height: 90, FileSize: 1024},
				{FileID: "large-photo", FileUniqueID: "large", Width: 1280, Height: 720, FileSize: 2048},
			},
		},
	})
	if err != nil {
		t.Fatalf("CabinetRequestFromUpdate() error = %v", err)
	}
	if req.Path != WebhookCapturePath {
		t.Fatalf("expected webhook capture path, got %q", req.Path)
	}
	body, ok := req.Body.(telegramcapture.WebhookUpdate)
	if !ok {
		t.Fatalf("expected telegramcapture.WebhookUpdate body, got %T", req.Body)
	}
	if body.UpdateID != 7002 || body.Message == nil || body.Message.Chat.ID != -5235769556 {
		t.Fatalf("webhook body did not preserve Telegram update/message context: %+v", body)
	}
}

func TestCabinetRequestFromUpdateRoutesCallbackQuery(t *testing.T) {
	t.Parallel()

	req, err := CabinetRequestFromUpdate(Update{
		UpdateID: 8001,
		CallbackQuery: &CallbackQuery{
			ID:   "callback-1",
			From: WebhookUser{ID: 12345},
			Message: &CallbackMessage{
				MessageID: 44,
				Chat:      WebhookChat{ID: -5235769556},
			},
			Data: "cabinet:catalog_capture:confirm:preview-1",
		},
	})
	if err != nil {
		t.Fatalf("CabinetRequestFromUpdate() callback error = %v", err)
	}
	if req.Path != CaptureCallbackPath {
		t.Fatalf("expected callback path, got %q", req.Path)
	}
	body, ok := req.Body.(CaptureCallbackRequest)
	if !ok {
		t.Fatalf("expected CaptureCallbackRequest body, got %T", req.Body)
	}
	if body.SenderID != "12345" || body.ChatID != "-5235769556" || body.CallbackData != "cabinet:catalog_capture:confirm:preview-1" {
		t.Fatalf("callback body did not preserve sender/chat/callback data: %+v", body)
	}
}

func TestCabinetRequestFromUpdateRoutesAgentTextCommand(t *testing.T) {
	t.Parallel()

	req, err := CabinetRequestFromUpdate(Update{
		UpdateID: 9001,
		Message: &WebhookMessage{
			MessageID: 51,
			From:      WebhookUser{ID: 12345},
			Chat:      WebhookChat{ID: -5235769556},
			Text:      "/agent cabinet.inventory.search_items query=AFX",
		},
	})
	if err != nil {
		t.Fatalf("CabinetRequestFromUpdate() agent text error = %v", err)
	}
	if req.Path != AgentTextPath {
		t.Fatalf("expected agent text path, got %q", req.Path)
	}
	body, ok := req.Body.(AgentTextRequest)
	if !ok {
		t.Fatalf("expected AgentTextRequest body, got %T", req.Body)
	}
	if body.SenderID != "12345" || body.ChatID != "-5235769556" || body.MessageID != "51" {
		t.Fatalf("agent text body did not preserve Telegram source identifiers: %+v", body)
	}
	if body.Text != "cabinet.inventory.search_items query=AFX" || body.SkillID != "cabinet.inventory.search_items" || body.Parameters["query"] != "AFX" {
		t.Fatalf("agent text body did not preserve command text/skill parameters: %+v", body)
	}
	if body.SourceMetadata["update_id"] != int64(9001) || body.SourceMetadata["command"] != "agent" {
		t.Fatalf("agent text body did not include non-secret source metadata: %+v", body.SourceMetadata)
	}
}

func TestCabinetRequestFromUpdateRoutesAgentCaptionMediaCommand(t *testing.T) {
	t.Parallel()

	req, err := CabinetRequestFromUpdate(Update{
		UpdateID: 9003,
		Message: &WebhookMessage{
			MessageID: 53,
			From:      WebhookUser{ID: 12345},
			Chat:      WebhookChat{ID: -5235769556},
			Caption:   "/agent cabinet.inventory.attach_media item_id=item-1 media_id=telegram-photo-large",
			Photo: []telegramcapture.WebhookPhotoSize{
				{FileID: "telegram-photo-small", FileUniqueID: "photo-small", Width: 90, Height: 90, FileSize: 1024},
				{FileID: "telegram-photo-large", FileUniqueID: "photo-large", Width: 1280, Height: 720, FileSize: 4096},
			},
		},
	})
	if err != nil {
		t.Fatalf("CabinetRequestFromUpdate() agent media caption error = %v", err)
	}
	if req.Path != AgentTextPath {
		t.Fatalf("expected agent text path, got %q", req.Path)
	}
	body, ok := req.Body.(AgentTextRequest)
	if !ok {
		t.Fatalf("expected AgentTextRequest body, got %T", req.Body)
	}
	if body.Text != "cabinet.inventory.attach_media item_id=item-1 media_id=telegram-photo-large" || body.SkillID != "cabinet.inventory.attach_media" {
		t.Fatalf("agent caption body did not preserve command text/skill: %+v", body)
	}
	if body.Parameters["item_id"] != "item-1" || body.Parameters["media_id"] != "telegram-photo-large" {
		t.Fatalf("agent caption body did not preserve parameters: %+v", body.Parameters)
	}
	if len(body.Media) != 1 ||
		body.Media[0].FileID != "telegram-photo-large" ||
		body.Media[0].FileUniqueID != "photo-large" ||
		body.Media[0].FileSize != 4096 ||
		body.Media[0].Filename != "photo-large.jpg" ||
		body.Media[0].MIMEType != "image/jpeg" ||
		body.Media[0].Kind != "photo" {
		t.Fatalf("agent caption body did not preserve non-secret Telegram media evidence: %+v", body.Media)
	}
}

func TestCabinetRequestFromUpdateRoutesAgentTextCallbackQuery(t *testing.T) {
	t.Parallel()

	req, err := CabinetRequestFromUpdate(Update{
		UpdateID: 9002,
		CallbackQuery: &CallbackQuery{
			ID:   "agent-callback-1",
			From: WebhookUser{ID: 12345},
			Message: &CallbackMessage{
				MessageID: 52,
				Chat:      WebhookChat{ID: -5235769556},
			},
			Data: "cabinet:agent_text:confirm:preview-1",
		},
	})
	if err != nil {
		t.Fatalf("CabinetRequestFromUpdate() agent callback error = %v", err)
	}
	if req.Path != AgentTextCallbackPath {
		t.Fatalf("expected agent text callback path, got %q", req.Path)
	}
	body, ok := req.Body.(AgentTextCallbackRequest)
	if !ok {
		t.Fatalf("expected AgentTextCallbackRequest body, got %T", req.Body)
	}
	if body.SenderID != "12345" || body.ChatID != "-5235769556" || body.MessageID != "52" || body.PreviewID != "preview-1" || body.Confirmation != "confirm" {
		t.Fatalf("agent callback body did not preserve confirmation identifiers: %+v", body)
	}
	if body.CallbackData != "cabinet:agent_text:confirm:preview-1" {
		t.Fatalf("agent callback body did not preserve callback data: %+v", body)
	}
}

func TestSendMessageFromReplyRendersInlineButtons(t *testing.T) {
	t.Parallel()

	call, err := SendMessageFromReply("-5235769556", telegramcapture.TelegramReply{
		Text:              "Draft ready for review: Slot car.",
		ReviewURL:         "/chats?profile_id=profile-1&thread_id=thread-1&preview_id=preview-1",
		ConfirmationState: "preview_required",
		ActionButtons: []telegramcapture.TelegramReplyButton{
			{Label: "Open Cabinet review", Kind: "url", Action: "open_cabinet_review", URL: "/chats?profile_id=profile-1&thread_id=thread-1&preview_id=preview-1"},
			{Label: "Confirm in Cabinet", Kind: "callback", Action: "confirm_in_cabinet", CallbackData: "cabinet:catalog_capture:confirm:preview-1"},
			{Label: "Cancel in Cabinet", Kind: "callback", Action: "cancel_in_cabinet", CallbackData: "cabinet:catalog_capture:cancel:preview-1"},
		},
	})
	if err != nil {
		t.Fatalf("SendMessageFromReply() error = %v", err)
	}
	if call.Method != "sendMessage" {
		t.Fatalf("expected sendMessage, got %q", call.Method)
	}
	body := mustJSONMap(t, call.Body)
	if body["chat_id"] != "-5235769556" || body["text"] != "Draft ready for review: Slot car." {
		t.Fatalf("sendMessage payload did not preserve chat/text: %+v", body)
	}
	markup := body["reply_markup"].(map[string]any)
	rows := markup["inline_keyboard"].([]any)
	if len(rows) != 3 {
		t.Fatalf("expected 3 inline button rows, got %#v", rows)
	}
	confirmRow := rows[1].([]any)[0].(map[string]any)
	if confirmRow["callback_data"] != "cabinet:catalog_capture:confirm:preview-1" {
		t.Fatalf("confirm callback button was not rendered: %#v", confirmRow)
	}
}

func TestSendMessageFromReplyRendersFollowUpReplyKeyboard(t *testing.T) {
	t.Parallel()

	call, err := SendMessageFromReply("-5235769556", telegramcapture.TelegramReply{
		Text:              "I need one more detail before I can draft this safely.",
		ConfirmationState: "follow_up_required",
		ActionButtons: []telegramcapture.TelegramReplyButton{
			{Label: "Send barcode", Kind: "reply", Action: "reply_with_barcode"},
			{Label: "Send part number", Kind: "reply", Action: "reply_with_part_number"},
			{Label: "Send item title", Kind: "reply", Action: "reply_with_item_title"},
		},
	})
	if err != nil {
		t.Fatalf("SendMessageFromReply() follow-up error = %v", err)
	}
	body := mustJSONMap(t, call.Body)
	markup := body["reply_markup"].(map[string]any)
	rows := markup["keyboard"].([]any)
	if len(rows) != 3 || !markup["resize_keyboard"].(bool) || !markup["one_time_keyboard"].(bool) {
		t.Fatalf("follow-up reply keyboard was not rendered: %#v", markup)
	}
}

func TestEditMessageFromReplyRendersCallbackResult(t *testing.T) {
	t.Parallel()

	call, err := EditMessageFromReply("-5235769556", "44", telegramcapture.TelegramReply{
		Text:              "Confirmed. Cabinet added the catalog item.",
		ConfirmationState: "confirmed",
		ActionButtons: []telegramcapture.TelegramReplyButton{
			{Label: "Open Cabinet review", Kind: "url", Action: "open_cabinet_review", URL: "/chats?profile_id=profile-1&thread_id=thread-1&preview_id=preview-1"},
		},
	})
	if err != nil {
		t.Fatalf("EditMessageFromReply() error = %v", err)
	}
	if call.Method != "editMessageText" {
		t.Fatalf("expected editMessageText, got %q", call.Method)
	}
	body := mustJSONMap(t, call.Body)
	if body["message_id"] != "44" || body["text"] != "Confirmed. Cabinet added the catalog item." {
		t.Fatalf("editMessageText payload did not preserve target/text: %+v", body)
	}
	markup := body["reply_markup"].(map[string]any)
	rows := markup["inline_keyboard"].([]any)
	button := rows[0].([]any)[0].(map[string]any)
	if button["url"] != "/chats?profile_id=profile-1&thread_id=thread-1&preview_id=preview-1" {
		t.Fatalf("review URL button was not rendered: %#v", button)
	}
}

func TestAnswerCallbackQueryFromReplyRendersCallbackAcknowledgement(t *testing.T) {
	t.Parallel()

	call, err := AnswerCallbackQueryFromReply("callback-1", telegramcapture.TelegramReply{
		Text:              "Confirmed. Cabinet added the catalog item.",
		ConfirmationState: "confirmed",
	})
	if err != nil {
		t.Fatalf("AnswerCallbackQueryFromReply() error = %v", err)
	}
	if call.Method != "answerCallbackQuery" {
		t.Fatalf("expected answerCallbackQuery, got %q", call.Method)
	}
	body := mustJSONMap(t, call.Body)
	if body["callback_query_id"] != "callback-1" || body["text"] != "Confirmed. Cabinet added the catalog item." {
		t.Fatalf("answerCallbackQuery payload did not preserve callback/text: %+v", body)
	}
	if body["show_alert"].(bool) {
		t.Fatalf("callback acknowledgement should be non-alert by default: %+v", body)
	}
}

func TestReplyRenderingCapsTelegramTextLengths(t *testing.T) {
	t.Parallel()

	longRunes := strings.Repeat("A", telegramMessageTextLimit+25)
	send, err := SendMessageFromReply("-5235769556", telegramcapture.TelegramReply{Text: longRunes})
	if err != nil {
		t.Fatalf("SendMessageFromReply() long text error = %v", err)
	}
	sendBody := mustJSONMap(t, send.Body)
	sendText := sendBody["text"].(string)
	if len([]rune(sendText)) != telegramMessageTextLimit || !strings.HasSuffix(sendText, "...") {
		t.Fatalf("sendMessage text was not capped at Telegram limit: len=%d suffix=%q", len([]rune(sendText)), sendText[len(sendText)-3:])
	}

	callbackText := strings.Repeat("B", telegramCallbackTextLimit+25)
	ack, err := AnswerCallbackQueryFromReply("callback-1", telegramcapture.TelegramReply{Text: callbackText})
	if err != nil {
		t.Fatalf("AnswerCallbackQueryFromReply() long text error = %v", err)
	}
	ackBody := mustJSONMap(t, ack.Body)
	ackText := ackBody["text"].(string)
	if len([]rune(ackText)) != telegramCallbackTextLimit || !strings.HasSuffix(ackText, "...") {
		t.Fatalf("answerCallbackQuery text was not capped at Telegram limit: len=%d suffix=%q", len([]rune(ackText)), ackText[len(ackText)-3:])
	}
}

func TestReplyRenderingCapsButtonTextAndSkipsOversizedCallbackData(t *testing.T) {
	t.Parallel()

	longLabel := strings.Repeat("Confirm ", 20)
	call, err := SendMessageFromReply("-5235769556", telegramcapture.TelegramReply{
		Text: "Draft ready.",
		ActionButtons: []telegramcapture.TelegramReplyButton{
			{Label: longLabel, Kind: "url", URL: "http://127.0.0.1:17882/chats?thread_id=thread-1"},
			{Label: "Overlong callback", Kind: "callback", CallbackData: "cabinet:catalog_capture:confirm:" + strings.Repeat("preview", 20)},
			{Label: "Confirm", Kind: "callback", CallbackData: "cabinet:catalog_capture:confirm:preview-1"},
		},
	})
	if err != nil {
		t.Fatalf("SendMessageFromReply() error = %v", err)
	}
	body := mustJSONMap(t, call.Body)
	markup := body["reply_markup"].(map[string]any)
	rows := markup["inline_keyboard"].([]any)
	if len(rows) != 2 {
		t.Fatalf("expected URL row and valid callback row only, got %#v", rows)
	}
	urlButton := rows[0].([]any)[0].(map[string]any)
	if text := urlButton["text"].(string); len([]rune(text)) != telegramButtonTextLimit || !strings.HasSuffix(text, "...") {
		t.Fatalf("URL button text was not capped visibly: %q", text)
	}
	callbackButton := rows[1].([]any)[0].(map[string]any)
	if callbackButton["callback_data"] != "cabinet:catalog_capture:confirm:preview-1" {
		t.Fatalf("valid callback button was not preserved after skipping oversized callback data: %#v", callbackButton)
	}

	replyCall, err := SendMessageFromReply("-5235769556", telegramcapture.TelegramReply{
		Text: "Need one more detail.",
		ActionButtons: []telegramcapture.TelegramReplyButton{
			{Label: longLabel, Kind: "reply", Action: "reply_with_barcode"},
		},
	})
	if err != nil {
		t.Fatalf("SendMessageFromReply() reply keyboard error = %v", err)
	}
	replyBody := mustJSONMap(t, replyCall.Body)
	replyMarkup := replyBody["reply_markup"].(map[string]any)
	replyRows := replyMarkup["keyboard"].([]any)
	replyButton := replyRows[0].([]any)[0].(map[string]any)
	if text := replyButton["text"].(string); len([]rune(text)) != telegramButtonTextLimit || !strings.HasSuffix(text, "...") {
		t.Fatalf("reply keyboard button text was not capped visibly: %q", text)
	}
}

func TestDispatchUpdateSendsCaptureReplyThroughBotAPI(t *testing.T) {
	t.Parallel()

	gateway := &fakeCabinetGateway{response: telegramcapture.TelegramReply{
		Text:              "Draft ready for review: Slot car.",
		ConfirmationState: "preview_required",
		ActionButtons: []telegramcapture.TelegramReplyButton{
			{Label: "Confirm in Cabinet", Kind: "callback", CallbackData: "cabinet:catalog_capture:confirm:preview-1"},
		},
	}}
	result, err := DispatchUpdate(context.Background(), gateway, Update{
		UpdateID: 7002,
		Message: &WebhookMessage{
			MessageID: 43,
			From:      WebhookUser{ID: 12345},
			Chat:      WebhookChat{ID: -5235769556},
			Text:      "Barcode 4904810900016",
		},
	})
	if err != nil {
		t.Fatalf("DispatchUpdate() error = %v", err)
	}
	if result.CabinetPath != WebhookCapturePath || gateway.path != WebhookCapturePath {
		t.Fatalf("dispatch did not route message through webhook capture path: result=%+v gateway=%q", result, gateway.path)
	}
	if len(result.BotCalls) != 1 || result.BotCalls[0].Method != "sendMessage" {
		t.Fatalf("expected one sendMessage call, got %+v", result.BotCalls)
	}
	body := mustJSONMap(t, result.BotCalls[0].Body)
	if body["chat_id"] != "-5235769556" || body["text"] != "Draft ready for review: Slot car." {
		t.Fatalf("sendMessage call did not preserve chat/text: %+v", body)
	}
}

func TestDispatchUpdateAcknowledgesAndEditsCallbackReply(t *testing.T) {
	t.Parallel()

	gateway := &fakeCabinetGateway{response: telegramcapture.TelegramReply{
		Text:              "Confirmed. Cabinet added the catalog item.",
		ConfirmationState: "confirmed",
		ActionButtons: []telegramcapture.TelegramReplyButton{
			{Label: "Open Cabinet review", Kind: "url", URL: "/chats?preview_id=preview-1"},
		},
	}}
	result, err := DispatchUpdate(context.Background(), gateway, Update{
		UpdateID: 8001,
		CallbackQuery: &CallbackQuery{
			ID:   "callback-1",
			From: WebhookUser{ID: 12345},
			Message: &CallbackMessage{
				MessageID: 44,
				Chat:      WebhookChat{ID: -5235769556},
			},
			Data: "cabinet:catalog_capture:confirm:preview-1",
		},
	})
	if err != nil {
		t.Fatalf("DispatchUpdate() callback error = %v", err)
	}
	if result.CabinetPath != CaptureCallbackPath || gateway.path != CaptureCallbackPath {
		t.Fatalf("dispatch did not route callback through callback path: result=%+v gateway=%q", result, gateway.path)
	}
	if len(result.BotCalls) != 2 || result.BotCalls[0].Method != "answerCallbackQuery" || result.BotCalls[1].Method != "editMessageText" {
		t.Fatalf("expected answerCallbackQuery then editMessageText, got %+v", result.BotCalls)
	}
	ack := mustJSONMap(t, result.BotCalls[0].Body)
	if ack["callback_query_id"] != "callback-1" || ack["text"] != "Confirmed. Cabinet added the catalog item." {
		t.Fatalf("callback acknowledgement did not preserve id/text: %+v", ack)
	}
	edit := mustJSONMap(t, result.BotCalls[1].Body)
	if edit["chat_id"] != "-5235769556" || edit["message_id"] != "44" {
		t.Fatalf("editMessageText did not target callback message: %+v", edit)
	}
}

func TestDispatchUpdateSendsStructuredFailureReplyWhenCabinetRejectsCapture(t *testing.T) {
	t.Parallel()

	gateway := &fakeCabinetGateway{
		response: telegramcapture.TelegramReply{
			Text:              "I need a barcode, part number, or clearer item title before I can draft this safely.",
			ConfirmationState: "follow_up_required",
			ActionButtons: []telegramcapture.TelegramReplyButton{
				{Label: "Send barcode", Kind: "reply", Action: "reply_with_barcode"},
			},
		},
		err: errors.New("cabinet returned 422 telegram_capture_needs_follow_up"),
	}
	result, err := DispatchUpdate(context.Background(), gateway, Update{
		UpdateID: 7003,
		Message: &WebhookMessage{
			MessageID: 45,
			From:      WebhookUser{ID: 12345},
			Chat:      WebhookChat{ID: -5235769556},
			Text:      "blue boxed one from the bench",
		},
	})
	if err != nil {
		t.Fatalf("DispatchUpdate() structured failure error = %v", err)
	}
	if result.CabinetPath != WebhookCapturePath || result.CabinetError == "" {
		t.Fatalf("dispatch did not preserve Cabinet path/error: %+v", result)
	}
	if len(result.BotCalls) != 1 || result.BotCalls[0].Method != "sendMessage" {
		t.Fatalf("expected one failure sendMessage call, got %+v", result.BotCalls)
	}
	body := mustJSONMap(t, result.BotCalls[0].Body)
	if body["chat_id"] != "-5235769556" || body["text"] != "I need a barcode, part number, or clearer item title before I can draft this safely." {
		t.Fatalf("failure sendMessage did not preserve structured Telegram reply: %+v", body)
	}
	markup := body["reply_markup"].(map[string]any)
	rows := markup["keyboard"].([]any)
	if len(rows) != 1 {
		t.Fatalf("structured follow-up controls were not rendered on failure: %#v", markup)
	}
}

func TestDispatchUpdateAcknowledgesAndEditsFallbackFailureForCallback(t *testing.T) {
	t.Parallel()

	gateway := &fakeCabinetGateway{err: errors.New("cabinet callback endpoint unavailable")}
	result, err := DispatchUpdate(context.Background(), gateway, Update{
		UpdateID: 8002,
		CallbackQuery: &CallbackQuery{
			ID:   "callback-2",
			From: WebhookUser{ID: 12345},
			Message: &CallbackMessage{
				MessageID: 46,
				Chat:      WebhookChat{ID: -5235769556},
			},
			Data: "cabinet:catalog_capture:confirm:preview-2",
		},
	})
	if err != nil {
		t.Fatalf("DispatchUpdate() callback fallback error = %v", err)
	}
	if result.CabinetPath != CaptureCallbackPath || result.CabinetError == "" {
		t.Fatalf("dispatch did not preserve callback path/error: %+v", result)
	}
	if len(result.BotCalls) != 2 || result.BotCalls[0].Method != "answerCallbackQuery" || result.BotCalls[1].Method != "editMessageText" {
		t.Fatalf("expected callback failure ack and edit calls, got %+v", result.BotCalls)
	}
	ack := mustJSONMap(t, result.BotCalls[0].Body)
	if ack["callback_query_id"] != "callback-2" || ack["text"] == "" {
		t.Fatalf("fallback callback ack did not preserve visible failure text: %+v", ack)
	}
	edit := mustJSONMap(t, result.BotCalls[1].Body)
	if edit["chat_id"] != "-5235769556" || edit["message_id"] != "46" || edit["text"] == "" {
		t.Fatalf("fallback callback edit did not target the original message with failure text: %+v", edit)
	}
}

func TestDispatchUpdateToBotAPIPostsRenderedMessageCall(t *testing.T) {
	t.Parallel()

	gateway := &fakeCabinetGateway{response: telegramcapture.TelegramReply{
		Text:              "Draft ready for review: Slot car.",
		ConfirmationState: "preview_required",
	}}
	var sawSendMessage bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/botbot-token-1/sendMessage" {
			t.Fatalf("unexpected Bot API path: %s", r.URL.Path)
		}
		sawSendMessage = true
		if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected Bot API request method/header: method=%s content-type=%s", r.Method, r.Header.Get("Content-Type"))
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode sendMessage body: %v", err)
		}
		if body["chat_id"] != "-5235769556" || body["text"] != "Draft ready for review: Slot car." {
			t.Fatalf("sendMessage body did not preserve rendered reply: %+v", body)
		}
		_, _ = w.Write([]byte("{\"ok\":true}"))
	}))
	defer server.Close()

	result, err := DispatchUpdateToBotAPI(context.Background(), gateway, BotAPIEndpoint{BaseURL: server.URL, Token: "bot-token-1"}, server.Client(), Update{
		UpdateID: 7004,
		Message: &WebhookMessage{
			MessageID: 47,
			From:      WebhookUser{ID: 12345},
			Chat:      WebhookChat{ID: -5235769556},
			Text:      "Barcode 4904810900016",
		},
	})
	if err != nil {
		t.Fatalf("DispatchUpdateToBotAPI() error = %v", err)
	}
	if !sawSendMessage || result.CabinetPath != WebhookCapturePath || len(result.BotCalls) != 1 || len(result.BotAPIErrors) != 0 {
		t.Fatalf("dispatch did not execute rendered Bot API message call cleanly: saw=%v result=%+v", sawSendMessage, result)
	}
}

func TestDispatchUpdateToBotAPIPostsCallbackAckAndEditCalls(t *testing.T) {
	t.Parallel()

	gateway := &fakeCabinetGateway{response: telegramcapture.TelegramReply{
		Text:              "Confirmed. Cabinet added the catalog item.",
		ConfirmationState: "confirmed",
	}}
	seen := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.URL.Path)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode Bot API body: %v", err)
		}
		switch r.URL.Path {
		case "/botbot-token-1/answerCallbackQuery":
			if body["callback_query_id"] != "callback-3" {
				t.Fatalf("answerCallbackQuery body did not preserve callback id: %+v", body)
			}
		case "/botbot-token-1/editMessageText":
			if body["chat_id"] != "-5235769556" || body["message_id"] != "48" {
				t.Fatalf("editMessageText body did not target original message: %+v", body)
			}
		default:
			t.Fatalf("unexpected Bot API path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte("{\"ok\":true}"))
	}))
	defer server.Close()

	result, err := DispatchUpdateToBotAPI(context.Background(), gateway, BotAPIEndpoint{BaseURL: server.URL, Token: "bot-token-1"}, server.Client(), Update{
		UpdateID: 8003,
		CallbackQuery: &CallbackQuery{
			ID:   "callback-3",
			From: WebhookUser{ID: 12345},
			Message: &CallbackMessage{
				MessageID: 48,
				Chat:      WebhookChat{ID: -5235769556},
			},
			Data: "cabinet:catalog_capture:confirm:preview-3",
		},
	})
	if err != nil {
		t.Fatalf("DispatchUpdateToBotAPI() callback error = %v", err)
	}
	if result.CabinetPath != CaptureCallbackPath || len(result.BotCalls) != 2 || len(result.BotAPIErrors) != 0 {
		t.Fatalf("callback dispatch did not execute cleanly: result=%+v", result)
	}
	if len(seen) != 2 || seen[0] != "/botbot-token-1/answerCallbackQuery" || seen[1] != "/botbot-token-1/editMessageText" {
		t.Fatalf("expected callback ack then edit Bot API calls, saw %#v", seen)
	}
}

func TestDispatchUpdateToBotAPIReportsOutboundFailure(t *testing.T) {
	t.Parallel()

	gateway := &fakeCabinetGateway{response: telegramcapture.TelegramReply{
		Text:              "Draft ready for review: Slot car.",
		ConfirmationState: "preview_required",
	}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/botbot-token-1/sendMessage" {
			t.Fatalf("unexpected Bot API path: %s", r.URL.Path)
		}
		http.Error(w, "{\"ok\":false,\"description\":\"chat not found\"}", http.StatusBadGateway)
	}))
	defer server.Close()

	result, err := DispatchUpdateToBotAPI(context.Background(), gateway, BotAPIEndpoint{BaseURL: server.URL, Token: "bot-token-1"}, server.Client(), Update{
		UpdateID: 7005,
		Message: &WebhookMessage{
			MessageID: 49,
			From:      WebhookUser{ID: 12345},
			Chat:      WebhookChat{ID: -5235769556},
			Text:      "Barcode 4904810900016",
		},
	})
	if err == nil {
		t.Fatalf("expected outbound Bot API failure")
	}
	if result.CabinetPath != WebhookCapturePath || len(result.BotCalls) != 1 || len(result.BotAPIErrors) != 1 {
		t.Fatalf("dispatch did not preserve outbound failure evidence: result=%+v err=%v", result, err)
	}
	if !strings.Contains(result.BotAPIErrors[0], "sendMessage") || !strings.Contains(result.BotAPIErrors[0], "chat not found") {
		t.Fatalf("outbound failure did not include method and response body: %+v", result.BotAPIErrors)
	}
}

func TestBotAPIEndpointNewRequestBindsTokenAndJSONBody(t *testing.T) {
	t.Parallel()

	call, err := SendMessageFromReply("-5235769556", telegramcapture.TelegramReply{Text: "Draft ready."})
	if err != nil {
		t.Fatalf("SendMessageFromReply() error = %v", err)
	}
	req, err := (BotAPIEndpoint{BaseURL: "https://telegram.example.test", Token: "bot-token-1"}).NewRequest(context.Background(), call)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	if req.Method != "POST" || req.URL.String() != "https://telegram.example.test/botbot-token-1/sendMessage" {
		t.Fatalf("unexpected Telegram Bot API request target: method=%s url=%s", req.Method, req.URL.String())
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("expected JSON content type, got %q", req.Header.Get("Content-Type"))
	}
	var body map[string]any
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	if body["chat_id"] != "-5235769556" || body["text"] != "Draft ready." {
		t.Fatalf("request body did not preserve sendMessage payload: %+v", body)
	}
}

func TestBotAPIEndpointBuildsGetFileAndDownloadRequests(t *testing.T) {
	t.Parallel()

	endpoint := BotAPIEndpoint{BaseURL: "https://telegram.example.test", Token: "bot-token-1"}
	getFileReq, err := endpoint.NewGetFileRequest(context.Background(), "telegram-file-1")
	if err != nil {
		t.Fatalf("NewGetFileRequest() error = %v", err)
	}
	if getFileReq.Method != http.MethodPost || getFileReq.URL.String() != "https://telegram.example.test/botbot-token-1/getFile" {
		t.Fatalf("unexpected getFile request: method=%s url=%s", getFileReq.Method, getFileReq.URL.String())
	}
	var getFileBody map[string]any
	if err := json.NewDecoder(getFileReq.Body).Decode(&getFileBody); err != nil {
		t.Fatalf("decode getFile body: %v", err)
	}
	if getFileBody["file_id"] != "telegram-file-1" {
		t.Fatalf("getFile body did not preserve file_id: %+v", getFileBody)
	}
	downloadReq, err := endpoint.NewFileDownloadRequest(context.Background(), "photos/file 1.jpg")
	if err != nil {
		t.Fatalf("NewFileDownloadRequest() error = %v", err)
	}
	if downloadReq.Method != http.MethodGet || downloadReq.URL.String() != "https://telegram.example.test/file/botbot-token-1/photos/file%201.jpg" {
		t.Fatalf("unexpected download request: method=%s url=%s", downloadReq.Method, downloadReq.URL.String())
	}
}

func TestBotAPIFileResolverDownloadsTelegramMediaBytes(t *testing.T) {
	t.Parallel()

	const mediaBytes = "resolved telegram photo bytes"
	var sawGetFile bool
	var sawDownload bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/botbot-token-1/getFile":
			sawGetFile = true
			if r.Method != http.MethodPost {
				t.Fatalf("getFile method = %s", r.Method)
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode getFile request: %v", err)
			}
			if body["file_id"] != "telegram-file-photo-1" {
				t.Fatalf("getFile request did not preserve file_id: %+v", body)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ok":true,"result":{"file_id":"telegram-file-photo-1","file_unique_id":"unique-photo-1","file_size":29,"file_path":"photos/photo-1.jpg"}}`))
		case "/file/botbot-token-1/photos/photo-1.jpg":
			sawDownload = true
			if r.Method != http.MethodGet {
				t.Fatalf("download method = %s", r.Method)
			}
			w.Header().Set("Content-Type", "image/jpeg")
			_, _ = w.Write([]byte(mediaBytes))
		default:
			t.Fatalf("unexpected Telegram test path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	resolver := BotAPIFileResolver{
		Endpoint:   BotAPIEndpoint{BaseURL: server.URL, Token: "bot-token-1"},
		HTTPClient: server.Client(),
	}
	resolved, err := resolver.ResolveTelegramMedia(context.Background(), telegramcapture.MediaInput{
		FileID: "telegram-file-photo-1",
		Kind:   "photo",
	})
	if err != nil {
		t.Fatalf("ResolveTelegramMedia() error = %v", err)
	}
	if !sawGetFile || !sawDownload {
		t.Fatalf("resolver did not call both getFile and file download: getFile=%v download=%v", sawGetFile, sawDownload)
	}
	if resolved.FileID != "telegram-file-photo-1" || resolved.FileUniqueID != "unique-photo-1" || resolved.FileSize != 29 || resolved.Filename != "photo-1.jpg" || resolved.MIMEType != "image/jpeg" || resolved.Kind != "photo" {
		t.Fatalf("resolved media did not preserve expected metadata: %+v", resolved)
	}
	raw, err := io.ReadAll(resolved.Reader)
	if err != nil {
		t.Fatalf("read resolved media: %v", err)
	}
	if string(raw) != mediaBytes {
		t.Fatalf("resolved media bytes = %q", string(raw))
	}
}

func mustJSONMap(t *testing.T, value any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	return out
}

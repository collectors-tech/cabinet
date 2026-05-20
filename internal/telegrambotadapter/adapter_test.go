package telegrambotadapter

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/collectors-tech/cabinet/internal/telegramcapture"
)

type fakeCabinetGateway struct {
	path     string
	body     any
	response telegramcapture.TelegramReply
}

func (f *fakeCabinetGateway) PostJSON(_ context.Context, path string, body any, response any) (int, error) {
	f.path = path
	f.body = body
	out := response.(*cabinetTelegramReplyResponse)
	out.TelegramReply = f.response
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

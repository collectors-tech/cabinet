package telegrambotadapter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/collectors-tech/cabinet/internal/telegramcapture"
)

const (
	WebhookCapturePath  = "/api/telegram/webhook/catalog-captures"
	CaptureCallbackPath = "/api/telegram/catalog-capture-callbacks"
)

type Update struct {
	UpdateID      int64           `json:"update_id"`
	Message       *WebhookMessage `json:"message,omitempty"`
	CallbackQuery *CallbackQuery  `json:"callback_query,omitempty"`
}

type WebhookMessage = telegramcapture.WebhookMessage

type CallbackQuery struct {
	ID      string           `json:"id"`
	From    WebhookUser      `json:"from"`
	Message *CallbackMessage `json:"message,omitempty"`
	Data    string           `json:"data"`
}

type WebhookUser = telegramcapture.WebhookUser
type WebhookChat = telegramcapture.WebhookChat

type CallbackMessage struct {
	MessageID int64       `json:"message_id"`
	Chat      WebhookChat `json:"chat"`
}

type CabinetRequest struct {
	Path string
	Body any
}

type CaptureCallbackRequest struct {
	SenderID     string `json:"sender_id"`
	ChatID       string `json:"chat_id"`
	CallbackData string `json:"callback_data"`
}

func CabinetRequestFromUpdate(update Update) (CabinetRequest, error) {
	if update.CallbackQuery != nil {
		callback := update.CallbackQuery
		if callback.Message == nil {
			return CabinetRequest{}, fmt.Errorf("telegram callback message is required")
		}
		body := CaptureCallbackRequest{
			SenderID:     telegramID(callback.From.ID),
			ChatID:       telegramID(callback.Message.Chat.ID),
			CallbackData: strings.TrimSpace(callback.Data),
		}
		if body.SenderID == "" || body.ChatID == "" || body.CallbackData == "" {
			return CabinetRequest{}, fmt.Errorf("telegram callback sender, chat, and callback data are required")
		}
		return CabinetRequest{Path: CaptureCallbackPath, Body: body}, nil
	}
	if update.Message != nil {
		return CabinetRequest{
			Path: WebhookCapturePath,
			Body: telegramcapture.WebhookUpdate{UpdateID: update.UpdateID, Message: update.Message},
		}, nil
	}
	return CabinetRequest{}, fmt.Errorf("telegram update does not contain a supported catalog capture message or callback")
}

type BotAPICall struct {
	Method string
	Body   any
}

type SendMessageRequest struct {
	ChatID      string `json:"chat_id"`
	Text        string `json:"text"`
	ReplyMarkup any    `json:"reply_markup,omitempty"`
}

type EditMessageTextRequest struct {
	ChatID      string `json:"chat_id"`
	MessageID   string `json:"message_id"`
	Text        string `json:"text"`
	ReplyMarkup any    `json:"reply_markup,omitempty"`
}

type AnswerCallbackQueryRequest struct {
	CallbackQueryID string `json:"callback_query_id"`
	Text            string `json:"text,omitempty"`
	ShowAlert       bool   `json:"show_alert"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	URL          string `json:"url,omitempty"`
	CallbackData string `json:"callback_data,omitempty"`
}

type ReplyKeyboardMarkup struct {
	Keyboard        [][]KeyboardButton `json:"keyboard"`
	ResizeKeyboard  bool               `json:"resize_keyboard"`
	OneTimeKeyboard bool               `json:"one_time_keyboard"`
}

type KeyboardButton struct {
	Text string `json:"text"`
}

func SendMessageFromReply(chatID string, reply telegramcapture.TelegramReply) (BotAPICall, error) {
	chatID = strings.TrimSpace(chatID)
	if chatID == "" {
		return BotAPICall{}, fmt.Errorf("telegram chat_id is required")
	}
	body := SendMessageRequest{
		ChatID:      chatID,
		Text:        replyText(reply),
		ReplyMarkup: replyMarkup(reply),
	}
	return BotAPICall{Method: "sendMessage", Body: body}, nil
}

func EditMessageFromReply(chatID, messageID string, reply telegramcapture.TelegramReply) (BotAPICall, error) {
	chatID = strings.TrimSpace(chatID)
	messageID = strings.TrimSpace(messageID)
	if chatID == "" || messageID == "" {
		return BotAPICall{}, fmt.Errorf("telegram chat_id and message_id are required")
	}
	body := EditMessageTextRequest{
		ChatID:      chatID,
		MessageID:   messageID,
		Text:        replyText(reply),
		ReplyMarkup: replyMarkup(reply),
	}
	return BotAPICall{Method: "editMessageText", Body: body}, nil
}

func AnswerCallbackQueryFromReply(callbackQueryID string, reply telegramcapture.TelegramReply) (BotAPICall, error) {
	callbackQueryID = strings.TrimSpace(callbackQueryID)
	if callbackQueryID == "" {
		return BotAPICall{}, fmt.Errorf("telegram callback_query_id is required")
	}
	body := AnswerCallbackQueryRequest{
		CallbackQueryID: callbackQueryID,
		Text:            replyText(reply),
		ShowAlert:       false,
	}
	return BotAPICall{Method: "answerCallbackQuery", Body: body}, nil
}

func MarshalBody(call BotAPICall) ([]byte, error) {
	return json.Marshal(call.Body)
}

func replyText(reply telegramcapture.TelegramReply) string {
	text := strings.TrimSpace(reply.Text)
	if text == "" {
		return "Cabinet updated the Telegram catalog capture."
	}
	return text
}

func replyMarkup(reply telegramcapture.TelegramReply) any {
	inlineRows := [][]InlineKeyboardButton{}
	replyRows := [][]KeyboardButton{}
	for _, button := range reply.ActionButtons {
		label := strings.TrimSpace(button.Label)
		if label == "" {
			continue
		}
		switch strings.TrimSpace(button.Kind) {
		case "url":
			if url := strings.TrimSpace(button.URL); url != "" {
				inlineRows = append(inlineRows, []InlineKeyboardButton{{Text: label, URL: url}})
			}
		case "callback":
			if data := strings.TrimSpace(button.CallbackData); data != "" {
				inlineRows = append(inlineRows, []InlineKeyboardButton{{Text: label, CallbackData: data}})
			}
		case "reply":
			replyRows = append(replyRows, []KeyboardButton{{Text: label}})
		}
	}
	if len(inlineRows) > 0 {
		return InlineKeyboardMarkup{InlineKeyboard: inlineRows}
	}
	if len(replyRows) > 0 {
		return ReplyKeyboardMarkup{Keyboard: replyRows, ResizeKeyboard: true, OneTimeKeyboard: true}
	}
	return nil
}

func telegramID(id int64) string {
	if id == 0 {
		return ""
	}
	return fmt.Sprintf("%d", id)
}

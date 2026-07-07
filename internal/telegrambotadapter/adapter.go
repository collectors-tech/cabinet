package telegrambotadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/collectors-tech/cabinet/internal/telegramcapture"
)

const (
	WebhookCapturePath        = "/api/telegram/webhook/catalog-captures"
	CaptureCallbackPath       = "/api/telegram/catalog-capture-callbacks"
	AgentTextPath             = "/api/telegram/agent-text"
	AgentTextCallbackPath     = "/api/telegram/agent-text-callbacks"
	telegramMessageTextLimit  = 4096
	telegramCallbackTextLimit = 200
	telegramButtonTextLimit   = 64
	telegramCallbackDataLimit = 64
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

type AgentTextRequest struct {
	SenderID       string         `json:"sender_id"`
	ChatID         string         `json:"chat_id"`
	MessageID      string         `json:"message_id"`
	Text           string         `json:"text"`
	SkillID        string         `json:"skill_id,omitempty"`
	Parameters     map[string]any `json:"parameters,omitempty"`
	SourceMetadata map[string]any `json:"source_metadata,omitempty"`
}

type AgentTextCallbackRequest struct {
	SenderID     string `json:"sender_id"`
	ChatID       string `json:"chat_id"`
	MessageID    string `json:"message_id"`
	ThreadID     string `json:"thread_id,omitempty"`
	PreviewID    string `json:"preview_id,omitempty"`
	Confirmation string `json:"confirmation,omitempty"`
	CallbackData string `json:"callback_data"`
}

func CabinetRequestFromUpdate(update Update) (CabinetRequest, error) {
	if update.CallbackQuery != nil {
		callback := update.CallbackQuery
		if callback.Message == nil {
			return CabinetRequest{}, fmt.Errorf("telegram callback message is required")
		}
		if isAgentTextCallback(callback.Data) {
			body := agentTextCallbackRequestFromCallback(callback)
			if body.SenderID == "" || body.ChatID == "" || body.CallbackData == "" {
				return CabinetRequest{}, fmt.Errorf("telegram agent callback sender, chat, and callback data are required")
			}
			return CabinetRequest{Path: AgentTextCallbackPath, Body: body}, nil
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
		if isAgentTextCommand(update.Message.Text) {
			body := agentTextRequestFromMessage(update.UpdateID, update.Message)
			if body.SenderID == "" || body.ChatID == "" || body.MessageID == "" || body.Text == "" {
				return CabinetRequest{}, fmt.Errorf("telegram agent text sender, chat, message, and text are required")
			}
			return CabinetRequest{Path: AgentTextPath, Body: body}, nil
		}
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

type CabinetGateway interface {
	PostJSON(ctx context.Context, path string, body any, response any) (int, error)
}

type DispatchResult struct {
	CabinetPath  string
	CabinetError string
	BotAPIErrors []string
	BotCalls     []BotAPICall
}

type cabinetTelegramReplyResponse struct {
	TelegramReply telegramcapture.TelegramReply `json:"telegram_reply"`
}

func DispatchUpdate(ctx context.Context, gateway CabinetGateway, update Update) (DispatchResult, error) {
	if gateway == nil {
		return DispatchResult{}, fmt.Errorf("cabinet gateway is required")
	}
	req, err := CabinetRequestFromUpdate(update)
	if err != nil {
		return DispatchResult{}, err
	}
	var response cabinetTelegramReplyResponse
	_, postErr := gateway.PostJSON(ctx, req.Path, req.Body, &response)
	result := DispatchResult{CabinetPath: req.Path}
	if postErr != nil {
		result.CabinetError = postErr.Error()
		if response.TelegramReply.Text == "" {
			response.TelegramReply = failureTelegramReply()
		}
	}
	if update.CallbackQuery != nil {
		callback := update.CallbackQuery
		ack, err := AnswerCallbackQueryFromReply(callback.ID, response.TelegramReply)
		if err != nil {
			return DispatchResult{}, err
		}
		edit, err := EditMessageFromReply(telegramID(callback.Message.Chat.ID), telegramID(callback.Message.MessageID), response.TelegramReply)
		if err != nil {
			return DispatchResult{}, err
		}
		result.BotCalls = []BotAPICall{ack, edit}
		return result, nil
	}
	if update.Message != nil {
		send, err := SendMessageFromReply(telegramID(update.Message.Chat.ID), response.TelegramReply)
		if err != nil {
			return DispatchResult{}, err
		}
		result.BotCalls = []BotAPICall{send}
		return result, nil
	}
	return DispatchResult{}, fmt.Errorf("telegram update does not contain a supported catalog capture message or callback")
}

func DispatchUpdateToBotAPI(ctx context.Context, gateway CabinetGateway, endpoint BotAPIEndpoint, client HTTPDoer, update Update) (DispatchResult, error) {
	result, err := DispatchUpdate(ctx, gateway, update)
	if err != nil {
		return result, err
	}
	if client == nil {
		client = http.DefaultClient
	}
	for _, call := range result.BotCalls {
		req, err := endpoint.NewRequest(ctx, call)
		if err != nil {
			result.BotAPIErrors = append(result.BotAPIErrors, err.Error())
			return result, err
		}
		resp, err := client.Do(req)
		if err != nil {
			result.BotAPIErrors = append(result.BotAPIErrors, err.Error())
			return result, err
		}
		if resp.Body != nil {
			defer resp.Body.Close()
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body := ""
			if resp.Body != nil {
				raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
				body = strings.TrimSpace(string(raw))
			}
			err := fmt.Errorf("telegram bot api %s returned status %d", call.Method, resp.StatusCode)
			if body != "" {
				err = fmt.Errorf("%w: %s", err, body)
			}
			result.BotAPIErrors = append(result.BotAPIErrors, err.Error())
			return result, err
		}
	}
	return result, nil
}

func failureTelegramReply() telegramcapture.TelegramReply {
	return telegramcapture.TelegramReply{
		Text:              "Cabinet could not process this Telegram catalog capture yet. Please retry or open Cabinet to review the capture.",
		ConfirmationState: "error",
		Actions:           []string{"retry", "open_cabinet"},
	}
}

type BotAPIEndpoint struct {
	BaseURL string
	Token   string
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type BotAPIFileResolver struct {
	Endpoint   BotAPIEndpoint
	HTTPClient HTTPDoer
}

type GetFileRequest struct {
	FileID string `json:"file_id"`
}

type getFileResponse struct {
	OK          bool           `json:"ok"`
	Description string         `json:"description,omitempty"`
	Result      BotAPIFileInfo `json:"result"`
}

type BotAPIFileInfo struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	FileSize     int    `json:"file_size"`
	FilePath     string `json:"file_path"`
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
		Text:        replyText(reply, telegramMessageTextLimit),
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
		Text:        replyText(reply, telegramMessageTextLimit),
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
		Text:            replyText(reply, telegramCallbackTextLimit),
		ShowAlert:       false,
	}
	return BotAPICall{Method: "answerCallbackQuery", Body: body}, nil
}

func MarshalBody(call BotAPICall) ([]byte, error) {
	return json.Marshal(call.Body)
}

func (endpoint BotAPIEndpoint) NewRequest(ctx context.Context, call BotAPICall) (*http.Request, error) {
	method := strings.TrimSpace(call.Method)
	if method == "" {
		return nil, fmt.Errorf("telegram bot api method is required")
	}
	token := strings.TrimSpace(endpoint.Token)
	if token == "" {
		return nil, fmt.Errorf("telegram bot token is required")
	}
	base := strings.TrimRight(strings.TrimSpace(endpoint.BaseURL), "/")
	if base == "" {
		base = "https://api.telegram.org"
	}
	parsed, err := url.Parse(base + "/bot" + url.PathEscape(token) + "/" + method)
	if err != nil {
		return nil, err
	}
	body, err := MarshalBody(call)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, parsed.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (endpoint BotAPIEndpoint) NewGetFileRequest(ctx context.Context, fileID string) (*http.Request, error) {
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return nil, fmt.Errorf("telegram file_id is required")
	}
	return endpoint.NewRequest(ctx, BotAPICall{
		Method: "getFile",
		Body:   GetFileRequest{FileID: fileID},
	})
}

func (endpoint BotAPIEndpoint) NewFileDownloadRequest(ctx context.Context, filePath string) (*http.Request, error) {
	filePath = strings.Trim(strings.TrimSpace(filePath), "/")
	if filePath == "" {
		return nil, fmt.Errorf("telegram file_path is required")
	}
	token := strings.TrimSpace(endpoint.Token)
	if token == "" {
		return nil, fmt.Errorf("telegram bot token is required")
	}
	base := strings.TrimRight(strings.TrimSpace(endpoint.BaseURL), "/")
	if base == "" {
		base = "https://api.telegram.org"
	}
	parsed, err := url.Parse(base + "/file/bot" + url.PathEscape(token) + "/" + escapeTelegramFilePath(filePath))
	if err != nil {
		return nil, err
	}
	return http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
}

func (resolver BotAPIFileResolver) ResolveTelegramMedia(ctx context.Context, media telegramcapture.MediaInput) (telegramcapture.MediaInput, error) {
	fileID := strings.TrimSpace(media.FileID)
	if fileID == "" {
		return telegramcapture.MediaInput{}, fmt.Errorf("telegram media file_id is required")
	}
	client := resolver.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	getFileReq, err := resolver.Endpoint.NewGetFileRequest(ctx, fileID)
	if err != nil {
		return telegramcapture.MediaInput{}, err
	}
	getFileResp, err := client.Do(getFileReq)
	if err != nil {
		return telegramcapture.MediaInput{}, err
	}
	defer getFileResp.Body.Close()
	if getFileResp.StatusCode < 200 || getFileResp.StatusCode >= 300 {
		return telegramcapture.MediaInput{}, fmt.Errorf("telegram getFile returned status %d", getFileResp.StatusCode)
	}
	var payload getFileResponse
	if err := json.NewDecoder(getFileResp.Body).Decode(&payload); err != nil {
		return telegramcapture.MediaInput{}, err
	}
	if !payload.OK || strings.TrimSpace(payload.Result.FilePath) == "" {
		description := strings.TrimSpace(payload.Description)
		if description == "" {
			description = "telegram getFile did not return a downloadable file path"
		}
		return telegramcapture.MediaInput{}, fmt.Errorf("%s", description)
	}
	downloadReq, err := resolver.Endpoint.NewFileDownloadRequest(ctx, payload.Result.FilePath)
	if err != nil {
		return telegramcapture.MediaInput{}, err
	}
	downloadResp, err := client.Do(downloadReq)
	if err != nil {
		return telegramcapture.MediaInput{}, err
	}
	defer downloadResp.Body.Close()
	if downloadResp.StatusCode < 200 || downloadResp.StatusCode >= 300 {
		return telegramcapture.MediaInput{}, fmt.Errorf("telegram file download returned status %d", downloadResp.StatusCode)
	}
	body, err := io.ReadAll(downloadResp.Body)
	if err != nil {
		return telegramcapture.MediaInput{}, err
	}
	resolved := media
	if id := strings.TrimSpace(payload.Result.FileID); id != "" {
		resolved.FileID = id
	}
	if uniqueID := strings.TrimSpace(payload.Result.FileUniqueID); uniqueID != "" {
		resolved.FileUniqueID = uniqueID
	}
	if payload.Result.FileSize > 0 {
		resolved.FileSize = payload.Result.FileSize
	}
	if strings.TrimSpace(resolved.Filename) == "" {
		resolved.Filename = telegramFileName(payload.Result.FilePath)
	}
	if strings.TrimSpace(resolved.MIMEType) == "" {
		resolved.MIMEType = strings.TrimSpace(downloadResp.Header.Get("Content-Type"))
	}
	resolved.Reader = bytes.NewReader(body)
	return resolved, nil
}

func escapeTelegramFilePath(filePath string) string {
	segments := strings.Split(strings.Trim(filePath, "/"), "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}

func telegramFileName(filePath string) string {
	segments := strings.Split(strings.Trim(filePath, "/"), "/")
	for i := len(segments) - 1; i >= 0; i-- {
		if name := strings.TrimSpace(segments[i]); name != "" {
			return name
		}
	}
	return ""
}

func replyText(reply telegramcapture.TelegramReply, limit int) string {
	text := strings.TrimSpace(reply.Text)
	if text == "" {
		text = "Cabinet updated the Telegram catalog capture."
	}
	return truncateTelegramText(text, limit)
}

func truncateTelegramText(text string, limit int) string {
	if limit <= 0 {
		return text
	}
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	if limit <= 3 {
		return string(runes[:limit])
	}
	return string(runes[:limit-3]) + "..."
}

func replyMarkup(reply telegramcapture.TelegramReply) any {
	inlineRows := [][]InlineKeyboardButton{}
	replyRows := [][]KeyboardButton{}
	for _, button := range reply.ActionButtons {
		label := truncateTelegramText(strings.TrimSpace(button.Label), telegramButtonTextLimit)
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
				if len([]byte(data)) <= telegramCallbackDataLimit {
					inlineRows = append(inlineRows, []InlineKeyboardButton{{Text: label, CallbackData: data}})
				}
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

func isAgentTextCommand(text string) bool {
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(trimmed)
	return lower == "/agent" || strings.HasPrefix(lower, "/agent ") || strings.HasPrefix(lower, "agent:")
}

func agentTextRequestFromMessage(updateID int64, message *WebhookMessage) AgentTextRequest {
	commandText := strings.TrimSpace(message.Text)
	commandText = strings.TrimSpace(strings.TrimPrefix(commandText, "/agent"))
	if strings.HasPrefix(strings.ToLower(commandText), "agent:") {
		commandText = strings.TrimSpace(commandText[len("agent:"):])
	}
	skillID, params := parseAgentTextCommand(commandText)
	return AgentTextRequest{
		SenderID:   telegramID(message.From.ID),
		ChatID:     telegramID(message.Chat.ID),
		MessageID:  telegramID(message.MessageID),
		Text:       commandText,
		SkillID:    skillID,
		Parameters: params,
		SourceMetadata: map[string]any{
			"update_id": updateID,
			"command":   "agent",
		},
	}
}

func parseAgentTextCommand(text string) (string, map[string]any) {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 {
		return "", nil
	}
	skillID := ""
	start := 0
	if strings.HasPrefix(fields[0], "cabinet.") {
		skillID = fields[0]
		start = 1
	}
	params := map[string]any{}
	for _, field := range fields[start:] {
		key, value, ok := strings.Cut(field, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key != "" {
			params[key] = value
		}
	}
	if len(params) == 0 {
		params = nil
	}
	return skillID, params
}

func isAgentTextCallback(data string) bool {
	return strings.HasPrefix(strings.TrimSpace(data), "cabinet:agent_text:")
}

func agentTextCallbackRequestFromCallback(callback *CallbackQuery) AgentTextCallbackRequest {
	confirmation, previewID := parseAgentTextCallbackData(callback.Data)
	return AgentTextCallbackRequest{
		SenderID:     telegramID(callback.From.ID),
		ChatID:       telegramID(callback.Message.Chat.ID),
		MessageID:    telegramID(callback.Message.MessageID),
		PreviewID:    previewID,
		Confirmation: confirmation,
		CallbackData: strings.TrimSpace(callback.Data),
	}
}

func parseAgentTextCallbackData(data string) (string, string) {
	parts := strings.Split(strings.TrimSpace(data), ":")
	if len(parts) < 4 || parts[0] != "cabinet" || parts[1] != "agent_text" {
		return "", ""
	}
	return strings.TrimSpace(parts[2]), strings.TrimSpace(parts[3])
}

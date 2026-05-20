package telegramcapture

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/collectors-tech/cabinet/internal/chat"
)

var ErrUnauthorizedSender = errors.New("telegram sender is not authorized for Cabinet capture")
var ErrDraftNeedsFollowUp = errors.New("telegram catalog capture needs follow-up before preview")

type Authorizer interface {
	AuthorizeTelegramCapture(ctx context.Context, senderID, chatID string) (AuthorizedProfile, error)
}

type AuthorizedProfile struct {
	ProfileID string
}

type ChatService interface {
	CreateThread(ctx context.Context, profileID, title string, metadata map[string]any) (chat.Thread, error)
	CreateMessage(ctx context.Context, profileID, threadID, role, content string, messageContext map[string]any) (chat.Message, error)
	SaveAttachment(ctx context.Context, profileID, threadID, filename, mimeType string, src io.Reader) (chat.Attachment, error)
	PreviewAction(ctx context.Context, in chat.PreviewActionInput) (chat.ActionPreview, error)
	CreateInboxItem(ctx context.Context, item chat.InboxItem) (chat.InboxItem, error)
}

type Service struct {
	authorizer Authorizer
	chat       ChatService
}

func NewService(authorizer Authorizer, chatSvc ChatService) *Service {
	return &Service{authorizer: authorizer, chat: chatSvc}
}

type CaptureInput struct {
	SenderID       string
	ChatID         string
	MessageID      string
	Text           string
	Barcode        string
	Draft          Draft
	Media          []MediaInput
	GroupingHint   string
	SourceMetadata map[string]any
}

type Draft struct {
	PartNumber string
	Title      string
	Brand      string
	Category   string
}

type MediaInput struct {
	FileID   string
	Filename string
	MIMEType string
	Kind     string
	Reader   io.Reader
}

type WebhookUpdate struct {
	UpdateID int64           `json:"update_id"`
	Message  *WebhookMessage `json:"message"`
}

type WebhookMessage struct {
	MessageID    int64              `json:"message_id"`
	Date         int64              `json:"date"`
	From         WebhookUser        `json:"from"`
	Chat         WebhookChat        `json:"chat"`
	Text         string             `json:"text"`
	Caption      string             `json:"caption"`
	MediaGroupID string             `json:"media_group_id"`
	Photo        []WebhookPhotoSize `json:"photo"`
	Document     *WebhookDocument   `json:"document"`
}

type WebhookUser struct {
	ID int64 `json:"id"`
}

type WebhookChat struct {
	ID int64 `json:"id"`
}

type WebhookPhotoSize struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FileSize     int    `json:"file_size"`
}

type WebhookDocument struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	FileName     string `json:"file_name"`
	MIMEType     string `json:"mime_type"`
	FileSize     int    `json:"file_size"`
}

type CaptureResult struct {
	ProfileID     string             `json:"profile_id"`
	Thread        chat.Thread        `json:"thread"`
	Message       chat.Message       `json:"message"`
	Attachments   []chat.Attachment  `json:"attachments"`
	Preview       chat.ActionPreview `json:"preview"`
	InboxItem     chat.InboxItem     `json:"inbox_item"`
	TelegramReply TelegramReply      `json:"telegram_reply"`
}

type TelegramReply struct {
	Text              string                `json:"text"`
	ReviewURL         string                `json:"review_url,omitempty"`
	ConfirmationState string                `json:"confirmation_state"`
	Actions           []string              `json:"actions"`
	ActionButtons     []TelegramReplyButton `json:"action_buttons,omitempty"`
}

type TelegramReplyButton struct {
	Label        string `json:"label"`
	Kind         string `json:"kind"`
	Action       string `json:"action"`
	URL          string `json:"url,omitempty"`
	CallbackData string `json:"callback_data,omitempty"`
}

type DraftNeedsFollowUpError struct {
	Reply         TelegramReply `json:"telegram_reply"`
	Reason        string        `json:"reason"`
	MissingFields []string      `json:"missing_fields"`
}

func (e DraftNeedsFollowUpError) Error() string {
	reason := strings.TrimSpace(e.Reason)
	if reason == "" {
		return ErrDraftNeedsFollowUp.Error()
	}
	return ErrDraftNeedsFollowUp.Error() + ": " + reason
}

func (e DraftNeedsFollowUpError) Unwrap() error {
	return ErrDraftNeedsFollowUp
}

func (s *Service) IngestCatalogCapture(ctx context.Context, in CaptureInput) (CaptureResult, error) {
	if s == nil || s.authorizer == nil || s.chat == nil {
		return CaptureResult{}, fmt.Errorf("telegram capture service is not configured")
	}
	senderID := strings.TrimSpace(in.SenderID)
	chatID := strings.TrimSpace(in.ChatID)
	if senderID == "" || chatID == "" {
		return CaptureResult{}, fmt.Errorf("telegram sender_id and chat_id are required")
	}
	authorized, err := s.authorizer.AuthorizeTelegramCapture(ctx, senderID, chatID)
	if err != nil {
		return CaptureResult{}, err
	}
	profileID := strings.TrimSpace(authorized.ProfileID)
	if profileID == "" {
		return CaptureResult{}, ErrUnauthorizedSender
	}

	draft := normalizeDraft(in.Draft, in.Barcode)
	if strings.TrimSpace(draft.PartNumber) == "" || strings.TrimSpace(draft.Title) == "" {
		return CaptureResult{}, draftNeedsFollowUp(in, draft)
	}

	thread, err := s.chat.CreateThread(ctx, profileID, threadTitle(draft), map[string]any{
		"source_channel":    "telegram",
		"source_chat_id":    chatID,
		"source_sender_id":  senderID,
		"source_message_id": strings.TrimSpace(in.MessageID),
		"grouping_hint":     strings.TrimSpace(in.GroupingHint),
	})
	if err != nil {
		return CaptureResult{}, err
	}

	attachments := make([]chat.Attachment, 0, len(in.Media))
	mediaContext := make([]map[string]any, 0, len(in.Media))
	for _, media := range in.Media {
		filename := strings.TrimSpace(media.Filename)
		if filename == "" {
			filename = strings.TrimSpace(media.FileID)
		}
		if filename == "" {
			return CaptureResult{}, fmt.Errorf("telegram media filename or file_id is required")
		}
		reader := media.Reader
		if reader == nil {
			reader = strings.NewReader("")
		}
		attachment, err := s.chat.SaveAttachment(ctx, profileID, thread.ID, filename, strings.TrimSpace(media.MIMEType), reader)
		if err != nil {
			return CaptureResult{}, err
		}
		attachments = append(attachments, attachment)
		mediaContext = append(mediaContext, map[string]any{
			"attachment_id": attachment.ID,
			"file_id":       strings.TrimSpace(media.FileID),
			"filename":      attachment.Filename,
			"mime_type":     attachment.MimeType,
			"kind":          strings.TrimSpace(media.Kind),
		})
	}

	messageContext := map[string]any{
		"source_channel":    "telegram",
		"source_chat_id":    chatID,
		"source_sender_id":  senderID,
		"source_message_id": strings.TrimSpace(in.MessageID),
		"barcode":           strings.TrimSpace(in.Barcode),
		"grouping_hint":     strings.TrimSpace(in.GroupingHint),
		"media":             mediaContext,
	}
	if len(in.SourceMetadata) > 0 {
		messageContext["source_metadata"] = in.SourceMetadata
	}
	message, err := s.chat.CreateMessage(ctx, profileID, thread.ID, "user", messageContent(in, draft), messageContext)
	if err != nil {
		return CaptureResult{}, err
	}

	preview, err := s.chat.PreviewAction(ctx, chat.PreviewActionInput{
		ProfileID: profileID,
		ThreadID:  thread.ID,
		Action:    "create_inventory_item",
		Payload: map[string]any{
			"part_number": strings.TrimSpace(draft.PartNumber),
			"title":       strings.TrimSpace(draft.Title),
			"brand":       strings.TrimSpace(draft.Brand),
			"category":    strings.TrimSpace(draft.Category),
			"source":      "telegram",
			"barcode":     strings.TrimSpace(in.Barcode),
		},
	})
	if err != nil {
		return CaptureResult{}, err
	}

	reviewURL := reviewURL(profileID, thread.ID, preview.ID)
	telegramReply := telegramReply(draft, reviewURL, preview.ID)
	inboxItem, err := s.chat.CreateInboxItem(ctx, chat.InboxItem{
		ProfileID: profileID,
		ThreadID:  thread.ID,
		Source:    "telegram_catalog_capture",
		Status:    "queued",
		Title:     "Telegram catalog draft ready",
		Summary:   "Review and confirm the Telegram catalog draft before it can mutate Cabinet inventory.",
		Metadata: map[string]any{
			"preview_id":         preview.ID,
			"source_channel":     "telegram",
			"source_chat_id":     chatID,
			"source_sender_id":   senderID,
			"source_message_id":  strings.TrimSpace(in.MessageID),
			"attachment_count":   len(attachments),
			"confirmation_state": "preview_required",
			"review_url":         reviewURL,
			"telegram_reply":     telegramReply,
		},
	})
	if err != nil {
		return CaptureResult{}, err
	}

	return CaptureResult{
		ProfileID:     profileID,
		Thread:        thread,
		Message:       message,
		Attachments:   attachments,
		Preview:       preview,
		InboxItem:     inboxItem,
		TelegramReply: telegramReply,
	}, nil
}

func CaptureInputFromWebhookUpdate(update WebhookUpdate, draft Draft) (CaptureInput, error) {
	if update.Message == nil {
		return CaptureInput{}, fmt.Errorf("telegram webhook update does not contain a message")
	}
	message := update.Message
	senderID := telegramID(message.From.ID)
	chatID := telegramID(message.Chat.ID)
	if senderID == "" || chatID == "" {
		return CaptureInput{}, fmt.Errorf("telegram webhook sender and chat are required")
	}
	text := strings.TrimSpace(message.Text)
	if text == "" {
		text = strings.TrimSpace(message.Caption)
	}
	media := webhookMediaInputs(message)
	return CaptureInput{
		SenderID:     senderID,
		ChatID:       chatID,
		MessageID:    telegramID(message.MessageID),
		Text:         text,
		Barcode:      inferBarcode(text),
		Draft:        draft,
		Media:        media,
		GroupingHint: strings.TrimSpace(message.MediaGroupID),
		SourceMetadata: map[string]any{
			"update_id":      update.UpdateID,
			"message_date":   message.Date,
			"media_group_id": strings.TrimSpace(message.MediaGroupID),
			"payload_type":   webhookPayloadType(message),
		},
	}, nil
}

func webhookMediaInputs(message *WebhookMessage) []MediaInput {
	if message == nil {
		return nil
	}
	media := []MediaInput{}
	if photo := largestPhoto(message.Photo); photo.FileID != "" {
		filename := strings.TrimSpace(photo.FileUniqueID)
		if filename == "" {
			filename = strings.TrimSpace(photo.FileID)
		}
		media = append(media, MediaInput{
			FileID:   strings.TrimSpace(photo.FileID),
			Filename: filename + ".jpg",
			MIMEType: "image/jpeg",
			Kind:     "photo",
		})
	}
	if doc := message.Document; doc != nil && strings.HasPrefix(strings.ToLower(strings.TrimSpace(doc.MIMEType)), "image/") {
		filename := strings.TrimSpace(doc.FileName)
		if filename == "" {
			filename = strings.TrimSpace(doc.FileUniqueID)
		}
		if filename == "" {
			filename = strings.TrimSpace(doc.FileID)
		}
		media = append(media, MediaInput{
			FileID:   strings.TrimSpace(doc.FileID),
			Filename: filename,
			MIMEType: strings.TrimSpace(doc.MIMEType),
			Kind:     "document_image",
		})
	}
	return media
}

func largestPhoto(photos []WebhookPhotoSize) WebhookPhotoSize {
	var selected WebhookPhotoSize
	for _, photo := range photos {
		if strings.TrimSpace(photo.FileID) == "" {
			continue
		}
		if selected.FileID == "" || photo.FileSize > selected.FileSize || (photo.FileSize == selected.FileSize && photo.Width*photo.Height > selected.Width*selected.Height) {
			selected = photo
		}
	}
	return selected
}

func webhookPayloadType(message *WebhookMessage) string {
	if message == nil {
		return "unknown"
	}
	parts := []string{}
	if strings.TrimSpace(message.Text) != "" {
		parts = append(parts, "text")
	}
	if strings.TrimSpace(message.Caption) != "" {
		parts = append(parts, "caption")
	}
	if len(message.Photo) > 0 {
		parts = append(parts, "photo")
	}
	if message.Document != nil {
		parts = append(parts, "document")
	}
	if len(parts) == 0 {
		return "unknown"
	}
	return strings.Join(parts, "+")
}

func inferBarcode(text string) string {
	var best strings.Builder
	var current strings.Builder
	flush := func() {
		if current.Len() >= 8 && current.Len() <= 14 && current.Len() > best.Len() {
			best.Reset()
			best.WriteString(current.String())
		}
		current.Reset()
	}
	for _, r := range text {
		if r >= '0' && r <= '9' {
			current.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	return best.String()
}

func telegramID(id int64) string {
	if id == 0 {
		return ""
	}
	return strconv.FormatInt(id, 10)
}

func normalizeDraft(draft Draft, barcode string) Draft {
	draft.PartNumber = strings.TrimSpace(draft.PartNumber)
	draft.Title = strings.TrimSpace(draft.Title)
	draft.Brand = strings.TrimSpace(draft.Brand)
	draft.Category = strings.TrimSpace(draft.Category)
	barcode = strings.TrimSpace(barcode)
	if draft.PartNumber == "" {
		draft.PartNumber = barcode
	}
	if draft.Title == "" && barcode != "" {
		draft.Title = "Barcode " + barcode
	}
	return draft
}

func threadTitle(draft Draft) string {
	if title := strings.TrimSpace(draft.Title); title != "" {
		return "Telegram capture: " + title
	}
	return "Telegram catalog capture"
}

func messageContent(in CaptureInput, draft Draft) string {
	parts := []string{}
	if text := strings.TrimSpace(in.Text); text != "" {
		parts = append(parts, text)
	}
	if barcode := strings.TrimSpace(in.Barcode); barcode != "" {
		parts = append(parts, "Barcode: "+barcode)
	}
	if len(in.Media) > 0 {
		parts = append(parts, fmt.Sprintf("Media attachments: %d", len(in.Media)))
	}
	parts = append(parts, "Draft: "+strings.TrimSpace(draft.PartNumber)+" - "+strings.TrimSpace(draft.Title))
	return strings.Join(parts, "\n")
}

func reviewURL(profileID, threadID, previewID string) string {
	return "/chats?profile_id=" + strings.TrimSpace(profileID) + "&thread_id=" + strings.TrimSpace(threadID) + "&preview_id=" + strings.TrimSpace(previewID)
}

func telegramReply(draft Draft, reviewURL, previewID string) TelegramReply {
	title := strings.TrimSpace(draft.Title)
	if title == "" {
		title = "catalog draft"
	}
	previewID = strings.TrimSpace(previewID)
	return TelegramReply{
		Text:              "Draft ready for review: " + title + ". Open Cabinet to confirm or cancel before anything is added to inventory.",
		ReviewURL:         reviewURL,
		ConfirmationState: "preview_required",
		Actions:           []string{"open_cabinet_review", "confirm_in_cabinet", "cancel_in_cabinet"},
		ActionButtons: []TelegramReplyButton{
			{Label: "Open Cabinet review", Kind: "url", Action: "open_cabinet_review", URL: strings.TrimSpace(reviewURL)},
			{Label: "Confirm in Cabinet", Kind: "callback", Action: "confirm_in_cabinet", CallbackData: telegramCallbackData("confirm", previewID)},
			{Label: "Cancel in Cabinet", Kind: "callback", Action: "cancel_in_cabinet", CallbackData: telegramCallbackData("cancel", previewID)},
		},
	}
}

func telegramCallbackData(action, previewID string) string {
	action = strings.TrimSpace(action)
	previewID = strings.TrimSpace(previewID)
	if action == "" || previewID == "" {
		return ""
	}
	return "cabinet:catalog_capture:" + action + ":" + previewID
}

func draftNeedsFollowUp(in CaptureInput, draft Draft) DraftNeedsFollowUpError {
	missing := []string{}
	if strings.TrimSpace(draft.PartNumber) == "" {
		missing = append(missing, "barcode_or_part_number")
	}
	if strings.TrimSpace(draft.Title) == "" {
		missing = append(missing, "title_or_description")
	}
	reason := "missing enough catalog identity to create a safe preview"
	if strings.TrimSpace(in.Text) != "" || len(in.Media) > 0 {
		reason = "ambiguous Telegram capture needs one more identifying detail before Cabinet can create a safe preview"
	}
	return DraftNeedsFollowUpError{
		Reason:        reason,
		MissingFields: missing,
		Reply: TelegramReply{
			Text:              "I need one more detail before I can draft this safely. Reply with a barcode, part number, or clearer item title and I will prepare a Cabinet review draft.",
			ConfirmationState: "follow_up_required",
			Actions:           []string{"reply_with_barcode", "reply_with_part_number", "reply_with_item_title"},
			ActionButtons: []TelegramReplyButton{
				{Label: "Send barcode", Kind: "reply", Action: "reply_with_barcode"},
				{Label: "Send part number", Kind: "reply", Action: "reply_with_part_number"},
				{Label: "Send item title", Kind: "reply", Action: "reply_with_item_title"},
			},
		},
	}
}

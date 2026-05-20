package telegramcapture

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/collectors-tech/cabinet/internal/chat"
)

var ErrUnauthorizedSender = errors.New("telegram sender is not authorized for Cabinet capture")

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
	Text              string   `json:"text"`
	ReviewURL         string   `json:"review_url"`
	ConfirmationState string   `json:"confirmation_state"`
	Actions           []string `json:"actions"`
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
		return CaptureResult{}, fmt.Errorf("telegram catalog capture requires a draft part_number and title before preview")
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
	telegramReply := telegramReply(draft, reviewURL)
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

func telegramReply(draft Draft, reviewURL string) TelegramReply {
	title := strings.TrimSpace(draft.Title)
	if title == "" {
		title = "catalog draft"
	}
	return TelegramReply{
		Text:              "Draft ready for review: " + title + ". Open Cabinet to confirm or cancel before anything is added to inventory.",
		ReviewURL:         reviewURL,
		ConfirmationState: "preview_required",
		Actions:           []string{"open_cabinet_review", "confirm_in_cabinet", "cancel_in_cabinet"},
	}
}

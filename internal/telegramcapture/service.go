package telegramcapture

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/collectors-tech/cabinet/internal/chat"
	"github.com/collectors-tech/cabinet/internal/media"
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
	GetActionPreview(ctx context.Context, profileID, previewID string) (chat.ActionPreview, error)
	ApplyAction(ctx context.Context, in chat.ApplyActionInput) (chat.ApplyActionResult, error)
	CancelAction(ctx context.Context, in chat.ApplyActionInput) (chat.ApplyActionResult, error)
	CreateWorkflowRun(ctx context.Context, in chat.CreateWorkflowRunInput) (chat.WorkflowRun, error)
	UpdateWorkflowRun(ctx context.Context, in chat.UpdateWorkflowRunInput) (chat.WorkflowRun, error)
	CreateInboxItem(ctx context.Context, item chat.InboxItem) (chat.InboxItem, error)
}

type MediaAttachmentService interface {
	SaveWorkspaceAttachment(ctx context.Context, profileID, threadID, filename, mimeType string, src io.Reader) (media.WorkspaceAttachment, error)
}

type Service struct {
	authorizer Authorizer
	chat       ChatService
	media      MediaAttachmentService
}

func NewService(authorizer Authorizer, chatSvc ChatService) *Service {
	return &Service{authorizer: authorizer, chat: chatSvc}
}

func NewServiceWithMedia(authorizer Authorizer, chatSvc ChatService, mediaSvc MediaAttachmentService) *Service {
	return &Service{authorizer: authorizer, chat: chatSvc, media: mediaSvc}
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

type CallbackInput struct {
	SenderID     string
	ChatID       string
	CallbackData string
}

type CallbackResult struct {
	ProfileID     string                 `json:"profile_id"`
	PreviewID     string                 `json:"preview_id"`
	ThreadID      string                 `json:"thread_id"`
	Action        string                 `json:"action"`
	ApplyResult   chat.ApplyActionResult `json:"apply_result"`
	TelegramReply TelegramReply          `json:"telegram_reply"`
}

type Draft struct {
	PartNumber       string
	Title            string
	Brand            string
	Category         string
	LookupSource     string
	LookupURL        string
	LookupConfidence string
}

type MediaInput struct {
	FileID       string
	FileUniqueID string
	FileSize     int
	Filename     string
	MIMEType     string
	Kind         string
	Reader       io.Reader
}

type WebhookMediaResolver interface {
	ResolveTelegramMedia(ctx context.Context, media MediaInput) (MediaInput, error)
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
	WorkflowRun   chat.WorkflowRun   `json:"workflow_run"`
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
		attachment, err := s.saveAttachment(ctx, profileID, thread.ID, filename, strings.TrimSpace(media.MIMEType), reader)
		if err != nil {
			return CaptureResult{}, err
		}
		attachments = append(attachments, attachment)
		mediaContext = append(mediaContext, map[string]any{
			"attachment_id":  attachment.ID,
			"file_id":        strings.TrimSpace(media.FileID),
			"file_unique_id": strings.TrimSpace(media.FileUniqueID),
			"file_size":      media.FileSize,
			"filename":       attachment.Filename,
			"mime_type":      attachment.MimeType,
			"kind":           strings.TrimSpace(media.Kind),
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
	if lookup := lookupMetadata(draft); len(lookup) > 0 {
		messageContext["lookup"] = lookup
	}
	message, err := s.chat.CreateMessage(ctx, profileID, thread.ID, "user", messageContent(in, draft), messageContext)
	if err != nil {
		return CaptureResult{}, err
	}

	preview, err := s.chat.PreviewAction(ctx, chat.PreviewActionInput{
		ProfileID: profileID,
		ThreadID:  thread.ID,
		Action:    "create_inventory_item",
		Payload:   previewPayload(draft, in.Barcode),
	})
	if err != nil {
		return CaptureResult{}, err
	}

	workflowRun, err := s.chat.CreateWorkflowRun(ctx, chat.CreateWorkflowRunInput{
		ProfileID:         profileID,
		WorkflowID:        "telegram-catalog-capture-preview",
		CapabilityID:      telegramCatalogCaptureCapability(in, draft),
		SourceChannel:     "telegram",
		SourceThreadID:    thread.ID,
		SourceMessageID:   strings.TrimSpace(in.MessageID),
		ConfirmationState: "pending",
		Input: map[string]any{
			"barcode":       strings.TrimSpace(in.Barcode),
			"draft":         previewPayload(draft, in.Barcode),
			"media_count":   len(attachments),
			"grouping_hint": strings.TrimSpace(in.GroupingHint),
		},
		ProviderTrace: map[string]any{
			"provider":       "openai",
			"source_channel": "telegram",
			"mode":           "governed_preview_before_apply",
			"live_provider":  false,
		},
		BulkItems: telegramCatalogCaptureBulkItems(in.Media, attachments),
	})
	if err != nil {
		return CaptureResult{}, err
	}
	workflowRun, err = s.chat.UpdateWorkflowRun(ctx, chat.UpdateWorkflowRunInput{
		ProfileID:         profileID,
		RunID:             workflowRun.ID,
		Status:            "completed",
		ConfirmationState: "pending",
		ProviderTrace:     workflowRun.ProviderTrace,
		Result: map[string]any{
			"preview_id":         preview.ID,
			"thread_id":          thread.ID,
			"confirmation_state": "preview_required",
			"review_url":         reviewURL(profileID, thread.ID, preview.ID),
		},
		BulkItems: workflowRun.BulkItems,
	})
	if err != nil {
		return CaptureResult{}, err
	}

	reviewURL := reviewURL(profileID, thread.ID, preview.ID)
	telegramReply := telegramReply(draft, reviewURL, preview.ID)
	inboxMetadata := map[string]any{
		"preview_id":         preview.ID,
		"workflow_run_id":    workflowRun.ID,
		"capability_id":      workflowRun.CapabilityID,
		"source_channel":     "telegram",
		"source_chat_id":     chatID,
		"source_sender_id":   senderID,
		"source_message_id":  strings.TrimSpace(in.MessageID),
		"attachment_count":   len(attachments),
		"media":              mediaContext,
		"confirmation_state": "preview_required",
		"review_url":         reviewURL,
		"telegram_reply":     telegramReply,
	}
	if len(in.SourceMetadata) > 0 {
		inboxMetadata["source_metadata"] = in.SourceMetadata
	}
	if lookup := lookupMetadata(draft); len(lookup) > 0 {
		inboxMetadata["lookup"] = lookup
	}
	inboxItem, err := s.chat.CreateInboxItem(ctx, chat.InboxItem{
		ProfileID: profileID,
		ThreadID:  thread.ID,
		Source:    "telegram_catalog_capture",
		Status:    "queued",
		Title:     "Telegram catalog draft ready",
		Summary:   "Review and confirm the Telegram catalog draft before it can mutate Cabinet inventory.",
		Metadata:  inboxMetadata,
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
		WorkflowRun:   workflowRun,
		InboxItem:     inboxItem,
		TelegramReply: telegramReply,
	}, nil
}

func (s *Service) saveAttachment(ctx context.Context, profileID, threadID, filename, mimeType string, reader io.Reader) (chat.Attachment, error) {
	if s.media == nil {
		return s.chat.SaveAttachment(ctx, profileID, threadID, filename, mimeType, reader)
	}
	attachment, err := s.media.SaveWorkspaceAttachment(ctx, profileID, threadID, filename, mimeType, reader)
	if err != nil {
		return chat.Attachment{}, err
	}
	return chat.Attachment{
		ID:        attachment.ID,
		ProfileID: attachment.ProfileID,
		ThreadID:  attachment.ThreadID,
		Filename:  attachment.Filename,
		MimeType:  attachment.MimeType,
		SizeBytes: attachment.SizeBytes,
		Path:      attachment.Path,
		CreatedAt: attachment.CreatedAt,
	}, nil
}

func (s *Service) HandleCatalogCaptureCallback(ctx context.Context, in CallbackInput) (CallbackResult, error) {
	if s == nil || s.authorizer == nil || s.chat == nil {
		return CallbackResult{}, fmt.Errorf("telegram capture service is not configured")
	}
	senderID := strings.TrimSpace(in.SenderID)
	chatID := strings.TrimSpace(in.ChatID)
	if senderID == "" || chatID == "" {
		return CallbackResult{}, fmt.Errorf("telegram sender_id and chat_id are required")
	}
	action, previewID, err := parseTelegramCallbackData(in.CallbackData)
	if err != nil {
		return CallbackResult{}, err
	}
	authorized, err := s.authorizer.AuthorizeTelegramCapture(ctx, senderID, chatID)
	if err != nil {
		return CallbackResult{}, err
	}
	profileID := strings.TrimSpace(authorized.ProfileID)
	if profileID == "" {
		return CallbackResult{}, ErrUnauthorizedSender
	}
	preview, err := s.chat.GetActionPreview(ctx, profileID, previewID)
	if err != nil {
		return CallbackResult{}, err
	}
	applyInput := chat.ApplyActionInput{ProfileID: profileID, ThreadID: preview.ThreadID, PreviewID: preview.ID, Confirm: action == "confirm"}
	var applyResult chat.ApplyActionResult
	switch action {
	case "confirm":
		applyResult, err = s.chat.ApplyAction(ctx, applyInput)
	case "cancel":
		applyResult, err = s.chat.CancelAction(ctx, applyInput)
	default:
		err = fmt.Errorf("unsupported telegram callback action: %s", action)
	}
	if err != nil {
		return CallbackResult{}, err
	}
	return CallbackResult{
		ProfileID:   profileID,
		PreviewID:   preview.ID,
		ThreadID:    preview.ThreadID,
		Action:      action,
		ApplyResult: applyResult,
		TelegramReply: TelegramReply{
			Text:              telegramCallbackReplyText(action, applyResult),
			ConfirmationState: telegramCallbackConfirmationState(action),
			Actions:           []string{"open_cabinet_review"},
			ReviewURL:         reviewURL(profileID, preview.ThreadID, preview.ID),
			ActionButtons: []TelegramReplyButton{
				{Label: "Open Cabinet review", Kind: "url", Action: "open_cabinet_review", URL: reviewURL(profileID, preview.ThreadID, preview.ID)},
			},
		},
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

func CaptureInputFromWebhookUpdateWithMedia(ctx context.Context, update WebhookUpdate, draft Draft, resolver WebhookMediaResolver) (CaptureInput, error) {
	input, err := CaptureInputFromWebhookUpdate(update, draft)
	if err != nil || resolver == nil || len(input.Media) == 0 {
		return input, err
	}
	resolvedMedia := make([]MediaInput, 0, len(input.Media))
	for _, media := range input.Media {
		resolved, err := resolver.ResolveTelegramMedia(ctx, media)
		if err != nil {
			return CaptureInput{}, err
		}
		resolvedMedia = append(resolvedMedia, mergeResolvedMedia(media, resolved))
	}
	input.Media = resolvedMedia
	return input, nil
}

func GroupCaptureInputs(inputs []CaptureInput) []CaptureInput {
	groups := make(map[string]int)
	out := make([]CaptureInput, 0, len(inputs))
	for _, input := range inputs {
		groupKey := captureGroupKey(input)
		if groupKey == "" {
			out = append(out, input)
			continue
		}
		if index, ok := groups[groupKey]; ok {
			out[index] = mergeCaptureInputGroup(out[index], input)
			continue
		}
		groups[groupKey] = len(out)
		out = append(out, normalizeGroupedCaptureInput(input))
	}
	return out
}

func captureGroupKey(input CaptureInput) string {
	senderID := strings.TrimSpace(input.SenderID)
	chatID := strings.TrimSpace(input.ChatID)
	groupingHint := strings.TrimSpace(input.GroupingHint)
	if senderID == "" || chatID == "" || groupingHint == "" {
		return ""
	}
	return senderID + "|" + chatID + "|" + groupingHint
}

func normalizeGroupedCaptureInput(input CaptureInput) CaptureInput {
	input.SenderID = strings.TrimSpace(input.SenderID)
	input.ChatID = strings.TrimSpace(input.ChatID)
	input.MessageID = strings.TrimSpace(input.MessageID)
	input.Text = strings.TrimSpace(input.Text)
	input.Barcode = strings.TrimSpace(input.Barcode)
	input.GroupingHint = strings.TrimSpace(input.GroupingHint)
	metadata := cloneMetadata(input.SourceMetadata)
	metadata["media_group_id"] = input.GroupingHint
	metadata["grouped_message_ids"] = appendStringValue(nil, input.MessageID)
	metadata["message_count"] = 1
	metadata["payload_type"] = groupedPayloadType(metadata["payload_type"])
	input.SourceMetadata = metadata
	return input
}

func mergeCaptureInputGroup(base, next CaptureInput) CaptureInput {
	next = normalizeGroupedCaptureInput(next)
	base.Text = joinNonEmpty("\n", base.Text, next.Text)
	if strings.TrimSpace(base.Barcode) == "" {
		base.Barcode = strings.TrimSpace(next.Barcode)
	}
	base.Draft = mergeDraft(base.Draft, next.Draft)
	base.Media = append(base.Media, next.Media...)
	base.SourceMetadata = mergeGroupedMetadata(base.SourceMetadata, next.SourceMetadata)
	return base
}

func mergeDraft(base, next Draft) Draft {
	if strings.TrimSpace(base.PartNumber) == "" {
		base.PartNumber = strings.TrimSpace(next.PartNumber)
	}
	if strings.TrimSpace(base.Title) == "" {
		base.Title = strings.TrimSpace(next.Title)
	}
	if strings.TrimSpace(base.Brand) == "" {
		base.Brand = strings.TrimSpace(next.Brand)
	}
	if strings.TrimSpace(base.Category) == "" {
		base.Category = strings.TrimSpace(next.Category)
	}
	if strings.TrimSpace(base.LookupSource) == "" {
		base.LookupSource = strings.TrimSpace(next.LookupSource)
	}
	if strings.TrimSpace(base.LookupURL) == "" {
		base.LookupURL = strings.TrimSpace(next.LookupURL)
	}
	if strings.TrimSpace(base.LookupConfidence) == "" {
		base.LookupConfidence = strings.TrimSpace(next.LookupConfidence)
	}
	return base
}

func mergeGroupedMetadata(base, next map[string]any) map[string]any {
	out := cloneMetadata(base)
	for _, messageID := range stringSlice(next["grouped_message_ids"]) {
		out["grouped_message_ids"] = appendStringValue(out["grouped_message_ids"], messageID)
	}
	out["message_count"] = len(stringSlice(out["grouped_message_ids"]))
	if updateIDs := appendUniqueMetadataValues(out["grouped_update_ids"], base["update_id"], next["update_id"]); len(updateIDs) > 0 {
		out["grouped_update_ids"] = updateIDs
	}
	out["payload_type"] = groupedPayloadType(joinNonEmpty("+", fmt.Sprint(base["payload_type"]), fmt.Sprint(next["payload_type"])))
	return out
}

func cloneMetadata(metadata map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range metadata {
		out[key] = value
	}
	return out
}

func appendMetadataValues(existing any, values ...any) []any {
	out := []any{}
	if current, ok := existing.([]any); ok {
		out = append(out, current...)
	}
	for _, value := range values {
		if value == nil {
			continue
		}
		out = append(out, value)
	}
	return out
}

func appendUniqueMetadataValues(existing any, values ...any) []any {
	out := appendMetadataValues(existing, values...)
	seen := map[string]bool{}
	unique := make([]any, 0, len(out))
	for _, value := range out {
		key := fmt.Sprintf("%T:%v", value, value)
		if seen[key] {
			continue
		}
		seen[key] = true
		unique = append(unique, value)
	}
	return unique
}

func stringSlice(value any) []string {
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			out = appendStringValue(out, fmt.Sprint(item))
		}
		return out
	default:
		return nil
	}
}

func appendStringValue(existing any, value string) []string {
	out := stringSlice(existing)
	value = strings.TrimSpace(value)
	if value == "" {
		return out
	}
	for _, current := range out {
		if current == value {
			return out
		}
	}
	return append(out, value)
}

func joinNonEmpty(separator string, values ...string) string {
	parts := []string{}
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			parts = append(parts, value)
		}
	}
	return strings.Join(parts, separator)
}

func groupedPayloadType(value any) string {
	parts := []string{"telegram_album"}
	for _, part := range strings.Split(fmt.Sprint(value), "+") {
		part = strings.TrimSpace(part)
		if part == "" || part == "<nil>" || part == "telegram_album" {
			continue
		}
		parts = append(parts, part)
	}
	return strings.Join(parts, "+")
}

func mergeResolvedMedia(original, resolved MediaInput) MediaInput {
	out := original
	if fileID := strings.TrimSpace(resolved.FileID); fileID != "" {
		out.FileID = fileID
	}
	if filename := strings.TrimSpace(resolved.Filename); filename != "" {
		out.Filename = filename
	}
	if fileUniqueID := strings.TrimSpace(resolved.FileUniqueID); fileUniqueID != "" {
		out.FileUniqueID = fileUniqueID
	}
	if resolved.FileSize > 0 {
		out.FileSize = resolved.FileSize
	}
	if mimeType := strings.TrimSpace(resolved.MIMEType); mimeType != "" {
		out.MIMEType = mimeType
	}
	if kind := strings.TrimSpace(resolved.Kind); kind != "" {
		out.Kind = kind
	}
	if resolved.Reader != nil {
		out.Reader = resolved.Reader
	}
	return out
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
			FileID:       strings.TrimSpace(photo.FileID),
			FileUniqueID: strings.TrimSpace(photo.FileUniqueID),
			FileSize:     photo.FileSize,
			Filename:     filename + ".jpg",
			MIMEType:     "image/jpeg",
			Kind:         "photo",
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
			FileID:       strings.TrimSpace(doc.FileID),
			FileUniqueID: strings.TrimSpace(doc.FileUniqueID),
			FileSize:     doc.FileSize,
			Filename:     filename,
			MIMEType:     strings.TrimSpace(doc.MIMEType),
			Kind:         "document_image",
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
	draft.LookupSource = strings.TrimSpace(draft.LookupSource)
	draft.LookupURL = strings.TrimSpace(draft.LookupURL)
	draft.LookupConfidence = strings.TrimSpace(draft.LookupConfidence)
	barcode = strings.TrimSpace(barcode)
	if draft.PartNumber == "" {
		draft.PartNumber = barcode
	}
	if draft.Title == "" && barcode != "" {
		draft.Title = "Barcode " + barcode
	}
	return draft
}

func previewPayload(draft Draft, barcode string) map[string]any {
	payload := map[string]any{
		"part_number": strings.TrimSpace(draft.PartNumber),
		"title":       strings.TrimSpace(draft.Title),
		"brand":       strings.TrimSpace(draft.Brand),
		"category":    strings.TrimSpace(draft.Category),
		"source":      "telegram",
		"barcode":     strings.TrimSpace(barcode),
	}
	if lookup := lookupMetadata(draft); len(lookup) > 0 {
		payload["lookup"] = lookup
	}
	return payload
}

func lookupMetadata(draft Draft) map[string]any {
	lookup := map[string]any{}
	if source := strings.TrimSpace(draft.LookupSource); source != "" {
		lookup["source"] = source
	}
	if url := strings.TrimSpace(draft.LookupURL); url != "" {
		lookup["url"] = url
	}
	if confidence := strings.TrimSpace(draft.LookupConfidence); confidence != "" {
		lookup["confidence"] = confidence
	}
	return lookup
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

func telegramCatalogCaptureCapability(in CaptureInput, draft Draft) string {
	if len(in.Media) > 0 {
		return "catalog_add_from_photo"
	}
	if strings.TrimSpace(in.Barcode) != "" || strings.TrimSpace(draft.PartNumber) != "" {
		return "catalog_add_from_barcode"
	}
	return "catalog_add_from_text"
}

func telegramCatalogCaptureBulkItems(media []MediaInput, attachments []chat.Attachment) []map[string]any {
	if len(media) == 0 && len(attachments) == 0 {
		return nil
	}
	count := len(media)
	if len(attachments) > count {
		count = len(attachments)
	}
	items := make([]map[string]any, 0, count)
	for i := 0; i < count; i++ {
		item := map[string]any{
			"item_key": fmt.Sprintf("telegram-media-%d", i+1),
			"status":   "completed",
		}
		if i < len(media) {
			item["file_id"] = strings.TrimSpace(media[i].FileID)
			item["file_unique_id"] = strings.TrimSpace(media[i].FileUniqueID)
			item["kind"] = strings.TrimSpace(media[i].Kind)
		}
		if i < len(attachments) {
			item["attachment_id"] = attachments[i].ID
			item["filename"] = attachments[i].Filename
		}
		items = append(items, item)
	}
	return items
}

func parseTelegramCallbackData(raw string) (string, string, error) {
	parts := strings.Split(strings.TrimSpace(raw), ":")
	if len(parts) != 4 || parts[0] != "cabinet" || parts[1] != "catalog_capture" {
		return "", "", fmt.Errorf("invalid telegram catalog capture callback data")
	}
	action := strings.TrimSpace(parts[2])
	previewID := strings.TrimSpace(parts[3])
	if (action != "confirm" && action != "cancel") || previewID == "" {
		return "", "", fmt.Errorf("invalid telegram catalog capture callback action")
	}
	return action, previewID, nil
}

func telegramCallbackReplyText(action string, result chat.ApplyActionResult) string {
	if action == "confirm" {
		if strings.TrimSpace(result.ItemID) != "" {
			return "Confirmed. Cabinet added the catalog item and kept the Telegram capture audit trail linked to the review thread."
		}
		return "Confirmed. Cabinet applied the catalog capture after explicit approval."
	}
	return "Cancelled. Cabinet left the catalog draft unapplied and kept the Telegram capture audit trail for review."
}

func telegramCallbackConfirmationState(action string) string {
	if action == "confirm" {
		return "confirmed"
	}
	return "cancelled"
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

package chat

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	db      *sql.DB
	dataDir string
}

type Thread struct {
	ID        string         `json:"id"`
	ProfileID string         `json:"profile_id"`
	Title     string         `json:"title"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
}

type Message struct {
	ID          string         `json:"id"`
	ProfileID   string         `json:"profile_id"`
	ThreadID    string         `json:"thread_id"`
	Role        string         `json:"role"`
	Content     string         `json:"content"`
	Attachments string         `json:"attachments_json"`
	Context     map[string]any `json:"context,omitempty"`
	CreatedAt   string         `json:"created_at"`
}

type Attachment struct {
	ID        string `json:"id"`
	ProfileID string `json:"profile_id"`
	ThreadID  string `json:"thread_id"`
	Filename  string `json:"filename"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
	Path      string `json:"path"`
	CreatedAt string `json:"created_at"`
}

type InboxItem struct {
	ID        string         `json:"id"`
	ProfileID string         `json:"profile_id"`
	ThreadID  string         `json:"thread_id"`
	Source    string         `json:"source"`
	Status    string         `json:"status"`
	Title     string         `json:"title"`
	Summary   string         `json:"summary"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
}

type PreviewActionInput struct {
	ProfileID string         `json:"profile_id"`
	ThreadID  string         `json:"thread_id"`
	Action    string         `json:"action"`
	Payload   map[string]any `json:"payload"`
}

type ActionPreview struct {
	ID        string         `json:"id"`
	ProfileID string         `json:"profile_id"`
	ThreadID  string         `json:"thread_id"`
	Action    string         `json:"action"`
	Payload   map[string]any `json:"payload,omitempty"`
	Status    string         `json:"status"`
	CreatedAt string         `json:"created_at"`
	AppliedAt string         `json:"applied_at,omitempty"`
}

type ApplyActionInput struct {
	ProfileID string `json:"profile_id"`
	ThreadID  string `json:"thread_id"`
	PreviewID string `json:"preview_id"`
	Confirm   bool   `json:"confirm"`
}

type ApplyActionResult struct {
	Applied        bool   `json:"applied"`
	Action         string `json:"action"`
	ItemID         string `json:"item_id,omitempty"`
	WishlistID     string `json:"wishlist_id,omitempty"`
	CollectionName string `json:"collection_name,omitempty"`
	PartNumber     string `json:"part_number,omitempty"`
	Title          string `json:"title,omitempty"`
	PreviewID      string `json:"preview_id"`
}

type WorkflowRun struct {
	ID                string           `json:"id"`
	ProfileID         string           `json:"profile_id"`
	WorkflowID        string           `json:"workflow_id"`
	CapabilityID      string           `json:"capability_id"`
	SourceChannel     string           `json:"source_channel"`
	SourceThreadID    string           `json:"source_thread_id,omitempty"`
	SourceMessageID   string           `json:"source_message_id,omitempty"`
	Status            string           `json:"status"`
	Input             map[string]any   `json:"input,omitempty"`
	ProviderTrace     map[string]any   `json:"provider_trace,omitempty"`
	Result            map[string]any   `json:"result,omitempty"`
	Error             map[string]any   `json:"error,omitempty"`
	ConfirmationState string           `json:"confirmation_state"`
	BulkItems         []map[string]any `json:"bulk_items,omitempty"`
	CreatedAt         string           `json:"created_at"`
	UpdatedAt         string           `json:"updated_at"`
	StartedAt         string           `json:"started_at,omitempty"`
	CompletedAt       string           `json:"completed_at,omitempty"`
}

type CreateWorkflowRunInput struct {
	ProfileID         string           `json:"profile_id"`
	WorkflowID        string           `json:"workflow_id"`
	CapabilityID      string           `json:"capability_id"`
	SourceChannel     string           `json:"source_channel"`
	SourceThreadID    string           `json:"source_thread_id"`
	SourceMessageID   string           `json:"source_message_id"`
	Input             map[string]any   `json:"input"`
	ProviderTrace     map[string]any   `json:"provider_trace"`
	ConfirmationState string           `json:"confirmation_state"`
	BulkItems         []map[string]any `json:"bulk_items"`
}

type UpdateWorkflowRunInput struct {
	ProfileID         string           `json:"profile_id"`
	RunID             string           `json:"run_id"`
	Status            string           `json:"status"`
	ProviderTrace     map[string]any   `json:"provider_trace"`
	Result            map[string]any   `json:"result"`
	Error             map[string]any   `json:"error"`
	ConfirmationState string           `json:"confirmation_state"`
	BulkItems         []map[string]any `json:"bulk_items"`
}

func NewService(db *sql.DB, dataDir string) *Service {
	return &Service{db: db, dataDir: dataDir}
}

func (s *Service) CreateThread(ctx context.Context, profileID, title string, metadata map[string]any) (Thread, error) {
	profileID = strings.TrimSpace(profileID)
	title = strings.TrimSpace(title)
	if profileID == "" {
		return Thread{}, fmt.Errorf("profile_id is required")
	}
	if title == "" {
		return Thread{}, fmt.Errorf("title is required")
	}
	id := uuid.NewString()
	metadataJSON := marshalContextJSON(metadata)
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO chat_threads(id, profile_id, title, metadata_json)
		VALUES (?, ?, ?, ?)
	`, id, profileID, title, metadataJSON); err != nil {
		return Thread{}, fmt.Errorf("create thread: %w", err)
	}
	return s.GetThread(ctx, profileID, id)
}

func (s *Service) ListThreads(ctx context.Context, profileID string) ([]Thread, error) {
	profileID = strings.TrimSpace(profileID)
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, profile_id, title, metadata_json, created_at, updated_at
		FROM chat_threads
		WHERE profile_id = ?
		ORDER BY updated_at DESC, created_at DESC
	`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Thread
	for rows.Next() {
		var t Thread
		var metadataJSON string
		if err := rows.Scan(&t.ID, &t.ProfileID, &t.Title, &metadataJSON, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		t.Metadata = parseContextJSON(metadataJSON)
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Service) GetThread(ctx context.Context, profileID, threadID string) (Thread, error) {
	var t Thread
	var metadataJSON string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, profile_id, title, metadata_json, created_at, updated_at
		FROM chat_threads
		WHERE id = ? AND profile_id = ?
	`, strings.TrimSpace(threadID), strings.TrimSpace(profileID)).Scan(&t.ID, &t.ProfileID, &t.Title, &metadataJSON, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return Thread{}, fmt.Errorf("thread not found")
		}
		return Thread{}, err
	}
	t.Metadata = parseContextJSON(metadataJSON)
	return t, nil
}

func (s *Service) CreateMessage(ctx context.Context, profileID, threadID, role, content string, messageContext map[string]any) (Message, error) {
	profileID = strings.TrimSpace(profileID)
	threadID = strings.TrimSpace(threadID)
	role = strings.TrimSpace(strings.ToLower(role))
	content = strings.TrimSpace(content)
	if profileID == "" || threadID == "" {
		return Message{}, fmt.Errorf("profile_id and thread_id are required")
	}
	if role != "user" && role != "assistant" && role != "system" {
		return Message{}, fmt.Errorf("invalid role")
	}
	if content == "" {
		return Message{}, fmt.Errorf("content is required")
	}
	if _, err := s.GetThread(ctx, profileID, threadID); err != nil {
		return Message{}, err
	}
	id := uuid.NewString()
	contextJSON := marshalContextJSON(messageContext)
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO chat_messages(id, profile_id, thread_id, role, content, attachments_json, context_json)
		VALUES (?, ?, ?, ?, ?, '[]', ?)
	`, id, profileID, threadID, role, content, contextJSON); err != nil {
		return Message{}, fmt.Errorf("create message: %w", err)
	}
	_, _ = s.db.ExecContext(ctx, `UPDATE chat_threads SET updated_at = CURRENT_TIMESTAMP WHERE id = ?`, threadID)
	return s.getMessage(ctx, profileID, id)
}

func (s *Service) ListMessages(ctx context.Context, profileID, threadID string) ([]Message, error) {
	profileID = strings.TrimSpace(profileID)
	threadID = strings.TrimSpace(threadID)
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, profile_id, thread_id, role, content, attachments_json, context_json, created_at
		FROM chat_messages
		WHERE profile_id = ? AND thread_id = ?
		ORDER BY created_at ASC
	`, profileID, threadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Message
	for rows.Next() {
		var m Message
		var contextJSON string
		if err := rows.Scan(&m.ID, &m.ProfileID, &m.ThreadID, &m.Role, &m.Content, &m.Attachments, &contextJSON, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.Context = parseContextJSON(contextJSON)
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Service) getMessage(ctx context.Context, profileID, messageID string) (Message, error) {
	var m Message
	var contextJSON string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, profile_id, thread_id, role, content, attachments_json, context_json, created_at
		FROM chat_messages
		WHERE id = ? AND profile_id = ?
	`, strings.TrimSpace(messageID), strings.TrimSpace(profileID)).Scan(&m.ID, &m.ProfileID, &m.ThreadID, &m.Role, &m.Content, &m.Attachments, &contextJSON, &m.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return Message{}, fmt.Errorf("message not found")
		}
		return Message{}, err
	}
	m.Context = parseContextJSON(contextJSON)
	return m, nil
}

func marshalContextJSON(messageContext map[string]any) string {
	if len(messageContext) == 0 {
		return "{}"
	}
	raw, err := json.Marshal(messageContext)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func parseContextJSON(raw string) map[string]any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil || out == nil {
		return map[string]any{}
	}
	return out
}

func marshalBulkItemsJSON(items []map[string]any) string {
	if len(items) == 0 {
		return "[]"
	}
	raw, err := json.Marshal(items)
	if err != nil {
		return "[]"
	}
	return string(raw)
}

func parseBulkItemsJSON(raw string) []map[string]any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []map[string]any{}
	}
	var out []map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil || out == nil {
		return []map[string]any{}
	}
	return out
}

func normalizeWorkflowStatus(status string) (string, error) {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case "":
		return "queued", nil
	case "queued", "running", "needs_input", "completed", "failed", "cancelled":
		return strings.TrimSpace(strings.ToLower(status)), nil
	default:
		return "", fmt.Errorf("unsupported workflow status: %s", status)
	}
}

func normalizeConfirmationState(state string) string {
	switch strings.TrimSpace(strings.ToLower(state)) {
	case "required", "pending", "confirmed", "cancelled", "not_required":
		return strings.TrimSpace(strings.ToLower(state))
	default:
		return "not_required"
	}
}

func (s *Service) CreateWorkflowRun(ctx context.Context, in CreateWorkflowRunInput) (WorkflowRun, error) {
	in.ProfileID = strings.TrimSpace(in.ProfileID)
	in.WorkflowID = strings.TrimSpace(in.WorkflowID)
	in.CapabilityID = strings.TrimSpace(in.CapabilityID)
	in.SourceChannel = strings.TrimSpace(in.SourceChannel)
	in.SourceThreadID = strings.TrimSpace(in.SourceThreadID)
	in.SourceMessageID = strings.TrimSpace(in.SourceMessageID)
	if in.ProfileID == "" || in.WorkflowID == "" || in.CapabilityID == "" {
		return WorkflowRun{}, fmt.Errorf("profile_id, workflow_id and capability_id are required")
	}
	if in.SourceChannel == "" {
		in.SourceChannel = "in_app_chat"
	}
	status, err := normalizeWorkflowStatus("queued")
	if err != nil {
		return WorkflowRun{}, err
	}
	id := uuid.NewString()
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO assistant_workflow_runs(
			id, profile_id, workflow_id, capability_id, source_channel, source_thread_id, source_message_id,
			status, input_json, provider_trace_json, confirmation_state, bulk_items_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, in.ProfileID, in.WorkflowID, in.CapabilityID, in.SourceChannel, in.SourceThreadID, in.SourceMessageID,
		status, marshalContextJSON(in.Input), marshalContextJSON(in.ProviderTrace), normalizeConfirmationState(in.ConfirmationState), marshalBulkItemsJSON(in.BulkItems)); err != nil {
		return WorkflowRun{}, fmt.Errorf("create workflow run: %w", err)
	}
	return s.GetWorkflowRun(ctx, in.ProfileID, id)
}

func (s *Service) ListWorkflowRuns(ctx context.Context, profileID, threadID string) ([]WorkflowRun, error) {
	profileID = strings.TrimSpace(profileID)
	threadID = strings.TrimSpace(threadID)
	if profileID == "" {
		return nil, fmt.Errorf("profile_id is required")
	}
	query := `
		SELECT id, profile_id, workflow_id, capability_id, source_channel, source_thread_id, source_message_id,
		       status, input_json, provider_trace_json, result_json, error_json, confirmation_state, bulk_items_json,
		       created_at, updated_at, started_at, completed_at
		FROM assistant_workflow_runs
		WHERE profile_id = ?`
	args := []any{profileID}
	if threadID != "" {
		query += ` AND source_thread_id = ?`
		args = append(args, threadID)
	}
	query += ` ORDER BY updated_at DESC, created_at DESC`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []WorkflowRun
	for rows.Next() {
		run, err := scanWorkflowRun(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, run)
	}
	return out, rows.Err()
}

func (s *Service) GetWorkflowRun(ctx context.Context, profileID, runID string) (WorkflowRun, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, profile_id, workflow_id, capability_id, source_channel, source_thread_id, source_message_id,
		       status, input_json, provider_trace_json, result_json, error_json, confirmation_state, bulk_items_json,
		       created_at, updated_at, started_at, completed_at
		FROM assistant_workflow_runs
		WHERE id = ? AND profile_id = ?
	`, strings.TrimSpace(runID), strings.TrimSpace(profileID))
	run, err := scanWorkflowRun(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return WorkflowRun{}, fmt.Errorf("workflow run not found")
		}
		return WorkflowRun{}, err
	}
	return run, nil
}

func (s *Service) UpdateWorkflowRun(ctx context.Context, in UpdateWorkflowRunInput) (WorkflowRun, error) {
	in.ProfileID = strings.TrimSpace(in.ProfileID)
	in.RunID = strings.TrimSpace(in.RunID)
	if in.ProfileID == "" || in.RunID == "" {
		return WorkflowRun{}, fmt.Errorf("profile_id and run_id are required")
	}
	status, err := normalizeWorkflowStatus(in.Status)
	if err != nil {
		return WorkflowRun{}, err
	}
	confirmationState := normalizeConfirmationState(in.ConfirmationState)
	if confirmationState == "not_required" {
		current, err := s.GetWorkflowRun(ctx, in.ProfileID, in.RunID)
		if err != nil {
			return WorkflowRun{}, err
		}
		confirmationState = current.ConfirmationState
	}
	startedAtExpr := "started_at"
	if status == "running" {
		startedAtExpr = "COALESCE(NULLIF(started_at, ''), CURRENT_TIMESTAMP)"
	}
	completedAtExpr := "completed_at"
	if status == "completed" || status == "failed" || status == "cancelled" {
		completedAtExpr = "COALESCE(NULLIF(completed_at, ''), CURRENT_TIMESTAMP)"
	}
	_, err = s.db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE assistant_workflow_runs
		SET status = ?, provider_trace_json = ?, result_json = ?, error_json = ?, confirmation_state = ?,
		    bulk_items_json = ?, updated_at = CURRENT_TIMESTAMP, started_at = %s, completed_at = %s
		WHERE id = ? AND profile_id = ?
	`, startedAtExpr, completedAtExpr),
		status, marshalContextJSON(in.ProviderTrace), marshalContextJSON(in.Result), marshalContextJSON(in.Error),
		confirmationState, marshalBulkItemsJSON(in.BulkItems), in.RunID, in.ProfileID)
	if err != nil {
		return WorkflowRun{}, fmt.Errorf("update workflow run: %w", err)
	}
	return s.GetWorkflowRun(ctx, in.ProfileID, in.RunID)
}

type workflowRunScanner interface {
	Scan(dest ...any) error
}

func scanWorkflowRun(scanner workflowRunScanner) (WorkflowRun, error) {
	var run WorkflowRun
	var inputJSON, providerTraceJSON, resultJSON, errorJSON, bulkItemsJSON string
	err := scanner.Scan(
		&run.ID, &run.ProfileID, &run.WorkflowID, &run.CapabilityID, &run.SourceChannel, &run.SourceThreadID, &run.SourceMessageID,
		&run.Status, &inputJSON, &providerTraceJSON, &resultJSON, &errorJSON, &run.ConfirmationState, &bulkItemsJSON,
		&run.CreatedAt, &run.UpdatedAt, &run.StartedAt, &run.CompletedAt,
	)
	if err != nil {
		return WorkflowRun{}, err
	}
	run.Input = parseContextJSON(inputJSON)
	run.ProviderTrace = parseContextJSON(providerTraceJSON)
	run.Result = parseContextJSON(resultJSON)
	run.Error = parseContextJSON(errorJSON)
	run.BulkItems = parseBulkItemsJSON(bulkItemsJSON)
	return run, nil
}

func (s *Service) CreateInboxItem(ctx context.Context, item InboxItem) (InboxItem, error) {
	item.ProfileID = strings.TrimSpace(item.ProfileID)
	item.ThreadID = strings.TrimSpace(item.ThreadID)
	item.Source = strings.TrimSpace(item.Source)
	item.Status = strings.TrimSpace(item.Status)
	item.Title = strings.TrimSpace(item.Title)
	item.Summary = strings.TrimSpace(item.Summary)
	if item.ProfileID == "" || item.ThreadID == "" {
		return InboxItem{}, fmt.Errorf("profile_id and thread_id are required")
	}
	if item.Source == "" {
		item.Source = "assistant_handoff"
	}
	if item.Status == "" {
		item.Status = "queued"
	}
	if item.Title == "" {
		item.Title = "Assistant handoff queued"
	}
	if _, err := s.GetThread(ctx, item.ProfileID, item.ThreadID); err != nil {
		return InboxItem{}, err
	}
	item.ID = uuid.NewString()
	metadataJSON := marshalContextJSON(item.Metadata)
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO chat_inbox_items(id, profile_id, thread_id, source, status, title, summary, metadata_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, item.ID, item.ProfileID, item.ThreadID, item.Source, item.Status, item.Title, item.Summary, metadataJSON); err != nil {
		return InboxItem{}, fmt.Errorf("create inbox item: %w", err)
	}
	return s.getInboxItem(ctx, item.ProfileID, item.ID)
}

func (s *Service) ListInboxItems(ctx context.Context, profileID string) ([]InboxItem, error) {
	profileID = strings.TrimSpace(profileID)
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, profile_id, thread_id, source, status, title, summary, metadata_json, created_at, updated_at
		FROM chat_inbox_items
		WHERE profile_id = ?
		ORDER BY updated_at DESC, created_at DESC
	`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []InboxItem
	for rows.Next() {
		var item InboxItem
		var metadataJSON string
		if err := rows.Scan(&item.ID, &item.ProfileID, &item.ThreadID, &item.Source, &item.Status, &item.Title, &item.Summary, &metadataJSON, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.Metadata = parseContextJSON(metadataJSON)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Service) UpdateInboxItemStatus(ctx context.Context, profileID, inboxID, status string) (InboxItem, error) {
	profileID = strings.TrimSpace(profileID)
	inboxID = strings.TrimSpace(inboxID)
	status = strings.TrimSpace(strings.ToLower(status))
	if profileID == "" || inboxID == "" {
		return InboxItem{}, fmt.Errorf("profile_id and inbox_id are required")
	}
	switch status {
	case "queued", "unread", "read", "archived":
	default:
		return InboxItem{}, fmt.Errorf("invalid inbox status")
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE chat_inbox_items
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND profile_id = ?
	`, status, inboxID, profileID)
	if err != nil {
		return InboxItem{}, fmt.Errorf("update inbox item status: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return InboxItem{}, fmt.Errorf("update inbox item status: %w", err)
	}
	if count == 0 {
		return InboxItem{}, fmt.Errorf("inbox item not found")
	}
	return s.getInboxItem(ctx, profileID, inboxID)
}

func (s *Service) getInboxItem(ctx context.Context, profileID, inboxID string) (InboxItem, error) {
	var item InboxItem
	var metadataJSON string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, profile_id, thread_id, source, status, title, summary, metadata_json, created_at, updated_at
		FROM chat_inbox_items
		WHERE id = ? AND profile_id = ?
	`, strings.TrimSpace(inboxID), strings.TrimSpace(profileID)).Scan(&item.ID, &item.ProfileID, &item.ThreadID, &item.Source, &item.Status, &item.Title, &item.Summary, &metadataJSON, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return InboxItem{}, fmt.Errorf("inbox item not found")
		}
		return InboxItem{}, err
	}
	item.Metadata = parseContextJSON(metadataJSON)
	return item, nil
}

func (s *Service) SaveAttachment(ctx context.Context, profileID, threadID, filename, mimeType string, src io.Reader) (Attachment, error) {
	profileID = strings.TrimSpace(profileID)
	threadID = strings.TrimSpace(threadID)
	filename = strings.TrimSpace(filename)
	if profileID == "" || threadID == "" {
		return Attachment{}, fmt.Errorf("profile_id and thread_id are required")
	}
	if filename == "" {
		return Attachment{}, fmt.Errorf("filename is required")
	}
	if _, err := s.GetThread(ctx, profileID, threadID); err != nil {
		return Attachment{}, err
	}
	if err := os.MkdirAll(s.dataDir, 0o755); err != nil {
		return Attachment{}, fmt.Errorf("create attachment dir: %w", err)
	}
	id := uuid.NewString()
	safeName := strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' {
			return '_'
		}
		return r
	}, filename)
	targetPath := filepath.Join(s.dataDir, id+"-"+safeName)
	out, err := os.Create(targetPath)
	if err != nil {
		return Attachment{}, fmt.Errorf("create attachment file: %w", err)
	}
	size, copyErr := io.Copy(out, src)
	closeErr := out.Close()
	if copyErr != nil {
		return Attachment{}, fmt.Errorf("write attachment: %w", copyErr)
	}
	if closeErr != nil {
		return Attachment{}, fmt.Errorf("close attachment: %w", closeErr)
	}
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO chat_attachments(id, profile_id, thread_id, filename, mime_type, size_bytes, stored_path)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, id, profileID, threadID, filename, mimeType, size, targetPath); err != nil {
		return Attachment{}, fmt.Errorf("save attachment: %w", err)
	}
	return s.getAttachment(ctx, profileID, id)
}

func (s *Service) getAttachment(ctx context.Context, profileID, attachmentID string) (Attachment, error) {
	var a Attachment
	err := s.db.QueryRowContext(ctx, `
		SELECT id, profile_id, thread_id, filename, mime_type, size_bytes, stored_path, created_at
		FROM chat_attachments
		WHERE id = ? AND profile_id = ?
	`, strings.TrimSpace(attachmentID), strings.TrimSpace(profileID)).Scan(
		&a.ID, &a.ProfileID, &a.ThreadID, &a.Filename, &a.MimeType, &a.SizeBytes, &a.Path, &a.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Attachment{}, fmt.Errorf("attachment not found")
		}
		return Attachment{}, err
	}
	return a, nil
}

func (s *Service) PreviewAction(ctx context.Context, in PreviewActionInput) (ActionPreview, error) {
	in.ProfileID = strings.TrimSpace(in.ProfileID)
	in.ThreadID = strings.TrimSpace(in.ThreadID)
	in.Action = strings.TrimSpace(in.Action)
	if in.ProfileID == "" || in.ThreadID == "" || in.Action == "" {
		return ActionPreview{}, fmt.Errorf("profile_id, thread_id and action are required")
	}
	if _, err := s.GetThread(ctx, in.ProfileID, in.ThreadID); err != nil {
		return ActionPreview{}, err
	}
	if err := validateActionPayload(in.Action, in.Payload); err != nil {
		return ActionPreview{}, err
	}
	rawPayload, err := json.Marshal(in.Payload)
	if err != nil {
		return ActionPreview{}, fmt.Errorf("marshal payload: %w", err)
	}
	id := uuid.NewString()
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO chat_action_previews(id, profile_id, thread_id, action, payload_json, status)
		VALUES (?, ?, ?, ?, ?, 'previewed')
	`, id, in.ProfileID, in.ThreadID, in.Action, string(rawPayload)); err != nil {
		return ActionPreview{}, fmt.Errorf("create preview: %w", err)
	}
	return s.getPreview(ctx, in.ProfileID, id)
}

func (s *Service) ApplyAction(ctx context.Context, in ApplyActionInput) (ApplyActionResult, error) {
	in.ProfileID = strings.TrimSpace(in.ProfileID)
	in.ThreadID = strings.TrimSpace(in.ThreadID)
	in.PreviewID = strings.TrimSpace(in.PreviewID)
	if in.ProfileID == "" || in.ThreadID == "" || in.PreviewID == "" {
		return ApplyActionResult{}, fmt.Errorf("profile_id, thread_id and preview_id are required")
	}
	if !in.Confirm {
		return ApplyActionResult{}, fmt.Errorf("confirm_required")
	}
	preview, payload, err := s.lookupPendingPreview(ctx, in.ProfileID, in.ThreadID, in.PreviewID)
	if err != nil {
		return ApplyActionResult{}, err
	}
	result := ApplyActionResult{Applied: true, Action: preview.Action, PreviewID: preview.ID}
	switch preview.Action {
	case "create_item_stub", "create_inventory_item":
		itemID, err := s.applyCreateItemStub(ctx, in.ProfileID, payload)
		if err != nil {
			return ApplyActionResult{}, err
		}
		result.ItemID = itemID
		result.PartNumber = trimPayloadString(payload, "part_number")
		result.Title = trimPayloadString(payload, "title")
	case "update_inventory_item":
		itemID, err := s.applyUpdateItem(ctx, in.ProfileID, payload)
		if err != nil {
			return ApplyActionResult{}, err
		}
		result.ItemID = itemID
		result.PartNumber = trimPayloadString(payload, "part_number")
		result.Title = trimPayloadString(payload, "title")
	case "create_wishlist_entry":
		itemID, wishlistID, err := s.applyCreateWishlistEntry(ctx, in.ProfileID, payload)
		if err != nil {
			return ApplyActionResult{}, err
		}
		result.ItemID = itemID
		result.WishlistID = wishlistID
		result.PartNumber = trimPayloadString(payload, "part_number")
		result.Title = trimPayloadString(payload, "title")
	case "assign_collection_item":
		itemID, collectionName, err := s.applyAssignCollectionItem(ctx, in.ProfileID, payload)
		if err != nil {
			return ApplyActionResult{}, err
		}
		result.ItemID = itemID
		result.CollectionName = collectionName
	default:
		return ApplyActionResult{}, fmt.Errorf("unsupported action: %s", preview.Action)
	}
	_, err = s.db.ExecContext(ctx, `
		UPDATE chat_action_previews
		SET status = 'applied', applied_at = CURRENT_TIMESTAMP
		WHERE id = ? AND profile_id = ? AND thread_id = ?
	`, in.PreviewID, in.ProfileID, in.ThreadID)
	if err != nil {
		return ApplyActionResult{}, fmt.Errorf("mark action applied: %w", err)
	}
	_, _ = s.CreateMessage(ctx, in.ProfileID, in.ThreadID, "assistant", applyActionMessage(result), map[string]any{
		"action_result": map[string]any{
			"preview_id":       result.PreviewID,
			"action":           result.Action,
			"item_id":          result.ItemID,
			"wishlist_id":      result.WishlistID,
			"collection_name":  result.CollectionName,
			"part_number":      result.PartNumber,
			"title":            result.Title,
			"confirmation":     "confirmed",
			"mutation_applied": result.Applied,
		},
	})
	return result, nil
}

func (s *Service) CancelAction(ctx context.Context, in ApplyActionInput) (ApplyActionResult, error) {
	in.ProfileID = strings.TrimSpace(in.ProfileID)
	in.ThreadID = strings.TrimSpace(in.ThreadID)
	in.PreviewID = strings.TrimSpace(in.PreviewID)
	if in.ProfileID == "" || in.ThreadID == "" || in.PreviewID == "" {
		return ApplyActionResult{}, fmt.Errorf("profile_id, thread_id and preview_id are required")
	}
	preview, payload, err := s.lookupPendingPreview(ctx, in.ProfileID, in.ThreadID, in.PreviewID)
	if err != nil {
		return ApplyActionResult{}, err
	}
	_, err = s.db.ExecContext(ctx, `
		UPDATE chat_action_previews
		SET status = 'cancelled'
		WHERE id = ? AND profile_id = ? AND thread_id = ?
	`, in.PreviewID, in.ProfileID, in.ThreadID)
	if err != nil {
		return ApplyActionResult{}, fmt.Errorf("mark action cancelled: %w", err)
	}
	result := ApplyActionResult{
		Applied:    false,
		Action:     preview.Action,
		PreviewID:  preview.ID,
		ItemID:     trimPayloadString(payload, "item_id"),
		PartNumber: trimPayloadString(payload, "part_number"),
		Title:      trimPayloadString(payload, "title"),
	}
	_, _ = s.CreateMessage(ctx, in.ProfileID, in.ThreadID, "assistant", cancelActionMessage(result), map[string]any{
		"action_result": map[string]any{
			"preview_id":       result.PreviewID,
			"action":           result.Action,
			"item_id":          result.ItemID,
			"part_number":      result.PartNumber,
			"title":            result.Title,
			"confirmation":     "cancelled",
			"mutation_applied": false,
		},
	})
	return result, nil
}

func (s *Service) lookupPendingPreview(ctx context.Context, profileID, threadID, previewID string) (ActionPreview, map[string]any, error) {
	var preview ActionPreview
	var payloadRaw string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, profile_id, thread_id, action, payload_json, status, created_at, COALESCE(applied_at, '')
		FROM chat_action_previews
		WHERE id = ? AND profile_id = ? AND thread_id = ?
	`, previewID, profileID, threadID).Scan(
		&preview.ID, &preview.ProfileID, &preview.ThreadID, &preview.Action, &payloadRaw, &preview.Status, &preview.CreatedAt, &preview.AppliedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return ActionPreview{}, nil, fmt.Errorf("preview not found")
		}
		return ActionPreview{}, nil, err
	}
	if preview.Status != "previewed" {
		return ActionPreview{}, nil, fmt.Errorf("preview already applied")
	}
	payload := map[string]any{}
	if err := json.Unmarshal([]byte(payloadRaw), &payload); err != nil {
		return ActionPreview{}, nil, fmt.Errorf("decode payload: %w", err)
	}
	return preview, payload, nil
}

func (s *Service) GetActionPreview(ctx context.Context, profileID, previewID string) (ActionPreview, error) {
	return s.getPreview(ctx, profileID, previewID)
}

func (s *Service) getPreview(ctx context.Context, profileID, previewID string) (ActionPreview, error) {
	var preview ActionPreview
	var payloadRaw string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, profile_id, thread_id, action, payload_json, status, created_at, COALESCE(applied_at, '')
		FROM chat_action_previews
		WHERE id = ? AND profile_id = ?
	`, strings.TrimSpace(previewID), strings.TrimSpace(profileID)).Scan(
		&preview.ID, &preview.ProfileID, &preview.ThreadID, &preview.Action, &payloadRaw, &preview.Status, &preview.CreatedAt, &preview.AppliedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return ActionPreview{}, fmt.Errorf("preview not found")
		}
		return ActionPreview{}, err
	}
	if strings.TrimSpace(payloadRaw) != "" {
		payload := map[string]any{}
		if err := json.Unmarshal([]byte(payloadRaw), &payload); err == nil {
			preview.Payload = payload
		}
	}
	return preview, nil
}

func validateActionPayload(action string, payload map[string]any) error {
	switch strings.TrimSpace(action) {
	case "create_item_stub", "create_inventory_item", "create_wishlist_entry":
		partNumber, _ := payload["part_number"].(string)
		title, _ := payload["title"].(string)
		if strings.TrimSpace(partNumber) == "" || strings.TrimSpace(title) == "" {
			return fmt.Errorf("part_number and title are required")
		}
		return nil
	case "assign_collection_item":
		itemID, _ := payload["item_id"].(string)
		collectionName, _ := payload["collection_name"].(string)
		if strings.TrimSpace(itemID) == "" || strings.TrimSpace(collectionName) == "" {
			return fmt.Errorf("item_id and collection_name are required")
		}
		return nil
	case "update_inventory_item":
		itemID, _ := payload["item_id"].(string)
		partNumber, _ := payload["part_number"].(string)
		title, _ := payload["title"].(string)
		brand, _ := payload["brand"].(string)
		category, _ := payload["category"].(string)
		if strings.TrimSpace(itemID) == "" {
			return fmt.Errorf("item_id is required")
		}
		if strings.TrimSpace(partNumber) == "" && strings.TrimSpace(title) == "" && strings.TrimSpace(brand) == "" && strings.TrimSpace(category) == "" {
			return fmt.Errorf("at least one mutable field is required")
		}
		return nil
	default:
		return fmt.Errorf("unsupported action: %s", action)
	}
}

func (s *Service) applyCreateItemStub(ctx context.Context, profileID string, payload map[string]any) (string, error) {
	partNumber, _ := payload["part_number"].(string)
	title, _ := payload["title"].(string)
	brand, _ := payload["brand"].(string)
	category, _ := payload["category"].(string)
	if strings.TrimSpace(brand) == "" {
		brand = "Unknown"
	}
	if strings.TrimSpace(category) == "" {
		category = "General"
	}
	itemID := uuid.NewString()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO canonical_items(
			id, profile_id, brand, category, part_number, title, make, model, year, scale, series, description, tags_json, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, '', '', '', '', '', '', '[]', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, itemID, profileID, strings.TrimSpace(brand), strings.TrimSpace(category), strings.TrimSpace(partNumber), strings.TrimSpace(title))
	if err != nil {
		return "", fmt.Errorf("create stub item: %w", err)
	}
	return itemID, nil
}

func (s *Service) applyUpdateItem(ctx context.Context, profileID string, payload map[string]any) (string, error) {
	itemID, _ := payload["item_id"].(string)
	partNumber, _ := payload["part_number"].(string)
	title, _ := payload["title"].(string)
	brand, _ := payload["brand"].(string)
	category, _ := payload["category"].(string)
	if strings.TrimSpace(brand) == "" {
		brand = "Unknown"
	}
	if strings.TrimSpace(category) == "" {
		category = "General"
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE canonical_items
		SET part_number = COALESCE(NULLIF(?, ''), part_number),
		    title = COALESCE(NULLIF(?, ''), title),
		    brand = COALESCE(NULLIF(?, ''), brand),
		    category = COALESCE(NULLIF(?, ''), category),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND profile_id = ?
	`, strings.TrimSpace(partNumber), strings.TrimSpace(title), strings.TrimSpace(brand), strings.TrimSpace(category), strings.TrimSpace(itemID), strings.TrimSpace(profileID))
	if err != nil {
		return "", fmt.Errorf("update item: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return "", fmt.Errorf("update item rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return "", fmt.Errorf("update item target not found")
	}
	return strings.TrimSpace(itemID), nil
}

func (s *Service) applyCreateWishlistEntry(ctx context.Context, profileID string, payload map[string]any) (string, string, error) {
	itemID, err := s.applyCreateItemStub(ctx, profileID, payload)
	if err != nil {
		return "", "", err
	}
	priority, _ := payload["priority"].(string)
	if strings.TrimSpace(priority) == "" {
		priority = "medium"
	}
	notes, _ := payload["notes"].(string)
	targetPrice := 0.0
	if value, ok := payload["target_price"]; ok {
		switch typed := value.(type) {
		case float64:
			targetPrice = typed
		case int:
			targetPrice = float64(typed)
		}
	}
	wishlistID := uuid.NewString()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO wishlist_entries(id, profile_id, item_id, target_price, priority, notes, highlight_hit)
		VALUES (?, ?, ?, ?, ?, ?, 0)
	`, wishlistID, strings.TrimSpace(profileID), itemID, targetPrice, strings.TrimSpace(priority), strings.TrimSpace(notes))
	if err != nil {
		return "", "", fmt.Errorf("create wishlist entry: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `
		UPDATE canonical_items
		SET status = 'wishlist', priority = ?, updated_at = CURRENT_TIMESTAMP, updated_by = 'chat.service'
		WHERE id = ? AND profile_id = ?
	`, strings.TrimSpace(priority), itemID, strings.TrimSpace(profileID))
	if err != nil {
		return "", "", fmt.Errorf("sync wishlist item: %w", err)
	}
	return itemID, wishlistID, nil
}

type workspaceCollectionsState struct {
	Collections      []string                  `json:"collections"`
	ActiveCollection string                    `json:"activeCollection"`
	Items            []workspaceCollectionItem `json:"items"`
}

type workspaceCollectionItem struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Detail         string `json:"detail"`
	CollectionName string `json:"collectionName,omitempty"`
}

func (s *Service) applyAssignCollectionItem(ctx context.Context, profileID string, payload map[string]any) (string, string, error) {
	itemID, _ := payload["item_id"].(string)
	collectionName, _ := payload["collection_name"].(string)
	title, _ := payload["title"].(string)
	partNumber, _ := payload["part_number"].(string)
	itemID = strings.TrimSpace(itemID)
	collectionName = strings.TrimSpace(collectionName)
	if itemID == "" || collectionName == "" || collectionName == "All Items" {
		return "", "", fmt.Errorf("item_id and assignable collection_name are required")
	}
	if strings.TrimSpace(title) == "" {
		title = itemID
	}
	detail := strings.TrimSpace(partNumber)
	if detail == "" {
		detail = "Assigned by chat copilot"
	}

	settingsKey := "collections.workspace.v1"
	state := workspaceCollectionsState{
		Collections:      []string{"All Items", collectionName},
		ActiveCollection: collectionName,
		Items:            []workspaceCollectionItem{},
	}
	var raw string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM profile_settings WHERE profile_id = ? AND key = ?`, strings.TrimSpace(profileID), settingsKey).Scan(&raw)
	if err != nil && err != sql.ErrNoRows {
		return "", "", fmt.Errorf("load workspace collections: %w", err)
	}
	if strings.TrimSpace(raw) != "" {
		var existing workspaceCollectionsState
		if err := json.Unmarshal([]byte(raw), &existing); err == nil {
			state = existing
		}
	}
	state.Collections = ensureCollectionName(state.Collections, "All Items")
	state.Collections = ensureCollectionName(state.Collections, collectionName)
	state.ActiveCollection = collectionName
	updated := false
	for i := range state.Items {
		if state.Items[i].ID == itemID {
			state.Items[i].CollectionName = collectionName
			if strings.TrimSpace(state.Items[i].Name) == "" {
				state.Items[i].Name = strings.TrimSpace(title)
			}
			if strings.TrimSpace(state.Items[i].Detail) == "" {
				state.Items[i].Detail = detail
			}
			updated = true
			break
		}
	}
	if !updated {
		state.Items = append(state.Items, workspaceCollectionItem{
			ID:             itemID,
			Name:           strings.TrimSpace(title),
			Detail:         detail,
			CollectionName: collectionName,
		})
	}
	nextRaw, err := json.Marshal(state)
	if err != nil {
		return "", "", fmt.Errorf("marshal workspace collections: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO profile_settings(profile_id, key, value, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(profile_id, key) DO UPDATE SET value=excluded.value, updated_at=CURRENT_TIMESTAMP
	`, strings.TrimSpace(profileID), settingsKey, string(nextRaw))
	if err != nil {
		return "", "", fmt.Errorf("assign collection item: %w", err)
	}
	return itemID, collectionName, nil
}

func ensureCollectionName(collections []string, name string) []string {
	name = strings.TrimSpace(name)
	if name == "" {
		return collections
	}
	for _, existing := range collections {
		if strings.EqualFold(strings.TrimSpace(existing), name) {
			return collections
		}
	}
	return append(collections, name)
}

func trimPayloadString(payload map[string]any, key string) string {
	value, _ := payload[key].(string)
	return strings.TrimSpace(value)
}

func applyActionMessage(result ApplyActionResult) string {
	fieldSummary := applyActionFieldSummary(result)
	switch result.Action {
	case "assign_collection_item":
		return fmt.Sprintf("Applied assign_collection_item to %s in %s.", result.ItemID, result.CollectionName)
	case "create_wishlist_entry":
		if strings.TrimSpace(result.ItemID) != "" {
			return fmt.Sprintf("Applied create_wishlist_entry to wishlist %s for item %s.", result.WishlistID, result.ItemID)
		}
		return fmt.Sprintf("Applied create_wishlist_entry to wishlist %s.", result.WishlistID)
	default:
		if strings.TrimSpace(result.ItemID) != "" {
			if fieldSummary != "" {
				return fmt.Sprintf("Applied %s to %s with %s.", result.Action, result.ItemID, fieldSummary)
			}
			return fmt.Sprintf("Applied %s to %s.", result.Action, result.ItemID)
		}
		return fmt.Sprintf("Applied %s.", result.Action)
	}
}

func cancelActionMessage(result ApplyActionResult) string {
	fieldSummary := applyActionFieldSummary(result)
	if strings.TrimSpace(result.ItemID) != "" {
		if fieldSummary != "" {
			return fmt.Sprintf("Canceled %s for %s with %s; no mutation applied.", result.Action, result.ItemID, fieldSummary)
		}
		return fmt.Sprintf("Canceled %s for %s; no mutation applied.", result.Action, result.ItemID)
	}
	if fieldSummary != "" {
		return fmt.Sprintf("Canceled %s with %s; no mutation applied.", result.Action, fieldSummary)
	}
	return fmt.Sprintf("Canceled %s; no mutation applied.", result.Action)
}

func applyActionFieldSummary(result ApplyActionResult) string {
	parts := []string{}
	if strings.TrimSpace(result.PartNumber) != "" {
		parts = append(parts, fmt.Sprintf("part_number=%s", strings.TrimSpace(result.PartNumber)))
	}
	if strings.TrimSpace(result.Title) != "" {
		parts = append(parts, fmt.Sprintf("title=%s", strings.TrimSpace(result.Title)))
	}
	return strings.Join(parts, " ")
}

func (s *Service) CleanupOldPreviews(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan).UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `DELETE FROM chat_action_previews WHERE status = 'previewed' AND created_at < ?`, cutoff)
	return err
}

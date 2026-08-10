package logging

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

type Entry struct {
	ID        string `json:"id"`
	Level     string `json:"level"`
	Action    string `json:"action"`
	Details   string `json:"details"`
	CreatedAt string `json:"created_at"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Log(ctx context.Context, level, action string, details any) {
	if strings.TrimSpace(level) == "" {
		level = "info"
	}
	raw, _ := json.Marshal(RedactValue(details))
	redacted := RedactSensitive(string(raw))
	_, _ = s.db.ExecContext(ctx, `
		INSERT INTO activity_logs(id, level, action, details, created_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, uuid.NewString(), level, action, redacted)
}

func (s *Service) List(ctx context.Context, limit int) ([]Entry, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, level, action, details, created_at
		FROM activity_logs
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Level, &e.Action, &e.Details, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Service) Export(ctx context.Context) (string, error) {
	items, err := s.List(ctx, 500)
	if err != nil {
		return "", err
	}
	raw, err := json.MarshalIndent(map[string]any{"logs": items}, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal logs export: %w", err)
	}
	return RedactSensitive(string(raw)), nil
}

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(authorization\s*[=:]\s*)(bearer\s+[^\s,"'}]+)`),
	regexp.MustCompile(`(?i)((?:api[_-]?key|token|secret|password|authorization|cookie)\"?\s*:\s*\")([^\"]+)`),
	regexp.MustCompile(`(?i)((?:api[_-]?key|token|secret|password|authorization|cookie)\s*[=:]\s*)([^\s,"'}]+)`),
	regexp.MustCompile(`(?i)(bearer\s+)([A-Za-z0-9\-_\.]+)`),
	regexp.MustCompile(`(?i)(sk-[A-Za-z0-9_\-]{6,})`),
	regexp.MustCompile(`(?i)(C:\\Users\\[^\\\s"]+(?:\\[^\s"]*)?)`),
	regexp.MustCompile(`(?i)(/(?:Users|home)/[^\s"]+)`),
}

func RedactSensitive(s string) string {
	out := s
	for _, re := range secretPatterns {
		if re.NumSubexp() >= 2 {
			out = re.ReplaceAllString(out, "${1}[REDACTED]")
			continue
		}
		out = re.ReplaceAllString(out, "[REDACTED]")
	}
	return out
}

func RedactValue(value any) any {
	return redactValue("", value)
}

func redactValue(key string, value any) any {
	if isSensitiveKey(key) {
		if value == nil {
			return nil
		}
		return "[REDACTED]"
	}
	if isPrivateContentKey(key) {
		if value == nil {
			return nil
		}
		return "[REDACTED_CONTENT]"
	}
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for k, v := range typed {
			out[k] = redactValue(k, v)
		}
		return out
	case map[string]string:
		out := make(map[string]string, len(typed))
		for k, v := range typed {
			if isSensitiveKey(k) {
				out[k] = "[REDACTED]"
				continue
			}
			if isPrivateContentKey(k) {
				out[k] = "[REDACTED_CONTENT]"
				continue
			}
			out[k] = RedactSensitive(v)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, v := range typed {
			out[i] = redactValue("", v)
		}
		return out
	case []string:
		out := make([]string, len(typed))
		for i, v := range typed {
			out[i] = RedactSensitive(v)
		}
		return out
	case string:
		return RedactSensitive(typed)
	default:
		return typed
	}
}

func isSensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(key), "-", "_"), " ", "_"))
	for _, token := range []string{"password", "passwd", "pwd", "cookie", "authorization", "auth_header", "bearer", "api_key", "apikey", "token", "secret", "credential", "session_id"} {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

func isPrivateContentKey(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(key), "-", "_"), " ", "_"))
	for _, token := range []string{"raw_page", "page_content", "page_html", "private_page", "document_body", "html_snapshot"} {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

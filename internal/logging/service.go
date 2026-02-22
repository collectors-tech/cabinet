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
	raw, _ := json.Marshal(details)
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
	return string(raw), nil
}

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(api[_-]?key\"?\s*:\s*\")([^\"]+)`),
	regexp.MustCompile(`(?i)(token\"?\s*:\s*\")([^\"]+)`),
	regexp.MustCompile(`(?i)(bearer\s+)([A-Za-z0-9\-_\.]+)`),
}

func RedactSensitive(s string) string {
	out := s
	for _, re := range secretPatterns {
		out = re.ReplaceAllString(out, "${1}[REDACTED]")
	}
	return out
}

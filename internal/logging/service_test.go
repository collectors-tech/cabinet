package logging

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

func TestLogListExportAndRedaction(t *testing.T) {
	t.Parallel()
	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	svc := NewService(conn)
	svc.Log(context.Background(), "info", "test_action", map[string]any{"api_key": "sk-secret", "token": "abc"})
	items, err := svc.List(context.Background(), 10)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected activity logs")
	}
	if strings.Contains(items[0].Details, "sk-secret") {
		t.Fatalf("expected redacted details, got %s", items[0].Details)
	}
	exported, err := svc.Export(context.Background())
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}
	if !strings.Contains(exported, "[REDACTED]") {
		t.Fatalf("expected redacted export, got %s", exported)
	}
}

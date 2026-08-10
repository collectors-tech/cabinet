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
	details := map[string]any{
		"api_key": "sk-secret",
		"token":   "abc",
		"nested": map[string]any{
			"password":      "hunter2",
			"cookie":        "sid=raw-cookie",
			"authorization": "Bearer raw-auth-token",
			"page_content":  "<html>private collector page</html>",
			"path":          `C:\Users\Max\Cabinet\private.db`,
			"message":       "provider failed with api_key=sk-nestedsecret and Authorization: Bearer nested-token",
		},
		"list": []any{
			map[string]any{"generic_secret": "secret-list-value"},
			"token=plain-list-token",
		},
	}
	svc.Log(context.Background(), "info", "test_action", details)
	items, err := svc.List(context.Background(), 10)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected activity logs")
	}
	for _, leaked := range []string{
		"sk-secret",
		"abc",
		"hunter2",
		"raw-cookie",
		"raw-auth-token",
		"private collector page",
		`C:\Users\Max\Cabinet\private.db`,
		"sk-nestedsecret",
		"nested-token",
		"secret-list-value",
		"plain-list-token",
	} {
		if strings.Contains(items[0].Details, leaked) {
			t.Fatalf("expected persisted details to redact %q, got %s", leaked, items[0].Details)
		}
	}
	if details["api_key"] != "sk-secret" || details["nested"].(map[string]any)["password"] != "hunter2" {
		t.Fatalf("redaction mutated source details: %+v", details)
	}
	if !strings.Contains(items[0].Details, "[REDACTED]") || !strings.Contains(items[0].Details, "[REDACTED_CONTENT]") {
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

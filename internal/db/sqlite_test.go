package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestOpenAndMigrateCreatesCoreTables(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "cabinet.db")
	conn, err := OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	assertTableExists(t, conn, "schema_migrations")
	assertTableExists(t, conn, "profiles")
}

func assertTableExists(t *testing.T, conn *sql.DB, table string) {
	t.Helper()

	var got string
	err := conn.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&got)
	if err != nil {
		t.Fatalf("table lookup failed: %v", err)
	}
	if got != table {
		t.Fatalf("expected table %q, got %q", table, got)
	}
}

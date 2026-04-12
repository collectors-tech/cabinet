package db

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"
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

func TestOpenAndMigrateParallelFreshDBsCompletesWithinContext(t *testing.T) {
	t.Parallel()

	const workers = 12
	root := t.TempDir()

	var wg sync.WaitGroup
	errCh := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			path := filepath.Join(root, fmt.Sprintf("db-%02d", i), "cabinet.db")
			conn, err := OpenAndMigrate(ctx, path)
			if err != nil {
				errCh <- fmt.Errorf("worker %d: %w", i, err)
				return
			}
			defer conn.Close()
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("OpenAndMigrate() parallel error = %v", err)
		}
	}
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

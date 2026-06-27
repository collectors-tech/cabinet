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

func TestOpenAndMigrateRebuildsLegacyScannerCandidateUniqueness(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "cabinet.db")
	legacy, err := sql.Open("sqlite", "file:"+dbPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("open legacy sqlite: %v", err)
	}
	if _, err := legacy.Exec(`CREATE TABLE scanner_query_sets (
		id TEXT PRIMARY KEY,
		profile_id TEXT NOT NULL DEFAULT '',
		name TEXT NOT NULL,
		keywords_json TEXT NOT NULL,
		exclusions_json TEXT NOT NULL DEFAULT '[]',
		provider_scope_json TEXT NOT NULL DEFAULT '[]',
		items_per_page INTEGER NOT NULL DEFAULT 24,
		max_price REAL NOT NULL DEFAULT 0,
		region TEXT NOT NULL DEFAULT '',
		condition_filter TEXT NOT NULL DEFAULT '',
		schedule_cron TEXT NOT NULL DEFAULT '',
		enabled INTEGER NOT NULL DEFAULT 1,
		rate_limit_rps INTEGER NOT NULL DEFAULT 2,
		max_retry_count INTEGER NOT NULL DEFAULT 2,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE scanner_candidates (
		id TEXT PRIMARY KEY,
		profile_id TEXT NOT NULL DEFAULT '',
		query_set_id TEXT NOT NULL,
		listing_id TEXT NOT NULL UNIQUE,
		title TEXT NOT NULL,
		price REAL NOT NULL DEFAULT 0,
		shipping REAL NOT NULL DEFAULT 0,
		url TEXT NOT NULL,
		image TEXT NOT NULL DEFAULT '',
		seller TEXT NOT NULL DEFAULT '',
		first_seen TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_seen TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		status TEXT NOT NULL DEFAULT 'new',
		source TEXT NOT NULL DEFAULT '',
		observed_currency TEXT NOT NULL DEFAULT '',
		reviewer_notes TEXT NOT NULL DEFAULT '',
		source_result_url TEXT NOT NULL DEFAULT '',
		stock_state TEXT NOT NULL DEFAULT 'unknown',
		stock_count INTEGER NOT NULL DEFAULT -1,
		FOREIGN KEY (query_set_id) REFERENCES scanner_query_sets(id) ON DELETE CASCADE
	);`); err != nil {
		t.Fatalf("seed legacy schema: %v", err)
	}
	if err := legacy.Close(); err != nil {
		t.Fatalf("close legacy db: %v", err)
	}

	conn, err := OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	assertTableExists(t, conn, "scanner_runs")
	if _, err := conn.Exec(`INSERT INTO scanner_query_sets(id, name, keywords_json) VALUES ('q1', 'Q1', '["afx"]'), ('q2', 'Q2', '["afx"]')`); err != nil {
		t.Fatalf("insert query sets: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_candidates(id, profile_id, query_set_id, listing_id, title, url, source) VALUES
		('c1', 'profile-a', 'q1', 'SHARED', 'Shared 1', 'https://example.test/1', 'ebay'),
		('c2', 'profile-a', 'q2', 'SHARED', 'Shared 2', 'https://example.test/2', 'ebay')`); err != nil {
		t.Fatalf("expected rebuilt scanner_candidates to allow per-watch shared listing ids: %v", err)
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

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

func TestOpenAndMigratePreservesRepresentativeLegacyReleaseData(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "cabinet.db")
	legacy, err := sql.Open("sqlite", "file:"+dbPath+"?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("open legacy sqlite: %v", err)
	}
	if _, err := legacy.Exec(`CREATE TABLE profiles (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE profile_settings (
		profile_id TEXT NOT NULL,
		key TEXT NOT NULL,
		value TEXT NOT NULL,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (profile_id, key),
		FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE
	);
	CREATE TABLE profile_licenses (
		profile_id TEXT PRIMARY KEY,
		license_json TEXT NOT NULL,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE
	);
	CREATE TABLE saved_filters (
		id TEXT PRIMARY KEY,
		profile_id TEXT NOT NULL,
		name TEXT NOT NULL,
		query_json TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE
	);
	CREATE TABLE canonical_items (
		id TEXT PRIMARY KEY,
		profile_id TEXT NOT NULL DEFAULT '',
		brand TEXT NOT NULL,
		category TEXT NOT NULL,
		item_type TEXT NOT NULL DEFAULT '',
		part_number TEXT NOT NULL,
		title TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'active',
		priority TEXT NOT NULL DEFAULT 'medium',
		grading_status TEXT NOT NULL DEFAULT 'ungraded',
		grader TEXT NOT NULL DEFAULT '',
		grade_numeric REAL NOT NULL DEFAULT 0,
		slabbed INTEGER NOT NULL DEFAULT 0,
		collector_classification TEXT NOT NULL DEFAULT '',
		car_grade_type TEXT NOT NULL DEFAULT '',
		packaging_grade_type TEXT NOT NULL DEFAULT '',
		make TEXT NOT NULL DEFAULT '',
		model TEXT NOT NULL DEFAULT '',
		year TEXT NOT NULL DEFAULT '',
		scale TEXT NOT NULL DEFAULT '',
		series TEXT NOT NULL DEFAULT '',
		description TEXT NOT NULL DEFAULT '',
		notes TEXT NOT NULL DEFAULT '',
		tags_json TEXT NOT NULL DEFAULT '[]',
		source_urls_json TEXT NOT NULL DEFAULT '[]',
		for_sale INTEGER NOT NULL DEFAULT 0,
		structured_offers_json TEXT NOT NULL DEFAULT '[]',
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		created_by TEXT NOT NULL DEFAULT 'system',
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_by TEXT NOT NULL DEFAULT 'system',
		deleted_at TEXT NOT NULL DEFAULT '',
		deleted_by TEXT NOT NULL DEFAULT ''
	);
	CREATE UNIQUE INDEX idx_canonical_items_part_number ON canonical_items(part_number);
	CREATE TABLE item_barcodes (
		id TEXT PRIMARY KEY,
		item_id TEXT NOT NULL,
		barcode TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (item_id) REFERENCES canonical_items(id) ON DELETE CASCADE
	);
	CREATE TABLE instances (
		id TEXT PRIMARY KEY,
		item_id TEXT NOT NULL,
		condition TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL,
		quantity INTEGER NOT NULL DEFAULT 1,
		storage_location TEXT NOT NULL DEFAULT '',
		acquisition_price REAL NOT NULL DEFAULT 0,
		acquisition_date TEXT NOT NULL DEFAULT '',
		notes TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (item_id) REFERENCES canonical_items(id) ON DELETE CASCADE
	);
	CREATE TABLE item_photos (
		id TEXT PRIMARY KEY,
		item_id TEXT NOT NULL,
		filename TEXT NOT NULL,
		original_path TEXT NOT NULL,
		preview_path TEXT NOT NULL,
		thumbnail_path TEXT NOT NULL,
		is_primary INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (item_id) REFERENCES canonical_items(id) ON DELETE CASCADE
	);
	CREATE TABLE wishlist_entries (
		id TEXT PRIMARY KEY,
		profile_id TEXT NOT NULL DEFAULT '',
		item_id TEXT NOT NULL UNIQUE,
		target_price REAL NOT NULL DEFAULT 0,
		priority TEXT NOT NULL DEFAULT 'normal',
		notes TEXT NOT NULL DEFAULT '',
		highlight_hit INTEGER NOT NULL DEFAULT 1,
		below_target_now INTEGER NOT NULL DEFAULT 0,
		owned INTEGER NOT NULL DEFAULT 0,
		delivered INTEGER NOT NULL DEFAULT 0,
		price_paid REAL NOT NULL DEFAULT 0,
		purchase_url TEXT NOT NULL DEFAULT '',
		purchase_date TEXT NOT NULL DEFAULT '',
		purchase_condition TEXT NOT NULL DEFAULT '',
		quantity INTEGER NOT NULL DEFAULT 0,
		needed_quantity INTEGER NOT NULL DEFAULT 1,
		deleted INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (item_id) REFERENCES canonical_items(id) ON DELETE CASCADE
	);
	CREATE TABLE scanner_query_sets (
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
		t.Fatalf("create representative legacy fixture: %v", err)
	}
	if _, err := legacy.Exec(`INSERT INTO profiles(id, name) VALUES ('profile-main', 'Release Upgrade Profile');
		INSERT INTO profile_settings(profile_id, key, value) VALUES ('profile-main', 'currency', 'AUD');
		INSERT INTO profile_licenses(profile_id, license_json) VALUES ('profile-main', '{"plan":"beta"}');
		INSERT INTO saved_filters(id, profile_id, name, query_json) VALUES ('filter-1', 'profile-main', 'Sealed AFX', '{"status":"sealed"}');
		INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, tags_json) VALUES ('item-1', 'profile-main', 'AFX', 'Slot Car', 'LEGACY-001', 'Legacy Item', '["release","upgrade"]');
		INSERT INTO item_barcodes(id, item_id, barcode) VALUES ('barcode-1', 'item-1', '9990001112223');
		INSERT INTO instances(id, item_id, condition, status, quantity, storage_location) VALUES ('instance-1', 'item-1', 'mint', 'owned', 2, 'Main Shelf');
		INSERT INTO item_photos(id, item_id, filename, original_path, preview_path, thumbnail_path, is_primary) VALUES ('photo-1', 'item-1', 'front.jpg', 'media/front.jpg', 'media/front-preview.jpg', 'media/front-thumb.jpg', 1);
		INSERT INTO wishlist_entries(id, profile_id, item_id, target_price, priority, notes) VALUES ('wish-1', 'profile-main', 'item-1', 25.5, 'high', 'buy spare');
		INSERT INTO scanner_query_sets(id, profile_id, name, keywords_json) VALUES ('watch-1', 'profile-main', 'AFX Watch', '["afx"]');
		INSERT INTO scanner_candidates(id, profile_id, query_set_id, listing_id, title, url, source) VALUES ('candidate-1', 'profile-main', 'watch-1', 'LISTING-1', 'Market Watch Result', 'https://example.test/listing-1', 'ebay');`); err != nil {
		t.Fatalf("seed representative legacy fixture: %v", err)
	}
	if err := legacy.Close(); err != nil {
		t.Fatalf("close legacy db: %v", err)
	}

	conn, err := OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	for table, want := range map[string]int{
		"profiles":           1,
		"profile_settings":   1,
		"profile_licenses":   1,
		"saved_filters":      1,
		"canonical_items":    1,
		"item_barcodes":      1,
		"instances":          1,
		"item_photos":        1,
		"wishlist_entries":   1,
		"scanner_query_sets": 1,
		"scanner_candidates": 1,
	} {
		if got := countRows(t, conn, table); got != want {
			t.Fatalf("expected %s count %d after upgrade, got %d", table, want, got)
		}
	}

	var displayOrder int
	if err := conn.QueryRow(`SELECT display_order FROM item_photos WHERE id = 'photo-1'`).Scan(&displayOrder); err != nil {
		t.Fatalf("read upgraded item photo display_order: %v", err)
	}
	if displayOrder != 0 {
		t.Fatalf("expected migrated photo display_order default 0, got %d", displayOrder)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_query_sets(id, profile_id, name, keywords_json) VALUES ('watch-2', 'profile-main', 'AFX Watch 2', '["afx"]')`); err != nil {
		t.Fatalf("insert second upgraded query set: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_candidates(id, profile_id, query_set_id, listing_id, title, url, source) VALUES ('candidate-2', 'profile-main', 'watch-2', 'LISTING-1', 'Same Listing New Watch', 'https://example.test/listing-2', 'ebay')`); err != nil {
		t.Fatalf("expected upgraded scanner_candidates to preserve data while allowing scoped duplicate listing ids: %v", err)
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

func countRows(t *testing.T, conn *sql.DB, table string) int {
	t.Helper()

	var got int
	if err := conn.QueryRow(fmt.Sprintf("SELECT COUNT(1) FROM %s", table)).Scan(&got); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return got
}

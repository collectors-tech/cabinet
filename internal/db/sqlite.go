package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func OpenAndMigrate(ctx context.Context, path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", path)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS app_state (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS profiles (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS profile_settings (
			profile_id TEXT NOT NULL,
			key TEXT NOT NULL,
			value TEXT NOT NULL,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (profile_id, key),
			FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS saved_filters (
			id TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL,
			name TEXT NOT NULL,
			query_json TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_saved_filters_profile_id ON saved_filters(profile_id);`,
		`CREATE TABLE IF NOT EXISTS profile_secrets (
			profile_id TEXT NOT NULL,
			key TEXT NOT NULL,
			value TEXT NOT NULL,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (profile_id, key),
			FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_profile_secrets_profile_id ON profile_secrets(profile_id);`,
		`CREATE TABLE IF NOT EXISTS profile_licenses (
			profile_id TEXT PRIMARY KEY,
			license_json TEXT NOT NULL,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS webauthn_credentials (
			id TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL,
			credential_json TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_webauthn_credentials_profile_id ON webauthn_credentials(profile_id);`,
		`CREATE TABLE IF NOT EXISTS canonical_items (
			id TEXT PRIMARY KEY,
			brand TEXT NOT NULL,
			category TEXT NOT NULL,
			part_number TEXT NOT NULL,
			title TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			make TEXT NOT NULL DEFAULT '',
			model TEXT NOT NULL DEFAULT '',
			year TEXT NOT NULL DEFAULT '',
			scale TEXT NOT NULL DEFAULT '',
			series TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			tags_json TEXT NOT NULL DEFAULT '[]',
			for_sale INTEGER NOT NULL DEFAULT 0,
			structured_offers_json TEXT NOT NULL DEFAULT '[]',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_canonical_items_part_number ON canonical_items(part_number);`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS canonical_items_fts USING fts5(
			item_id UNINDEXED,
			part_number,
			title,
			brand,
			category,
			description,
			tags
		);`,
		`CREATE TRIGGER IF NOT EXISTS trg_items_fts_insert AFTER INSERT ON canonical_items BEGIN
			INSERT INTO canonical_items_fts (item_id, part_number, title, brand, category, description, tags)
			VALUES (new.id, new.part_number, new.title, new.brand, new.category, new.description, new.tags_json);
		END;`,
		`CREATE TRIGGER IF NOT EXISTS trg_items_fts_update AFTER UPDATE ON canonical_items BEGIN
			UPDATE canonical_items_fts
			SET part_number = new.part_number,
				title = new.title,
				brand = new.brand,
				category = new.category,
				description = new.description,
				tags = new.tags_json
			WHERE item_id = new.id;
		END;`,
		`CREATE TRIGGER IF NOT EXISTS trg_items_fts_delete AFTER DELETE ON canonical_items BEGIN
			DELETE FROM canonical_items_fts WHERE item_id = old.id;
		END;`,
		`CREATE TABLE IF NOT EXISTS item_barcodes (
			id TEXT PRIMARY KEY,
			item_id TEXT NOT NULL,
			barcode TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (item_id) REFERENCES canonical_items(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_item_barcodes_item_id ON item_barcodes(item_id);`,
		`CREATE TABLE IF NOT EXISTS instances (
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
		);`,
		`CREATE INDEX IF NOT EXISTS idx_instances_item_id ON instances(item_id);`,
		`CREATE TABLE IF NOT EXISTS item_photos (
			id TEXT PRIMARY KEY,
			item_id TEXT NOT NULL,
			filename TEXT NOT NULL,
			original_path TEXT NOT NULL,
			preview_path TEXT NOT NULL,
			thumbnail_path TEXT NOT NULL,
			is_primary INTEGER NOT NULL DEFAULT 0,
			display_order INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (item_id) REFERENCES canonical_items(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_item_photos_item_id ON item_photos(item_id);`,
		`CREATE TABLE IF NOT EXISTS scanner_query_sets (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			keywords_json TEXT NOT NULL,
			exclusions_json TEXT NOT NULL DEFAULT '[]',
			provider_scope_json TEXT NOT NULL DEFAULT '[]',
			max_price REAL NOT NULL DEFAULT 0,
			region TEXT NOT NULL DEFAULT '',
			condition_filter TEXT NOT NULL DEFAULT '',
			schedule_cron TEXT NOT NULL DEFAULT '',
			enabled INTEGER NOT NULL DEFAULT 1,
			rate_limit_rps INTEGER NOT NULL DEFAULT 2,
			max_retry_count INTEGER NOT NULL DEFAULT 2,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS scanner_candidates (
			id TEXT PRIMARY KEY,
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
			stock_state TEXT NOT NULL DEFAULT 'unknown',
			stock_count INTEGER NOT NULL DEFAULT -1,
			FOREIGN KEY (query_set_id) REFERENCES scanner_query_sets(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_scanner_candidates_query_set_id ON scanner_candidates(query_set_id);`,
		`CREATE TABLE IF NOT EXISTS scanner_matches (
			candidate_id TEXT PRIMARY KEY,
			item_id TEXT NOT NULL DEFAULT '',
			state TEXT NOT NULL,
			confidence REAL NOT NULL DEFAULT 0,
			needs_review INTEGER NOT NULL DEFAULT 1,
			extracted_part_number TEXT NOT NULL DEFAULT '',
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (candidate_id) REFERENCES scanner_candidates(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS provider_health (
			provider TEXT PRIMARY KEY,
			status TEXT NOT NULL,
			message TEXT NOT NULL DEFAULT '',
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS scanner_failures (
			id TEXT PRIMARY KEY,
			query_set_id TEXT NOT NULL DEFAULT '',
			provider TEXT NOT NULL,
			message TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS discovery_actions (
			id TEXT PRIMARY KEY,
			candidate_id TEXT NOT NULL,
			action_type TEXT NOT NULL,
			payload_json TEXT NOT NULL DEFAULT '{}',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (candidate_id) REFERENCES scanner_candidates(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS ignored_candidates (
			candidate_id TEXT PRIMARY KEY,
			ignored_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (candidate_id) REFERENCES scanner_candidates(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS wishlist_entries (
			id TEXT PRIMARY KEY,
			item_id TEXT NOT NULL UNIQUE,
			target_price REAL NOT NULL DEFAULT 0,
			priority TEXT NOT NULL DEFAULT 'normal',
			notes TEXT NOT NULL DEFAULT '',
			highlight_hit INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (item_id) REFERENCES canonical_items(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS tracked_items (
			item_id TEXT PRIMARY KEY,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (item_id) REFERENCES canonical_items(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS price_snapshots (
			id TEXT PRIMARY KEY,
			item_id TEXT NOT NULL,
			snapshot_date TEXT NOT NULL,
			source TEXT NOT NULL DEFAULT '',
			min_price REAL NOT NULL DEFAULT 0,
			median_price REAL NOT NULL DEFAULT 0,
			latest_price REAL NOT NULL DEFAULT 0,
			stock_count INTEGER NOT NULL DEFAULT -1,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (item_id) REFERENCES canonical_items(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_price_snapshots_item_id ON price_snapshots(item_id);`,
		`CREATE TABLE IF NOT EXISTS ai_failures (
			id TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL,
			message TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS activity_logs (
			id TEXT PRIMARY KEY,
			level TEXT NOT NULL,
			action TEXT NOT NULL,
			details TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS chat_threads (
			id TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL,
			title TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_chat_threads_profile_id ON chat_threads(profile_id);`,
		`CREATE TABLE IF NOT EXISTS chat_messages (
			id TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL,
			thread_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			attachments_json TEXT NOT NULL DEFAULT '[]',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE,
			FOREIGN KEY (thread_id) REFERENCES chat_threads(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_chat_messages_thread_id ON chat_messages(thread_id);`,
		`CREATE TABLE IF NOT EXISTS chat_attachments (
			id TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL,
			thread_id TEXT NOT NULL,
			filename TEXT NOT NULL,
			mime_type TEXT NOT NULL DEFAULT 'application/octet-stream',
			size_bytes INTEGER NOT NULL DEFAULT 0,
			stored_path TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE,
			FOREIGN KEY (thread_id) REFERENCES chat_threads(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_chat_attachments_thread_id ON chat_attachments(thread_id);`,
		`CREATE TABLE IF NOT EXISTS chat_action_previews (
			id TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL,
			thread_id TEXT NOT NULL,
			action TEXT NOT NULL,
			payload_json TEXT NOT NULL DEFAULT '{}',
			status TEXT NOT NULL DEFAULT 'previewed',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			applied_at TEXT,
			FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE,
			FOREIGN KEY (thread_id) REFERENCES chat_threads(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_chat_action_previews_profile_id ON chat_action_previews(profile_id);`,
	}

	for _, q := range queries {
		if _, err := conn.ExecContext(ctx, q); err != nil {
			conn.Close()
			return nil, fmt.Errorf("run migration: %w", err)
		}
	}

	if err := ensureColumn(ctx, conn, "canonical_items", "profile_id", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure canonical_items.profile_id: %w", err)
	}
	if err := ensureColumn(ctx, conn, "canonical_items", "status", "TEXT NOT NULL DEFAULT 'active'"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure canonical_items.status: %w", err)
	}
	if _, err := conn.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_canonical_items_profile_id ON canonical_items(profile_id);`); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure canonical_items profile index: %w", err)
	}
	if err := ensureColumn(ctx, conn, "wishlist_entries", "profile_id", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure wishlist_entries.profile_id: %w", err)
	}
	if _, err := conn.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_wishlist_entries_profile_id ON wishlist_entries(profile_id);`); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure wishlist_entries profile index: %w", err)
	}
	if err := ensureColumn(ctx, conn, "scanner_query_sets", "profile_id", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_query_sets.profile_id: %w", err)
	}
	if err := ensureColumn(ctx, conn, "scanner_query_sets", "provider_scope_json", "TEXT NOT NULL DEFAULT '[]'"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_query_sets.provider_scope_json: %w", err)
	}
	if _, err := conn.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_scanner_query_sets_profile_id ON scanner_query_sets(profile_id);`); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_query_sets profile index: %w", err)
	}
	if err := ensureColumn(ctx, conn, "scanner_candidates", "profile_id", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_candidates.profile_id: %w", err)
	}
	if err := ensureColumn(ctx, conn, "scanner_candidates", "stock_state", "TEXT NOT NULL DEFAULT 'unknown'"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_candidates.stock_state: %w", err)
	}
	if err := ensureColumn(ctx, conn, "scanner_candidates", "stock_count", "INTEGER NOT NULL DEFAULT -1"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_candidates.stock_count: %w", err)
	}
	if err := ensureColumn(ctx, conn, "item_photos", "display_order", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure item_photos.display_order: %w", err)
	}
	if _, err := conn.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_scanner_candidates_profile_id ON scanner_candidates(profile_id);`); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_candidates profile index: %w", err)
	}
	if err := ensureColumn(ctx, conn, "scanner_failures", "query_set_id", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_failures.query_set_id: %w", err)
	}
	if err := ensureColumn(ctx, conn, "tracked_items", "profile_id", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure tracked_items.profile_id: %w", err)
	}
	if err := ensureColumn(ctx, conn, "price_snapshots", "stock_count", "INTEGER NOT NULL DEFAULT -1"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure price_snapshots.stock_count: %w", err)
	}
	if _, err := conn.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_tracked_items_profile_id ON tracked_items(profile_id);`); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure tracked_items profile index: %w", err)
	}

	return conn, nil
}

func ensureColumn(ctx context.Context, conn *sql.DB, table, column, definition string) error {
	rows, err := conn.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return fmt.Errorf("pragma table_info %s: %w", table, err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return fmt.Errorf("scan pragma table_info %s: %w", table, err)
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate pragma table_info %s: %w", table, err)
	}

	if _, err := conn.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition)); err != nil {
		return fmt.Errorf("alter table %s add column %s: %w", table, column, err)
	}
	return nil
}

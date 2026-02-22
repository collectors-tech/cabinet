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
			make TEXT NOT NULL DEFAULT '',
			model TEXT NOT NULL DEFAULT '',
			year TEXT NOT NULL DEFAULT '',
			scale TEXT NOT NULL DEFAULT '',
			series TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			tags_json TEXT NOT NULL DEFAULT '[]',
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
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (item_id) REFERENCES canonical_items(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_item_photos_item_id ON item_photos(item_id);`,
	}

	for _, q := range queries {
		if _, err := conn.ExecContext(ctx, q); err != nil {
			conn.Close()
			return nil, fmt.Errorf("run migration: %w", err)
		}
	}

	return conn, nil
}

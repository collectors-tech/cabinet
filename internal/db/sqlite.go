package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

type queryRower interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

type execer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func OpenAndMigrate(ctx context.Context, path string) (*sql.DB, error) {
	_, statErr := os.Stat(path)
	freshDB := os.IsNotExist(statErr)
	if statErr != nil && !os.IsNotExist(statErr) {
		return nil, fmt.Errorf("stat db path: %w", statErr)
	}

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

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("begin migration tx: %w", err)
	}
	defer tx.Rollback()

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
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
			,updated_by TEXT NOT NULL DEFAULT 'system'
			,deleted_at TEXT NOT NULL DEFAULT ''
			,deleted_by TEXT NOT NULL DEFAULT ''
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
		);`,
		`CREATE TABLE IF NOT EXISTS scanner_candidates (
			id TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL DEFAULT '',
			query_set_id TEXT NOT NULL,
			listing_id TEXT NOT NULL,
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
		);`,
		`CREATE INDEX IF NOT EXISTS idx_scanner_candidates_query_set_id ON scanner_candidates(query_set_id);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_scanner_candidates_result_scope ON scanner_candidates(profile_id, query_set_id, source, listing_id);`,
		`CREATE TABLE IF NOT EXISTS scanner_runs (
			id TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL DEFAULT '',
			query_set_id TEXT NOT NULL DEFAULT '',
			provider TEXT NOT NULL DEFAULT '',
			trigger_type TEXT NOT NULL DEFAULT 'manual',
			started_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			finished_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			status TEXT NOT NULL,
			result_count INTEGER NOT NULL DEFAULT 0,
			new_result_count INTEGER NOT NULL DEFAULT 0,
			error_category TEXT NOT NULL DEFAULT '',
			error_message TEXT NOT NULL DEFAULT '',
			retry_guidance TEXT NOT NULL DEFAULT '',
			FOREIGN KEY (query_set_id) REFERENCES scanner_query_sets(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_scanner_runs_query_set_id ON scanner_runs(query_set_id);`,
		`CREATE INDEX IF NOT EXISTS idx_scanner_runs_profile_id ON scanner_runs(profile_id);`,
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
			retry_after_seconds INTEGER NOT NULL DEFAULT 0,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS scanner_failures (
			id TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL DEFAULT '',
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
		);`,
		`CREATE TABLE IF NOT EXISTS tracked_items (
			item_id TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL DEFAULT '',
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
		`CREATE TABLE IF NOT EXISTS commerce_lifecycle_entries (
			id TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL DEFAULT '',
			item_id TEXT NOT NULL,
			state TEXT NOT NULL,
			source TEXT NOT NULL DEFAULT '',
			external_ref TEXT NOT NULL DEFAULT '',
			quantity INTEGER NOT NULL DEFAULT 1,
			amount REAL NOT NULL DEFAULT 0,
			currency TEXT NOT NULL DEFAULT 'AUD',
			notes TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (item_id) REFERENCES canonical_items(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_commerce_lifecycle_entries_profile_item ON commerce_lifecycle_entries(profile_id, item_id, created_at);`,
		`CREATE TABLE IF NOT EXISTS expected_arrivals (
			id TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL DEFAULT '',
			item_id TEXT NOT NULL,
			lifecycle_entry_id TEXT NOT NULL DEFAULT '',
			source TEXT NOT NULL DEFAULT '',
			external_ref TEXT NOT NULL DEFAULT '',
			quantity INTEGER NOT NULL DEFAULT 1,
			amount REAL NOT NULL DEFAULT 0,
			currency TEXT NOT NULL DEFAULT 'AUD',
			status TEXT NOT NULL DEFAULT 'expected',
			expected_on TEXT NOT NULL DEFAULT '',
			delivered_on TEXT NOT NULL DEFAULT '',
			reconciled_instance_id TEXT NOT NULL DEFAULT '',
			notes TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (item_id) REFERENCES canonical_items(id) ON DELETE CASCADE,
			FOREIGN KEY (lifecycle_entry_id) REFERENCES commerce_lifecycle_entries(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_expected_arrivals_profile_item_status ON expected_arrivals(profile_id, item_id, status, created_at);`,
		`CREATE INDEX IF NOT EXISTS idx_expected_arrivals_lifecycle_entry_id ON expected_arrivals(lifecycle_entry_id);`,
		`CREATE TABLE IF NOT EXISTS forwarder_packages (
			id TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL DEFAULT '',
			provider TEXT NOT NULL,
			source TEXT NOT NULL,
			external_package_id TEXT NOT NULL,
			shipment_id TEXT NOT NULL DEFAULT '',
			tracking_number TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL,
			received_at TEXT NOT NULL DEFAULT '',
			sender TEXT NOT NULL DEFAULT '',
			warehouse_location TEXT NOT NULL DEFAULT '',
			weight_grams INTEGER NOT NULL DEFAULT 0,
			provenance_key TEXT NOT NULL,
			raw_payload_json TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(profile_id, provider, source, external_package_id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_forwarder_packages_profile_status ON forwarder_packages(profile_id, status, updated_at);`,
		`CREATE TABLE IF NOT EXISTS forwarder_package_links (
			id TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL DEFAULT '',
			package_id TEXT NOT NULL UNIQUE,
			item_id TEXT NOT NULL,
			lifecycle_entry_id TEXT NOT NULL DEFAULT '',
			expected_arrival_id TEXT NOT NULL DEFAULT '',
			source TEXT NOT NULL DEFAULT 'manual',
			decision TEXT NOT NULL DEFAULT 'confirmed',
			notes TEXT NOT NULL DEFAULT '',
			audit_trail_json TEXT NOT NULL DEFAULT '[]',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (package_id) REFERENCES forwarder_packages(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_forwarder_package_links_profile_item ON forwarder_package_links(profile_id, item_id, updated_at);`,
		`CREATE TABLE IF NOT EXISTS forwarder_package_link_events (
			id TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL DEFAULT '',
			package_id TEXT NOT NULL,
			link_id TEXT NOT NULL DEFAULT '',
			action TEXT NOT NULL,
			item_id TEXT NOT NULL DEFAULT '',
			lifecycle_entry_id TEXT NOT NULL DEFAULT '',
			expected_arrival_id TEXT NOT NULL DEFAULT '',
			previous_item_id TEXT NOT NULL DEFAULT '',
			previous_lifecycle_entry_id TEXT NOT NULL DEFAULT '',
			previous_expected_arrival_id TEXT NOT NULL DEFAULT '',
			source TEXT NOT NULL DEFAULT '',
			notes TEXT NOT NULL DEFAULT '',
			audit_trail_json TEXT NOT NULL DEFAULT '[]',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (package_id) REFERENCES forwarder_packages(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_forwarder_package_link_events_package ON forwarder_package_link_events(profile_id, package_id, created_at);`,
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
			metadata_json TEXT NOT NULL DEFAULT '{}',
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
			context_json TEXT NOT NULL DEFAULT '{}',
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
		`CREATE TABLE IF NOT EXISTS media_asset_links (
			id TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL,
			asset_id TEXT NOT NULL,
			asset_type TEXT NOT NULL,
			target_type TEXT NOT NULL,
			target_id TEXT NOT NULL,
			source TEXT NOT NULL DEFAULT 'media.workspace',
			audit_summary TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(profile_id, asset_id, target_type, target_id)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_media_asset_links_profile_asset ON media_asset_links(profile_id, asset_id);`,
		`CREATE INDEX IF NOT EXISTS idx_media_asset_links_profile_target ON media_asset_links(profile_id, target_type, target_id);`,
		`CREATE TABLE IF NOT EXISTS chat_inbox_items (
			id TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL,
			thread_id TEXT NOT NULL,
			source TEXT NOT NULL,
			status TEXT NOT NULL,
			title TEXT NOT NULL,
			summary TEXT NOT NULL DEFAULT '',
			metadata_json TEXT NOT NULL DEFAULT '{}',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE,
			FOREIGN KEY (thread_id) REFERENCES chat_threads(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_chat_inbox_items_profile_id ON chat_inbox_items(profile_id, status);`,
		`CREATE INDEX IF NOT EXISTS idx_chat_inbox_items_thread_id ON chat_inbox_items(thread_id);`,
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
		`CREATE TABLE IF NOT EXISTS assistant_workflow_runs (
			id TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL,
			workflow_id TEXT NOT NULL,
			capability_id TEXT NOT NULL,
			source_channel TEXT NOT NULL DEFAULT 'in_app_chat',
			source_thread_id TEXT NOT NULL DEFAULT '',
			source_message_id TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL,
			input_json TEXT NOT NULL DEFAULT '{}',
			provider_trace_json TEXT NOT NULL DEFAULT '{}',
			result_json TEXT NOT NULL DEFAULT '{}',
			error_json TEXT NOT NULL DEFAULT '{}',
			confirmation_state TEXT NOT NULL DEFAULT 'not_required',
			bulk_items_json TEXT NOT NULL DEFAULT '[]',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			started_at TEXT NOT NULL DEFAULT '',
			completed_at TEXT NOT NULL DEFAULT '',
			FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_assistant_workflow_runs_profile_status ON assistant_workflow_runs(profile_id, status, updated_at);`,
		`CREATE INDEX IF NOT EXISTS idx_assistant_workflow_runs_thread ON assistant_workflow_runs(profile_id, source_thread_id, updated_at);`,
		`CREATE TABLE IF NOT EXISTS audit_events (
			id TEXT PRIMARY KEY,
			entity_type TEXT NOT NULL,
			entity_id TEXT NOT NULL,
			action TEXT NOT NULL,
			actor TEXT NOT NULL,
			source TEXT NOT NULL DEFAULT '',
			before_json TEXT NOT NULL DEFAULT '{}',
			after_json TEXT NOT NULL DEFAULT '{}',
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_audit_events_entity_timeline ON audit_events(entity_type, entity_id, created_at);`,
		`CREATE TABLE IF NOT EXISTS pokemon_graded_overrides (
			item_id TEXT PRIMARY KEY,
			profile_id TEXT NOT NULL DEFAULT '',
			grader TEXT NOT NULL DEFAULT '',
			grade_numeric REAL NOT NULL DEFAULT 0,
			cert_number TEXT NOT NULL DEFAULT '',
			slab_state TEXT NOT NULL DEFAULT '',
			valuation_override_amount REAL NOT NULL DEFAULT 0,
			currency TEXT NOT NULL DEFAULT 'AUD',
			source_note TEXT NOT NULL DEFAULT '',
			overridden_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (item_id) REFERENCES canonical_items(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_pokemon_graded_overrides_profile_id ON pokemon_graded_overrides(profile_id);`,
	}

	for _, q := range queries {
		if _, err := tx.ExecContext(ctx, q); err != nil {
			conn.Close()
			return nil, fmt.Errorf("run migration: %w", err)
		}
	}

	if freshDB {
		if _, err := tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_canonical_items_profile_id ON canonical_items(profile_id);`); err != nil {
			conn.Close()
			return nil, fmt.Errorf("ensure canonical_items profile index: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_wishlist_entries_profile_id ON wishlist_entries(profile_id);`); err != nil {
			conn.Close()
			return nil, fmt.Errorf("ensure wishlist_entries profile index: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_scanner_query_sets_profile_id ON scanner_query_sets(profile_id);`); err != nil {
			conn.Close()
			return nil, fmt.Errorf("ensure scanner_query_sets profile index: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_scanner_candidates_profile_id ON scanner_candidates(profile_id);`); err != nil {
			conn.Close()
			return nil, fmt.Errorf("ensure scanner_candidates profile index: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_tracked_items_profile_id ON tracked_items(profile_id);`); err != nil {
			conn.Close()
			return nil, fmt.Errorf("ensure tracked_items profile index: %w", err)
		}
		if err := tx.Commit(); err != nil {
			conn.Close()
			return nil, fmt.Errorf("commit migration tx: %w", err)
		}
		return conn, nil
	}
	if err := ensureColumn(ctx, tx, tx, "canonical_items", "profile_id", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure canonical_items.profile_id: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "canonical_items", "status", "TEXT NOT NULL DEFAULT 'active'"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure canonical_items.status: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "canonical_items", "priority", "TEXT NOT NULL DEFAULT 'medium'"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure canonical_items.priority: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "canonical_items", "item_type", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure canonical_items.item_type: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "canonical_items", "grading_status", "TEXT NOT NULL DEFAULT 'ungraded'"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure canonical_items.grading_status: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "canonical_items", "grader", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure canonical_items.grader: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "canonical_items", "grade_numeric", "REAL NOT NULL DEFAULT 0"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure canonical_items.grade_numeric: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "canonical_items", "slabbed", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure canonical_items.slabbed: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "canonical_items", "collector_classification", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure canonical_items.collector_classification: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "canonical_items", "car_grade_type", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure canonical_items.car_grade_type: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "canonical_items", "packaging_grade_type", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure canonical_items.packaging_grade_type: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "canonical_items", "created_by", "TEXT NOT NULL DEFAULT 'system'"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure canonical_items.created_by: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "canonical_items", "updated_by", "TEXT NOT NULL DEFAULT 'system'"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure canonical_items.updated_by: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "canonical_items", "deleted_at", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure canonical_items.deleted_at: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "canonical_items", "deleted_by", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure canonical_items.deleted_by: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "canonical_items", "notes", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure canonical_items.notes: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "canonical_items", "source_urls_json", "TEXT NOT NULL DEFAULT '[]'"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure canonical_items.source_urls_json: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_canonical_items_profile_id ON canonical_items(profile_id);`); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure canonical_items profile index: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "wishlist_entries", "profile_id", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure wishlist_entries.profile_id: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "wishlist_entries", "below_target_now", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure wishlist_entries.below_target_now: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "wishlist_entries", "owned", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure wishlist_entries.owned: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "wishlist_entries", "delivered", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure wishlist_entries.delivered: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "wishlist_entries", "price_paid", "REAL NOT NULL DEFAULT 0"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure wishlist_entries.price_paid: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "wishlist_entries", "purchase_url", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure wishlist_entries.purchase_url: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "wishlist_entries", "purchase_date", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure wishlist_entries.purchase_date: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "wishlist_entries", "purchase_condition", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure wishlist_entries.purchase_condition: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "wishlist_entries", "quantity", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure wishlist_entries.quantity: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "wishlist_entries", "needed_quantity", "INTEGER NOT NULL DEFAULT 1"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure wishlist_entries.needed_quantity: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "wishlist_entries", "deleted", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure wishlist_entries.deleted: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_wishlist_entries_profile_id ON wishlist_entries(profile_id);`); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure wishlist_entries profile index: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "scanner_query_sets", "profile_id", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_query_sets.profile_id: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "scanner_query_sets", "provider_scope_json", "TEXT NOT NULL DEFAULT '[]'"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_query_sets.provider_scope_json: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "scanner_query_sets", "items_per_page", "INTEGER NOT NULL DEFAULT 24"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_query_sets.items_per_page: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_scanner_query_sets_profile_id ON scanner_query_sets(profile_id);`); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_query_sets profile index: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "scanner_candidates", "profile_id", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_candidates.profile_id: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "scanner_candidates", "stock_state", "TEXT NOT NULL DEFAULT 'unknown'"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_candidates.stock_state: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "scanner_candidates", "stock_count", "INTEGER NOT NULL DEFAULT -1"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_candidates.stock_count: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "scanner_candidates", "observed_currency", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_candidates.observed_currency: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "scanner_candidates", "reviewer_notes", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_candidates.reviewer_notes: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "scanner_candidates", "source_result_url", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_candidates.source_result_url: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "item_photos", "display_order", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure item_photos.display_order: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_scanner_candidates_profile_id ON scanner_candidates(profile_id);`); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_candidates profile index: %w", err)
	}
	if err := rebuildScannerCandidatesWithoutGlobalListingUnique(ctx, tx); err != nil {
		conn.Close()
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS idx_scanner_candidates_result_scope ON scanner_candidates(profile_id, query_set_id, source, listing_id);`); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_candidates scoped result index: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS scanner_runs (
		id TEXT PRIMARY KEY,
		profile_id TEXT NOT NULL DEFAULT '',
		query_set_id TEXT NOT NULL DEFAULT '',
		provider TEXT NOT NULL DEFAULT '',
		trigger_type TEXT NOT NULL DEFAULT 'manual',
		started_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		finished_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		status TEXT NOT NULL,
		result_count INTEGER NOT NULL DEFAULT 0,
		new_result_count INTEGER NOT NULL DEFAULT 0,
		error_category TEXT NOT NULL DEFAULT '',
		error_message TEXT NOT NULL DEFAULT '',
		retry_guidance TEXT NOT NULL DEFAULT '',
		FOREIGN KEY (query_set_id) REFERENCES scanner_query_sets(id) ON DELETE CASCADE
	);`); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_runs table: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_scanner_runs_query_set_id ON scanner_runs(query_set_id);`); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_runs query set index: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_scanner_runs_profile_id ON scanner_runs(profile_id);`); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_runs profile index: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "scanner_failures", "query_set_id", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_failures.query_set_id: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "scanner_failures", "profile_id", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_failures.profile_id: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "provider_health", "retry_after_seconds", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure provider_health.retry_after_seconds: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_scanner_failures_profile_id ON scanner_failures(profile_id);`); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure scanner_failures profile index: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "tracked_items", "profile_id", "TEXT NOT NULL DEFAULT ''"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure tracked_items.profile_id: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "price_snapshots", "stock_count", "INTEGER NOT NULL DEFAULT -1"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure price_snapshots.stock_count: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_tracked_items_profile_id ON tracked_items(profile_id);`); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure tracked_items profile index: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "chat_threads", "metadata_json", "TEXT NOT NULL DEFAULT '{}'"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure chat_threads.metadata_json: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "chat_messages", "context_json", "TEXT NOT NULL DEFAULT '{}'"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure chat_messages.context_json: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "forwarder_package_links", "decision", "TEXT NOT NULL DEFAULT 'confirmed'"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure forwarder_package_links.decision: %w", err)
	}
	if err := ensureColumn(ctx, tx, tx, "forwarder_package_links", "audit_trail_json", "TEXT NOT NULL DEFAULT '[]'"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure forwarder_package_links.audit_trail_json: %w", err)
	}
	if err := tx.Commit(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("commit migration tx: %w", err)
	}

	return conn, nil
}

func rebuildScannerCandidatesWithoutGlobalListingUnique(ctx context.Context, tx *sql.Tx) error {
	var ddl string
	if err := tx.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type = 'table' AND name = 'scanner_candidates'`).Scan(&ddl); err != nil {
		return fmt.Errorf("inspect scanner_candidates schema: %w", err)
	}
	if !strings.Contains(strings.ToUpper(ddl), "LISTING_ID TEXT NOT NULL UNIQUE") {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `ALTER TABLE scanner_candidates RENAME TO scanner_candidates_legacy_unique_listing`); err != nil {
		return fmt.Errorf("rename legacy scanner_candidates: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `CREATE TABLE scanner_candidates (
		id TEXT PRIMARY KEY,
		profile_id TEXT NOT NULL DEFAULT '',
		query_set_id TEXT NOT NULL,
		listing_id TEXT NOT NULL,
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
		return fmt.Errorf("create rebuilt scanner_candidates: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO scanner_candidates(
		id, profile_id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source, observed_currency, reviewer_notes, source_result_url, stock_state, stock_count
	)
	SELECT id, profile_id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source, observed_currency, reviewer_notes, source_result_url, stock_state, stock_count
	FROM scanner_candidates_legacy_unique_listing;`); err != nil {
		return fmt.Errorf("copy rebuilt scanner_candidates: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DROP TABLE scanner_candidates_legacy_unique_listing`); err != nil {
		return fmt.Errorf("drop legacy scanner_candidates: %w", err)
	}
	return nil
}

func ensureColumn(ctx context.Context, query queryRower, exec execer, table, column, definition string) error {
	rows, err := query.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
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

	if _, err := exec.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition)); err != nil {
		return fmt.Errorf("alter table %s add column %s: %w", table, column, err)
	}
	return nil
}

package app

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/collectors-tech/cabinet/internal/auth"
	"github.com/collectors-tech/cabinet/internal/config"
)

type e2eBootstrapRequest struct {
	MinimalProfile bool `json:"minimal_profile"`
}

type e2eBootstrapResponse struct {
	ProfileID    string   `json:"profile_id"`
	ProfileName  string   `json:"profile_name"`
	SessionToken string   `json:"session_token"`
	ItemIDs      []string `json:"item_ids"`
	QuerySetID   string   `json:"query_set_id"`
	ThreadID     string   `json:"thread_id"`
}

type e2eScaleBootstrapRequest struct {
	Profile string `json:"profile"`
	Seed    int64  `json:"seed"`
}

type e2eScaleBootstrapResponse struct {
	Profile     string         `json:"profile"`
	Seed        int64          `json:"seed"`
	ProfileID   string         `json:"profile_id"`
	QuerySetID  string         `json:"query_set_id"`
	DatasetHash string         `json:"dataset_hash"`
	Counts      map[string]int `json:"counts"`
}

func registerE2ETestHooks(mux *http.ServeMux, conn *sql.DB, cfg config.Config, authService *auth.Service) {
	mux.HandleFunc("/api/test/reset", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		if err := resetE2EDatabase(r.Context(), conn); err != nil {
			http.Error(w, `{"error":"failed_to_reset_e2e_state"}`, http.StatusInternalServerError)
			return
		}
		resetAuthProviderOptionsOverride()
		_ = os.RemoveAll(filepath.Join(cfg.DataDir, "media"))
		_ = os.RemoveAll(filepath.Join(cfg.DataDir, "chat-attachments"))
		_ = os.RemoveAll(filepath.Join(cfg.DataDir, "backups"))
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})

	mux.HandleFunc("/api/test/bootstrap", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		var req e2eBootstrapRequest
		if r.Body != nil {
			defer r.Body.Close()
			_ = json.NewDecoder(r.Body).Decode(&req)
		}

		out, err := bootstrapE2EFixtures(r.Context(), conn, cfg, req)
		if err != nil {
			http.Error(w, `{"error":"failed_to_bootstrap_e2e_state"}`, http.StatusInternalServerError)
			return
		}
		out.SessionToken, err = authService.CreateUnlockedSession(out.ProfileID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_unlock_e2e_profile"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(out)
	})

	mux.HandleFunc("/api/test/scale/bootstrap", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req e2eScaleBootstrapRequest
		if r.Body != nil {
			defer r.Body.Close()
			_ = json.NewDecoder(r.Body).Decode(&req)
		}
		profile := strings.TrimSpace(strings.ToUpper(req.Profile))
		switch profile {
		case "S0", "S1", "S2", "S3":
		default:
			http.Error(w, `{"error":"invalid_scale_profile"}`, http.StatusBadRequest)
			return
		}
		if req.Seed == 0 {
			req.Seed = 1
		}
		if err := resetE2EDatabase(r.Context(), conn); err != nil {
			http.Error(w, `{"error":"failed_to_reset_e2e_state"}`, http.StatusInternalServerError)
			return
		}
		out, err := bootstrapE2EScaleFixtures(r.Context(), conn, profile, req.Seed)
		if err != nil {
			http.Error(w, `{"error":"failed_to_bootstrap_scale_state"}`, http.StatusInternalServerError)
			return
		}
		_ = os.MkdirAll(filepath.Join(cfg.DataDir, "media"), 0o755)
		_ = json.NewEncoder(w).Encode(out)
	})

	mux.HandleFunc("/api/test/runtime/setup-status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			State string `json:"state"`
		}
		if r.Body != nil {
			defer r.Body.Close()
			_ = json.NewDecoder(r.Body).Decode(&req)
		}
		state := strings.TrimSpace(strings.ToLower(req.State))
		switch state {
		case "missing":
			if err := os.Remove(runtimeSetupConfigPath(cfg)); err != nil && !os.IsNotExist(err) {
				http.Error(w, `{"error":"failed_to_remove_setup_config"}`, http.StatusInternalServerError)
				return
			}
		case "present":
			payload, err := buildRuntimeSetupConfig(cfg, runtimeSetupRequest{
				InstanceName: "E2E Primary",
				ProfileKey:   "e2e-primary",
				AuthMode:     "local",
			})
			if err != nil {
				http.Error(w, `{"error":"failed_to_build_setup_config"}`, http.StatusInternalServerError)
				return
			}
			if err := writeRuntimeSetupConfig(cfg, payload); err != nil {
				http.Error(w, `{"error":"failed_to_write_setup_config"}`, http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, `{"error":"invalid_setup_state"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":             true,
			"setup_required": runtimeSetupRequired(cfg),
		})
	})

	mux.HandleFunc("/api/test/runtime/setup-config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		raw, err := os.ReadFile(runtimeSetupConfigPath(cfg))
		if err != nil {
			if os.IsNotExist(err) {
				http.Error(w, `{"error":"setup_config_not_found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error":"failed_to_read_setup_config"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(raw)
	})

	mux.HandleFunc("/api/test/runtime/setup-import-source", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Mode string `json:"mode"`
		}
		if r.Body != nil {
			defer r.Body.Close()
			_ = json.NewDecoder(r.Body).Decode(&req)
		}
		mode := strings.TrimSpace(strings.ToLower(req.Mode))
		if mode == "" {
			mode = "valid"
		}
		sourcePath := filepath.Join(cfg.DataDir, "setup-import-source.json")
		if mode == "invalid" {
			if err := os.WriteFile(sourcePath, []byte(`{"invalid":true}`), 0o644); err != nil {
				http.Error(w, `{"error":"failed_to_write_import_source"}`, http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":          true,
				"mode":        "invalid",
				"source_path": sourcePath,
			})
			return
		}

		payload, err := buildRuntimeSetupConfig(cfg, runtimeSetupRequest{
			InstanceName: "Imported E2E Instance",
			ProfileKey:   "imported-e2e",
			AuthMode:     "local",
		})
		if err != nil {
			http.Error(w, `{"error":"failed_to_build_import_source"}`, http.StatusInternalServerError)
			return
		}
		raw, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			http.Error(w, `{"error":"failed_to_encode_import_source"}`, http.StatusInternalServerError)
			return
		}
		raw = append(raw, '\n')
		if err := os.WriteFile(sourcePath, raw, 0o644); err != nil {
			http.Error(w, `{"error":"failed_to_write_import_source"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":          true,
			"mode":        "valid",
			"source_path": sourcePath,
		})
	})

	mux.HandleFunc("/api/test/auth/provider-options", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req authProviderOptionsPayload
		if r.Body != nil {
			defer r.Body.Close()
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !strings.Contains(strings.ToLower(err.Error()), "eof") {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
		}
		setAuthProviderOptionsOverride(req)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
}

func resetE2EDatabase(ctx context.Context, conn *sql.DB) error {
	if _, err := conn.ExecContext(ctx, `PRAGMA foreign_keys = OFF`); err != nil {
		return fmt.Errorf("disable foreign keys: %w", err)
	}
	defer func() { _, _ = conn.ExecContext(context.Background(), `PRAGMA foreign_keys = ON`) }()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin reset tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	tables := []string{
		"activity_logs",
		"ai_failures",
		"app_state",
		"assistant_workflow_runs",
		"audit_events",
		"chat_action_previews",
		"chat_attachments",
		"chat_inbox_items",
		"chat_messages",
		"chat_threads",
		"discovery_actions",
		"expected_arrivals",
		"forwarder_package_link_events",
		"forwarder_package_links",
		"forwarder_packages",
		"ignored_candidates",
		"item_barcodes",
		"item_photos",
		"media_asset_links",
		"price_snapshots",
		"profile_licenses",
		"profile_secrets",
		"provider_health",
		"saved_filters",
		"scanner_matches",
		"scanner_failures",
		"scanner_candidates",
		"tracked_items",
		"webauthn_credentials",
		"wishlist_entries",
		"pokemon_graded_overrides",
		"commerce_lifecycle_entries",
		"instances",
		"profile_settings",
		"scanner_query_sets",
		"canonical_items_fts",
		"canonical_items",
		"profiles",
	}

	for _, table := range tables {
		exists, err := resetTableExists(ctx, tx, table)
		if err != nil {
			return fmt.Errorf("check table %s: %w", table, err)
		}
		if !exists {
			continue
		}
		query := "DELETE FROM " + table
		if _, err := tx.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("clear table %s: %w", table, err)
		}
	}

	if _, err := tx.ExecContext(ctx, `PRAGMA foreign_keys = ON`); err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit reset tx: %w", err)
	}
	return nil
}

func resetTableExists(ctx context.Context, tx *sql.Tx, table string) (bool, error) {
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type IN ('table', 'view') AND name = ?`, table).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func bootstrapE2EFixtures(ctx context.Context, conn *sql.DB, cfg config.Config, req e2eBootstrapRequest) (e2eBootstrapResponse, error) {
	profileID := "e2e-profile-001"
	profileName := "E2E Local"
	itemIDs := []string{"e2e-item-001", "e2e-item-002"}
	querySetID := "e2e-queryset-001"
	threadID := "e2e-thread-001"

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return e2eBootstrapResponse{}, fmt.Errorf("begin bootstrap tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	insertProfile := `INSERT INTO profiles(id, name) VALUES (?, ?)`
	if _, err := tx.ExecContext(ctx, insertProfile, profileID, profileName); err != nil {
		return e2eBootstrapResponse{}, fmt.Errorf("insert profile: %w", err)
	}

	storageDBPath := filepath.Join(cfg.DataDir, "profiles", profileID, "cabinet.db")
	storageMediaDir := filepath.Join(cfg.DataDir, "profiles", profileID, "media")
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO profile_settings(profile_id, key, value, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP), (?, ?, ?, CURRENT_TIMESTAMP), (?, ?, ?, CURRENT_TIMESTAMP), (?, ?, ?, CURRENT_TIMESTAMP), (?, ?, ?, CURRENT_TIMESTAMP)
	`, profileID, "storage.db_path", storageDBPath, profileID, "storage.media_dir", storageMediaDir, profileID, "backup_frequency", "daily", profileID, "scanner_schedule", "manual", profileID, "ai_enabled", "true"); err != nil {
		return e2eBootstrapResponse{}, fmt.Errorf("insert profile settings: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO app_state(key, value, updated_at)
		VALUES ('active_profile_id', ?, CURRENT_TIMESTAMP), ('clean_shutdown', '1', CURRENT_TIMESTAMP), ('recovery_required', '0', CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=CURRENT_TIMESTAMP
	`, profileID); err != nil {
		return e2eBootstrapResponse{}, fmt.Errorf("set app_state: %w", err)
	}

	if req.MinimalProfile {
		if err := tx.Commit(); err != nil {
			return e2eBootstrapResponse{}, fmt.Errorf("commit bootstrap minimal tx: %w", err)
		}
		return e2eBootstrapResponse{
			ProfileID:   profileID,
			ProfileName: profileName,
			ItemIDs:     itemIDs[:1],
			QuerySetID:  querySetID,
			ThreadID:    threadID,
		}, nil
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, make, model, year, scale, series, description, tags_json, created_at, updated_at)
		VALUES
		 (?, ?, ?, ?, ?, ?, '', '', '', '', 'Series A', 'Seed item 1', '["seed"]', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
		 (?, ?, ?, ?, ?, ?, '', '', '', '', 'Series B', 'Seed item 2', '["seed"]', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, itemIDs[0], profileID, "E2E Brand", "Cars", "E2E-PN-001", "E2E Starter Car", itemIDs[1], profileID, "E2E Brand", "Cars", "E2E-PN-002", "E2E Secondary Car"); err != nil {
		return e2eBootstrapResponse{}, fmt.Errorf("insert items: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, status, priority, make, model, year, scale, series, description, tags_json, created_at, updated_at)
		VALUES
		 ('e2e-wishlist-item-grail', ?, 'E2E Wishlist', 'Trading Cards', 'WISH-GRAIL-001', 'Wishlist Sample Grail Chase', 'wishlist', 'high', 'Pokemon', 'Charizard', '1999', 'Card', 'Wishlist Samples', 'Representative high-priority wishlist sample.', '["seed","wishlist","grail"]', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
		 ('e2e-wishlist-item-price-drop', ?, 'E2E Wishlist', 'Diecast', 'WISH-DROP-002', 'Wishlist Sample Price Drop Watch', 'wishlist', 'medium', 'Hot Wheels', 'Skyline', '2024', '1:64', 'Wishlist Samples', 'Representative below-target wishlist sample.', '["seed","wishlist","price-drop"]', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
		 ('e2e-wishlist-item-steady', ?, 'E2E Wishlist', 'Comics', 'WISH-WATCH-003', 'Wishlist Sample Steady Watch', 'wishlist', 'low', 'Marvel', 'Spider-Man', '1984', 'Issue', 'Wishlist Samples', 'Representative low-priority wishlist sample.', '["seed","wishlist","watch"]', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, profileID, profileID, profileID); err != nil {
		return e2eBootstrapResponse{}, fmt.Errorf("insert wishlist sample items: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO wishlist_entries(id, profile_id, item_id, target_price, priority, notes, highlight_hit, below_target_now, created_at, updated_at)
		VALUES
		 ('e2e-wishlist-grail', ?, 'e2e-wishlist-item-grail', 150.00, 'high', 'Sample grail chase row for Wishlist review.', 1, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
		 ('e2e-wishlist-price-drop', ?, 'e2e-wishlist-item-price-drop', 35.00, 'medium', 'Sample below-target row for Wishlist review.', 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
		 ('e2e-wishlist-steady', ?, 'e2e-wishlist-item-steady', 12.00, 'low', 'Sample steady watch row for Wishlist review.', 0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, profileID, profileID, profileID); err != nil {
		return e2eBootstrapResponse{}, fmt.Errorf("insert wishlist sample entries: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO tracked_items(item_id, profile_id, created_at)
		VALUES
		 ('e2e-wishlist-item-grail', ?, CURRENT_TIMESTAMP),
		 ('e2e-wishlist-item-price-drop', ?, CURRENT_TIMESTAMP),
		 ('e2e-wishlist-item-steady', ?, CURRENT_TIMESTAMP)
		ON CONFLICT(item_id) DO UPDATE SET profile_id=excluded.profile_id
	`, profileID, profileID, profileID); err != nil {
		return e2eBootstrapResponse{}, fmt.Errorf("insert wishlist sample tracked items: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO price_snapshots(id, item_id, snapshot_date, source, min_price, median_price, latest_price, stock_count, created_at)
		VALUES
		 ('e2e-price-grail-001', 'e2e-wishlist-item-grail', '2026-04-01', 'showcase-market', 130.00, 142.00, 145.00, 4, CURRENT_TIMESTAMP),
		 ('e2e-price-grail-002', 'e2e-wishlist-item-grail', '2026-04-08', 'showcase-market', 134.00, 146.00, 149.00, 5, CURRENT_TIMESTAMP),
		 ('e2e-price-grail-003', 'e2e-wishlist-item-grail', '2026-04-15', 'showcase-market', 138.00, 151.00, 156.00, 3, CURRENT_TIMESTAMP),
		 ('e2e-price-grail-004', 'e2e-wishlist-item-grail', '2026-04-22', 'showcase-market', 142.00, 158.00, 164.00, 2, CURRENT_TIMESTAMP),
		 ('e2e-price-drop-001', 'e2e-wishlist-item-price-drop', '2026-04-01', 'showcase-market', 42.00, 45.00, 48.00, 8, CURRENT_TIMESTAMP),
		 ('e2e-price-drop-002', 'e2e-wishlist-item-price-drop', '2026-04-08', 'showcase-market', 39.00, 42.00, 43.00, 7, CURRENT_TIMESTAMP),
		 ('e2e-price-drop-003', 'e2e-wishlist-item-price-drop', '2026-04-15', 'showcase-market', 34.00, 37.00, 36.00, 6, CURRENT_TIMESTAMP),
		 ('e2e-price-drop-004', 'e2e-wishlist-item-price-drop', '2026-04-22', 'showcase-market', 30.00, 33.00, 31.00, 5, CURRENT_TIMESTAMP),
		 ('e2e-price-steady-001', 'e2e-wishlist-item-steady', '2026-04-01', 'showcase-market', 10.00, 12.00, 12.00, 10, CURRENT_TIMESTAMP),
		 ('e2e-price-steady-002', 'e2e-wishlist-item-steady', '2026-04-08', 'showcase-market', 10.00, 12.00, 12.00, 11, CURRENT_TIMESTAMP),
		 ('e2e-price-steady-003', 'e2e-wishlist-item-steady', '2026-04-15', 'showcase-market', 10.00, 12.00, 12.00, 10, CURRENT_TIMESTAMP),
		 ('e2e-price-steady-004', 'e2e-wishlist-item-steady', '2026-04-22', 'showcase-market', 10.00, 12.00, 12.00, 12, CURRENT_TIMESTAMP)
	`); err != nil {
		return e2eBootstrapResponse{}, fmt.Errorf("insert wishlist sample pricing: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO instances(id, item_id, condition, status, quantity, storage_location, acquisition_price, acquisition_date, notes, created_at, updated_at)
		VALUES
		 ('e2e-instance-001', ?, 'mint', 'sealed', 1, 'Shelf A', 29.99, '2025-01-10', 'seed instance 1', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
		 ('e2e-instance-002', ?, 'excellent', 'loose', 1, 'Shelf B', 19.99, '2025-01-12', 'seed instance 2', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, itemIDs[0], itemIDs[1]); err != nil {
		return e2eBootstrapResponse{}, fmt.Errorf("insert instances: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO scanner_query_sets(id, profile_id, name, keywords_json, exclusions_json, max_price, region, condition_filter, schedule_cron, enabled, rate_limit_rps, max_retry_count, created_at, updated_at)
		VALUES (?, ?, 'E2E Query Set', '["slot car","ho"]', '[]', 120, 'AU', 'used', '', 1, 2, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, querySetID, profileID); err != nil {
		return e2eBootstrapResponse{}, fmt.Errorf("insert query set: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO scanner_candidates(id, profile_id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source, stock_state, stock_count)
		VALUES (?, ?, ?, 'e2e-listing-001', 'E2E Discovery Candidate', 44.95, 9.0, 'https://example.test/e2e-listing-001', '', 'e2e-seller', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'e2e-source', 'in_stock', 5)
	`, "e2e-candidate-001", profileID, querySetID); err != nil {
		return e2eBootstrapResponse{}, fmt.Errorf("insert candidate: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at)
		VALUES ('e2e-candidate-001', ?, 'not_in_collection', 0.9, 0, 'E2E-PN-900', CURRENT_TIMESTAMP)
	`, itemIDs[0]); err != nil {
		return e2eBootstrapResponse{}, fmt.Errorf("insert match: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO chat_threads(id, profile_id, title, created_at, updated_at)
		VALUES (?, ?, 'E2E Thread', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, threadID, profileID); err != nil {
		return e2eBootstrapResponse{}, fmt.Errorf("insert chat thread: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO chat_messages(id, profile_id, thread_id, role, content, attachments_json, created_at)
		VALUES
		 ('e2e-msg-001', ?, ?, 'assistant', 'Welcome to E2E seeded chat.', '[]', CURRENT_TIMESTAMP),
		 ('e2e-msg-002', ?, ?, 'assistant', 'Inventory baseline ready.', '[]', CURRENT_TIMESTAMP)
	`, profileID, threadID, profileID, threadID); err != nil {
		return e2eBootstrapResponse{}, fmt.Errorf("insert chat messages: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO activity_logs(id, level, action, details, created_at)
		VALUES ('e2e-log-001', 'info', 'e2e_bootstrap', ?, CURRENT_TIMESTAMP)
	`, `{"profile_id":"e2e-profile-001"}`); err != nil {
		return e2eBootstrapResponse{}, fmt.Errorf("insert activity log: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return e2eBootstrapResponse{}, fmt.Errorf("commit bootstrap tx: %w", err)
	}

	_ = os.MkdirAll(storageMediaDir, 0o755)
	_ = os.MkdirAll(filepath.Join(cfg.DataDir, "media"), 0o755)

	return e2eBootstrapResponse{
		ProfileID:   profileID,
		ProfileName: profileName,
		ItemIDs:     itemIDs,
		QuerySetID:  querySetID,
		ThreadID:    threadID,
	}, nil
}

func isE2EHooksEnabled(cfg config.Config) bool {
	if cfg.EnableE2EHooks {
		return true
	}
	raw := strings.TrimSpace(os.Getenv("CABINET_E2E_MODE"))
	return raw == "1" || strings.EqualFold(raw, "true")
}

func bootstrapE2EScaleFixtures(ctx context.Context, conn *sql.DB, profile string, seed int64) (e2eScaleBootstrapResponse, error) {
	profileID := "e2e-scale-" + strings.ToLower(profile)
	profileName := "E2E " + profile + " Scale"
	counts := map[string]int{
		"items":      0,
		"discovery":  0,
		"wishlist":   0,
		"query_sets": 1,
	}
	switch profile {
	case "S1":
		counts["items"] = 120
		counts["discovery"] = 60
		counts["wishlist"] = 25
	case "S2":
		counts["items"] = 800
		counts["discovery"] = 260
		counts["wishlist"] = 120
	case "S3":
		counts["items"] = 1600
		counts["discovery"] = 600
		counts["wishlist"] = 240
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return e2eScaleBootstrapResponse{}, fmt.Errorf("begin scale bootstrap tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `INSERT INTO profiles(id, name) VALUES (?, ?)`, profileID, profileName); err != nil {
		return e2eScaleBootstrapResponse{}, fmt.Errorf("insert scale profile: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO app_state(key, value, updated_at)
		VALUES ('active_profile_id', ?, CURRENT_TIMESTAMP), ('clean_shutdown', '1', CURRENT_TIMESTAMP), ('recovery_required', '0', CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=CURRENT_TIMESTAMP
	`, profileID); err != nil {
		return e2eScaleBootstrapResponse{}, fmt.Errorf("set active profile state: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO profile_settings(profile_id, key, value, updated_at)
		VALUES (?, 'scanner_schedule', 'manual', CURRENT_TIMESTAMP)
	`, profileID); err != nil {
		return e2eScaleBootstrapResponse{}, fmt.Errorf("insert scale profile settings: %w", err)
	}

	querySetID := "e2e-scale-queryset-" + strings.ToLower(profile)
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO scanner_query_sets(id, profile_id, name, keywords_json, exclusions_json, max_price, region, condition_filter, schedule_cron, enabled, rate_limit_rps, max_retry_count, created_at, updated_at)
		VALUES (?, ?, ?, '["scale","collector"]', '[]', 250, 'AU', 'used', '', 1, 3, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, querySetID, profileID, "E2E "+profile+" Query Set"); err != nil {
		return e2eScaleBootstrapResponse{}, fmt.Errorf("insert scale query set: %w", err)
	}

	rng := rand.New(rand.NewSource(seed))
	now := time.Now().UTC()
	itemIDs := make([]string, 0, counts["items"])
	for i := 0; i < counts["items"]; i++ {
		itemID := fmt.Sprintf("e2e-scale-item-%s-%05d", strings.ToLower(profile), i+1)
		itemIDs = append(itemIDs, itemID)
		partNumber := fmt.Sprintf("%s-%05d", profile, i+1)
		title := fmt.Sprintf("%s Item %05d", profile, i+1)
		tagSet := fmt.Sprintf("set:%s-core", strings.ToLower(profile))
		tagsJSON := fmt.Sprintf(`["%s","scale:%s","seed:%d"]`, tagSet, strings.ToLower(profile), seed)
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, status, priority, grading_status, grader, grade_numeric, slabbed, collector_classification, car_grade_type, packaging_grade_type, make, model, year, scale, series, description, tags_json, created_at, updated_at, created_by, updated_by)
			VALUES (?, ?, 'E2E Scale', 'Cards', ?, ?, 'active', 'medium', 'ungraded', '', 0, 0, '', '', '', 'Cabinet', ?, '2026', '1:64', ?, ?, ?, ?, ?, 'system', 'system')
		`, itemID, profileID, partNumber, title, "Model "+strconv.Itoa((i%90)+10), profile+" Series", "scale-seed", tagsJSON, now.Format(time.RFC3339), now.Format(time.RFC3339)); err != nil {
			return e2eScaleBootstrapResponse{}, fmt.Errorf("insert scale item: %w", err)
		}
	}

	for i := 0; i < counts["wishlist"] && i < len(itemIDs); i++ {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO wishlist_entries(id, profile_id, item_id, target_price, priority, notes, highlight_hit, created_at, updated_at)
			VALUES (?, ?, ?, ?, 'normal', 'scale bootstrap', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, fmt.Sprintf("e2e-scale-wishlist-%s-%04d", strings.ToLower(profile), i+1), profileID, itemIDs[i], roundPrice(20+rng.Float64()*80)); err != nil {
			return e2eScaleBootstrapResponse{}, fmt.Errorf("insert scale wishlist entry: %w", err)
		}
	}

	for i := 0; i < counts["discovery"]; i++ {
		candidateID := fmt.Sprintf("e2e-scale-candidate-%s-%05d", strings.ToLower(profile), i+1)
		itemID := ""
		if len(itemIDs) > 0 {
			itemID = itemIDs[i%len(itemIDs)]
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO scanner_candidates(id, profile_id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source, stock_state, stock_count)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, '', ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'e2e-scale', 'in_stock', ?)
		`, candidateID, profileID, querySetID, fmt.Sprintf("listing-%s-%05d", strings.ToLower(profile), i+1), fmt.Sprintf("%s Candidate %05d", profile, i+1), roundPrice(10+rng.Float64()*140), roundPrice(1+rng.Float64()*12), fmt.Sprintf("https://example.test/%s/%05d", strings.ToLower(profile), i+1), "e2e-scale-seller", 1+(i%9)); err != nil {
			return e2eScaleBootstrapResponse{}, fmt.Errorf("insert scale candidate: %w", err)
		}
		if itemID != "" {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at)
				VALUES (?, ?, 'suggested', ?, 0, ?, CURRENT_TIMESTAMP)
			`, candidateID, itemID, roundFloat(0.55+rng.Float64()*0.4), fmt.Sprintf("%s-%05d", profile, i+1)); err != nil {
				return e2eScaleBootstrapResponse{}, fmt.Errorf("insert scale match: %w", err)
			}
		}
	}

	hashInput, _ := json.Marshal(map[string]any{
		"profile": profile,
		"seed":    seed,
		"counts":  counts,
	})
	sum := sha256.Sum256(hashInput)
	datasetHash := hex.EncodeToString(sum[:])

	if err := tx.Commit(); err != nil {
		return e2eScaleBootstrapResponse{}, fmt.Errorf("commit scale bootstrap tx: %w", err)
	}
	return e2eScaleBootstrapResponse{
		Profile:     profile,
		Seed:        seed,
		ProfileID:   profileID,
		QuerySetID:  querySetID,
		DatasetHash: datasetHash,
		Counts:      counts,
	}, nil
}

func roundPrice(v float64) float64 {
	return roundFloat(v)
}

func roundFloat(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

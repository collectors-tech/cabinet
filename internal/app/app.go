package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/collectors-tech/cabinet/internal/auth"
	"github.com/collectors-tech/cabinet/internal/backup"
	"github.com/collectors-tech/cabinet/internal/barcode"
	"github.com/collectors-tech/cabinet/internal/collection"
	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/datamgmt"
	"github.com/collectors-tech/cabinet/internal/db"
	"github.com/collectors-tech/cabinet/internal/discovery"
	"github.com/collectors-tech/cabinet/internal/ebay"
	"github.com/collectors-tech/cabinet/internal/matching"
	"github.com/collectors-tech/cabinet/internal/media"
	"github.com/collectors-tech/cabinet/internal/pricing"
	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/collectors-tech/cabinet/internal/scanner"
	"github.com/collectors-tech/cabinet/internal/search"
	"github.com/collectors-tech/cabinet/internal/ui"
	"github.com/collectors-tech/cabinet/internal/wishlist"
)

type App struct {
	cfg       config.Config
	db        *sql.DB
	srv       *http.Server
	backupSvc *backup.Service
}

func New(cfg config.Config) (*App, error) {
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := db.OpenAndMigrate(ctx, cfg.DBPath)
	if err != nil {
		return nil, err
	}

	sub, err := fs.Sub(ui.Static, "static")
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("load embedded ui: %w", err)
	}

	profiles := profile.NewRepository(conn)
	collectionRepo := collection.NewRepository(conn)
	barcodeRepo := barcode.NewRepository(conn)
	mediaService := media.NewService(conn, filepath.Join(cfg.DataDir, "media"))
	dataService := datamgmt.NewService(conn)
	backupSvc := backup.NewService(cfg.DBPath, filepath.Join(cfg.DataDir, "backups"), cfg.BackupInterval)
	searchRepo := search.NewRepository(conn)
	scannerSvc := scanner.NewService(conn)
	matchingSvc := matching.NewService(conn)
	discoverySvc := discovery.NewService(conn)
	wishlistSvc := wishlist.NewService(conn)
	pricingSvc := pricing.NewService(conn)
	authService, err := auth.NewService(cfg, conn, profiles)
	if err != nil {
		conn.Close()
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(sub)))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/runtime", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"update_channel":               cfg.UpdateChannel,
			"update_public_key_configured": cfg.UpdatePublicKey != "",
		})
	})
	mux.HandleFunc("/api/profiles", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			list, err := profiles.List(r.Context())
			if err != nil {
				http.Error(w, `{"error":"failed_to_list_profiles"}`, http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"profiles": list})
		case http.MethodPost:
			var req struct {
				Name string `json:"name"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			created, err := profiles.Create(r.Context(), req.Name)
			if err != nil {
				http.Error(w, `{"error":"invalid_profile"}`, http.StatusBadRequest)
				return
			}
			profileDir := filepath.Join(cfg.DataDir, "profiles", created.ID)
			_ = os.MkdirAll(profileDir, 0o755)
			_ = os.MkdirAll(filepath.Join(profileDir, "media"), 0o755)
			_ = profiles.PutSettings(r.Context(), created.ID, map[string]string{
				"storage.db_path":   filepath.Join(profileDir, "cabinet.db"),
				"storage.media_dir": filepath.Join(profileDir, "media"),
			})
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(created)
		default:
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/profiles/active", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			active, err := profiles.GetActiveProfile(r.Context())
			if err != nil {
				http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(active)
		case http.MethodPut:
			var req struct {
				ProfileID string `json:"profile_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			if err := profiles.SetActiveProfile(r.Context(), req.ProfileID); err != nil {
				http.Error(w, `{"error":"invalid_profile_id"}`, http.StatusBadRequest)
				return
			}
			active, _ := profiles.GetActiveProfile(r.Context())
			_ = json.NewEncoder(w).Encode(active)
		default:
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/profiles/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := strings.TrimPrefix(r.URL.Path, "/api/profiles/")
		parts := strings.Split(path, "/")
		if len(parts) != 2 {
			http.Error(w, `{"error":"not_found"}`, http.StatusNotFound)
			return
		}
		profileID := strings.TrimSpace(parts[0])
		if profileID == "" {
			http.Error(w, `{"error":"invalid_profile_id"}`, http.StatusBadRequest)
			return
		}

		switch parts[1] {
		case "settings":
			switch r.Method {
			case http.MethodGet:
				settings, err := profiles.GetSettings(r.Context(), profileID)
				if err != nil {
					http.Error(w, `{"error":"failed_to_get_settings"}`, http.StatusBadRequest)
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"settings": settings})
			case http.MethodPut:
				var req struct {
					Settings map[string]string `json:"settings"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
					return
				}
				if err := profiles.PutSettings(r.Context(), profileID, req.Settings); err != nil {
					http.Error(w, `{"error":"failed_to_update_settings"}`, http.StatusBadRequest)
					return
				}
				settings, _ := profiles.GetSettings(r.Context(), profileID)
				_ = json.NewEncoder(w).Encode(map[string]any{"settings": settings})
			default:
				http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			}
		case "saved-filters":
			switch r.Method {
			case http.MethodGet:
				items, err := searchRepo.ListSavedFilters(r.Context(), profileID)
				if err != nil {
					http.Error(w, `{"error":"failed_to_list_saved_filters"}`, http.StatusBadRequest)
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"saved_filters": items})
			case http.MethodPost:
				var req struct {
					Name  string       `json:"name"`
					Query search.Query `json:"query"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
					return
				}
				created, err := searchRepo.SaveFilter(r.Context(), profileID, req.Name, req.Query)
				if err != nil {
					http.Error(w, `{"error":"failed_to_save_filter"}`, http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(created)
			case http.MethodPut:
				var req struct {
					ID    string       `json:"id"`
					Name  string       `json:"name"`
					Query search.Query `json:"query"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
					return
				}
				updated, err := searchRepo.UpdateFilter(r.Context(), req.ID, profileID, req.Name, req.Query)
				if err != nil {
					http.Error(w, `{"error":"failed_to_update_filter"}`, http.StatusBadRequest)
					return
				}
				_ = json.NewEncoder(w).Encode(updated)
			case http.MethodDelete:
				filterID := strings.TrimSpace(r.URL.Query().Get("id"))
				if filterID == "" {
					http.Error(w, `{"error":"missing_filter_id"}`, http.StatusBadRequest)
					return
				}
				if err := searchRepo.DeleteFilter(r.Context(), filterID, profileID); err != nil {
					http.Error(w, `{"error":"failed_to_delete_filter"}`, http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusNoContent)
			default:
				http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			}
		case "storage":
			if r.Method != http.MethodGet {
				http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			settings, err := profiles.GetSettings(r.Context(), profileID)
			if err != nil {
				http.Error(w, `{"error":"failed_to_get_storage"}`, http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"db_path":   settings["storage.db_path"],
				"media_dir": settings["storage.media_dir"],
			})
		case "secrets":
			if r.Method == http.MethodPut {
				var req struct {
					Key   string `json:"key"`
					Value string `json:"value"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
					return
				}
				if err := profiles.PutSecret(r.Context(), profileID, req.Key, req.Value); err != nil {
					http.Error(w, `{"error":"failed_to_put_secret"}`, http.StatusBadRequest)
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
				return
			}
			if r.Method == http.MethodGet {
				key := strings.TrimSpace(r.URL.Query().Get("key"))
				value, err := profiles.GetSecret(r.Context(), profileID, key)
				if err != nil {
					http.Error(w, `{"error":"failed_to_get_secret"}`, http.StatusBadRequest)
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"key": key, "value": value})
				return
			}
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		case "license":
			if r.Method == http.MethodPut {
				var req struct {
					LicenseJSON string `json:"license_json"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
					return
				}
				if err := profiles.PutLicense(r.Context(), profileID, req.LicenseJSON); err != nil {
					http.Error(w, `{"error":"failed_to_put_license"}`, http.StatusBadRequest)
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
				return
			}
			if r.Method == http.MethodGet {
				value, err := profiles.GetLicense(r.Context(), profileID)
				if err != nil {
					http.Error(w, `{"error":"failed_to_get_license"}`, http.StatusBadRequest)
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"license_json": value})
				return
			}
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		default:
			http.Error(w, `{"error":"not_found"}`, http.StatusNotFound)
		}
	})
	mux.HandleFunc("/api/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			items, err := collectionRepo.ListItems(r.Context())
			if err != nil {
				http.Error(w, `{"error":"failed_to_list_items"}`, http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
		case http.MethodPost:
			var req collection.Item
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			created, err := collectionRepo.CreateItem(r.Context(), req)
			if err != nil {
				http.Error(w, `{"error":"invalid_item"}`, http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(created)
		default:
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/barcodes/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		rest := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/barcodes/"))
		parts := strings.Split(rest, "/")
		code := strings.TrimSpace(parts[0])
		if code == "" {
			http.Error(w, `{"error":"invalid_barcode"}`, http.StatusBadRequest)
			return
		}
		if len(parts) == 2 && parts[1] == "external-search" {
			url, err := barcode.BuildExternalSearchURL(r.URL.Query().Get("source"), r.URL.Query().Get("region"), code)
			if err != nil {
				http.Error(w, `{"error":"failed_to_build_external_search_url"}`, http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"source": "ebay",
				"url":    url,
			})
			return
		}
		matches, err := barcodeRepo.Lookup(r.Context(), code)
		if err != nil {
			http.Error(w, `{"error":"failed_to_lookup_barcode"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"matches": matches})
	})
	mux.HandleFunc("/api/search/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		items, err := searchRepo.SearchItems(r.Context(), search.Query{
			Text:      r.URL.Query().Get("q"),
			Brand:     r.URL.Query().Get("brand"),
			Category:  r.URL.Query().Get("category"),
			Condition: r.URL.Query().Get("condition"),
			Status:    r.URL.Query().Get("status"),
			Tags:      r.URL.Query().Get("tags"),
			Scale:     r.URL.Query().Get("scale"),
			SortBy:    r.URL.Query().Get("sort"),
			Limit:     limit,
		})
		if err != nil {
			http.Error(w, `{"error":"failed_to_search_items"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
	})
	mux.HandleFunc("/api/scanner/query-sets", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			items, err := scannerSvc.ListQuerySets(r.Context())
			if err != nil {
				http.Error(w, `{"error":"failed_to_list_query_sets"}`, http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"query_sets": items})
		case http.MethodPost:
			var req scanner.QuerySet
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			created, err := scannerSvc.CreateQuerySet(r.Context(), req)
			if err != nil {
				http.Error(w, `{"error":"invalid_query_set"}`, http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(created)
		default:
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/scanner/run", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			QuerySetID string `json:"query_set_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil {
			http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusBadRequest)
			return
		}
		settings, err := profiles.GetSettings(r.Context(), active.ID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_get_settings"}`, http.StatusBadRequest)
			return
		}
		provider := ebay.NewProvider(ebay.ProviderConfig{
			BaseURL:     settings["ebay_base_url"],
			BearerToken: settings["ebay_bearer_token"],
			Marketplace: settings["ebay_marketplace"],
		})
		out, err := scannerSvc.RunNow(r.Context(), req.QuerySetID, provider)
		if err != nil {
			http.Error(w, `{"error":"failed_to_run_scanner"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(out)
	})
	mux.HandleFunc("/api/scanner/run/scheduled", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil {
			http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusBadRequest)
			return
		}
		settings, err := profiles.GetSettings(r.Context(), active.ID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_get_settings"}`, http.StatusBadRequest)
			return
		}
		provider := ebay.NewProvider(ebay.ProviderConfig{
			BaseURL:     settings["ebay_base_url"],
			BearerToken: settings["ebay_bearer_token"],
			Marketplace: settings["ebay_marketplace"],
		})
		ran, err := scannerSvc.RunScheduled(r.Context(), provider)
		if err != nil {
			http.Error(w, `{"error":"failed_to_run_scheduled_scanner"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"runs": ran})
	})
	mux.HandleFunc("/api/scanner/candidates", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		querySetID := strings.TrimSpace(r.URL.Query().Get("query_set_id"))
		if querySetID == "" {
			http.Error(w, `{"error":"missing_query_set_id"}`, http.StatusBadRequest)
			return
		}
		items, err := scannerSvc.ListCandidates(r.Context(), querySetID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_list_candidates"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"candidates": items})
	})
	mux.HandleFunc("/api/scanner/failures", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		items, err := scannerSvc.ListFailures(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed_to_list_failures"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"failures": items})
	})
	mux.HandleFunc("/api/provider/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		provider := strings.TrimSpace(r.URL.Query().Get("provider"))
		if provider == "" {
			provider = "ebay"
		}
		health, err := scannerSvc.ProviderHealth(r.Context(), provider)
		if err != nil {
			http.Error(w, `{"error":"failed_to_get_provider_health"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(health)
	})
	mux.HandleFunc("/api/matching/run", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		if err := matchingSvc.Run(r.Context()); err != nil {
			http.Error(w, `{"error":"failed_to_run_matching"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	mux.HandleFunc("/api/matching/results", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		items, err := matchingSvc.List(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed_to_list_matching_results"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"results": items})
	})
	mux.HandleFunc("/api/discovery/not-in-collection", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		priceMax, _ := strconv.ParseFloat(r.URL.Query().Get("price_max"), 64)
		items, err := discoverySvc.ListNotInCollection(r.Context(), discovery.Filter{
			Query:    r.URL.Query().Get("q"),
			PriceMax: priceMax,
			DateFrom: r.URL.Query().Get("date_from"),
		})
		if err != nil {
			http.Error(w, `{"error":"failed_to_list_not_in_collection"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
	})
	mux.HandleFunc("/api/discovery/action", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req discovery.Action
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		if err := discoverySvc.ApplyAction(r.Context(), req); err != nil {
			http.Error(w, `{"error":"failed_to_apply_discovery_action"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	mux.HandleFunc("/api/settings/reset-ignore-rules", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		if err := discoverySvc.ResetIgnored(r.Context()); err != nil {
			http.Error(w, `{"error":"failed_to_reset_ignore_rules"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	mux.HandleFunc("/api/wishlist", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			items, err := wishlistSvc.List(r.Context())
			if err != nil {
				http.Error(w, `{"error":"failed_to_list_wishlist"}`, http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
		case http.MethodPost:
			var req wishlist.Entry
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			created, err := wishlistSvc.Create(r.Context(), req)
			if err != nil {
				http.Error(w, `{"error":"failed_to_create_wishlist"}`, http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(created)
		case http.MethodPut:
			var req wishlist.Entry
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			if err := wishlistSvc.Update(r.Context(), req); err != nil {
				http.Error(w, `{"error":"failed_to_update_wishlist"}`, http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
		case http.MethodDelete:
			id := strings.TrimSpace(r.URL.Query().Get("id"))
			if id == "" {
				http.Error(w, `{"error":"missing_id"}`, http.StatusBadRequest)
				return
			}
			if err := wishlistSvc.Delete(r.Context(), id); err != nil {
				http.Error(w, `{"error":"failed_to_delete_wishlist"}`, http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/wishlist/hits", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		itemID := strings.TrimSpace(r.URL.Query().Get("item_id"))
		hits, err := wishlistSvc.Hits(r.Context(), itemID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_list_wishlist_hits"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"hits": hits})
	})
	mux.HandleFunc("/api/pricing/track", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ItemID string `json:"item_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		if err := pricingSvc.TrackItem(r.Context(), req.ItemID); err != nil {
			http.Error(w, `{"error":"failed_to_track_item"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	mux.HandleFunc("/api/pricing/snapshot/run", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		if err := pricingSvc.RunDailySnapshot(r.Context()); err != nil {
			http.Error(w, `{"error":"failed_to_run_price_snapshot"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	mux.HandleFunc("/api/pricing/history", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		itemID := strings.TrimSpace(r.URL.Query().Get("item_id"))
		history, err := pricingSvc.History(r.Context(), itemID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_get_price_history"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"history": history})
	})
	mux.HandleFunc("/api/pricing/graph", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		itemID := strings.TrimSpace(r.URL.Query().Get("item_id"))
		history, err := pricingSvc.History(r.Context(), itemID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_get_price_graph"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"points": history})
	})
	mux.HandleFunc("/api/pricing/by-source", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		itemID := strings.TrimSpace(r.URL.Query().Get("item_id"))
		bySource, err := pricingSvc.BySource(r.Context(), itemID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_get_price_by_source"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"sources": bySource})
	})
	mux.HandleFunc("/api/pricing/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		itemID := strings.TrimSpace(r.URL.Query().Get("item_id"))
		history, err := pricingSvc.History(r.Context(), itemID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_get_price_stats"}`, http.StatusBadRequest)
			return
		}
		if len(history) == 0 {
			_ = json.NewEncoder(w).Encode(map[string]any{"min": 0, "median": 0, "latest": 0})
			return
		}
		latest := history[len(history)-1]
		_ = json.NewEncoder(w).Encode(map[string]any{
			"min":    latest.MinPrice,
			"median": latest.MedianPrice,
			"latest": latest.LatestPrice,
		})
	})
	mux.HandleFunc("/api/pricing/history/export", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		itemID := strings.TrimSpace(r.URL.Query().Get("item_id"))
		csvText, err := pricingSvc.ExportCSV(r.Context(), itemID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_export_price_history"}`, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(csvText))
	})
	mux.HandleFunc("/api/data/export/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		snap, err := dataService.ExportSnapshot(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed_to_export_snapshot"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(snap)
	})
	mux.HandleFunc("/api/data/export/csv/items", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		csvText, err := dataService.ExportItemsCSV(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed_to_export_csv"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(csvText))
	})
	mux.HandleFunc("/api/data/import/json/dry-run", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Snapshot datamgmt.Snapshot `json:"snapshot"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		sum, err := dataService.DryRunImport(r.Context(), req.Snapshot)
		if err != nil {
			http.Error(w, `{"error":"failed_to_dry_run_import"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(sum)
	})
	mux.HandleFunc("/api/data/import/json/apply", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Snapshot datamgmt.Snapshot     `json:"snapshot"`
			Options  datamgmt.ApplyOptions `json:"options"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		if err := dataService.ApplyImport(r.Context(), req.Snapshot, req.Options); err != nil {
			http.Error(w, `{"error":"failed_to_apply_import"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	mux.HandleFunc("/api/data/import/csv/dry-run", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req datamgmt.CSVImportRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		snap, err := dataService.ParseCSVToSnapshot(req)
		if err != nil {
			http.Error(w, `{"error":"failed_to_parse_csv"}`, http.StatusBadRequest)
			return
		}
		sum, err := dataService.DryRunImport(r.Context(), snap)
		if err != nil {
			http.Error(w, `{"error":"failed_to_dry_run_import"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(sum)
	})
	mux.HandleFunc("/api/data/import/csv/apply", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			CSVImportRequest datamgmt.CSVImportRequest `json:"csv_import"`
			Options          datamgmt.ApplyOptions     `json:"options"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		snap, err := dataService.ParseCSVToSnapshot(req.CSVImportRequest)
		if err != nil {
			http.Error(w, `{"error":"failed_to_parse_csv"}`, http.StatusBadRequest)
			return
		}
		if err := dataService.ApplyImport(r.Context(), snap, req.Options); err != nil {
			http.Error(w, `{"error":"failed_to_apply_import"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	mux.HandleFunc("/api/data/reindex", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		if err := dataService.Reindex(r.Context()); err != nil {
			http.Error(w, `{"error":"failed_to_reindex"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	mux.HandleFunc("/api/data/repair", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		result, err := dataService.Repair(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed_to_repair_check"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"integrity_check": result})
	})
	mux.HandleFunc("/api/backup/run", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		path, err := backupSvc.CreateBackup(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed_to_create_backup"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"backup_path": path})
	})
	mux.HandleFunc("/api/backup/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		backups, err := backupSvc.ListBackups()
		if err != nil {
			http.Error(w, `{"error":"failed_to_list_backups"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"backups": backups})
	})
	mux.HandleFunc("/api/backup/restore", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			BackupPath string `json:"backup_path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		if err := backupSvc.RestoreBackup(req.BackupPath); err != nil {
			http.Error(w, `{"error":"failed_to_restore_backup"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	mux.HandleFunc("/api/items/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := strings.TrimPrefix(r.URL.Path, "/api/items/")
		parts := strings.Split(path, "/")
		if len(parts) < 2 {
			http.Error(w, `{"error":"not_found"}`, http.StatusNotFound)
			return
		}
		itemID := strings.TrimSpace(parts[0])
		if itemID == "" {
			http.Error(w, `{"error":"invalid_item_id"}`, http.StatusBadRequest)
			return
		}

		switch parts[1] {
		case "barcodes":
			switch r.Method {
			case http.MethodGet:
				records, err := barcodeRepo.ListByItem(r.Context(), itemID)
				if err != nil {
					http.Error(w, `{"error":"failed_to_list_barcodes"}`, http.StatusInternalServerError)
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"barcodes": records})
			case http.MethodPost:
				var req struct {
					Barcode string `json:"barcode"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
					return
				}
				created, err := barcodeRepo.Add(r.Context(), itemID, req.Barcode)
				if err != nil {
					http.Error(w, `{"error":"invalid_barcode"}`, http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(created)
			default:
				http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			}
		case "instances":
			switch r.Method {
			case http.MethodGet:
				instances, err := collectionRepo.ListInstancesByItemID(r.Context(), itemID)
				if err != nil {
					http.Error(w, `{"error":"failed_to_list_instances"}`, http.StatusInternalServerError)
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"instances": instances})
			case http.MethodPost:
				var req collection.Instance
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
					return
				}
				req.ItemID = itemID
				created, err := collectionRepo.CreateInstance(r.Context(), req)
				if err != nil {
					http.Error(w, `{"error":"invalid_instance"}`, http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(created)
			default:
				http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			}
		case "photos":
			if len(parts) == 2 {
				switch r.Method {
				case http.MethodGet:
					photos, err := mediaService.ListByItem(r.Context(), itemID)
					if err != nil {
						http.Error(w, `{"error":"failed_to_list_photos"}`, http.StatusInternalServerError)
						return
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"photos": photos})
				case http.MethodPost:
					if err := r.ParseMultipartForm(32 << 20); err != nil {
						http.Error(w, `{"error":"invalid_multipart_form"}`, http.StatusBadRequest)
						return
					}
					file, hdr, err := r.FormFile("file")
					if err != nil {
						http.Error(w, `{"error":"missing_file"}`, http.StatusBadRequest)
						return
					}
					defer file.Close()
					created, err := mediaService.Upload(r.Context(), itemID, hdr.Filename, file)
					if err != nil {
						http.Error(w, `{"error":"failed_to_upload_photo"}`, http.StatusBadRequest)
						return
					}
					w.WriteHeader(http.StatusCreated)
					_ = json.NewEncoder(w).Encode(created)
				default:
					http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
				}
				return
			}

			if len(parts) == 3 {
				photoID := strings.TrimSpace(parts[2])
				if photoID == "" {
					http.Error(w, `{"error":"invalid_photo_id"}`, http.StatusBadRequest)
					return
				}
				if r.Method == http.MethodDelete {
					if err := mediaService.Delete(r.Context(), itemID, photoID); err != nil {
						http.Error(w, `{"error":"failed_to_delete_photo"}`, http.StatusBadRequest)
						return
					}
					w.WriteHeader(http.StatusNoContent)
					return
				}
				http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
				return
			}

			if len(parts) == 4 && parts[3] == "primary" {
				photoID := strings.TrimSpace(parts[2])
				if photoID == "" {
					http.Error(w, `{"error":"invalid_photo_id"}`, http.StatusBadRequest)
					return
				}
				if r.Method != http.MethodPut {
					http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
					return
				}
				if err := mediaService.SetPrimary(r.Context(), itemID, photoID); err != nil {
					http.Error(w, `{"error":"failed_to_set_primary_photo"}`, http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
			if len(parts) == 4 && parts[3] == "file" {
				photoID := strings.TrimSpace(parts[2])
				if photoID == "" {
					http.Error(w, `{"error":"invalid_photo_id"}`, http.StatusBadRequest)
					return
				}
				if r.Method != http.MethodGet {
					http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
					return
				}
				path, err := mediaService.ResolveVariantPath(r.Context(), itemID, photoID, r.URL.Query().Get("variant"))
				if err != nil {
					http.Error(w, `{"error":"failed_to_resolve_photo_variant"}`, http.StatusBadRequest)
					return
				}
				w.Header().Del("Content-Type")
				http.ServeFile(w, r, path)
				return
			}
			http.Error(w, `{"error":"not_found"}`, http.StatusNotFound)
		case "photos-rebuild":
			if r.Method != http.MethodPost {
				http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			if err := mediaService.RebuildThumbnails(r.Context(), itemID); err != nil {
				http.Error(w, `{"error":"failed_to_rebuild_thumbnails"}`, http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
		default:
			http.Error(w, `{"error":"not_found"}`, http.StatusNotFound)
		}
	})
	mux.HandleFunc("/api/auth/webauthn/register/begin", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ProfileID string `json:"profile_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		resp, err := authService.BeginRegistration(r.Context(), req.ProfileID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_begin_registration"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/api/auth/webauthn/register/finish", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			SessionID  string `json:"session_id"`
			Credential any    `json:"credential"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		if err := authService.FinishRegistration(r.Context(), req.SessionID, req.Credential); err != nil {
			http.Error(w, `{"error":"failed_to_finish_registration"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	mux.HandleFunc("/api/auth/requirements", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		profileID := strings.TrimSpace(r.URL.Query().Get("profile_id"))
		required, err := authService.RequiresRegistration(r.Context(), profileID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_get_requirements"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"requires_registration": required})
	})
	mux.HandleFunc("/api/auth/recovery/passphrase", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ProfileID  string `json:"profile_id"`
			Passphrase string `json:"passphrase"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		if err := authService.SetRecoveryPassphrase(r.Context(), req.ProfileID, req.Passphrase); err != nil {
			http.Error(w, `{"error":"failed_to_set_recovery_passphrase"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	mux.HandleFunc("/api/auth/recovery/reset/begin", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ProfileID  string `json:"profile_id"`
			Passphrase string `json:"passphrase"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		resp, err := authService.BeginRecoveryRegistration(r.Context(), req.ProfileID, req.Passphrase)
		if err != nil {
			http.Error(w, `{"error":"failed_to_begin_recovery_reset"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/api/auth/webauthn/login/begin", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ProfileID string `json:"profile_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		resp, err := authService.BeginLogin(r.Context(), req.ProfileID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_begin_login"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/api/auth/webauthn/login/finish", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			SessionID  string `json:"session_id"`
			Credential any    `json:"credential"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		if err := authService.FinishLogin(r.Context(), req.SessionID, req.Credential); err != nil {
			http.Error(w, `{"error":"failed_to_finish_login"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	mux.HandleFunc("/api/auth/session/validate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			SessionToken string `json:"session_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		if err := authService.ValidateUnlockedSession(req.SessionToken); err != nil {
			http.Error(w, `{"error":"session_locked"}`, http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	mux.HandleFunc("/api/auth/session/lock", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			SessionToken string `json:"session_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		authService.LockUnlockedSession(req.SessionToken)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	a := &App{
		cfg:       cfg,
		db:        conn,
		srv:       srv,
		backupSvc: backupSvc,
	}

	return a, nil
}

func (a *App) Run(ctx context.Context) error {
	if a.backupSvc != nil {
		a.backupSvc.Start(ctx)
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- a.srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = a.srv.Shutdown(shutdownCtx)
		return a.close()
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return a.close()
		}
		_ = a.close()
		return err
	}
}

func (a *App) close() error {
	if a.db == nil {
		return nil
	}
	return a.db.Close()
}

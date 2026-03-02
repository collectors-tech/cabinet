package app

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/collectors-tech/cabinet/internal/ai"
	"github.com/collectors-tech/cabinet/internal/auth"
	"github.com/collectors-tech/cabinet/internal/backup"
	"github.com/collectors-tech/cabinet/internal/barcode"
	"github.com/collectors-tech/cabinet/internal/chat"
	"github.com/collectors-tech/cabinet/internal/collection"
	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/dashboard"
	"github.com/collectors-tech/cabinet/internal/datamgmt"
	"github.com/collectors-tech/cabinet/internal/db"
	"github.com/collectors-tech/cabinet/internal/discovery"
	"github.com/collectors-tech/cabinet/internal/ebay"
	"github.com/collectors-tech/cabinet/internal/licensing"
	"github.com/collectors-tech/cabinet/internal/logging"
	"github.com/collectors-tech/cabinet/internal/matching"
	"github.com/collectors-tech/cabinet/internal/media"
	"github.com/collectors-tech/cabinet/internal/pricing"
	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/collectors-tech/cabinet/internal/scanner"
	"github.com/collectors-tech/cabinet/internal/search"
	"github.com/collectors-tech/cabinet/internal/ui"
	"github.com/collectors-tech/cabinet/internal/update"
	"github.com/collectors-tech/cabinet/internal/wishlist"
)

type App struct {
	cfg         config.Config
	db          *sql.DB
	srv         *http.Server
	backupSvc   *backup.Service
	authService *auth.Service
	openapiSpec []byte
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
	dashboardSvc := dashboard.NewService(conn)
	chatSvc := chat.NewService(conn, filepath.Join(cfg.DataDir, "chat-attachments"))
	aiSvc := ai.NewService(ai.Config{})
	licenseSvc := licensing.NewService(conn, profiles, cfg.UpdatePublicKey)
	logSvc := logging.NewService(conn)
	authService, err := auth.NewService(cfg, conn, profiles)
	if err != nil {
		conn.Close()
		return nil, err
	}
	cloudLeases := newCloudLeaseStore()
	cloudEntitlements := newCloudEntitlementStore()

	mux := http.NewServeMux()
	if isE2EHooksEnabled(cfg) {
		registerE2ETestHooks(mux, conn, cfg)
	}
	var previousClean string
	_ = conn.QueryRowContext(ctx, `SELECT value FROM app_state WHERE key = 'clean_shutdown'`).Scan(&previousClean)
	if previousClean == "0" {
		_, _ = conn.ExecContext(ctx, `
			INSERT INTO app_state(key, value, updated_at)
			VALUES('recovery_required', '1', CURRENT_TIMESTAMP)
			ON CONFLICT(key) DO UPDATE SET value='1', updated_at=CURRENT_TIMESTAMP
		`)
	}
	_, _ = conn.ExecContext(ctx, `
		INSERT INTO app_state(key, value, updated_at)
		VALUES('clean_shutdown', '0', CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET value='0', updated_at=CURRENT_TIMESTAMP
	`)
	openapiSpec := loadOpenAPISpec(cfg)
	mux.HandleFunc("/apidocs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(apiDocsHTML))
	})
	mux.HandleFunc("/redoc.html", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		http.Redirect(w, r, "/apidocs", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/api/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		if len(openapiSpec) == 0 {
			http.Error(w, "openapi spec not available", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		_, _ = w.Write(openapiSpec)
	})
	fileServer := http.FileServer(http.FS(sub))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			fileServer.ServeHTTP(w, r)
			return
		}

		cleanPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if cleanPath == "." {
			cleanPath = ""
		}

		if cleanPath == "" || strings.Contains(cleanPath, ".") {
			fileServer.ServeHTTP(w, r)
			return
		}

		if _, err := fs.Stat(sub, cleanPath); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		rr := *r
		rr.URL = &url.URL{}
		*rr.URL = *r.URL
		rr.URL.Path = "/"
		fileServer.ServeHTTP(w, &rr)
	}))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/runtime", func(w http.ResponseWriter, _ *http.Request) {
		appVersion, buildDate := runtimeBuildMetadata()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"update_channel":               cfg.UpdateChannel,
			"update_public_key_configured": cfg.UpdatePublicKey != "",
			"app_version":                  appVersion,
			"build_date":                   buildDate,
		})
	})
	mux.HandleFunc("/api/runtime/update/install", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			PayloadBase64   string `json:"payload_base64"`
			SignatureBase64 string `json:"signature_base64"`
			ManifestChannel string `json:"manifest_channel"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(cfg.UpdatePublicKey) == "" {
			writeUpdateInstallError(w, http.StatusBadRequest, "UPDATE_SIGNATURE_UNAVAILABLE", "runtime update signature verification is not configured")
			return
		}
		payload, err := base64.StdEncoding.DecodeString(strings.TrimSpace(req.PayloadBase64))
		if err != nil {
			writeUpdateInstallError(w, http.StatusBadRequest, "INVALID_UPDATE_PAYLOAD", "unable to decode update payload")
			return
		}
		if err := update.VerifySignature(cfg.UpdatePublicKey, payload, strings.TrimSpace(req.SignatureBase64)); err != nil {
			writeUpdateInstallError(w, http.StatusBadRequest, "INVALID_UPDATE_SIGNATURE", "update signature verification failed")
			return
		}
		runtimeChannel := update.ParseChannel(string(cfg.UpdateChannel))
		manifestChannel := update.ParseChannel(strings.TrimSpace(strings.ToLower(req.ManifestChannel)))
		if manifestChannel != runtimeChannel {
			writeUpdateInstallError(w, http.StatusConflict, "UPDATE_CHANNEL_MISMATCH", fmt.Sprintf("manifest channel %s does not match runtime channel %s", manifestChannel, runtimeChannel))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":       true,
			"verified": true,
			"channel":  runtimeChannel,
			"action":   "install_approved",
		})
	})
	mux.HandleFunc("/api/runtime/recovery", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var value string
		_ = conn.QueryRowContext(r.Context(), `SELECT value FROM app_state WHERE key = 'recovery_required'`).Scan(&value)
		_ = json.NewEncoder(w).Encode(map[string]any{"recovery_required": value == "1"})
	})
	mux.HandleFunc("/api/diagnostics/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			cfg := diagnosticsConfigFromDB(r.Context(), conn)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"remote_opt_in":      cfg.RemoteOptIn,
				"provider":           cfg.Provider,
				"remote_url":         cfg.RemoteURL,
				"sentry_compatible":  true,
				"local_storage_only": !cfg.RemoteOptIn,
			})
		case http.MethodPut:
			var req struct {
				RemoteOptIn bool   `json:"remote_opt_in"`
				Provider    string `json:"provider"`
				RemoteURL   string `json:"remote_url"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			cfg := diagnosticsConfig{
				RemoteOptIn: req.RemoteOptIn,
				Provider:    strings.TrimSpace(strings.ToLower(req.Provider)),
				RemoteURL:   strings.TrimSpace(req.RemoteURL),
			}
			if cfg.Provider == "" {
				cfg.Provider = "sentry"
			}
			if err := persistDiagnosticsConfig(r.Context(), conn, cfg); err != nil {
				http.Error(w, `{"error":"failed_to_update_diagnostics_config"}`, http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"remote_opt_in":      cfg.RemoteOptIn,
				"provider":           cfg.Provider,
				"remote_url":         cfg.RemoteURL,
				"sentry_compatible":  true,
				"local_storage_only": !cfg.RemoteOptIn,
			})
		default:
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/diagnostics/event", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Type      string         `json:"type"`
			Category  string         `json:"category"`
			Message   string         `json:"message"`
			SessionID string         `json:"session_id"`
			Details   map[string]any `json:"details"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		cfg := diagnosticsConfigFromDB(r.Context(), conn)
		if req.Details == nil {
			req.Details = map[string]any{}
		}
		logSvc.Log(r.Context(), "error", "diagnostics_event_local", map[string]any{
			"type":       req.Type,
			"category":   req.Category,
			"message":    req.Message,
			"session_id": req.SessionID,
			"details":    req.Details,
		})
		remoteSent := false
		remoteStatus := "local_only"
		if cfg.RemoteOptIn {
			envelope := buildSentryCompatibleEnvelope(req.Type, req.Category, req.Message, req.SessionID, req.Details)
			if err := sendDiagnosticsEnvelope(r.Context(), cfg, envelope); err != nil {
				remoteStatus = "remote_error"
			} else {
				remoteSent = true
				remoteStatus = "remote_sent"
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"accepted":        true,
			"remote_sent":     remoteSent,
			"delivery_status": remoteStatus,
		})
	})
	mux.HandleFunc("/api/errors/classify", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ErrorCode string `json:"error_code"`
			Message   string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		category, nextAction, guidance := classifyErrorTaxonomy(req.ErrorCode, req.Message)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"category":    category,
			"next_action": nextAction,
			"guidance":    guidance,
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
	registerUsersRoutes(mux, conn, profiles)
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
	mux.HandleFunc("/api/inventory/grading/enums", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil || strings.TrimSpace(active.ID) == "" {
			http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			settings, getErr := profiles.GetSettings(r.Context(), active.ID)
			if getErr != nil {
				http.Error(w, `{"error":"failed_to_get_grading_enums"}`, http.StatusBadRequest)
				return
			}
			condition := parseStringArraySetting(settings["grading.enums.condition"], defaultConditionGrades())
			packaging := parseStringArraySetting(settings["grading.enums.packaging"], defaultPackagingGrades())
			_ = json.NewEncoder(w).Encode(map[string]any{
				"condition_grades": condition,
				"packaging_grades": packaging,
			})
		case http.MethodPut:
			var req struct {
				ConditionGrades []string `json:"condition_grades"`
				PackagingGrades []string `json:"packaging_grades"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			condition := normalizeStringList(req.ConditionGrades, defaultConditionGrades())
			packaging := normalizeStringList(req.PackagingGrades, defaultPackagingGrades())
			conditionRaw, _ := json.Marshal(condition)
			packagingRaw, _ := json.Marshal(packaging)
			if putErr := profiles.PutSettings(r.Context(), active.ID, map[string]string{
				"grading.enums.condition": string(conditionRaw),
				"grading.enums.packaging": string(packagingRaw),
			}); putErr != nil {
				http.Error(w, `{"error":"failed_to_save_grading_enums"}`, http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"condition_grades": condition,
				"packaging_grades": packaging,
			})
		default:
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/inventory/grading/defaults", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil || strings.TrimSpace(active.ID) == "" {
			http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			settings, getErr := profiles.GetSettings(r.Context(), active.ID)
			if getErr != nil {
				http.Error(w, `{"error":"failed_to_get_grading_defaults"}`, http.StatusBadRequest)
				return
			}
			gradingStatus := strings.TrimSpace(settings["grading.defaults.grading_status"])
			if gradingStatus == "" {
				gradingStatus = "ungraded"
			}
			priority := strings.TrimSpace(settings["grading.defaults.priority"])
			if priority == "" {
				priority = "medium"
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"grading_status": strings.ToLower(gradingStatus),
				"priority":       strings.ToLower(priority),
			})
		case http.MethodPut:
			var req struct {
				GradingStatus string `json:"grading_status"`
				Priority      string `json:"priority"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			gradingStatus := strings.ToLower(strings.TrimSpace(req.GradingStatus))
			if gradingStatus == "" {
				gradingStatus = "ungraded"
			}
			priority := strings.ToLower(strings.TrimSpace(req.Priority))
			if priority == "" {
				priority = "medium"
			}
			if putErr := profiles.PutSettings(r.Context(), active.ID, map[string]string{
				"grading.defaults.grading_status": gradingStatus,
				"grading.defaults.priority":       priority,
			}); putErr != nil {
				http.Error(w, `{"error":"failed_to_save_grading_defaults"}`, http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"grading_status": gradingStatus,
				"priority":       priority,
			})
		default:
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			items, err := collectionRepo.ListItems(r.Context())
			if active, activeErr := profiles.GetActiveProfile(r.Context()); activeErr == nil && strings.TrimSpace(active.ID) != "" {
				items, err = collectionRepo.ListItemsByProfile(r.Context(), active.ID)
			}
			if err != nil {
				http.Error(w, `{"error":"failed_to_list_items"}`, http.StatusInternalServerError)
				return
			}
			if items == nil {
				items = make([]collection.Item, 0)
			}
			requestedStatus := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("status")))
			if requestedStatus == "" {
				requestedStatus = "active"
			}
			items = filterItemsByLifecycleStatus(items, requestedStatus)
			_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
		case http.MethodPost:
			if strings.TrimSpace(cfg.UpdatePublicKey) != "" {
				if active, err := profiles.GetActiveProfile(r.Context()); err == nil {
					status, _ := licenseSvc.Status(r.Context(), active.ID)
					if status.State != "valid" || status.Tier != "pro" {
						var count int
						if err := conn.QueryRowContext(r.Context(), `SELECT COUNT(1) FROM canonical_items`).Scan(&count); err == nil && count >= 150 {
							http.Error(w, `{"error":"free_tier_item_limit_reached"}`, http.StatusPaymentRequired)
							return
						}
					}
				}
			}
			var req collection.Item
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			active, _ := profiles.GetActiveProfile(r.Context())
			if strings.TrimSpace(active.ID) != "" {
				settings, _ := profiles.GetSettings(r.Context(), active.ID)
				if strings.TrimSpace(req.GradingStatus) == "" {
					if v := strings.TrimSpace(settings["grading.defaults.grading_status"]); v != "" {
						req.GradingStatus = strings.ToLower(v)
					}
				}
				if strings.TrimSpace(req.Priority) == "" {
					if v := strings.TrimSpace(settings["grading.defaults.priority"]); v != "" {
						req.Priority = strings.ToLower(v)
					}
				}
			}
			created, err := collectionRepo.CreateItemForProfile(r.Context(), active.ID, req)
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
		query := search.Query{
			Text:      r.URL.Query().Get("q"),
			Brand:     r.URL.Query().Get("brand"),
			Category:  r.URL.Query().Get("category"),
			Condition: r.URL.Query().Get("condition"),
			Status:    r.URL.Query().Get("status"),
			Tags:      r.URL.Query().Get("tags"),
			Scale:     r.URL.Query().Get("scale"),
			SortBy:    r.URL.Query().Get("sort"),
			Limit:     limit,
		}
		items, err := searchRepo.SearchItems(r.Context(), query)
		if active, activeErr := profiles.GetActiveProfile(r.Context()); activeErr == nil && strings.TrimSpace(active.ID) != "" {
			items, err = searchRepo.SearchItemsByProfile(r.Context(), active.ID, query)
		}
		if err != nil {
			http.Error(w, `{"error":"failed_to_search_items"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
	})
	mux.HandleFunc("/api/scanner/query-sets", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		active, _ := profiles.GetActiveProfile(r.Context())
		profileID := strings.TrimSpace(active.ID)
		switch r.Method {
		case http.MethodGet:
			items, err := scannerSvc.ListQuerySetsByProfile(r.Context(), profileID)
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
			created, err := scannerSvc.CreateQuerySetForProfile(r.Context(), profileID, req)
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
	mux.HandleFunc("/api/scanner/query-sets/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		id := strings.TrimPrefix(r.URL.Path, "/api/scanner/query-sets/")
		id = strings.TrimSpace(id)
		if id == "" || strings.Contains(id, "/") {
			http.Error(w, `{"error":"invalid_query_set_id"}`, http.StatusBadRequest)
			return
		}
		active, _ := profiles.GetActiveProfile(r.Context())
		profileID := strings.TrimSpace(active.ID)
		switch r.Method {
		case http.MethodPut:
			var req scanner.QuerySet
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			updated, err := scannerSvc.UpdateQuerySetForProfile(r.Context(), profileID, id, req)
			if err != nil {
				http.Error(w, `{"error":"invalid_query_set"}`, http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(updated)
		case http.MethodDelete:
			if err := scannerSvc.DeleteQuerySetForProfile(r.Context(), profileID, id); err != nil {
				http.Error(w, `{"error":"invalid_query_set"}`, http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
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
		out, err := scannerSvc.RunNowForProfile(r.Context(), strings.TrimSpace(active.ID), req.QuerySetID, provider)
		if err != nil {
			logSvc.Log(r.Context(), "error", "scanner_run_failed", map[string]any{"query_set_id": req.QuerySetID, "error": err.Error()})
			http.Error(w, `{"error":"failed_to_run_scanner"}`, http.StatusBadRequest)
			return
		}
		logSvc.Log(r.Context(), "info", "scanner_run_completed", out)
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
		if !hasProFeatureAccess(r.Context(), conn, licenseSvc, cfg, active.ID, "scanner_automation") {
			writeProFeatureForbidden(w, "scanner_automation")
			return
		}
		provider := ebay.NewProvider(ebay.ProviderConfig{
			BaseURL:     settings["ebay_base_url"],
			BearerToken: settings["ebay_bearer_token"],
			Marketplace: settings["ebay_marketplace"],
		})
		querySets, err := scannerSvc.ListQuerySetsByProfile(r.Context(), strings.TrimSpace(active.ID))
		if err != nil {
			http.Error(w, `{"error":"failed_to_list_query_sets"}`, http.StatusBadRequest)
			return
		}
		summary := map[string]any{
			"run_id":               "scheduled-" + strconv.FormatInt(time.Now().UnixNano(), 10),
			"query_sets_executed":  0,
			"candidates_collected": 0,
			"failures":             0,
		}
		for _, qs := range querySets {
			if !qs.Enabled || strings.TrimSpace(qs.ScheduleCron) == "" {
				continue
			}
			out, runErr := scannerSvc.RunNowForProfile(r.Context(), strings.TrimSpace(active.ID), qs.ID, provider)
			if runErr != nil {
				logSvc.Log(r.Context(), "error", "scanner_run_scheduled_query_failed", map[string]any{
					"query_set_id": qs.ID,
					"error":        runErr.Error(),
				})
				summary["failures"] = summary["failures"].(int) + 1
				continue
			}
			summary["query_sets_executed"] = summary["query_sets_executed"].(int) + 1
			summary["candidates_collected"] = summary["candidates_collected"].(int) + out.Saved
		}
		logSvc.Log(r.Context(), "info", "scanner_run_scheduled_completed", summary)
		_ = json.NewEncoder(w).Encode(summary)
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
		active, _ := profiles.GetActiveProfile(r.Context())
		items, err := scannerSvc.ListCandidatesByProfile(r.Context(), strings.TrimSpace(active.ID), querySetID)
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
	mux.HandleFunc("/api/scanner/failures/retry", func(w http.ResponseWriter, r *http.Request) {
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
		req.QuerySetID = strings.TrimSpace(req.QuerySetID)
		if req.QuerySetID == "" {
			http.Error(w, `{"error":"missing_query_set_id"}`, http.StatusBadRequest)
			return
		}
		profileID := ""
		if active, err := profiles.GetActiveProfile(r.Context()); err == nil {
			profileID = strings.TrimSpace(active.ID)
		}
		_, err := scannerSvc.GetQuerySetForProfile(r.Context(), profileID, req.QuerySetID)
		if err != nil {
			http.Error(w, `{"error":"invalid_query_set_id"}`, http.StatusBadRequest)
			return
		}
		logSvc.Log(r.Context(), "info", "scanner_retry_requested", map[string]any{"query_set_id": req.QuerySetID})
		_ = json.NewEncoder(w).Encode(map[string]any{
			"retry_started": true,
			"query_set_id":  req.QuerySetID,
		})
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
	mux.HandleFunc("/api/providers/registry", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		amazonMode := amazonIntegrationMode(r.Context(), conn)
		settings := map[string]string{}
		if active, err := profiles.GetActiveProfile(r.Context()); err == nil {
			if profileSettings, settingsErr := profiles.GetSettings(r.Context(), strings.TrimSpace(active.ID)); settingsErr == nil {
				settings = profileSettings
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"providers": providerRegistryPayload(r.Context(), scannerSvc, amazonMode, settings),
		})
	})
	mux.HandleFunc("/api/providers/amazon/run", func(w http.ResponseWriter, r *http.Request) {
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
		req.QuerySetID = strings.TrimSpace(req.QuerySetID)
		if req.QuerySetID == "" {
			http.Error(w, `{"error":"missing_query_set_id"}`, http.StatusBadRequest)
			return
		}
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil {
			http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusBadRequest)
			return
		}
		qs, err := scannerSvc.GetQuerySetForProfile(r.Context(), strings.TrimSpace(active.ID), req.QuerySetID)
		if err != nil {
			http.Error(w, `{"error":"invalid_query_set_id"}`, http.StatusBadRequest)
			return
		}
		if amazonIntegrationMode(r.Context(), conn) != "program_api" {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error_code":   "PROVIDER_DISABLED",
				"provider":     "amazon",
				"message":      "Amazon provider is disabled for this profile",
				"next_action":  "enable_provider_or_choose_supported_source",
				"query_set_id": qs.ID,
			})
			return
		}
		candidates := buildAmazonCandidateContract(qs)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"query_set_id": qs.ID,
			"candidates":   candidates,
		})
	})
	mux.HandleFunc("/api/providers/au-webshops/parse-stock", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Domain string `json:"domain"`
			HTML   string `json:"html"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		domain := strings.TrimSpace(strings.ToLower(req.Domain))
		if domain == "" {
			http.Error(w, `{"error":"missing_domain"}`, http.StatusBadRequest)
			return
		}
		raw, state, count := parseStockSignal(req.HTML)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stock_signal": map[string]any{
				"raw":              raw,
				"normalized_state": state,
				"stock_count":      count,
				"source_domain":    domain,
				"observed_at":      time.Now().UTC().Format(time.RFC3339),
			},
		})
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
		active, _ := profiles.GetActiveProfile(r.Context())
		profileID := strings.TrimSpace(active.ID)
		switch r.Method {
		case http.MethodGet:
			items, err := wishlistSvc.ListByProfile(r.Context(), profileID)
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
			created, err := wishlistSvc.CreateForProfile(r.Context(), profileID, req)
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
			if err := wishlistSvc.UpdateForProfile(r.Context(), profileID, req); err != nil {
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
			if err := wishlistSvc.DeleteForProfile(r.Context(), profileID, id); err != nil {
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
		active, _ := profiles.GetActiveProfile(r.Context())
		profileID := strings.TrimSpace(active.ID)
		itemID := strings.TrimSpace(r.URL.Query().Get("item_id"))
		hits, err := wishlistSvc.HitsByProfile(r.Context(), profileID, itemID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_list_wishlist_hits"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"hits": hits})
	})
	itemOwnedByProfile := func(ctx context.Context, profileID, itemID string) bool {
		profileID = strings.TrimSpace(profileID)
		itemID = strings.TrimSpace(itemID)
		if profileID == "" || itemID == "" {
			return false
		}
		var count int
		if err := conn.QueryRowContext(ctx, `SELECT COUNT(1) FROM canonical_items WHERE id = ? AND profile_id = ?`, itemID, profileID).Scan(&count); err != nil {
			return false
		}
		return count > 0
	}
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
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil {
			http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusBadRequest)
			return
		}
		if !hasProFeatureAccess(r.Context(), conn, licenseSvc, cfg, active.ID, "price_tracking") {
			writeProFeatureForbidden(w, "price_tracking")
			return
		}
		if !itemOwnedByProfile(r.Context(), active.ID, req.ItemID) {
			http.Error(w, `{"error":"invalid_item_for_profile"}`, http.StatusBadRequest)
			return
		}
		if err := pricingSvc.TrackItemForProfile(r.Context(), active.ID, req.ItemID); err != nil {
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
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil {
			http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusBadRequest)
			return
		}
		if !hasProFeatureAccess(r.Context(), conn, licenseSvc, cfg, active.ID, "price_tracking") {
			writeProFeatureForbidden(w, "price_tracking")
			return
		}
		if err := pricingSvc.RunDailySnapshotForProfile(r.Context(), active.ID); err != nil {
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
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil {
			http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusBadRequest)
			return
		}
		if !hasProFeatureAccess(r.Context(), conn, licenseSvc, cfg, active.ID, "price_tracking") {
			writeProFeatureForbidden(w, "price_tracking")
			return
		}
		itemID := strings.TrimSpace(r.URL.Query().Get("item_id"))
		if !itemOwnedByProfile(r.Context(), active.ID, itemID) {
			http.Error(w, `{"error":"invalid_item_for_profile"}`, http.StatusBadRequest)
			return
		}
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
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil {
			http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusBadRequest)
			return
		}
		if !hasProFeatureAccess(r.Context(), conn, licenseSvc, cfg, active.ID, "price_tracking") {
			writeProFeatureForbidden(w, "price_tracking")
			return
		}
		itemID := strings.TrimSpace(r.URL.Query().Get("item_id"))
		if !itemOwnedByProfile(r.Context(), active.ID, itemID) {
			http.Error(w, `{"error":"invalid_item_for_profile"}`, http.StatusBadRequest)
			return
		}
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
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil {
			http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusBadRequest)
			return
		}
		if !hasProFeatureAccess(r.Context(), conn, licenseSvc, cfg, active.ID, "price_tracking") {
			writeProFeatureForbidden(w, "price_tracking")
			return
		}
		itemID := strings.TrimSpace(r.URL.Query().Get("item_id"))
		if !itemOwnedByProfile(r.Context(), active.ID, itemID) {
			http.Error(w, `{"error":"invalid_item_for_profile"}`, http.StatusBadRequest)
			return
		}
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
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil {
			http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusBadRequest)
			return
		}
		if !hasProFeatureAccess(r.Context(), conn, licenseSvc, cfg, active.ID, "price_tracking") {
			writeProFeatureForbidden(w, "price_tracking")
			return
		}
		itemID := strings.TrimSpace(r.URL.Query().Get("item_id"))
		if !itemOwnedByProfile(r.Context(), active.ID, itemID) {
			http.Error(w, `{"error":"invalid_item_for_profile"}`, http.StatusBadRequest)
			return
		}
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
	mux.HandleFunc("/api/pricing/trend", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil {
			http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusBadRequest)
			return
		}
		if !hasProFeatureAccess(r.Context(), conn, licenseSvc, cfg, active.ID, "price_tracking") {
			writeProFeatureForbidden(w, "price_tracking")
			return
		}
		itemID := strings.TrimSpace(r.URL.Query().Get("item_id"))
		if !itemOwnedByProfile(r.Context(), active.ID, itemID) {
			http.Error(w, `{"error":"invalid_item_for_profile"}`, http.StatusBadRequest)
			return
		}
		points, err := pricingSvc.Trend(r.Context(), itemID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_get_price_trend"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"points": points})
	})
	mux.HandleFunc("/api/pricing/history/export", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil {
			http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(cfg.UpdatePublicKey) != "" {
			if !hasProFeatureAccess(r.Context(), conn, licenseSvc, cfg, active.ID, "price_tracking") {
				writeProFeatureForbidden(w, "price_tracking")
				return
			}
		}
		itemID := strings.TrimSpace(r.URL.Query().Get("item_id"))
		if !itemOwnedByProfile(r.Context(), active.ID, itemID) {
			http.Error(w, `{"error":"invalid_item_for_profile"}`, http.StatusBadRequest)
			return
		}
		csvText, err := pricingSvc.ExportCSV(r.Context(), itemID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_export_price_history"}`, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(csvText))
	})
	mux.HandleFunc("/api/dashboard", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		sum, err := dashboardSvc.Summary(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed_to_load_dashboard"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(sum)
	})
	mux.HandleFunc("/api/logs/activity", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		items, err := logSvc.List(r.Context(), limit)
		if err != nil {
			http.Error(w, `{"error":"failed_to_list_activity_logs"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"logs": items})
	})
	mux.HandleFunc("/api/logs/export", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		text, err := logSvc.Export(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed_to_export_logs"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(text))
	})
	mux.HandleFunc("/api/logs/debug", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ProfileID string `json:"profile_id"`
			Enabled   bool   `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		val := "false"
		if req.Enabled {
			val = "true"
		}
		if err := profiles.PutSettings(r.Context(), req.ProfileID, map[string]string{"debug_mode": val}); err != nil {
			http.Error(w, `{"error":"failed_to_toggle_debug_mode"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	mux.HandleFunc("/api/license/import", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ProfileID string         `json:"profile_id"`
			License   licensing.File `json:"license"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		if err := licenseSvc.Import(r.Context(), req.ProfileID, req.License); err != nil {
			http.Error(w, `{"error":"failed_to_import_license"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	mux.HandleFunc("/api/license/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		profileID := strings.TrimSpace(r.URL.Query().Get("profile_id"))
		st, err := licenseSvc.Status(r.Context(), profileID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_get_license_status"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(st)
	})
	mux.HandleFunc("/api/ai/toggle", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ProfileID string `json:"profile_id"`
			Enabled   bool   `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		val := "false"
		if req.Enabled {
			val = "true"
		}
		if err := profiles.PutSettings(r.Context(), req.ProfileID, map[string]string{"ai_enabled": val}); err != nil {
			http.Error(w, `{"error":"failed_to_toggle_ai"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	mux.HandleFunc("/api/ai/test", func(w http.ResponseWriter, r *http.Request) {
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
		if !hasProFeatureAccess(r.Context(), conn, licenseSvc, cfg, req.ProfileID, "ai_assist") {
			writeProFeatureForbidden(w, "ai_assist")
			return
		}
		settings, _ := profiles.GetSettings(r.Context(), req.ProfileID)
		if settings["ai_enabled"] == "false" {
			http.Error(w, `{"error":"ai_disabled"}`, http.StatusBadRequest)
			return
		}
		key, err := profiles.GetSecret(r.Context(), req.ProfileID, "openai_api_key")
		if err != nil {
			http.Error(w, `{"error":"missing_openai_api_key"}`, http.StatusBadRequest)
			return
		}
		localAISvc := aiSvc
		if baseURL := strings.TrimSpace(settings["openai_base_url"]); baseURL != "" {
			localAISvc = ai.NewService(ai.Config{BaseURL: baseURL})
		}
		if err := localAISvc.TestConnectivity(r.Context(), key); err != nil {
			_, _ = conn.ExecContext(r.Context(), `INSERT INTO ai_failures(id, profile_id, message, created_at) VALUES (hex(randomblob(16)), ?, ?, CURRENT_TIMESTAMP)`, req.ProfileID, err.Error())
			logSvc.Log(r.Context(), "error", "ai_connectivity_failed", map[string]any{"profile_id": req.ProfileID, "error": err.Error()})
			http.Error(w, `{"error":"failed_ai_connectivity_test"}`, http.StatusBadRequest)
			return
		}
		logSvc.Log(r.Context(), "info", "ai_connectivity_ok", map[string]any{"profile_id": req.ProfileID})
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	mux.HandleFunc("/api/ai/suggest/title", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ProfileID string `json:"profile_id"`
			Title     string `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		if !hasProFeatureAccess(r.Context(), conn, licenseSvc, cfg, req.ProfileID, "ai_assist") {
			writeProFeatureForbidden(w, "ai_assist")
			return
		}
		settings, _ := profiles.GetSettings(r.Context(), req.ProfileID)
		if settings["ai_enabled"] == "false" {
			http.Error(w, `{"error":"ai_disabled"}`, http.StatusBadRequest)
			return
		}
		key, err := profiles.GetSecret(r.Context(), req.ProfileID, "openai_api_key")
		if err != nil {
			http.Error(w, `{"error":"missing_openai_api_key"}`, http.StatusBadRequest)
			return
		}
		localAISvc := aiSvc
		if baseURL := strings.TrimSpace(settings["openai_base_url"]); baseURL != "" {
			localAISvc = ai.NewService(ai.Config{BaseURL: baseURL})
		}
		suggestion, err := localAISvc.SuggestFromTitle(r.Context(), key, req.Title)
		if err != nil {
			_, _ = conn.ExecContext(r.Context(), `INSERT INTO ai_failures(id, profile_id, message, created_at) VALUES (hex(randomblob(16)), ?, ?, CURRENT_TIMESTAMP)`, req.ProfileID, err.Error())
			logSvc.Log(r.Context(), "error", "ai_suggest_title_failed", map[string]any{"profile_id": req.ProfileID, "error": err.Error()})
			http.Error(w, `{"error":"failed_ai_suggest_title"}`, http.StatusBadRequest)
			return
		}
		logSvc.Log(r.Context(), "info", "ai_suggest_title_ok", map[string]any{"profile_id": req.ProfileID, "confidence": suggestion.Confidence})
		_ = json.NewEncoder(w).Encode(suggestion)
	})
	mux.HandleFunc("/api/ai/suggest/photo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ProfileID string `json:"profile_id"`
			ImageURL  string `json:"image_url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		if !hasProFeatureAccess(r.Context(), conn, licenseSvc, cfg, req.ProfileID, "ai_assist") {
			writeProFeatureForbidden(w, "ai_assist")
			return
		}
		settings, _ := profiles.GetSettings(r.Context(), req.ProfileID)
		if settings["ai_enabled"] == "false" {
			http.Error(w, `{"error":"ai_disabled"}`, http.StatusBadRequest)
			return
		}
		key, err := profiles.GetSecret(r.Context(), req.ProfileID, "openai_api_key")
		if err != nil {
			http.Error(w, `{"error":"missing_openai_api_key"}`, http.StatusBadRequest)
			return
		}
		localAISvc := aiSvc
		if baseURL := strings.TrimSpace(settings["openai_base_url"]); baseURL != "" {
			localAISvc = ai.NewService(ai.Config{BaseURL: baseURL})
		}
		suggestion, err := localAISvc.SuggestFromPhoto(r.Context(), key, req.ImageURL)
		if err != nil {
			_, _ = conn.ExecContext(r.Context(), `INSERT INTO ai_failures(id, profile_id, message, created_at) VALUES (hex(randomblob(16)), ?, ?, CURRENT_TIMESTAMP)`, req.ProfileID, err.Error())
			logSvc.Log(r.Context(), "error", "ai_suggest_photo_failed", map[string]any{"profile_id": req.ProfileID, "error": err.Error()})
			http.Error(w, `{"error":"failed_ai_suggest_photo"}`, http.StatusBadRequest)
			return
		}
		logSvc.Log(r.Context(), "info", "ai_suggest_photo_ok", map[string]any{"profile_id": req.ProfileID, "confidence": suggestion.Confidence})
		_ = json.NewEncoder(w).Encode(suggestion)
	})
	mux.HandleFunc("/api/chat/threads", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			profileID := strings.TrimSpace(r.URL.Query().Get("profile_id"))
			if profileID == "" {
				http.Error(w, `{"error":"profile_id_required"}`, http.StatusBadRequest)
				return
			}
			threads, err := chatSvc.ListThreads(r.Context(), profileID)
			if err != nil {
				http.Error(w, `{"error":"failed_to_list_chat_threads"}`, http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"threads": threads})
		case http.MethodPost:
			var req struct {
				ProfileID string `json:"profile_id"`
				Title     string `json:"title"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			thread, err := chatSvc.CreateThread(r.Context(), req.ProfileID, req.Title)
			if err != nil {
				http.Error(w, `{"error":"failed_to_create_chat_thread"}`, http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(thread)
		default:
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/chat/messages", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			profileID := strings.TrimSpace(r.URL.Query().Get("profile_id"))
			threadID := strings.TrimSpace(r.URL.Query().Get("thread_id"))
			if profileID == "" || threadID == "" {
				http.Error(w, `{"error":"profile_id_and_thread_id_required"}`, http.StatusBadRequest)
				return
			}
			messages, err := chatSvc.ListMessages(r.Context(), profileID, threadID)
			if err != nil {
				http.Error(w, `{"error":"failed_to_list_chat_messages"}`, http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"messages": messages})
		case http.MethodPost:
			var req struct {
				ProfileID string `json:"profile_id"`
				ThreadID  string `json:"thread_id"`
				Role      string `json:"role"`
				Content   string `json:"content"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			message, err := chatSvc.CreateMessage(r.Context(), req.ProfileID, req.ThreadID, req.Role, req.Content)
			if err != nil {
				http.Error(w, `{"error":"failed_to_create_chat_message"}`, http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(message)
		default:
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/chat/attachments", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type"))), "multipart/form-data") {
			http.Error(w, `{"error":"multipart_form_data_required"}`, http.StatusBadRequest)
			return
		}
		if err := r.ParseMultipartForm(16 << 20); err != nil {
			http.Error(w, `{"error":"invalid_multipart"}`, http.StatusBadRequest)
			return
		}
		profileID := strings.TrimSpace(r.FormValue("profile_id"))
		threadID := strings.TrimSpace(r.FormValue("thread_id"))
		if profileID == "" || threadID == "" {
			http.Error(w, `{"error":"profile_id_and_thread_id_required"}`, http.StatusBadRequest)
			return
		}
		file, hdr, err := r.FormFile("file")
		if err != nil {
			http.Error(w, `{"error":"file_required"}`, http.StatusBadRequest)
			return
		}
		defer file.Close()
		attachment, err := chatSvc.SaveAttachment(r.Context(), profileID, threadID, hdr.Filename, hdr.Header.Get("Content-Type"), file)
		if err != nil {
			http.Error(w, `{"error":"failed_to_save_attachment"}`, http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(attachment)
	})
	mux.HandleFunc("/api/chat/actions/preview", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req chat.PreviewActionInput
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		preview, err := chatSvc.PreviewAction(r.Context(), req)
		if err != nil {
			http.Error(w, `{"error":"failed_to_preview_chat_action"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(preview)
	})
	mux.HandleFunc("/api/chat/actions/apply", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req chat.ApplyActionInput
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		result, err := chatSvc.ApplyAction(r.Context(), req)
		if err != nil {
			http.Error(w, `{"error":"failed_to_apply_chat_action"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(result)
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
		itemID := strings.TrimSpace(parts[0])
		if itemID == "" {
			http.Error(w, `{"error":"invalid_item_id"}`, http.StatusBadRequest)
			return
		}
		if itemID == "bulk-edit" {
			if r.Method != http.MethodPost {
				http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			var req struct {
				ItemIDs []string        `json:"item_ids"`
				Changes collection.Item `json:"changes"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			result, err := collectionRepo.BulkEditItems(r.Context(), req.ItemIDs, req.Changes)
			if err != nil {
				http.Error(w, `{"error":"invalid_bulk_edit"}`, http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(result)
			return
		}
		if len(parts) == 1 {
			if r.Method == http.MethodDelete {
				item, err := collectionRepo.GetItemByID(r.Context(), itemID)
				if err != nil {
					http.Error(w, `{"error":"item_not_found"}`, http.StatusNotFound)
					return
				}
				switch strings.ToLower(strings.TrimSpace(item.Status)) {
				case "", "active":
					updated, setErr := collectionRepo.SetItemLifecycleStatus(r.Context(), itemID, "deleted")
					if setErr != nil {
						http.Error(w, `{"error":"failed_to_update_item_status"}`, http.StatusBadRequest)
						return
					}
					_ = json.NewEncoder(w).Encode(map[string]any{
						"ok":             true,
						"lifecycle_step": "soft_deleted",
						"item":           updated,
					})
					return
				case "deleted":
					updated, setErr := collectionRepo.SetItemLifecycleStatus(r.Context(), itemID, "recycle")
					if setErr != nil {
						http.Error(w, `{"error":"failed_to_update_item_status"}`, http.StatusBadRequest)
						return
					}
					_ = json.NewEncoder(w).Encode(map[string]any{
						"ok":             true,
						"lifecycle_step": "moved_to_recycle",
						"item":           updated,
					})
					return
				case "recycle":
					dependencies, depErr := collectionRepo.ListItemDependencyCounts(r.Context(), itemID)
					if depErr != nil {
						http.Error(w, `{"error":"failed_to_check_dependencies"}`, http.StatusInternalServerError)
						return
					}
					if len(dependencies) > 0 {
						w.WriteHeader(http.StatusConflict)
						_ = json.NewEncoder(w).Encode(map[string]any{
							"error":        "item_has_dependencies",
							"message":      "item cannot be permanently deleted while dependencies exist",
							"dependencies": dependencies,
						})
						return
					}
					if delErr := collectionRepo.DeleteItemPermanent(r.Context(), itemID); delErr != nil {
						http.Error(w, `{"error":"failed_to_delete_item"}`, http.StatusInternalServerError)
						return
					}
					w.WriteHeader(http.StatusNoContent)
					return
				default:
					http.Error(w, `{"error":"invalid_item_status"}`, http.StatusBadRequest)
					return
				}
			}
			if r.Method != http.MethodPut {
				http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			var req collection.Item
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			updated, err := collectionRepo.UpdateItem(r.Context(), itemID, req)
			if err != nil {
				http.Error(w, `{"error":"invalid_item"}`, http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(updated)
			return
		}
		switch parts[1] {
		case "restore":
			if r.Method != http.MethodPost {
				http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			item, err := collectionRepo.GetItemByID(r.Context(), itemID)
			if err != nil {
				http.Error(w, `{"error":"item_not_found"}`, http.StatusNotFound)
				return
			}
			status := strings.ToLower(strings.TrimSpace(item.Status))
			if status != "deleted" && status != "recycle" {
				http.Error(w, `{"error":"item_not_restorable"}`, http.StatusConflict)
				return
			}
			restored, err := collectionRepo.SetItemLifecycleStatus(r.Context(), itemID, "active")
			if err != nil {
				http.Error(w, `{"error":"failed_to_restore_item"}`, http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":             true,
				"lifecycle_step": "restored",
				"item":           restored,
			})
			return
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
			if len(parts) == 3 {
				if r.Method != http.MethodPut {
					http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
					return
				}
				instanceID := strings.TrimSpace(parts[2])
				if instanceID == "" {
					http.Error(w, `{"error":"invalid_instance_id"}`, http.StatusBadRequest)
					return
				}
				var req collection.Instance
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
					return
				}
				updated, err := collectionRepo.UpdateInstance(r.Context(), instanceID, req)
				if err != nil {
					http.Error(w, `{"error":"invalid_instance"}`, http.StatusBadRequest)
					return
				}
				_ = json.NewEncoder(w).Encode(updated)
				return
			}
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
				if parts[2] == "reorder" {
					if r.Method != http.MethodPost {
						http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
						return
					}
					var req struct {
						PhotoIDs []string `json:"photo_ids"`
					}
					if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
						http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
						return
					}
					if err := mediaService.Reorder(r.Context(), itemID, req.PhotoIDs); err != nil {
						http.Error(w, `{"error":"failed_to_reorder_photos"}`, http.StatusBadRequest)
						return
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
					return
				}
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
	mux.HandleFunc("/api/onboarding/sample-data", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		result, err := seedOnboardingSampleData(r.Context(), profiles, collectionRepo, wishlistSvc, conn)
		if err != nil {
			http.Error(w, `{"error":"failed_to_seed_onboarding_sample_data"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(result)
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
		sessionToken, err := authService.FinishLogin(r.Context(), req.SessionID, req.Credential)
		if err != nil {
			http.Error(w, `{"error":"failed_to_finish_login"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "session_token": sessionToken})
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
	mux.HandleFunc("/api/auth/cloud/session/bootstrap", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Provider string `json:"provider"`
			Token    string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(strings.ToLower(req.Provider)) != "clerk" {
			http.Error(w, `{"error":"unsupported_provider"}`, http.StatusBadRequest)
			return
		}
		claims, err := parseCloudAuthClaims(req.Token)
		if err != nil {
			http.Error(w, `{"error":"invalid_token"}`, http.StatusUnauthorized)
			return
		}
		userID := strings.TrimSpace(claimAsString(claims, "sub"))
		if userID == "" {
			http.Error(w, `{"error":"invalid_token_subject"}`, http.StatusUnauthorized)
			return
		}
		email := strings.TrimSpace(claimAsString(claims, "email"))
		plan := strings.TrimSpace(strings.ToLower(claimAsString(claims, "plan")))
		entitlementSource := "billing"
		if plan == "" {
			plan = strings.TrimSpace(strings.ToLower(claimAsString(claims, "cabinet_plan")))
		}
		if override, ok := cloudEntitlements.Get(userID); ok {
			plan = override
			entitlementSource = "override"
		}
		if plan == "" {
			plan = "free"
			entitlementSource = "trial"
		}
		features := entitlementFeaturesFromPlan(plan)
		_ = persistCloudPlan(r.Context(), conn, userID, plan)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"provider":           "clerk",
			"user_id":            userID,
			"email":              email,
			"plan":               plan,
			"features":           features,
			"entitlement_source": entitlementSource,
		})
	})
	mux.HandleFunc("/api/auth/cloud/clerk/webhook", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, `{"error":"invalid_body"}`, http.StatusBadRequest)
			return
		}
		secret := strings.TrimSpace(os.Getenv("CABINET_CLERK_WEBHOOK_SECRET"))
		if secret == "" {
			secret = "dev-secret"
		}
		signature := strings.TrimSpace(r.Header.Get("X-Cabinet-Webhook-Signature"))
		if !verifyWebhookSignature(secret, body, signature) {
			http.Error(w, `{"error":"invalid_signature"}`, http.StatusUnauthorized)
			return
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		userID, plan, err := clerkWebhookPlanTransition(payload)
		if err != nil {
			http.Error(w, `{"error":"invalid_webhook_payload"}`, http.StatusBadRequest)
			return
		}
		cloudEntitlements.Set(userID, plan)
		_ = persistCloudPlan(r.Context(), conn, userID, plan)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user_id":  userID,
			"plan":     plan,
			"features": entitlementFeaturesFromPlan(plan),
		})
	})
	mux.HandleFunc("/api/auth/cloud/lease/issue", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Provider        string `json:"provider"`
			Token           string `json:"token"`
			DurationSeconds int    `json:"duration_seconds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(strings.ToLower(req.Provider)) != "clerk" {
			http.Error(w, `{"error":"unsupported_provider"}`, http.StatusBadRequest)
			return
		}
		claims, err := parseUnverifiedJWTPayload(req.Token)
		if err != nil {
			http.Error(w, `{"error":"invalid_token"}`, http.StatusBadRequest)
			return
		}
		userID := strings.TrimSpace(claimAsString(claims, "sub"))
		if userID == "" {
			http.Error(w, `{"error":"invalid_token_subject"}`, http.StatusBadRequest)
			return
		}
		plan := strings.TrimSpace(strings.ToLower(claimAsString(claims, "plan")))
		if plan == "" {
			plan = "free"
		}
		lease, err := cloudLeases.Issue(userID, plan, durationFromSeconds(req.DurationSeconds))
		if err != nil {
			http.Error(w, `{"error":"lease_issue_failed"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"lease_token": lease.Token,
			"user_id":     lease.UserID,
			"plan":        lease.Plan,
			"features":    entitlementFeaturesFromPlan(lease.Plan),
			"expires_at":  lease.ExpiresAt.UTC().Format(time.RFC3339),
		})
	})
	mux.HandleFunc("/api/auth/cloud/lease/renew", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			LeaseToken      string `json:"lease_token"`
			DurationSeconds int    `json:"duration_seconds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		lease, err := cloudLeases.Renew(strings.TrimSpace(req.LeaseToken), durationFromSeconds(req.DurationSeconds))
		if err != nil {
			http.Error(w, `{"error":"lease_expired"}`, http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"lease_token": lease.Token,
			"user_id":     lease.UserID,
			"plan":        lease.Plan,
			"features":    entitlementFeaturesFromPlan(lease.Plan),
			"expires_at":  lease.ExpiresAt.UTC().Format(time.RFC3339),
		})
	})
	mux.HandleFunc("/api/auth/cloud/lease/validate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		token := strings.TrimSpace(r.URL.Query().Get("lease_token"))
		lease, err := cloudLeases.Validate(token)
		if err != nil {
			_ = json.NewEncoder(w).Encode(map[string]any{"valid": false})
			return
		}
		remaining := int(time.Until(lease.ExpiresAt).Seconds())
		if remaining < 0 {
			remaining = 0
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"valid":             true,
			"user_id":           lease.UserID,
			"plan":              lease.Plan,
			"expires_at":        lease.ExpiresAt.UTC().Format(time.RFC3339),
			"seconds_remaining": remaining,
		})
	})
	mux.HandleFunc("/api/auth/cloud/offline/protected", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		token := strings.TrimSpace(r.Header.Get("X-Cabinet-Lease"))
		if token == "" {
			token = strings.TrimSpace(r.URL.Query().Get("lease_token"))
		}
		if _, err := cloudLeases.Validate(token); err != nil {
			http.Error(w, `{"error":"lease_expired"}`, http.StatusLocked)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})

	protectedMux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requiresUnlockedSession(r) {
			mux.ServeHTTP(w, r)
			return
		}
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil {
			mux.ServeHTTP(w, r)
			return
		}
		registrationRequired, err := authService.RequiresRegistration(r.Context(), active.ID)
		if err != nil {
			http.Error(w, `{"error":"session_locked"}`, http.StatusLocked)
			return
		}
		if registrationRequired {
			mux.ServeHTTP(w, r)
			return
		}
		token := sessionTokenFromRequest(r)
		if token != "" {
			if err := authService.ValidateUnlockedSession(token); err != nil {
				http.Error(w, `{"error":"session_locked"}`, http.StatusLocked)
				return
			}
			mux.ServeHTTP(w, r)
			return
		}
		if authService.HasUnlockedSession(active.ID) {
			mux.ServeHTTP(w, r)
			return
		}
		http.Error(w, `{"error":"session_locked"}`, http.StatusLocked)
	})

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           protectedMux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	a := &App{
		cfg:         cfg,
		db:          conn,
		srv:         srv,
		backupSvc:   backupSvc,
		authService: authService,
		openapiSpec: openapiSpec,
	}

	return a, nil
}

const apiDocsHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Cabinet API Docs</title>
  <style>
    body { margin: 0; font-family: Aptos, "Segoe UI", sans-serif; background: #f3f6fb; }
    header { padding: 12px 16px; border-bottom: 1px solid #dbe5f4; background: #fff; }
    header a { color: #1d4ed8; text-decoration: none; }
  </style>
</head>
<body>
  <header>
    <strong>Cabinet API Docs</strong> - <a href="/api/openapi.yaml">OpenAPI YAML</a>
  </header>
  <redoc spec-url="/api/openapi.yaml"></redoc>
  <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
</body>
</html>`

func loadOpenAPISpec(cfg config.Config) []byte {
	candidates := []string{
		filepath.Join("docs", "api", "openapi.yaml"),
		filepath.Join("..", "..", "docs", "api", "openapi.yaml"),
		filepath.Join(filepath.Dir(cfg.DBPath), "docs", "api", "openapi.yaml"),
	}
	if _, sourceFile, _, ok := runtime.Caller(0); ok {
		candidates = append(candidates, filepath.Join(filepath.Dir(sourceFile), "..", "..", "docs", "api", "openapi.yaml"))
	}
	if exePath, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exePath), "docs", "api", "openapi.yaml"))
	}
	for _, path := range candidates {
		b, err := os.ReadFile(path)
		if err == nil && len(b) > 0 {
			return b
		}
	}
	return nil
}

func requiresUnlockedSession(r *http.Request) bool {
	if !strings.HasPrefix(r.URL.Path, "/api/") {
		return false
	}
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodDelete:
	default:
		return false
	}
	if strings.HasPrefix(r.URL.Path, "/api/auth/") {
		return false
	}
	if strings.HasPrefix(r.URL.Path, "/api/onboarding/") {
		return false
	}
	if r.URL.Path == "/api/profiles" || r.URL.Path == "/api/profiles/active" {
		return false
	}
	return true
}

func sessionTokenFromRequest(r *http.Request) string {
	token := strings.TrimSpace(r.Header.Get("X-Cabinet-Session"))
	if token != "" {
		return token
	}
	authz := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
		return strings.TrimSpace(authz[7:])
	}
	return strings.TrimSpace(r.URL.Query().Get("session_token"))
}

type cloudLease struct {
	Token     string
	UserID    string
	Plan      string
	ExpiresAt time.Time
}

type cloudLeaseStore struct {
	mu     sync.Mutex
	leases map[string]cloudLease
}

type cloudEntitlementStore struct {
	mu    sync.Mutex
	plans map[string]string
}

func newCloudEntitlementStore() *cloudEntitlementStore {
	return &cloudEntitlementStore{plans: map[string]string{}}
}

func (s *cloudEntitlementStore) Set(userID, plan string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.plans[strings.TrimSpace(userID)] = normalizePlan(plan)
}

func (s *cloudEntitlementStore) Get(userID string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	plan, ok := s.plans[strings.TrimSpace(userID)]
	return plan, ok
}

func newCloudLeaseStore() *cloudLeaseStore {
	return &cloudLeaseStore{leases: map[string]cloudLease{}}
}

func (s *cloudLeaseStore) Issue(userID, plan string, ttl time.Duration) (cloudLease, error) {
	token, err := randomToken(24)
	if err != nil {
		return cloudLease{}, err
	}
	lease := cloudLease{
		Token:     token,
		UserID:    strings.TrimSpace(userID),
		Plan:      strings.TrimSpace(strings.ToLower(plan)),
		ExpiresAt: time.Now().Add(ttl),
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.leases[token] = lease
	return lease, nil
}

func (s *cloudLeaseStore) Validate(token string) (cloudLease, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	lease, ok := s.leases[strings.TrimSpace(token)]
	if !ok {
		return cloudLease{}, fmt.Errorf("lease_not_found")
	}
	if time.Now().After(lease.ExpiresAt) {
		return cloudLease{}, fmt.Errorf("lease_expired")
	}
	return lease, nil
}

func (s *cloudLeaseStore) Renew(token string, ttl time.Duration) (cloudLease, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	lease, ok := s.leases[strings.TrimSpace(token)]
	if !ok {
		return cloudLease{}, fmt.Errorf("lease_not_found")
	}
	if time.Now().After(lease.ExpiresAt) {
		return cloudLease{}, fmt.Errorf("lease_expired")
	}
	lease.ExpiresAt = time.Now().Add(ttl)
	s.leases[lease.Token] = lease
	return lease, nil
}

func randomToken(size int) (string, error) {
	if size <= 0 {
		size = 24
	}
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func durationFromSeconds(seconds int) time.Duration {
	if seconds == 0 {
		return 24 * time.Hour
	}
	return time.Duration(seconds) * time.Second
}

func verifyWebhookSignature(secret string, body []byte, signature string) bool {
	if strings.TrimSpace(signature) == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(strings.TrimSpace(signature)))
}

type diagnosticsConfig struct {
	RemoteOptIn bool
	Provider    string
	RemoteURL   string
}

func diagnosticsConfigFromDB(ctx context.Context, conn *sql.DB) diagnosticsConfig {
	cfg := diagnosticsConfig{
		RemoteOptIn: false,
		Provider:    "sentry",
	}
	var v string
	if err := conn.QueryRowContext(ctx, `SELECT value FROM app_state WHERE key = 'diagnostics.remote_opt_in'`).Scan(&v); err == nil {
		cfg.RemoteOptIn = strings.TrimSpace(v) == "1"
	}
	if err := conn.QueryRowContext(ctx, `SELECT value FROM app_state WHERE key = 'diagnostics.provider'`).Scan(&v); err == nil && strings.TrimSpace(v) != "" {
		cfg.Provider = strings.TrimSpace(strings.ToLower(v))
	}
	if err := conn.QueryRowContext(ctx, `SELECT value FROM app_state WHERE key = 'diagnostics.remote_url'`).Scan(&v); err == nil {
		cfg.RemoteURL = strings.TrimSpace(v)
	}
	return cfg
}

func persistDiagnosticsConfig(ctx context.Context, conn *sql.DB, cfg diagnosticsConfig) error {
	optVal := "0"
	if cfg.RemoteOptIn {
		optVal = "1"
	}
	updates := map[string]string{
		"diagnostics.remote_opt_in": optVal,
		"diagnostics.provider":      cfg.Provider,
		"diagnostics.remote_url":    cfg.RemoteURL,
	}
	for k, v := range updates {
		if _, err := conn.ExecContext(ctx, `
			INSERT INTO app_state(key, value, updated_at)
			VALUES(?, ?, CURRENT_TIMESTAMP)
			ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=CURRENT_TIMESTAMP
		`, k, v); err != nil {
			return err
		}
	}
	return nil
}

func buildSentryCompatibleEnvelope(eventType, category, message, sessionID string, details map[string]any) map[string]any {
	level := "info"
	if strings.EqualFold(strings.TrimSpace(eventType), "error") {
		level = "error"
	}
	eventID, _ := randomToken(12)
	return map[string]any{
		"event_id":  eventID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"level":     level,
		"message":   strings.TrimSpace(message),
		"tags": map[string]any{
			"category": strings.TrimSpace(strings.ToLower(category)),
			"type":     strings.TrimSpace(strings.ToLower(eventType)),
		},
		"contexts": map[string]any{
			"session": map[string]any{
				"id": strings.TrimSpace(sessionID),
			},
			"diagnostics": details,
		},
	}
}

func sendDiagnosticsEnvelope(ctx context.Context, cfg diagnosticsConfig, envelope map[string]any) error {
	if !cfg.RemoteOptIn {
		return nil
	}
	if cfg.Provider != "sentry" || strings.TrimSpace(cfg.RemoteURL) == "" {
		return fmt.Errorf("diagnostics provider not configured")
	}
	body, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.RemoteURL, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("diagnostics remote send status %d", resp.StatusCode)
	}
	return nil
}

func classifyErrorTaxonomy(errorCode, message string) (category, nextAction, guidance string) {
	input := strings.ToLower(strings.TrimSpace(errorCode) + " " + strings.TrimSpace(message))
	switch {
	case strings.Contains(input, "provider"), strings.Contains(input, "ebay"), strings.Contains(input, "amazon"), strings.Contains(input, "scanner"), strings.Contains(input, "timeout"):
		return "provider", "check_provider_health_and_credentials", "Check provider health, credentials, and retry the operation."
	case strings.Contains(input, "network"), strings.Contains(input, "connect"), strings.Contains(input, "dns"), strings.Contains(input, "offline"):
		return "connectivity", "check_network_and_retry", "Verify network connectivity and retry."
	case strings.Contains(input, "unauthorized"), strings.Contains(input, "forbidden"), strings.Contains(input, "session_locked"), strings.Contains(input, "invalid_signature"), strings.Contains(input, "invalid_token"):
		return "authorization", "re_authenticate_or_unlock_session", "Re-authenticate, unlock the session, or refresh access credentials."
	case strings.Contains(input, "invalid"), strings.Contains(input, "validation"), strings.Contains(input, "bad_request"), strings.Contains(input, "missing"):
		return "validation", "correct_request_payload", "Correct the request payload and try again."
	default:
		return "internal", "review_logs_and_report", "Review diagnostics logs and report with context."
	}
}

func parseCloudAuthClaims(token string) (map[string]any, error) {
	if os.Getenv("CABINET_CLOUD_AUTH_ENFORCE_SIGNED_TOKENS") != "1" {
		return parseUnverifiedJWTPayload(token)
	}
	secret := strings.TrimSpace(os.Getenv("CABINET_CLOUD_AUTH_HS256_SECRET"))
	if secret == "" {
		return nil, fmt.Errorf("strict cloud auth verification enabled but CABINET_CLOUD_AUTH_HS256_SECRET missing")
	}
	return parseVerifiedHS256JWTPayload(token, secret)
}

func parseVerifiedHS256JWTPayload(token, secret string) (map[string]any, error) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid jwt token")
	}
	decode := func(s string) ([]byte, error) {
		b, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(s))
		if err != nil {
			return nil, fmt.Errorf("decode jwt segment: %w", err)
		}
		return b, nil
	}
	headerRaw, err := decode(parts[0])
	if err != nil {
		return nil, err
	}
	var header map[string]any
	if err := json.Unmarshal(headerRaw, &header); err != nil {
		return nil, fmt.Errorf("decode jwt header: %w", err)
	}
	if strings.TrimSpace(strings.ToUpper(claimAsString(header, "alg"))) != "HS256" {
		return nil, fmt.Errorf("unsupported jwt algorithm")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))
	expected := mac.Sum(nil)
	gotSig, err := decode(parts[2])
	if err != nil {
		return nil, err
	}
	if !hmac.Equal(expected, gotSig) {
		return nil, fmt.Errorf("invalid jwt signature")
	}
	payloadRaw, err := decode(parts[1])
	if err != nil {
		return nil, err
	}
	var claims map[string]any
	if err := json.Unmarshal(payloadRaw, &claims); err != nil {
		return nil, fmt.Errorf("decode jwt payload: %w", err)
	}
	return claims, nil
}

func persistCloudPlan(ctx context.Context, conn *sql.DB, userID, plan string) error {
	updates := map[string]string{
		"cloud.user_id": strings.TrimSpace(userID),
		"cloud.plan":    normalizePlan(plan),
	}
	for k, v := range updates {
		if _, err := conn.ExecContext(ctx, `
			INSERT INTO app_state(key, value, updated_at)
			VALUES(?, ?, CURRENT_TIMESTAMP)
			ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=CURRENT_TIMESTAMP
		`, k, v); err != nil {
			return err
		}
	}
	return nil
}

func currentCloudPlanFromState(ctx context.Context, conn *sql.DB) (string, bool) {
	var plan string
	if err := conn.QueryRowContext(ctx, `SELECT value FROM app_state WHERE key = 'cloud.plan'`).Scan(&plan); err != nil {
		return "", false
	}
	plan = normalizePlan(plan)
	if strings.TrimSpace(plan) == "" {
		return "", false
	}
	return plan, true
}

func hasProFeatureAccess(ctx context.Context, conn *sql.DB, licenseSvc *licensing.Service, cfg config.Config, profileID, feature string) bool {
	if cloudPlan, ok := currentCloudPlanFromState(ctx, conn); ok {
		return cloudPlan == "pro"
	}
	if strings.TrimSpace(cfg.UpdatePublicKey) != "" {
		allowed, _ := licenseSvc.Allow(ctx, profileID, feature)
		return allowed
	}
	return true
}

func writeProFeatureForbidden(w http.ResponseWriter, feature string) {
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error":      "forbidden",
		"error_code": "PRO_FEATURE_REQUIRED",
		"feature":    strings.TrimSpace(strings.ToLower(feature)),
		"message":    fmt.Sprintf("pro feature required: %s", strings.TrimSpace(strings.ToLower(feature))),
	})
}

func writeUpdateInstallError(w http.ResponseWriter, status int, code, message string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error":      "update_validation_failed",
		"error_code": strings.TrimSpace(code),
		"message":    strings.TrimSpace(message),
	})
}

func filterItemsByLifecycleStatus(items []collection.Item, status string) []collection.Item {
	target := strings.TrimSpace(strings.ToLower(status))
	if target == "" {
		target = "active"
	}
	filtered := make([]collection.Item, 0, len(items))
	for _, item := range items {
		current := strings.TrimSpace(strings.ToLower(item.Status))
		if current == "" {
			current = "active"
		}
		if current == target {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func defaultConditionGrades() []string {
	return []string{"M", "NM", "EX", "VG", "G", "F", "P"}
}

func defaultPackagingGrades() []string {
	return []string{"sealed_mint", "sealed_good", "opened_complete", "loose"}
}

func parseStringArraySetting(raw string, fallback []string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return append([]string(nil), fallback...)
	}
	var values []string
	if err := json.Unmarshal([]byte(trimmed), &values); err != nil {
		return append([]string(nil), fallback...)
	}
	return normalizeStringList(values, fallback)
}

func normalizeStringList(input []string, fallback []string) []string {
	out := make([]string, 0, len(input))
	seen := map[string]struct{}{}
	for _, item := range input {
		clean := strings.TrimSpace(strings.ToLower(item))
		if clean == "" {
			continue
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		out = append(out, clean)
	}
	if len(out) == 0 {
		return append([]string(nil), fallback...)
	}
	return out
}

func clerkWebhookPlanTransition(payload map[string]any) (string, string, error) {
	data, _ := payload["data"].(map[string]any)
	if data == nil {
		return "", "", fmt.Errorf("missing data")
	}
	userID := strings.TrimSpace(claimAsString(data, "user_id"))
	if userID == "" {
		customer, _ := data["customer"].(map[string]any)
		userID = strings.TrimSpace(claimAsString(customer, "id"))
	}
	if userID == "" {
		return "", "", fmt.Errorf("missing user_id")
	}
	plan := normalizePlan(claimAsString(data, "plan"))
	if plan == "free" {
		meta, _ := data["public_metadata"].(map[string]any)
		plan = normalizePlan(claimAsString(meta, "cabinet_plan"))
	}
	if plan == "free" {
		plan = normalizePlan(claimAsString(data, "cabinet_plan"))
	}
	return userID, plan, nil
}

func normalizePlan(plan string) string {
	normalized := strings.TrimSpace(strings.ToLower(plan))
	if normalized == "" {
		return "free"
	}
	return normalized
}

func parseUnverifiedJWTPayload(token string) (map[string]any, error) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid token format")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("decode claims: %w", err)
	}
	return claims, nil
}

func claimAsString(claims map[string]any, key string) string {
	raw, ok := claims[key]
	if !ok || raw == nil {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func entitlementFeaturesFromPlan(plan string) []string {
	switch strings.TrimSpace(strings.ToLower(plan)) {
	case "pro", "paid", "plus":
		return []string{"collection_core", "ai_assist", "price_tracking", "scanner_automation"}
	default:
		return []string{"collection_core"}
	}
}

func amazonIntegrationMode(ctx context.Context, conn *sql.DB) string {
	var mode string
	if err := conn.QueryRowContext(ctx, `SELECT value FROM app_state WHERE key = 'provider.amazon.mode'`).Scan(&mode); err != nil {
		return "disabled"
	}
	mode = strings.TrimSpace(strings.ToLower(mode))
	if mode == "program_api" {
		return mode
	}
	return "disabled"
}

func providerRegistryPayload(ctx context.Context, scannerSvc *scanner.Service, amazonMode string, settings map[string]string) []map[string]any {
	base := []map[string]any{
		{
			"provider_id":      "ebay",
			"display_name":     "eBay",
			"base_domain":      "ebay.com",
			"integration_mode": "official_api",
			"api_available":    true,
			"auth_requirement": "api_key",
			"auth_mode":        "api_key",
			"capabilities": map[string]bool{
				"search":            true,
				"stock_observation": false,
				"pricing":           true,
				"health":            true,
			},
			"state":              "ready",
			"setup_instructions": "Add eBay API token and marketplace, validate health, then run scanner query sets.",
		},
		{
			"provider_id":          "amazon",
			"display_name":         "Amazon",
			"base_domain":          "amazon.com",
			"integration_mode":     amazonMode,
			"api_available":        amazonMode == "program_api",
			"auth_requirement":     "oauth",
			"auth_mode":            "hybrid",
			"eligibility_required": true,
			"policy_scope_note":    "Program API access eligibility controls availability.",
			"capabilities": map[string]bool{
				"search":            amazonMode == "program_api",
				"stock_observation": false,
				"pricing":           amazonMode == "program_api",
				"health":            true,
			},
			"state":              map[bool]string{true: "ready", false: "disabled"}[amazonMode == "program_api"],
			"setup_instructions": "Configure Amazon credentials and eligibility mode before running provider scans.",
		},
	}

	auDomains := []string{
		"bonzaslotcars.com.au",
		"frontlinehobbies.com.au",
		"hobbytechtoys.com.au",
		"andrewshobbies.com.au",
		"voglers.com.au",
		"acercmodels.com",
		"mrtoys.com.au",
		"hobbyco.com.au",
		"metrohobbies.com.au",
	}
	for _, d := range auDomains {
		base = append(base, map[string]any{
			"provider_id":      "au-webshop-" + strings.ReplaceAll(d, ".", "-"),
			"display_name":     d,
			"base_domain":      d,
			"integration_mode": "web_ingestion",
			"api_available":    false,
			"auth_requirement": "none",
			"auth_mode":        "none",
			"capabilities": map[string]bool{
				"search":            true,
				"stock_observation": true,
				"pricing":           true,
				"health":            true,
			},
			"state":              "ready",
			"setup_instructions": "Webshop ingestion uses crawl parsing and does not require API credentials.",
		})
	}

	for _, provider := range base {
		providerID := strings.TrimSpace(fmt.Sprintf("%v", provider["provider_id"]))
		if providerID == "" {
			continue
		}
		keys := providerSettingsKeys(providerID)
		hasToken := strings.TrimSpace(settings[keys.TokenKey]) != ""
		provider["has_token"] = hasToken

		healthStatus := "unknown"
		healthMessage := ""
		lastChecked := any(nil)
		lastRunStatus := "never"
		lastRunFinished := any(nil)
		if scannerSvc != nil {
			if health, err := scannerSvc.ProviderHealth(ctx, providerID); err == nil {
				if v := strings.TrimSpace(health["status"]); v != "" {
					healthStatus = v
					switch v {
					case "ok", "ready":
						lastRunStatus = "success"
					case "unknown":
						lastRunStatus = "never"
					default:
						lastRunStatus = "failed"
					}
				}
				if v := strings.TrimSpace(health["message"]); v != "" {
					healthMessage = v
				}
				if v := strings.TrimSpace(health["updated_at"]); v != "" {
					lastChecked = v
					lastRunFinished = v
				}
			}
		}
		provider["health"] = map[string]any{
			"status":          healthStatus,
			"message":         healthMessage,
			"last_checked_at": lastChecked,
		}
		provider["last_run"] = map[string]any{
			"status":      lastRunStatus,
			"finished_at": lastRunFinished,
		}
	}

	return base
}

type providerSettingKeySet struct {
	BaseURLKey     string
	TokenKey       string
	MarketplaceKey string
	EnabledKey     string
}

func providerSettingsKeys(providerID string) providerSettingKeySet {
	if strings.TrimSpace(strings.ToLower(providerID)) == "ebay" {
		return providerSettingKeySet{
			BaseURLKey:     "ebay_base_url",
			TokenKey:       "ebay_bearer_token",
			MarketplaceKey: "ebay_marketplace",
			EnabledKey:     "integration.ebay.enabled",
		}
	}
	slug := strings.TrimSpace(strings.ToLower(providerID))
	slug = strings.ReplaceAll(slug, "-", "_")
	slug = strings.ReplaceAll(slug, ".", "_")
	return providerSettingKeySet{
		BaseURLKey:     "integration." + slug + ".base_url",
		TokenKey:       "integration." + slug + ".token",
		MarketplaceKey: "integration." + slug + ".marketplace",
		EnabledKey:     "integration." + slug + ".enabled",
	}
}

func buildAmazonCandidateContract(qs scanner.QuerySet) []map[string]any {
	keyword := "collectible"
	if len(qs.Keywords) > 0 && strings.TrimSpace(qs.Keywords[0]) != "" {
		keyword = strings.TrimSpace(qs.Keywords[0])
	}
	return []map[string]any{
		{
			"listing_id": "amazon-" + strings.ToLower(strings.ReplaceAll(keyword, " ", "-")) + "-001",
			"title":      "Amazon " + keyword + " listing",
			"price": map[string]any{
				"amount":   39.99,
				"currency": "AUD",
			},
			"url":    "https://amazon.com/dp/example",
			"seller": "amazon-marketplace",
			"source": map[string]any{
				"provider_id": "amazon",
			},
		},
	}
}

func parseStockSignal(html string) (raw, normalized string, count int) {
	stripTags := regexp.MustCompile(`<[^>]*>`)
	compactSpaces := regexp.MustCompile(`\s+`)
	numberPattern := regexp.MustCompile(`\d+`)
	raw = strings.TrimSpace(compactSpaces.ReplaceAllString(stripTags.ReplaceAllString(html, " "), " "))
	if raw == "" {
		return "", "unknown", -1
	}
	lower := strings.ToLower(raw)
	count = -1
	if n := numberPattern.FindString(lower); n != "" {
		if parsed, err := strconv.Atoi(n); err == nil {
			count = parsed
		}
	}
	switch {
	case strings.Contains(lower, "out of stock"), strings.Contains(lower, "sold out"), count == 0:
		return raw, "out_of_stock", 0
	case strings.Contains(lower, "in stock"), strings.Contains(lower, "left in stock"), count > 0 && count <= 3:
		return raw, "low_stock", count
	case count > 3:
		return raw, "in_stock", count
	default:
		return raw, "unknown", count
	}
}

func runtimeBuildMetadata() (string, string) {
	version := "dev"
	buildDate := "unknown"

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return version, buildDate
	}

	if strings.TrimSpace(info.Main.Version) != "" && info.Main.Version != "(devel)" {
		version = info.Main.Version
	}

	var vcsRevision string
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			vcsRevision = strings.TrimSpace(setting.Value)
		case "vcs.time":
			if strings.TrimSpace(setting.Value) != "" {
				buildDate = strings.TrimSpace(setting.Value)
			}
		}
	}

	if vcsRevision != "" {
		short := vcsRevision
		if len(short) > 12 {
			short = short[:12]
		}
		version = "rev-" + short
	}

	return version, buildDate
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
	_, _ = a.db.Exec(`INSERT INTO app_state(key, value, updated_at) VALUES('clean_shutdown','1',CURRENT_TIMESTAMP) ON CONFLICT(key) DO UPDATE SET value='1', updated_at=CURRENT_TIMESTAMP`)
	return a.db.Close()
}

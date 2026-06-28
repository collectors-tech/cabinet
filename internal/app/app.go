package app

import (
	"bytes"
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
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"sort"
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
	"github.com/collectors-tech/cabinet/internal/commerce"
	"github.com/collectors-tech/cabinet/internal/companion"
	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/costing"
	"github.com/collectors-tech/cabinet/internal/dashboard"
	"github.com/collectors-tech/cabinet/internal/datamgmt"
	"github.com/collectors-tech/cabinet/internal/db"
	"github.com/collectors-tech/cabinet/internal/discovery"
	"github.com/collectors-tech/cabinet/internal/ebay"
	"github.com/collectors-tech/cabinet/internal/ebaypurchasecapture"
	"github.com/collectors-tech/cabinet/internal/forwarding"
	"github.com/collectors-tech/cabinet/internal/licensing"
	"github.com/collectors-tech/cabinet/internal/logging"
	"github.com/collectors-tech/cabinet/internal/matching"
	"github.com/collectors-tech/cabinet/internal/media"
	"github.com/collectors-tech/cabinet/internal/pricing"
	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/collectors-tech/cabinet/internal/scanner"
	"github.com/collectors-tech/cabinet/internal/search"
	"github.com/collectors-tech/cabinet/internal/telegramcapture"
	"github.com/collectors-tech/cabinet/internal/ui"
	"github.com/collectors-tech/cabinet/internal/update"
	"github.com/collectors-tech/cabinet/internal/wishlist"
)

func startupMigrationTimeout() time.Duration {
	const defaultTimeout = 3 * time.Minute
	value := strings.TrimSpace(os.Getenv("CABINET_STARTUP_TIMEOUT_SECONDS"))
	if value == "" {
		return defaultTimeout
	}
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return defaultTimeout
	}
	return time.Duration(seconds) * time.Second
}

func startupSampleDataSeedEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("CABINET_SEED_SAMPLE_DATA"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

type App struct {
	cfg           config.Config
	db            *sql.DB
	srv           *http.Server
	backupSvc     *backup.Service
	authService   *auth.Service
	openapiSpec   []byte
	runtimeLogs   *runtimeLogManager
	runtimeStopCh chan string
	startupNotice func(string)
	startupIsTTY  func() bool
}

func New(cfg config.Config) (*App, error) {
	if strings.TrimSpace(cfg.ValidationError) != "" {
		return nil, fmt.Errorf("invalid runtime config: %s", strings.TrimSpace(cfg.ValidationError))
	}
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), startupMigrationTimeout())
	defer cancel()

	conn, err := db.OpenAndMigrate(ctx, cfg.DBPath)
	if err != nil {
		return nil, err
	}
	if err := syncRuntimeSetupCurrentURL(cfg); err != nil {
		conn.Close()
		return nil, fmt.Errorf("sync runtime setup metadata: %w", err)
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
	commerceSvc := commerce.NewService(conn)
	forwarderInbox := forwarding.NewService(conn)
	pricingSvc := pricing.NewService(conn)
	dashboardSvc := dashboard.NewService(conn)
	chatSvc := chat.NewService(conn, filepath.Join(cfg.DataDir, "chat-attachments"))
	companionSvc := companion.DefaultService()
	aiSvc := ai.NewService(ai.Config{})
	licenseSvc := licensing.NewService(conn, profiles, cfg.UpdatePublicKey)
	logSvc := logging.NewService(conn)
	authService, err := auth.NewService(cfg, conn, profiles)
	if err != nil {
		conn.Close()
		return nil, err
	}
	if startupSampleDataSeedEnabled() {
		result, seedErr := seedOnboardingSampleData(ctx, profiles, collectionRepo, wishlistSvc, mediaService, conn)
		if seedErr != nil {
			log.Printf("sample data startup seed skipped: %v", seedErr)
		} else {
			log.Printf(
				"sample data startup seed complete: created_items=%d created_photos=%d created_wishlist_entries=%d total_wishlist_entries=%d",
				result.CreatedItems,
				result.CreatedPhotos,
				result.CreatedWishlistEntries,
				result.TotalWishlistEntries,
			)
		}
	}
	cloudLeases := newCloudLeaseStore()
	cloudEntitlements := newCloudEntitlementStore()
	runtimeStopCh := make(chan string, 1)

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
		_ = syncRuntimeLifecycleUncleanRecovery(cfg)
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
		host := strings.TrimSpace(cfg.Host)
		if host == "" {
			host = "127.0.0.1"
		}
		port := cfg.Port
		if port <= 0 {
			port = 17880
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"update_channel":               cfg.UpdateChannel,
			"update_public_key_configured": cfg.UpdatePublicKey != "",
			"app_version":                  appVersion,
			"build_date":                   buildDate,
			"bind_mode":                    strings.TrimSpace(strings.ToLower(cfg.BindMode)),
			"runtime_host":                 host,
			"runtime_port":                 port,
			"pid":                          os.Getpid(),
			"data_dir":                     cfg.DataDir,
		})
	})
	mux.HandleFunc("/api/runtime/shutdown", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		if !requestIsLoopback(r) {
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}
		var payload struct {
			Reason string `json:"reason"`
		}
		if r.Body != nil {
			defer r.Body.Close()
			_ = json.NewDecoder(r.Body).Decode(&payload)
		}
		reason := strings.TrimSpace(payload.Reason)
		if reason == "" {
			reason = "api_shutdown"
		}
		select {
		case runtimeStopCh <- reason:
		default:
		}
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":     true,
			"pid":    os.Getpid(),
			"reason": reason,
		})
	})
	mux.HandleFunc("/api/auth/provider-options", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		_ = json.NewEncoder(w).Encode(resolveAuthProviderOptions())
	})
	registerFutureHookRoutes(mux)
	mux.HandleFunc("/api/runtime/setup-status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		configPath := runtimeSetupConfigPath(cfg)
		defaultStorageDataDir := runtimeDefaultStorageDataDir(cfg)
		host := strings.TrimSpace(cfg.Host)
		if host == "" {
			host = "127.0.0.1"
		}
		port := cfg.Port
		if port <= 0 {
			port = 17880
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"setup_required":              runtimeSetupRequired(cfg),
			"config_path":                 configPath,
			"default_storage_data_dir":    defaultStorageDataDir,
			"default_storage_media_dir":   filepath.Join(defaultStorageDataDir, "media"),
			"default_storage_portable":    false,
			"default_storage_mode":        "exe_local",
			"default_storage_free_status": "unknown",
			"default_runtime_host":        host,
			"default_runtime_port":        port,
			"default_runtime_port_mode":   "auto",
			"default_runtime_url":         fmt.Sprintf("http://%s:%d", host, port),
		})
	})
	mux.HandleFunc("/api/runtime/setup-complete", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req runtimeSetupRequest
		if r.Body != nil {
			defer r.Body.Close()
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
		}
		if validationErr := validateRuntimeSetupRequest(req); validationErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":          "invalid_setup_payload",
				"error_code":     validationErr.Code,
				"message":        validationErr.Message,
				"field":          validationErr.Field,
				"setup_required": true,
			})
			return
		}
		payload, buildErr := buildRuntimeSetupConfig(cfg, req)
		if buildErr != nil {
			http.Error(w, `{"error":"failed_to_build_setup_config"}`, http.StatusInternalServerError)
			return
		}
		if err := writeRuntimeSetupConfig(cfg, payload); err != nil {
			http.Error(w, `{"error":"failed_to_write_setup_config"}`, http.StatusInternalServerError)
			return
		}
		resolvedURL := strings.TrimSpace(payload.Runtime.ResolvedURL)
		if resolvedURL == "" {
			resolvedURL = runtimeResolvedURLFromConfig(cfg)
		}
		_ = syncRuntimeLifecycleStartup(cfg, resolvedURL, cfg.Addr, os.Getpid())
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":             true,
			"setup_required": false,
			"instance_name":  payload.Instance.Name,
			"profile_key":    payload.Instance.Profile,
			"config_path":    runtimeSetupConfigPath(cfg),
			"auth_mode":      payload.Auth.Mode,
			"local_login_username": func() string {
				if payload.Auth.Mode != "local" {
					return ""
				}
				return "admin@cabinet.local"
			}(),
			"local_login_password": func() string {
				if payload.Auth.Mode != "local" {
					return ""
				}
				return "password123"
			}(),
			"data_dir":     payload.Storage.DataDir,
			"media_dir":    payload.Storage.MediaDir,
			"runtime_url":  payload.Runtime.ResolvedURL,
			"runtime_port": portFromResolvedURL(payload.Runtime.ResolvedURL),
		})
	})
	mux.HandleFunc("/api/runtime/setup-import", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req runtimeSetupImportRequest
		if r.Body != nil {
			defer r.Body.Close()
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
		}
		validationErr := validateRuntimeSetupImportRequest(req)
		if validationErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":          "invalid_import_source",
				"error_code":     validationErr.Code,
				"message":        validationErr.Message,
				"field":          validationErr.Field,
				"setup_required": true,
			})
			return
		}
		importedPayload, err := importRuntimeSetupConfigFromPath(cfg, req.SourcePath)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":          "failed_to_import_setup_config",
				"error_code":     "SETUP_IMPORT_FAILED",
				"message":        err.Error(),
				"field":          "source_path",
				"setup_required": true,
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":             true,
			"setup_required": false,
			"instance_name":  importedPayload.Instance.Name,
			"profile_key":    importedPayload.Instance.Profile,
			"config_path":    runtimeSetupConfigPath(cfg),
			"runtime_url":    importedPayload.Runtime.ResolvedURL,
			"runtime_port":   portFromResolvedURL(importedPayload.Runtime.ResolvedURL),
		})
	})
	mux.HandleFunc("/api/runtime/setup-storage-validate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			DataDir string `json:"data_dir"`
		}
		if r.Body != nil {
			defer r.Body.Close()
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
		}
		targetDir := strings.TrimSpace(req.DataDir)
		if targetDir == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":                false,
				"writable":          false,
				"free_space_ok":     false,
				"free_space_bytes":  nil,
				"free_space_status": "unknown",
				"error_code":        "SETUP_STORAGE_PATH_REQUIRED",
				"message":           "Storage data directory is required.",
			})
			return
		}
		writable, message := checkRuntimeSetupStorageWritable(targetDir)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":                writable,
			"writable":          writable,
			"free_space_ok":     true,
			"free_space_bytes":  nil,
			"free_space_status": "unknown",
			"message":           message,
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
			if r.Method == http.MethodDelete {
				key := strings.TrimSpace(r.URL.Query().Get("key"))
				if err := profiles.DeleteSecret(r.Context(), profileID, key); err != nil {
					http.Error(w, `{"error":"failed_to_delete_secret"}`, http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusNoContent)
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
			itemTypeConditionScales := parseItemTypeConditionScalesSetting(settings["grading.enums.item_type_condition_scales"])
			_ = json.NewEncoder(w).Encode(map[string]any{
				"condition_grades":           condition,
				"packaging_grades":           packaging,
				"item_type_condition_scales": itemTypeConditionScales,
			})
		case http.MethodPut:
			var req struct {
				ConditionGrades         []string                 `json:"condition_grades"`
				PackagingGrades         []string                 `json:"packaging_grades"`
				ItemTypeConditionScales []itemTypeConditionScale `json:"item_type_condition_scales"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			settings, _ := profiles.GetSettings(r.Context(), active.ID)
			condition := normalizeStringList(req.ConditionGrades, defaultConditionGrades())
			packaging := normalizeStringList(req.PackagingGrades, defaultPackagingGrades())
			itemTypeConditionScales := normalizeItemTypeConditionScales(req.ItemTypeConditionScales, parseItemTypeConditionScalesSetting(settings["grading.enums.item_type_condition_scales"]))
			conditionRaw, _ := json.Marshal(condition)
			packagingRaw, _ := json.Marshal(packaging)
			itemTypeConditionScalesRaw, _ := json.Marshal(itemTypeConditionScales)
			if putErr := profiles.PutSettings(r.Context(), active.ID, map[string]string{
				"grading.enums.condition":                  string(conditionRaw),
				"grading.enums.packaging":                  string(packagingRaw),
				"grading.enums.item_type_condition_scales": string(itemTypeConditionScalesRaw),
			}); putErr != nil {
				http.Error(w, `{"error":"failed_to_save_grading_enums"}`, http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"condition_grades":           condition,
				"packaging_grades":           packaging,
				"item_type_condition_scales": itemTypeConditionScales,
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
	mux.HandleFunc("/api/integrations/pokemon/set-progress", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		setID := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("set_id")))
		if setID == "" {
			http.Error(w, `{"error":"missing_set_id"}`, http.StatusBadRequest)
			return
		}
		totalCount := 0
		if raw := strings.TrimSpace(r.URL.Query().Get("total_count")); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed < 0 {
				http.Error(w, `{"error":"invalid_total_count"}`, http.StatusBadRequest)
				return
			}
			totalCount = parsed
		}
		items, err := collectionRepo.ListItems(r.Context())
		if active, activeErr := profiles.GetActiveProfile(r.Context()); activeErr == nil && strings.TrimSpace(active.ID) != "" {
			items, err = collectionRepo.ListItemsByProfile(r.Context(), active.ID)
		}
		if err != nil {
			http.Error(w, `{"error":"failed_to_list_items"}`, http.StatusInternalServerError)
			return
		}

		variantBreakdown := map[string]int{}
		languageBreakdown := map[string]int{}
		gradedBreakdown := map[string]int{
			"graded":   0,
			"ungraded": 0,
		}
		ownedCount := 0
		for _, item := range items {
			if strings.TrimSpace(strings.ToLower(item.Status)) == "deleted" || strings.TrimSpace(strings.ToLower(item.Status)) == "recycle" {
				continue
			}
			itemSet, itemVariant, itemLanguage := parsePokemonSetProgressTags(item.Tags)
			if itemSet != setID {
				continue
			}
			ownedCount++
			variantBreakdown[itemVariant]++
			languageBreakdown[itemLanguage]++
			if strings.EqualFold(strings.TrimSpace(item.GradingStatus), "graded") {
				gradedBreakdown["graded"]++
			} else {
				gradedBreakdown["ungraded"]++
			}
		}
		if totalCount <= 0 {
			totalCount = ownedCount
		}
		completionPercent := 0.0
		if totalCount > 0 {
			completionPercent = roundTo2((float64(ownedCount) / float64(totalCount)) * 100)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"set_id":             setID,
			"owned_count":        ownedCount,
			"total_count":        totalCount,
			"completion_percent": completionPercent,
			"breakdown": map[string]any{
				"variant":  variantBreakdown,
				"language": languageBreakdown,
				"graded":   gradedBreakdown,
			},
		})
	})
	mux.HandleFunc("/api/integrations/pokemon/progress-snapshot", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "method_not_allowed"})
			return
		}
		setID := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("set_id")))
		if setID == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "missing_set_id"})
			return
		}
		totalCount := 0
		if raw := strings.TrimSpace(r.URL.Query().Get("total_count")); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed < 0 {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid_total_count"})
				return
			}
			totalCount = parsed
		}
		visibility := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("visibility")))
		if visibility == "" {
			visibility = "private"
		}
		if visibility != "private" && visibility != "shared_link" && visibility != "team" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid_visibility"})
			return
		}
		items, err := collectionRepo.ListItems(r.Context())
		if active, activeErr := profiles.GetActiveProfile(r.Context()); activeErr == nil && strings.TrimSpace(active.ID) != "" {
			items, err = collectionRepo.ListItemsByProfile(r.Context(), active.ID)
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "failed_to_list_items"})
			return
		}
		ownedCount := 0
		for _, item := range items {
			if strings.TrimSpace(strings.ToLower(item.Status)) == "deleted" || strings.TrimSpace(strings.ToLower(item.Status)) == "recycle" {
				continue
			}
			itemSet, _, _ := parsePokemonSetProgressTags(item.Tags)
			if itemSet == setID {
				ownedCount++
			}
		}
		if totalCount <= 0 {
			totalCount = ownedCount
		}
		completionPercent := 0.0
		if totalCount > 0 {
			completionPercent = roundTo2((float64(ownedCount) / float64(totalCount)) * 100)
		}
		shareLink := fmt.Sprintf("cabinet://pokemon/progress/%s?visibility=%s", url.PathEscape(setID), visibility)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"set_id":             setID,
			"owned_count":        ownedCount,
			"total_count":        totalCount,
			"completion_percent": completionPercent,
			"generated_at":       time.Now().UTC().Format(time.RFC3339),
			"share_payload": map[string]any{
				"headline":     fmt.Sprintf("Set %s progress", strings.ToUpper(setID)),
				"summary":      fmt.Sprintf("%d/%d cards collected", ownedCount, totalCount),
				"visibility":   visibility,
				"share_link":   shareLink,
				"profile_hint": "active_profile",
			},
		})
	})
	mux.HandleFunc("/api/integrations/pokemon/milestone-evaluate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "method_not_allowed"})
			return
		}
		var req struct {
			SetID      string `json:"set_id"`
			TotalCount int    `json:"total_count"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid_json"})
			return
		}
		setID := strings.TrimSpace(strings.ToLower(req.SetID))
		if setID == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "missing_set_id"})
			return
		}
		totalCount := req.TotalCount
		if totalCount < 0 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid_total_count"})
			return
		}
		items, err := collectionRepo.ListItems(r.Context())
		if active, activeErr := profiles.GetActiveProfile(r.Context()); activeErr == nil && strings.TrimSpace(active.ID) != "" {
			items, err = collectionRepo.ListItemsByProfile(r.Context(), active.ID)
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "failed_to_list_items"})
			return
		}
		ownedCount := 0
		for _, item := range items {
			if strings.TrimSpace(strings.ToLower(item.Status)) == "deleted" || strings.TrimSpace(strings.ToLower(item.Status)) == "recycle" {
				continue
			}
			itemSet, _, _ := parsePokemonSetProgressTags(item.Tags)
			if itemSet == setID {
				ownedCount++
			}
		}
		if totalCount <= 0 {
			totalCount = ownedCount
		}
		completionPercent := 0.0
		if totalCount > 0 {
			completionPercent = roundTo2((float64(ownedCount) / float64(totalCount)) * 100)
		}
		thresholds := []int{25, 50, 75, 100}
		events := make([]map[string]any, 0, len(thresholds))
		triggeredAt := time.Now().UTC().Format(time.RFC3339)
		for _, threshold := range thresholds {
			if completionPercent >= float64(threshold) {
				events = append(events, map[string]any{
					"milestone_id":  fmt.Sprintf("milestone-%d", threshold),
					"threshold_pct": threshold,
					"triggered_at":  triggeredAt,
				})
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"set_id":             setID,
			"owned_count":        ownedCount,
			"completion_percent": completionPercent,
			"events":             events,
		})
	})
	mux.HandleFunc("/api/integrations/pokemon/price-alerts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		setID := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("set_id")))
		if setID == "" {
			http.Error(w, `{"error":"missing_set_id"}`, http.StatusBadRequest)
			return
		}
		dropThreshold := 10.0
		if raw := strings.TrimSpace(r.URL.Query().Get("drop_threshold_pct")); raw != "" {
			value, err := strconv.ParseFloat(raw, 64)
			if err != nil || value < 0 {
				http.Error(w, `{"error":"invalid_drop_threshold_pct"}`, http.StatusBadRequest)
				return
			}
			dropThreshold = value
		}

		items, err := collectionRepo.ListItems(r.Context())
		if active, activeErr := profiles.GetActiveProfile(r.Context()); activeErr == nil && strings.TrimSpace(active.ID) != "" {
			items, err = collectionRepo.ListItemsByProfile(r.Context(), active.ID)
		}
		if err != nil {
			http.Error(w, `{"error":"failed_to_list_items"}`, http.StatusInternalServerError)
			return
		}

		type sourceStats struct {
			Source   string  `json:"source"`
			MinPrice float64 `json:"min_price"`
			Median   float64 `json:"median_price"`
			Latest   float64 `json:"latest_price"`
		}
		type alertRecord struct {
			ItemID       string  `json:"item_id"`
			Source       string  `json:"source"`
			ChangePct    float64 `json:"change_pct"`
			ThresholdPct float64 `json:"threshold_pct"`
		}

		sourceLatestPrices := map[string][]float64{}
		alerts := make([]alertRecord, 0)
		consideredItemIDs := make([]string, 0)
		for _, item := range items {
			if strings.TrimSpace(strings.ToLower(item.Status)) == "deleted" || strings.TrimSpace(strings.ToLower(item.Status)) == "recycle" {
				continue
			}
			itemSet, _, _ := parsePokemonSetProgressTags(item.Tags)
			if itemSet != setID {
				continue
			}
			consideredItemIDs = append(consideredItemIDs, item.ID)
			history, historyErr := pricingSvc.History(r.Context(), item.ID)
			if historyErr != nil {
				http.Error(w, `{"error":"failed_to_load_price_history"}`, http.StatusInternalServerError)
				return
			}
			bySource := map[string][]pricing.Snapshot{}
			for _, snap := range history {
				source := strings.TrimSpace(strings.ToLower(snap.Source))
				if source == "" {
					source = "unknown"
				}
				bySource[source] = append(bySource[source], snap)
			}
			for source, snapshots := range bySource {
				if len(snapshots) == 0 {
					continue
				}
				latest := snapshots[len(snapshots)-1]
				sourceLatestPrices[source] = append(sourceLatestPrices[source], latest.LatestPrice)
				if len(snapshots) < 2 {
					continue
				}
				previous := snapshots[len(snapshots)-2]
				if previous.LatestPrice <= 0 {
					continue
				}
				changePct := roundTo2(((latest.LatestPrice - previous.LatestPrice) / previous.LatestPrice) * 100)
				if changePct <= -dropThreshold {
					alerts = append(alerts, alertRecord{
						ItemID:       item.ID,
						Source:       source,
						ChangePct:    changePct,
						ThresholdPct: dropThreshold,
					})
				}
			}
		}

		sources := make([]sourceStats, 0, len(sourceLatestPrices))
		for source, prices := range sourceLatestPrices {
			if len(prices) == 0 {
				continue
			}
			sort.Float64s(prices)
			sources = append(sources, sourceStats{
				Source:   source,
				MinPrice: roundTo2(prices[0]),
				Median:   roundTo2(prices[len(prices)/2]),
				Latest:   roundTo2(prices[len(prices)-1]),
			})
		}
		sort.Slice(sources, func(i, j int) bool {
			return sources[i].Source < sources[j].Source
		})
		sort.Slice(alerts, func(i, j int) bool {
			if alerts[i].ItemID == alerts[j].ItemID {
				return alerts[i].Source < alerts[j].Source
			}
			return alerts[i].ItemID < alerts[j].ItemID
		})

		_ = json.NewEncoder(w).Encode(map[string]any{
			"set_id":  setID,
			"items":   consideredItemIDs,
			"sources": sources,
			"alerts":  alerts,
		})
	})
	mux.HandleFunc("/api/integrations/pokemon/visibility-access", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		visibility := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("visibility")))
		actor := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("actor")))
		if actor == "" {
			actor = "anonymous"
		}
		shareToken := strings.TrimSpace(r.URL.Query().Get("share_token"))
		switch visibility {
		case "private":
			if actor == "anonymous" {
				http.Error(w, `{"error":"visibility_forbidden","visibility":"private","required":"authenticated"}`, http.StatusForbidden)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":         true,
				"visibility": "private",
				"actor":      actor,
			})
		case "shared_link":
			if shareToken == "" {
				http.Error(w, `{"error":"missing_share_token","visibility":"shared_link","required":"share_token"}`, http.StatusForbidden)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":                  true,
				"visibility":          "shared_link",
				"actor":               actor,
				"share_token_present": true,
			})
		case "team":
			if actor != "team_member" {
				http.Error(w, `{"error":"visibility_forbidden","visibility":"team","required":"team_member"}`, http.StatusForbidden)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":         true,
				"visibility": "team",
				"actor":      actor,
			})
		default:
			http.Error(w, `{"error":"invalid_visibility"}`, http.StatusBadRequest)
			return
		}
	})
	type pokemonListTemplate struct {
		ID             string         `json:"id"`
		Label          string         `json:"label"`
		DefaultFields  []string       `json:"default_fields"`
		DefaultFilters map[string]any `json:"default_filters"`
		SortOrder      []string       `json:"sort_order"`
	}
	type pokemonGoalBundle struct {
		ID       string         `json:"id"`
		Label    string         `json:"label"`
		Filters  map[string]any `json:"filters"`
		Actions  []string       `json:"actions"`
		Shortcut string         `json:"shortcut"`
	}
	pokemonTemplates := map[string]pokemonListTemplate{
		"wishlist": {
			ID:            "wishlist",
			Label:         "Wishlist",
			DefaultFields: []string{"thumbnail", "part_number", "title", "priority", "target_price"},
			DefaultFilters: map[string]any{
				"status": []string{"wishlist"},
			},
			SortOrder: []string{"priority:desc", "updated_at:desc"},
		},
		"trade_binder": {
			ID:            "trade_binder",
			Label:         "Trade Binder",
			DefaultFields: []string{"thumbnail", "part_number", "title", "grader", "grade_numeric", "collector_classification"},
			DefaultFilters: map[string]any{
				"status":         []string{"active"},
				"grading_status": []string{"graded"},
			},
			SortOrder: []string{"grade_numeric:desc", "updated_at:desc"},
		},
		"watchlist": {
			ID:            "watchlist",
			Label:         "Watchlist",
			DefaultFields: []string{"thumbnail", "part_number", "title", "status", "priority"},
			DefaultFilters: map[string]any{
				"status": []string{"discovered", "wishlist"},
			},
			SortOrder: []string{"updated_at:desc"},
		},
	}
	pokemonGoalBundles := map[string]pokemonGoalBundle{
		"finish-master-set": {
			ID:       "finish-master-set",
			Label:    "Finish Master Set",
			Filters:  map[string]any{"status": "missing", "priority": "high"},
			Actions:  []string{"run_scanner_now", "open_wishlist_focus", "track_price_drops"},
			Shortcut: "goal.master_set",
		},
		"optimize-trade-binder": {
			ID:       "optimize-trade-binder",
			Label:    "Optimize Trade Binder",
			Filters:  map[string]any{"status": "duplicate", "grading_status": "graded"},
			Actions:  []string{"open_trade_binder", "export_trade_cards"},
			Shortcut: "goal.trade_binder",
		},
		"price-drop-watch": {
			ID:       "price-drop-watch",
			Label:    "Price Drop Watch",
			Filters:  map[string]any{"watch": true, "price_drop": true},
			Actions:  []string{"open_pricing_panel", "run_price_snapshot"},
			Shortcut: "goal.price_drop_watch",
		},
	}
	mux.HandleFunc("/api/integrations/pokemon/list-templates", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		templates := make([]pokemonListTemplate, 0, len(pokemonTemplates))
		for _, template := range pokemonTemplates {
			templates = append(templates, template)
		}
		sort.Slice(templates, func(i, j int) bool {
			return templates[i].ID < templates[j].ID
		})
		_ = json.NewEncoder(w).Encode(map[string]any{
			"templates": templates,
		})
	})
	mux.HandleFunc("/api/integrations/pokemon/list-templates/apply", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			TemplateID string `json:"template_id"`
			ListName   string `json:"list_name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		templateID := strings.TrimSpace(strings.ToLower(req.TemplateID))
		template, ok := pokemonTemplates[templateID]
		if !ok {
			http.Error(w, `{"error":"invalid_template_id"}`, http.StatusBadRequest)
			return
		}
		listName := strings.TrimSpace(req.ListName)
		if listName == "" {
			listName = template.Label
		}
		listID := fmt.Sprintf("%s-%d", templateID, time.Now().UnixNano())
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"list_id":          listID,
			"list_name":        listName,
			"template_id":      template.ID,
			"default_fields":   template.DefaultFields,
			"default_filters":  template.DefaultFilters,
			"sort_order":       template.SortOrder,
			"share_visibility": "private",
		})
	})
	mux.HandleFunc("/api/integrations/pokemon/goal-bundles", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		bundles := make([]pokemonGoalBundle, 0, len(pokemonGoalBundles))
		for _, bundle := range pokemonGoalBundles {
			bundles = append(bundles, bundle)
		}
		sort.Slice(bundles, func(i, j int) bool {
			return bundles[i].ID < bundles[j].ID
		})
		_ = json.NewEncoder(w).Encode(map[string]any{
			"bundles": bundles,
		})
	})
	mux.HandleFunc("/api/integrations/pokemon/goal-bundles/apply", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			BundleID      string `json:"bundle_id"`
			WorkspaceName string `json:"workspace_name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		bundleID := strings.TrimSpace(req.BundleID)
		bundle, ok := pokemonGoalBundles[bundleID]
		if !ok {
			http.Error(w, `{"error":"invalid_bundle_id"}`, http.StatusBadRequest)
			return
		}
		workspaceName := strings.TrimSpace(req.WorkspaceName)
		if workspaceName == "" {
			workspaceName = bundle.Label
		}
		workspaceID := fmt.Sprintf("goal-%s-%d", bundleID, time.Now().UnixNano())
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"workspace_id":   workspaceID,
			"workspace_name": workspaceName,
			"bundle_id":      bundle.ID,
			"filters":        bundle.Filters,
			"actions":        bundle.Actions,
			"shortcut":       bundle.Shortcut,
		})
	})
	mux.HandleFunc("/api/integrations/pokemon/graded-overrides", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			itemID := strings.TrimSpace(r.URL.Query().Get("item_id"))
			if itemID == "" {
				http.Error(w, `{"error":"missing_item_id"}`, http.StatusBadRequest)
				return
			}
			active, _ := profiles.GetActiveProfile(r.Context())
			profileID := strings.TrimSpace(active.ID)
			var payload struct {
				ItemID                  string  `json:"item_id"`
				Grader                  string  `json:"grader"`
				GradeNumeric            float64 `json:"grade_numeric"`
				CertNumber              string  `json:"cert_number"`
				SlabState               string  `json:"slab_state"`
				ValuationOverrideAmount float64 `json:"valuation_override_amount"`
				Currency                string  `json:"currency"`
				SourceNote              string  `json:"source_note"`
				OverriddenAt            string  `json:"overridden_at"`
			}
			err := conn.QueryRowContext(r.Context(), `
				SELECT item_id, grader, grade_numeric, cert_number, slab_state, valuation_override_amount, currency, source_note, overridden_at
				FROM pokemon_graded_overrides
				WHERE item_id = ? AND (? = '' OR profile_id = ?)
			`, itemID, profileID, profileID).Scan(
				&payload.ItemID,
				&payload.Grader,
				&payload.GradeNumeric,
				&payload.CertNumber,
				&payload.SlabState,
				&payload.ValuationOverrideAmount,
				&payload.Currency,
				&payload.SourceNote,
				&payload.OverriddenAt,
			)
			if err == sql.ErrNoRows {
				http.Error(w, `{"error":"graded_override_not_found"}`, http.StatusNotFound)
				return
			}
			if err != nil {
				http.Error(w, `{"error":"failed_to_load_graded_override"}`, http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(payload)
		case http.MethodPost:
			var req struct {
				ItemID                  string  `json:"item_id"`
				Grader                  string  `json:"grader"`
				GradeNumeric            float64 `json:"grade_numeric"`
				CertNumber              string  `json:"cert_number"`
				SlabState               string  `json:"slab_state"`
				ValuationOverrideAmount float64 `json:"valuation_override_amount"`
				Currency                string  `json:"currency"`
				SourceNote              string  `json:"source_note"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			itemID := strings.TrimSpace(req.ItemID)
			if itemID == "" {
				http.Error(w, `{"error":"missing_item_id"}`, http.StatusBadRequest)
				return
			}
			active, _ := profiles.GetActiveProfile(r.Context())
			profileID := strings.TrimSpace(active.ID)
			var itemExists int
			if err := conn.QueryRowContext(r.Context(), `
				SELECT COUNT(1)
				FROM canonical_items
				WHERE id = ? AND (? = '' OR profile_id = ?)
			`, itemID, profileID, profileID).Scan(&itemExists); err != nil || itemExists == 0 {
				http.Error(w, `{"error":"item_not_found"}`, http.StatusNotFound)
				return
			}
			grader := strings.TrimSpace(req.Grader)
			certNumber := strings.TrimSpace(req.CertNumber)
			slabState := strings.TrimSpace(req.SlabState)
			currency := strings.ToUpper(strings.TrimSpace(req.Currency))
			if currency == "" {
				currency = "AUD"
			}
			sourceNote := strings.TrimSpace(req.SourceNote)
			_, err := conn.ExecContext(r.Context(), `
				INSERT INTO pokemon_graded_overrides(
					item_id, profile_id, grader, grade_numeric, cert_number, slab_state, valuation_override_amount, currency, source_note, overridden_at, updated_at
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
				ON CONFLICT(item_id) DO UPDATE SET
					profile_id = excluded.profile_id,
					grader = excluded.grader,
					grade_numeric = excluded.grade_numeric,
					cert_number = excluded.cert_number,
					slab_state = excluded.slab_state,
					valuation_override_amount = excluded.valuation_override_amount,
					currency = excluded.currency,
					source_note = excluded.source_note,
					overridden_at = CURRENT_TIMESTAMP,
					updated_at = CURRENT_TIMESTAMP
			`, itemID, profileID, grader, req.GradeNumeric, certNumber, slabState, req.ValuationOverrideAmount, currency, sourceNote)
			if err != nil {
				http.Error(w, `{"error":"failed_to_save_graded_override"}`, http.StatusBadRequest)
				return
			}

			slabbed := 0
			if slabState != "" && !strings.EqualFold(slabState, "unslabbed") {
				slabbed = 1
			}
			_, _ = conn.ExecContext(r.Context(), `
				UPDATE canonical_items
				SET grading_status = 'graded',
					grader = ?,
					grade_numeric = ?,
					slabbed = ?,
					updated_at = CURRENT_TIMESTAMP
				WHERE id = ? AND (? = '' OR profile_id = ?)
			`, grader, req.GradeNumeric, slabbed, itemID, profileID, profileID)

			var payload struct {
				ItemID                  string  `json:"item_id"`
				Grader                  string  `json:"grader"`
				GradeNumeric            float64 `json:"grade_numeric"`
				CertNumber              string  `json:"cert_number"`
				SlabState               string  `json:"slab_state"`
				ValuationOverrideAmount float64 `json:"valuation_override_amount"`
				Currency                string  `json:"currency"`
				SourceNote              string  `json:"source_note"`
				OverriddenAt            string  `json:"overridden_at"`
			}
			err = conn.QueryRowContext(r.Context(), `
				SELECT item_id, grader, grade_numeric, cert_number, slab_state, valuation_override_amount, currency, source_note, overridden_at
				FROM pokemon_graded_overrides
				WHERE item_id = ? AND (? = '' OR profile_id = ?)
			`, itemID, profileID, profileID).Scan(
				&payload.ItemID,
				&payload.Grader,
				&payload.GradeNumeric,
				&payload.CertNumber,
				&payload.SlabState,
				&payload.ValuationOverrideAmount,
				&payload.Currency,
				&payload.SourceNote,
				&payload.OverriddenAt,
			)
			if err != nil {
				http.Error(w, `{"error":"failed_to_load_graded_override"}`, http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(payload)
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
			settings := map[string]string{}
			if strings.TrimSpace(active.ID) != "" {
				settings, _ = profiles.GetSettings(r.Context(), active.ID)
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
			if validationErr := validateInventoryItemTaxonomy(req, nil, settings); validationErr != nil {
				writeInventoryTaxonomyValidationError(w, validationErr)
				return
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
		querySet, err := scannerSvc.GetQuerySetForProfile(r.Context(), strings.TrimSpace(active.ID), req.QuerySetID)
		if err != nil {
			http.Error(w, `{"error":"invalid_query_set"}`, http.StatusBadRequest)
			return
		}
		if providerID := firstProviderInScope(querySet.ProviderScope); providerID != "" {
			itemsPerPageKey := providerSettingsKeys(providerID).ItemsPerPageKey
			if raw := strings.TrimSpace(settings[itemsPerPageKey]); raw != "" {
				if value, parseErr := strconv.Atoi(raw); parseErr == nil {
					querySet.ItemsPerPage = value
					if _, updateErr := scannerSvc.UpdateQuerySetForProfile(
						r.Context(),
						strings.TrimSpace(active.ID),
						querySet.ID,
						querySet,
					); updateErr != nil {
						http.Error(w, `{"error":"failed_to_apply_provider_items_per_page"}`, http.StatusBadRequest)
						return
					}
				}
			}
		}
		provider := ebay.NewProvider(ebay.ProviderConfig{
			BaseURL:     settings["ebay_base_url"],
			BearerToken: settings["ebay_bearer_token"],
			Marketplace: settings["ebay_marketplace"],
		})
		out, err := scannerSvc.RunNowForProfile(r.Context(), strings.TrimSpace(active.ID), req.QuerySetID, provider)
		if err != nil {
			logSvc.Log(r.Context(), "error", "scanner_run_failed", map[string]any{"query_set_id": req.QuerySetID, "error": err.Error()})
			var providerErr *ebay.ProviderError
			if errors.As(err, &providerErr) && providerErr.ErrorCode != "" {
				status := providerErr.StatusCode
				if status <= 0 {
					status = http.StatusBadRequest
				}
				w.WriteHeader(status)
				payload := map[string]any{
					"error":        "failed_to_run_scanner",
					"error_code":   providerErr.ErrorCode,
					"provider":     "ebay",
					"message":      providerErr.Error(),
					"next_action":  ebayProviderErrorNextAction(providerErr),
					"query_set_id": req.QuerySetID,
				}
				if providerErr.RetryAfterSeconds > 0 {
					payload["retry_after_seconds"] = providerErr.RetryAfterSeconds
				}
				_ = json.NewEncoder(w).Encode(payload)
				return
			}
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
	mux.HandleFunc("/api/scanner/recognition-review/apply", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Candidates []scanner.RecognitionCandidateInput `json:"candidates"`
			Target     string                              `json:"target"`
			Confirmed  bool                                `json:"confirmed"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		target := normalizeScannerReviewApplyTarget(req.Target)
		for i := range req.Candidates {
			if strings.TrimSpace(req.Candidates[i].Target) == "" {
				req.Candidates[i].Target = target
			}
		}
		review, err := scanner.BuildRecognitionReview(req.Candidates)
		if err != nil {
			http.Error(w, `{"error":"invalid_recognition_review"}`, http.StatusBadRequest)
			return
		}
		review.Target = target
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil || strings.TrimSpace(active.ID) == "" {
			http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusBadRequest)
			return
		}
		if !req.Confirmed {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":              "scanner_review_confirmation_required",
				"confirmation_state": "required",
				"review":             review,
			})
			return
		}

		itemInput := scannerReviewApplyItem(review)
		createdItem, err := collectionRepo.CreateItemForProfile(r.Context(), strings.TrimSpace(active.ID), itemInput)
		if err != nil {
			http.Error(w, `{"error":"failed_to_create_scanner_item"}`, http.StatusBadRequest)
			return
		}
		result := map[string]any{
			"mode":               "scanner_review_apply",
			"profile_id":         strings.TrimSpace(active.ID),
			"target":             target,
			"confirmation_state": "confirmed",
			"review":             review,
			"item":               createdItem,
		}
		if target == "wishlist" {
			entry, err := wishlistSvc.CreateForProfile(r.Context(), strings.TrimSpace(active.ID), wishlist.Entry{
				ItemID:       createdItem.ID,
				Priority:     "medium",
				Notes:        scannerReviewApplyEvidenceNote(review),
				HighlightHit: true,
				Owned:        false,
			})
			if err != nil {
				http.Error(w, `{"error":"failed_to_create_scanner_wishlist_entry"}`, http.StatusBadRequest)
				return
			}
			result["wishlist_entry"] = entry
		}
		logSvc.Log(r.Context(), "info", "scanner_review_apply_confirmed", map[string]any{
			"profile_id": strings.TrimSpace(active.ID),
			"target":     target,
			"item_id":    createdItem.ID,
		})
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(result)
	})
	mux.HandleFunc("/api/scanner/failures", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":          "method_not_allowed",
				"error_code":     "method_not_allowed",
				"provider":       "ebay",
				"message":        "Use GET to list scanner failure snapshots before retrying eBay saved-search failures.",
				"next_action":    "retry_with_get",
				"allowed_method": http.MethodGet,
			})
			return
		}
		active, _ := profiles.GetActiveProfile(r.Context())
		items, err := scannerSvc.ListFailuresByProfile(r.Context(), strings.TrimSpace(active.ID))
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
		if strings.EqualFold(provider, "openai") {
			health := openAIProviderHealth(r.Context(), profiles)
			_ = json.NewEncoder(w).Encode(health)
			return
		}
		health, err := scannerSvc.ProviderHealth(r.Context(), provider)
		if err != nil {
			http.Error(w, `{"error":"failed_to_get_provider_health"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(providerHealthResponse(health))
	})
	mux.HandleFunc("/api/provider/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Provider  string `json:"provider"`
			ProfileID string `json:"profile_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		provider := strings.TrimSpace(req.Provider)
		if provider == "" {
			provider = "openai"
		}
		if !strings.EqualFold(provider, "openai") {
			http.Error(w, `{"error":"unsupported_provider_test"}`, http.StatusBadRequest)
			return
		}
		payload, statusCode := openAIProviderTest(r.Context(), profiles, aiSvc, strings.TrimSpace(req.ProfileID))
		if statusCode >= 400 {
			logSvc.Log(r.Context(), "error", "openai_provider_test_failed", map[string]any{
				"profile_id": strings.TrimSpace(req.ProfileID),
				"code":       payload["code"],
				"status":     payload["status"],
			})
			if statusCode == http.StatusBadRequest && payload["code"] == "OPENAI_PROVIDER_TEST_FAILED" {
				_, _ = conn.ExecContext(r.Context(), `INSERT INTO ai_failures(id, profile_id, message, created_at) VALUES (hex(randomblob(16)), ?, ?, CURRENT_TIMESTAMP)`, strings.TrimSpace(req.ProfileID), payload["message"])
			}
		} else {
			logSvc.Log(r.Context(), "info", "openai_provider_test_passed", map[string]any{
				"profile_id": strings.TrimSpace(req.ProfileID),
				"code":       payload["code"],
				"status":     payload["status"],
			})
		}
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(payload)
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
			if key, secretErr := profiles.GetSecret(r.Context(), strings.TrimSpace(active.ID), "openai_api_key"); secretErr == nil && strings.TrimSpace(key) != "" {
				settings["openai.api_key_secret_present"] = "true"
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"providers": providerRegistryPayload(r.Context(), conn, scannerSvc, amazonMode, settings),
		})
	})
	mux.HandleFunc("/api/companion/modules", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, "{\"error\":\"method_not_allowed\"}", http.StatusMethodNotAllowed)
			return
		}
		_ = json.NewEncoder(w).Encode(companionSvc.Registry())
	})
	mux.HandleFunc("/api/companion/payloads", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, "{\"error\":\"method_not_allowed\"}", http.StatusMethodNotAllowed)
			return
		}
		var req companion.PayloadSubmission
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "{\"error\":\"invalid_json\"}", http.StatusBadRequest)
			return
		}
		accepted, err := companionSvc.AcceptPayload(req, r.Header.Get("Authorization"))
		if err != nil {
			status := http.StatusBadRequest
			if err.Error() == "companion_auth_required" {
				status = http.StatusUnauthorized
			}
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(accepted)
	})
	mux.HandleFunc("/api/providers/ebay/buyer-interest/preview", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			SourceAccount       string                    `json:"source_account"`
			WriteBackCapability string                    `json:"write_back_capability"`
			Items               []ebay.BuyerInterestInput `json:"items"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		if len(req.Items) == 0 {
			http.Error(w, `{"error":"missing_items"}`, http.StatusBadRequest)
			return
		}
		mappings := make([]ebay.BuyerInterestMapping, 0, len(req.Items))
		summary := map[string]int{
			ebay.InterestDestinationWishlist:  0,
			ebay.InterestDestinationDiscovery: 0,
			"write_back_allowed":              0,
			"write_back_blocked":              0,
		}
		for _, item := range req.Items {
			if strings.TrimSpace(item.SourceAccount) == "" {
				item.SourceAccount = req.SourceAccount
			}
			if strings.TrimSpace(item.WriteBackCapability) == "" {
				item.WriteBackCapability = req.WriteBackCapability
			}
			mapped := ebay.MapBuyerInterest(item)
			mappings = append(mappings, mapped)
			summary[mapped.Destination]++
			if mapped.WriteBackAllowed {
				summary["write_back_allowed"]++
			} else {
				summary["write_back_blocked"]++
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"provider": "ebay",
			"mode":     "preview",
			"items":    mappings,
			"summary":  summary,
		})
	})
	mux.HandleFunc("/api/providers/ebay/seller-operations/preview", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req ebay.SellerOperationActionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		preview := ebay.PreviewSellerOperationAction(req)
		if preview.Operation == "" || preview.Action == "" {
			http.Error(w, `{"error":"missing_seller_operation_action"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"provider": "ebay",
			"mode":     "seller_operation_preview",
			"preview":  preview,
		})
	})
	mux.HandleFunc("/api/providers/ebay/seller-operations/execute", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req ebay.SellerOperationActionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		execution := ebay.ExecuteSellerOperationAction(req)
		if execution.Operation == "" || execution.Action == "" {
			http.Error(w, `{"error":"missing_seller_operation_action"}`, http.StatusBadRequest)
			return
		}
		status := http.StatusOK
		if !execution.Executed {
			status = http.StatusConflict
		}
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"provider":  "ebay",
			"mode":      "seller_operation_execute",
			"execution": execution,
		})
	})
	mux.HandleFunc("/api/providers/ebay/listing-lifecycle/preview", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req ebay.SellerListingLifecycleCommandRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		preview := ebay.PreviewSellerListingLifecycleCommand(req)
		if preview.Command == "" {
			http.Error(w, `{"error":"missing_listing_lifecycle_command"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"provider": "ebay",
			"mode":     "listing_lifecycle_preview",
			"preview":  preview,
		})
	})
	mux.HandleFunc("/api/providers/ebay/listing-lifecycle/execute", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req ebay.SellerListingLifecycleCommandRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		execution := ebay.ExecuteSellerListingLifecycleCommand(req, nil)
		if execution.Command == "" {
			http.Error(w, `{"error":"missing_listing_lifecycle_command"}`, http.StatusBadRequest)
			return
		}
		status := http.StatusOK
		if !execution.Executed {
			status = http.StatusConflict
		}
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"provider":  "ebay",
			"mode":      "listing_lifecycle_execute",
			"execution": execution,
		})
	})
	registerEbayBuyerInterestImportRoute(mux, conn, profiles)
	mux.HandleFunc("/api/providers/family-detect", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ProviderURL string `json:"provider_url"`
			HTML        string `json:"html"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		providerURL := strings.TrimSpace(req.ProviderURL)
		if providerURL == "" {
			http.Error(w, `{"error":"missing_provider_url"}`, http.StatusBadRequest)
			return
		}
		family, confidence, evidence, domain, err := detectProviderFamily(r.Context(), http.DefaultClient, providerURL, req.HTML)
		if err != nil {
			http.Error(w, `{"error":"failed_to_detect_provider_family"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"provider_domain":     domain,
			"proposed_api_family": family,
			"confidence":          confidence,
			"evidence":            evidence,
		})
	})
	mux.HandleFunc("/api/providers/family-override", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ProviderDomain string `json:"provider_domain"`
			APIFamily      string `json:"api_family"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		domain := normalizeProviderDomain(req.ProviderDomain)
		family := strings.TrimSpace(strings.ToLower(req.APIFamily))
		if domain == "" || family == "" {
			http.Error(w, `{"error":"missing_domain_or_family"}`, http.StatusBadRequest)
			return
		}
		if err := persistProviderFamilyOverride(r.Context(), conn, domain, family); err != nil {
			http.Error(w, `{"error":"failed_to_save_provider_family_override"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"saved":           true,
			"provider_domain": domain,
			"api_family":      family,
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
		run, err := scannerSvc.RunNowForProfile(r.Context(), strings.TrimSpace(active.ID), qs.ID, amazonContractProvider{candidates: candidates})
		if err != nil {
			http.Error(w, `{"error":"failed_to_run_amazon_provider"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"query_set_id": qs.ID,
			"candidates":   buildAmazonCandidateResponseContract(candidates),
			"run":          run,
		})
	})
	mux.HandleFunc("/api/providers/ebay/run", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		writeClientError := func(status int, errorCode, querySetID, nextAction, message string, fields map[string]any) {
			w.WriteHeader(status)
			payload := map[string]any{
				"error":        errorCode,
				"error_code":   errorCode,
				"provider":     "ebay",
				"query_set_id": strings.TrimSpace(querySetID),
				"next_action":  nextAction,
				"message":      message,
			}
			for key, value := range fields {
				payload[key] = value
			}
			_ = json.NewEncoder(w).Encode(payload)
		}
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":          "method_not_allowed",
				"error_code":     "method_not_allowed",
				"provider":       "ebay",
				"allowed_method": http.MethodPost,
				"next_action":    "retry_with_post",
				"message":        "Run eBay saved-search providers with POST and a query_set_id request body.",
			})
			return
		}
		var req struct {
			QuerySetID string `json:"query_set_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeClientError(http.StatusBadRequest, "invalid_json", "", "select_existing_ebay_query_set", "The eBay provider run request body must be valid JSON.", nil)
			return
		}
		req.QuerySetID = strings.TrimSpace(req.QuerySetID)
		if req.QuerySetID == "" {
			writeClientError(http.StatusBadRequest, "missing_query_set_id", req.QuerySetID, "select_existing_ebay_query_set", "Select an existing eBay query set before running the provider.", nil)
			return
		}
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil {
			writeClientError(http.StatusBadRequest, "active_profile_not_set", req.QuerySetID, "select_active_profile", "Select an active profile before running eBay saved searches.", nil)
			return
		}
		profileID := strings.TrimSpace(active.ID)
		settings, err := profiles.GetSettings(r.Context(), profileID)
		if err != nil {
			writeClientError(http.StatusBadRequest, "failed_to_get_settings", req.QuerySetID, "retry_profile_settings", "Profile settings could not be loaded for the eBay provider run.", nil)
			return
		}
		qs, err := scannerSvc.GetQuerySetForProfile(r.Context(), profileID, req.QuerySetID)
		if err != nil {
			writeClientError(http.StatusBadRequest, "invalid_query_set_id", req.QuerySetID, "select_existing_ebay_query_set", "The selected eBay query set no longer exists for the active profile.", nil)
			return
		}
		if !providerScopeIncludes(qs.ProviderScope, "ebay") {
			writeClientError(http.StatusBadRequest, "query_set_not_scoped_to_ebay", qs.ID, "choose_ebay_scoped_query_set", "Choose a query set whose provider scope includes eBay.", nil)
			return
		}
		if raw := strings.TrimSpace(settings[providerSettingsKeys("ebay").ItemsPerPageKey]); raw != "" {
			value, parseErr := strconv.Atoi(raw)
			if parseErr != nil || value <= 0 {
				writeClientError(http.StatusBadRequest, "invalid_ebay_items_per_page", qs.ID, "update_ebay_items_per_page", "Update the eBay setup page size to a positive integer before running this query.", map[string]any{
					"setting": providerSettingsKeys("ebay").ItemsPerPageKey,
				})
				return
			}
			qs.ItemsPerPage = value
			if _, updateErr := scannerSvc.UpdateQuerySetForProfile(
				r.Context(),
				profileID,
				qs.ID,
				qs,
			); updateErr != nil {
				writeClientError(http.StatusBadRequest, "failed_to_apply_provider_items_per_page", qs.ID, "update_ebay_items_per_page", "The eBay setup page size could not be applied to this query set.", map[string]any{
					"setting": providerSettingsKeys("ebay").ItemsPerPageKey,
				})
				return
			}
		}
		provider := ebay.NewProvider(ebay.ProviderConfig{
			BaseURL:     settings["ebay_base_url"],
			BearerToken: settings["ebay_bearer_token"],
			Marketplace: settings["ebay_marketplace"],
		})
		run, err := scannerSvc.RunNowForProfile(r.Context(), profileID, qs.ID, provider)
		if err != nil {
			var providerErr *ebay.ProviderError
			if errors.As(err, &providerErr) && providerErr.ErrorCode != "" {
				status := providerErr.StatusCode
				if status <= 0 {
					status = http.StatusBadRequest
				}
				w.WriteHeader(status)
				payload := map[string]any{
					"error":        "failed_to_run_ebay_provider",
					"error_code":   providerErr.ErrorCode,
					"provider":     "ebay",
					"message":      providerErr.Error(),
					"next_action":  ebayProviderErrorNextAction(providerErr),
					"query_set_id": qs.ID,
				}
				if providerErr.RetryAfterSeconds > 0 {
					payload["retry_after_seconds"] = providerErr.RetryAfterSeconds
				}
				_ = json.NewEncoder(w).Encode(payload)
				return
			}
			http.Error(w, `{"error":"failed_to_run_ebay_provider"}`, http.StatusBadRequest)
			return
		}
		candidates, err := scannerSvc.ListCandidatesByProfile(r.Context(), profileID, qs.ID)
		if err != nil {
			writeClientError(http.StatusBadRequest, "failed_to_list_ebay_candidates", qs.ID, "select_existing_ebay_query_set", "The eBay run completed but persisted candidates could not be reloaded.", nil)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"query_set_id": qs.ID,
			"provider":     "ebay",
			"candidates":   candidates,
			"run":          run,
		})
	})
	mux.HandleFunc("/api/providers/bonza/run", func(w http.ResponseWriter, r *http.Request) {
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
		profileID := strings.TrimSpace(active.ID)
		qs, err := scannerSvc.GetQuerySetForProfile(r.Context(), profileID, req.QuerySetID)
		if err != nil {
			http.Error(w, `{"error":"invalid_query_set_id"}`, http.StatusBadRequest)
			return
		}
		settings, err := profiles.GetSettings(r.Context(), profileID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_get_settings"}`, http.StatusBadRequest)
			return
		}
		baseURL := strings.TrimSpace(settings["integration.bonzaslotcars.base_url"])
		if baseURL == "" {
			baseURL = "https://bonzaslotcars.com.au"
		}
		requested := parsePositiveInt(settings["integration.bonzaslotcars.items_per_page"], 36)
		if requested > 36 {
			requested = 36
		}
		searchResult, err := runBonzaSearch(r.Context(), http.DefaultClient, baseURL, qs, requested)
		if err != nil {
			http.Error(w, `{"error":"failed_to_run_bonza"}`, http.StatusBadRequest)
			return
		}
		run, err := scannerSvc.PersistCandidatesForProfile(
			r.Context(),
			profileID,
			qs.ID,
			bonzaCandidatesForScanner(searchResult.Candidates),
			1,
			requested,
			searchResult.ItemsPerPageUsed,
			"",
		)
		if err != nil {
			http.Error(w, `{"error":"failed_to_persist_bonza_candidates"}`, http.StatusBadRequest)
			return
		}
		persisted, err := scannerSvc.ListCandidatesByProfile(r.Context(), profileID, qs.ID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_list_bonza_candidates"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"query_set_id":        qs.ID,
			"page_count":          searchResult.PageCount,
			"observed_page_size":  searchResult.ObservedPageSize,
			"items_per_page_used": searchResult.ItemsPerPageUsed,
			"candidates":          persisted,
			"run":                 run,
			"run_summary": map[string]any{
				"page_count":         searchResult.PageCount,
				"observed_page_size": searchResult.ObservedPageSize,
				"candidates_total":   run.Saved,
			},
		})
	})
	mux.HandleFunc("/api/providers/product-url/ingest", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		route, err := detectProviderProductURL(req.URL)
		if err != nil {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"mode":  "provider_product_url_ingest",
				"error": "unsupported_provider_url",
			})
			return
		}
		if route.Provider == "bonzaslotcars" && route.Action != "ingest_product_url" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"mode":     "provider_product_url_ingest",
				"error":    "supported_provider_unsupported_page",
				"provider": route.Provider,
				"family":   route.Family,
				"route":    route,
			})
			return
		}
		if route.Provider != "bonzaslotcars" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"mode":  "provider_product_url_ingest",
				"error": "unsupported_provider_url",
			})
			return
		}
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil || strings.TrimSpace(active.ID) == "" {
			http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusBadRequest)
			return
		}
		settings, err := profiles.GetSettings(r.Context(), strings.TrimSpace(active.ID))
		if err != nil {
			http.Error(w, `{"error":"failed_to_get_settings"}`, http.StatusBadRequest)
			return
		}
		baseURL := strings.TrimSpace(settings["integration.bonzaslotcars.base_url"])
		if baseURL == "" {
			baseURL = "https://bonzaslotcars.com.au"
		}
		draft, err := ingestBonzaProductURL(r.Context(), http.DefaultClient, baseURL, route)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"mode":     "provider_product_url_ingest",
				"error":    "failed_to_ingest_bonza_product_url",
				"provider": route.Provider,
				"family":   route.Family,
				"route":    route,
			})
			return
		}
		existingItems, err := collectionRepo.ListItemsByProfile(r.Context(), strings.TrimSpace(active.ID))
		if err != nil {
			http.Error(w, `{"error":"failed_to_check_provider_product_duplicates"}`, http.StatusInternalServerError)
			return
		}
		duplicates := providerProductDuplicateCandidates(existingItems, route, draft)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"mode":       "provider_product_url_ingest",
			"provider":   route.Provider,
			"family":     route.Family,
			"route":      route,
			"draft":      draft,
			"evidence":   draft.Evidence,
			"duplicates": duplicates,
		})
	})
	mux.HandleFunc("/api/providers/frontline/discovery", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			AssetURL          string   `json:"asset_url"`
			FallbackAssetURLs []string `json:"fallback_asset_urls"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		req.AssetURL = strings.TrimSpace(req.AssetURL)
		if req.AssetURL == "" {
			http.Error(w, `{"error":"missing_asset_url"}`, http.StatusBadRequest)
			return
		}
		config, fallbackUsed, warning, err := discoverFrontlineAlgoliaConfig(
			r.Context(),
			conn,
			http.DefaultClient,
			req.AssetURL,
			req.FallbackAssetURLs,
		)
		if err != nil {
			http.Error(w, `{"error":"failed_to_discover_frontline_config"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"config":        config,
			"fallback_used": fallbackUsed,
			"warning":       warning,
		})
	})
	mux.HandleFunc("/api/providers/frontline/run", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			QuerySetID                 string   `json:"query_set_id"`
			DiscoveryAssetURL          string   `json:"discovery_asset_url"`
			FallbackDiscoveryAssetURLs []string `json:"fallback_discovery_asset_urls"`
			SearchURL                  string   `json:"search_url"`
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
		profileID := strings.TrimSpace(active.ID)
		qs, err := scannerSvc.GetQuerySetForProfile(r.Context(), profileID, req.QuerySetID)
		if err != nil {
			http.Error(w, `{"error":"invalid_query_set_id"}`, http.StatusBadRequest)
			return
		}
		settings, err := profiles.GetSettings(r.Context(), profileID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_get_settings"}`, http.StatusBadRequest)
			return
		}
		baseURL := strings.TrimSpace(settings["integration.frontlinehobbies.base_url"])
		if baseURL == "" {
			baseURL = "https://frontlinehobbies.com.au"
		}
		itemsPerPage := parsePositiveInt(settings["integration.frontlinehobbies.items_per_page"], 24)
		if itemsPerPage > 50 {
			itemsPerPage = 50
		}
		discoveryAssetURL := strings.TrimSpace(req.DiscoveryAssetURL)
		if discoveryAssetURL == "" {
			discoveryAssetURL = strings.TrimRight(baseURL, "/") + "/assets/pd-search.js"
		}
		cfg, fallbackUsed, warning, err := discoverFrontlineAlgoliaConfig(
			r.Context(),
			conn,
			http.DefaultClient,
			discoveryAssetURL,
			req.FallbackDiscoveryAssetURLs,
		)
		if err != nil {
			http.Error(w, `{"error":"failed_to_discover_frontline_config"}`, http.StatusBadRequest)
			return
		}
		searchURL := strings.TrimSpace(req.SearchURL)
		if searchURL == "" {
			searchURL = defaultFrontlineAlgoliaSearchURL(cfg)
		}
		candidates, total, runErr := runFrontlineAlgoliaSearch(r.Context(), http.DefaultClient, searchURL, qs, cfg, baseURL, itemsPerPage)
		if runErr != nil {
			http.Error(w, `{"error":"failed_to_run_frontline_provider"}`, http.StatusBadRequest)
			return
		}
		if fallbackUsed && strings.TrimSpace(warning) == "" {
			warning = "frontline discovery used cached config"
		}
		run, err := scannerSvc.PersistCandidatesForProfile(
			r.Context(),
			profileID,
			qs.ID,
			frontlineCandidatesForScanner(candidates),
			1,
			itemsPerPage,
			itemsPerPage,
			"",
		)
		if err != nil {
			http.Error(w, `{"error":"failed_to_persist_frontline_candidates"}`, http.StatusBadRequest)
			return
		}
		persisted, err := scannerSvc.ListCandidatesByProfile(r.Context(), profileID, qs.ID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_list_frontline_candidates"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"query_set_id": qs.ID,
			"total":        total,
			"candidates":   persisted,
			"warning":      warning,
			"discovery_config": map[string]any{
				"application_id": cfg.ApplicationID,
				"index_names":    cfg.IndexNames,
				"source":         cfg.Source,
			},
			"run": run,
			"run_summary": map[string]any{
				"candidates_total": run.Saved,
				"total":            total,
			},
		})
	})
	mux.HandleFunc("/api/providers/doofinder/discovery", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			AssetURL          string   `json:"asset_url"`
			FallbackAssetURLs []string `json:"fallback_asset_urls"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		req.AssetURL = strings.TrimSpace(req.AssetURL)
		if req.AssetURL == "" {
			http.Error(w, `{"error":"missing_asset_url"}`, http.StatusBadRequest)
			return
		}
		config, fallbackUsed, warning, err := discoverDoofinderConfig(
			r.Context(),
			conn,
			http.DefaultClient,
			req.AssetURL,
			req.FallbackAssetURLs,
		)
		if err != nil {
			http.Error(w, `{"error":"failed_to_discover_doofinder_config"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"config":        config,
			"fallback_used": fallbackUsed,
			"warning":       warning,
		})
	})
	mux.HandleFunc("/api/providers/doofinder/run", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			QuerySetID        string   `json:"query_set_id"`
			AssetURL          string   `json:"asset_url"`
			SearchURL         string   `json:"search_url"`
			ProviderDomain    string   `json:"provider_domain"`
			FallbackAssetURLs []string `json:"fallback_asset_urls"`
			Page              int      `json:"page"`
			PageSize          int      `json:"page_size"`
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
		profileID := strings.TrimSpace(active.ID)
		qs, err := scannerSvc.GetQuerySetForProfile(r.Context(), profileID, req.QuerySetID)
		if err != nil {
			http.Error(w, `{"error":"invalid_query_set_id"}`, http.StatusBadRequest)
			return
		}
		settings, err := profiles.GetSettings(r.Context(), profileID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_get_settings"}`, http.StatusBadRequest)
			return
		}
		providerDomain := strings.TrimSpace(strings.ToLower(req.ProviderDomain))
		if providerDomain == "" {
			providerDomain = "mrtoys.com.au"
		}
		baseURL := strings.TrimSpace(settings["integration."+strings.ReplaceAll(providerDomain, ".", "-")+".base_url"])
		if baseURL == "" {
			baseURL = "https://" + providerDomain
		}
		discoveryAssetURL := strings.TrimSpace(req.AssetURL)
		if discoveryAssetURL == "" {
			discoveryAssetURL = strings.TrimRight(baseURL, "/") + "/assets/doofinder.js"
		}
		config, fallbackUsed, warning, err := discoverDoofinderConfig(
			r.Context(),
			conn,
			http.DefaultClient,
			discoveryAssetURL,
			req.FallbackAssetURLs,
		)
		if err != nil {
			http.Error(w, `{"error":"failed_to_discover_doofinder_config"}`, http.StatusBadRequest)
			return
		}
		page := req.Page
		if page <= 0 {
			page = 1
		}
		pageSize := req.PageSize
		if pageSize <= 0 {
			pageSize = 24
		}
		if pageSize > 50 {
			pageSize = 50
		}
		query := strings.TrimSpace(qs.Name)
		if len(qs.Keywords) > 0 && strings.TrimSpace(qs.Keywords[0]) != "" {
			query = strings.TrimSpace(qs.Keywords[0])
		}
		if query == "" {
			query = "collectible"
		}
		searchURL := strings.TrimSpace(req.SearchURL)
		if searchURL == "" {
			searchURL = strings.TrimRight(config.APIBase, "/") + "/5/search"
		}
		candidates, total, runErr := runDoofinderSearch(
			r.Context(),
			http.DefaultClient,
			searchURL,
			query,
			page,
			pageSize,
			config.HashID,
			baseURL,
			providerDomain,
		)
		if runErr != nil {
			http.Error(w, `{"error":"failed_to_run_doofinder_provider"}`, http.StatusBadRequest)
			return
		}
		if fallbackUsed && strings.TrimSpace(warning) == "" {
			warning = "doofinder discovery used cached config"
		}
		run, err := scannerSvc.PersistCandidatesForProfile(
			r.Context(),
			profileID,
			qs.ID,
			doofinderCandidatesForScanner(candidates, providerDomain),
			1,
			pageSize,
			pageSize,
			"",
		)
		if err != nil {
			http.Error(w, `{"error":"failed_to_persist_doofinder_candidates"}`, http.StatusBadRequest)
			return
		}
		persisted, err := scannerSvc.ListCandidatesByProfile(r.Context(), profileID, qs.ID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_list_doofinder_candidates"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"query_set_id":    qs.ID,
			"page":            page,
			"page_size":       pageSize,
			"total":           total,
			"candidate_count": len(persisted),
			"candidates":      persisted,
			"warning":         warning,
			"run":             run,
			"run_summary": map[string]any{
				"page":                page,
				"effective_page_size": pageSize,
				"candidates_total":    run.Saved,
			},
			"discovery": map[string]any{
				"store":    config.Store,
				"zone":     config.Zone,
				"hashid":   config.HashID,
				"source":   config.Source,
				"api_base": config.APIBase,
			},
		})
	})
	mux.HandleFunc("/api/providers/bigcommerce/run", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			QuerySetID     string `json:"query_set_id"`
			ProviderDomain string `json:"provider_domain"`
			SearchURL      string `json:"search_url"`
			GraphQLURL     string `json:"graphql_url"`
			Page           int    `json:"page"`
			PageSize       int    `json:"page_size"`
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
		profileID := strings.TrimSpace(active.ID)
		qs, err := scannerSvc.GetQuerySetForProfile(r.Context(), profileID, req.QuerySetID)
		if err != nil {
			http.Error(w, `{"error":"invalid_query_set_id"}`, http.StatusBadRequest)
			return
		}
		settings, err := profiles.GetSettings(r.Context(), profileID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_get_settings"}`, http.StatusBadRequest)
			return
		}
		providerDomain := strings.TrimSpace(strings.ToLower(req.ProviderDomain))
		if providerDomain == "" {
			providerDomain = "voglers.com.au"
		}
		providerID := "au-webshop-" + strings.ReplaceAll(providerDomain, ".", "-")
		keys := providerSettingsKeys(providerID)
		token := strings.TrimSpace(settings[keys.TokenKey])
		if token == "" {
			token = strings.TrimSpace(settings["integration."+strings.ReplaceAll(providerDomain, ".", "-")+".token"])
		}
		page := req.Page
		if page <= 0 {
			page = 1
		}
		pageSize := req.PageSize
		if pageSize <= 0 {
			pageSize = 24
		}
		if pageSize > 50 {
			pageSize = 50
		}
		query := strings.TrimSpace(qs.Name)
		if len(qs.Keywords) > 0 && strings.TrimSpace(qs.Keywords[0]) != "" {
			query = strings.TrimSpace(qs.Keywords[0])
		}
		if query == "" {
			query = "collectible"
		}

		authMode := "storefront_public"
		dataDepthSource := "storefront_public"
		capabilityLimits := []string{"stock_depth_limited_without_token"}
		candidates := []map[string]any{}

		if token != "" {
			authMode = "token_enabled"
			dataDepthSource = "token_enabled"
			capabilityLimits = []string{}
			graphURL := strings.TrimSpace(req.GraphQLURL)
			if graphURL == "" {
				graphURL = "https://" + providerDomain + "/graphql"
			}
			out, runErr := runBigCommerceTokenSearch(r.Context(), http.DefaultClient, graphURL, token, query, providerDomain)
			if runErr != nil {
				http.Error(w, `{"error":"failed_to_run_bigcommerce_token_mode"}`, http.StatusBadRequest)
				return
			}
			candidates = out
		} else {
			searchURL := strings.TrimSpace(req.SearchURL)
			if searchURL == "" {
				searchURL = "https://" + providerDomain + "/products/search"
			}
			out, runErr := runBigCommerceStorefrontSearch(r.Context(), http.DefaultClient, searchURL, query, page, pageSize, providerDomain)
			if runErr != nil {
				http.Error(w, `{"error":"failed_to_run_bigcommerce_storefront_mode"}`, http.StatusBadRequest)
				return
			}
			candidates = out
		}

		run, err := scannerSvc.PersistCandidatesForProfile(
			r.Context(),
			profileID,
			qs.ID,
			bigCommerceCandidatesForScanner(candidates, providerDomain),
			1,
			pageSize,
			pageSize,
			"",
		)
		if err != nil {
			http.Error(w, `{"error":"failed_to_persist_bigcommerce_candidates"}`, http.StatusBadRequest)
			return
		}
		persisted, err := scannerSvc.ListCandidatesByProfile(r.Context(), profileID, qs.ID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_list_bigcommerce_candidates"}`, http.StatusBadRequest)
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"query_set_id":      qs.ID,
			"provider_domain":   providerDomain,
			"auth_mode":         authMode,
			"data_depth_source": dataDepthSource,
			"capability_limits": capabilityLimits,
			"candidate_count":   len(persisted),
			"candidates":        persisted,
			"run":               run,
			"run_summary": map[string]any{
				"auth_mode":         authMode,
				"data_depth_source": dataDepthSource,
				"candidates_total":  run.Saved,
			},
		})
	})
	mux.HandleFunc("/api/providers/hobbytech/run", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			QuerySetID                 string   `json:"query_set_id"`
			DiscoveryAssetURL          string   `json:"discovery_asset_url"`
			FallbackDiscoveryAssetURLs []string `json:"fallback_discovery_asset_urls"`
			SearchURL                  string   `json:"search_url"`
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
		profileID := strings.TrimSpace(active.ID)
		qs, err := scannerSvc.GetQuerySetForProfile(r.Context(), profileID, req.QuerySetID)
		if err != nil {
			http.Error(w, `{"error":"invalid_query_set_id"}`, http.StatusBadRequest)
			return
		}
		settings, err := profiles.GetSettings(r.Context(), profileID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_get_settings"}`, http.StatusBadRequest)
			return
		}
		baseURL := strings.TrimSpace(settings["integration.hobbytechtoys.base_url"])
		if baseURL == "" {
			baseURL = "https://hobbytechtoys.com.au"
		}
		itemsPerPage := parsePositiveInt(settings["integration.hobbytechtoys.items_per_page"], 24)
		if itemsPerPage > 50 {
			itemsPerPage = 50
		}
		discoveryAssetURL := strings.TrimSpace(req.DiscoveryAssetURL)
		if discoveryAssetURL == "" {
			discoveryAssetURL = strings.TrimRight(baseURL, "/") + "/assets/hobby-search.js"
		}
		searchURL := strings.TrimSpace(req.SearchURL)
		if searchURL == "" {
			searchURL = strings.TrimRight(baseURL, "/") + "/services.mybcapps.com/bc-sf-filter/search"
		}
		cfg, discoveryFallbackUsed, discoveryWarning, err := discoverHobbytechBoostConfig(
			r.Context(),
			conn,
			http.DefaultClient,
			discoveryAssetURL,
			req.FallbackDiscoveryAssetURLs,
			searchURL,
		)
		if err != nil {
			http.Error(w, `{"error":"failed_to_discover_hobbytech_config"}`, http.StatusBadRequest)
			return
		}
		candidates, pageCount, runErr := runHobbytechSearch(r.Context(), http.DefaultClient, qs, cfg, itemsPerPage)
		driftRecovered := false
		warning := discoveryWarning
		if runErr != nil {
			recoveryAssetURL := discoveryAssetURL
			recoveryFallbackAssets := req.FallbackDiscoveryAssetURLs
			if len(req.FallbackDiscoveryAssetURLs) > 0 {
				recoveryAssetURL = strings.TrimSpace(req.FallbackDiscoveryAssetURLs[0])
				recoveryFallbackAssets = req.FallbackDiscoveryAssetURLs[1:]
			}
			recoveredCfg, fallbackUsed, fallbackWarning, fallbackErr := discoverHobbytechBoostConfig(
				r.Context(),
				conn,
				http.DefaultClient,
				recoveryAssetURL,
				recoveryFallbackAssets,
				searchURL,
			)
			if fallbackErr == nil {
				candidates, pageCount, runErr = runHobbytechSearch(r.Context(), http.DefaultClient, qs, recoveredCfg, itemsPerPage)
				if runErr == nil {
					driftRecovered = true
					if strings.TrimSpace(fallbackWarning) != "" {
						warning = fallbackWarning
					} else if fallbackUsed {
						warning = "hobbytech discovery fallback used cached config"
					} else {
						warning = "hobbytech session drift recovered via fallback discovery"
					}
				}
			}
		}
		if runErr != nil {
			http.Error(w, `{"error":"failed_to_run_hobbytech_provider"}`, http.StatusBadRequest)
			return
		}
		if discoveryFallbackUsed && strings.TrimSpace(warning) == "" {
			warning = "hobbytech discovery used cached config"
		}
		run, err := scannerSvc.PersistCandidatesForProfile(
			r.Context(),
			profileID,
			qs.ID,
			hobbytechCandidatesForScanner(candidates),
			1,
			itemsPerPage,
			itemsPerPage,
			"",
		)
		if err != nil {
			http.Error(w, `{"error":"failed_to_persist_hobbytech_candidates"}`, http.StatusBadRequest)
			return
		}
		persisted, err := scannerSvc.ListCandidatesByProfile(r.Context(), profileID, qs.ID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_list_hobbytech_candidates"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"query_set_id":     qs.ID,
			"page_count":       pageCount,
			"candidates":       persisted,
			"drift_recovered":  driftRecovered,
			"warning":          warning,
			"discovery_config": cfg,
			"run":              run,
			"run_summary": map[string]any{
				"page_count":       pageCount,
				"candidates_total": run.Saved,
			},
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
			Query:           r.URL.Query().Get("q"),
			PriceMax:        priceMax,
			DateFrom:        r.URL.Query().Get("date_from"),
			IncludeArchived: r.URL.Query().Get("include_archived") == "true",
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
		result, err := discoverySvc.ApplyActionWithResult(r.Context(), req)
		if err != nil {
			http.Error(w, `{"error":"failed_to_apply_discovery_action"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(result)
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
	mux.HandleFunc("/api/wishlist/convert-owned", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil || strings.TrimSpace(active.ID) == "" {
			http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusBadRequest)
			return
		}
		profileID := strings.TrimSpace(active.ID)
		var req struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		id := strings.TrimSpace(req.ID)
		if id == "" {
			http.Error(w, `{"error":"missing_id"}`, http.StatusBadRequest)
			return
		}
		if err := wishlistSvc.ConvertToOwnedForProfile(r.Context(), profileID, id); err != nil {
			http.Error(w, `{"error":"failed_to_convert_wishlist_to_owned"}`, http.StatusBadRequest)
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
			showDeleted := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("deleted")), "true")
			items, err := wishlistSvc.ListByProfileDeleted(r.Context(), profileID, showDeleted)
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
			var raw map[string]json.RawMessage
			if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			encoded, err := json.Marshal(raw)
			if err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			if err := json.Unmarshal(encoded, &req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(req.ID) != "" {
				existing, err := wishlistSvc.GetByIDForProfile(r.Context(), profileID, req.ID)
				if err == nil {
					if _, ok := raw["item_id"]; !ok {
						req.ItemID = existing.ItemID
					}
					if _, ok := raw["target_price"]; !ok {
						req.TargetPrice = existing.TargetPrice
					}
					if _, ok := raw["priority"]; !ok {
						req.Priority = existing.Priority
					}
					if _, ok := raw["notes"]; !ok {
						req.Notes = existing.Notes
					}
					if _, ok := raw["highlight_hit"]; !ok {
						req.HighlightHit = existing.HighlightHit
					}
					if _, ok := raw["below_target_now"]; !ok {
						req.BelowTargetNow = existing.BelowTargetNow
					}
					if _, ok := raw["owned"]; !ok {
						req.Owned = existing.Owned
					}
					if _, ok := raw["delivered"]; !ok {
						req.Delivered = existing.Delivered
					}
					if _, ok := raw["price_paid"]; !ok {
						req.PricePaid = existing.PricePaid
					}
					if _, ok := raw["purchase_url"]; !ok {
						req.PurchaseURL = existing.PurchaseURL
					}
					if _, ok := raw["purchase_date"]; !ok {
						req.PurchaseDate = existing.PurchaseDate
					}
					if _, ok := raw["purchase_condition"]; !ok {
						req.PurchaseCondition = existing.PurchaseCondition
					}
					if _, ok := raw["quantity"]; !ok {
						req.Quantity = existing.Quantity
					}
					if _, ok := raw["needed_quantity"]; !ok {
						req.NeededQuantity = existing.NeededQuantity
					}
				}
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
			var err error
			if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("permanent")), "true") {
				err = wishlistSvc.PermanentDeleteForProfile(r.Context(), profileID, id)
			} else {
				err = wishlistSvc.DeleteForProfile(r.Context(), profileID, id)
			}
			if err != nil {
				http.Error(w, `{"error":"failed_to_delete_wishlist"}`, http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/wishlist/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		active, _ := profiles.GetActiveProfile(r.Context())
		profileID := strings.TrimSpace(active.ID)
		path := strings.TrimPrefix(r.URL.Path, "/api/wishlist/")
		if !strings.HasSuffix(path, "/restore") || r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimSpace(strings.TrimSuffix(path, "/restore"))
		id = strings.Trim(id, "/")
		if id == "" {
			http.Error(w, `{"error":"missing_id"}`, http.StatusBadRequest)
			return
		}
		if err := wishlistSvc.RestoreForProfile(r.Context(), profileID, id); err != nil {
			http.Error(w, `{"error":"failed_to_restore_wishlist"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
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
	mux.HandleFunc("/api/commerce/landed-cost/plan", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			costing.AllocationRequest
			Consolidation struct {
				ShipmentFeeCents      int64 `json:"shipment_fee_cents"`
				DestinationLimitCents int64 `json:"destination_limit_cents"`
				WarningBufferCents    int64 `json:"warning_buffer_cents"`
			} `json:"consolidation"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		allocation, err := costing.AllocateLandedCosts(req.AllocationRequest)
		if err != nil {
			http.Error(w, `{"error":"invalid_landed_cost_allocation"}`, http.StatusBadRequest)
			return
		}
		plan, err := costing.PlanConsolidation(costing.ConsolidationRequest{
			Items:                 allocation.Items,
			ShipmentFeeCents:      req.Consolidation.ShipmentFeeCents,
			DestinationLimitCents: req.Consolidation.DestinationLimitCents,
			WarningBufferCents:    req.Consolidation.WarningBufferCents,
		})
		if err != nil {
			http.Error(w, `{"error":"invalid_consolidation_plan"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"mode":          "landed_cost_plan",
			"mutable":       false,
			"allocation":    allocation,
			"consolidation": plan,
		})
	})
	mux.HandleFunc("/api/commerce/lifecycle", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil {
			http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusBadRequest)
			return
		}
		profileID := strings.TrimSpace(active.ID)
		switch r.Method {
		case http.MethodGet:
			itemID := strings.TrimSpace(r.URL.Query().Get("item_id"))
			items, err := commerceSvc.ListLifecycleByProfile(r.Context(), profileID, itemID)
			if err != nil {
				http.Error(w, `{"error":"failed_to_list_commerce_lifecycle"}`, http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
		case http.MethodPost:
			var req commerce.LifecycleEntry
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			created, arrival, err := commerceSvc.CreateLifecycleForProfile(r.Context(), profileID, req)
			if err != nil {
				http.Error(w, `{"error":"failed_to_create_commerce_lifecycle"}`, http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"entry": created, "expected_arrival": arrival})
		default:
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/commerce/purchase-orders", func(w http.ResponseWriter, r *http.Request) {
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
		page, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("page")))
		pageSize, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("page_size")))
		orders, err := commerceSvc.ListPurchaseOrdersByProfile(
			r.Context(),
			strings.TrimSpace(active.ID),
			strings.TrimSpace(r.URL.Query().Get("status")),
			strings.TrimSpace(r.URL.Query().Get("search")),
			page,
			pageSize,
		)
		if err != nil {
			http.Error(w, `{"error":"failed_to_list_purchase_orders"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(orders)
	})
	mux.HandleFunc("/api/commerce/arrivals", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil {
			http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusBadRequest)
			return
		}
		profileID := strings.TrimSpace(active.ID)
		switch r.Method {
		case http.MethodGet:
			itemID := strings.TrimSpace(r.URL.Query().Get("item_id"))
			status := strings.TrimSpace(r.URL.Query().Get("status"))
			items, err := commerceSvc.ListArrivalsByProfile(r.Context(), profileID, itemID, status)
			if err != nil {
				http.Error(w, `{"error":"failed_to_list_expected_arrivals"}`, http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
		case http.MethodPost:
			var req commerce.ExpectedArrival
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			created, err := commerceSvc.CreateArrivalForProfile(r.Context(), profileID, req)
			if err != nil {
				http.Error(w, `{"error":"failed_to_create_expected_arrival"}`, http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(created)
		case http.MethodPut:
			var req commerce.ExpectedArrival
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			if err := commerceSvc.UpdateArrivalForProfile(r.Context(), profileID, req); err != nil {
				http.Error(w, `{"error":"failed_to_update_expected_arrival"}`, http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
		default:
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/forwarding/packages", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			profileID := strings.TrimSpace(r.URL.Query().Get("profile_id"))
			status := strings.TrimSpace(r.URL.Query().Get("status"))
			packages, err := forwarderInbox.ListPackages(r.Context(), profileID, status)
			if err != nil {
				http.Error(w, "{\"error\":\"failed_to_list_forwarder_packages\"}", http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"packages": packages,
				"summary":  map[string]int{"count": len(packages)},
			})
		case http.MethodPost:
			var req forwarding.PackageImport
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "{\"error\":\"invalid_json\"}", http.StatusBadRequest)
				return
			}
			pkg, err := forwarderInbox.UpsertPackage(r.Context(), req)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error":   "invalid_forwarder_package",
					"message": err.Error(),
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"package": pkg,
				"mode":    "forwarder_package_upsert",
			})
		default:
			http.Error(w, "{\"error\":\"method_not_allowed\"}", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/forwarding/package-links", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil {
			http.Error(w, "{\"error\":\"active_profile_not_set\"}", http.StatusBadRequest)
			return
		}
		profileID := strings.TrimSpace(active.ID)
		switch r.Method {
		case http.MethodGet:
			packageID := strings.TrimSpace(r.URL.Query().Get("package_id"))
			links, err := forwarderInbox.ListPackageLinks(r.Context(), profileID, packageID)
			if err != nil {
				http.Error(w, "{\"error\":\"failed_to_list_forwarder_package_links\"}", http.StatusInternalServerError)
				return
			}
			events, err := forwarderInbox.ListPackageLinkEvents(r.Context(), profileID, packageID)
			if err != nil {
				http.Error(w, "{\"error\":\"failed_to_list_forwarder_package_link_events\"}", http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"links":   links,
				"events":  events,
				"summary": map[string]int{"count": len(links), "events": len(events)},
			})
		case http.MethodPost:
			var req forwarding.PackageLinkRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "{\"error\":\"invalid_json\"}", http.StatusBadRequest)
				return
			}
			req.ProfileID = profileID
			link, err := forwarderInbox.LinkPackage(r.Context(), req)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error":   "invalid_forwarder_package_link",
					"message": err.Error(),
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"link": link,
				"mode": "forwarder_package_reconciliation_link",
			})
		case http.MethodDelete:
			var req forwarding.PackageUnlinkRequest
			if r.Body != nil {
				_ = json.NewDecoder(r.Body).Decode(&req)
			}
			req.ProfileID = profileID
			if req.PackageID == "" {
				req.PackageID = strings.TrimSpace(r.URL.Query().Get("package_id"))
			}
			event, err := forwarderInbox.UnlinkPackage(r.Context(), req)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error":   "invalid_forwarder_package_unlink",
					"message": err.Error(),
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"event": event,
				"mode":  "forwarder_package_reconciliation_unlink",
			})
		default:
			http.Error(w, "{\"error\":\"method_not_allowed\"}", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/forwarding/package-match-suggestions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, "{\"error\":\"method_not_allowed\"}", http.StatusMethodNotAllowed)
			return
		}
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil {
			http.Error(w, "{\"error\":\"active_profile_not_set\"}", http.StatusBadRequest)
			return
		}
		packageID := strings.TrimSpace(r.URL.Query().Get("package_id"))
		confidenceLabel, err := forwarding.NormalizePackageMatchConfidenceFilter(r.URL.Query().Get("confidence_label"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error":   "invalid_confidence_label",
				"message": err.Error(),
			})
			return
		}
		suggestions, err := forwarderInbox.SuggestPackageMatches(r.Context(), strings.TrimSpace(active.ID), packageID)
		if err != nil {
			http.Error(w, "{\"error\":\"failed_to_suggest_forwarder_package_matches\"}", http.StatusInternalServerError)
			return
		}
		suggestions = forwarding.FilterPackageMatchSuggestionsByConfidence(suggestions, confidenceLabel)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"mode":              "forwarder_package_match_suggestions",
			"mutable":           false,
			"confidence_filter": confidenceLabel,
			"suggestions":       suggestions,
			"summary":           forwarding.SummarizePackageMatchSuggestions(packageID, suggestions),
		})
	})
	mux.HandleFunc("/api/forwarding/packages/import-csv", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, "{\"error\":\"method_not_allowed\"}", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ProfileID string `json:"profile_id"`
			Provider  string `json:"provider"`
			CSV       string `json:"csv"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "{\"error\":\"invalid_json\"}", http.StatusBadRequest)
			return
		}
		imports, rowErrors, err := forwarding.ParsePackageCSV(req.ProfileID, req.Provider, strings.NewReader(req.CSV))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":   "invalid_forwarder_package_csv",
				"message": err.Error(),
			})
			return
		}
		imported := make([]forwarding.Package, 0, len(imports))
		for _, item := range imports {
			pkg, err := forwarderInbox.UpsertPackage(r.Context(), item)
			if err != nil {
				rowErrors = append(rowErrors, forwarding.PackageCSVRowError{Error: err.Error()})
				continue
			}
			imported = append(imported, pkg)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"mode":     "forwarder_package_csv_import",
			"imported": imported,
			"errors":   rowErrors,
			"summary": map[string]int{
				"imported": len(imported),
				"errors":   len(rowErrors),
			},
		})
	})
	mux.HandleFunc("/api/forwarding/packages/import-email", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, "{\"error\":\"method_not_allowed\"}", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ProfileID string `json:"profile_id"`
			Provider  string `json:"provider"`
			MessageID string `json:"message_id"`
			Body      string `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "{\"error\":\"invalid_json\"}", http.StatusBadRequest)
			return
		}
		imported, err := forwarding.ParsePackageEmail(req.ProfileID, req.Provider, req.MessageID, strings.NewReader(req.Body))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":   "invalid_forwarder_package_email",
				"message": err.Error(),
			})
			return
		}
		pkg, err := forwarderInbox.UpsertPackage(r.Context(), imported)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":   "invalid_forwarder_package_email",
				"message": err.Error(),
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"mode":    "forwarder_package_email_import",
			"package": pkg,
		})
	})
	mux.HandleFunc("/api/integrations/ebay/purchase-inbox/reviews", func(w http.ResponseWriter, r *http.Request) {
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
		var req struct {
			Cards []ebaypurchasecapture.PurchaseCard `json:"cards"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		reviews := ebaypurchasecapture.BuildPurchaseInboxReviews(req.Cards)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"profile_id": strings.TrimSpace(active.ID),
			"source":     "ebay_purchase_capture",
			"reviews":    reviews,
		})
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
	mux.HandleFunc("/api/integrations/ebay/purchase-inbox/actions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil || strings.TrimSpace(active.ID) == "" {
			http.Error(w, `{"error":"active_profile_not_set"}`, http.StatusBadRequest)
			return
		}
		profileID := strings.TrimSpace(active.ID)
		var req struct {
			ActionID       string                           `json:"action_id"`
			TargetKey      string                           `json:"target_key"`
			Confirmed      bool                             `json:"confirmed"`
			ExistingItemID string                           `json:"existing_item_id"`
			Card           ebaypurchasecapture.PurchaseCard `json:"card"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		actionID := strings.TrimSpace(req.ActionID)
		if !req.Confirmed {
			http.Error(w, `{"error":"purchase_inbox_action_requires_confirmation"}`, http.StatusConflict)
			return
		}
		if actionID != "link_existing_inventory_item" && actionID != "convert_to_inventory_item" {
			http.Error(w, `{"error":"unsupported_purchase_inbox_action"}`, http.StatusBadRequest)
			return
		}
		targetKey := ebaypurchasecapture.PurchaseItemKey(req.Card)
		if targetKey == "" || strings.TrimSpace(req.TargetKey) != targetKey {
			http.Error(w, `{"error":"purchase_inbox_target_mismatch"}`, http.StatusBadRequest)
			return
		}
		itemReview := ebaypurchasecapture.BuildPurchaseInboxReviews([]ebaypurchasecapture.PurchaseCard{req.Card})
		if len(itemReview) == 0 || len(itemReview[0].Items) == 0 || itemReview[0].Items[0].Status != "ready_to_link_or_convert" {
			http.Error(w, `{"error":"purchase_inbox_item_not_ready"}`, http.StatusBadRequest)
			return
		}

		switch actionID {
		case "link_existing_inventory_item":
			existingItemID := strings.TrimSpace(req.ExistingItemID)
			if !itemOwnedByProfile(r.Context(), profileID, existingItemID) {
				http.Error(w, `{"error":"purchase_inbox_existing_item_not_found"}`, http.StatusNotFound)
				return
			}
			existing, err := collectionRepo.GetItemByID(r.Context(), existingItemID)
			if err != nil {
				http.Error(w, `{"error":"purchase_inbox_existing_item_not_found"}`, http.StatusNotFound)
				return
			}
			existing.Notes = appendPurchaseInboxNote(existing.Notes, req.Card)
			existing.SourceURLs = appendPurchaseInboxSourceURL(existing.SourceURLs, req.Card.ItemURL)
			updated, err := collectionRepo.UpdateItem(r.Context(), existingItemID, existing)
			if err != nil {
				http.Error(w, `{"error":"failed_to_link_purchase_inbox_item"}`, http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":          true,
				"action_id":   actionID,
				"target_key":  targetKey,
				"profile_id":  profileID,
				"linked_item": updated,
				"audit":       purchaseInboxActionAudit(req.Card, existingItemID),
			})
		case "convert_to_inventory_item":
			created, err := collectionRepo.CreateItemForProfile(r.Context(), profileID, purchaseInboxCardToItem(req.Card))
			if err != nil {
				http.Error(w, `{"error":"failed_to_convert_purchase_inbox_item"}`, http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":           true,
				"action_id":    actionID,
				"target_key":   targetKey,
				"profile_id":   profileID,
				"created_item": created,
				"audit":        purchaseInboxActionAudit(req.Card, created.ID),
			})
		}
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
		profileID := ""
		if active, activeErr := profiles.GetActiveProfile(r.Context()); activeErr == nil {
			profileID = strings.TrimSpace(active.ID)
		}
		sum, err := dashboardSvc.Summary(r.Context(), profileID)
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
	mux.HandleFunc("/api/ai/suggestions/apply", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ProfileID    string `json:"profile_id"`
			SuggestionID string `json:"suggestion_id"`
			DraftID      string `json:"draft_id"`
			ConfirmToken string `json:"confirm_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.ProfileID) == "" || strings.TrimSpace(req.SuggestionID) == "" || strings.TrimSpace(req.DraftID) == "" {
			http.Error(w, `{"error":"invalid_apply_request"}`, http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.ConfirmToken) == "" {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":         "ai_confirm_required",
				"error_code":    "AI_CONFIRM_REQUIRED",
				"suggestion_id": strings.TrimSpace(req.SuggestionID),
				"draft_id":      strings.TrimSpace(req.DraftID),
				"next_action":   "confirm_before_apply",
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":            true,
			"applied":       true,
			"suggestion_id": strings.TrimSpace(req.SuggestionID),
			"draft_id":      strings.TrimSpace(req.DraftID),
		})
	})
	mux.HandleFunc("/api/telegram/catalog-captures", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req telegramCatalogCaptureRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.ProfileID) == "" {
			http.Error(w, `{"error":"profile_id_required"}`, http.StatusBadRequest)
			return
		}
		media, err := telegramCatalogCaptureMedia(req.Media)
		if err != nil {
			http.Error(w, `{"error":"invalid_telegram_media"}`, http.StatusBadRequest)
			return
		}
		svc := telegramcapture.NewService(profileSettingsTelegramAuthorizer{profiles: profiles, profileID: req.ProfileID}, chatSvc)
		result, err := svc.IngestCatalogCapture(r.Context(), telegramcapture.CaptureInput{
			SenderID:       req.SenderID,
			ChatID:         req.ChatID,
			MessageID:      req.MessageID,
			Text:           req.Text,
			Barcode:        req.Barcode,
			GroupingHint:   req.GroupingHint,
			SourceMetadata: req.SourceMetadata,
			Draft: telegramcapture.Draft{
				PartNumber:       req.Draft.PartNumber,
				Title:            req.Draft.Title,
				Brand:            req.Draft.Brand,
				Category:         req.Draft.Category,
				LookupSource:     req.Draft.LookupSource,
				LookupURL:        req.Draft.LookupURL,
				LookupConfidence: req.Draft.LookupConfidence,
			},
			Media: media,
		})
		if err != nil {
			if errors.Is(err, telegramcapture.ErrUnauthorizedSender) {
				http.Error(w, `{"error":"telegram_sender_not_authorized"}`, http.StatusForbidden)
				return
			}
			var followUp telegramcapture.DraftNeedsFollowUpError
			if errors.As(err, &followUp) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error":          "telegram_capture_needs_follow_up",
					"reason":         followUp.Reason,
					"missing_fields": followUp.MissingFields,
					"telegram_reply": followUp.Reply,
				})
				return
			}
			http.Error(w, `{"error":"failed_to_ingest_telegram_capture"}`, http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(result)
	})
	mux.HandleFunc("/api/telegram/webhook/catalog-captures", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var update telegramcapture.WebhookUpdate
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		input, err := telegramcapture.CaptureInputFromWebhookUpdate(update, telegramcapture.Draft{})
		if err != nil {
			http.Error(w, `{"error":"invalid_telegram_webhook_update"}`, http.StatusBadRequest)
			return
		}
		authorizer := allProfilesTelegramAuthorizer{profiles: profiles}
		authorized, err := authorizer.AuthorizeTelegramCapture(r.Context(), input.SenderID, input.ChatID)
		if err != nil {
			if errors.Is(err, telegramcapture.ErrUnauthorizedSender) {
				http.Error(w, `{"error":"telegram_sender_not_authorized"}`, http.StatusForbidden)
				return
			}
			http.Error(w, `{"error":"failed_to_authorize_telegram_sender"}`, http.StatusBadRequest)
			return
		}
		if draft, ok, err := telegramCatalogCaptureLocalBarcodeDraft(r.Context(), conn, authorized.ProfileID, input.Barcode); err != nil {
			http.Error(w, `{"error":"failed_to_lookup_telegram_barcode"}`, http.StatusBadRequest)
			return
		} else if ok {
			input.Draft = draft
		}
		svc := telegramcapture.NewService(authorizer, chatSvc)
		result, err := svc.IngestCatalogCapture(r.Context(), input)
		if err != nil {
			if errors.Is(err, telegramcapture.ErrUnauthorizedSender) {
				http.Error(w, `{"error":"telegram_sender_not_authorized"}`, http.StatusForbidden)
				return
			}
			var followUp telegramcapture.DraftNeedsFollowUpError
			if errors.As(err, &followUp) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error":          "telegram_capture_needs_follow_up",
					"reason":         followUp.Reason,
					"missing_fields": followUp.MissingFields,
					"telegram_reply": followUp.Reply,
				})
				return
			}
			http.Error(w, `{"error":"failed_to_ingest_telegram_capture"}`, http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(result)
	})
	mux.HandleFunc("/api/telegram/catalog-capture-callbacks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req telegramCatalogCaptureCallbackRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		svc := telegramcapture.NewService(allProfilesTelegramAuthorizer{profiles: profiles}, chatSvc)
		result, err := svc.HandleCatalogCaptureCallback(r.Context(), telegramcapture.CallbackInput{
			SenderID:     req.SenderID,
			ChatID:       req.ChatID,
			CallbackData: req.CallbackData,
		})
		if err != nil {
			if errors.Is(err, telegramcapture.ErrUnauthorizedSender) {
				http.Error(w, `{"error":"telegram_sender_not_authorized"}`, http.StatusForbidden)
				return
			}
			http.Error(w, `{"error":"failed_to_handle_telegram_capture_callback"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(result)
	})
	mux.HandleFunc("/api/telegram/external-intake-proofs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req telegramExternalIntakeProofRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		if missing := missingTelegramExternalIntakeProofFields(req); len(missing) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":          "invalid_external_intake_proof",
				"missing_fields": missing,
				"next_action":    "provide_authorized_runtime_source_and_non_secret_openai_provider_evidence",
			})
			return
		}
		authorizer := profileSettingsTelegramAuthorizer{profiles: profiles, profileID: req.ProfileID}
		if _, err := authorizer.AuthorizeTelegramCapture(r.Context(), req.SenderID, req.ChatID); err != nil {
			if errors.Is(err, telegramcapture.ErrUnauthorizedSender) {
				http.Error(w, `{"error":"telegram_sender_not_authorized"}`, http.StatusForbidden)
				return
			}
			http.Error(w, `{"error":"failed_to_authorize_telegram_sender"}`, http.StatusBadRequest)
			return
		}
		thread, err := chatSvc.GetThread(r.Context(), req.ProfileID, req.SourceThreadID)
		if err != nil {
			http.Error(w, `{"error":"source_thread_not_found"}`, http.StatusBadRequest)
			return
		}
		preview, err := chatSvc.GetActionPreview(r.Context(), req.ProfileID, req.PreviewID)
		if err != nil || preview.ThreadID != thread.ID {
			http.Error(w, `{"error":"preview_not_found_for_source_thread"}`, http.StatusBadRequest)
			return
		}
		run, err := chatSvc.CreateWorkflowRun(r.Context(), chat.CreateWorkflowRunInput{
			ProfileID:         req.ProfileID,
			WorkflowID:        "telegram-openai-external-intake-proof",
			CapabilityID:      strings.TrimSpace(req.CapabilityID),
			SourceChannel:     "telegram",
			SourceThreadID:    thread.ID,
			SourceMessageID:   strings.TrimSpace(req.SourceMessageID),
			ConfirmationState: req.ConfirmationState,
			Input: map[string]any{
				"sender_id":    strings.TrimSpace(req.SenderID),
				"chat_id":      strings.TrimSpace(req.ChatID),
				"preview_id":   strings.TrimSpace(req.PreviewID),
				"proof_source": "approved_runtime_packet",
			},
			ProviderTrace: nonSecretProviderTrace(req.ProviderTrace),
		})
		if err != nil {
			http.Error(w, `{"error":"failed_to_create_external_intake_proof"}`, http.StatusBadRequest)
			return
		}
		run, err = chatSvc.UpdateWorkflowRun(r.Context(), chat.UpdateWorkflowRunInput{
			ProfileID:         req.ProfileID,
			RunID:             run.ID,
			Status:            "completed",
			ConfirmationState: req.ConfirmationState,
			ProviderTrace:     run.ProviderTrace,
			Result: map[string]any{
				"proof_packet":       "authorized_telegram_openai_external_intake",
				"preview_id":         strings.TrimSpace(req.PreviewID),
				"thread_id":          thread.ID,
				"source_message_id":  strings.TrimSpace(req.SourceMessageID),
				"confirmation_state": strings.TrimSpace(req.ConfirmationState),
			},
		})
		if err != nil {
			http.Error(w, `{"error":"failed_to_complete_external_intake_proof"}`, http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":           true,
			"workflow_run": run,
		})
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
				ProfileID string         `json:"profile_id"`
				Title     string         `json:"title"`
				Metadata  map[string]any `json:"metadata"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			thread, err := chatSvc.CreateThread(r.Context(), req.ProfileID, req.Title, req.Metadata)
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
	mux.HandleFunc("/api/chat/capabilities", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		profileID := strings.TrimSpace(r.URL.Query().Get("profile_id"))
		if profileID == "" {
			http.Error(w, `{"error":"profile_id_required"}`, http.StatusBadRequest)
			return
		}
		route := strings.TrimSpace(r.URL.Query().Get("route"))
		if route == "" {
			route = "/"
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"profile_id":       profileID,
			"route":            route,
			"capabilities":     assistantCapabilityRegistry(),
			"guided_workflows": chat.GuidedWorkflowRegistry(),
			"policy":           "preview-before-apply",
			"confirm_apply":    true,
		})
	})
	mux.HandleFunc("/api/chat/workflow-runs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			profileID := strings.TrimSpace(r.URL.Query().Get("profile_id"))
			if profileID == "" {
				http.Error(w, `{"error":"profile_id_required"}`, http.StatusBadRequest)
				return
			}
			runs, err := chatSvc.ListWorkflowRuns(r.Context(), profileID, strings.TrimSpace(r.URL.Query().Get("thread_id")))
			if err != nil {
				http.Error(w, `{"error":"failed_to_list_workflow_runs"}`, http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"runs": runs})
		case http.MethodPost:
			var req chat.CreateWorkflowRunInput
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			run, err := chatSvc.CreateWorkflowRun(r.Context(), req)
			if err != nil {
				http.Error(w, `{"error":"failed_to_create_workflow_run"}`, http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(run)
		default:
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/chat/workflow-runs/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		runID := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/chat/workflow-runs/"))
		if runID == "" {
			http.Error(w, `{"error":"run_id_required"}`, http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			profileID := strings.TrimSpace(r.URL.Query().Get("profile_id"))
			if profileID == "" {
				http.Error(w, `{"error":"profile_id_required"}`, http.StatusBadRequest)
				return
			}
			run, err := chatSvc.GetWorkflowRun(r.Context(), profileID, runID)
			if err != nil {
				http.Error(w, `{"error":"workflow_run_not_found"}`, http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(run)
		case http.MethodPatch:
			var req chat.UpdateWorkflowRunInput
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			req.RunID = runID
			run, err := chatSvc.UpdateWorkflowRun(r.Context(), req)
			if err != nil {
				http.Error(w, `{"error":"failed_to_update_workflow_run"}`, http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(run)
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
				ProfileID string         `json:"profile_id"`
				ThreadID  string         `json:"thread_id"`
				Role      string         `json:"role"`
				Content   string         `json:"content"`
				Context   map[string]any `json:"context"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			message, err := chatSvc.CreateMessage(r.Context(), req.ProfileID, req.ThreadID, req.Role, req.Content, req.Context)
			if err != nil {
				http.Error(w, `{"error":"failed_to_create_chat_message"}`, http.StatusBadRequest)
				return
			}
			response := map[string]any{"message": message}
			if strings.EqualFold(strings.TrimSpace(req.Role), "user") {
				if assistantContext, ok := req.Context["assistant"].(map[string]any); ok && len(assistantContext) > 0 {
					if appControl, handled := dispatchChatMessageAppControl(r.Context(), chatSvc, req.ProfileID, req.ThreadID, req.Content, req.Context, message.ID); handled {
						response["app_control"] = appControl
					} else if chatMessageRequiresAssistantHandoff(req.Content) {
						inboxItem, inboxErr := chatSvc.CreateInboxItem(r.Context(), chat.InboxItem{
							ProfileID: req.ProfileID,
							ThreadID:  req.ThreadID,
							Source:    "assistant_handoff",
							Status:    "queued",
							Title:     "Assistant handoff queued",
							Summary:   strings.TrimSpace(req.Content),
							Metadata: map[string]any{
								"assistant": assistantContext,
								"route":     req.Context["route"],
								"selection": req.Context["selection"],
							},
						})
						if inboxErr == nil {
							assistantMessage, assistantErr := chatSvc.CreateMessage(r.Context(), req.ProfileID, req.ThreadID, "assistant", "Assistant handoff queued in Inbox.", map[string]any{
								"assistant_handoff": map[string]any{
									"status":        "queued",
									"inbox_item_id": inboxItem.ID,
								},
							})
							if assistantErr == nil {
								response["assistant_handoff"] = map[string]any{"thread_message": assistantMessage, "inbox_item": inboxItem}
							}
						}
					} else {
						assistantMessage, assistantErr := chatSvc.CreateMessage(r.Context(), req.ProfileID, req.ThreadID, "assistant", directAssistantChatResponse(req.Content), map[string]any{
							"assistant_response": map[string]any{
								"mode":          "direct",
								"source":        "deterministic_chat_fallback",
								"source_msg_id": message.ID,
							},
						})
						if assistantErr == nil {
							response["assistant_response"] = map[string]any{"mode": "direct", "thread_message": assistantMessage}
						}
					}
				}
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(response)
		default:
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/chat/inbox", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			var req struct {
				ProfileID string                          `json:"profile_id"`
				Records   []chat.NotificationHistoryInput `json:"records"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
				return
			}
			var items []chat.InboxItem
			for _, record := range req.Records {
				record.ProfileID = req.ProfileID
				item, err := chatSvc.CreateNotificationHistoryItem(r.Context(), record)
				if err != nil {
					http.Error(w, `{"error":"failed_to_record_notification_history"}`, http.StatusBadRequest)
					return
				}
				items = append(items, item)
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		profileID := strings.TrimSpace(r.URL.Query().Get("profile_id"))
		if profileID == "" {
			http.Error(w, `{"error":"profile_id_required"}`, http.StatusBadRequest)
			return
		}
		items, err := chatSvc.ListInboxItems(r.Context(), profileID)
		if err != nil {
			http.Error(w, `{"error":"failed_to_list_chat_inbox_items"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"items": items})
	})
	mux.HandleFunc("/api/chat/inbox/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPatch {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		inboxID := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/chat/inbox/"))
		if inboxID == "" {
			http.Error(w, `{"error":"inbox_id_required"}`, http.StatusBadRequest)
			return
		}
		var req struct {
			ProfileID string `json:"profile_id"`
			Status    string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		item, err := chatSvc.UpdateInboxItemStatus(r.Context(), req.ProfileID, inboxID, req.Status)
		if err != nil {
			http.Error(w, `{"error":"failed_to_update_chat_inbox_item"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(item)
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
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed_to_preview_chat_action", "detail": err.Error()})
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
	mux.HandleFunc("/api/chat/actions/cancel", func(w http.ResponseWriter, r *http.Request) {
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
		result, err := chatSvc.CancelAction(r.Context(), req)
		if err != nil {
			http.Error(w, `{"error":"failed_to_cancel_chat_action"}`, http.StatusBadRequest)
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
		w.Header().Set("Content-Disposition", `attachment; filename="cabinet-data-snapshot.json"`)
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
		w.Header().Set("Content-Disposition", `attachment; filename="cabinet-items.csv"`)
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
		sum, err := dataService.ApplyImport(r.Context(), req.Snapshot, req.Options)
		if err != nil {
			http.Error(w, `{"error":"failed_to_apply_import"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(sum)
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
		sum, err := dataService.ApplyImport(r.Context(), snap, req.Options)
		if err != nil {
			http.Error(w, `{"error":"failed_to_apply_import"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(sum)
	})
	mux.HandleFunc("/api/data/reindex", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		result, err := dataService.Reindex(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed_to_reindex"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(result)
	})
	mux.HandleFunc("/api/data/rebuild-thumbnails", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		rebuiltItems, rebuiltPhotos, err := mediaService.RebuildAllThumbnails(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed_to_rebuild_thumbnails"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":             true,
			"rebuilt_items":  rebuiltItems,
			"rebuilt_photos": rebuiltPhotos,
		})
	})
	mux.HandleFunc("/api/media/assets", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			active, err := profiles.GetActiveProfile(r.Context())
			if err != nil || strings.TrimSpace(active.ID) == "" {
				http.Error(w, `{"error":"active_profile_required"}`, http.StatusBadRequest)
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
			file, hdr, err := r.FormFile("file")
			if err != nil {
				http.Error(w, `{"error":"file_required"}`, http.StatusBadRequest)
				return
			}
			defer file.Close()
			mimeType := strings.TrimSpace(hdr.Header.Get("Content-Type"))
			if !isSupportedMediaUpload(mimeType, hdr.Filename) {
				http.Error(w, `{"error":"unsupported_media_type"}`, http.StatusUnsupportedMediaType)
				return
			}
			if mimeType == "" || strings.EqualFold(mimeType, "application/octet-stream") {
				mimeType = mediaUploadContentType(hdr.Filename)
			}
			threadID := ""
			threads, err := chatSvc.ListThreads(r.Context(), active.ID)
			if err != nil {
				http.Error(w, `{"error":"failed_to_list_media_upload_threads"}`, http.StatusBadRequest)
				return
			}
			for _, thread := range threads {
				if strings.EqualFold(strings.TrimSpace(thread.Title), "Media Uploads") {
					threadID = thread.ID
					break
				}
			}
			if threadID == "" {
				thread, err := chatSvc.CreateThread(r.Context(), active.ID, "Media Uploads", map[string]any{
					"source": "media.workspace",
				})
				if err != nil {
					http.Error(w, `{"error":"failed_to_create_media_upload_thread"}`, http.StatusBadRequest)
					return
				}
				threadID = thread.ID
			}
			attachment, err := chatSvc.SaveAttachment(r.Context(), active.ID, threadID, hdr.Filename, mimeType, file)
			if err != nil {
				http.Error(w, `{"error":"failed_to_save_media_asset"}`, http.StatusBadRequest)
				return
			}
			title := strings.TrimSpace(r.FormValue("title"))
			source := strings.TrimSpace(r.FormValue("source"))
			notes := strings.TrimSpace(r.FormValue("notes"))
			if title == "" {
				title = attachment.Filename
			}
			if _, err := chatSvc.CreateMessage(r.Context(), active.ID, threadID, "user", "Media asset added from Media workspace.", map[string]any{
				"source":        "media.workspace",
				"asset_id":      attachment.ID,
				"title":         title,
				"origin":        source,
				"notes":         notes,
				"filename":      attachment.Filename,
				"mime_type":     attachment.MimeType,
				"metadata_flow": "add-media-dialog",
			}); err != nil {
				http.Error(w, `{"error":"failed_to_save_media_metadata"}`, http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"asset_id":  attachment.ID,
				"filename":  attachment.Filename,
				"mime_type": attachment.MimeType,
				"title":     title,
				"source":    source,
				"notes":     notes,
			})
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil || strings.TrimSpace(active.ID) == "" {
			http.Error(w, `{"error":"active_profile_required"}`, http.StatusBadRequest)
			return
		}
		list, err := mediaService.ListWorkspaceAssets(r.Context(), active.ID, r.URL.Query().Get("filter"))
		if err != nil {
			http.Error(w, `{"error":"media_assets_unavailable"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(list)
	})
	mux.HandleFunc("/api/media/assets/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPatch {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		if !strings.HasSuffix(r.URL.Path, "/metadata") {
			http.Error(w, `{"error":"media_asset_route_not_found"}`, http.StatusNotFound)
			return
		}
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil || strings.TrimSpace(active.ID) == "" {
			http.Error(w, `{"error":"active_profile_required"}`, http.StatusBadRequest)
			return
		}
		assetID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/media/assets/"), "/metadata")
		assetID, err = url.PathUnescape(strings.Trim(assetID, "/"))
		if err != nil || strings.TrimSpace(assetID) == "" {
			http.Error(w, `{"error":"media_asset_id_required"}`, http.StatusBadRequest)
			return
		}
		var req media.WorkspaceAssetMetadataUpdate
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_media_metadata_update"}`, http.StatusBadRequest)
			return
		}
		asset, err := mediaService.UpdateWorkspaceAssetMetadata(r.Context(), active.ID, assetID, req)
		if err != nil {
			http.Error(w, `{"error":"media_metadata_update_failed"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(asset)
	})
	mux.HandleFunc("/api/media/assignments/preview", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil || strings.TrimSpace(active.ID) == "" {
			http.Error(w, `{"error":"active_profile_required"}`, http.StatusBadRequest)
			return
		}
		var req struct {
			AssetID    string `json:"asset_id"`
			TargetType string `json:"target_type"`
			TargetID   string `json:"target_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_media_assignment_preview"}`, http.StatusBadRequest)
			return
		}
		preview, err := mediaService.PreviewAssignment(r.Context(), active.ID, req.AssetID, req.TargetType, req.TargetID)
		if err != nil {
			http.Error(w, `{"error":"media_assignment_preview_unavailable"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(preview)
	})
	mux.HandleFunc("/api/media/assignments", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil || strings.TrimSpace(active.ID) == "" {
			http.Error(w, `{"error":"active_profile_required"}`, http.StatusBadRequest)
			return
		}
		var req struct {
			AssetID    string `json:"asset_id"`
			TargetType string `json:"target_type"`
			TargetID   string `json:"target_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_media_assignment"}`, http.StatusBadRequest)
			return
		}
		result, err := mediaService.ApplyAssignment(r.Context(), active.ID, req.AssetID, req.TargetType, req.TargetID)
		if err != nil {
			http.Error(w, `{"error":"media_assignment_unavailable"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(result)
	})
	mux.HandleFunc("/api/media/downloads/preview", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil || strings.TrimSpace(active.ID) == "" {
			http.Error(w, `{"error":"active_profile_required"}`, http.StatusBadRequest)
			return
		}
		var req struct {
			AssetIDs []string `json:"asset_ids"`
			Filter   string   `json:"filter"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_media_download_preview"}`, http.StatusBadRequest)
			return
		}
		preview, err := mediaService.PreviewDownload(r.Context(), active.ID, req.AssetIDs, req.Filter)
		if err != nil {
			http.Error(w, `{"error":"media_download_preview_unavailable"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(preview)
	})
	mux.HandleFunc("/api/media/downloads", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		active, err := profiles.GetActiveProfile(r.Context())
		if err != nil || strings.TrimSpace(active.ID) == "" {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"active_profile_required"}`, http.StatusBadRequest)
			return
		}
		var req struct {
			AssetIDs []string `json:"asset_ids"`
			Filter   string   `json:"filter"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"invalid_media_download"}`, http.StatusBadRequest)
			return
		}
		bundle, err := mediaService.BuildDownload(r.Context(), active.ID, req.AssetIDs, req.Filter)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"media_download_unavailable"}`, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", bundle.ContentType)
		w.Header().Set("Content-Disposition", contentDispositionAttachment(bundle.Filename))
		w.Header().Set("X-Cabinet-Media-Asset-Count", strconv.Itoa(len(bundle.AssetIDs)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(bundle.Bytes)
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
		_ = json.NewEncoder(w).Encode(result)
	})
	mux.HandleFunc("/api/backup/run", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		result, err := backupSvc.CreateBackup(r.Context())
		if err != nil {
			http.Error(w, `{"error":"failed_to_create_backup"}`, http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"backup": result})
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
	mux.HandleFunc("/api/backup/download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		backupPath, err := backupSvc.ResolveBackupPath(r.URL.Query().Get("file_name"))
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"backup_not_found"}`, http.StatusNotFound)
			return
		}
		if strings.EqualFold(filepath.Ext(backupPath), ".zip") {
			w.Header().Set("Content-Type", "application/zip")
		} else {
			w.Header().Set("Content-Type", "application/octet-stream")
		}
		w.Header().Set("Content-Disposition", `attachment; filename="`+filepath.Base(backupPath)+`"`)
		http.ServeFile(w, r, backupPath)
	})
	mux.HandleFunc("/api/backup/restore", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			BackupPath     string `json:"backup_path"`
			ConfirmRestore bool   `json:"confirm_restore"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
			return
		}
		if !req.ConfirmRestore {
			http.Error(w, `{"error":"restore_confirmation_required","recovery":"set confirm_restore to true after reviewing the selected backup"}`, http.StatusBadRequest)
			return
		}
		result, err := backupSvc.RestoreBackup(req.BackupPath)
		if err != nil {
			http.Error(w, `{"error":"failed_to_restore_backup"}`, http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"restore": result})
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
			settings := map[string]string{}
			if active, activeErr := profiles.GetActiveProfile(r.Context()); activeErr == nil && strings.TrimSpace(active.ID) != "" {
				settings, _ = profiles.GetSettings(r.Context(), active.ID)
			}
			if validationErr := validateInventoryItemTaxonomy(req.Changes, nil, settings); validationErr != nil {
				writeInventoryTaxonomyValidationError(w, validationErr)
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
			settings := map[string]string{}
			if active, activeErr := profiles.GetActiveProfile(r.Context()); activeErr == nil && strings.TrimSpace(active.ID) != "" {
				settings, _ = profiles.GetSettings(r.Context(), active.ID)
			}
			current, _ := collectionRepo.GetItemByID(r.Context(), itemID)
			if validationErr := validateInventoryItemTaxonomy(req, &current, settings); validationErr != nil {
				writeInventoryTaxonomyValidationError(w, validationErr)
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
			if len(parts) == 4 && parts[3] == "rotate" {
				photoID := strings.TrimSpace(parts[2])
				if photoID == "" {
					http.Error(w, `{"error":"invalid_photo_id"}`, http.StatusBadRequest)
					return
				}
				if r.Method != http.MethodPut {
					http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
					return
				}
				var req struct {
					Direction string `json:"direction"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
					return
				}
				rotated, err := mediaService.Rotate(r.Context(), itemID, photoID, req.Direction)
				if err != nil {
					http.Error(w, `{"error":"failed_to_rotate_photo"}`, http.StatusBadRequest)
					return
				}
				_ = json.NewEncoder(w).Encode(rotated)
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
		result, err := seedOnboardingSampleData(r.Context(), profiles, collectionRepo, wishlistSvc, mediaService, conn)
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
		role := strings.TrimSpace(strings.ToLower(claimAsString(claims, "role")))
		if role == "" {
			role = strings.TrimSpace(strings.ToLower(claimAsString(claims, "cabinet_role")))
		}
		if role == "" {
			role = "member"
		}
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
		_ = persistCloudSessionContext(r.Context(), conn, userID, email, role)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"provider":           "clerk",
			"user_id":            userID,
			"email":              email,
			"role":               role,
			"plan":               plan,
			"features":           features,
			"entitlement_source": entitlementSource,
		})
	})
	mux.HandleFunc("/api/auth/cloud/session/effective", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		plan, _ := currentCloudPlanFromState(r.Context(), conn)
		if strings.TrimSpace(plan) == "" {
			plan = "free"
		}
		userID := ""
		email := ""
		role := ""
		if conn != nil {
			userID, _ = appStateValue(r.Context(), conn, "cloud.user_id")
			email, _ = appStateValue(r.Context(), conn, "cloud.email")
			role, _ = appStateValue(r.Context(), conn, "cloud.role")
		}
		if strings.TrimSpace(role) == "" {
			role = "member"
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"provider": "clerk",
			"user_id":  strings.TrimSpace(userID),
			"email":    strings.TrimSpace(email),
			"role":     strings.TrimSpace(strings.ToLower(role)),
			"plan":     strings.TrimSpace(strings.ToLower(plan)),
			"features": entitlementFeaturesFromPlan(plan),
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

	runtimeLogs, err := newRuntimeLogManager(cfg)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("init runtime logs: %w", err)
	}

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           runtimeLogs.wrapHandler(protectedMux),
		ReadHeaderTimeout: 10 * time.Second,
		ErrorLog:          log.New(runtimeLogs.errorWriter(), "", 0),
	}

	a := &App{
		cfg:           cfg,
		db:            conn,
		srv:           srv,
		backupSvc:     backupSvc,
		authService:   authService,
		openapiSpec:   openapiSpec,
		runtimeLogs:   runtimeLogs,
		runtimeStopCh: runtimeStopCh,
		startupNotice: func(line string) {
			log.Print(line)
		},
		startupIsTTY: isRuntimeTTY,
	}

	return a, nil
}

type runtimeSetupRequest struct {
	InstanceName         string `json:"instance_name"`
	ProfileKey           string `json:"profile_key"`
	StorageMode          string `json:"storage_mode"`
	StorageDataDir       string `json:"storage_data_dir"`
	PortableMode         bool   `json:"portable_mode"`
	AuthMode             string `json:"auth_mode"`
	ClerkPublishableKey  string `json:"clerk_publishable_key"`
	RuntimePortMode      string `json:"runtime_port_mode"`
	RuntimeFixedPort     int    `json:"runtime_fixed_port"`
	FeatureChat          *bool  `json:"feature_chat"`
	FeatureProviders     *bool  `json:"feature_providers"`
	FeatureScanner       *bool  `json:"feature_scanner"`
	BootstrapWorkspace   string `json:"bootstrap_workspace"`
	BootstrapDatabaseRef string `json:"bootstrap_database_ref"`
}

type runtimeSetupImportRequest struct {
	SourcePath string `json:"source_path"`
}

type runtimeSetupValidationError struct {
	Code    string
	Message string
	Field   string
}

type runtimeSetupConfigFile struct {
	Version   int                         `json:"version"`
	Instance  runtimeSetupInstanceConfig  `json:"instance"`
	Storage   runtimeSetupStorageConfig   `json:"storage"`
	Runtime   runtimeSetupRuntimeConfig   `json:"runtime"`
	Auth      runtimeSetupAuthConfig      `json:"auth"`
	Bootstrap runtimeSetupBootstrapConfig `json:"bootstrap"`
	Features  runtimeSetupFeaturesConfig  `json:"features"`
	Meta      runtimeSetupMetaConfig      `json:"meta"`
}

type runtimeSetupInstanceConfig struct {
	Name    string `json:"name"`
	Profile string `json:"profile"`
}

type runtimeSetupStorageConfig struct {
	DataDir      string `json:"dataDir"`
	MediaDir     string `json:"mediaDir"`
	PortableMode bool   `json:"portableMode"`
}

type runtimeSetupRuntimeConfig struct {
	PortMode    string `json:"portMode"`
	Port        *int   `json:"port"`
	ResolvedURL string `json:"resolvedUrl"`
}

type runtimeSetupAuthConfig struct {
	Mode  string                      `json:"mode"`
	Clerk runtimeSetupClerkAuthConfig `json:"clerk"`
}

type runtimeSetupClerkAuthConfig struct {
	PublishableKey string `json:"publishableKey"`
	Enabled        bool   `json:"enabled"`
}

type runtimeSetupBootstrapConfig struct {
	Workspace       string `json:"workspace"`
	DatabaseProfile string `json:"databaseProfile"`
}

type runtimeSetupFeaturesConfig struct {
	Chat      bool `json:"chat"`
	Providers bool `json:"providers"`
	Scanner   bool `json:"scanner"`
}

type runtimeSetupMetaConfig struct {
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
	WizardVersion      string `json:"wizardVersion"`
	CurrentURL         string `json:"currentUrl"`
	StartedAt          string `json:"startedAt,omitempty"`
	StartedBy          string `json:"startedBy,omitempty"`
	LaunchSource       string `json:"launchSource,omitempty"`
	LastKnownPID       int    `json:"lastKnownPid,omitempty"`
	LastKnownURL       string `json:"lastKnownUrl,omitempty"`
	LastHeartbeatAt    string `json:"lastHeartbeatAt,omitempty"`
	LastShutdownAt     string `json:"lastShutdownAt,omitempty"`
	LastShutdownReason string `json:"lastShutdownReason,omitempty"`
	LastRunClean       bool   `json:"lastRunClean"`
}

func runtimeSetupConfigPath(cfg config.Config) string {
	return filepath.Join(cfg.DataDir, "cabinet.json")
}

func runtimePIDPath(cfg config.Config) string {
	return filepath.Join(cfg.DataDir, "cabinet.pid")
}

func runtimeSetupRequired(cfg config.Config) bool {
	_, err := os.Stat(runtimeSetupConfigPath(cfg))
	return errors.Is(err, os.ErrNotExist)
}

func writeRuntimeSetupConfig(cfg config.Config, payload runtimeSetupConfigFile) error {
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(runtimeSetupConfigPath(cfg), data, 0o644)
}

func writeRuntimePIDFile(cfg config.Config, pid int) error {
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return err
	}
	content := []byte(strconv.Itoa(pid) + "\n")
	return os.WriteFile(runtimePIDPath(cfg), content, 0o644)
}

func removeRuntimePIDFile(cfg config.Config) error {
	err := os.Remove(runtimePIDPath(cfg))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func runtimeResolvedURLFromConfig(cfg config.Config) string {
	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		host = "127.0.0.1"
	}
	port := cfg.Port
	if port <= 0 {
		port = 17880
	}
	return fmt.Sprintf("http://%s:%d", host, port)
}

func syncRuntimeSetupCurrentURL(cfg config.Config) error {
	return syncRuntimeSetupCurrentURLWithURL(cfg, runtimeResolvedURLFromConfig(cfg))
}

func syncRuntimeSetupCurrentURLWithURL(cfg config.Config, resolvedURL string) error {
	configPath := runtimeSetupConfigPath(cfg)
	raw, err := os.ReadFile(configPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var payload runtimeSetupConfigFile
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}
	changed := false
	resolved := strings.TrimSpace(resolvedURL)
	if resolved == "" {
		resolved = strings.TrimSpace(payload.Runtime.ResolvedURL)
		if resolved == "" {
			resolved = runtimeResolvedURLFromConfig(cfg)
		}
	}
	if strings.TrimSpace(payload.Runtime.ResolvedURL) != resolved {
		payload.Runtime.ResolvedURL = resolved
		changed = true
	}
	if strings.TrimSpace(payload.Meta.CurrentURL) != resolved {
		payload.Meta.CurrentURL = resolved
		payload.Meta.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		changed = true
	}
	if !changed {
		return nil
	}
	return writeRuntimeSetupConfig(cfg, payload)
}

func portFromResolvedURL(raw string) int {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return 0
	}
	port, err := strconv.Atoi(parsed.Port())
	if err != nil {
		return 0
	}
	return port
}

func validateRuntimeSetupRequest(req runtimeSetupRequest) *runtimeSetupValidationError {
	if strings.TrimSpace(req.InstanceName) == "" {
		return &runtimeSetupValidationError{
			Code:    "SETUP_INSTANCE_NAME_REQUIRED",
			Message: "Instance name is required.",
			Field:   "instance_name",
		}
	}
	storageMode := strings.TrimSpace(strings.ToLower(req.StorageMode))
	if storageMode == "" {
		storageMode = "exe_local"
	}
	if storageMode != "exe_local" && storageMode != "custom" {
		return &runtimeSetupValidationError{
			Code:    "SETUP_STORAGE_MODE_INVALID",
			Message: "Storage mode must be exe_local or custom.",
			Field:   "storage_mode",
		}
	}
	if storageMode == "custom" && strings.TrimSpace(req.StorageDataDir) == "" {
		return &runtimeSetupValidationError{
			Code:    "SETUP_STORAGE_PATH_REQUIRED",
			Message: "Custom storage path is required.",
			Field:   "storage_data_dir",
		}
	}

	authMode := strings.TrimSpace(strings.ToLower(req.AuthMode))
	if authMode == "" {
		authMode = "local"
	}
	if authMode != "local" && authMode != "clerk" {
		return &runtimeSetupValidationError{
			Code:    "SETUP_AUTH_MODE_INVALID",
			Message: "Auth mode must be local or clerk.",
			Field:   "auth_mode",
		}
	}
	if authMode == "clerk" && strings.TrimSpace(req.ClerkPublishableKey) == "" {
		return &runtimeSetupValidationError{
			Code:    "SETUP_CLERK_PUBLISHABLE_KEY_REQUIRED",
			Message: "Clerk publishable key is required.",
			Field:   "clerk_publishable_key",
		}
	}
	portMode := strings.TrimSpace(strings.ToLower(req.RuntimePortMode))
	if portMode == "" {
		portMode = "auto"
	}
	if portMode != "auto" && portMode != "fixed" {
		return &runtimeSetupValidationError{
			Code:    "SETUP_RUNTIME_PORT_MODE_INVALID",
			Message: "Runtime port mode must be auto or fixed.",
			Field:   "runtime_port_mode",
		}
	}
	if portMode == "fixed" && req.RuntimeFixedPort <= 0 {
		return &runtimeSetupValidationError{
			Code:    "SETUP_RUNTIME_FIXED_PORT_REQUIRED",
			Message: "Fixed port value is required when runtime port mode is fixed.",
			Field:   "runtime_fixed_port",
		}
	}
	return nil
}

func validateRuntimeSetupImportRequest(req runtimeSetupImportRequest) *runtimeSetupValidationError {
	if strings.TrimSpace(req.SourcePath) == "" {
		return &runtimeSetupValidationError{
			Code:    "SETUP_IMPORT_SOURCE_PATH_REQUIRED",
			Message: "Source path is required.",
			Field:   "source_path",
		}
	}
	return nil
}

func buildRuntimeSetupConfig(cfg config.Config, req runtimeSetupRequest) (runtimeSetupConfigFile, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	profileKey := deriveProfileKey(req.ProfileKey, req.InstanceName)
	instanceName := strings.TrimSpace(req.InstanceName)
	authMode := strings.TrimSpace(strings.ToLower(req.AuthMode))
	if authMode == "" {
		authMode = "local"
	}
	portMode := strings.TrimSpace(strings.ToLower(req.RuntimePortMode))
	if portMode == "" {
		portMode = "auto"
	}
	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		host = "127.0.0.1"
	}
	port := cfg.Port
	if port <= 0 {
		port = 17880
	}
	var runtimePort *int
	if portMode == "fixed" {
		runtimePort = &req.RuntimeFixedPort
		port = req.RuntimeFixedPort
	}

	workspace := strings.TrimSpace(req.BootstrapWorkspace)
	if workspace == "" {
		workspace = "Local Workspace"
	}
	databaseRef := strings.TrimSpace(req.BootstrapDatabaseRef)
	if databaseRef == "" {
		databaseRef = "Primary DB"
	}

	storageDataDir := resolveRuntimeSetupStorageDataDir(cfg, req, profileKey)
	storageMediaDir := filepath.Join(storageDataDir, "media")
	featureChat := true
	if req.FeatureChat != nil {
		featureChat = *req.FeatureChat
	}
	featureProviders := true
	if req.FeatureProviders != nil {
		featureProviders = *req.FeatureProviders
	}
	featureScanner := true
	if req.FeatureScanner != nil {
		featureScanner = *req.FeatureScanner
	}
	return runtimeSetupConfigFile{
		Version: 1,
		Instance: runtimeSetupInstanceConfig{
			Name:    instanceName,
			Profile: profileKey,
		},
		Storage: runtimeSetupStorageConfig{
			DataDir:      storageDataDir,
			MediaDir:     storageMediaDir,
			PortableMode: req.PortableMode,
		},
		Runtime: runtimeSetupRuntimeConfig{
			PortMode:    portMode,
			Port:        runtimePort,
			ResolvedURL: fmt.Sprintf("http://%s:%d", host, port),
		},
		Auth: runtimeSetupAuthConfig{
			Mode: authMode,
			Clerk: runtimeSetupClerkAuthConfig{
				PublishableKey: strings.TrimSpace(req.ClerkPublishableKey),
				Enabled:        authMode == "clerk",
			},
		},
		Bootstrap: runtimeSetupBootstrapConfig{
			Workspace:       workspace,
			DatabaseProfile: databaseRef,
		},
		Features: runtimeSetupFeaturesConfig{
			Chat:      featureChat,
			Providers: featureProviders,
			Scanner:   featureScanner,
		},
		Meta: runtimeSetupMetaConfig{
			CreatedAt:     now,
			UpdatedAt:     now,
			WizardVersion: "1",
			CurrentURL:    fmt.Sprintf("http://%s:%d", host, port),
		},
	}, nil
}

func runtimeDefaultStorageDataDir(cfg config.Config) string {
	exePath, err := os.Executable()
	if err != nil || strings.TrimSpace(exePath) == "" {
		return filepath.Join(cfg.DataDir, "data")
	}
	return filepath.Join(filepath.Dir(exePath), "data")
}

func resolveRuntimeSetupStorageDataDir(cfg config.Config, req runtimeSetupRequest, profileKey string) string {
	storageMode := strings.TrimSpace(strings.ToLower(req.StorageMode))
	if storageMode == "" {
		storageMode = "exe_local"
	}
	if storageMode == "custom" {
		custom := strings.TrimSpace(req.StorageDataDir)
		if custom != "" {
			return custom
		}
	}
	defaultDir := runtimeDefaultStorageDataDir(cfg)
	if strings.TrimSpace(defaultDir) == "" {
		return filepath.Join(cfg.DataDir, "profiles", profileKey)
	}
	return defaultDir
}

func checkRuntimeSetupStorageWritable(targetDir string) (bool, string) {
	cleanDir := strings.TrimSpace(targetDir)
	if cleanDir == "" {
		return false, "Storage data directory is required."
	}
	if err := os.MkdirAll(cleanDir, 0o755); err != nil {
		return false, "Storage path is not writable."
	}
	testFile := filepath.Join(cleanDir, ".cabinet-write-check.tmp")
	if err := os.WriteFile(testFile, []byte("ok"), 0o644); err != nil {
		return false, "Storage path is not writable."
	}
	_ = os.Remove(testFile)
	return true, "Storage path is writable."
}

func deriveProfileKey(rawProfileKey, instanceName string) string {
	candidate := strings.TrimSpace(rawProfileKey)
	if candidate == "" {
		candidate = strings.TrimSpace(instanceName)
	}
	if candidate == "" {
		return "primary"
	}
	normalized := strings.ToLower(candidate)
	normalized = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-")
	if normalized == "" {
		return "primary"
	}
	return normalized
}

func validateRuntimeSetupConfigFile(payload runtimeSetupConfigFile) error {
	if strings.TrimSpace(payload.Instance.Name) == "" {
		return fmt.Errorf("instance.name is required")
	}
	if strings.TrimSpace(payload.Instance.Profile) == "" {
		return fmt.Errorf("instance.profile is required")
	}
	if strings.TrimSpace(payload.Storage.DataDir) == "" {
		return fmt.Errorf("storage.dataDir is required")
	}
	if strings.TrimSpace(payload.Storage.MediaDir) == "" {
		return fmt.Errorf("storage.mediaDir is required")
	}
	if strings.TrimSpace(payload.Runtime.ResolvedURL) == "" {
		return fmt.Errorf("runtime.resolvedUrl is required")
	}
	authMode := strings.TrimSpace(strings.ToLower(payload.Auth.Mode))
	if authMode == "" {
		return fmt.Errorf("auth.mode is required")
	}
	if authMode != "local" && authMode != "clerk" {
		return fmt.Errorf("auth.mode must be local or clerk")
	}
	if authMode == "clerk" && strings.TrimSpace(payload.Auth.Clerk.PublishableKey) == "" {
		return fmt.Errorf("auth.clerk.publishableKey is required for clerk mode")
	}
	if strings.TrimSpace(payload.Meta.CreatedAt) == "" {
		return fmt.Errorf("meta.createdAt is required")
	}
	if strings.TrimSpace(payload.Meta.UpdatedAt) == "" {
		return fmt.Errorf("meta.updatedAt is required")
	}
	return nil
}

func importRuntimeSetupConfigFromPath(cfg config.Config, sourcePath string) (runtimeSetupConfigFile, error) {
	raw, err := os.ReadFile(strings.TrimSpace(sourcePath))
	if err != nil {
		return runtimeSetupConfigFile{}, fmt.Errorf("failed to read source config")
	}
	var payload runtimeSetupConfigFile
	if err := json.Unmarshal(raw, &payload); err != nil {
		return runtimeSetupConfigFile{}, fmt.Errorf("source config is not valid JSON")
	}
	if err := validateRuntimeSetupConfigFile(payload); err != nil {
		return runtimeSetupConfigFile{}, fmt.Errorf("source config validation failed: %v", err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if strings.TrimSpace(payload.Meta.CreatedAt) == "" {
		payload.Meta.CreatedAt = now
	}
	payload.Meta.UpdatedAt = now
	if strings.TrimSpace(payload.Meta.WizardVersion) == "" {
		payload.Meta.WizardVersion = "1"
	}
	if strings.TrimSpace(payload.Runtime.ResolvedURL) == "" {
		payload.Runtime.ResolvedURL = runtimeResolvedURLFromConfig(cfg)
	}
	if strings.TrimSpace(payload.Meta.CurrentURL) == "" {
		payload.Meta.CurrentURL = strings.TrimSpace(payload.Runtime.ResolvedURL)
	}
	if err := writeRuntimeSetupConfig(cfg, payload); err != nil {
		return runtimeSetupConfigFile{}, fmt.Errorf("failed to write imported setup config")
	}
	return payload, nil
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

func persistCloudSessionContext(ctx context.Context, conn *sql.DB, userID, email, role string) error {
	updates := map[string]string{
		"cloud.user_id": strings.TrimSpace(userID),
		"cloud.email":   strings.TrimSpace(strings.ToLower(email)),
		"cloud.role":    strings.TrimSpace(strings.ToLower(role)),
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

func appStateValue(ctx context.Context, conn *sql.DB, key string) (string, bool) {
	if conn == nil {
		return "", false
	}
	var value string
	if err := conn.QueryRowContext(ctx, `SELECT value FROM app_state WHERE key = ?`, strings.TrimSpace(key)).Scan(&value); err != nil {
		return "", false
	}
	return strings.TrimSpace(value), true
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
		return cloudPlanHasFeature(cloudPlan, feature)
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

func parsePokemonSetProgressTags(tags []string) (setID, variant, language string) {
	setID = ""
	variant = "unknown"
	language = "unknown"
	for _, tag := range tags {
		trimmed := strings.TrimSpace(strings.ToLower(tag))
		if strings.HasPrefix(trimmed, "set:") {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "set:"))
			if value != "" {
				setID = value
			}
			continue
		}
		if strings.HasPrefix(trimmed, "variant:") {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "variant:"))
			if value != "" {
				variant = value
			}
			continue
		}
		if strings.HasPrefix(trimmed, "language:") {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "language:"))
			if value != "" {
				language = value
			}
		}
	}
	return setID, variant, language
}

func roundTo2(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}

func defaultConditionGrades() []string {
	return []string{"M", "NM", "EX", "VG", "G", "F", "P"}
}

func defaultPackagingGrades() []string {
	return []string{"sealed_mint", "sealed_good", "opened_complete", "loose"}
}

type itemTypeConditionScale struct {
	ItemType   string   `json:"item_type"`
	Conditions []string `json:"conditions"`
}

func defaultItemTypeConditionScales() []itemTypeConditionScale {
	return []itemTypeConditionScale{
		{
			ItemType: "Slot Cars",
			Conditions: []string{
				"10+ - New, in packaging",
				"10 - New, with packaging separate",
				"9 - New, no packaging",
				"8 - Like new",
				"7 - Minor track-wear",
				"6 - Bumper-wear & scratches",
				"5 - Worn, with scratches & nicks",
				"4 - Cut wheel wells, but nice",
				"3 - Cut badly, but good runner",
				"2 - Good for parts only",
				"1 - Good for nothing",
			},
		},
		{
			ItemType: "Trading Cards",
			Conditions: []string{
				"Mint (M)",
				"Near Mint (NM)",
				"Excellent (EX)",
				"Good (GD)",
				"Light Played (LP)",
				"Played (PL)",
				"Poor (PO)",
			},
		},
	}
}

func parseItemTypeConditionScalesSetting(raw string) []itemTypeConditionScale {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return defaultItemTypeConditionScales()
	}
	var values []itemTypeConditionScale
	if err := json.Unmarshal([]byte(trimmed), &values); err != nil {
		return defaultItemTypeConditionScales()
	}
	return normalizeItemTypeConditionScales(values, defaultItemTypeConditionScales())
}

func normalizeItemTypeConditionScales(input []itemTypeConditionScale, fallback []itemTypeConditionScale) []itemTypeConditionScale {
	out := make([]itemTypeConditionScale, 0, len(input))
	seenTypes := map[string]struct{}{}
	for _, scale := range input {
		itemType := strings.TrimSpace(scale.ItemType)
		if itemType == "" {
			continue
		}
		typeKey := strings.ToLower(itemType)
		if _, ok := seenTypes[typeKey]; ok {
			continue
		}
		conditions := normalizeDisplayStringList(scale.Conditions)
		if len(conditions) == 0 {
			continue
		}
		seenTypes[typeKey] = struct{}{}
		out = append(out, itemTypeConditionScale{
			ItemType:   itemType,
			Conditions: conditions,
		})
	}
	if len(out) == 0 {
		return append([]itemTypeConditionScale(nil), fallback...)
	}
	return out
}

func normalizeDisplayStringList(input []string) []string {
	out := make([]string, 0, len(input))
	seen := map[string]struct{}{}
	for _, item := range input {
		clean := strings.TrimSpace(item)
		if clean == "" {
			continue
		}
		key := strings.ToLower(clean)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, clean)
	}
	return out
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

type inventoryTaxonomyValidationError struct {
	Field   string
	Value   string
	Message string
}

func (e *inventoryTaxonomyValidationError) Error() string {
	return e.Message
}

func writeInventoryTaxonomyValidationError(w http.ResponseWriter, validationErr *inventoryTaxonomyValidationError) {
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error":   "invalid_taxonomy_value",
		"field":   validationErr.Field,
		"value":   validationErr.Value,
		"message": validationErr.Message,
	})
}

func validateInventoryItemTaxonomy(item collection.Item, existing *collection.Item, settings map[string]string) *inventoryTaxonomyValidationError {
	scales := parseItemTypeConditionScalesSetting(settings["grading.enums.item_type_condition_scales"])
	packagingGrades := parseStringArraySetting(settings["grading.enums.packaging"], defaultPackagingGrades())

	itemType := strings.TrimSpace(item.ItemType)
	if itemType == "" && existing != nil {
		itemType = strings.TrimSpace(existing.ItemType)
	}
	if strings.TrimSpace(item.ItemType) != "" && !itemTypeExists(scales, item.ItemType) {
		return &inventoryTaxonomyValidationError{
			Field:   "item_type",
			Value:   strings.TrimSpace(item.ItemType),
			Message: "item_type must match a configured item type condition scale for the active profile",
		}
	}

	condition := strings.TrimSpace(item.Condition)
	if condition != "" && !conditionExistsForItemType(scales, itemType, condition) {
		return &inventoryTaxonomyValidationError{
			Field:   "condition",
			Value:   condition,
			Message: "condition must match the configured condition scale for the selected item type",
		}
	}

	packagingGrade := strings.TrimSpace(item.PackagingGradeType)
	if packagingGrade != "" && !displayListContains(packagingGrades, packagingGrade) {
		return &inventoryTaxonomyValidationError{
			Field:   "packaging_grade_type",
			Value:   packagingGrade,
			Message: "packaging_grade_type must match a configured packaging grade for the active profile",
		}
	}
	return nil
}

func itemTypeExists(scales []itemTypeConditionScale, value string) bool {
	clean := strings.TrimSpace(value)
	for _, scale := range scales {
		if strings.EqualFold(strings.TrimSpace(scale.ItemType), clean) {
			return true
		}
	}
	return false
}

func conditionExistsForItemType(scales []itemTypeConditionScale, itemType string, condition string) bool {
	clean := strings.TrimSpace(condition)
	selectedType := strings.TrimSpace(itemType)
	if selectedType == "" {
		for _, scale := range scales {
			if displayListContains(scale.Conditions, clean) {
				return true
			}
		}
		return false
	}
	for _, scale := range scales {
		if strings.EqualFold(strings.TrimSpace(scale.ItemType), selectedType) {
			return displayListContains(scale.Conditions, clean)
		}
	}
	return false
}

func displayListContains(values []string, value string) bool {
	clean := strings.TrimSpace(value)
	for _, candidate := range values {
		if strings.EqualFold(strings.TrimSpace(candidate), clean) {
			return true
		}
	}
	return false
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

func firstProviderInScope(scope []string) string {
	for _, raw := range scope {
		normalized := strings.TrimSpace(strings.ToLower(raw))
		if normalized != "" {
			return normalized
		}
	}
	return ""
}

func providerScopeIncludes(scope []string, providerID string) bool {
	want := strings.TrimSpace(strings.ToLower(providerID))
	if want == "" {
		return false
	}
	for _, raw := range scope {
		if strings.TrimSpace(strings.ToLower(raw)) == want {
			return true
		}
	}
	return false
}

func entitlementFeaturesFromPlan(plan string) []string {
	switch strings.TrimSpace(strings.ToLower(plan)) {
	case "teams", "pro", "paid", "plus":
		return []string{"collection_core", "ai_assist", "price_tracking", "scanner_automation"}
	case "creator":
		return []string{"collection_core", "ai_assist", "scanner_automation"}
	case "mvp":
		return []string{"collection_core"}
	default:
		return []string{"collection_core"}
	}
}

func cloudPlanHasFeature(plan, feature string) bool {
	target := strings.TrimSpace(strings.ToLower(feature))
	if target == "" {
		return true
	}
	for _, entry := range entitlementFeaturesFromPlan(plan) {
		if strings.TrimSpace(strings.ToLower(entry)) == target {
			return true
		}
	}
	return false
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

func ebayProviderErrorNextAction(providerErr *ebay.ProviderError) string {
	if providerErr == nil {
		return "check_provider_health_and_credentials"
	}
	switch providerErr.ErrorCode {
	case "PROVIDER_AUTH_MISSING", "PROVIDER_AUTH_INVALID":
		return "review_provider_credentials_and_health"
	default:
		return "check_provider_health_and_credentials"
	}
}

func openAIProviderHealth(ctx context.Context, profiles *profile.Repository) map[string]any {
	base := map[string]any{
		"provider": "openai",
	}
	if profiles == nil {
		base["status"] = "needs_config"
		base["code"] = "OPENAI_PROFILE_REQUIRED"
		base["message"] = "Select an active profile before validating OpenAI."
		base["next_action"] = "select_profile"
		return base
	}
	active, err := profiles.GetActiveProfile(ctx)
	if err != nil || strings.TrimSpace(active.ID) == "" {
		base["status"] = "needs_config"
		base["code"] = "OPENAI_PROFILE_REQUIRED"
		base["message"] = "Select an active profile before validating OpenAI."
		base["next_action"] = "select_profile"
		return base
	}
	settings, err := profiles.GetSettings(ctx, strings.TrimSpace(active.ID))
	if err != nil {
		settings = map[string]string{}
	}
	activeMethod := strings.TrimSpace(settings["openai.active_auth_method"])
	if activeMethod == "" {
		activeMethod = strings.TrimSpace(settings["openai_active_auth_method"])
	}
	base["auth_method"] = activeMethod

	switch activeMethod {
	case "api_key":
		key, err := profiles.GetSecret(ctx, strings.TrimSpace(active.ID), "openai_api_key")
		if err != nil || strings.TrimSpace(key) == "" {
			base["status"] = "needs_config"
			base["code"] = "OPENAI_API_KEY_MISSING"
			base["credential_present"] = false
			base["message"] = "OpenAI API-key mode is selected, but no API key secret is stored for the active profile."
			base["next_action"] = "connect_openai_api_key"
			return base
		}
		base["status"] = "ready"
		base["code"] = "OPENAI_API_KEY_PRESENT"
		base["credential_present"] = true
		base["message"] = "OpenAI API-key credentials are stored for the active profile."
		base["next_action"] = "run_openai_test"
		return base
	case "browser_auth":
		state := strings.TrimSpace(settings["openai.browser_auth_state"])
		artifactPresent := strings.EqualFold(strings.TrimSpace(settings["openai.browser_auth_artifact_present"]), "true")
		proofState, proofArtifactID, proofMessage, proofPassed := openAIBrowserAuthProviderTestProof(settings)
		base["browser_auth_state"] = map[bool]string{true: state, false: "setup_needed"}[state != ""]
		base["credential_present"] = artifactPresent
		base["provider_test_passed"] = proofPassed
		if proofArtifactID != "" {
			base["provider_test_artifact_id"] = proofArtifactID
		}
		if proofState != "" {
			base["provider_test_state"] = proofState
		}
		if strings.EqualFold(state, "connected") && artifactPresent && proofPassed {
			base["status"] = "ready"
			base["code"] = "OPENAI_BROWSER_AUTH_PROVIDER_TEST_PASSED"
			base["message"] = "OpenAI Browser Auth has verified profile proof and passed provider-test evidence."
			if proofMessage != "" {
				base["message"] = proofMessage
			}
			base["next_action"] = "run_openai_workflow"
			return base
		}
		if strings.EqualFold(state, "connected") && artifactPresent && strings.EqualFold(proofState, "failed") {
			base["status"] = "failed"
			base["code"] = "OPENAI_BROWSER_AUTH_PROVIDER_TEST_FAILED"
			base["message"] = "OpenAI Browser Auth provider-test proof failed."
			if proofMessage != "" {
				base["message"] = proofMessage
			}
			base["next_action"] = "review_browser_auth_provider_test_proof"
			return base
		}
		base["status"] = "needs_config"
		base["code"] = "OPENAI_BROWSER_AUTH_PROOF_REQUIRED"
		base["message"] = "Browser Auth requires a verifiable callback/artifact and provider-test proof before Cabinet marks OpenAI ready."
		base["next_action"] = "complete_browser_auth_provider_test_adapter"
		return base
	default:
		base["status"] = "needs_config"
		base["code"] = "OPENAI_AUTH_METHOD_REQUIRED"
		base["credential_present"] = false
		base["message"] = "Choose OpenAI API-key mode or complete verified Browser Auth before running OpenAI workflows."
		base["next_action"] = "connect_openai_api_key_or_browser_auth"
		return base
	}
}

func openAIProviderTest(ctx context.Context, profiles *profile.Repository, aiSvc *ai.Service, profileID string) (map[string]any, int) {
	base := map[string]any{
		"provider":             "openai",
		"provider_test_passed": false,
		"checked_at":           time.Now().UTC().Format(time.RFC3339),
	}
	if profiles == nil {
		base["status"] = "needs_config"
		base["code"] = "OPENAI_PROFILE_REQUIRED"
		base["message"] = "Select an active profile before testing OpenAI."
		base["next_action"] = "select_profile"
		return base, http.StatusBadRequest
	}
	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		active, err := profiles.GetActiveProfile(ctx)
		if err == nil {
			profileID = strings.TrimSpace(active.ID)
		}
	}
	if profileID == "" {
		base["status"] = "needs_config"
		base["code"] = "OPENAI_PROFILE_REQUIRED"
		base["message"] = "Select an active profile before testing OpenAI."
		base["next_action"] = "select_profile"
		return base, http.StatusBadRequest
	}
	base["profile_id"] = profileID
	settings, err := profiles.GetSettings(ctx, profileID)
	if err != nil {
		settings = map[string]string{}
	}
	activeMethod := strings.TrimSpace(settings["openai.active_auth_method"])
	if activeMethod == "" {
		activeMethod = strings.TrimSpace(settings["openai_active_auth_method"])
	}
	base["auth_method"] = activeMethod
	baseURL := strings.TrimSpace(settings["openai_base_url"])
	if baseURL != "" {
		base["base_domain"] = providerBaseDomain(baseURL)
	}

	switch activeMethod {
	case "api_key":
		key, err := profiles.GetSecret(ctx, profileID, "openai_api_key")
		if err != nil || strings.TrimSpace(key) == "" {
			base["status"] = "needs_config"
			base["code"] = "OPENAI_API_KEY_MISSING"
			base["credential_present"] = false
			base["message"] = "OpenAI API-key mode is selected, but no API key secret is stored for the active profile."
			base["next_action"] = "connect_openai_api_key"
			return base, http.StatusBadRequest
		}
		base["credential_present"] = true
		localAISvc := aiSvc
		if localAISvc == nil || baseURL != "" {
			localAISvc = ai.NewService(ai.Config{BaseURL: baseURL})
		}
		if err := localAISvc.TestConnectivity(ctx, key); err != nil {
			base["status"] = "failed"
			base["code"] = "OPENAI_PROVIDER_TEST_FAILED"
			base["message"] = "OpenAI provider test failed: " + err.Error()
			base["next_action"] = "review_openai_credentials_and_provider_status"
			return base, http.StatusBadRequest
		}
		base["status"] = "ready"
		base["code"] = "OPENAI_PROVIDER_TEST_PASSED"
		base["message"] = "OpenAI provider test passed for the active profile."
		base["next_action"] = "run_openai_workflow"
		base["provider_test_passed"] = true
		return base, http.StatusOK
	case "browser_auth":
		state := strings.TrimSpace(settings["openai.browser_auth_state"])
		artifactPresent := strings.EqualFold(strings.TrimSpace(settings["openai.browser_auth_artifact_present"]), "true")
		proofState, proofArtifactID, proofMessage, proofPassed := openAIBrowserAuthProviderTestProof(settings)
		base["browser_auth_state"] = map[bool]string{true: state, false: "setup_needed"}[state != ""]
		base["credential_present"] = artifactPresent
		if proofArtifactID != "" {
			base["provider_test_artifact_id"] = proofArtifactID
		}
		if proofState != "" {
			base["provider_test_state"] = proofState
		}
		if !strings.EqualFold(state, "connected") || !artifactPresent {
			base["status"] = "needs_config"
			base["code"] = "OPENAI_BROWSER_AUTH_PROOF_REQUIRED"
			base["message"] = "Browser Auth requires a verified callback or artifact before provider-test proof can be accepted."
			base["next_action"] = "complete_browser_auth_verification"
			return base, http.StatusBadRequest
		}
		if strings.EqualFold(proofState, "failed") && proofArtifactID != "" {
			base["status"] = "failed"
			base["code"] = "OPENAI_BROWSER_AUTH_PROVIDER_TEST_FAILED"
			base["message"] = "OpenAI Browser Auth provider-test proof failed."
			if proofMessage != "" {
				base["message"] = proofMessage
			}
			base["next_action"] = "review_browser_auth_provider_test_proof"
			return base, http.StatusBadRequest
		}
		if proofPassed {
			base["status"] = "ready"
			base["code"] = "OPENAI_BROWSER_AUTH_PROVIDER_TEST_PASSED"
			base["message"] = "OpenAI Browser Auth provider-test proof passed for the active profile."
			if proofMessage != "" {
				base["message"] = proofMessage
			}
			base["next_action"] = "run_openai_workflow"
			base["provider_test_passed"] = true
			return base, http.StatusOK
		}
		base["status"] = "needs_config"
		base["code"] = "OPENAI_BROWSER_AUTH_PROVIDER_TEST_PROOF_REQUIRED"
		base["message"] = "Browser Auth requires a verified runtime provider-test adapter proof before Cabinet can mark OpenAI live-provider tested."
		base["next_action"] = "complete_browser_auth_provider_test_adapter"
		return base, http.StatusBadRequest
	default:
		base["status"] = "needs_config"
		base["code"] = "OPENAI_AUTH_METHOD_REQUIRED"
		base["credential_present"] = false
		base["message"] = "Choose OpenAI API-key mode or complete verified Browser Auth before running OpenAI provider tests."
		base["next_action"] = "connect_openai_api_key_or_browser_auth"
		return base, http.StatusBadRequest
	}
}

func openAIBrowserAuthProviderTestProof(settings map[string]string) (string, string, string, bool) {
	state := strings.ToLower(strings.TrimSpace(settings["openai.browser_auth_provider_test_state"]))
	artifactID := strings.TrimSpace(settings["openai.browser_auth_provider_test_artifact_id"])
	message := strings.TrimSpace(settings["openai.browser_auth_provider_test_message"])
	return state, artifactID, message, state == "passed" && artifactID != ""
}

func providerBaseDomain(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || strings.TrimSpace(u.Host) == "" {
		return ""
	}
	return u.Host
}

func providerHealthResponse(health map[string]string) map[string]any {
	provider := strings.TrimSpace(health["provider"])
	status := strings.TrimSpace(health["status"])
	message := strings.TrimSpace(health["message"])
	updatedAt := strings.TrimSpace(health["updated_at"])
	retryAfterRaw := strings.TrimSpace(health["retry_after_seconds"])
	if status == "" {
		status = "unknown"
	}

	state := "disabled"
	switch strings.ToLower(status) {
	case "ok", "ready":
		state = "ready"
	case "error", "degraded":
		state = "degraded"
	}

	var lastError any
	if message != "" && state != "ready" {
		lastError = message
	}

	var updated any
	if updatedAt != "" {
		updated = updatedAt
	}

	var retryAfter any
	if retryAfterRaw != "" {
		if parsed, err := strconv.Atoi(retryAfterRaw); err == nil && parsed > 0 {
			retryAfter = parsed
		}
	}

	var nextAction any
	if state == "degraded" {
		switch strings.ToLower(provider) {
		case "ebay":
			nextAction = "check_provider_health_and_credentials"
		default:
			nextAction = "review_provider_status"
		}
	}

	return map[string]any{
		"provider":            provider,
		"status":              status,
		"state":               state,
		"message":             message,
		"last_error":          lastError,
		"retry_after_seconds": retryAfter,
		"next_action":         nextAction,
		"updated_at":          updated,
	}
}

func providerRegistryPayload(ctx context.Context, conn *sql.DB, scannerSvc *scanner.Service, amazonMode string, settings map[string]string) []map[string]any {
	familyOverrides := map[string]string{}
	if conn != nil {
		if loaded, err := loadProviderFamilyOverrides(ctx, conn); err == nil {
			familyOverrides = loaded
		}
	}
	openAIActiveMethod := strings.TrimSpace(settings["openai.active_auth_method"])
	openAIAPIKeyPresent := openAIActiveMethod == "api_key" && strings.EqualFold(strings.TrimSpace(settings["openai.api_key_secret_present"]), "true")
	openAIBrowserState := strings.TrimSpace(settings["openai.browser_auth_state"])
	openAIBrowserCredentialPresent := strings.EqualFold(strings.TrimSpace(settings["openai.browser_auth_artifact_present"]), "true")
	openAIBrowserProofState, openAIBrowserProofArtifactID, _, openAIBrowserProviderTestPassed := openAIBrowserAuthProviderTestProof(settings)
	openAIBrowserReady := openAIActiveMethod == "browser_auth" && strings.EqualFold(openAIBrowserState, "connected") && openAIBrowserCredentialPresent && openAIBrowserProviderTestPassed
	openAIReady := openAIAPIKeyPresent || openAIBrowserReady
	base := []map[string]any{
		{
			"provider_id":         "openai",
			"display_name":        "OpenAI / ChatGPT",
			"base_domain":         "platform.openai.com",
			"api_family":          "ai_provider",
			"api_support_profile": "browser_auth_or_api_key",
			"active_mode":         openAIActiveMethod,
			"integration_mode":    "assistant_workflows",
			"api_available":       true,
			"auth_requirement":    "browser_auth_or_api_key",
			"auth_mode":           "hybrid",
			"active_auth_method":  openAIActiveMethod,
			"auth_methods": map[string]any{
				"api_key": map[string]any{
					"state":              map[bool]string{true: "connected", false: "setup_needed"}[openAIAPIKeyPresent],
					"connected":          openAIAPIKeyPresent,
					"credential_present": openAIAPIKeyPresent,
				},
				"browser_auth": map[string]any{
					"state":                     map[bool]string{true: openAIBrowserState, false: "setup_needed"}[openAIBrowserState != ""],
					"connected":                 openAIBrowserReady,
					"credential_present":        openAIBrowserCredentialPresent,
					"provider_test_passed":      openAIBrowserProviderTestPassed,
					"provider_test_state":       openAIBrowserProofState,
					"provider_test_artifact_id": openAIBrowserProofArtifactID,
					"setup_message":             "Browser Auth requires a verifiable callback/artifact and provider-test proof before Cabinet marks OpenAI connected.",
				},
			},
			"model_options": []string{"gpt-4o-mini", "gpt-4.1-mini", "gpt-5.3-codex"},
			"capabilities": map[string]bool{
				"search":             false,
				"stock_observation":  false,
				"pricing":            false,
				"health":             true,
				"assistant":          true,
				"image_help":         true,
				"content_generation": true,
			},
			"state":              map[bool]string{true: "ready", false: "needs_config"}[openAIReady],
			"setup_instructions": "Configure OpenAI with Browser Auth or an API key. Browser Auth stays setup-needed until Cabinet verifies an auth artifact/callback; navigation alone is never connected proof.",
		},
		{
			"provider_id":         "telegram",
			"display_name":        "Telegram",
			"base_domain":         "telegram.org",
			"api_family":          "messaging_channel",
			"api_support_profile": "bot_webhook_sender_chat_v1",
			"active_mode":         map[bool]string{true: "authorized_sender_chat", false: "setup_needed"}[telegramCatalogCaptureConfigured(settings)],
			"integration_mode":    "assistant_capture_channel",
			"api_available":       true,
			"auth_requirement":    "sender_chat_authorization",
			"auth_mode":           "sender_chat",
			"auth_methods": map[string]any{
				"sender_chat": map[string]any{
					"state":              map[bool]string{true: "connected", false: "setup_needed"}[telegramCatalogCaptureConfigured(settings)],
					"connected":          telegramCatalogCaptureConfigured(settings),
					"credential_present": telegramCatalogCaptureConfigured(settings),
					"setup_message":      "Store the Telegram sender id and chat id on the active profile before Cabinet accepts capture messages.",
				},
			},
			"capabilities": map[string]bool{
				"search":        false,
				"health":        true,
				"assistant":     true,
				"media_capture": true,
				"text_capture":  true,
			},
			"state":              map[bool]string{true: "ready", false: "needs_config"}[telegramCatalogCaptureConfigured(settings)],
			"setup_instructions": "Configure Telegram sender/chat authorization in Profile settings, then route bot messages through the governed preview-before-apply capture channel.",
		},
		{
			"provider_id":         "ebay",
			"display_name":        "eBay",
			"base_domain":         "ebay.com",
			"api_family":          "official_api",
			"api_support_profile": "rest_v1",
			"active_mode":         "official_api",
			"integration_mode":    "official_api",
			"api_available":       true,
			"auth_requirement":    "api_key",
			"auth_mode":           "api_key",
			"capabilities": map[string]bool{
				"search":            true,
				"stock_observation": false,
				"pricing":           true,
				"health":            true,
			},
			"has_token":          strings.TrimSpace(settings["ebay_bearer_token"]) != "",
			"setup_status":       ebaySetupStatus(settings, ""),
			"seller_operations":  ebay.SellerOperationStatuses(nil),
			"state":              "ready",
			"setup_instructions": "Add eBay API token and marketplace, validate health, then run scanner query sets.",
		},
		{
			"provider_id":          "amazon",
			"display_name":         "Amazon",
			"base_domain":          "amazon.com",
			"api_family":           "official_api",
			"api_support_profile":  "program_api_v1",
			"active_mode":          map[bool]string{true: "program_api", false: "disabled"}[amazonMode == "program_api"],
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

	auDomains := resolveAUWebshopDomains(settings)
	for _, d := range auDomains {
		apiFamily := "web_ingestion"
		activeMode := "web_ingestion"
		integrationMode := "web_ingestion"
		supportProfile := "html_fallback"
		if d == "voglers.com.au" {
			apiFamily = "bigcommerce"
			activeMode = "storefront_public"
			integrationMode = "storefront_access"
			supportProfile = "bigcommerce_storefront_v1"
		}
		if d == "mrtoys.com.au" {
			apiFamily = "doofinder"
			activeMode = "hashid_search"
			integrationMode = "api_family_search"
			supportProfile = "doofinder_hashid_v1"
		}
		if d == "bonzaslotcars.com.au" {
			apiFamily = "woo_store_api"
			activeMode = "store_api_first"
			supportProfile = "store_v1"
		}
		if d == "frontlinehobbies.com.au" {
			apiFamily = "algolia"
			activeMode = "algolia_runtime"
			supportProfile = "algolia_runtime_v1"
		}
		if d == "hobbytechtoys.com.au" {
			apiFamily = "boost_shopify"
			activeMode = "boost_api"
			supportProfile = "boost_v2"
		}
		base = append(base, map[string]any{
			"provider_id":         "au-webshop-" + strings.ReplaceAll(d, ".", "-"),
			"display_name":        d,
			"base_domain":         d,
			"api_family":          apiFamily,
			"api_support_profile": supportProfile,
			"active_mode":         activeMode,
			"integration_mode":    integrationMode,
			"api_available":       d == "voglers.com.au" || d == "mrtoys.com.au",
			"auth_requirement":    "none",
			"auth_mode":           "none",
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
		if strings.EqualFold(fmt.Sprintf("%v", provider["api_family"]), "bigcommerce") {
			if hasToken {
				provider["active_mode"] = "token_enabled"
				provider["api_support_profile"] = "bigcommerce_token_v1"
				provider["auth_requirement"] = "api_token_optional"
				provider["auth_mode"] = "token_optional"
			} else {
				provider["active_mode"] = "storefront_public"
				provider["api_support_profile"] = "bigcommerce_storefront_v1"
				provider["auth_requirement"] = "none"
				provider["auth_mode"] = "none"
			}
		}

		healthStatus := "unknown"
		healthPayload := map[string]any{
			"status":              healthStatus,
			"state":               "disabled",
			"message":             "",
			"last_error":          nil,
			"retry_after_seconds": nil,
			"next_action":         nil,
			"last_checked_at":     nil,
			"updated_at":          nil,
		}
		lastChecked := any(nil)
		lastRunStatus := "never"
		lastRunFinished := any(nil)
		if scannerSvc != nil {
			if health, err := scannerSvc.ProviderHealth(ctx, providerID); err == nil {
				healthPayload = providerHealthResponse(health)
				healthPayload["last_checked_at"] = healthPayload["updated_at"]
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
				if v := strings.TrimSpace(health["updated_at"]); v != "" {
					lastChecked = v
					lastRunFinished = v
				}
			}
		}
		healthPayload["last_checked_at"] = lastChecked
		provider["health"] = healthPayload
		provider["last_run"] = map[string]any{
			"status":      lastRunStatus,
			"finished_at": lastRunFinished,
		}
		if strings.EqualFold(providerID, "ebay") {
			provider["setup_status"] = ebaySetupStatus(settings, healthStatus)
		}
		baseDomain := normalizeProviderDomain(fmt.Sprintf("%v", provider["base_domain"]))
		if overrideFamily, ok := familyOverrides[baseDomain]; ok && strings.TrimSpace(overrideFamily) != "" {
			provider["api_family"] = strings.TrimSpace(overrideFamily)
			switch strings.TrimSpace(strings.ToLower(overrideFamily)) {
			case "doofinder":
				provider["active_mode"] = "hashid_search"
				provider["integration_mode"] = "api_family_search"
				provider["api_available"] = true
				provider["api_support_profile"] = "doofinder_hashid_v1"
			case "bigcommerce":
				if hasToken {
					provider["active_mode"] = "token_enabled"
					provider["api_support_profile"] = "bigcommerce_token_v1"
				} else {
					provider["active_mode"] = "storefront_public"
					provider["api_support_profile"] = "bigcommerce_storefront_v1"
				}
			case "woocommerce":
				provider["active_mode"] = "store_api_first"
				provider["api_support_profile"] = "store_v1"
			case "boost_shopify":
				provider["active_mode"] = "boost_api"
				provider["api_support_profile"] = "boost_v2"
			case "algolia":
				provider["active_mode"] = "algolia_runtime"
				provider["api_support_profile"] = "algolia_runtime_v1"
			case "shopify_json":
				provider["active_mode"] = "products_json"
				provider["api_support_profile"] = "products_json_v1"
			default:
				provider["api_support_profile"] = "custom_override"
			}
		}
	}

	return base
}

func telegramCatalogCaptureConfigured(settings map[string]string) bool {
	return strings.TrimSpace(settings["telegram.catalog_capture.sender_id"]) != "" &&
		strings.TrimSpace(settings["telegram.catalog_capture.chat_id"]) != ""
}

func defaultAUWebshopDomains() []string {
	return []string{
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
}

func ebaySetupStatus(settings map[string]string, providerHealthStatus string) map[string]any {
	hasToken := strings.TrimSpace(settings["ebay_bearer_token"]) != ""
	marketplace := strings.TrimSpace(settings["ebay_marketplace"])
	if marketplace == "" {
		marketplace = "unset"
	}
	tokenState := "token_required"
	validationStatus := "needs_credentials"
	healthState := "disabled"
	nextAction := "save_ebay_credentials_and_marketplace"
	if hasToken {
		tokenState = "stored"
	}
	if hasToken && marketplace != "unset" {
		validationStatus = "ready"
		healthState = "ready"
		nextAction = "run_ebay_query_sets_from_market_watch"
	}
	if hasToken && marketplace != "unset" {
		switch strings.ToLower(strings.TrimSpace(providerHealthStatus)) {
		case "error", "failed", "degraded":
			validationStatus = "degraded"
			healthState = "degraded"
			nextAction = "check_provider_health_and_credentials"
		}
	}
	return map[string]any{
		"auth_mode":         "api_key",
		"marketplace":       marketplace,
		"token_state":       tokenState,
		"validation_status": validationStatus,
		"health_state":      healthState,
		"next_action":       nextAction,
		"base_url_set":      strings.TrimSpace(settings["ebay_base_url"]) != "",
	}
}

func resolveAUWebshopDomains(settings map[string]string) []string {
	defaults := defaultAUWebshopDomains()
	raw := strings.TrimSpace(settings["integration.au_webshops.domains"])
	if raw == "" {
		return defaults
	}

	seen := map[string]struct{}{}
	resolved := make([]string, 0, len(defaults))
	for _, part := range strings.Split(raw, ",") {
		domain := normalizeProviderDomain(part)
		if domain == "" {
			continue
		}
		if _, ok := seen[domain]; ok {
			continue
		}
		seen[domain] = struct{}{}
		resolved = append(resolved, domain)
	}
	if len(resolved) == 0 {
		return defaults
	}
	return resolved
}

type providerSettingKeySet struct {
	BaseURLKey      string
	TokenKey        string
	MarketplaceKey  string
	EnabledKey      string
	ItemsPerPageKey string
}

func providerSettingsKeys(providerID string) providerSettingKeySet {
	if strings.TrimSpace(strings.ToLower(providerID)) == "ebay" {
		return providerSettingKeySet{
			BaseURLKey:      "ebay_base_url",
			TokenKey:        "ebay_bearer_token",
			MarketplaceKey:  "ebay_marketplace",
			EnabledKey:      "integration.ebay.enabled",
			ItemsPerPageKey: "integration.ebay.items_per_page",
		}
	}
	slug := strings.TrimSpace(strings.ToLower(providerID))
	slug = strings.ReplaceAll(slug, "-", "_")
	slug = strings.ReplaceAll(slug, ".", "_")
	return providerSettingKeySet{
		BaseURLKey:      "integration." + slug + ".base_url",
		TokenKey:        "integration." + slug + ".token",
		MarketplaceKey:  "integration." + slug + ".marketplace",
		EnabledKey:      "integration." + slug + ".enabled",
		ItemsPerPageKey: "integration." + slug + ".items_per_page",
	}
}

type amazonContractProvider struct {
	candidates []scanner.CandidateInput
}

func (p amazonContractProvider) Search(context.Context, scanner.QuerySet) ([]scanner.CandidateInput, error) {
	return p.candidates, nil
}

func (p amazonContractProvider) ProviderID() string {
	return "amazon"
}

func buildAmazonCandidateContract(qs scanner.QuerySet) []scanner.CandidateInput {
	keyword := "collectible"
	if len(qs.Keywords) > 0 && strings.TrimSpace(qs.Keywords[0]) != "" {
		keyword = strings.TrimSpace(qs.Keywords[0])
	}
	return []scanner.CandidateInput{
		{
			ListingID:  "amazon-" + strings.ToLower(strings.ReplaceAll(keyword, " ", "-")) + "-001",
			Title:      "Amazon " + keyword + " listing",
			Price:      39.99,
			Currency:   "AUD",
			URL:        "https://amazon.com/dp/example",
			Seller:     "amazon-marketplace",
			Source:     "amazon",
			StockState: "available",
			StockCount: -1,
		},
	}
}

func buildAmazonCandidateResponseContract(candidates []scanner.CandidateInput) []map[string]any {
	out := make([]map[string]any, 0, len(candidates))
	for _, candidate := range candidates {
		out = append(out, map[string]any{
			"listing_id": candidate.ListingID,
			"title":      candidate.Title,
			"price": map[string]any{
				"amount":   candidate.Price,
				"currency": candidate.Currency,
			},
			"url":    candidate.URL,
			"seller": candidate.Seller,
			"source": map[string]any{
				"provider_id": candidate.Source,
			},
		})
	}
	return out
}

type frontlineAlgoliaConfig struct {
	ApplicationID string   `json:"application_id"`
	SearchKey     string   `json:"search_key"`
	IndexNames    []string `json:"index_names"`
	Source        string   `json:"source"`
	ConfigHash    string   `json:"config_hash"`
	DiscoveredAt  string   `json:"discovered_at"`
}

type doofinderConfig struct {
	Store        string `json:"store"`
	Zone         string `json:"zone"`
	HashID       string `json:"hashid"`
	APIBase      string `json:"api_base"`
	Source       string `json:"source"`
	ConfigHash   string `json:"config_hash"`
	DiscoveredAt string `json:"discovered_at"`
}

type hobbytechBoostConfig struct {
	Shop         string `json:"shop"`
	SID          string `json:"sid"`
	Template     string `json:"template"`
	Widget       string `json:"widget"`
	SearchURL    string `json:"search_url"`
	Source       string `json:"source"`
	ConfigHash   string `json:"config_hash"`
	DiscoveredAt string `json:"discovered_at"`
}

type bonzaProductResponse struct {
	ID                int                     `json:"id"`
	Name              string                  `json:"name"`
	Slug              string                  `json:"slug"`
	Permalink         string                  `json:"permalink"`
	Description       string                  `json:"description"`
	Prices            bonzaProductPrices      `json:"prices"`
	IsInStock         *bool                   `json:"is_in_stock"`
	LowStockRemaining *int                    `json:"low_stock_remaining"`
	Categories        []bonzaProductName      `json:"categories"`
	Attributes        []bonzaProductAttribute `json:"attributes"`
	Images            []bonzaProductImage     `json:"images"`
}

type bonzaSearchResult struct {
	PageCount        int              `json:"page_count"`
	ObservedPageSize int              `json:"observed_page_size"`
	ItemsPerPageUsed int              `json:"items_per_page_used"`
	Candidates       []map[string]any `json:"candidates"`
}

type bonzaProductPrices struct {
	CurrencyCode string `json:"currency_code"`
	Price        string `json:"price"`
}

type bonzaProductName struct {
	Name string `json:"name"`
}

type bonzaProductAttribute struct {
	Name    string                      `json:"name"`
	Terms   bonzaProductAttributeValues `json:"terms"`
	Options bonzaProductAttributeValues `json:"options"`
}

type bonzaProductImage struct {
	Src string `json:"src"`
}

type bonzaProductAttributeValues []string

func (values *bonzaProductAttributeValues) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*values = nil
		return nil
	}
	var stringsOnly []string
	if err := json.Unmarshal(data, &stringsOnly); err == nil {
		*values = compactBonzaAttributeValues(stringsOnly)
		return nil
	}
	var namedValues []bonzaProductName
	if err := json.Unmarshal(data, &namedValues); err != nil {
		return err
	}
	out := make([]string, 0, len(namedValues))
	for _, value := range namedValues {
		out = append(out, value.Name)
	}
	*values = compactBonzaAttributeValues(out)
	return nil
}

func compactBonzaAttributeValues(in []string) []string {
	out := make([]string, 0, len(in))
	for _, value := range in {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

type providerProductURLRoute struct {
	OriginalURL   string `json:"original_url"`
	NormalizedURL string `json:"normalized_url"`
	Host          string `json:"host"`
	Path          string `json:"path"`
	Provider      string `json:"provider"`
	Family        string `json:"family"`
	Slug          string `json:"slug,omitempty"`
	Action        string `json:"action"`
}

type providerProductDraft struct {
	ProviderProductID string            `json:"provider_product_id"`
	Title             string            `json:"title"`
	SourceURL         string            `json:"source_url"`
	Description       string            `json:"description"`
	Price             float64           `json:"price"`
	Currency          string            `json:"currency"`
	StockState        string            `json:"stock_state"`
	StockCount        int               `json:"stock_count"`
	Categories        []string          `json:"categories"`
	Attributes        map[string]string `json:"attributes"`
	ImageURLs         []string          `json:"image_urls"`
	Evidence          map[string]any    `json:"evidence"`
}

const frontlineAlgoliaCacheKey = "frontline_algolia_last_known_good"
const hobbytechBoostCacheKey = "hobbytech_boost_last_known_good"
const doofinderCacheKey = "doofinder_last_known_good"

func parsePositiveInt(raw string, fallback int) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func purchaseInboxCardToItem(card ebaypurchasecapture.PurchaseCard) collection.Item {
	title := strings.TrimSpace(firstNonEmptyString(card.PurchasedIdentity, card.ListingTitle))
	partNumber := strings.TrimSpace(firstNonEmptyString(card.ListingID, card.TransactionID, ebaypurchasecapture.PurchaseItemKey(card)))
	return collection.Item{
		Brand:      "Unknown",
		Category:   "General",
		PartNumber: partNumber,
		Title:      title,
		Status:     "active",
		Notes:      appendPurchaseInboxNote("", card),
		SourceURLs: appendPurchaseInboxSourceURL(nil, card.ItemURL),
		Tags:       []string{"purchase-inbox", "ebay-purchase"},
	}
}

func appendPurchaseInboxNote(existing string, card ebaypurchasecapture.PurchaseCard) string {
	parts := []string{"Purchase Inbox confirmed eBay capture"}
	if v := strings.TrimSpace(card.OrderID); v != "" {
		parts = append(parts, "order "+v)
	}
	if v := strings.TrimSpace(card.TransactionID); v != "" {
		parts = append(parts, "transaction "+v)
	}
	if v := strings.TrimSpace(card.ListingID); v != "" {
		parts = append(parts, "listing "+v)
	}
	if card.Quantity > 0 {
		parts = append(parts, fmt.Sprintf("quantity %d", card.Quantity))
	}
	if v := strings.TrimSpace(card.ItemPrice); v != "" {
		parts = append(parts, "price "+v)
	}
	note := strings.Join(parts, "; ") + "."
	if strings.TrimSpace(existing) == "" {
		return note
	}
	if strings.Contains(existing, note) {
		return existing
	}
	return strings.TrimSpace(existing) + "\n" + note
}

func appendPurchaseInboxSourceURL(existing []string, rawURL string) []string {
	trimmed := strings.TrimSpace(rawURL)
	out := append([]string{}, existing...)
	if trimmed == "" {
		return out
	}
	for _, value := range out {
		if strings.TrimSpace(value) == trimmed {
			return out
		}
	}
	return append(out, trimmed)
}

func purchaseInboxActionAudit(card ebaypurchasecapture.PurchaseCard, itemID string) map[string]any {
	return map[string]any{
		"source":         "ebay_purchase_capture",
		"item_id":        strings.TrimSpace(itemID),
		"target_key":     ebaypurchasecapture.PurchaseItemKey(card),
		"order_id":       strings.TrimSpace(card.OrderID),
		"listing_id":     strings.TrimSpace(card.ListingID),
		"transaction_id": strings.TrimSpace(card.TransactionID),
		"confirmed":      true,
	}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func discoverFrontlineAlgoliaConfig(
	ctx context.Context,
	conn *sql.DB,
	client *http.Client,
	assetURL string,
	fallbackAssetURLs []string,
) (frontlineAlgoliaConfig, bool, string, error) {
	if client == nil {
		client = http.DefaultClient
	}
	candidates := make([]string, 0, len(fallbackAssetURLs)+1)
	candidates = append(candidates, strings.TrimSpace(assetURL))
	for _, value := range fallbackAssetURLs {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			candidates = append(candidates, trimmed)
		}
	}
	var discoveryErr error
	for _, urlValue := range candidates {
		if urlValue == "" {
			continue
		}
		script, fetchErr := fetchProviderAsset(ctx, client, urlValue)
		if fetchErr != nil {
			discoveryErr = fetchErr
			continue
		}
		config, parseErr := parseFrontlineAlgoliaConfig(script, urlValue)
		if parseErr != nil {
			discoveryErr = parseErr
			continue
		}
		if persistErr := persistFrontlineAlgoliaCache(ctx, conn, config); persistErr != nil {
			discoveryErr = persistErr
			continue
		}
		return config, false, "", nil
	}
	cached, cacheErr := readFrontlineAlgoliaCache(ctx, conn)
	if cacheErr == nil {
		warning := "frontline discovery fallback used cached config"
		if discoveryErr != nil {
			warning = fmt.Sprintf("%s: %s", warning, strings.TrimSpace(discoveryErr.Error()))
		}
		return cached, true, warning, nil
	}
	if discoveryErr != nil {
		return frontlineAlgoliaConfig{}, false, "", discoveryErr
	}
	return frontlineAlgoliaConfig{}, false, "", cacheErr
}

func fetchProviderAsset(ctx context.Context, client *http.Client, assetURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSpace(assetURL), nil)
	if err != nil {
		return "", fmt.Errorf("build provider asset request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request provider asset: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		resp.Body.Close()
		return "", fmt.Errorf("provider asset returned status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", fmt.Errorf("read provider asset: %w", err)
	}
	return string(body), nil
}

func parseFrontlineAlgoliaConfig(script, source string) (frontlineAlgoliaConfig, error) {
	glgolia := regexp.MustCompile(`(?i)Glgoliasearch\s*\(\s*['"]([^'"]+)['"]\s*,\s*['"]([^'"]+)['"]\s*\)`)
	match := glgolia.FindStringSubmatch(script)
	if len(match) < 3 {
		return frontlineAlgoliaConfig{}, fmt.Errorf("missing algolia glgoliasearch credentials")
	}
	indexPattern := regexp.MustCompile(`(?i)index(?:Name)?\s*[:=]\s*['"]([^'"]+)['"]`)
	indexMatches := indexPattern.FindAllStringSubmatch(script, -1)
	seen := map[string]struct{}{}
	indexes := make([]string, 0, len(indexMatches))
	for _, entry := range indexMatches {
		if len(entry) < 2 {
			continue
		}
		value := strings.TrimSpace(entry[1])
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		indexes = append(indexes, value)
	}
	if len(indexes) == 0 {
		return frontlineAlgoliaConfig{}, fmt.Errorf("missing algolia index names")
	}
	hash := sha256.Sum256([]byte(script))
	return frontlineAlgoliaConfig{
		ApplicationID: strings.TrimSpace(match[1]),
		SearchKey:     strings.TrimSpace(match[2]),
		IndexNames:    indexes,
		Source:        strings.TrimSpace(source),
		ConfigHash:    fmt.Sprintf("%x", hash[:]),
		DiscoveredAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func persistFrontlineAlgoliaCache(ctx context.Context, conn *sql.DB, config frontlineAlgoliaConfig) error {
	raw, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal frontline cache config: %w", err)
	}
	_, execErr := conn.ExecContext(ctx, `
		INSERT INTO app_state(key, value, updated_at)
		VALUES(?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP
	`, frontlineAlgoliaCacheKey, string(raw))
	if execErr != nil {
		return fmt.Errorf("persist frontline cache config: %w", execErr)
	}
	return nil
}

func readFrontlineAlgoliaCache(ctx context.Context, conn *sql.DB) (frontlineAlgoliaConfig, error) {
	var raw string
	err := conn.QueryRowContext(ctx, `SELECT value FROM app_state WHERE key = ?`, frontlineAlgoliaCacheKey).Scan(&raw)
	if err != nil {
		return frontlineAlgoliaConfig{}, fmt.Errorf("load frontline cache config: %w", err)
	}
	var config frontlineAlgoliaConfig
	if decodeErr := json.Unmarshal([]byte(raw), &config); decodeErr != nil {
		return frontlineAlgoliaConfig{}, fmt.Errorf("decode frontline cache config: %w", decodeErr)
	}
	if strings.TrimSpace(config.ApplicationID) == "" || strings.TrimSpace(config.SearchKey) == "" {
		return frontlineAlgoliaConfig{}, fmt.Errorf("frontline cache config is incomplete")
	}
	return config, nil
}

func defaultFrontlineAlgoliaSearchURL(config frontlineAlgoliaConfig) string {
	appID := strings.ToLower(strings.TrimSpace(config.ApplicationID))
	index := ""
	if len(config.IndexNames) > 0 {
		index = strings.TrimSpace(config.IndexNames[0])
	}
	if appID == "" || index == "" {
		return ""
	}
	return "https://" + appID + "-dsn.algolia.net/1/indexes/" + url.PathEscape(index) + "/query"
}

func runFrontlineAlgoliaSearch(
	ctx context.Context,
	client *http.Client,
	searchURL string,
	qs scanner.QuerySet,
	config frontlineAlgoliaConfig,
	baseURL string,
	itemsPerPage int,
) ([]map[string]any, int, error) {
	if client == nil {
		client = http.DefaultClient
	}
	if strings.TrimSpace(searchURL) == "" {
		return nil, 0, fmt.Errorf("frontline search url is required")
	}
	query := strings.TrimSpace(qs.Name)
	if len(qs.Keywords) > 0 && strings.TrimSpace(qs.Keywords[0]) != "" {
		query = strings.TrimSpace(qs.Keywords[0])
	}
	if query == "" {
		query = "collectible"
	}
	if itemsPerPage <= 0 {
		itemsPerPage = 24
	}
	if itemsPerPage > 50 {
		itemsPerPage = 50
	}
	requestBody, err := json.Marshal(map[string]string{
		"params": url.Values{
			"query":       []string{query},
			"hitsPerPage": []string{strconv.Itoa(itemsPerPage)},
			"page":        []string{"0"},
		}.Encode(),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("marshal frontline search request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimSpace(searchURL), bytes.NewReader(requestBody))
	if err != nil {
		return nil, 0, fmt.Errorf("build frontline search request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Algolia-Application-Id", strings.TrimSpace(config.ApplicationID))
	req.Header.Set("X-Algolia-API-Key", strings.TrimSpace(config.SearchKey))
	req.Header.Set("User-Agent", "Cabinet/1.0 (+https://collectors.tech/cabinet)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("run frontline search request: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		resp.Body.Close()
		return nil, 0, fmt.Errorf("frontline search returned status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, 0, fmt.Errorf("read frontline search response: %w", err)
	}
	var payload struct {
		Hits   []map[string]any `json:"hits"`
		NBHits int              `json:"nbHits"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, 0, fmt.Errorf("decode frontline response: %w", err)
	}
	candidates := make([]map[string]any, 0, len(payload.Hits))
	for _, hit := range payload.Hits {
		id := firstNonEmptyString(
			stringCandidateValue(hit["objectID"]),
			stringCandidateValue(hit["id"]),
			stringCandidateValue(hit["sku"]),
		)
		if id == "" {
			continue
		}
		title := firstNonEmptyString(
			stringCandidateValue(hit["title"]),
			stringCandidateValue(hit["name"]),
			stringCandidateValue(hit["product_name"]),
		)
		link := firstNonEmptyString(
			stringCandidateValue(hit["url"]),
			stringCandidateValue(hit["link"]),
			stringCandidateValue(hit["handle"]),
		)
		if link != "" && !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
			link = strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(link, "/")
		}
		candidates = append(candidates, map[string]any{
			"listing_id": "frontline-" + id,
			"title":      title,
			"url":        link,
			"price":      numericCandidateValue(firstNonNil(hit["price"], hit["price_value"], hit["sale_price"])),
			"image":      firstNonEmptyString(stringCandidateValue(hit["image"]), stringCandidateValue(hit["image_url"])),
			"source":     "frontlinehobbies",
			"seller":     "frontlinehobbies.com.au",
		})
	}
	total := payload.NBHits
	if total == 0 {
		total = len(candidates)
	}
	return candidates, total, nil
}

func discoverDoofinderConfig(
	ctx context.Context,
	conn *sql.DB,
	client *http.Client,
	assetURL string,
	fallbackAssetURLs []string,
) (doofinderConfig, bool, string, error) {
	if client == nil {
		client = http.DefaultClient
	}
	candidates := make([]string, 0, len(fallbackAssetURLs)+1)
	candidates = append(candidates, strings.TrimSpace(assetURL))
	for _, value := range fallbackAssetURLs {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			candidates = append(candidates, trimmed)
		}
	}
	var discoveryErr error
	for _, urlValue := range candidates {
		if urlValue == "" {
			continue
		}
		script, fetchErr := fetchProviderAsset(ctx, client, urlValue)
		if fetchErr != nil {
			discoveryErr = fetchErr
			continue
		}
		config, parseErr := parseDoofinderConfig(script, urlValue)
		if parseErr != nil {
			discoveryErr = parseErr
			continue
		}
		if persistErr := persistDoofinderCache(ctx, conn, config); persistErr != nil {
			discoveryErr = persistErr
			continue
		}
		return config, false, "", nil
	}
	cached, cacheErr := readDoofinderCache(ctx, conn)
	if cacheErr == nil {
		warning := "doofinder discovery fallback used cached config"
		if discoveryErr != nil {
			warning = fmt.Sprintf("%s: %s", warning, strings.TrimSpace(discoveryErr.Error()))
		}
		return cached, true, warning, nil
	}
	if discoveryErr != nil {
		return doofinderConfig{}, false, "", discoveryErr
	}
	return doofinderConfig{}, false, "", cacheErr
}

func parseDoofinderConfig(script, source string) (doofinderConfig, error) {
	patterns := map[string]*regexp.Regexp{
		"store":  regexp.MustCompile(`(?i)store\s*[:=]\s*['"]([^'"]+)['"]`),
		"zone":   regexp.MustCompile(`(?i)zone\s*[:=]\s*['"]([^'"]+)['"]`),
		"hashid": regexp.MustCompile(`(?i)hashid\s*[:=]\s*['"]([^'"]+)['"]`),
	}
	values := map[string]string{}
	for key, pattern := range patterns {
		match := pattern.FindStringSubmatch(script)
		if len(match) >= 2 {
			values[key] = strings.TrimSpace(match[1])
		}
	}
	if values["store"] == "" || values["zone"] == "" || values["hashid"] == "" {
		return doofinderConfig{}, fmt.Errorf("missing doofinder store/zone/hashid config")
	}
	apiBase := "https://" + values["zone"] + "-search.doofinder.com"
	hash := sha256.Sum256([]byte(script))
	return doofinderConfig{
		Store:        values["store"],
		Zone:         values["zone"],
		HashID:       values["hashid"],
		APIBase:      apiBase,
		Source:       strings.TrimSpace(source),
		ConfigHash:   fmt.Sprintf("%x", hash[:]),
		DiscoveredAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func persistDoofinderCache(ctx context.Context, conn *sql.DB, config doofinderConfig) error {
	raw, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal doofinder cache config: %w", err)
	}
	_, execErr := conn.ExecContext(ctx, `
		INSERT INTO app_state(key, value, updated_at)
		VALUES(?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP
	`, doofinderCacheKey, string(raw))
	if execErr != nil {
		return fmt.Errorf("persist doofinder cache config: %w", execErr)
	}
	return nil
}

func readDoofinderCache(ctx context.Context, conn *sql.DB) (doofinderConfig, error) {
	var raw string
	err := conn.QueryRowContext(ctx, `SELECT value FROM app_state WHERE key = ?`, doofinderCacheKey).Scan(&raw)
	if err != nil {
		return doofinderConfig{}, fmt.Errorf("load doofinder cache config: %w", err)
	}
	var config doofinderConfig
	if decodeErr := json.Unmarshal([]byte(raw), &config); decodeErr != nil {
		return doofinderConfig{}, fmt.Errorf("decode doofinder cache config: %w", decodeErr)
	}
	if strings.TrimSpace(config.Store) == "" || strings.TrimSpace(config.Zone) == "" || strings.TrimSpace(config.HashID) == "" {
		return doofinderConfig{}, fmt.Errorf("doofinder cache config is incomplete")
	}
	return config, nil
}

func runDoofinderSearch(
	ctx context.Context,
	client *http.Client,
	searchURL, query string,
	page, pageSize int,
	hashID, baseURL, providerDomain string,
) ([]map[string]any, int, error) {
	if client == nil {
		client = http.DefaultClient
	}
	if strings.TrimSpace(searchURL) == "" {
		return nil, 0, fmt.Errorf("doofinder search url is required")
	}
	searchParsed, err := url.Parse(strings.TrimSpace(searchURL))
	if err != nil {
		return nil, 0, fmt.Errorf("parse doofinder search url: %w", err)
	}
	params := searchParsed.Query()
	params.Set("hashid", strings.TrimSpace(hashID))
	params.Set("query", strings.TrimSpace(query))
	params.Set("page", strconv.Itoa(page))
	params.Set("rpp", strconv.Itoa(pageSize))
	searchParsed.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchParsed.String(), nil)
	if err != nil {
		return nil, 0, fmt.Errorf("build doofinder search request: %w", err)
	}
	origin := strings.TrimSpace(baseURL)
	if origin == "" {
		origin = "https://" + strings.TrimSpace(providerDomain)
	}
	origin = strings.TrimRight(origin, "/")
	if strings.TrimSpace(providerDomain) == "" {
		if parsed, parseErr := url.Parse(origin); parseErr == nil {
			providerDomain = parsed.Hostname()
		}
	}
	req.Header.Set("Origin", origin)
	req.Header.Set("Referer", origin+"/")
	req.Header.Set("User-Agent", "Cabinet/1.0 (+https://collectors.tech/cabinet)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("run doofinder search request: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		resp.Body.Close()
		return nil, 0, fmt.Errorf("doofinder search returned status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, 0, fmt.Errorf("read doofinder search response: %w", err)
	}
	var payload struct {
		Results []map[string]any `json:"results"`
		Meta    struct {
			Total int `json:"total"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, 0, fmt.Errorf("decode doofinder response: %w", err)
	}
	if providerDomain == "" {
		providerDomain = "unknown"
	}
	candidates := make([]map[string]any, 0, len(payload.Results))
	for _, result := range payload.Results {
		id := strings.TrimSpace(fmt.Sprintf("%v", result["id"]))
		if id == "" {
			continue
		}
		title := strings.TrimSpace(fmt.Sprintf("%v", result["title"]))
		link := strings.TrimSpace(fmt.Sprintf("%v", result["link"]))
		price := strings.TrimSpace(fmt.Sprintf("%v", result["price"]))
		image := strings.TrimSpace(fmt.Sprintf("%v", result["image_link"]))
		candidates = append(candidates, map[string]any{
			"listing_id": "doofinder-" + id,
			"title":      title,
			"url":        link,
			"price":      price,
			"image":      image,
			"source":     providerDomain,
			"seller":     providerDomain,
		})
	}
	return candidates, payload.Meta.Total, nil
}

func runBigCommerceStorefrontSearch(
	ctx context.Context,
	client *http.Client,
	searchURL, query string,
	page, pageSize int,
	providerDomain string,
) ([]map[string]any, error) {
	if client == nil {
		client = http.DefaultClient
	}
	if strings.TrimSpace(searchURL) == "" {
		return nil, fmt.Errorf("bigcommerce storefront search url is required")
	}
	parsed, err := url.Parse(strings.TrimSpace(searchURL))
	if err != nil {
		return nil, fmt.Errorf("parse bigcommerce storefront url: %w", err)
	}
	params := parsed.Query()
	params.Set("q", strings.TrimSpace(query))
	params.Set("page", strconv.Itoa(page))
	params.Set("limit", strconv.Itoa(pageSize))
	parsed.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build bigcommerce storefront request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("run bigcommerce storefront request: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		resp.Body.Close()
		return nil, fmt.Errorf("bigcommerce storefront returned status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("read bigcommerce storefront response: %w", err)
	}
	var payload struct {
		Products []struct {
			ID    any    `json:"id"`
			Name  string `json:"name"`
			URL   string `json:"url"`
			Price any    `json:"price"`
			Image string `json:"image"`
		} `json:"products"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode bigcommerce storefront response: %w", err)
	}
	candidates := make([]map[string]any, 0, len(payload.Products))
	for _, product := range payload.Products {
		id := strings.TrimSpace(fmt.Sprintf("%v", product.ID))
		if id == "" {
			continue
		}
		candidates = append(candidates, map[string]any{
			"listing_id": "bigcommerce-" + id,
			"title":      strings.TrimSpace(product.Name),
			"url":        strings.TrimSpace(product.URL),
			"price":      strings.TrimSpace(fmt.Sprintf("%v", product.Price)),
			"image":      strings.TrimSpace(product.Image),
			"source":     providerDomain,
			"seller":     providerDomain,
		})
	}
	return candidates, nil
}

func runBigCommerceTokenSearch(
	ctx context.Context,
	client *http.Client,
	graphURL, token, query, providerDomain string,
) ([]map[string]any, error) {
	if client == nil {
		client = http.DefaultClient
	}
	if strings.TrimSpace(graphURL) == "" {
		return nil, fmt.Errorf("bigcommerce graphql url is required")
	}
	requestBody := fmt.Sprintf(`{"query":"query CabinetSearch { site { search { searchProducts(filters:{searchTerm:\\\"%s\\\"}) { products { entityId name path prices { price { value } } inventory { isInStock aggregated { availableToSell } } } } } } }"}`, strings.ReplaceAll(strings.TrimSpace(query), `"`, `\"`))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimSpace(graphURL), strings.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("build bigcommerce token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Token", strings.TrimSpace(token))
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("run bigcommerce token request: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		resp.Body.Close()
		return nil, fmt.Errorf("bigcommerce token mode returned status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("read bigcommerce token response: %w", err)
	}
	var payload struct {
		Data struct {
			Site struct {
				Search struct {
					SearchProducts struct {
						Products []struct {
							EntityID int    `json:"entityId"`
							Name     string `json:"name"`
							Path     string `json:"path"`
							Prices   struct {
								Price struct {
									Value any `json:"value"`
								} `json:"price"`
							} `json:"prices"`
							Inventory struct {
								IsInStock  bool `json:"isInStock"`
								Aggregated struct {
									AvailableToSell int `json:"availableToSell"`
								} `json:"aggregated"`
							} `json:"inventory"`
						} `json:"products"`
					} `json:"searchProducts"`
				} `json:"search"`
			} `json:"site"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode bigcommerce token response: %w", err)
	}
	candidates := make([]map[string]any, 0, len(payload.Data.Site.Search.SearchProducts.Products))
	for _, product := range payload.Data.Site.Search.SearchProducts.Products {
		if product.EntityID == 0 {
			continue
		}
		urlValue := strings.TrimSpace(product.Path)
		if strings.HasPrefix(urlValue, "/") {
			urlValue = "https://" + providerDomain + urlValue
		}
		candidates = append(candidates, map[string]any{
			"listing_id":  fmt.Sprintf("bigcommerce-%d", product.EntityID),
			"title":       strings.TrimSpace(product.Name),
			"url":         urlValue,
			"price":       strings.TrimSpace(fmt.Sprintf("%v", product.Prices.Price.Value)),
			"stock_state": map[bool]string{true: "in_stock", false: "out_of_stock"}[product.Inventory.IsInStock],
			"stock_count": product.Inventory.Aggregated.AvailableToSell,
			"source":      providerDomain,
			"seller":      providerDomain,
		})
	}
	return candidates, nil
}

func discoverHobbytechBoostConfig(
	ctx context.Context,
	conn *sql.DB,
	client *http.Client,
	assetURL string,
	fallbackAssetURLs []string,
	searchURL string,
) (hobbytechBoostConfig, bool, string, error) {
	if client == nil {
		client = http.DefaultClient
	}
	candidates := make([]string, 0, len(fallbackAssetURLs)+1)
	candidates = append(candidates, strings.TrimSpace(assetURL))
	for _, value := range fallbackAssetURLs {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			candidates = append(candidates, trimmed)
		}
	}
	var discoveryErr error
	for _, urlValue := range candidates {
		if urlValue == "" {
			continue
		}
		script, fetchErr := fetchProviderAsset(ctx, client, urlValue)
		if fetchErr != nil {
			discoveryErr = fetchErr
			continue
		}
		config, parseErr := parseHobbytechBoostConfig(script, urlValue, searchURL)
		if parseErr != nil {
			discoveryErr = parseErr
			continue
		}
		if persistErr := persistHobbytechBoostCache(ctx, conn, config); persistErr != nil {
			discoveryErr = persistErr
			continue
		}
		return config, false, "", nil
	}
	cached, cacheErr := readHobbytechBoostCache(ctx, conn)
	if cacheErr == nil {
		if strings.TrimSpace(searchURL) != "" {
			cached.SearchURL = strings.TrimSpace(searchURL)
		}
		warning := "hobbytech discovery fallback used cached config"
		if discoveryErr != nil {
			warning = fmt.Sprintf("%s: %s", warning, strings.TrimSpace(discoveryErr.Error()))
		}
		return cached, true, warning, nil
	}
	if discoveryErr != nil {
		return hobbytechBoostConfig{}, false, "", discoveryErr
	}
	return hobbytechBoostConfig{}, false, "", cacheErr
}

func parseHobbytechBoostConfig(script, source, searchURL string) (hobbytechBoostConfig, error) {
	patternByKey := map[string]*regexp.Regexp{
		"shop":     regexp.MustCompile(`(?i)shop\s*[:=]\s*['"]([^'"]+)['"]`),
		"sid":      regexp.MustCompile(`(?i)sid\s*[:=]\s*['"]([^'"]+)['"]`),
		"template": regexp.MustCompile(`(?i)template\s*[:=]\s*['"]([^'"]+)['"]`),
		"widget":   regexp.MustCompile(`(?i)widget\s*[:=]\s*['"]([^'"]+)['"]`),
	}
	values := map[string]string{}
	for key, pattern := range patternByKey {
		match := pattern.FindStringSubmatch(script)
		if len(match) >= 2 {
			values[key] = strings.TrimSpace(match[1])
		}
	}
	if values["shop"] == "" || values["sid"] == "" {
		return hobbytechBoostConfig{}, fmt.Errorf("missing hobbytech discovery values")
	}
	resolvedSearchURL := strings.TrimSpace(searchURL)
	if resolvedSearchURL == "" {
		urlMatch := regexp.MustCompile(`(?i)https?://[^'"\s]*mybcapps[^'"\s]*`).FindString(script)
		resolvedSearchURL = strings.TrimSpace(urlMatch)
	}
	if resolvedSearchURL == "" {
		return hobbytechBoostConfig{}, fmt.Errorf("missing hobbytech search endpoint")
	}
	hash := sha256.Sum256([]byte(script))
	return hobbytechBoostConfig{
		Shop:         values["shop"],
		SID:          values["sid"],
		Template:     values["template"],
		Widget:       values["widget"],
		SearchURL:    resolvedSearchURL,
		Source:       strings.TrimSpace(source),
		ConfigHash:   fmt.Sprintf("%x", hash[:]),
		DiscoveredAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func persistHobbytechBoostCache(ctx context.Context, conn *sql.DB, config hobbytechBoostConfig) error {
	raw, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal hobbytech cache config: %w", err)
	}
	_, execErr := conn.ExecContext(ctx, `
		INSERT INTO app_state(key, value, updated_at)
		VALUES(?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP
	`, hobbytechBoostCacheKey, string(raw))
	if execErr != nil {
		return fmt.Errorf("persist hobbytech cache config: %w", execErr)
	}
	return nil
}

func readHobbytechBoostCache(ctx context.Context, conn *sql.DB) (hobbytechBoostConfig, error) {
	var raw string
	err := conn.QueryRowContext(ctx, `SELECT value FROM app_state WHERE key = ?`, hobbytechBoostCacheKey).Scan(&raw)
	if err != nil {
		return hobbytechBoostConfig{}, fmt.Errorf("load hobbytech cache config: %w", err)
	}
	var config hobbytechBoostConfig
	if decodeErr := json.Unmarshal([]byte(raw), &config); decodeErr != nil {
		return hobbytechBoostConfig{}, fmt.Errorf("decode hobbytech cache config: %w", decodeErr)
	}
	if strings.TrimSpace(config.Shop) == "" || strings.TrimSpace(config.SID) == "" || strings.TrimSpace(config.SearchURL) == "" {
		return hobbytechBoostConfig{}, fmt.Errorf("hobbytech cache config is incomplete")
	}
	return config, nil
}

func runHobbytechSearch(
	ctx context.Context,
	client *http.Client,
	qs scanner.QuerySet,
	config hobbytechBoostConfig,
	itemsPerPage int,
) ([]map[string]any, int, error) {
	if client == nil {
		client = http.DefaultClient
	}
	if strings.TrimSpace(config.SearchURL) == "" {
		return nil, 0, fmt.Errorf("hobbytech search endpoint missing")
	}
	if itemsPerPage <= 0 {
		itemsPerPage = 24
	}
	query := strings.TrimSpace(qs.Name)
	if len(qs.Keywords) > 0 && strings.TrimSpace(qs.Keywords[0]) != "" {
		query = strings.TrimSpace(qs.Keywords[0])
	}
	if query == "" {
		query = "collectible"
	}
	candidateByID := map[string]map[string]any{}
	pageCount := 0
	totalPages := 0
	for page := 1; ; page++ {
		searchURL, err := url.Parse(strings.TrimSpace(config.SearchURL))
		if err != nil {
			return nil, 0, fmt.Errorf("parse hobbytech search url: %w", err)
		}
		params := searchURL.Query()
		params.Set("q", query)
		params.Set("page", strconv.Itoa(page))
		params.Set("limit", strconv.Itoa(itemsPerPage))
		params.Set("shop", config.Shop)
		params.Set("sid", config.SID)
		if strings.TrimSpace(config.Template) != "" {
			params.Set("template", config.Template)
		}
		if strings.TrimSpace(config.Widget) != "" {
			params.Set("widget", config.Widget)
		}
		searchURL.RawQuery = params.Encode()
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, searchURL.String(), nil)
		if reqErr != nil {
			return nil, 0, fmt.Errorf("build hobbytech search request: %w", reqErr)
		}
		resp, runErr := client.Do(req)
		if runErr != nil {
			return nil, 0, fmt.Errorf("run hobbytech search request: %w", runErr)
		}
		if resp.StatusCode >= http.StatusBadRequest {
			resp.Body.Close()
			return nil, pageCount, fmt.Errorf("hobbytech search returned status %d", resp.StatusCode)
		}
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, pageCount, fmt.Errorf("read hobbytech search response: %w", readErr)
		}
		var payload struct {
			Products []struct {
				ID    any    `json:"id"`
				Title string `json:"title"`
				URL   string `json:"url"`
				Price string `json:"price"`
				Image string `json:"image"`
			} `json:"products"`
			Pagination struct {
				TotalPages int `json:"total_pages"`
			} `json:"pagination"`
			HTML string `json:"html"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, pageCount, fmt.Errorf("decode hobbytech response: %w", err)
		}
		if totalPages == 0 && payload.Pagination.TotalPages > 0 {
			totalPages = payload.Pagination.TotalPages
		}
		pageCount++
		for _, product := range payload.Products {
			listingID := fmt.Sprintf("hobbytech-%v", product.ID)
			if strings.TrimSpace(listingID) == "" || strings.EqualFold(listingID, "hobbytech-<nil>") {
				continue
			}
			if _, exists := candidateByID[listingID]; exists {
				continue
			}
			candidateByID[listingID] = map[string]any{
				"listing_id": listingID,
				"title":      strings.TrimSpace(product.Title),
				"url":        strings.TrimSpace(product.URL),
				"price":      strings.TrimSpace(product.Price),
				"image":      strings.TrimSpace(product.Image),
				"source":     "hobbytechtoys",
				"seller":     "hobbytechtoys.com.au",
			}
		}
		if len(payload.Products) == 0 {
			break
		}
		if totalPages > 0 && page >= totalPages {
			break
		}
	}
	candidates := make([]map[string]any, 0, len(candidateByID))
	for _, candidate := range candidateByID {
		candidates = append(candidates, candidate)
	}
	return candidates, pageCount, nil
}

func runBonzaSearch(ctx context.Context, client *http.Client, baseURL string, qs scanner.QuerySet, itemsPerPage int) (bonzaSearchResult, error) {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return bonzaSearchResult{}, fmt.Errorf("bonza base_url is required")
	}
	if client == nil {
		client = http.DefaultClient
	}
	if itemsPerPage <= 0 {
		itemsPerPage = 36
	}
	if itemsPerPage > 36 {
		itemsPerPage = 36
	}

	searchTerm := ""
	if len(qs.Keywords) > 0 {
		searchTerm = strings.TrimSpace(qs.Keywords[0])
	}
	if searchTerm == "" {
		searchTerm = strings.TrimSpace(qs.Name)
	}

	type aggregate struct {
		order int
		data  map[string]any
	}
	candidateByID := map[string]aggregate{}
	order := []string{}
	pageCount := 0
	observedPageSize := 0
	totalPages := 0

	for page := 1; ; page++ {
		requestURL := fmt.Sprintf("%s/wp-json/wc/store/v1/products?search=%s&per_page=%d&page=%d",
			strings.TrimRight(baseURL, "/"),
			url.QueryEscape(searchTerm),
			itemsPerPage,
			page,
		)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
		if err != nil {
			return bonzaSearchResult{}, fmt.Errorf("build bonza request: %w", err)
		}
		resp, err := client.Do(req)
		if err != nil {
			return bonzaSearchResult{}, fmt.Errorf("request bonza products: %w", err)
		}
		if resp.StatusCode >= http.StatusBadRequest {
			resp.Body.Close()
			return bonzaSearchResult{}, fmt.Errorf("bonza products returned status %d", resp.StatusCode)
		}
		if totalPages == 0 {
			totalPages = parsePositiveInt(resp.Header.Get("X-WP-TotalPages"), 0)
		}
		var products []bonzaProductResponse
		decodeErr := json.NewDecoder(resp.Body).Decode(&products)
		resp.Body.Close()
		if decodeErr != nil {
			return bonzaSearchResult{}, fmt.Errorf("decode bonza products response: %w", decodeErr)
		}
		if len(products) == 0 {
			break
		}
		pageCount++
		if observedPageSize == 0 {
			observedPageSize = len(products)
		}
		for _, product := range products {
			listingID := fmt.Sprintf("bonza-%d", product.ID)
			if strings.TrimSpace(listingID) == "" {
				continue
			}
			permalink := normalizeBonzaURL(baseURL, strings.TrimSpace(product.Permalink))
			rawStock, stockState, stockCount := deriveBonzaStockSignal(ctx, client, permalink, product)
			if existing, exists := candidateByID[listingID]; exists {
				if strings.EqualFold(strings.TrimSpace(existing.data["stock_state"].(string)), "unknown") && !strings.EqualFold(stockState, "unknown") {
					existing.data["stock_state"] = stockState
					existing.data["stock_count"] = stockCount
					existing.data["stock_raw"] = rawStock
					candidateByID[listingID] = existing
				}
				continue
			}
			candidate := map[string]any{
				"listing_id":  listingID,
				"title":       strings.TrimSpace(product.Name),
				"url":         permalink,
				"price":       parseWooCommerceMinorUnitPrice(product.Prices.Price),
				"currency":    strings.TrimSpace(product.Prices.CurrencyCode),
				"image":       firstBonzaImageURL(product.Images),
				"category":    firstBonzaCategoryName(product.Categories),
				"categories":  bonzaCategoryNames(product.Categories),
				"source":      "bonzaslotcars",
				"seller":      "bonzaslotcars.com.au",
				"stock_state": stockState,
				"stock_count": stockCount,
				"stock_raw":   rawStock,
			}
			candidateByID[listingID] = aggregate{order: len(order), data: candidate}
			order = append(order, listingID)
		}
		if totalPages > 0 {
			if page >= totalPages {
				break
			}
		} else if len(products) < itemsPerPage {
			break
		}
	}

	candidates := make([]map[string]any, 0, len(order))
	for _, listingID := range order {
		candidates = append(candidates, candidateByID[listingID].data)
	}
	return bonzaSearchResult{
		PageCount:        pageCount,
		ObservedPageSize: observedPageSize,
		ItemsPerPageUsed: itemsPerPage,
		Candidates:       candidates,
	}, nil
}

func detectProviderProductURL(raw string) (providerProductURLRoute, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return providerProductURLRoute{}, fmt.Errorf("url is required")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Host == "" {
		return providerProductURLRoute{}, fmt.Errorf("invalid provider url")
	}
	host := strings.ToLower(strings.TrimPrefix(parsed.Hostname(), "www."))
	pathValue := "/" + strings.Trim(strings.TrimSpace(parsed.EscapedPath()), "/")
	if pathValue == "/" {
		pathValue = "/"
	}
	route := providerProductURLRoute{
		OriginalURL: strings.TrimSpace(raw),
		Host:        host,
		Path:        pathValue,
	}
	switch host {
	case "bonzaslotcars.com.au":
		route.Provider = "bonzaslotcars"
		route.Family = "woocommerce"
	default:
		return providerProductURLRoute{}, fmt.Errorf("unsupported provider host")
	}
	parts := strings.Split(strings.Trim(pathValue, "/"), "/")
	if len(parts) >= 2 && strings.EqualFold(parts[0], "product") && strings.TrimSpace(parts[1]) != "" {
		route.Slug = strings.TrimSpace(parts[1])
		route.Action = "ingest_product_url"
		route.NormalizedURL = "https://" + host + "/product/" + route.Slug + "/"
		return route, nil
	}
	route.Action = "unsupported_page"
	route.NormalizedURL = "https://" + host + pathValue
	if !strings.HasSuffix(route.NormalizedURL, "/") {
		route.NormalizedURL += "/"
	}
	return route, nil
}

func ingestBonzaProductURL(ctx context.Context, client *http.Client, baseURL string, route providerProductURLRoute) (providerProductDraft, error) {
	if client == nil {
		client = http.DefaultClient
	}
	search := strings.ReplaceAll(strings.TrimSpace(route.Slug), "-", " ")
	requestURL := fmt.Sprintf("%s/wp-json/wc/store/v1/products?search=%s&per_page=5",
		strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		url.QueryEscape(search),
	)
	products, err := fetchBonzaProductURLProducts(ctx, client, requestURL)
	if err != nil {
		return providerProductDraft{}, err
	}
	for _, product := range products {
		if !bonzaProductMatchesRoute(product, route) {
			continue
		}
		return bonzaProductDraft(product, route), nil
	}
	return providerProductDraft{}, fmt.Errorf("bonza product %q not found", route.Slug)
}

func fetchBonzaProductURLProducts(ctx context.Context, client *http.Client, requestURL string) ([]bonzaProductResponse, error) {
	if client == nil {
		client = http.DefaultClient
	}
	requestClient := *client
	requestClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := fetchBonzaProductURLProductsResponse(ctx, &requestClient, requestURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		resp.Body.Close()
		return nil, fmt.Errorf("bonza product ingest returned status %d", resp.StatusCode)
	}
	var products []bonzaProductResponse
	decodeErr := json.NewDecoder(resp.Body).Decode(&products)
	resp.Body.Close()
	if decodeErr != nil {
		return nil, fmt.Errorf("decode bonza product ingest response: %w", decodeErr)
	}
	return products, nil
}

func fetchBonzaProductURLProductsResponse(ctx context.Context, client *http.Client, requestURL string) (*http.Response, error) {
	const maxChallengeRetries = 5
	cookies := map[string]string{}
	var lastChallengeErr error

	for attempt := 0; attempt <= maxChallengeRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
		if err != nil {
			return nil, fmt.Errorf("build bonza product ingest request: %w", err)
		}
		applyBonzaProductURLRequestHeaders(req, cookies)
		resp, err := client.Do(req)
		if err != nil {
			if attempt == 0 {
				return nil, fmt.Errorf("request bonza product ingest: %w", err)
			}
			return nil, fmt.Errorf("retry bonza product ingest after challenge: %w", err)
		}
		challengeBody, readChallenge := readBonzaSucuriChallenge(resp)
		if !readChallenge {
			return resp, nil
		}
		cookie, cookieErr := bonzaSucuriChallengeCookie(challengeBody)
		if cookieErr != nil || cookie == "" {
			if cookieErr == nil {
				cookieErr = fmt.Errorf("empty challenge cookie")
			}
			return nil, fmt.Errorf("solve bonza product ingest challenge: %w", cookieErr)
		}
		name, value, ok := strings.Cut(cookie, "=")
		if !ok || strings.TrimSpace(name) == "" || strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("solve bonza product ingest challenge: invalid cookie")
		}
		cookies[strings.TrimSpace(name)] = strings.TrimSpace(value)
		lastChallengeErr = fmt.Errorf("bonza product ingest challenge did not clear after %d attempt(s)", attempt+1)
	}
	return nil, lastChallengeErr
}

func applyBonzaProductURLRequestHeaders(req *http.Request, cookies map[string]string) {
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-AU,en;q=0.9")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36 Cabinet/1.0")
	if req.URL != nil {
		req.Header.Set("Referer", strings.TrimRight(req.URL.Scheme+"://"+req.URL.Host, "/")+"/")
	}
	if len(cookies) > 0 {
		names := make([]string, 0, len(cookies))
		for name := range cookies {
			names = append(names, name)
		}
		sort.Strings(names)
		parts := make([]string, 0, len(names))
		for _, name := range names {
			parts = append(parts, name+"="+cookies[name])
		}
		req.Header.Set("Cookie", strings.Join(parts, "; "))
	}
}

func readBonzaSucuriChallenge(resp *http.Response) (string, bool) {
	if resp == nil || resp.Body == nil || resp.StatusCode < http.StatusMultipleChoices || resp.StatusCode >= http.StatusBadRequest {
		return "", false
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", false
	}
	content := string(body)
	return content, strings.Contains(content, "sucuri_cloudproxy_js") && strings.Contains(content, "S='")
}

func bonzaSucuriChallengeCookie(body string) (string, error) {
	scriptMatch := regexp.MustCompile(`S='([^']+)'`).FindStringSubmatch(body)
	if len(scriptMatch) != 2 {
		return "", fmt.Errorf("missing challenge script")
	}
	decoded, err := base64.StdEncoding.DecodeString(scriptMatch[1])
	if err != nil {
		return "", fmt.Errorf("decode challenge script: %w", err)
	}
	challenge := string(decoded)
	assignment := regexp.MustCompile(`^([A-Za-z])=(.+?);document\.cookie=(.+)$`).FindStringSubmatch(challenge)
	if len(assignment) != 4 {
		return "", fmt.Errorf("unsupported challenge assignment")
	}
	value, err := evalSucuriConcatExpression(assignment[2])
	if err != nil {
		return "", fmt.Errorf("decode challenge value: %w", err)
	}
	cookiePattern := regexp.MustCompile(`\+\s*["']=["']\s*\+\s*` + regexp.QuoteMeta(assignment[1]) + `\s*\+`)
	cookieParts := cookiePattern.Split(assignment[3], 2)
	if len(cookieParts) != 2 {
		return "", fmt.Errorf("unsupported challenge cookie expression")
	}
	name, err := evalSucuriConcatExpression(cookieParts[0])
	if err != nil {
		return "", fmt.Errorf("decode challenge cookie name: %w", err)
	}
	if name == "" || value == "" {
		return "", fmt.Errorf("empty challenge cookie")
	}
	return name + "=" + value, nil
}

func evalSucuriConcatExpression(expr string) (string, error) {
	parts := strings.Split(expr, "+")
	var out strings.Builder
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		if len(trimmed) >= 2 && ((trimmed[0] == '\'' && trimmed[len(trimmed)-1] == '\'') || (trimmed[0] == '"' && trimmed[len(trimmed)-1] == '"')) {
			out.WriteString(trimmed[1 : len(trimmed)-1])
			continue
		}
		match := regexp.MustCompile(`^String\.fromCharCode\((\d+)\)$`).FindStringSubmatch(trimmed)
		if len(match) == 2 {
			codepoint, err := strconv.Atoi(match[1])
			if err != nil {
				return "", err
			}
			out.WriteRune(rune(codepoint))
			continue
		}
		return "", fmt.Errorf("unsupported concat token %q", trimmed)
	}
	return out.String(), nil
}

func bonzaProductMatchesRoute(product bonzaProductResponse, route providerProductURLRoute) bool {
	if strings.EqualFold(strings.TrimSpace(product.Slug), strings.TrimSpace(route.Slug)) {
		return true
	}
	permalink := strings.TrimRight(strings.ToLower(strings.TrimSpace(product.Permalink)), "/")
	normalized := strings.TrimRight(strings.ToLower(strings.TrimSpace(route.NormalizedURL)), "/")
	return permalink != "" && permalink == normalized
}

func bonzaProductDraft(product bonzaProductResponse, route providerProductURLRoute) providerProductDraft {
	rawStock, stockState, stockCount := deriveBonzaStockSignal(context.Background(), nil, "", product)
	_ = rawStock
	if product.IsInStock != nil && *product.IsInStock && stockCount > 0 {
		stockState = "in_stock"
	}
	categories := make([]string, 0, len(product.Categories))
	for _, category := range product.Categories {
		if value := strings.TrimSpace(category.Name); value != "" {
			categories = append(categories, value)
		}
	}
	attributes := map[string]string{}
	for _, attribute := range product.Attributes {
		name := strings.TrimSpace(attribute.Name)
		if name == "" {
			continue
		}
		values := append([]string{}, attribute.Terms...)
		values = append(values, attribute.Options...)
		for _, value := range values {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				attributes[name] = trimmed
				break
			}
		}
	}
	images := make([]string, 0, len(product.Images))
	for _, image := range product.Images {
		if src := strings.TrimSpace(image.Src); src != "" {
			images = append(images, src)
		}
	}
	observedAt := time.Now().UTC().Format(time.RFC3339)
	return providerProductDraft{
		ProviderProductID: strconv.Itoa(product.ID),
		Title:             strings.TrimSpace(product.Name),
		SourceURL:         route.NormalizedURL,
		Description:       stripHTMLText(product.Description),
		Price:             parseWooCommerceMinorUnitPrice(product.Prices.Price),
		Currency:          strings.TrimSpace(product.Prices.CurrencyCode),
		StockState:        stockState,
		StockCount:        stockCount,
		Categories:        categories,
		Attributes:        attributes,
		ImageURLs:         images,
		Evidence: map[string]any{
			"provider":            "bonzaslotcars",
			"family":              "woocommerce",
			"extraction_method":   "store_api",
			"provider_product_id": strconv.Itoa(product.ID),
			"original_url":        route.OriginalURL,
			"normalized_url":      route.NormalizedURL,
			"observed_at":         observedAt,
			"source_summary":      "WooCommerce Store API product detail",
		},
	}
}

type providerProductDuplicateCandidate struct {
	ItemID     string   `json:"item_id"`
	Title      string   `json:"title"`
	SourceURLs []string `json:"source_urls"`
	Reasons    []string `json:"reasons"`
}

func providerProductDuplicateCandidates(items []collection.Item, route providerProductURLRoute, draft providerProductDraft) []providerProductDuplicateCandidate {
	normalizedSource := strings.TrimRight(strings.ToLower(strings.TrimSpace(firstNonEmptyString(draft.SourceURL, route.NormalizedURL))), "/")
	providerProductID := strings.TrimSpace(draft.ProviderProductID)
	out := make([]providerProductDuplicateCandidate, 0)
	for _, item := range items {
		reasons := make([]string, 0, 2)
		for _, sourceURL := range item.SourceURLs {
			if normalizedSource != "" && strings.TrimRight(strings.ToLower(strings.TrimSpace(sourceURL)), "/") == normalizedSource {
				reasons = append(reasons, "source_url")
				break
			}
		}
		if providerProductID != "" && strings.Contains(strings.ToLower(item.Notes), strings.ToLower("provider_product_id="+providerProductID)) {
			reasons = append(reasons, "provider_product_id")
		}
		if len(reasons) == 0 {
			continue
		}
		out = append(out, providerProductDuplicateCandidate{
			ItemID:     item.ID,
			Title:      item.Title,
			SourceURLs: item.SourceURLs,
			Reasons:    reasons,
		})
	}
	if out == nil {
		return []providerProductDuplicateCandidate{}
	}
	return out
}

func parseWooCommerceMinorUnitPrice(raw string) float64 {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0
	}
	if strings.Contains(value, ".") {
		parsed, _ := strconv.ParseFloat(value, 64)
		return parsed
	}
	cents, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return cents / 100
}

func stripHTMLText(raw string) string {
	withoutTags := regexp.MustCompile(`<[^>]*>`).ReplaceAllString(raw, " ")
	return strings.Join(strings.Fields(withoutTags), " ")
}

func bonzaCandidatesForScanner(candidates []map[string]any) []scanner.CandidateInput {
	return providerCandidatesForScanner(candidates, "bonzaslotcars")
}

func hobbytechCandidatesForScanner(candidates []map[string]any) []scanner.CandidateInput {
	return providerCandidatesForScanner(candidates, "hobbytechtoys")
}

func frontlineCandidatesForScanner(candidates []map[string]any) []scanner.CandidateInput {
	return providerCandidatesForScanner(candidates, "frontlinehobbies")
}

func doofinderCandidatesForScanner(candidates []map[string]any, providerDomain string) []scanner.CandidateInput {
	return providerCandidatesForScanner(candidates, providerDomain)
}

func bigCommerceCandidatesForScanner(candidates []map[string]any, providerDomain string) []scanner.CandidateInput {
	return providerCandidatesForScanner(candidates, providerDomain)
}

func providerCandidatesForScanner(candidates []map[string]any, defaultSource string) []scanner.CandidateInput {
	out := make([]scanner.CandidateInput, 0, len(candidates))
	for _, candidate := range candidates {
		listingID := strings.TrimSpace(fmt.Sprint(candidate["listing_id"]))
		title := stringCandidateValue(candidate["title"])
		sourceURL := stringCandidateValue(candidate["url"])
		if listingID == "" || title == "" || sourceURL == "" {
			continue
		}
		out = append(out, scanner.CandidateInput{
			ListingID:  listingID,
			Title:      title,
			Price:      numericCandidateValue(candidate["price"]),
			Currency:   stringCandidateValue(candidate["currency"]),
			URL:        sourceURL,
			Image:      stringCandidateValue(candidate["image"]),
			Seller:     stringCandidateValue(candidate["seller"]),
			Source:     firstNonEmptyString(stringCandidateValue(candidate["source"]), defaultSource),
			StockState: stringCandidateValue(candidate["stock_state"]),
			StockCount: int(numericCandidateValue(candidate["stock_count"])),
		})
	}
	return out
}

func firstBonzaCategoryName(categories []bonzaProductName) string {
	names := bonzaCategoryNames(categories)
	if len(names) == 0 {
		return ""
	}
	return names[0]
}

func bonzaCategoryNames(categories []bonzaProductName) []string {
	out := make([]string, 0, len(categories))
	for _, category := range categories {
		if value := strings.TrimSpace(category.Name); value != "" {
			out = append(out, value)
		}
	}
	return out
}

func firstBonzaImageURL(images []bonzaProductImage) string {
	for _, image := range images {
		if src := strings.TrimSpace(image.Src); src != "" {
			return src
		}
	}
	return ""
}

func normalizeScannerReviewApplyTarget(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "wishlist") {
		return "wishlist"
	}
	return "inventory"
}

func scannerReviewApplyItem(review scanner.RecognitionReview) collection.Item {
	selected := review.SelectedCandidate
	status := "active"
	if review.Target == "wishlist" {
		status = "wishlist"
	}
	return collection.Item{
		Brand:      "Unknown",
		Category:   "Cards",
		ItemType:   "Trading Cards",
		PartNumber: selected.ID,
		Title:      selected.Title,
		Status:     status,
		Notes:      scannerReviewApplyEvidenceNote(review),
		SourceURLs: scannerReviewApplySourceURLs(review),
		Tags:       []string{"scanner-review"},
	}
}

func scannerReviewApplyEvidenceNote(review scanner.RecognitionReview) string {
	selected := review.SelectedCandidate
	values := []string{
		"scanner_review_apply",
		"selected_candidate=" + selected.ID,
		"confidence=" + strconv.FormatFloat(selected.Confidence, 'f', 2, 64),
		"confidence_label=" + review.ConfidenceLabel,
		"target=" + review.Target,
	}
	if review.ManualOverrideApplied {
		values = append(values, "manual_override=true")
	}
	if selected.OverrideNote != "" {
		values = append(values, "override_note="+selected.OverrideNote)
	}
	if mediaID := strings.TrimSpace(review.MediaEvidence["media_id"]); mediaID != "" {
		values = append(values, "media_id="+mediaID)
	} else if mediaID := strings.TrimSpace(review.TopCandidate.MediaID); mediaID != "" {
		values = append(values, "media_id="+mediaID)
	}
	if mediaURL := strings.TrimSpace(review.MediaEvidence["media_url"]); mediaURL != "" {
		values = append(values, "media_url="+mediaURL)
	} else if mediaURL := strings.TrimSpace(review.TopCandidate.MediaURL); mediaURL != "" {
		values = append(values, "media_url="+mediaURL)
	}
	if len(review.Provenance) > 0 {
		values = append(values, "provenance="+strings.Join(review.Provenance, "|"))
	}
	return strings.Join(values, "; ")
}

func scannerReviewApplySourceURLs(review scanner.RecognitionReview) []string {
	if value := strings.TrimSpace(review.MediaEvidence["media_url"]); value != "" {
		return []string{value}
	}
	if value := strings.TrimSpace(review.TopCandidate.MediaURL); value != "" {
		return []string{value}
	}
	return nil
}

func stringCandidateValue(raw any) string {
	if raw == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(raw))
}

func firstNonNil(values ...any) any {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func numericCandidateValue(raw any) float64 {
	switch value := raw.(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int64:
		return float64(value)
	case json.Number:
		parsed, _ := value.Float64()
		return parsed
	case string:
		parsed, _ := strconv.ParseFloat(strings.TrimSpace(value), 64)
		return parsed
	default:
		return 0
	}
}

func normalizeBonzaURL(baseURL, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return strings.TrimRight(baseURL, "/")
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	if strings.HasPrefix(raw, "/") {
		return strings.TrimRight(baseURL, "/") + raw
	}
	baseParsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(raw, "/")
	}
	if strings.Contains(raw, "/") && strings.Contains(raw, ":") && !strings.Contains(raw, "://") {
		return baseParsed.Scheme + "://" + raw
	}
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(raw, "/")
}

func deriveBonzaStockSignal(ctx context.Context, client *http.Client, permalink string, product bonzaProductResponse) (raw, state string, count int) {
	if product.LowStockRemaining != nil {
		if *product.LowStockRemaining <= 0 {
			return "Out of stock", "out_of_stock", 0
		}
		if *product.LowStockRemaining <= 3 {
			return fmt.Sprintf("Only %d left in stock", *product.LowStockRemaining), "low_stock", *product.LowStockRemaining
		}
		return fmt.Sprintf("%d in stock", *product.LowStockRemaining), "in_stock", *product.LowStockRemaining
	}
	if product.IsInStock != nil {
		if *product.IsInStock {
			return "In stock", "in_stock", -1
		}
		return "Out of stock", "out_of_stock", 0
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, permalink, nil)
	if err != nil {
		return "", "unknown", -1
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", "unknown", -1
	}
	if resp.StatusCode >= http.StatusBadRequest {
		resp.Body.Close()
		return "", "unknown", -1
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", "unknown", -1
	}
	return parseStockSignal(string(body))
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

func normalizeProviderDomain(input string) string {
	value := strings.TrimSpace(strings.ToLower(input))
	if value == "" {
		return ""
	}
	if strings.Contains(value, "://") {
		if parsed, err := url.Parse(value); err == nil {
			value = strings.TrimSpace(strings.ToLower(parsed.Hostname()))
		}
	}
	value = strings.TrimPrefix(value, "www.")
	return strings.TrimSpace(value)
}

func persistProviderFamilyOverride(ctx context.Context, conn *sql.DB, domain, family string) error {
	domain = normalizeProviderDomain(domain)
	family = strings.TrimSpace(strings.ToLower(family))
	if domain == "" || family == "" {
		return fmt.Errorf("domain and family are required")
	}
	key := "provider.family.override." + domain
	_, err := conn.ExecContext(ctx, `
		INSERT INTO app_state(key, value, updated_at)
		VALUES(?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP
	`, key, family)
	if err != nil {
		return fmt.Errorf("persist provider family override: %w", err)
	}
	return nil
}

func loadProviderFamilyOverrides(ctx context.Context, conn *sql.DB) (map[string]string, error) {
	rows, err := conn.QueryContext(ctx, `
		SELECT key, value
		FROM app_state
		WHERE key LIKE 'provider.family.override.%'
	`)
	if err != nil {
		return nil, fmt.Errorf("list provider family overrides: %w", err)
	}
	defer rows.Close()
	result := map[string]string{}
	for rows.Next() {
		var key string
		var value string
		if scanErr := rows.Scan(&key, &value); scanErr != nil {
			return nil, fmt.Errorf("scan provider family override: %w", scanErr)
		}
		domain := strings.TrimPrefix(strings.TrimSpace(strings.ToLower(key)), "provider.family.override.")
		domain = normalizeProviderDomain(domain)
		family := strings.TrimSpace(strings.ToLower(value))
		if domain == "" || family == "" {
			continue
		}
		result[domain] = family
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate provider family overrides: %w", err)
	}
	return result, nil
}

func detectProviderFamily(ctx context.Context, client *http.Client, providerURL, htmlInput string) (string, float64, []string, string, error) {
	if client == nil {
		client = http.DefaultClient
	}
	providerURL = strings.TrimSpace(providerURL)
	if providerURL == "" {
		return "", 0, nil, "", fmt.Errorf("provider_url is required")
	}
	parsedURL, err := url.Parse(providerURL)
	if err != nil {
		return "", 0, nil, "", fmt.Errorf("parse provider_url: %w", err)
	}
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}
	domain := normalizeProviderDomain(parsedURL.String())
	html := strings.TrimSpace(htmlInput)
	if html == "" {
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
		if reqErr != nil {
			return "", 0, nil, domain, fmt.Errorf("build detect request: %w", reqErr)
		}
		resp, runErr := client.Do(req)
		if runErr != nil {
			return "", 0, nil, domain, fmt.Errorf("request detect provider_url: %w", runErr)
		}
		if resp.StatusCode >= http.StatusBadRequest {
			resp.Body.Close()
			return "", 0, nil, domain, fmt.Errorf("provider_url returned status %d", resp.StatusCode)
		}
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return "", 0, nil, domain, fmt.Errorf("read provider_url body: %w", readErr)
		}
		html = string(body)
	}
	lower := strings.ToLower(html)
	type familyScore struct {
		name     string
		score    float64
		evidence []string
	}
	candidates := []familyScore{}

	buildEvidence := func(name string, checks map[string]bool) {
		evidence := []string{}
		for marker, hit := range checks {
			if hit {
				evidence = append(evidence, marker)
			}
		}
		if len(evidence) == 0 {
			return
		}
		score := 0.55 + (float64(len(evidence)) * 0.12)
		if score > 0.99 {
			score = 0.99
		}
		candidates = append(candidates, familyScore{name: name, score: score, evidence: evidence})
	}

	buildEvidence("woocommerce", map[string]bool{
		"/wp-json/wc/store/v1": strings.Contains(lower, "/wp-json/wc/store/v1"),
		"woocommerce marker":   strings.Contains(lower, "woocommerce"),
	})
	buildEvidence("boost_shopify", map[string]bool{
		"services.mybcapps.com/bc-sf-filter": strings.Contains(lower, "services.mybcapps.com/bc-sf-filter"),
		"boost script signature":             strings.Contains(lower, "mybcapps"),
	})
	buildEvidence("algolia", map[string]bool{
		"algoliasearch(": strings.Contains(lower, "algoliasearch("),
		"glgoliasearch":  strings.Contains(lower, "glgoliasearch"),
	})
	buildEvidence("shopify_json", map[string]bool{
		"/products.json":               strings.Contains(lower, "/products.json"),
		"/collections/*/products.json": strings.Contains(lower, "/collections/") && strings.Contains(lower, "/products.json"),
	})
	buildEvidence("doofinder", map[string]bool{
		"cdn.doofinder.com":       strings.Contains(lower, "cdn.doofinder.com"),
		"hashid/search_engines":   strings.Contains(lower, "hashid") && strings.Contains(lower, "search_engines"),
		"doofinder loader/config": strings.Contains(lower, "doofinder"),
	})

	if len(candidates) == 0 {
		return "unknown", 0.1, []string{"no_known_markers"}, domain, nil
	}
	best := candidates[0]
	for _, candidate := range candidates[1:] {
		if candidate.score > best.score {
			best = candidate
		}
	}
	return best.name, best.score, best.evidence, domain, nil
}

var (
	buildRevision string
	buildDate     string
)

func runtimeBuildMetadata() (string, string) {
	version := "dev"
	resolvedBuildDate := "unknown"

	if strings.TrimSpace(buildRevision) != "" {
		short := strings.TrimSpace(buildRevision)
		if len(short) > 12 {
			short = short[:12]
		}
		version = "rev-" + short
	}
	if strings.TrimSpace(buildDate) != "" {
		resolvedBuildDate = strings.TrimSpace(buildDate)
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return version, resolvedBuildDate
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
				resolvedBuildDate = strings.TrimSpace(setting.Value)
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

	return version, resolvedBuildDate
}

func (a *App) Run(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return a.close()
	}
	if a.backupSvc != nil {
		a.backupSvc.Start(ctx)
	}
	pid := os.Getpid()
	if err := writeRuntimePIDFile(a.cfg, pid); err != nil {
		return fmt.Errorf("write runtime pid file: %w", err)
	}
	defer func() {
		_ = removeRuntimePIDFile(a.cfg)
	}()
	listener, err := listenWithPortFallback(a.srv.Addr, 50)
	if err != nil {
		_ = a.closeRuntime(false, "listen_error", "")
		return err
	}
	resolvedAddr := listener.Addr().String()
	resolvedURL := startupURLFromResolvedAddr(resolvedAddr)
	_ = syncRuntimeSetupCurrentURLWithURL(a.cfg, resolvedURL)
	_ = syncRuntimeLifecycleStartup(a.cfg, resolvedURL, resolvedAddr, pid)
	if a.runtimeLogs != nil {
		a.runtimeLogs.writeRuntimeEvent("info", "startup", "runtime started", map[string]any{
			"url":           resolvedURL,
			"requestedAddr": strings.TrimSpace(a.cfg.Addr),
			"resolvedAddr":  strings.TrimSpace(resolvedAddr),
		})
	}
	if a.startupNotice != nil {
		for _, line := range buildStartupConsoleLines(a.cfg, resolvedAddr, a.isTTYRuntimeOutput()) {
			a.startupNotice(line)
		}
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.srv.Serve(listener)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = a.srv.Shutdown(shutdownCtx)
		return a.closeRuntime(true, "shutdown", resolvedURL)
	case reason := <-a.runtimeStopCh:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = a.srv.Shutdown(shutdownCtx)
		return a.closeRuntime(true, cleanReason(reason), resolvedURL)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return a.closeRuntime(true, "server_closed", resolvedURL)
		}
		if a.runtimeLogs != nil {
			a.runtimeLogs.writeErrorEvent("serve_error", err.Error(), map[string]any{"url": resolvedURL})
		}
		_ = a.closeRuntime(false, "serve_error", resolvedURL)
		return err
	}
}

func listenWithPortFallback(addr string, maxFallbackAttempts int) (net.Listener, error) {
	listener, err := net.Listen("tcp", addr)
	if err == nil {
		return listener, nil
	}
	if !isAddressInUseError(err) {
		return nil, err
	}
	host, requestedPort := splitHostPort(addr)
	if strings.TrimSpace(host) == "" || requestedPort <= 0 || maxFallbackAttempts <= 0 {
		return nil, err
	}
	for offset := 1; offset <= maxFallbackAttempts; offset++ {
		candidatePort := requestedPort + offset
		if candidatePort > 65535 {
			break
		}
		candidateAddr := net.JoinHostPort(host, strconv.Itoa(candidatePort))
		candidateListener, candidateErr := net.Listen("tcp", candidateAddr)
		if candidateErr == nil {
			return candidateListener, nil
		}
		if !isAddressInUseError(candidateErr) {
			return nil, candidateErr
		}
	}
	return nil, err
}

func requestIsLoopback(r *http.Request) bool {
	if r == nil {
		return false
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		host = strings.TrimSpace(r.RemoteAddr)
	}
	ip := net.ParseIP(strings.TrimSpace(host))
	if ip == nil {
		return false
	}
	return ip.IsLoopback()
}

func splitHostPort(addr string) (string, int) {
	host, portRaw, err := net.SplitHostPort(strings.TrimSpace(addr))
	if err != nil {
		return "", 0
	}
	port, err := strconv.Atoi(strings.TrimSpace(portRaw))
	if err != nil {
		return host, 0
	}
	return host, port
}

func isAddressInUseError(err error) bool {
	if err == nil {
		return false
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "address already in use") ||
		strings.Contains(lower, "only one usage of each socket address")
}

func startupURLFromResolvedAddr(addr string) string {
	host, port, err := net.SplitHostPort(strings.TrimSpace(addr))
	if err != nil {
		return runtimeResolvedURLFromConfig(config.Config{Host: "127.0.0.1", Port: 17880})
	}
	host = strings.TrimSpace(host)
	if host == "" || host == "0.0.0.0" || host == "::" || host == "[::]" {
		host = "127.0.0.1"
	}
	return fmt.Sprintf("http://%s:%s", host, port)
}

func splitPortFromAddr(addr string) int {
	_, portRaw, err := net.SplitHostPort(strings.TrimSpace(addr))
	if err != nil {
		return 0
	}
	port, err := strconv.Atoi(portRaw)
	if err != nil {
		return 0
	}
	return port
}

func readRuntimeSetupIdentity(cfg config.Config) (instanceName, profileKey string) {
	raw, err := os.ReadFile(runtimeSetupConfigPath(cfg))
	if err != nil {
		return "", ""
	}
	var payload runtimeSetupConfigFile
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", ""
	}
	return strings.TrimSpace(payload.Instance.Name), strings.TrimSpace(payload.Instance.Profile)
}

func buildStartupConsoleLine(cfg config.Config, resolvedAddr string) string {
	instanceName, profileKey := readRuntimeSetupIdentity(cfg)
	if instanceName == "" {
		instanceName = "unknown"
	}
	if profileKey == "" {
		profileKey = "unknown"
	}
	requestedPort := splitPortFromAddr(cfg.Addr)
	resolvedPort := splitPortFromAddr(resolvedAddr)
	return fmt.Sprintf(
		"CABINET_STARTUP url=%s requested_addr=%s resolved_addr=%s instance=%s profile=%s data_dir=%s requested_port=%d resolved_port=%d",
		startupURLFromResolvedAddr(resolvedAddr),
		strings.TrimSpace(cfg.Addr),
		strings.TrimSpace(resolvedAddr),
		instanceName,
		profileKey,
		strings.TrimSpace(cfg.DataDir),
		requestedPort,
		resolvedPort,
	)
}

func buildStartupConsoleJSONLine(cfg config.Config, resolvedAddr string) string {
	instanceName, profileKey := readRuntimeSetupIdentity(cfg)
	if instanceName == "" {
		instanceName = "unknown"
	}
	if profileKey == "" {
		profileKey = "unknown"
	}
	payload := map[string]any{
		"url":            startupURLFromResolvedAddr(resolvedAddr),
		"requested_addr": strings.TrimSpace(cfg.Addr),
		"resolved_addr":  strings.TrimSpace(resolvedAddr),
		"instance":       instanceName,
		"profile":        profileKey,
		"data_dir":       strings.TrimSpace(cfg.DataDir),
		"requested_port": splitPortFromAddr(cfg.Addr),
		"resolved_port":  splitPortFromAddr(resolvedAddr),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "CABINET_STARTUP_JSON {}"
	}
	return "CABINET_STARTUP_JSON " + string(raw)
}

func buildStartupConsoleLines(cfg config.Config, resolvedAddr string, isTTY bool) []string {
	instanceName, profileKey := readRuntimeSetupIdentity(cfg)
	if instanceName == "" {
		instanceName = "unknown"
	}
	if profileKey == "" {
		profileKey = "unknown"
	}
	requestedPort := splitPortFromAddr(cfg.Addr)
	resolvedPort := splitPortFromAddr(resolvedAddr)
	bannerTitle := "Cabinet Started"
	if isTTY {
		bannerTitle = "🚀 Cabinet Started"
	}
	return []string{
		bannerTitle,
		fmt.Sprintf("URL: %s", startupURLFromResolvedAddr(resolvedAddr)),
		fmt.Sprintf("Instance: %s", instanceName),
		fmt.Sprintf("Profile: %s", profileKey),
		fmt.Sprintf("Data Dir: %s", strings.TrimSpace(cfg.DataDir)),
		fmt.Sprintf("Port: %d (requested %d)", resolvedPort, requestedPort),
		fmt.Sprintf("Bind: %s (requested %s)", strings.TrimSpace(resolvedAddr), strings.TrimSpace(cfg.Addr)),
		buildStartupConsoleLine(cfg, resolvedAddr),
		buildStartupConsoleJSONLine(cfg, resolvedAddr),
	}
}

func isRuntimeTTY() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func (a *App) isTTYRuntimeOutput() bool {
	if a.startupIsTTY == nil {
		return false
	}
	return a.startupIsTTY()
}

func (a *App) closeRuntime(clean bool, reason, resolvedURL string) error {
	if a.runtimeLogs != nil {
		level := "info"
		event := "shutdown"
		message := "runtime stopped"
		if !clean {
			level = "error"
			event = "shutdown_unclean"
			message = "runtime stopped unexpectedly"
		}
		a.runtimeLogs.writeRuntimeEvent(level, event, message, map[string]any{
			"reason": cleanReason(reason),
			"url":    strings.TrimSpace(resolvedURL),
		})
	}
	_ = syncRuntimeLifecycleShutdown(a.cfg, resolvedURL, cleanReason(reason), clean)
	if a.db == nil {
		if a.runtimeLogs != nil {
			_ = a.runtimeLogs.Close()
		}
		return nil
	}
	cleanValue := "0"
	if clean {
		cleanValue = "1"
	}
	_, _ = a.db.Exec(`INSERT INTO app_state(key, value, updated_at) VALUES('clean_shutdown', ?, CURRENT_TIMESTAMP) ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=CURRENT_TIMESTAMP`, cleanValue)
	err := a.db.Close()
	if a.runtimeLogs != nil {
		if closeErr := a.runtimeLogs.Close(); err == nil {
			err = closeErr
		}
	}
	return err
}

func cleanReason(reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return "shutdown"
	}
	return reason
}

func contentDispositionAttachment(filename string) string {
	name := strings.TrimSpace(filename)
	if name == "" {
		name = "download"
	}
	name = strings.ReplaceAll(name, `\`, "_")
	name = strings.ReplaceAll(name, `"`, "_")
	return `attachment; filename="` + name + `"`
}

func isSupportedMediaUpload(mimeType, filename string) bool {
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	if strings.HasPrefix(mimeType, "image/") {
		return true
	}
	return mediaUploadContentType(filename) != "application/octet-stream"
}

func mediaUploadContentType(filename string) string {
	switch strings.ToLower(filepath.Ext(strings.TrimSpace(filename))) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

type telegramCatalogCaptureRequest struct {
	ProfileID    string `json:"profile_id"`
	SenderID     string `json:"sender_id"`
	ChatID       string `json:"chat_id"`
	MessageID    string `json:"message_id"`
	Text         string `json:"text"`
	Barcode      string `json:"barcode"`
	GroupingHint string `json:"grouping_hint"`
	Draft        struct {
		PartNumber       string `json:"part_number"`
		Title            string `json:"title"`
		Brand            string `json:"brand"`
		Category         string `json:"category"`
		LookupSource     string `json:"lookup_source"`
		LookupURL        string `json:"lookup_url"`
		LookupConfidence string `json:"lookup_confidence"`
	} `json:"draft"`
	Media          []telegramCatalogCaptureMediaRequest `json:"media"`
	SourceMetadata map[string]any                       `json:"source_metadata"`
}

type telegramCatalogCaptureMediaRequest struct {
	FileID        string `json:"file_id"`
	Filename      string `json:"filename"`
	MIMEType      string `json:"mime_type"`
	Kind          string `json:"kind"`
	ContentBase64 string `json:"content_base64"`
}

type telegramCatalogCaptureCallbackRequest struct {
	SenderID     string `json:"sender_id"`
	ChatID       string `json:"chat_id"`
	CallbackData string `json:"callback_data"`
}

type telegramExternalIntakeProofRequest struct {
	ProfileID         string         `json:"profile_id"`
	SenderID          string         `json:"sender_id"`
	ChatID            string         `json:"chat_id"`
	SourceThreadID    string         `json:"source_thread_id"`
	SourceMessageID   string         `json:"source_message_id"`
	CapabilityID      string         `json:"capability_id"`
	PreviewID         string         `json:"preview_id"`
	ConfirmationState string         `json:"confirmation_state"`
	ProofApproved     bool           `json:"proof_approved"`
	ProviderTrace     map[string]any `json:"provider_trace"`
}

type profileSettingsTelegramAuthorizer struct {
	profiles  *profile.Repository
	profileID string
}

type allProfilesTelegramAuthorizer struct {
	profiles *profile.Repository
}

func (a profileSettingsTelegramAuthorizer) AuthorizeTelegramCapture(ctx context.Context, senderID, chatID string) (telegramcapture.AuthorizedProfile, error) {
	profileID := strings.TrimSpace(a.profileID)
	if profileID == "" || a.profiles == nil {
		return telegramcapture.AuthorizedProfile{}, telegramcapture.ErrUnauthorizedSender
	}
	settings, err := a.profiles.GetSettings(ctx, profileID)
	if err != nil {
		return telegramcapture.AuthorizedProfile{}, err
	}
	if strings.TrimSpace(settings["telegram.catalog_capture.sender_id"]) != strings.TrimSpace(senderID) {
		return telegramcapture.AuthorizedProfile{}, telegramcapture.ErrUnauthorizedSender
	}
	if strings.TrimSpace(settings["telegram.catalog_capture.chat_id"]) != strings.TrimSpace(chatID) {
		return telegramcapture.AuthorizedProfile{}, telegramcapture.ErrUnauthorizedSender
	}
	return telegramcapture.AuthorizedProfile{ProfileID: profileID}, nil
}

func (a allProfilesTelegramAuthorizer) AuthorizeTelegramCapture(ctx context.Context, senderID, chatID string) (telegramcapture.AuthorizedProfile, error) {
	if a.profiles == nil || strings.TrimSpace(senderID) == "" || strings.TrimSpace(chatID) == "" {
		return telegramcapture.AuthorizedProfile{}, telegramcapture.ErrUnauthorizedSender
	}
	profiles, err := a.profiles.List(ctx)
	if err != nil {
		return telegramcapture.AuthorizedProfile{}, err
	}
	for _, candidate := range profiles {
		settings, err := a.profiles.GetSettings(ctx, candidate.ID)
		if err != nil {
			return telegramcapture.AuthorizedProfile{}, err
		}
		if strings.TrimSpace(settings["telegram.catalog_capture.sender_id"]) != strings.TrimSpace(senderID) {
			continue
		}
		if strings.TrimSpace(settings["telegram.catalog_capture.chat_id"]) != strings.TrimSpace(chatID) {
			continue
		}
		return telegramcapture.AuthorizedProfile{ProfileID: candidate.ID}, nil
	}
	return telegramcapture.AuthorizedProfile{}, telegramcapture.ErrUnauthorizedSender
}

func telegramCatalogCaptureLocalBarcodeDraft(ctx context.Context, conn *sql.DB, profileID, barcodeValue string) (telegramcapture.Draft, bool, error) {
	profileID = strings.TrimSpace(profileID)
	barcodeValue = strings.TrimSpace(barcodeValue)
	if conn == nil || profileID == "" || barcodeValue == "" {
		return telegramcapture.Draft{}, false, nil
	}

	var itemID string
	var draft telegramcapture.Draft
	err := conn.QueryRowContext(ctx, `
		SELECT c.id, c.part_number, c.title, c.brand, c.category
		FROM item_barcodes b
		JOIN canonical_items c ON c.id = b.item_id
		WHERE b.barcode = ?
			AND c.profile_id = ?
			AND COALESCE(c.deleted_at, '') = ''
		ORDER BY b.created_at ASC
		LIMIT 1
	`, barcodeValue, profileID).Scan(&itemID, &draft.PartNumber, &draft.Title, &draft.Brand, &draft.Category)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return telegramcapture.Draft{}, false, nil
		}
		return telegramcapture.Draft{}, false, err
	}
	draft.LookupSource = "barcode_local"
	draft.LookupURL = "/api/barcodes/" + url.PathEscape(barcodeValue)
	draft.LookupConfidence = "high"
	return draft, strings.TrimSpace(itemID) != "", nil
}

func telegramCatalogCaptureMedia(media []telegramCatalogCaptureMediaRequest) ([]telegramcapture.MediaInput, error) {
	out := make([]telegramcapture.MediaInput, 0, len(media))
	for _, item := range media {
		var reader io.Reader = strings.NewReader("")
		if raw := strings.TrimSpace(item.ContentBase64); raw != "" {
			decoded, err := base64.StdEncoding.DecodeString(raw)
			if err != nil {
				return nil, err
			}
			reader = bytes.NewReader(decoded)
		}
		out = append(out, telegramcapture.MediaInput{
			FileID:   strings.TrimSpace(item.FileID),
			Filename: strings.TrimSpace(item.Filename),
			MIMEType: strings.TrimSpace(item.MIMEType),
			Kind:     strings.TrimSpace(item.Kind),
			Reader:   reader,
		})
	}
	return out, nil
}

func missingTelegramExternalIntakeProofFields(req telegramExternalIntakeProofRequest) []string {
	var missing []string
	requiredStrings := map[string]string{
		"profile_id":         req.ProfileID,
		"sender_id":          req.SenderID,
		"chat_id":            req.ChatID,
		"source_thread_id":   req.SourceThreadID,
		"source_message_id":  req.SourceMessageID,
		"capability_id":      req.CapabilityID,
		"preview_id":         req.PreviewID,
		"confirmation_state": req.ConfirmationState,
	}
	names := make([]string, 0, len(requiredStrings))
	for name := range requiredStrings {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if strings.TrimSpace(requiredStrings[name]) == "" {
			missing = append(missing, name)
		}
	}
	if !req.ProofApproved {
		missing = append(missing, "proof_approved")
	}
	if !isTelegramCatalogCapability(req.CapabilityID) {
		missing = append(missing, "capability_id.catalog_add_from")
	}
	if !isAllowedExternalProofConfirmation(req.ConfirmationState) {
		missing = append(missing, "confirmation_state.valid")
	}
	if provider, _ := req.ProviderTrace["provider"].(string); strings.TrimSpace(provider) != "openai" {
		missing = append(missing, "provider_trace.provider")
	}
	if !truthy(req.ProviderTrace["live_provider"]) {
		missing = append(missing, "provider_trace.live_provider")
	}
	if requestID, _ := req.ProviderTrace["request_id"].(string); strings.TrimSpace(requestID) == "" {
		missing = append(missing, "provider_trace.request_id")
	}
	if resultID, _ := req.ProviderTrace["result_id"].(string); strings.TrimSpace(resultID) == "" {
		missing = append(missing, "provider_trace.result_id")
	}
	if value, ok := req.ProviderTrace["credential_returned"]; !ok || truthy(value) {
		missing = append(missing, "provider_trace.credential_returned_false")
	}
	for _, key := range []string{"api_key", "token", "secret", "authorization"} {
		if _, ok := req.ProviderTrace[key]; ok {
			missing = append(missing, "provider_trace.no_"+key)
		}
	}
	return missing
}

func isTelegramCatalogCapability(capabilityID string) bool {
	switch strings.TrimSpace(capabilityID) {
	case "catalog_add_from_photo", "catalog_add_from_barcode", "catalog_add_from_text":
		return true
	default:
		return false
	}
}

func isAllowedExternalProofConfirmation(state string) bool {
	switch strings.TrimSpace(state) {
	case "pending", "confirmed", "cancelled":
		return true
	default:
		return false
	}
}

func truthy(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "true")
	default:
		return false
	}
}

func nonSecretProviderTrace(in map[string]any) map[string]any {
	out := map[string]any{}
	for key, value := range in {
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "api_key", "token", "secret", "authorization":
			continue
		default:
			out[key] = value
		}
	}
	out["credential_returned"] = false
	out["proof_packet"] = "authorized_telegram_openai_external_intake"
	return out
}

func (a *App) close() error {
	return a.closeRuntime(true, "shutdown", "")
}

package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/collectors-tech/cabinet/internal/config"
)

type runtimeLogManager struct {
	cfg          config.Config
	mu           sync.Mutex
	startedAtUTC time.Time
	logSetID     string
	runtimePath  string
	errorPath    string
	accessPath   string
	runtimeLog   *os.File
	errorLog     *os.File
	accessLog    *os.File
}

type runtimeStatusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *runtimeStatusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func runtimeLogSetID(startedAtUTC time.Time) string {
	return startedAtUTC.UTC().Format("20060102T150405.000Z")
}

func runtimeLogsDir(cfg config.Config) string {
	return filepath.Join(cfg.DataDir, "logs")
}

func runtimeLogPath(cfg config.Config, startedAtUTC time.Time) string {
	return filepath.Join(runtimeLogsDir(cfg), fmt.Sprintf("cabinet.runtime.%s.log", runtimeLogSetID(startedAtUTC)))
}

func runtimeErrorLogPath(cfg config.Config, startedAtUTC time.Time) string {
	return filepath.Join(runtimeLogsDir(cfg), fmt.Sprintf("cabinet.error.%s.log", runtimeLogSetID(startedAtUTC)))
}

func runtimeAccessLogPath(cfg config.Config, startedAtUTC time.Time) string {
	return filepath.Join(runtimeLogsDir(cfg), fmt.Sprintf("cabinet.access.%s.log", runtimeLogSetID(startedAtUTC)))
}

func newRuntimeLogManager(cfg config.Config) (*runtimeLogManager, error) {
	if err := os.MkdirAll(runtimeLogsDir(cfg), 0o755); err != nil {
		return nil, err
	}
	startedAtUTC := time.Now().UTC()
	logSetID := runtimeLogSetID(startedAtUTC)
	runtimePath := runtimeLogPath(cfg, startedAtUTC)
	errorPath := runtimeErrorLogPath(cfg, startedAtUTC)
	accessPath := runtimeAccessLogPath(cfg, startedAtUTC)
	runtimeFile, err := os.OpenFile(runtimePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	errorFile, err := os.OpenFile(errorPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		_ = runtimeFile.Close()
		return nil, err
	}
	accessFile, err := os.OpenFile(accessPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		_ = runtimeFile.Close()
		_ = errorFile.Close()
		return nil, err
	}
	return &runtimeLogManager{
		cfg:          cfg,
		startedAtUTC: startedAtUTC,
		logSetID:     logSetID,
		runtimePath:  runtimePath,
		errorPath:    errorPath,
		accessPath:   accessPath,
		runtimeLog:   runtimeFile,
		errorLog:     errorFile,
		accessLog:    accessFile,
	}, nil
}

func (m *runtimeLogManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var firstErr error
	for _, f := range []*os.File{m.runtimeLog, m.errorLog, m.accessLog} {
		if f == nil {
			continue
		}
		if err := f.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	m.runtimeLog = nil
	m.errorLog = nil
	m.accessLog = nil
	return firstErr
}

func (m *runtimeLogManager) writeJSONLine(file *os.File, payload map[string]any) {
	if file == nil {
		return
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	_, _ = file.Write(append(raw, '\n'))
}

func (m *runtimeLogManager) writeRuntimeEvent(level, event, message string, extra map[string]any) {
	payload := map[string]any{
		"ts":         time.Now().UTC().Format(time.RFC3339),
		"level":      level,
		"type":       "runtime",
		"event":      event,
		"message":    message,
		"pid":        os.Getpid(),
		"dataDir":    strings.TrimSpace(m.cfg.DataDir),
		"logSetId":   m.logSetID,
		"runtimeLog": m.runtimePath,
		"errorLog":   m.errorPath,
		"accessLog":  m.accessPath,
	}
	instanceName, profileKey := readRuntimeSetupIdentity(m.cfg)
	if instanceName != "" {
		payload["instance"] = instanceName
	}
	if profileKey != "" {
		payload["profile"] = profileKey
	}
	for key, value := range extra {
		payload[key] = value
	}
	m.writeJSONLine(m.runtimeLog, payload)
}

func (m *runtimeLogManager) writeErrorEvent(event, message string, extra map[string]any) {
	payload := map[string]any{
		"ts":        time.Now().UTC().Format(time.RFC3339),
		"level":     "error",
		"type":      "error",
		"event":     event,
		"message":   message,
		"pid":       os.Getpid(),
		"logSetId":  m.logSetID,
		"errorLog":  m.errorPath,
		"dataDir":   strings.TrimSpace(m.cfg.DataDir),
	}
	for key, value := range extra {
		payload[key] = value
	}
	m.writeJSONLine(m.errorLog, payload)
}

func (m *runtimeLogManager) writeAccessEvent(r *http.Request, status int, duration time.Duration) {
	payload := map[string]any{
		"ts":         time.Now().UTC().Format(time.RFC3339),
		"level":      "info",
		"type":       "access",
		"message":    "request completed",
		"pid":        os.Getpid(),
		"method":     r.Method,
		"path":       r.URL.Path,
		"status":     status,
		"durationMs": duration.Milliseconds(),
		"remoteAddr": strings.TrimSpace(r.RemoteAddr),
		"logSetId":   m.logSetID,
		"accessLog":  m.accessPath,
		"dataDir":    strings.TrimSpace(m.cfg.DataDir),
	}
	m.writeJSONLine(m.accessLog, payload)
}

func (m *runtimeLogManager) errorWriter() *runtimeErrorLogWriter {
	return &runtimeErrorLogWriter{manager: m}
}

func (m *runtimeLogManager) wrapHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		recorder := &runtimeStatusRecorder{ResponseWriter: w, status: http.StatusOK}
		defer func() {
			if recovered := recover(); recovered != nil {
				m.writeErrorEvent("panic", fmt.Sprintf("%v", recovered), map[string]any{
					"method": r.Method,
					"path":   r.URL.Path,
				})
				recorder.WriteHeader(http.StatusInternalServerError)
			}
			m.writeAccessEvent(r, recorder.status, time.Since(started))
		}()
		next.ServeHTTP(recorder, r)
	})
}

type runtimeErrorLogWriter struct {
	manager *runtimeLogManager
}

func (w *runtimeErrorLogWriter) Write(p []byte) (int, error) {
	msg := strings.TrimSpace(string(p))
	if msg != "" {
		w.manager.writeErrorEvent("http_server", msg, nil)
	}
	return len(p), nil
}

func syncRuntimeLifecycleStartup(cfg config.Config, resolvedURL, resolvedAddr string, pid int) error {
	configPath := runtimeSetupConfigPath(cfg)
	raw, err := os.ReadFile(configPath)
	if errorsIsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var payload runtimeSetupConfigFile
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	payload.Runtime.ResolvedURL = strings.TrimSpace(resolvedURL)
	payload.Meta.CurrentURL = strings.TrimSpace(resolvedURL)
	payload.Meta.UpdatedAt = now
	payload.Meta.StartedAt = now
	payload.Meta.StartedBy = "cabinet"
	payload.Meta.LaunchSource = strings.TrimSpace(resolvedAddr)
	payload.Meta.LastKnownPID = pid
	payload.Meta.LastKnownURL = strings.TrimSpace(resolvedURL)
	payload.Meta.LastHeartbeatAt = now
	payload.Meta.LastRunClean = false
	return writeRuntimeSetupConfig(cfg, payload)
}

func syncRuntimeLifecycleShutdown(cfg config.Config, resolvedURL, reason string, clean bool) error {
	configPath := runtimeSetupConfigPath(cfg)
	raw, err := os.ReadFile(configPath)
	if errorsIsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var payload runtimeSetupConfigFile
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if strings.TrimSpace(resolvedURL) != "" {
		payload.Runtime.ResolvedURL = strings.TrimSpace(resolvedURL)
		payload.Meta.CurrentURL = strings.TrimSpace(resolvedURL)
		payload.Meta.LastKnownURL = strings.TrimSpace(resolvedURL)
	}
	payload.Meta.UpdatedAt = now
	payload.Meta.LastHeartbeatAt = now
	payload.Meta.LastShutdownAt = now
	payload.Meta.LastShutdownReason = strings.TrimSpace(reason)
	payload.Meta.LastRunClean = clean
	return writeRuntimeSetupConfig(cfg, payload)
}

func syncRuntimeLifecycleUncleanRecovery(cfg config.Config) error {
	configPath := runtimeSetupConfigPath(cfg)
	raw, err := os.ReadFile(configPath)
	if errorsIsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var payload runtimeSetupConfigFile
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	payload.Meta.UpdatedAt = now
	payload.Meta.LastHeartbeatAt = now
	payload.Meta.LastRunClean = false
	if strings.TrimSpace(payload.Meta.LastShutdownReason) == "" {
		payload.Meta.LastShutdownReason = "unclean_previous_exit"
	}
	return writeRuntimeSetupConfig(cfg, payload)
}

func errorsIsNotExist(err error) bool {
	return err != nil && os.IsNotExist(err)
}

func readStructuredLogLines(path string) ([]map[string]any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return nil, nil
	}
	lines := strings.Split(trimmed, "\n")
	result := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(line), &payload); err != nil {
			return nil, err
		}
		result = append(result, payload)
	}
	return result, nil
}

func getIntFromMeta(v any) int {
	switch value := v.(type) {
	case float64:
		return int(value)
	case int:
		return value
	case string:
		parsed, _ := strconv.Atoi(strings.TrimSpace(value))
		return parsed
	default:
		return 0
	}
}

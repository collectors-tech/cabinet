package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/collectors-tech/cabinet/internal/config"
)

func TestRuntimeLifecycleMetadataAndStructuredLogs(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	cfg := config.Config{
		Addr:    "127.0.0.1:0",
		Host:    "127.0.0.1",
		Port:    17882,
		DataDir: base,
		DBPath:  base + "/cabinet.db",
	}
	payload, err := buildRuntimeSetupConfig(cfg, runtimeSetupRequest{
		InstanceName: "Runtime Logging Test",
		ProfileKey:   "runtime-logging-test",
		AuthMode:     "local",
	})
	if err != nil {
		t.Fatalf("buildRuntimeSetupConfig() error = %v", err)
	}
	if err := writeRuntimeSetupConfig(cfg, payload); err != nil {
		t.Fatalf("writeRuntimeSetupConfig() error = %v", err)
	}

	a, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	startupLines := make(chan string, 16)
	a.startupNotice = func(line string) {
		select {
		case startupLines <- line:
		default:
		}
	}
	a.startupIsTTY = func() bool { return false }

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runErr := make(chan error, 1)
	go func() {
		runErr <- a.Run(ctx)
	}()

	var resolvedURL string
	deadline := time.After(5 * time.Second)
	for resolvedURL == "" {
		select {
		case line := <-startupLines:
			if strings.HasPrefix(line, "CABINET_STARTUP ") {
				parts := strings.Split(line, " ")
				for _, part := range parts {
					if strings.HasPrefix(part, "url=") {
						resolvedURL = strings.TrimPrefix(part, "url=")
						break
					}
				}
			}
		case err := <-runErr:
			if err != nil {
				t.Fatalf("Run() early error = %v", err)
			}
			t.Fatal("Run() returned before startup line")
		case <-deadline:
			t.Fatal("timeout waiting for runtime startup line")
		}
	}

	for _, path := range []string{"/healthz", "/api/runtime", "/missing"} {
		resp, err := http.Get(resolvedURL + path) //nolint:gosec // local runtime test endpoint
		if err != nil {
			t.Fatalf("GET %s error = %v", path, err)
		}
		_ = resp.Body.Close()
	}

	cancel()
	select {
	case err := <-runErr:
		if err != nil {
			t.Fatalf("Run() shutdown error = %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for runtime shutdown")
	}

	raw, err := os.ReadFile(runtimeSetupConfigPath(cfg))
	if err != nil {
		t.Fatalf("read cabinet.json: %v", err)
	}
	var updated runtimeSetupConfigFile
	if err := json.Unmarshal(raw, &updated); err != nil {
		t.Fatalf("unmarshal cabinet.json: %v", err)
	}
	if strings.TrimSpace(updated.Meta.StartedAt) == "" {
		t.Fatalf("expected meta.startedAt in cabinet.json, got %+v", updated.Meta)
	}
	if strings.TrimSpace(updated.Meta.LastKnownURL) == "" {
		t.Fatalf("expected meta.lastKnownUrl in cabinet.json, got %+v", updated.Meta)
	}
	if updated.Meta.LastKnownPID <= 0 {
		t.Fatalf("expected meta.lastKnownPid > 0, got %+v", updated.Meta)
	}
	if strings.TrimSpace(updated.Meta.LastShutdownAt) == "" {
		t.Fatalf("expected meta.lastShutdownAt in cabinet.json, got %+v", updated.Meta)
	}
	if !updated.Meta.LastRunClean {
		t.Fatalf("expected meta.lastRunClean=true after clean shutdown, got %+v", updated.Meta)
	}

	runtimeLines, err := readStructuredLogLines(runtimeLogPath(cfg))
	if err != nil {
		t.Fatalf("read runtime log: %v", err)
	}
	if len(runtimeLines) == 0 {
		t.Fatal("expected runtime log lines")
	}
	if !containsEvent(runtimeLines, "startup") {
		t.Fatalf("expected startup event in runtime log, got %+v", runtimeLines)
	}
	if !containsEvent(runtimeLines, "shutdown") {
		t.Fatalf("expected shutdown event in runtime log, got %+v", runtimeLines)
	}

	accessLines, err := readStructuredLogLines(runtimeAccessLogPath(cfg))
	if err != nil {
		t.Fatalf("read access log: %v", err)
	}
	if len(accessLines) < 3 {
		t.Fatalf("expected access log entries for exercised requests, got %+v", accessLines)
	}
	if !containsPath(accessLines, "/healthz") {
		t.Fatalf("expected /healthz access log entry, got %+v", accessLines)
	}
	if !containsPath(accessLines, "/api/runtime") {
		t.Fatalf("expected /api/runtime access log entry, got %+v", accessLines)
	}
}

func containsEvent(lines []map[string]any, event string) bool {
	for _, line := range lines {
		if fmt.Sprint(line["event"]) == event {
			return true
		}
	}
	return false
}

func containsPath(lines []map[string]any, path string) bool {
	for _, line := range lines {
		if fmt.Sprint(line["path"]) == path {
			return true
		}
	}
	return false
}

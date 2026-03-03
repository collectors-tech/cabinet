package app

import (
	"encoding/json"
	"context"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/collectors-tech/cabinet/internal/config"
)

func TestRuntimeStartupConsoleOutputsResolvedURLAndContext(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	setupPayload := runtimeSetupConfigFile{
		Version: 1,
		Instance: runtimeSetupInstanceConfig{
			Name:    "Demo Instance",
			Profile: "demo-profile",
		},
		Storage: runtimeSetupStorageConfig{
			DataDir:  filepath.Join(a.cfg.DataDir, "profiles", "demo-profile"),
			MediaDir: filepath.Join(a.cfg.DataDir, "profiles", "demo-profile", "media"),
		},
		Runtime: runtimeSetupRuntimeConfig{
			PortMode:    "auto",
			ResolvedURL: "http://127.0.0.1:17880",
		},
		Auth: runtimeSetupAuthConfig{
			Mode: "local",
			Clerk: runtimeSetupClerkAuthConfig{
				PublishableKey: "",
				Enabled:        false,
			},
		},
		Bootstrap: runtimeSetupBootstrapConfig{
			Workspace:       "Local Workspace",
			DatabaseProfile: "Primary DB",
		},
		Features: runtimeSetupFeaturesConfig{
			Chat:      true,
			Providers: true,
			Scanner:   true,
		},
		Meta: runtimeSetupMetaConfig{
			CreatedAt:     "2026-01-01T00:00:00Z",
			UpdatedAt:     "2026-01-01T00:00:00Z",
			WizardVersion: "1",
			CurrentURL:    "http://127.0.0.1:17880",
		},
	}
	if err := writeRuntimeSetupConfig(a.cfg, setupPayload); err != nil {
		t.Fatalf("writeRuntimeSetupConfig() error = %v", err)
	}

	startupLineCh := make(chan string, 16)
	a.startupNotice = func(line string) {
		select {
		case startupLineCh <- line:
		default:
		}
	}
	a.startupIsTTY = func() bool { return false }

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- a.Run(ctx)
	}()

	lines := make([]string, 0, 8)
	var startupLine string
	var startupJSONLine string
	var startupBannerLine string
	select {
	case <-time.After(4 * time.Second):
		cancel()
		<-errCh
		t.Fatal("timed out waiting for runtime startup console lines")
	case line := <-startupLineCh:
		lines = append(lines, line)
	}

	deadline := time.After(4 * time.Second)
	for {
		for _, line := range lines {
			if strings.HasPrefix(line, "CABINET_STARTUP ") {
				startupLine = line
			}
			if strings.HasPrefix(line, "CABINET_STARTUP_JSON ") {
				startupJSONLine = line
			}
			if strings.Contains(line, "Cabinet Started") {
				startupBannerLine = line
			}
		}
		if startupLine != "" && startupJSONLine != "" && startupBannerLine != "" {
			break
		}
		select {
		case line := <-startupLineCh:
			lines = append(lines, line)
		case <-deadline:
			cancel()
			<-errCh
			t.Fatalf("timed out waiting for startup output set; lines=%q", lines)
		}
	}

	if !strings.Contains(startupLine, "CABINET_STARTUP") {
		t.Fatalf("expected CABINET_STARTUP marker, got %q", startupLine)
	}
	if !strings.Contains(startupLine, "url=http://127.0.0.1:") {
		t.Fatalf("expected resolved runtime url in startup line, got %q", startupLine)
	}
	if !strings.Contains(startupLine, "instance=Demo Instance") {
		t.Fatalf("expected instance context in startup line, got %q", startupLine)
	}
	if !strings.Contains(startupLine, "profile=demo-profile") {
		t.Fatalf("expected profile context in startup line, got %q", startupLine)
	}
	if !strings.Contains(startupLine, "data_dir="+a.cfg.DataDir) {
		t.Fatalf("expected data_dir context in startup line, got %q", startupLine)
	}

	portPattern := regexp.MustCompile(`requested_port=\d+\s+resolved_port=\d+`)
	if !portPattern.MatchString(startupLine) {
		t.Fatalf("expected requested/resolved port fields in startup line, got %q", startupLine)
	}
	if !strings.HasPrefix(startupJSONLine, "CABINET_STARTUP_JSON ") {
		t.Fatalf("expected CABINET_STARTUP_JSON marker, got %q", startupJSONLine)
	}

	payloadRaw := strings.TrimPrefix(startupJSONLine, "CABINET_STARTUP_JSON ")
	var startupPayload map[string]any
	if err := json.Unmarshal([]byte(payloadRaw), &startupPayload); err != nil {
		t.Fatalf("startup json line should decode: %v; line=%q", err, startupJSONLine)
	}
	if startupPayload["instance"] != "Demo Instance" {
		t.Fatalf("expected instance in startup json payload, got %v", startupPayload["instance"])
	}
	if startupPayload["profile"] != "demo-profile" {
		t.Fatalf("expected profile in startup json payload, got %v", startupPayload["profile"])
	}
	if startupPayload["data_dir"] != a.cfg.DataDir {
		t.Fatalf("expected data_dir in startup json payload, got %v", startupPayload["data_dir"])
	}
	if !strings.Contains(startupBannerLine, "Cabinet Started") {
		t.Fatalf("expected plain startup banner line, got %q", startupBannerLine)
	}

	cancel()
	if err := <-errCh; err != nil {
		t.Fatalf("Run() error after cancel = %v", err)
	}
}

func TestBuildStartupConsoleLinesUsesEmojiForTTYOutput(t *testing.T) {
	t.Parallel()

	cfg := configForStartupLineTest(t)
	if err := writeRuntimeSetupConfig(cfg, runtimeSetupConfigFile{
		Version: 1,
		Instance: runtimeSetupInstanceConfig{
			Name:    "TTY Instance",
			Profile: "tty-profile",
		},
		Storage: runtimeSetupStorageConfig{
			DataDir:  cfg.DataDir,
			MediaDir: filepath.Join(cfg.DataDir, "media"),
		},
		Meta: runtimeSetupMetaConfig{
			CreatedAt:     "2026-01-01T00:00:00Z",
			UpdatedAt:     "2026-01-01T00:00:00Z",
			WizardVersion: "1",
		},
	}); err != nil {
		t.Fatalf("writeRuntimeSetupConfig() error = %v", err)
	}

	lines := buildStartupConsoleLines(cfg, "127.0.0.1:17880", true)
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 startup console lines, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "🚀 Cabinet Started") {
		t.Fatalf("expected emoji banner for tty output, got %q", lines[0])
	}
	if !strings.HasPrefix(lines[len(lines)-1], "CABINET_STARTUP_JSON ") {
		t.Fatalf("expected machine json startup line, got %q", lines[len(lines)-1])
	}
}

func configForStartupLineTest(t *testing.T) config.Config {
	t.Helper()

	tempDir := t.TempDir()
	return config.Config{
		Host:    "127.0.0.1",
		Port:    17880,
		Addr:    "127.0.0.1:17880",
		DataDir: tempDir,
	}
}

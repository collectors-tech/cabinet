package app

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
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

	startupLineCh := make(chan string, 1)
	a.startupNotice = func(line string) {
		select {
		case startupLineCh <- line:
		default:
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- a.Run(ctx)
	}()

	var startupLine string
	select {
	case startupLine = <-startupLineCh:
	case <-time.After(4 * time.Second):
		cancel()
		<-errCh
		t.Fatal("timed out waiting for runtime startup console line")
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

	cancel()
	if err := <-errCh; err != nil {
		t.Fatalf("Run() error after cancel = %v", err)
	}
}

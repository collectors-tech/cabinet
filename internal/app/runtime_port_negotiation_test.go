package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/collectors-tech/cabinet/internal/config"
)

func TestRunFallsBackToNextAvailablePortWhenRequestedPortIsOccupied(t *testing.T) {
	t.Parallel()

	blocker, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen blocker: %v", err)
	}
	defer blocker.Close()
	requestedPort := blocker.Addr().(*net.TCPAddr).Port

	cfg := config.Config{
		Addr:    fmt.Sprintf("127.0.0.1:%d", requestedPort),
		Host:    "127.0.0.1",
		Port:    requestedPort,
		DataDir: t.TempDir(),
		DBPath:  t.TempDir() + "/cabinet.db",
	}
	a, err := New(cfg)
	if err != nil {
		t.Fatalf("new app: %v", err)
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

	var startupLine string
	var resolvedURL string
	deadline := time.After(5 * time.Second)
waitLoop:
	for {
		select {
		case line := <-startupLines:
			if strings.HasPrefix(line, "CABINET_STARTUP ") {
				startupLine = line
				if idx := strings.Index(line, "url="); idx >= 0 {
					urlPart := strings.TrimSpace(line[idx+len("url="):])
					if space := strings.IndexByte(urlPart, ' '); space >= 0 {
						urlPart = urlPart[:space]
					}
					resolvedURL = urlPart
				}
				break waitLoop
			}
		case err := <-runErr:
			if err != nil {
				t.Fatalf("run returned before startup with error: %v", err)
			}
			t.Fatal("run returned before startup notice")
		case <-deadline:
			t.Fatal("timeout waiting for startup notice")
		}
	}

	if strings.Contains(startupLine, fmt.Sprintf("requested_port=%d resolved_port=%d", requestedPort, requestedPort)) {
		t.Fatalf("expected fallback to different resolved port, line=%q", startupLine)
	}
	if resolvedURL == "" {
		t.Fatalf("expected resolved URL in startup line, line=%q", startupLine)
	}

	resp, err := http.Get(resolvedURL + "/healthz") //nolint:gosec // local runtime test endpoint
	if err != nil {
		t.Fatalf("healthz request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected healthz 200, got %d", resp.StatusCode)
	}

	cancel()
	select {
	case err := <-runErr:
		if err != nil {
			t.Fatalf("run shutdown error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for runtime shutdown")
	}
}


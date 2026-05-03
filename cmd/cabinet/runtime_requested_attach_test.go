package main

import (
	"testing"

	"github.com/collectors-tech/cabinet/internal/config"
)

func TestResolveRequestedRuntimeAttachWhenEndpointHealthyAndParallelDisabled(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Addr:    "127.0.0.1:17882",
		Host:    "127.0.0.1",
		Port:    17882,
		DataDir: t.TempDir(),
	}

	decision := resolveRequestedRuntimeAttach(cfg, false, func(runtimeURL string) bool {
		return runtimeURL == "http://127.0.0.1:17882/"
	})

	if !decision.Attach {
		t.Fatalf("expected attach=true when requested endpoint already serves cabinet")
	}
	if decision.URL != "http://127.0.0.1:17882/" {
		t.Fatalf("expected requested endpoint URL, got %q", decision.URL)
	}
}

func TestResolveRequestedRuntimeAttachSkipsAttachWhenParallelAllowed(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Addr:    "127.0.0.1:17882",
		Host:    "127.0.0.1",
		Port:    17882,
		DataDir: t.TempDir(),
	}

	healthyChecks := 0
	decision := resolveRequestedRuntimeAttach(cfg, true, func(runtimeURL string) bool {
		healthyChecks++
		return runtimeURL == "http://127.0.0.1:17882/"
	})

	if decision.Attach {
		t.Fatalf("expected attach=false when explicit parallel mode is enabled")
	}
	if healthyChecks != 0 {
		t.Fatalf("expected no endpoint health checks in explicit parallel mode, got %d", healthyChecks)
	}
}

func TestResolveRequestedRuntimeAttachUsesLoopbackForLanBind(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Addr:    "0.0.0.0:17882",
		Host:    "0.0.0.0",
		Port:    17882,
		DataDir: t.TempDir(),
	}

	decision := resolveRequestedRuntimeAttach(cfg, false, func(runtimeURL string) bool {
		return runtimeURL == "http://127.0.0.1:17882/"
	})

	if !decision.Attach {
		t.Fatalf("expected attach=true when loopback-mapped endpoint is healthy")
	}
	if decision.URL != "http://127.0.0.1:17882/" {
		t.Fatalf("expected loopback attach URL, got %q", decision.URL)
	}
}

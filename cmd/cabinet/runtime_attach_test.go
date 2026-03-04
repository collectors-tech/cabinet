package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveRunningRuntimeAttachWhenPIDAliveAndEndpointHealthy(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dataDir, "cabinet.pid"), []byte("4242\n"), 0o644); err != nil {
		t.Fatalf("write pid: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "cabinet.json"), []byte(`{
  "runtime": {"resolvedURL":"http://127.0.0.1:19090"},
  "meta": {"currentUrl":"http://127.0.0.1:19090"}
}`), 0o644); err != nil {
		t.Fatalf("write setup metadata: %v", err)
	}

	decision, err := resolveRunningRuntimeAttach(dataDir, func(pid int) bool {
		return pid == 4242
	}, func(url string) bool {
		return url == "http://127.0.0.1:19090"
	})
	if err != nil {
		t.Fatalf("resolveRunningRuntimeAttach error: %v", err)
	}
	if !decision.Attach {
		t.Fatalf("expected attach=true, got false")
	}
	if decision.URL != "http://127.0.0.1:19090" {
		t.Fatalf("expected attach URL, got %q", decision.URL)
	}
	pidRaw, err := os.ReadFile(filepath.Join(dataDir, "cabinet.pid"))
	if err != nil {
		t.Fatalf("read pid file: %v", err)
	}
	if string(pidRaw) != "4242\n" {
		t.Fatalf("expected pid file to stay PID-only, got %q", string(pidRaw))
	}
}

func TestResolveRunningRuntimeAttachRemovesStalePIDWhenEndpointUnhealthy(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	pidPath := filepath.Join(dataDir, "cabinet.pid")
	if err := os.WriteFile(pidPath, []byte("4242\n"), 0o644); err != nil {
		t.Fatalf("write pid: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "cabinet.json"), []byte(`{
  "runtime": {"resolvedURL":"http://127.0.0.1:19091"},
  "meta": {"currentUrl":"http://127.0.0.1:19091"}
}`), 0o644); err != nil {
		t.Fatalf("write setup metadata: %v", err)
	}

	decision, err := resolveRunningRuntimeAttach(dataDir, func(pid int) bool {
		return pid == 4242
	}, func(url string) bool {
		return false
	})
	if err != nil {
		t.Fatalf("resolveRunningRuntimeAttach error: %v", err)
	}
	if decision.Attach {
		t.Fatalf("expected attach=false for stale lock")
	}
	if _, err := os.Stat(pidPath); !os.IsNotExist(err) {
		t.Fatalf("expected stale pid file removed, stat err=%v", err)
	}
}

func TestResolveRunningRuntimeAttachNoCrossDataDirAttach(t *testing.T) {
	t.Parallel()

	targetDataDir := t.TempDir()
	otherDataDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(otherDataDir, "cabinet.pid"), []byte("777\n"), 0o644); err != nil {
		t.Fatalf("write other pid: %v", err)
	}
	if err := os.WriteFile(filepath.Join(otherDataDir, "cabinet.json"), []byte(`{
  "runtime": {"resolvedURL":"http://127.0.0.1:19100"},
  "meta": {"currentUrl":"http://127.0.0.1:19100"}
}`), 0o644); err != nil {
		t.Fatalf("write other setup metadata: %v", err)
	}

	decision, err := resolveRunningRuntimeAttach(targetDataDir, func(pid int) bool {
		return true
	}, func(url string) bool {
		return true
	})
	if err != nil {
		t.Fatalf("resolveRunningRuntimeAttach error: %v", err)
	}
	if decision.Attach {
		t.Fatalf("expected no attach for unrelated data dir")
	}
}

func TestResolveRunningRuntimeAttachUsesMetaCurrentURLFallback(t *testing.T) {
	t.Parallel()

	dataDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dataDir, "cabinet.pid"), []byte("3131\n"), 0o644); err != nil {
		t.Fatalf("write pid: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "cabinet.json"), []byte(`{
  "runtime": {"resolvedURL":""},
  "meta": {"currentUrl":"http://127.0.0.1:19331"}
}`), 0o644); err != nil {
		t.Fatalf("write setup metadata: %v", err)
	}

	decision, err := resolveRunningRuntimeAttach(dataDir, func(pid int) bool {
		return pid == 3131
	}, func(url string) bool {
		return url == "http://127.0.0.1:19331"
	})
	if err != nil {
		t.Fatalf("resolveRunningRuntimeAttach error: %v", err)
	}
	if !decision.Attach {
		t.Fatalf("expected attach=true")
	}
	if decision.URL != "http://127.0.0.1:19331" {
		t.Fatalf("expected fallback URL, got %q", decision.URL)
	}
}

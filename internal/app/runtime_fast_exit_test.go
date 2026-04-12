package app

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/collectors-tech/cabinet/internal/config"
)

func TestRunReturnsQuicklyWhenContextAlreadyCanceled(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	cfg := config.Config{
		Addr:    "127.0.0.1:0",
		Host:    "127.0.0.1",
		Port:    0,
		DataDir: base,
		DBPath:  filepath.Join(base, "cabinet.db"),
	}
	a, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	if err := a.Run(ctx); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if took := time.Since(start); took > 500*time.Millisecond {
		t.Fatalf("expected pre-canceled Run to exit quickly, took %s", took)
	}
}

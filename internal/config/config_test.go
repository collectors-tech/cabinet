package config

import "testing"

func TestLoadDefaultsAndChannelFallback(t *testing.T) {
	t.Setenv("CABINET_ADDR", "")
	t.Setenv("CABINET_DATA_DIR", "")
	t.Setenv("CABINET_DB_PATH", "")
	t.Setenv("CABINET_UPDATE_CHANNEL", "invalid")

	cfg := Load()
	if cfg.Addr == "" {
		t.Fatal("expected default address")
	}
	if cfg.DataDir == "" {
		t.Fatal("expected default data dir")
	}
	if cfg.DBPath == "" {
		t.Fatal("expected default db path")
	}
	if string(cfg.UpdateChannel) != "stable" {
		t.Fatalf("expected stable fallback, got %q", cfg.UpdateChannel)
	}
}

package main

import (
	"context"
	"testing"
)

func TestParseLauncherArgsRequiresExplicitProfile(t *testing.T) {
	if _, err := parseLauncherArgs([]string{}); err == nil {
		t.Fatal("parseLauncherArgs() should reject missing profile binding")
	}
}

func TestParseLauncherArgsBuildsProfileBoundConfig(t *testing.T) {
	cfg, err := parseLauncherArgs([]string{
		"--profile-id", "profile-main",
		"--profile-label", "Main collection",
		"--version", "1.2.3",
		"--version-digest", "git:abc123",
	})
	if err != nil {
		t.Fatalf("parseLauncherArgs() error = %v", err)
	}
	if cfg.ProfileID != "profile-main" || cfg.ProfileLabel != "Main collection" || cfg.Version != "1.2.3" || cfg.VersionDigest != "git:abc123" {
		t.Fatalf("unexpected launcher config: %#v", cfg)
	}
}

func TestRunLauncherRejectsMissingProfileBeforeTransport(t *testing.T) {
	err := runLauncher(context.Background(), launcherConfig{})
	if err == nil {
		t.Fatal("runLauncher() should reject missing profile binding")
	}
}

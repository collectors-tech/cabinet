package app

import (
	"runtime/debug"
	"testing"
)

func TestRuntimeBuildMetadataPrefersExplicitBetaVersion(t *testing.T) {
	originalVersion := buildVersion
	originalRevision := buildRevision
	originalDate := buildDate
	t.Cleanup(func() {
		buildVersion = originalVersion
		buildRevision = originalRevision
		buildDate = originalDate
	})

	buildVersion = "0.1.0-beta.1"
	buildRevision = "303754c3cf2e940615817a53e2c496f0b99ef143"
	buildDate = "2026-07-11T17:11:22Z"

	version, revision, date := runtimeBuildMetadata()
	if version != "0.1.0-beta.1" {
		t.Fatalf("expected explicit beta version, got %q", version)
	}
	if revision != "303754c3cf2e940615817a53e2c496f0b99ef143" {
		t.Fatalf("expected full build revision from ldflags, got %q", revision)
	}
	if date != "2026-07-11T17:11:22Z" {
		t.Fatalf("expected build date from ldflags, got %q", date)
	}
}

func TestRuntimeBuildMetadataUsesFullVCSRevisionFallback(t *testing.T) {
	originalVersion := buildVersion
	originalRevision := buildRevision
	originalDate := buildDate
	t.Cleanup(func() {
		buildVersion = originalVersion
		buildRevision = originalRevision
		buildDate = originalDate
	})

	buildVersion = ""
	buildRevision = ""
	buildDate = ""
	fullRevision := "abcdef1234567890abcdef1234567890abcdef12"
	info := &debug.BuildInfo{
		Main: debug.Module{Version: "(devel)"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: fullRevision},
			{Key: "vcs.time", Value: "2026-08-12T00:00:00Z"},
		},
	}

	version, revision, date := runtimeBuildMetadataFromBuildInfo(info, true)
	if version != "rev-abcdef123456" || revision != fullRevision || date != "2026-08-12T00:00:00Z" {
		t.Fatalf("unexpected VCS fallback metadata: version=%q revision=%q date=%q", version, revision, date)
	}
}

func TestRuntimeBuildMetadataUsesUnknownForUnstampedBuild(t *testing.T) {
	originalVersion := buildVersion
	originalRevision := buildRevision
	originalDate := buildDate
	t.Cleanup(func() {
		buildVersion = originalVersion
		buildRevision = originalRevision
		buildDate = originalDate
	})

	buildVersion = ""
	buildRevision = "not-a-full-commit"
	buildDate = ""

	version, revision, date := runtimeBuildMetadataFromBuildInfo(nil, false)
	if version != "dev" || revision != "unknown" || date != "unknown" {
		t.Fatalf("unexpected unstamped metadata: version=%q revision=%q date=%q", version, revision, date)
	}
}

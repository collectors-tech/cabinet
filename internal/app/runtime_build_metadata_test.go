package app

import "testing"

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

	version, date := runtimeBuildMetadata()
	if version != "0.1.0-beta.1" {
		t.Fatalf("expected explicit beta version, got %q", version)
	}
	if date != "2026-07-11T17:11:22Z" {
		t.Fatalf("expected build date from ldflags, got %q", date)
	}
}

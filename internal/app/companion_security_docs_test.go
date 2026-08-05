package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrowserCompanionSecurityDocumentation(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	assertFileContains := func(path string, fragments ...string) {
		t.Helper()
		raw, err := os.ReadFile(filepath.Join(repoRoot, path))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		text := string(raw)
		for _, fragment := range fragments {
			if !strings.Contains(text, fragment) {
				t.Errorf("%s missing security/recovery fragment %q", path, fragment)
			}
		}
	}

	assertFileContains(
		"docs/help-center/sections/integrations.md",
		"six-digit pairing code",
		"Revoke all",
		"pair them separately",
		"must not export browser cookies",
		"cannot protect a credential from malware",
		"known-good backup",
	)
	assertFileContains(
		"openspec/specs/integrations/browser-companion/spec.md",
		"INTEGRATION-067",
		"INTEGRATION-068",
		"INTEGRATION-069",
		"INTEGRATION-070",
		"INTEGRATION-071",
		"companion_media_persistence_not_implemented",
	)
	assertFileContains(
		"browser-extension/contracts/companion-protocol-v1.json",
		`"protocol_version": "1"`,
		`"credential_prefix": "cabcmp_"`,
		`"forbidden_transports"`,
	)
}

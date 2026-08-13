package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProfileActivationRecoveryIsBoundedAndPreservesSessionUntilSuccess(t *testing.T) {
	t.Parallel()

	helperPath := filepath.Join("..", "ui.web", "src", "lib", "profile-activation.ts")
	helperRaw, err := os.ReadFile(helperPath)
	if err != nil {
		t.Fatalf("read profile activation helper: %v", err)
	}
	helper := string(helperRaw)
	for _, fragment := range []string{
		"profile_activation_unavailable",
		"maxActivationAttempts = 2",
		"response.status !== 503",
		"await wait(retryAfterMs)",
	} {
		if !strings.Contains(helper, fragment) {
			t.Fatalf("expected bounded activation helper to include %q", fragment)
		}
	}

	switcherPath := filepath.Join("..", "ui.web", "src", "components", "layout", "team-switcher.tsx")
	switcherRaw, err := os.ReadFile(switcherPath)
	if err != nil {
		t.Fatalf("read team switcher: %v", err)
	}
	switcher := string(switcherRaw)
	activationCall := strings.Index(switcher, "activateProfile(profileID)")
	lockCall := strings.Index(switcher, "lockLocalServerSession()")
	if activationCall == -1 || lockCall == -1 || activationCall > lockCall {
		t.Fatalf("profile activation must succeed before local session state is cleared")
	}
	if strings.Count(switcher, "fetch('/api/profiles/active'") != 1 {
		t.Fatalf("team switcher must delegate activation without duplicate direct requests")
	}
}

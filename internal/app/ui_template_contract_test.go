package app

import (
	"os"
	"strings"
	"testing"
)

func TestUITemplateSidebarMappingContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/components/layout/data/sidebar-data.ts")
	if err != nil {
		t.Fatalf("read sidebar data: %v", err)
	}
	src := string(b)

	required := []string{
		"title: 'Inventory'",
		"title: 'Wishlist'",
		"title: 'Integrations'",
		"title: 'Chats'",
		"title: 'Users'",
		"title: 'Settings'",
		"title: 'Help Center'",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("missing required nav token: %s", token)
		}
	}

	forbidden := []string{
		"title: 'Tasks'",
		"title: 'Apps'",
		"title: 'Secured by Clerk'",
		"title: 'Auth'",
		"title: 'Errors'",
		"title: 'Upgrade to Pro'",
	}
	for _, token := range forbidden {
		if strings.Contains(src, token) {
			t.Fatalf("forbidden template token still present: %s", token)
		}
	}
}

func TestUITemplateProfileMenusContract(t *testing.T) {
	t.Parallel()

	files := []string{
		"../../ui.web/src/components/profile-dropdown.tsx",
		"../../ui.web/src/components/layout/nav-user.tsx",
	}

	for _, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		src := string(b)
		forbidden := []string{
			"Upgrade to Pro",
			"Billing",
			"New Team",
		}
		for _, token := range forbidden {
			if strings.Contains(src, token) {
				t.Fatalf("%s contains forbidden token: %s", file, token)
			}
		}
	}
}

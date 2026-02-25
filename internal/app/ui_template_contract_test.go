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

func TestUIMigratedScreensRouteBindingContract(t *testing.T) {
	t.Parallel()

	cases := []struct {
		file      string
		forbidden []string
		required  []string
	}{
		{
			file: "../../ui.web/src/features/inventory/index.tsx",
			forbidden: []string{
				"/_authenticated/tasks/",
			},
			required: []string{
				"from '@/features/tasks'",
			},
		},
		{
			file: "../../ui.web/src/features/wishlist/index.tsx",
			forbidden: []string{
				"/_authenticated/tasks/",
			},
			required: []string{
				"from '@/features/tasks'",
			},
		},
		{
			file: "../../ui.web/src/features/integrations/index.tsx",
			forbidden: []string{
				"/_authenticated/apps/",
			},
			required: []string{
				"from '@/features/apps'",
			},
		},
		{
			file: "../../ui.web/src/features/tasks/components/tasks-table.tsx",
			forbidden: []string{
				"getRouteApi('/_authenticated/tasks/')",
			},
			required: []string{
				"getRouteApi('/_authenticated/inventory/')",
				"getRouteApi('/_authenticated/wishlist/')",
			},
		},
		{
			file: "../../ui.web/src/features/apps/index.tsx",
			forbidden: []string{
				"getRouteApi('/_authenticated/apps/')",
			},
			required: []string{
				"getRouteApi('/_authenticated/integrations/')",
			},
		},
	}

	for _, tc := range cases {
		b, err := os.ReadFile(tc.file)
		if err != nil {
			t.Fatalf("read %s: %v", tc.file, err)
		}
		src := string(b)

		for _, token := range tc.required {
			if !strings.Contains(src, token) {
				t.Fatalf("%s missing required token: %s", tc.file, token)
			}
		}
		for _, token := range tc.forbidden {
			if strings.Contains(src, token) {
				t.Fatalf("%s contains forbidden stale route token: %s", tc.file, token)
			}
		}
	}
}

func TestDashboardUsesRuntimeSummaryContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/dashboard/index.tsx")
	if err != nil {
		t.Fatalf("read dashboard feature: %v", err)
	}
	src := string(b)

	required := []string{
		"/api/dashboard",
		"new_discoveries",
		"wishlist_hits",
		"price_drops",
		"total_items",
		"total_instances",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("dashboard missing required runtime token: %s", token)
		}
	}

	forbidden := []string{
		"Total Revenue",
		"Recent Sales",
		"Customers",
		"Products",
	}
	for _, token := range forbidden {
		if strings.Contains(src, token) {
			t.Fatalf("dashboard contains stale template token: %s", token)
		}
	}
}

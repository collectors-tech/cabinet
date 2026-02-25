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
				"from '@/features/tasks'",
			},
			required: []string{
				"from '@/features/collection'",
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

func TestInventoryWishlistViewToggleContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/tasks/components/tasks-table.tsx")
	if err != nil {
		t.Fatalf("read tasks table: %v", err)
	}
	src := string(b)

	required := []string{
		"Rows",
		"Cards",
		"cabinet.viewMode.",
		"routePath",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("tasks table missing view-toggle contract token: %s", token)
		}
	}
}

func TestI18nBootstrapContract(t *testing.T) {
	t.Parallel()

	checkContains := func(path string, required []string) {
		t.Helper()
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		src := string(b)
		for _, token := range required {
			if !strings.Contains(src, token) {
				t.Fatalf("%s missing required i18n token: %s", path, token)
			}
		}
	}

	checkContains("../../ui.web/src/main.tsx", []string{"./i18n"})
	checkContains("../../ui.web/src/components/layout/app-sidebar.tsx", []string{"useTranslation"})
	checkContains("../../ui.web/src/i18n/index.ts", []string{"initReactI18next", "i18next"})
	checkContains("../../ui.web/src/locales/en/nav.json", []string{"Dashboard", "Inventory", "Wishlist"})
}

func TestLegacyTasksAppsRoutesRemovedContract(t *testing.T) {
	t.Parallel()

	legacyFiles := []string{
		"../../ui.web/src/routes/_authenticated/tasks/index.tsx",
		"../../ui.web/src/routes/_authenticated/apps/index.tsx",
	}

	for _, file := range legacyFiles {
		if _, err := os.Stat(file); err == nil {
			t.Fatalf("legacy route file should be removed: %s", file)
		}
	}
}

func TestCollectionWorkspaceSemanticContract(t *testing.T) {
	t.Parallel()

	checkContains := func(path string, required []string) {
		t.Helper()
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		src := string(b)
		for _, token := range required {
			if !strings.Contains(src, token) {
				t.Fatalf("%s missing required collection workspace token: %s", path, token)
			}
		}
	}

	checkContains("../../ui.web/src/features/inventory/index.tsx", []string{
		"from '@/features/collection'",
	})
	checkContains("../../ui.web/src/features/collection/index.tsx", []string{
		"Collection",
		"Command Row",
		"Summary Strip",
		"Folders",
		"Collection Browser",
	})
}

func TestI18nShellSharedLabelsContract(t *testing.T) {
	t.Parallel()

	checkContains := func(path string, required []string, forbidden []string) {
		t.Helper()
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		src := string(b)
		for _, token := range required {
			if !strings.Contains(src, token) {
				t.Fatalf("%s missing required i18n shell token: %s", path, token)
			}
		}
		for _, token := range forbidden {
			if strings.Contains(src, token) {
				t.Fatalf("%s contains hardcoded shell text token: %s", path, token)
			}
		}
	}

	checkContains(
		"../../ui.web/src/components/search.tsx",
		[]string{"useTranslation", "common:search.placeholder"},
		[]string{"placeholder = 'Search'"},
	)
	checkContains(
		"../../ui.web/src/components/layout/team-switcher.tsx",
		[]string{"useTranslation", "common:workspace.label", "common:workspace.add"},
		[]string{"Teams", "Add team"},
	)
	checkContains(
		"../../ui.web/src/locales/en/common.json",
		[]string{"\"search\"", "\"workspace\""},
		nil,
	)
}

package app

import (
	"os"
	"strings"
	"testing"
)

func TestUITemplateSidebarMappingContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/lib/route-metadata.ts")
	if err != nil {
		t.Fatalf("read route metadata: %v", err)
	}
	src := string(b)

	b, err = os.ReadFile("../../ui.web/src/lib/route-navigation.ts")
	if err != nil {
		t.Fatalf("read route navigation: %v", err)
	}
	src += "\n" + string(b)

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

	b, err = os.ReadFile("../../ui.web/src/components/layout/data/sidebar-data.ts")
	if err != nil {
		t.Fatalf("read sidebar data: %v", err)
	}
	sidebarSrc := string(b)
	if !strings.Contains(sidebarSrc, "buildSidebarNavigationGroups()") {
		t.Fatal("sidebar data must use canonical route metadata groups")
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
		if strings.Contains(sidebarSrc, token) {
			t.Fatalf("forbidden template token still present: %s", token)
		}
	}
}

func TestUITemplateActionRegionContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/lib/action-placement.ts")
	if err != nil {
		t.Fatalf("read action placement: %v", err)
	}
	src := string(b)

	required := []string{
		"routeActionRegionContracts",
		"route: '/dashboard'",
		"route: '/inventory'",
		"route: '/collections'",
		"route: '/wishlist'",
		"route: '/media'",
		"route: '/purchases'",
		"route: '/integrations'",
		"route: '/chats'",
		"route: '/inbox'",
		"route: '/discoveries'",
		"route: '/reports'",
		"route: '/scanner'",
		"route: '/settings/profile'",
		"route: '/settings/account'",
		"route: '/settings/appearance'",
		"route: '/settings/display'",
		"route: '/settings/billing'",
		"route: '/settings/categories'",
		"route: '/settings/integrations'",
		"route: '/settings/notifications'",
		"route: '/settings/operations'",
		"route: '/settings/skills'",
		"route: '/settings/storage'",
		"route: '/users'",
		"route: '/help-center'",
		"pageActionRegionTestId: 'reports-global-header-actions'",
		"pageActionRegionTestId: 'market-watch-global-header-actions'",
		"wholePageActionIds: ['reports-refresh', 'reports-export']",
		"wholePageActionIds: ['market-watch-create', 'market-watch-run']",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("missing action-region contract token: %s", token)
		}
	}

	files := []string{
		"../../ui.web/src/features/reports/index.tsx",
		"../../ui.web/src/features/scanner/index.tsx",
	}
	for _, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		pageSrc := string(b)
		for _, token := range []string{
			"global-header-actions",
			"<HeaderTitle",
		} {
			if !strings.Contains(pageSrc, token) {
				t.Fatalf("%s missing packaged shell action region token: %s", file, token)
			}
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

func TestWishlistLocalizationContract(t *testing.T) {
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
				t.Fatalf("%s missing required wishlist localization token: %s", path, token)
			}
		}
	}

	checkContains("../../ui.web/src/features/wishlist/index.tsx", []string{"t('wishlist.title')", "t('wishlist.description')"})
	checkContains("../../ui.web/src/locales/en/pages.json", []string{`"wishlist"`, `"title": "Wishlist"`, `"description": "Track wanted items, target prices, and planning decisions before they become owned inventory."`})
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
		"New",
		"Create",
		"inventory-new-action",
		"inventory-create-menu-trigger",
		"Folders",
		"Active Brand",
		"Active Category",
		"inventory-table-card",
		"collection-summary-line",
	})
}

func TestInventoryFolderTreeStructuredAffordancesContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/collection/index.tsx")
	if err != nil {
		t.Fatalf("read collection workspace: %v", err)
	}
	src := string(b)

	required := []string{
		"folder-tree-secondary-",
		"folder-tree-count-",
		"folder-tree-badge-",
		"folder-tree-row-actions-",
		"folder-tree-trailing-",
		"folder-tree-drag-handle-",
		"folder-tree-sort-root-az",
		"onPointerDown",
		"onMouseDown",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("collection workspace missing folder tree structured affordance token: %s", token)
		}
	}
}

func TestInventoryFolderTreePersistenceAndDragContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/collection/index.tsx")
	if err != nil {
		t.Fatalf("read collection workspace: %v", err)
	}
	src := string(b)

	required := []string{
		"inventoryTreeStorageKey",
		"inventoryWorkspaceSettingsStorageKeyPrefix",
		"inventory.folder-tree.v2",
		"loadPersistedWorkspaceSnapshot",
		"loadProfileWorkspaceSnapshot",
		"savePersistedWorkspaceSnapshot",
		"saveProfileWorkspaceSnapshot",
		"parsePersistedWorkspaceSnapshot",
		"loadInventoryTreeState",
		"/api/profiles/${encodeURIComponent(profileID)}/settings",
		"window.localStorage.setItem",
		"data-draggable-row",
		"data-invalid-drop-target",
		"draggedFolderID",
		"dragTarget",
		"folderPointerDragRef",
		"resolvePointerFolderDropTarget",
		"startFolderPointerDrag",
		"window.addEventListener('pointermove'",
		"isInvalidFolderDropTarget",
		"moveFolderNode(",
		"moveFolderNodeRelative(",
		"moveFolderNodeToRoot(",
		"folder-tree-drop-before-",
		"folder-tree-drop-after-",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("collection workspace missing folder tree persistence/drag token: %s", token)
		}
	}
}

func TestInventorySavedViewsAndFilterContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/tasks/components/tasks-table.tsx")
	if err != nil {
		t.Fatalf("read tasks table: %v", err)
	}
	src := string(b)

	required := []string{
		"inventory.saved-views.v1",
		"inventory-saved-view-select",
		"inventory-saved-view-save",
		"inventory-saved-view-name",
		"inventory-saved-view-submit",
		"Filter by title or part number...",
		"Condition",
		"Category",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("tasks table missing inventory saved view/filter token: %s", token)
		}
	}
}

func TestSettingsOperationsQueueControlsContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/settings/operations/index.tsx")
	if err != nil {
		t.Fatalf("read settings operations: %v", err)
	}
	src := string(b)

	required := []string{
		"useProfileSettings",
		"scanner_schedule",
		"operations_queue_resume_schedule",
		"settings-operations-queue-card",
		"settings-operations-queue-status",
		"settings-operations-queue-pause",
		"settings-operations-queue-resume",
		"Workers paused.",
		"Workers scheduled:",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("settings operations missing queue control token: %s", token)
		}
	}

	forbidden := []string{
		"Pause Workers (Coming soon)",
		"Resume Workers (Coming soon)",
	}
	for _, token := range forbidden {
		if strings.Contains(src, token) {
			t.Fatalf("settings operations still contains placeholder queue control token: %s", token)
		}
	}
}

func TestSettingsOperationsRecoveryWorkflowContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/settings/operations/index.tsx")
	if err != nil {
		t.Fatalf("read settings operations: %v", err)
	}
	src := string(b)

	required := []string{
		"settings-operations-auth-recovery-card",
		"settings-operations-recovery-passphrase-input",
		"settings-operations-recovery-passphrase-submit",
		"settings-operations-recovery-reset-submit",
		"settings-operations-auth-recovery-status",
		"settings-operations-auth-recovery-summary",
		"/api/auth/recovery/passphrase",
		"/api/auth/recovery/reset/begin",
		"Recovery reset session started.",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("settings operations missing recovery workflow token: %s", token)
		}
	}
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

func TestInventoryWishlistViewToggleAccessibilityContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/tasks/components/tasks-table.tsx")
	if err != nil {
		t.Fatalf("read tasks table: %v", err)
	}
	src := string(b)

	required := []string{
		"aria-pressed",
		"aria-label='Switch to rows view'",
		"aria-label='Switch to cards view'",
		"'cabinet.viewMode.inventory'",
		"'cabinet.viewMode.wishlist'",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("tasks table missing accessibility/persistence token: %s", token)
		}
	}
}

func TestSetupWizardAuthModeOptionsRetireClerk(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/auth/sign-in/index.tsx")
	if err != nil {
		t.Fatalf("read sign-in setup feature: %v", err)
	}
	src := string(b)
	required := []string{
		"<option value='local'>local</option>",
		"<option value='zitadel'>zitadel</option>",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("setup wizard missing active auth option token: %s", token)
		}
	}
	forbidden := []string{
		"<option value='clerk'>clerk</option>",
		"setup-clerk-built-in-key",
		"BUILT_IN_CLERK_PUBLISHABLE_KEY",
	}
	for _, token := range forbidden {
		if strings.Contains(src, token) {
			t.Fatalf("setup wizard still exposes retired Clerk auth token: %s", token)
		}
	}
}

func TestFrontendRetiresClerkRouteDependencySurface(t *testing.T) {
	t.Parallel()

	if _, err := os.Stat("../../ui.web/src/routes/clerk"); err == nil {
		t.Fatal("frontend still ships retired Clerk route source directory")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat Clerk route source directory: %v", err)
	}
	for _, file := range []string{
		"../../ui.web/src/assets/clerk-full-logo.tsx",
		"../../ui.web/src/assets/clerk-logo.tsx",
		"../../scripts/runtime/start-exploration-clerk.ps1",
	} {
		if _, err := os.Stat(file); err == nil {
			t.Fatalf("retired Clerk surface still exists: %s", file)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat retired Clerk surface %s: %v", file, err)
		}
	}

	files := []string{
		"../../ui.web/src/routeTree.gen.ts",
		"../../ui.web/src/features/auth/sign-in/components/user-auth-form.tsx",
		"../../ui.web/package.json",
		"../../ui.web/package-lock.json",
		"../../scripts/runtime/start-exploration-local.ps1",
	}
	for _, file := range files {
		b, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		src := string(b)
		forbidden := []string{
			"@clerk/clerk-react",
			"/clerk",
			"VITE_CLERK_PUBLISHABLE_KEY",
			"CABINET_AUTH_IDENTITY_MODE=clerk",
			"identity_mode === 'clerk'",
		}
		for _, token := range forbidden {
			if strings.Contains(src, token) {
				t.Fatalf("%s still contains retired Clerk frontend token: %s", file, token)
			}
		}
	}
}

func TestIntegrationsEditPersistenceContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/apps/index.tsx")
	if err != nil {
		t.Fatalf("read integrations feature: %v", err)
	}
	src := string(b)
	required := []string{
		"/api/profiles/active",
		"/api/profiles/${activeProfileId}/settings",
		"method: 'PUT'",
		"Save Integration",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("integrations edit contract missing token: %s", token)
		}
	}
}

func TestIntegrationsProviderConfigInputsHaveLabels(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/apps/index.tsx")
	if err != nil {
		t.Fatalf("read integrations feature: %v", err)
	}
	src := string(b)
	required := []string{
		"data-testid='integration-schema-form'",
		"function setupFieldID(field: IntegrationSetupField)",
		"return `provider-schema-${notificationHistoryID(field.key)}`",
		"const fieldID = setupFieldID(field)",
		"<Label htmlFor={fieldID}>",
		"id={fieldID}",
		"data-testid={`provider-schema-field-${field.key}`}",
		"field.type === 'select'",
		"field.type === 'textarea'",
		"field.type === 'checkbox'",
		"field.type === 'browser-auth-status'",
		"htmlFor='provider-token'",
		"OpenAI API key",
		"id='provider-token'",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("integrations provider input label contract missing token: %s", token)
		}
	}
}

func TestIntegrationsValidateHealthReconcilesVisibleStateContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/apps/index.tsx")
	if err != nil {
		t.Fatalf("read integrations feature: %v", err)
	}
	src := string(b)
	required := []string{
		"/api/provider/health?provider=",
		"const checkedAt = payload.updated_at ?? new Date().toISOString()",
		"health: nextProvider.health",
		"last_run: nextProvider.last_run",
		"last_error: payload.last_error",
		"retry_after_seconds: payload.retry_after_seconds",
		"next_action: payload.next_action",
		"Last error:",
		"Retry after:",
		"Next action:",
		"{actionMessage ? <p>{actionMessage}</p> : null}",
		"Validated ${editingProvider.display_name} health: ${healthStatus}.",
		"Validating...",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("integrations validate health reconciliation missing token: %s", token)
		}
	}
}

func TestMarketWatchProviderControlsUseProviderRegistryContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/scanner/index.tsx")
	if err != nil {
		t.Fatalf("read scanner feature: %v", err)
	}
	src := string(b)
	required := []string{
		"/api/providers/registry",
		"marketWatchProviderOptionsFromRegistry",
		"market_watch_scope",
		"workflowRefs.includes('market_watch.run')",
		"setMarketWatchProviderOptions(registryProviderOptions)",
		"marketWatchProviderOptions.map((provider) =>",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("Market Watch provider registry contract missing token: %s", token)
		}
	}
}

func TestIntegrationsProviderDetailActionsUseProviderRegistryContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/apps/index.tsx")
	if err != nil {
		t.Fatalf("read integrations feature: %v", err)
	}
	src := string(b)
	required := []string{
		"provider_category?: string",
		"provider_type?: string",
		"config_schema_ref?: string",
		"workflow_refs?: string[]",
		"providerManifestActions",
		"data-testid='provider-detail-category'",
		"data-testid='provider-detail-config-schema'",
		"data-testid='provider-detail-workflows'",
		"data-testid='provider-detail-manifest-actions'",
		"data-testid='integrations-row-details-provider-contract'",
		"data-testid='integrations-row-details-provider-actions'",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("integrations provider detail/action registry contract missing token: %s", token)
		}
	}
}

func TestIntegrationsEbaySetupStatusPanelContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/apps/index.tsx")
	if err != nil {
		t.Fatalf("read integrations feature: %v", err)
	}
	src := string(b)
	required := []string{
		"data-testid='ebay-setup-status-panel'",
		"eBay setup status",
		"Verify credentials, marketplace, token state,",
		"and provider health before running eBay query",
		"data-testid='ebay-setup-auth-mode'",
		"setupStatus?.auth_mode ??",
		"editingProvider.auth_mode",
		"data-testid='ebay-setup-marketplace'",
		"setupStatus?.marketplace ||",
		"form.marketplace ||",
		"data-testid='ebay-setup-token-state'",
		"stored token on file",
		"new token pending save",
		"token required",
		"data-testid='ebay-setup-health-state'",
		"Validation status:",
		"setupStatus?.validation_status",
		"data-testid='ebay-setup-readiness-state'",
		"Health state:",
		"data-testid='ebay-setup-next-action'",
		"setupStatus?.next_action ??",
		"editingProvider.health?.next_action",
		"formatEbaySetupNextAction(setupNextAction)",
		"formatEbaySetupBaseURLState",
		"base_url_set?: boolean",
		"setupStatus?.base_url_set",
		"data-testid='ebay-setup-base-url-override'",
		"Ready for Market Watch runs",
		"Check provider health and credentials",
		"Save credentials, then validate health",
		"Using default eBay Browse API base URL",
		"Base URL override configured",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("integrations eBay setup status panel contract missing token: %s", token)
		}
	}
}

func TestIntegrationsEbaySellerOperationsPanelContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/apps/index.tsx")
	if err != nil {
		t.Fatalf("read integrations feature: %v", err)
	}
	src := string(b)
	required := []string{
		"data-testid='ebay-seller-operations-panel'",
		"editingProvider.seller_operations",
		"ebay-seller-operation-",
		"previewSellerOperation",
		"executeSellerOperation",
		"/api/providers/ebay/seller-operations/preview",
		"/api/providers/ebay/seller-operations/execute",
		"ebay-seller-operation-preview-result",
		"ebay-seller-operation-execute-result",
		"ebay-seller-operation-read-result",
		"ebay-seller-operation-read-records",
		"sellerOperationResult.preview.remote_write",
		"sellerOperationExecution.execution.local_only",
		"sellerOperationExecution.execution.result",
		"External writes require confirmation",
		"status.blocker",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("integrations eBay seller operations panel contract missing token: %s", token)
		}
	}
}

func TestIntegrationsEbayListingLifecyclePanelContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/apps/index.tsx")
	if err != nil {
		t.Fatalf("read integrations feature: %v", err)
	}
	src := string(b)
	required := []string{
		"data-testid='ebay-listing-lifecycle-panel'",
		"listingLifecycleCommands",
		"previewListingLifecycle",
		"executeListingLifecycle",
		"/api/providers/ebay/listing-lifecycle/preview",
		"/api/providers/ebay/listing-lifecycle/execute",
		"ebay-listing-lifecycle-preview-result",
		"ebay-listing-lifecycle-execute-result",
		"listingLifecycleResult.preview.remote_write",
		"listingLifecycleExecution.execution.local_only",
		"Publish, revise, end, and relist require",
		"ebay_listing_lifecycle_adapter_required",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("integrations eBay listing lifecycle panel contract missing token: %s", token)
		}
	}
}

func TestIntegrationsEbayLandedCostPlannerPanelContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/apps/index.tsx")
	if err != nil {
		t.Fatalf("read integrations feature: %v", err)
	}
	src := string(b)
	required := []string{
		"data-testid='ebay-landed-cost-planner-panel'",
		"defaultLandedCostPayload",
		"previewLandedCostPlan",
		"/api/commerce/landed-cost/plan",
		"ebay-landed-cost-preview",
		"ebay-landed-cost-result",
		"Preview only / no mutation",
		"Landed-cost plan previewed without mutating inventory or shipment state.",
		"landedCostResult.allocation",
		"landedCostResult.consolidation",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("integrations eBay landed-cost planner panel contract missing token: %s", token)
		}
	}
}

func TestSettingsCategoriesTaxonomyControlsContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/settings/categories/index.tsx")
	if err != nil {
		t.Fatalf("read settings categories: %v", err)
	}
	src := string(b)
	required := []string{
		"inventoryPackagingGradesSettingsKey",
		"settings-packaging-grade-new",
		"settings-packaging-grade-add",
		"settings-packaging-grades-list",
		"settings-packaging-grade-remove-",
		"settings-item-type-scales-list",
		"settings-item-type-conditions-",
		"Saved categories, packaging grades, and item type condition scales.",
		"Save taxonomy settings",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("settings categories taxonomy contract missing token: %s", token)
		}
	}
}

func TestInventoryEditorTaxonomyFieldsContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/collection/index.tsx")
	if err != nil {
		t.Fatalf("read inventory collection feature: %v", err)
	}
	src := string(b)
	required := []string{
		"loadInventoryItemTypeConditionScales",
		"loadInventoryPackagingGrades",
		"inventoryItemTypeOptions",
		"inventoryConditionOptions",
		"inventoryPackagingGrades",
		"data-testid='inventory-item-type'",
		"data-testid='inventory-instance-condition'",
		"data-testid='inventory-item-packaging-grade'",
		"packaging_grade_type: itemDraft.packaging_grade_type.trim()",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("inventory editor taxonomy fields contract missing token: %s", token)
		}
	}
}

func TestInventoryTaxonomySearchFilterSortContract(t *testing.T) {
	t.Parallel()

	files := []struct {
		path     string
		required []string
	}{
		{
			path: "../../ui.web/src/features/collection/index.tsx",
			required: []string{
				"itemType: item.item_type",
				"packagingGradeType: item.packaging_grade_type",
				"condition: item.condition",
			},
		},
		{
			path: "../../ui.web/src/features/tasks/components/tasks-table.tsx",
			required: []string{
				"itemTypeFilters",
				"packagingGradeFilters",
				"columnId: 'itemType'",
				"columnId: 'packagingGradeType'",
				"Filter by title, part number, type, condition, or packaging...",
				"testIdPrefix: 'inventory-table-item-type'",
				"testIdPrefix: 'inventory-table-packaging'",
			},
		},
		{
			path: "../../ui.web/src/features/tasks/components/tasks-columns.tsx",
			required: []string{
				"accessorKey: 'itemType'",
				"accessorKey: 'packagingGradeType'",
				"data-testid='inventory-row-item-type'",
				"data-testid='inventory-row-packaging-grade'",
			},
		},
	}

	for _, file := range files {
		b, err := os.ReadFile(file.path)
		if err != nil {
			t.Fatalf("read %s: %v", file.path, err)
		}
		src := string(b)
		for _, token := range file.required {
			if !strings.Contains(src, token) {
				t.Fatalf("%s inventory taxonomy search/filter/sort contract missing token: %s", file.path, token)
			}
		}
	}
}

func TestWishlistTaxonomyFieldsContract(t *testing.T) {
	t.Parallel()

	files := []struct {
		path     string
		required []string
	}{
		{
			path: "../../ui.web/src/features/tasks/index.tsx",
			required: []string{
				"loadWishlistItemTypeConditionScales",
				"loadWishlistPackagingGrades",
				"item_type: draft.itemType",
				"packaging_grade_type: draft.packagingGradeType",
				"condition: draft.condition",
				"itemType: item.item_type?.trim()",
				"packagingGradeType: item.packaging_grade_type?.trim()",
				"condition: item.condition?.trim()",
			},
		},
		{
			path: "../../ui.web/src/features/tasks/components/tasks-mutate-drawer.tsx",
			required: []string{
				"data-testid='wishlist-item-type'",
				"data-testid='wishlist-condition'",
				"data-testid='wishlist-packaging-grade'",
				"itemType: data.itemType.trim()",
				"packagingGradeType: data.packagingGradeType.trim()",
				"condition: data.condition.trim()",
			},
		},
	}
	for _, file := range files {
		b, err := os.ReadFile(file.path)
		if err != nil {
			t.Fatalf("read %s: %v", file.path, err)
		}
		src := string(b)
		for _, token := range file.required {
			if !strings.Contains(src, token) {
				t.Fatalf("%s wishlist taxonomy fields contract missing token: %s", file.path, token)
			}
		}
	}
}

func TestGeneralErrorChunkRecoveryContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/errors/general-error.tsx")
	if err != nil {
		t.Fatalf("read general error component: %v", err)
	}
	src := string(b)
	required := []string{
		"dynamically imported module",
		"window.location.reload()",
		"sessionStorage",
		"cabinet.chunk-reload-once",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("general error chunk recovery missing token: %s", token)
		}
	}
}

func TestGeneralErrorSafeNavigationContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/errors/general-error.tsx")
	if err != nil {
		t.Fatalf("read general error component: %v", err)
	}
	src := string(b)
	required := []string{
		"history.go(-1)",
		"Go Back",
		"Back to Home",
		"navigate({ to: '/' })",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("general error safe navigation missing token: %s", token)
		}
	}
}

func TestCollectionWorkspaceUsesTasksProviderContract(t *testing.T) {
	t.Parallel()

	b, err := os.ReadFile("../../ui.web/src/features/collection/index.tsx")
	if err != nil {
		t.Fatalf("read collection workspace: %v", err)
	}
	src := string(b)
	required := []string{
		"TasksProvider",
		"<TasksProvider>",
		"TasksDialogs",
	}
	for _, token := range required {
		if !strings.Contains(src, token) {
			t.Fatalf("collection workspace missing tasks context token: %s", token)
		}
	}
}

func TestHeaderLanguageSwitchContract(t *testing.T) {
	t.Parallel()

	checkOrder := func(path string) {
		t.Helper()
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		src := string(b)
		langIdx := strings.Index(src, "<LanguageSwitch")
		themeIdx := strings.Index(src, "<ThemeSwitch")
		if langIdx == -1 {
			t.Fatalf("%s missing LanguageSwitch in header actions", path)
		}
		if themeIdx == -1 {
			t.Fatalf("%s missing ThemeSwitch in header actions", path)
		}
		if langIdx > themeIdx {
			t.Fatalf("%s has LanguageSwitch after ThemeSwitch; expected before", path)
		}
	}

	checkOrder("../../ui.web/src/features/dashboard/index.tsx")
	checkOrder("../../ui.web/src/features/collection/index.tsx")
	checkOrder("../../ui.web/src/features/apps/index.tsx")
}

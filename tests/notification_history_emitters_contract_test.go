package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSubmittedDataFeedbackCarriesInboxHistoryMetadata(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	toastHistoryPath := filepath.Join(root, "ui.web", "src", "lib", "toast-history.ts")
	submittedDataPath := filepath.Join(root, "ui.web", "src", "lib", "show-submitted-data.tsx")
	notificationSpecPath := filepath.Join(root, "openspec", "specs", "chats", "notification-inbox", "spec.md")
	traceabilityPath := filepath.Join(root, "openspec", "traceability.md")

	toastHistory := mustReadContractFile(t, toastHistoryPath)
	submittedData := mustReadContractFile(t, submittedDataPath)
	notificationSpec := mustReadContractFile(t, notificationSpecPath)
	traceability := mustReadContractFile(t, traceabilityPath)

	requiredToastHistorySnippets := []string{
		"type ToastHistoryMetadata",
		"historyMetadataFromOptions",
		"source_label: history.source_label",
		"category: history.category",
		"summary: history.summary || summaryFromOptions(args[1])",
	}
	for _, snippet := range requiredToastHistorySnippets {
		if !strings.Contains(toastHistory, snippet) {
			t.Fatalf("toast history capture missing submitted-data metadata support %q in %s", snippet, toastHistoryPath)
		}
	}

	requiredSubmittedDataSnippets := []string{
		"SubmittedDataHistoryOptions",
		"submittedDataSummary(data)",
		"source_label: history.source_label ?? 'Submitted Data'",
		"category: history.category ?? 'system'",
		"<code className='text-white'>{summary}</code>",
	}
	for _, snippet := range requiredSubmittedDataSnippets {
		if !strings.Contains(submittedData, snippet) {
			t.Fatalf("submitted-data helper missing durable Inbox history metadata %q in %s", snippet, submittedDataPath)
		}
	}

	if !strings.Contains(notificationSpec, "shared confirmation or warning dialogs") ||
		!strings.Contains(notificationSpec, "source label, event time, level/category metadata") {
		t.Fatalf("notification Inbox spec must keep #1438 notification-like feedback metadata contract in %s", notificationSpecPath)
	}

	if !strings.Contains(traceability, "TestSubmittedDataFeedbackCarriesInboxHistoryMetadata") {
		t.Fatalf("traceability must list submitted-data Inbox history contract test in %s", traceabilityPath)
	}
}

func TestWishlistToastFeedbackCarriesInboxHistoryMetadata(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	tasksPath := filepath.Join(root, "ui.web", "src", "features", "tasks", "index.tsx")
	taskBulkPath := filepath.Join(root, "ui.web", "src", "features", "tasks", "components", "data-table-bulk-actions.tsx")
	taskDeletePath := filepath.Join(root, "ui.web", "src", "features", "tasks", "components", "tasks-multi-delete-dialog.tsx")
	notificationSpecPath := filepath.Join(root, "openspec", "specs", "chats", "notification-inbox", "spec.md")
	traceabilityPath := filepath.Join(root, "openspec", "traceability.md")

	tasksSource := mustReadContractFile(t, tasksPath)
	taskBulkSource := mustReadContractFile(t, taskBulkPath)
	taskDeleteSource := mustReadContractFile(t, taskDeletePath)
	notificationSpec := mustReadContractFile(t, notificationSpecPath)
	traceability := mustReadContractFile(t, traceabilityPath)

	requiredTasksSnippets := []string{
		"function wishlistToastHistory(",
		"source_label: 'Wishlist / Tasks'",
		"category: 'wishlist'",
		"'wishlist-save-success'",
		"'wishlist-inline-update-failed'",
		"'wishlist-import-success'",
		"'wishlist-bulk-delete-failed'",
		"'wishlist-screenshot-save-failed'",
		"'wishlist-image-drop-success'",
		"'wishlist-barcode-save-failed'",
	}
	for _, snippet := range requiredTasksSnippets {
		if !strings.Contains(tasksSource, snippet) {
			t.Fatalf("wishlist/tasks toast feedback missing durable Inbox history metadata %q in %s", snippet, tasksPath)
		}
	}

	requiredTaskBulkSnippets := []string{
		"function taskBulkHistory(",
		"source_label: 'Task bulk actions'",
		"category: 'tasks'",
		"'tasks-bulk-status'",
		"'tasks-bulk-priority'",
		"'tasks-bulk-export'",
	}
	for _, snippet := range requiredTaskBulkSnippets {
		if !strings.Contains(taskBulkSource, snippet) {
			t.Fatalf("task bulk feedback missing durable Inbox history metadata %q in %s", snippet, taskBulkPath)
		}
	}

	requiredTaskDeleteSnippets := []string{
		"function taskDeleteHistory(",
		"source_label: 'Task delete dialog'",
		"category: 'tasks'",
		"'tasks-delete-confirmation-invalid'",
		"'tasks-bulk-delete'",
	}
	for _, snippet := range requiredTaskDeleteSnippets {
		if !strings.Contains(taskDeleteSource, snippet) {
			t.Fatalf("task delete feedback missing durable Inbox history metadata %q in %s", snippet, taskDeletePath)
		}
	}

	if !strings.Contains(notificationSpec, "Wishlist/Tasks save, import, bulk, screenshot, image drop, and barcode feedback") ||
		!strings.Contains(notificationSpec, "generic Task table bulk status, priority, export, and delete feedback") {
		t.Fatalf("notification Inbox spec must list Wishlist/Tasks feedback in #1438 emitter scope in %s", notificationSpecPath)
	}

	if !strings.Contains(traceability, "TestWishlistToastFeedbackCarriesInboxHistoryMetadata") {
		t.Fatalf("traceability must list Wishlist/Tasks Inbox history contract test in %s", traceabilityPath)
	}
}

func TestUsersBulkFeedbackCarriesInboxHistoryMetadata(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	toastHistoryPath := filepath.Join(root, "ui.web", "src", "lib", "toast-history.ts")
	usersBulkPath := filepath.Join(root, "ui.web", "src", "features", "users", "components", "data-table-bulk-actions.tsx")
	usersDeletePath := filepath.Join(root, "ui.web", "src", "features", "users", "components", "users-multi-delete-dialog.tsx")
	notificationSpecPath := filepath.Join(root, "openspec", "specs", "chats", "notification-inbox", "spec.md")
	traceabilityPath := filepath.Join(root, "openspec", "traceability.md")

	toastHistory := mustReadContractFile(t, toastHistoryPath)
	usersBulk := mustReadContractFile(t, usersBulkPath)
	usersDelete := mustReadContractFile(t, usersDeletePath)
	notificationSpec := mustReadContractFile(t, notificationSpecPath)
	traceability := mustReadContractFile(t, traceabilityPath)

	requiredToastHistorySnippets := []string{
		"promiseHistoryMetadata",
		"title: history.title ?? loading.title",
		"source_label: history.source_label",
		"category: history.category",
		"history.id ? `${history.id}-${level}` : undefined",
	}
	for _, snippet := range requiredToastHistorySnippets {
		if !strings.Contains(toastHistory, snippet) {
			t.Fatalf("toast promise history capture missing Users metadata support %q in %s", snippet, toastHistoryPath)
		}
	}

	requiredUsersSnippets := []string{
		"function usersBulkHistory(",
		"source_label: 'Users bulk actions'",
		"category: 'users'",
		"'users-bulk-invite'",
		"'users-bulk-active'",
		"'users-bulk-inactive'",
	}
	for _, snippet := range requiredUsersSnippets {
		if !strings.Contains(usersBulk, snippet) {
			t.Fatalf("Users bulk feedback missing durable Inbox history metadata %q in %s", snippet, usersBulkPath)
		}
	}

	requiredDeleteSnippets := []string{
		"function usersDeleteHistory(",
		"source_label: 'Users delete dialog'",
		"category: 'users'",
		"'users-delete-confirmation-invalid'",
		"'users-bulk-delete'",
	}
	for _, snippet := range requiredDeleteSnippets {
		if !strings.Contains(usersDelete, snippet) {
			t.Fatalf("Users delete feedback missing durable Inbox history metadata %q in %s", snippet, usersDeletePath)
		}
	}

	if !strings.Contains(notificationSpec, "Users admin bulk action/delete confirmation feedback") ||
		!strings.Contains(notificationSpec, "Promise feedback MUST preserve configured source/category metadata") {
		t.Fatalf("notification Inbox spec must list Users feedback in #1438 emitter scope in %s", notificationSpecPath)
	}

	if !strings.Contains(traceability, "TestUsersBulkFeedbackCarriesInboxHistoryMetadata") {
		t.Fatalf("traceability must list Users Inbox history contract test in %s", traceabilityPath)
	}
}

func TestAuthAndGlobalErrorFeedbackCarriesInboxHistoryMetadata(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	signInPath := filepath.Join(root, "ui.web", "src", "features", "auth", "sign-in", "components", "user-auth-form.tsx")
	signUpPath := filepath.Join(root, "ui.web", "src", "features", "auth", "sign-up", "components", "sign-up-form.tsx")
	forgotPasswordPath := filepath.Join(root, "ui.web", "src", "features", "auth", "forgot-password", "components", "forgot-password-form.tsx")
	handleServerErrorPath := filepath.Join(root, "ui.web", "src", "lib", "handle-server-error.ts")
	mainPath := filepath.Join(root, "ui.web", "src", "main.tsx")
	notificationSpecPath := filepath.Join(root, "openspec", "specs", "chats", "notification-inbox", "spec.md")
	traceabilityPath := filepath.Join(root, "openspec", "traceability.md")

	signIn := mustReadContractFile(t, signInPath)
	signUp := mustReadContractFile(t, signUpPath)
	forgotPassword := mustReadContractFile(t, forgotPasswordPath)
	handleServerError := mustReadContractFile(t, handleServerErrorPath)
	mainSource := mustReadContractFile(t, mainPath)
	notificationSpec := mustReadContractFile(t, notificationSpecPath)
	traceability := mustReadContractFile(t, traceabilityPath)

	requiredSignInSnippets := []string{
		"source_label: 'Auth sign-in'",
		"category: 'auth'",
		"'auth-sign-in'",
		"'auth-passkey-sign-in-failed'",
		"recordNotificationHistory({",
	}
	for _, snippet := range requiredSignInSnippets {
		if !strings.Contains(signIn, snippet) {
			t.Fatalf("auth sign-in feedback missing durable Inbox history metadata %q in %s", snippet, signInPath)
		}
	}

	requiredAuthPromiseSnippets := map[string][]string{
		signUpPath: {
			"source_label: 'Auth sign-up'",
			"category: 'auth'",
			"'auth-sign-up'",
		},
		forgotPasswordPath: {
			"source_label: 'Auth forgot password'",
			"category: 'auth'",
			"'auth-forgot-password'",
		},
	}
	for path, snippets := range requiredAuthPromiseSnippets {
		source := signUp
		if path == forgotPasswordPath {
			source = forgotPassword
		}
		for _, snippet := range snippets {
			if !strings.Contains(source, snippet) {
				t.Fatalf("auth feedback missing durable Inbox history metadata %q in %s", snippet, path)
			}
		}
	}

	requiredGlobalSnippets := []string{
		"source_label: 'Global server error'",
		"source_label: 'Global query client'",
		"category: 'system'",
		"category: 'auth'",
	}
	combinedGlobal := handleServerError + "\n" + mainSource
	for _, snippet := range requiredGlobalSnippets {
		if !strings.Contains(combinedGlobal, snippet) {
			t.Fatalf("global error feedback missing durable Inbox history metadata %q in %s or %s", snippet, handleServerErrorPath, mainPath)
		}
	}

	if !strings.Contains(notificationSpec, "Auth entry sign-in/sign-up/forgot-password/passkey feedback") ||
		!strings.Contains(notificationSpec, "global query/server error feedback") {
		t.Fatalf("notification Inbox spec must list Auth/global feedback in #1438 emitter scope in %s", notificationSpecPath)
	}

	if !strings.Contains(traceability, "TestAuthAndGlobalErrorFeedbackCarriesInboxHistoryMetadata") {
		t.Fatalf("traceability must list Auth/global Inbox history contract test in %s", traceabilityPath)
	}
}

func TestIntegrationsFeedbackCarriesInboxHistoryMetadata(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	appsPath := filepath.Join(root, "ui.web", "src", "features", "apps", "index.tsx")
	notificationSpecPath := filepath.Join(root, "openspec", "specs", "chats", "notification-inbox", "spec.md")
	traceabilityPath := filepath.Join(root, "openspec", "traceability.md")

	appsSource := mustReadContractFile(t, appsPath)
	notificationSpec := mustReadContractFile(t, notificationSpecPath)
	traceability := mustReadContractFile(t, traceabilityPath)

	requiredAppsSnippets := []string{
		"function recordIntegrationsStatusHistory(",
		"source_label: 'Integrations'",
		"category: 'system'",
		"'integrations-openai-save-success'",
		"integrations-token-field-${notificationHistoryID(message)}",
		"integrations-provider-save-${notificationHistoryID(editingProvider.provider_id)}",
		"'integrations-openai-api-key-disconnect-success'",
		"integrations-buyer-interest-${mode}-success",
		"integrations-seller-operation-${notificationHistoryID(status.operation)}-${confirmed ? 'confirmed-' : ''}preview-success",
		"integrations-seller-operation-${notificationHistoryID(status.operation)}-${confirmed ? 'confirmed-' : ''}execute-failed",
		"integrations-listing-lifecycle-${notificationHistoryID(command)}-${confirmed ? 'confirmed-' : ''}preview-success",
		"integrations-listing-lifecycle-${notificationHistoryID(command)}-${confirmed ? 'confirmed-' : ''}execute-failed",
		"'integrations-landed-cost-preview-success'",
		"'integrations-landed-cost-preview-failed'",
	}
	for _, snippet := range requiredAppsSnippets {
		if !strings.Contains(appsSource, snippet) {
			t.Fatalf("Integrations feedback missing durable Inbox history metadata %q in %s", snippet, appsPath)
		}
	}

	if !strings.Contains(notificationSpec, "Integrations provider configuration save/disconnect feedback") ||
		!strings.Contains(notificationSpec, "Integrations buyer-interest, seller operation, listing lifecycle, and landed-cost action feedback") {
		t.Fatalf("notification Inbox spec must list Integrations action feedback in #1438 emitter scope in %s", notificationSpecPath)
	}

	if !strings.Contains(traceability, "TestIntegrationsFeedbackCarriesInboxHistoryMetadata") {
		t.Fatalf("traceability must list Integrations Inbox history contract test in %s", traceabilityPath)
	}
}

func mustReadContractFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

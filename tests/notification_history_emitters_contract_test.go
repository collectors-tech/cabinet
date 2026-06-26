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
	notificationSpecPath := filepath.Join(root, "openspec", "specs", "chats", "notification-inbox", "spec.md")
	traceabilityPath := filepath.Join(root, "openspec", "traceability.md")

	tasksSource := mustReadContractFile(t, tasksPath)
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

	if !strings.Contains(notificationSpec, "Wishlist/Tasks save, import, bulk, screenshot, image drop, and barcode feedback") {
		t.Fatalf("notification Inbox spec must list Wishlist/Tasks feedback in #1438 emitter scope in %s", notificationSpecPath)
	}

	if !strings.Contains(traceability, "TestWishlistToastFeedbackCarriesInboxHistoryMetadata") {
		t.Fatalf("traceability must list Wishlist/Tasks Inbox history contract test in %s", traceabilityPath)
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

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

func mustReadContractFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

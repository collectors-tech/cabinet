package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssistantUITargetRegistryBindsInventoryWalkthroughTargets(t *testing.T) {
	t.Parallel()

	registryPath := filepath.Join("..", "ui.web", "src", "lib", "ui-target-registry.ts")
	registryRaw, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatalf("read UI target registry: %v", err)
	}
	registry := string(registryRaw)

	requiredRegistryFragments := []string{
		"id: 'inventory.surface'",
		"id: 'inventory.item.row'",
		"id: 'inventory.item.actions'",
		"id: 'inventory.item.editor'",
		"id: 'inventory.item.title'",
		"id: 'inventory.item.status'",
		"id: 'inventory.item.category'",
		"id: 'inventory.item.save'",
		"id: 'inventory.item.cancel'",
		"selector: '[data-testid=\"inventory-edit-title\"]'",
		"safeActions: ['highlight', 'scroll', 'focus']",
		"fallbackInstruction:",
		"calloutPlacement:",
	}
	for _, fragment := range requiredRegistryFragments {
		if !strings.Contains(registry, fragment) {
			t.Fatalf("expected UI target registry to include %q", fragment)
		}
	}

	overlayPath := filepath.Join("..", "ui.web", "src", "components", "guidance", "guidance-overlay.tsx")
	overlayRaw, err := os.ReadFile(overlayPath)
	if err != nil {
		t.Fatalf("read guidance overlay: %v", err)
	}
	overlay := string(overlayRaw)
	for _, fragment := range []string{
		"uiGuidanceEventName",
		"data-testid='ui-guidance-overlay'",
		"data-testid='ui-guidance-highlight'",
		"data-testid='ui-guidance-fallback'",
		"aria-label='Cancel guidance'",
		"aria-label='Previous guidance step'",
		"aria-label='Next guidance step'",
	} {
		if !strings.Contains(overlay, fragment) {
			t.Fatalf("expected guidance overlay to include %q", fragment)
		}
	}
}

func TestAssistantUITargetTraceabilityIsImplemented(t *testing.T) {
	t.Parallel()

	tracePath := filepath.Join("..", "openspec", "traceability.md")
	traceRaw, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}

	var row string
	for _, line := range strings.Split(string(traceRaw), "\n") {
		if strings.HasPrefix(line, "| `ASSISTANT-EXECUTION-012` ") {
			row = line
			break
		}
	}
	if row == "" {
		t.Fatal("expected traceability row for ASSISTANT-EXECUTION-012")
	}

	for _, fragment := range []string{
		"#1511",
		"ui.web/src/lib/ui-target-registry.ts",
		"ui.web/src/components/guidance/guidance-overlay.tsx",
		"inventory.item.title",
		"TestAssistantUITargetRegistryBindsInventoryWalkthroughTargets",
		"| implemented |",
	} {
		if !strings.Contains(row, fragment) {
			t.Fatalf("expected ASSISTANT-EXECUTION-012 traceability row to include %q; row: %s", fragment, row)
		}
	}
}

func TestAssistantShellCommandBusTraceabilityIsImplemented(t *testing.T) {
	t.Parallel()

	busPath := filepath.Join("..", "ui.web", "src", "lib", "shell-command-bus.ts")
	busRaw, err := os.ReadFile(busPath)
	if err != nil {
		t.Fatalf("read shell command bus: %v", err)
	}
	bus := string(busRaw)
	for _, fragment := range []string{
		"'navigate.open_surface'",
		"'ui.highlight_target'",
		"'ui.clear_guidance'",
		"'walkthrough.cancel'",
		"ShellCommandEvent",
		"allowedRoutes",
		"requestUiGuidance",
		"clearUiGuidance",
		"Route is not in the Cabinet app-control allowlist",
		"UI target is not registered for guided walkthroughs",
	} {
		if !strings.Contains(bus, fragment) {
			t.Fatalf("expected shell command bus to include %q", fragment)
		}
	}

	panelPath := filepath.Join("..", "ui.web", "src", "components", "layout", "assistant-workspace-panel.tsx")
	panelRaw, err := os.ReadFile(panelPath)
	if err != nil {
		t.Fatalf("read assistant workspace panel: %v", err)
	}
	panel := string(panelRaw)
	for _, fragment := range []string{
		"dispatchShellCommand",
		"data-testid='shell-assistant-command-timeline'",
		"data-testid='shell-assistant-command-event'",
		"data-command-type={event.type}",
		"data-command-status={event.status}",
		"setActiveWorkspace('assistant')",
	} {
		if !strings.Contains(panel, fragment) {
			t.Fatalf("expected assistant workspace panel to include %q", fragment)
		}
	}

	tracePath := filepath.Join("..", "openspec", "traceability.md")
	traceRaw, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}
	trace := string(traceRaw)
	for _, fragment := range []string{
		"`ASSISTANT-EXECUTION-013`",
		"`ASSISTANT-WORKSPACE-008`",
		"#1512/#1514",
		"`ui.web/src/lib/shell-command-bus.ts`",
		"TestAssistantShellCommandBusTraceabilityIsImplemented",
		"| implemented |",
	} {
		if !strings.Contains(trace, fragment) {
			t.Fatalf("expected shell command bus traceability to include %q", fragment)
		}
	}
}

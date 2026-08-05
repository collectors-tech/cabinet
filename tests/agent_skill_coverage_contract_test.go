package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentSkillCoverageMatrixCoversRequiredSurfacesAndFields(t *testing.T) {
	t.Parallel()

	matrixPath := filepath.Join("..", "openspec", "traceability", "agent-skill-coverage.md")
	raw, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("read agent skill coverage matrix: %v", err)
	}
	content := string(raw)

	requiredFragments := []string{
		"Issue: #1702",
		"Parent: #1701",
		"Related registry epic: #1666",
		"Main Chat `/chats`",
		"Side-panel Chat",
		"Inbox review",
		"Telegram/external channel",
		"Skills page detail/actions",
		"| Dashboard |",
		"| Inventory |",
		"| Wishlist |",
		"| Collections |",
		"| Media |",
		"| Discoveries |",
		"| Market Watch / Scanner |",
		"| Purchases |",
		"| Integrations |",
		"| Inbox |",
		"| Users / workspace admin |",
		"| Settings / Profile |",
		"| Settings / Account |",
		"| Settings / Appearance |",
		"| Settings / Storage |",
		"| Import / Export / Backup / Restore / Maintenance |",
		"| Chats / Agent itself |",
		"Expected skill ids",
		"User-facing request examples",
		"Safety",
		"Required context/selection",
		"Required setup",
		"Bound capability ids",
		"Bound guided workflow ids",
		"Missing issue/PR links",
		"Validation evidence",
		"Blocked/deferred reason",
		"#1707",
		"#1708",
		"#1709",
		"#1710",
		"#1711",
		"#1715",
		"#1773",
		"cabinet.navigate.open_surface",
		"cabinet.guided.inventory.update_item",
		"cabinet.inbox.summarise_unhandled",
		"cabinet.users.remove_user",
		"Marketplace discovery/publishing/payments/reviews remain deferred",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(content, fragment) {
			t.Fatalf("expected agent skill coverage matrix to include %q", fragment)
		}
	}
}

func TestAgentSkillCoverageTraceabilityStaysBoundToOpenSpec(t *testing.T) {
	t.Parallel()

	tracePath := filepath.Join("..", "openspec", "traceability.md")
	traceRaw, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("read traceability: %v", err)
	}
	trace := string(traceRaw)

	for _, fragment := range []string{
		"AGENT-SKILL-COVERAGE-001",
		"#1701/#1702/#1666",
		"openspec/traceability/agent-skill-coverage.md",
		"TestAgentSkillCoverageMatrixCoversRequiredSurfacesAndFields",
		"TestAgentSkillCoverageTraceabilityStaysBoundToOpenSpec",
		"TestAgentSkillCoverageMatrixBindsSkillsPageToMergedIssue1670Evidence",
		"TestAgentSkillCoverageMatrixBindsInboxReviewToIssue1987Evidence",
		"TestAgentSkillCoverageMatrixBindsInboxMissingStaleContextToIssue1987Evidence",
		"TestAgentSkillCoverageMatrixBindsSettingsProfilePersistenceToIssue2009Evidence",
		"TestAgentSkillCoverageMatrixBindsSettingsProfileUIChannelDispatchToIssue2011Evidence",
		"TestAgentSkillCoverageMatrixBindsSettingsStorageUIChannelDispatchToIssue2019Evidence",
		"TestAgentSkillCoverageMatrixBindsSettingsDataMaintenanceUIChannelDispatchToIssue2021Evidence",
		"#1985 reconciles the Skills page detail/actions entry point",
		"#1987 binds Inbox review Agent launch context",
		"#1989 reconciles Inbox missing/stale context evidence",
		"#2009 reconciles Settings/Profile",
		"#2011 reconciles Settings/Profile UI/channel dispatch",
		"#2019 reconciles Settings Storage UI/channel dispatch",
		"#2021 reconciles Settings Data/Maintenance UI/channel dispatch",
		"main Chat, side-panel Chat, Inbox review, and Telegram/external channels",
		"| partial |",
	} {
		if !strings.Contains(trace, fragment) {
			t.Fatalf("expected agent skill coverage traceability to include %q", fragment)
		}
	}

	specPath := filepath.Join("..", "openspec", "specs", "agent-skills-registry", "spec.md")
	specRaw, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read agent skill registry spec: %v", err)
	}
	spec := string(specRaw)

	for _, fragment := range []string{
		"### Requirement: Agent skill coverage SHALL stay mapped by Cabinet surface",
		"#### Scenario: Maintain per-surface coverage matrix",
		"skill ids, safety levels, dependencies, statuses, issue links, and validation evidence",
		"distinguish in-app main Chat, side-panel Chat, Inbox review, and Telegram/external channels",
	} {
		if !strings.Contains(spec, fragment) {
			t.Fatalf("expected agent skill registry spec to include %q", fragment)
		}
	}
}

func TestAgentSkillCoverageMatrixBindsInventoryGuidedUpdateToIssue1513Evidence(t *testing.T) {
	t.Parallel()

	matrixPath := filepath.Join("..", "openspec", "traceability", "agent-skill-coverage.md")
	raw, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("read agent skill coverage matrix: %v", err)
	}
	content := string(raw)

	var inventoryRow string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "| Inventory |") {
			inventoryRow = line
			break
		}
	}
	if inventoryRow == "" {
		t.Fatalf("expected Inventory surface coverage row")
	}
	for _, required := range []string{
		"#1513",
		"`cabinet.guided.inventory.update_item`",
		"`guided.inventory.update_item`",
		"TestGuidedWorkflowRegistryMatchesInventoryItemUpdateRecipe",
		"TestChatMessageAppControlPlannerStartsGuidedInventoryWalkthrough",
		"TestGuidedInventoryUpdatePersistsTimelineAndConfirmedMutation",
		"guided inventory update proof",
		"live Telegram production-channel validation remains #1773",
		"broader external intake adapter routing remains planned",
	} {
		if !strings.Contains(inventoryRow, required) {
			t.Fatalf("Inventory row must include #1513 guided update evidence %q; row: %s", required, inventoryRow)
		}
	}
	for _, stale := range []string{
		"Guided update remains blocked until #1513",
		"guided update remains blocked until #1513",
	} {
		if strings.Contains(inventoryRow, stale) {
			t.Fatalf("Inventory row must not keep stale guided update blocker %q; row: %s", stale, inventoryRow)
		}
	}
}

func TestAgentSkillCoverageMatrixBindsSettingsProfilePersistenceToIssue2009Evidence(t *testing.T) {
	t.Parallel()

	matrixPath := filepath.Join("..", "openspec", "traceability", "agent-skill-coverage.md")
	raw, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("read agent skill coverage matrix: %v", err)
	}
	content := string(raw)

	var settingsProfileRow string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "| Settings / Profile |") {
			settingsProfileRow = line
			break
		}
	}
	if settingsProfileRow == "" {
		t.Fatalf("expected Settings / Profile surface coverage row")
	}
	for _, required := range []string{
		"#1711, #2009",
		"`cabinet.settings.update_profile`",
		"structured `settings_profile` payload",
		"source surface/channel/thread/message context",
		"TestAgentSkillApplyAPIHandlesSettingsProfilePersistenceEvidence",
		"no preview-time profile settings mutation",
		"confirmed structured `settings_profile` persistence into `profile_settings`",
		"key-only `settings_persisted` evidence",
		"raw setting value redaction",
		"broader Settings Account/Storage/Data and live Telegram/external-channel validation remain separate",
	} {
		if !strings.Contains(settingsProfileRow, required) {
			t.Fatalf("Settings/Profile row must include %q; row: %s", required, settingsProfileRow)
		}
	}
	for _, stale := range []string{
		"persisted profile settings integration remain planned",
		"persisted profile settings integration remains planned",
	} {
		if strings.Contains(settingsProfileRow, stale) {
			t.Fatalf("Settings/Profile row must not keep stale planned wording %q; row: %s", stale, settingsProfileRow)
		}
	}

	for _, required := range []string{
		"The #2009 Settings/Profile slice adds focused direct API proof",
		"without claiming broader Settings Account/Storage/Data coverage, live Telegram/external-channel validation, or #1701 parent closure",
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("coverage matrix must include #2009 narrative %q", required)
		}
	}
}

func TestAgentSkillCoverageMatrixBindsSettingsProfileUIChannelDispatchToIssue2011Evidence(t *testing.T) {
	t.Parallel()

	matrixPath := filepath.Join("..", "openspec", "traceability", "agent-skill-coverage.md")
	raw, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("read agent skill coverage matrix: %v", err)
	}
	content := string(raw)

	var settingsProfileRow string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "| Settings / Profile |") {
			settingsProfileRow = line
			break
		}
	}
	if settingsProfileRow == "" {
		t.Fatalf("expected Settings / Profile surface coverage row")
	}
	for _, required := range []string{
		"#1711, #2009, #2011",
		"`cabinet.settings.update_profile`",
		"source surface/channel/thread/message context",
		"ui.web/cypress/e2e/chats/assistant-workspace-agent-skills/spec.cy.ts",
		"ASSISTANT-WORKSPACE-016/#2011 dispatches Settings Profile Agent Skills with in-app source context",
		"side-panel Agent Skill preview/apply dispatch",
		"`source_channel=in-app`",
		"structured `settings_profile`",
		"private-note redaction",
		"broader Settings Account/Storage/Data and live Telegram/external-channel validation remain separate",
	} {
		if !strings.Contains(settingsProfileRow, required) {
			t.Fatalf("Settings/Profile row must include #2011 UI/channel evidence %q; row: %s", required, settingsProfileRow)
		}
	}
	for _, stale := range []string{
		"UI/channel dispatch for Settings/Profile remains planned",
		"Settings/Profile UI/channel dispatch remains planned separately",
	} {
		if strings.Contains(content, stale) {
			t.Fatalf("coverage matrix must not keep stale Settings/Profile UI/channel planned wording %q", stale)
		}
	}

	for _, required := range []string{
		"The #2011 Settings/Profile UI/channel dispatch slice adds focused side-panel proof",
		"without claiming broader Settings Account/Storage/Data coverage, live Telegram/external-channel validation, or #1701 parent closure",
		"Settings/Profile direct API persistence is implemented under #2009 and side-panel UI/channel dispatch is implemented under #2011",
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("coverage matrix must include #2011 narrative %q", required)
		}
	}
}

func TestAgentSkillCoverageMatrixBindsSettingsAccountPersistenceToIssue2015Evidence(t *testing.T) {
	t.Parallel()

	matrixPath := filepath.Join("..", "openspec", "traceability", "agent-skill-coverage.md")
	raw, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("read agent skill coverage matrix: %v", err)
	}
	content := string(raw)

	var settingsAccountRow string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "| Settings / Account |") {
			settingsAccountRow = line
			break
		}
	}
	if settingsAccountRow == "" {
		t.Fatalf("expected Settings / Account surface coverage row")
	}
	for _, required := range []string{
		"#1711, #2015, #2017",
		"`cabinet.settings.update_account`",
		"structured `settings_account` payload",
		"source surface/channel/thread/message context",
		"TestAgentSkillApplyAPIHandlesSettingsAccountPersistenceEvidence",
		"no preview-time account settings mutation",
		"confirmed structured `settings_account` persistence into `profile_settings`",
		"key-only `settings_persisted` evidence",
		"raw setting value redaction",
		"Settings Storage/Data and live Telegram/external-channel validation remain separate",
	} {
		if !strings.Contains(settingsAccountRow, required) {
			t.Fatalf("Settings/Account row must include %q; row: %s", required, settingsAccountRow)
		}
	}
	for _, stale := range []string{
		"persisted account settings integration remain planned",
		"persisted account settings integration remains planned",
	} {
		if strings.Contains(settingsAccountRow, stale) {
			t.Fatalf("Settings/Account row must not keep stale planned wording %q; row: %s", stale, settingsAccountRow)
		}
	}

	for _, required := range []string{
		"The #2015 Settings/Account slice adds focused direct API proof",
		"without claiming broader Settings Storage/Data coverage, live Telegram/external-channel validation, or #1701 parent closure",
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("coverage matrix must include #2015 narrative %q", required)
		}
	}
}

func TestAgentSkillCoverageMatrixBindsSettingsAccountUIChannelDispatchToIssue2017Evidence(t *testing.T) {
	t.Parallel()

	matrixPath := filepath.Join("..", "openspec", "traceability", "agent-skill-coverage.md")
	raw, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("read agent skill coverage matrix: %v", err)
	}
	content := string(raw)

	var settingsAccountRow string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "| Settings / Account |") {
			settingsAccountRow = line
			break
		}
	}
	if settingsAccountRow == "" {
		t.Fatalf("expected Settings / Account surface coverage row")
	}
	for _, required := range []string{
		"#1711, #2015, #2017",
		"`cabinet.settings.update_account`",
		"source surface/channel/thread/message context",
		"ui.web/cypress/e2e/chats/assistant-workspace-agent-skills/spec.cy.ts",
		"ASSISTANT-WORKSPACE-016/#2017 dispatches Settings Account Agent Skills with in-app source context",
		"side-panel Agent Skill preview/apply dispatch",
		"`source_channel=in-app`",
		"structured `settings_account` parameters",
		"private-note redaction",
		"Settings Storage/Data and live Telegram/external-channel validation remain separate",
	} {
		if !strings.Contains(settingsAccountRow, required) {
			t.Fatalf("Settings/Account row must include #2017 UI/channel evidence %q; row: %s", required, settingsAccountRow)
		}
	}
	for _, stale := range []string{
		"Settings Account UI/channel dispatch, Settings Storage/Data, and live Telegram/external-channel validation remain separate",
		"Settings Account UI/channel dispatch remains planned separately",
	} {
		if strings.Contains(content, stale) {
			t.Fatalf("coverage matrix must not keep stale Settings/Account UI/channel planned wording %q", stale)
		}
	}

	for _, required := range []string{
		"The #2017 Settings/Account UI/channel dispatch slice adds focused side-panel proof",
		"without claiming broader Settings Storage/Data coverage, live Telegram/external-channel validation, or #1701 parent closure",
		"Settings/Account side-panel UI/channel dispatch is implemented under #2017",
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("coverage matrix must include #2017 narrative %q", required)
		}
	}
}

func TestAgentSkillCoverageMatrixBindsSettingsAppearanceUIChannelDispatchToIssue2013Evidence(t *testing.T) {
	t.Parallel()

	matrixPath := filepath.Join("..", "openspec", "traceability", "agent-skill-coverage.md")
	raw, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("read agent skill coverage matrix: %v", err)
	}
	content := string(raw)

	var settingsAppearanceRow string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "| Settings / Appearance |") {
			settingsAppearanceRow = line
			break
		}
	}
	if settingsAppearanceRow == "" {
		t.Fatalf("expected Settings / Appearance surface coverage row")
	}
	for _, required := range []string{
		"#1711, #2013",
		"`cabinet.settings.update_appearance`",
		"source surface/channel/thread/message context",
		"ui.web/cypress/e2e/chats/assistant-workspace-agent-skills/spec.cy.ts",
		"ASSISTANT-WORKSPACE-016/#2013 dispatches Settings Appearance Agent Skills with in-app source context",
		"side-panel Agent Skill preview/apply dispatch",
		"`source_channel=in-app`",
		"`setting_key`/`setting_scope`/`setting_value` parameters",
		"setting-value redaction",
		"broader Settings Account/Storage/Data and live Telegram/external-channel validation remain separate",
	} {
		if !strings.Contains(settingsAppearanceRow, required) {
			t.Fatalf("Settings/Appearance row must include #2013 UI/channel evidence %q; row: %s", required, settingsAppearanceRow)
		}
	}
	for _, stale := range []string{
		"UI/channel dispatch and broader settings surface coverage remain planned",
		"Settings/Appearance UI/channel dispatch remains planned separately",
	} {
		if strings.Contains(content, stale) {
			t.Fatalf("coverage matrix must not keep stale Settings/Appearance UI/channel planned wording %q", stale)
		}
	}

	for _, required := range []string{
		"The #2013 Settings/Appearance UI/channel dispatch slice adds focused side-panel proof",
		"without claiming broader Settings Account/Storage/Data coverage, live Telegram/external-channel validation, or #1701 parent closure",
		"Settings/Appearance side-panel UI/channel dispatch is implemented under #2013",
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("coverage matrix must include #2013 narrative %q", required)
		}
	}
}

func TestAgentSkillCoverageMatrixBindsSettingsStorageUIChannelDispatchToIssue2019Evidence(t *testing.T) {
	t.Parallel()

	matrixPath := filepath.Join("..", "openspec", "traceability", "agent-skill-coverage.md")
	raw, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("read agent skill coverage matrix: %v", err)
	}
	content := string(raw)

	var settingsStorageRow string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "| Settings / Storage |") {
			settingsStorageRow = line
			break
		}
	}
	if settingsStorageRow == "" {
		t.Fatalf("expected Settings / Storage surface coverage row")
	}
	for _, required := range []string{
		"#1711, #2019",
		"`cabinet.storage.configure_backup`",
		"source surface/channel/thread/message context",
		"ui.web/cypress/e2e/chats/assistant-workspace-agent-skills/spec.cy.ts",
		"ASSISTANT-WORKSPACE-016/#2019 dispatches Settings Storage Agent Skills with in-app source context",
		"side-panel Agent Skill preview/apply dispatch",
		"`source_channel=in-app`",
		"`backup_target`/`backup_schedule`/`storage_note` parameters",
		"private-note redaction",
		"local filesystem path redaction",
		"restore drill, Settings Data/Maintenance, and live Telegram/external-channel validation remain separate",
	} {
		if !strings.Contains(settingsStorageRow, required) {
			t.Fatalf("Settings/Storage row must include #2019 UI/channel evidence %q; row: %s", required, settingsStorageRow)
		}
	}
	for _, stale := range []string{
		"UI/channel dispatch and restore drill evidence remain planned",
		"Settings Storage UI/channel dispatch remains planned separately",
	} {
		if strings.Contains(content, stale) {
			t.Fatalf("coverage matrix must not keep stale Settings/Storage UI/channel planned wording %q", stale)
		}
	}

	for _, required := range []string{
		"The #2019 Settings/Storage UI/channel dispatch slice adds focused side-panel proof",
		"without claiming restore drill, Settings Data/Maintenance coverage, live Telegram/external-channel validation, or #1701 parent closure",
		"Settings/Storage side-panel UI/channel dispatch is implemented under #2019",
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("coverage matrix must include #2019 narrative %q", required)
		}
	}
}

func TestAgentSkillCoverageMatrixBindsSettingsDataMaintenanceUIChannelDispatchToIssue2021Evidence(t *testing.T) {
	t.Parallel()

	matrixPath := filepath.Join("..", "openspec", "traceability", "agent-skill-coverage.md")
	raw, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("read agent skill coverage matrix: %v", err)
	}
	content := string(raw)

	var dataMaintenanceRow string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "| Import / Export / Backup / Restore / Maintenance |") {
			dataMaintenanceRow = line
			break
		}
	}
	if dataMaintenanceRow == "" {
		t.Fatalf("expected Import / Export / Backup / Restore / Maintenance surface coverage row")
	}
	for _, required := range []string{
		"#1711, #2021, #2023",
		"`cabinet.maintenance.run_safe_check`",
		"source surface/channel/thread/message context",
		"ui.web/cypress/e2e/chats/assistant-workspace-agent-skills/spec.cy.ts",
		"ASSISTANT-WORKSPACE-016/#2021 dispatches Settings Data/Maintenance Agent Skills with in-app source context",
		"side-panel Agent Skill preview/apply dispatch",
		"`source_channel=in-app`",
		"`maintenance_scope`/`check_level`/`maintenance_note` parameters",
		"private-note redaction",
		"local filesystem path redaction",
		"no external-write claim",
		"Live Telegram/external-channel validation remains separate",
	} {
		if !strings.Contains(dataMaintenanceRow, required) {
			t.Fatalf("Settings Data/Maintenance row must include #2021 UI/channel evidence %q; row: %s", required, dataMaintenanceRow)
		}
	}
	for _, stale := range []string{
		"stronger confirmation copy in UI/channel surfaces, and restore drill evidence remain planned",
		"broader Settings Data/Maintenance, and live Telegram/external-channel validation remain separate",
	} {
		if strings.Contains(content, stale) {
			t.Fatalf("coverage matrix must not keep stale Settings Data/Maintenance UI/channel wording %q", stale)
		}
	}

	for _, required := range []string{
		"The #2021 Settings Data/Maintenance UI/channel dispatch slice adds focused side-panel proof",
		"without claiming real import/restore persistence, destructive restore drill, live Telegram/external-channel validation, packaged/release acceptance, or #1701 parent closure",
		"Settings Data/Maintenance side-panel UI/channel dispatch is implemented under #2021",
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("coverage matrix must include #2021 narrative %q", required)
		}
	}
}

func TestAgentSkillCoverageMatrixBindsDataImportRestorePersistenceToIssue2023Evidence(t *testing.T) {
	t.Parallel()

	matrixPath := filepath.Join("..", "openspec", "traceability", "agent-skill-coverage.md")
	raw, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("read agent skill coverage matrix: %v", err)
	}
	content := string(raw)

	var dataMaintenanceRow string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "| Import / Export / Backup / Restore / Maintenance |") {
			dataMaintenanceRow = line
			break
		}
	}
	if dataMaintenanceRow == "" {
		t.Fatalf("expected Import / Export / Backup / Restore / Maintenance surface coverage row")
	}
	for _, required := range []string{
		"#2023",
		"`cabinet.data.import_file`",
		"`cabinet.data.restore_backup`",
		"TestAgentSkillApplyAPIHandlesDataImportRestorePersistenceEvidence",
		"confirmed import persistence for selected JSON fixtures",
		"profile-scoped setting/item changes",
		"key-only persisted-setting evidence",
		"selected path/raw payload redaction",
		"restore-drill SHA-256/byte-count integrity evidence",
		"profile isolation without touching live user data",
	} {
		if !strings.Contains(dataMaintenanceRow, required) {
			t.Fatalf("Settings Data/Maintenance row must include #2023 persistence evidence %q; row: %s", required, dataMaintenanceRow)
		}
	}
	for _, stale := range []string{
		"Real import/restore persistence, destructive restore drill, and live Telegram/external-channel validation remain separate",
	} {
		if strings.Contains(content, stale) {
			t.Fatalf("coverage matrix must not keep stale #2023 gap wording %q", stale)
		}
	}
	for _, required := range []string{
		"The #2023 Settings Data import/restore drill slice adds focused direct API proof",
		"confirmed import from an explicit selected JSON fixture persists whitelisted profile settings and canonical item rows",
		"confirmed restore drill verifies selected backup bytes with SHA-256/byte-count integrity evidence",
		"without claiming live Telegram/external-channel validation, packaged/release acceptance, or #1701 parent closure",
		"Settings Data import/restore drill proof is implemented under #2023",
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("coverage matrix must include #2023 narrative %q", required)
		}
	}
}

func TestAgentSkillCoverageMatrixBindsChatActionTimelineToIssue2029Evidence(t *testing.T) {
	t.Parallel()

	matrixPath := filepath.Join("..", "openspec", "traceability", "agent-skill-coverage.md")
	raw, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("read agent skill coverage matrix: %v", err)
	}
	content := string(raw)

	var chatAgentRow string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "| Chats / Agent itself |") {
			chatAgentRow = line
			break
		}
	}
	if chatAgentRow == "" {
		t.Fatalf("expected Chats / Agent itself surface coverage row")
	}
	for _, required := range []string{
		"#2029",
		"`cabinet.chat.action_timeline.view`",
		"TestAgentSkillApplyAPIHandlesChatActionTimelineSkill",
		"scoped action-timeline evidence",
		"read-only",
		"without mutation previews or confirmation tokens",
		"live Telegram/external-channel validation remains separate",
	} {
		if !strings.Contains(chatAgentRow, required) {
			t.Fatalf("Chats / Agent itself row must include #2029 action-timeline evidence %q; row: %s", required, chatAgentRow)
		}
	}
	for _, stale := range []string{
		"chat-action-timeline planned",
		"action timeline and broader acceptance remain tracked by #1703/#1716",
	} {
		if strings.Contains(chatAgentRow, stale) {
			t.Fatalf("Chats / Agent itself row must not keep stale action-timeline wording %q; row: %s", stale, chatAgentRow)
		}
	}
}

func TestAgentSkillCoverageMatrixBindsInboxReviewToIssue1987Evidence(t *testing.T) {
	t.Parallel()

	matrixPath := filepath.Join("..", "openspec", "traceability", "agent-skill-coverage.md")
	raw, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("read agent skill coverage matrix: %v", err)
	}
	content := string(raw)

	var inboxReviewRow string
	var inboxSurfaceRow string
	for _, line := range strings.Split(content, "\n") {
		switch {
		case strings.HasPrefix(line, "| Inbox review |"):
			inboxReviewRow = line
		case strings.HasPrefix(line, "| Inbox |"):
			inboxSurfaceRow = line
		}
	}
	if inboxReviewRow == "" {
		t.Fatalf("expected Inbox review channel coverage row")
	}
	if inboxSurfaceRow == "" {
		t.Fatalf("expected Inbox surface coverage row")
	}

	for _, row := range []string{inboxReviewRow, inboxSurfaceRow} {
		for _, fragment := range []string{
			"#1987",
			"assistant-inbox-agent-context/spec.cy.ts",
			"AGENT-UNIVERSAL-CHANNELS-001/#1987 preserves Inbox review notification context in the Agent envelope",
			"notification-inbox-open-agent",
			"selected notification",
		} {
			if !strings.Contains(row, fragment) {
				t.Fatalf("expected Inbox coverage row to include #1987 evidence %q: %s", fragment, row)
			}
		}
	}
	if strings.Contains(inboxReviewRow, "Inbox review and live Telegram/external-channel launch proof remain partial follow-up evidence") {
		t.Fatalf("Inbox review row still contains stale future-only #1987 wording: %s", inboxReviewRow)
	}
}

func TestAgentSkillCoverageMatrixBindsInboxMissingStaleContextToIssue1987Evidence(t *testing.T) {
	t.Parallel()

	matrixPath := filepath.Join("..", "openspec", "traceability", "agent-skill-coverage.md")
	raw, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("read agent skill coverage matrix: %v", err)
	}
	content := string(raw)

	var inboxReviewRow string
	var inboxSurfaceRow string
	for _, line := range strings.Split(content, "\n") {
		switch {
		case strings.HasPrefix(line, "| Inbox review |"):
			inboxReviewRow = line
		case strings.HasPrefix(line, "| Inbox |"):
			inboxSurfaceRow = line
		}
	}
	if inboxReviewRow == "" {
		t.Fatalf("expected Inbox review channel coverage row")
	}
	if inboxSurfaceRow == "" {
		t.Fatalf("expected Inbox surface coverage row")
	}

	for _, row := range []string{inboxReviewRow, inboxSurfaceRow} {
		for _, fragment := range []string{
			"#1987",
			"TestAgentSkillInboxReviewContextClarifiesMissingOrStaleNotification",
			"agent_context.selected_notification",
			"missing_context",
			"stale_selected_notification",
		} {
			if !strings.Contains(row, fragment) {
				t.Fatalf("expected Inbox coverage row to include #1987 missing/stale evidence %q: %s", fragment, row)
			}
		}
	}

	for _, stale := range []string{
		"missing/stale Inbox notification clarification remain partial follow-up evidence",
		"Missing/stale Inbox notification context clarification",
	} {
		if strings.Contains(content, stale) {
			t.Fatalf("agent skill coverage matrix still contains stale #1987 missing/stale wording %q", stale)
		}
	}
}

func TestAgentSkillCoverageMatrixBindsSkillsPageToMergedIssue1670Evidence(t *testing.T) {
	t.Parallel()

	matrixPath := filepath.Join("..", "openspec", "traceability", "agent-skill-coverage.md")
	raw, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatalf("read agent skill coverage matrix: %v", err)
	}
	content := string(raw)

	var skillsPageRow string
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "| Skills page detail/actions |") {
			skillsPageRow = line
			break
		}
	}
	if skillsPageRow == "" {
		t.Fatalf("expected Skills page detail/actions channel coverage row")
	}

	for _, stale := range []string{
		"| planned |",
		"Planned Cypress",
	} {
		if strings.Contains(skillsPageRow, stale) {
			t.Fatalf("expected Skills page coverage row to avoid stale planned #1670 wording %q: %s", stale, skillsPageRow)
		}
	}

	for _, fragment := range []string{
		"| implemented |",
		"#1670",
		"TestAgentSkillStateAPIEnablesAndDisablesImportedSkill",
		"TestAgentSkillStateAPIBlocksBuiltInAndHighRiskWithoutConfirmation",
		"AGENT-SKILLS-REGISTRY-008",
		"lists skills opens details and imports a local archive disabled by default",
	} {
		if !strings.Contains(skillsPageRow, fragment) {
			t.Fatalf("expected Skills page coverage row to include merged #1670 evidence %q: %s", fragment, skillsPageRow)
		}
	}
}

package profile

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

func TestRepositoryCreateAndList(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	repo := NewRepository(conn)
	if _, err := repo.Create(context.Background(), " "); err == nil {
		t.Fatal("expected error for blank name")
	}

	created, err := repo.Create(context.Background(), "Default")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == "" || created.CreatedAt == "" {
		t.Fatalf("unexpected created profile: %+v", created)
	}

	all, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(all))
	}
	if all[0].Name != "Default" {
		t.Fatalf("expected profile name Default, got %q", all[0].Name)
	}

	if err := repo.SetActiveProfile(context.Background(), created.ID); err != nil {
		t.Fatalf("SetActiveProfile() error = %v", err)
	}
	active, err := repo.GetActiveProfile(context.Background())
	if err != nil {
		t.Fatalf("GetActiveProfile() error = %v", err)
	}
	if active.ID != created.ID {
		t.Fatalf("expected active profile %q, got %q", created.ID, active.ID)
	}

	if err := repo.PutSettings(context.Background(), created.ID, map[string]string{
		"theme":            "dark",
		"scanner_schedule": "0 */6 * * *",
	}); err != nil {
		t.Fatalf("PutSettings() error = %v", err)
	}
	settings, err := repo.GetSettings(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetSettings() error = %v", err)
	}
	if settings["theme"] != "dark" {
		t.Fatalf("unexpected theme setting: %q", settings["theme"])
	}
}

func TestRepositoryPersistsDefaultAgentAuthorityPolicy(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	repo := NewRepository(conn)
	created, err := repo.Create(context.Background(), "Agent Authority")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	policy, err := repo.GetAgentAuthorityPolicy(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetAgentAuthorityPolicy() error = %v", err)
	}
	if policy.ProfileID != created.ID ||
		policy.Mode != AgentAuthorityModeAskBeforeLocalChanges ||
		policy.ExternalWriteApproved {
		t.Fatalf("expected default ask-before-local-changes policy, got %+v", policy)
	}

	var storedMode string
	if err := conn.QueryRowContext(context.Background(), `SELECT value FROM profile_settings WHERE profile_id = ? AND key = ?`, created.ID, AgentAuthorityModeSettingKey).Scan(&storedMode); err != nil {
		t.Fatalf("query stored default agent authority mode: %v", err)
	}
	if storedMode != AgentAuthorityModeAskBeforeLocalChanges {
		t.Fatalf("stored mode = %q, want %q", storedMode, AgentAuthorityModeAskBeforeLocalChanges)
	}

	legacy, err := repo.Create(context.Background(), "Legacy Agent Authority")
	if err != nil {
		t.Fatalf("Create legacy profile error = %v", err)
	}
	if _, err := conn.ExecContext(context.Background(), `DELETE FROM profile_settings WHERE profile_id = ? AND key IN (?, ?)`, legacy.ID, AgentAuthorityModeSettingKey, AgentAuthorityExternalWriteApprovedSettingKey); err != nil {
		t.Fatalf("delete legacy authority settings: %v", err)
	}

	legacyPolicy, err := repo.GetAgentAuthorityPolicy(context.Background(), legacy.ID)
	if err != nil {
		t.Fatalf("GetAgentAuthorityPolicy() legacy error = %v", err)
	}
	if legacyPolicy.Mode != AgentAuthorityModeAskBeforeLocalChanges || legacyPolicy.ExternalWriteApproved {
		t.Fatalf("expected legacy profile to backfill default policy, got %+v", legacyPolicy)
	}
	if err := conn.QueryRowContext(context.Background(), `SELECT value FROM profile_settings WHERE profile_id = ? AND key = ?`, legacy.ID, AgentAuthorityModeSettingKey).Scan(&storedMode); err != nil {
		t.Fatalf("query backfilled legacy default mode: %v", err)
	}
	if storedMode != AgentAuthorityModeAskBeforeLocalChanges {
		t.Fatalf("backfilled mode = %q, want %q", storedMode, AgentAuthorityModeAskBeforeLocalChanges)
	}

	updated, err := repo.PutAgentAuthorityPolicy(context.Background(), created.ID, AgentAuthorityPolicy{
		Mode:                  AgentAuthorityModeReadOnly,
		ExternalWriteApproved: true,
	})
	if err != nil {
		t.Fatalf("PutAgentAuthorityPolicy() error = %v", err)
	}
	if updated.Mode != AgentAuthorityModeReadOnly || !updated.ExternalWriteApproved {
		t.Fatalf("expected updated read-only policy, got %+v", updated)
	}
	if _, err := repo.PutAgentAuthorityPolicy(context.Background(), created.ID, AgentAuthorityPolicy{Mode: "silent_write"}); err == nil {
		t.Fatal("expected invalid Agent authority mode to be rejected")
	}
}

func TestRepositoryAuditsAgentAuthorityPolicyChanges(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	repo := NewRepository(conn)
	created, err := repo.Create(context.Background(), "Audited Agent Authority")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if _, err := repo.PutAgentAuthorityPolicy(context.Background(), created.ID, AgentAuthorityPolicy{
		Mode:                  AgentAuthorityModeApprovedExternalActions,
		ExternalWriteApproved: true,
	}); err != nil {
		t.Fatalf("PutAgentAuthorityPolicy() error = %v", err)
	}

	rows, err := conn.QueryContext(context.Background(), `
		SELECT entity_type, entity_id, action, actor, source, before_json, after_json
		FROM audit_events
		WHERE entity_type = 'profile_agent_authority_policy' AND entity_id = ?
		ORDER BY created_at ASC, id ASC
	`, created.ID)
	if err != nil {
		t.Fatalf("query audit events: %v", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		count++
		var entityType, entityID, action, actor, source, beforeRaw, afterRaw string
		if err := rows.Scan(&entityType, &entityID, &action, &actor, &source, &beforeRaw, &afterRaw); err != nil {
			t.Fatalf("scan audit event: %v", err)
		}
		if entityType != "profile_agent_authority_policy" || entityID != created.ID {
			t.Fatalf("unexpected audited entity: %s/%s", entityType, entityID)
		}
		if action != "agent_authority_policy.update" || actor != "cabinet.agent_authority" || source != "settings.skills" {
			t.Fatalf("unexpected audit metadata: action=%q actor=%q source=%q", action, actor, source)
		}
		var before map[string]any
		var after map[string]any
		if err := json.Unmarshal([]byte(beforeRaw), &before); err != nil {
			t.Fatalf("unmarshal before_json: %v", err)
		}
		if err := json.Unmarshal([]byte(afterRaw), &after); err != nil {
			t.Fatalf("unmarshal after_json: %v", err)
		}
		if before["mode"] != AgentAuthorityModeAskBeforeLocalChanges || before["external_write_approved"] != false {
			t.Fatalf("unexpected before policy: %+v", before)
		}
		if after["mode"] != AgentAuthorityModeApprovedExternalActions || after["external_write_approved"] != true {
			t.Fatalf("unexpected after policy: %+v", after)
		}
		if _, ok := after["secret"]; ok {
			t.Fatalf("audit payload must not include secret fields: %+v", after)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate audit events: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one policy-change audit event, got %d", count)
	}
}

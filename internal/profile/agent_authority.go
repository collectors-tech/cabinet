package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	AgentAuthorityModeSettingKey                  = "agent.authority.mode"
	AgentAuthorityExternalWriteApprovedSettingKey = "agent.authority.external_write_approved"

	AgentAuthorityModeReadOnly                = "read_only"
	AgentAuthorityModeAskBeforeLocalChanges   = "ask_before_local_changes"
	AgentAuthorityModeApprovedExternalActions = "approved_external_actions"
)

type AgentAuthorityPolicy struct {
	ProfileID             string `json:"profile_id"`
	Mode                  string `json:"mode"`
	ExternalWriteApproved bool   `json:"external_write_approved"`
}

func (r *Repository) GetAgentAuthorityPolicy(ctx context.Context, profileID string) (AgentAuthorityPolicy, error) {
	if _, err := r.GetByID(ctx, profileID); err != nil {
		return AgentAuthorityPolicy{}, err
	}
	settings, err := r.GetSettings(ctx, profileID)
	if err != nil {
		return AgentAuthorityPolicy{}, err
	}

	mode := strings.TrimSpace(settings[AgentAuthorityModeSettingKey])
	if mode == "" {
		mode = AgentAuthorityModeAskBeforeLocalChanges
		if err := r.ensureDefaultAgentAuthorityPolicy(ctx, profileID); err != nil {
			return AgentAuthorityPolicy{}, err
		}
	} else if !validAgentAuthorityMode(mode) {
		return AgentAuthorityPolicy{}, fmt.Errorf("invalid agent authority mode")
	}

	return AgentAuthorityPolicy{
		ProfileID:             strings.TrimSpace(profileID),
		Mode:                  mode,
		ExternalWriteApproved: parseAgentAuthorityBool(settings[AgentAuthorityExternalWriteApprovedSettingKey]),
	}, nil
}

func (r *Repository) PutAgentAuthorityPolicy(ctx context.Context, profileID string, policy AgentAuthorityPolicy) (AgentAuthorityPolicy, error) {
	if _, err := r.GetByID(ctx, profileID); err != nil {
		return AgentAuthorityPolicy{}, err
	}
	current, err := r.GetAgentAuthorityPolicy(ctx, profileID)
	if err != nil {
		return AgentAuthorityPolicy{}, err
	}
	mode := strings.TrimSpace(policy.Mode)
	if mode == "" {
		mode = AgentAuthorityModeAskBeforeLocalChanges
	}
	if !validAgentAuthorityMode(mode) {
		return AgentAuthorityPolicy{}, fmt.Errorf("invalid agent authority mode")
	}
	if err := r.PutSettings(ctx, profileID, map[string]string{
		AgentAuthorityModeSettingKey:                  mode,
		AgentAuthorityExternalWriteApprovedSettingKey: formatAgentAuthorityBool(policy.ExternalWriteApproved),
	}); err != nil {
		return AgentAuthorityPolicy{}, err
	}
	updated := AgentAuthorityPolicy{
		ProfileID:             strings.TrimSpace(profileID),
		Mode:                  mode,
		ExternalWriteApproved: policy.ExternalWriteApproved,
	}
	if current.Mode != updated.Mode || current.ExternalWriteApproved != updated.ExternalWriteApproved {
		if err := r.appendAgentAuthorityPolicyAudit(ctx, current, updated); err != nil {
			return AgentAuthorityPolicy{}, err
		}
	}
	return updated, nil
}

func (r *Repository) ensureDefaultAgentAuthorityPolicy(ctx context.Context, profileID string) error {
	settings, err := r.GetSettings(ctx, profileID)
	if err != nil {
		return err
	}
	defaults := map[string]string{}
	if strings.TrimSpace(settings[AgentAuthorityModeSettingKey]) == "" {
		defaults[AgentAuthorityModeSettingKey] = AgentAuthorityModeAskBeforeLocalChanges
	}
	if strings.TrimSpace(settings[AgentAuthorityExternalWriteApprovedSettingKey]) == "" {
		defaults[AgentAuthorityExternalWriteApprovedSettingKey] = "false"
	}
	if len(defaults) == 0 {
		return nil
	}
	return r.PutSettings(ctx, profileID, defaults)
}

func validAgentAuthorityMode(mode string) bool {
	switch strings.TrimSpace(mode) {
	case AgentAuthorityModeReadOnly, AgentAuthorityModeAskBeforeLocalChanges, AgentAuthorityModeApprovedExternalActions:
		return true
	default:
		return false
	}
}

func parseAgentAuthorityBool(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func formatAgentAuthorityBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func (r *Repository) appendAgentAuthorityPolicyAudit(ctx context.Context, before, after AgentAuthorityPolicy) error {
	beforeJSON, err := json.Marshal(agentAuthorityPolicyAuditMap(before))
	if err != nil {
		return fmt.Errorf("marshal agent authority policy audit before: %w", err)
	}
	afterJSON, err := json.Marshal(agentAuthorityPolicyAuditMap(after))
	if err != nil {
		return fmt.Errorf("marshal agent authority policy audit after: %w", err)
	}
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO audit_events(id, entity_type, entity_id, action, actor, source, before_json, after_json, created_at)
		VALUES (?, 'profile_agent_authority_policy', ?, 'agent_authority_policy.update', 'cabinet.agent_authority', 'settings.skills', ?, ?, ?)
	`, uuid.NewString(), strings.TrimSpace(after.ProfileID), string(beforeJSON), string(afterJSON), time.Now().UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("append agent authority policy audit: %w", err)
	}
	return nil
}

func agentAuthorityPolicyAuditMap(policy AgentAuthorityPolicy) map[string]any {
	return map[string]any{
		"profile_id":              strings.TrimSpace(policy.ProfileID),
		"mode":                    strings.TrimSpace(policy.Mode),
		"external_write_approved": policy.ExternalWriteApproved,
	}
}

package profile

import (
	"context"
	"fmt"
	"strings"
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
	return AgentAuthorityPolicy{
		ProfileID:             strings.TrimSpace(profileID),
		Mode:                  mode,
		ExternalWriteApproved: policy.ExternalWriteApproved,
	}, nil
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

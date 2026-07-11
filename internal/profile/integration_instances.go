package profile

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

const integrationInstancesSettingsKey = "integration.instances.v1"

type IntegrationRequiredAction struct {
	Code     string `json:"code,omitempty"`
	Message  string `json:"message,omitempty"`
	Workflow string `json:"workflow,omitempty"`
	Guidance string `json:"guidance,omitempty"`
}

type IntegrationInstance struct {
	ID             string                    `json:"id"`
	ProfileID      string                    `json:"profile_id"`
	ProviderID     string                    `json:"provider_id"`
	DisplayName    string                    `json:"display_name,omitempty"`
	Enabled        bool                      `json:"enabled"`
	Config         map[string]string         `json:"config,omitempty"`
	SecretRefs     map[string]string         `json:"secret_refs,omitempty"`
	AuthState      string                    `json:"auth_state,omitempty"`
	HealthState    string                    `json:"health_state,omitempty"`
	LastCheckedAt  string                    `json:"last_checked_at,omitempty"`
	LastSuccessAt  string                    `json:"last_success_at,omitempty"`
	LastError      string                    `json:"last_error,omitempty"`
	RequiredAction IntegrationRequiredAction `json:"required_action,omitempty"`
	CreatedAt      string                    `json:"created_at"`
	UpdatedAt      string                    `json:"updated_at"`
}

type IntegrationInstancePatch struct {
	ID             string                     `json:"id"`
	ProviderID     string                     `json:"provider_id"`
	DisplayName    *string                    `json:"display_name"`
	Enabled        *bool                      `json:"enabled"`
	Config         map[string]string          `json:"config"`
	Secrets        map[string]string          `json:"secrets"`
	AuthState      *string                    `json:"auth_state"`
	HealthState    *string                    `json:"health_state"`
	LastCheckedAt  *string                    `json:"last_checked_at"`
	LastSuccessAt  *string                    `json:"last_success_at"`
	LastError      *string                    `json:"last_error"`
	RequiredAction *IntegrationRequiredAction `json:"required_action"`
}

func (r *Repository) ListIntegrationInstances(ctx context.Context, profileID string) ([]IntegrationInstance, error) {
	if _, err := r.GetByID(ctx, profileID); err != nil {
		return nil, err
	}
	settings, err := r.GetSettings(ctx, profileID)
	if err != nil {
		return nil, err
	}
	instances, err := decodeIntegrationInstances(settings[integrationInstancesSettingsKey])
	if err != nil {
		return nil, err
	}
	for i := range instances {
		instances[i].ProfileID = profileID
	}
	sort.Slice(instances, func(i, j int) bool {
		if instances[i].ProviderID == instances[j].ProviderID {
			return instances[i].ID < instances[j].ID
		}
		return instances[i].ProviderID < instances[j].ProviderID
	})
	return instances, nil
}

func (r *Repository) UpsertIntegrationInstance(ctx context.Context, profileID string, patch IntegrationInstancePatch) (IntegrationInstance, error) {
	if _, err := r.GetByID(ctx, profileID); err != nil {
		return IntegrationInstance{}, err
	}
	instances, err := r.ListIntegrationInstances(ctx, profileID)
	if err != nil {
		return IntegrationInstance{}, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	id := strings.TrimSpace(patch.ID)
	index := -1
	for i := range instances {
		if instances[i].ID == id && id != "" {
			index = i
			break
		}
	}
	if id == "" {
		id = "inst_" + randomHex(8)
	}

	instance := IntegrationInstance{ID: id, ProfileID: profileID, Enabled: true, CreatedAt: now, UpdatedAt: now}
	if index >= 0 {
		instance = instances[index]
		instance.UpdatedAt = now
	}
	if providerID := strings.TrimSpace(patch.ProviderID); providerID != "" {
		instance.ProviderID = providerID
	}
	if instance.ProviderID == "" {
		return IntegrationInstance{}, fmt.Errorf("provider_id is required")
	}
	if patch.DisplayName != nil {
		instance.DisplayName = strings.TrimSpace(*patch.DisplayName)
	}
	if patch.Enabled != nil {
		instance.Enabled = *patch.Enabled
	}
	if patch.AuthState != nil {
		instance.AuthState = strings.TrimSpace(*patch.AuthState)
	}
	if patch.HealthState != nil {
		instance.HealthState = strings.TrimSpace(*patch.HealthState)
	}
	if patch.LastCheckedAt != nil {
		instance.LastCheckedAt = strings.TrimSpace(*patch.LastCheckedAt)
	}
	if patch.LastSuccessAt != nil {
		instance.LastSuccessAt = strings.TrimSpace(*patch.LastSuccessAt)
	}
	if patch.LastError != nil {
		instance.LastError = strings.TrimSpace(*patch.LastError)
	}
	if patch.RequiredAction != nil {
		instance.RequiredAction = *patch.RequiredAction
	}
	if instance.Config == nil {
		instance.Config = map[string]string{}
	}
	for key, value := range patch.Config {
		key = strings.TrimSpace(key)
		if key == "" || looksSecretConfigKey(key) {
			continue
		}
		instance.Config[key] = value
	}
	if instance.SecretRefs == nil {
		instance.SecretRefs = map[string]string{}
	}
	for key, value := range patch.Secrets {
		key = strings.TrimSpace(key)
		if key == "" || strings.TrimSpace(value) == "" {
			continue
		}
		secretKey := "integration." + instance.ID + "." + key
		if err := r.PutSecret(ctx, profileID, secretKey, value); err != nil {
			return IntegrationInstance{}, err
		}
		instance.SecretRefs[key] = secretKey
	}

	if index >= 0 {
		instances[index] = instance
	} else {
		instances = append(instances, instance)
	}
	if err := r.storeIntegrationInstances(ctx, profileID, instances); err != nil {
		return IntegrationInstance{}, err
	}
	return instance, nil
}

func (r *Repository) DeleteIntegrationInstance(ctx context.Context, profileID, instanceID string) error {
	instanceID = strings.TrimSpace(instanceID)
	if instanceID == "" {
		return fmt.Errorf("instance id is required")
	}
	instances, err := r.ListIntegrationInstances(ctx, profileID)
	if err != nil {
		return err
	}
	next := make([]IntegrationInstance, 0, len(instances))
	found := false
	for _, instance := range instances {
		if instance.ID == instanceID {
			found = true
			for _, secretKey := range instance.SecretRefs {
				_ = r.DeleteSecret(ctx, profileID, secretKey)
			}
			continue
		}
		next = append(next, instance)
	}
	if !found {
		return fmt.Errorf("integration instance not found")
	}
	return r.storeIntegrationInstances(ctx, profileID, next)
}

func (r *Repository) storeIntegrationInstances(ctx context.Context, profileID string, instances []IntegrationInstance) error {
	raw, err := json.Marshal(instances)
	if err != nil {
		return fmt.Errorf("encode integration instances: %w", err)
	}
	return r.PutSettings(ctx, profileID, map[string]string{integrationInstancesSettingsKey: string(raw)})
}

func decodeIntegrationInstances(raw string) ([]IntegrationInstance, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var instances []IntegrationInstance
	if err := json.Unmarshal([]byte(raw), &instances); err != nil {
		return nil, fmt.Errorf("decode integration instances: %w", err)
	}
	return instances, nil
}

func randomHex(bytesLen int) string {
	buf := make([]byte, bytesLen)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}

func looksSecretConfigKey(key string) bool {
	lower := strings.ToLower(strings.TrimSpace(key))
	return strings.Contains(lower, "secret") || strings.Contains(lower, "token") || strings.Contains(lower, "api_key") || strings.Contains(lower, "password")
}

package companion

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/collectors-tech/cabinet/internal/profile"
)

const (
	SyncModePassiveCapture = "passive_capture"
	AuthSchemeBearer       = "Bearer "
)

type Module struct {
	ID                    string            `json:"id"`
	Site                  string            `json:"site"`
	ProviderID            string            `json:"provider_id"`
	IntegrationInstanceID string            `json:"integration_instance_id,omitempty"`
	Actions               []string          `json:"actions"`
	PassiveOnly           bool              `json:"passive_only"`
	SafeConfig            map[string]string `json:"safe_config,omitempty"`
}

type Registry struct {
	ProtocolVersion string   `json:"protocol_version"`
	ProfileID       string   `json:"profile_id"`
	Modules         []Module `json:"modules"`
}

type PayloadSubmission struct {
	ProfileID       string         `json:"profile_id"`
	ModuleID        string         `json:"module_id"`
	URL             string         `json:"url"`
	PayloadType     string         `json:"payload_type"`
	CapturedAt      string         `json:"captured_at,omitempty"`
	Passive         bool           `json:"passive"`
	AttemptedWrite  bool           `json:"attempted_write"`
	ConfidenceScore float64        `json:"confidence_score"`
	Data            map[string]any `json:"data,omitempty"`
}

type AcceptedPayload struct {
	Accepted        bool     `json:"accepted"`
	ProfileID       string   `json:"profile_id"`
	ModuleID        string   `json:"module_id"`
	PayloadType     string   `json:"payload_type"`
	SyncMode        string   `json:"sync_mode"`
	RemoteWrite     bool     `json:"remote_write"`
	AuditTrail      []string `json:"audit_trail"`
	ConfidenceLabel string   `json:"confidence_label"`
}

type Service struct {
	modules         map[string]Module
	db              *sql.DB
	profiles        *profile.Repository
	instanceID      string
	options         Options
	rateMu          sync.Mutex
	rateWindows     map[string]rateWindow
	concurrencyMu   sync.Mutex
	activeBySession map[string]int
}

func NewService(modules []Module) *Service {
	return newConfiguredService(modules, Options{})
}

func newConfiguredService(modules []Module, options Options) *Service {
	configured := map[string]Module{}
	for _, module := range modules {
		module.ID = strings.TrimSpace(module.ID)
		module.Site = strings.TrimSpace(module.Site)
		module.ProviderID = strings.TrimSpace(module.ProviderID)
		if module.ProviderID == "" {
			module.ProviderID = module.Site
		}
		if module.ID == "" || module.Site == "" {
			continue
		}
		module.PassiveOnly = true
		module.Actions = append([]string(nil), module.Actions...)
		sort.Strings(module.Actions)
		module.SafeConfig = safeModuleConfig(module.SafeConfig)
		configured[module.ID] = module
	}
	return &Service{
		modules: configured, options: defaultOptions(options),
		rateWindows: map[string]rateWindow{}, activeBySession: map[string]int{},
	}
}

func DefaultService() *Service {
	return NewService(DefaultModules())
}

func DefaultModules() []Module {
	return []Module{
		{
			ID:          "ebay-purchase-capture",
			Site:        "ebay",
			ProviderID:  "ebay",
			Actions:     []string{"capture_order", "capture_item", "capture_tracking"},
			PassiveOnly: true,
		},
	}
}

func (s *Service) Registry() Registry {
	modules := make([]Module, 0, len(s.modules))
	for _, module := range s.modules {
		modules = append(modules, module)
	}
	sort.Slice(modules, func(i, j int) bool { return modules[i].ID < modules[j].ID })
	return Registry{ProtocolVersion: ProtocolVersionV1, Modules: modules}
}

func (s *Service) RegistryForSession(ctx context.Context, authorization string, metadata RequestMetadata) (Registry, error) {
	session, err := s.Authenticate(ctx, authorization, metadata, CapabilityModulesRead)
	if err != nil {
		return Registry{}, err
	}
	return s.registryForProfile(ctx, session.ProfileID)
}

func (s *Service) registryForProfile(ctx context.Context, profileID string) (Registry, error) {
	instances, err := s.profiles.ListIntegrationInstances(ctx, profileID)
	if err != nil {
		return Registry{}, protocolError("companion_module_discovery_failed")
	}
	byProvider := map[string][]profile.IntegrationInstance{}
	for _, instance := range instances {
		if instance.Enabled {
			byProvider[instance.ProviderID] = append(byProvider[instance.ProviderID], instance)
		}
	}
	modules := []Module{}
	for _, module := range s.modules {
		for _, instance := range byProvider[module.ProviderID] {
			projected := module
			projected.IntegrationInstanceID = instance.ID
			projected.SafeConfig = safeModuleConfig(instance.Config)
			projected.Actions = append([]string(nil), module.Actions...)
			modules = append(modules, projected)
		}
	}
	sort.Slice(modules, func(i, j int) bool {
		if modules[i].ID == modules[j].ID {
			return modules[i].IntegrationInstanceID < modules[j].IntegrationInstanceID
		}
		return modules[i].ID < modules[j].ID
	})
	return Registry{ProtocolVersion: ProtocolVersionV1, ProfileID: profileID, Modules: modules}, nil
}

func (s *Service) AcceptPayload(ctx context.Context, in PayloadSubmission, authorization string, metadata RequestMetadata) (AcceptedPayload, error) {
	profileID := strings.TrimSpace(in.ProfileID)
	moduleID := strings.TrimSpace(in.ModuleID)
	payloadType := strings.TrimSpace(in.PayloadType)
	if profileID == "" {
		return AcceptedPayload{}, protocolError("companion_profile_required")
	}
	session, err := s.Authenticate(ctx, authorization, metadata, CapabilityCapturesSubmit)
	if err != nil {
		return AcceptedPayload{}, err
	}
	if session.ProfileID != profileID {
		return AcceptedPayload{}, protocolError("companion_profile_mismatch")
	}
	release, err := s.acquireSession(session.ID)
	if err != nil {
		return AcceptedPayload{}, err
	}
	defer release()
	registry, err := s.registryForProfile(ctx, session.ProfileID)
	if err != nil {
		return AcceptedPayload{}, err
	}
	moduleRegistered := false
	for _, module := range registry.Modules {
		if module.ID == moduleID {
			moduleRegistered = true
			break
		}
	}
	if !moduleRegistered {
		return AcceptedPayload{}, protocolError("companion_module_not_registered")
	}
	if payloadType == "" {
		return AcceptedPayload{}, protocolError("companion_payload_type_required")
	}
	if !validCaptureURL(in.URL) {
		return AcceptedPayload{}, protocolError("companion_capture_url_required")
	}
	if !in.Passive || in.AttemptedWrite {
		return AcceptedPayload{}, protocolError("companion_payload_must_be_passive")
	}
	if in.ConfidenceScore < 0 || in.ConfidenceScore > 1 {
		return AcceptedPayload{}, protocolError("companion_confidence_score_out_of_range")
	}
	if raw, err := json.Marshal(in.Data); err != nil || len(raw) > 1024*1024 {
		return AcceptedPayload{}, protocolError("companion_payload_too_large")
	}
	s.recordAudit(ctx, session.ProfileID, session.ID, "capture.transport.accepted", "accepted", metadata)
	return AcceptedPayload{
		Accepted:        true,
		ProfileID:       profileID,
		ModuleID:        moduleID,
		PayloadType:     payloadType,
		SyncMode:        SyncModePassiveCapture,
		RemoteWrite:     false,
		ConfidenceLabel: confidenceLabel(in.ConfidenceScore),
		AuditTrail: []string{
			"companion_module=" + moduleID,
			"companion_session=" + session.ID,
			"protocol_version=" + session.ProtocolVersion,
			"sync_mode=" + SyncModePassiveCapture,
			"remote_write=false",
		},
	}, nil
}

func validCaptureURL(rawURL string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Host == "" {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func safeModuleConfig(input map[string]string) map[string]string {
	result := map[string]string{}
	for key, value := range input {
		lower := strings.ToLower(strings.TrimSpace(key))
		if lower == "" || strings.Contains(lower, "secret") || strings.Contains(lower, "token") ||
			strings.Contains(lower, "password") || strings.Contains(lower, "cookie") || strings.Contains(lower, "api_key") {
			continue
		}
		if len(key) <= 128 && len(value) <= 2048 {
			result[key] = value
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func confidenceLabel(score float64) string {
	switch {
	case score >= 0.8:
		return "high"
	case score >= 0.5:
		return "medium"
	default:
		return "low"
	}
}

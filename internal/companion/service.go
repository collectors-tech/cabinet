package companion

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

const (
	SyncModePassiveCapture = "passive_capture"
	AuthSchemeBearer       = "Bearer "
)

type Module struct {
	ID          string   `json:"id"`
	Site        string   `json:"site"`
	Actions     []string `json:"actions"`
	PassiveOnly bool     `json:"passive_only"`
}

type Registry struct {
	Modules []Module `json:"modules"`
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
	modules map[string]Module
}

func NewService(modules []Module) *Service {
	configured := map[string]Module{}
	for _, module := range modules {
		module.ID = strings.TrimSpace(module.ID)
		module.Site = strings.TrimSpace(module.Site)
		if module.ID == "" || module.Site == "" {
			continue
		}
		module.PassiveOnly = true
		module.Actions = append([]string(nil), module.Actions...)
		sort.Strings(module.Actions)
		configured[module.ID] = module
	}
	return &Service{modules: configured}
}

func DefaultService() *Service {
	return NewService([]Module{
		{
			ID:          "ebay-purchase-capture",
			Site:        "ebay",
			Actions:     []string{"capture_order", "capture_item", "capture_tracking"},
			PassiveOnly: true,
		},
	})
}

func (s *Service) Registry() Registry {
	modules := make([]Module, 0, len(s.modules))
	for _, module := range s.modules {
		modules = append(modules, module)
	}
	sort.Slice(modules, func(i, j int) bool { return modules[i].ID < modules[j].ID })
	return Registry{Modules: modules}
}

func (s *Service) AcceptPayload(in PayloadSubmission, authorization string) (AcceptedPayload, error) {
	profileID := strings.TrimSpace(in.ProfileID)
	moduleID := strings.TrimSpace(in.ModuleID)
	payloadType := strings.TrimSpace(in.PayloadType)
	if profileID == "" {
		return AcceptedPayload{}, fmt.Errorf("profile_id_required")
	}
	if !validCompanionAuthorization(profileID, authorization) {
		return AcceptedPayload{}, fmt.Errorf("companion_auth_required")
	}
	if _, ok := s.modules[moduleID]; !ok {
		return AcceptedPayload{}, fmt.Errorf("companion_module_not_registered")
	}
	if payloadType == "" {
		return AcceptedPayload{}, fmt.Errorf("payload_type_required")
	}
	if !validCaptureURL(in.URL) {
		return AcceptedPayload{}, fmt.Errorf("capture_url_required")
	}
	if !in.Passive || in.AttemptedWrite {
		return AcceptedPayload{}, fmt.Errorf("companion_payload_must_be_passive")
	}
	if in.ConfidenceScore < 0 || in.ConfidenceScore > 1 {
		return AcceptedPayload{}, fmt.Errorf("confidence_score_out_of_range")
	}
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
			"sync_mode=" + SyncModePassiveCapture,
			"remote_write=false",
		},
	}, nil
}

func validCompanionAuthorization(profileID, authorization string) bool {
	if !strings.HasPrefix(authorization, AuthSchemeBearer) {
		return false
	}
	token := strings.TrimSpace(strings.TrimPrefix(authorization, AuthSchemeBearer))
	return token == "companion:"+profileID
}

func validCaptureURL(rawURL string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Host == "" {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
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

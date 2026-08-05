package companion

import (
	"context"
	"database/sql"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/collectors-tech/cabinet/internal/profile"
)

const (
	SyncModePassiveCapture             = "passive_capture"
	AuthSchemeBearer                   = "Bearer "
	DurableCapturePersistenceAvailable = true
)

type Module struct {
	ID                    string              `json:"id"`
	ModuleVersion         string              `json:"module_version"`
	Site                  string              `json:"site"`
	ProviderID            string              `json:"provider_id"`
	IntegrationInstanceID string              `json:"integration_instance_id,omitempty"`
	Actions               []string            `json:"actions"`
	PassiveOnly           bool                `json:"passive_only"`
	CaptureSchemas        []CaptureSchema     `json:"capture_schemas"`
	Workflows             []string            `json:"workflows"`
	RedactionRules        []string            `json:"redaction_rules"`
	FixtureVersion        string              `json:"fixture_version"`
	Display               ModuleDisplay       `json:"display"`
	Browser               BrowserContract     `json:"browser"`
	Configuration         ModuleConfiguration `json:"configuration"`
	SafeConfig            map[string]string   `json:"safe_config,omitempty"`
}

type CaptureSchema struct {
	PayloadType string   `json:"payload_type"`
	Fields      []string `json:"fields"`
	MediaFields []string `json:"media_fields"`
}

type ModuleConfiguration struct {
	CaptureMode        string   `json:"capture_mode"`
	ItemFields         []string `json:"item_fields"`
	MediaPolicy        string   `json:"media_policy"`
	ReviewDestination  string   `json:"review_destination"`
	RateLimitPerMinute int      `json:"rate_limit_per_minute"`
	HelpURL            string   `json:"help_url"`
	SetupRequired      bool     `json:"setup_required"`
	SyncAvailable      bool     `json:"sync_available"`
}

type ModuleDisplay struct {
	Name string `json:"name"`
}

type BrowserReadiness struct {
	Ready     []string `json:"ready"`
	LoggedOut []string `json:"logged_out"`
	Challenge []string `json:"challenge"`
}

type BrowserContract struct {
	StartURL      string           `json:"start_url"`
	Origins       []string         `json:"origins"`
	URLPatterns   []string         `json:"url_patterns"`
	CaptureScript string           `json:"capture_script,omitempty"`
	Readiness     BrowserReadiness `json:"readiness"`
}

type Registry struct {
	ProtocolVersion string   `json:"protocol_version"`
	ProfileID       string   `json:"profile_id"`
	Modules         []Module `json:"modules"`
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
		module.ModuleVersion = strings.TrimSpace(module.ModuleVersion)
		module.Site = strings.TrimSpace(module.Site)
		module.ProviderID = strings.TrimSpace(module.ProviderID)
		module.FixtureVersion = strings.TrimSpace(module.FixtureVersion)
		module.Display.Name = strings.TrimSpace(module.Display.Name)
		if module.ProviderID == "" {
			module.ProviderID = module.Site
		}
		if module.ID == "" || module.Site == "" || module.ModuleVersion == "" || module.FixtureVersion == "" || module.Display.Name == "" ||
			!module.PassiveOnly || !validBrowserContract(module.Browser) || !validCaptureContract(module) {
			continue
		}
		module.Actions = append([]string(nil), module.Actions...)
		sort.Strings(module.Actions)
		module.Workflows = append([]string(nil), module.Workflows...)
		sort.Strings(module.Workflows)
		module.RedactionRules = append([]string(nil), module.RedactionRules...)
		sort.Strings(module.RedactionRules)
		module.CaptureSchemas = copyCaptureSchemas(module.CaptureSchemas)
		module.Configuration.ItemFields = append([]string(nil), module.Configuration.ItemFields...)
		module.Configuration.SyncAvailable = module.Configuration.SyncAvailable && module.Browser.CaptureScript != "" && DurableCapturePersistenceAvailable
		module.Browser = copyBrowserContract(module.Browser)
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
			ID:            "ebay-purchase-capture",
			ModuleVersion: "1.0.0",
			Site:          "ebay",
			ProviderID:    "ebay",
			Actions:       []string{"capture_order", "capture_item", "capture_tracking"},
			PassiveOnly:   true,
			CaptureSchemas: []CaptureSchema{{
				PayloadType: "purchase_order",
				Fields:      []string{"cards"},
				MediaFields: []string{"cards.image_url"},
			}},
			Workflows:      []string{"manual_purchase_capture"},
			RedactionRules: []string{"no_cookies", "no_raw_page", "no_tokens"},
			FixtureVersion: "1",
			Display:        ModuleDisplay{Name: "eBay purchases"},
			Browser: BrowserContract{
				StartURL:    "https://www.ebay.com/mye/myebay/purchase",
				Origins:     []string{"https://www.ebay.com/*"},
				URLPatterns: []string{"https://www.ebay.com/mye/myebay/purchase*"},
				Readiness: BrowserReadiness{
					Ready:     []string{"#purchase-history", "[data-testid=\"purchase-history\"]"},
					LoggedOut: []string{"a[href*=\"signin\"]", "form[action*=\"signin\"]"},
					Challenge: []string{"#captcha-box", "iframe[src*=\"challenge\"]"},
				},
			},
			Configuration: ModuleConfiguration{
				CaptureMode: "manual_user_present", ItemFields: []string{"order_id", "item_id", "title", "seller", "price", "currency", "tracking_number"},
				MediaPolicy: "review_before_canonical_persistence", ReviewDestination: "purchases", RateLimitPerMinute: 6,
				HelpURL: "/help-center/integrations", SetupRequired: false, SyncAvailable: false,
			},
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
			projected.Workflows = append([]string(nil), module.Workflows...)
			projected.RedactionRules = append([]string(nil), module.RedactionRules...)
			projected.CaptureSchemas = copyCaptureSchemas(module.CaptureSchemas)
			projected.Configuration.ItemFields = append([]string(nil), module.Configuration.ItemFields...)
			projected.Browser = copyBrowserContract(module.Browser)
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

func validBrowserContract(contract BrowserContract) bool {
	start, err := url.Parse(strings.TrimSpace(contract.StartURL))
	if err != nil || start.Scheme != "https" || start.Hostname() == "" || start.User != nil || len(contract.Origins) == 0 || len(contract.Origins) > 10 ||
		len(contract.URLPatterns) == 0 || len(contract.URLPatterns) > 20 {
		return false
	}
	startAllowed := false
	for _, rawOrigin := range contract.Origins {
		origin := strings.TrimSpace(rawOrigin)
		base := strings.TrimSuffix(origin, "/*")
		if !strings.HasPrefix(origin, "https://") || !strings.HasSuffix(origin, "/*") || strings.Contains(base, "*") {
			return false
		}
		parsed, parseErr := url.Parse(base)
		if parseErr != nil || parsed.Hostname() == "" || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			return false
		}
		if start.Scheme == parsed.Scheme && start.Host == parsed.Host {
			startAllowed = true
		}
	}
	if !startAllowed {
		return false
	}
	for _, pattern := range contract.URLPatterns {
		if strings.Count(pattern, "*") != 1 || !strings.HasSuffix(pattern, "*") {
			return false
		}
		prefix := strings.TrimSuffix(pattern, "*")
		parsed, parseErr := url.Parse(prefix)
		if parseErr != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil {
			return false
		}
		allowed := false
		for _, origin := range contract.Origins {
			originURL, _ := url.Parse(strings.TrimSuffix(origin, "/*"))
			if parsed.Scheme == originURL.Scheme && parsed.Host == originURL.Host {
				allowed = true
			}
		}
		if !allowed {
			return false
		}
	}
	if contract.CaptureScript != "" && (!strings.HasPrefix(contract.CaptureScript, "modules/") || !strings.HasSuffix(contract.CaptureScript, ".js") || strings.Contains(contract.CaptureScript, "..")) {
		return false
	}
	return validSelectors(contract.Readiness.Ready) && validSelectors(contract.Readiness.LoggedOut) && validSelectors(contract.Readiness.Challenge)
}

func validCaptureContract(module Module) bool {
	if len(module.CaptureSchemas) == 0 || len(module.CaptureSchemas) > 20 || !validNames(module.Workflows, 20) || !validNames(module.RedactionRules, 20) {
		return false
	}
	for _, schema := range module.CaptureSchemas {
		if strings.TrimSpace(schema.PayloadType) == "" || !validNames(schema.Fields, 64) || len(schema.MediaFields) > 32 || (len(schema.MediaFields) > 0 && !validNames(schema.MediaFields, 32)) {
			return false
		}
	}
	config := module.Configuration
	if config.CaptureMode != "manual_user_present" && config.CaptureMode != "browser_open_scheduled" {
		return false
	}
	if !validNames(config.ItemFields, 64) || strings.TrimSpace(config.MediaPolicy) == "" || !validDestination(config.ReviewDestination) ||
		config.RateLimitPerMinute < 1 || config.RateLimitPerMinute > 60 || len(config.HelpURL) < 2 || !strings.HasPrefix(config.HelpURL, "/") || strings.HasPrefix(config.HelpURL, "//") {
		return false
	}
	return !config.SyncAvailable || module.Browser.CaptureScript != ""
}

func validDestination(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if (character < 'a' || character > 'z') && (character < '0' || character > '9') && character != '-' {
			return false
		}
	}
	return true
}

func validNames(values []string, maximum int) bool {
	if len(values) == 0 || len(values) > maximum {
		return false
	}
	for _, value := range values {
		if strings.TrimSpace(value) == "" || len(value) > 128 {
			return false
		}
	}
	return true
}

func validSelectors(selectors []string) bool {
	if len(selectors) == 0 || len(selectors) > 20 {
		return false
	}
	for _, selector := range selectors {
		if strings.TrimSpace(selector) == "" || len(selector) > 256 {
			return false
		}
	}
	return true
}

func copyBrowserContract(contract BrowserContract) BrowserContract {
	contract.Origins = append([]string(nil), contract.Origins...)
	contract.URLPatterns = append([]string(nil), contract.URLPatterns...)
	contract.Readiness.Ready = append([]string(nil), contract.Readiness.Ready...)
	contract.Readiness.LoggedOut = append([]string(nil), contract.Readiness.LoggedOut...)
	contract.Readiness.Challenge = append([]string(nil), contract.Readiness.Challenge...)
	return contract
}

func copyCaptureSchemas(schemas []CaptureSchema) []CaptureSchema {
	copied := make([]CaptureSchema, len(schemas))
	for index, schema := range schemas {
		copied[index] = schema
		copied[index].Fields = append([]string(nil), schema.Fields...)
		copied[index].MediaFields = append([]string(nil), schema.MediaFields...)
	}
	return copied
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

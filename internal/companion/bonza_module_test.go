package companion

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/profile"
)

const (
	bonzaModuleID   = "bonzaslotcars-search-capture"
	bonzaProviderID = "au-webshop-bonzaslotcars-com-au"
	bonzaScope      = "bonzaslotcars"
)

func TestDefaultBonzaModulePublishesExactFailClosedBrowserContract(t *testing.T) {
	t.Parallel()

	registry := NewService(DefaultModules()).Registry()
	var module Module
	for _, candidate := range registry.Modules {
		if candidate.ID == bonzaModuleID {
			module = candidate
			break
		}
	}
	if module.ID == "" {
		t.Fatalf("default companion registry is missing %s: %+v", bonzaModuleID, registry.Modules)
	}
	if module.ProviderID != bonzaProviderID || module.Site != bonzaScope || module.Display.Name != "Bonza Slot Cars" {
		t.Fatalf("Bonza identities are not canonical: %+v", module)
	}
	if module.Browser.StartURL != "https://www.bonzaslotcars.com.au/" || module.Browser.CaptureScript != "modules/bonzaslotcars.js" {
		t.Fatalf("Bonza browser entry point is not packaged: %+v", module.Browser)
	}
	wantOrigins := map[string]bool{
		"https://www.bonzaslotcars.com.au/*": false,
		"https://bonzaslotcars.com.au/*":     false,
	}
	for _, origin := range module.Browser.Origins {
		if _, ok := wantOrigins[origin]; !ok {
			t.Fatalf("unexpected Bonza optional origin %q", origin)
		}
		wantOrigins[origin] = true
	}
	for origin, found := range wantOrigins {
		if !found {
			t.Fatalf("missing exact Bonza origin %q in %+v", origin, module.Browser.Origins)
		}
	}
	if len(module.Browser.Readiness.Ready) == 0 || len(module.Browser.Readiness.LoggedOut) == 0 || len(module.Browser.Readiness.Challenge) == 0 {
		t.Fatalf("Bonza readiness must fail closed across ready/login/challenge states: %+v", module.Browser.Readiness)
	}
	if !containsString(module.Browser.Readiness.Challenge, "script-marker:sucuri_cloudproxy_js") {
		t.Fatalf("Bonza readiness must recognise the Sucuri script marker without exporting its contents: %+v", module.Browser.Readiness)
	}
	if !module.Configuration.SyncAvailable || module.Configuration.CaptureMode != "manual_user_present" || module.Configuration.ReviewDestination != "discoveries" {
		t.Fatalf("Bonza sync contract is not truthful: %+v", module.Configuration)
	}
	if len(module.CaptureSchemas) != 2 || module.FixtureVersion != "1" || !module.PassiveOnly {
		t.Fatalf("Bonza capture contract is incomplete: %+v", module)
	}
}

func TestBonzaCompanionSearchPersistsCanonicalMarketWatchProvenance(t *testing.T) {
	t.Parallel()

	svc, profiles, profileID, metadata := newPersistentCompanionTestService(t, Options{})
	instance, err := profiles.UpsertIntegrationInstance(context.Background(), profileID, profile.IntegrationInstancePatch{
		ProviderID: bonzaProviderID,
		Enabled:    boolPointer(true),
	})
	if err != nil {
		t.Fatalf("create Bonza integration instance: %v", err)
	}
	svc, err = NewPersistentService(context.Background(), svc.db, profiles, DefaultModules(), Options{})
	if err != nil {
		t.Fatalf("restart companion with Bonza module: %v", err)
	}
	authorization := pairCompanionTestSession(t, svc, profileID, metadata, []string{CapabilityCapturesSubmit})

	raw, err := os.ReadFile(filepath.Join("testdata", "bonza-search-results-v1.json"))
	if err != nil {
		t.Fatalf("read Bonza fixture: %v", err)
	}
	data := map[string]any{}
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode Bonza fixture: %v", err)
	}
	data["complete"] = true
	data["total_pages"] = float64(1)
	payload := PayloadSubmission{
		ProtocolVersion: ProtocolVersionV1, ProfileID: profileID, ModuleID: bonzaModuleID,
		ModuleVersion: "1.0.0", SchemaVersion: "1", IntegrationInstanceID: instance.ID, ProviderID: bonzaProviderID,
		URL: "https://www.bonzaslotcars.com.au/?s=Scalextric&utm_source=private", PayloadType: "search_results",
		CapturedAt: svc.options.Now().UTC().Format("2006-01-02T15:04:05Z07:00"), PageComplete: true, Passive: true,
		ConfidenceScore: 0.95, RedactionSummary: []string{"no_cookies", "no_raw_page", "no_tokens"},
		PayloadHash: PayloadDigest(data), IdempotencyKey: "bonza-module-fixture-v1", Data: data,
	}
	accepted, err := svc.AcceptPayload(context.Background(), payload, authorization, metadata)
	if err != nil || !accepted.Committed || accepted.State != captureStateCompleted {
		t.Fatalf("Bonza fixture acceptance = %+v, %v", accepted, err)
	}

	var source, providerID, moduleVersion, schemaVersion, sourceResultURL, currency, stock string
	var candidates int
	if err := svc.db.QueryRow(`SELECT COUNT(*), MIN(source), MIN(source_result_url), MIN(observed_currency), MIN(stock_state)
		FROM scanner_candidates WHERE profile_id = ?`, profileID).Scan(&candidates, &source, &sourceResultURL, &currency, &stock); err != nil {
		t.Fatalf("load Bonza Market Watch candidates: %v", err)
	}
	if candidates != 2 || source != bonzaScope || currency != "AUD" || stock == "" || sourceResultURL != "https://www.bonzaslotcars.com.au/" {
		t.Fatalf("Bonza candidate provenance is not canonical: count=%d source=%q result=%q currency=%q stock=%q", candidates, source, sourceResultURL, currency, stock)
	}
	if err := svc.db.QueryRow(`SELECT provider_id, module_version, schema_version FROM companion_captures WHERE id = ?`, accepted.CaptureID).
		Scan(&providerID, &moduleVersion, &schemaVersion); err != nil {
		t.Fatalf("load Bonza transport provenance: %v", err)
	}
	if providerID != bonzaProviderID || moduleVersion != "1.0.0" || schemaVersion != "1" {
		t.Fatalf("Bonza transport provenance = provider %q module %q schema %q", providerID, moduleVersion, schemaVersion)
	}
	var healthStatus, healthMessage string
	if err := svc.db.QueryRow(`SELECT status, message FROM provider_health WHERE provider = ?`, bonzaScope).
		Scan(&healthStatus, &healthMessage); err != nil {
		t.Fatalf("load Bonza companion health evidence: %v", err)
	}
	if healthStatus != "ok" || !strings.Contains(healthMessage, "browser_companion") || !strings.Contains(healthMessage, bonzaModuleID) {
		t.Fatalf("Bonza companion health evidence status=%q message=%q", healthStatus, healthMessage)
	}
}

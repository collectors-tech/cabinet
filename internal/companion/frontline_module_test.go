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
	frontlineModuleID   = "frontlinehobbies-search-capture"
	frontlineProviderID = "au-webshop-frontlinehobbies-com-au"
	frontlineScope      = "frontlinehobbies"
)

func TestDefaultFrontlineModulePublishesExactFailClosedBrowserContract(t *testing.T) {
	t.Parallel()

	registry := NewService(DefaultModules()).Registry()
	var module Module
	for _, candidate := range registry.Modules {
		if candidate.ID == frontlineModuleID {
			module = candidate
			break
		}
	}
	if module.ID == "" {
		t.Fatalf("default companion registry is missing %s: %+v", frontlineModuleID, registry.Modules)
	}
	if module.ProviderID != frontlineProviderID || module.Site != frontlineScope || module.Display.Name != "Frontline Hobbies" {
		t.Fatalf("Frontline identities are not canonical: %+v", module)
	}
	if module.Browser.StartURL != "https://www.frontlinehobbies.com.au/" || module.Browser.CaptureScript != "modules/frontlinehobbies.js" {
		t.Fatalf("Frontline browser entry point is not packaged: %+v", module.Browser)
	}
	wantOrigins := map[string]bool{
		"https://www.frontlinehobbies.com.au/*": false,
		"https://frontlinehobbies.com.au/*":     false,
		"https://cdn.frontlinehobbies.com.au/*": false,
	}
	for _, origin := range module.Browser.Origins {
		if _, ok := wantOrigins[origin]; !ok {
			t.Fatalf("unexpected Frontline optional origin %q", origin)
		}
		wantOrigins[origin] = true
	}
	for origin, found := range wantOrigins {
		if !found {
			t.Fatalf("missing exact Frontline origin %q in %+v", origin, module.Browser.Origins)
		}
	}
	if len(module.Browser.Readiness.Ready) == 0 || len(module.Browser.Readiness.LoggedOut) == 0 || len(module.Browser.Readiness.Challenge) == 0 {
		t.Fatalf("Frontline readiness must fail closed across ready/login/challenge states: %+v", module.Browser.Readiness)
	}
	if !module.Configuration.SyncAvailable || module.Configuration.CaptureMode != "manual_user_present" || module.Configuration.ReviewDestination != "discoveries" {
		t.Fatalf("Frontline sync contract is not truthful: %+v", module.Configuration)
	}
	if len(module.CaptureSchemas) != 2 || module.FixtureVersion != "1" || !module.PassiveOnly {
		t.Fatalf("Frontline capture contract is incomplete: %+v", module)
	}
}

func TestFrontlineCompanionSearchPersistsCanonicalMarketWatchProvenance(t *testing.T) {
	t.Parallel()

	svc, profiles, profileID, metadata := newPersistentCompanionTestService(t, Options{})
	instance, err := profiles.UpsertIntegrationInstance(context.Background(), profileID, profile.IntegrationInstancePatch{
		ProviderID: frontlineProviderID,
		Enabled:    boolPointer(true),
	})
	if err != nil {
		t.Fatalf("create Frontline integration instance: %v", err)
	}
	svc, err = NewPersistentService(context.Background(), svc.db, profiles, DefaultModules(), Options{})
	if err != nil {
		t.Fatalf("restart companion with Frontline module: %v", err)
	}
	authorization := pairCompanionTestSession(t, svc, profileID, metadata, []string{CapabilityCapturesSubmit})

	raw, err := os.ReadFile(filepath.Join("testdata", "frontline-search-results-v1.json"))
	if err != nil {
		t.Fatalf("read Frontline fixture: %v", err)
	}
	data := map[string]any{}
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("decode Frontline fixture: %v", err)
	}
	payload := PayloadSubmission{
		ProtocolVersion: ProtocolVersionV1, ProfileID: profileID, ModuleID: frontlineModuleID,
		ModuleVersion: "1.0.0", SchemaVersion: "1", IntegrationInstanceID: instance.ID, ProviderID: frontlineProviderID,
		URL: "https://www.frontlinehobbies.com.au/?s=AFX+slot+car&utm_source=private", PayloadType: "search_results",
		CapturedAt: svc.options.Now().UTC().Format("2006-01-02T15:04:05Z07:00"), PageComplete: true, Passive: true,
		ConfidenceScore: 0.95, RedactionSummary: []string{"no_cookies", "no_raw_page", "no_tokens"},
		PayloadHash: PayloadDigest(data), IdempotencyKey: "frontline-module-fixture-v1", Data: data,
	}
	accepted, err := svc.AcceptPayload(context.Background(), payload, authorization, metadata)
	if err != nil || !accepted.Committed || accepted.State != captureStateCompleted {
		t.Fatalf("Frontline fixture acceptance = %+v, %v", accepted, err)
	}

	var source, providerID, moduleVersion, schemaVersion, sourceResultURL, currency, stock string
	var candidates int
	if err := svc.db.QueryRow(`SELECT COUNT(*), MIN(source), MIN(source_result_url), MIN(observed_currency), MIN(stock_state)
		FROM scanner_candidates WHERE profile_id = ?`, profileID).Scan(&candidates, &source, &sourceResultURL, &currency, &stock); err != nil {
		t.Fatalf("load Frontline Market Watch candidates: %v", err)
	}
	if candidates != 2 || source != frontlineScope || currency != "AUD" || stock == "" || sourceResultURL != "https://www.frontlinehobbies.com.au/" {
		t.Fatalf("Frontline candidate provenance is not canonical: count=%d source=%q result=%q currency=%q stock=%q", candidates, source, sourceResultURL, currency, stock)
	}
	if err := svc.db.QueryRow(`SELECT provider_id, module_version, schema_version FROM companion_captures WHERE id = ?`, accepted.CaptureID).
		Scan(&providerID, &moduleVersion, &schemaVersion); err != nil {
		t.Fatalf("load Frontline transport provenance: %v", err)
	}
	if providerID != frontlineProviderID || moduleVersion != "1.0.0" || schemaVersion != "1" {
		t.Fatalf("Frontline transport provenance = provider %q module %q schema %q", providerID, moduleVersion, schemaVersion)
	}
	var healthStatus, healthMessage string
	if err := svc.db.QueryRow(`SELECT status, message FROM provider_health WHERE provider = ?`, frontlineScope).
		Scan(&healthStatus, &healthMessage); err != nil {
		t.Fatalf("load Frontline companion health evidence: %v", err)
	}
	if healthStatus != "ok" || !strings.Contains(healthMessage, "browser_companion") || !strings.Contains(healthMessage, frontlineModuleID) {
		t.Fatalf("Frontline companion health evidence status=%q message=%q", healthStatus, healthMessage)
	}
}

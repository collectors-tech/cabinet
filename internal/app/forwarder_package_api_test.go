package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestForwarderPackageAPIUpsertsAndListsPackages(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	body := "{\"profile_id\":\"profile-api\",\"provider\":\"Stackry\",\"source\":\"manual\",\"external_package_id\":\"PKG-API-1\",\"status\":\"received\",\"weight_grams\":250,\"raw_payload\":{\"note\":\"manual entry\"}}"
	created := doRequest(t, a, http.MethodPost, "/api/forwarding/packages", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if created.Code != http.StatusOK {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body.String())
	}
	updated := doRequest(t, a, http.MethodPost, "/api/forwarding/packages", strings.NewReader("{\"profile_id\":\"profile-api\",\"provider\":\"stackry\",\"source\":\"manual\",\"external_package_id\":\"PKG-API-1\",\"status\":\"ready_to_ship\",\"weight_grams\":300}"), map[string]string{"Content-Type": "application/json"})
	if updated.Code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", updated.Code, updated.Body.String())
	}
	listed := doRequest(t, a, http.MethodGet, "/api/forwarding/packages?profile_id=profile-api&status=ready_to_ship", nil, nil)
	if listed.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listed.Code, listed.Body.String())
	}
	var payload struct {
		Packages []map[string]any
		Summary  map[string]int
	}
	if err := json.NewDecoder(listed.Body).Decode(&payload); err != nil {
		t.Fatalf("decode list payload: %v", err)
	}
	if payload.Summary["count"] != 1 || len(payload.Packages) != 1 {
		t.Fatalf("expected one package summary, got %+v", payload)
	}
	pkg := payload.Packages[0]
	if pkg["provider"] != "stackry" || pkg["source"] != "manual" || pkg["external_package_id"] != "PKG-API-1" || pkg["status"] != "ready_to_ship" || int(pkg["weight_grams"].(float64)) != 300 {
		t.Fatalf("unexpected listed package: %+v", pkg)
	}
}

func TestForwarderPackageAPIRejectsInvalidImports(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodPost, "/api/forwarding/packages", strings.NewReader("{\"profile_id\":\"profile-api\",\"provider\":\"stackry\",\"source\":\"manual\",\"status\":\"received\"}"), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid package import to fail, status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestForwarderPackageLinkAPIReconcilesPackageToPurchaseArrival(t *testing.T) {
	t.Parallel()

	a, profileID := newCommerceProfileApp(t)
	if _, err := a.db.Exec(`INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title) VALUES ('fwd-item-1', ?, 'AFX', 'Slot', 'FWD-1', 'Forwarder Reconcile Item')`, profileID); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	createLifecycle := doRequest(t, a, http.MethodPost, "/api/commerce/lifecycle", strings.NewReader(`{"item_id":"fwd-item-1","state":"purchase","source":"ebay","external_ref":"order-link-1","quantity":1,"amount":42,"currency":"aud","notes":"awaiting forwarder package"}`), map[string]string{"Content-Type": "application/json"})
	if createLifecycle.Code != http.StatusCreated {
		t.Fatalf("create lifecycle status=%d body=%s", createLifecycle.Code, createLifecycle.Body.String())
	}
	var lifecyclePayload struct {
		Entry struct {
			ID string `json:"id"`
		} `json:"entry"`
		ExpectedArrival struct {
			ID string `json:"id"`
		} `json:"expected_arrival"`
	}
	if err := json.NewDecoder(createLifecycle.Body).Decode(&lifecyclePayload); err != nil {
		t.Fatalf("decode lifecycle payload: %v", err)
	}
	createPackage := doRequest(t, a, http.MethodPost, "/api/forwarding/packages", strings.NewReader(`{"profile_id":"`+profileID+`","provider":"stackry","source":"manual","external_package_id":"PKG-LINK-API","status":"received"}`), map[string]string{"Content-Type": "application/json"})
	if createPackage.Code != http.StatusOK {
		t.Fatalf("create package status=%d body=%s", createPackage.Code, createPackage.Body.String())
	}
	var packagePayload struct {
		Package struct {
			ID string `json:"id"`
		} `json:"package"`
	}
	if err := json.NewDecoder(createPackage.Body).Decode(&packagePayload); err != nil {
		t.Fatalf("decode package payload: %v", err)
	}

	linkBody := `{"package_id":"` + packagePayload.Package.ID + `","item_id":"fwd-item-1","lifecycle_entry_id":"` + lifecyclePayload.Entry.ID + `","expected_arrival_id":"` + lifecyclePayload.ExpectedArrival.ID + `","source":"manual-review","notes":"tracking number matched"}`
	linkResp := doRequest(t, a, http.MethodPost, "/api/forwarding/package-links", strings.NewReader(linkBody), map[string]string{"Content-Type": "application/json"})
	if linkResp.Code != http.StatusOK {
		t.Fatalf("link package status=%d body=%s", linkResp.Code, linkResp.Body.String())
	}
	var linkPayload struct {
		Mode string `json:"mode"`
		Link struct {
			PackageID         string `json:"package_id"`
			ItemID            string `json:"item_id"`
			ExpectedArrivalID string `json:"expected_arrival_id"`
			Source            string `json:"source"`
		} `json:"link"`
	}
	if err := json.NewDecoder(linkResp.Body).Decode(&linkPayload); err != nil {
		t.Fatalf("decode link payload: %v", err)
	}
	if linkPayload.Mode != "forwarder_package_reconciliation_link" || linkPayload.Link.PackageID != packagePayload.Package.ID || linkPayload.Link.ExpectedArrivalID != lifecyclePayload.ExpectedArrival.ID || linkPayload.Link.Source != "manual-review" {
		t.Fatalf("unexpected link payload %+v", linkPayload)
	}
	listLinks := doRequest(t, a, http.MethodGet, "/api/forwarding/package-links?package_id="+packagePayload.Package.ID, nil, nil)
	if listLinks.Code != http.StatusOK {
		t.Fatalf("list links status=%d body=%s", listLinks.Code, listLinks.Body.String())
	}
	var listPayload struct {
		Links   []map[string]any `json:"links"`
		Summary map[string]int   `json:"summary"`
	}
	if err := json.NewDecoder(listLinks.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode link list: %v", err)
	}
	if listPayload.Summary["count"] != 1 || len(listPayload.Links) != 1 || listPayload.Links[0]["item_id"] != "fwd-item-1" {
		t.Fatalf("expected one persisted reconciliation link, got %+v", listPayload)
	}
}

func TestForwarderPackageLinkAPIOverridesAndUnlinksWithEvents(t *testing.T) {
	t.Parallel()

	a, profileID := newCommerceProfileApp(t)
	for _, itemID := range []string{"fwd-override-item-1", "fwd-override-item-2"} {
		if _, err := a.db.Exec(`INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title) VALUES (?, ?, 'AFX', 'Slot', ?, 'Forwarder Override Item')`, itemID, profileID, itemID); err != nil {
			t.Fatalf("seed item %s: %v", itemID, err)
		}
	}
	createLifecycleOne := doRequest(t, a, http.MethodPost, "/api/commerce/lifecycle", strings.NewReader(`{"item_id":"fwd-override-item-1","state":"purchase","source":"ebay","external_ref":"order-override-1","quantity":1,"amount":42,"currency":"aud"}`), map[string]string{"Content-Type": "application/json"})
	if createLifecycleOne.Code != http.StatusCreated {
		t.Fatalf("create lifecycle one status=%d body=%s", createLifecycleOne.Code, createLifecycleOne.Body.String())
	}
	createLifecycleTwo := doRequest(t, a, http.MethodPost, "/api/commerce/lifecycle", strings.NewReader(`{"item_id":"fwd-override-item-2","state":"purchase","source":"ebay","external_ref":"order-override-2","quantity":1,"amount":43,"currency":"aud"}`), map[string]string{"Content-Type": "application/json"})
	if createLifecycleTwo.Code != http.StatusCreated {
		t.Fatalf("create lifecycle two status=%d body=%s", createLifecycleTwo.Code, createLifecycleTwo.Body.String())
	}
	var lifeOne, lifeTwo struct {
		Entry struct {
			ID string `json:"id"`
		} `json:"entry"`
		ExpectedArrival struct {
			ID string `json:"id"`
		} `json:"expected_arrival"`
	}
	if err := json.NewDecoder(createLifecycleOne.Body).Decode(&lifeOne); err != nil {
		t.Fatalf("decode lifecycle one: %v", err)
	}
	if err := json.NewDecoder(createLifecycleTwo.Body).Decode(&lifeTwo); err != nil {
		t.Fatalf("decode lifecycle two: %v", err)
	}
	createPackage := doRequest(t, a, http.MethodPost, "/api/forwarding/packages", strings.NewReader(`{"profile_id":"`+profileID+`","provider":"stackry","source":"manual","external_package_id":"PKG-OVERRIDE-API","status":"received"}`), map[string]string{"Content-Type": "application/json"})
	if createPackage.Code != http.StatusOK {
		t.Fatalf("create package status=%d body=%s", createPackage.Code, createPackage.Body.String())
	}
	var packagePayload struct {
		Package struct {
			ID string `json:"id"`
		} `json:"package"`
	}
	if err := json.NewDecoder(createPackage.Body).Decode(&packagePayload); err != nil {
		t.Fatalf("decode package payload: %v", err)
	}

	confirmBody := `{"package_id":"` + packagePayload.Package.ID + `","item_id":"fwd-override-item-1","lifecycle_entry_id":"` + lifeOne.Entry.ID + `","expected_arrival_id":"` + lifeOne.ExpectedArrival.ID + `","source":"suggested-match","decision":"confirmed","actor":"reviewer","notes":"accepted suggestion"}`
	confirmResp := doRequest(t, a, http.MethodPost, "/api/forwarding/package-links", strings.NewReader(confirmBody), map[string]string{"Content-Type": "application/json"})
	if confirmResp.Code != http.StatusOK {
		t.Fatalf("confirm package status=%d body=%s", confirmResp.Code, confirmResp.Body.String())
	}
	overrideBody := `{"package_id":"` + packagePayload.Package.ID + `","item_id":"fwd-override-item-2","lifecycle_entry_id":"` + lifeTwo.Entry.ID + `","expected_arrival_id":"` + lifeTwo.ExpectedArrival.ID + `","source":"manual-override","decision":"override","override":true,"actor":"reviewer","notes":"corrected target"}`
	overrideResp := doRequest(t, a, http.MethodPost, "/api/forwarding/package-links", strings.NewReader(overrideBody), map[string]string{"Content-Type": "application/json"})
	if overrideResp.Code != http.StatusOK {
		t.Fatalf("override package status=%d body=%s", overrideResp.Code, overrideResp.Body.String())
	}
	var overridePayload struct {
		Link struct {
			ItemID     string   `json:"item_id"`
			Decision   string   `json:"decision"`
			AuditTrail []string `json:"audit_trail"`
		} `json:"link"`
	}
	if err := json.NewDecoder(overrideResp.Body).Decode(&overridePayload); err != nil {
		t.Fatalf("decode override payload: %v", err)
	}
	if overridePayload.Link.ItemID != "fwd-override-item-2" || overridePayload.Link.Decision != "override" || !strings.Contains(strings.Join(overridePayload.Link.AuditTrail, " "), "previous target item=fwd-override-item-1") {
		t.Fatalf("expected override link payload, got %+v", overridePayload)
	}
	unlinkResp := doRequest(t, a, http.MethodDelete, "/api/forwarding/package-links?package_id="+packagePayload.Package.ID, strings.NewReader(`{"source":"manual-unlink","actor":"reviewer","notes":"remove stale link"}`), map[string]string{"Content-Type": "application/json"})
	if unlinkResp.Code != http.StatusOK {
		t.Fatalf("unlink package status=%d body=%s", unlinkResp.Code, unlinkResp.Body.String())
	}
	listLinks := doRequest(t, a, http.MethodGet, "/api/forwarding/package-links?package_id="+packagePayload.Package.ID, nil, nil)
	if listLinks.Code != http.StatusOK {
		t.Fatalf("list links status=%d body=%s", listLinks.Code, listLinks.Body.String())
	}
	var listPayload struct {
		Links   []map[string]any `json:"links"`
		Events  []map[string]any `json:"events"`
		Summary map[string]int   `json:"summary"`
	}
	if err := json.NewDecoder(listLinks.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode list payload: %v", err)
	}
	if listPayload.Summary["count"] != 0 || listPayload.Summary["events"] != 3 || len(listPayload.Links) != 0 || len(listPayload.Events) != 3 {
		t.Fatalf("expected no active link and three audit events, got %+v", listPayload)
	}
}

func TestForwarderPackageMatchSuggestionsAPIIsNonMutating(t *testing.T) {
	t.Parallel()

	a, profileID := newCommerceProfileApp(t)
	if _, err := a.db.Exec(`INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title) VALUES ('fwd-match-item', ?, 'AFX', 'Slot', 'AFX-901', 'AFX Turbo slot car')`, profileID); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	createLifecycle := doRequest(t, a, http.MethodPost, "/api/commerce/lifecycle", strings.NewReader(`{"item_id":"fwd-match-item","state":"purchase","source":"ebay","external_ref":"ORDER-901","quantity":2,"amount":52,"currency":"aud","notes":"Seller: slot shop; tracking TRACK-901; package PKG-901; AFX Turbo"}`), map[string]string{"Content-Type": "application/json"})
	if createLifecycle.Code != http.StatusCreated {
		t.Fatalf("create lifecycle status=%d body=%s", createLifecycle.Code, createLifecycle.Body.String())
	}
	createPackage := doRequest(t, a, http.MethodPost, "/api/forwarding/packages", strings.NewReader(`{"profile_id":"`+profileID+`","provider":"stackry","source":"email","external_package_id":"PKG-901","status":"received","tracking_number":"TRACK-901","received_at":"2026-05-27T09:15:00Z","sender":"slot shop","raw_payload":{"title":"AFX Turbo slot car","quantity":2}}`), map[string]string{"Content-Type": "application/json"})
	if createPackage.Code != http.StatusOK {
		t.Fatalf("create package status=%d body=%s", createPackage.Code, createPackage.Body.String())
	}
	var packagePayload struct {
		Package struct {
			ID string `json:"id"`
		} `json:"package"`
	}
	if err := json.NewDecoder(createPackage.Body).Decode(&packagePayload); err != nil {
		t.Fatalf("decode package payload: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title) VALUES ('fwd-medium-item', ?, 'AFX', 'Slot', 'AFX-902', 'Medium match item')`, profileID); err != nil {
		t.Fatalf("seed medium item: %v", err)
	}
	createMediumLifecycle := doRequest(t, a, http.MethodPost, "/api/commerce/lifecycle", strings.NewReader(`{"item_id":"fwd-medium-item","state":"purchase","source":"ebay","external_ref":"ORDER-902","quantity":1,"amount":20,"currency":"aud","notes":"tracking TRACK-902 package PKG-902"}`), map[string]string{"Content-Type": "application/json"})
	if createMediumLifecycle.Code != http.StatusCreated {
		t.Fatalf("create medium lifecycle status=%d body=%s", createMediumLifecycle.Code, createMediumLifecycle.Body.String())
	}
	createMediumPackage := doRequest(t, a, http.MethodPost, "/api/forwarding/packages", strings.NewReader(`{"profile_id":"`+profileID+`","provider":"stackry","source":"email","external_package_id":"PKG-902","status":"received","tracking_number":"TRACK-902","sender":"unrelated sender","raw_payload":{"title":"unrelated text","quantity":7}}`), map[string]string{"Content-Type": "application/json"})
	if createMediumPackage.Code != http.StatusOK {
		t.Fatalf("create medium package status=%d body=%s", createMediumPackage.Code, createMediumPackage.Body.String())
	}
	var mediumPackagePayload struct {
		Package struct {
			ID string `json:"id"`
		} `json:"package"`
	}
	if err := json.NewDecoder(createMediumPackage.Body).Decode(&mediumPackagePayload); err != nil {
		t.Fatalf("decode medium package payload: %v", err)
	}

	resp := doRequest(t, a, http.MethodGet, "/api/forwarding/package-match-suggestions", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("suggestions status=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload struct {
		Mode             string           `json:"mode"`
		Mutable          bool             `json:"mutable"`
		ConfidenceFilter string           `json:"confidence_filter"`
		Suggestions      []map[string]any `json:"suggestions"`
		Summary          map[string]int   `json:"summary"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode suggestions payload: %v", err)
	}
	if payload.Mode != "forwarder_package_match_suggestions" || payload.Mutable {
		t.Fatalf("expected non-mutating suggestions mode, got %+v", payload)
	}
	if payload.Summary["count"] != 3 || len(payload.Suggestions) != 3 {
		t.Fatalf("expected three suggestions, got %+v", payload)
	}
	if payload.Summary["high_confidence"] != 1 || payload.Summary["medium_confidence"] != 1 || payload.Summary["low_confidence"] != 1 {
		t.Fatalf("expected confidence-bucketed suggestion summary, got %+v", payload.Summary)
	}
	if payload.Summary["scoped_packages"] != 2 {
		t.Fatalf("expected summary to count packages represented by returned suggestions, got %+v", payload.Summary)
	}
	suggestion := payload.Suggestions[0]
	if suggestion["item_id"] != "fwd-match-item" || suggestion["confidence_label"] != "high" {
		t.Fatalf("expected high confidence item suggestion, got %+v", suggestion)
	}
	if suggestion["confidence_score"] != float64(0.95) {
		t.Fatalf("expected API confidence_score for UI rendering, got %+v", suggestion)
	}
	if _, ok := suggestion["confidence"]; ok {
		t.Fatalf("suggestion API must not expose stale confidence field, got %+v", suggestion)
	}
	signals, ok := suggestion["signals"].([]any)
	if !ok || len(signals) == 0 {
		t.Fatalf("expected scored suggestion signals, got %+v", suggestion["signals"])
	}
	firstSignal, ok := signals[0].(map[string]any)
	if !ok {
		t.Fatalf("expected object suggestion signal, got %+v", signals[0])
	}
	if _, ok := firstSignal["score"]; !ok {
		t.Fatalf("expected signal score for UI rendering, got %+v", firstSignal)
	}
	if _, ok := firstSignal["weight"]; ok {
		t.Fatalf("suggestion signal API must not expose stale weight field, got %+v", firstSignal)
	}
	filteredResp := doRequest(t, a, http.MethodGet, "/api/forwarding/package-match-suggestions?confidence_label=medium", nil, nil)
	if filteredResp.Code != http.StatusOK {
		t.Fatalf("filtered suggestions status=%d body=%s", filteredResp.Code, filteredResp.Body.String())
	}
	var filteredPayload struct {
		ConfidenceFilter string           `json:"confidence_filter"`
		Suggestions      []map[string]any `json:"suggestions"`
		Summary          map[string]int   `json:"summary"`
	}
	if err := json.NewDecoder(filteredResp.Body).Decode(&filteredPayload); err != nil {
		t.Fatalf("decode filtered suggestions payload: %v", err)
	}
	if filteredPayload.ConfidenceFilter != "medium" || len(filteredPayload.Suggestions) != 1 || filteredPayload.Summary["medium_confidence"] != 1 || filteredPayload.Summary["high_confidence"] != 0 {
		t.Fatalf("expected medium-only filtered suggestions with filtered summary, got %+v", filteredPayload)
	}
	if filteredPayload.Suggestions[0]["item_id"] != "fwd-medium-item" || filteredPayload.Suggestions[0]["confidence_label"] != "medium" {
		t.Fatalf("expected medium confidence item suggestion, got %+v", filteredPayload.Suggestions[0])
	}
	scopedResp := doRequest(t, a, http.MethodGet, "/api/forwarding/package-match-suggestions?package_id="+packagePayload.Package.ID+"&confidence_label=high", nil, nil)
	if scopedResp.Code != http.StatusOK {
		t.Fatalf("scoped suggestions status=%d body=%s", scopedResp.Code, scopedResp.Body.String())
	}
	var scopedPayload struct {
		ConfidenceFilter string           `json:"confidence_filter"`
		Suggestions      []map[string]any `json:"suggestions"`
		Summary          map[string]int   `json:"summary"`
	}
	if err := json.NewDecoder(scopedResp.Body).Decode(&scopedPayload); err != nil {
		t.Fatalf("decode scoped suggestions payload: %v", err)
	}
	if scopedPayload.ConfidenceFilter != "high" || len(scopedPayload.Suggestions) != 1 {
		t.Fatalf("expected one scoped high-confidence suggestion, got %+v", scopedPayload)
	}
	if scopedPayload.Summary["count"] != 1 || scopedPayload.Summary["scoped_packages"] != 1 || scopedPayload.Summary["high_confidence"] != 1 || scopedPayload.Summary["medium_confidence"] != 0 {
		t.Fatalf("expected scoped filtered summary, got %+v", scopedPayload.Summary)
	}
	if scopedPayload.Suggestions[0]["package_id"] != packagePayload.Package.ID || scopedPayload.Suggestions[0]["confidence_label"] != "high" {
		t.Fatalf("expected scoped package high-confidence suggestion, got %+v", scopedPayload.Suggestions[0])
	}
	scopedEmptyResp := doRequest(t, a, http.MethodGet, "/api/forwarding/package-match-suggestions?package_id="+mediumPackagePayload.Package.ID+"&confidence_label=high", nil, nil)
	if scopedEmptyResp.Code != http.StatusOK {
		t.Fatalf("scoped empty suggestions status=%d body=%s", scopedEmptyResp.Code, scopedEmptyResp.Body.String())
	}
	var scopedEmptyPayload struct {
		ConfidenceFilter string           `json:"confidence_filter"`
		Suggestions      []map[string]any `json:"suggestions"`
		Summary          map[string]int   `json:"summary"`
	}
	if err := json.NewDecoder(scopedEmptyResp.Body).Decode(&scopedEmptyPayload); err != nil {
		t.Fatalf("decode scoped empty suggestions payload: %v", err)
	}
	if scopedEmptyPayload.ConfidenceFilter != "high" || len(scopedEmptyPayload.Suggestions) != 0 {
		t.Fatalf("expected no high-confidence suggestions for medium package, got %+v", scopedEmptyPayload)
	}
	if scopedEmptyPayload.Summary["count"] != 0 || scopedEmptyPayload.Summary["scoped_packages"] != 0 || scopedEmptyPayload.Summary["high_confidence"] != 0 {
		t.Fatalf("expected empty scoped summary derived from returned suggestions, got %+v", scopedEmptyPayload.Summary)
	}
	invalidFilter := doRequest(t, a, http.MethodGet, "/api/forwarding/package-match-suggestions?confidence_label=certain", nil, nil)
	if invalidFilter.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid confidence filter to return 400, got %d body=%s", invalidFilter.Code, invalidFilter.Body.String())
	}
	listLinks := doRequest(t, a, http.MethodGet, "/api/forwarding/package-links", nil, nil)
	if listLinks.Code != http.StatusOK {
		t.Fatalf("list links status=%d body=%s", listLinks.Code, listLinks.Body.String())
	}
	var linksPayload struct {
		Summary map[string]int `json:"summary"`
	}
	if err := json.NewDecoder(listLinks.Body).Decode(&linksPayload); err != nil {
		t.Fatalf("decode links payload: %v", err)
	}
	if linksPayload.Summary["count"] != 0 {
		t.Fatalf("suggestions must not create reconciliation links, got %+v", linksPayload)
	}
}

func TestForwarderPackageCSVImportAPIUpsertsValidRowsAndReportsRowErrors(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	body := `{
		"profile_id":"profile-csv-api",
		"provider":"Stackry",
		"csv":"Stackry Package ID,Status,Shipment ID,Tracking Number,Warehouse Location,Weight Grams\nPKG-CSV-API-1,received,SHIP-1,TRACK-1,A-12,425\n,received,SHIP-2,TRACK-2,B-7,510\nPKG-CSV-API-3,ready_to_ship,SHIP-3,TRACK-3,C-4,invalid\n"
	}`
	resp := doRequest(t, a, http.MethodPost, "/api/forwarding/packages/import-csv", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusOK {
		t.Fatalf("csv import status=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload struct {
		Imported []map[string]any `json:"imported"`
		Errors   []map[string]any `json:"errors"`
		Summary  map[string]int   `json:"summary"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode csv import payload: %v", err)
	}
	if payload.Summary["imported"] != 1 || payload.Summary["errors"] != 2 {
		t.Fatalf("expected one imported row and two errors, got %+v", payload.Summary)
	}
	if len(payload.Imported) != 1 || payload.Imported[0]["external_package_id"] != "PKG-CSV-API-1" || payload.Imported[0]["source"] != "csv" {
		t.Fatalf("expected imported CSV package response, got %+v", payload.Imported)
	}
	if len(payload.Errors) != 2 {
		t.Fatalf("expected two row errors, got %+v", payload.Errors)
	}
	listed := doRequest(t, a, http.MethodGet, "/api/forwarding/packages?profile_id=profile-csv-api", nil, nil)
	if listed.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listed.Code, listed.Body.String())
	}
	var listPayload struct {
		Packages []map[string]any
		Summary  map[string]int
	}
	if err := json.NewDecoder(listed.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode list payload: %v", err)
	}
	if listPayload.Summary["count"] != 1 || listPayload.Packages[0]["provenance_key"] != "stackry:csv:PKG-CSV-API-1" {
		t.Fatalf("expected only valid CSV package to persist, got %+v", listPayload)
	}
}

func TestForwarderPackageEmailImportAPIUpsertsNotice(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	body := `{
		"profile_id":"profile-email-api",
		"provider":"Stackry",
		"message_id":"msg-stackry-001",
		"body":"Package ID: PKG-EMAIL-API-1\nStatus: Received\nShipment ID: SHIP-EMAIL-1\nTracking Number: TRACK-EMAIL-1\nWarehouse Location: Locker E-5\nWeight Grams: 640\nSender: Stackry Intake"
	}`
	resp := doRequest(t, a, http.MethodPost, "/api/forwarding/packages/import-email", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusOK {
		t.Fatalf("email import status=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload struct {
		Package map[string]any `json:"package"`
		Mode    string         `json:"mode"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode email import payload: %v", err)
	}
	if payload.Mode != "forwarder_package_email_import" || payload.Package["external_package_id"] != "PKG-EMAIL-API-1" || payload.Package["source"] != "email" {
		t.Fatalf("expected imported email package response, got %+v", payload)
	}
	listed := doRequest(t, a, http.MethodGet, "/api/forwarding/packages?profile_id=profile-email-api", nil, nil)
	if listed.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listed.Code, listed.Body.String())
	}
	var listPayload struct {
		Packages []map[string]any
		Summary  map[string]int
	}
	if err := json.NewDecoder(listed.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode list payload: %v", err)
	}
	if listPayload.Summary["count"] != 1 || listPayload.Packages[0]["provenance_key"] != "stackry:email:PKG-EMAIL-API-1" {
		t.Fatalf("expected email package to persist, got %+v", listPayload)
	}
}

func TestForwarderPackageEmailImportAPIRejectsInvalidNotice(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	body := `{"profile_id":"profile-email-api","provider":"Stackry","message_id":"msg-stackry-invalid","body":"Status: Received\nTracking Number: TRACK-EMAIL-2"}`
	resp := doRequest(t, a, http.MethodPost, "/api/forwarding/packages/import-email", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid email import to fail, status=%d body=%s", resp.Code, resp.Body.String())
	}
}

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

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

package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/logging"
)

func TestWave6ProfilesIsolationStorageAndSecrets(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createP1 := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wave6P1"}`), map[string]string{"Content-Type": "application/json"})
	createP2 := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wave6P2"}`), map[string]string{"Content-Type": "application/json"})
	if createP1.Code != http.StatusCreated || createP2.Code != http.StatusCreated {
		t.Fatalf("create profiles failed p1=%d p2=%d", createP1.Code, createP2.Code)
	}
	var p1 struct {
		ID string `json:"id"`
	}
	var p2 struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(createP1.Body).Decode(&p1)
	_ = json.NewDecoder(createP2.Body).Decode(&p2)

	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p1.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	_ = doRequest(t, a, http.MethodPut, "/api/profiles/"+p1.ID+"/secrets", strings.NewReader(`{"key":"ebay_token","value":"token-p1"}`), map[string]string{"Content-Type": "application/json"})
	createP1Item := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"W6-P1-001","title":"P1 Item","brand":"AFX","category":"Slot"}`), map[string]string{"Content-Type": "application/json"})
	if createP1Item.Code != http.StatusCreated {
		t.Fatalf("create p1 item status=%d body=%s", createP1Item.Code, createP1Item.Body.String())
	}

	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p2.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	listP2Items := doRequest(t, a, http.MethodGet, "/api/items", nil, nil)
	if listP2Items.Code != http.StatusOK {
		t.Fatalf("list p2 items status=%d body=%s", listP2Items.Code, listP2Items.Body.String())
	}
	var p2Items struct {
		Items []map[string]any `json:"items"`
	}
	_ = json.NewDecoder(listP2Items.Body).Decode(&p2Items)
	if len(p2Items.Items) != 0 {
		t.Fatalf("expected isolated p2 items view, got %+v", p2Items.Items)
	}
	getP2Secret := doRequest(t, a, http.MethodGet, "/api/profiles/"+p2.ID+"/secrets?key=ebay_token", nil, nil)
	if getP2Secret.Code != http.StatusBadRequest {
		t.Fatalf("expected missing secret isolation response 400, got code=%d body=%s", getP2Secret.Code, getP2Secret.Body.String())
	}
	if strings.Contains(getP2Secret.Body.String(), "token-p1") {
		t.Fatalf("secret from p1 leaked into p2 response: %s", getP2Secret.Body.String())
	}
}

func TestDataExportsAreScopedToActiveProfile(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createP1 := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"ExportP1"}`), map[string]string{"Content-Type": "application/json"})
	createP2 := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"ExportP2"}`), map[string]string{"Content-Type": "application/json"})
	if createP1.Code != http.StatusCreated || createP2.Code != http.StatusCreated {
		t.Fatalf("create profiles failed p1=%d p2=%d", createP1.Code, createP2.Code)
	}
	var p1 struct {
		ID string `json:"id"`
	}
	var p2 struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(createP1.Body).Decode(&p1)
	_ = json.NewDecoder(createP2.Body).Decode(&p2)

	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p1.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	createP1Item := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"EXPORT-P1-ONLY","title":"P1 export item","brand":"AFX","category":"Slot"}`), map[string]string{"Content-Type": "application/json"})
	if createP1Item.Code != http.StatusCreated {
		t.Fatalf("create p1 item status=%d body=%s", createP1Item.Code, createP1Item.Body.String())
	}

	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p2.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	createP2Item := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"EXPORT-P2-ONLY","title":"P2 export item","brand":"AFX","category":"Slot"}`), map[string]string{"Content-Type": "application/json"})
	if createP2Item.Code != http.StatusCreated {
		t.Fatalf("create p2 item status=%d body=%s", createP2Item.Code, createP2Item.Body.String())
	}

	jsonExport := doRequest(t, a, http.MethodGet, "/api/data/export/json", nil, nil)
	if jsonExport.Code != http.StatusOK {
		t.Fatalf("json export status=%d body=%s", jsonExport.Code, jsonExport.Body.String())
	}
	if body := jsonExport.Body.String(); !strings.Contains(body, "EXPORT-P2-ONLY") || strings.Contains(body, "EXPORT-P1-ONLY") {
		t.Fatalf("json export should include only active profile records, got %s", body)
	}

	csvExport := doRequest(t, a, http.MethodGet, "/api/data/export/csv/items", nil, nil)
	if csvExport.Code != http.StatusOK {
		t.Fatalf("csv export status=%d body=%s", csvExport.Code, csvExport.Body.String())
	}
	if body := csvExport.Body.String(); !strings.Contains(body, "EXPORT-P2-ONLY") || strings.Contains(body, "EXPORT-P1-ONLY") {
		t.Fatalf("csv export should include only active profile records, got %s", body)
	}
}

func TestWave6CollectionMetadataInstancesAndNoAutoMerge(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/items",
		strings.NewReader(`{"part_number":"W6-COL-001","title":"Collector Item","brand":"AFX","category":"Slot","make":"Chevrolet","model":"Camaro","year":"1970","scale":"HO","series":"Mega G+","description":"Wave6 metadata","tags":["mint","sealed"]}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if create.Code != http.StatusCreated {
		t.Fatalf("create item status=%d body=%s", create.Code, create.Body.String())
	}
	var item map[string]any
	_ = json.NewDecoder(create.Body).Decode(&item)
	itemID, _ := item["id"].(string)
	for _, field := range []string{"brand", "category", "part_number", "title", "make", "model", "year", "scale", "series", "description", "created_at", "updated_at"} {
		if _, ok := item[field]; !ok {
			t.Fatalf("item missing metadata field %q: %+v", field, item)
		}
	}

	createInstance := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/items/"+itemID+"/instances",
		strings.NewReader(`{"condition":"mint","status":"sealed","quantity":1,"storage_location":"Shelf A","acquisition_price":49.95}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if createInstance.Code != http.StatusCreated {
		t.Fatalf("create instance status=%d body=%s", createInstance.Code, createInstance.Body.String())
	}
	instances := doRequest(t, a, http.MethodGet, "/api/items/"+itemID+"/instances", nil, nil)
	if instances.Code != http.StatusOK {
		t.Fatalf("list instances status=%d body=%s", instances.Code, instances.Body.String())
	}
	var instancePayload struct {
		Instances []map[string]any `json:"instances"`
	}
	_ = json.NewDecoder(instances.Body).Decode(&instancePayload)
	if len(instancePayload.Instances) != 1 {
		t.Fatalf("expected one instance for item, got %d", len(instancePayload.Instances))
	}

	duplicate := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"W6-COL-001","title":"Collector Item Duplicate","brand":"AFX","category":"Slot"}`), map[string]string{"Content-Type": "application/json"})
	if duplicate.Code != http.StatusBadRequest {
		t.Fatalf("duplicate create should require explicit action (no auto-merge); expected 400 got %d body=%s", duplicate.Code, duplicate.Body.String())
	}
	allItems := doRequest(t, a, http.MethodGet, "/api/items", nil, nil)
	var allPayload struct {
		Items []map[string]any `json:"items"`
	}
	_ = json.NewDecoder(allItems.Body).Decode(&allPayload)
	if len(allPayload.Items) != 1 {
		t.Fatalf("expected no implicit merge/create for duplicate identity; got %+v", allPayload.Items)
	}
}

func TestWave6ScannerQuerySetCreateAndFailureRetry(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wave6Scanner"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(createProfile.Body).Decode(&profile)
	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})

	createQuery := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/scanner/query-sets",
		strings.NewReader(`{"name":"W6 Query","keywords":["afx","camaro"],"exclusions":["broken"],"max_price":70,"region":"AU","condition":"new","schedule_cron":"*/10 * * * *","enabled":true,"rate_limit_rps":2,"max_retry_count":2}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if createQuery.Code != http.StatusCreated {
		t.Fatalf("create query status=%d body=%s", createQuery.Code, createQuery.Body.String())
	}
	var query map[string]any
	_ = json.NewDecoder(createQuery.Body).Decode(&query)
	querySetID, _ := query["id"].(string)
	if querySetID == "" {
		t.Fatal("expected stable query set id")
	}
	for _, field := range []string{"keywords", "exclusions", "max_price", "region", "condition", "schedule_cron", "rate_limit_rps", "max_retry_count"} {
		if _, ok := query[field]; !ok {
			t.Fatalf("query set missing field %q: %+v", field, query)
		}
	}
	if _, err := a.db.Exec(`INSERT INTO scanner_failures(id, query_set_id, provider, message, created_at) VALUES ('w6-f1', ?, 'ebay', 'timeout', CURRENT_TIMESTAMP)`, querySetID); err != nil {
		t.Fatalf("seed scanner failure: %v", err)
	}
	retry := doRequest(t, a, http.MethodPost, "/api/scanner/failures/retry", strings.NewReader(`{"query_set_id":"`+querySetID+`"}`), map[string]string{"Content-Type": "application/json"})
	if retry.Code != http.StatusOK {
		t.Fatalf("retry status=%d body=%s", retry.Code, retry.Body.String())
	}
	if !strings.Contains(retry.Body.String(), `"retry_started":true`) {
		t.Fatalf("expected retry_started contract, got %s", retry.Body.String())
	}
}

func TestWave6DataImportDryRunAndMaintenanceContracts(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	seed := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"W6-DATA-001","title":"Seed Item","brand":"AFX","category":"Slot"}`), map[string]string{"Content-Type": "application/json"})
	if seed.Code != http.StatusCreated {
		t.Fatalf("seed item status=%d body=%s", seed.Code, seed.Body.String())
	}

	dryRun := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/data/import/json/dry-run",
		strings.NewReader(`{"snapshot":{"schema_version":1,"items":[{"brand":"AFX","category":"Slot","part_number":"W6-DATA-001","title":"Conflict"},{"brand":"AFX","category":"Slot","part_number":"W6-DATA-NEW","title":"New Item"}]}}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if dryRun.Code != http.StatusOK {
		t.Fatalf("dry-run status=%d body=%s", dryRun.Code, dryRun.Body.String())
	}
	var dryPayload map[string]any
	_ = json.NewDecoder(dryRun.Body).Decode(&dryPayload)
	if dryPayload["conflicts"].(float64) < 1 {
		t.Fatalf("expected at least one dry-run conflict, got %+v", dryPayload)
	}

	itemsAfterDryRun := doRequest(t, a, http.MethodGet, "/api/items", nil, nil)
	var itemsPayload struct {
		Items []map[string]any `json:"items"`
	}
	_ = json.NewDecoder(itemsAfterDryRun.Body).Decode(&itemsPayload)
	if len(itemsPayload.Items) != 1 {
		t.Fatalf("dry-run must not mutate persisted records, got %+v", itemsPayload.Items)
	}

	apply := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/data/import/json/apply",
		strings.NewReader(`{"snapshot":{"schema_version":1,"items":[{"brand":"AFX","category":"Slot","part_number":"W6-DATA-001","title":"Conflict"},{"brand":"AFX","category":"Slot","part_number":"W6-DATA-NEW","title":"New Item"}]},"options":{"default_action":"merge"}}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if apply.Code != http.StatusOK {
		t.Fatalf("apply status=%d body=%s", apply.Code, apply.Body.String())
	}
	var applyPayload struct {
		TotalItems int `json:"total_items"`
		Created    int `json:"created"`
		Merged     int `json:"merged"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	}
	_ = json.NewDecoder(apply.Body).Decode(&applyPayload)
	if applyPayload.TotalItems != 2 || applyPayload.Created != 1 || applyPayload.Merged != 1 || applyPayload.Skipped != 0 || applyPayload.Failed != 0 {
		t.Fatalf("unexpected apply summary: %+v body=%s", applyPayload, apply.Body.String())
	}

	jsonExport := doRequest(t, a, http.MethodGet, "/api/data/export/json", nil, nil)
	if jsonExport.Code != http.StatusOK {
		t.Fatalf("json export status=%d body=%s", jsonExport.Code, jsonExport.Body.String())
	}
	if got := jsonExport.Header().Get("Content-Disposition"); !strings.Contains(got, "cabinet-data-snapshot.json") {
		t.Fatalf("json export missing download filename, got %q", got)
	}
	if !strings.Contains(jsonExport.Body.String(), "W6-DATA-NEW") {
		t.Fatalf("json export missing applied item evidence: %s", jsonExport.Body.String())
	}

	csvExport := doRequest(t, a, http.MethodGet, "/api/data/export/csv/items", nil, nil)
	if csvExport.Code != http.StatusOK {
		t.Fatalf("csv export status=%d body=%s", csvExport.Code, csvExport.Body.String())
	}
	if got := csvExport.Header().Get("Content-Disposition"); !strings.Contains(got, "cabinet-items.csv") {
		t.Fatalf("csv export missing download filename, got %q", got)
	}
	if !strings.Contains(csvExport.Body.String(), "W6-DATA-NEW") {
		t.Fatalf("csv export missing applied item evidence: %s", csvExport.Body.String())
	}

	reindex := doRequest(t, a, http.MethodPost, "/api/data/reindex", strings.NewReader(`{}`), map[string]string{"Content-Type": "application/json"})
	if reindex.Code != http.StatusOK {
		t.Fatalf("reindex status=%d body=%s", reindex.Code, reindex.Body.String())
	}
	if !strings.Contains(reindex.Body.String(), `"operation":"reindex_search"`) || !strings.Contains(reindex.Body.String(), `"rebuilt_search_index":true`) {
		t.Fatalf("reindex response missing operation metadata: %s", reindex.Body.String())
	}
	repair := doRequest(t, a, http.MethodPost, "/api/data/repair", strings.NewReader(`{}`), map[string]string{"Content-Type": "application/json"})
	if repair.Code != http.StatusOK {
		t.Fatalf("repair status=%d body=%s", repair.Code, repair.Body.String())
	}
	if !strings.Contains(repair.Body.String(), `"operation":"integrity_check"`) || !strings.Contains(repair.Body.String(), `"integrity_check":"ok"`) {
		t.Fatalf("repair response missing integrity metadata: %s", repair.Body.String())
	}
}

func TestWave6LoggingDebugToggleAndRedactedExport(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wave6Logs"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(createProfile.Body).Decode(&profile)

	toggle := doRequest(t, a, http.MethodPost, "/api/logs/debug", strings.NewReader(`{"profile_id":"`+profile.ID+`","enabled":true}`), map[string]string{"Content-Type": "application/json"})
	if toggle.Code != http.StatusOK {
		t.Fatalf("toggle debug status=%d body=%s", toggle.Code, toggle.Body.String())
	}
	logging.NewService(a.db).Log(t.Context(), "error", "wave6_sensitive", map[string]any{
		"api_key": "sk-wave6-secret",
		"token":   "tok-wave6-secret",
	})
	export := doRequest(t, a, http.MethodGet, "/api/logs/export", nil, nil)
	if export.Code != http.StatusOK {
		t.Fatalf("logs export status=%d body=%s", export.Code, export.Body.String())
	}
	if !strings.Contains(export.Body.String(), "[REDACTED]") {
		t.Fatalf("expected redacted export content, got %s", export.Body.String())
	}
	if strings.Contains(export.Body.String(), "sk-wave6-secret") || strings.Contains(export.Body.String(), "tok-wave6-secret") {
		t.Fatalf("expected sensitive values redacted in export, got %s", export.Body.String())
	}
}

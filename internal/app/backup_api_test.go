package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestBackupRunAndRestoreEndpoints(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	if _, err := a.db.Exec(`INSERT INTO canonical_items (id, brand, category, part_number, title) VALUES ('item-a','AFX','Slot Car','P-1','Car')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}

	runResp := doRequest(t, a, http.MethodPost, "/api/backup/run", nil, nil)
	if runResp.Code != http.StatusOK {
		t.Fatalf("backup run status = %d body=%s", runResp.Code, runResp.Body.String())
	}
	var runPayload struct {
		BackupPath string `json:"backup_path"`
	}
	if err := json.NewDecoder(runResp.Body).Decode(&runPayload); err != nil {
		t.Fatalf("decode backup run response: %v", err)
	}
	if strings.TrimSpace(runPayload.BackupPath) == "" {
		t.Fatal("expected backup path in response")
	}

	if _, err := a.db.Exec(`INSERT INTO canonical_items (id, brand, category, part_number, title) VALUES ('item-b','AFX','Slot Car','P-2','Car')`); err != nil {
		t.Fatalf("seed second item: %v", err)
	}

	restoreBody := bytes.NewBufferString(`{"backup_path":"` + strings.ReplaceAll(runPayload.BackupPath, `\`, `\\`) + `"}`)
	restoreResp := doRequest(t, a, http.MethodPost, "/api/backup/restore", restoreBody, map[string]string{"Content-Type": "application/json"})
	if restoreResp.Code != http.StatusOK {
		t.Fatalf("backup restore status = %d body=%s", restoreResp.Code, restoreResp.Body.String())
	}

	itemsResp := doRequest(t, a, http.MethodGet, "/api/items", nil, nil)
	if itemsResp.Code != http.StatusOK {
		t.Fatalf("items status = %d body=%s", itemsResp.Code, itemsResp.Body.String())
	}
	var itemsPayload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(itemsResp.Body).Decode(&itemsPayload); err != nil {
		t.Fatalf("decode items response: %v", err)
	}
	if len(itemsPayload.Items) != 1 {
		t.Fatalf("expected 1 item after restore, got %d", len(itemsPayload.Items))
	}
}

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
		Backup struct {
			Path           string `json:"path"`
			FileName       string `json:"file_name"`
			SizeBytes      int64  `json:"size_bytes"`
			CreatedAt      string `json:"created_at"`
			IntegrityCheck string `json:"integrity_check"`
		} `json:"backup"`
	}
	if err := json.NewDecoder(runResp.Body).Decode(&runPayload); err != nil {
		t.Fatalf("decode backup run response: %v", err)
	}
	if strings.TrimSpace(runPayload.Backup.Path) == "" || runPayload.Backup.FileName == "" || runPayload.Backup.SizeBytes == 0 || runPayload.Backup.IntegrityCheck != "ok" {
		t.Fatalf("expected backup metadata in response, got %+v", runPayload.Backup)
	}

	listResp := doRequest(t, a, http.MethodGet, "/api/backup/list", nil, nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("backup list status = %d body=%s", listResp.Code, listResp.Body.String())
	}
	var listPayload struct {
		Backups []struct {
			Path      string `json:"path"`
			FileName  string `json:"file_name"`
			SizeBytes int64  `json:"size_bytes"`
			CreatedAt string `json:"created_at"`
		} `json:"backups"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode backup list response: %v", err)
	}
	if len(listPayload.Backups) != 1 || listPayload.Backups[0].Path != runPayload.Backup.Path || listPayload.Backups[0].SizeBytes == 0 {
		t.Fatalf("expected listed backup metadata, got %+v", listPayload.Backups)
	}

	if _, err := a.db.Exec(`INSERT INTO canonical_items (id, brand, category, part_number, title) VALUES ('item-b','AFX','Slot Car','P-2','Car')`); err != nil {
		t.Fatalf("seed second item: %v", err)
	}

	unconfirmedBody := bytes.NewBufferString(`{"backup_path":"` + strings.ReplaceAll(runPayload.Backup.Path, `\`, `\\`) + `"}`)
	unconfirmedResp := doRequest(t, a, http.MethodPost, "/api/backup/restore", unconfirmedBody, map[string]string{"Content-Type": "application/json"})
	if unconfirmedResp.Code != http.StatusBadRequest {
		t.Fatalf("backup restore without confirmation status = %d body=%s", unconfirmedResp.Code, unconfirmedResp.Body.String())
	}
	if !strings.Contains(unconfirmedResp.Body.String(), "restore_confirmation_required") {
		t.Fatalf("expected confirmation error, got %s", unconfirmedResp.Body.String())
	}

	restoreBody := bytes.NewBufferString(`{"backup_path":"` + strings.ReplaceAll(runPayload.Backup.Path, `\`, `\\`) + `","confirm_restore":true}`)
	restoreResp := doRequest(t, a, http.MethodPost, "/api/backup/restore", restoreBody, map[string]string{"Content-Type": "application/json"})
	if restoreResp.Code != http.StatusOK {
		t.Fatalf("backup restore status = %d body=%s", restoreResp.Code, restoreResp.Body.String())
	}
	var restorePayload struct {
		Restore struct {
			RestoredPath   string `json:"restored_path"`
			RestoredAt     string `json:"restored_at"`
			IntegrityCheck string `json:"integrity_check"`
		} `json:"restore"`
	}
	if err := json.NewDecoder(restoreResp.Body).Decode(&restorePayload); err != nil {
		t.Fatalf("decode backup restore response: %v", err)
	}
	if restorePayload.Restore.RestoredPath != runPayload.Backup.Path || restorePayload.Restore.IntegrityCheck != "ok" || restorePayload.Restore.RestoredAt == "" {
		t.Fatalf("expected restore metadata, got %+v", restorePayload.Restore)
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

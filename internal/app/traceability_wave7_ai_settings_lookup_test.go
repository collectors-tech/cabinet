package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWave7AIContractsMissingKeySuggestAndToggle(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wave7AI"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(createProfile.Body).Decode(&profile)

	_ = doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(`{"settings":{"ai_enabled":"true"}}`), map[string]string{"Content-Type": "application/json"})

	missingKey := doRequest(t, a, http.MethodPost, "/api/ai/test", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if missingKey.Code != http.StatusBadRequest {
		t.Fatalf("ai test missing key expected 400, got %d body=%s", missingKey.Code, missingKey.Body.String())
	}

	aiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o-mini"}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"part_number\":\"W7-1\",\"brand\":\"AFX\",\"title\":\"AFX W7-1\",\"confidence\":0.84}"}}]}`))
	}))
	defer aiServer.Close()

	_ = doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(`{"settings":{"openai_base_url":"`+aiServer.URL+`","ai_enabled":"true"}}`), map[string]string{"Content-Type": "application/json"})
	_ = doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/secrets", strings.NewReader(`{"key":"openai_api_key","value":"sk-wave7"}`), map[string]string{"Content-Type": "application/json"})

	suggest := doRequest(t, a, http.MethodPost, "/api/ai/suggest/title", strings.NewReader(`{"profile_id":"`+profile.ID+`","title":"AFX W7-1 listing"}`), map[string]string{"Content-Type": "application/json"})
	if suggest.Code != http.StatusOK {
		t.Fatalf("ai suggest status=%d body=%s", suggest.Code, suggest.Body.String())
	}
	var suggestion map[string]any
	_ = json.NewDecoder(suggest.Body).Decode(&suggestion)
	if _, ok := suggestion["confidence"]; !ok {
		t.Fatalf("expected confidence in suggestion payload: %+v", suggestion)
	}

	toggleOff := doRequest(t, a, http.MethodPost, "/api/ai/toggle", strings.NewReader(`{"profile_id":"`+profile.ID+`","enabled":false}`), map[string]string{"Content-Type": "application/json"})
	if toggleOff.Code != http.StatusOK {
		t.Fatalf("ai toggle off status=%d body=%s", toggleOff.Code, toggleOff.Body.String())
	}
	disabled := doRequest(t, a, http.MethodPost, "/api/ai/test", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if disabled.Code != http.StatusBadRequest {
		t.Fatalf("disabled ai expected 400, got %d body=%s", disabled.Code, disabled.Body.String())
	}
}

func TestWave7BarcodeLookupAndVariantDuplicateResolution(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createA := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"W7-BC-A","title":"Barcode A","brand":"AFX","category":"Slot"}`), map[string]string{"Content-Type": "application/json"})
	createB := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"W7-BC-B","title":"Barcode B","brand":"AFX","category":"Slot"}`), map[string]string{"Content-Type": "application/json"})
	if createA.Code != http.StatusCreated || createB.Code != http.StatusCreated {
		t.Fatalf("create items failed A=%d B=%d", createA.Code, createB.Code)
	}
	var itemA struct{ ID string `json:"id"` }
	var itemB struct{ ID string `json:"id"` }
	_ = json.NewDecoder(createA.Body).Decode(&itemA)
	_ = json.NewDecoder(createB.Body).Decode(&itemB)

	addA := doRequest(t, a, http.MethodPost, "/api/items/"+itemA.ID+"/barcodes", strings.NewReader(`{"barcode":"1234567890"}`), map[string]string{"Content-Type": "application/json"})
	addB := doRequest(t, a, http.MethodPost, "/api/items/"+itemB.ID+"/barcodes", strings.NewReader(`{"barcode":"1234567890"}`), map[string]string{"Content-Type": "application/json"})
	if addA.Code != http.StatusCreated || addB.Code != http.StatusCreated {
		t.Fatalf("add barcode failed A=%d B=%d", addA.Code, addB.Code)
	}

	localLookup := doRequest(t, a, http.MethodGet, "/api/barcodes/1234567890", nil, nil)
	if localLookup.Code != http.StatusOK {
		t.Fatalf("local barcode lookup status=%d body=%s", localLookup.Code, localLookup.Body.String())
	}
	var lookup struct {
		Matches []map[string]any `json:"matches"`
	}
	_ = json.NewDecoder(localLookup.Body).Decode(&lookup)
	if len(lookup.Matches) < 2 {
		t.Fatalf("expected duplicate barcode links preserved for variant resolution, got %+v", lookup.Matches)
	}

	external := doRequest(t, a, http.MethodGet, "/api/barcodes/1234567890/external-search?source=ebay&region=AU", nil, nil)
	if external.Code != http.StatusOK {
		t.Fatalf("external lookup status=%d body=%s", external.Code, external.Body.String())
	}
	if !strings.Contains(external.Body.String(), "ebay") {
		t.Fatalf("expected external lookup provider url, got %s", external.Body.String())
	}
}

func TestWave7SettingsProfileScopedPersistenceContract(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createP1 := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wave7Settings1"}`), map[string]string{"Content-Type": "application/json"})
	createP2 := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wave7Settings2"}`), map[string]string{"Content-Type": "application/json"})
	if createP1.Code != http.StatusCreated || createP2.Code != http.StatusCreated {
		t.Fatalf("create profiles failed p1=%d p2=%d", createP1.Code, createP2.Code)
	}
	var p1 struct{ ID string `json:"id"` }
	var p2 struct{ ID string `json:"id"` }
	_ = json.NewDecoder(createP1.Body).Decode(&p1)
	_ = json.NewDecoder(createP2.Body).Decode(&p2)

	saveP1 := doRequest(t, a, http.MethodPut, "/api/profiles/"+p1.ID+"/settings", strings.NewReader(`{"settings":{"theme":"dark","scanner_schedule":"0 */6 * * *","update_channel":"stable","backup_frequency":"daily","database_location":"profile_store"}}`), map[string]string{"Content-Type": "application/json"})
	if saveP1.Code != http.StatusOK {
		t.Fatalf("save p1 settings status=%d body=%s", saveP1.Code, saveP1.Body.String())
	}
	_ = doRequest(t, a, http.MethodPut, "/api/profiles/"+p1.ID+"/secrets", strings.NewReader(`{"key":"openai_api_key","value":"sk-settings-p1"}`), map[string]string{"Content-Type": "application/json"})

	getP1 := doRequest(t, a, http.MethodGet, "/api/profiles/"+p1.ID+"/settings", nil, nil)
	if getP1.Code != http.StatusOK || !strings.Contains(getP1.Body.String(), `"theme":"dark"`) {
		t.Fatalf("expected persisted p1 settings, got code=%d body=%s", getP1.Code, getP1.Body.String())
	}

	getP2 := doRequest(t, a, http.MethodGet, "/api/profiles/"+p2.ID+"/settings", nil, nil)
	if getP2.Code != http.StatusOK {
		t.Fatalf("get p2 settings status=%d body=%s", getP2.Code, getP2.Body.String())
	}
	if strings.Contains(getP2.Body.String(), `"theme":"dark"`) {
		t.Fatalf("p2 should not inherit p1 settings: %s", getP2.Body.String())
	}

	getP2Secret := doRequest(t, a, http.MethodGet, "/api/profiles/"+p2.ID+"/secrets?key=openai_api_key", nil, nil)
	if getP2Secret.Code != http.StatusBadRequest {
		t.Fatalf("expected missing p2 secret, got code=%d body=%s", getP2Secret.Code, getP2Secret.Body.String())
	}
}

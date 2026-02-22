package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConnectivityAndSuggest(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"part_number\":\"P-1\",\"brand\":\"AFX\",\"title\":\"AFX P-1\",\"confidence\":0.91}"}}]}`))
	}))
	defer srv.Close()

	svc := NewService(Config{BaseURL: srv.URL})
	if err := svc.TestConnectivity(context.Background(), "sk-test"); err != nil {
		t.Fatalf("TestConnectivity() error = %v", err)
	}
	out, err := svc.SuggestFromTitle(context.Background(), "sk-test", "AFX P-1 slot car")
	if err != nil {
		t.Fatalf("SuggestFromTitle() error = %v", err)
	}
	if out.PartNumber == "" || out.Confidence <= 0 {
		t.Fatalf("unexpected AI output: %+v", out)
	}
	if !out.RequiresConfirmation {
		t.Fatal("expected requires confirmation true")
	}
}

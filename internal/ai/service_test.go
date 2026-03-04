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

func TestProposeOperationStructuredPayload(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o-mini"}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"action\":\"create_wishlist_entry\",\"target\":\"wishlist\",\"payload\":{\"part_number\":\"CP-233\",\"title\":\"Copilot Wishlist\"},\"confidence\":0.88}"}}]}`))
	}))
	defer srv.Close()

	svc := NewService(Config{BaseURL: srv.URL})
	proposal, err := svc.ProposeOperation(context.Background(), "sk-test", "Create wishlist entry for CP-233")
	if err != nil {
		t.Fatalf("ProposeOperation() error = %v", err)
	}
	if proposal.Action != "create_wishlist_entry" {
		t.Fatalf("expected action create_wishlist_entry, got %q", proposal.Action)
	}
	if proposal.Target != "wishlist" {
		t.Fatalf("expected target wishlist, got %q", proposal.Target)
	}
	if proposal.Payload["part_number"] != "CP-233" {
		t.Fatalf("expected payload part_number CP-233, got %+v", proposal.Payload)
	}
	if !proposal.RequiresConfirmation {
		t.Fatal("expected requires confirmation true")
	}
}

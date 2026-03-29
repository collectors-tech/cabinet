package app

import (
  "encoding/json"
  "net/http"
)

type futureHookDescriptor struct {
  ID          string `json:"id"`
  Category    string `json:"category"`
  Status      string `json:"status"`
  Active      bool   `json:"active"`
  Operative   bool   `json:"operative"`
  Description string `json:"description"`
}

func futureHookRegistry() []futureHookDescriptor {
  return []futureHookDescriptor{
    {
      ID:          "ai-provider-anthropic",
      Category:    "ai_provider",
      Status:      "disabled",
      Active:      false,
      Operative:   false,
      Description: "Reserved scaffold for future Anthropic provider integration.",
    },
    {
      ID:          "scanner-provider-google-vision",
      Category:    "scanner_provider",
      Status:      "disabled",
      Active:      false,
      Operative:   false,
      Description: "Reserved scaffold for future Google Vision scanner integration.",
    },
    {
      ID:          "share-compare",
      Category:    "workflow_hook",
      Status:      "disabled",
      Active:      false,
      Operative:   false,
      Description: "Reserved scaffold for future share/compare flows.",
    },
    {
      ID:          "for-sale-flag",
      Category:    "inventory_hook",
      Status:      "disabled",
      Active:      false,
      Operative:   false,
      Description: "Reserved scaffold for future for-sale item state.",
    },
    {
      ID:          "structured-offers",
      Category:    "inventory_hook",
      Status:      "disabled",
      Active:      false,
      Operative:   false,
      Description: "Reserved scaffold for future structured offers support.",
    },
  }
}

func registerFutureHookRoutes(mux *http.ServeMux) {
  mux.HandleFunc("/api/future-hooks", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    if r.Method != http.MethodGet {
      http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
      return
    }
    _ = json.NewEncoder(w).Encode(map[string]any{
      "hooks": futureHookRegistry(),
    })
  })

  mux.HandleFunc("/api/future-hooks/invoke", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    if r.Method != http.MethodPost {
      http.Error(w, `{"error":"method_not_allowed"}`, http.StatusMethodNotAllowed)
      return
    }

    var req struct {
      HookID string `json:"hook_id"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
      http.Error(w, `{"error":"invalid_json"}`, http.StatusBadRequest)
      return
    }

    w.WriteHeader(http.StatusConflict)
    _ = json.NewEncoder(w).Encode(map[string]any{
      "error":   "hook_not_active",
      "hook_id": req.HookID,
      "status":  "disabled",
    })
  })
}

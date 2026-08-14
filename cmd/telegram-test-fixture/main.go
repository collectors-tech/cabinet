// Command telegram-test-fixture is a deterministic, loopback-only Telegram Bot
// API double used by the governed release-candidate Cypress pack. It never
// contacts Telegram and must not be used as live-provider evidence.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
)

type fixture struct {
	mu      sync.Mutex
	updates []map[string]any
	counts  map[string]int
	hold    bool
	webhook string
}

func newFixture() *fixture {
	return &fixture{counts: map[string]int{}, hold: true, webhook: "https://old.example/telegram"}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func readBody(r *http.Request) map[string]any {
	var value map[string]any
	_ = json.NewDecoder(r.Body).Decode(&value)
	return value
}

func int64Value(value any) int64 {
	switch number := value.(type) {
	case float64:
		return int64(number)
	case int64:
		return number
	case int:
		return int64(number)
	default:
		return 0
	}
}

func (f *fixture) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	switch r.URL.Path {
	case "/control/status":
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "counts": f.counts, "queued": len(f.updates), "hold": f.hold})
		return
	case "/control/reset":
		f.updates = nil
		f.counts = map[string]int{}
		f.hold = true
		f.webhook = "https://old.example/telegram"
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	case "/control/hold":
		f.hold = true
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	case "/control/release":
		f.hold = false
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	case "/control/updates":
		body := readBody(r)
		updateID := int64Value(body["update_id"])
		peerID := int64Value(body["sender_id"])
		chatID := int64Value(body["chat_id"])
		chatType, _ := body["chat_type"].(string)
		var update map[string]any
		if callbackData, _ := body["callback_data"].(string); callbackData != "" {
			update = map[string]any{
				"update_id": updateID,
				"callback_query": map[string]any{
					"id": body["callback_query_id"], "from": map[string]any{"id": peerID},
					"message": map[string]any{"message_id": int64Value(body["message_id"]), "chat": map[string]any{"id": chatID, "type": chatType}},
					"data":    callbackData,
				},
			}
		} else {
			update = map[string]any{
				"update_id": updateID,
				"message":   map[string]any{"message_id": updateID, "from": map[string]any{"id": peerID}, "chat": map[string]any{"id": chatID, "type": chatType}, "text": body["text"]},
			}
		}
		f.updates = append(f.updates, update)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}

	method := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	f.counts[method]++
	switch method {
	case "getMe":
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "result": map[string]any{"id": 2086, "is_bot": true, "first_name": "Cabinet", "username": "cabinet_fixture_bot"}})
	case "getWebhookInfo":
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "result": map[string]any{"url": f.webhook, "pending_update_count": len(f.updates)}})
	case "deleteWebhook":
		f.webhook = ""
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "result": true})
	case "getUpdates":
		offset := int64Value(readBody(r)["offset"])
		result := []map[string]any{}
		if !f.hold {
			for _, update := range f.updates {
				if int64Value(update["update_id"]) >= offset {
					result = append(result, update)
				}
			}
			sort.Slice(result, func(i, j int) bool { return int64Value(result[i]["update_id"]) < int64Value(result[j]["update_id"]) })
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "result": result})
	default:
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "result": true})
	}
}

func main() {
	listen := flag.String("listen", "127.0.0.1:17994", "loopback listen address")
	flag.Parse()
	if !strings.HasPrefix(*listen, "127.0.0.1:") {
		log.Fatal("telegram test fixture must bind to 127.0.0.1")
	}
	log.Printf("controlled Telegram fixture listening on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, newFixture()))
}

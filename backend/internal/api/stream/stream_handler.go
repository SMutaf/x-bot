package stream

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func StreamHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	viewID := strings.TrimSpace(r.URL.Query().Get("view"))
	events := SubscribePublished()
	defer UnsubscribePublished(events)

	sendEvent(w, flusher, "connected", map[string]any{
		"status": "ok",
		"view":   viewID,
		"time":   time.Now().Format(time.RFC3339),
	})

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case item := <-events:
			if !matchesView(item, viewID) {
				continue
			}
			sendEvent(w, flusher, "news.published", item)
		case <-heartbeat.C:
			sendEvent(w, flusher, "heartbeat", map[string]any{
				"time": time.Now().Format(time.RFC3339),
			})
		}
	}
}

func sendEvent(w http.ResponseWriter, flusher http.Flusher, eventName string, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "event: %s\n", eventName)
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

func matchesView(item PublishedItem, viewID string) bool {
	switch viewID {
	case "turkey-critical":
		switch item.Category {
		case "BREAKING":
			return item.Virality >= 35
		case "GENERAL":
			return item.Virality >= 25
		case "ECONOMY":
			return item.Virality >= 24
		default:
			return false
		}

	case "global-high-impact":
		switch item.Category {
		case "BREAKING":
			return item.Virality >= 38
		case "GENERAL":
			return item.Virality >= 35
		case "ECONOMY":
			return item.Virality >= 30
		default:
			return false
		}

	case "economy-markets":
		switch item.Category {
		case "ECONOMY":
			return true
		case "BREAKING":
			return item.Virality >= 40
		case "GENERAL":
			return item.Virality >= 38
		default:
			return false
		}

	case "tech-watch":
		return item.Category == "TECH"

	default:
		return true
	}
}

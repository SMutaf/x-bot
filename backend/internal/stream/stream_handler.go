package stream

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type PublishedItem struct {
	Time         string `json:"time"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	Source       string `json:"source"`
	Link         string `json:"link"`
	Virality     int    `json:"virality"`
	ClusterCount int    `json:"clusterCount"`
}

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
	filePath := "data/published.jsonl"

	sendEvent(w, flusher, "connected", map[string]any{
		"status": "ok",
		"view":   viewID,
		"time":   time.Now().Format(time.RFC3339),
	})

	var lastLineCount int
	lastHeartbeat := time.Now()

	for {
		select {
		case <-r.Context().Done():
			fmt.Println("Client disconnected")
			return
		default:
			file, err := os.Open(filePath)
			if err == nil {
				scanner := bufio.NewScanner(file)
				currentLine := 0
				for scanner.Scan() {
					currentLine++
					if currentLine <= lastLineCount {
						continue
					}
					var item PublishedItem
					if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
						continue
					}
					if !matchesView(item, viewID) {
						continue
					}
					sendEvent(w, flusher, "news.published", item)
				}
				lastLineCount = currentLine
				_ = file.Close()
			}

			if time.Since(lastHeartbeat) >= 15*time.Second {
				sendEvent(w, flusher, "heartbeat", map[string]any{
					"time": time.Now().Format(time.RFC3339),
				})
				lastHeartbeat = time.Now()
			}
			time.Sleep(2 * time.Second)
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

// matchesView — hangi haberin hangi view'da görüneceğini belirler.
func matchesView(item PublishedItem, viewID string) bool {
	/*	switch viewID {

		case "turkey-critical":
			// Türkiye'yi etkileyen her kategoriden önemli haber
			// BREAKING: tüm kritik breaking haberler
			// GENERAL: Türkiye'ye doğrudan dokunan haberler
			// ECONOMY: Türkiye ekonomisini etkileyen haberler
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
			// Küresel ölçekte yüksek etkili haberler
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
			// Ekonomi ve piyasa haberleri
			switch item.Category {
			case "ECONOMY":
				return true // tüm economy haberleri
			case "BREAKING":
				return item.Virality >= 40 // çok güçlü breaking ekonomi etkili haberler
			case "GENERAL":
				return item.Virality >= 38
			default:
				return false
			}

		case "tech-watch":
			return item.Category == "TECH"

		default:
			return true
		}*/
	return true
}

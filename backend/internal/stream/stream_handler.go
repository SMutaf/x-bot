package stream

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type PublishedItem struct {
	Time         string `json:"time"`
	Title        string `json:"title"`
	Category     string `json:"category"`
	Source       string `json:"source"`
	Link         string `json:"link"`
	Virality     int    `json:"virality"`
	ClusterCount int    `json:"clusterCount"`
}

func StreamHandler(w http.ResponseWriter, r *http.Request) {
	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	filePath := "data/published.jsonl"

	var lastLineCount int

	for {
		select {
		case <-r.Context().Done():
			fmt.Println("Client disconnected")
			return
		default:
			file, err := os.Open(filePath)
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}

			scanner := bufio.NewScanner(file)
			var lines []string

			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			file.Close()

			// yeni satırlar varsa gönder
			if len(lines) > lastLineCount {
				newLines := lines[lastLineCount:]

				for _, line := range newLines {
					var item PublishedItem
					err := json.Unmarshal([]byte(line), &item)
					if err != nil {
						continue
					}

					payload, _ := json.Marshal(item)

					fmt.Fprintf(w, "event: news.published\n")
					fmt.Fprintf(w, "data: %s\n\n", payload)
				}

				lastLineCount = len(lines)
				flusher.Flush()
			}

			time.Sleep(2 * time.Second)
		}
	}
}

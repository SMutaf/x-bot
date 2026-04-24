package dashboardapi

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/SMutaf/twitter-bot/backend/internal/monitoring"
	"github.com/SMutaf/twitter-bot/backend/internal/sourcehealth"
)

type Handler struct {
	Monitoring    *monitoring.Manager
	HealthManager *sourcehealth.Manager
	Status        *StatusProvider
}

func NewHandler(m *monitoring.Manager, h *sourcehealth.Manager, s *StatusProvider) *Handler {
	return &Handler{
		Monitoring:    m,
		HealthManager: h,
		Status:        s,
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/feed", h.handleFeedSnapshot)

	mux.HandleFunc("/api/dashboard/summary", h.handleSummary)
	mux.HandleFunc("/api/dashboard/status", h.handleStatus)
	mux.HandleFunc("/api/dashboard/published", h.handlePublished)
	mux.HandleFunc("/api/dashboard/rejected", h.handleRejected)
	mux.HandleFunc("/api/dashboard/sources", h.handleSources)
	mux.HandleFunc("/api/dashboard/health-events", h.handleHealthEvents)

	mux.HandleFunc("/api/dashboard/download/published", h.handleDownloadPublished)
	mux.HandleFunc("/api/dashboard/download/rejected", h.handleDownloadRejected)
	mux.HandleFunc("/api/dashboard/download/source-health", h.handleDownloadSourceHealth)
}

func (h *Handler) handleFeedSnapshot(w http.ResponseWriter, r *http.Request) {
	viewID := strings.TrimSpace(r.URL.Query().Get("view"))
	limit := parseLimit(r.URL.Query().Get("limit"), 50)

	items := h.Monitoring.GetPublished()
	filtered := filterPublishedByView(items, viewID)

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Time.After(filtered[j].Time)
	})

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	writeJSON(w, filtered)
}

func (h *Handler) handleSummary(w http.ResponseWriter, r *http.Request) {
	snapshot := h.HealthManager.Snapshot()
	summary := h.Monitoring.BuildSummary(snapshot)
	writeJSON(w, summary)
}

func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, h.Status.Build())
}

func (h *Handler) handlePublished(w http.ResponseWriter, r *http.Request) {
	items := h.Monitoring.GetPublished()
	viewID := strings.TrimSpace(r.URL.Query().Get("view"))
	limit := parseLimit(r.URL.Query().Get("limit"), len(items))

	if viewID != "" {
		items = filterPublishedByView(items, viewID)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Time.After(items[j].Time)
	})

	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}

	writeJSON(w, items)
}

func (h *Handler) handleRejected(w http.ResponseWriter, r *http.Request) {
	items := h.Monitoring.GetRejected()
	limit := parseLimit(r.URL.Query().Get("limit"), len(items))

	sort.Slice(items, func(i, j int) bool {
		return items[i].Time.After(items[j].Time)
	})

	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}

	writeJSON(w, items)
}

func (h *Handler) handleSources(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, h.HealthManager.Snapshot())
}

func (h *Handler) handleHealthEvents(w http.ResponseWriter, r *http.Request) {
	items := h.Monitoring.GetSourceHealth()
	limit := parseLimit(r.URL.Query().Get("limit"), len(items))

	sort.Slice(items, func(i, j int) bool {
		return items[i].Time.After(items[j].Time)
	})

	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}

	writeJSON(w, items)
}

func (h *Handler) handleDownloadPublished(w http.ResponseWriter, r *http.Request) {
	serveDownload(w, r, h.Monitoring.PublishedPath(), "published.jsonl")
}

func (h *Handler) handleDownloadRejected(w http.ResponseWriter, r *http.Request) {
	serveDownload(w, r, h.Monitoring.RejectedPath(), "rejected.jsonl")
}

func (h *Handler) handleDownloadSourceHealth(w http.ResponseWriter, r *http.Request) {
	serveDownload(w, r, h.Monitoring.HealthPath(), "source_health.jsonl")
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func serveDownload(w http.ResponseWriter, r *http.Request, path string, filename string) {
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	http.ServeFile(w, r, filepath.Clean(path))
}

func parseLimit(raw string, fallback int) int {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}

	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return fallback
	}

	if n > 500 {
		return 500
	}

	return n
}

func filterPublishedByView(items []monitoring.PublishedNewsEvent, viewID string) []monitoring.PublishedNewsEvent {
	if viewID == "" {
		return items
	}

	out := make([]monitoring.PublishedNewsEvent, 0, len(items))

	for _, item := range items {
		if matchesView(item, viewID) {
			out = append(out, item)
		}
	}

	return out
}

func matchesView(item monitoring.PublishedNewsEvent, viewID string) bool {
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

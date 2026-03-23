package dashboardapi

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/monitoring"
	"github.com/SMutaf/twitter-bot/backend/internal/sourcehealth"
)

type Handler struct {
	Monitoring    *monitoring.Manager
	HealthManager *sourcehealth.Manager
}

func NewHandler(m *monitoring.Manager, h *sourcehealth.Manager) *Handler {
	return &Handler{
		Monitoring:    m,
		HealthManager: h,
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/dashboard/summary", h.handleSummary)
	mux.HandleFunc("/api/dashboard/published", h.handlePublished)
	mux.HandleFunc("/api/dashboard/rejected", h.handleRejected)
	mux.HandleFunc("/api/dashboard/sources", h.handleSources)
	mux.HandleFunc("/api/dashboard/health-events", h.handleHealthEvents)

	mux.HandleFunc("/api/dashboard/download/published", h.handleDownloadPublished)
	mux.HandleFunc("/api/dashboard/download/rejected", h.handleDownloadRejected)
	mux.HandleFunc("/api/dashboard/download/source-health", h.handleDownloadSourceHealth)
}

func (h *Handler) handleSummary(w http.ResponseWriter, r *http.Request) {
	snapshot := h.HealthManager.Snapshot()
	summary := h.Monitoring.BuildSummary(snapshot)
	writeJSON(w, summary)
}

func (h *Handler) handlePublished(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, h.Monitoring.GetPublished())
}

func (h *Handler) handleRejected(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, h.Monitoring.GetRejected())
}

func (h *Handler) handleSources(w http.ResponseWriter, r *http.Request) {
	type SourceView struct {
		SourceName       string `json:"sourceName"`
		URL              string `json:"url"`
		Category         string `json:"category"`
		State            string `json:"state"`
		ConsecutiveFails int    `json:"consecutiveFails"`
		LastErrorType    string `json:"lastErrorType"`
		LastErrorMessage string `json:"lastErrorMessage"`
		DisabledUntil    string `json:"disabledUntil"`
		LastSuccessAt    string `json:"lastSuccessAt"`
	}

	now := time.Now()
	snapshot := h.HealthManager.Snapshot()
	out := make([]SourceView, 0, len(snapshot))

	for _, s := range snapshot {
		state := "healthy"
		if s.IsDisabled(now) {
			state = "disabled"
		} else if s.ConsecutiveFails > 0 {
			state = "degraded"
		}

		out = append(out, SourceView{
			SourceName:       s.SourceName,
			URL:              s.URL,
			Category:         string(s.Category),
			State:            state,
			ConsecutiveFails: s.ConsecutiveFails,
			LastErrorType:    s.LastErrorType,
			LastErrorMessage: s.LastErrorMessage,
			DisabledUntil:    formatTime(s.DisabledUntil),
			LastSuccessAt:    formatTime(s.LastSuccessAt),
		})
	}

	writeJSON(w, out)
}

func (h *Handler) handleHealthEvents(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, h.Monitoring.GetSourceHealthEvents())
}

func (h *Handler) handleDownloadPublished(w http.ResponseWriter, r *http.Request) {
	serveJSONLFile(w, r, h.Monitoring.GetPublishedPath(), "published.jsonl")
}

func (h *Handler) handleDownloadRejected(w http.ResponseWriter, r *http.Request) {
	serveJSONLFile(w, r, h.Monitoring.GetRejectedPath(), "rejected.jsonl")
}

func (h *Handler) handleDownloadSourceHealth(w http.ResponseWriter, r *http.Request) {
	serveJSONLFile(w, r, h.Monitoring.GetHealthPath(), "source_health.jsonl")
}

func serveJSONLFile(w http.ResponseWriter, r *http.Request, path, downloadName string) {
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Content-Disposition", `attachment; filename="`+downloadName+`"`)
	http.ServeFile(w, r, filepath.Clean(path))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

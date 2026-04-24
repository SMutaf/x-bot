package dashboardapi

import (
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/monitoring"
)

type SystemStatus struct {
	PublishedCount              int    `json:"publishedCount"`
	RejectedCount               int    `json:"rejectedCount"`
	HealthyRSSSources           int    `json:"healthyRssSources"`
	UnhealthyRSSSources         int    `json:"unhealthyRssSources"`
	DisabledRSSSources          int    `json:"disabledRssSources"`
	TrackedRSSSources           int    `json:"trackedRssSources"`
	RedisConnected              bool   `json:"redisConnected"`
	RedisState                  string `json:"redisState"`
	PythonConnected             bool   `json:"pythonConnected"`
	PythonState                 string `json:"pythonState"`
	PythonInFlight              int    `json:"pythonInFlight"`
	RedisError                  string `json:"redisError"`
	PythonError                 string `json:"pythonError"`
	RedisConsecutiveFailures    int    `json:"redisConsecutiveFailures"`
	PythonConsecutiveFailures   int    `json:"pythonConsecutiveFailures"`
	RedisLastSuccessfulCheckAt  string `json:"redisLastSuccessfulCheckAt"`
	RedisLastFailedCheckAt      string `json:"redisLastFailedCheckAt"`
	RedisLastStateChangeAt      string `json:"redisLastStateChangeAt"`
	PythonLastSuccessfulCheckAt string `json:"pythonLastSuccessfulCheckAt"`
	PythonLastFailedCheckAt     string `json:"pythonLastFailedCheckAt"`
	PythonLastStateChangeAt     string `json:"pythonLastStateChangeAt"`
	LastPublishedAt             string `json:"lastPublishedAt"`
	LastRejectedAt              string `json:"lastRejectedAt"`
}

type StatusProvider struct {
	Monitoring *monitoring.Manager
	Services   *ServiceStatusManager
}

func (p *StatusProvider) Build() SystemStatus {
	summary := p.Monitoring.BuildSummary()

	status := SystemStatus{
		PublishedCount:      summary.PublishedCount,
		RejectedCount:       summary.RejectedCount,
		HealthyRSSSources:   summary.HealthySources,
		UnhealthyRSSSources: summary.DegradedSources + summary.DisabledSources,
		DisabledRSSSources:  summary.DisabledSources,
		TrackedRSSSources:   summary.TrackedSourceSize,
		LastPublishedAt:     latestPublishedAt(p.Monitoring),
		LastRejectedAt:      latestRejectedAt(p.Monitoring),
	}

	if p.Services != nil {
		serviceStatus := p.Services.Snapshot()
		status.RedisConnected = serviceStatus.RedisConnected
		status.RedisState = serviceStatus.RedisState
		status.RedisError = serviceStatus.RedisError
		status.RedisConsecutiveFailures = serviceStatus.RedisConsecutiveFailures
		status.RedisLastSuccessfulCheckAt = serviceStatus.RedisLastSuccessfulCheckAt
		status.RedisLastFailedCheckAt = serviceStatus.RedisLastFailedCheckAt
		status.RedisLastStateChangeAt = serviceStatus.RedisLastStateChangeAt
		status.PythonConnected = serviceStatus.PythonConnected
		status.PythonState = serviceStatus.PythonState
		status.PythonInFlight = serviceStatus.PythonInFlight
		status.PythonError = serviceStatus.PythonError
		status.PythonConsecutiveFailures = serviceStatus.PythonConsecutiveFailures
		status.PythonLastSuccessfulCheckAt = serviceStatus.PythonLastSuccessfulCheckAt
		status.PythonLastFailedCheckAt = serviceStatus.PythonLastFailedCheckAt
		status.PythonLastStateChangeAt = serviceStatus.PythonLastStateChangeAt
	}

	return status
}

func latestPublishedAt(m *monitoring.Manager) string {
	items := m.GetPublished()
	if len(items) == 0 {
		return ""
	}

	latest := items[0].Time
	for _, item := range items[1:] {
		if item.Time.After(latest) {
			latest = item.Time
		}
	}

	return latest.Format(time.RFC3339)
}

func latestRejectedAt(m *monitoring.Manager) string {
	items := m.GetRejected()
	if len(items) == 0 {
		return ""
	}

	latest := items[0].Time
	for _, item := range items[1:] {
		if item.Time.After(latest) {
			latest = item.Time
		}
	}

	return latest.Format(time.RFC3339)
}

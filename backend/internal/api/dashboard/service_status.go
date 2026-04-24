package dashboardapi

import (
	"sync"
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/infra/ai"
	"github.com/SMutaf/twitter-bot/backend/internal/ingestion/dedup"
)

const (
	serviceStateOnline   = "online"
	serviceStateBusy     = "busy"
	serviceStateDegraded = "degraded"
	serviceStateOffline  = "offline"
)

type ServiceStatusSnapshot struct {
	RedisConnected             bool
	RedisState                 string
	RedisError                 string
	RedisConsecutiveFailures   int
	RedisLastSuccessfulCheckAt string
	RedisLastFailedCheckAt     string
	RedisLastStateChangeAt     string

	PythonConnected             bool
	PythonState                 string
	PythonInFlight              int
	PythonError                 string
	PythonConsecutiveFailures   int
	PythonLastSuccessfulCheckAt string
	PythonLastFailedCheckAt     string
	PythonLastStateChangeAt     string
}

type ServiceStatusManager struct {
	redis *dedup.Deduplicator
	ai    *ai.Client

	mu       sync.RWMutex
	snapshot ServiceStatusSnapshot
}

func NewServiceStatusManager(redis *dedup.Deduplicator, aiClient *ai.Client) *ServiceStatusManager {
	manager := &ServiceStatusManager{
		redis: redis,
		ai:    aiClient,
		snapshot: ServiceStatusSnapshot{
			RedisState:  serviceStateOffline,
			PythonState: serviceStateOffline,
		},
	}

	manager.refresh()
	return manager
}

func (m *ServiceStatusManager) Start(interval time.Duration) {
	if interval <= 0 {
		interval = 10 * time.Second
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			m.refresh()
		}
	}()
}

func (m *ServiceStatusManager) Snapshot() ServiceStatusSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.snapshot
}

func (m *ServiceStatusManager) refresh() {
	m.mu.RLock()
	prev := m.snapshot
	m.mu.RUnlock()

	now := time.Now().Format(time.RFC3339)
	next := ServiceStatusSnapshot{
		RedisState:                  serviceStateOffline,
		PythonState:                 serviceStateOffline,
		RedisConsecutiveFailures:    prev.RedisConsecutiveFailures,
		PythonConsecutiveFailures:   prev.PythonConsecutiveFailures,
		RedisLastSuccessfulCheckAt:  prev.RedisLastSuccessfulCheckAt,
		RedisLastFailedCheckAt:      prev.RedisLastFailedCheckAt,
		RedisLastStateChangeAt:      prev.RedisLastStateChangeAt,
		PythonLastSuccessfulCheckAt: prev.PythonLastSuccessfulCheckAt,
		PythonLastFailedCheckAt:     prev.PythonLastFailedCheckAt,
		PythonLastStateChangeAt:     prev.PythonLastStateChangeAt,
	}

	if m.redis != nil {
		if err := m.redis.HealthCheck(); err != nil {
			next.RedisError = err.Error()
			next.RedisConsecutiveFailures++
			next.RedisLastFailedCheckAt = now
			if next.RedisConsecutiveFailures >= 3 {
				next.RedisState = serviceStateOffline
			} else {
				next.RedisState = serviceStateDegraded
			}
		} else {
			next.RedisConnected = true
			next.RedisState = serviceStateOnline
			next.RedisConsecutiveFailures = 0
			next.RedisLastSuccessfulCheckAt = now
		}
	}

	if m.ai != nil {
		next.PythonInFlight = m.ai.InFlight()
		if next.PythonInFlight > 0 {
			next.PythonConnected = true
			next.PythonState = serviceStateBusy
			next.PythonConsecutiveFailures = 0
			next.PythonLastSuccessfulCheckAt = now
		} else if err := m.ai.HealthCheck(); err != nil {
			next.PythonError = err.Error()
			next.PythonConsecutiveFailures++
			next.PythonLastFailedCheckAt = now
			if next.PythonConsecutiveFailures >= 3 {
				next.PythonState = serviceStateOffline
			} else {
				next.PythonState = serviceStateDegraded
			}
		} else {
			next.PythonConnected = true
			next.PythonState = serviceStateOnline
			next.PythonConsecutiveFailures = 0
			next.PythonLastSuccessfulCheckAt = now
		}
	}

	if next.RedisState != prev.RedisState {
		next.RedisLastStateChangeAt = now
	}
	if next.PythonState != prev.PythonState {
		next.PythonLastStateChangeAt = now
	}

	m.mu.Lock()
	m.snapshot = next
	m.mu.Unlock()
}

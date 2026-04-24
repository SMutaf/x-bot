package dashboardapi

import (
	"sync"
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/ai"
	"github.com/SMutaf/twitter-bot/backend/internal/dedup"
)

const (
	serviceStateOnline  = "online"
	serviceStateBusy    = "busy"
	serviceStateOffline = "offline"
)

type ServiceStatusSnapshot struct {
	RedisConnected bool
	RedisState     string
	RedisError     string

	PythonConnected bool
	PythonState     string
	PythonInFlight  int
	PythonError     string
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
	next := ServiceStatusSnapshot{
		RedisState:  serviceStateOffline,
		PythonState: serviceStateOffline,
	}

	if m.redis != nil {
		if err := m.redis.HealthCheck(); err != nil {
			next.RedisError = err.Error()
		} else {
			next.RedisConnected = true
			next.RedisState = serviceStateOnline
		}
	}

	if m.ai != nil {
		next.PythonInFlight = m.ai.InFlight()
		if next.PythonInFlight > 0 {
			next.PythonConnected = true
			next.PythonState = serviceStateBusy
		} else if err := m.ai.HealthCheck(); err != nil {
			next.PythonError = err.Error()
		} else {
			next.PythonConnected = true
			next.PythonState = serviceStateOnline
		}
	}

	m.mu.Lock()
	m.snapshot = next
	m.mu.Unlock()
}

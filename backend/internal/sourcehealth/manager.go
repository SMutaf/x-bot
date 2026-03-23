package sourcehealth

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/SMutaf/twitter-bot/backend/config"
)

type Manager struct {
	mu     sync.RWMutex
	states map[string]*Status
}

func NewManager() *Manager {
	return &Manager{
		states: make(map[string]*Status),
	}
}

func (m *Manager) ShouldSkip(source config.RSSSource, sourceName string) (bool, Status) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, ok := m.states[source.URL]
	if !ok {
		return false, Status{}
	}

	now := time.Now()
	if state.IsDisabled(now) {
		return true, *state
	}

	return false, *state
}

func (m *Manager) RecordSuccess(source config.RSSSource, sourceName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state := m.getOrCreateLocked(source, sourceName)
	state.ConsecutiveFails = 0
	state.LastErrorType = ""
	state.LastErrorMessage = ""
	state.LastErrorAt = time.Time{}
	state.DisabledUntil = time.Time{}
	state.LastSuccessAt = time.Now()
}

func (m *Manager) RecordFailure(source config.RSSSource, sourceName, errType, errMsg string) Status {
	m.mu.Lock()
	defer m.mu.Unlock()

	state := m.getOrCreateLocked(source, sourceName)
	state.ConsecutiveFails++
	state.LastErrorType = errType
	state.LastErrorMessage = errMsg
	state.LastErrorAt = time.Now()
	state.DisabledUntil = time.Now().Add(m.cooldownFor(state.ConsecutiveFails, errType))

	return *state
}

func (m *Manager) Snapshot() []Status {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]Status, 0, len(m.states))
	for _, state := range m.states {
		out = append(out, *state)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Category != out[j].Category {
			return out[i].Category < out[j].Category
		}
		return out[i].SourceName < out[j].SourceName
	})

	return out
}

func (m *Manager) getOrCreateLocked(source config.RSSSource, sourceName string) *Status {
	if existing, ok := m.states[source.URL]; ok {
		return existing
	}

	state := &Status{
		SourceName: sourceName,
		URL:        source.URL,
		Category:   source.Category,
	}

	m.states[source.URL] = state
	return state
}

func (m *Manager) cooldownFor(consecutiveFails int, errType string) time.Duration {
	base := 0 * time.Minute

	switch {
	case consecutiveFails <= 1:
		base = 0
	case consecutiveFails == 2:
		base = 2 * time.Minute
	case consecutiveFails == 3:
		base = 5 * time.Minute
	default:
		base = 10 * time.Minute
	}

	switch errType {
	case "INVALID_UTF8":
		if base < 10*time.Minute {
			base = 10 * time.Minute
		}
	case "DNS_ERROR":
		if base < 5*time.Minute {
			base = 5 * time.Minute
		}
	}

	return base
}

func FormatSnapshot(snapshot []Status) string {
	if len(snapshot) == 0 {
		return "[SOURCE HEALTH] henüz kayıt yok"
	}

	now := time.Now()
	result := "[SOURCE HEALTH SNAPSHOT]\n"

	for _, s := range snapshot {
		healthState := "healthy"
		if s.IsDisabled(now) {
			healthState = fmt.Sprintf("disabled until %s", s.DisabledUntil.Format("15:04:05"))
		} else if s.ConsecutiveFails > 0 {
			healthState = "degraded"
		}

		line := fmt.Sprintf(
			"- %s | %s | state=%s | fails=%d",
			s.SourceName,
			s.Category,
			healthState,
			s.ConsecutiveFails,
		)

		if s.LastErrorType != "" {
			line += fmt.Sprintf(" | lastError=%s", s.LastErrorType)
		}
		if !s.LastSuccessAt.IsZero() {
			line += fmt.Sprintf(" | lastSuccess=%s", s.LastSuccessAt.Format("15:04:05"))
		}

		result += line + "\n"
	}

	return result
}

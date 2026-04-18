package monitoring

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/sourcehealth"
)

type Manager struct {
	mu sync.RWMutex

	dataDir string

	publishedPath string
	rejectedPath  string
	healthPath    string

	published []PublishedNewsEvent
	rejected  []RejectedNewsEvent
	health    []SourceHealthEvent

	maxInMemory int
}

func NewManager(dataDir string) (*Manager, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}

	m := &Manager{
		dataDir:       dataDir,
		publishedPath: filepath.Join(dataDir, "published.jsonl"),
		rejectedPath:  filepath.Join(dataDir, "rejected.jsonl"),
		healthPath:    filepath.Join(dataDir, "source_health.jsonl"),
		maxInMemory:   300,
		published:     make([]PublishedNewsEvent, 0),
		rejected:      make([]RejectedNewsEvent, 0),
		health:        make([]SourceHealthEvent, 0),
	}

	return m, nil
}

func (m *Manager) RecordPublished(event PublishedNewsEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.published = append(m.published, event)
	m.trimPublished()
	_ = appendJSONL(m.publishedPath, event)
}

func (m *Manager) RecordRejected(event RejectedNewsEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.rejected = append(m.rejected, event)
	m.trimRejected()
	_ = appendJSONL(m.rejectedPath, event)
}

func (m *Manager) RecordSourceHealth(event SourceHealthEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.health = append(m.health, event)
	m.trimHealth()
	_ = appendJSONL(m.healthPath, event)
}

func (m *Manager) GetPublished() []PublishedNewsEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return cloneSlice(m.published)
}

func (m *Manager) GetRejected() []RejectedNewsEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return cloneSlice(m.rejected)
}

func (m *Manager) GetSourceHealthEvents() []SourceHealthEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return cloneSlice(m.health)
}

func (m *Manager) GetPublishedPath() string {
	return m.publishedPath
}

func (m *Manager) GetRejectedPath() string {
	return m.rejectedPath
}

func (m *Manager) GetHealthPath() string {
	return m.healthPath
}

func (m *Manager) PublishedPath() string {
	return m.publishedPath
}

func (m *Manager) RejectedPath() string {
	return m.rejectedPath
}

func (m *Manager) HealthPath() string {
	return m.healthPath
}

func (m *Manager) GetSourceHealth() string {
	return m.GetSourceHealth()
}

func (m *Manager) BuildSummary(snapshot []sourcehealth.Status) Summary {
	healthy := 0
	disabled := 0
	degraded := 0
	now := time.Now()

	for _, s := range snapshot {
		switch {
		case s.IsDisabled(now):
			disabled++
		case s.ConsecutiveFails > 0:
			degraded++
		default:
			healthy++
		}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	return Summary{
		PublishedCount:    len(m.published),
		RejectedCount:     len(m.rejected),
		HealthySources:    healthy,
		DisabledSources:   disabled,
		DegradedSources:   degraded,
		TrackedSourceSize: len(snapshot),
	}
}

func (m *Manager) trimPublished() {
	if len(m.published) > m.maxInMemory {
		m.published = m.published[len(m.published)-m.maxInMemory:]
	}
}

func (m *Manager) trimRejected() {
	if len(m.rejected) > m.maxInMemory {
		m.rejected = m.rejected[len(m.rejected)-m.maxInMemory:]
	}
}

func (m *Manager) trimHealth() {
	if len(m.health) > m.maxInMemory {
		m.health = m.health[len(m.health)-m.maxInMemory:]
	}
}

func appendJSONL[T any](path string, value T) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(value)
}

func cloneSlice[T any](in []T) []T {
	out := make([]T, len(in))
	copy(out, in)
	return out
}

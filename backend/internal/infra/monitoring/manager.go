package monitoring

import (
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/api/stream"
	"github.com/SMutaf/twitter-bot/backend/internal/ingestion/dedup"
)

type Manager struct {
	store *RedisStore
}

func NewManager(cache *dedup.Deduplicator) (*Manager, error) {
	return &Manager{
		store: NewRedisStore(cache, retentionDays),
	}, nil
}

func (m *Manager) RecordPublished(event PublishedNewsEvent) {
	if m.store != nil {
		_ = m.store.SavePublished(event)
	}

	stream.PublishPublished(stream.PublishedItem{
		Time:         event.Time.Format(time.RFC3339),
		Title:        event.Title,
		Description:  event.Description,
		DescriptionTR: event.DescriptionTR,
		Hook:         event.Hook,
		Summary:      event.Summary,
		Importance:   event.Importance,
		Sentiment:    event.Sentiment,
		Category:     event.Category,
		Source:       event.Source,
		Link:         event.Link,
		Virality:     event.Virality,
		ClusterCount: event.ClusterCount,
	})
}

func (m *Manager) RecordRejected(event RejectedNewsEvent) {
	if m.store != nil {
		_ = m.store.SaveRejected(event)
	}
}

func (m *Manager) RecordSourceHealth(event SourceHealthEvent) {
	if m.store != nil {
		_ = m.store.SaveSourceHealth(event)
	}
}

func (m *Manager) GetPublished() []PublishedNewsEvent {
	if m.store == nil {
		return nil
	}

	items, err := m.store.GetPublished()
	if err != nil {
		return nil
	}

	return items
}

func (m *Manager) GetRejected() []RejectedNewsEvent {
	if m.store == nil {
		return nil
	}

	items, err := m.store.GetRejected()
	if err != nil {
		return nil
	}

	return items
}

func (m *Manager) GetCurrentSourceHealth() []SourceHealthEvent {
	if m.store == nil {
		return nil
	}

	items, err := m.store.GetSourceHealthCurrent()
	if err != nil {
		return nil
	}

	return items
}

func (m *Manager) GetSourceHealth() []SourceHealthEvent {
	return m.GetCurrentSourceHealth()
}

func (m *Manager) BuildSummary() Summary {
	sources := m.GetCurrentSourceHealth()

	healthy := 0
	disabled := 0
	degraded := 0

	for _, source := range sources {
		switch {
		case isFutureTimestamp(source.DisabledUntil):
			disabled++
		case source.ConsecutiveFails > 0:
			degraded++
		default:
			healthy++
		}
	}

	publishedCount := 0
	rejectedCount := 0
	if m.store != nil {
		publishedCount = m.store.CountPublished()
		rejectedCount = m.store.CountRejected()
	}

	return Summary{
		PublishedCount:    publishedCount,
		RejectedCount:     rejectedCount,
		HealthySources:    healthy,
		DisabledSources:   disabled,
		DegradedSources:   degraded,
		TrackedSourceSize: len(sources),
	}
}

func (m *Manager) ExportPublishedJSONL() ([]byte, error) {
	if m.store == nil {
		return nil, nil
	}
	return m.store.ExportPublishedJSONL()
}

func (m *Manager) ExportRejectedJSONL() ([]byte, error) {
	if m.store == nil {
		return nil, nil
	}
	return m.store.ExportRejectedJSONL()
}

func (m *Manager) ExportSourceHealthJSONL() ([]byte, error) {
	if m.store == nil {
		return nil, nil
	}
	return m.store.ExportSourceHealthJSONL()
}

func isFutureTimestamp(raw string) bool {
	if raw == "" {
		return false
	}

	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return false
	}

	return t.After(time.Now())
}

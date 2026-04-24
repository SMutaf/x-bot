package monitoring

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/ingestion/dedup"
	"github.com/redis/go-redis/v9"
)

const (
	retentionDays          = 7
	dateLayout             = "2006-01-02"
	publishedKeyPrefix     = "dashboard:published"
	rejectedKeyPrefix      = "dashboard:rejected"
	publishedCountPrefix   = "dashboard:counts:published"
	rejectedCountPrefix    = "dashboard:counts:rejected"
	sourceHealthCurrentKey = "dashboard:source-health:current"
)

type RedisStore struct {
	client        *redis.Client
	ctx           context.Context
	retentionDays int
}

func NewRedisStore(cache *dedup.Deduplicator, days int) *RedisStore {
	if cache == nil || cache.Client == nil {
		return nil
	}

	if days <= 0 {
		days = retentionDays
	}

	return &RedisStore{
		client:        cache.Client,
		ctx:           cache.Ctx,
		retentionDays: days,
	}
}

func (s *RedisStore) SavePublished(event PublishedNewsEvent) error {
	return s.pushDailyJSON(publishedKeyPrefix, event.Time, event)
}

func (s *RedisStore) SaveRejected(event RejectedNewsEvent) error {
	return s.pushDailyJSON(rejectedKeyPrefix, event.Time, event)
}

func (s *RedisStore) SaveSourceHealth(event SourceHealthEvent) error {
	if s == nil {
		return nil
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return s.client.HSet(s.ctx, sourceHealthCurrentKey, event.URL, data).Err()
}

func (s *RedisStore) GetPublished() ([]PublishedNewsEvent, error) {
	return readDailyJSON[PublishedNewsEvent](s, publishedKeyPrefix)
}

func (s *RedisStore) GetRejected() ([]RejectedNewsEvent, error) {
	return readDailyJSON[RejectedNewsEvent](s, rejectedKeyPrefix)
}

func (s *RedisStore) GetSourceHealthCurrent() ([]SourceHealthEvent, error) {
	if s == nil {
		return nil, nil
	}

	values, err := s.client.HGetAll(s.ctx, sourceHealthCurrentKey).Result()
	if err != nil {
		return nil, err
	}

	items := make([]SourceHealthEvent, 0, len(values))
	for _, raw := range values {
		var item SourceHealthEvent
		if err := json.Unmarshal([]byte(raw), &item); err != nil {
			continue
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Category != items[j].Category {
			return items[i].Category < items[j].Category
		}
		return items[i].SourceName < items[j].SourceName
	})

	return items, nil
}

func (s *RedisStore) CountPublished() int {
	return s.countDailyKeys(publishedKeyPrefix)
}

func (s *RedisStore) CountRejected() int {
	return s.countDailyKeys(rejectedKeyPrefix)
}

func (s *RedisStore) ExportPublishedJSONL() ([]byte, error) {
	items, err := s.GetPublished()
	if err != nil {
		return nil, err
	}
	return encodeJSONLLines(items)
}

func (s *RedisStore) ExportRejectedJSONL() ([]byte, error) {
	items, err := s.GetRejected()
	if err != nil {
		return nil, err
	}
	return encodeJSONLLines(items)
}

func (s *RedisStore) ExportSourceHealthJSONL() ([]byte, error) {
	items, err := s.GetSourceHealthCurrent()
	if err != nil {
		return nil, err
	}
	return encodeJSONLLines(items)
}

func (s *RedisStore) pushDailyJSON(prefix string, t time.Time, value any) error {
	if s == nil {
		return nil
	}

	if t.IsZero() {
		t = time.Now()
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	key := s.dailyKey(prefix, t)
	pipe := s.client.TxPipeline()
	pipe.LPush(s.ctx, key, data)
	pipe.Expire(s.ctx, key, time.Duration(s.retentionDays)*24*time.Hour)
	if counterPrefix := counterPrefixFor(prefix); counterPrefix != "" {
		countKey := s.dailyKey(counterPrefix, t)
		pipe.Incr(s.ctx, countKey)
		pipe.Expire(s.ctx, countKey, time.Duration(s.retentionDays)*24*time.Hour)
	}
	_, err = pipe.Exec(s.ctx)
	return err
}

func (s *RedisStore) countDailyKeys(prefix string) int {
	if s == nil {
		return 0
	}

	total := 0
	for _, key := range s.dailyKeys(counterPrefixFor(prefix)) {
		n, err := s.client.Get(s.ctx, key).Int()
		if err != nil {
			continue
		}
		total += n
	}
	return total
}

func (s *RedisStore) dailyKey(prefix string, t time.Time) string {
	return fmt.Sprintf("%s:%s", prefix, t.Format(dateLayout))
}

func (s *RedisStore) dailyKeys(prefix string) []string {
	keys := make([]string, 0, s.retentionDays)
	now := time.Now()

	for i := 0; i < s.retentionDays; i++ {
		keys = append(keys, s.dailyKey(prefix, now.AddDate(0, 0, -i)))
	}

	return keys
}

func counterPrefixFor(prefix string) string {
	switch prefix {
	case publishedKeyPrefix:
		return publishedCountPrefix
	case rejectedKeyPrefix:
		return rejectedCountPrefix
	default:
		return ""
	}
}

func readDailyJSON[T any](s *RedisStore, prefix string) ([]T, error) {
	if s == nil {
		return nil, nil
	}

	items := make([]T, 0)
	for _, key := range s.dailyKeys(prefix) {
		values, err := s.client.LRange(s.ctx, key, 0, -1).Result()
		if err != nil {
			return nil, err
		}

		for _, raw := range values {
			var item T
			if err := json.Unmarshal([]byte(raw), &item); err != nil {
				continue
			}
			items = append(items, item)
		}
	}

	sortByTimeDesc(items)
	return items, nil
}

func sortByTimeDesc[T any](items []T) {
	sort.Slice(items, func(i, j int) bool {
		return itemTime(items[i]).After(itemTime(items[j]))
	})
}

func itemTime[T any](item T) time.Time {
	switch v := any(item).(type) {
	case PublishedNewsEvent:
		return v.Time
	case RejectedNewsEvent:
		return v.Time
	case SourceHealthEvent:
		return v.Time
	default:
		return time.Time{}
	}
}

func encodeJSONLLines[T any](items []T) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	for _, item := range items {
		if err := enc.Encode(item); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

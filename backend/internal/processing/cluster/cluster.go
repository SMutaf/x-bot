package eventcluster

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/domain/models"
	"github.com/redis/go-redis/v9"
)

const (
	clusterKeyPrefix    = "event:"
	sentKeyPrefix       = "event_sent:"
	similarityThreshold = 0.3
)

var stopwords = map[string]bool{
	"the": true, "a": true, "an": true, "on": true, "of": true,
	"for": true, "at": true, "in": true, "to": true, "and": true,
	"is": true, "it": true, "its": true, "by": true, "with": true,
	"as": true, "be": true, "are": true, "was": true, "were": true,
	"has": true, "have": true, "had": true, "that": true, "this": true,
	"from": true, "or": true, "but": true, "not": true, "what": true,
	"why": true, "how": true, "who": true, "will": true, "can": true,
	"bir": true, "bu": true, "ve": true, "da": true, "de": true,
	"ile": true, "için": true, "mi": true, "mı": true, "mu": true,
	"mü": true, "ne": true, "en": true, "çok": true, "daha": true,
	"olan": true, "var": true, "her": true,
}

type Event struct {
	Key     string          `json:"key"`
	Sources map[string]bool `json:"sources"`
	Count   int             `json:"count"`
	Created time.Time       `json:"created"`
}

type EventClusterer struct {
	client *redis.Client
	ctx    context.Context
	ttl    map[models.NewsCategory]time.Duration
}

func NewEventClusterer(client *redis.Client) *EventClusterer {
	return &EventClusterer{
		client: client,
		ctx:    context.Background(),
		ttl: map[models.NewsCategory]time.Duration{
			models.CategoryBreaking: 30 * time.Minute,
			models.CategoryTech:     2 * time.Hour,
			models.CategoryGeneral:  2 * time.Hour,
			models.CategoryEconomy:  2 * time.Hour,
			models.CategorySports:   1 * time.Hour,
			models.CategoryScience:  2 * time.Hour,
		},
	}
}

func normalizeTitle(title string) []string {
	title = strings.ToLower(title)
	replacer := strings.NewReplacer(
		".", "", ",", "", ":", "", ";", "", "!", "", "?", "",
		"'", "", "\"", "", "-", " ", "_", " ", "/", " ",
		"(", "", ")", "", "[", "", "]", "",
	)
	title = replacer.Replace(title)

	words := strings.Fields(title)
	tokens := make([]string, 0, len(words))
	for _, w := range words {
		if !stopwords[w] && len(w) > 2 {
			tokens = append(tokens, w)
		}
	}
	return tokens
}

func jaccardSimilarity(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	setA := make(map[string]bool, len(a))
	for _, t := range a {
		setA[t] = true
	}

	intersection := 0
	setB := make(map[string]bool, len(b))
	for _, t := range b {
		setB[t] = true
		if setA[t] {
			intersection++
		}
	}

	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

func tokenKey(tokens []string) string {
	sorted := make([]string, len(tokens))
	copy(sorted, tokens)

	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return clusterKeyPrefix + strings.Join(sorted, "_")
}

func (ec *EventClusterer) findSimilarEvent(tokens []string) (event *Event, redisKey string) {
	exactKey := tokenKey(tokens)

	data, err := ec.client.Get(ec.ctx, exactKey).Result()
	if err == nil {
		var e Event
		if json.Unmarshal([]byte(data), &e) == nil {
			return &e, exactKey
		}
	}

	keys, err := ec.client.Keys(ec.ctx, clusterKeyPrefix+"*").Result()
	if err != nil || len(keys) == 0 {
		return nil, exactKey
	}

	bestSimilarity := 0.0
	var bestEvent *Event
	var bestRedisKey string

	for _, key := range keys {
		data, err := ec.client.Get(ec.ctx, key).Result()
		if err != nil {
			continue
		}

		var e Event
		if json.Unmarshal([]byte(data), &e) != nil {
			continue
		}

		keyTokens := strings.Split(strings.TrimPrefix(key, clusterKeyPrefix), "_")
		sim := jaccardSimilarity(tokens, keyTokens)

		if sim > bestSimilarity && sim >= similarityThreshold {
			bestSimilarity = sim
			bestRedisKey = key
			bestEvent = &e
		}
	}

	if bestEvent != nil {
		return bestEvent, bestRedisKey
	}
	return nil, exactKey
}

func (ec *EventClusterer) AddEvent(news models.RawNewsItem) (boost int, sourceCount int, clusterKey string) {
	tokens := normalizeTitle(news.Title)
	if len(tokens) == 0 {
		return 0, 1, ""
	}

	ttl := ec.ttl[news.Category]
	if ttl == 0 {
		ttl = 2 * time.Hour
	}

	existingEvent, redisKey := ec.findSimilarEvent(tokens)

	if existingEvent == nil {
		event := &Event{
			Key:     redisKey,
			Sources: map[string]bool{news.Source: true},
			Count:   1,
			Created: time.Now(),
		}
		data, _ := json.Marshal(event)
		ec.client.Set(ec.ctx, redisKey, string(data), ttl)
		fmt.Printf("[CLUSTER] Yeni event (count:1): %s\n", news.Title)
		return 0, 1, redisKey
	}

	if existingEvent.Sources[news.Source] {
		fmt.Printf("[CLUSTER] Aynı kaynak tekrarı, boost yok (%s): %s\n", news.Source, news.Title)
		return 0, existingEvent.Count, redisKey
	}

	existingEvent.Sources[news.Source] = true
	existingEvent.Count++

	data, _ := json.Marshal(existingEvent)
	ec.client.Set(ec.ctx, redisKey, string(data), ttl)

	boost = calculateBoost(existingEvent.Count)
	fmt.Printf("[CLUSTER] Event güncellendi (count:%d, boost:+%d, kaynaklar:%v): %s\n",
		existingEvent.Count, boost, sourceKeys(existingEvent.Sources), news.Title)

	return boost, existingEvent.Count, redisKey
}

func calculateBoost(count int) int {
	switch {
	case count >= 5:
		return 60
	case count >= 3:
		return 40
	case count >= 2:
		return 20
	default:
		return 0
	}
}

func sourceKeys(sources map[string]bool) []string {
	keys := make([]string, 0, len(sources))
	for k := range sources {
		keys = append(keys, k)
	}
	return keys
}

func (ec *EventClusterer) WasSentRecently(clusterKey string) bool {
	if clusterKey == "" {
		return false
	}
	key := sentKeyPrefix + clusterKey
	exists, err := ec.client.Exists(ec.ctx, key).Result()
	if err != nil {
		return false
	}
	return exists > 0
}

func (ec *EventClusterer) MarkSent(clusterKey string, ttl time.Duration) {
	if clusterKey == "" {
		return
	}
	key := sentKeyPrefix + clusterKey
	ec.client.Set(ec.ctx, key, time.Now().Format(time.RFC3339), ttl)
}

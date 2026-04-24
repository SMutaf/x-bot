package scoring

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/models"
	"github.com/redis/go-redis/v9"
)

type BurstProvider struct {
	redisClient *redis.Client
	ctx         context.Context
}

func NewBurstProvider(redisClient *redis.Client) *BurstProvider {
	return &BurstProvider{
		redisClient: redisClient,
		ctx:         context.Background(),
	}
}

func (b *BurstProvider) Score(env models.NewsEnvelope) float64 {
	if b.redisClient == nil {
		return 0
	}

	if !b.shouldTrackBurst(env) {
		return 0
	}

	if requiresClusterGate(env.News.Category) && env.Cluster.ClusterCount < 2 {
		return 0
	}

	now := time.Now()
	window := now.Unix() / 300
	currentKey := fmt.Sprintf("burst:%s:%d", env.News.Category, window)

	currentCount, err := b.redisClient.Incr(b.ctx, currentKey).Result()
	if err != nil {
		return 0
	}
	b.redisClient.Expire(b.ctx, currentKey, 45*time.Minute)

	var total float64
	var samples float64

	for i := int64(1); i <= 6; i++ {
		prevKey := fmt.Sprintf("burst:%s:%d", env.News.Category, window-i)
		val, err := b.redisClient.Get(b.ctx, prevKey).Result()
		if err != nil {
			continue
		}

		n, convErr := strconv.Atoi(val)
		if convErr != nil {
			continue
		}

		total += float64(n)
		samples++
	}

	baseline := 1.0
	if samples > 0 {
		baseline = total / samples
		if baseline < 1 {
			baseline = 1
		}
	}

	ratio := float64(currentCount) / baseline

	switch {
	case currentCount >= 8 && ratio >= 2.2:
		return 100
	case currentCount >= 6 && ratio >= 1.8:
		return 75
	case currentCount >= 4 && ratio >= 1.5:
		return 50
	case currentCount >= 3 && ratio >= 1.3:
		return 30
	default:
		return 0
	}
}

func (b *BurstProvider) shouldTrackBurst(env models.NewsEnvelope) bool {
	switch env.News.Category {
	case models.CategoryBreaking, models.CategoryGeneral:
		return true
	case models.CategoryEconomy, models.CategoryTech:
		return hasBurstTrigger(env.News.Title + " " + env.News.Description)
	default:
		return false
	}
}

func requiresClusterGate(category models.NewsCategory) bool {
	return category == models.CategoryBreaking || category == models.CategoryGeneral
}

func hasBurstTrigger(text string) bool {
	lower := strings.ToLower(text)

	triggers := []string{
		"flash crash", "market crash", "rate hike", "rate cut", "faiz kararı", "faiz karari",
		"tcmb", "fed", "bankruptcy", "iflas", "sanctions", "yaptırım", "yaptirim",
		"mass layoff", "lay off", "acquisition", "merger", "breach", "hack",
		"openai", "meta", "microsoft", "apple", "google", "tesla", "nvidia",
		"trump", "putin", "nato", "israel", "iran", "gaza", "gazze",
	}

	for _, trigger := range triggers {
		if strings.Contains(lower, trigger) {
			return true
		}
	}

	return false
}

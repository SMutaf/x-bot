package virality

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/SMutaf/twitter-bot/backend/internal/models"
	"github.com/redis/go-redis/v9"
)

var lightKeywords = []string{
	"killed", "died", "arrested", "collapsed", "attacked",
	"struck", "seized", "resigned", "launched", "exploded",
	"sanctions", "evacuated", "invaded", "crashed",
	"öldü", "yaralandı", "tutuklandı", "saldırı", "patlama",
	"deprem", "çöktü", "istifa", "vuruldu",
	"faiz", "interest rate", "inflation", "enflasyon", "oil", "petrol", "lng",
	"fed", "gas facility", "energy", "radar", "resmi gazete", "dışişleri",
}

type ViralityScorer struct {
	redisClient *redis.Client
	ctx         context.Context
}

func NewViralityScorer(redisClient *redis.Client) *ViralityScorer {
	return &ViralityScorer{
		redisClient: redisClient,
		ctx:         context.Background(),
	}
}

func clusterScore(clusterCount int) float64 {
	switch {
	case clusterCount >= 5:
		return 100
	case clusterCount >= 4:
		return 92
	case clusterCount >= 3:
		return 82
	case clusterCount >= 2:
		return 65
	default:
		return 0
	}
}

func recencyScore(publishedAt time.Time) float64 {
	if publishedAt.IsZero() {
		return 0
	}

	diff := time.Since(publishedAt).Minutes()
	switch {
	case diff < 5:
		return 100
	case diff < 15:
		return 85
	case diff < 30:
		return 70
	case diff < 60:
		return 55
	case diff < 120:
		return 35
	case diff < 240:
		return 22
	default:
		return 8
	}
}

func (v *ViralityScorer) burstScore(env models.NewsEnvelope) float64 {
	if v.redisClient == nil {
		return 0
	}

	if env.News.Category != models.CategoryBreaking && env.News.Category != models.CategoryGeneral {
		return 0
	}

	if env.Cluster.ClusterCount < 2 {
		return 0
	}

	now := time.Now()
	window := now.Unix() / 300
	currentKey := fmt.Sprintf("burst:%s:%d", env.News.Category, window)

	currentCount, err := v.redisClient.Incr(v.ctx, currentKey).Result()
	if err != nil {
		return 0
	}
	_ = v.redisClient.Expire(v.ctx, currentKey, 45*time.Minute).Err()

	var total float64
	var samples float64

	for i := int64(1); i <= 6; i++ {
		prevKey := fmt.Sprintf("burst:%s:%d", env.News.Category, window-i)
		val, err := v.redisClient.Get(v.ctx, prevKey).Result()
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

func keywordScore(text string) float64 {
	count := 0
	for _, kw := range lightKeywords {
		if strings.Contains(text, kw) {
			count++
		}
	}

	if hasLargeNumber(text) {
		count++
	}

	if count == 0 {
		return 0
	}

	return math.Min(float64(count*15), 100)
}

func hasLargeNumber(text string) bool {
	signals := []string{"billion", "trillion", "million", "milyar", "trilyon", "milyon", "%"}
	for _, s := range signals {
		if strings.Contains(text, s) {
			return true
		}
	}

	digits := 0
	for _, r := range text {
		if unicode.IsDigit(r) {
			digits++
			if digits >= 2 {
				return true
			}
		}
	}

	return false
}

func (v *ViralityScorer) CalculateScore(env models.NewsEnvelope) int {
	text := strings.ToLower(env.News.Title + " " + env.News.Description)

	cScore := clusterScore(env.Cluster.ClusterCount)
	rScore := recencyScore(env.News.PublishedAt)
	bScore := v.burstScore(env)
	kScore := keywordScore(text)

	var raw float64

	switch env.News.Category {
	case models.CategoryBreaking:
		raw = (cScore * 0.58) + (rScore * 0.24) + (bScore * 0.12) + (kScore * 0.06)
	case models.CategoryGeneral:
		raw = (cScore * 0.10) + (rScore * 0.42) + (bScore * 0.03) + (kScore * 0.45)
	case models.CategoryEconomy:
		raw = (cScore * 0.08) + (rScore * 0.38) + (bScore * 0.00) + (kScore * 0.54)
	case models.CategoryTech:
		raw = (cScore * 0.10) + (rScore * 0.30) + (bScore * 0.00) + (kScore * 0.60)
	default:
		raw = (cScore * 0.20) + (rScore * 0.35) + (bScore * 0.05) + (kScore * 0.40)
	}

	final := int(math.Round(raw))
	if final > 100 {
		final = 100
	}
	if final < 0 {
		final = 0
	}

	fmt.Printf(
		"[VIRALITY DETAIL] cluster:%.0f recency:%.0f burst:%.0f keyword:%.0f => %d | %s\n",
		cScore, rScore, bScore, kScore, final, env.News.Title,
	)

	return final
}

func (v *ViralityScorer) GetViralityLevel(score int) string {
	switch {
	case score >= 70:
		return "ULTRA_VIRAL"
	case score >= 50:
		return "HIGH_VIRAL"
	case score >= 30:
		return "MEDIUM_VIRAL"
	case score >= 15:
		return "LOW_VIRAL"
	default:
		return "NOT_VIRAL"
	}
}

package scoring

import (
	"fmt"

	"github.com/SMutaf/twitter-bot/backend/internal/models"
	"github.com/redis/go-redis/v9"
)

type NewsScorer struct {
	burstProvider *BurstProvider
}

func NewNewsScorer(redisClient *redis.Client) *NewsScorer {
	return &NewsScorer{
		burstProvider: NewBurstProvider(redisClient),
	}
}

func (s *NewsScorer) Calculate(item models.NewsItem) ScoreBreakdown {
	cScore := ClusterScore(item.ClusterCount)
	rScore := RecencyScore(item.PublishedAt)
	bScore := s.burstProvider.Score(item)
	kScore := KeywordScore(item.Title + " " + item.Description)

	var raw float64

	switch item.Category {
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

	final := clampScore(raw)

	fmt.Printf(
		"[VIRALITY DETAIL] cluster:%.0f recency:%.0f burst:%.0f keyword:%.0f => %d | %s\n",
		cScore, rScore, bScore, kScore, final, item.Title,
	)

	return ScoreBreakdown{
		Cluster: cScore,
		Recency: rScore,
		Burst:   bScore,
		Keyword: kScore,
		Final:   final,
	}
}

func (s *NewsScorer) GetViralityLevel(score int) string {
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

func clampScore(v float64) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return int(v + 0.5)
}

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

func (s *NewsScorer) Calculate(env models.NewsEnvelope) models.ScoreBreakdown {
	cScore := ClusterScore(env.Cluster.ClusterCount)
	rScore := RecencyScore(env.News.PublishedAt)
	bScore := s.burstProvider.Score(env)
	text := env.News.Title + " " + env.News.Description
	kScore := KeywordScore(text)
	tScore := TurkeyRelevanceScore(text)
	mScore := MagnitudeScore(text)

	var raw float64

	switch env.News.Category {
	case models.CategoryBreaking:
		raw = (cScore * 0.48) + (rScore * 0.20) + (bScore * 0.12) + (kScore * 0.12) + (tScore * 0.04) + (mScore * 0.04)
	case models.CategoryGeneral:
		raw = (cScore * 0.08) + (rScore * 0.30) + (bScore * 0.05) + (kScore * 0.32) + (tScore * 0.18) + (mScore * 0.07)
	case models.CategoryEconomy:
		raw = (cScore * 0.06) + (rScore * 0.22) + (bScore * 0.08) + (kScore * 0.34) + (tScore * 0.20) + (mScore * 0.10)
	case models.CategoryTech:
		raw = (cScore * 0.08) + (rScore * 0.22) + (bScore * 0.07) + (kScore * 0.43) + (tScore * 0.08) + (mScore * 0.12)
	default:
		raw = (cScore * 0.16) + (rScore * 0.28) + (bScore * 0.06) + (kScore * 0.34) + (tScore * 0.10) + (mScore * 0.06)
	}

	final := clampScore(raw)

	fmt.Printf(
		"[VIRALITY DETAIL] cluster:%.0f recency:%.0f burst:%.0f keyword:%.0f turkey:%.0f magnitude:%.0f => %d | %s\n",
		cScore, rScore, bScore, kScore, tScore, mScore, final, env.News.Title,
	)

	return models.ScoreBreakdown{
		Cluster:   cScore,
		Recency:   rScore,
		Burst:     bScore,
		Keyword:   kScore,
		Turkey:    tScore,
		Magnitude: mScore,
		Final:     final,
		Boost:     0,
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

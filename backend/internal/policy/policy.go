package policy

import (
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/models"
)

type CategoryPolicy struct {
	MinClusterCount int
	MinVirality     int
	DedupeCooldown  time.Duration
	MaxAge          time.Duration
}

var categoryPolicies = map[models.NewsCategory]CategoryPolicy{
	models.CategoryBreaking: {
		MinClusterCount: 2,
		MinVirality:     38,
		DedupeCooldown:  45 * time.Minute,
		MaxAge:          90 * time.Minute,
	},
	models.CategoryEconomy: {
		MinClusterCount: 1,
		MinVirality:     26,
		DedupeCooldown:  90 * time.Minute,
		MaxAge:          4 * time.Hour,
	},
	models.CategoryTech: {
		MinClusterCount: 1,
		MinVirality:     22,
		DedupeCooldown:  2 * time.Hour,
		MaxAge:          8 * time.Hour,
	},
	models.CategoryGeneral: {
		MinClusterCount: 1,
		MinVirality:     20,
		DedupeCooldown:  90 * time.Minute,
		MaxAge:          3 * time.Hour,
	},
	models.CategorySports: {
		MinClusterCount: 1,
		MinVirality:     24,
		DedupeCooldown:  90 * time.Minute,
		MaxAge:          2 * time.Hour,
	},
	models.CategoryScience: {
		MinClusterCount: 1,
		MinVirality:     22,
		DedupeCooldown:  2 * time.Hour,
		MaxAge:          8 * time.Hour,
	},
}

func Get(category models.NewsCategory) CategoryPolicy {
	p, ok := categoryPolicies[category]
	if ok {
		return p
	}
	return CategoryPolicy{
		MinClusterCount: 1,
		MinVirality:     20,
		DedupeCooldown:  90 * time.Minute,
		MaxAge:          4 * time.Hour,
	}
}

func IsFreshEnough(item models.NewsItem, policy CategoryPolicy) bool {
	if item.PublishedAt.IsZero() {
		return true
	}

	diff := time.Since(item.PublishedAt)
	if diff <= policy.MaxAge {
		return true
	}

	if IsCriticalEvent(item) && IsAcceptableCriticalAge(item, policy) {
		return true
	}

	return false
}

func IsAcceptableCriticalAge(item models.NewsItem, _ CategoryPolicy) bool {
	if item.PublishedAt.IsZero() {
		return true
	}

	diff := time.Since(item.PublishedAt)

	switch item.Category {
	case models.CategoryBreaking:
		return diff <= 3*time.Hour
	case models.CategoryEconomy:
		return diff <= 6*time.Hour
	case models.CategoryGeneral:
		return diff <= 6*time.Hour
	default:
		return diff <= 4*time.Hour
	}
}

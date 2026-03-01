package filter

import (
	"strings"

	"github.com/SMutaf/twitter-bot/backend/internal/models"
	"github.com/SMutaf/twitter-bot/backend/internal/virality"
)

// FilterOptions filtreleme için opsiyonlar
type FilterOptions struct {
	MinVirality    int
	IncludeTech    bool
	IncludeGeneral bool
}

// NewsFilter struct
type NewsFilter struct {
	scorer   *virality.ViralityScorer
	MinScore int
}

func NewNewsFilter(vs *virality.ViralityScorer, minScore int) *NewsFilter {
	return &NewsFilter{
		scorer:   virality.NewViralityScorer(),
		MinScore: minScore,
	}
}

// FilterNews haberleri filtreler
func (f *NewsFilter) FilterNews(news []models.NewsItem, opts FilterOptions) []models.NewsItem {
	var result []models.NewsItem

	for _, item := range news {
		score := f.scorer.CalculateScore(item.Title, item.Description, item.Category)
		if score < opts.MinVirality {
			continue
		}

		if !opts.IncludeTech && item.Category == models.CategoryTech {
			continue
		}

		if !opts.IncludeGeneral && item.Category == models.CategoryGeneral {
			continue
		}

		result = append(result, item)
	}

	return result
}

// ContainsKeyword filtreleme yardımı
func ContainsKeyword(text string, keywords []string) bool {
	text = strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(text, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

func (f *NewsFilter) IsAllowed(title, description string) bool {
	score := f.scorer.CalculateScore(title, description, "")
	return score >= 30 // ya f.MinScore kullanabilirsin
}

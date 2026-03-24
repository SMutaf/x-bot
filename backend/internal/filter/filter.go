package filter

import "github.com/SMutaf/twitter-bot/backend/internal/models"

type NewsFilter struct{}

func NewNewsFilter() *NewsFilter {
	return &NewsFilter{}
}

func (f *NewsFilter) ShouldProcess(item models.NewsItem) (bool, string) {
	title := normalize(item.Title)
	text := normalize(item.Title + " " + item.Description)

	if IsBackground(title) {
		return false, "background-analysis"
	}

	switch item.Category {
	case models.CategoryBreaking:
		if !IsBreakingRelevant(text) {
			return false, "breaking-not-relevant"
		}
	case models.CategoryEconomy:
		if !IsEconomyRelevant(text) {
			return false, "economy-not-relevant"
		}
	case models.CategoryTech:
		if !IsTechRelevant(text) {
			return false, "tech-not-relevant"
		}
	case models.CategoryGeneral:
		if !IsGeneralRelevant(text) {
			return false, "general-not-relevant"
		}
	}

	return true, "filter-ok"
}

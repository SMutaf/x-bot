package filter

import "github.com/SMutaf/twitter-bot/backend/internal/models"

type NewsFilter struct{}

func NewNewsFilter() *NewsFilter {
	return &NewsFilter{}
}

func (f *NewsFilter) ShouldProcess(env models.NewsEnvelope) (bool, string) {
	title := normalize(env.News.Title)
	text := normalize(env.News.Title + " " + env.News.Description)

	if IsBackground(title) {
		return false, "background-analysis"
	}

	switch env.News.Category {
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

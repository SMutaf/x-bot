package filter

import (
	"fmt"
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/models"
)

type NewsFilter struct {
	BreakingThreshold int
	NormalThreshold   int
}

func NewNewsFilter(breaking int, normal int) *NewsFilter {
	return &NewsFilter{
		BreakingThreshold: breaking,
		NormalThreshold:   normal,
	}
}

func calculateScore(item models.NewsItem) int {
	title := normalize(item.Title)
	fullText := normalize(item.Title + " " + item.Description)

	if IsBackground(title) {
		return 0
	}

	score := 0

	if !item.PublishedAt.IsZero() {
		diff := time.Since(item.PublishedAt)
		switch {
		case diff < 10*time.Minute:
			score += 20
		case diff < 30*time.Minute:
			score += 14
		case diff < 2*time.Hour:
			score += 8
		case diff < 4*time.Hour:
			score += 4
		}
	}

	for verb, pts := range actionVerbs {
		if contains(fullText, verb) {
			score += pts
		}
	}

	for actor, pts := range majorActors {
		if contains(fullText, actor) {
			score += pts
		}
	}

	for actor, pts := range shortActors {
		if containsWord(fullText, actor) {
			score += pts
		}
	}

	if containsAny(fullText, turkeyKeywords) {
		score += 18
	}

	if hasNumber(fullText) {
		score += 8
	}

	switch item.Category {
	case models.CategoryBreaking:
		score += 16
	case models.CategoryEconomy:
		score += 16
	case models.CategoryTech:
		score += 10
	case models.CategoryGeneral:
		score += 14
	}

	if score > 100 {
		score = 100
	}
	return score
}

func (f *NewsFilter) ShouldProcess(item models.NewsItem, boost int) (bool, string) {
	score := calculateScore(item)
	if score == 0 {
		return false, "background-analysis"
	}

	text := normalize(item.Title + " " + item.Description)

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

	total := score + boost
	if total > 100 {
		total = 100
	}

	if item.Category == models.CategoryBreaking {
		if total < f.BreakingThreshold {
			return false, fmt.Sprintf("breaking-score-low(%d+%d=%d)", score, boost, total)
		}
		return true, fmt.Sprintf("breaking-ok(%d+%d=%d)", score, boost, total)
	}

	if total < f.NormalThreshold {
		return false, fmt.Sprintf("normal-score-low(%d+%d=%d)", score, boost, total)
	}

	return true, fmt.Sprintf("normal-ok(%d+%d=%d)", score, boost, total)
}

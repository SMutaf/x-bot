package filter

import (
	"strings"
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

func contains(text string, word string) bool {
	return strings.Contains(text, word)
}

func containsAny(text string, words []string) bool {
	for _, w := range words {
		if contains(text, w) {
			return true
		}
	}
	return false
}

var backgroundPatterns = []string{
	"why ",
	"what is ",
	"how ",
	"analysis",
	"opinion",
	"explained",
	"timeline",
	"profile",
	"watch:",
}

var actionVerbs = map[string]int{

	"attack":   30,
	"strike":   30,
	"launch":   25,
	"kill":     35,
	"die":      35,
	"warn":     20,
	"sanction": 20,
	"ban":      20,
	"explode":  30,
	"crash":    30,
	"resign":   25,
	"arrest":   25,
	"close":    20,
	"halt":     20,

	"saldırı": 30,
	"deprem":  35,
	"patlama": 30,
}

var majorActors = map[string]int{

	"us":      20,
	"america": 20,
	"china":   20,
	"russia":  20,
	"iran":    20,
	"israel":  20,
	"turkey":  25,
	"türkiye": 25,

	"nato":     20,
	"un":       20,
	"pentagon": 20,
	"fed":      25,

	"erdogan": 25,
	"trump":   25,
}

var turkeyKeywords = []string{
	"turkey",
	"türkiye",
	"istanbul",
	"ankara",
	"bosphorus",
}

func isBackground(text string) bool {

	for _, p := range backgroundPatterns {

		if contains(text, p) {
			return true
		}
	}

	return false
}

func calculateScore(item models.NewsItem) int {

	text := strings.ToLower(item.Title + " " + item.Description)

	if isBackground(text) {
		return 0
	}

	score := 0

	if !item.PublishedAt.IsZero() {

		diff := time.Since(item.PublishedAt)

		switch {

		case diff < 10*time.Minute:
			score += 30

		case diff < 30*time.Minute:
			score += 20

		case diff < 2*time.Hour:
			score += 10
		}
	}

	for verb, pts := range actionVerbs {

		if contains(text, verb) {
			score += pts
		}
	}

	for actor, pts := range majorActors {

		if contains(text, actor) {
			score += pts
		}
	}

	if containsAny(text, turkeyKeywords) {
		score += 25
	}

	for i := 0; i <= 9; i++ {

		if contains(text, string('0'+i)) {

			score += 10
			break
		}
	}

	switch item.Category {

	case models.CategoryBreaking:
		score += 25

	case models.CategoryEconomy:
		score += 20

	case models.CategoryTech:
		score += 15

	case models.CategoryGeneral:
		score += 10
	}

	if score > 100 {
		score = 100
	}

	return score
}

func (f *NewsFilter) ShouldProcess(item models.NewsItem) (bool, string) {

	score := calculateScore(item)

	if score == 0 {
		return false, "background-analysis"
	}

	if item.Category == models.CategoryBreaking {

		if score < f.BreakingThreshold {
			return false, "breaking-score-low"
		}

		return true, "breaking-score-ok"
	}

	if score < f.NormalThreshold {
		return false, "normal-score-low"
	}

	return true, "normal-score-ok"
}

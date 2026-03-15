package virality

import (
	"strings"

	"github.com/SMutaf/twitter-bot/backend/internal/models"
)

// IMPORTANCE_KEYWORDS — Breaking haberlerin önem skoru için
var IMPORTANCE_KEYWORDS = map[string]int{
	// Çatışma & Güvenlik
	"öldü": 20, "ölü": 20, "hayatını kaybetti": 20,
	"savaş": 20, "saldırı": 20, "deprem": 20,
	"füze": 15, "nükleer": 15, "terör": 15, "darbe": 15,
	// İngilizce
	"died": 20, "death": 20, "war": 20, "attack": 20,
	"earthquake": 20, "missile": 15, "nuclear": 15, "terror": 15, "coup": 15,
	// Liderler
	"erdoğan": 15, "cumhurbaşkanı": 15, "president": 15, "prime minister": 15,
	// Ekonomi (global)
	"recession": 15, "collapse": 15, "sanctions": 15,
	"kriz": 10, "çöküş": 10, "yaptırım": 10,
}

// ENGAGEMENT_KEYWORDS — Normal haberlerin etkileşim skoru için
var ENGAGEMENT_KEYWORDS = map[string]int{
	// Ürün & Tech
	"iphone": 20, "chatgpt": 20, "openai": 20, "gemini": 15, "claude": 15,
	"tanıttı": 15, "duyurdu": 15, "launched": 15, "announced": 15,
	// Ekonomik etki
	"zam": 20, "faiz": 15, "enflasyon": 15, "bitcoin": 15, "kripto": 10,
	"dolar": 10, "indirim": 15, "maaş": 15,
	// Rekor
	"rekor": 15, "ilk kez": 15, "tarihi": 10, "record": 15, "historic": 10,
	// Figürler
	"elon musk": 15, "trump": 10,
}

// Negatif — engagement düşürür
var ANTI_ENGAGEMENT_KEYWORDS = map[string]int{
	"top 10": -20, "top 5": -20, "tutorial": -15,
	"nasıl": -10, "how to": -10, "review": -10, "inceleme": -10,
}

var CATEGORY_BASE_SCORE = map[models.NewsCategory]int{
	models.CategoryBreaking: 30,
	models.CategoryTech:     15,
	models.CategoryGeneral:  15,
	models.CategoryEconomy:  20,
	models.CategorySports:   10,
	models.CategoryScience:  10,
}

type ViralityScorer struct{}

func NewViralityScorer() *ViralityScorer {
	return &ViralityScorer{}
}

// CalculateScore — Artık sadece loglama için, filtre kararı vermiyor
func (v *ViralityScorer) CalculateScore(title, content string, category models.NewsCategory) int {
	text := strings.ToLower(title + " " + content)
	score := CATEGORY_BASE_SCORE[category]

	// Kategori bazlı keyword seti seç
	if category == models.CategoryBreaking {
		for kw, pts := range IMPORTANCE_KEYWORDS {
			if strings.Contains(text, kw) {
				score += pts
			}
		}
	} else {
		for kw, pts := range ENGAGEMENT_KEYWORDS {
			if strings.Contains(text, kw) {
				score += pts
			}
		}
		for kw, pts := range ANTI_ENGAGEMENT_KEYWORDS {
			if strings.Contains(text, kw) {
				score += pts
			}
		}
	}

	// Türkiye bonusu her iki kategoride de geçerli
	turkeyKws := []string{"türkiye", "turkey", "istanbul", "ankara", "türk", "lira"}
	for _, kw := range turkeyKws {
		if strings.Contains(text, kw) {
			score += 15
			break
		}
	}

	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}
	return score
}

func (v *ViralityScorer) GetViralityLevel(score int) string {
	switch {
	case score >= 80:
		return "ULTRA_VIRAL"
	case score >= 60:
		return "HIGH_VIRAL"
	case score >= 40:
		return "MEDIUM_VIRAL"
	case score >= 20:
		return "LOW_VIRAL"
	default:
		return "NOT_VIRAL"
	}
}

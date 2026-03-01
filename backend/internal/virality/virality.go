package virality

import (
	"strings"

	"github.com/SMutaf/twitter-bot/backend/internal/models"
)

// Viral potansiyeli yüksek kelimeler ve skorları
var VIRAL_KEYWORDS = map[string]int{
	// Çok yüksek viral potansiyel (50+ puan)
	"öldü":              50,
	"ölü":               50,
	"hayatını kaybetti": 50,
	"savaş":             50,
	"saldırı":           50,
	"deprem":            50,
	"füze":              45,
	"nükleer":           45,
	"terör":             45,

	// İngilizce karşılıklar
	"died":          50,
	"death":         50,
	"lost his life": 50,
	"war":           50,
	"attack":        50,
	"earthquake":    50,
	"missile":       45,
	"nuclear":       45,
	"terror":        45,

	// Yüksek viral potansiyel (30-40 puan)
	"erdoğan":       40,
	"cumhurbaşkanı": 40,
	"faiz":          35,
	"enflasyon":     35,
	"dolar":         35,
	"bitcoin":       35,
	"kripto":        30,
	"iphone":        30,
	"tesla":         30,
	"elon musk":     30,
	"openai":        30,
	"chatgpt":       30,

	// İngilizce karşılıklar
	"president": 35,
	"inflation": 35,
	"dollar":    35,
	"crypto":    30,
	"stock":     30,
}

// Negatif viral (puan azaltır)
var ANTI_VIRAL_KEYWORDS = map[string]int{
	"top 10":   -20,
	"top 5":    -20,
	"listicle": -20,
	"tutorial": -15,
	"nasıl":    -10,
	"how to":   -10,
	"review":   -10,
	"inceleme": -10,
}

// Kategori bazlı bonus puanlar
var CATEGORY_BASE_SCORE = map[models.NewsCategory]int{
	models.CategoryBreaking: 40, // Breaking news otomatik yüksek skor
	models.CategoryTech:     15,
	models.CategoryGeneral:  20,
	models.CategoryEconomy:  20,
	models.CategorySports:   15,
	models.CategoryScience:  15,
}

type ViralityScorer struct{}

func NewViralityScorer() *ViralityScorer {
	return &ViralityScorer{}
}

// CalculateScore haberin viral potansiyelini 0-100 arası skorlar
func (v *ViralityScorer) CalculateScore(title, content string, category models.NewsCategory) int {
	text := strings.ToLower(title + " " + content)

	// Base score (kategori bazlı)
	score := CATEGORY_BASE_SCORE[category]

	// Viral keyword'leri ara ve puan ekle
	for keyword, points := range VIRAL_KEYWORDS {
		if strings.Contains(text, keyword) {
			score += points
		}
	}

	// Anti-viral keyword'leri ara ve puan çıkar
	for keyword, points := range ANTI_VIRAL_KEYWORDS {
		if strings.Contains(text, keyword) {
			score += points // points zaten negatif
		}
	}

	// Türkiye/İngiltere/ABD gibi region keywords bonus
	relevantKeywords := []string{"türkiye", "turkey", "istanbul", "ankara", "türk", "lira", "usa", "america", "london", "uk", "england"}
	for _, keyword := range relevantKeywords {
		if strings.Contains(text, keyword) {
			score += 15
			break
		}
	}

	// Emoji/punctuation bonusu
	if strings.Contains(text, "!") {
		score += 5
	}
	if strings.Contains(text, "?") {
		score += 3
	}

	// Score'u 0-100 arası sınırla
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return score
}

// GetViralityLevel score'a göre seviye döndür
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

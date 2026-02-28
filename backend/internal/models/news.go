package models

import "time"

// NewsCategory haberin türünü belirtir
type NewsCategory string

const (
	CategoryBreaking NewsCategory = "BREAKING" // Son dakika - kısa, çarpıcı
	CategoryTech     NewsCategory = "TECH"     // Teknoloji haberi - bilgilendirici
	CategoryGeneral  NewsCategory = "GENERAL"  // Genel haber
	CategoryEconomy  NewsCategory = "ECONOMY"  // Ekonomi & finans
	CategorySports   NewsCategory = "SPORTS"   // Spor haberleri
	CategoryScience  NewsCategory = "SCIENCE"  // Bilim & uzay
)

type NewsItem struct {
	Title       string
	Description string
	Link        string
	Source      string
	Category    NewsCategory // AI bu alana göre prompt seçecek
	PublishedAt time.Time
}

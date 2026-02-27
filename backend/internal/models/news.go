package models

// NewsCategory haberin türünü belirtir
type NewsCategory string

const (
	CategoryBreaking NewsCategory = "BREAKING" // Son dakika - kısa, çarpıcı
	CategoryTech     NewsCategory = "TECH"     // Teknoloji haberi - bilgilendirici
	CategoryGeneral  NewsCategory = "GENERAL"  // Genel haber
)

type NewsItem struct {
	Title       string
	Description string
	Link        string
	Source      string
	Category    NewsCategory // AI bu alana göre prompt seçecek
}

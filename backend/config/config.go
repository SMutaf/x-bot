package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/models"
	"github.com/joho/godotenv"
)

// RSSSource her RSS kaynağının ayarlarını tutar
type RSSSource struct {
	URL      string
	Category models.NewsCategory
	Interval time.Duration // Bu kaynağın tarama sıklığı
}

type Config struct {
	RSSSources       []RSSSource
	RedisAddr        string
	TelegramToken    string
	TelegramChatID   int64
	MaxNewsPerSource int
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env dosyası bulunamadı, sistem değişkenleri kullanılacak.")
	}

	chatID, _ := strconv.ParseInt(os.Getenv("TELEGRAM_CHAT_ID"), 10, 64)

	return &Config{
		RSSSources: []RSSSource{
			// ========================================
			// ULUSLARARASI HABER AJANSLARI (BREAKING)
			// ========================================
			// En hızlı breaking news kaynakları - Her 2 dakikada bir tara

			{
				URL:      "https://feeds.bbci.co.uk/news/world/rss.xml",
				Category: models.CategoryBreaking,
				Interval: 2 * time.Minute,
			},
			{
				URL:      "https://www.aa.com.tr/tr/rss/default?cat=guncel",
				Category: models.CategoryBreaking,
				Interval: 2 * time.Minute,
			},
			{
				URL:      "https://www.aljazeera.com/xml/rss/all.xml",
				Category: models.CategoryBreaking,
				Interval: 2 * time.Minute,
			},

			// ========================================
			// TEKNOLOJİ HABERLERİ (NORMAL)
			// ========================================
			// Teknoloji odaklı kaynaklar - Her 10 dakikada bir tara

			{
				URL:      "https://www.webtekno.com/rss.xml",
				Category: models.CategoryTech,
				Interval: 10 * time.Minute,
			},
			{
				URL:      "https://shiftdelete.net/feed",
				Category: models.CategoryTech,
				Interval: 10 * time.Minute,
			},
			{
				URL:      "https://www.chip.com.tr/rss",
				Category: models.CategoryTech,
				Interval: 10 * time.Minute,
			},
			{
				URL:      "https://www.technopat.net/feed/",
				Category: models.CategoryTech,
				Interval: 10 * time.Minute,
			},
			{
				URL:      "https://feeds.feedburner.com/TechCrunch/",
				Category: models.CategoryTech,
				Interval: 10 * time.Minute,
			},
			{
				URL:      "https://www.theverge.com/rss/index.xml",
				Category: models.CategoryTech,
				Interval: 10 * time.Minute,
			},

			// ========================================
			// TÜRKİYE GÜNDEM (GENEL)
			// ========================================
			// Türkiye gündem haberleri - Her 5 dakikada bir tara

			{
				URL:      "https://www.aa.com.tr/tr/rss/default?cat=ekonomi",
				Category: models.CategoryGeneral,
				Interval: 5 * time.Minute,
			},
			{
				URL:      "https://www.aa.com.tr/tr/rss/default?cat=turkiye",
				Category: models.CategoryGeneral,
				Interval: 5 * time.Minute,
			},
			{
				URL:      "https://www.hurriyet.com.tr/rss/anasayfa",
				Category: models.CategoryGeneral,
				Interval: 5 * time.Minute,
			},

			// ========================================
			// EKONOMİ & FİNANS (GENEL)
			// ========================================
			// Ekonomi, finans, kripto - Her 5 dakikada bir tara

			// burası eksik KAYNAK BULUNMASI LAZIM
		},
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		TelegramToken:    os.Getenv("TELEGRAM_TOKEN"),
		TelegramChatID:   chatID,
		MaxNewsPerSource: 5, // Artırıldı (3 → 5)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

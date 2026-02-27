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
			// --- SON DAKİKA KAYNAKLARI (5 dakikada bir taranır) ---
			{
				URL:      "https://www.webtekno.com/rss.xml",
				Category: models.CategoryBreaking,
				Interval: 5 * time.Minute,
			},
			{
				URL:      "https://shiftdelete.net/feed",
				Category: models.CategoryBreaking,
				Interval: 5 * time.Minute,
			},
			// --- NORMAL TEKNOLOJİ KAYNAKLARI (15 dakikada bir taranır) ---
			{
				URL:      "https://www.chip.com.tr/rss",
				Category: models.CategoryTech,
				Interval: 15 * time.Minute,
			},
			{
				URL:      "https://www.technopat.net/feed/",
				Category: models.CategoryTech,
				Interval: 15 * time.Minute,
			},
			{
				URL:      "https://feeds.feedburner.com/TechCrunch/",
				Category: models.CategoryGeneral,
				Interval: 15 * time.Minute,
			},
		},
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		TelegramToken:    os.Getenv("TELEGRAM_TOKEN"),
		TelegramChatID:   chatID,
		MaxNewsPerSource: 3,
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

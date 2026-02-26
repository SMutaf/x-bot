package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	RSSUrls          []string
	RedisAddr        string
	TelegramToken    string
	TelegramChatID   int64
	MaxNewsPerSource int // Her RSS kaynağından tek turda alınacak max haber sayısı
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env dosyası bulunamadı, sistem değişkenleri kullanılacak.")
	}

	chatID, _ := strconv.ParseInt(os.Getenv("TELEGRAM_CHAT_ID"), 10, 64)

	return &Config{
		RSSUrls: []string{
			"https://feeds.feedburner.com/TechCrunch/",
			"https://news.ycombinator.com/rss",
			"https://openai.com/blog/rss.xml",
			"https://feeds.bbci.co.uk/news/technology/rss.xml",
		},
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		TelegramToken:    os.Getenv("TELEGRAM_TOKEN"),
		TelegramChatID:   chatID,
		MaxNewsPerSource: 3, // Kaynak başına max 3 haber (4 kaynak × 3 = max 12 haber/tur)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

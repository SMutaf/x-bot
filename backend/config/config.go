package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/models"
	"github.com/joho/godotenv"
)

type RSSSource struct {
	URL      string
	Category models.NewsCategory
	Interval time.Duration
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
			{
				URL:      "https://feeds.bbci.co.uk/news/world/rss.xml",
				Category: models.CategoryBreaking,
				Interval: 2 * time.Minute,
			},
			{
				URL:      "https://rss.nytimes.com/services/xml/rss/nyt/World.xml",
				Category: models.CategoryBreaking,
				Interval: 2 * time.Minute,
			},
			{
				URL:      "https://feeds.npr.org/1001/rss.xml",
				Category: models.CategoryBreaking,
				Interval: 2 * time.Minute,
			},
			{
				URL:      "https://www.aljazeera.com/xml/rss/all.xml",
				Category: models.CategoryBreaking,
				Interval: 2 * time.Minute,
			},
			{
				URL:      "https://www.theguardian.com/world/rss",
				Category: models.CategoryBreaking,
				Interval: 3 * time.Minute,
			},
			{
				URL:      "https://rss.cnn.com/rss/edition.rss",
				Category: models.CategoryBreaking,
				Interval: 3 * time.Minute,
			},
			{
				URL:      "https://feeds.bloomberg.com/markets/news.rss",
				Category: models.CategoryEconomy,
				Interval: 5 * time.Minute,
			},
			{
				URL:      "https://www.ft.com/rss/home",
				Category: models.CategoryEconomy,
				Interval: 5 * time.Minute,
			},
			{
				URL:      "https://www.cnbc.com/id/100003114/device/rss/rss.html",
				Category: models.CategoryEconomy,
				Interval: 5 * time.Minute,
			},
			{
				URL:      "https://feeds.marketwatch.com/marketwatch/topstories",
				Category: models.CategoryEconomy,
				Interval: 5 * time.Minute,
			},
			{
				URL:      "https://www.webtekno.com/rss.xml",
				Category: models.CategoryTech,
				Interval: 10 * time.Minute,
			},
			{
				URL:      "https://techcrunch.com/feed/",
				Category: models.CategoryTech,
				Interval: 10 * time.Minute,
			},
			{
				URL:      "https://www.theverge.com/rss/index.xml",
				Category: models.CategoryTech,
				Interval: 10 * time.Minute,
			},
			{
				URL:      "https://feeds.arstechnica.com/arstechnica/index",
				Category: models.CategoryTech,
				Interval: 10 * time.Minute,
			},
			{
				URL:      "https://www.aa.com.tr/tr/rss/default?cat=guncel",
				Category: models.CategoryGeneral,
				Interval: 3 * time.Minute,
			},
			{
				URL:      "https://www.trthaber.com/sondakika.rss",
				Category: models.CategoryGeneral,
				Interval: 3 * time.Minute,
			},
			{
				URL:      "https://www.bloomberght.com/rss",
				Category: models.CategoryGeneral,
				Interval: 5 * time.Minute,
			},
		},
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		TelegramToken:    os.Getenv("TELEGRAM_TOKEN"),
		TelegramChatID:   chatID,
		MaxNewsPerSource: 8,
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

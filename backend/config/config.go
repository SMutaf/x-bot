package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/domain/models"
	"github.com/joho/godotenv"
)

type RSSSource struct {
	URL      string              `json:"url"`
	Category models.NewsCategory `json:"category"`
	Interval time.Duration       `json:"interval"`
}

type rawRSSSource struct {
	URL      string              `json:"url"`
	Category models.NewsCategory `json:"category"`
	Interval string              `json:"interval"`
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

	sources, err := loadSources(getEnv("SOURCES_FILE", "config/sources.json"))
	if err != nil {
		log.Fatalf("RSS kaynakları yüklenemedi: %v", err)
	}

	return &Config{
		RSSSources:       sources,
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		TelegramToken:    os.Getenv("TELEGRAM_TOKEN"),
		TelegramChatID:   chatID,
		MaxNewsPerSource: getEnvAsInt("MAX_NEWS_PER_SOURCE", 8),
	}
}

func loadSources(path string) ([]RSSSource, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("sources dosyası okunamadı (%s): %w", path, err)
	}

	var rawSources []rawRSSSource
	if err := json.Unmarshal(data, &rawSources); err != nil {
		return nil, fmt.Errorf("sources dosyası parse edilemedi (%s): %w", path, err)
	}

	if len(rawSources) == 0 {
		return nil, fmt.Errorf("sources dosyası boş (%s)", path)
	}

	sources := make([]RSSSource, 0, len(rawSources))

	for i, raw := range rawSources {
		if raw.URL == "" {
			return nil, fmt.Errorf("sources[%d] için url boş", i)
		}
		if raw.Category == "" {
			return nil, fmt.Errorf("sources[%d] için category boş", i)
		}
		if raw.Interval == "" {
			return nil, fmt.Errorf("sources[%d] için interval boş", i)
		}

		interval, err := time.ParseDuration(raw.Interval)
		if err != nil {
			return nil, fmt.Errorf("sources[%d] interval parse edilemedi (%s): %w", i, raw.Interval, err)
		}

		sources = append(sources, RSSSource{
			URL:      raw.URL,
			Category: raw.Category,
			Interval: interval,
		})
	}

	log.Printf("Toplam %d RSS kaynağı yüklendi: %s", len(sources), path)
	return sources, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("%s parse edilemedi (%s), fallback kullanılacak: %d", key, value, fallback)
		return fallback
	}

	return parsed
}

func (r RSSSource) String() string {
	return fmt.Sprintf("[%s | %s] %s", r.Category, r.Interval, r.URL)
}

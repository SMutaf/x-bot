package main

import (
	"context"
	"fmt"
	"time"

	"github.com/SMutaf/twitter-bot/backend/config"
	"github.com/SMutaf/twitter-bot/backend/internal/ai"
	"github.com/SMutaf/twitter-bot/backend/internal/dedup"
	"github.com/SMutaf/twitter-bot/backend/internal/filter"
	"github.com/SMutaf/twitter-bot/backend/internal/middleware"
	"github.com/SMutaf/twitter-bot/backend/internal/models"
	"github.com/SMutaf/twitter-bot/backend/internal/scraper"
	"github.com/SMutaf/twitter-bot/backend/internal/telegram"
	"github.com/SMutaf/twitter-bot/backend/internal/virality"
	"golang.org/x/time/rate"
)

func main() {
	fmt.Println("Twitter Bot Backend Başlatılıyor.")

	cfg := config.LoadConfig()

	cache := dedup.NewDeduplicator(cfg.RedisAddr)
	cache.Client.FlushAll(cache.Ctx)
	fmt.Println("Redis Hafızası TEMİZLENDİ!")

	aiClient := ai.NewClient("http://localhost:8000")
	tgBot := telegram.NewApprovalBot(cfg.TelegramToken, cfg.TelegramChatID)

	go tgBot.ListenForApproval()
	fmt.Println("Telegram Onay Servisi Aktif!")

	breakingChannel := make(chan models.NewsItem, 100)
	normalChannel := make(chan models.NewsItem, 200)

	sc := scraper.NewRSSScraper(cache, breakingChannel, normalChannel, cfg.MaxNewsPerSource)

	viralityScorer := virality.NewViralityScorer()
	newsFilter := filter.NewNewsFilter(viralityScorer, 30) // 30 altındaki skorları geçirme

	limiter := rate.NewLimiter(rate.Every(3*time.Second), 1)

	go func() {
		for {
			select {
			case item := <-breakingChannel:
				limiter.Wait(context.Background())
				middleware.RecoveryWrapper("Breaking News Worker", func() {
					processNews(item, aiClient, tgBot, newsFilter, viralityScorer)
				})
				continue
			default:
			}

			select {
			case item := <-breakingChannel:
				limiter.Wait(context.Background())
				middleware.RecoveryWrapper("Breaking News Worker", func() {
					processNews(item, aiClient, tgBot, newsFilter, viralityScorer)
				})
			case item := <-normalChannel:
				limiter.Wait(context.Background())
				middleware.RecoveryWrapper("Normal News Worker", func() {
					processNews(item, aiClient, tgBot, newsFilter, viralityScorer)
				})
			case <-time.After(100 * time.Millisecond):
				continue
			}
		}
	}()

	fmt.Println("Priority Worker Başlatıldı! (Breaking > Normal)")

	for _, source := range cfg.RSSSources {
		src := source
		go func() {
			fmt.Printf("Kaynak başlatıldı [%s | %s]: %s\n", src.Category, src.Interval, src.URL)
			for {
				middleware.RecoveryWrapper("Tarama", func() {
					sc.Fetch(src)
				})
				time.Sleep(src.Interval)
			}
		}()
	}

	fmt.Println("Tüm kaynaklar aktif. Bot çalışıyor...")
	select {}
}

func processNews(item models.NewsItem, aiClient *ai.Client, tgBot *telegram.ApprovalBot, nf *filter.NewsFilter, vs *virality.ViralityScorer) {
	// Virality Skoru hesapla
	score := vs.CalculateScore(item.Title, item.Description, item.Category)
	level := vs.GetViralityLevel(score)
	fmt.Printf("[%s] Skor: %d (%s) | %s\n", item.Category, score, level, item.Title)

	// Filter kontrolü
	if !nf.IsAllowed(item.Title, item.Description) {
		fmt.Printf("Haber filtrelendi (skor düşük veya anti-viral): %s\n", item.Title)
		return
	}

	// Saat bazlı gönderim (BREAKING her zaman, diğerleri belirli saatlerde)
	now := time.Now()
	hour := now.Hour()
	if item.Category == models.CategoryTech || item.Category == models.CategoryGeneral {
		if !((hour >= 8 && hour < 10) || (hour >= 12 && hour < 14) || (hour >= 18 && hour <= 21)) {
			fmt.Printf("⏳ Haber saat filtresine takıldı, gönderilmiyor: %s\n", item.Title)
			return
		}
	}

	// Yayınlanma zamanı etiketi
	publishedTime := ""
	if !item.PublishedAt.IsZero() {
		diff := time.Since(item.PublishedAt)
		switch {
		case diff < 5*time.Minute:
			publishedTime = "🔴 ŞU AN"
		case diff < 30*time.Minute:
			publishedTime = fmt.Sprintf("%d dk önce", int(diff.Minutes()))
		case diff < 2*time.Hour:
			publishedTime = fmt.Sprintf("%d saat önce", int(diff.Hours()))
		default:
			publishedTime = item.PublishedAt.Format("15:04")
		}
	}

	response, err := aiClient.GenerateTweet(item.Title, item.Description, item.Link, item.Source, string(item.Category), item.PublishedAt)
	if err != nil {
		fmt.Printf("AI Hatası (%s): %v\n", item.Title, err)
		return
	}

	if response.Tweet == "" {
		fmt.Printf("AI boş tweet döndü: %s\n", item.Title)
		return
	}

	fmt.Printf("AI cevap aldı - Tweet: %s... | Reply: %s...\n",
		response.Tweet[:min(30, len(response.Tweet))],
		response.Reply[:min(30, len(response.Reply))])

	err = tgBot.RequestApproval(response.Tweet, response.Reply, item.Source, string(item.Category), publishedTime)
	if err != nil {
		fmt.Printf("Telegram Hatası: %v\n", err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/SMutaf/twitter-bot/backend/config"
	"github.com/SMutaf/twitter-bot/backend/internal/ai"
	"github.com/SMutaf/twitter-bot/backend/internal/dedup"
	"github.com/SMutaf/twitter-bot/backend/internal/middleware"
	"github.com/SMutaf/twitter-bot/backend/internal/models"
	"github.com/SMutaf/twitter-bot/backend/internal/scraper"
	"github.com/SMutaf/twitter-bot/backend/internal/telegram"
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

	// İki ayrı kanal: BREAKING için öncelikli, diğerleri için normal
	// Buffer size artırıldı (daha fazla kaynak için)
	breakingChannel := make(chan models.NewsItem, 100) // 50 → 100
	normalChannel := make(chan models.NewsItem, 200)   // 100 → 200

	sc := scraper.NewRSSScraper(cache, breakingChannel, normalChannel, cfg.MaxNewsPerSource)

	// Rate limiter: 3 saniyede 1 istek
	limiter := rate.NewLimiter(rate.Every(3*time.Second), 1)

	// Priority Worker: BREAKING haberleri MUTLAKA öncelikli işlenir
	go func() {
		for {
			// ÖNCE breaking kanalını non-blocking kontrol et
			select {
			case item := <-breakingChannel:
				limiter.Wait(context.Background())
				middleware.RecoveryWrapper("Breaking News Worker", func() {
					processNews(item, aiClient, tgBot)
				})
				continue // Döngünün başına dön, tekrar breaking kontrol et
			default:
				// Breaking kanalda bir şey yok, normal kanala bak
			}

			// Breaking yoksa normal kanala bak
			select {
			case item := <-breakingChannel:
				// Normal kanalı beklerken breaking geldi, onu önceliklendir
				limiter.Wait(context.Background())
				middleware.RecoveryWrapper("Breaking News Worker", func() {
					processNews(item, aiClient, tgBot)
				})

			case item := <-normalChannel:
				limiter.Wait(context.Background())
				middleware.RecoveryWrapper("Normal News Worker", func() {
					processNews(item, aiClient, tgBot)
				})

			case <-time.After(100 * time.Millisecond):
				// Kısa süre bekle, CPU'yu meşgul etme
				continue
			}
		}
	}()

	fmt.Println("Priority Worker Başlatıldı! (Breaking > Normal)")

	// Her kaynak için ayrı goroutine başlatıyoruz
	for _, source := range cfg.RSSSources {
		src := source // closure için kopyala
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

	// Ana goroutine'i canlı tut
	select {}
}

func processNews(item models.NewsItem, aiClient *ai.Client, tgBot *telegram.ApprovalBot) {
	//  Yayınlanma saatini hesapla (eğer varsa)
	publishedTime := ""
	if !item.PublishedAt.IsZero() {
		now := time.Now()
		diff := now.Sub(item.PublishedAt)

		if diff < 5*time.Minute {
			publishedTime = "🔴 ŞU AN" // Çok yeni
		} else if diff < 30*time.Minute {
			publishedTime = fmt.Sprintf("%d dk önce", int(diff.Minutes()))
		} else if diff < 2*time.Hour {
			publishedTime = fmt.Sprintf("%d saat önce", int(diff.Hours()))
		} else {
			publishedTime = item.PublishedAt.Format("15:04")
		}
	}

	fmt.Printf("[%s] İşleniyor (%s): %s\n", item.Category, publishedTime, item.Title)

	response, err := aiClient.GenerateTweet(item.Title, item.Description, item.Link, item.Source, string(item.Category), item.PublishedAt)
	if err != nil {
		fmt.Printf("AI Hatası (%s): %v\n", item.Title, err)
		return
	}

	// AI response'unu kontrol et
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

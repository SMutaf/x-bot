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

	// Telegram onay dinleyicisini asenkron başlatıyoruz
	go tgBot.ListenForApproval()
	fmt.Println("Telegram Onay Servisi Aktif!")

	// 1. Haber Kanalı Oluşturma (Buffer size: 100)
	newsChannel := make(chan models.NewsItem, 100)

	// 2. Scraper Tanımlama
	sc := scraper.NewRSSScraper(cache, newsChannel, cfg.MaxNewsPerSource)

	// 3. Rate Limiter: Dakikada max 20 istek (Gemini limitinin altında güvenli tampon)
	// Her istek arası en az 3 saniye bekler
	limiter := rate.NewLimiter(rate.Every(3*time.Second), 1)

	// 4. Worker Başlatma
	go func() {
		for item := range newsChannel {
			// Rate limiter: token hazır olana kadar bekler, sonra devam eder
			limiter.Wait(context.Background())

			middleware.RecoveryWrapper("Worker İşlemi", func() {
				processNews(item, aiClient, tgBot)
			})
		}
	}()

	fmt.Println("Asenkron Worker Başlatıldı! (Rate Limit: 3sn/istek)")
	fmt.Println("Bot Sürekli Tarama Moduna Geçiyor...")

	for {
		middleware.RecoveryWrapper("Tarama Turu", func() {
			fmt.Println("\n--- Yeni Tarama Turu Başlıyor ---")

			for _, url := range cfg.RSSUrls {
				fmt.Printf(">> Kaynak Taranıyor: %s\n", url)
				sc.Fetch(url)
				time.Sleep(2 * time.Second) // Kaynaklar arası kısa bekleme
			}
		})

		fmt.Println("Bu tur bitti. 15 dakika dinleniliyor...")
		time.Sleep(15 * time.Minute)
	}
}

// processNews tekil bir haberin AI ve Telegram süreçlerini yönetir
func processNews(item models.NewsItem, aiClient *ai.Client, tgBot *telegram.ApprovalBot) {
	fmt.Printf("İşleniyor: %s\n", item.Title)

	response, err := aiClient.GenerateTweet(item.Title, item.Description, item.Link, item.Source)
	if err != nil {
		fmt.Printf("AI Hatası (%s): %v\n", item.Title, err)
		return
	}

	err = tgBot.RequestApproval(response.Tweet, response.Reply, item.Source)
	if err != nil {
		fmt.Printf("Telegram Hatası: %v\n", err)
	}
}

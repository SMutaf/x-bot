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

	newsChannel := make(chan models.NewsItem, 100)
	sc := scraper.NewRSSScraper(cache, newsChannel, cfg.MaxNewsPerSource)

	// Rate limiter: 3 saniyede 1 istek → dakikada max 20 (Gemini limitinin altında)
	limiter := rate.NewLimiter(rate.Every(3*time.Second), 1)

	// Worker
	go func() {
		for item := range newsChannel {
			limiter.Wait(context.Background())
			middleware.RecoveryWrapper("Worker İşlemi", func() {
				processNews(item, aiClient, tgBot)
			})
		}
	}()

	fmt.Println("Worker Başlatıldı! (Rate Limit: 3sn/istek)")

	// Her kaynak için ayrı goroutine başlatıyoruz
	// Her kaynak kendi Interval'ına göre çalışır
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
	fmt.Printf("[%s] İşleniyor: %s\n", item.Category, item.Title)

	response, err := aiClient.GenerateTweet(item.Title, item.Description, item.Link, item.Source, string(item.Category))
	if err != nil {
		fmt.Printf("AI Hatası (%s): %v\n", item.Title, err)
		return
	}

	err = tgBot.RequestApproval(response.Tweet, response.Reply, item.Source, string(item.Category))
	if err != nil {
		fmt.Printf("Telegram Hatası: %v\n", err)
	}
}

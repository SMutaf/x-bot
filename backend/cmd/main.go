package main

import (
	"fmt"
	"sync"

	"github.com/SMutaf/twitter-bot/backend/config"
	"github.com/SMutaf/twitter-bot/backend/internal/ai"
	"github.com/SMutaf/twitter-bot/backend/internal/dedup"
	"github.com/SMutaf/twitter-bot/backend/internal/scraper"
	"github.com/SMutaf/twitter-bot/backend/internal/telegram"
)

func main() {
	fmt.Println("Twitter Bot Backend Başlatılıyor...")

	cfg := config.LoadConfig()

	// 1. Redis (Hafıza)
	cache := dedup.NewDeduplicator(cfg.RedisAddr)
	fmt.Println("Redis Hafızası Devrede!")

	// 2. AI İstemcisi (İletişim)
	aiClient := ai.NewClient("http://localhost:8000")
	fmt.Println("AI Servisine Bağlanıldı!")

	// 3. Telegram Onay Botu (YENİ)
	tgBot := telegram.NewApprovalBot(cfg.TelegramToken, cfg.TelegramChatID)
	go tgBot.ListenForApproval()
	fmt.Println("Telegram Onay Servisi Aktif!")

	// 4. Scraper (Redis + AI + Telegram)
	// Parametre sayısını 3'e çıkardık:
	sc := scraper.NewRSSScraper(cache, aiClient, tgBot)

	// 5. Tarama Başlasın
	var wg sync.WaitGroup
	for _, url := range cfg.RSSUrls {
		wg.Add(1)
		go func(targetUrl string) {
			defer wg.Done()
			sc.Fetch(targetUrl)
		}(url)
	}

	wg.Wait()
	fmt.Println("Tüm işlemler tamamlandı.")
}

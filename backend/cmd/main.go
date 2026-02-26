package main

import (
	"fmt"
	"time"

	"github.com/SMutaf/twitter-bot/backend/config"
	"github.com/SMutaf/twitter-bot/backend/internal/ai"
	"github.com/SMutaf/twitter-bot/backend/internal/dedup"
	"github.com/SMutaf/twitter-bot/backend/internal/middleware"
	"github.com/SMutaf/twitter-bot/backend/internal/scraper"
	"github.com/SMutaf/twitter-bot/backend/internal/telegram"
)

func main() {
	fmt.Println("Twitter Bot Backend Başlatılıyor (SIRALI MOD)...")

	cfg := config.LoadConfig()

	cache := dedup.NewDeduplicator(cfg.RedisAddr)
	// Test aşamasında her şeyi yeni görmek için temizlik yapar
	cache.Client.FlushAll(cache.Ctx)
	fmt.Println("Redis Hafızası TEMİZLENDİ!")
	fmt.Println("Redis Hafızası Devrede!")

	aiClient := ai.NewClient("http://localhost:8000")
	fmt.Println("AI Servisine Bağlanıldı!")

	tgBot := telegram.NewApprovalBot(cfg.TelegramToken, cfg.TelegramChatID)
	go tgBot.ListenForApproval()
	fmt.Println("Telegram Onay Servisi Aktif!")

	sc := scraper.NewRSSScraper(cache, aiClient, tgBot)

	fmt.Println("Bot Sürekli Tarama Moduna Geçiyor...")

	for {
		// RecoveryWrapper sayesinde tarama turundaki hatalar programı durdurmaz
		middleware.RecoveryWrapper("Tarama Turu", func() {
			fmt.Println("\n--- Yeni Tarama Turu Başlıyor ---")

			for _, url := range cfg.RSSUrls {
				fmt.Printf(">> Kaynak Taranıyor: %s\n", url)
				sc.Fetch(url)
				fmt.Println("Diğer kaynağa geçmeden 5 saniye bekleniyor...")
				time.Sleep(5 * time.Second)
			}
		})

		fmt.Println("Bu tur bitti. 15 dakika dinleniliyor...")
		time.Sleep(15 * time.Minute)
	}
}

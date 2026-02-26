package main

import (
	"fmt"
	"time"

	"github.com/SMutaf/twitter-bot/backend/config"
	"github.com/SMutaf/twitter-bot/backend/internal/ai"
	"github.com/SMutaf/twitter-bot/backend/internal/dedup"
	"github.com/SMutaf/twitter-bot/backend/internal/scraper"
	"github.com/SMutaf/twitter-bot/backend/internal/telegram"
)

func main() {
	fmt.Println("🚀 Twitter Bot Backend Başlatılıyor (SIRALI MOD)...")

	cfg := config.LoadConfig()

	cache := dedup.NewDeduplicator(cfg.RedisAddr) // silincek

	cache.Client.FlushAll(cache.Ctx)
	fmt.Println("Redis Hafızası TEMİZLENDİ! (Tüm haberler yeni sayılacak)")

	fmt.Println("Redis Hafızası Devrede!")

	// 2. AI İstemcisi
	aiClient := ai.NewClient("http://localhost:8000")
	fmt.Println("AI Servisine Bağlanıldı!")

	// 3. Telegram Botu
	tgBot := telegram.NewApprovalBot(cfg.TelegramToken, cfg.TelegramChatID)
	go tgBot.ListenForApproval()
	fmt.Println("Telegram Onay Servisi Aktif!")

	// 4. Scraper
	sc := scraper.NewRSSScraper(cache, aiClient, tgBot)

	fmt.Println("Bot Sürekli Tarama Moduna Geçiyor...")

	// --- SONSUZ DÖNGÜ ---
	for {
		fmt.Println("\n--- Yeni Tarama Turu Başlıyor ---")

		// DİKKAT: "go func" ve "WaitGroup" YOK.
		// Kaynakları tek tek, sırayla tarıyoruz.
		for _, url := range cfg.RSSUrls {
			fmt.Printf(">> Kaynak Taranıyor: %s\n", url)
			sc.Fetch(url)

			// Her kaynak arasında 5 saniye nefes alıyoruz
			fmt.Println("Diğer kaynağa geçmeden 5 saniye bekleniyor...")
			time.Sleep(5 * time.Second)
		}

		fmt.Println("Bu tur bitti. 15 dakika dinleniliyor...")
		time.Sleep(15 * time.Minute)
	}
}

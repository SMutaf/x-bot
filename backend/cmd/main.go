package main

import (
	"fmt"
	"sync"

	"github.com/SMutaf/twitter-bot/backend/config"
	"github.com/SMutaf/twitter-bot/backend/internal/ai"
	"github.com/SMutaf/twitter-bot/backend/internal/dedup"
	"github.com/SMutaf/twitter-bot/backend/internal/scraper"
)

func main() {
	fmt.Println("Twitter Bot Backend Başlatılıyor...")

	cfg := config.LoadConfig()

	// 1. Redis (Hafıza)
	cache := dedup.NewDeduplicator("localhost:6379")
	fmt.Println("Redis Hafızası Devrede!")

	// 2. AI İstemcisi (İletişim)
	// Python servisinin adresi: http://localhost:8000
	aiClient := ai.NewClient("http://localhost:8000")
	fmt.Println("AI Servisine Bağlanıldı!")

	// 3. Scraper (Redis + AI)
	sc := scraper.NewRSSScraper(cache, aiClient)

	// 4. Tarama Başlasın
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

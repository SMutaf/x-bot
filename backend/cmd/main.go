package main

import (
	"fmt"
	"sync"

	"github.com/SMutaf/twitter-bot/backend/config"
	"github.com/SMutaf/twitter-bot/backend/internal/dedup"
	"github.com/SMutaf/twitter-bot/backend/internal/scraper"
)

func main() {
	fmt.Println("Twitter Bot Backend Başlatılıyor...")

	// 1. Ayarları Yükle
	cfg := config.LoadConfig()

	// 2. Redis Bağlantısını Başlat (Hafıza)
	cache := dedup.NewDeduplicator("localhost:6379")
	fmt.Println("Redis Hafızası Devrede!")

	// 3. Scraper'ı Başlat (Hafızayı içine veriyoruz)
	sc := scraper.NewRSSScraper(cache)

	// 4. Haberleri Tara
	var wg sync.WaitGroup
	for _, url := range cfg.RSSUrls {
		wg.Add(1)
		go func(targetUrl string) {
			defer wg.Done()
			sc.Fetch(targetUrl)
		}(url)
	}

	wg.Wait()
	fmt.Println("Tüm taramalar tamamlandı.")
}

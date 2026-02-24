package scraper

import (
	"context"
	"fmt"
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/dedup"
	"github.com/mmcdole/gofeed"
)

type RSSScraper struct {
	Parser *gofeed.Parser
	Cache  *dedup.Deduplicator // <--- Scaraper Hafızası
}

func NewRSSScraper(cache *dedup.Deduplicator) *RSSScraper {
	return &RSSScraper{
		Parser: gofeed.NewParser(),
		Cache:  cache,
	}
}

func (s *RSSScraper) Fetch(url string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	feed, err := s.Parser.ParseURLWithContext(url, ctx)
	if err != nil {
		fmt.Printf("Hata (%s): %v\n", url, err)
		return
	}

	fmt.Printf("Kaynak Taranıyor: %s\n", feed.Title)

	newItemsCount := 0

	for _, item := range feed.Items {
		// --- KRİTİK NOKTA: REDIS KONTROLÜ ---
		// Linki Redis'e soruyoruz. Eğer varsa (true dönerse) atlıyoruz.
		if s.Cache.IsDuplicate(item.Link) {
			continue
		}

		// Eğer buraya geldiyse haber YENİDİR!
		newItemsCount++
		fmt.Printf("YENİ HABER BULUNDU: %s\n", item.Title)
		fmt.Printf("Link: %s\n", item.Link)

		// Python servisi eklenicek
	}

	if newItemsCount == 0 {
		fmt.Println("Yeni içerik yok, hepsi eski.")
	}
	fmt.Println("--------------------------------------------------")
}

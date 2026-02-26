package scraper

import (
	"context"
	"fmt"
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/dedup"
	"github.com/SMutaf/twitter-bot/backend/internal/models"
	"github.com/mmcdole/gofeed"
)

type RSSScraper struct {
	Parser       *gofeed.Parser
	Cache        *dedup.Deduplicator
	Out          chan<- models.NewsItem
	MaxPerSource int // Kaynak başına max haber limiti
}

// NewRSSScraper artık maxPerSource parametresi de alıyor
func NewRSSScraper(cache *dedup.Deduplicator, out chan<- models.NewsItem, maxPerSource int) *RSSScraper {
	return &RSSScraper{
		Parser:       gofeed.NewParser(),
		Cache:        cache,
		Out:          out,
		MaxPerSource: maxPerSource,
	}
}

func (s *RSSScraper) Fetch(url string) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	feed, err := s.Parser.ParseURLWithContext(url, ctx)
	if err != nil {
		fmt.Printf("RSS Hatası (%s): %v\n", url, err)
		return
	}

	count := 0
	for _, item := range feed.Items {
		// Kaynak başına limit kontrolü
		if count >= s.MaxPerSource {
			fmt.Printf("Kaynak limiti doldu (%d/%d): %s\n", count, s.MaxPerSource, feed.Title)
			break
		}

		// 1. Link bazlı kontrol
		if s.Cache.IsDuplicate(item.Link) {
			continue
		}

		// 2. Başlık bazlı kontrol
		if s.Cache.IsTitleDuplicate(item.Title) {
			fmt.Printf("Benzer haber pas geçildi: %s\n", item.Title)
			continue
		}

		s.Out <- models.NewsItem{
			Title:       item.Title,
			Description: item.Description,
			Link:        item.Link,
			Source:      feed.Title,
		}

		fmt.Printf("Haber kanala gönderildi [%d/%d]: %s\n", count+1, s.MaxPerSource, item.Title)
		count++
	}
}

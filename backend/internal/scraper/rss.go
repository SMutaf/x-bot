package scraper

import (
	"context"
	"fmt"
	"time"

	"github.com/SMutaf/twitter-bot/backend/config"
	"github.com/SMutaf/twitter-bot/backend/internal/dedup"
	"github.com/SMutaf/twitter-bot/backend/internal/models"
	"github.com/mmcdole/gofeed"
)

type RSSScraper struct {
	Parser       *gofeed.Parser
	Cache        *dedup.Deduplicator
	Out          chan<- models.NewsItem
	MaxPerSource int
}

func NewRSSScraper(cache *dedup.Deduplicator, out chan<- models.NewsItem, maxPerSource int) *RSSScraper {
	return &RSSScraper{
		Parser:       gofeed.NewParser(),
		Cache:        cache,
		Out:          out,
		MaxPerSource: maxPerSource,
	}
}

// Fetch tek bir kaynağı tarar, kategoriyi NewsItem'a ekler
func (s *RSSScraper) Fetch(source config.RSSSource) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	feed, err := s.Parser.ParseURLWithContext(source.URL, ctx)
	if err != nil {
		fmt.Printf("RSS Hatası (%s): %v\n", source.URL, err)
		return
	}

	count := 0
	for _, item := range feed.Items {
		if count >= s.MaxPerSource {
			fmt.Printf("Kaynak limiti doldu (%d/%d): %s\n", count, s.MaxPerSource, feed.Title)
			break
		}

		if s.Cache.IsDuplicate(item.Link) {
			continue
		}

		if s.Cache.IsTitleDuplicate(item.Title) {
			fmt.Printf("Benzer haber pas geçildi: %s\n", item.Title)
			continue
		}

		s.Out <- models.NewsItem{
			Title:       item.Title,
			Description: item.Description,
			Link:        item.Link,
			Source:      feed.Title,
			Category:    source.Category, // Kategori buradan geliyor
		}

		fmt.Printf("[%s] Haber kanala gönderildi [%d/%d]: %s\n", source.Category, count+1, s.MaxPerSource, item.Title)
		count++
	}
}

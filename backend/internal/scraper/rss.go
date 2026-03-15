package scraper

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/SMutaf/twitter-bot/backend/config"
	"github.com/SMutaf/twitter-bot/backend/internal/dedup"
	"github.com/SMutaf/twitter-bot/backend/internal/filter"
	"github.com/SMutaf/twitter-bot/backend/internal/models"
	"github.com/mmcdole/gofeed"
)

type RSSScraper struct {
	Parser          *gofeed.Parser
	Cache           *dedup.Deduplicator
	BreakingChannel chan<- models.NewsItem
	NormalChannel   chan<- models.NewsItem
	MaxPerSource    int
	Filter          *filter.NewsFilter // Filtre scraper'da
}

func NewRSSScraper(
	cache *dedup.Deduplicator,
	breakingCh chan<- models.NewsItem,
	normalCh chan<- models.NewsItem,
	maxPerSource int,
	f *filter.NewsFilter,
) *RSSScraper {
	return &RSSScraper{
		Parser:          gofeed.NewParser(),
		Cache:           cache,
		BreakingChannel: breakingCh,
		NormalChannel:   normalCh,
		MaxPerSource:    maxPerSource,
		Filter:          f,
	}
}

func (s *RSSScraper) Fetch(source config.RSSSource) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	feed, err := s.Parser.ParseURLWithContext(source.URL, ctx)
	if err != nil {
		fmt.Printf("RSS Hatası (%s): %v\n", source.URL, err)
		return
	}

	// En yeni haberleri üste al
	sort.Slice(feed.Items, func(i, j int) bool {
		var t1, t2 time.Time
		if feed.Items[i].PublishedParsed != nil {
			t1 = *feed.Items[i].PublishedParsed
		} else if feed.Items[i].UpdatedParsed != nil {
			t1 = *feed.Items[i].UpdatedParsed
		}
		if feed.Items[j].PublishedParsed != nil {
			t2 = *feed.Items[j].PublishedParsed
		} else if feed.Items[j].UpdatedParsed != nil {
			t2 = *feed.Items[j].UpdatedParsed
		}
		return t1.After(t2)
	})

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

		var publishedAt time.Time
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			publishedAt = *item.UpdatedParsed
		} else {
			publishedAt = time.Now()
		}

		newsItem := models.NewsItem{
			Title:       item.Title,
			Description: item.Description,
			Link:        item.Link,
			Source:      feed.Title,
			Category:    source.Category,
			PublishedAt: publishedAt,
		}

		// --- FİLTRE BURADA — Channel'a girmeden önce karar ver ---
		allowed, reason := s.Filter.ShouldProcess(newsItem)
		if !allowed {
			fmt.Printf("[FİLTRE] Elendi (%s): %s\n", reason, item.Title)
			continue
		}
		fmt.Printf("[FİLTRE] Geçti (%s): %s\n", reason, item.Title)
		// ----------------------------------------------------------

		if source.Category == models.CategoryBreaking {
			select {
			case s.BreakingChannel <- newsItem:
				fmt.Printf("[BREAKING] Kanala eklendi [%d/%d]: %s\n",
					count+1, s.MaxPerSource, item.Title)
			default:
				fmt.Println("[BREAKING] Channel dolu, atlandı:", item.Title)
			}
		} else {
			select {
			case s.NormalChannel <- newsItem:
				fmt.Printf("[%s] Kanala eklendi [%d/%d]: %s\n",
					source.Category, count+1, s.MaxPerSource, item.Title)
			default:
				fmt.Println("[NORMAL] Channel dolu, atlandı:", item.Title)
			}
		}

		count++
	}
}

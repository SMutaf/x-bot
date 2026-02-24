package scraper

import (
	"context"
	"fmt"
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/ai" // <--- YENİ
	"github.com/SMutaf/twitter-bot/backend/internal/dedup"
	"github.com/mmcdole/gofeed"
)

type RSSScraper struct {
	Parser   *gofeed.Parser
	Cache    *dedup.Deduplicator
	AIClient *ai.Client // <--- Scraper artık AI ile konuşabiliyor
}

// NewRSSScraper güncellendi: Artık AI Client istiyor
func NewRSSScraper(cache *dedup.Deduplicator, aiClient *ai.Client) *RSSScraper {
	return &RSSScraper{
		Parser:   gofeed.NewParser(),
		Cache:    cache,
		AIClient: aiClient,
	}
}

func (s *RSSScraper) Fetch(url string) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	feed, err := s.Parser.ParseURLWithContext(url, ctx)
	if err != nil {
		fmt.Printf("RSS Hatası (%s): %v\n", url, err)
		return
	}

	fmt.Printf("Kaynak: %s\n", feed.Title)

	for _, item := range feed.Items {
		// 1. Daha önce işledik mi?
		if s.Cache.IsDuplicate(item.Link) {
			continue
		}

		fmt.Printf("\nYENİ HABER: %s\n", item.Title)

		// 2. Haberi AI Servisine Gönder
		response, err := s.AIClient.GenerateTweet(item.Title, item.Description, item.Link, feed.Title)
		if err != nil {
			fmt.Printf("AI Hatası: %v\n", err)
			continue
		}

		// 3. Sonucu Ekrana Bas (Simülasyon)
		fmt.Println("AI Tarafından Oluşturulan Tweet:")
		fmt.Println("   ------------------------------------------------")
		fmt.Printf("Tweet: %s\n", response.Tweet)
		fmt.Printf("Reply: %s\n", response.Reply)
		fmt.Printf("   mood: %s\n", response.Sentiment)
		fmt.Println("   ------------------------------------------------")

		// Demo olduğu için çok spam yapmasın diye biraz bekletelim
		time.Sleep(1 * time.Second)
	}
}

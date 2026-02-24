package scraper

import (
	"context"
	"fmt"
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/ai"
	"github.com/SMutaf/twitter-bot/backend/internal/dedup"
	"github.com/SMutaf/twitter-bot/backend/internal/telegram"
	"github.com/mmcdole/gofeed"
)

type RSSScraper struct {
	Parser   *gofeed.Parser
	Cache    *dedup.Deduplicator
	AIClient *ai.Client
	Telegram *telegram.ApprovalBot
}

func NewRSSScraper(cache *dedup.Deduplicator, aiClient *ai.Client, tgBot *telegram.ApprovalBot) *RSSScraper {
	return &RSSScraper{
		Parser:   gofeed.NewParser(),
		Cache:    cache,
		AIClient: aiClient,
		Telegram: tgBot, // Botu struct'a bağladık
	}
}

func (s *RSSScraper) Fetch(url string) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	feed, err := s.Parser.ParseURLWithContext(url, ctx)
	if err != nil {
		fmt.Printf("RSS Hatası (%s): %v\n", url, err)
		return
	}

	fmt.Printf("Kaynak Taranıyor: %s\n", feed.Title)

	for _, item := range feed.Items {
		// 1. Daha önce işledik mi? (Redis kontrolü)
		if s.Cache.IsDuplicate(item.Link) {
			continue
		}

		fmt.Printf("\nYENİ HABER BULUNDU: %s\n", item.Title)

		// 2. Haberi AI Servisine (Python/Gemma 3) Gönder
		response, err := s.AIClient.GenerateTweet(item.Title, item.Description, item.Link, feed.Title)
		if err != nil {
			fmt.Printf("AI Hatası: %v\n", err)
			continue
		}

		// log
		fmt.Println("AI Tweeti Hazırladı...")

		// 3. Telegram Onayına Gönder
		err = s.Telegram.RequestApproval(response.Tweet, response.Reply, feed.Title)
		if err != nil {
			fmt.Printf("Telegram Onay Mesajı Gönderilemedi: %v\n", err)
		} else {
			fmt.Println("Onay mesajı telefonuna gönderildi!")
		}

		// Kaynakları yormamak için kısa bir bekleme
		time.Sleep(1 * time.Second)
	}
}

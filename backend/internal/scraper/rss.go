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
		Telegram: tgBot,
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

	fmt.Printf("Kaynak Taranıyor: %s\n", feed.Title)

	for _, item := range feed.Items {
		// 1. Daha önce işledik mi? (Redis kontrolü)
		if s.Cache.IsDuplicate(item.Link) {
			continue
		}

		fmt.Printf("\nYENİ HABER BULUNDU: %s\n", item.Title)

		// 2. Haberi AI Servisine Gönder
		response, err := s.AIClient.GenerateTweet(item.Title, item.Description, item.Link, feed.Title)
		if err != nil {
			fmt.Printf("AI Hatası: %v\n", err)
			// Hata olsa bile diğer habere geçmeden önce biraz bekle ki API tamamen kilitlenmesin
			time.Sleep(5 * time.Second)
			continue
		}

		fmt.Println("AI Tweeti Hazırladı...")

		// 3. Telegram Onayına Gönder
		err = s.Telegram.RequestApproval(response.Tweet, response.Reply, feed.Title)
		if err != nil {
			fmt.Printf("Telegram Onay Mesajı Gönderilemedi: %v\n", err)
		} else {
			fmt.Println("Onay mesajı telefonuna gönderildi!")
		}

		// Google Gemini Free Tier için her istek arasında 10 saniye bekle
		fmt.Println("Kota aşımını önlemek için 10 saniye bekleniyor...")
		time.Sleep(10 * time.Second)
	}
}

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
		// 1. Link bazlı kontrol
		if s.Cache.IsDuplicate(item.Link) {
			continue
		}

		// 2. Başlık bazlı kontrol (Farklı kaynaklardaki aynı haberi engeller)
		if s.Cache.IsTitleDuplicate(item.Title) {
			fmt.Printf("Benzer haber pas geçildi: %s\n", item.Title)
			continue
		}

		fmt.Printf("\nYENİ HABER BULUNDU: %s\n", item.Title)

		response, err := s.AIClient.GenerateTweet(item.Title, item.Description, item.Link, feed.Title)
		if err != nil {
			fmt.Printf("AI Hatası: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		fmt.Println("AI Tweeti Hazırladı...")

		err = s.Telegram.RequestApproval(response.Tweet, response.Reply, feed.Title)
		if err != nil {
			fmt.Printf("Telegram Onay Mesajı Gönderilemedi: %v\n", err)
		} else {
			fmt.Println("Onay mesajı telefonuna gönderildi!")
		}

		fmt.Println("Kota aşımını önlemek için 10 saniye bekleniyor...")
		time.Sleep(10 * time.Second)
	}
}

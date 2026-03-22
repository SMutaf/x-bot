package main

import (
	"context"
	"fmt"
	"time"

	"github.com/SMutaf/twitter-bot/backend/config"
	"github.com/SMutaf/twitter-bot/backend/internal/ai"
	"github.com/SMutaf/twitter-bot/backend/internal/dedup"
	"github.com/SMutaf/twitter-bot/backend/internal/eventcluster"
	"github.com/SMutaf/twitter-bot/backend/internal/filter"
	"github.com/SMutaf/twitter-bot/backend/internal/middleware"
	"github.com/SMutaf/twitter-bot/backend/internal/models"
	"github.com/SMutaf/twitter-bot/backend/internal/policy"
	"github.com/SMutaf/twitter-bot/backend/internal/scoring"
	"github.com/SMutaf/twitter-bot/backend/internal/scraper"
	"github.com/SMutaf/twitter-bot/backend/internal/telegram"
	"golang.org/x/time/rate"
)

func main() {
	fmt.Println("Twitter Bot Backend Başlatılıyor.")

	cfg := config.LoadConfig()

	cache := dedup.NewDeduplicator(cfg.RedisAddr)
	cache.Client.FlushAll(cache.Ctx)
	fmt.Println("Redis Hafızası Silindi")

	clusterer := eventcluster.NewEventClusterer(cache.Client)
	newsScorer := scoring.NewNewsScorer(cache.Client)

	aiClient := ai.NewClient("http://localhost:8000")
	tgBot := telegram.NewApprovalBot(cfg.TelegramToken, cfg.TelegramChatID)

	go tgBot.ListenForApproval()
	fmt.Println("Telegram Onay Servisi Aktif!")

	breakingChannel := make(chan models.NewsItem, 100)
	normalChannel := make(chan models.NewsItem, 200)

	newsFilter := filter.NewNewsFilter(44, 38)
	sc := scraper.NewRSSScraper(cache, breakingChannel, normalChannel, cfg.MaxNewsPerSource, newsFilter, clusterer)

	limiter := rate.NewLimiter(rate.Every(3*time.Second), 1)

	go func() {
		breakingStreak := 0
		const maxBreakingStreak = 3

		for {
			if breakingStreak >= maxBreakingStreak {
				select {
				case item := <-normalChannel:
					fmt.Println("[DENGE] Normal kanala geçiliyor...")
					limiter.Wait(context.Background())
					middleware.RecoveryWrapper("Normal News Worker", func() {
						processNews(item, aiClient, tgBot, newsScorer, clusterer)
					})
					breakingStreak = 0
					continue
				default:
					fmt.Println("[DENGE] Normal kanalda haber yok, breaking'e dönülüyor.")
					breakingStreak = 0
				}
			}

			select {
			case item := <-breakingChannel:
				limiter.Wait(context.Background())
				middleware.RecoveryWrapper("Breaking News Worker", func() {
					processNews(item, aiClient, tgBot, newsScorer, clusterer)
				})
				breakingStreak++
				fmt.Printf("Breaking streak: %d/%d\n", breakingStreak, maxBreakingStreak)

			case item := <-normalChannel:
				limiter.Wait(context.Background())
				middleware.RecoveryWrapper("Normal News Worker", func() {
					processNews(item, aiClient, tgBot, newsScorer, clusterer)
				})
				breakingStreak = 0

			case <-time.After(100 * time.Millisecond):
				continue
			}
		}
	}()

	fmt.Println("Priority Worker Başlatıldı! (Breaking 3:1 Normal oranında)")

	for _, source := range cfg.RSSSources {
		src := source
		go func() {
			fmt.Printf("Kaynak başlatıldı [%s | %s]: %s\n", src.Category, src.Interval, src.URL)
			for {
				middleware.RecoveryWrapper("Tarama", func() {
					sc.Fetch(src)
				})
				time.Sleep(src.Interval)
			}
		}()
	}

	fmt.Println("Tüm kaynaklar aktif. Bot çalışıyor...")
	select {}
}

func processNews(
	item models.NewsItem,
	aiClient *ai.Client,
	tgBot *telegram.ApprovalBot,
	newsScorer *scoring.NewsScorer,
	clusterer *eventcluster.EventClusterer,
) {
	catPolicy := policy.Get(item.Category)

	if item.ClusterCount < catPolicy.MinClusterCount {
		fmt.Printf("[HARD FILTER] %s için yetersiz kaynak (%d < %d): %s\n",
			item.Category, item.ClusterCount, catPolicy.MinClusterCount, item.Title)
		return
	}

	if item.ClusterKey != "" && clusterer.WasSentRecently(item.ClusterKey) {
		fmt.Printf("[EVENT DEDUPE] Aynı event yakın zamanda gönderilmiş, atlandı: %s\n", item.Title)
		return
	}

	score := newsScorer.Calculate(item)
	fmt.Printf("[%s] Virality: %d (%s) | ClusterCount: %d | Boost: +%d | %s\n",
		item.Category, score.Final, newsScorer.GetViralityLevel(score.Final), item.ClusterCount, item.Score, item.Title)

	if score.Final < catPolicy.MinVirality {
		if !(policy.IsCriticalEvent(item) && policy.IsAcceptableCriticalAge(item, catPolicy)) {
			fmt.Printf("[VIRALITY FILTER] Elendi (score:%d < min:%d): %s\n",
				score.Final, catPolicy.MinVirality, item.Title)
			return
		}
		fmt.Printf("[CRITICAL OVERRIDE] Düşük skora rağmen geçirildi: %s\n", item.Title)
	}

	loc, _ := time.LoadLocation("Europe/Istanbul")
	now := time.Now().In(loc)
	hour := now.Hour()

	if item.Category == models.CategoryTech {
		if !((hour >= 8 && hour < 11) || (hour >= 13 && hour < 15) || (hour >= 18 && hour <= 22)) {
			fmt.Printf("[SAAT FİLTRE] Gönderilmiyor: %s\n", item.Title)
			return
		}
	}

	publishedTime := ""
	if !item.PublishedAt.IsZero() {
		diff := time.Since(item.PublishedAt)
		switch {
		case diff < 5*time.Minute:
			publishedTime = "🔴 ŞU AN"
		case diff < 30*time.Minute:
			publishedTime = fmt.Sprintf("%d dk önce", int(diff.Minutes()))
		case diff < 2*time.Hour:
			publishedTime = fmt.Sprintf("%d saat önce", int(diff.Hours()))
		default:
			publishedTime = item.PublishedAt.In(loc).Format("15:04")
		}
	}

	response, err := aiClient.GenerateTweet(
		item.Title,
		item.Description,
		item.Link,
		item.Source,
		string(item.Category),
		item.PublishedAt,
	)
	if err != nil {
		fmt.Printf("AI Hatası (%s): %v\n", item.Title, err)
		return
	}

	if response.Tweet == "" {
		fmt.Printf("AI boş tweet döndü: %s\n", item.Title)
		return
	}

	fmt.Printf("AI cevap aldı - Tweet: %s... | Reply: %s...\n",
		response.Tweet[:min(30, len(response.Tweet))],
		response.Reply[:min(30, len(response.Reply))])

	err = tgBot.RequestApproval(response.Tweet, item.Link, item.Source, string(item.Category), publishedTime)
	if err != nil {
		fmt.Printf("Telegram Hatası: %v\n", err)
		return
	}

	if item.ClusterKey != "" {
		clusterer.MarkSent(item.ClusterKey, catPolicy.DedupeCooldown)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

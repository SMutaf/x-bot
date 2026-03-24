package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/SMutaf/twitter-bot/backend/config"
	"github.com/SMutaf/twitter-bot/backend/internal/ai"
	"github.com/SMutaf/twitter-bot/backend/internal/dashboardapi"
	"github.com/SMutaf/twitter-bot/backend/internal/dedup"
	"github.com/SMutaf/twitter-bot/backend/internal/eventcluster"
	"github.com/SMutaf/twitter-bot/backend/internal/filter"
	"github.com/SMutaf/twitter-bot/backend/internal/middleware"
	"github.com/SMutaf/twitter-bot/backend/internal/models"
	"github.com/SMutaf/twitter-bot/backend/internal/monitoring"
	"github.com/SMutaf/twitter-bot/backend/internal/pipeline"
	"github.com/SMutaf/twitter-bot/backend/internal/scoring"
	"github.com/SMutaf/twitter-bot/backend/internal/scraper"
	"github.com/SMutaf/twitter-bot/backend/internal/sourcehealth"
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
	healthManager := sourcehealth.NewManager()

	monitor, err := monitoring.NewManager("data")
	if err != nil {
		panic(err)
	}

	aiClient := ai.NewClient("http://localhost:8000")
	tgBot := telegram.NewApprovalBot(cfg.TelegramToken, cfg.TelegramChatID)

	processor := pipeline.NewProcessor(
		newsScorer,
		aiClient,
		tgBot,
		clusterer,
		monitor,
	)

	breakingChannel := make(chan models.NewsItem, 100)
	normalChannel := make(chan models.NewsItem, 200)

	newsFilter := filter.NewNewsFilter(44, 38)
	sc := scraper.NewRSSScraper(
		cache,
		breakingChannel,
		normalChannel,
		cfg.MaxNewsPerSource,
		newsFilter,
		clusterer,
		healthManager,
		monitor,
	)

	go func() {
		mux := http.NewServeMux()
		api := dashboardapi.NewHandler(monitor, healthManager)
		api.Register(mux)

		handler := dashboardapi.WithCORS(mux)

		fmt.Println("Dashboard API aktif: http://localhost:8081")
		if err := http.ListenAndServe(":8081", handler); err != nil {
			fmt.Printf("Dashboard API hatası: %v\n", err)
		}
	}()

	limiter := rate.NewLimiter(rate.Every(3*time.Second), 1)

	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			snapshot := healthManager.Snapshot()
			fmt.Println(sourcehealth.FormatSnapshot(snapshot))
		}
	}()

	go func() {
		breakingStreak := 0
		const maxBreakingStreak = 3

		for {
			if breakingStreak >= maxBreakingStreak {
				select {
				case item := <-normalChannel:
					fmt.Println("[DENGE] Normal kanala geçiliyor...")
					_ = limiter.Wait(context.Background())
					middleware.RecoveryWrapper("Normal News Worker", func() {
						_ = processor.Process(item)
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
				_ = limiter.Wait(context.Background())
				middleware.RecoveryWrapper("Breaking News Worker", func() {
					_ = processor.Process(item)
				})
				breakingStreak++
				fmt.Printf("Breaking streak: %d/%d\n", breakingStreak, maxBreakingStreak)

			case item := <-normalChannel:
				_ = limiter.Wait(context.Background())
				middleware.RecoveryWrapper("Normal News Worker", func() {
					_ = processor.Process(item)
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

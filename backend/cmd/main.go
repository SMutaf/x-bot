package main

import (
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
	"github.com/SMutaf/twitter-bot/backend/internal/render"
	"github.com/SMutaf/twitter-bot/backend/internal/scoring"
	"github.com/SMutaf/twitter-bot/backend/internal/scraper"
	"github.com/SMutaf/twitter-bot/backend/internal/sourcehealth"
	"github.com/SMutaf/twitter-bot/backend/internal/stream"
	"github.com/SMutaf/twitter-bot/backend/internal/telegram"
	"github.com/SMutaf/twitter-bot/backend/internal/translation"
	"golang.org/x/time/rate"
)

func main() {
	fmt.Println("Twitter Bot Backend Baslatiliyor.")

	cfg := config.LoadConfig()

	cache := dedup.NewDeduplicator(cfg.RedisAddr)
	fmt.Println("Redis baglantisi hazir")

	clusterer := eventcluster.NewEventClusterer(cache.Client)
	newsScorer := scoring.NewNewsScorer(cache.Client)
	healthManager := sourcehealth.NewManager()

	cache.Client.FlushAll(cache.Ctx)
	fmt.Println("Redis Hafizasi Silindi")

	monitor, err := monitoring.NewManager(cache)
	if err != nil {
		panic(err)
	}

	aiClient := ai.NewClient("http://localhost:8000")
	tgBot := telegram.NewApprovalBot(cfg.TelegramToken, cfg.TelegramChatID)
	telegramRenderer := render.NewTelegramRenderer()
	translator := translation.NewLibreTranslator("http://localhost:5000")
	serviceStatus := dashboardapi.NewServiceStatusManager(cache, aiClient)
	serviceStatus.Start(10 * time.Second)

	processor := pipeline.NewProcessor(
		newsScorer,
		aiClient,
		tgBot,
		clusterer,
		monitor,
		telegramRenderer,
		translator,
	)

	channels := pipeline.CategoryChannels{
		Breaking: make(chan models.NewsEnvelope, 50),
		Economy:  make(chan models.NewsEnvelope, 100),
		General:  make(chan models.NewsEnvelope, 100),
		Tech:     make(chan models.NewsEnvelope, 150),
	}

	newsFilter := filter.NewNewsFilter()
	sc := scraper.NewRSSScraper(
		cache,
		channels,
		cfg.MaxNewsPerSource,
		newsFilter,
		clusterer,
		healthManager,
		monitor,
	)

	go func() {
		mux := http.NewServeMux()
		statusProvider := &dashboardapi.StatusProvider{
			Monitoring: monitor,
			Services:   serviceStatus,
		}
		api := dashboardapi.NewHandler(monitor, healthManager, statusProvider)
		mux.HandleFunc("/api/feed/stream", stream.StreamHandler)
		api.Register(mux)

		handler := dashboardapi.WithCORS(mux)

		fmt.Println("Dashboard API aktif: http://localhost:8081")
		if err := http.ListenAndServe(":8081", handler); err != nil {
			fmt.Printf("Dashboard API hatasi: %v\n", err)
		}
	}()

	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			snapshot := healthManager.Snapshot()
			fmt.Println(sourcehealth.FormatSnapshot(snapshot))
		}
	}()

	limiter := rate.NewLimiter(rate.Every(3*time.Second), 1)
	dispatcher := pipeline.NewDispatcher(channels, processor, limiter)
	go dispatcher.Run()

	for _, source := range cfg.RSSSources {
		src := source
		go func() {
			fmt.Printf("Kaynak baslatildi [%s | %s]: %s\n", src.Category, src.Interval, src.URL)
			for {
				middleware.RecoveryWrapper("Tarama", func() {
					sc.Fetch(src)
				})
				time.Sleep(src.Interval)
			}
		}()
	}

	fmt.Println("Tum kaynaklar aktif. Bot calisiyor...")
	select {}
}

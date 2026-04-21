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
	telegramRenderer := render.NewTelegramRenderer()

	processor := pipeline.NewProcessor(
		newsScorer,
		aiClient,
		tgBot,
		clusterer,
		monitor,
		telegramRenderer,
	)

	// Her kategori için ayrı kanal — buffer boyutları kategori karakteristiğine göre ayarlandı.
	// Breaking: 2dk polling, 30dk TTL → küçük buffer yeterli.
	// Economy/General: 3-5dk polling, orta buffer.
	// Tech: 10dk polling, 8 saat MaxAge → büyük buffer.
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

	// Dashboard API
	go func() {
		mux := http.NewServeMux()
		api := dashboardapi.NewHandler(monitor, healthManager)
		mux.HandleFunc("/api/feed/stream", stream.StreamHandler)
		api.Register(mux)

		handler := dashboardapi.WithCORS(mux)

		fmt.Println("Dashboard API aktif: http://localhost:8081")
		if err := http.ListenAndServe(":8081", handler); err != nil {
			fmt.Printf("Dashboard API hatası: %v\n", err)
		}
	}()

	// Source health snapshot — her 2 dakikada bir loglanır.
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			snapshot := healthManager.Snapshot()
			fmt.Println(sourcehealth.FormatSnapshot(snapshot))
		}
	}()

	// Dispatcher: Tiered Priority + Starvation Protection
	// Tek rate limiter tüm kategoriler için ortaktır — Telegram rate limit korunur.
	limiter := rate.NewLimiter(rate.Every(3*time.Second), 1)
	dispatcher := pipeline.NewDispatcher(channels, processor, limiter)
	go dispatcher.Run()

	// RSS kaynaklarını başlat — her kaynak kendi goroutine'inde döner.
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

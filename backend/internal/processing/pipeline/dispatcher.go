package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/domain/models"
	"github.com/SMutaf/twitter-bot/backend/internal/infra/middleware"
	"golang.org/x/time/rate"
)

// CategoryChannels her kategori için ayrı bir kanal tutar.
type CategoryChannels struct {
	Breaking chan models.NewsEnvelope
	Economy  chan models.NewsEnvelope
	General  chan models.NewsEnvelope
	Tech     chan models.NewsEnvelope
}

// Dispatcher tiered priority + starvation protection ile çalışır.
// Öncelik sırası: Breaking > Economy > General > Tech
// Bir kategori maxWait süresince hiç işlenmediyse starvation koruması devreye girer
// ve o kategori, bir üst önceliklinin önüne geçer (Breaking hariç, o her zaman ilk).
type Dispatcher struct {
	channels      CategoryChannels
	processor     *Processor
	limiter       *rate.Limiter
	lastProcessed map[string]time.Time
	maxWait       map[string]time.Duration
}

func NewDispatcher(ch CategoryChannels, processor *Processor, limiter *rate.Limiter) *Dispatcher {
	now := time.Now()
	return &Dispatcher{
		channels:  ch,
		processor: processor,
		limiter:   limiter,
		lastProcessed: map[string]time.Time{
			"BREAKING": now,
			"ECONOMY":  now,
			"GENERAL":  now,
			"TECH":     now,
		},
		// Breaking için maxWait yok — her zaman en önce işlenir.
		// Diğerleri için: eğer bu süre geçmişse starvation override devreye girer.
		maxWait: map[string]time.Duration{
			"ECONOMY": 2 * time.Minute,
			"GENERAL": 4 * time.Minute,
			"TECH":    8 * time.Minute,
		},
	}
}

// Run dispatcher döngüsünü başlatır. Goroutine olarak çağrılmalı.
func (d *Dispatcher) Run() {
	fmt.Println("[DISPATCHER] Tiered Priority Worker başlatıldı. (Breaking > Economy > General > Tech)")

	for {
		env, category, ok := d.pickNext()
		if !ok {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		_ = d.limiter.Wait(context.Background())

		cat := category
		item := env
		middleware.RecoveryWrapper(cat+" Worker", func() {
			if err := d.processor.Process(item); err != nil {
				fmt.Printf("[DISPATCHER HATA] %s: %v (haber: %s)\n", cat, err, item.News.Title)
			}
		})

		d.lastProcessed[cat] = time.Now()
	}
}

// pickNext bir sonraki işlenecek haberi seçer.
// Önce Breaking kontrol edilir (her zaman önce).
// Ardından starvation durumuna göre Economy > General > Tech sırası ile devam edilir.
func (d *Dispatcher) pickNext() (models.NewsEnvelope, string, bool) {
	// 1. Breaking her zaman en önce — starvation protection gereksiz.
	select {
	case item := <-d.channels.Breaking:
		fmt.Printf("[DISPATCHER] ✅ BREAKING işleniyor: %s\n", item.News.Title)
		return item, "BREAKING", true
	default:
	}

	// 2. Starvation kontrolü — priority sırasına göre (Economy önce, sonra General, sonra Tech).
	//    Starving olan en yüksek öncelikli kategori önce işlenir.
	if d.isStarving("ECONOMY") {
		select {
		case item := <-d.channels.Economy:
			fmt.Printf("[DISPATCHER] ⚠️  ECONOMY starvation override: %s\n", item.News.Title)
			return item, "ECONOMY", true
		default:
		}
	}

	if d.isStarving("GENERAL") {
		select {
		case item := <-d.channels.General:
			fmt.Printf("[DISPATCHER] ⚠️  GENERAL starvation override: %s\n", item.News.Title)
			return item, "GENERAL", true
		default:
		}
	}

	if d.isStarving("TECH") {
		select {
		case item := <-d.channels.Tech:
			fmt.Printf("[DISPATCHER] ⚠️  TECH starvation override: %s\n", item.News.Title)
			return item, "TECH", true
		default:
		}
	}

	// 3. Normal priority: Economy > General > Tech (Breaking zaten 1. adımda işlendi).
	select {
	case item := <-d.channels.Economy:
		fmt.Printf("[DISPATCHER] 📈 ECONOMY işleniyor: %s\n", item.News.Title)
		return item, "ECONOMY", true
	default:
	}

	select {
	case item := <-d.channels.General:
		fmt.Printf("[DISPATCHER] 📰 GENERAL işleniyor: %s\n", item.News.Title)
		return item, "GENERAL", true
	default:
	}

	select {
	case item := <-d.channels.Tech:
		fmt.Printf("[DISPATCHER] 💻 TECH işleniyor: %s\n", item.News.Title)
		return item, "TECH", true
	default:
	}

	return models.NewsEnvelope{}, "", false
}

// isStarving verilen kategori için maxWait süresinin aşılıp aşılmadığını kontrol eder.
func (d *Dispatcher) isStarving(category string) bool {
	maxWait, ok := d.maxWait[category]
	if !ok {
		return false
	}
	last, ok := d.lastProcessed[category]
	if !ok {
		return false
	}
	return time.Since(last) > maxWait
}

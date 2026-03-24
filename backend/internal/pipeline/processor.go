package pipeline

import (
	"fmt"
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/ai"
	"github.com/SMutaf/twitter-bot/backend/internal/eventcluster"
	"github.com/SMutaf/twitter-bot/backend/internal/models"
	"github.com/SMutaf/twitter-bot/backend/internal/monitoring"
	"github.com/SMutaf/twitter-bot/backend/internal/policy"
	"github.com/SMutaf/twitter-bot/backend/internal/scoring"
	"github.com/SMutaf/twitter-bot/backend/internal/telegram"
)

type Processor struct {
	scorer   *scoring.NewsScorer
	ai       *ai.Client
	telegram *telegram.ApprovalBot
	cluster  *eventcluster.EventClusterer
	monitor  *monitoring.Manager
	istLoc   *time.Location
}

func NewProcessor(
	scorer *scoring.NewsScorer,
	aiClient *ai.Client,
	tgBot *telegram.ApprovalBot,
	clusterer *eventcluster.EventClusterer,
	monitor *monitoring.Manager,
) *Processor {
	loc, _ := time.LoadLocation("Europe/Istanbul")
	return &Processor{
		scorer:   scorer,
		ai:       aiClient,
		telegram: tgBot,
		cluster:  clusterer,
		monitor:  monitor,
		istLoc:   loc,
	}
}

func (p *Processor) Process(item models.NewsItem) error {
	catPolicy := policy.Get(item.Category)

	if item.ClusterCount < catPolicy.MinClusterCount {
		fmt.Printf("[HARD FILTER] %s için yetersiz kaynak (%d < %d): %s\n",
			item.Category, item.ClusterCount, catPolicy.MinClusterCount, item.Title)
		p.recordRejected(item, "process-min-cluster")
		return nil
	}

	if item.ClusterKey != "" && p.cluster.WasSentRecently(item.ClusterKey) {
		fmt.Printf("[EVENT DEDUPE] Aynı event yakın zamanda gönderilmiş, atlandı: %s\n", item.Title)
		p.recordRejected(item, "event-dedupe")
		return nil
	}

	score := p.scorer.Calculate(item)
	fmt.Printf("[%s] Virality: %d (%s) | ClusterCount: %d | %s\n",
		item.Category, score.Final, p.scorer.GetViralityLevel(score.Final), item.ClusterCount, item.Title)

	if score.Final < catPolicy.MinVirality {
		if !(policy.IsCriticalEvent(item) && policy.IsAcceptableCriticalAge(item, catPolicy)) {
			fmt.Printf("[VIRALITY FILTER] Elendi (score:%d < min:%d): %s\n",
				score.Final, catPolicy.MinVirality, item.Title)
			p.recordRejected(item, "virality-filter")
			return nil
		}
		fmt.Printf("[CRITICAL OVERRIDE] Düşük skora rağmen geçirildi: %s\n", item.Title)
	}

	if item.Category == models.CategoryTech && !p.isAllowedTechHour() {
		fmt.Printf("[SAAT FİLTRE] Gönderilmiyor: %s\n", item.Title)
		p.recordRejected(item, "tech-time-filter")
		return nil
	}

	publishedTime := p.buildPublishedTime(item)

	response, err := p.ai.GenerateTelegramPost(
		item.Title,
		item.Description,
		item.Link,
		item.Source,
		string(item.Category),
		item.PublishedAt,
	)
	if err != nil {
		fmt.Printf("AI Hatası (%s): %v\n", item.Title, err)
		p.recordRejected(item, "ai-error")
		return err
	}

	if response.Message == "" {
		fmt.Printf("AI boş cevap döndü: %s\n", item.Title)
		p.recordRejected(item, "ai-empty-message")
		return nil
	}

	fmt.Printf("AI cevap aldı - Message: %s...\n", response.Message[:min(60, len(response.Message))])

	if err := p.telegram.RequestApproval(response.Message, string(item.Category), publishedTime); err != nil {
		fmt.Printf("Telegram Hatası: %v\n", err)
		p.recordRejected(item, "telegram-error")
		return err
	}

	p.monitor.RecordPublished(monitoring.PublishedNewsEvent{
		Time:         time.Now(),
		Title:        item.Title,
		Category:     string(item.Category),
		Source:       item.Source,
		Link:         item.Link,
		Virality:     score.Final,
		ClusterCount: item.ClusterCount,
	})

	if item.ClusterKey != "" {
		p.cluster.MarkSent(item.ClusterKey, catPolicy.DedupeCooldown)
	}

	return nil
}

func (p *Processor) recordRejected(item models.NewsItem, reason string) {
	p.monitor.RecordRejected(monitoring.RejectedNewsEvent{
		Time:     time.Now(),
		Title:    item.Title,
		Category: string(item.Category),
		Source:   item.Source,
		Reason:   reason,
	})
}

func (p *Processor) isAllowedTechHour() bool {
	hour := time.Now().In(p.istLoc).Hour()
	return (hour >= 8 && hour < 11) || (hour >= 13 && hour < 15) || (hour >= 18 && hour <= 22)
}

func (p *Processor) buildPublishedTime(item models.NewsItem) string {
	if item.PublishedAt.IsZero() {
		return ""
	}
	diff := time.Since(item.PublishedAt)
	switch {
	case diff < 5*time.Minute:
		return "🔴 ŞU AN"
	case diff < 30*time.Minute:
		return fmt.Sprintf("%d dk önce", int(diff.Minutes()))
	case diff < 2*time.Hour:
		return fmt.Sprintf("%d saat önce", int(diff.Hours()))
	default:
		return item.PublishedAt.In(p.istLoc).Format("15:04")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

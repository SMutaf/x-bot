package pipeline

import (
	"fmt"
	"strings"
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/ai"
	"github.com/SMutaf/twitter-bot/backend/internal/eventcluster"
	"github.com/SMutaf/twitter-bot/backend/internal/models"
	"github.com/SMutaf/twitter-bot/backend/internal/monitoring"
	"github.com/SMutaf/twitter-bot/backend/internal/policy"
	"github.com/SMutaf/twitter-bot/backend/internal/render"
	"github.com/SMutaf/twitter-bot/backend/internal/scoring"
	"github.com/SMutaf/twitter-bot/backend/internal/telegram"
	"github.com/SMutaf/twitter-bot/backend/internal/translation"
)

type Processor struct {
	scorer     *scoring.NewsScorer
	ai         *ai.Client
	telegram   *telegram.ApprovalBot
	cluster    *eventcluster.EventClusterer
	monitor    *monitoring.Manager
	renderer   *render.TelegramRenderer
	translator *translation.LibreTranslator
	istLoc     *time.Location
}

func NewProcessor(
	scorer *scoring.NewsScorer,
	aiClient *ai.Client,
	tgBot *telegram.ApprovalBot,
	clusterer *eventcluster.EventClusterer,
	monitor *monitoring.Manager,
	renderer *render.TelegramRenderer,
	translator *translation.LibreTranslator,
) *Processor {
	loc, _ := time.LoadLocation("Europe/Istanbul")
	return &Processor{
		scorer:     scorer,
		ai:         aiClient,
		telegram:   tgBot,
		cluster:    clusterer,
		monitor:    monitor,
		renderer:   renderer,
		translator: translator,
		istLoc:     loc,
	}
}

func (p *Processor) Process(env models.NewsEnvelope) error {
	catPolicy := policy.Get(env.News.Category)

	if env.Cluster.ClusterCount < catPolicy.MinClusterCount {
		fmt.Printf("[HARD FILTER] %s için yetersiz kaynak (%d < %d): %s\n",
			env.News.Category, env.Cluster.ClusterCount, catPolicy.MinClusterCount, env.News.Title)
		p.recordRejected(env, "process-min-cluster")
		return nil
	}

	if env.Cluster.ClusterKey != "" && p.cluster.WasSentRecently(env.Cluster.ClusterKey) {
		fmt.Printf("[EVENT DEDUPE] Aynı event yakın zamanda gönderilmiş, atlandı: %s\n", env.News.Title)
		p.recordRejected(env, "event-dedupe")
		return nil
	}

	score := p.scorer.Calculate(env)
	env.Score = score
	env.Stage = models.StageScored

	fmt.Printf("[%s] Virality: %d (%s) | ClusterCount: %d | %s\n",
		env.News.Category, score.Final, p.scorer.GetViralityLevel(score.Final), env.Cluster.ClusterCount, env.News.Title)

	if score.Final < catPolicy.MinVirality {
		if !(policy.IsCriticalEvent(env, catPolicy) && policy.IsAcceptableCriticalAge(env, catPolicy)) {
			fmt.Printf("[VIRALITY FILTER] Elendi (score:%d < min:%d): %s\n",
				score.Final, catPolicy.MinVirality, env.News.Title)
			p.recordRejected(env, "virality-filter")
			return nil
		}
		fmt.Printf("[CRITICAL OVERRIDE] Düşük skora rağmen geçirildi: %s\n", env.News.Title)
	}

	if env.News.Category == models.CategoryTech && !p.isAllowedTechHour() {
		fmt.Printf("[SAAT FİLTRE] Gönderilmiyor: %s\n", env.News.Title)
		p.recordRejected(env, "tech-time-filter")
		return nil
	}

	req := models.EditorialAnalysisRequest{
		Title:        env.News.Title,
		Description:  env.News.Description,
		Category:     string(env.News.Category),
		Source:       env.News.Source,
		PublishedAt:  env.News.PublishedAt,
		ClusterCount: env.Cluster.ClusterCount,
		Virality:     env.Score.Final,
	}

	res, err := p.ai.Analyze(req)
	if err != nil {
		fmt.Printf("AI Hatası (%s): %v\n", env.News.Title, err)
		p.recordRejected(env, "ai-error")
		return err
	}

	env.Stage = models.StageLLMAnalyzed

	decision := p.mapToEditorialDecision(env, res)

	fmt.Printf("[AI] Decision: %s | Reason: %s | %s\n",
		decision.Decision, decision.RejectReason, env.News.Title)

	if decision.Decision == models.DecisionReject {
		reason := strings.TrimSpace(decision.RejectReason)
		if reason == "" {
			reason = "llm-editorial-reject"
		}

		fmt.Printf("LLM editoryal olarak reddetti (%s): %s\n", reason, env.News.Title)
		p.recordRejected(env, "llm-"+reason)
		return nil
	}

	if decision.Decision != models.DecisionPublish {
		fmt.Printf("AI geçersiz decision döndü (%s): %s\n", decision.Decision, env.News.Title)
		p.recordRejected(env, "ai-invalid-decision")
		return nil
	}

	message := p.renderer.Render(env, decision)
	if strings.TrimSpace(message) == "" {
		fmt.Printf("AI boş veya geçersiz içerik döndü: %s\n", env.News.Title)
		p.recordRejected(env, "ai-empty-message")
		return nil
	}

	publishedTime := p.buildPublishedTime(env)

	fmt.Printf("AI cevap aldı - Message: %s...\n", message[:min(60, len(message))])

	if err := p.telegram.RequestApproval(message, string(env.News.Category), publishedTime); err != nil {
		fmt.Printf("Telegram Hatası: %v\n", err)
		p.recordRejected(env, "telegram-error")
		return err
	}

	translatedDesc, err := p.translator.Translate(env.News.Description, "en", "tr")
	if err == nil {
		env.News.Description = translatedDesc
	}

	p.monitor.RecordPublished(monitoring.PublishedNewsEvent{
		Time:         time.Now(),
		Title:        env.News.Title,
		Description:  env.News.Description,
		Hook:         decision.Hook,
		Summary:      decision.Summary,
		Importance:   decision.Importance,
		Sentiment:    decision.Sentiment,
		Category:     string(env.News.Category),
		Source:       env.News.Source,
		Link:         env.News.Link,
		Virality:     score.Final,
		ClusterCount: env.Cluster.ClusterCount,
	})

	if env.Cluster.ClusterKey != "" {
		p.cluster.MarkSent(env.Cluster.ClusterKey, catPolicy.DedupeCooldown)
	}

	return nil
}

func (p *Processor) mapToEditorialDecision(env models.NewsEnvelope, res *models.EditorialAnalysisResponse) models.EditorialDecision {
	decisionType := models.EditorialDecisionType(strings.ToUpper(strings.TrimSpace(res.Decision)))

	if decisionType != models.DecisionPublish &&
		decisionType != models.DecisionReject &&
		decisionType != models.DecisionReview {
		decisionType = ""
	}

	return models.EditorialDecision{
		ID:              env.News.ID,
		NewsID:          env.News.ID,
		Decision:        decisionType,
		RejectReason:    strings.TrimSpace(res.RejectReason),
		NewsType:        "",
		Sentiment:       strings.TrimSpace(res.Sentiment),
		Hook:            strings.TrimSpace(res.Hook),
		Summary:         strings.TrimSpace(res.Summary),
		Importance:      strings.TrimSpace(res.Importance),
		SourceLine:      fmt.Sprintf("Kaynak: %s", env.News.Source),
		ApprovalStatus:  models.ApprovalPending,
		ApprovalChannel: "telegram",
		CreatedAt:       time.Now(),
	}
}

func (p *Processor) recordRejected(env models.NewsEnvelope, reason string) {
	p.monitor.RecordRejected(monitoring.RejectedNewsEvent{
		Time:     time.Now(),
		Title:    env.News.Title,
		Category: string(env.News.Category),
		Source:   env.News.Source,
		Reason:   reason,
	})
}

func (p *Processor) isAllowedTechHour() bool {
	hour := time.Now().In(p.istLoc).Hour()
	return (hour >= 8 && hour < 11) || (hour >= 13 && hour < 15) || (hour >= 18 && hour <= 22)
}

func (p *Processor) buildPublishedTime(env models.NewsEnvelope) string {
	if env.News.PublishedAt.IsZero() {
		return ""
	}
	diff := time.Since(env.News.PublishedAt)
	switch {
	case diff < 5*time.Minute:
		return "🔴 ŞU AN"
	case diff < 30*time.Minute:
		return fmt.Sprintf("%d dk önce", int(diff.Minutes()))
	case diff < 2*time.Hour:
		return fmt.Sprintf("%d saat önce", int(diff.Hours()))
	default:
		return env.News.PublishedAt.In(p.istLoc).Format("15:04")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

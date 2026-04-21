package scraper

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/SMutaf/twitter-bot/backend/config"
	"github.com/SMutaf/twitter-bot/backend/internal/dedup"
	"github.com/SMutaf/twitter-bot/backend/internal/eventcluster"
	"github.com/SMutaf/twitter-bot/backend/internal/filter"
	"github.com/SMutaf/twitter-bot/backend/internal/models"
	"github.com/SMutaf/twitter-bot/backend/internal/monitoring"
	"github.com/SMutaf/twitter-bot/backend/internal/pipeline"
	"github.com/SMutaf/twitter-bot/backend/internal/policy"
	"github.com/SMutaf/twitter-bot/backend/internal/sourcehealth"
	"github.com/mmcdole/gofeed"
)

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36"

type customTransport struct {
	rt http.RoundTripper
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cloned := req.Clone(req.Context())
	cloned.Header.Set("User-Agent", defaultUserAgent)
	cloned.Header.Set("Accept", "application/rss+xml, application/xml, text/xml;q=0.9, */*;q=0.8")
	cloned.Header.Set("Accept-Language", "en-US,en;q=0.9,tr;q=0.8")
	cloned.Header.Set("Cache-Control", "no-cache")
	return t.rt.RoundTrip(cloned)
}

type RSSScraper struct {
	Parser        *gofeed.Parser
	Cache         *dedup.Deduplicator
	Channels      pipeline.CategoryChannels // ← BreakingChannel + NormalChannel yerine
	MaxPerSource  int
	Filter        *filter.NewsFilter
	Clusterer     *eventcluster.EventClusterer
	HealthManager *sourcehealth.Manager
	Monitoring    *monitoring.Manager
}

func NewRSSScraper(
	cache *dedup.Deduplicator,
	channels pipeline.CategoryChannels, // ← iki ayrı kanal yerine tek struct
	maxPerSource int,
	f *filter.NewsFilter,
	ec *eventcluster.EventClusterer,
	healthManager *sourcehealth.Manager,
	monitor *monitoring.Manager,
) *RSSScraper {
	parser := gofeed.NewParser()
	parser.Client = &http.Client{
		Timeout: 15 * time.Second,
		Transport: &customTransport{
			rt: http.DefaultTransport,
		},
	}

	return &RSSScraper{
		Parser:        parser,
		Cache:         cache,
		Channels:      channels,
		MaxPerSource:  maxPerSource,
		Filter:        f,
		Clusterer:     ec,
		HealthManager: healthManager,
		Monitoring:    monitor,
	}
}

// routeToChannel haberi kategorisine göre doğru kanala yönlendirir.
func (s *RSSScraper) routeToChannel(env models.NewsEnvelope, itemTitle string, count int) bool {
	var ch chan<- models.NewsEnvelope
	cat := string(env.News.Category)

	switch env.News.Category {
	case models.CategoryBreaking:
		ch = s.Channels.Breaking
	case models.CategoryEconomy:
		ch = s.Channels.Economy
	case models.CategoryTech:
		ch = s.Channels.Tech
	default: // GENERAL ve tanımsız kategoriler
		ch = s.Channels.General
		cat = "GENERAL"
	}

	select {
	case ch <- env:
		fmt.Printf("[%s] Kanala eklendi [%d/%d]: %s\n", cat, count+1, s.MaxPerSource, itemTitle)
		return true
	default:
		fmt.Printf("[%s] Channel dolu, atlandı: %s\n", cat, itemTitle)
		s.recordRejected(env, strings.ToLower(cat)+"-channel-full")
		return false
	}
}

func (s *RSSScraper) Fetch(source config.RSSSource) {
	sourceName := feedSourceName(source.URL)

	if s.HealthManager != nil {
		shouldSkip, status := s.HealthManager.ShouldSkip(source, sourceName)
		if shouldSkip {
			fmt.Printf("[SOURCE HEALTH] source=%s category=%s skipped=true disabledUntil=%s consecutiveFails=%d lastErrorType=%s\n",
				sourceName,
				source.Category,
				status.DisabledUntil.Format(time.RFC3339),
				status.ConsecutiveFails,
				status.LastErrorType,
			)
			s.recordHealthEvent(status, "disabled")
			return
		}
	}

	feed, err := s.fetchWithRetry(source)
	if err != nil {
		errType := classifyRSSError(err)

		if s.HealthManager != nil {
			status := s.HealthManager.RecordFailure(source, sourceName, errType, err.Error())
			fmt.Printf("[RSS ERROR] source=%s category=%s type=%s url=%s err=%v consecutiveFails=%d disabledUntil=%s\n",
				sourceName,
				source.Category,
				errType,
				source.URL,
				err,
				status.ConsecutiveFails,
				status.DisabledUntil.Format(time.RFC3339),
			)
			s.recordHealthEvent(status, "degraded")
		} else {
			fmt.Printf("[RSS ERROR] source=%s category=%s type=%s url=%s err=%v\n",
				sourceName,
				source.Category,
				errType,
				source.URL,
				err,
			)
		}

		return
	}

	if s.HealthManager != nil {
		s.HealthManager.RecordSuccess(source, sourceName)
		_, status := s.HealthManager.ShouldSkip(source, sourceName)
		s.recordHealthEvent(status, "healthy")
	}

	sort.Slice(feed.Items, func(i, j int) bool {
		var t1, t2 time.Time
		if feed.Items[i].PublishedParsed != nil {
			t1 = *feed.Items[i].PublishedParsed
		} else if feed.Items[i].UpdatedParsed != nil {
			t1 = *feed.Items[i].UpdatedParsed
		}
		if feed.Items[j].PublishedParsed != nil {
			t2 = *feed.Items[j].PublishedParsed
		} else if feed.Items[j].UpdatedParsed != nil {
			t2 = *feed.Items[j].UpdatedParsed
		}
		return t1.After(t2)
	})

	count := 0
	for _, item := range feed.Items {
		if count >= s.MaxPerSource {
			fmt.Printf("Kaynak limiti doldu (%d/%d): %s\n", count, s.MaxPerSource, feed.Title)
			break
		}

		if s.Cache.IsDuplicate(item.Link) {
			continue
		}
		if s.Cache.IsTitleDuplicate(item.Title) {
			fmt.Printf("Benzer haber pas geçildi: %s\n", item.Title)
			continue
		}

		var publishedAt time.Time
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			publishedAt = *item.UpdatedParsed
		} else {
			publishedAt = time.Now()
		}

		raw := models.RawNewsItem{
			Title:       item.Title,
			Description: item.Description,
			Link:        item.Link,
			Source:      feed.Title,
			Category:    source.Category,
			PublishedAt: publishedAt,
			FetchedAt:   time.Now(),
		}
		raw.ID = raw.BuildID()

		env := models.NewEnvelope(raw)

		boost, sourceCount, clusterKey := s.Clusterer.AddEvent(raw)
		env.Cluster = models.ClusterInfo{
			ClusterKey:   clusterKey,
			ClusterCount: sourceCount,
			IsClustered:  sourceCount > 1,
		}
		env.Score.Boost = boost
		env.Stage = models.StageClustered

		allowed, reason := s.Filter.ShouldProcess(env)
		env.Filter = models.FilterDecision{
			Passed: allowed,
			Reason: reason,
		}

		if !allowed {
			fmt.Printf("[FİLTRE] Elendi (%s): %s\n", reason, item.Title)
			s.recordRejected(env, reason)
			continue
		}

		catPolicy := policy.Get(env.News.Category)
		if !policy.IsFreshEnough(env, catPolicy) {
			fmt.Printf("[FRESHNESS FILTER] Çok eski haber, atıldı: %s\n", env.News.Title)
			s.recordRejected(env, "freshness-filter")
			continue
		}

		if env.News.Category == models.CategoryBreaking && env.Cluster.ClusterCount < 2 {
			fmt.Printf("[HARD FILTER] BREAKING için yetersiz kaynak (%d < 2): %s\n",
				env.Cluster.ClusterCount, env.News.Title)
			s.recordRejected(env, "breaking-min-cluster")
			continue
		}

		fmt.Printf("[FİLTRE] Geçti (%s, %d kaynak): %s\n", reason, sourceCount, item.Title)

		if s.routeToChannel(env, item.Title, count) {
			count++
		}
	}
}

func (s *RSSScraper) fetchWithRetry(source config.RSSSource) (*gofeed.Feed, error) {
	var lastErr error

	for attempt := 1; attempt <= 3; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		feed, err := s.Parser.ParseURLWithContext(source.URL, ctx)
		cancel()

		if err == nil {
			if attempt > 1 {
				fmt.Printf("[RSS RETRY OK] source=%s category=%s attempt=%d url=%s\n",
					feedSourceName(source.URL), source.Category, attempt, source.URL)
			}
			return feed, nil
		}

		lastErr = err
		errType := classifyRSSError(err)

		fmt.Printf("[RSS RETRY] source=%s category=%s attempt=%d/3 type=%s url=%s err=%v\n",
			feedSourceName(source.URL),
			source.Category,
			attempt,
			errType,
			source.URL,
			err,
		)

		if errType == "HTTP_404" || errType == "HTTP_403" {
			fmt.Printf("[RSS SKIP RETRY] Kalıcı HTTP hatası (%s), retry atlanıyor: %s\n", errType, source.URL)
			break
		}

		if attempt < 3 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
	}

	return nil, lastErr
}

func (s *RSSScraper) recordRejected(env models.NewsEnvelope, reason string) {
	if s.Monitoring == nil {
		return
	}

	s.Monitoring.RecordRejected(monitoring.RejectedNewsEvent{
		Time:     time.Now(),
		Title:    env.News.Title,
		Category: string(env.News.Category),
		Source:   env.News.Source,
		Reason:   reason,
	})
}

func (s *RSSScraper) recordHealthEvent(status sourcehealth.Status, state string) {
	if s.Monitoring == nil {
		return
	}

	s.Monitoring.RecordSourceHealth(monitoring.SourceHealthEvent{
		Time:             time.Now(),
		SourceName:       status.SourceName,
		URL:              status.URL,
		Category:         string(status.Category),
		State:            state,
		ConsecutiveFails: status.ConsecutiveFails,
		LastErrorType:    status.LastErrorType,
		LastErrorMessage: status.LastErrorMessage,
		DisabledUntil:    formatTime(status.DisabledUntil),
		LastSuccessAt:    formatTime(status.LastSuccessAt),
	})
}

func classifyRSSError(err error) string {
	msg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(msg, "404") || strings.Contains(msg, "not found"):
		return "HTTP_404"
	case strings.Contains(msg, "403") || strings.Contains(msg, "forbidden"):
		return "HTTP_403"
	case strings.Contains(msg, "401") || strings.Contains(msg, "unauthorized"):
		return "HTTP_401"
	case strings.Contains(msg, "500") || strings.Contains(msg, "internal server"):
		return "HTTP_500"
	case strings.Contains(msg, "invalid utf-8"):
		return "INVALID_UTF8"
	case strings.Contains(msg, "no such host"):
		return "DNS_ERROR"
	case strings.Contains(msg, "eof"):
		return "EOF"
	case strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline exceeded"):
		return "TIMEOUT"
	case strings.Contains(msg, "connection refused"):
		return "CONN_REFUSED"
	default:
		return "OTHER"
	}
}

func feedSourceName(url string) string {
	switch {
	case strings.Contains(url, "trthaber.com"):
		return "TRT Haber"
	case strings.Contains(url, "bloomberght.com"):
		return "BloombergHT"
	case strings.Contains(url, "aa.com.tr"):
		return "Anadolu Ajansı"
	case strings.Contains(url, "webtekno.com"):
		return "Webtekno"
	case strings.Contains(url, "t24.com.tr"):
		return "T24"
	case strings.Contains(url, "ntv.com.tr"):
		return "NTV"
	case strings.Contains(url, "cumhuriyet.com"):
		return "Cumhuriyet"
	case strings.Contains(url, "haberturk.com"):
		return "Habertürk"
	case strings.Contains(url, "bbci.co.uk"):
		return "BBC"
	case strings.Contains(url, "nytimes.com"):
		return "NYT"
	case strings.Contains(url, "npr.org"):
		return "NPR"
	case strings.Contains(url, "theguardian.com"):
		return "The Guardian"
	case strings.Contains(url, "aljazeera.com"):
		return "Al Jazeera"
	case strings.Contains(url, "bloomberg.com"):
		return "Bloomberg"
	case strings.Contains(url, "marketwatch.com"):
		return "MarketWatch"
	case strings.Contains(url, "cnbc.com"):
		return "CNBC"
	case strings.Contains(url, "ft.com"):
		return "FT"
	case strings.Contains(url, "techcrunch.com"):
		return "TechCrunch"
	case strings.Contains(url, "theverge.com"):
		return "The Verge"
	case strings.Contains(url, "arstechnica.com"):
		return "Ars Technica"
	case strings.Contains(url, "skynews.com"):
		return "Sky News"
	case strings.Contains(url, "dw.com"):
		return "Deutsche Welle"
	case strings.Contains(url, "politico.eu"):
		return "Politico EU"
	default:
		return "Unknown"
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

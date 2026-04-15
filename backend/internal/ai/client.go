package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/models"
)

type EditorialRequest struct {
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	URL         string     `json:"url"`
	Source      string     `json:"source"`
	Category    string     `json:"category"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

type EditorialResponse struct {
	Decision     string `json:"decision"`
	RejectReason string `json:"reject_reason,omitempty"`

	Message    string `json:"message,omitempty"`
	Hook       string `json:"hook,omitempty"`
	Summary    string `json:"summary,omitempty"`
	Importance string `json:"importance,omitempty"`
	SourceLine string `json:"source_line,omitempty"`
	Sentiment  string `json:"sentiment,omitempty"`
	NewsType   string `json:"news_type,omitempty"`
}

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(apiURL string) *Client {
	return &Client{
		BaseURL: apiURL,
		HTTPClient: &http.Client{
			Timeout: 180 * time.Second,
		},
	}
}

func (c *Client) AnalyzeEditorial(env models.NewsEnvelope) (*models.EditorialDecision, *EditorialResponse, error) {
	var pubAt *time.Time
	if !env.News.PublishedAt.IsZero() {
		pubAt = &env.News.PublishedAt
	}

	reqBody := EditorialRequest{
		Title:       env.News.Title,
		Content:     env.News.Description,
		URL:         env.News.Link,
		Source:      env.News.Source,
		Category:    string(env.News.Category),
		PublishedAt: pubAt,
	}

	jsonValue, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("istek JSON'a çevrilemedi: %v", err)
	}

	resp, err := c.HTTPClient.Post(c.BaseURL+"/generate-message", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, nil, fmt.Errorf("AI servisine ulaşılamadı: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 2048))
		if err != nil {
			return nil, nil, fmt.Errorf("AI servisi hata döndü [%d]: (body okunamadı: %v)", resp.StatusCode, err)
		}
		return nil, nil, fmt.Errorf("AI servisi hata döndü [%d]: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	var out EditorialResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, nil, fmt.Errorf("cevap okunamadı: %v", err)
	}

	if out.Decision == "" {
		return nil, nil, fmt.Errorf("AI cevabında decision alanı yok")
	}

	decision := &models.EditorialDecision{
		ID:              env.News.ID,
		NewsID:          env.News.ID,
		Decision:        models.EditorialDecisionType(out.Decision),
		RejectReason:    out.RejectReason,
		NewsType:        out.NewsType,
		Sentiment:       out.Sentiment,
		Hook:            out.Hook,
		Summary:         out.Summary,
		Importance:      out.Importance,
		SourceLine:      out.SourceLine,
		ApprovalStatus:  models.ApprovalPending,
		ApprovalChannel: "telegram",
		CreatedAt:       time.Now(),
	}

	return decision, &out, nil
}

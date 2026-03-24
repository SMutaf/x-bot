package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type MessageRequest struct {
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	URL         string     `json:"url"`
	Source      string     `json:"source"`
	Category    string     `json:"category"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

type MessageResponse struct {
	Message    string `json:"message"`
	Hook       string `json:"hook"`
	Summary    string `json:"summary"`
	Importance string `json:"importance"`
	SourceLine string `json:"source_line"`
	Sentiment  string `json:"sentiment"`
	NewsType   string `json:"news_type"`
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

func (c *Client) GenerateTelegramPost(title, content, url, source, category string, publishedAt time.Time) (*MessageResponse, error) {
	var pubAt *time.Time
	if !publishedAt.IsZero() {
		pubAt = &publishedAt
	}

	reqBody := MessageRequest{
		Title:       title,
		Content:     content,
		URL:         url,
		Source:      source,
		Category:    category,
		PublishedAt: pubAt,
	}

	jsonValue, _ := json.Marshal(reqBody)

	resp, err := c.HTTPClient.Post(c.BaseURL+"/generate-message", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, fmt.Errorf("AI servisine ulaşılamadı: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI servisi hata döndü: %d", resp.StatusCode)
	}

	var out MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("cevap okunamadı: %v", err)
	}

	return &out, nil
}

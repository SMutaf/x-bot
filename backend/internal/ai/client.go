package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

	jsonValue, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("istek JSON'a çevrilemedi: %v", err)
	}

	resp, err := c.HTTPClient.Post(c.BaseURL+"/generate-message", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, fmt.Errorf("AI servisine ulaşılamadı: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 2048))
		if err != nil {
			return nil, fmt.Errorf("AI servisi hata döndü [%d]: (body okunamadı: %v)", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("AI servisi hata döndü [%d]: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}
	var out MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("cevap okunamadı: %v", err)
	}

	if out.Decision == "" {
		return nil, fmt.Errorf("AI cevabında decision alanı yok")
	}

	return &out, nil
}

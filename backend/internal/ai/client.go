package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type TweetRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	URL      string `json:"url"`
	Source   string `json:"source"`
	Category string `json:"category"` // BREAKING / TECH / GENERAL
}

type TweetResponse struct {
	Tweet     string `json:"tweet"`
	Reply     string `json:"reply"`
	Sentiment string `json:"sentiment"`
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

func (c *Client) GenerateTweet(title, content, url, source, category string) (*TweetResponse, error) {
	reqBody := TweetRequest{
		Title:    title,
		Content:  content,
		URL:      url,
		Source:   source,
		Category: category,
	}

	jsonValue, _ := json.Marshal(reqBody)

	resp, err := c.HTTPClient.Post(c.BaseURL+"/generate-tweet", "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, fmt.Errorf("AI servisine ulaşılamadı: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI servisi hata döndü: %d", resp.StatusCode)
	}

	var tweetResp TweetResponse
	if err := json.NewDecoder(resp.Body).Decode(&tweetResp); err != nil {
		return nil, fmt.Errorf("cevap okunamadı: %v", err)
	}

	return &tweetResp, nil
}

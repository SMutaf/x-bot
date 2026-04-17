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

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 180 * time.Second,
		},
	}
}

func (c *Client) Analyze(req models.EditorialAnalysisRequest) (*models.EditorialAnalysisResponse, error) {
	url := c.baseURL + "/analyze"

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("request marshal hatası: %w", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("AI servisine istek atılamadı: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("AI servis hatası [%d]: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	}

	var result models.EditorialAnalysisResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("response decode hatası: %w", err)
	}

	return &result, nil
}

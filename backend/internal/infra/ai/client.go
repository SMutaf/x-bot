package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/domain/models"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	inFlight   atomic.Int32
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
	c.inFlight.Add(1)
	defer c.inFlight.Add(-1)

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

func (c *Client) HealthCheck() error {
	client := &http.Client{Timeout: 2 * time.Second}

	resp, err := client.Get(c.baseURL + "/")
	if err != nil {
		return fmt.Errorf("AI servis health check hatasi: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("AI servis health check status: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) InFlight() int {
	return int(c.inFlight.Load())
}

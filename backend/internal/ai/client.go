package ai

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/SMutaf/twitter-bot/backend/internal/models"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

func (c *Client) Analyze(req models.EditorialAnalysisRequest) (*models.EditorialAnalysisResponse, error) {
	url := c.baseURL + "/analyze"

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result models.EditorialAnalysisResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

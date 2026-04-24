package translation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type LibreTranslator struct {
	baseURL string
	client  *http.Client
}

func NewLibreTranslator(baseURL string) *LibreTranslator {
	return &LibreTranslator{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type libreRequest struct {
	Q      string `json:"q"`
	Source string `json:"source"`
	Target string `json:"target"`
	Format string `json:"format"`
}

type libreResponse struct {
	TranslatedText string `json:"translatedText"`
}

func (t *LibreTranslator) Translate(text, source, target string) (string, error) {
	reqBody := libreRequest{
		Q:      text,
		Source: source,
		Target: target,
		Format: "text",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := t.baseURL + "/translate"

	resp, err := t.client.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("translate failed, status: %d", resp.StatusCode)
	}

	var result libreResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.TranslatedText, nil
}

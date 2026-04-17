package models

import "time"

type EditorialAnalysisRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Source      string    `json:"source"`
	PublishedAt time.Time `json:"published_at"`

	ClusterCount int `json:"cluster_count"`
	Virality     int `json:"virality"`
}

type EditorialAnalysisResponse struct {
	Decision     string `json:"decision"` // PUBLISH | REJECT
	RejectReason string `json:"reject_reason"`

	Summary    string `json:"summary"`
	Importance string `json:"importance"`
	Sentiment  string `json:"sentiment"`

	Hook string `json:"hook"`
}

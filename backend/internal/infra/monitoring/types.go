package monitoring

import "time"

type PublishedNewsEvent struct {
	Time         time.Time `json:"time"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	DescriptionTR string   `json:"descriptionTr"`
	Hook         string    `json:"hook"`
	Summary      string    `json:"summary"`
	Importance   string    `json:"importance"`
	Sentiment    string    `json:"sentiment"`
	Category     string    `json:"category"`
	Source       string    `json:"source"`
	Link         string    `json:"link"`
	Virality     int       `json:"virality"`
	ClusterCount int       `json:"clusterCount"`
}

type RejectedNewsEvent struct {
	Time     time.Time `json:"time"`
	Title    string    `json:"title"`
	Category string    `json:"category"`
	Source   string    `json:"source"`
	Reason   string    `json:"reason"`
}

type SourceHealthEvent struct {
	Time             time.Time `json:"time"`
	SourceName       string    `json:"sourceName"`
	URL              string    `json:"url"`
	Category         string    `json:"category"`
	State            string    `json:"state"`
	ConsecutiveFails int       `json:"consecutiveFails"`
	LastErrorType    string    `json:"lastErrorType"`
	LastErrorMessage string    `json:"lastErrorMessage"`
	DisabledUntil    string    `json:"disabledUntil"`
	LastSuccessAt    string    `json:"lastSuccessAt"`
}

type Summary struct {
	PublishedCount    int `json:"publishedCount"`
	RejectedCount     int `json:"rejectedCount"`
	HealthySources    int `json:"healthySources"`
	DisabledSources   int `json:"disabledSources"`
	DegradedSources   int `json:"degradedSources"`
	TrackedSourceSize int `json:"trackedSourceSize"`
}

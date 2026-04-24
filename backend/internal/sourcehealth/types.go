package sourcehealth

import (
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/models"
)

type Status struct {
	SourceName       string              `json:"sourceName"`
	URL              string              `json:"url"`
	Category         models.NewsCategory `json:"category"`
	ConsecutiveFails int                 `json:"consecutiveFails"`
	LastErrorType    string              `json:"lastErrorType"`
	LastErrorMessage string              `json:"lastErrorMessage"`
	LastErrorAt      time.Time           `json:"lastErrorAt"`
	LastSuccessAt    time.Time           `json:"lastSuccessAt"`
	DisabledUntil    time.Time           `json:"disabledUntil"`
}

func (s Status) IsDisabled(now time.Time) bool {
	return !s.DisabledUntil.IsZero() && now.Before(s.DisabledUntil)
}

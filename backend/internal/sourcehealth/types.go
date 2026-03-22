package sourcehealth

import (
	"time"

	"github.com/SMutaf/twitter-bot/backend/internal/models"
)

type Status struct {
	SourceName       string
	URL              string
	Category         models.NewsCategory
	ConsecutiveFails int
	LastErrorType    string
	LastErrorMessage string
	LastErrorAt      time.Time
	LastSuccessAt    time.Time
	DisabledUntil    time.Time
}

func (s Status) IsDisabled(now time.Time) bool {
	return !s.DisabledUntil.IsZero() && now.Before(s.DisabledUntil)
}

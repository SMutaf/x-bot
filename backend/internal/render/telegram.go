package render

import (
	"fmt"
	"strings"

	"github.com/SMutaf/twitter-bot/backend/internal/models"
)

type TelegramRenderer struct{}

func NewTelegramRenderer() *TelegramRenderer {
	return &TelegramRenderer{}
}

func (r *TelegramRenderer) Render(env models.NewsEnvelope, decision models.EditorialDecision) string {
	parts := make([]string, 0, 5)

	hook := strings.TrimSpace(decision.Hook)
	summary := strings.TrimSpace(decision.Summary)
	importance := strings.TrimSpace(decision.Importance)
	sourceLine := strings.TrimSpace(decision.SourceLine)
	link := strings.TrimSpace(env.News.Link)

	if hook != "" {
		parts = append(parts, fmt.Sprintf("**%s**", hook))
	}
	if summary != "" {
		parts = append(parts, summary)
	}
	if importance != "" {
		parts = append(parts, fmt.Sprintf("Önem: %s", importance))
	}
	if sourceLine != "" {
		parts = append(parts, sourceLine)
	}
	if link != "" {
		parts = append(parts, link)
	}

	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

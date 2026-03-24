package telegram

import (
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ApprovalBot struct {
	Bot    *tgbotapi.BotAPI
	ChatID int64
}

func NewApprovalBot(token string, chatID int64) *ApprovalBot {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panicf("Telegram bot başlatılamadı: %v", err)
	}
	return &ApprovalBot{Bot: bot, ChatID: chatID}
}

func categoryLabel(category string) string {
	switch category {
	case "BREAKING":
		return "🚨 SON DAKİKA"
	case "TECH":
		return "💻 TEKNOLOJİ"
	case "GENERAL":
		return "📰 GENEL"
	case "ECONOMY":
		return "💹 EKONOMİ"
	case "SPORTS":
		return "⚽ SPOR"
	default:
		return "📌 HABER"
	}
}

func escapeMarkdown(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
}

func (b *ApprovalBot) RequestApproval(message, category, publishedTime string) error {
	safeMessage := escapeMarkdown(message)
	safeTime := escapeMarkdown(publishedTime)

	finalText := fmt.Sprintf(
		"%s\n\n%s\n\n*⏰ Yayınlanma:* %s",
		categoryLabel(category),
		safeMessage,
		safeTime,
	)

	msg := tgbotapi.NewMessage(b.ChatID, finalText)
	msg.ParseMode = "MarkdownV2"
	msg.DisableWebPagePreview = false

	_, err := b.Bot.Send(msg)
	if err != nil {
		fmt.Printf("Telegram mesajı gönderilemedi: %v\n", err)
		return err
	}

	fmt.Printf("✅ Telegram'a gönderildi: %s\n", message[:min(50, len(message))])
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

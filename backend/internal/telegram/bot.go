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

// Kategori emojisi ve etiketi
func categoryLabel(category string) string {
	switch category {
	case "BREAKING":
		return "🚨 SON DAKİKA"
	case "TECH":
		return "💻 TEKNOLOJİ"
	case "GENERAL":
		return "📰 GENEL"
	default:
		return "📌 HABER"
	}
}

// ✅ Markdown özel karakterlerini escape et
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

func (b *ApprovalBot) RequestApproval(tweet, reply, source, category, publishedTime string) error {
	// Yayınlanma zamanı varsa ekle
	timeInfo := ""
	if publishedTime != "" {
		timeInfo = fmt.Sprintf("\n*⏰ Yayınlanma:* %s", escapeMarkdown(publishedTime))
	}

	// Tweet ve reply içeriğini escape et
	safeTweet := escapeMarkdown(tweet)
	safeReply := escapeMarkdown(reply)
	safeSource := escapeMarkdown(source)

	text := fmt.Sprintf(
		"%s\n\n"+
			"*Kaynak:* %s%s\n\n"+
			"*📝 Tweet:*\n%s\n\n"+
			"*🔗 Yanıt \\(Link\\):*\n%s\n\n"+
			"Onaylıyor musun?",
		categoryLabel(category), safeSource, timeInfo, safeTweet, safeReply,
	)

	msg := tgbotapi.NewMessage(b.ChatID, text)
	msg.ParseMode = "MarkdownV2" // ✅ MarkdownV2 kullan
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Onayla", "approve"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Reddet", "reject"),
		),
	)

	_, err := b.Bot.Send(msg)
	if err != nil {
		// Hata detayını logla
		fmt.Printf("Telegram Gönderim Hatası: %v\nMesaj: %s\n", err, text)
		return err
	}

	fmt.Printf("Telegram'a gönderildi: %s\n", tweet[:min(50, len(tweet))])
	return nil
}

func (b *ApprovalBot) ListenForApproval() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.Bot.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery == nil {
			continue
		}

		callback := update.CallbackQuery
		b.Bot.Request(tgbotapi.NewCallback(callback.ID, "İşlem yapılıyor..."))

		if callback.Data == "approve" {
			newText := callback.Message.Text + "\n\n✅ *ONAYLANDI VE PAYLAŞILDI\\!*"
			editMsg := tgbotapi.NewEditMessageText(b.ChatID, callback.Message.MessageID, newText)
			editMsg.ParseMode = "MarkdownV2"
			b.Bot.Send(editMsg)
			fmt.Println("🚀 Tweet onaylandı.")
		} else if callback.Data == "reject" {
			newText := callback.Message.Text + "\n\n❌ *REDDEDİLDİ\\.*"
			editMsg := tgbotapi.NewEditMessageText(b.ChatID, callback.Message.MessageID, newText)
			editMsg.ParseMode = "MarkdownV2"
			b.Bot.Send(editMsg)
			fmt.Println("🗑️ İçerik reddedildi.")
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

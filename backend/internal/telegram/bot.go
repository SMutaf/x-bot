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

// Kategori etiketi
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

// MarkdownV2 escape
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

func (b *ApprovalBot) RequestApproval(tweet, link, source, category, publishedTime string) error {

	// MESAJ → SADECE TWEET
	safeTweet := escapeMarkdown(tweet)

	tweetMsg := tgbotapi.NewMessage(b.ChatID, safeTweet)
	tweetMsg.ParseMode = "MarkdownV2"

	_, err := b.Bot.Send(tweetMsg)
	if err != nil {
		fmt.Printf("Tweet mesajı gönderilemedi: %v\n", err)
		return err
	}

	// MESAJ → META BİLGİLER + BUTONLAR

	safeSource := escapeMarkdown(source)
	safeLink := escapeMarkdown(link)
	safeTime := escapeMarkdown(publishedTime)

	metaText := fmt.Sprintf(
		"%s\n\n"+
			"*Kaynak:* %s\n"+
			"*⏰ Yayınlanma:* %s\n"+
			"*🔗 Link:*\n%s\n\n"+
			"Onaylıyor musun?",
		categoryLabel(category),
		safeSource,
		safeTime,
		safeLink,
	)

	metaMsg := tgbotapi.NewMessage(b.ChatID, metaText)
	metaMsg.ParseMode = "MarkdownV2"

	metaMsg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Onayla", "approve"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Reddet", "reject"),
		),
	)

	_, err = b.Bot.Send(metaMsg)
	if err != nil {
		fmt.Printf("Meta mesaj gönderilemedi: %v\n", err)
		return err
	}

	fmt.Println("Telegram'a 2 ayrı mesaj gönderildi.")
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

		originalText := callback.Message.Text

		var newText string

		if callback.Data == "approve" {
			newText = originalText + "\n\n✅ *ONAYLANDI VE PAYLAŞILDI\\!*"
			fmt.Println("🚀 Tweet onaylandı.")
		} else if callback.Data == "reject" {
			newText = originalText + "\n\n❌ *REDDEDİLDİ\\.*"
			fmt.Println("🗑️ İçerik reddedildi.")
		}

		editMsg := tgbotapi.NewEditMessageText(
			b.ChatID,
			callback.Message.MessageID,
			newText,
		)
		editMsg.ParseMode = "MarkdownV2"

		b.Bot.Send(editMsg)
	}
}

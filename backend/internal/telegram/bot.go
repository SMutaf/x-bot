package telegram

import (
	"fmt"
	"log"

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

func (b *ApprovalBot) RequestApproval(tweet, reply, source, category string) error {
	text := fmt.Sprintf(
		"%s\n\n"+
			"*Kaynak:* %s\n\n"+
			"*📝 Tweet:*\n%s\n\n"+
			"*🔗 Yanıt (Link):*\n%s\n\n"+
			"Onaylıyor musun?",
		categoryLabel(category), source, tweet, reply,
	)

	msg := tgbotapi.NewMessage(b.ChatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Onayla", "approve"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Reddet", "reject"),
		),
	)

	_, err := b.Bot.Send(msg)
	return err
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
			newText := callback.Message.Text + "\n\n✅ *ONAYLANDI VE PAYLAŞILDI!*"
			editMsg := tgbotapi.NewEditMessageText(b.ChatID, callback.Message.MessageID, newText)
			editMsg.ParseMode = "Markdown"
			b.Bot.Send(editMsg)
			fmt.Println("🚀 Tweet onaylandı.")
		} else if callback.Data == "reject" {
			newText := callback.Message.Text + "\n\n❌ *REDDEDİLDİ.*"
			editMsg := tgbotapi.NewEditMessageText(b.ChatID, callback.Message.MessageID, newText)
			editMsg.ParseMode = "Markdown"
			b.Bot.Send(editMsg)
			fmt.Println("🗑️ İçerik reddedildi.")
		}
	}
}

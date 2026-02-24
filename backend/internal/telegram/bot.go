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

// NewApprovalBot yeni bir Telegram bot istemcisi başlatır
func NewApprovalBot(token string, chatID int64) *ApprovalBot {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panicf("Telegram bot başlatılamadı: %v", err)
	}

	return &ApprovalBot{
		Bot:    bot,
		ChatID: chatID,
	}
}

// RequestApproval hazırlanan tweeti onay için Telegram'a gönderir
func (b *ApprovalBot) RequestApproval(tweet, reply, source string) error {
	// Mesaj metnini oluşturuyoruz
	text := fmt.Sprintf(
		"🔔 *YENİ TWEET ONAYI BEKLİYOR*\n\n"+
			"*Kaynak:* %s\n\n"+
			"*📝 Tweet:* \n%s\n\n"+
			"*🔗 Yanıt (Link):* \n%s\n\n"+
			"Onaylıyor musun?",
		source, tweet, reply,
	)

	msg := tgbotapi.NewMessage(b.ChatID, text)
	msg.ParseMode = "Markdown" // Kalın yazılar için

	// Onay ve Red butonlarını  (Inline Keyboard)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Onayla ve Paylaş", "approve"),
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

		// Butona basıldığında burası çalışır
		callback := update.CallbackQuery

		// Kullanıcıya "İşlem alınıyor" bildirimi gönderir
		callbackCfg := tgbotapi.NewCallback(callback.ID, "İşlem yapılıyor...")
		b.Bot.Request(callbackCfg)

		if callback.Data == "approve" {
			// BURASI GELECEKTE X (TWITTER) API'SİNİ ÇAĞIRACAK
			newText := callback.Message.Text + "\n\n✅ **BU TWEET ONAYLANDI VE PAYLAŞILDI!**"
			editMsg := tgbotapi.NewEditMessageText(b.ChatID, callback.Message.MessageID, newText)
			b.Bot.Send(editMsg)
			fmt.Println("🚀 Tweet onaylandı, X'e gönderiliyor (X API bekleniyor...)")
		} else if callback.Data == "reject" {
			newText := callback.Message.Text + "\n\n❌ **BU İÇERİK REDDEDİLDİ.**"
			editMsg := tgbotapi.NewEditMessageText(b.ChatID, callback.Message.MessageID, newText)
			b.Bot.Send(editMsg)
			fmt.Println("🗑️ İçerik reddedildi.")
		}
	}
}

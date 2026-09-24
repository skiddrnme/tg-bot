package telegram

import (
	"strings"

	"github.com/askemblerrr/pr-review-bot/internal/types"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ToDomain конвертирует апдейт Telegram в доменное сообщение.
// Возвращает nil, если апдейт нам не интересен.

func ToDomain(u tgbotapi.Update) *types.Message{
	if u.Message == nil {
        return nil
    }

    m := u.Message
    text := m.Text
    if text == "" {
        text = m.Caption
    }
    if text == "" {
        return nil
    }

    msg := &types.Message{
        Text:   text,
        ChatID: m.Chat.ID,
    }
    if m.From != nil {
        msg.UserID = m.From.ID
        msg.Username = m.From.UserName
    }

	  if m.IsCommand() {
        msg.Command = m.Command()
        msg.Args = strings.Fields(m.Text)[1:]
    }

	return msg
}
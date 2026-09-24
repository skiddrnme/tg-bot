package telegram

import (
	"context"

	"log/slog"

	
	"github.com/askemblerrr/pr-review-bot/internal/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)


type Client struct {
	bot    *tgbotapi.BotAPI
	logger *slog.Logger
}

func NewBot(bot *tgbotapi.BotAPI, logger *slog.Logger) *Client {
	return &Client{
		bot:    bot,
		logger: logger,
	}
}


// Updates возвращает канал доменных сообщений.
// GetUpdatesChan живёт ЗДЕСЬ, а не в delivery.
func (c *Client) Updates(ctx context.Context) <-chan *types.Message {
    out := make(chan *types.Message)

    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60
    raw := c.bot.GetUpdatesChan(u)

    go func() {
        defer close(out)
        for {
            select {
            case <-ctx.Done():
                c.bot.StopReceivingUpdates()
                return
            case upd, ok := <-raw:
                if !ok {
                    return
                }
                if msg := ToDomain(upd); msg != nil {
                    select {
                    case out <- msg:
                    case <-ctx.Done():
                        return
                    }
                }
            }
        }
    }()

    return out
}

// Send — исходящий вызов в Telegram API (для уведомлений).
func (c *Client) Send(chatID int64, text string) error {
    _, err := c.bot.Send(tgbotapi.NewMessage(chatID, text))
    return err
}

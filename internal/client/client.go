package client

import (
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// type TelegramClient interface {
// 	SendMessage(chatID int64, text string) error
// }

type Client struct {
	bot    *tgbotapi.BotAPI
	logger *slog.Logger
}

func NewBot(bot *tgbotapi.BotAPI, logger *slog.Logger) *Client{
	return &Client{
		bot: bot,
		logger: logger,
	}
}

func (c *Client) Start() error{
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := c.bot.GetUpdatesChan(u)
	
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		
		if update.Message.IsCommand() {
			if err := .handleCommand(update.Message); err != nil {
				b.handleError(update.Message.Chat.ID, err)
			}

			continue
		}

		
		if err := b.handleMessage(update.Message); err != nil {
			b.handleError(update.Message.Chat.ID, err)
		}
	}
}
package client

import (
	"fmt"
	"log/slog"

	"github.com/askemblerrr/pr-review-bot/internal/delivery/telegram"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// type TelegramClient interface {
// 	SendMessage(chatID int64, text string) error
// }

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

func (c *Client) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	handler := telegram.NewHandler(c.bot, c.logger)
	updates := c.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		if update.Message.IsCommand() {
			if err := handler.HandleCommand(update.Message); err != nil {
				fmt.Errorf(err.Error())
			}

			continue
		}

		if err := handler.HandleMessage(update.Message); err != nil {
			fmt.Errorf(err.Error())
		}
	}

	return nil
}

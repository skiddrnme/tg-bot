package telegram

import (
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot    *tgbotapi.BotAPI
	logger *slog.Logger
}

const (
	commandStart = "start"
)

func NewHandler(bot *tgbotapi.BotAPI, logger *slog.Logger) *Handler {
	return &Handler{
		bot:    bot,
		logger: logger,
	}
}

func (h *Handler) HandleUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			h.HandleCommand(update.Message)
			continue
		}

		h.HandleMessage(update.Message)
	}

}

func (h *Handler) HandleMessage(message *tgbotapi.Message) error {
	h.logger.Info("[%s] %s", message.From.UserName, message.Text)

	msg := tgbotapi.NewMessage(message.Chat.ID, message.Text)
	_, err := h.bot.Send(msg)
	return err
}

func (h *Handler) HandleCommand(message *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Я не знаю такой команды")

	switch message.Command() {
	case commandStart:
		msg.Text = "Ты ввел команду /start"
		_, err := h.bot.Send(msg)
		return err
	default:
		_, err := h.bot.Send(msg)
		return err
	}
}

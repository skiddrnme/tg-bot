package config

import (
	"errors"
	"log/slog"
	"os"

	"github.com/askemblerrr/pr-review-bot/internal/validators"
)

type Config struct {
	TelegramBotToken string 
}

func LoadConfig(logger *slog.Logger) (*Config, error) {
	if logger == nil {
		logger = slog.Default()
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	if botToken == "" {
		logger.Error("Необходимые переменные окружения отсутствуют", "bot_token", maskToken(botToken))
		return nil, errors.New("необходимые переменные окружения отсутствуют")
	}

	if err := validators.ValidateBotToken(botToken); err != nil {
		logger.Error("Невалидный токен бота", "err", err)
		return nil, err
	}

	return &Config{
		TelegramBotToken: botToken,
	}, nil
}


// maskToken маскирует токен для безопасного логирования
func maskToken(token string) string {
	if token == "" {
		return "***"
	}
	if len(token) <= 8 {
		return "***"
	}

	return token[:4] + "***" + token[len(token)-4:]
}

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/askemblerrr/pr-review-bot/internal/config"
	"github.com/askemblerrr/pr-review-bot/internal/delivery/telegram"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

// setupLogger логгер
func setupLogger() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	return logger
}

func main() {
	logger := setupLogger()

	if err := godotenv.Load(); err != nil {
		logger.Warn("Файл .env не найден или не загружен")
	}

	_, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Загрузка конфигурации
	cfg, err := config.LoadConfig(logger)
	if err != nil {
		logger.Error("Ошибка загрузки конфигурации", "err", err)
		os.Exit(1)
	}

	botApi, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	

	botApi.Debug = true

	bot := telegram.NewBot(botApi, logger)

}

// 1. Загрузка конфигурации
// 2. Инициализация логгера
// 3. Инициализация репозитория
// 4. Инициализация клиентов
// 5. Инициализация сервисов
// 6. Настройка роутеров
// 7. Запуск HTTP сервера
// 8. Запуск Telegram бота (long polling)

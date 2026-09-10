package service

import (
	"github.com/askemblerrr/pr-review-bot/internal/repository"
	"log/slog"
)

type NotificationService struct {
	userRepo repository.UserRepository
	// tgClient telegram.TelegramClient
	logger   *slog.Logger
}

// func (s *NotificationService) NotifyPRAssigned(githubLogin, prTitle, prURL string) error {
// 	// 1. Найти telegramID по githubLogin
// 	// 2. Отправить сообщение
// }
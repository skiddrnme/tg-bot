package service

import (
	"github.com/askemblerrr/pr-review-bot/internal/repository"
	"log/slog"
)

type UserService struct {
	repo   repository.UserRepository
	// client telegram.TelegramClient
	logger *slog.Logger
}

// func (s *UserService) SetGitHubNick(telegramID int64, githubLogin string) error {
// 	// Валидация, сохранение
// }
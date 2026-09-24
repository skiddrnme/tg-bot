package service

import (
    "context"
    "errors"
    "fmt"
    "log/slog"

    "github.com/askemblerrr/pr-review-bot/internal/repository"
)

type Notifier interface {
    Send(chatID int64, text string) error
}

type NotificationService struct {
    userRepo repository.UserRepository
    notifier Notifier
    logger   *slog.Logger
}

func NewNotificationService(
    userRepo repository.UserRepository,
    notifier Notifier,
    logger *slog.Logger,
) *NotificationService {
    return &NotificationService{userRepo: userRepo, notifier: notifier, logger: logger}
}

func (s *NotificationService) NotifyPRAssigned(
    ctx context.Context,
    githubLogin, prTitle, prURL string,
) error {
    chatID, err := s.userRepo.GetTelegramID(ctx, githubLogin)
    if err != nil {
        if errors.Is(err, repository.ErrUserNotFound) {
            s.logger.Warn("no telegram user for github login", "github_login", githubLogin)
            return err
        }
        return fmt.Errorf("get telegram id: %w", err)
    }

    text := fmt.Sprintf("🔔 На вас назначен pull request:\n%s\n%s", prTitle, prURL)
    if err := s.notifier.Send(chatID, text); err != nil {
        return fmt.Errorf("send notification: %w", err)
    }
    s.logger.Info("notification sent", "github_login", githubLogin, "chat_id", chatID)
    return nil
}
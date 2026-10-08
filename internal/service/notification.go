package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/askemblerrr/pr-review-bot/internal/repository"
	"github.com/askemblerrr/pr-review-bot/internal/types"
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

func (s *NotificationService) NotifyPRAssigned(ctx context.Context, ev types.PRAssignedEvent) error {
	chatID, err := s.userRepo.GetTelegramID(ctx, ev.Assignee)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			s.logger.Warn("no telegram user for github login", "github_login", ev.Assignee)
			return err
		}
		return fmt.Errorf("get telegram id: %w", err)
	}

	text := fmt.Sprintf("🔔 На вас назначен pull request:\n%s\n%s", ev.PRTitle, ev.PRURL)
	if err := s.notifier.Send(chatID, text); err != nil {
		return fmt.Errorf("send notification: %w", err)
	}
	s.logger.Info("notification sent", "github_login", ev.Assignee, "chat_id", chatID)
	return nil
}

func (s *NotificationService) NotifyPRReview(ctx context.Context, ev types.PRReviewEvent) error {
	chatID, err := s.userRepo.GetTelegramID(ctx, ev.AuthorLogin)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			s.logger.Warn("no telegram user for github login", "github_login", ev.AuthorLogin)
			return err
		}
		return fmt.Errorf("get telegram id: %w", err)
	}

	var header string
	switch ev.Action {
	case "submitted":
		header = "💬 Новое ревью на ваш PR"
	case "dismissed":
		header = "⚠️ Ревью снято с вашего PR"
	}

	var status string
	switch ev.ReviewState {
	case "approved":
		status = "✅ Одобрено"
	case "changes_requested":
		status = "❌ Запрошены изменения"
	case "commented":
		status = "💬 Комментарий"
	default:
		status = ev.ReviewState
	}

	var body string
	runes := []rune(ev.ReviewBody)
	switch {
	case ev.ReviewBody == "":
		body = "Без комментария"
	case len(ev.ReviewBody) > 1000:
		body = string(runes[:1000]) + "... (обрезано)"
	default:
		body = ev.ReviewBody
	}

	text := fmt.Sprintf("%s: %s\nОт: @%s\nСтатус: %s\nКомментарий: %s\n%s",
		header,
		ev.PRTitle,
		ev.ReviewerLogin,
		status,
		body,
		ev.PRURL,
	)

	if err := s.notifier.Send(chatID, text); err != nil {
		return fmt.Errorf("send notification: %w", err)
	}
	s.logger.Info("notification sent", "github_login", ev.AuthorLogin, "chat_id", chatID)
	return nil
}

func (s *NotificationService) NotifyPRReviewComment(ctx context.Context, ev types.PRReviewCommentEvent) error {
	chatID, err := s.userRepo.GetTelegramID(ctx, ev.AuthorLogin)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			s.logger.Warn("no telegram user for github login", "github_login", ev.AuthorLogin)
			return err
		}
		return fmt.Errorf("get telegram id: %w", err)
	}

	var body string
	runes := []rune(ev.CommentBody)
	switch {
	case ev.CommentBody == "":
		return nil
	case len(ev.CommentBody) > 1000:
		body = string(runes[:1000]) + "... (обрезано)"
	default:
		body = ev.CommentBody
	}

	text := fmt.Sprintf("Новый комментарий к вашему PR: %s\nОт: @%s\nФайл: %s :%d\nКомментарий: %s\n%s",
		ev.PRTitle,
		ev.CommenterLogin,
		ev.FilePath,
		body,
		ev.CommentURL,
	)
	if err := s.notifier.Send(chatID, text); err != nil {
		return fmt.Errorf("send notification: %w", err)
	}
	s.logger.Info("notification sent", "github_login", ev.AuthorLogin, "chat_id", chatID)
	return nil

}
